# 📦 Helm Charts 已迁移

此目录的内容已整合到新的部署方式中。

## 🔄 新的部署结构

为了更清晰地区分两种部署方式，本项目现在使用以下结构：

### Terraform + Helm 部署方式
👉 **请使用**: [`deployment-terraform/`](../deployment-terraform/)

此目录包含：
- Terraform 基础设施配置
- Helm Charts (elasticsearch, control-plane, monitoring)

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

## 📚 Helm Charts 说明

本目录包含以下 Helm Charts：

### 1. elasticsearch/
Elasticsearch 集群 + IVF 向量搜索插件

**主要配置**:
- `values.yaml` - 默认配置
- `templates/statefulset.yaml` - ES StatefulSet
- `templates/service.yaml` - ClusterIP + Headless Service

**使用方式**:
```bash
helm install elasticsearch ./helm/elasticsearch \
  --set replicaCount=3 \
  --set ivfPlugin.enabled=true
```

### 2. control-plane/
控制平面服务（Manager + ShardController + Reporting）

**包含服务**:
- ES Serverless Manager
- Shard Controller
- Reporting Service

**使用方式**:
```bash
helm install control-plane ./helm/control-plane
```

### 3. monitoring/
监控栈（Prometheus + Grafana）

**使用方式**:
```bash
helm install monitoring ./helm/monitoring
```

---

## 📖 部署方式对比

不确定使用哪种方式？参考 [DEPLOYMENT.md](../DEPLOYMENT.md) 了解两种部署方式的详细对比。

---

**注意**: 此目录仍然存在是为了保持向后兼容。建议新用户直接使用 `deployment-terraform/` 或 `deployment-scripts/`。
