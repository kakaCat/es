# ES Serverless 多租户架构说明

## 概述

ES Serverless 平台支持完整的多租户架构，通过 **租户组织ID（Tenant Org ID）** 实现租户间的资源隔离和管理。

## 多租户隔离机制

### 1. 租户组织ID（Tenant Org ID）

**作用**：
- 作为顶层租户标识，实现组织级别的资源隔离
- 支持一个组织下有多个用户和服务
- 用于资源配额管理和计费

**格式建议**：
- 字符串类型，建议使用UUID或组织编码
- 例如：`org-001`, `company-abc`, `uuid-xxx-xxx`

### 2. 命名空间（Namespace）隔离

**自动生成规则**：

如果不指定namespace参数，系统会自动生成：

```
{tenant_org_id}-{user}-{service_name}
```

**示例**：
```
tenant_org_id: org-001
user: alice
service_name: vector-search

自动生成的namespace: org-001-alice-vector-search
```

**手动指定**：

也可以手动指定namespace，但仍需提供tenant_org_id用于管理和追踪。

### 3. Kubernetes标签体系

每个Namespace都会被标记以下标签：

```yaml
labels:
  es-cluster: "true"                    # 标识为ES集群
  tenant-org-id: "org-001"              # 租户组织ID
  user: "alice"                          # 用户ID
  service-name: "vector-search"          # 服务名称
```

**用途**：
- 资源查询和过滤
- 配额管理
- 监控和告警
- 计费统计

## 创建容器组（多租户模式）

### API请求

**端点**：`POST /clusters`

**必需参数**：

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `tenant_org_id` | string | ✅ | 租户组织ID（顶层隔离标识） |
| `user` | string | ✅ | 用户ID |
| `service_name` | string | ✅ | 服务名称 |
| `replicas` | int | ❌ | 副本数（默认1） |
| `cpu_request` | string | ❌ | CPU请求（默认500m） |
| `cpu_limit` | string | ❌ | CPU限制（默认2） |
| `mem_request` | string | ❌ | 内存请求（默认1Gi） |
| `mem_limit` | string | ❌ | 内存限制（默认2Gi） |
| `disk_size` | string | ❌ | 磁盘大小（默认10Gi） |
| `gpu_count` | int | ❌ | GPU数量（默认0） |
| `dimension` | int | ❌ | 向量维度（默认128） |
| `vector_count` | int | ❌ | 向量数量（默认10000） |
| `index_limit` | int | ❌ | 索引限制（默认0） |
| `gitlab_url` | string | ❌ | GitLab配置URL |
| `namespace` | string | ❌ | 自定义命名空间（不提供则自动生成） |

### 请求示例

```bash
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_org_id": "org-001",
    "user": "alice",
    "service_name": "vector-search",
    "replicas": 2,
    "cpu_request": "1",
    "cpu_limit": "4",
    "mem_request": "2Gi",
    "mem_limit": "8Gi",
    "disk_size": "50Gi",
    "gpu_count": 1,
    "dimension": 512,
    "vector_count": 1000000,
    "index_limit": 10
  }'
```

### 创建流程

```
1. 验证必需参数
   ├── tenant_org_id 不能为空
   ├── user 不能为空
   └── service_name 不能为空

2. 检查租户配额
   └── 验证是否超出组织配额限制

3. ⭐ 记录租户元数据（第一步）
   ├── 保存TenantContainer（包含tenant_org_id）
   └── 保存DeploymentStatus（包含tenant_org_id）

4. 生成/验证Namespace
   └── 默认: {tenant_org_id}-{user}-{service_name}

5. 创建Kubernetes资源
   ├── 创建Namespace
   ├── 添加标签（tenant-org-id, user, service-name）
   ├── 创建StatefulSet
   ├── 创建Service
   └── 创建PVC

6. 配置资源限制
   ├── CPU/内存配置
   ├── 磁盘大小
   └── GPU数量（如果>0）

7. 等待Pod就绪

8. 同步到租户容器管理

9. 更新状态为created
```

