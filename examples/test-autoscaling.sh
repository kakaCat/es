#!/bin/bash

# Test script for autoscaling component
# This script demonstrates how to test the autoscaling service

set -e

echo "Testing Autoscaling Component"
echo "============================"

# Configuration
MANAGER_URL="${MANAGER_URL:-http://localhost:8080}"
TEST_NAMESPACE="test-autoscaling"

echo "1. Creating a test cluster..."
curl -X POST "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "service_name": "test-autoscaling-service",
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

echo -e "\n\n2. Waiting for cluster to be ready..."
sleep 15

echo -e "\n3. Setting up autoscaling policy for the user..."
curl -X POST "$MANAGER_URL/autoscale-policy" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "testuser",
    "enable_auto_scale_up": true,
    "enable_auto_scale_down": true,
    "scale_up_threshold": 70.0,
    "scale_down_threshold": 30.0,
    "max_replicas": 5,
    "min_replicas": 1
  }'

echo -e "\n\n4. Generating load to trigger autoscaling..."
# This would typically involve sending requests to the Elasticsearch cluster
# For demonstration purposes, we'll simulate this by updating metrics directly

echo -e "\n5. Checking current cluster status..."
curl -X GET "$MANAGER_URL/clusters/$TEST_NAMESPACE"

echo -e "\n\n6. Simulating high load metrics..."
# In a real scenario, these metrics would be collected automatically
# For testing, we can manually update metrics to trigger scaling

echo -e "\n7. Waiting for autoscaling to occur..."
sleep 60

echo -e "\n8. Checking cluster status after scaling..."
curl -X GET "$MANAGER_URL/clusters/$TEST_NAMESPACE"

echo -e "\n\n9. Checking autoscaling logs..."
kubectl -n es-serverless logs -l app=es-serverless-manager --tail=50 | grep -i autoscal || echo "No autoscaling logs found"

echo -e "\n\n10. Cleaning up test cluster..."
curl -X DELETE "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d '{"namespace": "'"$TEST_NAMESPACE"'"}'

echo -e "\n\nTest completed!"