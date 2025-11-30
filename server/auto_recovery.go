package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// RecoveryAction represents a recovery action
type RecoveryAction struct {
	ID          string    `json:"id"`
	Index       string    `json:"index"`
	ActionType  string    `json:"action_type"` // rebalance, reallocate, resync, verify
	Status      string    `json:"status"`      // pending, running, success, failed, retrying
	Attempts    int       `json:"attempts"`
	MaxRetries  int       `json:"max_retries"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Error       string    `json:"error,omitempty"`
	Description string    `json:"description"`
}

// RecoveryHistory represents the history of recovery actions
type RecoveryHistory struct {
	Actions      []*RecoveryAction
	actionsMutex sync.RWMutex
}

// AutoRecoveryManager manages automatic recovery of ES shards
type AutoRecoveryManager struct {
	esClient            *ESClient
	replicationMonitor  *ReplicationMonitor
	consistencyChecker  *ConsistencyChecker
	ticker              *time.Ticker
	stopChan            chan bool
	history             *RecoveryHistory
	checkInterval       time.Duration
	maxRetries          int
	retryDelay          time.Duration
	enableAutoRecovery  bool
	activeRecoveries    map[string]*RecoveryAction
	recoveriesMutex     sync.RWMutex
}

// NewAutoRecoveryManager creates a new auto recovery manager
func NewAutoRecoveryManager(esClient *ESClient, rm *ReplicationMonitor, cc *ConsistencyChecker) *AutoRecoveryManager {
	return &AutoRecoveryManager{
		esClient:           esClient,
		replicationMonitor: rm,
		consistencyChecker: cc,
		stopChan:           make(chan bool),
		history: &RecoveryHistory{
			Actions: []*RecoveryAction{},
		},
		checkInterval:      1 * time.Minute, // 每分钟检查一次
		maxRetries:         3,
		retryDelay:         30 * time.Second,
		enableAutoRecovery: true,
		activeRecoveries:   make(map[string]*RecoveryAction),
	}
}

// Start starts the auto recovery manager
func (arm *AutoRecoveryManager) Start() {
	log.Println("[AutoRecoveryManager] Starting automatic recovery management...")
	
	arm.ticker = time.NewTicker(arm.checkInterval)
	
	// 立即执行一次检查
	go arm.checkAndRecover()
	
	// 定期检查
	go func() {
		for {
			select {
			case <-arm.ticker.C:
				arm.checkAndRecover()
			case <-arm.stopChan:
				log.Println("[AutoRecoveryManager] Stopping...")
				return
			}
		}
	}()
}

// Stop stops the auto recovery manager
func (arm *AutoRecoveryManager) Stop() {
	if arm.ticker != nil {
		arm.ticker.Stop()
	}
	arm.stopChan <- true
}

// checkAndRecover checks for issues and triggers recovery if needed
func (arm *AutoRecoveryManager) checkAndRecover() {
	if !arm.enableAutoRecovery {
		return
	}
	
	log.Println("[AutoRecoveryManager] Checking for recovery opportunities...")
	
	// 1. 检查副本同步状态
	arm.checkReplicationIssues()
	
	// 2. 检查数据一致性
	arm.checkConsistencyIssues()
	
	// 3. 执行待处理的恢复任务
	arm.processRecoveryQueue()
}

// checkReplicationIssues checks for replication issues
func (arm *AutoRecoveryManager) checkReplicationIssues() {
	statuses := arm.replicationMonitor.GetAllReplicationStatuses()
	
	for index, status := range statuses {
		// 跳过已在恢复中的索引
		if arm.isRecovering(index) {
			continue
		}
		
		switch status.Status {
		case "failed":
			log.Printf("[AutoRecoveryManager] Detected failed replication for index: %s", index)
			arm.triggerRecovery(index, "reallocate", 
				fmt.Sprintf("Replication failed: %.1f%% complete", status.ReplicationProgress))
			
		case "degraded":
			log.Printf("[AutoRecoveryManager] Detected degraded replication for index: %s", index)
			arm.triggerRecovery(index, "reallocate",
				fmt.Sprintf("Degraded: %d unassigned shards", status.UnreplicatedShards))
		}
	}
}

// checkConsistencyIssues checks for consistency issues
func (arm *AutoRecoveryManager) checkConsistencyIssues() {
	reports := arm.consistencyChecker.GetAllConsistencyReports()
	
	for index, report := range reports {
		// 跳过已在恢复中的索引
		if arm.isRecovering(index) {
			continue
		}
		
		switch report.Status {
		case "inconsistent":
			log.Printf("[AutoRecoveryManager] Detected data inconsistency for index: %s", index)
			arm.triggerRecovery(index, "resync",
				fmt.Sprintf("Inconsistent: %d/%d shards", report.InconsistentShards, report.TotalShards))
			
		case "error":
			log.Printf("[AutoRecoveryManager] Detected severe inconsistency for index: %s", index)
			arm.triggerRecovery(index, "resync",
				fmt.Sprintf("Severe: %d/%d shards inconsistent", report.InconsistentShards, report.TotalShards))
		}
	}
}

// triggerRecovery triggers a recovery action
func (arm *AutoRecoveryManager) triggerRecovery(index, actionType, description string) {
	action := &RecoveryAction{
		ID:          fmt.Sprintf("%s-%s-%d", index, actionType, time.Now().Unix()),
		Index:       index,
		ActionType:  actionType,
		Status:      "pending",
		Attempts:    0,
		MaxRetries:  arm.maxRetries,
		StartTime:   time.Now(),
		Description: description,
	}
	
	log.Printf("[AutoRecoveryManager] Triggering recovery: %s for index %s", actionType, index)
	
	// 添加到活跃恢复列表
	arm.recoveriesMutex.Lock()
	arm.activeRecoveries[index] = action
	arm.recoveriesMutex.Unlock()
	
	// 添加到历史记录
	arm.history.actionsMutex.Lock()
	arm.history.Actions = append(arm.history.Actions, action)
	arm.history.actionsMutex.Unlock()
	
	// 异步执行恢复
	go arm.executeRecovery(action)
}

// executeRecovery executes a recovery action
func (arm *AutoRecoveryManager) executeRecovery(action *RecoveryAction) {
	action.Status = "running"
	action.Attempts++
	
	log.Printf("[AutoRecoveryManager] Executing recovery %s (attempt %d/%d)", 
		action.ID, action.Attempts, action.MaxRetries)
	
	var err error
	
	switch action.ActionType {
	case "reallocate":
		err = arm.performReallocation(action.Index)
	case "resync":
		err = arm.performResync(action.Index)
	case "rebalance":
		err = arm.performRebalance(action.Index)
	case "verify":
		err = arm.performVerification(action.Index)
	default:
		err = fmt.Errorf("unknown action type: %s", action.ActionType)
	}
	
	if err != nil {
		action.Error = err.Error()
		
		// 判断是否需要重试
		if action.Attempts < action.MaxRetries {
			action.Status = "retrying"
			log.Printf("[AutoRecoveryManager] Recovery failed, will retry: %s (error: %v)", action.ID, err)
			
			// 延迟后重试
			time.Sleep(arm.retryDelay)
			go arm.executeRecovery(action)
		} else {
			action.Status = "failed"
			action.EndTime = time.Now()
			log.Printf("[AutoRecoveryManager] Recovery failed after %d attempts: %s (error: %v)", 
				action.Attempts, action.ID, err)
			
			// 从活跃列表移除
			arm.recoveriesMutex.Lock()
			delete(arm.activeRecoveries, action.Index)
			arm.recoveriesMutex.Unlock()
		}
	} else {
		action.Status = "success"
		action.EndTime = time.Now()
		log.Printf("[AutoRecoveryManager] Recovery succeeded: %s (attempts: %d, duration: %v)",
			action.ID, action.Attempts, action.EndTime.Sub(action.StartTime))
		
		// 从活跃列表移除
		arm.recoveriesMutex.Lock()
		delete(arm.activeRecoveries, action.Index)
		arm.recoveriesMutex.Unlock()
		
		// 触发验证
		go arm.scheduleVerification(action.Index)
	}
}

// performReallocation performs shard reallocation
func (arm *AutoRecoveryManager) performReallocation(index string) error {
	log.Printf("[AutoRecoveryManager] Performing reallocation for index: %s", index)
	
	// 启用分片分配
	settings := map[string]interface{}{
		"transient": map[string]interface{}{
			"cluster.routing.allocation.enable": "all",
			"cluster.routing.rebalance.enable":  "all",
		},
	}
	
	err := arm.esClient.UpdateClusterSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to enable allocation: %v", err)
	}
	
	// 等待一段时间让分配生效
	time.Sleep(10 * time.Second)
	
	// 检查分配状态
	shards, err := arm.esClient.GetShardAllocation()
	if err != nil {
		return fmt.Errorf("failed to check allocation status: %v", err)
	}
	
	// 统计未分配的分片
	unassignedCount := 0
	for _, shard := range shards {
		if shard.Index == index && shard.State == "UNASSIGNED" {
			unassignedCount++
		}
	}
	
	if unassignedCount > 0 {
		return fmt.Errorf("still have %d unassigned shards", unassignedCount)
	}
	
	return nil
}

// performResync performs data resynchronization
func (arm *AutoRecoveryManager) performResync(index string) error {
	log.Printf("[AutoRecoveryManager] Performing resync for index: %s", index)
	
	// 触发分片刷新
	settings := map[string]interface{}{
		"transient": map[string]interface{}{
			"indices.recovery.max_bytes_per_sec": "100mb",
			"cluster.routing.allocation.node_concurrent_recoveries": 4,
		},
	}
	
	err := arm.esClient.UpdateClusterSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to update recovery settings: %v", err)
	}
	
	// 等待同步完成
	maxWait := 5 * time.Minute
	checkInterval := 10 * time.Second
	deadline := time.Now().Add(maxWait)
	
	for time.Now().Before(deadline) {
		// 检查一致性
		report, err := arm.consistencyChecker.CheckIndexNow(index)
		if err != nil {
			log.Printf("[AutoRecoveryManager] Error checking consistency: %v", err)
			time.Sleep(checkInterval)
			continue
		}
		
		if report.Status == "consistent" {
			log.Printf("[AutoRecoveryManager] Resync completed for index: %s", index)
			return nil
		}
		
		log.Printf("[AutoRecoveryManager] Resync in progress for %s: %d/%d shards consistent",
			index, report.ConsistentShards, report.TotalShards)
		
		time.Sleep(checkInterval)
	}
	
	return fmt.Errorf("resync timeout after %v", maxWait)
}

// performRebalance performs shard rebalancing
func (arm *AutoRecoveryManager) performRebalance(index string) error {
	log.Printf("[AutoRecoveryManager] Performing rebalance for index: %s", index)
	
	settings := map[string]interface{}{
		"transient": map[string]interface{}{
			"cluster.routing.rebalance.enable": "all",
			"cluster.routing.allocation.balance.shard": 0.45,
			"cluster.routing.allocation.balance.index": 0.55,
		},
	}
	
	err := arm.esClient.UpdateClusterSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to trigger rebalance: %v", err)
	}
	
	// 等待重平衡完成
	time.Sleep(30 * time.Second)
	
	return nil
}

// performVerification performs verification after recovery
func (arm *AutoRecoveryManager) performVerification(index string) error {
	log.Printf("[AutoRecoveryManager] Performing verification for index: %s", index)
	
	// 检查副本状态
	status, err := arm.replicationMonitor.GetReplicationStatus(index)
	if err == nil {
		if status.Status != "healthy" {
			return fmt.Errorf("replication status is %s", status.Status)
		}
	}
	
	// 检查数据一致性
	report, err := arm.consistencyChecker.CheckIndexNow(index)
	if err != nil {
		return fmt.Errorf("failed to check consistency: %v", err)
	}
	
	if report.Status != "consistent" {
		return fmt.Errorf("consistency check failed: %s", report.Status)
	}
	
	log.Printf("[AutoRecoveryManager] Verification passed for index: %s", index)
	return nil
}

// scheduleVerification schedules a verification after recovery
func (arm *AutoRecoveryManager) scheduleVerification(index string) {
	// 等待一段时间后验证
	time.Sleep(1 * time.Minute)
	
	action := &RecoveryAction{
		ID:          fmt.Sprintf("%s-verify-%d", index, time.Now().Unix()),
		Index:       index,
		ActionType:  "verify",
		Status:      "pending",
		Attempts:    0,
		MaxRetries:  1,
		StartTime:   time.Now(),
		Description: "Post-recovery verification",
	}
	
	log.Printf("[AutoRecoveryManager] Scheduling verification for index: %s", index)
	
	arm.history.actionsMutex.Lock()
	arm.history.Actions = append(arm.history.Actions, action)
	arm.history.actionsMutex.Unlock()
	
	arm.executeRecovery(action)
}

// processRecoveryQueue processes pending recovery actions
func (arm *AutoRecoveryManager) processRecoveryQueue() {
	// 这个方法可以用于处理排队的恢复任务
	// 目前恢复任务是立即异步执行的，这里可以扩展为队列模式
}

// isRecovering checks if an index is currently being recovered
func (arm *AutoRecoveryManager) isRecovering(index string) bool {
	arm.recoveriesMutex.RLock()
	defer arm.recoveriesMutex.RUnlock()
	
	action, exists := arm.activeRecoveries[index]
	if !exists {
		return false
	}
	
	// 检查是否是活跃状态
	return action.Status == "running" || action.Status == "retrying" || action.Status == "pending"
}

// GetRecoveryHistory returns the recovery history
func (arm *AutoRecoveryManager) GetRecoveryHistory() []*RecoveryAction {
	arm.history.actionsMutex.RLock()
	defer arm.history.actionsMutex.RUnlock()
	
	// 返回副本
	history := make([]*RecoveryAction, len(arm.history.Actions))
	copy(history, arm.history.Actions)
	
	return history
}

// GetActiveRecoveries returns currently active recoveries
func (arm *AutoRecoveryManager) GetActiveRecoveries() []*RecoveryAction {
	arm.recoveriesMutex.RLock()
	defer arm.recoveriesMutex.RUnlock()
	
	active := make([]*RecoveryAction, 0, len(arm.activeRecoveries))
	for _, action := range arm.activeRecoveries {
		active = append(active, action)
	}
	
	return active
}

// EnableAutoRecovery enables automatic recovery
func (arm *AutoRecoveryManager) EnableAutoRecovery() {
	arm.enableAutoRecovery = true
	log.Println("[AutoRecoveryManager] Auto recovery enabled")
}

// DisableAutoRecovery disables automatic recovery
func (arm *AutoRecoveryManager) DisableAutoRecovery() {
	arm.enableAutoRecovery = false
	log.Println("[AutoRecoveryManager] Auto recovery disabled")
}

// SetMaxRetries sets the maximum number of retries
func (arm *AutoRecoveryManager) SetMaxRetries(maxRetries int) {
	arm.maxRetries = maxRetries
	log.Printf("[AutoRecoveryManager] Max retries set to %d", maxRetries)
}

// SetRetryDelay sets the delay between retries
func (arm *AutoRecoveryManager) SetRetryDelay(delay time.Duration) {
	arm.retryDelay = delay
	log.Printf("[AutoRecoveryManager] Retry delay set to %v", delay)
}

// SetCheckInterval sets the check interval
func (arm *AutoRecoveryManager) SetCheckInterval(interval time.Duration) {
	arm.checkInterval = interval
	log.Printf("[AutoRecoveryManager] Check interval set to %v", interval)
	
	// 重启定时器
	if arm.ticker != nil {
		arm.ticker.Stop()
		arm.ticker = time.NewTicker(interval)
	}
}
