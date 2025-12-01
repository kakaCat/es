# ES Serverless Platform

> 基于 IVF 算法的 Serverless Elasticsearch 向量搜索平台

**版本**: v2.0
**更新日期**: 2025-12-01

---

## 📖 项目概览

ES Serverless Platform 是一个完全托管的 Elasticsearch 平台，提供：
- 🚀 **Serverless 架构**：自动资源分配和弹性伸缩
- 🔍 **向量检索**：基于 IVF 算法的高维向量搜索（类似 FAISS）
- 🏢 **多租户隔离**：组织级别的资源和数据隔离
- 📊 **智能监控**：实时监控和自动告警
- 🔄 **自动恢复**：分片同步监控和故障自动恢复
- 💾 **数据安全**：自动备份和灾难恢复

---

## 🚀 快速开始

### 前置条件

在部署系统前，请确保已安装以下环境：

- **Docker Desktop** with Kubernetes enabled
- **kubectl** CLI
- **Go 1.21+** (本地开发)
- **Bash** shell

### Kubernetes 环境配置

1. **启用 Kubernetes**（Docker Desktop）:
   - 打开 Docker Desktop
   - 进入 Settings > Kubernetes
   - 勾选 "Enable Kubernetes"
   - 点击 "Apply & Restart"

2. **验证 Kubernetes 运行状态**:
   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

如遇 Kubernetes 配置问题，请参考 [KUBERNETES_SETUP_ISSUES.md](KUBERNETES_SETUP_ISSUES.md)。

## ✨ 核心特性

### 多租户与隔离
- 🏢 **组织级隔离**：通过租户组织ID实现资源和数据隔离 ([详细说明](docs/architecture/multi-tenancy.md))
- 🔒 **配额管理**：自动扩展时检查租户配额，防止资源超限 ([配额管理](docs/architecture/auto-scaling.md))
- 🗑️ **逻辑删除**：安全的数据删除机制，支持恢复 ([逻辑删除](docs/operations/logical-deletion.md))

### 向量检索与性能
- 🔍 **IVF 算法**：高维向量搜索，支持亿级向量规模
- 🎯 **混合检索**：向量检索 + 结构化过滤
- ⚡ **GPU 加速**：可选 GPU 加速向量计算

### 自动化与智能运维
- 🚀 **自动扩缩容**：基于 CPU、内存、QPS 的智能扩缩容
- 🔄 **分片管理**：自动分片均衡和副本同步 ([分片复制](docs/architecture/shard-replication.md))
- 🔧 **故障恢复**：自动检测和恢复失败节点
- 📊 **监控告警**：Prometheus + Grafana 实时监控 ([监控配置](docs/operations/monitoring.md))

### 数据安全与可靠性
- 💾 **自动备份**：ES 快照和元数据自动备份到 MinIO/S3 ([灾难恢复](docs/operations/disaster-recovery.md))
- 🗄️ **双存储模式**：支持 PostgreSQL 或文件系统存储元数据
- 🔐 **安全认证**：支持 Token 认证和访问审计

## 📦 部署指南

本项目提供两种部署方式，请根据使用场景选择：

| 部署方式 | 适用场景 | 复杂度 | 生产就绪 |
|---------|---------|--------|---------|
| **Shell 脚本** | 开发测试、快速验证 | ⭐ 简单 | ❌ |
| **Terraform + Helm** | 生产环境、多租户 | ⭐⭐⭐ 中等 | ✅ |

### 🚀 方式一：Shell 脚本部署（推荐开发测试）

```bash
# 一键部署
./scripts/deploy/deploy.sh install

# 查看状态
./scripts/deploy/deploy.sh status

# 卸载
./scripts/deploy/deploy.sh uninstall
```

👉 **详细文档**: [Shell 脚本部署指南](docs/deployment/shell-scripts.md)

---

### 🔧 方式二：Terraform + Helm 部署（推荐生产环境）

```bash
cd deployments/terraform

# 初始化
terraform init

# 预览变更
terraform plan

# 执行部署
terraform apply
```

👉 **详细文档**: [Terraform + Helm 部署指南](docs/deployment/terraform-helm.md)

---

**📚 完整对比和选择指南**: 查看 [部署总览](docs/deployment/README.md)

## 💻 本地开发

### 控制平面开发

```bash
cd src/control-plane
go build -o manager .
./manager
```

控制平面提供 REST API (端口 8080)，管理集群生命周期、自动扩缩容和监控。

### ES 插件开发

```bash
cd src/es-plugin
./gradlew build

# 生成的插件位于
# build/distributions/es-ivf-plugin-*.zip
```

### 前端开发

```bash
cd src/frontend

# 启动 HTTP 服务器
python -m http.server 8000

# 访问
# http://localhost:8000
```

**前端功能**：
- 创建/删除集群
- 查询集群状态
- 多租户管理

👉 **详细开发指南**: [开发环境搭建](docs/development/setup.md)

## 🏗️ 系统架构

ES Serverless Platform 采用三层架构设计：

### 控制平面 (Control Plane)
- **管理器** (Manager): 集群生命周期管理
- **自动扩缩容** (AutoScaler): 基于指标的智能扩缩容
- **分片控制器** (ShardController): 分片均衡和迁移
- **副本监控器** (ReplicationMonitor): 副本同步监控
- **一致性检查器** (ConsistencyChecker): 数据一致性验证

### 数据平面 (Data Plane)
- **Elasticsearch 集群**: 基于 StatefulSet 部署
- **IVF 向量插件**: 自定义向量检索算法
- **持久化存储**: PersistentVolume 数据持久化

### 监控与日志
- **Prometheus**: 指标采集
- **Grafana**: 可视化监控
- **Fluentd**: 日志聚合

👉 **详细架构**: [系统架构总览](docs/architecture.md)

---

## 📡 API 接口

控制平面提供以下 REST API (端口 8080):

### 集群管理
- `POST /clusters` - 创建集群
- `DELETE /clusters` - 删除集群
- `GET /clusters` - 列出所有集群
- `GET /clusters/{namespace}` - 获取集群详情
- `POST /clusters/scale` - 扩缩容集群

### 向量索引
- `POST /vector-indexes` - 创建向量索引
- `GET /vector-indexes` - 列出向量索引
- `DELETE /vector-indexes` - 删除向量索引

### 监控与状态
- `GET /deployments` - 部署状态
- `GET /metrics` - 监控指标
- `GET /qps/{namespace}` - QPS 指标

👉 **完整 API 文档**: [REST API 参考](docs/api.md)