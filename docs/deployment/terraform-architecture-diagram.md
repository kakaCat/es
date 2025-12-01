# Terraform/Helm 架构图

## 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Terraform 管理层                                 │
│                                                                          │
│  terraform/                                                              │
│  ├── main.tf           (主配置,orchestrates所有模块)                      │
│  ├── variables.tf      (全局变量定义)                                     │
│  ├── outputs.tf        (输出服务 URLs)                                   │
│  └── modules/          (可复用模块)                                      │
│      ├── elasticsearch/                                                  │
│      ├── control-plane/                                                  │
│      ├── monitoring/                                                     │
│      ├── logging/                                                        │
│      └── tenant/       (多租户核心)                                      │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ deploys via
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Helm Charts 层                                  │
│                                                                          │
│  helm/                                                                   │
│  ├── elasticsearch/     (ES集群 + IVF插件)                               │
│  ├── control-plane/     (Manager + ShardController + Reporting)         │
│  └── monitoring/        (Prometheus + Grafana)                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ creates
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      Kubernetes 资源层                                   │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Namespace: es-serverless (平台命名空间)                          │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │   │
│  │  │ Elasticsearch│  │ Control Plane│  │  Monitoring  │          │   │
│  │  │  StatefulSet │  │  Deployments │  │  Deployments │          │   │
│  │  │  (3 replicas)│  │              │  │              │          │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘          │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Namespace: org-001-alice-vector-search (租户1)                   │   │
│  │  ┌──────────────┐  ┌──────────────┐                             │   │
│  │  │ Elasticsearch│  │  ConfigMap   │                             │   │
│  │  │  StatefulSet │  │  (metadata)  │                             │   │
│  │  └──────────────┘  └──────────────┘                             │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Namespace: org-002-bob-analytics (租户2)                         │   │
│  │  ┌──────────────┐  ┌──────────────┐                             │   │
│  │  │ Elasticsearch│  │  ConfigMap   │                             │   │
│  │  │  StatefulSet │  │  (metadata)  │                             │   │
│  │  └──────────────┘  └──────────────┘                             │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 部署流程

```
┌─────────────┐
│   开发者     │
└──────┬──────┘
       │
       │ 1. 编辑配置
       │    terraform.tfvars
       ▼
┌─────────────────┐
│ Terraform Init  │
│  - 下载providers │
│  - 初始化后端    │
└────────┬────────┘
         │
         │ 2. terraform plan
         ▼
┌─────────────────┐
│ Terraform Plan  │
│  - 计算变更     │
│  - 显示预览     │
└────────┬────────┘
         │
         │ 3. terraform apply (用户确认)
         ▼
┌─────────────────┐
│ Terraform Apply │
│  - 创建资源     │
│  - 调用Helm     │
└────────┬────────┘
         │
         ├───────────────┬───────────────┬───────────────┐
         │               │               │               │
         ▼               ▼               ▼               ▼
    ┌────────┐     ┌────────┐     ┌────────┐     ┌────────┐
    │ module │     │ module │     │ module │     │ module │
    │   ES   │     │control │     │monitor │     │logging │
    │        │     │ plane  │     │        │     │        │
    └───┬────┘     └───┬────┘     └───┬────┘     └───┬────┘
        │              │              │              │
        │ helm install │              │              │
        ▼              ▼              ▼              ▼
    ┌────────────────────────────────────────────────────┐
    │           Kubernetes API Server                    │
    │  - 创建 Namespaces                                 │
    │  - 部署 StatefulSets/Deployments                   │
    │  - 创建 Services                                   │
    │  - 分配 PersistentVolumeClaims                     │
    └────────────────────────────────────────────────────┘
                         │
                         ▼
    ┌────────────────────────────────────────────────────┐
    │           运行中的集群                              │
    │  ✓ Elasticsearch pods running                     │
    │  ✓ Manager API available                          │
    │  ✓ Prometheus collecting metrics                  │
    │  ✓ Grafana dashboards ready                       │
    └────────────────────────────────────────────────────┘
```

## 租户创建流程

