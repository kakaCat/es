# Terraform + Helm 部署方式

这个目录包含使用 **Terraform + Helm** 进行基础设施即代码（IaC）的部署方式。

## 📋 概述

这种部署方式使用：
- **Terraform**: 管理 Kubernetes 资源（命名空间、配额、网络策略等）
- **Helm**: 管理应用部署（Elasticsearch、Control Plane、Monitoring）

## 🏗️ 架构优势

✅ **基础设施即代码**: 所有配置可版本控制、可复现
✅ **模块化设计**: Terraform 模块化管理多租户隔离
✅ **声明式管理**: Helm Charts 管理应用配置
✅ **状态管理**: Terraform 状态文件追踪资源变更
✅ **企业级部署**: 适合生产环境大规模部署

## 📁 目录结构

```
deployment-terraform/
├── README.md                 # 本文档
├── terraform/                # Terraform 配置
│   ├── main.tf              # 主配置文件
│   ├── variables.tf         # 变量定义
│   ├── outputs.tf           # 输出定义
│   ├── terraform.tfvars     # 变量值（自定义）
│   └── modules/             # Terraform 模块
│       ├── tenant/          # 租户模块（命名空间、配额、隔离）
│       ├── control-plane/   # 控制平面模块
│       ├── monitoring/      # 监控模块
│       ├── logging/         # 日志模块
│       └── gpu/             # GPU 调度模块
│
├── helm/                    # Helm Charts
│   ├── elasticsearch/       # ES + IVF 插件
│   ├── control-plane/       # Manager + ShardController + Reporting
│   └── monitoring/          # Prometheus + Grafana
│
└── scripts/                 # 辅助脚本
    ├── deploy.sh            # 一键部署脚本
    └── destroy.sh           # 一键清理脚本
```

## 🚀 快速开始

### 前提条件

```bash
# 1. 安装工具
brew install terraform kubectl helm

# 2. 配置 Kubernetes 上下文
kubectl config use-context docker-desktop  # 或 kind-kind

# 3. 验证连接
kubectl cluster-info
```

### 部署步骤

#### 1️⃣ 初始化 Terraform

```bash
cd deployment-terraform/terraform
terraform init
```

#### 2️⃣ 配置变量

复制示例配置并修改：

```bash
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars
```

关键配置项：
```hcl
# Kubernetes 配置
kubeconfig_path = "~/.kube/config"
kube_context    = "docker-desktop"

# Elasticsearch 配置
elasticsearch_replicas = 3
elasticsearch_storage_size = "10Gi"

# GPU 配置（可选）
gpu_enabled = true
gpu_count = 1
```

#### 3️⃣ 预览变更

```bash
terraform plan
```

#### 4️⃣ 执行部署

```bash
terraform apply
```

或使用快捷脚本：

```bash
cd deployment-terraform
./scripts/deploy.sh
```

#### 5️⃣ 验证部署

```bash
# 查看所有资源
kubectl get all -n es-serverless

# 查看租户命名空间
kubectl get ns | grep es-

# 访问服务
kubectl -n es-serverless port-forward svc/elasticsearch 9200:9200
curl http://localhost:9200
```

## 🎯 核心功能

### 1. 多租户隔离

Terraform 自动创建隔离的租户环境：

```hcl
module "tenant_alice" {
  source = "./modules/tenant"

  tenant_org_id = "org-001"
  user          = "alice"
  service_name  = "vector-search"

  # 资源配额
  replicas = 3
  cpu      = "2000m"
  memory   = "4Gi"
  storage  = "50Gi"

  # GPU 支持
  gpu_enabled = true
  gpu_count   = 1
}
```

生成的资源：
- Namespace: `org-001-alice-vector-search`
- ResourceQuota: CPU/内存/存储限制
- NetworkPolicy: 租户间网络隔离
- GPU NodeSelector: 自动调度到 GPU 节点

### 2. Helm Charts 管理

#### Elasticsearch + IVF 插件

```bash
helm install elasticsearch ./helm/elasticsearch \
  --namespace es-serverless \
  --set replicaCount=3 \
  --set ivfPlugin.enabled=true \
  --set persistence.size=10Gi
```

#### Control Plane

```bash
helm install control-plane ./helm/control-plane \
  --namespace es-serverless \
  --set manager.replicas=2 \
  --set shardController.enabled=true
```

### 3. GPU 加速配置

Terraform 自动配置 GPU 资源：

```hcl
# 在 terraform.tfvars 中启用
gpu_enabled = true
gpu_count   = 1

# Terraform 会自动：
# 1. 添加 GPU nodeSelector
# 2. 配置 nvidia.com/gpu 资源限制
# 3. 设置 GPU 容忍度（Tolerations）
```

生成的 Pod 配置：
```yaml
spec:
  nodeSelector:
    nvidia.com/gpu: "true"
  resources:
    limits:
      nvidia.com/gpu: 1
```

### 4. 监控与日志

```bash
# Prometheus + Grafana
helm install monitoring ./helm/monitoring \
  --namespace es-serverless

# 访问 Grafana
kubectl -n es-serverless port-forward svc/grafana 3000:3000
# 默认账号: admin / admin
```

## 🔧 常用操作

### 扩容 Elasticsearch

```bash
# 方式1: 修改 terraform.tfvars
elasticsearch_replicas = 5

terraform apply

# 方式2: Helm 升级
helm upgrade elasticsearch ./helm/elasticsearch \
  --set replicaCount=5
```

