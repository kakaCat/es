# 📦 目录已迁移

此目录 `helm/` 已迁移到新位置。

## ✅ 新位置

```
helm/ → deployments/helm/
```

## 📖 为什么迁移？

项目已进行标准化重构，采用更清晰的目录结构：
- `src/` - 所有源代码
- `deployments/` - 所有部署配置
  - `terraform/` - Terraform IaC
  - `helm/` - Helm Charts
  - `kubernetes/` - K8s YAML
  - `docker/` - Docker Compose
- `scripts/` - 工具脚本
- `docs/` - 文档中心

## 🚀 使用 Helm 部署

新路径下的部署命令：

```bash
cd deployments/helm

# 部署 Elasticsearch 集群
helm install es-cluster elasticsearch/ \
  --set replicas=3

# 部署控制平面
helm install control-plane control-plane/

# 部署监控
helm install monitoring monitoring/
```

## 🔗 更多信息

查看完整的部署文档：
- [Terraform + Helm 部署指南](../docs/deployment/terraform-helm.md)
- [Helm Charts 参考](../docs/helm-charts-reference.md)
- [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md)

---

**注意**：此目录仅保留用于向后兼容，后续版本可能移除。请尽快更新您的引用路径。
