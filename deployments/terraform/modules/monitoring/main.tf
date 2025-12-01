resource "helm_release" "monitoring" {
  name       = var.release_name
  namespace  = var.namespace
  chart      = "${path.module}/../../../helm/monitoring"
  version    = var.chart_version
  wait       = true
  timeout    = 300

  values = [
    yamlencode({
      prometheus = {
        enabled = var.prometheus_enabled
        retention = {
          days = var.prometheus_retention_days
        }
      }
      grafana = {
        enabled = var.grafana_enabled
      }
    })
  ]
}
