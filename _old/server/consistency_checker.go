package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ConsistencyReport represents a consistency check report for an index
type ConsistencyReport struct {
	Index              string                   `json:"index"`
	CheckTime          time.Time                `json:"check_time"`
	Status             string                   `json:"status"` // consistent, inconsistent, checking, error
	TotalShards        int                      `json:"total_shards"`
	ConsistentShards   int                      `json:"consistent_shards"`
	InconsistentShards int                      `json:"inconsistent_shards"`
	ShardReports       []ShardConsistencyReport `json:"shard_reports"`
	Issues             []string                 `json:"issues"`
}

// ShardConsistencyReport represents consistency check for a single shard
type ShardConsistencyReport struct {
	ShardID          int     `json:"shard_id"`
	PrimaryNode      string  `json:"primary_node"`
	ReplicaNodes     []string `json:"replica_nodes"`
	PrimaryDocCount  int64   `json:"primary_doc_count"`
	ReplicaDocCounts []int64 `json:"replica_doc_counts"`
	PrimaryStoreSize string  `json:"primary_store_size"`
	ReplicaStoreSizes []string `json:"replica_store_sizes"`
	IsConsistent     bool    `json:"is_consistent"`
	Issues           []string `json:"issues"`
}

// ConsistencyChecker checks data consistency between primary and replica shards
type ConsistencyChecker struct {
	esClient       *ESClient
	ticker         *time.Ticker
	stopChan       chan bool
	reports        map[string]*ConsistencyReport // index -> report
	reportsMutex   sync.RWMutex
	checkInterval  time.Duration
	docCountTolerance int64 // 文档数量允许的差异阈值
}

// NewConsistencyChecker creates a new consistency checker
func NewConsistencyChecker(esClient *ESClient) *ConsistencyChecker {
	return &ConsistencyChecker{
		esClient:         esClient,
		stopChan:         make(chan bool),
		reports:          make(map[string]*ConsistencyReport),
		checkInterval:    5 * time.Minute, // 默认每5分钟检查一次
		docCountTolerance: 10,               // 允许10个文档的差异
	}
}

// Start starts the consistency checker
func (cc *ConsistencyChecker) Start() {
	log.Println("[ConsistencyChecker] Starting data consistency checking...")
	
	cc.ticker = time.NewTicker(cc.checkInterval)
	
	// 立即执行一次检查
	go cc.checkAllIndices()
	
	// 定期检查
	go func() {
		for {
			select {
			case <-cc.ticker.C:
				cc.checkAllIndices()
			case <-cc.stopChan:
				log.Println("[ConsistencyChecker] Stopping...")
				return
			}
		}
	}()
}

// Stop stops the consistency checker
func (cc *ConsistencyChecker) Stop() {
	if cc.ticker != nil {
		cc.ticker.Stop()
	}
	cc.stopChan <- true
}

// checkAllIndices checks consistency for all indices
func (cc *ConsistencyChecker) checkAllIndices() {
	log.Println("[ConsistencyChecker] Starting consistency check for all indices...")
	
	// 获取所有分片信息
	shards, err := cc.esClient.GetShardAllocation()
	if err != nil {
		log.Printf("[ConsistencyChecker] Error getting shard allocation: %v", err)
		return
	}
	
	// 按索引分组
	indexShards := make(map[string][]ShardInfo)
	for _, shard := range shards {
		indexShards[shard.Index] = append(indexShards[shard.Index], shard)
	}
	
	// 对每个索引进行一致性检查
	for index, shards := range indexShards {
		report := cc.checkIndexConsistency(index, shards)
		
		// 保存报告
		cc.reportsMutex.Lock()
		cc.reports[index] = report
		cc.reportsMutex.Unlock()
	}
	
	// 打印总结报告
	cc.printSummaryReport()
}

