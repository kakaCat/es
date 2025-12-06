# Kubernetes 容器详解

## 什么是 k8s_ 开头的容器?

所有以 `k8s_` 开头的容器都是 **Kubernetes 创建和管理的容器**。

### 容器命名规则

```
k8s_<容器名>_<Pod名>_<命名空间>_<Pod UID>_<重启次数>
```

**示例**:
```
k8s_coredns_coredns-66bc5c9577-bhhps_kube-system_4030aff8-e40a-494e-9868-238cb2e3f3b7_2
│   │       │                        │           │                                    │
│   │       │                        │           └─ Pod 唯一 ID                       └─ 重启次数 (2次)
│   │       │                        └─ 命名空间 (kube-system)
│   │       └─ Pod 名称
│   └─ 容器名称
└─ Kubernetes 前缀
```

## 两类容器

### 1. POD 容器 (Pause 容器)

**名称特征**: `k8s_POD_...`

**作用**:
- 每个 Pod 的 **基础容器** (Infrastructure Container)
- 使用 `pause` 镜像 (`registry.k8s.io/pause:3.10`)
- **持有网络命名空间**,提供网络共享
- 为 Pod 中的其他容器提供共享的网络栈

**为什么需要它**:
```
Pod (逻辑概念)
  ├─ POD 容器 (pause) ← 网络命名空间的持有者
  ├─ 应用容器 1       ← 共享 POD 容器的网络
  └─ 应用容器 2       ← 共享 POD 容器的网络
```

**示例**:
```bash
# POD 容器
k8s_POD_coredns-66bc5c9577-bhhps_...
  └─ 镜像: registry.k8s.io/pause:3.10
  └─ 作用: 为 coredns Pod 提供网络命名空间
```

### 2. 应用容器 (实际工作的容器)

**名称特征**: `k8s_<服务名>_...`

**作用**:
- 运行实际的应用程序
- 共享对应 POD 容器的网络

**示例**:
```bash
# 应用容器
k8s_coredns_coredns-66bc5c9577-bhhps_...
  └─ 镜像: coredns 镜像
  └─ 作用: 运行 DNS 服务
  └─ 网络: 共享 POD 容器的网络命名空间
```

## 当前系统中的 Kubernetes 容器

### Kubernetes 核心组件 (不可删除)

#### 1. **kube-apiserver**
```
k8s_kube-apiserver_kube-apiserver-docker-desktop_kube-system_...
```
- **作用**: Kubernetes API 服务器
- **功能**:
  - 所有 kubectl 命令的入口
  - 集群的控制中心
  - 处理 REST 请求
  - 验证和配置数据

#### 2. **etcd**
```
k8s_etcd_etcd-docker-desktop_kube-system_...
```
- **作用**: 分布式键值存储
- **功能**:
  - 存储集群的所有配置数据
  - 存储集群状态
  - Kubernetes 的"数据库"

#### 3. **kube-controller-manager**
```
k8s_kube-controller-manager_kube-controller-manager-docker-desktop_...
```
- **作用**: 控制器管理器
- **功能**:
  - 节点控制器 (监控节点状态)
  - 副本控制器 (维护 Pod 副本数)
  - 端点控制器 (填充 Endpoints 对象)
  - 服务账户控制器

#### 4. **kube-scheduler**
```
k8s_kube-scheduler_kube-scheduler-docker-desktop_...
```
- **作用**: 调度器
- **功能**:
  - 监听新创建的 Pod
  - 为 Pod 选择合适的节点
  - 考虑资源需求、硬件约束等

#### 5. **kube-proxy**
```
k8s_kube-proxy_kube-proxy-95rsb_...
```
- **作用**: 网络代理
- **功能**:
  - 维护节点上的网络规则
  - 实现 Kubernetes Service 概念
  - 负责流量转发

#### 6. **coredns**
```
k8s_coredns_coredns-66bc5c9577-bhhps_...
k8s_coredns_coredns-66bc5c9577-8s8xx_...
```
- **作用**: DNS 服务器
- **功能**:
  - 提供集群内部 DNS 解析
  - 解析 Service 名称到 ClusterIP
  - 示例: `elasticsearch.es-serverless.svc.cluster.local`

#### 7. **storage-provisioner**
```
k8s_storage-provisioner_storage-provisioner_kube-system_...
```
- **作用**: 存储供应器 (Docker Desktop 特有)
- **功能**:
  - 自动创建 PersistentVolume
  - 处理 PVC 请求
  - 提供 hostpath 存储类

#### 8. **vpnkit-controller**
```
k8s_vpnkit-controller_vpnkit-controller_kube-system_...
```
- **作用**: VPN 控制器 (Docker Desktop 特有)
- **功能**:
  - Docker Desktop 网络桥接
  - 连接容器网络和主机网络

### 之前的应用容器 (已删除)

这些是你刚才删除的 ES Serverless 应用:

#### Elasticsearch
```
k8s_elasticsearch_elasticsearch-0_es-serverless_...
  └─ 运行 Elasticsearch 数据库
```

#### Grafana
```
k8s_grafana_grafana-xxx_es-serverless_...
  └─ 运行 Grafana 可视化
```

#### Kibana
```
k8s_kibana_kibana-xxx_es-serverless_...
  └─ 运行 Kibana 界面
```

#### Prometheus
```
k8s_prometheus_prometheus-xxx_es-serverless_...
  └─ 运行 Prometheus 监控
```

#### Manager
```
k8s_manager_es-serverless-manager-xxx_es-serverless_...
  └─ 运行 ES Serverless Manager API
```

## 容器之间的关系

### Pod 结构示例

