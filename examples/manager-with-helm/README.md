# Go 集成 Helm SDK 示例

本目录展示如何在 Go 应用中使用 Helm SDK 管理 Kubernetes 应用。

## 文件说明

- `main.go` - Helm SDK 基础使用示例
- `helm_manager.go` - 租户 Helm 管理器封装
- `api_handler.go` - HTTP API 处理器
- `go.mod` - Go 模块依赖

## 依赖安装

```bash
go mod download
```

## 核心依赖

```go
import (
    "helm.sh/helm/v3/pkg/action"      // Helm 操作
    "helm.sh/helm/v3/pkg/chart/loader" // Chart 加载
    "helm.sh/helm/v3/pkg/cli"          // CLI 环境
    "k8s.io/client-go/kubernetes"      // Kubernetes 客户端
)
```

## 使用方式

### 方式 1: 直接使用 Helm SDK

```go
package main

import (
    "helm.sh/helm/v3/pkg/action"
    "helm.sh/helm/v3/pkg/chart/loader"
    "helm.sh/helm/v3/pkg/cli"
)

func main() {
    settings := cli.New()
    actionConfig := new(action.Configuration)

    // 初始化
    actionConfig.Init(
        settings.RESTClientGetter(),
        "default",
        os.Getenv("HELM_DRIVER"),
        log.Printf,
    )

    // 安装 Chart
    client := action.NewInstall(actionConfig)
    client.Namespace = "default"
    client.ReleaseName = "my-release"

    chart, _ := loader.Load("/path/to/chart")
    values := map[string]interface{}{
        "replicaCount": 3,
    }

    release, _ := client.Run(chart, values)
}
```

### 方式 2: 使用封装的 TenantHelmManager

```go
manager := NewTenantHelmManager()

// 创建租户集群
req := &TenantClusterRequest{
    TenantOrgID: "org-001",
    User: "alice",
    ServiceName: "vector-search",
    Replicas: 3,
}

resp, err := manager.CreateTenantCluster(req)
```

### 方式 3: HTTP API

启动 API 服务器:

```bash
go run *.go
```

调用 API:

```bash
# 创建集群
curl -X POST http://localhost:8080/api/v1/tenant-clusters \
  -H 'Content-Type: application/json' \
  -d '{
    "tenant_org_id": "org-001",
    "user": "alice",
    "service_name": "vector-search",
    "replicas": 3
  }'

# 获取状态
curl http://localhost:8080/api/v1/tenant-clusters/org-001-alice-vector-search

# 扩容
curl -X POST http://localhost:8080/api/v1/tenant-clusters/org-001-alice-vector-search/scale \
  -d '{"replicas": 5}'
```

## 主要功能

### 1. 安装 Chart

```go
func (h *HelmClient) InstallChart(releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
    client := action.NewInstall(h.actionConfig)
    client.ReleaseName = releaseName
    client.Wait = true

    chart, _ := loader.Load(chartPath)
    return client.Run(chart, values)
}
```

### 2. 升级 Chart

```go
func (h *HelmClient) UpgradeChart(releaseName, chartPath string, values map[string]interface{}) (*release.Release, error) {
    client := action.NewUpgrade(h.actionConfig)
    chart, _ := loader.Load(chartPath)
    return client.Run(releaseName, chart, values)
}
```

### 3. 卸载 Chart

```go
func (h *HelmClient) UninstallChart(releaseName string) error {
    client := action.NewUninstall(h.actionConfig)
    _, err := client.Run(releaseName)
    return err
}
```

### 4. 列出 Releases

```go
func (h *HelmClient) ListReleases() ([]*release.Release, error) {
    client := action.NewList(h.actionConfig)
    return client.Run()
}
```

### 5. 回滚

```go
func (h *HelmClient) RollbackRelease(releaseName string, version int) error {
    client := action.NewRollback(h.actionConfig)
    client.Version = version
    return client.Run(releaseName)
}
```

## 集成到现有 Manager

### 替换 kubectl 命令

**之前 (kubectl)**:
```go
cmd := exec.Command("kubectl", "apply", "-f", "manifest.yaml")
cmd.Run()
```

**之后 (Helm SDK)**:
```go
helmManager := NewTenantHelmManager()
helmManager.CreateTenantCluster(req)
```

### 优势

1. **类型安全**: Go 类型检查,编译时发现错误
2. **更好的错误处理**: 详细的错误信息
3. **无需外部命令**: 不依赖 helm CLI
4. **性能更好**: 直接 API 调用,无进程开销
5. **易于测试**: 可以 mock Helm 操作

## 环境配置

### Kubeconfig

Helm SDK 使用 kubeconfig 连接 Kubernetes:

```go
// 默认使用 ~/.kube/config
settings := cli.New()

// 或指定路径
os.Setenv("KUBECONFIG", "/path/to/kubeconfig")
```

### Helm Driver

存储 release 信息的方式:

```bash
# ConfigMap (默认)
export HELM_DRIVER=configmap

# Secret (推荐生产环境)
export HELM_DRIVER=secret

# Memory (测试)
export HELM_DRIVER=memory
```

## 完整示例

参考 `server/` 目录中的实际集成示例:

```
server/
├── main.go                 # Manager 主程序
├── helm_integration.go     # Helm 集成模块
├── api_handlers.go         # API 处理器
└── cluster_manager.go      # 集群管理器
```

## 测试

```bash
# 运行示例
go run main.go

# 或使用特定文件
go run helm_manager.go api_handler.go
```

## 调试

启用详细日志:

```go
actionConfig.Init(
    settings.RESTClientGetter(),
    namespace,
    os.Getenv("HELM_DRIVER"),
    func(format string, v ...interface{}) {
        log.Printf("[HELM] " + format, v...) // 添加前缀
    },
)
```

## 常见问题

### Q: Helm SDK vs Terraform Helm Provider?

- **Helm SDK**: 在 Go 代码中动态管理
- **Terraform**: 声明式,适合基础设施

**推荐**: 两者结合
- Terraform: 部署平台基础设施
- Helm SDK: 运行时动态创建租户

### Q: Chart 路径如何指定?

```go
// 本地路径
chartPath := "./helm/elasticsearch"

// 或嵌入到二进制
//go:embed helm/elasticsearch
var chartFS embed.FS
```

### Q: 如何处理 values 文件?

```go
// 方式 1: map
values := map[string]interface{}{
    "replicaCount": 3,
}

// 方式 2: 从 YAML 加载
data, _ := os.ReadFile("values.yaml")
yaml.Unmarshal(data, &values)
```

## 生产环境建议

1. **使用 Secret 存储**: `HELM_DRIVER=secret`
2. **设置超时**: `client.Timeout = 300 * time.Second`
3. **启用等待**: `client.Wait = true`
4. **错误重试**: 实现重试逻辑
5. **并发控制**: 使用 semaphore 限制并发

## 参考资料

- [Helm Go SDK 文档](https://helm.sh/docs/topics/advanced/)
- [Helm Action Package](https://pkg.go.dev/helm.sh/helm/v3/pkg/action)
- [Kubernetes Client-Go](https://github.com/kubernetes/client-go)
