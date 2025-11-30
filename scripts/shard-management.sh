#!/usr/bin/env bash
set -euo pipefail

# Elasticsearch分片管理脚本

NAMESPACE=${NAMESPACE:-es-serverless}
ELASTICSEARCH_SVC=${ELASTICSEARCH_SVC:-elasticsearch}

# 获取Elasticsearch集群状态
get_cluster_state() {
    kubectl -n "$NAMESPACE" exec -it deploy/es-serverless-manager -- \
        curl -s "http://$ELASTICSEARCH_SVC:9200/_cluster/stats" | jq '.'
}

# 获取分片分布情况
get_shard_allocation() {
    kubectl -n "$NAMESPACE" exec -it deploy/es-serverless-manager -- \
        curl -s "http://$ELASTICSEARCH_SVC:9200/_cat/shards?v" | column -t
}

# 重新平衡分片
rebalance_shards() {
    echo "触发分片重新平衡..."
    kubectl -n "$NAMESPACE" exec -it deploy/es-serverless-manager -- \
        curl -X PUT "http://$ELASTICSEARCH_SVC:9200/_cluster/settings" -H 'Content-Type: application/json' -d'
{
  "transient": {
    "cluster.routing.rebalance.enable": "all",
    "cluster.routing.allocation.node_concurrent_recoveries": 2,
    "indices.recovery.max_bytes_per_sec": "50mb"
  }
}
'
}

# 优化分片分配
optimize_allocation() {
    echo "优化分片分配策略..."
    kubectl -n "$NAMESPACE" exec -it deploy/es-serverless-manager -- \
        curl -X PUT "http://$ELASTICSEARCH_SVC:9200/_cluster/settings" -H 'Content-Type: application/json' -d'
{
  "transient": {
    "cluster.routing.allocation.balance.shard": 0.45,
    "cluster.routing.allocation.balance.index": 0.55,
    "cluster.routing.allocation.balance.threshold": 1.0
  }
}
'
}

# 检查热点分片
check_hot_shards() {
    echo "检查热点分片..."
    kubectl -n "$NAMESPACE" exec -it deploy/es-serverless-manager -- \
        curl -s "http://$ELASTICSEARCH_SVC:9200/_nodes/stats/indices/search" | jq '.'
}

case "${1:-help}" in
    state)
        get_cluster_state
        ;;
    shards)
        get_shard_allocation
        ;;
    rebalance)
        rebalance_shards
        ;;
    optimize)
        optimize_allocation
        ;;
    hot)
        check_hot_shards
        ;;
    *)
        echo "Usage: $0 {state|shards|rebalance|optimize|hot}"
        echo "  state     - 获取集群状态"
        echo "  shards    - 获取分片分配情况"
        echo "  rebalance - 触发分片重新平衡"
        echo "  optimize  - 优化分片分配策略"
        echo "  hot       - 检查热点分片"
        ;;
esac