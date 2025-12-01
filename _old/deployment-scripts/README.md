# 纯脚本部署方式

这个目录包含使用 **Shell 脚本 + Kubernetes YAML** 的轻量级部署方式。

## 📋 概述

这种部署方式使用：
- **Shell 脚本**: 自动化部署流程
- **Kubernetes YAML**: 直接定义资源配置
- **kubectl**: 直接操作 Kubernetes API

## 🎯 架构优势

✅ **简单直接**: 无需学习 Terraform/Helm，Shell 脚本即可
✅ **快速上手**: 适合开发测试和快速验证
✅ **灵活调试**: 容易修改和调试
✅ **轻量级**: 无需额外工具，只需 kubectl
✅ **透明**: 直接查看和修改 YAML 配置

## 📁 目录结构

```
deployment-scripts/
├── README.md                 # 本文档
├── scripts/                  # 部署脚本
│   ├── deploy.sh            # 主部署脚本
│   ├── cluster.sh           # 集群管理脚本
│   ├── create-tenant.sh     # 创建租户脚本
│   ├── build-plugin.sh      # 构建 IVF 插件
│   ├── monitor.sh           # 监控脚本
│   ├── shard-management.sh  # 分片管理脚本
│   ├── backup-*.sh          # 备份脚本
│   └── test-ivf.sh          # IVF 测试脚本
│
├── k8s/                      # Kubernetes 配置
│   ├── base/                # 基础配置
│   │   ├── namespace.yaml
│   │   ├── elasticsearch.yaml
│   │   ├── manager.yaml
│   │   ├── shard-controller.yaml
│   │   └── reporting.yaml
│   │
│   └── overlays/            # 环境配置
│       ├── dev/             # 开发环境
│       └── prod/            # 生产环境
│
└── config/                  # 配置文件
    ├── elasticsearch.yml
    └── ivf-config.json
```

## 🚀 快速开始

### 前提条件

```bash
# 1. 确保已安装 kubectl
kubectl version --client

# 2. 配置 Kubernetes 上下文
kubectl config use-context docker-desktop

# 3. 验证连接
kubectl cluster-info
```

### 一键部署

```bash
cd deployment-scripts

# 完整安装
./scripts/deploy.sh install

# 查看状态
./scripts/deploy.sh status

# 卸载
./scripts/deploy.sh uninstall
```

## 📋 脚本说明

### 1. deploy.sh - 主部署脚本

**功能**: 一键部署整个 ES Serverless 系统

```bash
# 查看帮助
./scripts/deploy.sh help

# 部署到自定义命名空间
NAMESPACE=my-es ./scripts/deploy.sh install

# 查看状态
./scripts/deploy.sh status

# 完全卸载
./scripts/deploy.sh uninstall
```

**部署流程**:
1. 创建命名空间 `es-serverless`
2. 部署 Elasticsearch StatefulSet (3副本)
3. 部署 Manager 服务
4. 部署 ShardController
5. 部署 Reporting Service
6. 等待所有服务就绪
7. 显示访问方式

### 2. cluster.sh - 集群管理脚本

**功能**: 管理 Elasticsearch 集群生命周期

```bash
# 创建集群（3副本）
./scripts/cluster.sh create my-namespace 3

# 查看集群状态
./scripts/cluster.sh status my-namespace

# 扩容集群到5副本
./scripts/cluster.sh scale my-namespace 5

# 删除集群
./scripts/cluster.sh delete my-namespace
```

**实现原理**:
- 使用 kubectl 动态创建 StatefulSet
- 自动配置 Service 和 Headless Service
- 创建 PVC 持久化存储
- 等待 Pod 就绪

### 3. create-tenant.sh - 租户创建脚本

**功能**: 为新用户创建隔离的 ES 集群

```bash
# 创建租户
./scripts/create-tenant.sh \
  --tenant-org-id org-001 \
  --user alice \
  --service vector-search \
  --replicas 3 \
  --cpu 2000m \
  --memory 4Gi \
  --storage 50Gi

# 查看租户资源
kubectl get all -n org-001-alice-vector-search
```

**自动创建**:
- ✅ Namespace: `org-001-alice-vector-search`
- ✅ ResourceQuota: CPU/内存/存储限制
- ✅ NetworkPolicy: 租户间网络隔离
- ✅ Elasticsearch StatefulSet
- ✅ Service (ClusterIP + Headless)
- ✅ PVC 持久化卷

### 4. build-plugin.sh - IVF 插件构建

**功能**: 编译 ES IVF 向量搜索插件

```bash
# 构建插件
./scripts/build-plugin.sh

# 输出: es-plugin/build/distributions/es-ivf-plugin-*.zip
```

**构建流程**:
1. 进入 `es-plugin/` 目录
2. 执行 `gradle build`
3. 生成插件 ZIP 包
4. 可选：上传到 ConfigMap 或镜像仓库

### 5. monitor.sh - 监控脚本

**功能**: 实时监控集群健康状态

