package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"es-serverless-manager/internal/model"
)

// TerraformManager handles Terraform operations for tenant clusters
// TerraformManager 处理租户集群的 Terraform 操作
type TerraformManager struct {
	BaseDir string
}

// NewTerraformManager creates a new TerraformManager
// NewTerraformManager 创建一个新的 TerraformManager
func NewTerraformManager(baseDir string) *TerraformManager {
	return &TerraformManager{
		BaseDir: baseDir,
	}
}

const tenantMainTfTemplate = `
module "tenant_cluster" {
  source = "../../modules/tenant"

  tenant_org_id    = "{{.TenantOrgID}}"
  user             = "{{.User}}"
  service_name     = "{{.ServiceName}}"
  replicas         = {{.Replicas}}
  cpu              = "{{.CPU}}"
  memory           = "{{.Memory}}"
  disk_size        = "{{.DiskSize}}"
  storage_class    = "{{.StorageClass}}"
  gpu_count        = {{.GPUCount}}
  vector_dimension = {{.VectorDimension}}
  vector_count     = {{.VectorCount}}
}

output "namespace" {
  value = module.tenant_cluster.namespace
}
`

// CreateCluster creates a new cluster using Terraform
// CreateCluster 使用 Terraform 创建新集群
func (m *TerraformManager) CreateCluster(config model.TenantConfig) error {
	// Check if terraform is installed
	// 检查 Terraform 是否安装
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform not found in PATH")
	}

	// Create tenant directory
	// 创建租户目录
	tenantDir := filepath.Join(m.BaseDir, "tenants", fmt.Sprintf("%s-%s-%s", config.TenantOrgID, config.User, config.ServiceName))
	if err := os.MkdirAll(tenantDir, 0755); err != nil {
		return fmt.Errorf("failed to create tenant directory: %w", err)
	}

	// Generate main.tf
	// 生成 main.tf 文件
	tmpl, err := template.New("main.tf").Parse(tenantMainTfTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(filepath.Join(tenantDir, "main.tf"))
	if err != nil {
		return fmt.Errorf("failed to create main.tf: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Initialize Terraform
	// 初始化 Terraform
	if err := m.runTerraform(tenantDir, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Apply Terraform
	// 应用 Terraform 配置
	if err := m.runTerraform(tenantDir, "apply", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// DeleteCluster deletes a cluster using Terraform
// DeleteCluster 使用 Terraform 删除集群
func (m *TerraformManager) DeleteCluster(namespace string) error {
	// Check if terraform is installed
	// 检查 Terraform 是否安装
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform not found in PATH")
	}

	tenantDir := filepath.Join(m.BaseDir, "tenants", namespace)

	// Check if directory exists
	// 检查目录是否存在
	if _, err := os.Stat(tenantDir); os.IsNotExist(err) {
		return fmt.Errorf("tenant directory does not exist: %s", tenantDir)
	}

	// Destroy Terraform
	// 销毁 Terraform 资源
	if err := m.runTerraform(tenantDir, "destroy", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	// Remove directory
	// 删除目录
	return os.RemoveAll(tenantDir)
}

func (m *TerraformManager) runTerraform(dir string, args ...string) error {
	// Check if terraform is installed
	// 检查 Terraform 是否安装
	_, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("terraform not found in PATH")
	}

	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
