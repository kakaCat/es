# ES 分片数据同步实现方案

## 现状分析

### ❌ 当前问题

根据代码审查，ES Serverless 平台的分片数据同步功能**尚未完全实现**：

1. **`/server/shard_controller.go` 中的问题**：
   - Line 117-119: `rebalanceShards()` 只打印配置，没有实际调用 Elasticsearch API
   - Line 134-136: `optimizeShardAllocation()` 同样只打印配置
   - 缺少实际的 HTTP 请求到 Elasticsearch

2. **缺失的功能**：
   - ❌ 主分片到副本分片的实时数据同步
   - ❌ 分片恢复进度监控
   - ❌ 副本故障自动恢复
   - ❌ 数据一致性验证
   - ❌ 同步失败重试机制

### ✅ 已有基础

- ✅ 分片控制器框架（ShardController）
- ✅ 定时监控机制（每30秒）
- ✅ 分片重平衡决策逻辑
- ✅ Elasticsearch 客户端（ESClient）
- ✅ 副本配置（number_of_replicas）

---

## 完整实现方案

### 阶段一：实现实际的API调用

#### 1.1 完善 ESClient

**文件**：`/server/es_client.go`

需要添加以下方法：

```go
// UpdateClusterSettings updates Elasticsearch cluster settings
func (c *ESClient) UpdateClusterSettings(settings map[string]interface{}) error {
    url := fmt.Sprintf("%s/_cluster/settings", c.baseURL)
    
    jsonData, err := json.Marshal(settings)
    if err != nil {
        return fmt.Errorf("failed to marshal settings: %v", err)
    }
    
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %v", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to update cluster settings: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("ES API error: %s, body: %s", resp.Status, string(body))
    }
    
    return nil
}

// GetRecoveryStatus gets shard recovery status
func (c *ESClient) GetRecoveryStatus() ([]ShardRecovery, error) {
    url := fmt.Sprintf("%s/_recovery?active_only=true", c.baseURL)
    
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to get recovery status: %v", err)
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }
    
    // Parse recovery information
    recoveries := parseRecoveryInfo(result)
    return recoveries, nil
}

// GetShardAllocation gets shard allocation information
func (c *ESClient) GetShardAllocation() ([]ShardInfo, error) {
    url := fmt.Sprintf("%s/_cat/shards?format=json", c.baseURL)
    
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to get shard allocation: %v", err)
    }
    defer resp.Body.Close()
    
    var shards []ShardInfo
    if err := json.NewDecoder(resp.Body).Decode(&shards); err != nil {
        return nil, fmt.Errorf("failed to decode shards: %v", err)
    }
    
    return shards, nil
}

// VerifyShardConsistency verifies data consistency between primary and replica
func (c *ESClient) VerifyShardConsistency(index string) (bool, error) {
    url := fmt.Sprintf("%s/%s/_search_shards", c.baseURL, index)
    
    resp, err := c.client.Get(url)
    if err != nil {
        return false, fmt.Errorf("failed to verify consistency: %v", err)
    }
    defer resp.Body.Close()
    
    // Check if all shards are in sync
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false, fmt.Errorf("failed to decode response: %v", err)
    }
    
    // Analyze shard states
    return analyzeShardStates(result), nil
}
```

#### 1.2 定义数据结构

```go
// ShardRecovery represents shard recovery information
type ShardRecovery struct {
    Index         string  `json:"index"`
    Shard         int     `json:"shard"`
    Type          string  `json:"type"`          // primary or replica
    Stage         string  `json:"stage"`         // init, index, verify_index, translog, finalize, done
    SourceNode    string  `json:"source_node"`
    TargetNode    string  `json:"target_node"`
    BytesRecovered int64  `json:"bytes_recovered"`
    BytesTotal    int64   `json:"bytes_total"`
    Percent       float64 `json:"percent"`
    StartTime     string  `json:"start_time"`
}

// ShardInfo represents shard information
type ShardInfo struct {
    Index    string `json:"index"`
    Shard    string `json:"shard"`
    Prirep   string `json:"prirep"`   // p=primary, r=replica
    State    string `json:"state"`     // STARTED, RELOCATING, INITIALIZING, UNASSIGNED
    Docs     string `json:"docs"`
    Store    string `json:"store"`
    IP       string `json:"ip"`
    Node     string `json:"node"`
}

// ReplicationStatus represents replication status
type ReplicationStatus struct {
    Index           string             `json:"index"`
    Shard           int                `json:"shard"`
    PrimaryNode     string             `json:"primary_node"`
    ReplicaNodes    []string           `json:"replica_nodes"`
    InSync          bool               `json:"in_sync"`
    SyncLag         int64              `json:"sync_lag_bytes"`
    LastSyncTime    time.Time          `json:"last_sync_time"`
}
```

