#!/usr/bin/env bash
set -euo pipefail

# Create a new tenant cluster using Terraform

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_TENANTS_DIR="$SCRIPT_DIR/../terraform/tenants"

show_help() {
    echo "Usage: scripts/create-tenant.sh --org ORG_ID --user USER --service SERVICE [OPTIONS]"
    echo ""
    echo "Required arguments:"
    echo "  --org ORG_ID          Tenant organization ID"
    echo "  --user USER           User name"
    echo "  --service SERVICE     Service name"
    echo ""
    echo "Optional arguments:"
    echo "  --cpu CPU             CPU allocation (default: 1000m)"
    echo "  --memory MEMORY       Memory allocation (default: 2Gi)"
    echo "  --disk DISK           Disk size (default: 10Gi)"
    echo "  --gpu GPU             GPU count (default: 0)"
    echo "  --dimension DIM       Vector dimension (default: 128)"
    echo "  --vectors COUNT       Vector count (default: 1000000)"
    echo "  --replicas COUNT      Number of replicas (default: 3)"
    echo ""
    echo "Example:"
    echo "  ./scripts/create-tenant.sh --org org-001 --user alice --service vector-search \\"
    echo "    --cpu 2000m --memory 4Gi --disk 20Gi --dimension 256 --replicas 3"
}

# Parse arguments
TENANT_ORG_ID=""
USER=""
SERVICE_NAME=""
CPU="1000m"
MEMORY="2Gi"
DISK="10Gi"
GPU="0"
DIMENSION="128"
VECTORS="1000000"
REPLICAS="3"

while [[ $# -gt 0 ]]; do
    case $1 in
        --org)
            TENANT_ORG_ID="$2"
            shift 2
            ;;
        --user)
            USER="$2"
            shift 2
            ;;
        --service)
            SERVICE_NAME="$2"
            shift 2
            ;;
        --cpu)
            CPU="$2"
            shift 2
            ;;
        --memory)
            MEMORY="$2"
            shift 2
            ;;
        --disk)
            DISK="$2"
            shift 2
            ;;
        --gpu)
            GPU="$2"
            shift 2
            ;;
        --dimension)
            DIMENSION="$2"
            shift 2
            ;;
        --vectors)
            VECTORS="$2"
            shift 2
            ;;
        --replicas)
            REPLICAS="$2"
            shift 2
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate required arguments
if [[ -z "$TENANT_ORG_ID" ]] || [[ -z "$USER" ]] || [[ -z "$SERVICE_NAME" ]]; then
    echo "Error: Missing required arguments"
    show_help
    exit 1
fi

# Create tenant directory
TENANT_NAME="${TENANT_ORG_ID}-${USER}-${SERVICE_NAME}"
TENANT_DIR="$TERRAFORM_TENANTS_DIR/$TENANT_NAME"

mkdir -p "$TENANT_DIR"

# Generate Terraform configuration for tenant
cat > "$TENANT_DIR/main.tf" <<EOF
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
  config_path = "~/.kube/config"
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}

module "tenant" {
  source = "../../modules/tenant"

  tenant_org_id    = "${TENANT_ORG_ID}"
  user             = "${USER}"
  service_name     = "${SERVICE_NAME}"

  cpu              = "${CPU}"
  memory           = "${MEMORY}"
  disk_size        = "${DISK}"
  gpu_count        = ${GPU}

  vector_dimension = ${DIMENSION}
  vector_count     = ${VECTORS}
  replicas         = ${REPLICAS}

  storage_class    = "hostpath"

  enable_quota          = true
  enable_network_policy = true
}

output "namespace" {
  value = module.tenant.namespace
}

output "elasticsearch_url" {
  value = module.tenant.elasticsearch_service_url
}

output "resource_specs" {
  value = module.tenant.resource_specs
}
EOF

echo "Tenant configuration created at: $TENANT_DIR"
echo ""
echo "Deploying tenant cluster..."
cd "$TENANT_DIR"

terraform init
terraform plan
terraform apply -auto-approve

echo ""
echo "Tenant cluster created successfully!"
echo ""
echo "Tenant details:"
echo "  Organization: $TENANT_ORG_ID"
echo "  User: $USER"
echo "  Service: $SERVICE_NAME"
echo "  Namespace: $TENANT_NAME"
echo ""
echo "To access the cluster:"
echo "  kubectl -n $TENANT_NAME get pods"
echo "  kubectl -n $TENANT_NAME port-forward svc/elasticsearch 9200:9200"
