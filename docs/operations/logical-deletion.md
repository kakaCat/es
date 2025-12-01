# 逻辑删除实现说明

## 概述

为了保证数据的安全性和可恢复性，我们将租户容器的删除操作从物理删除改为逻辑删除。逻辑删除不会真正从文件系统中删除文件，而是通过设置一个删除标记来标识记录已被删除。

## 实现细节

### 1. 数据结构修改

在 `TenantContainer` 结构体中添加了一个 `Deleted` 字段：

```go
type TenantContainer struct {
    // ... 其他字段 ...
    Deleted bool `json:"deleted"` // 逻辑删除标记
}
```

### 2. 删除操作修改

原来的物理删除：
```go
func (ms *MetadataService) DeleteTenantContainer(user, serviceName string) error {
    filename := filepath.Join(ms.dataDir, fmt.Sprintf("tenant_%s_%s.json", user, serviceName))
    return os.Remove(filename)  // 直接删除文件
}
```

修改为逻辑删除：
```go
func (ms *MetadataService) DeleteTenantContainer(user, serviceName string) error {
    // 读取现有租户容器信息
    filename := filepath.Join(ms.dataDir, fmt.Sprintf("tenant_%s_%s.json", user, serviceName))
    file, err := os.ReadFile(filename)
    if err != nil {
        return err
    }
    
    var container TenantContainer
    if err := json.Unmarshal(file, &container); err != nil {
        return err
    }
    
    // 设置逻辑删除标记
    container.Deleted = true
    container.Status = "deleted"
    container.SyncTime = time.Now()
    
    // 保存更新后的租户容器信息
    updatedFile, err := json.MarshalIndent(container, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, updatedFile, 0644)
}
```

### 3. 查询过滤

在所有查询租户容器的方法中添加了对已删除记录的过滤：

```go
// 在 GetTenantContainer 中
if container.Deleted {
    return nil, fmt.Errorf("tenant container for user %s and service %s has been deleted", user, serviceName)
}

// 在 ListTenantContainers 中
if !container.Deleted {
    containers = append(containers, &container)
}

// 在 ListTenantContainersByOrgID 中
if container.TenantOrgID == tenantOrgID && !container.Deleted {
    containers = append(containers, &container)
}
```

### 4. 创建时的初始化

在创建新的租户容器时，确保 `Deleted` 字段被初始化为 `false`：

```go
tenantContainer := &TenantContainer{
    // ... 其他字段 ...
    Deleted: false, // 初始化为未删除状态
}
```

## 优势

1. **数据安全性**：避免误删重要数据
2. **可恢复性**：可以通过清除删除标记来恢复数据
3. **审计追踪**：保留删除操作的历史记录
4. **一致性**：保持文件系统的稳定性

## 注意事项

1. 磁盘空间：逻辑删除的记录仍占用磁盘空间
2. 查询性能：需要额外检查删除标记
3. 数据清理：需要定期清理长期未使用的已删除记录
