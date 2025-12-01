package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ReplicationStatus represents the replication status of an index
type ReplicationStatus struct {
	Index               string    `json:"index"`
	TotalShards         int       `json:"total_shards"`
	ReplicatedShards    int       `json:"replicated_shards"`
	UnreplicatedShards  int       `json:"unreplicated_shards"`
	ReplicationProgress float64   `json:"replication_progress"` // 0-100%
	LastCheckTime       time.Time `json:"last_check_time"`
	Status              string    `json:"status"` // healthy, syncing, degraded, failed
}

// ReplicaHealth represents the health status of a replica shard
type ReplicaHealth struct {
	Index        string `json:"index"`
	Shard        int    `json:"shard"`
	Primary      bool   `json:"primary"`
	State        string `json:"state"`        // STARTED, INITIALIZING, RELOCATING, UNASSIGNED
	Node         string `json:"node"`
	SyncDelay    int64  `json:"sync_delay"`   // 同步延迟（毫秒）
	DocsCount    int64  `json:"docs_count"`
	StoreSize    string `json:"store_size"`
	LastSyncTime string `json:"last_sync_time"`
}

// ReplicationMonitor monitors replica synchronization status
type ReplicationMonitor struct {
	esClient          *ESClient
	ticker            *time.Ticker
	stopChan          chan bool
	statuses          map[string]*ReplicationStatus // index -> status
	statusMutex       sync.RWMutex
	syncDelayThreshold int64 // 同步延迟阈值（毫秒）
	checkInterval     time.Duration
}

// NewReplicationMonitor creates a new replication monitor
func NewReplicationMonitor(esClient *ESClient) *ReplicationMonitor {
	return &ReplicationMonitor{
		esClient:          esClient,
		stopChan:          make(chan bool),
		statuses:          make(map[string]*ReplicationStatus),
		syncDelayThreshold: 5000, // 默认5秒阈值
		checkInterval:     30 * time.Second, // 每30秒检查一次
	}
}

// Start starts the replication monitor
func (rm *ReplicationMonitor) Start() {
	log.Println("[ReplicationMonitor] Starting replica synchronization monitoring...")
	
	rm.ticker = time.NewTicker(rm.checkInterval)
	
	// 立即执行一次检查
	go rm.checkReplicationStatus()
	
	// 定期检查
	go func() {
		for {
			select {
			case <-rm.ticker.C:
				rm.checkReplicationStatus()
			case <-rm.stopChan:
				log.Println("[ReplicationMonitor] Stopping...")
				return
			}
		}
	}()
}

// Stop stops the replication monitor
func (rm *ReplicationMonitor) Stop() {
	if rm.ticker != nil {
		rm.ticker.Stop()
	}
	rm.stopChan <- true
}

// checkReplicationStatus checks the replication status of all indices
func (rm *ReplicationMonitor) checkReplicationStatus() {
	log.Println("[ReplicationMonitor] Checking replication status...")
	
	// 1. 获取所有分片信息
	shards, err := rm.esClient.GetShardAllocation()
	if err != nil {
		log.Printf("[ReplicationMonitor] Error getting shard allocation: %v", err)
		return
	}
	
	// 2. 按索引分组统计
	indexStats := make(map[string]*ReplicationStatus)
	replicaHealths := make(map[string][]ReplicaHealth)
	
	for _, shard := range shards {
		if _, exists := indexStats[shard.Index]; !exists {
			indexStats[shard.Index] = &ReplicationStatus{
				Index:         shard.Index,
				LastCheckTime: time.Now(),
				Status:        "healthy",
			}
		}
		
		stats := indexStats[shard.Index]
		stats.TotalShards++
		
		// 构建副本健康信息
		health := ReplicaHealth{
			Index: shard.Index,
			State: shard.State,
			Node:  shard.Node,
		}
		
		// 判断是否是副本
		isPrimary := shard.Prirep == "p"
		health.Primary = isPrimary
		
		replicaHealths[shard.Index] = append(replicaHealths[shard.Index], health)
		
		// 统计复制状态
		if shard.State == "STARTED" {
			stats.ReplicatedShards++
		} else {
			stats.UnreplicatedShards++
			// 更新索引状态
			if shard.State == "INITIALIZING" {
				stats.Status = "syncing"
			} else if shard.State == "UNASSIGNED" {
				stats.Status = "degraded"
			}
		}
	}
	
	// 3. 计算复制进度
	for _, stats := range indexStats {
		if stats.TotalShards > 0 {
			stats.ReplicationProgress = float64(stats.ReplicatedShards) / float64(stats.TotalShards) * 100
		}
		
		// 判断整体状态
		if stats.UnreplicatedShards > 0 {
			if stats.ReplicationProgress < 50 {
				stats.Status = "failed"
			} else if stats.ReplicationProgress < 100 {
				stats.Status = "syncing"
			}
		}
	}
	
	// 4. 更新状态缓存
	rm.statusMutex.Lock()
	rm.statuses = indexStats
	rm.statusMutex.Unlock()
	
	// 5. 打印报告
	rm.printReplicationReport(indexStats)
	
	// 6. 检查并处理异常
	rm.handleReplicationIssues(indexStats, replicaHealths)
}

