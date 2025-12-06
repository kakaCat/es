terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "1.213.0"
    }
  }
}

# 配置阿里云 Provider
# 凭证通过 variables.tf 传入，建议在 terraform.tfvars 中设置
provider "alicloud" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
}

# 1. VPC 网络 (Virtual Private Cloud)
# 创建专有网络，用于隔离云上资源
resource "alicloud_vpc" "main" {
  vpc_name   = "${var.cluster_name}-vpc"
  cidr_block = "10.0.0.0/8" # VPC 网段
}

# 创建交换机 (VSwitch)，位于指定可用区
resource "alicloud_vswitch" "main" {
  vswitch_name = "${var.cluster_name}-vswitch"
  vpc_id       = alicloud_vpc.main.id
  cidr_block   = "10.1.0.0/16" # 子网网段
  zone_id      = var.zone_id
}

# 2. 托管版 Kubernetes (ACK)
# 创建 ACK Pro 版集群
resource "alicloud_cs_managed_kubernetes" "main" {
  name                      = var.cluster_name
  cluster_spec              = "ack.pro.small"
  version                   = "1.28.3-aliyun.1"
  worker_vswitch_ids        = [alicloud_vswitch.main.id]
  new_nat_gateway           = true # 自动创建 NAT 网关，允许节点访问公网
  pod_vswitch_ids           = [alicloud_vswitch.main.id]
  service_cidr              = "172.21.0.0/20"
  slb_internet_enabled      = true # 允许 API Server 公网访问
  
  # 集群组件 (Addons)
  addons {
    name = "terway-eniip" # 网络插件，高性能 Terway
  }
  addons {
    name = "csi-plugin" # 存储插件 CSI
  }
  addons {
    name = "csi-provisioner" # 存储动态供应
  }
}

# ACK 节点池 (Node Pool)
# 管理集群的工作节点，支持自动伸缩
resource "alicloud_cs_kubernetes_node_pool" "default" {
  cluster_id            = alicloud_cs_managed_kubernetes.main.id
  node_pool_name        = "default-pool"
  vswitch_ids           = [alicloud_vswitch.main.id]
  instance_types        = [var.worker_instance_type] # ECS 实例规格
  system_disk_category  = "cloud_essd" # 系统盘类型：ESSD
  system_disk_size      = 40           # 系统盘大小：40GB
  
  # 自动伸缩配置
  scaling_config {
    min_size = 1 # 最小节点数
    max_size = 5 # 最大节点数
  }
}

# 3. RDS PostgreSQL 数据库
# 创建高可用版云数据库实例，用于存储元数据
resource "alicloud_db_instance" "metadata" {
  engine           = "PostgreSQL"
  engine_version   = "15.0"
  instance_type    = "pg.n2.small.2c" # 规格：2核 4GB
  instance_storage = 20               # 存储空间：20GB
  instance_name    = "${var.cluster_name}-db"
  vswitch_id       = alicloud_vswitch.main.id
  security_ips     = [alicloud_vpc.main.cidr_block] # 白名单：允许 VPC 内访问
}

# 创建数据库账号
resource "alicloud_db_account" "default" {
  db_instance_id    = alicloud_db_instance.metadata.id
  account_name      = var.db_user
  account_password  = var.db_password
  account_type      = "Normal"
}

# 创建初始数据库
resource "alicloud_db_database" "default" {
  instance_id = alicloud_db_instance.metadata.id
  name        = var.db_name
}

# 4. 保存 Kubeconfig
# 将集群访问凭证保存到本地，供 Helm 部署使用
resource "local_file" "kubeconfig" {
  content  = alicloud_cs_managed_kubernetes.main.config
  filename = "${path.module}/kubeconfig"
}
