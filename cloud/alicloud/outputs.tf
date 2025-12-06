output "cluster_id" {
  value = alicloud_cs_managed_kubernetes.main.id
}

output "kubeconfig_path" {
  value = abspath(local_file.kubeconfig.filename)
}

output "db_connection_string" {
  value = alicloud_db_instance.metadata.connection_string
}

output "db_port" {
  value = alicloud_db_instance.metadata.port
}