```bash
# 启动监控
./scripts/monitor.sh

# 输出示例:
# === ES Cluster Health ===
# Status: green
# Nodes: 3
# Indices: 5
# Shards: 15 (primary: 5, replica: 10)
#
# === Pod Status ===
# elasticsearch-0  Running  1/1
# elasticsearch-1  Running  1/1
# elasticsearch-2  Running  1/1
```

**监控指标**:
- Elasticsearch 集群健康
- Pod 运行状态
- 分片分布情况
- CPU/内存使用率

### 6. shard-management.sh - 分片管理

**功能**: 手动触发分片重平衡和优化

```bash
# 查看分片分布
./scripts/shard-management.sh status

# 触发重平衡
./scripts/shard-management.sh rebalance

# 优化热点分片
./scripts/shard-management.sh optimize
```

### 7. backup-*.sh - 备份脚本

#### backup-es-snapshot.sh - ES 快照备份

```bash
# 创建快照仓库
./scripts/backup-es-snapshot.sh register-repo s3://my-bucket/snapshots

# 创建快照
./scripts/backup-es-snapshot.sh create my-snapshot-1

# 查看快照
./scripts/backup-es-snapshot.sh list

# 恢复快照
./scripts/restore-from-snapshot.sh my-snapshot-1
```

#### backup-metadata.sh - 元数据备份

```bash
# 备份控制平面元数据
./scripts/backup-metadata.sh backup

# 恢复元数据
./scripts/backup-metadata.sh restore backup-2024-12-01.tar.gz
```

### 8. test-ivf.sh - IVF 功能测试

**功能**: 测试向量搜索功能

```bash
# 运行完整测试
./scripts/test-ivf.sh

# 测试流程:
# 1. 创建 IVF 索引
# 2. 插入测试向量
# 3. 执行向量搜索
# 4. 验证结果准确性
# 5. 性能基准测试
```

## 🔧 常用操作

### 扩容 Elasticsearch

```bash
# 方式1: 使用 cluster.sh
./scripts/cluster.sh scale es-serverless 5

# 方式2: 直接修改 StatefulSet
kubectl -n es-serverless scale sts/elasticsearch --replicas=5

# 验证
kubectl -n es-serverless get pods -l app=elasticsearch
```

### 创建新租户

```bash
./scripts/create-tenant.sh \
  --tenant-org-id org-002 \
  --user bob \
  --service text-search \
  --replicas 2 \
  --cpu 1000m \
  --memory 2Gi

# 查看新租户
kubectl get ns | grep org-002
kubectl get all -n org-002-bob-text-search
```

### 访问服务

```bash
# Elasticsearch
kubectl -n es-serverless port-forward svc/elasticsearch 9200:9200

# Manager API
kubectl -n es-serverless port-forward svc/es-serverless-manager 8080:8080

# Kibana (如果部署)
kubectl -n es-serverless port-forward svc/kibana 5601:5601
```

### 查看日志

```bash
# Elasticsearch 日志
kubectl -n es-serverless logs -l app=elasticsearch --tail=100 -f

# Manager 日志
kubectl -n es-serverless logs -l app=es-serverless-manager -f

# 所有容器日志
kubectl -n es-serverless logs --all-containers=true -l app=elasticsearch
```

## 📊 Kubernetes 配置说明

### base/ - 基础配置

#### namespace.yaml
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: es-serverless
  labels:
    name: es-serverless
```

#### elasticsearch.yaml
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
spec:
  serviceName: elasticsearch-headless
  replicas: 3
  selector:
    matchLabels:
      app: elasticsearch
  template:
    spec:
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
        ports:
        - containerPort: 9200
          name: http
        - containerPort: 9300
          name: transport
        env:
        - name: cluster.name
          value: "es-serverless"
        - name: discovery.seed_hosts
          value: "elasticsearch-headless"
        volumeMounts:
        - name: data
          mountPath: /usr/share/elasticsearch/data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

### overlays/ - 环境覆盖

#### dev/ - 开发环境
- 单副本
- 小资源配额
- 不启用持久化（可选）

#### prod/ - 生产环境
- 3副本
- 高资源配额
- 启用持久化和备份

## 🔄 升级流程

### 滚动升级 Elasticsearch

```bash
# 1. 修改镜像版本
kubectl -n es-serverless set image sts/elasticsearch \
  elasticsearch=docker.elastic.co/elasticsearch/elasticsearch:8.12.0

# 2. 监控升级进度
kubectl -n es-serverless rollout status sts/elasticsearch

# 3. 验证集群健康
kubectl -n es-serverless exec elasticsearch-0 -- \
  curl -s http://localhost:9200/_cluster/health?pretty
```

### 升级控制平面

```bash
# 1. 重新构建镜像
cd server
docker build -t es-serverless-manager:v2.0 .

# 2. 更新 Deployment
kubectl -n es-serverless set image deployment/es-serverless-manager \
  manager=es-serverless-manager:v2.0

# 3. 监控部署
kubectl -n es-serverless rollout status deployment/es-serverless-manager
```

## 🗑️ 清理资源

### 删除特定租户

```bash
# 删除租户命名空间（包含所有资源）
kubectl delete ns org-001-alice-vector-search

# 确认删除
kubectl get ns | grep org-001
```

### 完全清理

```bash
# 方式1: 使用脚本
./scripts/deploy.sh uninstall

