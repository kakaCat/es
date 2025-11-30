#!/bin/bash

# Test script for container monitoring functionality

echo "Testing Container Monitoring..."

# Test getting all container metrics
echo "Getting all container metrics..."
curl -X GET http://localhost:8080/monitoring/container-metrics

echo -e "\n\nGetting container metrics for a specific namespace..."
curl -X GET http://localhost:8080/monitoring/container-metrics/test-namespace

echo -e "\n\nComparing with regular metrics..."
curl -X GET http://localhost:8080/monitoring/metrics

echo -e "\n\nGetting regular metrics for a specific namespace..."
curl -X GET http://localhost:8080/monitoring/metrics/test-namespace

echo -e "\n\nAll tests completed."