# ES Serverless 部署上报机制说明

## 概述

ES Serverless 平台在创建容器资源的每个关键步骤都实现了详细的上报机制，确保部署过程的可追溯性和透明度。

## 创建容器组的完整流程及上报步骤

### 📋 流程总览

创建容器组的整个过程包括以下主要阶段：

```
元数据记录 → GitLab配置拉取 → Namespace创建 → K8s资源创建 → 资源配置 → 就绪等待 → 状态同步 → 完成
```

每个阶段都有对应的上报机制，总共**10次上报**。

---

### 🔍 详细步骤及上报说明

#### **阶段0：元数据记录**（在Manager服务中）

**执行内容**：
- 检查租户配额
- 记录租户容器信息到元数据服务
- 保存部署状态（status=creating）

**关键代码位置**：`/server/main.go` - `handleCreate()` 函数

**元数据记录内容**：
```json
{
  "user": "用户ID",
  "service_name": "服务名称",
  "namespace": "命名空间",
  "replicas": 副本数,
  "cpu": "CPU请求/CPU限制",
  "memory": "内存请求/内存限制",
  "disk": "磁盘大小",
  "gpu_count": GPU数量,
  "dimension": 向量维度,
  "vector_count": 向量数量,
  "status": "creating",
  "created_at": "创建时间"
}
```

**⚠️ 重要**：如果元数据记录失败，整个创建流程会终止，不会继续创建K8s资源。

---

#### **第1次上报：starting**

**执行时机**：在创建任何K8s资源之前

**执行脚本**：`/scripts/cluster.sh` - Line 121

**上报内容**：
```json
{
  "status": "starting",
  "message": "Starting cluster creation",
  "timestamp": "2024-01-01T00:00:00Z",
  "details": {
    "user": "用户ID",
    "service_name": "服务名称",
    "namespace": "命名空间",
    "replicas": 1,
    "cpu_request": "500m",
    "cpu_limit": "2",
    "mem_request": "1Gi",
    "mem_limit": "2Gi",
    "disk_size": "10Gi",
    "gpu_count": 0,
    "dimension": 128,
    "vector_count": 10000
  }
}
```

**存储位置**：
- `/server/deployment_reports/{user}_{service_name}_{timestamp}.json`
- `/tmp/deployment.log`

---

#### **第2次上报：namespace_created**

**执行时机**：Kubernetes Namespace创建完成后

**执行脚本**：`/scripts/cluster.sh` - Line 124

**上报内容**：
```json
{
  "status": "namespace_created",
  "message": "Namespace created successfully",
  "timestamp": "2024-01-01T00:00:01Z"
}
```

**对应K8s操作**：
```bash
kubectl create namespace $NAMESPACE
kubectl label namespace $NAMESPACE es-cluster=true
```

---

#### **第3次上报：gitlab_pulled**

**执行时机**：从GitLab拉取配置文件完成后（如果提供了`GITLAB_URL`）

**执行脚本**：`/scripts/cluster.sh` - Line 127

**上报内容**：
```json
{
  "status": "gitlab_pulled",
  "message": "GitLab resources pulled successfully",
  "timestamp": "2024-01-01T00:00:02Z"
}
```

**对应操作**：
- 从GitLab拉取`docker-compose.yml`或其他配置文件
- 如果未提供GitLab URL，这一步会被跳过但仍然上报

---

#### **第4次上报：k8s_applied**

**执行时机**：Kubernetes核心资源（StatefulSet、Service、PVC）应用完成后

**执行脚本**：`/scripts/cluster.sh` - Line 130

**上报内容**：
```json
{
  "status": "k8s_applied",
  "message": "Kubernetes resources applied successfully",
  "timestamp": "2024-01-01T00:00:05Z"
}
```

**对应K8s操作**：
```bash
kubectl apply -k k8s/overlays/dev
```

**创建的资源包括**：
- StatefulSet: `elasticsearch`
- Service: `elasticsearch`, `kibana`
- PersistentVolumeClaim: `elasticsearch-data-elasticsearch-0`
- Deployment: `kibana`
- ConfigMap: Elasticsearch配置

---

#### **第5次上报：resources_configured**

**执行时机**：CPU和内存资源限制设置完成后

**执行脚本**：`/scripts/cluster.sh` - Line 135

**上报内容**：
```json
{
  "status": "resources_configured",
  "message": "Cluster resources configured successfully",
  "timestamp": "2024-01-01T00:00:06Z"
}
```

