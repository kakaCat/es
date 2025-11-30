# Elasticsearch IVF Vector Plugin Documentation

## Overview

The Elasticsearch IVF Vector Plugin adds support for vector search capabilities to Elasticsearch using the Inverted File (IVF) algorithm. This plugin provides a custom field type for storing vectors and a query type for performing approximate nearest neighbor (ANN) searches.

## Features

- Custom `vector` field type for storing high-dimensional vectors
- ANN search using IVF algorithm
- Support for L2, cosine, and dot product distance metrics
- Configurable IVF parameters (nlist, nprobe)

## Installation

### Prerequisites

- Elasticsearch 8.x
- Java 17+

### Building the Plugin

1. Clone the repository:
   ```bash
   cd es-plugin
   ```

2. Build the plugin:
   ```bash
   ./gradlew assemble
   ```

3. The plugin ZIP file will be created in `build/distributions/`.

### Installing the Plugin

1. Install the plugin in Elasticsearch:
   ```bash
   bin/elasticsearch-plugin install file:///path/to/es-vector-ivf-plugin.zip
   ```

2. Restart Elasticsearch.

## Usage

### Creating an Index with a Vector Field

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

### Indexing Documents with Vector Data

```json
POST /my_index/_doc
{
  "text": "Sample document",
  "embedding": [0.1, 0.2, 0.3, ..., 0.128]
}
```

### Searching with ANN Query

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

## Configuration

### Vector Field Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `dimension` | Dimensionality of the vectors | None (required) |
| `metric` | Distance metric (l2, cosine, dot) | l2 |
| `nlist` | Number of clusters for IVF | 100 |
| `nprobe` | Number of clusters to search | 10 |

### ANN Query Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `field` | Field name containing vectors | None (required) |
| `vector` | Query vector | None (required) |
| `algorithm` | Search algorithm (ivf) | ivf |
| `nprobe` | Number of clusters to search | 10 |

## Performance Tuning

### IVF Parameters

- **nlist**: Number of clusters. Higher values provide better accuracy but slower indexing.
- **nprobe**: Number of clusters to search. Higher values provide better accuracy but slower search.

### Recommendations

- For 1M vectors, use nlist=1024
- For search, use nprobe=16-32 for a good balance of accuracy and speed
- Adjust based on your specific use case and performance requirements

## Limitations

- Currently supports only float vectors
- Maximum dimensionality: 4096
- Single-threaded indexing (planned improvement)

## Troubleshooting

### Common Issues

1. **Plugin not loading**: Ensure Elasticsearch version compatibility and Java version.
2. **Indexing errors**: Check vector dimensionality matches the field definition.
3. **Search errors**: Verify query vector dimensionality matches the index.

### Logs

Check Elasticsearch logs for plugin-related errors:
```bash
tail -f /var/log/elasticsearch/elasticsearch.log
```