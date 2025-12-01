# 快速开始指南

本指南帮助你从零开始设置和运行 ES Serverless 平台。

## 前置要求

### 必需软件

```bash
# 1. Docker Desktop (包含 Kubernetes)
# 下载: https://www.docker.com/products/docker-desktop

# 2. Terraform
brew install terraform
terraform version  # >= 1.0

# 3. Helm
brew install helm
helm version  # >= 3.0

# 4. kubectl
brew install kubectl
kubectl version --client

# 5. Git
brew install git

# 6. Go (如果需要修改代码)
brew install go
go version  # >= 1.21

# 7. jq (用于脚本)
brew install jq
```

### 启用 Kubernetes

```bash
# Docker Desktop -> Settings -> Kubernetes -> Enable Kubernetes
# 等待 Kubernetes 启动完成

# 验证
kubectl cluster-info
kubectl get nodes
```

## 克隆代码

### 方式 1: 如果代码已在 Git 仓库

```bash
# 克隆仓库
git clone <your-repo-url>
cd es项目

# 查看文件结构
tree -L 2
```

### 方式 2: 如果是本地目录

代码已在: `/Users/yunpeng/Documents/es项目`

```bash
# 进入项目目录
cd /Users/yunpeng/Documents/es项目

# 查看项目结构
ls -la

# 应该看到:
# - terraform/
# - helm/
# - scripts/
# - server/
# - docs/
# - Makefile
```

## 初始化项目

### 步骤 1: 检查环境

```bash
# 运行环境检查
make test-cluster

# 输出应该显示:
# ✓ Kubernetes 连接正常
# ✓ Helm 版本正确
```

### 步骤 2: 配置 Terraform

```bash
# 进入 terraform 目录
cd terraform

# 复制示例配置
cp terraform.tfvars.example terraform.tfvars

# 编辑配置文件
vim terraform.tfvars
```

**重要配置项**:

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

# 资源配置 (根据你的机器调整)
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

# 控制平面镜像 (需要先构建,见下文)
manager_image           = "es-serverless-manager:latest"
shard_controller_image  = "shard-controller:latest"
reporting_service_image = "reporting-service:latest"

# 监控
prometheus_enabled = true
grafana_enabled   = true

# 日志
fluentd_enabled = true
```

### 步骤 3: 构建服务镜像 (可选)

如果需要使用控制平面服务,先构建镜像:

```bash
# 构建 Manager
cd server
docker build -t es-serverless-manager:latest .

# 构建 Shard Controller
# (如果有单独的 Dockerfile)
docker build -f Dockerfile.shard-controller -t shard-controller:latest .

# 构建 Reporting Service
docker build -f Dockerfile.reporting -t reporting-service:latest .

# 回到项目根目录
cd ..
```

**如果没有 Dockerfile,可以暂时禁用控制平面**:

```hcl
# terraform.tfvars
# 注释掉控制平面相关配置,或在 main.tf 中禁用该模块
```

## 部署平台

### 方式 1: 使用 Makefile (推荐)

```bash
# 一键部署
make quick-start

# 包含:
# 1. terraform init
# 2. terraform apply
# 3. 显示服务访问地址
```

### 方式 2: 手动步骤

```bash
# 1. 初始化 Terraform
cd terraform
terraform init

# 2. 查看执行计划
terraform plan

# 3. 应用配置
terraform apply

# 输入 'yes' 确认
```

### 步骤 4: 验证部署

```bash
# 查看状态
make status

# 或手动检查
kubectl get pods -n es-serverless

# 应该看到:
# - elasticsearch-0, elasticsearch-1, elasticsearch-2 (Running)
# - monitoring-prometheus-xxx (Running)
# - monitoring-grafana-xxx (Running)
# - 其他组件...
```

## 访问服务

### Elasticsearch

```bash
# 端口转发
make port-forward-es

# 或手动
kubectl -n es-serverless port-forward svc/elasticsearch 9200:9200

