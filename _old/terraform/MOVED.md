# 📦 目录已迁移

此目录 `terraform/` 已迁移到新位置。

## ✅ 新位置

```
terraform/ → deployments/terraform/
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

## 🚀 使用 Terraform 部署

新路径下的部署命令：

```bash
cd deployments/terraform
terraform init
terraform plan
terraform apply
```

## 🔗 更多信息

查看完整的部署文档：
- [Terraform + Helm 部署指南](../docs/deployment/terraform-helm.md)
- [部署总览](../docs/deployment/README.md)
- [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md)

---

**注意**：此目录仅保留用于向后兼容，后续版本可能移除。请尽快更新您的引用路径。