**对应K8s操作**：
```bash
kubectl -n $NAMESPACE annotate sts/elasticsearch es.yunpeng.cn/max-indices="$INDEX_LIMIT" --overwrite
kubectl -n $NAMESPACE scale sts/elasticsearch --replicas $REPLICAS
kubectl -n $NAMESPACE set resources sts/elasticsearch \
  --requests=cpu="$CPU_REQUEST",memory="$MEM_REQUEST" \
  --limits=cpu="$CPU_LIMIT",memory="$MEM_LIMIT"
```

---

#### **第6次上报：disk_configured**

**执行时机**：PVC磁盘大小配置完成后

**执行脚本**：`/scripts/cluster.sh` - Line 140

**上报内容**：
```json
{
  "status": "disk_configured",
  "message": "Disk size configured to 10Gi",
  "timestamp": "2024-01-01T00:00:07Z"
}
```

**对应K8s操作**：
```bash
kubectl -n $NAMESPACE patch pvc elasticsearch-data-elasticsearch-0 \
  -p '{"spec":{"resources":{"requests":{"storage":"'$DISK_SIZE'"}}}}'
```

---

#### **第7次上报：gpu_configured**

**执行时机**：GPU资源配置完成后（仅当`GPU_COUNT > 0`时）

**执行脚本**：`/scripts/cluster.sh` - Line 146

**上报内容**：
```json
{
  "status": "gpu_configured",
  "message": "GPU count configured to 2",
  "timestamp": "2024-01-01T00:00:08Z"
}
```

**对应K8s操作**：
```bash
kubectl -n $NAMESPACE patch sts elasticsearch \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"elasticsearch","resources":{"limits":{"nvidia.com/gpu":"'$GPU_COUNT'"}}}]}}]}'
```

---

#### **第8次上报：rollout_completed**

**执行时机**：Pod成功启动并运行后

**执行脚本**：`/scripts/cluster.sh` - Line 151

**上报内容**：
```json
{
  "status": "rollout_completed",
  "message": "Cluster rollout completed successfully",
  "timestamp": "2024-01-01T00:01:30Z"
}
```

**对应K8s操作**：
```bash
kubectl -n $NAMESPACE rollout status sts/elasticsearch
kubectl -n $NAMESPACE rollout status deploy/kibana
```

**等待条件**：
- StatefulSet中的所有Pod都处于Running状态
- Deployment中的所有Pod都处于Running状态

---

#### **第9次上报：tenant_synced**

**执行时机**：数据同步到租户容器管理系统完成后

**执行脚本**：`/scripts/cluster.sh` - Line 155

**上报内容**：
```json
{
  "status": "tenant_synced",
  "message": "Data synced to tenant container management",
  "timestamp": "2024-01-01T00:01:31Z"
}
```

**同步操作**：
- 保存租户容器数据到 `/server/tenant_data/{user}_{service_name}.json`
- 记录完整的资源配置和创建时间

---

#### **第10次上报：completed**

**执行时机**：所有步骤成功完成后

**执行脚本**：`/scripts/cluster.sh` - Line 161

**上报内容**：
```json
{
  "status": "completed",
  "message": "Cluster creation completed successfully",
  "timestamp": "2024-01-01T00:01:32Z",
  "details": {
    "user": "用户ID",
    "service_name": "服务名称",
    "namespace": "命名空间",
    "replicas": 1,
    "cpu_request": "500m",
    "cpu_limit": "2",
    "mem_request": "1Gi",
    "mem_limit": "2Gi",
    "disk_size": "10Gi",
    "gpu_count": 0,
    "dimension": 128,
    "vector_count": 10000
  }
}
```

**最终状态更新**（在Manager服务中）：
```go
// 更新部署状态为 "created"
deploymentStatus.Status = "created"
deploymentStatus.UpdatedAt = time.Now()
metadataService.SaveDeploymentStatus(deploymentStatus)

// 更新租户容器状态为 "created"
tenantContainer.Status = "created"
tenantContainer.SyncTime = time.Now()
metadataService.SaveTenantContainer(tenantContainer)
```

---

## 🔄 错误处理和回滚机制

### 元数据记录失败

**场景**：保存租户容器或部署状态失败

