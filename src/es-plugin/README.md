# Elasticsearch IVF Vector Plugin

This plugin adds support for IVF (Inverted File) vector search to Elasticsearch.

## Features

- Custom `vector` field type for storing high-dimensional vectors
- ANN (Approximate Nearest Neighbor) search using IVF algorithm
- Support for L2, cosine, and dot product distance metrics
- Configurable IVF parameters (nlist, nprobe)

## Installation

1. Build the plugin:
   ```
   ./gradlew assemble
   ```

2. Install the plugin in Elasticsearch:
   ```
   bin/elasticsearch-plugin install file:///path/to/es-vector-ivf-plugin.zip
   ```

## Usage

### Creating an index with a vector field

```json
PUT /my_index
{
  "mappings": {
    "properties": {
      "embedding": {
        "type": "vector",
        "dimension": 128,
        "metric": "l2",
        "nlist": 100,
        "nprobe": 10
      }
    }
  }
}
```

### Indexing documents with vector data

```json
POST /my_index/_doc
{
  "text": "Sample document",
  "embedding": [0.1, 0.2, 0.3, ..., 0.128]
}
```

### Searching with ANN query

```json
POST /my_index/_search
{
  "size": 10,
  "query": {
    "ann": {
      "field": "embedding",
      "vector": [0.15, 0.25, 0.35, ..., 0.135],
      "algorithm": "ivf",
      "nprobe": 8
    }
  }
}
```