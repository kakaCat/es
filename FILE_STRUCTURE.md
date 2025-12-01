# Terraform/Helm 文件结构

本文档列出所有新增的 Terraform 和 Helm 相关文件。

## 完整文件树

```
es项目/
├── terraform/                              # Terraform 配置目录
│   ├── main.tf                            # 主配置文件 (编排所有模块)
│   ├── variables.tf                       # 全局变量定义 (42 个变量)
│   ├── outputs.tf                         # 输出定义 (服务 URLs)
│   ├── terraform.tfvars.example           # 配置示例文件
│   │
│   ├── modules/                           # Terraform 模块
│   │   ├── elasticsearch/                 # Elasticsearch 模块
│   │   │   ├── main.tf                   # Helm release 定义
│   │   │   ├── variables.tf              # 模块变量
│   │   │   └── outputs.tf                # 模块输出
│   │   │
│   │   ├── control-plane/                # 控制平面模块
│   │   │   ├── main.tf                   # Manager/ShardController/Reporting
│   │   │   ├── variables.tf              # 模块变量
│   │   │   └── outputs.tf                # 模块输出
│   │   │
│   │   ├── monitoring/                   # 监控模块
│   │   │   ├── main.tf                   # Prometheus/Grafana
│   │   │   ├── variables.tf              # 模块变量
│   │   │   └── outputs.tf                # 模块输出
│   │   │
│   │   ├── logging/                      # 日志模块
│   │   │   ├── main.tf                   # Fluentd DaemonSet
│   │   │   ├── variables.tf              # 模块变量
│   │   │   └── outputs.tf                # 模块输出
│   │   │
│   │   └── tenant/                       # 租户模块 ⭐
│   │       ├── main.tf                   # 租户资源定义
│   │       ├── variables.tf              # 租户配置 (20+ 变量)
│   │       └── outputs.tf                # 租户信息输出
│   │
│   └── tenants/                          # 租户实例 (自动生成)
│       └── {org}-{user}-{service}/       # 每个租户一个目录
│           └── main.tf                   # 租户 Terraform 配置
│
├── helm/                                  # Helm Charts 目录
│   ├── elasticsearch/                     # Elasticsearch Chart
│   │   ├── Chart.yaml                    # Chart 元数据
│   │   ├── values.yaml                   # 默认配置 (50+ 参数)
│   │   └── templates/                    # Kubernetes 模板
│   │       ├── statefulset.yaml         # ES StatefulSet
│   │       ├── service.yaml             # Service (ClusterIP + Headless)
│   │       ├── configmap.yaml           # ES 配置
│   │       └── serviceaccount.yaml      # ServiceAccount
│   │
│   ├── control-plane/                    # 控制平面 Chart
│   │   ├── Chart.yaml                    # Chart 元数据
│   │   ├── values.yaml                   # 默认配置 (40+ 参数)
│   │   └── templates/                    # Kubernetes 模板
│   │       ├── _helpers.tpl             # 模板辅助函数
│   │       ├── manager-deployment.yaml  # Manager Deployment
│   │       ├── manager-service.yaml     # Manager Service
│   │       ├── manager-pvc.yaml         # Manager PVC
│   │       ├── shard-controller-deployment.yaml
│   │       ├── reporting-deployment.yaml
│   │       ├── reporting-service.yaml
│   │       ├── serviceaccount.yaml
│   │       └── rbac.yaml                # RBAC 配置
│   │
│   └── monitoring/                       # 监控 Chart
│       ├── Chart.yaml                    # Chart 元数据
│       ├── values.yaml                   # 默认配置 (60+ 参数)
│       └── templates/                    # Kubernetes 模板
│           ├── prometheus-deployment.yaml
│           ├── prometheus-service.yaml
│           ├── prometheus-pvc.yaml
│           ├── prometheus-configmap.yaml
│           ├── grafana-deployment.yaml
│           ├── grafana-service.yaml
│           ├── grafana-pvc.yaml
│           ├── grafana-configmap.yaml
│           ├── serviceaccount.yaml
│           └── rbac.yaml
│
├── scripts/                              # 部署脚本
│   ├── deploy-terraform.sh              # Terraform 部署脚本
│   └── create-tenant.sh                 # 租户创建脚本
│
├── docs/                                 # 文档目录
│   ├── terraform-helm-guide.md          # 完整使用指南 (10,000+ 字)
│   ├── helm-charts-reference.md         # Helm Charts 参考 (5,000+ 字)
│   └── terraform-architecture-diagram.md # 架构图和设计 (3,000+ 字)
│
├── Makefile                              # Make 命令定义 (30+ 命令)
├── TERRAFORM_HELM_README.md             # 主文档
├── QUICK_REFERENCE.md                   # 快速参考
├── IMPLEMENTATION_SUMMARY.md            # 实现总结
└── FILE_STRUCTURE.md                    # 本文件
```

