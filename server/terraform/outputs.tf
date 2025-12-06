output "namespace" {
  description = "The Kubernetes namespace where ES Serverless is deployed"
  value       = kubernetes_namespace.es_serverless.metadata[0].name
}

output "elasticsearch_service_url" {
  description = "Elasticsearch service URL"
  value       = module.elasticsearch.service_url
}

output "manager_service_url" {
  description = "Manager API service URL"
  value       = module.control_plane.manager_service_url
}

output "grafana_service_url" {
  description = "Grafana dashboard URL"
  value       = module.monitoring.grafana_service_url
}

output "prometheus_service_url" {
  description = "Prometheus service URL"
  value       = module.monitoring.prometheus_service_url
}
