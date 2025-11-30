package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CreateRequest struct {
	TenantOrgID string `json:"tenant_org_id"` // 租户组织ID（多租户隔离）
	User        string `json:"user"`
	ServiceName string `json:"service_name"`
	Namespace   string `json:"namespace"`
	Replicas    int    `json:"replicas"`
	CPURequest  string `json:"cpu_request"`
	CPULimit    string `json:"cpu_limit"`
	MemRequest  string `json:"mem_request"`
	MemLimit    string `json:"mem_limit"`
	DiskSize    string `json:"disk_size"`
	GPUCount    int    `json:"gpu_count"`
	Dimension   int    `json:"dimension"`
	VectorCount int    `json:"vector_count"`
	IndexLimit  int    `json:"index_limit"`
	GitlabURL   string `json:"gitlab_url"`
}

type DeleteRequest struct {
	Namespace string `json:"namespace"`
}

type ScaleRequest struct {
	Namespace string `json:"namespace"`
	Replicas  int    `json:"replicas"`
}

type ClusterStatus struct {
	Namespace   string                 `json:"namespace"`
	User        string                 `json:"user"`
	ServiceName string                 `json:"service_name"`
	Status      string                 `json:"status"`
	CPUUsage    float64                `json:"cpu_usage"`
	MemoryUsage float64                `json:"memory_usage"`
	DiskUsage   float64                `json:"disk_usage"`
	QPS         float64                `json:"qps"`
	GPUCount    int                    `json:"gpu_count"`
	Dimension   int                    `json:"dimension"`
	VectorCount int                    `json:"vector_count"`
	Replicas    int                    `json:"replicas"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Details     map[string]interface{} `json:"details"`
}

type VectorIndexRequest struct {
	IndexName    string            `json:"index_name"`
	Dimension    int               `json:"dimension"`
	Metric       string            `json:"metric"`     // L2, cosine, dot
	IVFParams    map[string]int    `json:"ivf_params"` // nlist, nprobe
	FieldMapping map[string]string `json:"field_mapping"`
}

type VectorIndexStatus struct {
	IndexName     string         `json:"index_name"`
	Dimension     int            `json:"dimension"`
	Metric        string         `json:"metric"`
	IVFParams     map[string]int `json:"ivf_params"`
	Status        string         `json:"status"`
	DocumentCount int            `json:"document_count"`
	CreatedAt     time.Time      `json:"created_at"`
}

// IVFParams represents IVF algorithm parameters
type IVFParams struct {
	NList  int `json:"nlist"`
	NProbe int `json:"nprobe"`
}

// IndexMetadata represents index metadata
type IndexMetadata struct {
	ID            string    `json:"id"`
	IndexName     string    `json:"index_name"`
	Namespace     string    `json:"namespace"`
	Dimension     int       `json:"dimension"`
	Metric        string    `json:"metric"`
	IVFParams     IVFParams `json:"ivf_params"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     string    `json:"created_by"`
	Status        string    `json:"status"` // active, deleted, building
	DocumentCount int       `json:"document_count"`
	StorageSize   string    `json:"storage_size"`
}

// TenantQuota represents tenant quota information
type TenantQuota struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	MaxIndices     int       `json:"max_indices"`
	MaxStorage     string    `json:"max_storage"`
	CurrentIndices int       `json:"current_indices"`
	CurrentStorage string    `json:"current_storage"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DeploymentStatus represents deployment status information
type DeploymentStatus struct {
	ID          string                 `json:"id"`
	TenantOrgID string                 `json:"tenant_org_id"` // 租户组织ID
	Namespace   string                 `json:"namespace"`
	User        string                 `json:"user"`
	ServiceName string                 `json:"service_name"`
	Status      string                 `json:"status"` // created, running, scaling, deleting, error
	CPUUsage    float64                `json:"cpu_usage"`
	MemoryUsage float64                `json:"memory_usage"`
	DiskUsage   float64                `json:"disk_usage"`
	QPS         float64                `json:"qps"`
	GPUCount    int                    `json:"gpu_count"`
	Dimension   int                    `json:"dimension"`
	VectorCount int                    `json:"vector_count"`
	Replicas    int                    `json:"replicas"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Details     map[string]interface{} `json:"details"`
}

// Global services
var (
	metadataService *MetadataService
	autoscaler      *Autoscaler
)

