package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"es-serverless-manager/internal/model"
)

// MonitoringService handles container monitoring
// MonitoringService 处理容器监控
type MonitoringService struct {
	metadataService *MetadataService
	ticker          *time.Ticker
}

// NewMonitoringService creates a new monitoring service
// NewMonitoringService 创建一个新的监控服务
func NewMonitoringService(metadataService *MetadataService) *MonitoringService {
	return &MonitoringService{
		metadataService: metadataService,
		ticker:          time.NewTicker(30 * time.Second),
	}
}

// Start begins the monitoring loop
// Start 启动监控循环
func (ms *MonitoringService) Start() {
	go func() {
		for range ms.ticker.C {
			ms.collectMetrics()
		}
	}()
}

// Stop stops the monitoring loop
// Stop 停止监控循环
func (ms *MonitoringService) Stop() {
	ms.ticker.Stop()
}

// collectMetrics collects metrics from all namespaces
// collectMetrics 收集所有命名空间的指标
func (ms *MonitoringService) collectMetrics() {
	// Get list of namespaces with ES clusters
	// 获取具有 ES 集群的命名空间列表
	cmd := exec.Command("kubectl", "get", "namespaces", "-l", "es-cluster=true", "-o", "jsonpath={.items[*].metadata.name}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error getting namespaces: %v", err)
		return
	}

	namespaces := strings.Fields(string(out))
	for _, ns := range namespaces {
		containerMetrics, err := ms.getContainerMetricsForNamespace(ns)
		if err != nil {
			log.Printf("Error getting container metrics for namespace %s: %v", ns, err)
			continue
		}

		// Save container metrics to metadata service
		// 将容器指标保存到元数据服务
		err = ms.saveContainerMetricsToMetadataService(containerMetrics)
		if err != nil {
			log.Printf("Error saving container metrics to metadata service for namespace %s: %v", ns, err)
		}

		// Update deployment status with latest metrics
		// 使用最新指标更新部署状态
		ms.updateDeploymentStatusWithMetrics(ns, containerMetrics)
	}
}

// updateDeploymentStatusWithMetrics updates deployment status with latest metrics
// updateDeploymentStatusWithMetrics 使用最新指标更新部署状态
func (ms *MonitoringService) updateDeploymentStatusWithMetrics(namespace string, metrics *model.ContainerMetrics) {
	// Get current deployment status
	// 获取当前部署状态
	deployment, err := ms.metadataService.GetDeploymentStatus(namespace)
	if err != nil {
		// If deployment status doesn't exist, that's okay - we'll skip updating it
		// 如果部署状态不存在，没关系 - 我们将跳过更新它
		return
	}

	// Update deployment status with latest metrics
	// 使用最新指标更新部署状态
	deployment.CPUUsage = metrics.CPUUsage
	deployment.MemoryUsage = metrics.MemoryUsage
	deployment.DiskUsage = metrics.DiskUsage
	deployment.QPS = metrics.QPS
	deployment.UpdatedAt = time.Now()

	// Update status based on metrics
	// 根据指标更新状态
	if metrics.CPUUsage > 80 || metrics.MemoryUsage > 80 {
		deployment.Status = "high_load"
	} else if metrics.CPUUsage < 20 && metrics.MemoryUsage < 20 {
		deployment.Status = "low_load"
	} else {
		deployment.Status = "normal"
	}

	// Save updated deployment status
	// 保存更新后的部署状态
	err = ms.metadataService.SaveDeploymentStatus(deployment)
	if err != nil {
		log.Printf("Error updating deployment status for namespace %s: %v", namespace, err)
	}
}