#### 1.3 更新 ShardController

**文件**：`/server/shard_controller.go`

```go
// rebalanceShards triggers shard rebalancing with actual API call
func (sc *ShardController) rebalanceShards() {
    log.Println("Triggering shard rebalancing...")
    
    settings := map[string]interface{}{
        "transient": map[string]interface{}{
            "cluster.routing.rebalance.enable":                    "all",
            "cluster.routing.allocation.node_concurrent_recoveries": 2,
            "indices.recovery.max_bytes_per_sec":                 "50mb",
        },
    }
    
    // 实际调用 Elasticsearch API
    err := sc.esClient.UpdateClusterSettings(settings)
    if err != nil {
        log.Printf("Error triggering rebalance: %v", err)
        return
    }
    
    log.Println("Shard rebalancing triggered successfully")
    
    // 启动异步监控重平衡进度
    go sc.monitorRebalanceProgress()
}

// optimizeShardAllocation optimizes shard allocation with actual API call
func (sc *ShardController) optimizeShardAllocation() {
    log.Println("Optimizing shard allocation...")
    
    settings := map[string]interface{}{
        "transient": map[string]interface{}{
            "cluster.routing.allocation.balance.shard":      0.45,
            "cluster.routing.allocation.balance.index":      0.55,
            "cluster.routing.allocation.balance.threshold":  1.0,
        },
    }
    
    // 实际调用 Elasticsearch API
    err := sc.esClient.UpdateClusterSettings(settings)
    if err != nil {
        log.Printf("Error optimizing allocation: %v", err)
        return
    }
    
    log.Println("Shard allocation optimized successfully")
}

// monitorRebalanceProgress monitors shard rebalancing progress
func (sc *ShardController) monitorRebalanceProgress() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    timeout := time.After(30 * time.Minute) // 30分钟超时
    
    for {
        select {
        case <-ticker.C:
            recoveries, err := sc.esClient.GetRecoveryStatus()
            if err != nil {
                log.Printf("Error getting recovery status: %v", err)
                continue
            }
            
            if len(recoveries) == 0 {
                log.Println("Rebalancing completed")
                return
            }
            
            // 打印进度
            for _, recovery := range recoveries {
                log.Printf("Rebalancing %s[%d]: %.2f%% (stage: %s)",
                    recovery.Index, recovery.Shard, recovery.Percent, recovery.Stage)
            }
            
        case <-timeout:
            log.Println("Rebalancing monitoring timeout")
            return
        }
    }
}
```

---

### 阶段二：实现副本数据同步监控

#### 2.1 副本同步状态检查

