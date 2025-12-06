output "namespace" {
  description = "Tenant namespace"
  value       = kubernetes_namespace.tenant.metadata[0].name
}

output "elasticsearch_service_url" {
  description = "Elasticsearch service URL"
  value       = "http://elasticsearch.${kubernetes_namespace.tenant.metadata[0].name}.svc.cluster.local:9200"
}

output "tenant_org_id" {
  description = "Tenant organization ID"
  value       = var.tenant_org_id
}

output "user" {
  description = "User name"
  value       = var.user
}

output "service_name" {
  description = "Service name"
  value       = var.service_name
}

output "cluster_endpoint" {
  description = "Elasticsearch cluster endpoint"
  value = {
    internal_url = "http://elasticsearch.${kubernetes_namespace.tenant.metadata[0].name}.svc.cluster.local:9200"
    namespace    = kubernetes_namespace.tenant.metadata[0].name
  }
}

output "resource_specs" {
  description = "Resource specifications"
  value = {
    cpu             = var.cpu
    memory          = var.memory
    disk_size       = var.disk_size
    gpu_count       = var.gpu_count
    vector_dimension = var.vector_dimension
    vector_count    = var.vector_count
    replicas        = var.replicas
  }
}
