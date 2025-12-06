package handler

import (
	"net/http"
	"time"

	"es-serverless-manager/internal/model"
	"es-serverless-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type VectorHandler struct {
	esService *service.ESService
}

func NewVectorHandler(esService *service.ESService) *VectorHandler {
	return &VectorHandler{
		esService: esService,
	}
}

// CreateVectorIndex creates a new vector index
// CreateVectorIndex 创建新的向量索引
// @Summary Create a new vector index
// @Description Create a new vector index in Elasticsearch
// @Tags vectors
// @Accept json
// @Produce json
// @Param index body model.VectorIndexRequest true "Vector Index configuration"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /vectors [post]
func (h *VectorHandler) CreateVectorIndex(c *gin.Context) {
	var req model.VectorIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IndexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index_name is required"})
		return
	}

	// Construct mapping for vector index
	// 构建向量索引的映射
	// This is a simplified example. In a real scenario, you'd construct the mapping based on dimension, metric, etc.
	// 这是一个简化的示例。在实际场景中，您需要根据维度、度量等构建映射。
	mapping := model.VectorIndexMapping{
		Properties: map[string]interface{}{
			"vector": map[string]interface{}{
				"type":       "dense_vector",
				"dims":       req.Dimension,
				"index":      true,
				"similarity": req.Metric, // e.g., "l2_norm", "cosine", "dot_product"
			},
			// Add other fields from request if needed
			// 如果需要，添加其他字段
		},
	}

	// Add user defined fields
	// 添加用户定义的字段
	for k, v := range req.FieldMapping {
		mapping.Properties[k] = map[string]interface{}{
			"type": v,
		}
	}

	err := h.esService.CreateVectorIndex(req.IndexName, mapping)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"message": "Vector index created successfully",
		"index":   req.IndexName,
		"status":  "created",
	}
	c.JSON(http.StatusOK, response)
}

// IndexDocument indexes a document
// IndexDocument 索引文档（插入/更新）
func (h *VectorHandler) IndexDocument(c *gin.Context) {
	indexName := c.Param("index_name")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index_name is required"})
		return
	}

	var doc map[string]interface{}
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Optional: allow specifying doc ID via query param or field
	// 可选：通过查询参数或字段指定文档 ID
	docID := c.Query("id")

	err := h.esService.IndexDocument(indexName, docID, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document indexed successfully",
		"index":   indexName,
		"id":      docID,
	})
}

// Search performs a vector search
// Search 执行向量搜索
func (h *VectorHandler) Search(c *gin.Context) {
	indexName := c.Param("index_name")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index_name is required"})
		return
	}

	var query map[string]interface{}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.esService.Search(indexName, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetIndexStats gets index statistics
// GetIndexStats 获取索引统计信息
func (h *VectorHandler) GetIndexStats(c *gin.Context) {
	indexName := c.Param("index_name")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index_name is required"})
		return
	}

	stats, err := h.esService.GetIndexStats(indexName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ListVectorIndexes lists all vector indexes
// ListVectorIndexes 列出所有向量索引
// @Summary List all vector indexes
// @Description List all vector indexes in Elasticsearch
// @Tags vectors
// @Produce json
// @Success 200 {array} model.VectorIndexStatus
// @Failure 500 {string} string "Internal Server Error"
// @Router /vectors [get]
func (h *VectorHandler) ListVectorIndexes(c *gin.Context) {
	// In a real implementation, you might want to query ES metadata or store index metadata in your own DB
	// For now, we'll just list all indexes from ES
	// 在实际实现中，您可能需要查询 ES 元数据或将索引元数据存储在自己的数据库中
	// 目前，我们只是列出 ES 中的所有索引
	indexNames, err := h.esService.ListIndexes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	indexes := make([]model.VectorIndexStatus, len(indexNames))
	for i, name := range indexNames {
		indexes[i] = model.VectorIndexStatus{
			IndexName: name,
			Status:    "active",   // Assuming active if it exists
			CreatedAt: time.Now(), // Placeholder as ES _cat API doesn't give creation time easily without more queries
		}
	}

	c.JSON(http.StatusOK, indexes)
}

// DeleteVectorIndex deletes a vector index
// DeleteVectorIndex 删除向量索引
// @Summary Delete a vector index
// @Description Delete a vector index in Elasticsearch
// @Tags vectors
// @Accept json
// @Produce json
// @Param index body model.VectorIndexRequest true "Vector Index deletion info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /vectors [delete]
func (h *VectorHandler) DeleteVectorIndex(c *gin.Context) {
	var req model.VectorIndexRequest // Reusing request struct for delete parameters
	if err := c.ShouldBindJSON(&req); err != nil {
		// Try getting from query param if JSON body fails or is empty
		// 如果 JSON 正文失败或为空，尝试从查询参数获取
		indexName := c.Query("index_name")
		if indexName != "" {
			req.IndexName = indexName
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if req.IndexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index_name is required"})
		return
	}

	err := h.esService.DeleteIndex(req.IndexName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"message": "Vector index deleted successfully",
		"index":   req.IndexName,
		"status":  "deleted",
	}
	c.JSON(http.StatusOK, response)
}
