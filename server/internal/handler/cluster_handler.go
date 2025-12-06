package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"es-serverless-manager/internal/model"
	"es-serverless-manager/internal/service"
)

type ClusterHandler struct {
	metadataService  *service.MetadataService
	terraformManager *service.TerraformManager
}

func NewClusterHandler(metadata *service.MetadataService, terraform *service.TerraformManager) *ClusterHandler {
	return &ClusterHandler{
		metadataService:  metadata,
		terraformManager: terraform,
	}
}

// CreateCluster creates a new cluster
// CreateCluster 创建新集群
// @Summary Create a new cluster
// @Description Create a new Elasticsearch cluster
// @Tags clusters
// @Accept json
// @Produce json
// @Param cluster body model.CreateRequest true "Cluster configuration"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /clusters [post]
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req model.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必需参数
	if req.TenantOrgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_org_id is required for multi-tenancy"})
		return
	}
	if req.User == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is required"})
		return
	}
	if req.ServiceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_name is required"})
		return
	}

	// Check tenant quota before creating cluster
	// 创建集群前检查租户配额
	if req.User != "" {
		hasQuota, quota, err := h.metadataService.CheckTenantQuota(req.User)
		if err != nil {
			log.Printf("Warning: Failed to check tenant quota for user %s: %v", req.User, err)
		} else if !hasQuota {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Tenant quota exceeded. Max indices: %d, Current indices: %d", quota.MaxIndices, quota.CurrentIndices),
			})
			return
		}
	}

	// 构建基于租户组织ID的命名空间（实现多租户隔离）
	ns := req.Namespace
	if ns == "" {
		ns = fmt.Sprintf("%s-%s-%s", req.TenantOrgID, req.User, req.ServiceName)
		log.Printf("Auto-generated namespace based on tenant_org_id: %s", ns)
	}

	// ⭐ STEP 1: 首先记录租户元数据到元数据服务（在创建K8s资源之前）
	log.Printf("Recording tenant metadata for tenant_org_id: %s, namespace: %s, user: %s, service: %s", req.TenantOrgID, ns, req.User, req.ServiceName)

	tenantContainer := &model.TenantContainer{
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
		Deleted:     false,
	}
	err := h.metadataService.SaveTenantContainer(tenantContainer)
	if err != nil {
		log.Printf("Error: Failed to save tenant container metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save tenant metadata: %v", err)})
		return
	}

	deploymentStatus := &model.DeploymentStatus{
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
	err = h.metadataService.SaveDeploymentStatus(deploymentStatus)
	if err != nil {
		log.Printf("Error: Failed to save deployment status: %v", err)
		h.metadataService.DeleteTenantContainer(req.User, req.ServiceName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save deployment status: %v", err)})
		return
	}

	// STEP 2: Use Terraform to create K8s resources
	// STEP 2: 使用 Terraform 创建 K8s 资源
	tenantConfig := model.TenantConfig{
		TenantOrgID:     req.TenantOrgID,
		User:            req.User,
		ServiceName:     req.ServiceName,
		Replicas:        req.Replicas,
		CPU:             req.CPURequest,
		Memory:          req.MemRequest,
		DiskSize:        req.DiskSize,
		StorageClass:    "hostpath",
		GPUCount:        req.GPUCount,
		VectorDimension: req.Dimension,
		VectorCount:     req.VectorCount,
	}

	if tenantConfig.Replicas <= 0 {
		tenantConfig.Replicas = 1
	}
	if tenantConfig.CPU == "" {
		tenantConfig.CPU = "500m"
	}
	if tenantConfig.Memory == "" {
		tenantConfig.Memory = "1Gi"
	}
	if _, err := strconv.Atoi(tenantConfig.Memory); err == nil {
		tenantConfig.Memory += "Gi"
	}
	if tenantConfig.DiskSize == "" {
		tenantConfig.DiskSize = "10Gi"
	}
	if _, err := strconv.Atoi(tenantConfig.DiskSize); err == nil {
		tenantConfig.DiskSize += "Gi"
	}
	if tenantConfig.VectorDimension <= 0 {
		tenantConfig.VectorDimension = 128
	}
	if tenantConfig.VectorCount <= 0 {
		tenantConfig.VectorCount = 10000
	}

	err = h.terraformManager.CreateCluster(tenantConfig)
	if err != nil {
		log.Printf("Error: Failed to create K8s resources via Terraform: %v", err)
		h.metadataService.DeleteTenantContainer(req.User, req.ServiceName)
		deploymentStatus.Status = "failed"
		deploymentStatus.UpdatedAt = time.Now()
		h.metadataService.SaveDeploymentStatus(deploymentStatus)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create cluster: %v", err)})
		return
	}

	// Update tenant quota usage
	// 更新租户配额使用量
	if req.User != "" {
		h.metadataService.UpdateTenantQuotaUsage(req.User, true, req.DiskSize)
	}

	// Update status to created
	// 更新状态为已创建
	deploymentStatus.Status = "created"
	deploymentStatus.UpdatedAt = time.Now()
	h.metadataService.SaveDeploymentStatus(deploymentStatus)

	tenantContainer.Status = "created"
	tenantContainer.SyncTime = time.Now()
	h.metadataService.SaveTenantContainer(tenantContainer)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cluster creation initiated successfully",
		"namespace": ns,
		"status":    "created",
	})
}

