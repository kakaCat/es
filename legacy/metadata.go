package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TenantContainer represents a tenant's container metadata
type TenantContainer struct {
	ID          string    `json:"id"`
	TenantOrgID string    `json:"tenant_org_id"` // 租户组织ID
	User        string    `json:"user"`
	ServiceName string    `json:"service_name"`
	Namespace   string    `json:"namespace"`
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
	Deleted     bool      `json:"deleted"`
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
}

// MetadataService provides file-based metadata storage
type MetadataService struct {
	dataDir string
	mu      sync.RWMutex
}

// NewMetadataService creates a new metadata service
func NewMetadataService(dataDir string) *MetadataService {
	return &MetadataService{
		dataDir: dataDir,
	}
}

// SaveTenantContainer saves tenant container metadata to a file
func (m *MetadataService) SaveTenantContainer(container *TenantContainer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("tenant_%s_%s.json", container.User, container.ServiceName))
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetTenantContainer retrieves tenant container metadata
func (m *MetadataService) GetTenantContainer(user, serviceName string) (*TenantContainer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("tenant_%s_%s.json", user, serviceName))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var container TenantContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	return &container, nil
}

// DeleteTenantContainer marks a tenant container as deleted (logical deletion)
func (m *MetadataService) DeleteTenantContainer(user, serviceName string) error {
	container, err := m.GetTenantContainer(user, serviceName)
	if err != nil {
		return err
	}

	container.Deleted = true
	container.DeletedAt = time.Now()
	return m.SaveTenantContainer(container)
}

// ListTenantContainersByOrgID lists all tenant containers for an organization
func (m *MetadataService) ListTenantContainersByOrgID(tenantOrgID string) ([]*TenantContainer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files, err := os.ReadDir(m.dataDir)
	if err != nil {
		return nil, err
	}

	var containers []*TenantContainer
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && filepath.Base(file.Name())[:7] == "tenant_" {
			data, err := os.ReadFile(filepath.Join(m.dataDir, file.Name()))
			if err != nil {
				continue
			}

			var container TenantContainer
			if err := json.Unmarshal(data, &container); err != nil {
				continue
			}

			if container.TenantOrgID == tenantOrgID && !container.Deleted {
				containers = append(containers, &container)
			}
		}
	}

	return containers, nil
}

// SaveDeploymentStatus saves deployment status metadata
func (m *MetadataService) SaveDeploymentStatus(status *DeploymentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("deploy_%s.json", status.Namespace))
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetDeploymentStatus retrieves deployment status metadata
func (m *MetadataService) GetDeploymentStatus(namespace string) (*DeploymentStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("deploy_%s.json", namespace))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var status DeploymentStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// ListDeploymentStatus lists all deployment statuses
func (m *MetadataService) ListDeploymentStatus() ([]*DeploymentStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files, err := os.ReadDir(m.dataDir)
	if err != nil {
		return nil, err
	}

	var deployments []*DeploymentStatus
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && filepath.Base(file.Name())[:7] == "deploy_" {
			data, err := os.ReadFile(filepath.Join(m.dataDir, file.Name()))
			if err != nil {
				continue
			}

			var status DeploymentStatus
			if err := json.Unmarshal(data, &status); err != nil {
				continue
			}

			deployments = append(deployments, &status)
		}
	}

	return deployments, nil
}

// SaveIndexMetadata saves index metadata
func (m *MetadataService) SaveIndexMetadata(metadata *IndexMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("index_%s_%s.json", metadata.Namespace, metadata.IndexName))
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetIndexMetadata retrieves index metadata
func (m *MetadataService) GetIndexMetadata(id string) (*IndexMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files, err := os.ReadDir(m.dataDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && filepath.Base(file.Name())[:6] == "index_" {
			data, err := os.ReadFile(filepath.Join(m.dataDir, file.Name()))
			if err != nil {
				continue
			}

			var metadata IndexMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue
			}

			if metadata.ID == id {
				return &metadata, nil
			}
		}
	}

	return nil, errors.New("index metadata not found")
}

// ListIndexMetadata lists all index metadata
func (m *MetadataService) ListIndexMetadata() ([]*IndexMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files, err := os.ReadDir(m.dataDir)
	if err != nil {
		return nil, err
	}

	var metadataList []*IndexMetadata
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && filepath.Base(file.Name())[:6] == "index_" {
			data, err := os.ReadFile(filepath.Join(m.dataDir, file.Name()))
			if err != nil {
				continue
			}

			var metadata IndexMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue
			}

			metadataList = append(metadataList, &metadata)
		}
	}

	return metadataList, nil
}

// DeleteIndexMetadata deletes index metadata
func (m *MetadataService) DeleteIndexMetadata(id string) error {
	metadata, err := m.GetIndexMetadata(id)
	if err != nil {
		return err
	}

	filename := filepath.Join(m.dataDir, fmt.Sprintf("index_%s_%s.json", metadata.Namespace, metadata.IndexName))
	return os.Remove(filename)
}

// SaveTenantQuota saves tenant quota
func (m *MetadataService) SaveTenantQuota(quota *TenantQuota) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("quota_%s.json", quota.TenantID))
	data, err := json.MarshalIndent(quota, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetTenantQuota retrieves tenant quota
func (m *MetadataService) GetTenantQuota(tenantID string) (*TenantQuota, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("quota_%s.json", tenantID))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var quota TenantQuota
	if err := json.Unmarshal(data, &quota); err != nil {
		return nil, err
	}

	return &quota, nil
}

// CheckTenantQuota checks if tenant has available quota
func (m *MetadataService) CheckTenantQuota(tenantID string) (bool, *TenantQuota, error) {
	quota, err := m.GetTenantQuota(tenantID)
	if err != nil {
		// If quota doesn't exist, create default quota
		quota = &TenantQuota{
			TenantID:       tenantID,
			MaxIndices:     100,
			MaxStorage:     "1Ti",
			CurrentIndices: 0,
			CurrentStorage: "0Gi",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		m.SaveTenantQuota(quota)
		return true, quota, nil
	}

	// Check if current usage is within limits
	hasQuota := quota.CurrentIndices < quota.MaxIndices
	return hasQuota, quota, nil
}

// UpdateTenantQuotaUsage updates tenant quota usage
func (m *MetadataService) UpdateTenantQuotaUsage(tenantID string, increase bool, storage string) error {
	quota, err := m.GetTenantQuota(tenantID)
	if err != nil {
		// Create default quota if it doesn't exist
		quota = &TenantQuota{
			TenantID:       tenantID,
			MaxIndices:     100,
			MaxStorage:     "1Ti",
			CurrentIndices: 0,
			CurrentStorage: "0Gi",
			CreatedAt:      time.Now(),
		}
	}

	if increase {
		quota.CurrentIndices++
	} else {
		quota.CurrentIndices--
	}

	quota.UpdatedAt = time.Now()
	return m.SaveTenantQuota(quota)
}

// SaveMetrics saves monitoring metrics
func (m *MetadataService) SaveMetrics(metrics *Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("metrics_%s.json", metrics.Namespace))
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetLatestMetrics retrieves the latest metrics for a namespace
func (m *MetadataService) GetLatestMetrics(namespace string) (*Metrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("metrics_%s.json", namespace))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var metrics Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// GetContainerMetrics retrieves container metrics for a namespace
func (m *MetadataService) GetContainerMetrics(namespace string) (*ContainerMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filename := filepath.Join(m.dataDir, fmt.Sprintf("container_metrics_%s.json", namespace))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var metrics ContainerMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}
