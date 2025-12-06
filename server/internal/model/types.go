package model

import (
	"time"
)

// CreateRequest represents the request body for creating a cluster
// CreateRequest 创建集群的请求体
type CreateRequest struct {
	TenantOrgID string `json:"tenant_org_id"` // 租户组织ID（多租户隔离）
	User        string `json:"user"`          // 用户名
	ServiceName string `json:"service_name"`  // 服务名称
	Namespace   string `json:"namespace"`     // 命名空间
	Replicas    int    `json:"replicas"`      // 副本数
	CPURequest  string `json:"cpu_request"`   // CPU 请求量
	CPULimit    string `json:"cpu_limit"`     // CPU 限制量
	MemRequest  string `json:"mem_request"`   // 内存请求量
	MemLimit    string `json:"mem_limit"`     // 内存限制量
	DiskSize    string `json:"disk_size"`     // 磁盘大小
	GPUCount    int    `json:"gpu_count"`     // GPU 数量
	Dimension   int    `json:"dimension"`     // 向量维度
	VectorCount int    `json:"vector_count"`  // 向量数量估计
	IndexLimit  int    `json:"index_limit"`   // 索引数量限制
	GitlabURL   string `json:"gitlab_url"`    // Gitlab 地址（可选）
}

// DeleteRequest represents the request body for deleting a cluster
// DeleteRequest 删除集群的请求体
type DeleteRequest struct {
	Namespace string `json:"namespace"` // 命名空间
}

// ScaleRequest represents the request body for scaling a cluster
// ScaleRequest 扩缩容集群的请求体
type ScaleRequest struct {
	Namespace string `json:"namespace"` // 命名空间
	Replicas  int    `json:"replicas"`  // 目标副本数
}

// ClusterStatus represents the status of a cluster
// ClusterStatus 集群状态信息
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

// VectorIndexRequest represents the request body for creating a vector index
// VectorIndexRequest 创建向量索引的请求体
type VectorIndexRequest struct {
	IndexName    string            `json:"index_name"`
	Dimension    int               `json:"dimension"`
	Metric       string            `json:"metric"`     // L2, cosine, dot
	IVFParams    map[string]int    `json:"ivf_params"` // nlist, nprobe
	FieldMapping map[string]string `json:"field_mapping"`
}

// VectorIndexStatus represents the status of a vector index
// VectorIndexStatus 向量索引状态
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
// IVFParams IVF 算法参数
type IVFParams struct {
	NList  int `json:"nlist" gorm:"column:ivf_nlist"`   // 聚类中心数
	NProbe int `json:"nprobe" gorm:"column:ivf_nprobe"` // 搜索探针数
}

// VectorIndexMapping represents the mapping for a vector index
// VectorIndexMapping 向量索引映射配置
type VectorIndexMapping struct {
	Properties map[string]interface{} `json:"properties" gorm:"-"`
}

// IndexMetadata represents index metadata
// IndexMetadata 索引元数据
type IndexMetadata struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	IndexName     string    `json:"index_name" gorm:"index"`
	Namespace     string    `json:"namespace" gorm:"index"`
	Dimension     int       `json:"dimension"`
	Metric        string    `json:"metric"`
	IVFParams     IVFParams `json:"ivf_params" gorm:"embedded"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     string    `json:"created_by"`
	Status        string    `json:"status"` // active, deleted, building
	DocumentCount int       `json:"document_count"`
	StorageSize   string    `json:"storage_size"`
}

func (IndexMetadata) TableName() string {
	return "index_metadata"
}

// TenantQuota represents tenant quota information
// TenantQuota 租户配额信息
type TenantQuota struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	TenantID       string    `json:"tenant_id" gorm:"uniqueIndex"`
	MaxIndices     int       `json:"max_indices"`     // 最大索引数
	MaxStorage     string    `json:"max_storage"`     // 最大存储空间
	CurrentIndices int       `json:"current_indices"` // 当前索引数
	CurrentStorage string    `json:"current_storage"` // 当前存储空间
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (TenantQuota) TableName() string {
	return "tenant_quota"
}

// DeploymentStatus represents deployment status information
// DeploymentStatus 部署状态信息
type DeploymentStatus struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	TenantOrgID string                 `json:"tenant_org_id"` // 租户组织ID
	Namespace   string                 `json:"namespace" gorm:"uniqueIndex"`
	User        string                 `json:"user" gorm:"index"`
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
	Details     map[string]interface{} `json:"details" gorm:"-"`
}

func (DeploymentStatus) TableName() string {
	return "deployment_status"
}

// TenantContainer represents a tenant's container metadata
// TenantContainer 租户容器元数据
type TenantContainer struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantOrgID string    `json:"tenant_org_id" gorm:"index"` // 租户组织ID
	User        string    `json:"user" gorm:"index:idx_user_service"`
	ServiceName string    `json:"service_name" gorm:"index:idx_user_service"`
	Namespace   string    `json:"namespace" gorm:"index"`
	Replicas    int       `json:"replicas"`
	CPU         string    `json:"cpu"`
	Memory      string    `json:"memory"`
	Disk        string    `json:"disk"`
	GPUCount    int       `json:"gpu_count"`
	Dimension   int       `json:"dimension"`
	VectorCount int       `json:"vector_count"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	SyncTime    time.Time `json:"sync_time"`
	Deleted     bool      `json:"deleted" gorm:"index"`
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
}

