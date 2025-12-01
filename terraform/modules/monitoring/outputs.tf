output "prometheus_service_url" {
  description = "Prometheus service URL"
  value       = "http://${var.release_name}-prometheus.${var.namespace}.svc.cluster.local:9090"
}

output "grafana_service_url" {
  description = "Grafana service URL"
  value       = "http://${var.release_name}-grafana.${var.namespace}.svc.cluster.local:3000"
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.monitoring.name
}
