# Terraform vs Helm Go SDK 使用场景对比

## 概述

你的项目现在同时支持两种方式管理 Helm Charts:

1. **Terraform** - 通过 Terraform Helm Provider
2. **Helm Go SDK** - 在 Go 代码中直接调用

## 快速对比

| 特性 | Terraform | Helm Go SDK |
|------|-----------|-------------|
| 使用方式 | 声明式配置 | 编程式调用 |
| 状态管理 | Terraform State | Helm Release |
| 学习曲线 | 中等 | 低 (如果懂 Go) |
| 灵活性 | 中等 | 高 |
| 适用场景 | 基础设施部署 | 运行时动态管理 |
| 版本控制 | Git (配置文件) | Git (代码) |
| 预览变更 | `terraform plan` | 需要自己实现 |
| 回滚 | `terraform apply` | `helm rollback` |
| 性能 | 较慢 (需要 provider) | 快 (直接 API) |
| 依赖管理 | 自动 | 需要手动处理 |

## 使用场景

### ✅ 使用 Terraform 的场景

#### 1. 部署平台基础设施

```bash
# 一次性部署整个平台
make quick-start
```

**优势**:
- 声明式配置,易于理解
- 自动依赖管理
- 可以预览变更 (`terraform plan`)
- 状态管理
- 支持多环境

**示例**:
```hcl
# terraform/main.tf
module "elasticsearch" { ... }
module "control_plane" { ... }
module "monitoring" { ... }
```

#### 2. 长期存在的资源

如果资源需要长期存在且很少变更:
- 平台 Elasticsearch 集群
- Prometheus/Grafana 监控
- 基础网络和存储配置

#### 3. 团队协作

多人团队通过 Git 协作:
- Pull Request review 配置变更
- 统一的部署流程
- 避免手动操作错误

#### 4. 合规和审计

需要追踪所有变更:
- Git 历史记录所有变更
- Terraform State 保存完整状态
- 易于审计

### ✅ 使用 Helm Go SDK 的场景

#### 1. 动态创建租户集群

当用户通过 API 请求创建租户时:

```go
// POST /api/v1/tenant-clusters
func CreateTenantCluster(w http.ResponseWriter, r *http.Request) {
    manager := NewTenantHelmManager()
    resp, err := manager.CreateTenantCluster(req)
    // 立即返回结果
}
```

**优势**:
- 即时响应 API 请求
- 无需等待 Terraform
- 可以返回详细的进度信息
- 更好的错误处理

#### 2. 频繁的运行时操作

需要频繁操作的场景:
- 自动扩缩容
- 故障自愈
- 定期健康检查和修复

```go
// 自动扩容
if cpuUsage > 80% {
    helmManager.ScaleTenantCluster(namespace, currentReplicas + 1)
}
```

#### 3. 需要精细控制

需要在操作过程中进行判断:

```go
// 升级前检查
status := helmManager.GetTenantClusterStatus(namespace)
if status.Status == "deployed" {
    // 执行升级
    helmManager.UpgradeChart(...)
}
```

#### 4. 与现有 Go 应用集成

如果你的 Manager 服务是 Go 写的:
- 统一技术栈
- 无需外部工具
- 类型安全
- 易于测试

## 推荐架构: 混合使用

### 🎯 最佳实践: Terraform + Helm SDK

```
┌─────────────────────────────────────────────┐
│  Terraform (基础设施层)                       │
│  - 部署平台组件                               │
│  - 创建命名空间                               │
│  - 配置 RBAC                                 │
│  - 设置监控                                   │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│  Manager Service (Go + Helm SDK)            │
│  - 接收 API 请求                             │
│  - 动态创建租户集群                           │
│  - 运行时扩缩容                               │
│  - 自动故障恢复                               │
└─────────────────────────────────────────────┘
```

### 实现方式

#### 步骤 1: 使用 Terraform 部署平台

```bash
# 部署平台基础设施
cd terraform
terraform apply

# 部署后得到:
# - Namespace: es-serverless
# - Manager Service (运行中)
# - Monitoring Stack
# - RBAC 配置
```

#### 步骤 2: Manager 使用 Helm SDK 管理租户

```go
// server/main.go
package main

import (
    "github.com/your/manager/helm"
)

func main() {
    helmManager := helm.NewTenantHelmManager()

    http.HandleFunc("/api/v1/tenant-clusters", func(w http.ResponseWriter, r *http.Request) {
        // 动态创建租户集群
        resp, err := helmManager.CreateTenantCluster(req)
        json.NewEncoder(w).Encode(resp)
    })

    http.ListenAndServe(":8080", nil)
}
```

## 具体示例对比

### 示例 1: 创建租户集群

#### Terraform 方式

```bash
# 1. 创建配置文件
cat > terraform/tenants/org-alice-app1/main.tf <<EOF
module "tenant" {
  source = "../../modules/tenant"
  tenant_org_id = "org-001"
  user = "alice"
  service_name = "app1"
  replicas = 3
}
EOF

# 2. 初始化和应用
cd terraform/tenants/org-alice-app1
terraform init
terraform apply

# 耗时: ~2-3 分钟
```

