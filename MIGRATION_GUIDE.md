# 项目迁移指南

> ES Serverless Platform v2.0 标准化结构迁移说明

**迁移日期**: 2025-12-01
**版本**: v1.0 → v2.0

---

## 📋 迁移概述

项目已完成标准化重构，采用清晰的开源项目目录结构。所有文件已迁移到新位置，旧目录保留用于向后兼容。

---

## 🗂️ 目录映射表

### 源代码迁移

| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `server/` | `src/control-plane/` | Go 控制平面服务 |
| `es-plugin/` | `src/es-plugin/` | ES IVF 向量插件 |
| `frontend/` | `src/frontend/` | Web 管理界面 |

### 部署配置迁移

| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `terraform/` | `deployments/terraform/` | Terraform IaC 配置 |
| `helm/` | `deployments/helm/` | Helm Charts |
| `k8s/` | `deployments/kubernetes/` | Kubernetes YAML |
| `docker/` | `deployments/docker/` | Docker Compose |

### 脚本迁移

| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `scripts/deploy.sh` | `scripts/deploy/deploy.sh` | 部署脚本 |
| `scripts/cluster.sh` | `scripts/deploy/cluster.sh` | 集群管理 |
| `scripts/create-tenant.sh` | `scripts/deploy/create-tenant.sh` | 租户创建 |
| `scripts/build-plugin.sh` | `scripts/build/build-plugin.sh` | 构建 ES 插件 |
| `scripts/build-reporting.sh` | `scripts/build/build-reporting.sh` | 构建上报服务 |
| `scripts/backup*.sh` | `scripts/ops/` | 运维备份脚本 |
| `scripts/test-ivf.sh` | `scripts/dev/test-ivf.sh` | 开发测试脚本 |

### 文档迁移

| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `docs/多租户架构说明.md` | `docs/architecture/multi-tenancy.md` | 架构文档 |
| `docs/分片数据同步*.md` | `docs/architecture/shard-replication.md` | 架构文档 |
| `docs/自动扩展配额管理*.md` | `docs/architecture/auto-scaling.md` | 架构文档 |
| `docs/灾难恢复手册.md` | `docs/operations/disaster-recovery.md` | 运维文档 |
| `docs/逻辑删除*.md` | `docs/operations/logical-deletion.md` | 运维文档 |
| `docs/部署上报*.md` | `docs/operations/deployment-reporting.md` | 运维文档 |
| `IVF实现*.md` | `docs/archive/implementation-summary/` | 归档文档 |
| `具体要求.md` | `docs/archive/requirements/` | 归档文档 |

### 示例迁移

| 旧路径 | 新路径 | 说明 |
|--------|--------|------|
| `demo/` | `examples/` | 使用示例和演示 |

---

## 🚀 新目录结构

```
es-paas/es/
│
├── README.md                     # 项目总览（已更新）
├── CONTRIBUTING.md               # 贡献指南（新增）
├── PROJECT_STRUCTURE.md          # 项目结构说明（新增）
├── MIGRATION_GUIDE.md            # 迁移指南（本文档）
├── CLAUDE.md                     # AI 助手配置
│
├── src/                          # 💻 源代码
│   ├── control-plane/            # Go 控制平面（原 server/）
│   ├── es-plugin/                # ES 插件（原 es-plugin/）
│   └── frontend/                 # 前端（原 frontend/）
│
├── deployments/                  # 🚀 部署配置
│   ├── terraform/                # Terraform（原 terraform/）
│   ├── helm/                     # Helm Charts（原 helm/）
│   ├── kubernetes/               # K8s YAML（原 k8s/）
│   └── docker/                   # Docker Compose（原 docker/）
│
├── scripts/                      # 🛠️ 工具脚本
│   ├── deploy/                   # 部署脚本
│   ├── build/                    # 构建脚本
│   ├── ops/                      # 运维脚本
│   └── dev/                      # 开发辅助
│
├── docs/                         # 📚 文档中心
│   ├── README.md                 # 文档索引（新增）
│   ├── architecture/             # 架构设计
│   ├── deployment/               # 部署指南
│   ├── development/              # 开发文档
│   ├── operations/               # 运维手册
│   ├── api/                      # API 文档
│   └── archive/                  # 归档文档
│
├── tests/                        # 🧪 测试
│   ├── unit/                     # 单元测试
│   ├── integration/              # 集成测试
│   └── e2e/                      # 端到端测试
│
├── examples/                     # 📖 示例代码（原 demo/）
├── configs/                      # ⚙️ 配置文件
└── tools/                        # 🔧 开发工具
```

---

## ✅ 迁移检查清单

### 对于开发人员

- [ ] 更新本地仓库路径引用
  ```bash
  # 旧路径
  cd server && go build

  # 新路径
  cd src/control-plane && go build
  ```

- [ ] 更新 IDE 项目配置
  - Go 项目路径: `src/control-plane`
  - Java 项目路径: `src/es-plugin`
  - 前端项目路径: `src/frontend`

- [ ] 更新脚本中的路径引用
  - 部署脚本引用
  - 构建脚本引用
  - CI/CD 配置

