# 自动扩展配额管理说明

## 概述

在 ES Serverless 平台中，自动扩展会检查租户配额管理，确保资源使用不会超出租户的配额限制。

## 配额检查机制

### 1. 检查时机

自动扩展会**在每次扩展操作前**检查租户配额：

1. **创建新集群时** ✅（已实现）
2. **自动扩展容器副本时** ✅（新增功能）

### 2. 检查内容

检查以下配额限制：
- 最大索引数（MaxIndices）
- 最大存储空间（MaxStorage）- 未来扩展
- 最大计算资源（CPU/Memory）- 未来扩展

### 3. 检查逻辑

```go
// 扩展前检查配额（仅扩容时检查）
if newReplicas > currentReplicas && userID != "" {
    hasQuota, quota, err := metadataService.CheckTenantQuota(userID)
    if err != nil {
        log.Printf("Warning: Failed to check tenant quota for user %s: %v", userID, err)
    } else if !hasQuota {
        log.Printf("Tenant quota exceeded for user %s. Max indices: %d, Current indices: %d", 
            userID, quota.MaxIndices, quota.CurrentIndices)
        return  // 阻止扩展
    }
}
```

## 实现细节

### 文件位置
- `/server/autoscaler.go` - 第180-230行

### 核心函数
- `scaleNamespace()` - 扩展命名空间函数

### 检查流程

```
开始扩展 → 检查指标 → 计算新副本数 → 
    ↓
新副本数 > 当前副本数？ → 是 → 检查租户配额 → 
    ↓                        ↓
   否                       配额充足？ → 是 → 执行扩展
    ↓                        ↓
执行扩展                    否 → 阻止扩展
```

## API 接口

### 查询租户配额
```bash
curl -X GET http://localhost:8080/metadata/tenants/{tenant_id}
```

### 更新租户配额
```bash
curl -X PUT http://localhost:8080/metadata/tenants/{tenant_id} \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "testuser",
    "max_indices": 10,
    "max_storage": "100Gi",
    "current_indices": 2,
    "current_storage": "20Gi"
  }'
```

## 日志记录

### 配额检查成功
```
2024/01/01 10:00:00 Updated tenant quota check for user testuser after scaling
```

### 配额超出
```
2024/01/01 10:00:00 Tenant quota exceeded for user testuser. Max indices: 10, Current indices: 10
```

### 配额检查失败
```
2024/01/01 10:00:00 Warning: Failed to check tenant quota for user testuser: error message
```

## 配置参数

### 默认配额限制
```go
MaxReplicas: 10    // 最大副本数
MinReplicas: 1     // 最小副本数
ScaleUpFactor: 1.5 // 扩容因子
ScaleDownFactor: 0.5 // 缩容因子
```

### 用户自定义配额
```go
type ScalingPolicy struct {
    UserID              string  `json:"user_id"`
    EnableAutoScaleUp   bool    `json:"enable_auto_scale_up"`
    EnableAutoScaleDown bool    `json:"enable_auto_scale_down"`
    ScaleUpThreshold    float64 `json:"scale_up_threshold"`
    ScaleDownThreshold  float64 `json:"scale_down_threshold"`
    MaxReplicas         int     `json:"max_replicas"`  // 用户特定最大副本数
    MinReplicas         int     `json:"min_replicas"`  // 用户特定最小副本数
}
```

## 测试场景

### 场景1：正常扩展
1. 租户配额充足
2. 系统负载增加
3. 自动扩展副本数
4. 配额检查通过
5. 扩展成功执行

### 场景2：配额超出阻止扩展
1. 租户已达到最大索引数
2. 系统负载增加
3. 自动扩展尝试增加副本
4. 配额检查失败
5. 扩展被阻止，记录日志

### 场景3：配额检查失败
1. 元数据服务不可用
2. 配额检查失败
3. 系统记录警告日志
4. 扩展继续执行（容错机制）

## 未来扩展

### 1. 更精确的资源跟踪
```go
// 计划实现
func (ms *MetadataService) UpdateTenantResourceUsage(tenantID string, cpu, memory, storage string) error
```

### 2. 实时配额监控
- 每分钟检查配额使用情况
- 提供配额使用率仪表板
- 配额接近限制时发送告警

### 3. 配额策略
```json
{
  "tenant_id": "org-001",
  "quota": {
    "max_clusters": 10,
    "max_cpu": "100",
    "max_memory": "200Gi",
    "max_disk": "1Ti",
    "max_gpu": 5
  }
}
```

## 相关文档

- [多租户架构说明](/docs/多租户架构说明.md)
- [分片数据同步实现进度](/docs/分片数据同步实现进度.md)
- [自动扩缩容服务](/server/autoscaler.go)
