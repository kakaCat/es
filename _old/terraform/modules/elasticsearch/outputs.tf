output "service_url" {
  description = "Elasticsearch service URL"
  value       = "http://${var.release_name}.${var.namespace}.svc.cluster.local:9200"
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.elasticsearch.name
}

output "namespace" {
  description = "Namespace"
  value       = var.namespace
}
