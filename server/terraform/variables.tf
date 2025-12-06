variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kube_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = "docker-desktop"
}

variable "namespace" {
  description = "Kubernetes namespace for ES Serverless platform"
  type        = string
  default     = "es-serverless"
}

# Elasticsearch variables
variable "elasticsearch_chart_version" {
  description = "Elasticsearch Helm chart version"
  type        = string
  default     = "1.0.0"
}

variable "elasticsearch_replicas" {
  description = "Number of Elasticsearch replicas"
  type        = number
  default     = 3
}

variable "elasticsearch_storage_size" {
  description = "Storage size for each Elasticsearch node"
  type        = string
  default     = "10Gi"
}

variable "storage_class" {
  description = "Storage class for persistent volumes"
  type        = string
  default     = "hostpath"
}

variable "elasticsearch_resources" {
  description = "Resource requests and limits for Elasticsearch"
  type = object({
    requests = object({
      cpu    = string
      memory = string
    })
    limits = object({
      cpu    = string
      memory = string
    })
  })
  default = {
    requests = {
      cpu    = "1000m"
      memory = "2Gi"
    }
    limits = {
      cpu    = "2000m"
      memory = "4Gi"
    }
  }
}

# Control plane variables
variable "control_plane_chart_version" {
  description = "Control plane Helm chart version"
  type        = string
  default     = "1.0.0"
}

variable "manager_image" {
  description = "Manager service container image"
  type        = string
  default     = "es-serverless-manager:latest"
}

variable "shard_controller_image" {
  description = "Shard controller container image"
  type        = string
  default     = "shard-controller:latest"
}

variable "reporting_service_image" {
  description = "Reporting service container image"
  type        = string
  default     = "reporting-service:latest"
}

# Monitoring variables
variable "monitoring_chart_version" {
  description = "Monitoring stack Helm chart version"
  type        = string
  default     = "1.0.0"
}

variable "prometheus_enabled" {
  description = "Enable Prometheus monitoring"
  type        = bool
  default     = true
}

variable "grafana_enabled" {
  description = "Enable Grafana dashboards"
  type        = bool
  default     = true
}

variable "prometheus_retention_days" {
  description = "Prometheus data retention in days"
  type        = number
  default     = 15
}

# Logging variables
variable "logging_chart_version" {
  description = "Logging stack Helm chart version"
  type        = string
  default     = "1.0.0"
}

variable "fluentd_enabled" {
  description = "Enable Fluentd logging"
  type        = bool
  default     = true
}
