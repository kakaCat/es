package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// ESClient is a simple Elasticsearch client
type ESClient struct {
    baseURL    string
    httpClient *http.Client
}

// NewESClient creates a new Elasticsearch client
func NewESClient(baseURL string) *ESClient {
    return &ESClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// CreateVectorIndex creates a new vector index in Elasticsearch
func (c *ESClient) CreateVectorIndex(indexName string, mapping VectorIndexMapping) error {
    url := fmt.Sprintf("%s/%s", c.baseURL, indexName)
    
    // Create the index mapping
    indexMapping := map[string]interface{}{
        "mappings": map[string]interface{}{
            "properties": mapping.Properties,
        },
    }
    
    body, err := json.Marshal(indexMapping)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("ES request failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    return nil
}

// DeleteIndex deletes an index from Elasticsearch
func (c *ESClient) DeleteIndex(indexName string) error {
    url := fmt.Sprintf("%s/%s", c.baseURL, indexName)
    
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("ES request failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    return nil
}

// IndexDocument indexes a document in Elasticsearch
func (c *ESClient) IndexDocument(indexName, docID string, document map[string]interface{}) error {
    var url string
    if docID != "" {
        url = fmt.Sprintf("%s/%s/_doc/%s", c.baseURL, indexName, docID)
    } else {
        url = fmt.Sprintf("%s/%s/_doc", c.baseURL, indexName)
    }
    
    body, err := json.Marshal(document)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("ES request failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    return nil
}

// Search performs a search query in Elasticsearch
func (c *ESClient) Search(indexName string, query map[string]interface{}) (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/%s/_search", c.baseURL, indexName)
    
    body, err := json.Marshal(query)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ES request failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result, nil
}

// GetIndexStats gets statistics for an index
func (c *ESClient) GetIndexStats(indexName string) (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/%s/_stats", c.baseURL, indexName)
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ES request failed with status %d: %s", resp.StatusCode, string(body))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result, nil
}

// VectorIndexMapping represents the mapping for a vector index
type VectorIndexMapping struct {
    Properties map[string]interface{} `json:"properties"`
}

// ShardInfo represents shard information
type ShardInfo struct {
    Index    string `json:"index"`
    Shard    string `json:"shard"`
    Prirep   string `json:"prirep"`   // p=primary, r=replica
    State    string `json:"state"`     // STARTED, RELOCATING, INITIALIZING, UNASSIGNED
    Docs     string `json:"docs"`
    Store    string `json:"store"`
    IP       string `json:"ip"`
    Node     string `json:"node"`
}

// ShardRecovery represents shard recovery information
type ShardRecovery struct {
    Index          string                 `json:"index"`
    Shard          int                    `json:"shard"`
    Type           string                 `json:"type"`          // primary or replica
    Stage          string                 `json:"stage"`         // init, index, verify_index, translog, finalize, done
    SourceNode     string                 `json:"source_node"`
    TargetNode     string                 `json:"target_node"`
    BytesRecovered int64                  `json:"bytes_recovered"`
    BytesTotal     int64                  `json:"bytes_total"`
    Percent        string                 `json:"percent"`
    StartTime      string                 `json:"start_time"`
    Files          map[string]interface{} `json:"files"`
}

// UpdateClusterSettings updates Elasticsearch cluster settings
func (c *ESClient) UpdateClusterSettings(settings map[string]interface{}) error {
    url := fmt.Sprintf("%s/_cluster/settings", c.baseURL)
    
    jsonData, err := json.Marshal(settings)
    if err != nil {
        return fmt.Errorf("failed to marshal settings: %v", err)
    }
    
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %v", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to update cluster settings: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("ES API error: %s, body: %s", resp.Status, string(body))
    }
    
    return nil
}

// GetRecoveryStatus gets shard recovery status
func (c *ESClient) GetRecoveryStatus() (map[string][]ShardRecovery, error) {
    url := fmt.Sprintf("%s/_recovery?active_only=true", c.baseURL)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to get recovery status: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ES API error: %s, body: %s", resp.Status, string(body))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }
    
    // Parse recovery information
    recoveries := make(map[string][]ShardRecovery)
    for index, data := range result {
        if shardData, ok := data.(map[string]interface{}); ok {
            if shardsArray, ok := shardData["shards"].([]interface{}); ok {
                for _, shardInfo := range shardsArray {
                    if shard, ok := shardInfo.(map[string]interface{}); ok {
                        recovery := parseShardRecovery(shard)
                        recovery.Index = index
                        recoveries[index] = append(recoveries[index], recovery)
                    }
                }
            }
        }
    }
    
    return recoveries, nil
}

// GetShardAllocation gets shard allocation information
func (c *ESClient) GetShardAllocation() ([]ShardInfo, error) {
    url := fmt.Sprintf("%s/_cat/shards?format=json", c.baseURL)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to get shard allocation: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ES API error: %s, body: %s", resp.Status, string(body))
    }
    
    var shards []ShardInfo
    if err := json.NewDecoder(resp.Body).Decode(&shards); err != nil {
        return nil, fmt.Errorf("failed to decode shards: %v", err)
    }
    
    return shards, nil
}

// GetClusterStats gets cluster statistics
func (c *ESClient) GetClusterStats() (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/_cluster/stats", c.baseURL)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to get cluster stats: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ES API error: %s, body: %s", resp.Status, string(body))
    }
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }
    
    return result, nil
}

// parseShardRecovery parses shard recovery information from ES response
func parseShardRecovery(shard map[string]interface{}) ShardRecovery {
    recovery := ShardRecovery{}
    
    if id, ok := shard["id"].(float64); ok {
        recovery.Shard = int(id)
    }
    if typ, ok := shard["type"].(string); ok {
        recovery.Type = typ
    }
    if stage, ok := shard["stage"].(string); ok {
        recovery.Stage = stage
    }
    if source, ok := shard["source"].(map[string]interface{}); ok {
        if name, ok := source["name"].(string); ok {
            recovery.SourceNode = name
        }
    }
    if target, ok := shard["target"].(map[string]interface{}); ok {
        if name, ok := target["name"].(string); ok {
            recovery.TargetNode = name
        }
    }
    if index, ok := shard["index"].(map[string]interface{}); ok {
        if size, ok := index["size"].(map[string]interface{}); ok {
            if recovered, ok := size["recovered_in_bytes"].(float64); ok {
                recovery.BytesRecovered = int64(recovered)
            }
            if total, ok := size["total_in_bytes"].(float64); ok {
                recovery.BytesTotal = int64(total)
            }
            if percent, ok := size["percent"].(string); ok {
                recovery.Percent = percent
            }
        }
    }
    if startTime, ok := shard["start_time_in_millis"].(float64); ok {
        recovery.StartTime = time.Unix(int64(startTime)/1000, 0).Format(time.RFC3339)
    }
    
    return recovery
}