// getContainerMetricsForNamespace gets detailed container metrics for a specific namespace
// getContainerMetricsForNamespace 获取特定命名空间的详细容器指标
func (ms *MonitoringService) getContainerMetricsForNamespace(namespace string) (*model.ContainerMetrics, error) {
	// Get CPU and memory usage
	// 获取 CPU 和内存使用情况
	cpu, memory, err := ms.getResourceUsage(namespace)
	if err != nil {
		return nil, fmt.Errorf("error getting resource usage: %v", err)
	}

	// Get disk usage
	// 获取磁盘使用情况
	disk, err := ms.getDiskUsage(namespace)
	if err != nil {
		log.Printf("Warning: error getting disk usage for namespace %s: %v", namespace, err)
		disk = 0.0
	}

	// Get QPS (in a real implementation, this would come from Elasticsearch)
	// 获取 QPS（在实际实现中，这将来自 Elasticsearch）
	qps := ms.getQPS(namespace)

	// Get plugin QPS
	// 获取插件 QPS
	pluginQPS := ms.getPluginQPS(namespace)

	// Get startup metrics
	// 获取启动指标
	startupCPU, startupMemory, startupDisk, err := ms.getStartupMetrics(namespace)
	if err != nil {
		log.Printf("Warning: error getting startup metrics for namespace %s: %v", namespace, err)
		startupCPU, startupMemory, startupDisk = 0.0, 0.0, 0.0
	}

	// Get resource limits and requests
	// 获取资源限制和请求
	resourceLimits, resourceRequests, err := ms.getResourceLimitsAndRequests(namespace)
	if err != nil {
		log.Printf("Warning: error getting resource limits and requests for namespace %s: %v", namespace, err)
	}

	containerMetrics := &model.ContainerMetrics{
		ID:            fmt.Sprintf("container_metrics_%s_%d", namespace, time.Now().UnixNano()),
		Namespace:     namespace,
		ContainerName: "elasticsearch",
		CPUUsage:      cpu,
		MemoryUsage:   memory,
		DiskUsage:     disk,
		QPS:           qps,
		StartupCPU:    startupCPU,
		StartupMemory: startupMemory,
		StartupDisk:   startupDisk,
		PluginQPS:     pluginQPS,
		Timestamp:     time.Now(),
		Status:        "running",
		ResourceLimits: model.ResourceLimits{
			CPU:    resourceLimits.CPU,
			Memory: resourceLimits.Memory,
		},
		ResourceRequests: model.ResourceRequests{
			CPU:    resourceRequests.CPU,
			Memory: resourceRequests.Memory,
		},
	}

	return containerMetrics, nil
}

// getResourceUsage gets CPU and memory usage for a namespace
// getResourceUsage 获取命名空间的 CPU 和内存使用情况
func (ms *MonitoringService) getResourceUsage(namespace string) (cpu float64, memory float64, err error) {
	// Get pod metrics using kubectl top
	// 使用 kubectl top 获取 pod 指标
	cmd := exec.Command("kubectl", "top", "pods", "-n", namespace, "--no-headers")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return 0, 0, nil
	}

	// Parse the first pod's metrics (in a real implementation, you would aggregate all pods)
	// 解析第一个 pod 的指标（在实际实现中，您将聚合所有 pod）
	fields := strings.Fields(lines[0])
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("unexpected output format")
	}

	// Parse CPU usage (e.g., "100m" -> 0.1)
	// 解析 CPU 使用情况（例如 "100m" -> 0.1）
	cpuStr := strings.TrimSuffix(fields[1], "m")
	if cpuVal, err := parseFloat(cpuStr); err == nil {
		cpu = cpuVal / 1000 // Convert millicores to cores
	}

	// Parse memory usage (e.g., "100Mi" -> 100)
	// 解析内存使用情况（例如 "100Mi" -> 100）
	memStr := strings.TrimSuffix(fields[2], "Mi")
	if memVal, err := parseFloat(memStr); err == nil {
		memory = memVal // In MB
	}

	return cpu, memory, nil
}

// getDiskUsage gets disk usage for a namespace
// getDiskUsage 获取命名空间的磁盘使用情况
func (ms *MonitoringService) getDiskUsage(namespace string) (float64, error) {
	// Get PVC usage
	// 获取 PVC 使用情况
	cmd := exec.Command("kubectl", "exec", "-n", namespace, "elasticsearch-0", "--", "df", "-h", "/usr/share/elasticsearch/data")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}

	// Parse disk usage percentage (e.g., "75%" -> 75)
	// 解析磁盘使用百分比（例如 "75%" -> 75）
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	percentStr := strings.TrimSuffix(fields[4], "%")
	if percent, err := parseFloat(percentStr); err == nil {
		return percent, nil
	}

	return 0, fmt.Errorf("unable to parse disk usage")
}