// DeleteCluster deletes a cluster
// DeleteCluster 删除集群
// @Summary Delete a cluster
// @Description Delete an Elasticsearch cluster
// @Tags clusters
// @Accept json
// @Produce json
// @Param cluster body model.DeleteRequest true "Cluster deletion info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /clusters [delete]
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ns := req.Namespace
	if ns == "" {
		ns = os.Getenv("NAMESPACE")
		if ns == "" {
			ns = "es-serverless"
		}
	}

	deployment, err := h.metadataService.GetDeploymentStatus(ns)
	if err != nil {
		log.Printf("Warning: Could not find deployment status for namespace %s: %v", ns, err)
	}

	// Delete K8s resources via Terraform
	// 通过 Terraform 删除 K8s 资源
	err = h.terraformManager.DeleteCluster(ns)
	if err != nil {
		log.Printf("Error: Failed to delete cluster via Terraform: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete cluster: %v", err)})
		return
	}

	if deployment != nil {
		// Mark tenant container as deleted
		// 标记租户容器为已删除
		h.metadataService.DeleteTenantContainer(deployment.User, deployment.ServiceName)

		deployment.Status = "deleted"
		deployment.UpdatedAt = time.Now()
		h.metadataService.SaveDeploymentStatus(deployment)

		// Release quota
		// 释放配额
		h.metadataService.UpdateTenantQuotaUsage(deployment.User, false, "")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cluster deleted successfully",
		"namespace": ns,
		"status":    "deleted",
	})
}

// ScaleCluster scales a cluster
// ScaleCluster 扩缩容集群
// @Summary Scale a cluster
// @Description Scale an Elasticsearch cluster
// @Tags clusters
// @Accept json
// @Produce json
// @Param cluster body model.ScaleRequest true "Cluster scaling info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /clusters/scale [post]
func (h *ClusterHandler) ScaleCluster(c *gin.Context) {
	var req model.ScaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ns := req.Namespace
	if ns == "" {
		ns = os.Getenv("NAMESPACE")
		if ns == "" {
			ns = "es-serverless"
		}
	}

	deployment, err := h.metadataService.GetDeploymentStatus(ns)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Deployment not found: %v", err)})
		return
	}

	getString := func(v interface{}) string {
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}

	getInt := func(v interface{}) int {
		if i, ok := v.(int); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
		return 0
	}

	tenantConfig := model.TenantConfig{
		TenantOrgID:     deployment.TenantOrgID,
		User:            deployment.User,
		ServiceName:     deployment.ServiceName,
		Replicas:        req.Replicas,
		CPU:             getString(deployment.Details["cpu_request"]),
		Memory:          getString(deployment.Details["mem_request"]),
		DiskSize:        getString(deployment.Details["disk_size"]),
		StorageClass:    "hostpath",
		GPUCount:        getInt(deployment.Details["gpu_count"]),
		VectorDimension: getInt(deployment.Details["dimension"]),
		VectorCount:     getInt(deployment.Details["vector_count"]),
	}

	if tenantConfig.CPU == "" {
		tenantConfig.CPU = "500m"
	}
	if tenantConfig.Memory == "" {
		tenantConfig.Memory = "1Gi"
	}
	if tenantConfig.DiskSize == "" {
		tenantConfig.DiskSize = "10Gi"
	}

	// Apply changes via Terraform
	// 通过 Terraform 应用变更
	err = h.terraformManager.CreateCluster(tenantConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scale cluster: %v", err)})
		return
	}

	deployment.Replicas = req.Replicas
	deployment.Details["replicas"] = req.Replicas
	deployment.UpdatedAt = time.Now()
	deployment.Status = "scaling"

	h.metadataService.SaveDeploymentStatus(deployment)

	if tenantContainer, err := h.metadataService.GetTenantContainer(deployment.User, deployment.ServiceName); err == nil {
		tenantContainer.Replicas = req.Replicas
		tenantContainer.SyncTime = time.Now()
		h.metadataService.SaveTenantContainer(tenantContainer)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cluster scaling initiated successfully via Terraform",
		"namespace": ns,
		"replicas":  req.Replicas,
		"status":    "scaling",
	})
}

// ListClusters lists all clusters
// ListClusters 列出所有集群
// @Summary List all clusters
// @Description List all Elasticsearch clusters
// @Tags clusters
// @Produce json
// @Success 200 {array} model.ClusterStatus
// @Failure 500 {string} string "Internal Server Error"
// @Router /clusters [get]
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	deployments, err := h.metadataService.ListDeploymentStatus()
	if err != nil {
		// Fallback to kubectl
		cmd := exec.Command("kubectl", "get", "namespaces", "-l", "es-cluster=true", "-o", "jsonpath={.items[*].metadata.name}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": string(out)})
			return
		}

		namespaces := strings.Fields(string(out))
		clusters := make([]model.ClusterStatus, len(namespaces))

		for i, ns := range namespaces {
			statusCmd := exec.Command("kubectl", "-n", ns, "get", "sts/elasticsearch", "-o", "jsonpath={.status.readyReplicas}/{.spec.replicas}")
			statusOut, _ := statusCmd.CombinedOutput()
			status := string(statusOut)
			if status == "" {
				status = "unknown"
			}

			clusters[i] = model.ClusterStatus{
				Namespace:   ns,
				User:        "unknown",
				ServiceName: "unknown",
				Status:      status,
				Replicas:    1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
		}
		c.JSON(http.StatusOK, clusters)
		return
	}

	clusters := make([]model.ClusterStatus, len(deployments))
	for i, deployment := range deployments {
		statusCmd := exec.Command("kubectl", "-n", deployment.Namespace, "get", "sts/elasticsearch", "-o", "jsonpath={.status.readyReplicas}/{.spec.replicas}")
		statusOut, _ := statusCmd.CombinedOutput()
		status := string(statusOut)
		if status == "" {
			status = "unknown"
		}

		clusters[i] = model.ClusterStatus{
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

	c.JSON(http.StatusOK, clusters)
}