```go
// ReplicationMonitor monitors replica synchronization
type ReplicationMonitor struct {
    esClient *ESClient
    ticker   *time.Ticker
}

// NewReplicationMonitor creates a new replication monitor
func NewReplicationMonitor(esClient *ESClient) *ReplicationMonitor {
    return &ReplicationMonitor{
        esClient: esClient,
        ticker:   time.NewTicker(10 * time.Second), // 每10秒检查一次
    }
}

// Start begins monitoring replication
func (rm *ReplicationMonitor) Start() {
    go func() {
        for range rm.ticker.C {
            rm.checkReplication()
        }
    }()
}

// Stop stops the monitoring
func (rm *ReplicationMonitor) Stop() {
    rm.ticker.Stop()
}

// checkReplication checks replication status
func (rm *ReplicationMonitor) checkReplication() {
    shards, err := rm.esClient.GetShardAllocation()
    if err != nil {
        log.Printf("Error getting shard allocation: %v", err)
        return
    }
    
    // 分析主副本关系
    replicationStatus := rm.analyzeReplication(shards)
    
    // 检查不健康的副本
    for _, status := range replicationStatus {
        if !status.InSync {
            log.Printf("WARNING: Replica out of sync: index=%s, shard=%d, lag=%d bytes",
                status.Index, status.Shard, status.SyncLag)
            
            // 触发恢复
            rm.triggerRecovery(status)
        }
    }
}

// analyzeReplication analyzes replica synchronization
func (rm *ReplicationMonitor) analyzeReplication(shards []ShardInfo) []ReplicationStatus {
    // 按索引和分片编号分组
    shardMap := make(map[string]map[int]*ReplicationStatus)
    
    for _, shard := range shards {
        if shardMap[shard.Index] == nil {
            shardMap[shard.Index] = make(map[int]*ReplicationStatus)
        }
        
        shardNum, _ := strconv.Atoi(shard.Shard)
        
        if shardMap[shard.Index][shardNum] == nil {
            shardMap[shard.Index][shardNum] = &ReplicationStatus{
                Index:        shard.Index,
                Shard:        shardNum,
                ReplicaNodes: []string{},
                InSync:       true,
            }
        }
        
        status := shardMap[shard.Index][shardNum]
        
        if shard.Prirep == "p" {
            status.PrimaryNode = shard.Node
        } else {
            status.ReplicaNodes = append(status.ReplicaNodes, shard.Node)
        }
        
        // 检查状态
        if shard.State != "STARTED" {
            status.InSync = false
        }
    }
    
    // 转换为数组
    var results []ReplicationStatus
    for _, indexShards := range shardMap {
        for _, status := range indexShards {
            results = append(results, *status)
        }
    }
    
    return results
}

// triggerRecovery triggers replica recovery
func (rm *ReplicationMonitor) triggerRecovery(status ReplicationStatus) {
    log.Printf("Triggering recovery for %s[%d]", status.Index, status.Shard)
    
    // 调用 Elasticsearch API 触发恢复
    settings := map[string]interface{}{
        "transient": map[string]interface{}{
            fmt.Sprintf("cluster.routing.allocation.enable"): "all",
        },
    }
    
    err := rm.esClient.UpdateClusterSettings(settings)
    if err != nil {
        log.Printf("Error triggering recovery: %v", err)
    }
}
```

---

### 阶段三：实现数据一致性验证

#### 3.1 一致性检查服务

```go
// ConsistencyChecker checks data consistency
type ConsistencyChecker struct {
    esClient *ESClient
}

// NewConsistencyChecker creates a new consistency checker
func NewConsistencyChecker(esClient *ESClient) *ConsistencyChecker {
    return &ConsistencyChecker{
        esClient: esClient,
    }
}

// VerifyIndexConsistency verifies index consistency
func (cc *ConsistencyChecker) VerifyIndexConsistency(index string) (*ConsistencyReport, error) {
    report := &ConsistencyReport{
        Index:       index,
        CheckTime:   time.Now(),
        Consistent:  true,
        Issues:      []string{},
    }
    
    // 1. 检查分片状态
    shards, err := cc.esClient.GetShardAllocation()
    if err != nil {
        return nil, err
    }
    
    // 过滤该索引的分片
    indexShards := filterShardsByIndex(shards, index)
    
    // 2. 检查每个分片的主副本
    for shardNum, shardGroup := range groupShardsByNumber(indexShards) {
        primary := findPrimaryShard(shardGroup)
        replicas := findReplicaShards(shardGroup)
        
        if primary == nil {
            report.Consistent = false
            report.Issues = append(report.Issues, 
                fmt.Sprintf("Shard %d: No primary shard found", shardNum))
            continue
        }
        
        // 3. 比较主副本文档数量
        for _, replica := range replicas {
            if primary.Docs != replica.Docs {
                report.Consistent = false
                report.Issues = append(report.Issues,
                    fmt.Sprintf("Shard %d: Doc count mismatch - primary: %s, replica on %s: %s",
                        shardNum, primary.Docs, replica.Node, replica.Docs))
            }
            
            // 4. 比较存储大小
            if primary.Store != replica.Store {
                report.Issues = append(report.Issues,
                    fmt.Sprintf("Shard %d: Store size difference - primary: %s, replica on %s: %s",
                        shardNum, primary.Store, replica.Node, replica.Store))
            }
        }
    }
    
    return report, nil
}

// ConsistencyReport represents consistency check report
type ConsistencyReport struct {
    Index      string    `json:"index"`
    CheckTime  time.Time `json:"check_time"`
    Consistent bool      `json:"consistent"`
    Issues     []string  `json:"issues"`
}
```

