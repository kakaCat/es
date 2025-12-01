# ES Serverless Manager API Documentation

## Overview

The ES Serverless Manager provides a RESTful API for managing Serverless Elasticsearch clusters and vector indexes.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, the API does not require authentication. In production environments, authentication should be implemented.

## Clusters

### Create Cluster

Creates a new Serverless Elasticsearch cluster.

**Endpoint:** `POST /clusters`

**Request Body:**
```json
{
  "namespace": "string",
  "replicas": "integer",
  "cpu_request": "string",
  "cpu_limit": "string",
  "mem_request": "string",
  "mem_limit": "string",
  "index_limit": "integer"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "my-cluster",
    "replicas": 1,
    "cpu_request": "500m",
    "cpu_limit": "2",
    "mem_request": "1Gi",
    "mem_limit": "2Gi",
    "index_limit": 10
  }'
```

### List Clusters

Lists all Serverless Elasticsearch clusters.

**Endpoint:** `GET /clusters`

**Example:**
```bash
curl -X GET http://localhost:8080/clusters
```

### Delete Cluster

Deletes a Serverless Elasticsearch cluster.

**Endpoint:** `DELETE /clusters`

**Request Body:**
```json
{
  "namespace": "string"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/clusters \
  -H "Content-Type: application/json" \
  -d '{"namespace": "my-cluster"}'
```

### Scale Cluster

Scales a Serverless Elasticsearch cluster.

**Endpoint:** `POST /clusters/scale`

**Request Body:**
```json
{
  "namespace": "string",
  "replicas": "integer"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/clusters/scale \
  -H "Content-Type: application/json" \
  -d '{"namespace": "my-cluster", "replicas": 3}'
```

## Vector Indexes

### Create Vector Index

Creates a new vector index.

**Endpoint:** `POST /vector-indexes`

**Request Body:**
```json
{
  "index_name": "string",
  "dimension": "integer",
  "metric": "string",
  "ivf_params": {
    "nlist": "integer",
    "nprobe": "integer"
  },
  "field_mapping": {
    "field_name": "field_type"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/vector-indexes \
  -H "Content-Type: application/json" \
  -d '{
    "index_name": "my_vector_index",
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
  }'
```

### List Vector Indexes

Lists all vector indexes.

**Endpoint:** `GET /vector-indexes`

**Example:**
```bash
curl -X GET http://localhost:8080/vector-indexes
```

### Delete Vector Index

Deletes a vector index.

**Endpoint:** `DELETE /vector-indexes`

**Request Body:**
```json
{
  "index_name": "string"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/vector-indexes \
  -H "Content-Type: application/json" \
  -d '{"index_name": "my_vector_index"}'
```

## Shards

### Get Shard Information

Gets information about shard distribution and cluster state.

**Endpoint:** `GET /shards`

**Example:**
```bash
curl -X GET http://localhost:8080/shards
```

### Manage Shards

Triggers shard management operations.

**Endpoint:** `POST /shards`

**Request Body:**
```json
{
  "action": "string"  // "rebalance" or "optimize"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/shards \
  -H "Content-Type: application/json" \
  -d '{"action": "rebalance"}'
```

## Health Check

### Health

Checks the health of the Manager service.

**Endpoint:** `GET /health`

**Example:**
```bash
curl -X GET http://localhost:8080/health
```

## Error Responses

The API uses standard HTTP status codes to indicate the success or failure of requests:

- `200 OK` - The request was successful
- `400 Bad Request` - The request was invalid
- `404 Not Found` - The requested resource was not found
- `500 Internal Server Error` - An error occurred on the server

Error responses include a JSON object with an error message:

```json
{
  "error": "Error message"
}
```