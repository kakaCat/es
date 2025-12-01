# 📦 Terraform 配置已迁移

此目录的内容已整合到新的部署方式中。

## 🔄 新的部署结构

为了更清晰地区分两种部署方式，本项目现在使用以下结构：

### Terraform + Helm 部署方式
👉 **请使用**: [`deployment-terraform/`](../deployment-terraform/)

```bash
cd deployment-terraform
./deploy.sh init
./deploy.sh apply
```

📖 [查看详细文档](../deployment-terraform/README.md)

---

### Shell 脚本部署方式
👉 **请使用**: [`deployment-scripts/`](../deployment-scripts/)

```bash
cd deployment-scripts
./scripts/deploy.sh install
```

📖 [查看详细文档](../deployment-scripts/README.md)

---

## 📚 部署方式对比

不确定使用哪种方式？参考 [DEPLOYMENT.md](../DEPLOYMENT.md) 了解两种部署方式的详细对比。

---

**注意**: 此目录仍然存在是为了保持向后兼容。建议新用户直接使用上述新的部署目录。