**优点**: 配置即文档,易于版本控制
**缺点**: 需要创建文件,流程较长

#### Helm SDK 方式

```go
// API 调用
resp, err := helmManager.CreateTenantCluster(&TenantClusterRequest{
    TenantOrgID: "org-001",
    User: "alice",
    ServiceName: "app1",
    Replicas: 3,
})

// 耗时: ~1-2 分钟 (无需初始化)
```

**优点**: 即时响应,无需创建文件
**缺点**: 需要编写代码

### 示例 2: 扩容集群

#### Terraform 方式

```bash
# 1. 修改配置
vim terraform/tenants/org-alice-app1/main.tf
# 修改 replicas = 5

# 2. 预览和应用
terraform plan
terraform apply
```

**优点**: 可以预览变更
**缺点**: 需要手动修改文件

#### Helm SDK 方式

```go
// 直接调用
err := helmManager.ScaleTenantCluster("org-001-alice-app1", 5)
```

**优点**: 代码控制,可以自动化
**缺点**: 无内置预览功能

### 示例 3: 自动扩容

#### Terraform 方式

**不适合**: Terraform 不适合运行时自动化

#### Helm SDK 方式 ✅

```go
// 监控循环
go func() {
    for {
        metrics := getClusterMetrics(namespace)

        if metrics.CPUUsage > 80% {
            currentReplicas := getCurrentReplicas(namespace)
            helmManager.ScaleTenantCluster(namespace, currentReplicas + 1)
            log.Printf("Auto-scaled %s to %d replicas", namespace, currentReplicas + 1)
        }

        time.Sleep(30 * time.Second)
    }
}()
```

## 决策流程图

```
收到请求
    │
    ├─ 是基础设施部署? ─────────► 使用 Terraform
    │   (平台、监控、网络)
    │
    ├─ 是运行时操作? ───────────► 使用 Helm SDK
    │   (创建租户、扩容、修复)
    │
    ├─ 需要频繁变更? ───────────► 使用 Helm SDK
    │
    ├─ 需要审计和版本控制? ─────► 使用 Terraform
    │
    └─ 需要与 Go 应用集成? ─────► 使用 Helm SDK
```

## 实际项目中的应用

### 推荐分工

```
Terraform 负责:
├── 平台基础设施
│   ├── Namespace (es-serverless)
│   ├── Manager Deployment
│   ├── Prometheus + Grafana
│   └── RBAC 配置
│
└── 开发/测试租户 (预创建)
    ├── dev-team1-app1
    └── test-team2-app2

Manager (Helm SDK) 负责:
├── 生产租户 (动态创建)
│   ├── 用户请求创建
│   ├── API 驱动
│   └── 自动化管理
│
└── 运行时操作
    ├── 自动扩缩容
    ├── 故障恢复
    └── 健康检查
```

### 代码组织

```go
// server/cluster_manager.go
package main

type ClusterManager struct {
    helmManager *HelmManager
}

// 创建租户集群 (Helm SDK)
func (cm *ClusterManager) CreateTenantCluster(req *CreateClusterRequest) error {
    return cm.helmManager.CreateTenantCluster(req)
}

// 自动扩容 (Helm SDK)
func (cm *ClusterManager) AutoScale() {
    // 监控并自动扩容
}

// 平台初始化使用 Terraform (外部)
// terraform apply
```

## 迁移策略

### 从 kubectl 迁移到 Helm SDK

**之前**:
```go
cmd := exec.Command("kubectl", "apply", "-f", "tenant.yaml")
cmd.Run()
```

**之后**:
```go
helmManager.CreateTenantCluster(req)
```

**优势**:
- 去除外部依赖
- 更好的错误处理
- 类型安全
- 易于测试

### 保留 Terraform 用于平台

继续使用 Terraform 管理:
- 平台部署
- 监控配置
- 网络和存储
- RBAC 权限

## 总结

### 何时使用 Terraform

- ✅ 部署平台基础设施
- ✅ 长期存在的资源
- ✅ 需要版本控制和审计
- ✅ 团队协作
- ✅ 多环境管理 (dev/staging/prod)

### 何时使用 Helm Go SDK

- ✅ 动态创建租户集群
- ✅ 运行时自动化 (扩容、恢复)
- ✅ API 驱动的操作
- ✅ 需要精细控制
- ✅ 与 Go 应用深度集成

### 最佳实践

```
Terraform 管理 "基础设施"
    +
Helm SDK 管理 "工作负载"
    =
完美组合 🎯
```

## 下一步

1. **查看示例代码**: [examples/manager-with-helm/](../examples/manager-with-helm/)
2. **阅读 Helm SDK 文档**: [README.md](../examples/manager-with-helm/README.md)
3. **尝试集成**: 在你的 Manager 服务中集成 Helm SDK
4. **保持 Terraform**: 继续使用 Terraform 管理平台

两者结合,发挥各自优势! 🚀
