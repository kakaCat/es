#!/bin/bash
# 创建 Elasticsearch 快照，按租户隔离存储

set -euo pipefail

# 环境变量
ES_HOST=${ES_HOST:-"localhost:9200"}
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"minio:9000"}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-"es_minio"}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-"es_minio_P@ssw0rd_2025_11_28"}
NAMESPACE=${NAMESPACE:-""}
TENANT_ORG_ID=${TENANT_ORG_ID:-""}

# 如果没有提供命名空间和租户组织ID，则退出
if [[ -z "$NAMESPACE" || -z "$TENANT_ORG_ID" ]]; then
    echo "Error: NAMESPACE and TENANT_ORG_ID must be provided"
    exit 1
fi

# 创建按租户隔离的存储桶路径
BUCKET_NAME="es-backups"
TENANT_PATH="$TENANT_ORG_ID/$NAMESPACE"

echo "Creating Elasticsearch snapshot for namespace: $NAMESPACE"
echo "Tenant Org ID: $TENANT_ORG_ID"
echo "Storage path: $BUCKET_NAME/$TENANT_PATH"

# 1. 注册快照仓库（按租户隔离）
curl -X PUT "$ES_HOST/_snapshot/${TENANT_ORG_ID}_${NAMESPACE}" \
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

# 2. 创建快照
SNAPSHOT_NAME="snapshot_$(date +%Y%m%d_%H%M%S)_${NAMESPACE}"
echo "Creating snapshot: $SNAPSHOT_NAME"
curl -X PUT "$ES_HOST/_snapshot/${TENANT_ORG_ID}_${NAMESPACE}/$SNAPSHOT_NAME?wait_for_completion=true" \
  -H 'Content-Type: application/json' \
  -d '{
    "indices": "*",
    "ignore_unavailable": true,
    "include_global_state": false
  }'

# 3. 验证快照创建成功
echo "Verifying snapshot creation..."
curl -X GET "$ES_HOST/_snapshot/${TENANT_ORG_ID}_${NAMESPACE}/$SNAPSHOT_NAME"

echo "Snapshot created successfully for namespace: $NAMESPACE"
echo "Tenant Org ID: $TENANT_ORG_ID"
echo "Snapshot name: $SNAPSHOT_NAME"