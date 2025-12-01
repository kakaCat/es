output "manager_service_url" {
  description = "Manager service URL"
  value       = "http://${var.release_name}-manager.${var.namespace}.svc.cluster.local:8080"
}

output "reporting_service_url" {
  description = "Reporting service URL"
  value       = "http://${var.release_name}-reporting.${var.namespace}.svc.cluster.local:8081"
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.control_plane.name
}
