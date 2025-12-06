# Tenant identification
variable "tenant_org_id" {
  description = "Tenant organization ID"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.tenant_org_id))
    error_message = "tenant_org_id must contain only lowercase letters, numbers, and hyphens"
  }
}

variable "user" {
  description = "User name"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.user))
    error_message = "user must contain only lowercase letters, numbers, and hyphens"
  }
}

variable "service_name" {
  description = "Service name"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.service_name))
    error_message = "service_name must contain only lowercase letters, numbers, and hyphens"
  }
}

# Resource specifications
variable "cpu" {
  description = "CPU allocation (e.g., '2000m', '2')"
  type        = string
  default     = "1000m"
}

variable "memory" {
  description = "Memory allocation (e.g., '2Gi', '2048Mi')"
  type        = string
  default     = "2Gi"
}

variable "disk_size" {
  description = "Disk size (e.g., '10Gi', '100Gi')"
  type        = string
  default     = "10Gi"
}

variable "gpu_count" {
  description = "Number of GPUs"
  type        = number
  default     = 0

  validation {
    condition     = var.gpu_count >= 0
    error_message = "gpu_count must be non-negative"
  }
}

# Vector configuration
variable "vector_dimension" {
  description = "Vector dimension"
  type        = number
  default     = 128

  validation {
    condition     = var.vector_dimension > 0
    error_message = "vector_dimension must be positive"
  }
}

variable "vector_count" {
  description = "Number of vectors in the database"
  type        = number
  default     = 1000000

  validation {
    condition     = var.vector_count > 0
    error_message = "vector_count must be positive"
  }
}

# IVF algorithm parameters
variable "nlist" {
  description = "IVF nlist parameter (number of clusters)"
  type        = number
  default     = 100
}

variable "nprobe" {
  description = "IVF nprobe parameter (number of clusters to search)"
  type        = number
  default     = 10
}

# Cluster configuration
variable "replicas" {
  description = "Number of Elasticsearch replicas"
  type        = number
  default     = 3

  validation {
    condition     = var.replicas > 0
    error_message = "replicas must be positive"
  }
}

variable "storage_class" {
  description = "Storage class for persistent volumes"
  type        = string
  default     = "hostpath"
}

# Quota settings
variable "enable_quota" {
  description = "Enable resource quotas"
  type        = bool
  default     = true
}

variable "quota_cpu" {
  description = "CPU quota for namespace"
  type        = string
  default     = "10"
}

variable "quota_memory" {
  description = "Memory quota for namespace"
  type        = string
  default     = "20Gi"
}

variable "quota_storage" {
  description = "Storage quota for namespace"
  type        = string
  default     = "100Gi"
}

variable "quota_pvcs" {
  description = "Maximum number of PVCs"
  type        = number
  default     = 10
}

variable "quota_pods" {
  description = "Maximum number of pods"
  type        = number
  default     = 20
}

# Network policy
variable "enable_network_policy" {
  description = "Enable network policy for tenant isolation"
  type        = bool
  default     = true
}