```
┌───────────────┐
│ create-tenant │
│  script       │
└───────┬───────┘
        │
        │ 生成租户配置
        ▼
┌──────────────────────────┐
│ terraform/tenants/       │
│   org-user-service/      │
│     main.tf              │
└───────┬──────────────────┘
        │
        │ terraform init & apply
        ▼
┌──────────────────────────┐
│  Tenant Module           │
│  ┌──────────────────┐    │
│  │ 1. Namespace     │    │
│  │    + Labels      │    │
│  └──────────────────┘    │
│  ┌──────────────────┐    │
│  │ 2. Helm Release  │    │
│  │    (ES Cluster)  │    │
│  └──────────────────┘    │
│  ┌──────────────────┐    │
│  │ 3. ConfigMap     │    │
│  │    (metadata)    │    │
│  └──────────────────┘    │
│  ┌──────────────────┐    │
│  │ 4. ResourceQuota │    │
│  └──────────────────┘    │
│  ┌──────────────────┐    │
│  │ 5. NetworkPolicy │    │
│  └──────────────────┘    │
└───────┬──────────────────┘
        │
        ▼
┌──────────────────────────┐
│ Kubernetes Namespace     │
│ org-001-alice-vector     │
│                          │
│ Labels:                  │
│   es-cluster: "true"     │
│   tenant-org-id: "org-001"│
│   user: "alice"          │
│   service-name: "vector" │
│                          │
│ Resources:               │
│  - ES StatefulSet (3)    │
│  - ES Service            │
│  - PVCs (3x 20Gi)        │
│  - ConfigMap             │
│  - ResourceQuota         │
│  - NetworkPolicy         │
└──────────────────────────┘
```

## 模块依赖关系

```
main.tf
   │
   ├─► module.elasticsearch
   │      │
   │      └─► helm/elasticsearch chart
   │             │
   │             ├─► StatefulSet
   │             ├─► Service (ClusterIP + Headless)
   │             ├─► ConfigMap (ES config)
   │             └─► PVCs (per replica)
   │
   ├─► module.control_plane
   │      │
   │      └─► helm/control-plane chart
   │             │
   │             ├─► Manager Deployment + Service
   │             ├─► Shard Controller Deployment
   │             ├─► Reporting Deployment + Service
   │             ├─► ServiceAccount
   │             └─► RBAC (ClusterRole + Binding)
   │
   ├─► module.monitoring
   │      │
   │      └─► helm/monitoring chart
   │             │
   │             ├─► Prometheus Deployment + Service + PVC
   │             ├─► Grafana Deployment + Service + PVC
   │             ├─► ConfigMaps (datasources, dashboards)
   │             └─► RBAC (for Prometheus service discovery)
   │
   └─► module.logging
          │
          └─► kubernetes resources (直接创建)
                 │
                 ├─► Fluentd DaemonSet
                 └─► Fluentd ConfigMap
```

## 状态管理

```
┌────────────────────────────────────────┐
│  Terraform State (terraform.tfstate)   │
│                                        │
│  {                                     │
│    "resources": [                      │
│      {                                 │
│        "module": "elasticsearch",      │
│        "type": "helm_release",         │
│        "instances": [...]              │
│      },                                │
│      {                                 │
│        "module": "control_plane",      │
│        "type": "helm_release",         │
│        "instances": [...]              │
│      }                                 │
│    ]                                   │
│  }                                     │
└────────────────────────────────────────┘
         │
         │ tracks
         ▼
┌────────────────────────────────────────┐
│  Kubernetes 实际状态                    │
│                                        │
│  - Helm releases deployed              │
│  - Namespaces created                  │
│  - Resources running                   │
│  - Services exposed                    │
└────────────────────────────────────────┘
         ▲
         │ reconciles
         │
    terraform apply
```

## 变量流动

```
terraform.tfvars
    │
    ├─ kubeconfig_path ───────────► Providers
    ├─ namespace ─────────────────► Namespace resource
    ├─ elasticsearch_replicas ────► module.elasticsearch
    │                                  │
    │                                  └─► helm values
    │                                        │
    │                                        └─► Chart templates
    │                                              │
    │                                              └─► StatefulSet.spec.replicas
    │
    ├─ elasticsearch_resources ───► module.elasticsearch
    │                                  │
    │                                  └─► helm values
    │                                        │
    │                                        └─► Pod resources
    │
    └─ manager_image ────────────────► module.control_plane
                                          │
                                          └─► helm values
                                                │
                                                └─► Deployment.spec.template.spec.containers[0].image
```