## 查询接口

### 1. 查询所有租户容器

```bash
curl http://localhost:8080/tenant/containers
```

### 2. 查询特定租户容器

```bash
curl http://localhost:8080/tenant/containers/{user}/{service_name}
```

### 3. 🆕 查询特定组织的所有容器

```bash
curl http://localhost:8080/tenant/containers/org/{tenant_org_id}
```

**示例**：

```bash
# 查询org-001组织的所有容器
curl http://localhost:8080/tenant/containers/org/org-001
```

**返回数据示例**：

```json
[
  {
    "id": "tenant_alice_vector-search_1234567890",
    "tenant_org_id": "org-001",
    "user": "alice",
    "service_name": "vector-search",
    "namespace": "org-001-alice-vector-search",
    "replicas": 2,
    "cpu": "1/4",
    "memory": "2Gi/8Gi",
    "disk": "50Gi",
    "gpu_count": 1,
    "dimension": 512,
    "vector_count": 1000000,
    "status": "created",
    "created_at": "2024-01-01T00:00:00Z",
    "sync_time": "2024-01-01T00:00:00Z"
  },
  {
    "id": "tenant_bob_image-search_1234567891",
    "tenant_org_id": "org-001",
    "user": "bob",
    "service_name": "image-search",
    "namespace": "org-001-bob-image-search",
    "replicas": 1,
    "cpu": "500m/2",
    "memory": "1Gi/2Gi",
    "disk": "10Gi",
    "gpu_count": 0,
    "dimension": 128,
    "vector_count": 100000,
    "status": "created",
    "created_at": "2024-01-01T01:00:00Z",
    "sync_time": "2024-01-01T01:00:00Z"
  }
]
```

## 数据存储结构

### TenantContainer 元数据

**位置**：`server/data/tenant_{user}_{service_name}.json`

**结构**：

```json
{
  "id": "tenant_alice_vector-search_1234567890",
  "tenant_org_id": "org-001",
  "user": "alice",
  "service_name": "vector-search",
  "namespace": "org-001-alice-vector-search",
  "replicas": 2,
  "cpu": "1/4",
  "memory": "2Gi/8Gi",
  "disk": "50Gi",
  "gpu_count": 1,
  "dimension": 512,
  "vector_count": 1000000,
  "status": "created",
  "created_at": "2024-01-01T00:00:00Z",
  "sync_time": "2024-01-01T00:00:00Z"
}
```

### DeploymentStatus 元数据

**位置**：`server/data/deploy_{namespace}.json`

**结构**：

```json
{
  "id": "deploy_org-001-alice-vector-search_1234567890",
  "tenant_org_id": "org-001",
  "namespace": "org-001-alice-vector-search",
  "user": "alice",
  "service_name": "vector-search",
  "status": "created",
  "cpu_usage": 0.0,
  "memory_usage": 0.0,
  "disk_usage": 0.0,
  "qps": 0.0,
  "gpu_count": 1,
  "dimension": 512,
  "vector_count": 1000000,
  "replicas": 2,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "details": {
    "cpu_request": "1",
    "cpu_limit": "4",
    "mem_request": "2Gi",
    "mem_limit": "8Gi",
    "disk_size": "50Gi",
    "gpu_count": 1,
    "dimension": 512,
    "vector_count": 1000000,
    "index_limit": 10
  }
}
```

## 多租户场景示例

### 场景1：同一组织下多个用户

```bash
# 组织org-001下，用户alice创建向量搜索服务
curl -X POST http://localhost:8080/clusters -d '{
  "tenant_org_id": "org-001",
  "user": "alice",
  "service_name": "vector-search",
  ...
}'

# 组织org-001下，用户bob创建图像搜索服务
curl -X POST http://localhost:8080/clusters -d '{
  "tenant_org_id": "org-001",
  "user": "bob",
  "service_name": "image-search",
  ...
}'

# 查询org-001组织的所有容器
curl http://localhost:8080/tenant/containers/org/org-001
```

