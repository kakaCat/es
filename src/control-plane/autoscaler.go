package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// AutoscalerConfig holds the configuration for autoscaling
type AutoscalerConfig struct {
	// CPU thresholds
	HighCPUThreshold    float64 `json:"high_cpu_threshold"`     // Percentage (0-100)
	LowCPUThreshold     float64 `json:"low_cpu_threshold"`      // Percentage (0-100)
	
	// Memory thresholds
	HighMemoryThreshold float64 `json:"high_memory_threshold"`  // Percentage (0-100)
	LowMemoryThreshold  float64 `json:"low_memory_threshold"`   // Percentage (0-100)
	
	// QPS thresholds
	HighQPSThreshold    float64 `json:"high_qps_threshold"`
	LowQPSThreshold     float64 `json:"low_qps_threshold"`
	
	// Disk thresholds
	HighDiskThreshold   float64 `json:"high_disk_threshold"`    // Percentage (0-100)
	LowDiskThreshold    float64 `json:"low_disk_threshold"`     // Percentage (0-100)
	
	// Scaling factors
	ScaleUpFactor       float64 `json:"scale_up_factor"`        // Multiplier for scale up (e.g., 1.5)
	ScaleDownFactor     float64 `json:"scale_down_factor"`      // Multiplier for scale down (e.g., 0.5)
	
	// Limits
	MinReplicas         int     `json:"min_replicas"`
	MaxReplicas         int     `json:"max_replicas"`
	
	// Cooldown period in seconds
	ScaleUpCooldown     int     `json:"scale_up_cooldown"`      // Cooldown period after scaling up
	ScaleDownCooldown    int     `json:"scale_down_cooldown"`    // Cooldown period after scaling down
	
	// User-specific scaling policies
	ScalingPolicies     map[string]ScalingPolicy `json:"scaling_policies"`
}

// ScalingPolicy holds user-specific scaling policies
type ScalingPolicy struct {
	UserID              string  `json:"user_id"`
	EnableAutoScaleUp   bool    `json:"enable_auto_scale_up"`
	EnableAutoScaleDown bool    `json:"enable_auto_scale_down"`
	ScaleUpThreshold    float64 `json:"scale_up_threshold"`
	ScaleDownThreshold  float64 `json:"scale_down_threshold"`
	MaxReplicas         int     `json:"max_replicas"`
	MinReplicas         int     `json:"min_replicas"`
}

// HistoricalMetrics stores historical metrics for trend analysis
type HistoricalMetrics struct {
	Metrics []Metrics `json:"metrics"`
	Window  int       `json:"window"` // Number of metrics to keep for trend analysis
}

// Autoscaler handles automatic scaling of Elasticsearch clusters
type Autoscaler struct {
	config *AutoscalerConfig
	ticker *time.Ticker
	// Map to store last scaling time for each namespace
	lastScalingTime map[string]time.Time
	// Map to store historical metrics for each namespace
	historicalMetrics map[string]*HistoricalMetrics
}

// NewAutoscaler creates a new autoscaler with default configuration
func NewAutoscaler() *Autoscaler {
	config := &AutoscalerConfig{
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
		ScaleUpCooldown:     300,  // 5 minutes cooldown after scaling up
		ScaleDownCooldown:    600,  // 10 minutes cooldown after scaling down
		ScalingPolicies:     make(map[string]ScalingPolicy),
	}
	
	return &Autoscaler{
		config: config,
		ticker: time.NewTicker(60 * time.Second), // Check every minute
		lastScalingTime: make(map[string]time.Time),
		historicalMetrics: make(map[string]*HistoricalMetrics),
	}
}

// Start begins the autoscaling loop
func (a *Autoscaler) Start() {
	go func() {
		for range a.ticker.C {
			a.checkAndScale()
		}
	}()
}

// Stop stops the autoscaling loop
func (a *Autoscaler) Stop() {
	a.ticker.Stop()
}

