resource "helm_release" "elasticsearch" {
  name       = var.release_name
  namespace  = var.namespace
  chart      = "${path.module}/../../../helm/elasticsearch"
  version    = var.chart_version
  wait       = true
  timeout    = 600

  values = [
    yamlencode({
      replicaCount = var.replicas
      persistence = {
        enabled      = true
        storageClass = var.storage_class
        size         = var.storage_size
      }
      resources = var.resources
    })
  ]
}
