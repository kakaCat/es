output "fluentd_daemonset_name" {
  description = "Fluentd DaemonSet name"
  value       = var.fluentd_enabled ? kubernetes_daemonset.fluentd[0].metadata[0].name : ""
}
