resource "helm_release" "control_plane" {
  name       = var.release_name
  namespace  = var.namespace
  chart      = "${path.module}/../../../helm/control-plane"
  version    = var.chart_version
  wait       = true
  timeout    = 300

  values = [
    yamlencode({
      manager = {
        image = {
          repository = split(":", var.manager_image)[0]
          tag        = length(split(":", var.manager_image)) > 1 ? split(":", var.manager_image)[1] : "latest"
        }
        env = [
          {
            name  = "ELASTICSEARCH_URL"
            value = var.elasticsearch_url
          }
        ]
      }
      shardController = {
        image = {
          repository = split(":", var.shard_controller_image)[0]
          tag        = length(split(":", var.shard_controller_image)) > 1 ? split(":", var.shard_controller_image)[1] : "latest"
        }
        env = [
          {
            name  = "ELASTICSEARCH_URL"
            value = var.elasticsearch_url
          }
        ]
      }
      reportingService = {
        image = {
          repository = split(":", var.reporting_service_image)[0]
          tag        = length(split(":", var.reporting_service_image)) > 1 ? split(":", var.reporting_service_image)[1] : "latest"
        }
      }
    })
  ]
}