```
Pod: coredns-66bc5c9577-bhhps
├─ k8s_POD_coredns-...              (pause 容器)
│  └─ 作用: 持有网络命名空间
│  └─ IP: 10.1.0.5
│
└─ k8s_coredns_coredns-...          (应用容器)
   └─ 作用: 运行 CoreDNS
   └─ 网络: 共享 POD 容器的 IP (10.1.0.5)
```

### Kubernetes 架构图

```
┌─────────────────────────────────────────────────┐
│  Control Plane (控制平面)                        │
│  ┌──────────────┐  ┌──────────────┐            │
│  │ kube-apiserver│  │     etcd     │            │
│  │              │  │  (数据存储)   │            │
│  └──────────────┘  └──────────────┘            │
│  ┌──────────────┐  ┌──────────────┐            │
│  │kube-scheduler│  │kube-controller│            │
│  │   (调度器)    │  │   -manager   │            │
│  └──────────────┘  └──────────────┘            │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│  Worker Node (工作节点 - Docker Desktop)         │
│  ┌──────────────┐  ┌──────────────┐            │
│  │  kube-proxy  │  │   coredns    │            │
│  │  (网络代理)   │  │  (DNS服务)   │            │
│  └──────────────┘  └──────────────┘            │
│  ┌──────────────┐                              │
│  │storage-      │                              │
│  │provisioner   │                              │
│  └──────────────┘                              │
│                                                 │
│  应用 Pods (已删除)                              │
│  ┌────────────────────────────────────────┐    │
│  │ elasticsearch, grafana, kibana, etc.   │    │
│  └────────────────────────────────────────┘    │
└─────────────────────────────────────────────────┘
```

## 为什么不能直接删除?

### Kubernetes 的自愈机制

```bash
# 场景 1: 删除 Pod
kubectl delete pod coredns-xxx
  → Deployment 检测到 Pod 丢失
  → 自动创建新的 Pod
  → 新容器启动

# 场景 2: 删除 Docker 容器
docker rm -f k8s_coredns_...
  → Kubelet 检测到容器丢失
  → 自动重启容器
  → 容器恢复运行
```

### 正确的删除顺序

```
1. Kubernetes 资源 (最高层)
   kubectl delete deployment/statefulset/pod

2. Kubernetes 自动停止容器

3. Docker 容器自动清理

❌ 不要跳过第1步直接删除 Docker 容器!
```

## 如何查看容器详情

### 查看容器所属的 Pod

```bash
# 从容器名提取信息
容器名: k8s_coredns_coredns-66bc5c9577-bhhps_kube-system_...
         │      │                        │
         │      └─ Pod 名                └─ 命名空间
         └─ 容器名

# 查看对应的 Pod
kubectl get pod coredns-66bc5c9577-bhhps -n kube-system
```

### 查看容器日志

```bash
# 方式 1: 通过 kubectl (推荐)
kubectl logs coredns-66bc5c9577-bhhps -n kube-system

# 方式 2: 通过 docker
docker logs k8s_coredns_coredns-66bc5c9577-bhhps_kube-system_...
```

### 进入容器

```bash
# 方式 1: 通过 kubectl (推荐)
kubectl exec -it coredns-66bc5c9577-bhhps -n kube-system -- sh

# 方式 2: 通过 docker
docker exec -it k8s_coredns_coredns-66bc5c9577-bhhps_... sh
```

## 容器资源使用

### 查看容器资源

```bash
# 查看 Pod 资源使用
kubectl top pods -n kube-system

# 查看 Docker 容器资源
docker stats
```

### 限制容器资源

通过 Kubernetes 配置:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

## 常见问题

### Q1: 为什么有这么多 POD 容器?

**A**: 每个 Pod 都有一个 POD (pause) 容器,用于:
- 持有网络命名空间
- 作为 Pod 中所有容器的网络基础
- 即使应用容器重启,网络配置也保持不变

### Q2: 可以删除 POD 容器吗?

**A**: 不建议!如果删除:
- Pod 中的所有容器会失去网络连接
- Kubernetes 会重新创建整个 Pod

### Q3: kube-system 的容器可以删除吗?

**A**: **绝对不要删除!** 这些是 Kubernetes 核心组件:
- 删除后集群会崩溃
- kubectl 命令会失效
- 所有应用会停止

### Q4: 容器一直重启怎么办?

```bash
# 查看容器日志
kubectl logs <pod-name> -n <namespace>

# 查看 Pod 事件
kubectl describe pod <pod-name> -n <namespace>

# 查看 Pod 状态
kubectl get pod <pod-name> -n <namespace> -o yaml
```

## 总结

### 容器类型

| 类型 | 前缀 | 作用 | 可删除 |
|------|------|------|--------|
| Kubernetes 核心 | `k8s_kube-*` | 集群控制 | ❌ 不可 |
| DNS 服务 | `k8s_coredns_` | 内部 DNS | ❌ 不可 |
| Pause 容器 | `k8s_POD_` | 网络基础 | ❌ 不可 |
| 应用容器 | `k8s_<app>_` | 你的应用 | ✅ 可以 (通过 kubectl) |

### 管理建议

1. **不要直接操作 Docker 容器** - 使用 kubectl
2. **不要删除 kube-system 容器** - 这是核心组件
3. **删除应用使用 kubectl delete** - 让 Kubernetes 管理生命周期
4. **定期清理已停止的容器** - `docker container prune`

### 有用的命令

```bash
# 查看所有容器
docker ps -a | grep k8s

# 查看容器和对应的 Pod
kubectl get pods -A -o wide

# 清理已停止的容器
docker container prune -f

# 查看容器详细信息
docker inspect k8s_<container-name>
```