- [ ] 阅读新的文档结构
  - [README.md](README.md) - 项目总览
  - [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - 结构说明
  - [CONTRIBUTING.md](CONTRIBUTING.md) - 贡献指南

### 对于运维人员

- [ ] 更新部署脚本路径
  ```bash
  # 旧路径
  ./scripts/deploy.sh install

  # 新路径
  ./scripts/deploy/deploy.sh install
  ```

- [ ] 更新 Terraform 配置路径
  ```bash
  # 旧路径
  cd terraform && terraform init

  # 新路径
  cd deployments/terraform && terraform init
  ```

- [ ] 更新 Helm 命令路径
  ```bash
  # 旧路径
  cd helm && helm install ...

  # 新路径
  cd deployments/helm && helm install ...
  ```

- [ ] 更新运维脚本引用
  - 备份脚本: `scripts/ops/backup*.sh`
  - 监控脚本: `scripts/ops/monitor.sh`

### 对于文档维护者

- [ ] 更新文档内部链接
  - 检查所有相对路径
  - 更新跨文档引用

- [ ] 验证文档分类
  - 架构文档 → `docs/architecture/`
  - 部署文档 → `docs/deployment/`
  - 运维文档 → `docs/operations/`

- [ ] 归档旧文档
  - 已废弃文档 → `docs/archive/`

---

## 🔄 常见问题

### Q1: 旧目录什么时候会删除？

**A**: 旧目录暂时保留用于向后兼容。建议在 v2.1 版本（预计 1 个月后）之前完成迁移。届时旧目录将被标记为废弃，并在 v3.0 版本中移除。

### Q2: 我的脚本引用了旧路径，会不会失效？

**A**: 短期内不会。旧目录中的文件仍然存在（只是添加了 MOVED.md 说明）。但为了避免未来问题，请尽快更新引用。

### Q3: 如何快速查找某个文件的新位置？

**A**: 参考本文档的"目录映射表"，或查看旧目录中的 `MOVED.md` 文件。

### Q4: 我需要同时维护多个分支，如何处理？

**A**: 建议：
- 主分支（main）: 使用新结构
- 功能分支: 基于主分支创建
- 旧版本分支: 保持旧结构不变

### Q5: CI/CD 配置需要更新吗？

**A**: 是的。请更新：
- 构建路径: `src/control-plane`, `src/es-plugin`
- 部署路径: `deployments/`
- 脚本路径: `scripts/deploy/`, `scripts/build/`

---

## 📖 详细文档链接

### 新用户

1. **开始使用**:
   - [README.md](README.md) - 项目总览和快速开始
   - [项目结构说明](PROJECT_STRUCTURE.md) - 详细的目录结构
   - [部署总览](docs/deployment/README.md) - 选择部署方式

2. **开发环境**:
   - [开发环境搭建](docs/development/setup.md)
   - [贡献指南](CONTRIBUTING.md)

### 现有用户

1. **迁移相关**:
   - [本文档](MIGRATION_GUIDE.md) - 迁移指南（您正在阅读）
   - 旧目录中的 `MOVED.md` - 单个目录的迁移说明

2. **功能文档**（路径已更新）:
   - [系统架构](docs/architecture.md)
   - [多租户架构](docs/architecture/multi-tenancy.md)
   - [自动扩展](docs/architecture/auto-scaling.md)
   - [监控运维](docs/operations/monitoring.md)

---

## 🛠️ 自动化迁移脚本

如果您有大量脚本需要更新路径，可以使用以下命令批量替换：

```bash
# 备份当前脚本
cp -r your-scripts/ your-scripts.backup/

# 批量替换路径（示例）
find your-scripts/ -type f -name "*.sh" -exec sed -i '' 's|cd server|cd src/control-plane|g' {} +
find your-scripts/ -type f -name "*.sh" -exec sed -i '' 's|cd es-plugin|cd src/es-plugin|g' {} +
find your-scripts/ -type f -name "*.sh" -exec sed -i '' 's|scripts/deploy.sh|scripts/deploy/deploy.sh|g' {} +
find your-scripts/ -type f -name "*.sh" -exec sed -i '' 's|cd terraform|cd deployments/terraform|g' {} +
```

**注意**: 运行前请先备份，并根据实际情况调整替换规则。

---

## 📞 获取帮助

如果您在迁移过程中遇到问题：

1. 查看 [项目结构说明](PROJECT_STRUCTURE.md)
2. 查看旧目录中的 `MOVED.md` 文件
3. 提交 GitHub Issue（标注 `migration` 标签）
4. 联系项目维护者

---

## 🎯 总结

项目重构的核心目标是：
- ✅ **更清晰**：源代码、部署、脚本、文档分类明确
- ✅ **更标准**：符合开源项目最佳实践
- ✅ **更易用**：新用户快速上手，老用户平滑迁移
- ✅ **更可维护**：文档和代码组织合理，便于长期维护

感谢您的理解和配合！

---

**版本历史**:
- v1.0 (2025-12-01) - 初始迁移指南发布
