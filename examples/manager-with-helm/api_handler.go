package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// APIServer 提供 Helm 管理的 HTTP API
type APIServer struct {
	helmManager *TenantHelmManager
}

// NewAPIServer 创建 API 服务器
func NewAPIServer() *APIServer {
	return &APIServer{
		helmManager: NewTenantHelmManager(),
	}
}

// SetupRoutes 设置路由
func (s *APIServer) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// 租户集群管理
	r.HandleFunc("/api/v1/tenant-clusters", s.CreateTenantClusterHandler).Methods("POST")
	r.HandleFunc("/api/v1/tenant-clusters", s.ListTenantClustersHandler).Methods("GET")
	r.HandleFunc("/api/v1/tenant-clusters/{namespace}", s.GetTenantClusterHandler).Methods("GET")
	r.HandleFunc("/api/v1/tenant-clusters/{namespace}", s.DeleteTenantClusterHandler).Methods("DELETE")
	r.HandleFunc("/api/v1/tenant-clusters/{namespace}/scale", s.ScaleTenantClusterHandler).Methods("POST")

	return r
}

// CreateTenantClusterHandler 创建租户集群
func (s *APIServer) CreateTenantClusterHandler(w http.ResponseWriter, r *http.Request) {
	var req TenantClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// 验证必需字段
	if req.TenantOrgID == "" || req.User == "" || req.ServiceName == "" {
		http.Error(w, "tenant_org_id, user, and service_name are required", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.Replicas == 0 {
		req.Replicas = 3
	}
	if req.CPU == "" {
		req.CPU = "1000m"
	}
	if req.Memory == "" {
		req.Memory = "2Gi"
	}
	if req.DiskSize == "" {
		req.DiskSize = "10Gi"
	}
	if req.VectorDimension == 0 {
		req.VectorDimension = 128
	}
	if req.VectorCount == 0 {
		req.VectorCount = 1000000
	}

	// 创建集群
	resp, err := s.helmManager.CreateTenantCluster(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create cluster: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetTenantClusterHandler 获取租户集群状态
func (s *APIServer) GetTenantClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	status, err := s.helmManager.GetTenantClusterStatus(namespace)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get cluster status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ListTenantClustersHandler 列出所有租户集群
func (s *APIServer) ListTenantClustersHandler(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.helmManager.ListTenantClusters()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list clusters: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clusters": clusters,
		"count":    len(clusters),
	})
}

// DeleteTenantClusterHandler 删除租户集群
func (s *APIServer) DeleteTenantClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	if err := s.helmManager.DeleteTenantCluster(namespace); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete cluster: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Cluster deleted successfully",
		"namespace": namespace,
	})
}

// ScaleTenantClusterHandler 扩缩容租户集群
func (s *APIServer) ScaleTenantClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	var req struct {
		Replicas int `json:"replicas"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Replicas <= 0 {
		http.Error(w, "replicas must be greater than 0", http.StatusBadRequest)
		return
	}

	if err := s.helmManager.ScaleTenantCluster(namespace, req.Replicas); err != nil {
		http.Error(w, fmt.Sprintf("Failed to scale cluster: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Cluster scaled successfully",
		"namespace": namespace,
		"replicas":  req.Replicas,
	})
}

// Run 启动 API 服务器
func (s *APIServer) Run(addr string) error {
	router := s.SetupRoutes()
	fmt.Printf("Starting API server on %s\n", addr)
	return http.ListenAndServe(addr, router)
}

// 使用示例
func mainAPI() {
	server := NewAPIServer()
	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}

/*
使用示例:

1. 创建租户集群:
curl -X POST http://localhost:8080/api/v1/tenant-clusters \
  -H 'Content-Type: application/json' \
  -d '{
    "tenant_org_id": "org-001",
    "user": "alice",
    "service_name": "vector-search",
    "replicas": 3,
    "cpu": "2000m",
    "memory": "4Gi",
    "disk_size": "20Gi",
    "vector_dimension": 256,
    "vector_count": 10000000
  }'

2. 获取集群状态:
curl http://localhost:8080/api/v1/tenant-clusters/org-001-alice-vector-search

3. 列出所有集群:
curl http://localhost:8080/api/v1/tenant-clusters

4. 扩容集群:
curl -X POST http://localhost:8080/api/v1/tenant-clusters/org-001-alice-vector-search/scale \
  -H 'Content-Type: application/json' \
  -d '{"replicas": 5}'

5. 删除集群:
curl -X DELETE http://localhost:8080/api/v1/tenant-clusters/org-001-alice-vector-search
*/
