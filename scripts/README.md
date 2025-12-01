# 📦 部署脚本已迁移

此目录的内容已整合到新的部署方式中。

## 🔄 新的部署结构

为了更清晰地区分两种部署方式，本项目现在使用以下结构：

### 🚀 Shell 脚本部署方式（推荐开发测试）
👉 **请使用**: [`deployment-scripts/`](../deployment-scripts/)

```bash
cd deployment-scripts
./scripts/deploy.sh install
```

📖 [查看详细文档](../deployment-scripts/README.md)

---

### 🔧 Terraform + Helm 部署方式（推荐生产环境）
👉 **请使用**: [`deployment-terraform/`](../deployment-terraform/)

```bash
cd deployment-terraform
./deploy.sh init
./deploy.sh apply
```

📖 [查看详细文档](../deployment-terraform/README.md)

---

## 📋 脚本说明

本目录包含以下部署和管理脚本：

### 核心部署脚本

| 脚本 | 功能 | 使用方式 |
|------|------|---------|
| `deploy.sh` | 主部署脚本 | `./scripts/deploy.sh install` |
| `cluster.sh` | 集群管理 | `./scripts/cluster.sh create <ns> <replicas>` |
| `create-tenant.sh` | 创建租户 | `./scripts/create-tenant.sh --tenant-org-id org-001 --user alice` |

### 插件和构建

| 脚本 | 功能 | 使用方式 |
|------|------|---------|
| `build-plugin.sh` | 构建 IVF 插件 | `./scripts/build-plugin.sh` |
| `build-reporting.sh` | 构建报告服务 | `./scripts/build-reporting.sh` |
| `build.sh` | 构建所有组件 | `./scripts/build.sh` |

### 运维管理

| 脚本 | 功能 | 使用方式 |
|------|------|---------|
| `monitor.sh` | 监控集群状态 | `./scripts/monitor.sh` |
| `shard-management.sh` | 分片管理 | `./scripts/shard-management.sh rebalance` |

### 备份恢复

| 脚本 | 功能 | 使用方式 |
|------|------|---------|
| `backup-es-snapshot.sh` | ES 快照备份 | `./scripts/backup-es-snapshot.sh create` |
| `backup-metadata.sh` | 元数据备份 | `./scripts/backup-metadata.sh backup` |
| `restore-from-snapshot.sh` | 从快照恢复 | `./scripts/restore-from-snapshot.sh <snapshot>` |

### 测试

| 脚本 | 功能 | 使用方式 |
|------|------|---------|
| `test-ivf.sh` | IVF 功能测试 | `./scripts/test-ivf.sh` |
| `deploy-terraform.sh` | Terraform 部署 | `./scripts/deploy-terraform.sh apply` |

---

## 📖 快速开始

### 开发测试部署

```bash
# 1. 进入新的部署目录
cd deployment-scripts

# 2. 一键部署
./scripts/deploy.sh install

# 3. 查看状态
./scripts/deploy.sh status

# 4. 创建租户
./scripts/create-tenant.sh \
  --tenant-org-id org-001 \
  --user alice \
  --service vector-search \
  --replicas 3
```

### 生产环境部署

```bash
# 使用 Terraform + Helm
cd deployment-terraform
./deploy.sh init
./deploy.sh apply
```

---

## 📚 部署方式对比

| 特性 | Shell 脚本 | Terraform + Helm |
|------|-----------|-----------------|
| 学习曲线 | ✅ 低 | ⚠️ 中等 |
| 部署速度 | ✅ 快 | ⚠️ 较慢 |
| 适用场景 | 开发测试 | 生产环境 |
| 状态管理 | ❌ 手动 | ✅ 自动 |
| 回滚能力 | ❌ 困难 | ✅ 简单 |

📖 [查看详细对比](../DEPLOYMENT.md)

---

**注意**: 此目录仍然存在是为了保持向后兼容。建议新用户直接使用 `deployment-scripts/`。
