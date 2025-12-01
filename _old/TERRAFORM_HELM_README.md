# ES Serverless - Terraform 和 Helm 部署

本项目现已支持使用 **Terraform** 和 **Helm** 进行基础设施即代码 (Infrastructure as Code) 部署。

## 快速开始

### 前置要求

- Terraform >= 1.0
- Helm >= 3.0
- kubectl
- Kubernetes 集群 (Docker Desktop / Kind / GKE / EKS / AKS)

### 5 分钟部署

```bash
# 1. 配置
cd terraform
cp terraform.tfvars.example terraform.tfvars
# 编辑 terraform.tfvars 设置你的配置

# 2. 部署平台
./scripts/deploy-terraform.sh init
./scripts/deploy-terraform.sh apply

# 3. 创建租户集群
./scripts/create-tenant.sh \
  --org org-001 \
  --user alice \
  --service vector-search \
  --cpu 2000m \
  --memory 4Gi \
  --disk 20Gi

# 4. 访问服务
kubectl -n es-serverless port-forward svc/es-control-plane-manager 8080:8080
```

## 项目结构

```
.
├── terraform/                      # Terraform 配置
│   ├── main.tf                    # 主配置文件
│   ├── variables.tf               # 变量定义
│   ├── outputs.tf                 # 输出定义
│   ├── terraform.tfvars.example   # 配置示例
│   ├── modules/                   # Terraform 模块
│   │   ├── elasticsearch/         # Elasticsearch 模块
│   │   ├── control-plane/         # 控制平面模块
│   │   ├── monitoring/            # 监控模块
│   │   ├── logging/               # 日志模块
│   │   └── tenant/                # 租户资源模块
│   └── tenants/                   # 租户实例 (自动生成)
│
├── helm/                          # Helm Charts
│   ├── elasticsearch/             # Elasticsearch Chart
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   ├── control-plane/             # 控制平面 Chart
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   └── monitoring/                # 监控 Chart
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│
├── scripts/
│   ├── deploy-terraform.sh        # Terraform 部署脚本
│   └── create-tenant.sh           # 租户创建脚本
│
└── docs/
    ├── terraform-helm-guide.md    # 完整使用指南
    └── helm-charts-reference.md   # Helm Charts 参考
```

## 核心功能

### ✅ 平台部署

使用 Terraform 一键部署整个平台:

- Elasticsearch 集群 (带 IVF 向量搜索插件)
- 控制平面服务 (Manager, Shard Controller, Reporting)
- 监控栈 (Prometheus, Grafana)
- 日志收集 (Fluentd)

### ✅ 多租户管理

轻松创建和管理租户集群:

```bash
./scripts/create-tenant.sh \
  --org myorg \
  --user john \
  --service app1 \
  --replicas 3
```

每个租户获得:
- 独立的 Kubernetes 命名空间
- 专用的 Elasticsearch 集群
- 资源配额和网络隔离
- 自定义向量搜索配置

### ✅ 声明式配置

所有基础设施配置都是声明式的:

```hcl
# terraform.tfvars
elasticsearch_replicas = 3
elasticsearch_storage_size = "20Gi"

elasticsearch_resources = {
  requests = {
    cpu    = "2000m"
    memory = "4Gi"
  }
}
```

### ✅ 版本控制和回滚

```bash
# 查看变更计划
terraform plan

# 应用变更
terraform apply

# 回滚到之前的状态
terraform apply -var="elasticsearch_replicas=3"
```

## 架构优势

### 传统方式 vs Terraform/Helm

| 特性 | 传统 YAML | Terraform + Helm |
|------|-----------|------------------|
| 配置管理 | 分散的 YAML 文件 | 统一的变量管理 |
| 状态追踪 | 手动记录 | 自动状态管理 |
| 变更预览 | 无 | `terraform plan` |
| 模块化 | 复制粘贴 | 可复用模块 |
| 多环境 | 复杂 | 轻松切换 |
| 依赖管理 | 手动控制 | 自动处理 |
| 回滚 | 手动 | 一条命令 |

### 基础设施即代码的好处

1. **可重复性**: 环境一致,避免配置漂移
2. **版本控制**: 所有变更可追溯
3. **协作**: 团队可以 review 基础设施变更
4. **自动化**: CI/CD 集成
5. **文档化**: 配置即文档

## 使用场景

### 场景 1: 部署新的开发环境

```bash
# 1. 创建环境配置
mkdir -p terraform/environments/dev
cd terraform/environments/dev

# 2. 创建配置
cat > terraform.tfvars <<EOF
namespace = "es-dev"
elasticsearch_replicas = 1
elasticsearch_storage_size = "5Gi"
EOF

# 3. 部署
terraform init -from-module=../..
terraform apply
```

### 场景 2: 扩容生产环境

```bash
# 编辑 terraform.tfvars
elasticsearch_replicas = 5  # 从 3 增加到 5

# 预览变更
terraform plan

# 应用变更
terraform apply
```

### 场景 3: 创建多个租户

```bash
# 批量创建租户
for user in alice bob charlie; do
  ./scripts/create-tenant.sh \
    --org company-001 \
    --user $user \
    --service analytics \
    --replicas 3
done
```

### 场景 4: 灾难恢复

```bash
# 销毁并重建整个环境
terraform destroy
terraform apply

# 或仅重建特定组件
terraform destroy -target=module.elasticsearch
terraform apply -target=module.elasticsearch
```

## Helm Charts

项目提供 3 个主要 Helm Charts:

### 1. Elasticsearch Chart