// getQPS gets QPS for a namespace (mock implementation)
// getQPS 获取命名空间的 QPS（模拟实现）
func (ms *MonitoringService) getQPS(namespace string) float64 {
	// In a real implementation, this would query Elasticsearch for QPS metrics
	// For now, we'll return a mock value
	// 在实际实现中，这将查询 Elasticsearch 的 QPS 指标
	// 目前，我们返回一个模拟值
	return 100.0 + float64(time.Now().UnixNano()%1000)/100.0
}

// getPluginQPS gets plugin QPS for a namespace (mock implementation)
// getPluginQPS 获取命名空间的插件 QPS（模拟实现）
func (ms *MonitoringService) getPluginQPS(namespace string) float64 {
	// In a real implementation, this would query the Elasticsearch plugin for QPS metrics
	// For now, we'll return a mock value
	// 在实际实现中，这将查询 Elasticsearch 插件的 QPS 指标
	// 目前，我们返回一个模拟值
	return 50.0 + float64(time.Now().UnixNano()%500)/100.0
}

// getStartupMetrics gets startup metrics for a namespace
// getStartupMetrics 获取命名空间的启动指标
func (ms *MonitoringService) getStartupMetrics(namespace string) (cpu float64, memory float64, disk float64, err error) {
	// In a real implementation, this would get the metrics from when the container was started
	// For now, we'll return mock values based on current metrics
	// 在实际实现中，这将获取容器启动时的指标
	// 目前，我们根据当前指标返回模拟值
	cpu, memory, err = ms.getResourceUsage(namespace)
	if err != nil {
		return 0, 0, 0, err
	}

	disk, err = ms.getDiskUsage(namespace)
	if err != nil {
		return cpu, memory, 0, nil
	}

	return cpu, memory, disk, nil
}

// getResourceLimitsAndRequests gets resource limits and requests for a namespace
// getResourceLimitsAndRequests 获取命名空间的资源限制和请求
func (ms *MonitoringService) getResourceLimitsAndRequests(namespace string) (model.ResourceLimits, model.ResourceRequests, error) {
	// Get resource limits and requests from the statefulset
	// 从 StatefulSet 获取资源限制和请求
	cmd := exec.Command("kubectl", "-n", namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.spec.template.spec.containers[0].resources}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return model.ResourceLimits{}, model.ResourceRequests{}, err
	}

	// Parse the JSON output
	// 解析 JSON 输出
	var resources map[string]map[string]string
	if err := json.Unmarshal(out, &resources); err != nil {
		return model.ResourceLimits{}, model.ResourceRequests{}, err
	}

	limits := model.ResourceLimits{}
	requests := model.ResourceRequests{}

	if limitsMap, ok := resources["limits"]; ok {
		limits.CPU = limitsMap["cpu"]
		limits.Memory = limitsMap["memory"]
	}

	if requestsMap, ok := resources["requests"]; ok {
		requests.CPU = requestsMap["cpu"]
		requests.Memory = requestsMap["memory"]
	}

	return limits, requests, nil
}

// saveContainerMetricsToMetadataService saves container metrics to the metadata service
// saveContainerMetricsToMetadataService 将容器指标保存到元数据服务
func (ms *MonitoringService) saveContainerMetricsToMetadataService(metrics *model.ContainerMetrics) error {
	// Convert ContainerMetrics to regular Metrics for backward compatibility
	// 将 ContainerMetrics 转换为常规 Metrics 以实现向后兼容性
	regularMetrics := &model.Metrics{
		ID:          metrics.ID,
		Namespace:   metrics.Namespace,
		CPUUsage:    metrics.CPUUsage,
		MemoryUsage: metrics.MemoryUsage,
		DiskUsage:   metrics.DiskUsage,
		QPS:         metrics.QPS,
		Timestamp:   metrics.Timestamp,
	}

	// Save metrics to metadata service
	// 将指标保存到元数据服务
	err := ms.metadataService.SaveMetrics(regularMetrics)
	if err != nil {
		return fmt.Errorf("failed to save metrics to metadata service: %v", err)
	}

	log.Printf("Successfully saved container metrics to metadata service for namespace %s", metrics.Namespace)
	return nil
}

// parseFloat is a helper function to parse float64 from string
// parseFloat 是一个辅助函数，用于从字符串解析 float64
func parseFloat(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}
