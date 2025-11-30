#!/bin/bash

# Test script for deployment reports functionality

echo "Testing Deployment Reports..."

# Test creating a cluster which should generate deployment reports
echo "Creating a cluster..."
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "user": "reportuser",
    "service_name": "report-service",
    "namespace": "report-namespace",
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

echo -e "\n\nListing all deployment reports..."
curl -X GET http://localhost:8080/deployment/reports

echo -e "\n\nGetting latest deployment report for specific user and service..."
curl -X GET http://localhost:8080/deployment/reports/reportuser/report-service

echo -e "\n\nTesting with another user..."
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "user": "anotherreportuser",
    "service_name": "another-report-service",
    "namespace": "another-report-namespace",
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

echo -e "\n\nListing all deployment reports after creating another cluster..."
curl -X GET http://localhost:8080/deployment/reports

echo -e "\n\nGetting latest deployment report for another user and service..."
curl -X GET http://localhost:8080/deployment/reports/anotherreportuser/another-report-service

echo -e "\n\nAll tests completed."