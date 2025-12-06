package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"es-serverless-manager/internal/model"
)

// AutoscalerService handles automatic scaling of Elasticsearch clusters
// AutoscalerService 处理 Elasticsearch 集群的自动扩缩容
type AutoscalerService struct {
	config *model.AutoscalerConfig
	ticker *time.Ticker
	// Map to store last scaling time for each namespace
	// 存储每个命名空间的最后一次扩缩容时间的映射
	lastScalingTime map[string]time.Time
	// Map to store historical metrics for each namespace
	// 存储每个命名空间的历史指标的映射
	historicalMetrics map[string]*model.HistoricalMetrics
	metadataService   *MetadataService
	mu                sync.RWMutex
	stopChan          chan struct{}
}

// NewAutoscalerService creates a new autoscaler with default configuration
// NewAutoscalerService 创建一个具有默认配置的新自动扩缩容服务
func NewAutoscalerService(metadataService *MetadataService) *AutoscalerService {
	config := &model.AutoscalerConfig{
		HighCPUThreshold:    70.0,
		LowCPUThreshold:     30.0,
		HighMemoryThreshold: 70.0,
		LowMemoryThreshold:  30.0,
		HighQPSThreshold:    2000.0,
		LowQPSThreshold:     500.0,
		HighDiskThreshold:   80.0,
		LowDiskThreshold:    20.0,
		ScaleUpFactor:       1.5,
		ScaleDownFactor:     0.5,
		MinReplicas:         1,
		MaxReplicas:         10,
		ScaleUpCooldown:     300, // 5 minutes cooldown after scaling up
		ScaleDownCooldown:   600, // 10 minutes cooldown after scaling down
		ScalingPolicies:     make(map[string]model.ScalingPolicy),
	}

	return &AutoscalerService{
		config:            config,
		lastScalingTime:   make(map[string]time.Time),
		historicalMetrics: make(map[string]*model.HistoricalMetrics),
		metadataService:   metadataService,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the autoscaling loop
// Start 启动自动扩缩容循环
func (a *AutoscalerService) Start() {
	a.ticker = time.NewTicker(60 * time.Second) // Check every minute
	go func() {
		for {
			select {
			case <-a.ticker.C:
				a.checkAndScale()
			case <-a.stopChan:
				a.ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the autoscaling loop
// Stop 停止自动扩缩容循环
func (a *AutoscalerService) Stop() {
	close(a.stopChan)
}

// checkAndScale checks metrics and scales clusters if needed
// checkAndScale 检查指标并在需要时扩缩容集群
func (a *AutoscalerService) checkAndScale() {
	// Get list of namespaces with ES clusters from metadata service
	// 从元数据服务获取具有 ES 集群的命名空间列表
	deployments, err := a.metadataService.ListDeploymentStatus()
	if err != nil {
		// Fallback to kubectl if metadata service fails
		// 如果元数据服务失败，回退到 kubectl
		log.Printf("Error getting deployments from metadata service: %v, falling back to kubectl", err)
		cmd := exec.Command("kubectl", "get", "namespaces", "-l", "es-cluster=true", "-o", "jsonpath={.items[*].metadata.name}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error getting namespaces: %v", err)
			return
		}

		namespaces := parseNamespaces(string(out))
		for _, ns := range namespaces {
			a.scaleNamespace(ns, "")
		}
		return
	}

	// Scale each deployment
	// 对每个部署进行扩缩容
	for _, deployment := range deployments {
		a.scaleNamespace(deployment.Namespace, deployment.User)
	}
}

// scaleNamespace scales a specific namespace based on its metrics and user policy
// scaleNamespace 根据指标和用户策略扩缩容特定命名空间
func (a *AutoscalerService) scaleNamespace(namespace string, userID string) {
	// Get current replicas
	// 获取当前副本数
	currentReplicas, err := a.getCurrentReplicas(namespace)
	if err != nil {
		log.Printf("Error getting current replicas for namespace %s: %v", namespace, err)
		return
	}

	// Get metrics for namespace
	// 获取命名空间的指标
	metrics, err := a.getMetricsForNamespace(namespace)
	if err != nil {
		log.Printf("Error getting metrics for namespace %s: %v", namespace, err)
		return
	}

	// Update historical metrics
	// 更新历史指标
	a.updateHistoricalMetrics(namespace, metrics)

	// Get trend analysis
	// 获取趋势分析
	cpuTrend, memoryTrend, diskTrend, qpsTrend := a.getTrendAnalysis(namespace)

	// Log trend analysis
	// 记录趋势分析
	log.Printf("Namespace %s trends - CPU: %.2f, Memory: %.2f, Disk: %.2f, QPS: %.2f", namespace, cpuTrend, memoryTrend, diskTrend, qpsTrend)

	// Get user-specific scaling policy
	// 获取用户特定的扩缩容策略
	policy := a.getUserScalingPolicy(userID)

	// Check if auto-scaling is enabled for this user
	// 检查是否为此用户启用了自动扩缩容
	if !policy.EnableAutoScaleUp && !policy.EnableAutoScaleDown {
		log.Printf("Auto-scaling disabled for user %s in namespace %s", userID, namespace)
		return
	}

	// Adjust scaling decision based on trends
	// 根据趋势调整扩缩容决策
	adjustedMetrics := a.adjustMetricsBasedOnTrends(metrics, cpuTrend, memoryTrend, diskTrend, qpsTrend)

	// Determine if we need to scale based on user policy and metrics
	// 根据用户策略和指标确定是否需要扩缩容
	newReplicas := a.calculateNewReplicas(currentReplicas, adjustedMetrics, policy)
	if newReplicas != currentReplicas {
		// Check if we're still in cooldown period
		// 检查是否仍处于冷却期
		if a.isInCooldown(namespace, currentReplicas, newReplicas) {
			log.Printf("Skipping scaling for namespace %s due to cooldown period", namespace)
			return
		}

		// Check user-specific limits
		// 检查用户特定的限制
		if newReplicas > policy.MaxReplicas {
			newReplicas = policy.MaxReplicas
		}
		if newReplicas < policy.MinReplicas {
			newReplicas = policy.MinReplicas
		}

		// Check tenant quota before scaling (only for scale up)
		// 扩容前检查租户配额（仅针对扩容）
		if newReplicas > currentReplicas && userID != "" {
			hasQuota, quota, err := a.metadataService.CheckTenantQuota(userID)
			if err != nil {
				log.Printf("Warning: Failed to check tenant quota for user %s: %v", userID, err)
			} else if !hasQuota {
				log.Printf("Tenant quota exceeded for user %s. Max indices: %d, Current indices: %d",
					userID, quota.MaxIndices, quota.CurrentIndices)
				return
			}
		}

		err := a.scaleCluster(namespace, newReplicas)
		if err != nil {
			log.Printf("Error scaling cluster in namespace %s: %v", namespace, err)
		} else {
			log.Printf("Scaled cluster in namespace %s from %d to %d replicas", namespace, currentReplicas, newReplicas)

			// Update deployment status in metadata service
			// 更新元数据服务中的部署状态
			deployment, err := a.metadataService.GetDeploymentStatus(namespace)
			if err == nil {
				deployment.Replicas = newReplicas
				deployment.Status = "scaling"
				deployment.UpdatedAt = time.Now()
				a.metadataService.SaveDeploymentStatus(deployment)
			}

			// Update last scaling time
			// 更新最后一次扩缩容时间
			a.mu.Lock()
			a.lastScalingTime[namespace] = time.Now()
			a.mu.Unlock()

			// Update tenant quota usage if scaled up
			// 如果扩容，更新租户配额使用情况
			if newReplicas > currentReplicas && userID != "" {
				// Note: In a real implementation, you might want to track resource usage more precisely
				// For now, we're just checking quota, not updating usage for replica scaling
				// 注意：在实际实现中，您可能希望更精确地跟踪资源使用情况
				// 目前，我们只是检查配额，而不是更新副本扩容的使用情况
				log.Printf("Updated tenant quota check for user %s after scaling", userID)
			}
		}
	}
}

// getUserScalingPolicy gets the scaling policy for a user
// getUserScalingPolicy 获取用户的扩缩容策略
func (a *AutoscalerService) getUserScalingPolicy(userID string) model.ScalingPolicy {
	if userID == "" {
		// Return default policy if no user ID
		// 如果没有用户 ID，返回默认策略
		return model.ScalingPolicy{
			UserID:              "",
			EnableAutoScaleUp:   true,
			EnableAutoScaleDown: true,
			ScaleUpThreshold:    a.config.HighCPUThreshold,
			ScaleDownThreshold:  a.config.LowCPUThreshold,
			MaxReplicas:         a.config.MaxReplicas,
			MinReplicas:         a.config.MinReplicas,
		}
	}

	a.mu.RLock()
	defer a.mu.RUnlock()
	// Check if user has a specific policy
	// 检查用户是否有特定策略
	if policy, exists := a.config.ScalingPolicies[userID]; exists {
		return policy
	}

	// Return default policy
	// 返回默认策略
	return model.ScalingPolicy{
		UserID:              userID,
		EnableAutoScaleUp:   true,
		EnableAutoScaleDown: true,
		ScaleUpThreshold:    a.config.HighCPUThreshold,
		ScaleDownThreshold:  a.config.LowCPUThreshold,
		MaxReplicas:         a.config.MaxReplicas,
		MinReplicas:         a.config.MinReplicas,
	}
}

// getCurrentReplicas gets the current number of replicas for a namespace
// getCurrentReplicas 获取命名空间的当前副本数
func (a *AutoscalerService) getCurrentReplicas(namespace string) (int, error) {
	cmd := exec.Command("kubectl", "-n", namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.spec.replicas}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	replicas, err := strconv.Atoi(string(out))
	if err != nil {
		return 0, fmt.Errorf("error parsing replicas: %v", err)
	}

	return replicas, nil
}

// getMetricsForNamespace gets the latest metrics for a namespace
// getMetricsForNamespace 获取命名空间的最新指标
func (a *AutoscalerService) getMetricsForNamespace(namespace string) (*model.Metrics, error) {
	// Try to get metrics from metadata service first
	// 首先尝试从元数据服务获取指标
	metrics, err := a.metadataService.GetLatestMetrics(namespace)
	if err != nil {
		// Fallback to file if metadata service fails
		// 如果元数据服务失败，回退到文件
		filename := fmt.Sprintf("server/metrics_%s.json", namespace)

		// Try to read metrics from file
		// 尝试从文件读取指标
		file, err := exec.Command("cat", filename).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("error reading metrics file: %v", err)
		}

		var fileMetrics model.Metrics
		if err := json.Unmarshal(file, &fileMetrics); err != nil {
			return nil, fmt.Errorf("error unmarshaling metrics: %v", err)
		}

		return &fileMetrics, nil
	}

	return metrics, nil
}

// calculateNewReplicas calculates the new number of replicas based on metrics and user policy
// calculateNewReplicas 根据指标和用户策略计算新的副本数
func (a *AutoscalerService) calculateNewReplicas(currentReplicas int, metrics *model.Metrics, policy model.ScalingPolicy) int {
	// Check if we need to scale up
	// 检查是否需要扩容
	if policy.EnableAutoScaleUp && a.shouldScaleUp(metrics, policy) {
		newReplicas := int(float64(currentReplicas) * a.config.ScaleUpFactor)
		if newReplicas > a.config.MaxReplicas {
			newReplicas = a.config.MaxReplicas
		}
		if newReplicas <= currentReplicas {
			newReplicas = currentReplicas + 1
		}
		return newReplicas
	}

	// Check if we need to scale down
	// 检查是否需要缩容
	if policy.EnableAutoScaleDown && a.shouldScaleDown(metrics, policy) {
		newReplicas := int(float64(currentReplicas) * a.config.ScaleDownFactor)
		if newReplicas < a.config.MinReplicas {
			newReplicas = a.config.MinReplicas
		}
		if newReplicas >= currentReplicas {
			newReplicas = currentReplicas - 1
		}
		return newReplicas
	}

	// No scaling needed
	// 不需要扩缩容
	return currentReplicas
}

// shouldScaleUp determines if we should scale up based on metrics and user policy
// shouldScaleUp 确定是否应根据指标和用户策略进行扩容
func (a *AutoscalerService) shouldScaleUp(metrics *model.Metrics, policy model.ScalingPolicy) bool {
	// Scale up if CPU, memory, disk, or QPS is above high threshold
	// 如果 CPU、内存、磁盘或 QPS 高于高阈值，则扩容
	return metrics.CPUUsage > policy.ScaleUpThreshold ||
		metrics.MemoryUsage > policy.ScaleUpThreshold ||
		metrics.DiskUsage > policy.ScaleUpThreshold ||
		metrics.QPS > policy.ScaleUpThreshold
}

// shouldScaleDown determines if we should scale down based on metrics and user policy
// shouldScaleDown 确定是否应根据指标和用户策略进行缩容
func (a *AutoscalerService) shouldScaleDown(metrics *model.Metrics, policy model.ScalingPolicy) bool {
	// Scale down if CPU, memory, disk, and QPS are below low threshold
	// 如果 CPU、内存、磁盘和 QPS 低于低阈值，则缩容
	return metrics.CPUUsage < policy.ScaleDownThreshold &&
		metrics.MemoryUsage < policy.ScaleDownThreshold &&
		metrics.DiskUsage < policy.ScaleDownThreshold &&
		metrics.QPS < policy.ScaleDownThreshold
}

// isInCooldown checks if the namespace is still in cooldown period after scaling
// isInCooldown 检查命名空间在扩缩容后是否仍处于冷却期
func (a *AutoscalerService) isInCooldown(namespace string, currentReplicas, newReplicas int) bool {
	a.mu.RLock()
	lastTime, exists := a.lastScalingTime[namespace]
	a.mu.RUnlock()
	if !exists {
		return false // No previous scaling, not in cooldown
	}

	// Determine cooldown period based on scaling direction
	// 根据扩缩容方向确定冷却期
	var cooldownSeconds int
	if newReplicas > currentReplicas {
		// Scale up
		// 扩容
		cooldownSeconds = a.config.ScaleUpCooldown
	} else {
		// Scale down
		// 缩容
		cooldownSeconds = a.config.ScaleDownCooldown
	}

	// Check if cooldown period has passed
	// 检查冷却期是否已过
	cooldownDuration := time.Duration(cooldownSeconds) * time.Second
	return time.Since(lastTime) < cooldownDuration
}

// updateHistoricalMetrics updates the historical metrics for a namespace
// updateHistoricalMetrics 更新命名空间的历史指标
func (a *AutoscalerService) updateHistoricalMetrics(namespace string, metrics *model.Metrics) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Initialize historical metrics for this namespace if not exists
	// 如果不存在，则初始化此命名空间的历史指标
	if _, exists := a.historicalMetrics[namespace]; !exists {
		a.historicalMetrics[namespace] = &model.HistoricalMetrics{
			Metrics: make([]model.Metrics, 0),
			Window:  5, // Keep last 5 metrics for trend analysis
		}
	}

	// Add new metrics
	// 添加新指标
	hm := a.historicalMetrics[namespace]
	hm.Metrics = append(hm.Metrics, *metrics)

	// Trim to window size if needed
	// 如果需要，修剪到窗口大小
	if len(hm.Metrics) > hm.Window {
		hm.Metrics = hm.Metrics[len(hm.Metrics)-hm.Window:]
	}
}

// adjustMetricsBasedOnTrends adjusts metrics based on trend analysis to make more informed scaling decisions
// adjustMetricsBasedOnTrends 根据趋势分析调整指标，以做出更明智的扩缩容决策
func (a *AutoscalerService) adjustMetricsBasedOnTrends(metrics *model.Metrics, cpuTrend, memoryTrend, diskTrend, qpsTrend float64) *model.Metrics {
	// Create a copy of the metrics to avoid modifying the original
	// 创建指标的副本以避免修改原始指标
	adjusted := &model.Metrics{
		CPUUsage:    metrics.CPUUsage,
		MemoryUsage: metrics.MemoryUsage,
		DiskUsage:   metrics.DiskUsage,
		QPS:         metrics.QPS,
	}

	// Apply trend-based adjustments
	// If there's a positive trend, we might want to scale up sooner
	// If there's a negative trend, we might want to scale down later
	// 应用基于趋势的调整
	// 如果有正趋势，我们可能希望更早扩容
	// 如果有负趋势，我们可能希望更晚缩容

	// CPU adjustment
	// CPU 调整
	if cpuTrend > 0.5 { // Increasing trend
		adjusted.CPUUsage *= 1.1 // Increase weight for scaling decision
	} else if cpuTrend < -0.5 { // Decreasing trend
		adjusted.CPUUsage *= 0.9 // Decrease weight for scaling decision
	}

	// Memory adjustment
	// 内存调整
	if memoryTrend > 0.5 { // Increasing trend
		adjusted.MemoryUsage *= 1.1
	} else if memoryTrend < -0.5 { // Decreasing trend
		adjusted.MemoryUsage *= 0.9
	}

	// Disk adjustment
	// 磁盘调整
	if diskTrend > 0.5 { // Increasing trend
		adjusted.DiskUsage *= 1.1
	} else if diskTrend < -0.5 { // Decreasing trend
		adjusted.DiskUsage *= 0.9
	}

	// QPS adjustment
	// QPS 调整
	if qpsTrend > 10 { // Increasing trend
		adjusted.QPS *= 1.1
	} else if qpsTrend < -10 { // Decreasing trend
		adjusted.QPS *= 0.9
	}

	return adjusted
}

// getTrendAnalysis analyzes the trend of metrics over time
// getTrendAnalysis 分析指标随时间的趋势
func (a *AutoscalerService) getTrendAnalysis(namespace string) (cpuTrend, memoryTrend, diskTrend, qpsTrend float64) {
	a.mu.RLock()
	hm, exists := a.historicalMetrics[namespace]
	a.mu.RUnlock()

	// Return 0 trends if no historical data
	// 如果没有历史数据，返回 0 趋势
	if !exists || len(hm.Metrics) < 2 {
		return 0.0, 0.0, 0.0, 0.0
	}

	// Calculate average trends
	// 计算平均趋势
	count := len(hm.Metrics) - 1
	if count <= 0 {
		return 0.0, 0.0, 0.0, 0.0
	}

	for i := 1; i < len(hm.Metrics); i++ {
		cpuTrend += (hm.Metrics[i].CPUUsage - hm.Metrics[i-1].CPUUsage)
		memoryTrend += (hm.Metrics[i].MemoryUsage - hm.Metrics[i-1].MemoryUsage)
		diskTrend += (hm.Metrics[i].DiskUsage - hm.Metrics[i-1].DiskUsage)
		qpsTrend += (hm.Metrics[i].QPS - hm.Metrics[i-1].QPS)
	}

	cpuTrend /= float64(count)
	memoryTrend /= float64(count)
	diskTrend /= float64(count)
	qpsTrend /= float64(count)

	return cpuTrend, memoryTrend, diskTrend, qpsTrend
}

// scaleCluster scales the Elasticsearch cluster in a namespace
// scaleCluster 扩缩容命名空间中的 Elasticsearch 集群
func (a *AutoscalerService) scaleCluster(namespace string, replicas int) error {
	cmd := exec.Command("kubectl", "-n", namespace, "scale", "sts/elasticsearch", "--replicas", strconv.Itoa(replicas))
	_, err := cmd.CombinedOutput()
	return err
}

// SetUserScalingPolicy sets a user-specific scaling policy
// SetUserScalingPolicy 设置用户特定的扩缩容策略
func (a *AutoscalerService) SetUserScalingPolicy(policy model.ScalingPolicy) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config.ScalingPolicies[policy.UserID] = policy
}

// GetUserScalingPolicy gets a user-specific scaling policy
// GetUserScalingPolicy 获取用户特定的扩缩容策略
func (a *AutoscalerService) GetUserScalingPolicy(userID string) (model.ScalingPolicy, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	policy, exists := a.config.ScalingPolicies[userID]
	return policy, exists
}

// RemoveUserScalingPolicy removes a user-specific scaling policy
// RemoveUserScalingPolicy 删除用户特定的扩缩容策略
func (a *AutoscalerService) RemoveUserScalingPolicy(userID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.config.ScalingPolicies, userID)
}

// parseNamespaces parses namespace names from kubectl output
// parseNamespaces 从 kubectl 输出中解析命名空间名称
func parseNamespaces(output string) []string {
	// Split by whitespace and filter out empty strings
	// 按空白字符分割并过滤掉空字符串
	names := []string{}
	for _, name := range strings.Fields(output) {
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}