### 场景2：不同组织完全隔离

```bash
# 组织org-001的服务
curl -X POST http://localhost:8080/clusters -d '{
  "tenant_org_id": "org-001",
  "user": "alice",
  "service_name": "search",
  ...
}'

# 组织org-002的服务（完全隔离）
curl -X POST http://localhost:8080/clusters -d '{
  "tenant_org_id": "org-002",
  "user": "alice",  # 即使用户名相同，也是不同租户
  "service_name": "search",
  ...
}'
```

## 资源隔离保证

### 1. Namespace级别隔离

- 每个服务运行在独立的Kubernetes Namespace
- 通过Namespace实现网络隔离和资源配额
- Namespace命名包含tenant_org_id，确保唯一性

### 2. 元数据级别隔离

- TenantContainer和DeploymentStatus都记录tenant_org_id
- 支持按组织ID快速查询和过滤
- 便于组织级别的配额管理和计费

### 3. 标签级别管理

- Kubernetes标签体系支持多维度查询
- 支持按组织、用户、服务名筛选资源
- 便于监控、告警和运维管理

## Kubernetes命令查询

### 查询特定组织的所有Namespace

```bash
kubectl get namespaces -l tenant-org-id=org-001
```

### 查询特定用户的所有Namespace

```bash
kubectl get namespaces -l user=alice
```

### 查询特定服务的Namespace

```bash
kubectl get namespaces -l service-name=vector-search
```

### 组合查询

```bash
# 查询org-001组织下alice用户的所有服务
kubectl get namespaces -l tenant-org-id=org-001,user=alice
```

## 配额管理

多租户架构下，可以基于tenant_org_id实现：

1. **组织级配额**：限制每个组织的总资源使用
2. **用户级配额**：限制单个用户的资源使用
3. **服务级配额**：限制单个服务的资源使用

**未来扩展**：

```json
{
  "tenant_org_id": "org-001",
  "quota": {
    "max_clusters": 10,
    "max_cpu": "100",
    "max_memory": "200Gi",
    "max_disk": "1Ti",
    "max_gpu": 5
  },
  "current_usage": {
    "clusters": 2,
    "cpu": "5",
    "memory": "10Gi",
    "disk": "60Gi",
    "gpu": 1
  }
}
```

## 计费和成本管理

基于tenant_org_id，可以实现：

1. **按组织计费**：统计组织总消耗
2. **成本分摊**：组织内部按用户或服务分摊
3. **账单生成**：自动生成组织级别账单

## 错误处理

### 缺少tenant_org_id

**错误信息**：
```json
{
  "error": "tenant_org_id is required for multi-tenancy"
}
```

**HTTP状态码**：400 Bad Request

### 缺少user或service_name

**错误信息**：
```json
{
  "error": "user is required"
}
```

```json
{
  "error": "service_name is required"
}
```

**HTTP状态码**：400 Bad Request

## 相关文档

- **部署上报机制说明**：`/docs/部署上报机制说明.md`
- **时序图集合**：`/docs/时序图集合.md`
- **API文档**：`/README.md`

## 总结

ES Serverless的多租户架构通过以下机制实现完整的租户隔离：

✅ **租户组织ID（tenant_org_id）**：顶层租户标识  
✅ **自动命名空间生成**：{tenant_org_id}-{user}-{service_name}  
✅ **Kubernetes标签体系**：支持多维度查询和管理  
✅ **元数据完整记录**：所有资源都记录tenant_org_id  
✅ **专用查询接口**：支持按组织ID查询所有容器  
✅ **配额和计费基础**：为组织级配额和计费提供基础  

这种设计确保了不同组织之间的完全隔离，同时支持组织内部的灵活管理。
