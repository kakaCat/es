package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// Metrics represents container resource usage metrics
type Metrics struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	QPS         float64   `json:"qps"`
	Timestamp   time.Time `json:"timestamp"`
}

// ContainerMetrics represents detailed container metrics including startup data
type ContainerMetrics struct {
	ID              string            `json:"id"`
	Namespace       string            `json:"namespace"`
	ContainerName   string            `json:"container_name"`
	CPUUsage        float64           `json:"cpu_usage"`
	MemoryUsage     float64           `json:"memory_usage"`
	DiskUsage       float64           `json:"disk_usage"`
	QPS             float64           `json:"qps"`
	StartupCPU      float64           `json:"startup_cpu"`
	StartupMemory   float64           `json:"startup_memory"`
	StartupDisk     float64           `json:"startup_disk"`
	PluginQPS       float64           `json:"plugin_qps"`
	Timestamp       time.Time         `json:"timestamp"`
	Status          string            `json:"status"`
	ResourceLimits  ResourceLimits    `json:"resource_limits"`
	ResourceRequests ResourceRequests  `json:"resource_requests"`
}

// ResourceLimits represents container resource limits
type ResourceLimits struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// ResourceRequests represents container resource requests
type ResourceRequests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// MonitoringService handles container monitoring
type MonitoringService struct {
	ticker *time.Ticker
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		ticker: time.NewTicker(30 * time.Second),
	}
}

// Start begins the monitoring loop
func (ms *MonitoringService) Start() {
	go func() {
		for range ms.ticker.C {
			ms.collectMetrics()
		}
	}()
}

// Stop stops the monitoring loop
func (ms *MonitoringService) Stop() {
	ms.ticker.Stop()
}

// collectMetrics collects metrics from all namespaces
func (ms *MonitoringService) collectMetrics() {
	// Get list of namespaces with ES clusters
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
		err = ms.saveContainerMetricsToMetadataService(containerMetrics)
		if err != nil {
			log.Printf("Error saving container metrics to metadata service for namespace %s: %v", ns, err)
		}

		// Also save metrics to file for backward compatibility
		err = ms.saveMetricsToFile(containerMetrics)
		if err != nil {
			log.Printf("Error saving metrics to file for namespace %s: %v", ns, err)
		}
		
		// Update deployment status with latest metrics
		ms.updateDeploymentStatusWithMetrics(ns, containerMetrics)
	}
}

// updateDeploymentStatusWithMetrics updates deployment status with latest metrics
func (ms *MonitoringService) updateDeploymentStatusWithMetrics(namespace string, metrics *ContainerMetrics) {
	// Get current deployment status
	deployment, err := metadataService.GetDeploymentStatus(namespace)
	if err != nil {
		// If deployment status doesn't exist, that's okay - we'll skip updating it
		return
	}
	
	// Update deployment status with latest metrics
	deployment.CPUUsage = metrics.CPUUsage
	deployment.MemoryUsage = metrics.MemoryUsage
	deployment.DiskUsage = metrics.DiskUsage
	deployment.QPS = metrics.QPS
	deployment.UpdatedAt = time.Now()
	
	// Update status based on metrics
	if metrics.CPUUsage > 80 || metrics.MemoryUsage > 80 {
		deployment.Status = "high_load"
	} else if metrics.CPUUsage < 20 && metrics.MemoryUsage < 20 {
		deployment.Status = "low_load"
	} else {
		deployment.Status = "normal"
	}
	
	// Save updated deployment status
	err = metadataService.SaveDeploymentStatus(deployment)
	if err != nil {
		log.Printf("Error updating deployment status for namespace %s: %v", namespace, err)
	}
}

// getContainerMetricsForNamespace gets detailed container metrics for a specific namespace
func (ms *MonitoringService) getContainerMetricsForNamespace(namespace string) (*ContainerMetrics, error) {
	// Get CPU and memory usage
	cpu, memory, err := ms.getResourceUsage(namespace)
	if err != nil {
		return nil, fmt.Errorf("error getting resource usage: %v", err)
	}

	// Get disk usage
	disk, err := ms.getDiskUsage(namespace)
	if err != nil {
		log.Printf("Warning: error getting disk usage for namespace %s: %v", namespace, err)
		disk = 0.0
	}

	// Get QPS (in a real implementation, this would come from Elasticsearch)
	qps := ms.getQPS(namespace)
	
	// Get plugin QPS
	pluginQPS := ms.getPluginQPS(namespace)

	// Get startup metrics
	startupCPU, startupMemory, startupDisk, err := ms.getStartupMetrics(namespace)
	if err != nil {
		log.Printf("Warning: error getting startup metrics for namespace %s: %v", namespace, err)
		startupCPU, startupMemory, startupDisk = 0.0, 0.0, 0.0
	}

	// Get resource limits and requests
	resourceLimits, resourceRequests, err := ms.getResourceLimitsAndRequests(namespace)
	if err != nil {
		log.Printf("Warning: error getting resource limits and requests for namespace %s: %v", namespace, err)
	}

	containerMetrics := &ContainerMetrics{
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
		ResourceLimits: ResourceLimits{
			CPU:    resourceLimits.CPU,
			Memory: resourceLimits.Memory,
		},
		ResourceRequests: ResourceRequests{
			CPU:    resourceRequests.CPU,
			Memory: resourceRequests.Memory,
		},
	}

	return containerMetrics, nil
}