### 创建新租户

```bash
# 1. 在 terraform/main.tf 添加模块
module "tenant_bob" {
  source = "./modules/tenant"

  tenant_org_id = "org-002"
  user          = "bob"
  service_name  = "text-search"
  replicas      = 2
}

# 2. 应用变更
terraform apply
```

### 更新配置

```bash
# 修改 Helm values
vim helm/elasticsearch/values.yaml

# 升级部署
helm upgrade elasticsearch ./helm/elasticsearch
```

### 查看 Terraform 状态

```bash
# 查看所有资源
terraform state list

# 查看特定资源详情
terraform state show module.tenant_alice.kubernetes_namespace.tenant

# 查看输出
terraform output
```

## 🔄 升级与回滚

### Terraform 升级

```bash
# 查看变更
terraform plan

# 应用升级
terraform apply
```

### Helm 升级

```bash
# 查看历史版本
helm history elasticsearch -n es-serverless

# 升级到新版本
helm upgrade elasticsearch ./helm/elasticsearch

# 回滚到上一版本
helm rollback elasticsearch -n es-serverless
```

## 🗑️ 清理资源

### 删除特定租户

```bash
# 方式1: 注释掉模块后 apply
# terraform/main.tf 中注释掉 module "tenant_bob"
terraform apply

# 方式2: 直接删除
terraform destroy -target=module.tenant_bob
```

### 完全清理

```bash
# 使用 Terraform
cd deployment-terraform/terraform
terraform destroy

# 或使用脚本
cd deployment-terraform
./scripts/destroy.sh
```

## 📊 监控与调试

### 查看日志

```bash
# Elasticsearch 日志
kubectl -n es-serverless logs -l app=elasticsearch --tail=100 -f

# Manager 日志
kubectl -n es-serverless logs -l app=es-serverless-manager -f

# Shard Controller 日志
kubectl -n es-serverless logs -l app=shard-controller -f
```

### 健康检查

```bash
# 集群健康
kubectl -n es-serverless exec -it elasticsearch-0 -- \
  curl -s http://localhost:9200/_cluster/health?pretty

# IVF 插件状态
kubectl -n es-serverless exec -it elasticsearch-0 -- \
  curl -s http://localhost:9200/_cat/plugins
```

## 🔐 安全配置

### 网络隔离

Terraform 自动配置 NetworkPolicy：

```yaml
# 允许租户内部通信
# 拒绝跨租户访问
# 允许 DNS 查询
```

### 资源配额

```hcl
# 在模块中配置
quota_cpu     = "10"      # 10 核
quota_memory  = "20Gi"    # 20GB 内存
quota_storage = "100Gi"   # 100GB 存储
quota_pods    = "50"      # 最多 50 个 Pod
```

## 📚 参考文档

- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)
- [Helm 官方文档](https://helm.sh/docs/)
- [Elasticsearch Helm Chart](https://github.com/elastic/helm-charts)
- [GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/getting-started.html)

## ❓ 常见问题

### Q: Terraform 和 Helm 的区别？

**Terraform**:
- 管理基础设施资源（命名空间、配额、网络）
- 适合静态资源和租户隔离
- 状态管理和依赖追踪

**Helm**:
- 管理应用部署（Elasticsearch、控制平面）
- 适合应用配置和版本管理
- 支持模板化和升级回滚

### Q: 为什么选择 Terraform + Helm？

- **关注点分离**: Terraform 管基础设施，Helm 管应用
- **模块化**: Terraform 模块复用多租户配置
- **GitOps**: 配置文件版本控制
- **企业级**: 适合大规模生产环境

### Q: GPU 支持需要什么前提？

```bash
# 1. 安装 NVIDIA GPU Operator
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/gpu-operator/master/deployments/gpu-operator.yaml

# 2. 验证 GPU 节点
kubectl get nodes -o json | jq '.items[].status.allocatable'

# 3. 在 Terraform 中启用
gpu_enabled = true
```

### Q: 如何备份 Terraform 状态？

```bash
# 本地备份
cp terraform.tfstate terraform.tfstate.backup

# 使用远程后端（推荐）
# terraform/main.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "es-serverless/terraform.tfstate"
    region = "us-west-2"
  }
}
```

## 🆚 对比：Terraform+Helm vs 纯脚本

| 特性 | Terraform+Helm | 纯脚本 |
|------|---------------|--------|
| 学习曲线 | 中等（需学习 Terraform/Helm） | 低（Shell 脚本） |
| 适用场景 | 生产环境、多租户、大规模 | 开发测试、快速验证 |
| 状态管理 | ✅ 自动状态追踪 | ❌ 手动管理 |
| 回滚能力 | ✅ 支持 | ❌ 需手动处理 |
| 模块化 | ✅ 高度模块化 | ⚠️ 脚本复用有限 |
| 多云支持 | ✅ 支持多云 | ❌ 需为每个平台写脚本 |
| 配置复杂度 | 中等 | 低 |
| 维护成本 | 低（声明式） | 高（命令式） |

**推荐使用场景**:
- **Terraform+Helm**: 生产环境、多租户部署、需要审计和回滚
- **纯脚本**: 本地开发、快速测试、概念验证（POC）

---

**下一步**: 参考 [../deployment-scripts/README.md](../deployment-scripts/README.md) 了解纯脚本部署方式
