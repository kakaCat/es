# Tenant-specific Elasticsearch cluster
# This module creates a dedicated ES cluster for a single tenant

locals {
  namespace = "${var.tenant_org_id}-${var.user}-${var.service_name}"
  labels = {
    "es-cluster"      = "true"
    "tenant-org-id"   = var.tenant_org_id
    "user"            = var.user
    "service-name"    = var.service_name
    "managed-by"      = "terraform"
  }
}

# Create dedicated namespace for tenant
resource "kubernetes_namespace" "tenant" {
  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Deploy Elasticsearch cluster for tenant
resource "helm_release" "tenant_elasticsearch" {
  name       = "elasticsearch"
  namespace  = kubernetes_namespace.tenant.metadata[0].name
  chart      = "${path.module}/../../../helm/elasticsearch"
  wait       = true
  timeout    = 600

  values = [
    yamlencode({
      replicaCount = var.replicas

      clusterName = local.namespace

      resources = {
        requests = {
          cpu    = var.cpu
          memory = var.memory
        }
        limits = {
          cpu    = var.cpu
          memory = var.memory
        }
      }

      persistence = {
        enabled      = true
        storageClass = var.storage_class
        size         = var.disk_size
      }

      # IVF plugin configuration
      ivfPlugin = {
        enabled = true
        config = {
          dimension    = var.vector_dimension
          vectorCount  = var.vector_count
          nlist        = var.nlist
          nprobe       = var.nprobe
        }
      }

      # GPU configuration if requested
      nodeSelector = var.gpu_count > 0 ? {
        "nvidia.com/gpu" = "true"
      } : {}

      # Add resource limits for GPU
      resources_gpu = var.gpu_count > 0 ? {
        limits = {
          "nvidia.com/gpu" = var.gpu_count
        }
      } : {}
    })
  ]

  depends_on = [kubernetes_namespace.tenant]
}

# Create ConfigMap for tenant metadata
resource "kubernetes_config_map" "tenant_metadata" {
  metadata {
    name      = "tenant-metadata"
    namespace = kubernetes_namespace.tenant.metadata[0].name
    labels    = local.labels
  }

  data = {
    tenant_org_id   = var.tenant_org_id
    user            = var.user
    service_name    = var.service_name
    cpu             = var.cpu
    memory          = var.memory
    disk_size       = var.disk_size
    gpu_count       = tostring(var.gpu_count)
    vector_dimension = tostring(var.vector_dimension)
    vector_count    = tostring(var.vector_count)
    replicas        = tostring(var.replicas)
    created_at      = timestamp()
  }
}

# Resource quota for tenant namespace
resource "kubernetes_resource_quota" "tenant_quota" {
  count = var.enable_quota ? 1 : 0

  metadata {
    name      = "tenant-quota"
    namespace = kubernetes_namespace.tenant.metadata[0].name
  }

  spec {
    hard = {
      "requests.cpu"    = var.quota_cpu
      "requests.memory" = var.quota_memory
      "requests.storage" = var.quota_storage
      "persistentvolumeclaims" = var.quota_pvcs
      "pods" = var.quota_pods
    }
  }
}

# Network policy for tenant isolation
resource "kubernetes_network_policy" "tenant_isolation" {
  count = var.enable_network_policy ? 1 : 0

  metadata {
    name      = "tenant-isolation"
    namespace = kubernetes_namespace.tenant.metadata[0].name
  }

  spec {
    pod_selector {}

    policy_types = ["Ingress", "Egress"]

    # Allow traffic within namespace
    ingress {
      from {
        pod_selector {}
      }
    }

    # Allow egress to DNS and within namespace
    egress {
      to {
        pod_selector {}
      }
    }

    egress {
      to {
        namespace_selector {
          match_labels = {
            name = "kube-system"
          }
        }
      }
      ports {
        port     = "53"
        protocol = "UDP"
      }
    }
  }
}
