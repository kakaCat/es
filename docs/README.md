# ES Serverless 文档中心

欢迎使用 ES Serverless 平台文档！

## 📚 文档导航

### 🏗️ 架构设计
- [系统架构总览](architecture.md)
- [多租户架构](architecture/multi-tenancy.md)
- [分片复制机制](architecture/shard-replication.md)
- [自动扩展与配额管理](architecture/auto-scaling.md)

### 🚀 部署指南
- [部署总览](deployment/README.md) - 两种部署方式对比和选择
- [Terraform + Helm 部署](deployment/terraform-helm.md)
- [Shell 脚本部署](deployment/shell-scripts.md)
- [GPU 加速配置](deployment/gpu-setup.md)

### 💻 开发文档
- [开发环境搭建](development/setup.md)
- [ES 插件开发指南](development/es-plugin.md)
- [控制平面开发](development/control-plane.md)
- [前端开发](development/frontend.md)
- [测试指南](development/testing.md)

### 🛠️ 运维手册
- [监控与告警](operations/monitoring.md)
- [备份与恢复](operations/disaster-recovery.md)
- [部署上报机制](operations/deployment-reporting.md)
- [逻辑删除机制](operations/logical-deletion.md)
- [故障排查指南](operations/troubleshooting.md)

### 📡 API 文档
- [REST API 参考](api.md)
- [向量搜索 API](api/vector-search.md)

### 📦 其他资源
- [Helm Charts 参考](helm-charts-reference.md)
- [Terraform 架构图](terraform-architecture-diagram.md)
- [时序图集合](时序图集合.md)

### 🗄️ 归档文档
- [归档中心](archive/) - 历史文档和项目记录
- [项目整理归档 (2025-12-01)](archive/2025-12-01-project-cleanup/) - v2.0 重构和整理记录
- [实现总结归档](archive/implementation-summary/)
- [需求文档归档](archive/requirements/)

## 🎯 快速开始

### 新用户推荐阅读顺序

1. **了解系统**
   - 先阅读 [README.md](../README.md)
   - 然后查看 [系统架构总览](architecture.md)

2. **选择部署方式**
   - 阅读 [部署总览](deployment/README.md)
   - 选择适合您的部署方式

3. **开始部署**
   - 开发测试：[Shell 脚本部署](deployment/shell-scripts.md)
   - 生产环境：[Terraform + Helm 部署](deployment/terraform-helm.md)

4. **深入了解**
   - [多租户架构](architecture/multi-tenancy.md)
   - [自动扩展机制](architecture/auto-scaling.md)
   - [GPU 加速配置](deployment/gpu-setup.md)

## 🔍 按角色查找文档

### 开发人员
- [开发环境搭建](development/setup.md)
- [ES 插件开发](development/es-plugin.md)
- [控制平面开发](development/control-plane.md)
- [API 文档](api.md)

### 运维人员
- [部署指南](deployment/README.md)
- [监控运维](operations/monitoring.md)
- [备份恢复](operations/disaster-recovery.md)
- [故障排查](operations/troubleshooting.md)

### 架构师
- [系统架构](architecture.md)
- [多租户设计](architecture/multi-tenancy.md)
- [自动扩展](architecture/auto-scaling.md)
- [分片复制](architecture/shard-replication.md)

## 📝 文档贡献

发现文档问题或想要改进？请查看 [CONTRIBUTING.md](../CONTRIBUTING.md)

## 📖 文档版本

- 当前版本：v1.0.0
- 最后更新：2025-12-01

---

💡 **提示**：使用左侧目录导航或搜索功能快速找到您需要的文档。
