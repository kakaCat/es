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

variable "manager_image" {
  description = "Manager service image"
  type        = string
}

variable "shard_controller_image" {
  description = "Shard controller image"
  type        = string
}

variable "reporting_service_image" {
  description = "Reporting service image"
  type        = string
}

variable "elasticsearch_url" {
  description = "Elasticsearch service URL"
  type        = string
}
