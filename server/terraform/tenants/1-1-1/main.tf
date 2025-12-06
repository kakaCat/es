
module "tenant_cluster" {
  source = "../../modules/tenant"

  tenant_org_id    = "1"
  user             = "1"
  service_name     = "1"
  replicas         = 1
  cpu              = "500m"
  memory           = "1Gi"
  disk_size        = "10Gi"
  storage_class    = "hostpath"
  gpu_count        = 0
  vector_dimension = 128
  vector_count     = 10000
}

output "namespace" {
  value = module.tenant_cluster.namespace
}