---

### 阶段四：实现故障自动恢复

#### 4.1 自动恢复机制

```go
// AutoRecoveryManager manages automatic recovery
type AutoRecoveryManager struct {
    esClient   *ESClient
    ticker     *time.Ticker
    maxRetries int
}

// NewAutoRecoveryManager creates a new auto recovery manager
func NewAutoRecoveryManager(esClient *ESClient) *AutoRecoveryManager {
    return &AutoRecoveryManager{
        esClient:   esClient,
        ticker:     time.NewTicker(30 * time.Second),
        maxRetries: 3,
    }
}

// Start begins auto recovery monitoring
func (arm *AutoRecoveryManager) Start() {
    go func() {
        for range arm.ticker.C {
            arm.checkAndRecover()
        }
    }()
}

// Stop stops the auto recovery
func (arm *AutoRecoveryManager) Stop() {
    arm.ticker.Stop()
}

// checkAndRecover checks for failed shards and recovers them
func (arm *AutoRecoveryManager) checkAndRecover() {
    shards, err := arm.esClient.GetShardAllocation()
    if err != nil {
        log.Printf("Error getting shards: %v", err)
        return
    }
    
    // 查找未分配的分片
    unassignedShards := filterUnassignedShards(shards)
    
    if len(unassignedShards) > 0 {
        log.Printf("Found %d unassigned shards, triggering recovery", len(unassignedShards))
        arm.recoverUnassignedShards(unassignedShards)
    }
    
    // 查找初始化失败的分片
    failedShards := filterFailedShards(shards)
    
    if len(failedShards) > 0 {
        log.Printf("Found %d failed shards, attempting recovery", len(failedShards))
        arm.recoverFailedShards(failedShards)
    }
}

// recoverUnassignedShards recovers unassigned shards
func (arm *AutoRecoveryManager) recoverUnassignedShards(shards []ShardInfo) {
    settings := map[string]interface{}{
        "transient": map[string]interface{}{
            "cluster.routing.allocation.enable": "all",
            "cluster.routing.rebalance.enable":  "all",
        },
    }
    
    err := arm.esClient.UpdateClusterSettings(settings)
    if err != nil {
        log.Printf("Error enabling allocation: %v", err)
        return
    }
    
    // 等待分配完成
    time.Sleep(10 * time.Second)
    
    // 验证恢复结果
    arm.verifyRecovery(shards)
}

// recoverFailedShards recovers failed shards
func (arm *AutoRecoveryManager) recoverFailedShards(shards []ShardInfo) {
    for _, shard := range shards {
        log.Printf("Attempting to recover failed shard: %s[%s]", shard.Index, shard.Shard)
        
        // 尝试重新分配
        // 在实际实现中，可能需要删除并重新创建分片
        // 这取决于具体的故障原因
    }
}

// verifyRecovery verifies recovery success
func (arm *AutoRecoveryManager) verifyRecovery(originalShards []ShardInfo) {
    time.Sleep(30 * time.Second)
    
    currentShards, err := arm.esClient.GetShardAllocation()
    if err != nil {
        log.Printf("Error verifying recovery: %v", err)
        return
    }
    
    stillUnassigned := filterUnassignedShards(currentShards)
    
    if len(stillUnassigned) > 0 {
        log.Printf("WARNING: %d shards still unassigned after recovery attempt", len(stillUnassigned))
    } else {
        log.Println("All shards successfully recovered")
    }
}
```

---

## 实现步骤

### 第一步：完善 ESClient（1-2天）

1. 在 `es_client.go` 中添加新方法
2. 定义相关数据结构
3. 编写单元测试

### 第二步：更新 ShardController（1天）

