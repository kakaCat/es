package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"es-serverless-manager/internal/model"
)

// ESService handles Elasticsearch operations
// ESService 处理 Elasticsearch 操作
type ESService struct {
	baseURL    string
	httpClient *http.Client
}

// NewESService creates a new Elasticsearch service
// NewESService 创建一个新的 Elasticsearch 服务
func NewESService(baseURL string) *ESService {
	return &ESService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateVectorIndex creates a new vector index in Elasticsearch
// CreateVectorIndex 在 Elasticsearch 中创建一个新的向量索引
func (s *ESService) CreateVectorIndex(indexName string, mapping model.VectorIndexMapping) error {
	url := fmt.Sprintf("%s/%s", s.baseURL, indexName)

	// Create the index mapping
	// 创建索引映射
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

	resp, err := s.httpClient.Do(req)
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
// DeleteIndex 从 Elasticsearch 中删除索引
func (s *ESService) DeleteIndex(indexName string) error {
	url := fmt.Sprintf("%s/%s", s.baseURL, indexName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		// If index not found (404), consider it deleted
		// 如果索引未找到 (404)，则认为已删除
		if resp.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("ES request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IndexDocument indexes a document in Elasticsearch
// IndexDocument 在 Elasticsearch 中索引文档
func (s *ESService) IndexDocument(indexName, docID string, document map[string]interface{}) error {
	var url string
	if docID != "" {
		url = fmt.Sprintf("%s/%s/_doc/%s", s.baseURL, indexName, docID)
	} else {
		url = fmt.Sprintf("%s/%s/_doc", s.baseURL, indexName)
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

	resp, err := s.httpClient.Do(req)
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

// ListIndexes lists all indexes (simple implementation)
// ListIndexes 列出所有索引（简单实现）
func (s *ESService) ListIndexes() ([]string, error) {
	url := fmt.Sprintf("%s/_cat/indices?format=json", s.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to list indexes, status: %d", resp.StatusCode)
	}

	var indices []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&indices); err != nil {
		return nil, err
	}

	var names []string
	for _, idx := range indices {
		if name, ok := idx["index"].(string); ok {
			names = append(names, name)
		}
	}

	return names, nil
}

// Search performs a search query in Elasticsearch
// Search 执行 Elasticsearch 搜索查询
func (s *ESService) Search(indexName string, query map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/_search", s.baseURL, indexName)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
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
// GetIndexStats 获取索引统计信息
func (s *ESService) GetIndexStats(indexName string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/_stats", s.baseURL, indexName)

	resp, err := s.httpClient.Get(url)
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