func envOrDefault(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 验证必需参数
	if req.TenantOrgID == "" {
		http.Error(w, "tenant_org_id is required for multi-tenancy", http.StatusBadRequest)
		return
	}
	if req.User == "" {
		http.Error(w, "user is required", http.StatusBadRequest)
		return
	}
	if req.ServiceName == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}

	// Check tenant quota before creating cluster
	if req.User != "" {
		hasQuota, quota, err := metadataService.CheckTenantQuota(req.User)
		if err != nil {
			// Log the error but don't fail the request if quota checking fails
			log.Printf("Warning: Failed to check tenant quota for user %s: %v", req.User, err)
		} else if !hasQuota {
			http.Error(w, fmt.Sprintf("Tenant quota exceeded. Max indices: %d, Current indices: %d", quota.MaxIndices, quota.CurrentIndices), http.StatusForbidden)
			return
		}
	}

	// 构建基于租户组织ID的命名空间（实现多租户隔离）
	ns := req.Namespace
	if ns == "" {
		// 默认命名空间格式：{tenant_org_id}-{user}-{service_name}
		ns = fmt.Sprintf("%s-%s-%s", req.TenantOrgID, req.User, req.ServiceName)
		log.Printf("Auto-generated namespace based on tenant_org_id: %s", ns)
	}

	// ⭐ STEP 1: 首先记录租户元数据到元数据服务（在创建K8s资源之前）
	log.Printf("Recording tenant metadata for tenant_org_id: %s, namespace: %s, user: %s, service: %s", req.TenantOrgID, ns, req.User, req.ServiceName)

	// 创建租户容器记录
	tenantContainer := &TenantContainer{
		TenantOrgID: req.TenantOrgID,
		User:        req.User,
		ServiceName: req.ServiceName,
		Namespace:   ns,
		Replicas:    req.Replicas,
		CPU:         req.CPURequest + "/" + req.CPULimit,
		Memory:      req.MemRequest + "/" + req.MemLimit,
		Disk:        req.DiskSize,
		GPUCount:    req.GPUCount,
		Dimension:   req.Dimension,
		VectorCount: req.VectorCount,
		Status:      "creating",
		CreatedAt:   time.Now(),
		SyncTime:    time.Now(),
		Deleted:     false, // 初始化为未删除状态
	}
	err := metadataService.SaveTenantContainer(tenantContainer)
	if err != nil {
		log.Printf("Error: Failed to save tenant container metadata: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save tenant metadata: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully saved tenant metadata for namespace: %s", ns)

	// 保存部署状态到元数据服务
	deploymentStatus := &DeploymentStatus{
		TenantOrgID: req.TenantOrgID,
		Namespace:   ns,
		User:        req.User,
		ServiceName: req.ServiceName,
		Status:      "creating",
		Replicas:    req.Replicas,
		CPUUsage:    0.0,
		MemoryUsage: 0.0,
		DiskUsage:   0.0,
		QPS:         0.0,
		GPUCount:    req.GPUCount,
		Dimension:   req.Dimension,
		VectorCount: req.VectorCount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Details: map[string]interface{}{
			"cpu_request":  req.CPURequest,
			"cpu_limit":    req.CPULimit,
			"mem_request":  req.MemRequest,
			"mem_limit":    req.MemLimit,
			"disk_size":    req.DiskSize,
			"gpu_count":    req.GPUCount,
			"dimension":    req.Dimension,
			"vector_count": req.VectorCount,
			"index_limit":  req.IndexLimit,
			"gitlab_url":   req.GitlabURL,
		},
	}
	err = metadataService.SaveDeploymentStatus(deploymentStatus)
	if err != nil {
		log.Printf("Error: Failed to save deployment status: %v", err)
		// 如果保存部署状态失败，回滚租户容器记录
		metadataService.DeleteTenantContainer(req.User, req.ServiceName)
		http.Error(w, fmt.Sprintf("Failed to save deployment status: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully saved deployment status for namespace: %s", ns)

	// STEP 2: 然后创建K8s资源

	env := os.Environ()
	env = append(env, "NAMESPACE="+ns)
	if req.TenantOrgID != "" {
		env = append(env, "TENANT_ORG_ID="+req.TenantOrgID)
	}
	if req.User != "" {
		env = append(env, "USER="+req.User)
	}
	if req.ServiceName != "" {
		env = append(env, "SERVICE_NAME="+req.ServiceName)
	}
	if req.Replicas > 0 {
		env = append(env, "REPLICAS="+strconv.Itoa(req.Replicas))
	}
	if req.CPURequest != "" {
		env = append(env, "CPU_REQUEST="+req.CPURequest)
	}
	if req.CPULimit != "" {
		env = append(env, "CPU_LIMIT="+req.CPULimit)
	}
	if req.MemRequest != "" {
		env = append(env, "MEM_REQUEST="+req.MemRequest)
	}
	if req.MemLimit != "" {
		env = append(env, "MEM_LIMIT="+req.MemLimit)
	}
	if req.DiskSize != "" {
		env = append(env, "DISK_SIZE="+req.DiskSize)
	}
	if req.GPUCount > 0 {
		env = append(env, "GPU_COUNT="+strconv.Itoa(req.GPUCount))
	}
	if req.Dimension > 0 {
		env = append(env, "DIMENSION="+strconv.Itoa(req.Dimension))
	}
	if req.VectorCount > 0 {
		env = append(env, "VECTOR_COUNT="+strconv.Itoa(req.VectorCount))
	}
	if req.IndexLimit > 0 {
		env = append(env, "INDEX_LIMIT="+strconv.Itoa(req.IndexLimit))
	}
	if req.GitlabURL != "" {
		env = append(env, "GITLAB_URL="+req.GitlabURL)
	}

	cmd := exec.Command("bash", "scripts/cluster.sh", "create")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: Failed to create K8s resources: %v", err)
		// 回滚：删除元数据记录
		metadataService.DeleteTenantContainer(req.User, req.ServiceName)
		deploymentStatus.Status = "failed"
		deploymentStatus.UpdatedAt = time.Now()
		metadataService.SaveDeploymentStatus(deploymentStatus)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(out)
		return
	}
	log.Printf("Successfully created K8s resources for namespace: %s", ns)

	// STEP 3: 更新部署状态为创建成功
	// Record deployment status
	deploymentInfo := map[string]interface{}{
		"namespace":    ns,
		"user":         req.User,
		"service_name": req.ServiceName,
		"status":       "created",
		"created_at":   time.Now(),
		"details": map[string]interface{}{
			"replicas":     req.Replicas,
			"cpu_request":  req.CPURequest,
			"cpu_limit":    req.CPULimit,
			"mem_request":  req.MemRequest,
			"mem_limit":    req.MemLimit,
			"disk_size":    req.DiskSize,
			"gpu_count":    req.GPUCount,
			"dimension":    req.Dimension,
			"vector_count": req.VectorCount,
			"gitlab_url":   req.GitlabURL,
		},
	}

	// Save deployment info to file
	saveDeploymentInfo(deploymentInfo)

	// Update tenant quota usage
	if req.User != "" {
		metadataService.UpdateTenantQuotaUsage(req.User, true, req.DiskSize)
	}

	// 更新部署状态为创建成功
	deploymentStatus.Status = "created"
	deploymentStatus.UpdatedAt = time.Now()
	metadataService.SaveDeploymentStatus(deploymentStatus)
	log.Printf("Updated deployment status to 'created' for namespace: %s", ns)

	// 更新租户容器状态
	tenantContainer.Status = "created"
	tenantContainer.SyncTime = time.Now()
	metadataService.SaveTenantContainer(tenantContainer)
	log.Printf("Updated tenant container status to 'created' for namespace: %s", ns)

	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// saveDeploymentInfo saves deployment information to a JSON file
func saveDeploymentInfo(info map[string]interface{}) error {
	// Read existing deployments
	file, err := os.OpenFile("server/deployments.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read existing data
	var deployments []map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&deployments); err != nil && err != io.EOF {
		return err
	}

	// Append new deployment
	deployments = append(deployments, info)

	// Write back to file
	file.Truncate(0)
	file.Seek(0, 0)
	encoder := json.NewEncoder(file)
	return encoder.Encode(deployments)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := req.Namespace
	if ns == "" {
		ns = envOrDefault("NAMESPACE", "es-serverless")
	}

	env := os.Environ()
	env = append(env, "NAMESPACE="+ns)
	cmd := exec.Command("bash", "scripts/cluster.sh", "delete")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(out)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func handleScale(w http.ResponseWriter, r *http.Request) {
	var req ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := req.Namespace
	if ns == "" {
		ns = envOrDefault("NAMESPACE", "es-serverless")
	}

	env := os.Environ()
	env = append(env, "NAMESPACE="+ns)
	env = append(env, "REPLICAS="+strconv.Itoa(req.Replicas))

	// Scale the Elasticsearch cluster
	cmd := exec.Command("kubectl", "-n", ns, "scale", "sts/elasticsearch", "--replicas", strconv.Itoa(req.Replicas))
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(out)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// handleListClusters now uses metadata service for more accurate information
func handleListClusters(w http.ResponseWriter, r *http.Request) {
	// Get deployment statuses from metadata service
	deployments, err := metadataService.ListDeploymentStatus()
	if err != nil {
		// Fallback to kubectl if metadata service fails
		cmd := exec.Command("kubectl", "get", "namespaces", "-l", "es-cluster=true", "-o", "jsonpath={.items[*].metadata.name}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(w, string(out), http.StatusInternalServerError)
			return
		}

		namespaces := strings.Fields(string(out))
		clusters := make([]ClusterStatus, len(namespaces))

		for i, ns := range namespaces {
			// Get cluster status
			statusCmd := exec.Command("kubectl", "-n", ns, "get", "sts/elasticsearch", "-o", "jsonpath={.status.readyReplicas}/{.spec.replicas}")
			statusOut, _ := statusCmd.CombinedOutput()
			status := string(statusOut)
			if status == "" {
				status = "unknown"
			}

			// In a real implementation, you would retrieve these values from a database or config map
			// For now, we'll use default values
			clusters[i] = ClusterStatus{
				Namespace:   ns,
				User:        "unknown",
				ServiceName: "unknown",
				Status:      status,
				CPUUsage:    0.0,
				MemoryUsage: 0.0,
				DiskUsage:   0.0,
				QPS:         0.0,
				GPUCount:    0,
				Dimension:   128,
				VectorCount: 10000,
				Replicas:    1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Details: map[string]interface{}{
					"cpu_request": "500m",
					"cpu_limit":   "2",
					"mem_request": "1Gi",
					"mem_limit":   "2Gi",
					"disk_size":   "10Gi",
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusters)
		return
	}

	// Convert deployment statuses to cluster statuses
	clusters := make([]ClusterStatus, len(deployments))
	for i, deployment := range deployments {
		// Get current replica status from k8s
		statusCmd := exec.Command("kubectl", "-n", deployment.Namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.status.readyReplicas}/{.spec.replicas}")
		statusOut, _ := statusCmd.CombinedOutput()
		status := string(statusOut)
		if status == "" {
			status = "unknown"
		}

		clusters[i] = ClusterStatus{
			Namespace:   deployment.Namespace,
			User:        deployment.User,
			ServiceName: deployment.ServiceName,
			Status:      status,
			CPUUsage:    deployment.CPUUsage,
			MemoryUsage: deployment.MemoryUsage,
			DiskUsage:   deployment.DiskUsage,
			QPS:         deployment.QPS,
			GPUCount:    deployment.GPUCount,
			Dimension:   deployment.Dimension,
			VectorCount: deployment.VectorCount,
			Replicas:    deployment.Replicas,
			CreatedAt:   deployment.CreatedAt,
			UpdatedAt:   deployment.UpdatedAt,
			Details:     deployment.Details,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func handleCreateVectorIndex(w http.ResponseWriter, r *http.Request) {
	var req VectorIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// This would call the Elasticsearch API to create a vector index
	// For now, we'll just return a success response
	response := map[string]interface{}{
		"message": "Vector index creation initiated",
		"index":   req.IndexName,
		"status":  "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListVectorIndexes(w http.ResponseWriter, r *http.Request) {
	// This would call the Elasticsearch API to list vector indexes
	// For now, we'll just return a mock response
	indexes := []VectorIndexStatus{
		{
			IndexName:     "sample_vector_index",
			Dimension:     128,
			Metric:        "l2",
			IVFParams:     map[string]int{"nlist": 100, "nprobe": 10},
			Status:        "ready",
			DocumentCount: 1000,
			CreatedAt:     time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indexes)
}

func handleDeleteVectorIndex(w http.ResponseWriter, r *http.Request) {
	// This would call the Elasticsearch API to delete a vector index
	// For now, we'll just return a success response
	response := map[string]interface{}{
		"message": "Vector index deletion initiated",
		"status":  "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListDeployments(w http.ResponseWriter, r *http.Request) {
	// Read deployment info from file
	file, err := os.Open("server/deployments.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var deployments []map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&deployments); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

// handleListMetrics returns the latest metrics for all namespaces
func handleListMetrics(w http.ResponseWriter, r *http.Request) {
	// Get list of namespaces with ES clusters
	cmd := exec.Command("kubectl", "get", "namespaces", "-l", "es-cluster=true", "-o", "jsonpath={.items[*].metadata.name}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}

	namespaces := strings.Fields(string(out))
	metricsList := make([]Metrics, 0, len(namespaces))

	for _, ns := range namespaces {
		// Try to read metrics from file
		filename := fmt.Sprintf("server/metrics_%s.json", ns)
		file, err := os.Open(filename)
		if err != nil {
			// If file doesn't exist, create empty metrics
			metricsList = append(metricsList, Metrics{
				Namespace: ns,
				Timestamp: time.Now(),
			})
			continue
		}

		var metrics Metrics
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&metrics); err != nil {
			file.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Close()

		metricsList = append(metricsList, metrics)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricsList)
}

// saveIndexMetadata saves index metadata to the metadata service
func saveIndexMetadata(req VectorIndexRequest, user string) error {
	metadata := &IndexMetadata{
		IndexName: req.IndexName,
		// Namespace would be extracted from context in a real implementation
		Dimension: req.Dimension,
		Metric:    req.Metric,
		IVFParams: IVFParams{
			NList:  req.IVFParams["nlist"],
			NProbe: req.IVFParams["nprobe"],
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: user,
		Status:    "active",
	}

	return metadataService.SaveIndexMetadata(metadata)
}

// handleCreateIndexMetadata handles creating index metadata
func handleCreateIndexMetadata(w http.ResponseWriter, r *http.Request) {
	var metadata IndexMetadata
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := metadataService.SaveIndexMetadata(&metadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// handleGetIndexMetadata handles getting index metadata by ID
func handleGetIndexMetadata(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/metadata/indexes/")
	if id == "" {
		http.Error(w, "Missing index ID", http.StatusBadRequest)
		return
	}

	metadata, err := metadataService.GetIndexMetadata(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// handleListIndexMetadata handles listing all index metadata
func handleListIndexMetadata(w http.ResponseWriter, r *http.Request) {
	metadataList, err := metadataService.ListIndexMetadata()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadataList)
}

// handleUpdateIndexMetadata handles updating index metadata
func handleUpdateIndexMetadata(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/metadata/indexes/")
	if id == "" {
		http.Error(w, "Missing index ID", http.StatusBadRequest)
		return
	}

	var metadata IndexMetadata
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Preserve the ID
	metadata.ID = id
	metadata.UpdatedAt = time.Now()

	if err := metadataService.SaveIndexMetadata(&metadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// handleDeleteIndexMetadata handles deleting index metadata
func handleDeleteIndexMetadata(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/metadata/indexes/")
	if id == "" {
		http.Error(w, "Missing index ID", http.StatusBadRequest)
		return
	}

	if err := metadataService.DeleteIndexMetadata(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Index metadata deleted successfully"))
}

// handleCreateTenantQuota handles creating tenant quota
func handleCreateTenantQuota(w http.ResponseWriter, r *http.Request) {
	var quota TenantQuota
	if err := json.NewDecoder(r.Body).Decode(&quota); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := metadataService.SaveTenantQuota(&quota); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quota)
}

// handleGetTenantQuota handles getting tenant quota by tenant ID
func handleGetTenantQuota(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from URL path
	tenantID := strings.TrimPrefix(r.URL.Path, "/metadata/tenants/")
	if tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	quota, err := metadataService.GetTenantQuota(tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quota)
}

// handleUpdateTenantQuota handles updating tenant quota
func handleUpdateTenantQuota(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from URL path
	tenantID := strings.TrimPrefix(r.URL.Path, "/metadata/tenants/")
	if tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	var quota TenantQuota
	if err := json.NewDecoder(r.Body).Decode(&quota); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Preserve the ID and tenant ID
	quota.TenantID = tenantID
	quota.UpdatedAt = time.Now()

	if err := metadataService.SaveTenantQuota(&quota); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quota)
}

// handleGetDeploymentStatus handles getting deployment status by namespace
func handleGetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	// Extract namespace from URL path
	namespace := strings.TrimPrefix(r.URL.Path, "/metadata/deployments/")
	if namespace == "" {
		http.Error(w, "Missing namespace", http.StatusBadRequest)
		return
	}

	status, err := metadataService.GetDeploymentStatus(namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleUpdateDeploymentStatus handles updating deployment status
func handleUpdateDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	// Extract namespace from URL path
	namespace := strings.TrimPrefix(r.URL.Path, "/metadata/deployments/")
	if namespace == "" {
		http.Error(w, "Missing namespace", http.StatusBadRequest)
		return
	}

	var status DeploymentStatus
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Preserve the namespace
	status.Namespace = namespace
	status.UpdatedAt = time.Now()

	if err := metadataService.SaveDeploymentStatus(&status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	// Create data directory for metadata
	dataDir := envOrDefault("METADATA_DIR", "server/data")
	os.MkdirAll(dataDir, 0755)

	// Initialize metadata service
	metadataService = NewMetadataService(dataDir)

	// Create Elasticsearch client
	// Use localhost when running outside Docker, use service name when running in Docker
	esBaseURL := envOrDefault("ES_BASE_URL", "http://localhost:9200")
	esClient := NewESClient(esBaseURL)

	// Create and start shard controller
	shardController := NewShardController(esClient)
	shardController.Start()
	defer shardController.Stop()

	// Create and start replication monitor
	replicationMonitor := NewReplicationMonitor(esClient)
	replicationMonitor.Start()
	defer replicationMonitor.Stop()

	// Create and start consistency checker
	consistencyChecker := NewConsistencyChecker(esClient)
	consistencyChecker.Start()
	defer consistencyChecker.Stop()

	// Create and start auto recovery manager
	autoRecoveryManager := NewAutoRecoveryManager(esClient, replicationMonitor, consistencyChecker)
	autoRecoveryManager.Start()
	defer autoRecoveryManager.Stop()

	// Create and start monitoring service
	monitoringService := NewMonitoringService()
	monitoringService.Start()
	defer monitoringService.Stop()

	// Create and start autoscaler
	autoscaler = NewAutoscaler()
	autoscaler.Start()
	defer autoscaler.Stop()

	// Create and start reporting service
	reportingService := NewReportingService(esClient, "") // No reporting URL for now
	reportingService.Start()
	defer reportingService.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/clusters", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(w, r)
		case http.MethodDelete:
			handleDelete(w, r)
		case http.MethodGet:
			handleListClusters(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/clusters/scale", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleScale(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/vector-indexes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateVectorIndex(w, r)
			// Report index creation
			var req VectorIndexRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				ivfParams := IVFParams{
					NList:  req.IVFParams["nlist"],
					NProbe: req.IVFParams["nprobe"],
				}
				reportingService.ReportIndexCreation(req.IndexName, req.Dimension, req.Metric, ivfParams)

				// Save index metadata
				saveIndexMetadata(req, "unknown_user") // In a real implementation, you would get the user from auth context
			}
		case http.MethodGet:
			handleListVectorIndexes(w, r)
		case http.MethodDelete:
			handleDeleteVectorIndex(w, r)
			// Report index deletion
			// In a real implementation, you would extract the index name from the request
			reportingService.ReportIndexDeletion("unknown_index")
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add deployment status endpoint
	mux.HandleFunc("/deployments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListDeployments(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListMetrics(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add shard management endpoint
	mux.HandleFunc("/shards", shardController.ShardManagementHandler)

	// Add metadata management endpoints
	mux.HandleFunc("/metadata/indexes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateIndexMetadata(w, r)
		case http.MethodGet:
			handleListIndexMetadata(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/metadata/indexes/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetIndexMetadata(w, r)
		case http.MethodPut:
			handleUpdateIndexMetadata(w, r)
		case http.MethodDelete:
			handleDeleteIndexMetadata(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/metadata/tenants/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateTenantQuota(w, r)
		case http.MethodGet:
			handleGetTenantQuota(w, r)
		case http.MethodPut:
			handleUpdateTenantQuota(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/metadata/deployments/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetDeploymentStatus(w, r)
		case http.MethodPut:
			handleUpdateDeploymentStatus(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add monitoring metrics endpoints
	mux.HandleFunc("/monitoring/metrics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListAllMetrics(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/monitoring/metrics/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetNamespaceMetrics(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add detailed container monitoring endpoints
	mux.HandleFunc("/monitoring/container-metrics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListContainerMetrics(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/monitoring/container-metrics/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetContainerMetrics(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add autoscaler policy management endpoints
	mux.HandleFunc("/autoscaler/policies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateAutoscalerPolicy(w, r)
		case http.MethodGet:
			handleListAutoscalerPolicies(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/autoscaler/policies/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetAutoscalerPolicy(w, r)
		case http.MethodPut:
			handleUpdateAutoscalerPolicy(w, r)
		case http.MethodDelete:
			handleDeleteAutoscalerPolicy(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add tenant container management endpoints
	mux.HandleFunc("/tenant/containers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListTenantContainers(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/tenant/containers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// 判断是否是查询组织的接口
			if strings.HasPrefix(r.URL.Path, "/tenant/containers/org/") {
				handleListTenantContainersByOrgID(w, r)
			} else {
				handleGetTenantContainer(w, r)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add deployment reports endpoints
	mux.HandleFunc("/deployment/reports", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListDeploymentReports(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/deployment/reports/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetDeploymentReport(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add cluster details endpoint
	mux.HandleFunc("/clusters/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetClusterDetails(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add QPS query endpoint
	mux.HandleFunc("/qps/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetQPSMetrics(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add replication status endpoint
	mux.HandleFunc("/replication/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetAllReplicationStatus(w, r, replicationMonitor)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/replication/status/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetIndexReplicationStatus(w, r, replicationMonitor)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add consistency check endpoint
	mux.HandleFunc("/consistency/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetAllConsistencyReports(w, r, consistencyChecker)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/consistency/reports/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetConsistencyReport(w, r, consistencyChecker)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/consistency/check/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCheckIndexConsistency(w, r, consistencyChecker)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Add recovery management endpoint
	mux.HandleFunc("/recovery/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetRecoveryHistory(w, r, autoRecoveryManager)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/recovery/active", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetActiveRecoveries(w, r, autoRecoveryManager)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/recovery/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetRecoveryConfig(w, r, autoRecoveryManager)
		case http.MethodPost:
			handleUpdateRecoveryConfig(w, r, autoRecoveryManager)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	addr := ":8080"
	// Check if PORT environment variable is set
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// handleListAllMetrics handles listing all metrics
func handleListAllMetrics(w http.ResponseWriter, r *http.Request) {
	// Get list of namespaces with ES clusters
	cmd := exec.Command("kubectl", "get", "namespaces", "-l", "es-cluster=true", "-o", "jsonpath={.items[*].metadata.name}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}

	namespaces := strings.Fields(string(out))
	metricsList := make([]*Metrics, 0, len(namespaces))

	for _, ns := range namespaces {
		// Try to get latest metrics from metadata service
		metrics, err := metadataService.GetLatestMetrics(ns)
		if err != nil {
			// If we can't get metrics from metadata service, try to get from file
			filename := fmt.Sprintf("server/metrics_%s.json", ns)
			file, err := os.ReadFile(filename)
			if err != nil {
				continue
			}

			var fileMetrics Metrics
			if err := json.Unmarshal(file, &fileMetrics); err != nil {
				continue
			}

			metricsList = append(metricsList, &fileMetrics)
		} else {
			metricsList = append(metricsList, metrics)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricsList)
}

// handleGetNamespaceMetrics handles getting metrics for a specific namespace
func handleGetNamespaceMetrics(w http.ResponseWriter, r *http.Request) {
	// Extract namespace from URL path
	namespace := strings.TrimPrefix(r.URL.Path, "/monitoring/metrics/")
	if namespace == "" {
		http.Error(w, "Missing namespace", http.StatusBadRequest)
		return
	}

	// Get metrics for the namespace
	metrics, err := metadataService.GetLatestMetrics(namespace)
	if err != nil {
		// If we can't get metrics from metadata service, try to get from file
		filename := fmt.Sprintf("server/metrics_%s.json", namespace)
		file, err := os.ReadFile(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("No metrics found for namespace %s", namespace), http.StatusNotFound)
			return
		}

		var fileMetrics Metrics
		if err := json.Unmarshal(file, &fileMetrics); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing metrics for namespace %s: %v", namespace, err), http.StatusInternalServerError)
			return
		}

		metrics = &fileMetrics
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleListContainerMetrics handles listing all container metrics
func handleListContainerMetrics(w http.ResponseWriter, r *http.Request) {
	// This would query the monitoring service for all container metrics
	// For now, we'll return a mock response with detailed container metrics

	// In a real implementation, this would call the monitoring service to get actual metrics
	mockMetrics := []map[string]interface{}{
		{
			"id":             "container_metrics_ns1_1234567890",
			"namespace":      "test-namespace-1",
			"container_name": "elasticsearch",
			"cpu_usage":      45.5,
			"memory_usage":   1024.0,
			"disk_usage":     65.0,
			"qps":            150.5,
			"startup_cpu":    50.0,
			"startup_memory": 1000.0,
			"startup_disk":   60.0,
			"plugin_qps":     75.2,
			"timestamp":      time.Now().Format(time.RFC3339),
			"status":         "running",
			"resource_limits": map[string]string{
				"cpu":    "2",
				"memory": "2Gi",
			},
			"resource_requests": map[string]string{
				"cpu":    "500m",
				"memory": "1Gi",
			},
		},
		{
			"id":             "container_metrics_ns2_1234567891",
			"namespace":      "test-namespace-2",
			"container_name": "elasticsearch",
			"cpu_usage":      30.2,
			"memory_usage":   768.0,
			"disk_usage":     45.0,
			"qps":            95.8,
			"startup_cpu":    35.0,
			"startup_memory": 750.0,
			"startup_disk":   40.0,
			"plugin_qps":     48.1,
			"timestamp":      time.Now().Format(time.RFC3339),
			"status":         "running",
			"resource_limits": map[string]string{
				"cpu":    "1",
				"memory": "1Gi",
			},
			"resource_requests": map[string]string{
				"cpu":    "250m",
				"memory": "512Mi",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockMetrics)
}

// handleGetContainerMetrics handles getting container metrics for a specific namespace
func handleGetContainerMetrics(w http.ResponseWriter, r *http.Request) {
	// Extract namespace from URL path
	namespace := strings.TrimPrefix(r.URL.Path, "/monitoring/container-metrics/")
	if namespace == "" {
		http.Error(w, "Missing namespace", http.StatusBadRequest)
		return
	}

	// This would query the monitoring service for container metrics for the specific namespace
	// For now, we'll return a mock response with detailed container metrics

	// In a real implementation, this would call the monitoring service to get actual metrics for the namespace
	mockMetric := map[string]interface{}{
		"id":             fmt.Sprintf("container_metrics_%s_1234567890", namespace),
		"namespace":      namespace,
		"container_name": "elasticsearch",
		"cpu_usage":      45.5,
		"memory_usage":   1024.0,
		"disk_usage":     65.0,
		"qps":            150.5,
		"startup_cpu":    50.0,
		"startup_memory": 1000.0,
		"startup_disk":   60.0,
		"plugin_qps":     75.2,
		"timestamp":      time.Now().Format(time.RFC3339),
		"status":         "running",
		"resource_limits": map[string]string{
			"cpu":    "2",
			"memory": "2Gi",
		},
		"resource_requests": map[string]string{
			"cpu":    "500m",
			"memory": "1Gi",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockMetric)
}

// handleListTenantContainers handles listing all tenant containers
func handleListTenantContainers(w http.ResponseWriter, r *http.Request) {
	// Get tenant data directory
	tenantDir := "server/tenant_data"

	// Check if directory exists
	if _, err := os.Stat(tenantDir); os.IsNotExist(err) {
		http.Error(w, "Tenant data directory does not exist", http.StatusNotFound)
		return
	}

	// Read all tenant data files
	files, err := os.ReadDir(tenantDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading tenant data directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Collect all tenant container data
	var containers []map[string]interface{}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			// Read tenant data file
			content, err := os.ReadFile(filepath.Join(tenantDir, file.Name()))
			if err != nil {
				log.Printf("Error reading tenant data file %s: %v", file.Name(), err)
				continue
			}

			// Parse JSON data
			var containerData map[string]interface{}
			if err := json.Unmarshal(content, &containerData); err != nil {
				log.Printf("Error parsing tenant data file %s: %v", file.Name(), err)
				continue
			}

			containers = append(containers, containerData)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// handleGetTenantContainer handles getting a specific tenant container
func handleGetTenantContainer(w http.ResponseWriter, r *http.Request) {
	// Extract user and service name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/tenant/containers/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path format. Expected: /tenant/containers/{user}/{service_name}", http.StatusBadRequest)
		return
	}

	user := parts[0]
	serviceName := parts[1]

	// Construct tenant data file path
	tenantFile := fmt.Sprintf("server/tenant_data/%s_%s.json", user, serviceName)

	// Check if file exists
	if _, err := os.Stat(tenantFile); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Tenant container data not found for user %s and service %s", user, serviceName), http.StatusNotFound)
		return
	}

	// Read tenant data file
	content, err := os.ReadFile(tenantFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading tenant data file: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse JSON data
	var containerData map[string]interface{}
	if err := json.Unmarshal(content, &containerData); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing tenant data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containerData)
}

// handleListDeploymentReports handles listing all deployment reports
func handleListDeploymentReports(w http.ResponseWriter, r *http.Request) {
	// Get deployment reports directory
	reportsDir := "server/deployment_reports"

	// Check if directory exists
	if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
		http.Error(w, "Deployment reports directory does not exist", http.StatusNotFound)
		return
	}

	// Read all deployment report files
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading deployment reports directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Collect all deployment reports
	var reports []map[string]interface{}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			// Read deployment report file
			content, err := os.ReadFile(filepath.Join(reportsDir, file.Name()))
			if err != nil {
				log.Printf("Error reading deployment report file %s: %v", file.Name(), err)
				continue
			}

			// Parse JSON data
			var reportData map[string]interface{}
			if err := json.Unmarshal(content, &reportData); err != nil {
				log.Printf("Error parsing deployment report file %s: %v", file.Name(), err)
				continue
			}

			reports = append(reports, reportData)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// handleGetDeploymentReport handles getting a specific deployment report
func handleGetDeploymentReport(w http.ResponseWriter, r *http.Request) {
	// Extract user and service name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/deployment/reports/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path format. Expected: /deployment/reports/{user}/{service_name}", http.StatusBadRequest)
		return
	}

	user := parts[0]
	serviceName := parts[1]

	// Find the latest deployment report for this user and service
	reportsDir := "server/deployment_reports"
	if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
		http.Error(w, "Deployment reports directory does not exist", http.StatusNotFound)
		return
	}

	// Read all deployment report files
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading deployment reports directory: %v", err), http.StatusInternalServerError)
		return
	}

	var latestReport map[string]interface{}
	var latestTime time.Time

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && strings.HasPrefix(file.Name(), fmt.Sprintf("%s_%s_", user, serviceName)) {
			// Read deployment report file
			content, err := os.ReadFile(filepath.Join(reportsDir, file.Name()))
			if err != nil {
				log.Printf("Error reading deployment report file %s: %v", file.Name(), err)
				continue
			}

			// Parse JSON data
			var reportData map[string]interface{}
			if err := json.Unmarshal(content, &reportData); err != nil {
				log.Printf("Error parsing deployment report file %s: %v", file.Name(), err)
				continue
			}

			// Check if this is the latest report
			if timestampStr, ok := reportData["timestamp"].(string); ok {
				if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
					if timestamp.After(latestTime) {
						latestTime = timestamp
						latestReport = reportData
					}
				}
			}
		}
	}

	if latestReport == nil {
		http.Error(w, fmt.Sprintf("Deployment report not found for user %s and service %s", user, serviceName), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latestReport)
}

// handleGetClusterDetails handles getting details for a specific cluster by namespace
func handleGetClusterDetails(w http.ResponseWriter, r *http.Request) {
	// Extract namespace from URL path
	namespace := strings.TrimPrefix(r.URL.Path, "/clusters/")
	if namespace == "" {
		http.Error(w, "Missing namespace", http.StatusBadRequest)
		return
	}

	// Get deployment status from metadata service
	status, err := metadataService.GetDeploymentStatus(namespace)
	if err != nil {
		// Fallback to kubectl if metadata service fails
		cmd := exec.Command("kubectl", "-n", namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.status.readyReplicas}/{.spec.replicas}")
		statusOut, cmdErr := cmd.CombinedOutput()
		statusStr := string(statusOut)
		if statusStr == "" || cmdErr != nil {
			statusStr = "unknown"
		}

		// Get resource info from kubectl
		resourceCmd := exec.Command("kubectl", "-n", namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.spec.template.spec.containers[0].resources}")
		resourceOut, _ := resourceCmd.CombinedOutput()
		resourceInfo := string(resourceOut)

		// Create cluster status with available information
		clusterStatus := ClusterStatus{
			Namespace:   namespace,
			User:        "unknown",
			ServiceName: "unknown",
			Status:      statusStr,
			CPUUsage:    0.0,
			MemoryUsage: 0.0,
			DiskUsage:   0.0,
			QPS:         0.0,
			GPUCount:    0,
			Dimension:   128,
			VectorCount: 10000,
			Replicas:    1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Details: map[string]interface{}{
				"resource_info": resourceInfo,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterStatus)
		return
	}

	// Convert deployment status to cluster status
	clusterStatus := ClusterStatus{
		Namespace:   status.Namespace,
		User:        status.User,
		ServiceName: status.ServiceName,
		Status:      status.Status,
		CPUUsage:    status.CPUUsage,
		MemoryUsage: status.MemoryUsage,
		DiskUsage:   status.DiskUsage,
		QPS:         status.QPS,
		GPUCount:    status.GPUCount,
		Dimension:   status.Dimension,
		VectorCount: status.VectorCount,
		Replicas:    status.Replicas,
		CreatedAt:   status.CreatedAt,
		UpdatedAt:   status.UpdatedAt,
		Details:     status.Details,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusterStatus)
}

// handleGetQPSMetrics handles getting QPS metrics for a specific namespace
func handleGetQPSMetrics(w http.ResponseWriter, r *http.Request) {
	// Extract namespace from URL path
	namespace := strings.TrimPrefix(r.URL.Path, "/qps/")
	if namespace == "" {
		http.Error(w, "Missing namespace", http.StatusBadRequest)
		return
	}

	// Try to get QPS metrics from metadata service
	metrics, err := metadataService.GetLatestMetrics(namespace)
	if err != nil {
		// If we can't get metrics from metadata service, try to get from file
		filename := fmt.Sprintf("server/metrics_%s.json", namespace)
		file, err := os.ReadFile(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("No QPS metrics found for namespace %s", namespace), http.StatusNotFound)
			return
		}

		if err := json.Unmarshal(file, &metrics); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing QPS metrics for namespace %s: %v", namespace, err), http.StatusInternalServerError)
			return
		}
	}

	// Create QPS response
	qpsResponse := map[string]interface{}{
		"namespace": namespace,
		"qps":       metrics.QPS,
		"timestamp": metrics.Timestamp,
		"id":        metrics.ID,
	}

	// Try to get container metrics for plugin QPS
	containerMetrics, err := metadataService.GetContainerMetrics(namespace)
	if err == nil {
		qpsResponse["plugin_qps"] = containerMetrics.PluginQPS
		qpsResponse["avg_latency"] = containerMetrics.PluginQPS * 0.5 // Mock value
		qpsResponse["p95_latency"] = containerMetrics.PluginQPS * 0.8 // Mock value
		qpsResponse["p99_latency"] = containerMetrics.PluginQPS * 0.9 // Mock value
	} else {
		// Fallback to monitoring service
		monitoringService := NewMonitoringService()
		containerMetrics, err := monitoringService.getContainerMetricsForNamespace(namespace)
		if err == nil {
			qpsResponse["plugin_qps"] = containerMetrics.PluginQPS
			qpsResponse["avg_latency"] = containerMetrics.PluginQPS * 0.5 // Mock value
			qpsResponse["p95_latency"] = containerMetrics.PluginQPS * 0.8 // Mock value
			qpsResponse["p99_latency"] = containerMetrics.PluginQPS * 0.9 // Mock value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(qpsResponse)
}

// handleCreateAutoscalerPolicy handles creating an autoscaler policy
func handleCreateAutoscalerPolicy(w http.ResponseWriter, r *http.Request) {
	var policy ScalingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the policy in the autoscaler
	autoscaler.SetUserScalingPolicy(policy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// handleGetAutoscalerPolicy handles getting an autoscaler policy by user ID
func handleGetAutoscalerPolicy(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/autoscaler/policies/")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Get the policy from the autoscaler
	policy, exists := autoscaler.GetUserScalingPolicy(userID)
	if !exists {
		http.Error(w, fmt.Sprintf("Autoscaler policy not found for user %s", userID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// handleListAutoscalerPolicies handles listing all autoscaler policies
func handleListAutoscalerPolicies(w http.ResponseWriter, r *http.Request) {
	// This would query the autoscaler for all policies
	// For now, we'll return a mock response

	// In a real implementation, this would call the autoscaler to get actual policies
	mockPolicies := []ScalingPolicy{
		{
			UserID:              "user1",
			EnableAutoScaleUp:   true,
			EnableAutoScaleDown: true,
			ScaleUpThreshold:    75.0,
			ScaleDownThreshold:  25.0,
			MaxReplicas:         5,
			MinReplicas:         1,
		},
		{
			UserID:              "user2",
			EnableAutoScaleUp:   true,
			EnableAutoScaleDown: false,
			ScaleUpThreshold:    80.0,
			ScaleDownThreshold:  30.0,
			MaxReplicas:         3,
			MinReplicas:         2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockPolicies)
}

// handleUpdateAutoscalerPolicy handles updating an autoscaler policy
func handleUpdateAutoscalerPolicy(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/autoscaler/policies/")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	var policy ScalingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure the user ID in the policy matches the URL
	policy.UserID = userID

	// Update the policy in the autoscaler
	autoscaler.SetUserScalingPolicy(policy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// handleDeleteAutoscalerPolicy handles deleting an autoscaler policy
func handleDeleteAutoscalerPolicy(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/autoscaler/policies/")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Remove the policy from the autoscaler
	autoscaler.RemoveUserScalingPolicy(userID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Autoscaler policy deleted successfully"))
}

// handleListTenantContainersByOrgID handles listing all tenant containers by organization ID
func handleListTenantContainersByOrgID(w http.ResponseWriter, r *http.Request) {
	// Extract tenant org ID from URL path
	tenantOrgID := strings.TrimPrefix(r.URL.Path, "/tenant/containers/org/")
	if tenantOrgID == "" {
		http.Error(w, "Missing tenant organization ID", http.StatusBadRequest)
		return
	}

	log.Printf("Querying tenant containers for org ID: %s", tenantOrgID)

	// Get containers from metadata service
	containers, err := metadataService.ListTenantContainersByOrgID(tenantOrgID)
	if err != nil {
		log.Printf("Error listing tenant containers for org %s: %v", tenantOrgID, err)
		http.Error(w, fmt.Sprintf("Error listing tenant containers: %v", err), http.StatusInternalServerError)
		return
	}

	if len(containers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// handleGetAllReplicationStatus handles getting all replication statuses
func handleGetAllReplicationStatus(w http.ResponseWriter, r *http.Request, rm *ReplicationMonitor) {
	statuses := rm.GetAllReplicationStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

// handleGetIndexReplicationStatus handles getting replication status for a specific index
func handleGetIndexReplicationStatus(w http.ResponseWriter, r *http.Request, rm *ReplicationMonitor) {
	// Extract index name from URL path
	indexName := strings.TrimPrefix(r.URL.Path, "/replication/status/")
	if indexName == "" {
		http.Error(w, "Missing index name", http.StatusBadRequest)
		return
	}

	status, err := rm.GetReplicationStatus(indexName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleGetAllConsistencyReports handles getting all consistency reports
func handleGetAllConsistencyReports(w http.ResponseWriter, r *http.Request, cc *ConsistencyChecker) {
	reports := cc.GetAllConsistencyReports()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// handleGetConsistencyReport handles getting consistency report for a specific index
func handleGetConsistencyReport(w http.ResponseWriter, r *http.Request, cc *ConsistencyChecker) {
	// Extract index name from URL path
	indexName := strings.TrimPrefix(r.URL.Path, "/consistency/reports/")
	if indexName == "" {
		http.Error(w, "Missing index name", http.StatusBadRequest)
		return
	}

	report, err := cc.GetConsistencyReport(indexName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// handleCheckIndexConsistency handles triggering an immediate consistency check
func handleCheckIndexConsistency(w http.ResponseWriter, r *http.Request, cc *ConsistencyChecker) {
	// Extract index name from URL path
	indexName := strings.TrimPrefix(r.URL.Path, "/consistency/check/")
	if indexName == "" {
		http.Error(w, "Missing index name", http.StatusBadRequest)
		return
	}

	report, err := cc.CheckIndexNow(indexName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking consistency: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// handleGetRecoveryHistory handles getting recovery history
func handleGetRecoveryHistory(w http.ResponseWriter, r *http.Request, arm *AutoRecoveryManager) {
	history := arm.GetRecoveryHistory()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// handleGetActiveRecoveries handles getting active recoveries
func handleGetActiveRecoveries(w http.ResponseWriter, r *http.Request, arm *AutoRecoveryManager) {
	active := arm.GetActiveRecoveries()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(active)
}

// handleGetRecoveryConfig handles getting recovery configuration
func handleGetRecoveryConfig(w http.ResponseWriter, r *http.Request, arm *AutoRecoveryManager) {
	// 这里可以返回当前配置
	config := map[string]interface{}{
		"auto_recovery_enabled":  true,
		"max_retries":            3,
		"retry_delay_seconds":    30,
		"check_interval_seconds": 60,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleUpdateRecoveryConfig handles updating recovery configuration
func handleUpdateRecoveryConfig(w http.ResponseWriter, r *http.Request, arm *AutoRecoveryManager) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 更新配置
	if enabled, ok := config["auto_recovery_enabled"].(bool); ok {
		if enabled {
			arm.EnableAutoRecovery()
		} else {
			arm.DisableAutoRecovery()
		}
	}

	if maxRetries, ok := config["max_retries"].(float64); ok {
		arm.SetMaxRetries(int(maxRetries))
	}

	if retryDelay, ok := config["retry_delay_seconds"].(float64); ok {
		arm.SetRetryDelay(time.Duration(retryDelay) * time.Second)
	}

	if checkInterval, ok := config["check_interval_seconds"].(float64); ok {
		arm.SetCheckInterval(time.Duration(checkInterval) * time.Second)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Recovery configuration updated successfully"))
}
