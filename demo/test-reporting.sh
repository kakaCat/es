#!/bin/bash

# Test script for reporting component
# This script demonstrates how to test the reporting service

set -e

echo "Testing Reporting Component"
echo "=========================="

# Configuration
MANAGER_URL="${MANAGER_URL:-http://localhost:8080}"

echo "1. Checking if reporting service is deployed..."
kubectl -n es-serverless get deploy reporting-service || echo "Reporting service not found"

echo -e "\n2. Checking reporting service pods..."
kubectl -n es-serverless get pods -l app=reporting-service || echo "No reporting service pods found"

echo -e "\n3. Checking reporting service logs..."
kubectl -n es-serverless logs -l app=reporting-service --tail=10 || echo "No logs available"

echo -e "\n4. Testing reporting service endpoint..."
# Port forward the reporting service
kubectl -n es-serverless port-forward svc/reporting-service 8081:8080 &
PORT_FORWARD_PID=$!

# Wait a moment for port forwarding to establish
sleep 3

# Test the endpoint
curl -X GET http://localhost:8081/health || echo "Reporting service endpoint not available"

# Kill the port forwarding process
kill $PORT_FORWARD_PID 2>/dev/null || true

echo -e "\n5. Creating a test index to trigger reporting..."
curl -X POST "$MANAGER_URL/vector-indexes" \
  -H "Content-Type: application/json" \
  -d '{
    "index_name": "test_reporting_index",
    "dimension": 128,
    "metric": "l2",
    "ivf_params": {
      "nlist": 100,
      "nprobe": 10
    }
  }'

echo -e "\n\n6. Waiting for reporting to occur..."
sleep 10

echo -e "\n7. Checking logs for reporting activity..."
kubectl -n es-serverless logs -l app=reporting-service --tail=20 || echo "No logs available"

echo -e "\n\nTest completed!"