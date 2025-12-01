# ES Serverless - 快速参考

## 一分钟上手

```bash
# 1. 部署平台
make quick-start

# 2. 创建租户
make tenant-create ORG=myorg USER=alice SERVICE=app1

# 3. 访问服务
make port-forward-manager   # http://localhost:8080
make port-forward-grafana   # http://localhost:3000
```

## 常用命令速查

### Makefile 快捷方式

```bash
# 平台管理
make init              # 初始化 Terraform
make plan              # 查看执行计划
make apply             # 部署平台
make status            # 查看状态
make destroy           # 销毁资源

# 租户管理
make tenant-create ORG=org USER=user SERVICE=svc
make tenant-list       # 列出所有租户
make tenant-status TENANT=org-user-svc
make tenant-delete TENANT=org-user-svc

# 日志查看
make logs-manager      # Manager 日志
make logs-elasticsearch
make logs-tenant TENANT=org-user-svc

# 访问服务
make port-forward-manager    # localhost:8080
make port-forward-grafana    # localhost:3000
make port-forward-prometheus # localhost:9090
make port-forward-es         # localhost:9200

# 快速开始
make quick-start       # 初始化 + 部署
make quick-demo        # 部署 + 创建示例租户
```

### 原生 Terraform 命令

```bash
cd terraform

# 基本操作
terraform init         # 初始化
terraform plan         # 查看计划
terraform apply        # 应用变更
terraform destroy      # 销毁资源

# 查看资源
terraform show         # 显示状态
terraform output       # 显示输出
terraform state list   # 列出资源

# 指定目标
terraform apply -target=module.elasticsearch
terraform destroy -target=module.monitoring
```

### Helm 命令

```bash
# 列出 releases
helm list -n es-serverless

# 查看状态
helm status elasticsearch -n es-serverless

# 查看值
helm get values elasticsearch -n es-serverless

# 升级
helm upgrade elasticsearch ./helm/elasticsearch \
  --set replicaCount=5 \
  -n es-serverless

# 回滚
helm rollback elasticsearch -n es-serverless

# 卸载
helm uninstall elasticsearch -n es-serverless
```

### Kubectl 命令

```bash
# 查看资源
kubectl get pods -n es-serverless
kubectl get svc -n es-serverless
kubectl get pvc -n es-serverless

# 查看租户
kubectl get ns -l es-cluster=true
kubectl get all -n org-001-alice-vector-search

# 查看日志
kubectl logs -f elasticsearch-0 -n es-serverless
kubectl logs -l app=es-control-plane-manager -n es-serverless -f

# 进入容器
kubectl exec -it elasticsearch-0 -n es-serverless -- bash

# 端口转发
kubectl port-forward svc/elasticsearch 9200:9200 -n es-serverless

# 资源使用
kubectl top nodes
kubectl top pods -n es-serverless
```

## API 快速参考

### Manager API (localhost:8080)

```bash
# 集群管理
# 创建集群
curl -X POST http://localhost:8080/clusters \
  -H 'Content-Type: application/json' \
  -d '{
    "tenant_org_id": "org-001",
    "user": "alice",
    "service_name": "vector-search",
    "replicas": 3,
    "cpu": "2000m",
    "memory": "4Gi"
  }'

# 列出集群
curl http://localhost:8080/clusters

# 获取集群详情
curl http://localhost:8080/clusters/org-001-alice-vector-search

# 删除集群
curl -X DELETE http://localhost:8080/clusters \
  -H 'Content-Type: application/json' \
  -d '{"namespace": "org-001-alice-vector-search"}'

# 扩容集群
curl -X POST http://localhost:8080/clusters/scale \
  -H 'Content-Type: application/json' \
  -d '{
    "namespace": "org-001-alice-vector-search",
    "replicas": 5
  }'

# 向量索引管理
# 创建索引
curl -X POST http://localhost:8080/vector-indexes \
  -H 'Content-Type: application/json' \
  -d '{
    "namespace": "org-001-alice-vector-search",
    "index_name": "products",
    "dimension": 256,
    "nlist": 100,
    "nprobe": 10
  }'

# 列出索引
curl http://localhost:8080/vector-indexes

# 监控
# 查看部署状态
curl http://localhost:8080/deployments

# 查看指标
curl http://localhost:8080/metrics

# 查看 QPS
curl http://localhost:8080/qps/org-001-alice-vector-search

# 租户查询
# 所有租户容器
curl http://localhost:8080/tenant/containers

# 特定用户的容器
curl http://localhost:8080/tenant/containers/alice/vector-search

# 组织下所有容器
curl http://localhost:8080/tenant/containers/org/org-001
```

### Elasticsearch API (localhost:9200)

```bash
# 集群健康
curl http://localhost:9200/_cluster/health?pretty

# 节点信息
curl http://localhost:9200/_cat/nodes?v

# 索引列表
curl http://localhost:9200/_cat/indices?v

# 分片信息
curl http://localhost:9200/_cat/shards?v

# 集群设置
curl http://localhost:9200/_cluster/settings?pretty
```

## 配置文件模板

### terraform.tfvars

```hcl
# Kubernetes 配置
kubeconfig_path = "~/.kube/config"
kube_context    = "docker-desktop"
namespace       = "es-serverless"

# Elasticsearch
elasticsearch_replicas     = 3
elasticsearch_storage_size = "10Gi"
storage_class             = "hostpath"

elasticsearch_resources = {
  requests = { cpu = "1000m", memory = "2Gi" }
  limits   = { cpu = "2000m", memory = "4Gi" }
}

# 控制平面镜像
manager_image           = "es-serverless-manager:latest"
shard_controller_image  = "shard-controller:latest"
reporting_service_image = "reporting-service:latest"

# 监控
prometheus_enabled         = true
grafana_enabled           = true
prometheus_retention_days = 15

# 日志
fluentd_enabled = true
```

