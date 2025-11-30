package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
)

// ShardController manages Elasticsearch shard allocation and rebalancing
type ShardController struct {
    esClient *ESClient
    ticker   *time.Ticker
}

// NewShardController creates a new shard controller
func NewShardController(esClient *ESClient) *ShardController {
    return &ShardController{
        esClient: esClient,
        ticker:   time.NewTicker(30 * time.Second),
    }
}

// Start begins the shard management loop
func (sc *ShardController) Start() {
    go func() {
        for range sc.ticker.C {
            sc.manageShards()
        }
    }()
}

// Stop stops the shard management loop
func (sc *ShardController) Stop() {
    sc.ticker.Stop()
}

// manageShards handles shard allocation and rebalancing
func (sc *ShardController) manageShards() {
    // Get cluster stats
    stats, err := sc.getClusterStats()
    if err != nil {
        fmt.Printf("Error getting cluster stats: %v\n", err)
        return
    }

    // Check if rebalancing is needed
    if sc.shouldRebalance(stats) {
        sc.rebalanceShards()
    }

    // Check for hot shards
    if sc.hasHotShards(stats) {
        sc.optimizeShardAllocation()
    }
}

// getClusterStats retrieves cluster statistics
func (sc *ShardController) getClusterStats() (map[string]interface{}, error) {
    // Call actual Elasticsearch API to get cluster statistics
    return sc.esClient.GetClusterStats()
}

// shouldRebalance determines if shard rebalancing is needed
func (sc *ShardController) shouldRebalance(stats map[string]interface{}) bool {
    // Check shard distribution balance
    // In a real implementation, this would analyze the actual shard distribution
    // and determine if rebalancing is needed
    
    // Example logic:
    // - If shards are unevenly distributed
    // - If nodes have significantly different loads
    // - If new nodes have been added
    
    nodes := stats["nodes"].(map[string]interface{})
    nodeCount := nodes["count"].(float64)
    
    indices := stats["indices"].(map[string]interface{})
    shards := indices["shards"].(map[string]interface{})
    shardCount := shards["total"].(float64)
    
    // Simple heuristic: if average shards per node > 5, consider rebalancing
    avgShardsPerNode := shardCount / nodeCount
    return avgShardsPerNode > 5
}

// hasHotShards checks if there are hot shards
func (sc *ShardController) hasHotShards(stats map[string]interface{}) bool {
    // In a real implementation, this would check for shards with
    // disproportionately high query rates or resource usage
    return false
}

// rebalanceShards triggers shard rebalancing
func (sc *ShardController) rebalanceShards() {
    log.Println("Triggering shard rebalancing...")
    
    settings := map[string]interface{}{
        "transient": map[string]interface{}{
            "cluster.routing.rebalance.enable":                    "all",
            "cluster.routing.allocation.node_concurrent_recoveries": 2,
            "indices.recovery.max_bytes_per_sec":                 "50mb",
        },
    }
    
    // 实际调用 Elasticsearch API 更新集群设置
    err := sc.esClient.UpdateClusterSettings(settings)
    if err != nil {
        log.Printf("Error triggering rebalance: %v", err)
        return
    }
    
    log.Println("Shard rebalancing triggered successfully")
    
    // 启动异步监控重平衡进度
    go sc.monitorRebalanceProgress()
}

// optimizeShardAllocation optimizes shard allocation
func (sc *ShardController) optimizeShardAllocation() {
    log.Println("Optimizing shard allocation...")
    
    settings := map[string]interface{}{
        "transient": map[string]interface{}{
            "cluster.routing.allocation.balance.shard":  0.45,
            "cluster.routing.allocation.balance.index":  0.55,
            "cluster.routing.allocation.balance.threshold": 1.0,
        },
    }
    
    // 实际调用 Elasticsearch API 更新分配设置
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
                log.Println("Rebalancing completed - no active recoveries")
                return
            }
            
            // 打印进度
            totalRecoveries := 0
            for index, indexRecoveries := range recoveries {
                for _, recovery := range indexRecoveries {
                    log.Printf("[REBALANCE] %s[%d]: %s%% (stage: %s, %s -> %s)",
                        index, recovery.Shard, recovery.Percent, recovery.Stage,
                        recovery.SourceNode, recovery.TargetNode)
                    totalRecoveries++
                }
            }
            log.Printf("[REBALANCE] Total active recoveries: %d", totalRecoveries)
            
        case <-timeout:
            log.Println("[REBALANCE] Monitoring timeout after 30 minutes")
            return
        }
    }
}

// ShardManagementHandler handles shard management API requests
func (sc *ShardController) ShardManagementHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        sc.handleGetShardInfo(w, r)
    case http.MethodPost:
        sc.handleManageShards(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// handleGetShardInfo returns shard information
func (sc *ShardController) handleGetShardInfo(w http.ResponseWriter, r *http.Request) {
    stats, err := sc.getClusterStats()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

// handleManageShards triggers manual shard management
func (sc *ShardController) handleManageShards(w http.ResponseWriter, r *http.Request) {
    var req map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    action, ok := req["action"].(string)
    if !ok {
        http.Error(w, "Missing action parameter", http.StatusBadRequest)
        return
    }
    
    switch action {
    case "rebalance":
        sc.rebalanceShards()
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Shard rebalancing triggered"))
    case "optimize":
        sc.optimizeShardAllocation()
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Shard allocation optimized"))
    default:
        http.Error(w, "Unknown action", http.StatusBadRequest)
    }
}