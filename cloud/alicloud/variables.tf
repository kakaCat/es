variable "access_key" {
  description = "Alibaba Cloud Access Key"
  type        = string
  sensitive   = true
}

variable "secret_key" {
  description = "Alibaba Cloud Secret Key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "Alibaba Cloud Region"
  default     = "cn-hangzhou"
}

variable "zone_id" {
  description = "Availability Zone ID"
  default     = "cn-hangzhou-j"
}

variable "cluster_name" {
  description = "Name of the K8s cluster"
  default     = "es-serverless-prod"
}

variable "worker_instance_type" {
  description = "ECS Instance Type for K8s Workers"
  default     = "ecs.g7.xlarge" # 4 vCPU 16GB
}

variable "db_user" {
  description = "Database User"
  default     = "es_user"
}

variable "db_password" {
  description = "Database Password"
  default     = "Es_password_2025" # Must meet complexity requirements
}

variable "db_name" {
  description = "Database Name"
  default     = "es_metadata"
}
