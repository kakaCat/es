#!/usr/bin/env bash
set -euo pipefail

# Terraform-based deployment script for ES Serverless system

ACTION=${1:-help}
TERRAFORM_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../terraform" && pwd)"

show_help() {
    echo "Usage: scripts/deploy-terraform.sh [ACTION]"
    echo ""
    echo "Actions:"
    echo "  help        - Show this help message"
    echo "  init        - Initialize Terraform"
    echo "  plan        - Plan infrastructure changes"
    echo "  apply       - Apply infrastructure changes"
    echo "  destroy     - Destroy all infrastructure"
    echo "  output      - Show Terraform outputs"
    echo "  status      - Show deployment status"
    echo ""
    echo "Examples:"
    echo "  # Initial setup"
    echo "  ./scripts/deploy-terraform.sh init"
    echo "  ./scripts/deploy-terraform.sh plan"
    echo "  ./scripts/deploy-terraform.sh apply"
    echo ""
    echo "  # Check status"
    echo "  ./scripts/deploy-terraform.sh status"
    echo ""
    echo "  # Clean up"
    echo "  ./scripts/deploy-terraform.sh destroy"
}

init_terraform() {
    echo "Initializing Terraform..."
    cd "$TERRAFORM_DIR"
    terraform init
    echo "Terraform initialized successfully!"
}

plan_terraform() {
    echo "Planning Terraform changes..."
    cd "$TERRAFORM_DIR"
    terraform plan
}

apply_terraform() {
    echo "Applying Terraform configuration..."
    cd "$TERRAFORM_DIR"
    terraform apply -auto-approve

    echo ""
    echo "ES Serverless system deployed successfully!"
    echo ""
    echo "Access the services:"
    terraform output -json | jq -r '
        "  Elasticsearch: kubectl -n " + .namespace.value + " port-forward svc/elasticsearch 9200:9200",
        "  Manager API: kubectl -n " + .namespace.value + " port-forward svc/es-control-plane-manager 8080:8080",
        "  Grafana: kubectl -n " + .namespace.value + " port-forward svc/monitoring-grafana 3000:3000",
        "  Prometheus: kubectl -n " + .namespace.value + " port-forward svc/monitoring-prometheus 9090:9090"
    '
}

destroy_terraform() {
    echo "WARNING: This will destroy all infrastructure managed by Terraform!"
    read -p "Are you sure you want to continue? (yes/no): " -r
    echo
    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo "Destroying Terraform-managed infrastructure..."
        cd "$TERRAFORM_DIR"
        terraform destroy -auto-approve
        echo "Infrastructure destroyed successfully!"
    else
        echo "Destroy cancelled."
        exit 0
    fi
}

show_output() {
    echo "Terraform outputs:"
    cd "$TERRAFORM_DIR"
    terraform output
}

show_status() {
    cd "$TERRAFORM_DIR"

    NAMESPACE=$(terraform output -raw namespace 2>/dev/null || echo "es-serverless")

    echo "ES Serverless system status:"
    echo ""
    echo "Namespace: $NAMESPACE"
    echo ""
    echo "Helm releases:"
    helm list -n "$NAMESPACE" 2>/dev/null || echo "No Helm releases found"
    echo ""
    echo "Pods:"
    kubectl get pods -n "$NAMESPACE" 2>/dev/null || echo "No pods found"
    echo ""
    echo "Services:"
    kubectl get svc -n "$NAMESPACE" 2>/dev/null || echo "No services found"
    echo ""
    echo "PVCs:"
    kubectl get pvc -n "$NAMESPACE" 2>/dev/null || echo "No PVCs found"
}

case "$ACTION" in
    help)
        show_help
        ;;
    init)
        init_terraform
        ;;
    plan)
        plan_terraform
        ;;
    apply)
        apply_terraform
        ;;
    destroy)
        destroy_terraform
        ;;
    output)
        show_output
        ;;
    status)
        show_status
        ;;
    *)
        echo "Unknown action: $ACTION"
        show_help
        exit 1
        ;;
esac