// checkIndexConsistency checks consistency for a single index
func (cc *ConsistencyChecker) checkIndexConsistency(index string, shards []ShardInfo) *ConsistencyReport {
	log.Printf("[ConsistencyChecker] Checking consistency for index: %s", index)
	
	report := &ConsistencyReport{
		Index:        index,
		CheckTime:    time.Now(),
		Status:       "checking",
		ShardReports: []ShardConsistencyReport{},
		Issues:       []string{},
	}
	
	// 按分片ID分组
	shardGroups := make(map[string][]ShardInfo)
	for _, shard := range shards {
		shardGroups[shard.Shard] = append(shardGroups[shard.Shard], shard)
	}
	
	report.TotalShards = len(shardGroups)
	
	// 检查每个分片组（主分片 + 副本）
	for shardID, shardGroup := range shardGroups {
		shardReport := cc.checkShardConsistency(index, shardID, shardGroup)
		report.ShardReports = append(report.ShardReports, shardReport)
		
		if shardReport.IsConsistent {
			report.ConsistentShards++
		} else {
			report.InconsistentShards++
			report.Issues = append(report.Issues, shardReport.Issues...)
		}
	}
	
	// 判断整体状态
	if report.InconsistentShards == 0 {
		report.Status = "consistent"
	} else if report.InconsistentShards < report.TotalShards/2 {
		report.Status = "inconsistent"
	} else {
		report.Status = "error"
	}
	
	return report
}

// checkShardConsistency checks consistency for a single shard (primary + replicas)
func (cc *ConsistencyChecker) checkShardConsistency(index, shardID string, shards []ShardInfo) ShardConsistencyReport {
	report := ShardConsistencyReport{
		IsConsistent: true,
		Issues:       []string{},
	}
	
	// 解析 shardID
	var sid int
	fmt.Sscanf(shardID, "%d", &sid)
	report.ShardID = sid
	
	// 分离主分片和副本
	var primary *ShardInfo
	var replicas []ShardInfo
	
	for i := range shards {
		if shards[i].Prirep == "p" {
			primary = &shards[i]
		} else {
			replicas = append(replicas, shards[i])
		}
	}
	
	// 检查是否有主分片
	if primary == nil {
		report.IsConsistent = false
		report.Issues = append(report.Issues, fmt.Sprintf("No primary shard found for shard %s", shardID))
		return report
	}
	
	report.PrimaryNode = primary.Node
	report.PrimaryStoreSize = primary.Store
	
	// 解析主分片文档数
	var primaryDocCount int64
	if primary.Docs != "" {
		fmt.Sscanf(primary.Docs, "%d", &primaryDocCount)
	}
	report.PrimaryDocCount = primaryDocCount
	
	// 检查副本一致性
	for _, replica := range replicas {
		report.ReplicaNodes = append(report.ReplicaNodes, replica.Node)
		report.ReplicaStoreSizes = append(report.ReplicaStoreSizes, replica.Store)
		
		// 解析副本文档数
		var replicaDocCount int64
		if replica.Docs != "" {
			fmt.Sscanf(replica.Docs, "%d", &replicaDocCount)
		}
		report.ReplicaDocCounts = append(report.ReplicaDocCounts, replicaDocCount)
		
		// 检查副本状态
		if replica.State != "STARTED" {
			report.IsConsistent = false
			report.Issues = append(report.Issues, 
				fmt.Sprintf("Replica on %s is not STARTED (state: %s)", replica.Node, replica.State))
			continue
		}
		
		// 检查文档数量一致性
		docCountDiff := abs(primaryDocCount - replicaDocCount)
		if docCountDiff > cc.docCountTolerance {
			report.IsConsistent = false
			report.Issues = append(report.Issues,
				fmt.Sprintf("Doc count mismatch: primary=%d, replica on %s=%d (diff=%d)",
					primaryDocCount, replica.Node, replicaDocCount, docCountDiff))
		}
		
		// 检查存储大小一致性（仅警告，不标记为不一致）
		if primary.Store != replica.Store {
			log.Printf("[ConsistencyChecker] Warning: Store size mismatch for index %s shard %s: primary=%s, replica on %s=%s",
				index, shardID, primary.Store, replica.Node, replica.Store)
		}
	}
	
	// 检查副本数量
	if len(replicas) == 0 {
		report.Issues = append(report.Issues, 
			fmt.Sprintf("No replica shards found for shard %s", shardID))
		// 注意：没有副本不一定是不一致，可能是配置的副本数为0
	}
	
	return report
}