## 文件统计

### Terraform 文件
- **配置文件**: 4 个 (main.tf, variables.tf, outputs.tf, tfvars.example)
- **模块**: 5 个 (elasticsearch, control-plane, monitoring, logging, tenant)
- **模块文件**: 15 个 (每个模块 3 个文件)
- **总计**: 19 个 .tf 文件

### Helm 文件
- **Charts**: 3 个 (elasticsearch, control-plane, monitoring)
- **Chart.yaml**: 3 个
- **values.yaml**: 3 个
- **模板文件**: 22 个
- **总计**: 28 个 YAML 文件

### 脚本文件
- **Bash 脚本**: 2 个 (deploy-terraform.sh, create-tenant.sh)
- **Makefile**: 1 个

### 文档文件
- **主文档**: 4 个 Markdown 文件
- **技术文档**: 3 个 (在 docs/ 目录)
- **总计**: 7 个文档文件

## 文件用途说明

### 核心 Terraform 文件

#### terraform/main.tf
**用途**: 主配置文件,编排所有模块

**内容**:
- Provider 配置 (kubernetes, helm)
- Namespace 创建
- 调用 5 个模块
- 模块间依赖关系

**关键代码**:
```hcl
module "elasticsearch" { ... }
module "control_plane" { ... }
module "monitoring" { ... }
module "logging" { ... }
```

#### terraform/variables.tf
**用途**: 全局变量定义

**变量分类**:
- Kubernetes 配置 (3 个)
- Elasticsearch 配置 (6 个)
- Control Plane 配置 (4 个)
- Monitoring 配置 (4 个)
- Logging 配置 (2 个)

#### terraform/outputs.tf
**用途**: 定义输出,暴露服务 URLs

**输出**:
- namespace
- elasticsearch_service_url
- manager_service_url
- grafana_service_url
- prometheus_service_url

### 模块文件

#### modules/*/main.tf
**用途**: 模块主逻辑

**elasticsearch**: 部署 ES Helm Chart
**control-plane**: 部署控制平面 Chart
**monitoring**: 部署监控 Chart
**logging**: 直接创建 Kubernetes 资源
**tenant**: 创建租户资源 (最复杂的模块)

#### modules/*/variables.tf
**用途**: 模块输入参数

**tenant 模块变量最多** (20+ 个):
- 租户标识 (org, user, service)
- 资源配置 (cpu, memory, disk, gpu)
- 向量配置 (dimension, vector_count, nlist, nprobe)
- 配额和隔离设置

#### modules/*/outputs.tf
**用途**: 模块输出

**典型输出**:
- service_url
- namespace
- release_name
- resource_specs

### Helm Chart 文件

#### Chart.yaml
**用途**: Chart 元数据

**内容**:
- Chart 名称和版本
- 应用版本
- 描述和关键词
- 维护者信息

#### values.yaml
**用途**: 默认配置值

**参数数量**:
- elasticsearch: 50+ 参数
- control-plane: 40+ 参数
- monitoring: 60+ 参数

#### templates/*.yaml
**用途**: Kubernetes 资源模板

**使用 Go 模板语法**:
- `{{ .Values.xxx }}` - 引用配置
- `{{ .Release.Name }}` - Release 名称
- `{{ .Chart.Name }}` - Chart 名称
- `{{- if .Values.xxx }}` - 条件渲染

#### templates/_helpers.tpl
**用途**: 模板辅助函数

**常用函数**:
- `control-plane.name` - Chart 名称
- `control-plane.fullname` - 完整名称
- `control-plane.labels` - 标签集
- `control-plane.selectorLabels` - 选择器标签

### 脚本文件

#### scripts/deploy-terraform.sh
**用途**: Terraform 生命周期管理

**操作**:
- init - 初始化
- plan - 查看计划
- apply - 部署
- destroy - 销毁
- status - 状态检查
- output - 显示输出

**特性**:
- 彩色输出
- 错误处理
- 用户确认
- 自动显示服务 URLs

#### scripts/create-tenant.sh
**用途**: 快速创建租户

**流程**:
1. 解析命令行参数
2. 验证必需参数
3. 创建租户目录
4. 生成 main.tf
5. 运行 terraform init/apply
6. 显示访问信息

**参数**:
- 必需: --org, --user, --service
- 可选: --cpu, --memory, --disk, --gpu, --dimension, --vectors, --replicas

### Makefile

**用途**: 简化常用操作

**命令分组**:
- 平台管理 (8 个)
- 租户管理 (5 个)
- 监控和日志 (6 个)
- 访问服务 (4 个)
- 开发和测试 (4 个)
- 清理和维护 (3 个)
- 快速开始 (2 个)

**高级特性**:
- 彩色输出
- 参数验证
- 错误处理
- 内置帮助文档

### 文档文件

#### TERRAFORM_HELM_README.md
**用途**: 项目主文档