// checkAndScale checks metrics and scales clusters if needed
func (a *Autoscaler) checkAndScale() {
	// Get list of namespaces with ES clusters from metadata service
	deployments, err := metadataService.ListDeploymentStatus()
	if err != nil {
		// Fallback to kubectl if metadata service fails
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
	for _, deployment := range deployments {
		a.scaleNamespace(deployment.Namespace, deployment.User)
	}
}

// scaleNamespace scales a specific namespace based on its metrics and user policy
func (a *Autoscaler) scaleNamespace(namespace string, userID string) {
	// Get current replicas
	currentReplicas, err := a.getCurrentReplicas(namespace)
	if err != nil {
		log.Printf("Error getting current replicas for namespace %s: %v", namespace, err)
		return
	}
	
	// Get metrics for namespace
	metrics, err := a.getMetricsForNamespace(namespace)
	if err != nil {
		log.Printf("Error getting metrics for namespace %s: %v", namespace, err)
		return
	}
	
	// Update historical metrics
	a.updateHistoricalMetrics(namespace, metrics)
	
	// Get trend analysis
	cpuTrend, memoryTrend, diskTrend, qpsTrend := a.getTrendAnalysis(namespace)
	
	// Log trend analysis
	log.Printf("Namespace %s trends - CPU: %.2f, Memory: %.2f, Disk: %.2f, QPS: %.2f", namespace, cpuTrend, memoryTrend, diskTrend, qpsTrend)
	
	// Get user-specific scaling policy
	policy := a.getUserScalingPolicy(userID)
	
	// Check if auto-scaling is enabled for this user
	if !policy.EnableAutoScaleUp && !policy.EnableAutoScaleDown {
		log.Printf("Auto-scaling disabled for user %s in namespace %s", userID, namespace)
		return
	}
	
	// Adjust scaling decision based on trends
	adjustedMetrics := a.adjustMetricsBasedOnTrends(metrics, cpuTrend, memoryTrend, diskTrend, qpsTrend)
	
	// Determine if we need to scale based on user policy and metrics
	newReplicas := a.calculateNewReplicas(currentReplicas, adjustedMetrics, policy)
	if newReplicas != currentReplicas {
		// Check if we're still in cooldown period
		if a.isInCooldown(namespace, currentReplicas, newReplicas) {
			log.Printf("Skipping scaling for namespace %s due to cooldown period", namespace)
			return
		}
		
		// Check user-specific limits
		if newReplicas > policy.MaxReplicas {
			newReplicas = policy.MaxReplicas
		}
		if newReplicas < policy.MinReplicas {
			newReplicas = policy.MinReplicas
		}
		
		// ⭐ Check tenant quota before scaling (only for scale up)
		if newReplicas > currentReplicas && userID != "" {
			hasQuota, quota, err := metadataService.CheckTenantQuota(userID)
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
			deployment, err := metadataService.GetDeploymentStatus(namespace)
			if err == nil {
				deployment.Replicas = newReplicas
				deployment.Status = "scaling"
				deployment.UpdatedAt = time.Now()
				metadataService.SaveDeploymentStatus(deployment)
			}
			
			// Update last scaling time
			a.lastScalingTime[namespace] = time.Now()
			
			// Update tenant quota usage if scaled up
			if newReplicas > currentReplicas && userID != "" {
				// Note: In a real implementation, you might want to track resource usage more precisely
				// For now, we're just checking quota, not updating usage for replica scaling
				log.Printf("Updated tenant quota check for user %s after scaling", userID)
			}
		}
	}
}

// getUserScalingPolicy gets the scaling policy for a user
func (a *Autoscaler) getUserScalingPolicy(userID string) ScalingPolicy {
	if userID == "" {
		// Return default policy if no user ID
		return ScalingPolicy{
			UserID:              "",
			EnableAutoScaleUp:   true,
			EnableAutoScaleDown: true,
			ScaleUpThreshold:    a.config.HighCPUThreshold,
			ScaleDownThreshold:  a.config.LowCPUThreshold,
			MaxReplicas:         a.config.MaxReplicas,
			MinReplicas:         a.config.MinReplicas,
		}
	}
	
	// Check if user has a specific policy
	if policy, exists := a.config.ScalingPolicies[userID]; exists {
		return policy
	}
	
	// Return default policy
	return ScalingPolicy{
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
func (a *Autoscaler) getCurrentReplicas(namespace string) (int, error) {
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
func (a *Autoscaler) getMetricsForNamespace(namespace string) (*Metrics, error) {
	// Try to get metrics from metadata service first
	metrics, err := metadataService.GetLatestMetrics(namespace)
	if err != nil {
		// Fallback to file if metadata service fails
		filename := fmt.Sprintf("server/metrics_%s.json", namespace)
		
		// Try to read metrics from file
		file, err := exec.Command("cat", filename).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("error reading metrics file: %v", err)
		}
		
		var fileMetrics Metrics
		if err := json.Unmarshal(file, &fileMetrics); err != nil {
			return nil, fmt.Errorf("error unmarshaling metrics: %v", err)
		}
		
		return &fileMetrics, nil
	}
	
	return metrics, nil
}

// calculateNewReplicas calculates the new number of replicas based on metrics and user policy
func (a *Autoscaler) calculateNewReplicas(currentReplicas int, metrics *Metrics, policy ScalingPolicy) int {
	// Check if we need to scale up
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
	return currentReplicas
}

// shouldScaleUp determines if we should scale up based on metrics and user policy
func (a *Autoscaler) shouldScaleUp(metrics *Metrics, policy ScalingPolicy) bool {
	// Scale up if CPU, memory, disk, or QPS is above high threshold
	return metrics.CPUUsage > policy.ScaleUpThreshold ||
		metrics.MemoryUsage > policy.ScaleUpThreshold ||
		metrics.DiskUsage > policy.ScaleUpThreshold ||
		metrics.QPS > policy.ScaleUpThreshold
}

// shouldScaleDown determines if we should scale down based on metrics and user policy
func (a *Autoscaler) shouldScaleDown(metrics *Metrics, policy ScalingPolicy) bool {
	// Scale down if CPU, memory, disk, and QPS are below low threshold
	return metrics.CPUUsage < policy.ScaleDownThreshold &&
		metrics.MemoryUsage < policy.ScaleDownThreshold &&
		metrics.DiskUsage < policy.ScaleDownThreshold &&
		metrics.QPS < policy.ScaleDownThreshold
}

// isInCooldown checks if the namespace is still in cooldown period after scaling
func (a *Autoscaler) isInCooldown(namespace string, currentReplicas, newReplicas int) bool {
	lastTime, exists := a.lastScalingTime[namespace]
	if !exists {
		return false // No previous scaling, not in cooldown
	}
	
	// Determine cooldown period based on scaling direction
	var cooldownSeconds int
	if newReplicas > currentReplicas {
		// Scale up
		cooldownSeconds = a.config.ScaleUpCooldown
	} else {
		// Scale down
		cooldownSeconds = a.config.ScaleDownCooldown
	}
	
	// Check if cooldown period has passed
	cooldownDuration := time.Duration(cooldownSeconds) * time.Second
	return time.Since(lastTime) < cooldownDuration
}

// updateHistoricalMetrics updates the historical metrics for a namespace
func (a *Autoscaler) updateHistoricalMetrics(namespace string, metrics *Metrics) {
	// Initialize historical metrics for this namespace if not exists
	if _, exists := a.historicalMetrics[namespace]; !exists {
		a.historicalMetrics[namespace] = &HistoricalMetrics{
			Metrics: make([]Metrics, 0),
			Window:  5, // Keep last 5 metrics for trend analysis
		}
	}
	
	// Add new metrics
	hm := a.historicalMetrics[namespace]
	hm.Metrics = append(hm.Metrics, *metrics)
	
	// Trim to window size if needed
	if len(hm.Metrics) > hm.Window {
		hm.Metrics = hm.Metrics[len(hm.Metrics)-hm.Window:]
	}
}

// adjustMetricsBasedOnTrends adjusts metrics based on trend analysis to make more informed scaling decisions
func (a *Autoscaler) adjustMetricsBasedOnTrends(metrics *Metrics, cpuTrend, memoryTrend, diskTrend, qpsTrend float64) *Metrics {
	// Create a copy of the metrics to avoid modifying the original
	adjusted := &Metrics{
		CPUUsage:    metrics.CPUUsage,
		MemoryUsage: metrics.MemoryUsage,
		DiskUsage:   metrics.DiskUsage,
		QPS:         metrics.QPS,
	}
	
	// Apply trend-based adjustments
	// If there's a positive trend, we might want to scale up sooner
	// If there's a negative trend, we might want to scale down later
	
	// CPU adjustment
	if cpuTrend > 0.5 { // Increasing trend
		adjusted.CPUUsage *= 1.1 // Increase weight for scaling decision
	} else if cpuTrend < -0.5 { // Decreasing trend
		adjusted.CPUUsage *= 0.9 // Decrease weight for scaling decision
	}
	
	// Memory adjustment
	if memoryTrend > 0.5 { // Increasing trend
		adjusted.MemoryUsage *= 1.1
	} else if memoryTrend < -0.5 { // Decreasing trend
		adjusted.MemoryUsage *= 0.9
	}
	
	// Disk adjustment
	if diskTrend > 0.5 { // Increasing trend
		adjusted.DiskUsage *= 1.1
	} else if diskTrend < -0.5 { // Decreasing trend
		adjusted.DiskUsage *= 0.9
	}
	
	// QPS adjustment
	if qpsTrend > 10 { // Increasing trend
		adjusted.QPS *= 1.1
	} else if qpsTrend < -10 { // Decreasing trend
		adjusted.QPS *= 0.9
	}
	
	return adjusted
}

// getTrendAnalysis analyzes the trend of metrics over time
func (a *Autoscaler) getTrendAnalysis(namespace string) (cpuTrend, memoryTrend, diskTrend, qpsTrend float64) {
	// Return 0 trends if no historical data
	hm, exists := a.historicalMetrics[namespace]
	if !exists || len(hm.Metrics) < 2 {
		return 0.0, 0.0, 0.0, 0.0
	}
	
	// Calculate average trends
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
func (a *Autoscaler) scaleCluster(namespace string, replicas int) error {
	cmd := exec.Command("kubectl", "-n", namespace, "scale", "sts/elasticsearch", "--replicas", strconv.Itoa(replicas))
	_, err := cmd.CombinedOutput()
	return err
}

// SetUserScalingPolicy sets a user-specific scaling policy
func (a *Autoscaler) SetUserScalingPolicy(policy ScalingPolicy) {
	a.config.ScalingPolicies[policy.UserID] = policy
}

// GetUserScalingPolicy gets a user-specific scaling policy
func (a *Autoscaler) GetUserScalingPolicy(userID string) (ScalingPolicy, bool) {
	policy, exists := a.config.ScalingPolicies[userID]
	return policy, exists
}

// RemoveUserScalingPolicy removes a user-specific scaling policy
func (a *Autoscaler) RemoveUserScalingPolicy(userID string) {
	delete(a.config.ScalingPolicies, userID)
}

// parseNamespaces parses namespace names from kubectl output
func parseNamespaces(output string) []string {
	// Split by whitespace and filter out empty strings
	names := []string{}
	for _, name := range strings.Fields(output) {
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}