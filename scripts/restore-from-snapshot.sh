#!/bin/bash
# 从快照恢复数据，支持按租户隔离的存储

set -euo pipefail

# 环境变量
ES_HOST=${ES_HOST:-"localhost:9200"}
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"minio:9000"}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-"es_minio"}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-"es_minio_P@ssw0rd_2025_11_28"}
NAMESPACE=${NAMESPACE:-""}
TENANT_ORG_ID=${TENANT_ORG_ID:-""}
SNAPSHOT_NAME=${SNAPSHOT_NAME:-""}

# 检查必需参数
if [[ -z "$NAMESPACE" || -z "$TENANT_ORG_ID" ]]; then
    echo "Error: NAMESPACE and TENANT_ORG_ID must be provided"
    exit 1
fi

# 创建按租户隔离的存储桶路径
BUCKET_NAME="es-backups"
TENANT_PATH="$TENANT_ORG_ID/$NAMESPACE"
REPOSITORY_NAME="${TENANT_ORG_ID}_${NAMESPACE}"

echo "Restoring Elasticsearch snapshot for namespace: $NAMESPACE"
echo "Tenant Org ID: $TENANT_ORG_ID"
echo "Storage path: $BUCKET_NAME/$TENANT_PATH"

# 1. 注册快照仓库（按租户隔离）
echo "Registering snapshot repository..."
curl -X PUT "$ES_HOST/_snapshot/$REPOSITORY_NAME" \
  -H 'Content-Type: application/json' \
  -d "{
    \"type\": \"s3\",
    \"settings\": {
      \"bucket\": \"$BUCKET_NAME\",
      \"client\": \"default\",
      \"base_path\": \"$TENANT_PATH\",
      \"endpoint\": \"$MINIO_ENDPOINT\",
      \"protocol\": \"http\",
      \"access_key\": \"$MINIO_ACCESS_KEY\",
      \"secret_key\": \"$MINIO_SECRET_KEY\"
    }
  }"

# 2. 如果提供了快照名称，则恢复指定快照，否则列出可用快照
if [[ -n "$SNAPSHOT_NAME" ]]; then
    echo "Restoring snapshot: $SNAPSHOT_NAME"
    curl -X POST "$ES_HOST/_snapshot/$REPOSITORY_NAME/$SNAPSHOT_NAME/_restore" \
      -H 'Content-Type: application/json' \
      -d '{
        "indices": "*",
        "ignore_unavailable": true,
        "include_global_state": false
      }'
    
    echo "Restore initiated for snapshot: $SNAPSHOT_NAME"
else
    echo "Listing available snapshots..."
    curl -X GET "$ES_HOST/_snapshot/$REPOSITORY_NAME/_all"
fi

echo "Restore process completed for namespace: $NAMESPACE"