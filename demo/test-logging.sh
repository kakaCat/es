#!/bin/bash

# Test script for logging component
# This script demonstrates how to test the logging service

set -e

echo "Testing Logging Component"
echo "========================"

# Configuration
MANAGER_URL="${MANAGER_URL:-http://localhost:8080}"
TEST_NAMESPACE="test-logging"

echo "1. Checking if fluentd daemonset is deployed..."
kubectl -n es-serverless get daemonset fluentd || echo "Fluentd daemonset not found"

echo -e "\n2. Checking fluentd pods..."
kubectl -n es-serverless get pods -l app=fluentd || echo "No fluentd pods found"

echo -e "\n3. Checking fluentd pod logs..."
FLUENTD_POD=$(kubectl -n es-serverless get pods -l app=fluentd -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$FLUENTD_POD" ]; then
  echo "Checking logs for pod: $FLUENTD_POD"
  kubectl -n es-serverless logs "$FLUENTD_POD" --tail=20 || echo "Failed to get logs"
else
  echo "No fluentd pod found"
fi

echo -e "\n4. Creating a test cluster to generate logs..."
curl -X POST "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "service_name": "test-logging-service",
    "namespace": "'"$TEST_NAMESPACE"'",
    "replicas": 1,
    "cpu_request": "500m",
    "cpu_limit": "1",
    "mem_request": "1Gi",
    "mem_limit": "2Gi",
    "disk_size": "10Gi",
    "gpu_count": 0,
    "dimension": 128,
    "vector_count": 10000,
    "index_limit": 10,
    "gitlab_url": ""
  }'

echo -e "\n\n5. Waiting for cluster to be ready..."
sleep 15

echo -e "\n6. Checking Elasticsearch for logs..."
# Port forward Elasticsearch
kubectl -n es-serverless port-forward svc/elasticsearch 9201:9200 &
ES_PORT_FORWARD_PID=$!

# Wait a moment for port forwarding to establish
sleep 3

# Check if we can connect to Elasticsearch
curl -X GET http://localhost:9201/_cluster/health || echo "Cannot connect to Elasticsearch"

# Kill the port forwarding process
kill $ES_PORT_FORWARD_PID 2>/dev/null || true

echo -e "\n7. Creating a test index to generate more logs..."
curl -X POST "$MANAGER_URL/vector-indexes" \
  -H "Content-Type: application/json" \
  -d '{
    "index_name": "test_logging_index",
    "dimension": 128,
    "metric": "l2",
    "ivf_params": {
      "nlist": 100,
      "nprobe": 10
    }
  }'

echo -e "\n\n8. Waiting for log collection..."
sleep 10

echo -e "\n9. Checking fluentd logs for activity..."
if [ -n "$FLUENTD_POD" ]; then
  kubectl -n es-serverless logs "$FLUENTD_POD" --tail=30 | grep -E "(elasticsearch|kubernetes)" || echo "No relevant logs found"
fi

echo -e "\n10. Cleaning up test cluster..."
curl -X DELETE "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d '{"namespace": "'"$TEST_NAMESPACE"'"}'

echo -e "\n\nTest completed!"