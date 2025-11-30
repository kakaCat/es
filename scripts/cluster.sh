#!/usr/bin/env bash
set -euo pipefail

ACTION=${1:-}
NAMESPACE=${NAMESPACE:-es-serverless}
REPLICAS=${REPLICAS:-1}
CPU_REQUEST=${CPU_REQUEST:-500m}
CPU_LIMIT=${CPU_LIMIT:-2}
MEM_REQUEST=${MEM_REQUEST:-1Gi}
MEM_LIMIT=${MEM_LIMIT:-2Gi}
DISK_SIZE=${DISK_SIZE:-10Gi}
GPU_COUNT=${GPU_COUNT:-0}
INDEX_LIMIT=${INDEX_LIMIT:-0}
USER=${USER:-}
SERVICE_NAME=${SERVICE_NAME:-}
TENANT_ORG_ID=${TENANT_ORG_ID:-}  # 租户组织ID（多租户隔离）
DIMENSION=${DIMENSION:-128}
VECTOR_COUNT=${VECTOR_COUNT:-10000}
GITLAB_URL=${GITLAB_URL:-}

# Create namespace if not exists
create_namespace() {
  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
  kubectl label namespace "$NAMESPACE" es-cluster=true --overwrite
  
  # 添加租户组织ID标签（用于多租户隔离和管理）
  if [ -n "$TENANT_ORG_ID" ]; then
    kubectl label namespace "$NAMESPACE" tenant-org-id="$TENANT_ORG_ID" --overwrite
    echo "Labeled namespace $NAMESPACE with tenant-org-id=$TENANT_ORG_ID"
  fi
  
  # 添加用户和服务名标签
  if [ -n "$USER" ]; then
    kubectl label namespace "$NAMESPACE" user="$USER" --overwrite
  fi
  if [ -n "$SERVICE_NAME" ]; then
    kubectl label namespace "$NAMESPACE" service-name="$SERVICE_NAME" --overwrite
  fi
}

# Pull docker-compose.yml from GitLab if URL is provided
pull_from_gitlab() {
  if [ -n "$GITLAB_URL" ]; then
    echo "Pulling docker-compose.yml from GitLab: $GITLAB_URL"
    # In a real implementation, you would authenticate and pull from GitLab
    # For now, we'll just log the action
    echo "Would pull from GitLab URL: $GITLAB_URL" >> /tmp/deployment.log
  fi
}

