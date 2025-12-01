#!/bin/bash

# Test script for IVF vector search functionality
# This script creates an index, trains it, adds vectors, and performs searches

set -e

ES_URL="${ES_URL:-http://localhost:9200}"
INDEX_NAME="test_vectors"
DIMENSION=128

echo "========================================="
echo "IVF Vector Search Test Script"
echo "========================================="
echo "Elasticsearch URL: $ES_URL"
echo "Index: $INDEX_NAME"
echo "Vector Dimension: $DIMENSION"
echo ""

# Step 1: Delete existing index if it exists
echo "[1/5] Cleaning up existing index..."
curl -X DELETE "$ES_URL/$INDEX_NAME" 2>/dev/null || echo "No existing index to delete"
echo ""

# Step 2: Create index with vector field
echo "[2/5] Creating index with vector field..."
curl -X PUT "$ES_URL/$INDEX_NAME" -H 'Content-Type: application/json' -d'{
  "mappings": {
    "properties": {
      "title": {
        "type": "text"
      },
      "embedding": {
        "type": "dense_vector",
        "dims": 128,
        "index": true,
        "similarity": "l2_norm"
      },
      "metadata": {
        "type": "object"
      }
    }
  },
  "settings": {
    "index": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "ivf": {
        "nlist": 10,
        "nprobe": 3,
        "metric": "l2"
      }
    }
  }
}'
echo -e "\n"

# Step 3: Generate and index test vectors
echo "[3/5] Indexing test vectors..."

# Function to generate a random float between -1 and 1
random_float() {
  echo "scale=6; ($RANDOM - 16384) / 16384" | bc
}

# Function to generate a random vector
generate_vector() {
  local vec="["
  for i in $(seq 1 $DIMENSION); do
    vec="$vec$(random_float)"
    if [ $i -lt $DIMENSION ]; then
      vec="$vec,"
    fi
  done
  vec="$vec]"
  echo "$vec"
}

# Index 100 test vectors
for i in $(seq 1 100); do
  VECTOR=$(generate_vector)

  curl -s -X POST "$ES_URL/$INDEX_NAME/_doc/$i" -H 'Content-Type: application/json' -d"{
    \"title\": \"Document $i\",
    \"embedding\": $VECTOR,
    \"metadata\": {
      \"category\": \"test\",
      \"id\": $i
    }
  }" > /dev/null

  if [ $((i % 20)) -eq 0 ]; then
    echo "  Indexed $i documents..."
  fi
done

echo "  Indexed 100 documents total"
echo ""

# Step 4: Refresh index to make documents searchable
echo "[4/5] Refreshing index..."
curl -X POST "$ES_URL/$INDEX_NAME/_refresh"
echo -e "\n"

# Step 5: Perform vector search
echo "[5/5] Performing vector search..."

QUERY_VECTOR=$(generate_vector)

echo "Query vector (first 5 dimensions): $(echo $QUERY_VECTOR | jq -r '.[0:5]')"
echo ""

# Standard kNN search (if supported)
echo "=== Standard kNN Search ==="
curl -X POST "$ES_URL/$INDEX_NAME/_search" -H 'Content-Type: application/json' -d"{
  \"knn\": {
    \"field\": \"embedding\",
    \"query_vector\": $QUERY_VECTOR,
    \"k\": 5,
    \"num_candidates\": 100
  },
  \"size\": 5
}" | jq -r '.hits.hits[] | {id: ._id, score: ._score, title: ._source.title}'
echo ""

# IVF ANN search (using our custom query)
echo "=== IVF ANN Search ==="
curl -X POST "$ES_URL/$INDEX_NAME/_search" -H 'Content-Type: application/json' -d"{
  \"query\": {
    \"ann\": {
      \"field\": \"embedding\",
      \"vector\": $QUERY_VECTOR,
      \"algorithm\": \"ivf\",
      \"nprobe\": 3,
      \"k\": 5
    }
  },
  \"size\": 5
}" | jq -r '.hits.hits[] | {id: ._id, score: ._score, title: ._source.title}'
echo ""

echo "========================================="
echo "Test completed!"
echo "========================================="
echo ""
echo "Summary:"
echo "  - Created index: $INDEX_NAME"
echo "  - Indexed 100 random vectors ($DIMENSION dimensions)"
echo "  - Performed kNN search (top-5 results)"
echo "  - Performed IVF ANN search (nprobe=3, k=5)"
echo ""
echo "To clean up:"
echo "  curl -X DELETE $ES_URL/$INDEX_NAME"