### 自定义 Helm values

```yaml
# custom-values.yaml
replicaCount: 5

resources:
  requests:
    cpu: 2000m
    memory: 4Gi
  limits:
    cpu: 4000m
    memory: 8Gi

persistence:
  size: 50Gi

ivfPlugin:
  config:
    dimension: 512
    vectorCount: 10000000
```

## 环境要求

### 开发环境
- **CPU**: 4 cores
- **内存**: 8 GB
- **磁盘**: 50 GB
- **Kubernetes**: Docker Desktop / Kind

### 测试环境
- **CPU**: 8 cores
- **内存**: 16 GB
- **磁盘**: 100 GB
- **Kubernetes**: 任何托管 K8s

### 生产环境
- **CPU**: 16+ cores
- **内存**: 32+ GB
- **磁盘**: 500+ GB
- **Kubernetes**: GKE / EKS / AKS

## 常见问题速查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl get pods -n es-serverless

# 查看 Pod 事件
kubectl describe pod <pod-name> -n es-serverless

# 查看日志
kubectl logs <pod-name> -n es-serverless

# 常见原因:
# 1. 资源不足 -> 增加 limits
# 2. 镜像拉取失败 -> 检查镜像仓库
# 3. 配置错误 -> 检查 ConfigMap
```

### PVC 无法绑定

```bash
# 检查 StorageClass
kubectl get sc

# 设置默认 StorageClass
kubectl patch storageclass hostpath \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

### Helm release 失败

```bash
# 查看 release 状态
helm status <release-name> -n es-serverless

# 查看 release 历史
helm history <release-name> -n es-serverless

# 回滚到上一个版本
helm rollback <release-name> -n es-serverless

# 完全删除并重新安装
helm uninstall <release-name> -n es-serverless
terraform apply -target=module.<module-name>
```

### Terraform 状态锁定

```bash
# 查看锁定信息
cd terraform
terraform force-unlock <LOCK_ID>

# 谨慎: 只在确认没有其他 terraform 进程运行时使用
```

### 网络连接问题

```bash
# 测试 DNS 解析
kubectl run -it --rm debug --image=busybox -n es-serverless -- \
  nslookup elasticsearch

# 测试服务连接
kubectl run -it --rm debug --image=curlimages/curl -n es-serverless -- \
  curl http://elasticsearch:9200

# 检查 NetworkPolicy
kubectl get networkpolicy -n <namespace>
```

## 性能调优速查

### Elasticsearch JVM

```yaml
# helm/elasticsearch/values.yaml
env:
  - name: ES_JAVA_OPTS
    value: "-Xms4g -Xmx4g"  # 设为容器内存的 50%
```

### 资源限制建议

| 组件 | CPU (requests/limits) | 内存 (requests/limits) |
|------|----------------------|----------------------|
| ES (开发) | 1/2 | 2Gi/4Gi |
| ES (生产) | 4/8 | 8Gi/16Gi |
| Manager | 500m/1 | 512Mi/1Gi |
| Prometheus | 500m/1 | 1Gi/2Gi |
| Grafana | 200m/500m | 256Mi/512Mi |

### 存储性能

```bash
# 使用 SSD StorageClass (生产环境)
# GKE: pd-ssd
# EKS: gp3
# AKS: managed-premium
```

## 监控指标速查

### Prometheus 查询

```promql
# Elasticsearch 堆内存使用率
es_jvm_mem_heap_used_percent

# Pod CPU 使用率
rate(container_cpu_usage_seconds_total[5m])

# Pod 内存使用
container_memory_working_set_bytes

# 磁盘使用率
(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100

# QPS
rate(http_requests_total[1m])
```

### Grafana Dashboard IDs

- Elasticsearch: 2322
- Kubernetes Cluster: 7249
- Node Exporter: 1860

## 备份恢复速查

### Elasticsearch 快照

```bash
# 创建快照仓库
curl -X PUT "localhost:9200/_snapshot/backup" \
  -H 'Content-Type: application/json' \
  -d '{"type": "fs", "settings": {"location": "/backups"}}'

# 创建快照
curl -X PUT "localhost:9200/_snapshot/backup/snapshot_1?wait_for_completion=true"

# 查看快照
curl "localhost:9200/_snapshot/backup/_all?pretty"

# 恢复快照
curl -X POST "localhost:9200/_snapshot/backup/snapshot_1/_restore"
```

### Terraform 状态备份

```bash
# 备份
make backup-state

# 或手动
cp terraform/terraform.tfstate \
   backups/terraform.tfstate.$(date +%Y%m%d)
```

## 版本信息

- **Elasticsearch**: 8.11.0
- **Prometheus**: 2.47.0
- **Grafana**: 10.1.0
- **Terraform**: >= 1.0
- **Helm**: >= 3.0

## 文档链接

- [完整部署指南](docs/terraform-helm-guide.md)
- [Helm Charts 参考](docs/helm-charts-reference.md)
- [架构图](docs/terraform-architecture-diagram.md)
- [主 README](TERRAFORM_HELM_README.md)

## 获取帮助

```bash
# 显示所有 make 命令
make help

# 检查环境
make test-cluster

# 查看文档
make docs
```

---

💡 **提示**: 将此文件加入书签,随时查阅!