# 在新终端测试
curl http://localhost:9200
curl http://localhost:9200/_cluster/health
```

### Manager API (如果部署了)

```bash
# 端口转发
make port-forward-manager

# 测试
curl http://localhost:8080/clusters
```

### Grafana

```bash
# 端口转发
make port-forward-grafana

# 浏览器访问: http://localhost:3000
# 用户名/密码: admin/admin
```

### Prometheus

```bash
# 端口转发
make port-forward-prometheus

# 浏览器访问: http://localhost:9090
```

## 创建测试租户

```bash
# 使用 Makefile
make tenant-create \
  ORG=demo \
  USER=test \
  SERVICE=app1 \
  REPLICAS=3

# 或使用脚本
./scripts/create-tenant.sh \
  --org demo \
  --user test \
  --service app1 \
  --replicas 3

# 查看租户
make tenant-list

# 查看租户状态
make tenant-status TENANT=demo-test-app1
```

## 常用命令

### Makefile 命令

```bash
# 查看所有命令
make help

# 平台管理
make init              # 初始化
make plan              # 查看计划
make apply             # 部署
make status            # 查看状态
make destroy           # 销毁

# 租户管理
make tenant-create     # 创建租户
make tenant-list       # 列出租户
make tenant-status     # 租户状态
make tenant-delete     # 删除租户

# 日志查看
make logs-manager      # Manager 日志
make logs-elasticsearch
make logs-tenant TENANT=demo-test-app1

# 快速开始
make quick-start       # 初始化+部署
make quick-demo        # 部署+创建示例租户
```

### Terraform 命令

```bash
cd terraform

# 初始化
terraform init

# 查看计划
terraform plan

# 应用
terraform apply

# 查看输出
terraform output

# 查看状态
terraform show

# 销毁
terraform destroy
```

### Helm 命令

```bash
# 列出 releases
helm list -n es-serverless

# 查看状态
helm status elasticsearch -n es-serverless

# 查看 values
helm get values elasticsearch -n es-serverless

# 升级
helm upgrade elasticsearch ./helm/elasticsearch -n es-serverless

# 回滚
helm rollback elasticsearch -n es-serverless
```

### kubectl 命令

```bash
# 查看 Pods
kubectl get pods -n es-serverless

# 查看服务
kubectl get svc -n es-serverless

# 查看 PVC
kubectl get pvc -n es-serverless

# 查看日志
kubectl logs -f elasticsearch-0 -n es-serverless

# 进入容器
kubectl exec -it elasticsearch-0 -n es-serverless -- bash

# 查看租户
kubectl get ns -l es-cluster=true
```

## 项目目录结构

```
es项目/
├── terraform/                 # Terraform 配置
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── terraform.tfvars      # (需创建)
│   └── modules/              # Terraform 模块
│
├── helm/                     # Helm Charts
│   ├── elasticsearch/
│   ├── control-plane/
│   └── monitoring/
│
├── scripts/                  # 部署脚本
│   ├── deploy-terraform.sh
│   └── create-tenant.sh
│
├── server/                   # Go 服务代码
│   ├── main.go
│   ├── metadata.go
│   └── ...
│
├── examples/                 # 示例代码
│   └── manager-with-helm/   # Helm SDK 集成示例
│
├── docs/                     # 文档
│   ├── terraform-helm-guide.md
│   ├── helm-charts-reference.md
│   └── terraform-vs-helm-sdk.md
│
├── Makefile                  # Make 命令
├── QUICK_REFERENCE.md        # 快速参考
└── README.md                 # 项目说明
```

## 故障排查

### 问题 1: Terraform init 失败

```bash
# 清理并重新初始化
rm -rf terraform/.terraform
rm -f terraform/.terraform.lock.hcl
cd terraform && terraform init
```

### 问题 2: Pod 无法启动

```bash
# 查看 Pod 详情
kubectl describe pod elasticsearch-0 -n es-serverless

# 查看日志
kubectl logs elasticsearch-0 -n es-serverless

