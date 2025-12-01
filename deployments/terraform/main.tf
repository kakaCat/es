terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}

provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kube_context
  }
}

# Create main namespace for ES Serverless platform
resource "kubernetes_namespace" "es_serverless" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/name"       = "es-serverless"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Deploy Elasticsearch cluster using Helm
module "elasticsearch" {
  source = "./modules/elasticsearch"

  namespace       = kubernetes_namespace.es_serverless.metadata[0].name
  release_name    = "elasticsearch"
  chart_version   = var.elasticsearch_chart_version
  replicas        = var.elasticsearch_replicas
  storage_size    = var.elasticsearch_storage_size
  storage_class   = var.storage_class
  resources       = var.elasticsearch_resources

  depends_on = [kubernetes_namespace.es_serverless]
}

# Deploy control plane services
module "control_plane" {
  source = "./modules/control-plane"

  namespace     = kubernetes_namespace.es_serverless.metadata[0].name
  release_name  = "es-control-plane"
  chart_version = var.control_plane_chart_version

  manager_image           = var.manager_image
  shard_controller_image  = var.shard_controller_image
  reporting_service_image = var.reporting_service_image

  elasticsearch_url = module.elasticsearch.service_url

  depends_on = [module.elasticsearch]
}

# Deploy monitoring stack
module "monitoring" {
  source = "./modules/monitoring"

  namespace     = kubernetes_namespace.es_serverless.metadata[0].name
  release_name  = "monitoring"
  chart_version = var.monitoring_chart_version

  prometheus_enabled         = var.prometheus_enabled
  grafana_enabled           = var.grafana_enabled
  prometheus_retention_days = var.prometheus_retention_days

  depends_on = [kubernetes_namespace.es_serverless]
}

# Deploy logging stack
module "logging" {
  source = "./modules/logging"

  namespace     = kubernetes_namespace.es_serverless.metadata[0].name
  release_name  = "logging"
  chart_version = var.logging_chart_version

  fluentd_enabled = var.fluentd_enabled

  depends_on = [kubernetes_namespace.es_serverless]
}
