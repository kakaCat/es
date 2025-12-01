#!/bin/bash

# Test script for tenant container management functionality

echo "Testing Tenant Container Management..."

# Test creating a cluster which should sync data to tenant container management
echo "Creating a cluster..."
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "service_name": "test-service",
    "namespace": "test-namespace",
    "replicas": 1,
    "cpu_request": "500m",
    "cpu_limit": "2",
    "mem_request": "1Gi",
    "mem_limit": "2Gi",
    "disk_size": "10Gi",
    "gpu_count": 0,
    "dimension": 128,
    "vector_count": 10000
  }'

echo -e "\n\nListing all tenant containers..."
curl -X GET http://localhost:8080/tenant/containers

echo -e "\n\nGetting specific tenant container..."
curl -X GET http://localhost:8080/tenant/containers/testuser/test-service

echo -e "\n\nTesting with another user..."
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "user": "anotheruser",
    "service_name": "another-service",
    "namespace": "another-namespace",
    "replicas": 2,
    "cpu_request": "1",
    "cpu_limit": "4",
    "mem_request": "2Gi",
    "mem_limit": "4Gi",
    "disk_size": "20Gi",
    "gpu_count": 1,
    "dimension": 256,
    "vector_count": 50000
  }'

echo -e "\n\nListing all tenant containers after creating another cluster..."
curl -X GET http://localhost:8080/tenant/containers

echo -e "\n\nGetting another tenant container..."
curl -X GET http://localhost:8080/tenant/containers/anotheruser/another-service

echo -e "\n\nAll tests completed."