**内容**:
- 快速开始
- 项目结构
- 核心功能
- 架构优势
- 使用场景
- 升级和维护

#### docs/terraform-helm-guide.md
**用途**: 完整使用指南

**章节**:
- 前置要求
- 快速开始
- 部署平台
- 租户管理
- 监控和运维
- 故障排查
- 最佳实践

#### docs/helm-charts-reference.md
**用途**: Helm Charts 配置参考

**内容**:
- 每个 Chart 的详细配置
- 参数说明和默认值
- 使用示例
- Chart 维护

#### docs/terraform-architecture-diagram.md
**用途**: 架构图和设计说明

**内容**:
- ASCII 架构图
- 流程图
- 设计决策
- 技术选型

#### QUICK_REFERENCE.md
**用途**: 快速参考手册

**内容**:
- 常用命令速查
- API 参考
- 配置模板
- 常见问题
- 性能调优

#### IMPLEMENTATION_SUMMARY.md
**用途**: 实现总结

**内容**:
- 项目概述
- 实现内容
- 技术栈
- 架构亮点
- 成果统计

## 依赖关系

### Terraform 模块依赖

```
main.tf
  │
  ├──► module.elasticsearch (独立)
  │
  ├──► module.control_plane
  │       └── depends_on: module.elasticsearch
  │
  ├──► module.monitoring (独立)
  │
  └──► module.logging (独立)
```

### Helm Chart 依赖

```
elasticsearch Chart
  ├── 依赖: 无 (基础)
  └── 被依赖: control-plane

control-plane Chart
  ├── 依赖: elasticsearch (需要 ES URL)
  └── 被依赖: 无

monitoring Chart
  ├── 依赖: elasticsearch (scrape target)
  └── 被依赖: 无
```

### 文件间引用

```
terraform/main.tf
  ├── variables.tf (读取变量)
  ├── outputs.tf (定义输出)
  └── modules/* (调用模块)
      ├── modules/*/main.tf
      ├── modules/*/variables.tf
      └── modules/*/outputs.tf

modules/*/main.tf
  └── ../../helm/*/ (引用 Chart 路径)
      ├── Chart.yaml
      ├── values.yaml
      └── templates/*.yaml
```

## 配置覆盖顺序

### Terraform
```
1. variables.tf (默认值)
   ↓
2. terraform.tfvars (用户配置)
   ↓
3. 环境变量 TF_VAR_*
   ↓
4. 命令行 -var 参数
```

### Helm
```
1. values.yaml (Chart 默认值)
   ↓
2. Terraform module 传递的 values
   ↓
3. helm install -f custom-values.yaml
   ↓
4. helm install --set key=value
```

## 版本控制建议

### 应该提交的文件
✅ 所有 .tf 文件
✅ 所有 Chart 文件 (yaml, tpl)
✅ 所有脚本文件
✅ 所有文档文件
✅ Makefile
✅ terraform.tfvars.example

### 不应该提交的文件
❌ terraform.tfstate
❌ terraform.tfstate.backup
❌ .terraform/
❌ terraform.tfvars (包含敏感信息)
❌ tenants/ (自动生成)
❌ *.tfvars (除了 .example)

### .gitignore 建议
```gitignore
# Terraform
*.tfstate
*.tfstate.*
.terraform/
.terraform.lock.hcl
terraform.tfvars
!terraform.tfvars.example

# 租户配置 (自动生成)
terraform/tenants/

# 备份
backups/
*.backup

# OS
.DS_Store
```

## 文件大小统计

### Terraform
- 小文件 (<50 行): variables.tf, outputs.tf
- 中文件 (50-150 行): main.tf, modules/*/main.tf
- 大文件 (>150 行): modules/tenant/main.tf, modules/tenant/variables.tf

### Helm
- 小文件 (<30 行): Chart.yaml, serviceaccount.yaml
- 中文件 (30-100 行): service.yaml, deployment.yaml
- 大文件 (>100 行): values.yaml, statefulset.yaml

### 文档
- 小文件 (<1000 字): IMPLEMENTATION_SUMMARY.md
- 中文件 (1000-5000 字): helm-charts-reference.md
- 大文件 (>5000 字): terraform-helm-guide.md

## 更新频率

### 高频更新
- terraform.tfvars (每次部署)
- tenants/*/main.tf (创建租户时)

### 中频更新
- values.yaml (调整配置)
- main.tf (添加新模块)

### 低频更新
- Chart.yaml (版本发布)
- templates/*.yaml (功能变更)
- 文档 (重大更新)

## 总结

本项目新增:
- **54 个文件** (不含自动生成)
- **~5000 行代码** (Terraform + Helm + 脚本)
- **~20000 字文档**

完整的基础设施即代码实现,支持:
- ✅ 平台一键部署
- ✅ 租户自动创建
- ✅ 配置版本控制
- ✅ 状态管理
- ✅ 完整文档