1. 修改 `rebalanceShards()` 使用实际API调用
2. 修改 `optimizeShardAllocation()` 使用实际API调用
3. 添加 `monitorRebalanceProgress()` 方法
4. 测试重平衡功能

### 第三步：实现副本监控（2天）

1. 创建 `ReplicationMonitor` 结构体
2. 实现副本同步检查逻辑
3. 集成到 main.go
4. 测试副本同步监控

### 第四步：实现一致性检查（1-2天）

1. 创建 `ConsistencyChecker` 结构体
2. 实现一致性验证逻辑
3. 添加 API 接口
4. 测试一致性检查

### 第五步：实现自动恢复（2-3天）

1. 创建 `AutoRecoveryManager` 结构体
2. 实现自动恢复逻辑
3. 添加重试机制
4. 测试故障恢复

### 第六步：集成和测试（2-3天）

1. 将所有组件集成到系统
2. 编写集成测试
3. 性能测试
4. 文档更新

---

## 监控和可观测性

### 1. 添加监控指标

```go
// Metrics for replication monitoring
type ReplicationMetrics struct {
    // 副本同步延迟
    SyncLagBytes prometheus.Gauge
    
    // 未分配分片数
    UnassignedShards prometheus.Gauge
    
    // 正在恢复的分片数
    RecoveringShards prometheus.Gauge
    
    // 数据一致性状态
    ConsistencyStatus prometheus.Gauge
    
    // 自动恢复次数
    AutoRecoveryAttempts prometheus.Counter
    
    // 恢复成功次数
    RecoverySuccesses prometheus.Counter
    
    // 恢复失败次数
    RecoveryFailures prometheus.Counter
}
```

### 2. 添加日志记录

```go
// 详细的日志记录
log.Printf("[REPLICATION] Starting replication check for index: %s", index)
log.Printf("[REPLICATION] Primary shard on node: %s, docs: %s", primary.Node, primary.Docs)
log.Printf("[REPLICATION] Replica shard on node: %s, docs: %s, in_sync: %v", replica.Node, replica.Docs, inSync)
log.Printf("[RECOVERY] Triggering recovery for %d unassigned shards", len(unassigned))
log.Printf("[RECOVERY] Recovery completed for shard %s[%d], time taken: %v", index, shard, duration)
```

### 3. 添加告警

```go
// 告警规则
if syncLag > 1GB {
    alertManager.SendAlert("High replication lag", fmt.Sprintf("Lag: %d bytes", syncLag))
}

if unassignedCount > 0 {
    alertManager.SendAlert("Unassigned shards detected", fmt.Sprintf("Count: %d", unassignedCount))
}
```

---

## API 接口扩展

### 1. 副本状态查询

```
GET /shards/replication
```

返回所有索引的副本同步状态

### 2. 一致性检查

```
GET /shards/consistency/{index}
```

检查指定索引的数据一致性

### 3. 触发恢复

```
POST /shards/recovery
{
  "index": "my-index",
  "shard": 0
}
```

手动触发特定分片的恢复

### 4. 查看恢复进度

```
GET /shards/recovery/status
```

查看当前正在进行的恢复任务

---

## 测试计划

### 单元测试

```go
func TestRebalanceShards(t *testing.T) {
    // Test rebalancing with mock ES client
}

func TestReplicationMonitor(t *testing.T) {
    // Test replication monitoring
}

func TestConsistencyChecker(t *testing.T) {
    // Test consistency checking
}

func TestAutoRecovery(t *testing.T) {
    // Test auto recovery
}
```

### 集成测试

```bash
# 创建索引并插入数据
# 模拟节点故障
# 验证自动恢复
# 检查数据一致性
```

### 性能测试

- 测试大规模数据同步性能
- 测试并发恢复性能
- 测试监控开销

---

## 预期成果

实现完成后，系统将具备：

✅ 实际的分片重平衡功能（调用真实的 ES API）  
✅ 实时的副本同步监控  
✅ 自动的数据一致性验证  
✅ 智能的故障自动恢复  
✅ 完整的监控和告警  
✅ 详细的恢复进度追踪  

---

## 相关文档

- Elasticsearch 官方文档：https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-cluster.html
- 分片分配策略：https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-cluster.html#shards-allocation
- 副本恢复：https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-recovery.html