// getResourceUsage gets CPU and memory usage for a namespace
func (ms *MonitoringService) getResourceUsage(namespace string) (cpu float64, memory float64, err error) {
	// Get pod metrics using kubectl top
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
	fields := strings.Fields(lines[0])
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("unexpected output format")
	}

	// Parse CPU usage (e.g., "100m" -> 0.1)
	cpuStr := strings.TrimSuffix(fields[1], "m")
	if cpuVal, err := parseFloat(cpuStr); err == nil {
		cpu = cpuVal / 1000 // Convert millicores to cores
	}

	// Parse memory usage (e.g., "100Mi" -> 100)
	memStr := strings.TrimSuffix(fields[2], "Mi")
	if memVal, err := parseFloat(memStr); err == nil {
		memory = memVal // In MB
	}

	return cpu, memory, nil
}

// getDiskUsage gets disk usage for a namespace
func (ms *MonitoringService) getDiskUsage(namespace string) (float64, error) {
	// Get PVC usage
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
func (ms *MonitoringService) getQPS(namespace string) float64 {
	// In a real implementation, this would query Elasticsearch for QPS metrics
	// For now, we'll return a mock value
	return 100.0 + float64(time.Now().UnixNano()%1000)/100.0
}

// getPluginQPS gets plugin QPS for a namespace (mock implementation)
func (ms *MonitoringService) getPluginQPS(namespace string) float64 {
	// In a real implementation, this would query the Elasticsearch plugin for QPS metrics
	// For now, we'll return a mock value
	return 50.0 + float64(time.Now().UnixNano()%500)/100.0
}

// getStartupMetrics gets startup metrics for a namespace
func (ms *MonitoringService) getStartupMetrics(namespace string) (cpu float64, memory float64, disk float64, err error) {
	// In a real implementation, this would get the metrics from when the container was started
	// For now, we'll return mock values based on current metrics
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
func (ms *MonitoringService) getResourceLimitsAndRequests(namespace string) (ResourceLimits, ResourceRequests, error) {
	// Get resource limits and requests from the statefulset
	cmd := exec.Command("kubectl", "-n", namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.spec.template.spec.containers[0].resources}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ResourceLimits{}, ResourceRequests{}, err
	}

	// Parse the JSON output
	var resources map[string]map[string]string
	if err := json.Unmarshal(out, &resources); err != nil {
		return ResourceLimits{}, ResourceRequests{}, err
	}

	limits := ResourceLimits{}
	requests := ResourceRequests{}

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
func (ms *MonitoringService) saveContainerMetricsToMetadataService(metrics *ContainerMetrics) error {
	// Convert ContainerMetrics to regular Metrics for backward compatibility
	regularMetrics := &Metrics{
		ID:          metrics.ID,
		Namespace:   metrics.Namespace,
		CPUUsage:    metrics.CPUUsage,
		MemoryUsage: metrics.MemoryUsage,
		DiskUsage:   metrics.DiskUsage,
		QPS:         metrics.QPS,
		Timestamp:   metrics.Timestamp,
	}
	
	// Save metrics to metadata service
	err := metadataService.SaveMetrics(regularMetrics)
	if err != nil {
		return fmt.Errorf("failed to save metrics to metadata service: %v", err)
	}
	
	log.Printf("Successfully saved container metrics to metadata service for namespace %s", metrics.Namespace)
	return nil
}

// saveMetricsToFile saves metrics to a file (backward compatibility)
func (ms *MonitoringService) saveMetricsToFile(metrics *ContainerMetrics) error {
	// Convert ContainerMetrics to regular Metrics for backward compatibility
	regularMetrics := &Metrics{
		ID:          metrics.ID,
		Namespace:   metrics.Namespace,
		CPUUsage:    metrics.CPUUsage,
		MemoryUsage: metrics.MemoryUsage,
		DiskUsage:   metrics.DiskUsage,
		QPS:         metrics.QPS,
		Timestamp:   metrics.Timestamp,
	}

	// In a real implementation, you would save to a database
	// For now, we'll save to a JSON file
	filename := fmt.Sprintf("server/metrics_%s.json", metrics.Namespace)
	file, err := json.MarshalIndent(regularMetrics, "", "  ")
	if err != nil {
		return err
	}

	return writeToFile(filename, file)
}

// writeToFile writes data to a file
func writeToFile(filename string, data []byte) error {
	return exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > %s", string(data), filename)).Run()
}

// parseFloat is a helper function to parse float64 from string
func parseFloat(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}