# 方式2: 手动删除
kubectl delete ns es-serverless

# 清理 PVC（持久化卷）
kubectl get pvc -A | grep elasticsearch
kubectl delete pvc -n es-serverless data-elasticsearch-0
```

## 🔐 安全配置

### 网络隔离

手动创建 NetworkPolicy：

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tenant-isolation
  namespace: org-001-alice-vector-search
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector: {}  # 只允许命名空间内部通信
  egress:
  - to:
    - podSelector: {}
  - to:  # 允许 DNS
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - port: 53
      protocol: UDP
```

```bash
kubectl apply -f network-policy.yaml
```

### 资源配额

手动创建 ResourceQuota：

```yaml
# resource-quota.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-quota
  namespace: org-001-alice-vector-search
spec:
  hard:
    requests.cpu: "10"
    requests.memory: "20Gi"
    requests.storage: "100Gi"
    persistentvolumeclaims: "10"
    pods: "50"
```

```bash
kubectl apply -f resource-quota.yaml
```

## 📚 脚本示例

### 自定义部署脚本

```bash
#!/usr/bin/env bash
# my-deploy.sh - 自定义部署流程

set -euo pipefail

NAMESPACE="my-es"

# 1. 创建命名空间
kubectl create namespace $NAMESPACE || true

# 2. 部署 Elasticsearch
kubectl apply -f k8s/base/elasticsearch.yaml -n $NAMESPACE

# 3. 等待就绪
kubectl -n $NAMESPACE rollout status sts/elasticsearch --timeout=300s

# 4. 验证集群健康
kubectl -n $NAMESPACE exec elasticsearch-0 -- \
  curl -s http://localhost:9200/_cluster/health?pretty

echo "Deployment completed!"
```

## ❓ 常见问题

### Q: Pod 一直处于 Pending 状态？

```bash
# 查看原因
kubectl -n es-serverless describe pod elasticsearch-0

# 常见原因:
# 1. 资源不足 -> 减少 replicas 或增加节点
# 2. PVC 无法绑定 -> 检查 StorageClass
# 3. 镜像拉取失败 -> 检查网络或使用本地镜像
```

### Q: 如何查看 Elasticsearch 集群状态？

```bash
# 集群健康
kubectl -n es-serverless exec elasticsearch-0 -- \
  curl -s http://localhost:9200/_cluster/health?pretty

# 节点信息
kubectl -n es-serverless exec elasticsearch-0 -- \
  curl -s http://localhost:9200/_cat/nodes?v

# 索引信息
kubectl -n es-serverless exec elasticsearch-0 -- \
  curl -s http://localhost:9200/_cat/indices?v
```

### Q: 如何重启 Elasticsearch Pod？

```bash
# 滚动重启 StatefulSet
kubectl -n es-serverless rollout restart sts/elasticsearch

# 删除特定 Pod（自动重建）
kubectl -n es-serverless delete pod elasticsearch-0

# 等待就绪
kubectl -n es-serverless wait --for=condition=ready pod -l app=elasticsearch --timeout=300s
```

### Q: 数据持久化配置？

```bash
# 查看 PVC
kubectl -n es-serverless get pvc

# 查看 PV
kubectl get pv

# 查看 StorageClass
kubectl get sc

# 修改 StorageClass（如果使用 hostpath）
kubectl patch storageclass hostpath \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

## 🆚 对比：纯脚本 vs Terraform+Helm

| 特性 | 纯脚本 | Terraform+Helm |
|------|--------|---------------|
| 学习曲线 | ✅ 低（Shell） | ⚠️ 中等 |
| 部署速度 | ✅ 快 | ⚠️ 较慢 |
| 配置复杂度 | ✅ 简单 | ⚠️ 复杂 |
| 适用场景 | 开发/测试 | 生产环境 |
| 状态管理 | ❌ 手动 | ✅ 自动 |
| 回滚能力 | ❌ 困难 | ✅ 简单 |
| 模块化 | ⚠️ 有限 | ✅ 强大 |
| 多租户管理 | ⚠️ 需手动 | ✅ 自动化 |
| 审计追踪 | ❌ 无 | ✅ 有 |

**推荐使用场景**:
- **纯脚本**: 本地开发、功能测试、快速验证POC
- **Terraform+Helm**: 生产部署、多租户环境、企业级应用

## 📖 脚本开发指南

### 编写新脚本的最佳实践

```bash
#!/usr/bin/env bash
set -euo pipefail  # 错误时退出，未定义变量报错，管道错误传播

# 变量定义
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE=${NAMESPACE:-es-serverless}

# 帮助信息
show_help() {
    cat <<EOF
Usage: $0 [OPTIONS]

Options:
  -h, --help     Show this help
  -n NAMESPACE   Kubernetes namespace (default: es-serverless)

Examples:
  $0 -n my-namespace
EOF
}

# 主函数
main() {
    echo "Starting deployment..."
    # 实现逻辑
}

# 调用主函数
main "$@"
```

---

**下一步**: 参考 [../deployment-terraform/README.md](../deployment-terraform/README.md) 了解 Terraform+Helm 部署方式
