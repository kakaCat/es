#!/usr/bin/env bash

# Test script for deployment API

echo "Testing deployment creation..."

# Create a test deployment
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "service_name": "test-service",
    "namespace": "test-namespace",
    "replicas": 2,
    "cpu_request": "1",
    "cpu_limit": "2",
    "mem_request": "2Gi",
    "mem_limit": "4Gi",
    "disk_size": "20Gi",
    "gpu_count": 1,
    "dimension": 256,
    "vector_count": 50000,
    "index_limit": 5,
    "gitlab_url": "https://gitlab.example.com/docker-compose.yml"
  }'

echo ""
echo "Checking deployments..."
curl -X GET http://localhost:8080/deployments

echo ""
echo "Test completed."