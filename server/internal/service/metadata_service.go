package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"es-serverless-manager/internal/model"
)

// MetadataService provides database-backed metadata storage
// MetadataService 提供基于数据库的元数据存储服务
type MetadataService struct {
	db *gorm.DB
}

// NewMetadataService creates a new metadata service
// NewMetadataService 创建一个新的元数据服务实例
func NewMetadataService(db *gorm.DB) *MetadataService {
	return &MetadataService{
		db: db,
	}
}

// SaveTenantContainer saves tenant container metadata
// SaveTenantContainer 保存租户容器元数据
func (m *MetadataService) SaveTenantContainer(container *model.TenantContainer) error {
	return m.db.Save(container).Error
}

// GetTenantContainer retrieves tenant container metadata
// GetTenantContainer 获取租户容器元数据
func (m *MetadataService) GetTenantContainer(user, serviceName string) (*model.TenantContainer, error) {
	var container model.TenantContainer
	// Find by user and service_name
	// 通过用户名和服务名查找
	result := m.db.Where("user = ? AND service_name = ?", user, serviceName).First(&container)
	if result.Error != nil {
		return nil, result.Error
	}
	return &container, nil
}

// DeleteTenantContainer marks a tenant container as deleted (logical deletion)
// DeleteTenantContainer 标记租户容器为已删除（逻辑删除）
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
// ListTenantContainersByOrgID 列出指定组织的所有租户容器（不包含已删除的）
func (m *MetadataService) ListTenantContainersByOrgID(tenantOrgID string) ([]*model.TenantContainer, error) {
	var containers []*model.TenantContainer
	result := m.db.Where("tenant_org_id = ? AND deleted = ?", tenantOrgID, false).Find(&containers)
	if result.Error != nil {
		return nil, result.Error
	}
	return containers, nil
}

// SaveDeploymentStatus saves deployment status metadata
// SaveDeploymentStatus 保存部署状态元数据
func (m *MetadataService) SaveDeploymentStatus(status *model.DeploymentStatus) error {
	return m.db.Save(status).Error
}

// GetDeploymentStatus retrieves deployment status metadata
// GetDeploymentStatus 获取部署状态元数据
func (m *MetadataService) GetDeploymentStatus(namespace string) (*model.DeploymentStatus, error) {
	var status model.DeploymentStatus
	result := m.db.Where("namespace = ?", namespace).First(&status)
	if result.Error != nil {
		return nil, result.Error
	}
	return &status, nil
}

// ListDeploymentStatus lists all deployment statuses
// ListDeploymentStatus 列出所有部署状态
func (m *MetadataService) ListDeploymentStatus() ([]*model.DeploymentStatus, error) {
	var deployments []*model.DeploymentStatus
	result := m.db.Find(&deployments)
	if result.Error != nil {
		return nil, result.Error
	}
	return deployments, nil
}

// SaveIndexMetadata saves index metadata
// SaveIndexMetadata 保存索引元数据
func (m *MetadataService) SaveIndexMetadata(metadata *model.IndexMetadata) error {
	return m.db.Save(metadata).Error
}

// GetIndexMetadata retrieves index metadata
// GetIndexMetadata 获取索引元数据
func (m *MetadataService) GetIndexMetadata(id string) (*model.IndexMetadata, error) {
	var metadata model.IndexMetadata
	result := m.db.Where("id = ?", id).First(&metadata)
	if result.Error != nil {
		return nil, result.Error
	}
	return &metadata, nil
}

// ListIndexMetadata lists all index metadata
// ListIndexMetadata 列出所有索引元数据
func (m *MetadataService) ListIndexMetadata() ([]*model.IndexMetadata, error) {
	var metadataList []*model.IndexMetadata
	result := m.db.Find(&metadataList)
	if result.Error != nil {
		return nil, result.Error
	}
	return metadataList, nil
}

// DeleteIndexMetadata deletes index metadata
// DeleteIndexMetadata 删除索引元数据
func (m *MetadataService) DeleteIndexMetadata(id string) error {
	return m.db.Delete(&model.IndexMetadata{}, "id = ?", id).Error
}

// SaveTenantQuota saves tenant quota
// SaveTenantQuota 保存租户配额
func (m *MetadataService) SaveTenantQuota(quota *model.TenantQuota) error {
	return m.db.Save(quota).Error
}

// GetTenantQuota retrieves tenant quota
// GetTenantQuota 获取租户配额
func (m *MetadataService) GetTenantQuota(tenantID string) (*model.TenantQuota, error) {
	var quota model.TenantQuota
	result := m.db.Where("tenant_id = ?", tenantID).First(&quota)
	if result.Error != nil {
		return nil, result.Error
	}
	return &quota, nil
}

// CheckTenantQuota checks if tenant has available quota
// CheckTenantQuota 检查租户是否有可用配额
func (m *MetadataService) CheckTenantQuota(tenantID string) (bool, *model.TenantQuota, error) {
	quota, err := m.GetTenantQuota(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If quota doesn't exist, create default quota
			// 如果配额不存在，创建默认配额
			quota = &model.TenantQuota{
				TenantID:       tenantID,
				MaxIndices:     100,
				MaxStorage:     "1Ti",
				CurrentIndices: 0,
				CurrentStorage: "0Gi",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			// Generate ID for new quota
			// 为新配额生成 ID
			quota.ID = "quota_" + tenantID

			m.SaveTenantQuota(quota)
			return true, quota, nil
		}
		return false, nil, err
	}

	// Check if current usage is within limits
	// 检查当前使用量是否在限制范围内
	hasQuota := quota.CurrentIndices < quota.MaxIndices
	return hasQuota, quota, nil
}

// UpdateTenantQuotaUsage updates tenant quota usage
// UpdateTenantQuotaUsage 更新租户配额使用量
func (m *MetadataService) UpdateTenantQuotaUsage(tenantID string, increase bool, storage string) error {
	quota, err := m.GetTenantQuota(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default quota if it doesn't exist
			// 如果不存在，创建默认配额
			quota = &model.TenantQuota{
				TenantID:       tenantID,
				MaxIndices:     100,
				MaxStorage:     "1Ti",
				CurrentIndices: 0,
				CurrentStorage: "0Gi",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			quota.ID = "quota_" + tenantID
		} else {
			return err
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
func (m *MetadataService) SaveMetrics(metrics *model.Metrics) error {
	return m.db.Create(metrics).Error
}

// GetLatestMetrics retrieves the latest metrics for a namespace
func (m *MetadataService) GetLatestMetrics(namespace string) (*model.Metrics, error) {
	var metrics model.Metrics
	// Order by timestamp desc, limit 1
	result := m.db.Where("namespace = ?", namespace).Order("timestamp desc").First(&metrics)
	if result.Error != nil {
		return nil, result.Error
	}
	return &metrics, nil
}