// printSummaryReport prints a summary of consistency check results
func (cc *ConsistencyChecker) printSummaryReport() {
	cc.reportsMutex.RLock()
	defer cc.reportsMutex.RUnlock()
	
	log.Println("========================================")
	log.Println("[ConsistencyChecker] Consistency Check Summary Report")
	log.Println("========================================")
	
	totalIndices := len(cc.reports)
	consistentIndices := 0
	inconsistentIndices := 0
	errorIndices := 0
	
	for index, report := range cc.reports {
		log.Printf("[INDEX] %s: %s (%d/%d shards consistent)",
			index, report.Status, report.ConsistentShards, report.TotalShards)
		
		// 打印问题
		if len(report.Issues) > 0 {
			for _, issue := range report.Issues {
				log.Printf("  ⚠️  %s", issue)
			}
		}
		
		// 统计
		switch report.Status {
		case "consistent":
			consistentIndices++
		case "inconsistent":
			inconsistentIndices++
		case "error":
			errorIndices++
		}
	}
	
	log.Println("========================================")
	log.Printf("[SUMMARY] Total: %d | Consistent: %d | Inconsistent: %d | Error: %d",
		totalIndices, consistentIndices, inconsistentIndices, errorIndices)
	log.Println("========================================")
	
	// 如果有不一致的索引，触发告警
	if inconsistentIndices > 0 || errorIndices > 0 {
		cc.handleInconsistencies()
	}
}

// handleInconsistencies handles detected inconsistencies
func (cc *ConsistencyChecker) handleInconsistencies() {
	cc.reportsMutex.RLock()
	defer cc.reportsMutex.RUnlock()
	
	for index, report := range cc.reports {
		if report.Status == "inconsistent" || report.Status == "error" {
			log.Printf("[ALERT] Inconsistency detected in index: %s", index)
			
			// 记录详细信息
			for _, shardReport := range report.ShardReports {
				if !shardReport.IsConsistent {
					log.Printf("[ALERT] Shard %d issues:", shardReport.ShardID)
					for _, issue := range shardReport.Issues {
						log.Printf("  - %s", issue)
					}
				}
			}
			
			// 这里可以触发自动修复或发送告警
			// 暂时只记录日志
			log.Printf("[ACTION] Please investigate index %s for data inconsistencies", index)
		}
	}
}

// GetConsistencyReport returns the consistency report for an index
func (cc *ConsistencyChecker) GetConsistencyReport(index string) (*ConsistencyReport, error) {
	cc.reportsMutex.RLock()
	defer cc.reportsMutex.RUnlock()
	
	report, exists := cc.reports[index]
	if !exists {
		return nil, fmt.Errorf("no consistency report found for index: %s", index)
	}
	
	return report, nil
}

// GetAllConsistencyReports returns all consistency reports
func (cc *ConsistencyChecker) GetAllConsistencyReports() map[string]*ConsistencyReport {
	cc.reportsMutex.RLock()
	defer cc.reportsMutex.RUnlock()
	
	// 返回副本以避免并发修改
	reports := make(map[string]*ConsistencyReport)
	for k, v := range cc.reports {
		reports[k] = v
	}
	
	return reports
}

// CheckIndexNow performs an immediate consistency check for a specific index
func (cc *ConsistencyChecker) CheckIndexNow(index string) (*ConsistencyReport, error) {
	log.Printf("[ConsistencyChecker] Performing immediate consistency check for index: %s", index)
	
	// 获取该索引的所有分片
	allShards, err := cc.esClient.GetShardAllocation()
	if err != nil {
		return nil, fmt.Errorf("failed to get shard allocation: %v", err)
	}
	
	// 过滤出指定索引的分片
	var indexShards []ShardInfo
	for _, shard := range allShards {
		if shard.Index == index {
			indexShards = append(indexShards, shard)
		}
	}
	
	if len(indexShards) == 0 {
		return nil, fmt.Errorf("no shards found for index: %s", index)
	}
	
	// 执行检查
	report := cc.checkIndexConsistency(index, indexShards)
	
	// 保存报告
	cc.reportsMutex.Lock()
	cc.reports[index] = report
	cc.reportsMutex.Unlock()
	
	return report, nil
}

// SetCheckInterval sets the check interval
func (cc *ConsistencyChecker) SetCheckInterval(interval time.Duration) {
	cc.checkInterval = interval
	log.Printf("[ConsistencyChecker] Check interval set to %v", interval)
	
	// 重启定时器
	if cc.ticker != nil {
		cc.ticker.Stop()
		cc.ticker = time.NewTicker(interval)
	}
}

// SetDocCountTolerance sets the document count tolerance
func (cc *ConsistencyChecker) SetDocCountTolerance(tolerance int64) {
	cc.docCountTolerance = tolerance
	log.Printf("[ConsistencyChecker] Document count tolerance set to %d", tolerance)
}

// abs returns the absolute value of an int64
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
