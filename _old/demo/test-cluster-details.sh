#!/bin/bash

# Test script for cluster details API endpoint
# This script demonstrates how to query cluster details by namespace

set -e

echo "Testing Cluster Details API Endpoint"
echo "===================================="

# Configuration
MANAGER_URL="${MANAGER_URL:-http://localhost:8080}"
TEST_NAMESPACE="test-cluster-details"

echo "1. Creating a test cluster..."
curl -X POST "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "service_name": "test-service",
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

echo -e "\n\nWaiting for cluster to be ready..."
sleep 10

echo -e "\n2. Getting cluster details..."
curl -X GET "$MANAGER_URL/clusters/$TEST_NAMESPACE"

echo -e "\n\n3. Listing all clusters..."
curl -X GET "$MANAGER_URL/clusters"

echo -e "\n\n4. Cleaning up test cluster..."
curl -X DELETE "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d '{"namespace": "'"$TEST_NAMESPACE"'"}'

echo -e "\n\nTest completed!"