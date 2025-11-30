#!/usr/bin/env bash
set -euo pipefail

# Demo script for ES Serverless system

NAMESPACE="es-demo"
MANAGER_URL="http://localhost:8080"

echo "=== ES Serverless System Demo ==="
echo ""

# Step 1: Create a cluster
echo "Step 1: Creating a Serverless Elasticsearch cluster..."
cat <<EOF > /tmp/create-cluster.json
{
  "namespace": "$NAMESPACE",
  "replicas": 1,
  "cpu_request": "500m",
  "cpu_limit": "2",
  "mem_request": "1Gi",
  "mem_limit": "2Gi",
  "index_limit": 10
}
EOF

curl -X POST "$MANAGER_URL/clusters" \
  -H "Content-Type: application/json" \
  -d @/tmp/create-cluster.json

echo ""
echo "Waiting for cluster to be ready..."
sleep 30

# Step 2: Port forward to access services
echo ""
echo "Step 2: Setting up port forwarding..."
kubectl -n "$NAMESPACE" port-forward svc/elasticsearch 9200:9200 &
kubectl -n "$NAMESPACE" port-forward svc/kibana 5601:5601 &
kubectl -n "$NAMESPACE" port-forward svc/es-serverless-manager 8080:8080 &

echo "Services are now accessible:"
echo "  Elasticsearch: http://localhost:9200"
echo "  Kibana: http://localhost:5601"
echo "  Manager API: http://localhost:8080"
sleep 5

# Step 3: Create a vector index
echo ""
echo "Step 3: Creating a vector index..."
cat <<EOF > /tmp/create-index.json
{
  "index_name": "demo_vector_index",
  "dimension": 128,
  "metric": "l2",
  "ivf_params": {
    "nlist": 100,
    "nprobe": 10
  },
  "field_mapping": {
    "title": "text",
    "embedding": "vector"
  }
}
EOF

curl -X POST "$MANAGER_URL/vector-indexes" \
  -H "Content-Type: application/json" \
  -d @/tmp/create-index.json

# Step 4: Insert sample data
echo ""
echo "Step 4: Inserting sample vector data..."
# Generate a random vector
VECTOR_DATA="["
for i in {1..128}; do
  if [ $i -eq 128 ]; then
    VECTOR_DATA+="$(echo "scale=2; $RANDOM/32767*2-1" | bc)"
  else
    VECTOR_DATA+="$(echo "scale=2; $RANDOM/32767*2-1" | bc),"
  fi
done
VECTOR_DATA+="]"

cat <<EOF > /tmp/sample-doc.json
{
  "title": "Sample Document",
  "embedding": $VECTOR_DATA
}
EOF

curl -X POST "http://localhost:9200/demo_vector_index/_doc" \
  -H "Content-Type: application/json" \
  -d @/tmp/sample-doc.json

echo ""
echo "Sample document inserted successfully!"

# Step 5: Perform a vector search
echo ""
echo "Step 5: Performing a vector search..."
# Generate a query vector
QUERY_VECTOR="["
for i in {1..128}; do
  if [ $i -eq 128 ]; then
    QUERY_VECTOR+="$(echo "scale=2; $RANDOM/32767*2-1" | bc)"
  else
    QUERY_VECTOR+="$(echo "scale=2; $RANDOM/32767*2-1" | bc),"
  fi
done
QUERY_VECTOR+="]"

cat <<EOF > /tmp/search-query.json
{
  "size": 10,
  "query": {
    "ann": {
      "field": "embedding",
      "vector": $QUERY_VECTOR,
      "algorithm": "ivf",
      "nprobe": 8
    }
  }
}
EOF

curl -X POST "http://localhost:9200/demo_vector_index/_search" \
  -H "Content-Type: application/json" \
  -d @/tmp/search-query.json | jq '.'

echo ""
echo "Demo completed successfully!"
echo ""
echo "To clean up, run: curl -X DELETE $MANAGER_URL/clusters -H 'Content-Type: application/json' -d '{\"namespace\":\"$NAMESPACE\"}'"