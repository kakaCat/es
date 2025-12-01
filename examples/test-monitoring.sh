#!/usr/bin/env bash

# Test script for monitoring and autoscaling API

echo "Testing metrics collection..."

# Get cluster metrics
echo "Getting cluster metrics..."
curl -X GET http://localhost:8080/metrics

echo ""
echo "Test completed."