# Terraform 和 Helm 部署指南

本文档介绍如何使用 Terraform 和 Helm 部署和管理 ES Serverless 平台。

## 目录

- [架构概述](#架构概述)
- [前置要求](#前置要求)
- [快速开始](#快速开始)
- [部署平台](#部署平台)
- [租户管理](#租户管理)
- [监控和运维](#监控和运维)
- [故障排查](#故障排查)

## 架构概述

新的部署架构采用基础设施即代码 (IaC) 方案:

### Terraform 层
- **主配置** (`terraform/main.tf`): 定义整个平台的基础设施
- **模块化设计**:
  - `modules/elasticsearch`: Elasticsearch 集群模块
  - `modules/control-plane`: 控制平面服务 (Manager, Shard Controller, Reporting)
  - `modules/monitoring`: 监控栈 (Prometheus, Grafana)
  - `modules/logging`: 日志收集 (Fluentd)
  - `modules/tenant`: 租户资源管理

### Helm 层
- **Elasticsearch Chart** (`helm/elasticsearch`): ES 集群及 IVF 插件
- **Control Plane Chart** (`helm/control-plane`): 控制平面服务
- **Monitoring Chart** (`helm/monitoring`): Prometheus 和 Grafana

### 优势
✅ 声明式配置,版本可控
✅ 模块化设计,便于复用
✅ 自动化部署,减少人为错误
✅ 统一的状态管理
✅ 支持多环境部署 (dev/staging/prod)

## 前置要求

### 必需工具

```bash
# Terraform (>= 1.0)
brew install terraform
terraform version

# Helm (>= 3.0)
brew install helm
helm version

# kubectl
brew install kubectl
kubectl version --client

# jq (用于脚本中处理 JSON)
brew install jq
```

### Kubernetes 集群

支持以下 Kubernetes 环境:

**选项 1: Docker Desktop (推荐用于本地开发)**
```bash
# 启用 Kubernetes
# Docker Desktop -> Settings -> Kubernetes -> Enable Kubernetes
```

**选项 2: Kind (Kubernetes in Docker)**
```bash
brew install kind
kind create cluster --name es-serverless
```

**选项 3: 云平台 (生产环境)**
- GKE (Google Kubernetes Engine)
- EKS (Amazon Elastic Kubernetes Service)
- AKS (Azure Kubernetes Service)

### 验证环境

```bash
# 检查 Kubernetes 连接
kubectl cluster-info
kubectl get nodes

# 检查存储类 (StorageClass)
kubectl get sc
```

## 快速开始

### 1. 克隆项目并配置

```bash
cd terraform

# 复制示例配置文件
cp terraform.tfvars.example terraform.tfvars

# 编辑配置文件
vim terraform.tfvars
```

### 2. 配置文件说明

编辑 `terraform.tfvars`:

```hcl
# Kubernetes 配置
kubeconfig_path = "~/.kube/config"
kube_context    = "docker-desktop"  # 或 "kind-kind"

# 命名空间
namespace = "es-serverless"

# Elasticsearch 配置
elasticsearch_replicas     = 3
elasticsearch_storage_size = "10Gi"
storage_class             = "hostpath"

# 资源配置
elasticsearch_resources = {
  requests = {
    cpu    = "1000m"
    memory = "2Gi"
  }
  limits = {
    cpu    = "2000m"
    memory = "4Gi"
  }
}

# 控制平面镜像 (需要先构建)
manager_image           = "es-serverless-manager:latest"
shard_controller_image  = "shard-controller:latest"
reporting_service_image = "reporting-service:latest"

# 监控
prometheus_enabled = true
grafana_enabled   = true
```

### 3. 初始化和部署

```bash
# 使用便捷脚本
./scripts/deploy-terraform.sh init
./scripts/deploy-terraform.sh plan
./scripts/deploy-terraform.sh apply

# 或者直接使用 Terraform
cd terraform
terraform init
terraform plan
terraform apply
```

## 部署平台

### 完整部署流程

#### 步骤 1: 初始化 Terraform

```bash
./scripts/deploy-terraform.sh init
```

这会:
- 下载 Terraform provider (Kubernetes, Helm)
- 初始化后端状态
- 准备模块

#### 步骤 2: 查看执行计划

```bash
./scripts/deploy-terraform.sh plan
```

输出示例:
```
Plan: 15 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + elasticsearch_service_url = "http://elasticsearch.es-serverless.svc.cluster.local:9200"
  + grafana_service_url      = "http://monitoring-grafana.es-serverless.svc.cluster.local:3000"
  + manager_service_url      = "http://es-control-plane-manager.es-serverless.svc.cluster.local:8080"
  + namespace                = "es-serverless"
  + prometheus_service_url   = "http://monitoring-prometheus.es-serverless.svc.cluster.local:9090"
```

#### 步骤 3: 应用配置

```bash
./scripts/deploy-terraform.sh apply
```

部署时间: 约 5-10 分钟

#### 步骤 4: 验证部署

```bash
./scripts/deploy-terraform.sh status
```

或手动检查:

```bash
# 查看命名空间
kubectl get ns es-serverless

# 查看 Helm releases
helm list -n es-serverless

# 查看 Pods
kubectl get pods -n es-serverless

# 应该看到:
# - elasticsearch-0, elasticsearch-1, elasticsearch-2
# - es-control-plane-manager-xxx
# - es-control-plane-shard-controller-xxx
# - es-control-plane-reporting-xxx
# - monitoring-prometheus-xxx
# - monitoring-grafana-xxx
```

### 访问服务

#### Elasticsearch

```bash
kubectl -n es-serverless port-forward svc/elasticsearch 9200:9200

# 测试连接
curl http://localhost:9200
curl http://localhost:9200/_cluster/health
```

#### Manager API

```bash
kubectl -n es-serverless port-forward svc/es-control-plane-manager 8080:8080

# 测试 API
curl http://localhost:8080/clusters
```

#### Grafana Dashboard

```bash
kubectl -n es-serverless port-forward svc/monitoring-grafana 3000:3000

# 访问 http://localhost:3000
# 默认用户名/密码: admin/admin
```

#### Prometheus

```bash
kubectl -n es-serverless port-forward svc/monitoring-prometheus 9090:9090

# 访问 http://localhost:9090
```

## 租户管理

### 创建租户集群

使用便捷脚本创建租户:

```bash
./scripts/create-tenant.sh \
  --org org-001 \
  --user alice \
  --service vector-search \
  --cpu 2000m \
  --memory 4Gi \
  --disk 20Gi \
  --dimension 256 \
  --vectors 10000000 \
  --replicas 3
```

### 参数说明

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `--org` | 组织 ID (必需) | - | org-001 |
| `--user` | 用户名 (必需) | - | alice |
| `--service` | 服务名 (必需) | - | vector-search |
| `--cpu` | CPU 分配 | 1000m | 2000m, 4 |
| `--memory` | 内存分配 | 2Gi | 4Gi, 8Gi |
| `--disk` | 磁盘大小 | 10Gi | 20Gi, 100Gi |
| `--gpu` | GPU 数量 | 0 | 1, 2, 4 |
| `--dimension` | 向量维度 | 128 | 256, 512, 1024 |
| `--vectors` | 向量数量 | 1000000 | 10000000 |
| `--replicas` | 副本数 | 3 | 1, 3, 5 |

### 租户命名规则

租户命名空间遵循规则: `{tenant_org_id}-{user}-{service_name}`

示例:
- 输入: `--org org-001 --user alice --service vector-search`
- 命名空间: `org-001-alice-vector-search`

### 手动创建租户 (使用 Terraform)

如需更精细的控制,可直接使用 Terraform 模块:

```bash
# 创建租户目录
mkdir -p terraform/tenants/org-001-alice-vector-search
cd terraform/tenants/org-001-alice-vector-search

# 创建 main.tf
cat > main.tf <<'EOF'
module "tenant" {
  source = "../../modules/tenant"

  tenant_org_id = "org-001"
  user          = "alice"
  service_name  = "vector-search"

  cpu       = "2000m"
  memory    = "4Gi"
  disk_size = "20Gi"
  gpu_count = 0

  vector_dimension = 256
  vector_count     = 10000000
  replicas         = 3

  # 资源配额
  enable_quota   = true
  quota_cpu      = "10"
  quota_memory   = "20Gi"
  quota_storage  = "100Gi"

  # 网络隔离
  enable_network_policy = true
}

output "namespace" {
  value = module.tenant.namespace
}

output "elasticsearch_url" {
  value = module.tenant.elasticsearch_service_url
}
EOF

# 部署
terraform init
terraform apply
```

### 查看租户

```bash
# 列出所有租户命名空间
kubectl get ns -l es-cluster=true

# 查看特定租户
TENANT_NS="org-001-alice-vector-search"
kubectl get all -n $TENANT_NS

# 查看租户元数据
kubectl get configmap tenant-metadata -n $TENANT_NS -o yaml
```

### 扩容租户集群

```bash
cd terraform/tenants/org-001-alice-vector-search

# 编辑 main.tf, 修改 replicas
vim main.tf

# 应用变更
terraform apply
```

### 删除租户

```bash
cd terraform/tenants/org-001-alice-vector-search

# 销毁资源
terraform destroy

# 或强制删除命名空间
kubectl delete ns org-001-alice-vector-search
```

## 监控和运维

### Prometheus 监控

访问 Prometheus UI:
```bash
kubectl -n es-serverless port-forward svc/monitoring-prometheus 9090:9090
```

常用查询:
```promql
# Elasticsearch JVM 堆内存使用
es_jvm_mem_heap_used_percent

# Pod CPU 使用率
rate(container_cpu_usage_seconds_total[5m])

# Pod 内存使用
container_memory_working_set_bytes
```

### Grafana Dashboard

访问 Grafana:
```bash
kubectl -n es-serverless port-forward svc/monitoring-grafana 3000:3000
```

默认凭证: `admin` / `admin`

添加 Dashboard:
1. 导入官方 Elasticsearch Dashboard (ID: 2322)
2. 导入 Kubernetes Dashboard (ID: 7249)

### 日志查看

查看特定服务日志:

```bash
# Manager 日志
kubectl -n es-serverless logs -l app=es-control-plane-manager -f

# Elasticsearch 日志
kubectl -n es-serverless logs elasticsearch-0 -f

# Shard Controller 日志
kubectl -n es-serverless logs -l component=shard-controller -f
```

查看租户日志:
```bash
TENANT_NS="org-001-alice-vector-search"
kubectl -n $TENANT_NS logs -l app=elasticsearch -f
```

### 资源使用统计

```bash
# 查看节点资源使用
kubectl top nodes

# 查看 Pod 资源使用
kubectl top pods -n es-serverless

# 查看租户资源使用
kubectl top pods -n org-001-alice-vector-search
```

## 故障排查

### 常见问题

#### 1. Helm Release 安装失败

**症状**:
```
Error: timed out waiting for the condition
```

**解决方案**:
```bash
# 查看 Pod 状态
kubectl get pods -n es-serverless

# 查看 Pod 详情
kubectl describe pod <pod-name> -n es-serverless

# 查看 Pod 日志
kubectl logs <pod-name> -n es-serverless

# 删除失败的 Helm release
helm uninstall <release-name> -n es-serverless

# 重新应用 Terraform
cd terraform
terraform apply
```

#### 2. Elasticsearch 无法启动

**症状**:
```
CrashLoopBackOff
```

**检查步骤**:
```bash
# 查看 Pod 事件
kubectl describe pod elasticsearch-0 -n es-serverless

# 查看日志
kubectl logs elasticsearch-0 -n es-serverless

# 常见原因:
# - 内存不足: 增加 memory limits
# - 磁盘空间不足: 检查 PVC
# - JVM 配置错误: 检查 ES_JAVA_OPTS
```

**解决方案**:
```bash
# 增加资源限制
# 编辑 terraform/terraform.tfvars
elasticsearch_resources = {
  limits = {
    memory = "4Gi"  # 增加内存
  }
}

terraform apply
```

#### 3. PVC 无法绑定

**症状**:
```
PersistentVolumeClaim is not bound
```

**解决方案**:
```bash
# 检查 StorageClass
kubectl get sc

# 如果没有默认 StorageClass
kubectl patch storageclass hostpath -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

# 或修改 terraform.tfvars
storage_class = "standard"  # 使用可用的 StorageClass
```

#### 4. 网络连接问题

**症状**: 服务间无法通信

**检查步骤**:
```bash
# 检查服务
kubectl get svc -n es-serverless

# 测试 DNS 解析
kubectl run -it --rm debug --image=busybox --restart=Never -n es-serverless -- nslookup elasticsearch

# 测试连接
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -n es-serverless -- curl http://elasticsearch:9200
```

#### 5. Terraform State 锁定

**症状**:
```
Error: Error acquiring the state lock
```

**解决方案**:
```bash
# 强制解锁 (谨慎使用)
cd terraform
terraform force-unlock <LOCK_ID>
```

### 调试工具

#### 进入 Pod 调试

```bash
# 进入 Elasticsearch Pod
kubectl exec -it elasticsearch-0 -n es-serverless -- bash

# 在 Pod 内测试
curl localhost:9200
curl localhost:9200/_cat/nodes?v
```

#### 查看 Terraform 状态

```bash
cd terraform

# 查看当前状态
terraform show

# 查看特定资源
terraform state list
terraform state show module.elasticsearch.helm_release.elasticsearch
```

#### 清理和重建

```bash
# 完全销毁并重建
./scripts/deploy-terraform.sh destroy
./scripts/deploy-terraform.sh apply

# 仅重建特定模块
cd terraform
terraform destroy -target=module.monitoring
terraform apply -target=module.monitoring
```

### 性能优化

#### Elasticsearch 调优

编辑 `helm/elasticsearch/values.yaml`:

```yaml
# JVM 堆内存 (设为容器内存的 50%)
env:
  - name: ES_JAVA_OPTS
    value: "-Xms2g -Xmx2g"

# 线程池
esConfig:
  elasticsearch.yml: |
    thread_pool:
      write:
        queue_size: 1000
      search:
        queue_size: 1000
```

#### 资源限制建议

| 环境 | CPU | 内存 | 磁盘 |
|------|-----|------|------|
| 开发 | 1-2 | 2-4Gi | 10-20Gi |
| 测试 | 2-4 | 4-8Gi | 50-100Gi |
| 生产 | 4-8 | 8-16Gi | 100-500Gi |

### 备份和恢复

#### Elasticsearch 快照

```bash
# 创建快照仓库 (通过 Manager API)
curl -X POST http://localhost:8080/snapshots/repo \
  -H 'Content-Type: application/json' \
  -d '{
    "type": "fs",
    "settings": {
      "location": "/backups"
    }
  }'

# 创建快照
curl -X POST http://localhost:8080/snapshots/snapshot_1

# 恢复快照
curl -X POST http://localhost:8080/snapshots/snapshot_1/_restore
```

#### Terraform State 备份

```bash
cd terraform

# 备份状态文件
cp terraform.tfstate terraform.tfstate.backup.$(date +%Y%m%d)

# 或使用远程后端 (S3/GCS)
# 编辑 main.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "es-serverless/terraform.tfstate"
    region = "us-west-2"
  }
}
```

## 最佳实践

### 1. 使用版本控制

```bash
# 将 Terraform 配置纳入 Git
git add terraform/ helm/
git commit -m "Add Terraform/Helm configurations"
```

### 2. 环境隔离

为不同环境创建独立配置:

```
terraform/
  environments/
    dev/
      terraform.tfvars
      main.tf
    staging/
      terraform.tfvars
      main.tf
    prod/
      terraform.tfvars
      main.tf
```

### 3. 使用变量和密钥管理

```bash
# 使用环境变量
export TF_VAR_grafana_admin_password="secure-password"

# 或使用 Kubernetes Secrets
kubectl create secret generic grafana-admin \
  --from-literal=password=secure-password \
  -n es-serverless
```

### 4. 持续集成/持续部署 (CI/CD)

在 CI/CD 流水线中使用 Terraform:

```yaml
# .github/workflows/deploy.yml
name: Deploy ES Serverless
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: hashicorp/setup-terraform@v1
      - name: Terraform Init
        run: cd terraform && terraform init
      - name: Terraform Apply
        run: cd terraform && terraform apply -auto-approve
```

### 5. 成本优化

```bash
# 使用节点亲和性将 Pods 调度到低成本节点
# 编辑 helm/elasticsearch/values.yaml
nodeSelector:
  node.kubernetes.io/instance-type: n1-standard-2
```

## 下一步

- 阅读 [多租户架构说明](多租户架构说明.md)
- 了解 [自动扩展配额管理](自动扩展配额管理说明.md)
- 查看 [分片数据同步实现方案](分片数据同步实现方案.md)
- 探索 [API 文档](../README.md#rest-api-endpoints)

## 支持

如有问题,请:
1. 查看本文档的故障排查部分
2. 检查 GitHub Issues
3. 联系平台管理员
