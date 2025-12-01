variable "namespace" {
  description = "Kubernetes namespace"
  type        = string
}

variable "release_name" {
  description = "Helm release name"
  type        = string
}

variable "chart_version" {
  description = "Chart version"
  type        = string
}

variable "prometheus_enabled" {
  description = "Enable Prometheus"
  type        = bool
  default     = true
}

variable "grafana_enabled" {
  description = "Enable Grafana"
  type        = bool
  default     = true
}

variable "prometheus_retention_days" {
  description = "Prometheus retention in days"
  type        = number
  default     = 15
}