// printReplicationReport prints replication status report
func (rm *ReplicationMonitor) printReplicationReport(stats map[string]*ReplicationStatus) {
	log.Println("========================================")
	log.Println("[ReplicationMonitor] Replication Status Report")
	log.Println("========================================")
	
	totalIndices := len(stats)
	healthyIndices := 0
	syncingIndices := 0
	degradedIndices := 0
	failedIndices := 0
	
	for index, status := range stats {
		log.Printf("[INDEX] %s: %.1f%% (%d/%d shards) - Status: %s",
			index, status.ReplicationProgress,
			status.ReplicatedShards, status.TotalShards,
			status.Status)
		
		switch status.Status {
		case "healthy":
			healthyIndices++
		case "syncing":
			syncingIndices++
		case "degraded":
			degradedIndices++
		case "failed":
			failedIndices++
		}
	}
	
	log.Println("========================================")
	log.Printf("[SUMMARY] Total: %d | Healthy: %d | Syncing: %d | Degraded: %d | Failed: %d",
		totalIndices, healthyIndices, syncingIndices, degradedIndices, failedIndices)
	log.Println("========================================")
}

// handleReplicationIssues handles replication issues
func (rm *ReplicationMonitor) handleReplicationIssues(stats map[string]*ReplicationStatus, healths map[string][]ReplicaHealth) {
	for index, status := range stats {
		switch status.Status {
		case "failed":
			log.Printf("[ALERT] Index %s replication FAILED: %.1f%% complete", index, status.ReplicationProgress)
			rm.triggerRecovery(index, healths[index])
			
		case "degraded":
			log.Printf("[WARNING] Index %s replication DEGRADED: %d unassigned shards", 
				index, status.UnreplicatedShards)
			
		case "syncing":
			log.Printf("[INFO] Index %s is SYNCING: %.1f%% complete", 
				index, status.ReplicationProgress)
		}
	}
}

// triggerRecovery triggers recovery for failed replication
func (rm *ReplicationMonitor) triggerRecovery(index string, healths []ReplicaHealth) {
	log.Printf("[RECOVERY] Triggering recovery for index: %s", index)
	
	// 统计未分配的分片
	unassignedShards := []int{}
	for _, health := range healths {
		if health.State == "UNASSIGNED" {
			unassignedShards = append(unassignedShards, health.Shard)
		}
	}
	
	if len(unassignedShards) > 0 {
		log.Printf("[RECOVERY] Found %d unassigned shards: %v", len(unassignedShards), unassignedShards)
		
		// 尝试重新分配
		settings := map[string]interface{}{
			"transient": map[string]interface{}{
				"cluster.routing.allocation.enable": "all",
			},
		}
		
		err := rm.esClient.UpdateClusterSettings(settings)
		if err != nil {
			log.Printf("[RECOVERY] Error enabling allocation: %v", err)
		} else {
			log.Printf("[RECOVERY] Allocation enabled for index: %s", index)
		}
	}
	
	// 统计初始化中的分片
	initializingShards := []int{}
	for _, health := range healths {
		if health.State == "INITIALIZING" {
			initializingShards = append(initializingShards, health.Shard)
		}
	}
	
	if len(initializingShards) > 0 {
		log.Printf("[RECOVERY] %d shards initializing: %v", len(initializingShards), initializingShards)
	}
}

// GetReplicationStatus returns the current replication status for an index
func (rm *ReplicationMonitor) GetReplicationStatus(index string) (*ReplicationStatus, error) {
	rm.statusMutex.RLock()
	defer rm.statusMutex.RUnlock()
	
	status, exists := rm.statuses[index]
	if !exists {
		return nil, fmt.Errorf("no status found for index: %s", index)
	}
	
	return status, nil
}

// GetAllReplicationStatuses returns all replication statuses
func (rm *ReplicationMonitor) GetAllReplicationStatuses() map[string]*ReplicationStatus {
	rm.statusMutex.RLock()
	defer rm.statusMutex.RUnlock()
	
	// 返回副本以避免并发修改
	statuses := make(map[string]*ReplicationStatus)
	for k, v := range rm.statuses {
		statuses[k] = v
	}
	
	return statuses
}

// SetSyncDelayThreshold sets the sync delay threshold
func (rm *ReplicationMonitor) SetSyncDelayThreshold(threshold int64) {
	rm.syncDelayThreshold = threshold
	log.Printf("[ReplicationMonitor] Sync delay threshold set to %d ms", threshold)
}

// SetCheckInterval sets the check interval
func (rm *ReplicationMonitor) SetCheckInterval(interval time.Duration) {
	rm.checkInterval = interval
	log.Printf("[ReplicationMonitor] Check interval set to %v", interval)
	
	// 重启定时器
	if rm.ticker != nil {
		rm.ticker.Stop()
		rm.ticker = time.NewTicker(interval)
	}
}