## 监控集成

```
┌──────────────────────────────────────────────────────────┐
│                    Prometheus                            │
│                                                          │
│  Scrape Configs:                                         │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 1. kubernetes-pods (auto-discovery)                │  │
│  │    - Discovers all pods with annotation            │  │
│  │      prometheus.io/scrape: "true"                  │  │
│  └────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 2. elasticsearch (static)                          │  │
│  │    - http://elasticsearch:9200                     │  │
│  └────────────────────────────────────────────────────┘  │
└───────────────────────┬──────────────────────────────────┘
                        │
                        │ metrics
                        ▼
┌──────────────────────────────────────────────────────────┐
│                    Grafana                               │
│                                                          │
│  Datasources:                                            │
│  - Prometheus (http://monitoring-prometheus:9090)        │
│                                                          │
│  Dashboards:                                             │
│  - Elasticsearch Cluster Health                          │
│  - Kubernetes Resource Usage                             │
│  - IVF Plugin Metrics                                    │
└──────────────────────────────────────────────────────────┘
```

## 网络拓扑

```
┌─────────────────────────────────────────────────────────────┐
│  Kubernetes Cluster                                         │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Namespace: es-serverless                             │  │
│  │                                                       │  │
│  │  ┌──────────┐           ┌──────────┐                 │  │
│  │  │ Manager  │◄─────────►│   ES     │                 │  │
│  │  │   API    │   9200    │ Cluster  │                 │  │
│  │  └─────┬────┘           └─────▲────┘                 │  │
│  │        │                      │                      │  │
│  │   8080 │                      │ 9200                 │  │
│  │        │                      │                      │  │
│  │        ▼                      │                      │  │
│  │  ┌──────────┐           ┌─────┴────┐                 │  │
│  │  │  Shard   │           │Prometheus│                 │  │
│  │  │Controller│           │          │                 │  │
│  │  └──────────┘           └─────┬────┘                 │  │
│  │                               │ 9090                 │  │
│  │                               ▼                      │  │
│  │                         ┌──────────┐                 │  │
│  │                         │ Grafana  │                 │  │
│  │                         │          │                 │  │
│  │                         └──────────┘                 │  │
│  │                              3000                    │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Namespace: org-001-alice-vector-search               │  │
│  │  (Network Policy: isolated)                           │  │
│  │                                                       │  │
│  │  ┌──────────┐                                         │  │
│  │  │   ES     │◄─── Only accessible within namespace   │  │
│  │  │ Cluster  │     + from Manager (via service)       │  │
│  │  └──────────┘                                         │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                             │
│  External Access (via port-forward):                        │
│  kubectl port-forward svc/es-control-plane-manager 8080    │
│  kubectl port-forward svc/monitoring-grafana 3000          │
└─────────────────────────────────────────────────────────────┘
```

## 关键设计决策

### 1. 为什么使用 Terraform + Helm?

- **Terraform**:
  - 管理 Kubernetes 基础资源 (Namespaces, RBAC, etc.)
  - 编排多个 Helm releases
  - 统一状态管理

- **Helm**:
  - 应用打包和模板化
  - 版本管理和回滚
  - 复杂应用的参数化配置

### 2. 模块化设计

每个模块负责特定功能:
- **独立性**: 可以单独测试和部署
- **可复用性**: 租户模块可以多次实例化
- **维护性**: 变更隔离,影响范围小

### 3. 租户隔离

- **Namespace 隔离**: 每个租户独立命名空间
- **Resource Quotas**: 限制资源使用
- **Network Policies**: 网络层面隔离
- **Labels**: 便于查询和管理

### 4. 声明式 vs 命令式

```
# 声明式 (Terraform)
resource "helm_release" "elasticsearch" {
  replicas = 3  # 期望状态
}

# Terraform 自动计算并执行从当前状态到期望状态的变更
```

## 下一步

- 查看 [完整部署指南](terraform-helm-guide.md)
- 了解 [Helm Charts 配置](helm-charts-reference.md)
- 探索 [租户管理最佳实践](terraform-helm-guide.md#租户管理)