# 常见原因:
# - 资源不足: 降低 limits
# - 存储问题: 检查 StorageClass
```

### 问题 3: PVC 无法绑定

```bash
# 检查 StorageClass
kubectl get sc

# 设置默认 StorageClass
kubectl patch storageclass hostpath \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

### 问题 4: 镜像拉取失败

```bash
# 检查镜像是否存在
docker images | grep es-serverless

# 如果没有,需要先构建或禁用相关组件
```

## 清理环境

### 完全清理

```bash
# 销毁所有资源
make destroy

# 或
cd terraform
terraform destroy

# 删除命名空间
kubectl delete ns es-serverless

# 删除租户
kubectl delete ns -l es-cluster=true
```

### 部分清理

```bash
# 只删除特定租户
make tenant-delete TENANT=demo-test-app1

# 只删除特定模块
cd terraform
terraform destroy -target=module.monitoring
```

## 下一步

### 1. 学习文档

- [Terraform/Helm 完整指南](docs/terraform-helm-guide.md) - 详细使用说明
- [快速参考](QUICK_REFERENCE.md) - 常用命令速查
- [Helm Charts 参考](docs/helm-charts-reference.md) - 配置参数

### 2. 尝试功能

```bash
# 创建多个租户
for i in {1..3}; do
  make tenant-create ORG=demo USER=user$i SERVICE=app1
done

# 查看所有租户
make tenant-list

# 扩容租户
make tenant-status TENANT=demo-user1-app1
cd terraform/tenants/demo-user1-app1
# 修改 replicas
terraform apply
```

### 3. 集成 Helm SDK

查看示例:
```bash
cd examples/manager-with-helm
cat README.md
go run *.go
```

### 4. 自定义配置

修改 Helm values:
```bash
cd helm/elasticsearch
vim values.yaml
# 修改配置

cd ../../terraform
terraform apply
```

## 获取帮助

### 文档

- `make help` - 显示所有命令
- `make docs` - 查看文档列表
- [完整文档](docs/)

### 日志

```bash
# 查看所有组件日志
make logs-manager
make logs-elasticsearch
make logs-tenant TENANT=xxx
```

### 调试

```bash
# 详细输出
export TF_LOG=DEBUG
terraform apply

# Helm 调试
helm install --debug --dry-run elasticsearch ./helm/elasticsearch
```

## 开发环境推荐配置

### 最小配置 (本地测试)

```hcl
# terraform.tfvars
elasticsearch_replicas = 1
prometheus_enabled = false
grafana_enabled = false
fluentd_enabled = false

elasticsearch_resources = {
  requests = { cpu = "500m", memory = "1Gi" }
  limits = { cpu = "1000m", memory = "2Gi" }
}
```

### 标准配置 (开发环境)

```hcl
# terraform.tfvars
elasticsearch_replicas = 3
prometheus_enabled = true
grafana_enabled = true
fluentd_enabled = false

elasticsearch_resources = {
  requests = { cpu = "1000m", memory = "2Gi" }
  limits = { cpu = "2000m", memory = "4Gi" }
}
```

### 完整配置 (生产模拟)

```hcl
# terraform.tfvars
elasticsearch_replicas = 3
prometheus_enabled = true
grafana_enabled = true
fluentd_enabled = true

elasticsearch_resources = {
  requests = { cpu = "2000m", memory = "4Gi" }
  limits = { cpu = "4000m", memory = "8Gi" }
}
```

## 成功指标

部署成功后,你应该能够:

- ✅ `kubectl get pods -n es-serverless` 显示所有 Pod 为 Running
- ✅ `curl http://localhost:9200` 返回 Elasticsearch 信息
- ✅ 访问 Grafana (http://localhost:3000) 看到 Dashboard
- ✅ `make tenant-create` 成功创建租户
- ✅ `make tenant-list` 看到租户列表

恭喜! 🎉 你已成功部署 ES Serverless 平台!