**处理方式**：
```go
if err != nil {
    log.Printf("Error: Failed to save tenant container metadata: %v", err)
    http.Error(w, fmt.Sprintf("Failed to save tenant metadata: %v", err), http.StatusInternalServerError)
    return  // 直接返回，不创建K8s资源
}
```

### K8s资源创建失败

**场景**：执行 `cluster.sh create` 脚本失败

**处理方式**：
```go
if err != nil {
    log.Printf("Error: Failed to create K8s resources: %v", err)
    
    // 回滚：删除元数据记录
    metadataService.DeleteTenantContainer(req.User, req.ServiceName)
    
    // 更新部署状态为 "failed"
    deploymentStatus.Status = "failed"
    deploymentStatus.UpdatedAt = time.Now()
    metadataService.SaveDeploymentStatus(deploymentStatus)
    
    w.WriteHeader(http.StatusInternalServerError)
    w.Write(out)
    return
}
```

**上报记录**：错误信息会记录在部署报告和日志文件中

---

## 📊 上报数据存储

### 1. 部署报告文件

**路径**：`/server/deployment_reports/{user}_{service_name}_{timestamp}.json`

**特点**：
- 每次上报都会生成一个独立的JSON文件
- 文件名包含时间戳，便于追踪历史记录
- 包含完整的配置详情

### 2. 部署日志文件

**路径**：`/tmp/deployment.log`

**特点**：
- 追加写入模式
- 每条记录包含时间戳、用户、服务名、状态、消息
- 便于快速查看部署历史

**日志格式**：
```
2024-01-01T00:00:00Z - User: test_user, Service: test_service, Namespace: es-test, Status: starting, Message: Starting cluster creation
2024-01-01T00:00:01Z - User: test_user, Service: test_service, Namespace: es-test, Status: namespace_created, Message: Namespace created successfully
...
```

### 3. 元数据服务存储

**TenantContainer**：
- 存储在 `/server/tenant_data/{user}_{service_name}.json`
- 记录租户容器的配置和状态

**DeploymentStatus**：
- 存储在元数据服务的 `deployments.json`
- 记录部署的详细状态和配置

---

## 🛠️ 代码实现位置

### Manager服务（Go）

**文件**：`/server/main.go`

**关键函数**：
- `handleCreate()` - 处理创建请求，记录元数据，调用cluster.sh
- `handleList()` - 查询部署状态

**元数据服务**：
- `SaveTenantContainer()` - 保存租户容器信息
- `SaveDeploymentStatus()` - 保存部署状态
- `DeleteTenantContainer()` - 删除租户容器记录（回滚用）

### 部署脚本（Bash）

**文件**：`/scripts/cluster.sh`

**关键函数**：
- `report_deployment_status()` - 上报部署状态（Line 74-117）
- `sync_to_tenant_management()` - 同步租户数据（Line 37-71）
- `create_namespace()` - 创建命名空间（Line 21-24）
- `pull_from_gitlab()` - 拉取GitLab配置（Line 27-34）

---

## 📈 查询上报记录

### 1. 查询特定部署的所有上报记录

```bash
ls -lt /server/deployment_reports/test_user_test_service_*.json
```

### 2. 查看最新的上报内容

```bash
cat /server/deployment_reports/test_user_test_service_$(ls -t /server/deployment_reports/test_user_test_service_*.json | head -1 | xargs basename | cut -d'_' -f4 | cut -d'.' -f1).json | jq .
```

### 3. 查看部署日志

```bash
tail -f /tmp/deployment.log
```

### 4. 通过API查询部署状态

```bash
curl http://localhost:8080/deployments?user=test_user&service_name=test_service
```

---

## 🔍 时序图参考

详细的创建容器组时序图请参考：`/docs/时序图集合.md` - **1. 创建容器组**

该时序图清晰展示了：
- 元数据记录（⭐标记）
- 10次上报步骤（📊标记）
- 各个服务之间的交互
- 错误处理流程

---

## 📝 总结

ES Serverless平台的部署上报机制具有以下特点：

✅ **完整性**：覆盖创建流程的每个关键步骤，共10次上报  
✅ **可追溯性**：每次上报都有时间戳和详细信息  
✅ **多重存储**：同时保存到文件、日志和元数据服务  
✅ **错误处理**：失败时有明确的回滚机制  
✅ **易于查询**：支持API和文件系统两种查询方式  

这种详细的上报机制确保了部署过程的透明度和可维护性，便于问题排查和状态监控。