功能:
- Elasticsearch 8.x 集群部署
- IVF 向量搜索插件集成
- 自动化集群发现和配置
- 持久化存储管理

```bash
helm install elasticsearch ./helm/elasticsearch \
  --set replicaCount=5 \
  --set persistence.size=50Gi
```

### 2. Control Plane Chart

包含:
- Manager API (集群管理)
- Shard Controller (分片管理)
- Reporting Service (状态上报)

```bash
helm install control-plane ./helm/control-plane \
  --set manager.image.tag=v1.0.0
```

### 3. Monitoring Chart

包含:
- Prometheus (指标收集)
- Grafana (可视化)

```bash
helm install monitoring ./helm/monitoring \
  --set prometheus.retention.days=30
```

## Terraform 模块

### 核心模块

1. **elasticsearch**: ES 集群部署
2. **control-plane**: 控制平面服务
3. **monitoring**: 监控栈
4. **logging**: 日志收集
5. **tenant**: 租户资源管理 (多租户关键)

### 租户模块特性

租户模块 (`modules/tenant`) 提供:

```hcl
module "tenant" {
  source = "./modules/tenant"

  # 租户标识
  tenant_org_id = "org-001"
  user          = "alice"
  service_name  = "vector-search"

  # 资源配置
  cpu       = "2000m"
  memory    = "4Gi"
  disk_size = "20Gi"
  gpu_count = 1

  # 向量配置
  vector_dimension = 256
  vector_count     = 10000000
  replicas         = 3

  # 隔离和配额
  enable_quota          = true
  enable_network_policy = true
}
```

自动创建:
- ✅ 专用命名空间 (`org-001-alice-vector-search`)
- ✅ Elasticsearch 集群
- ✅ 资源配额
- ✅ 网络策略 (租户隔离)
- ✅ 元数据 ConfigMap

## 监控和运维

### 访问 Grafana

```bash
kubectl -n es-serverless port-forward svc/monitoring-grafana 3000:3000
# 访问 http://localhost:3000
# 用户名/密码: admin/admin
```

### 访问 Prometheus

```bash
kubectl -n es-serverless port-forward svc/monitoring-prometheus 9090:9090
# 访问 http://localhost:9090
```

### 查看日志

```bash
# Manager 日志
kubectl -n es-serverless logs -l app=es-control-plane-manager -f

# 租户日志
kubectl -n org-001-alice-vector-search logs -l app=elasticsearch -f
```

### 资源监控

```bash
# 查看资源使用
kubectl top pods -n es-serverless

# 查看租户资源使用
kubectl top pods -n org-001-alice-vector-search
```

## 升级和维护

### 升级 Elasticsearch 版本

```bash
# 编辑 terraform.tfvars 或 helm/elasticsearch/values.yaml
# image.tag: "8.12.0"

terraform apply
# 或
helm upgrade elasticsearch ./helm/elasticsearch
```

### 更新配置

```bash
# 修改任何 .tf 或 values.yaml 文件后
terraform plan  # 预览变更
terraform apply # 应用变更
```

### 备份和恢复

```bash
# 备份 Terraform 状态
cp terraform.tfstate terraform.tfstate.backup

# 备份 Helm releases
helm list -n es-serverless > helm-releases.txt
```

## 最佳实践

1. **使用版本控制**: 将 Terraform 配置提交到 Git
2. **环境隔离**: 为 dev/staging/prod 创建独立配置
3. **使用变量**: 避免硬编码值
4. **定期备份**: 备份 Terraform state 和 Elasticsearch 数据
5. **CI/CD 集成**: 在流水线中使用 Terraform

## 文档

- [Terraform 和 Helm 完整指南](docs/terraform-helm-guide.md)
- [Helm Charts 参考文档](docs/helm-charts-reference.md)
- [多租户架构说明](docs/多租户架构说明.md)
- [原有 README](README.md)

## 迁移指南

如果你正在从旧的 Kustomize 部署迁移:

### 步骤 1: 备份现有数据

```bash
# 创建 Elasticsearch 快照
kubectl exec -n es-serverless elasticsearch-0 -- \
  curl -X PUT "localhost:9200/_snapshot/backup/snapshot_1?wait_for_completion=true"
```

### 步骤 2: 卸载旧部署

```bash
./scripts/deploy.sh uninstall
```

### 步骤 3: 使用 Terraform 重新部署

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
# 编辑配置

terraform init
terraform apply
```

### 步骤 4: 恢复数据

```bash
# 恢复快照
kubectl exec -n es-serverless elasticsearch-0 -- \
  curl -X POST "localhost:9200/_snapshot/backup/snapshot_1/_restore"
```

## 故障排查

### 常见问题

**Q: Terraform apply 超时**
```bash
# 增加超时时间
terraform apply -timeout=30m
```

**Q: Helm release 安装失败**
```bash
# 查看详细日志
helm install elasticsearch ./helm/elasticsearch --debug

# 删除失败的 release
helm uninstall elasticsearch -n es-serverless
```

**Q: PVC 无法绑定**
```bash
# 检查 StorageClass
kubectl get sc

# 创建默认 StorageClass (Docker Desktop)
kubectl patch storageclass hostpath \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

更多故障排查,请查看[完整指南](docs/terraform-helm-guide.md#故障排查)。

## 贡献

欢迎贡献! 提交 Pull Request 或 Issue。

## 许可证

[根据项目许可证]

---

**快速链接**:
- [部署指南](docs/terraform-helm-guide.md)
- [Charts 参考](docs/helm-charts-reference.md)
- [API 文档](README.md#rest-api-endpoints)
