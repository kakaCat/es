# 灾难恢复手册

## 概述

本文档描述了ES Serverless平台的灾难恢复流程，包括数据备份、恢复和验证步骤。

## RTO/RPO 目标

- **RTO (Recovery Time Objective)**: 30 分钟
- **RPO (Recovery Point Objective)**: 24 小时

## 数据备份策略

### 1. Elasticsearch 快照备份

- **频率**: 每天凌晨 2 点自动执行
- **存储位置**: MinIO 对象存储 (S3 兼容)
- **隔离方式**: 按租户组织ID和命名空间进行路径隔离
- **保留策略**: 保留最近 7 天的快照

### 2. 元数据备份

- **频率**: 每天凌晨 2 点自动执行
- **存储位置**: MinIO 对象存储 (S3 兼容)
- **隔离方式**: 按租户组织ID进行路径隔离
- **加密**: 使用 AES-256 加密

## 完整恢复流程

### 1. 部署新的 Kubernetes 集群

```bash
# 部署基础组件
kubectl apply -k k8s/overlays/dev
```

### 2. 部署 Elasticsearch

等待 Elasticsearch StatefulSet 完全启动。

### 3. 恢复元数据数据库

```bash
# 下载备份文件
mc cp minio/es-backups/metadata/all/metadata_backup_all_*.sql.enc ./metadata_backup.sql.enc

# 解密备份文件
openssl enc -d -aes-256-cbc -salt \
    -in metadata_backup.sql.enc \
    -out metadata_backup.sql \
    -k $DB_PASSWORD

# 恢复数据库
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < metadata_backup.sql
```

### 4. 恢复 Elasticsearch 快照

```bash
# 执行恢复脚本
./scripts/restore-from-snapshot.sh
```

### 5. 验证数据完整性

```bash
# 验证租户数据
curl -X GET "localhost:9200/_cat/indices?v"

# 验证元数据
kubectl -n es-serverless exec -it postgres -- psql -U es_user -d es_metadata -c "SELECT COUNT(*) FROM tenant_containers;"
```

### 6. 切换流量

更新 DNS 记录或负载均衡器配置，将流量切换到恢复的环境。

## 演练计划

- **频率**: 每季度一次
- **记录**: 恢复时间和遇到的问题
- **改进**: 根据演练结果优化恢复流程

## 常见问题和解决方案

### 1. 快照仓库注册失败

**问题**: 无法连接到 MinIO 存储

**解决方案**:
```bash
# 检查 MinIO 服务状态
kubectl -n es-serverless get pods -l app=minio

# 检查网络连接
kubectl -n es-serverless exec -it elasticsearch-0 -- curl -v http://minio:9000
```

### 2. 数据库恢复失败

**问题**: 数据库连接或权限问题

**解决方案**:
```bash
# 检查数据库连接
kubectl -n es-serverless exec -it postgres -- pg_isready

# 检查用户权限
kubectl -n es-serverless exec -it postgres -- psql -U es_user -d es_metadata -c "SELECT * FROM tenant_containers LIMIT 1;"
```

## 联系信息

- **系统管理员**: [邮箱/电话]
- **数据库管理员**: [邮箱/电话]
- **存储管理员**: [邮箱/电话]

## 附录

### 1. 备份脚本位置

- Elasticsearch 快照: `/scripts/backup-es-snapshot.sh`
- 元数据备份: `/scripts/backup-metadata.sh`
- 恢复脚本: `/scripts/restore-from-snapshot.sh`

### 2. Kubernetes 资源

- 备份 CronJob: `kubectl -n es-serverless get cronjob es-backup`
- 备份历史: `kubectl -n es-serverless get jobs`