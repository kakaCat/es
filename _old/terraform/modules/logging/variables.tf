variable "namespace" {
  description = "Kubernetes namespace"
  type        = string
}

variable "release_name" {
  description = "Release name prefix"
  type        = string
}

variable "chart_version" {
  description = "Chart version"
  type        = string
}

variable "fluentd_enabled" {
  description = "Enable Fluentd logging"
  type        = bool
  default     = true
}

variable "service_account_name" {
  description = "Service account name for Fluentd"
  type        = string
  default     = "fluentd"
}