func (TenantContainer) TableName() string {
	return "tenant_containers"
}

// TenantConfig holds configuration for a tenant cluster
// TenantConfig 租户集群配置
type TenantConfig struct {
	TenantOrgID     string
	User            string
	ServiceName     string
	Replicas        int
	CPU             string
	Memory          string
	DiskSize        string
	StorageClass    string
	GPUCount        int
	VectorDimension int
	VectorCount     int
}

// Metrics represents container resource usage metrics
// Metrics 容器资源使用指标
type Metrics struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Namespace   string    `json:"namespace" gorm:"index"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	QPS         float64   `json:"qps"`
	Timestamp   time.Time `json:"timestamp" gorm:"index"`
}

func (Metrics) TableName() string {
	return "monitoring_metrics"
}

// ContainerMetrics represents detailed container metrics including startup data
// ContainerMetrics 详细容器指标（包含启动数据）
type ContainerMetrics struct {
	ID               string           `json:"id"`
	Namespace        string           `json:"namespace"`
	ContainerName    string           `json:"container_name"`
	CPUUsage         float64          `json:"cpu_usage"`
	MemoryUsage      float64          `json:"memory_usage"`
	DiskUsage        float64          `json:"disk_usage"`
	QPS              float64          `json:"qps"`
	StartupCPU       float64          `json:"startup_cpu"`
	StartupMemory    float64          `json:"startup_memory"`
	StartupDisk      float64          `json:"startup_disk"`
	PluginQPS        float64          `json:"plugin_qps"`
	Timestamp        time.Time        `json:"timestamp"`
	Status           string           `json:"status"`
	ResourceLimits   ResourceLimits   `json:"resource_limits"`
	ResourceRequests ResourceRequests `json:"resource_requests"`
}

// ResourceLimits represents container resource limits
type ResourceLimits struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// AutoscalerConfig holds the configuration for autoscaling
type AutoscalerConfig struct {
	// CPU thresholds
	HighCPUThreshold float64 `json:"high_cpu_threshold"` // Percentage (0-100)
	LowCPUThreshold  float64 `json:"low_cpu_threshold"`  // Percentage (0-100)

	// Memory thresholds
	HighMemoryThreshold float64 `json:"high_memory_threshold"` // Percentage (0-100)
	LowMemoryThreshold  float64 `json:"low_memory_threshold"`  // Percentage (0-100)

	// QPS thresholds
	HighQPSThreshold float64 `json:"high_qps_threshold"`
	LowQPSThreshold  float64 `json:"low_qps_threshold"`

	// Disk thresholds
	HighDiskThreshold float64 `json:"high_disk_threshold"` // Percentage (0-100)
	LowDiskThreshold  float64 `json:"low_disk_threshold"`  // Percentage (0-100)

	// Scaling factors
	ScaleUpFactor   float64 `json:"scale_up_factor"`   // Multiplier for scale up (e.g., 1.5)
	ScaleDownFactor float64 `json:"scale_down_factor"` // Multiplier for scale down (e.g., 0.5)

	// Limits
	MinReplicas int `json:"min_replicas"`
	MaxReplicas int `json:"max_replicas"`

	// Cooldown period in seconds
	ScaleUpCooldown   int `json:"scale_up_cooldown"`   // Cooldown period after scaling up
	ScaleDownCooldown int `json:"scale_down_cooldown"` // Cooldown period after scaling down

	// User-specific scaling policies
	ScalingPolicies map[string]ScalingPolicy `json:"scaling_policies"`
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

// ResourceRequests represents container resource requests
type ResourceRequests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}
