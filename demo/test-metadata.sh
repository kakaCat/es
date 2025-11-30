#!/bin/bash

# Test script for metadata management functionality

echo "Testing Index Metadata Management..."

# Test creating index metadata
echo "Creating index metadata..."
curl -X POST http://localhost:8080/metadata/indexes \
  -H "Content-Type: application/json" \
  -d '{
    "index_name": "test_vector_index",
    "namespace": "test-namespace",
    "dimension": 128,
    "metric": "l2",
    "ivf_params": {
      "nlist": 100,
      "nprobe": 10
    },
    "created_by": "test-user",
    "status": "active",
    "document_count": 1000,
    "storage_size": "50MB"
  }'

echo -e "\n\nListing all index metadata..."
curl -X GET http://localhost:8080/metadata/indexes

echo -e "\n\nTesting Tenant Quota Management..."

# Test creating tenant quota
echo "Creating tenant quota..."
curl -X POST http://localhost:8080/metadata/tenants/test-tenant \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "max_indices": 10,
    "max_storage": "100GB",
    "current_indices": 2,
    "current_storage": "20GB"
  }'

echo -e "\n\nGetting tenant quota..."
curl -X GET http://localhost:8080/metadata/tenants/test-tenant

echo -e "\n\nTesting Deployment Status Management..."

# Test updating deployment status
echo "Updating deployment status..."
curl -X PUT http://localhost:8080/metadata/deployments/test-namespace \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "test-namespace",
    "user": "test-user",
    "service_name": "test-service",
    "status": "running",
    "cpu_usage": 45.5,
    "memory_usage": 60.2,
    "disk_usage": 30.0,
    "qps": 1200.5,
    "gpu_count": 1,
    "dimension": 128,
    "vector_count": 10000,
    "replicas": 3,
    "details": {
      "version": "1.0.0",
      "region": "us-west-1"
    }
  }'

echo -e "\n\nGetting deployment status..."
curl -X GET http://localhost:8080/metadata/deployments/test-namespace

echo -e "\n\nTesting Monitoring Metrics..."

# Test getting monitoring metrics
echo "Getting all monitoring metrics..."
curl -X GET http://localhost:8080/monitoring/metrics

echo -e "\n\nGetting monitoring metrics for a specific namespace..."
curl -X GET http://localhost:8080/monitoring/metrics/test-namespace

echo -e "\n\nAll tests completed."