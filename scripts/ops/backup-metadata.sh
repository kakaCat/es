#!/bin/bash
# 备份 PostgreSQL 元数据，按租户隔离存储

set -euo pipefail

# 环境变量
DB_HOST=${DB_HOST:-"postgres"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"es_user"}
DB_PASSWORD=${DB_PASSWORD:-"es_password_2025"}
DB_NAME=${DB_NAME:-"es_metadata"}
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"minio:9000"}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-"es_minio"}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-"es_minio_P@ssw0rd_2025_11_28"}
TENANT_ORG_ID=${TENANT_ORG_ID:-"all"}

# 设置环境变量用于pg_dump
export PGPASSWORD=$DB_PASSWORD

echo "Backing up metadata for tenant: $TENANT_ORG_ID"

# 创建备份文件名
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="metadata_backup_${TENANT_ORG_ID}_${TIMESTAMP}.sql"

# 1. 备份 PostgreSQL 元数据
if [[ "$TENANT_ORG_ID" == "all" ]]; then
    echo "Backing up all metadata..."
    pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > $BACKUP_FILE
else
    echo "Backing up metadata for tenant: $TENANT_ORG_ID"
    # 只备份特定租户的数据
    pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
        -t "tenant_containers" \
        -t "index_metadata" \
        -t "tenant_quota" \
        -t "deployment_status" \
        --where="tenant_org_id = '$TENANT_ORG_ID'" > $BACKUP_FILE
fi

# 2. 加密备份文件
ENCRYPTED_FILE="${BACKUP_FILE}.enc"
echo "Encrypting backup file..."
openssl enc -aes-256-cbc -salt \
    -in $BACKUP_FILE \
    -out $ENCRYPTED_FILE \
    -k $DB_PASSWORD

# 3. 上传到 MinIO（按租户隔离）
echo "Uploading to MinIO..."
# 使用mc命令上传文件
mc alias set minio-local http://$MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
mc mb -p minio-local/es-backups/metadata/$TENANT_ORG_ID
mc cp $ENCRYPTED_FILE minio-local/es-backups/metadata/$TENANT_ORG_ID/

# 4. 清理本地文件
rm $BACKUP_FILE $ENCRYPTED_FILE

echo "Metadata backup completed for tenant: $TENANT_ORG_ID"
echo "Backup file: $ENCRYPTED_FILE"
echo "Uploaded to: es-backups/metadata/$TENANT_ORG_ID/"