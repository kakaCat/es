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

variable "replicas" {
  description = "Number of Elasticsearch replicas"
  type        = number
}

variable "storage_size" {
  description = "Storage size"
  type        = string
}

variable "storage_class" {
  description = "Storage class"
  type        = string
}

variable "resources" {
  description = "Resource requests and limits"
  type        = any
}
