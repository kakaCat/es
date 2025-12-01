# Helm Charts 参考文档

本文档详细说明 ES Serverless 平台提供的 Helm Charts 配置选项。

## 目录

- [Elasticsearch Chart](#elasticsearch-chart)
- [Control Plane Chart](#control-plane-chart)
- [Monitoring Chart](#monitoring-chart)

## Elasticsearch Chart

位置: `helm/elasticsearch/`

### 基本信息

- **Chart 名称**: elasticsearch
- **Chart 版本**: 1.0.0
- **应用版本**: 8.11.0

### 配置选项

#### 副本和镜像

```yaml
# 副本数量
replicaCount: 3

# Elasticsearch 镜像
image:
  repository: docker.elastic.co/elasticsearch/elasticsearch
  tag: "8.11.0"
  pullPolicy: IfNotPresent
```

#### IVF 插件配置

```yaml
ivfPlugin:
  enabled: true
  image:
    repository: es-ivf-plugin
    tag: latest
  config:
    dimension: 128      # 向量维度
    vectorCount: 1000000  # 向量数量
    nlist: 100          # 倒排簇数
    nprobe: 10          # 搜索簇数
```

#### 集群配置

```yaml
clusterName: elasticsearch

# Elasticsearch 配置文件
esConfig:
  elasticsearch.yml: |
    cluster.name: ${CLUSTER_NAME}
    network.host: 0.0.0.0
    discovery.seed_hosts: elasticsearch-headless
    cluster.initial_master_nodes: elasticsearch-0,elasticsearch-1,elasticsearch-2
    xpack.security.enabled: false

    # IVF 插件设置
    ivf.enabled: true
    ivf.nlist: 100
    ivf.nprobe: 10
```

#### 资源配置

```yaml
resources:
  requests:
    cpu: 1000m
    memory: 2Gi
  limits:
    cpu: 2000m
    memory: 4Gi
```

#### 持久化存储

```yaml
persistence:
  enabled: true
  storageClass: "hostpath"
  size: 10Gi
  accessModes:
    - ReadWriteOnce
```

#### JVM 配置

```yaml
env:
  - name: ES_JAVA_OPTS
    value: "-Xms1g -Xmx1g"
  - name: CLUSTER_NAME
    value: "elasticsearch"
```

#### 服务配置

```yaml
service:
  type: ClusterIP
  port: 9200
  transportPort: 9300

headlessService:
  enabled: true
```

#### 探针配置

```yaml
livenessProbe:
  httpGet:
    path: /_cluster/health?local=true
    port: 9200
  initialDelaySeconds: 90
  periodSeconds: 10
  timeoutSeconds: 5

readinessProbe:
  httpGet:
    path: /_cluster/health?local=true
    port: 9200
  initialDelaySeconds: 60
  periodSeconds: 10
  timeoutSeconds: 5
```

#### 监控集成

```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
```

### 使用示例

#### 基本部署

```bash
helm install elasticsearch ./helm/elasticsearch \
  --namespace es-serverless \
  --create-namespace
```

#### 自定义配置

```bash
# 创建 values.yaml
cat > custom-values.yaml <<EOF
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
EOF

# 使用自定义配置部署
helm install elasticsearch ./helm/elasticsearch \
  -f custom-values.yaml \
  --namespace es-serverless
```

#### 升级配置

```bash
helm upgrade elasticsearch ./helm/elasticsearch \
  -f custom-values.yaml \
  --namespace es-serverless
```

## Control Plane Chart

位置: `helm/control-plane/`

### 基本信息

- **Chart 名称**: es-control-plane
- **Chart 版本**: 1.0.0
- **应用版本**: 1.0.0

### 配置选项

#### Manager 服务

```yaml
manager:
  enabled: true
  replicaCount: 1

  image:
    repository: es-serverless-manager
    tag: latest
    pullPolicy: IfNotPresent

  service:
    type: ClusterIP
    port: 8080

  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1Gi

  env:
    - name: ELASTICSEARCH_URL
      value: "http://elasticsearch:9200"
    - name: USE_POSTGRES
      value: "false"
    - name: DATA_DIR
      value: "/data"

  persistence:
    enabled: true
    storageClass: "hostpath"
    size: 5Gi
    mountPath: /data
```

#### Shard Controller

```yaml
shardController:
  enabled: true
  replicaCount: 1

  image:
    repository: shard-controller
    tag: latest
    pullPolicy: IfNotPresent

  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi

  env:
    - name: ELASTICSEARCH_URL
      value: "http://elasticsearch:9200"
    - name: CHECK_INTERVAL
      value: "30s"
```

#### Reporting Service

```yaml
reportingService:
  enabled: true
  replicaCount: 1

  image:
    repository: reporting-service
    tag: latest
    pullPolicy: IfNotPresent

  service:
    type: ClusterIP
    port: 8081

  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi

  env:
    - name: MANAGER_URL
      value: "http://es-serverless-manager:8080"
    - name: REPORT_INTERVAL
      value: "60s"
```

#### RBAC 配置

```yaml
serviceAccount:
  create: true
  name: es-control-plane

rbac:
  create: true
  rules:
    - apiGroups: [""]
      resources: ["namespaces", "pods", "services", "persistentvolumeclaims"]
      verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
    - apiGroups: ["apps"]
      resources: ["statefulsets", "deployments"]
      verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

### 使用示例

```bash
helm install es-control-plane ./helm/control-plane \
  --namespace es-serverless \
  --set manager.image.tag=v1.0.0 \
  --set shardController.image.tag=v1.0.0 \
  --set reportingService.image.tag=v1.0.0
```

## Monitoring Chart

位置: `helm/monitoring/`

### 基本信息

- **Chart 名称**: es-monitoring
- **Chart 版本**: 1.0.0

### 配置选项

#### Prometheus

```yaml
prometheus:
  enabled: true
  replicaCount: 1

  image:
    repository: prom/prometheus
    tag: v2.47.0
    pullPolicy: IfNotPresent

  service:
    type: ClusterIP
    port: 9090

  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1000m
      memory: 2Gi

  persistence:
    enabled: true
    storageClass: "hostpath"
    size: 10Gi

  retention:
    days: 15

  scrapeInterval: 30s
  evaluationInterval: 30s

  # 采集配置
  scrapeConfigs:
    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
        - role: pod
      relabel_configs:
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
          action: keep
          regex: true

    - job_name: 'elasticsearch'
      static_configs:
        - targets: ['elasticsearch:9200']
```

#### Grafana

```yaml
grafana:
  enabled: true
  replicaCount: 1

  image:
    repository: grafana/grafana
    tag: 10.1.0
    pullPolicy: IfNotPresent

  service:
    type: ClusterIP
    port: 3000

  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi

  persistence:
    enabled: true
    storageClass: "hostpath"
    size: 5Gi

  adminUser: admin
  adminPassword: admin

  # 数据源配置
  datasources:
    - name: Prometheus
      type: prometheus
      url: http://monitoring-prometheus:9090
      isDefault: true
      access: proxy
```

### 使用示例

```bash
# 启用所有监控组件
helm install monitoring ./helm/monitoring \
  --namespace es-serverless

# 仅启用 Prometheus
helm install monitoring ./helm/monitoring \
  --namespace es-serverless \
  --set grafana.enabled=false

# 自定义保留期
helm install monitoring ./helm/monitoring \
  --namespace es-serverless \
  --set prometheus.retention.days=30
```

## Chart 维护

### 打包 Chart

```bash
# 打包单个 Chart
helm package helm/elasticsearch
helm package helm/control-plane
helm package helm/monitoring

# 输出: elasticsearch-1.0.0.tgz
```

### 验证 Chart

```bash
# 语法检查
helm lint helm/elasticsearch

# 模板渲染测试
helm template elasticsearch helm/elasticsearch \
  --namespace es-serverless \
  --debug
```

### 发布 Chart

```bash
# 创建 Chart 仓库索引
helm repo index . --url https://your-repo.com/charts

# 上传到 Chart 仓库 (示例)
# 可以使用 ChartMuseum, Harbor, 或 GitHub Pages
```

## 常见问题

### 如何更新镜像版本?

```bash
helm upgrade elasticsearch ./helm/elasticsearch \
  --set image.tag=8.12.0 \
  --namespace es-serverless
```

### 如何增加副本数?

```bash
helm upgrade elasticsearch ./helm/elasticsearch \
  --set replicaCount=5 \
  --namespace es-serverless
```

### 如何修改资源限制?

```yaml
# custom-values.yaml
resources:
  limits:
    cpu: 4000m
    memory: 8Gi
```

```bash
helm upgrade elasticsearch ./helm/elasticsearch \
  -f custom-values.yaml \
  --namespace es-serverless
```

### 如何禁用某个组件?

```bash
# 禁用 Grafana
helm upgrade monitoring ./helm/monitoring \
  --set grafana.enabled=false \
  --namespace es-serverless
```

## 下一步

- 查看 [Terraform 和 Helm 部署指南](terraform-helm-guide.md)
- 了解 [多租户架构](多租户架构说明.md)
- 探索 [API 文档](../README.md)