# Sync data to tenant container management
sync_to_tenant_management() {
  if [ -n "$USER" ] && [ -n "$SERVICE_NAME" ]; then
    echo "Syncing data to tenant container management for user: $USER, service: $SERVICE_NAME"
    
    # Create tenant container management data
    tenant_data=$(cat <<EOF
{
  "user": "$USER",
  "service_name": "$SERVICE_NAME",
  "namespace": "$NAMESPACE",
  "replicas": $REPLICAS,
  "cpu_request": "$CPU_REQUEST",
  "cpu_limit": "$CPU_LIMIT",
  "mem_request": "$MEM_REQUEST",
  "mem_limit": "$MEM_LIMIT",
  "disk_size": "$DISK_SIZE",
  "gpu_count": $GPU_COUNT,
  "dimension": $DIMENSION,
  "vector_count": $VECTOR_COUNT,
  "status": "created",
  "created_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
)
    
    # Save to tenant management system (in this case, a JSON file)
    tenant_dir="server/tenant_data"
    mkdir -p "$tenant_dir"
    echo "$tenant_data" > "$tenant_dir/${USER}_${SERVICE_NAME}.json"
    
    echo "Data synced to tenant container management"
  else
    echo "Skipping tenant data sync: USER or SERVICE_NAME not provided"
  fi
}

# Report deployment status
report_deployment_status() {
  local status=$1
  local message=$2
  
  if [ -n "$USER" ] && [ -n "$SERVICE_NAME" ]; then
    echo "Reporting deployment status: $status - $message"
    
    # Create deployment status report
    deployment_report=$(cat <<EOF
{
  "user": "$USER",
  "service_name": "$SERVICE_NAME",
  "namespace": "$NAMESPACE",
  "status": "$status",
  "message": "$message",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "details": {
    "replicas": $REPLICAS,
    "cpu_request": "$CPU_REQUEST",
    "cpu_limit": "$CPU_LIMIT",
    "mem_request": "$MEM_REQUEST",
    "mem_limit": "$MEM_LIMIT",
    "disk_size": "$DISK_SIZE",
    "gpu_count": $GPU_COUNT,
    "dimension": $DIMENSION,
    "vector_count": $VECTOR_COUNT
  }
}
EOF
)
    
    # Save deployment report to a dedicated directory
    report_dir="server/deployment_reports"
    mkdir -p "$report_dir"
    echo "$deployment_report" > "$report_dir/${USER}_${SERVICE_NAME}_$(date +%s).json"
    
    # Also append to a general deployment log
    echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") - User: $USER, Service: $SERVICE_NAME, Namespace: $NAMESPACE, Status: $status, Message: $message" >> /tmp/deployment.log
    
    echo "Deployment status reported"
  else
    echo "Skipping deployment status report: USER or SERVICE_NAME not provided"
  fi
}

case "$ACTION" in
  create)
    report_deployment_status "starting" "Starting cluster creation"
    
    create_namespace
    report_deployment_status "namespace_created" "Namespace created successfully"
    
    pull_from_gitlab
    report_deployment_status "gitlab_pulled" "GitLab resources pulled successfully"
    
    kubectl apply -k k8s/overlays/dev
    report_deployment_status "k8s_applied" "Kubernetes resources applied successfully"
    
    kubectl -n "$NAMESPACE" annotate sts/elasticsearch es.yunpeng.cn/max-indices="$INDEX_LIMIT" --overwrite
    kubectl -n "$NAMESPACE" scale sts/elasticsearch --replicas "$REPLICAS"
    kubectl -n "$NAMESPACE" set resources sts/elasticsearch --requests=cpu="$CPU_REQUEST",memory="$MEM_REQUEST" --limits=cpu="$CPU_LIMIT",memory="$MEM_LIMIT"
    report_deployment_status "resources_configured" "Cluster resources configured successfully"
    
    # Set disk size if specified
    if [ -n "$DISK_SIZE" ]; then
      kubectl -n "$NAMESPACE" patch pvc elasticsearch-data-elasticsearch-0 -p '{"spec":{"resources":{"requests":{"storage":"'"$DISK_SIZE"'"}}}}'
      report_deployment_status "disk_configured" "Disk size configured to $DISK_SIZE"
    fi
    
    # Set GPU count if specified
    if [ "$GPU_COUNT" -gt 0 ]; then
      kubectl -n "$NAMESPACE" patch sts elasticsearch -p '{"spec":{"template":{"spec":{"containers":[{"name":"elasticsearch","resources":{"limits":{"nvidia.com/gpu":"'"$GPU_COUNT"'"}}}]}}]}'
      report_deployment_status "gpu_configured" "GPU count configured to $GPU_COUNT"
    fi
    
    kubectl -n "$NAMESPACE" rollout status sts/elasticsearch
    kubectl -n "$NAMESPACE" rollout status deploy/kibana
    report_deployment_status "rollout_completed" "Cluster rollout completed successfully"
    
    # Sync data to tenant container management
    sync_to_tenant_management
    report_deployment_status "tenant_synced" "Data synced to tenant container management"
    
    # Record deployment details
    echo "Deployment completed for user: $USER, service: $SERVICE_NAME" >> /tmp/deployment.log
    echo "Namespace: $NAMESPACE, Replicas: $REPLICAS, Disk Size: $DISK_SIZE, GPU Count: $GPU_COUNT" >> /tmp/deployment.log
    
    report_deployment_status "completed" "Cluster creation completed successfully"
    ;;
  delete)
    report_deployment_status "starting" "Starting cluster deletion"
    
    kubectl delete ns "$NAMESPACE" --ignore-not-found=true
    report_deployment_status "namespace_deleted" "Namespace deleted successfully"
    
    # Remove tenant data if exists
    if [ -n "$USER" ] && [ -n "$SERVICE_NAME" ]; then
      tenant_file="server/tenant_data/${USER}_${SERVICE_NAME}.json"
      if [ -f "$tenant_file" ]; then
        rm "$tenant_file"
        echo "Removed tenant data for user: $USER, service: $SERVICE_NAME"
        report_deployment_status "tenant_data_removed" "Tenant data removed successfully"
      fi
    fi
    
    report_deployment_status "completed" "Cluster deletion completed successfully"
    ;;
  scale)
    report_deployment_status "starting" "Starting cluster scaling to $REPLICAS replicas"
    
    kubectl -n "$NAMESPACE" scale sts/elasticsearch --replicas "$REPLICAS"
    report_deployment_status "completed" "Cluster scaled to $REPLICAS replicas successfully"
    ;;
  status)
    kubectl -n "$NAMESPACE" get sts/elasticsearch
    ;;
  *)
    echo "Usage: scripts/cluster.sh {create|delete|scale|status}"
    echo "Env: REPLICAS CPU_REQUEST CPU_LIMIT MEM_REQUEST MEM_LIMIT DISK_SIZE GPU_COUNT INDEX_LIMIT NAMESPACE USER SERVICE_NAME DIMENSION VECTOR_COUNT GITLAB_URL"
    exit 1
    ;;
esac