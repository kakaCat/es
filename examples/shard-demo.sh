#!/usr/bin/env bash
set -euo pipefail

# Shard Management Demo Script

NAMESPACE="es-shard-demo"
MANAGER_URL="http://localhost:8080"

echo "=== ES Serverless Shard Management Demo ==="
echo ""

# Step 1: Create a cluster
echo "Step 1: Creating a Serverless Elasticsearch cluster..."
cat <<EOF > /tmp/create-cluster.json
{
  "namespace": "$NAMESPACE",
  "replicas": 3,
  "cpu_request": "500m",
  "cpu_limit": "2",
  "mem_request": "1Gi",
  "mem_limit": "2Gi"
}
EOF

curl -X POST "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d @/tmp/create-cluster.json

echo ""
echo "Waiting for cluster to be ready..."
sleep 60

# Step 2: Port forward to access services
echo ""
echo "Step 2: Setting up port forwarding..."
kubectl -n "$NAMESPACE" port-forward svc/elasticsearch 9200:9200 &
kubectl -n "$NAMESPACE" port-forward svc/es-serverless-manager 8080:8080 &
ES_PID=$!
MANAGER_PID=$!

# Cleanup function
cleanup() {
    kill $ES_PID $MANAGER_PID 2>/dev/null || true
    curl -X DELETE "$MANAGER_URL/clusters" \
      -H "Content-Type: application/json" \
      -d "{\"namespace\":\"$NAMESPACE\"}"
}
trap cleanup EXIT

sleep 5

# Step 3: Create multiple indices to generate shards
echo ""
echo "Step 3: Creating indices to generate shards..."
for i in {1..5}; do
    curl -X PUT "http://localhost:9200/index_$i" \
      -H "Content-Type: application/json" \
      -d '{
        "settings": {
          "number_of_shards": 3,
          "number_of_replicas": 1
        }
      }'
    echo ""
done

# Step 4: Insert sample data
echo ""
echo "Step 4: Inserting sample data..."
for i in {1..100}; do
    curl -X POST "http://localhost:9200/index_$(( (i % 5) + 1 ))/_doc" \
      -H "Content-Type: application/json" \
      -d "{
        \"title\": \"Document $i\",
        \"content\": \"This is sample content for document $i\"
      }" >/dev/null 2>&1
done

echo "Sample data inserted successfully!"

# Step 5: Check shard distribution
echo ""
echo "Step 5: Checking shard distribution..."
curl -X GET "$MANAGER_URL/shards"

echo ""
echo "Shard distribution in Elasticsearch:"
curl -s "http://localhost:9200/_cat/shards?v" | column -t

# Step 6: Trigger shard rebalancing
echo ""
echo "Step 6: Triggering shard rebalancing..."
curl -X POST "$MANAGER_URL/shards" \
  -H "Content-Type: application/json" \
  -d '{"action": "rebalance"}'

echo ""
echo "Shard rebalancing triggered!"

# Step 7: Check shard distribution after rebalancing
echo ""
echo "Step 7: Checking shard distribution after rebalancing..."
sleep 30  # Wait for rebalancing to complete

echo "Shard distribution after rebalancing:"
curl -s "http://localhost:9200/_cat/shards?v" | column -t

# Step 8: Optimize shard allocation
echo ""
echo "Step 8: Optimizing shard allocation..."
curl -X POST "$MANAGER_URL/shards" \
  -H "Content-Type: application/json" \
  -d '{"action": "optimize"}'

echo ""
echo "Shard allocation optimization triggered!"

echo ""
echo "=== Shard Management Demo Completed Successfully ==="
echo ""
echo "To clean up, run: curl -X DELETE $MANAGER_URL/clusters -H 'Content-Type: application/json' -d '{\"namespace\":\"$NAMESPACE\"}'"