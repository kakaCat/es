package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// IndexStats represents index statistics
type IndexStats struct {
	IndexName        string    `json:"index_name"`
	DocumentCount    int       `json:"document_count"`
	StorageSize      string    `json:"storage_size"`
	Dimension        int       `json:"dimension"`
	Metric           string    `json:"metric"`
	IVFParams        IVFParams `json:"ivf_params"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdatedAt    time.Time `json:"last_updated_at"`
}

// QPSStats represents query per second statistics
type QPSStats struct {
	IndexName     string    `json:"index_name"`
	QPS           float64   `json:"qps"`
	AvgLatency    float64   `json:"avg_latency"`
	P95Latency    float64   `json:"p95_latency"`
	P99Latency    float64   `json:"p99_latency"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// ReportingService handles data reporting to external systems
type ReportingService struct {
	esClient     *ESClient
	reportingURL string
	ticker       *time.Ticker
}

// NewReportingService creates a new reporting service
func NewReportingService(esClient *ESClient, reportingURL string) *ReportingService {
	return &ReportingService{
		esClient:     esClient,
		reportingURL: reportingURL,
		ticker:       time.NewTicker(60 * time.Second), // Report every minute
	}
}

// Start begins the reporting loop
func (rs *ReportingService) Start() {
	go func() {
		for range rs.ticker.C {
			rs.reportIndexStats()
			rs.reportQPSStats()
		}
	}()
}

// Stop stops the reporting loop
func (rs *ReportingService) Stop() {
	rs.ticker.Stop()
}

// reportIndexStats collects and reports index statistics
func (rs *ReportingService) reportIndexStats() {
	// In a real implementation, you would query Elasticsearch for index stats
	// For now, we'll create mock data
	stats := []IndexStats{
		{
			IndexName:     "sample_vector_index",
			DocumentCount: 10000,
			StorageSize:   "50MB",
			Dimension:     128,
			Metric:        "l2",
			IVFParams: IVFParams{
				NList:  100,
				NProbe: 10,
			},
			CreatedAt:     time.Now().Add(-2 * time.Hour),
			LastUpdatedAt: time.Now(),
		},
	}

	// Send stats to reporting endpoint
	rs.sendReport("index_stats", stats)
}

// reportQPSStats collects and reports QPS statistics
func (rs *ReportingService) reportQPSStats() {
	// In a real implementation, you would query Elasticsearch for QPS stats
	// For now, we'll create mock data
	stats := []QPSStats{
		{
			IndexName:     "sample_vector_index",
			QPS:           1500.5,
			AvgLatency:    15.2,
			P95Latency:    25.8,
			P99Latency:    35.1,
			LastUpdatedAt: time.Now(),
		},
	}

	// Send stats to reporting endpoint
	rs.sendReport("qps_stats", stats)
}

// sendReport sends report data to the reporting endpoint
func (rs *ReportingService) sendReport(reportType string, data interface{}) {
	if rs.reportingURL == "" {
		// If no reporting URL is configured, just log the data
		log.Printf("Report [%s]: %+v", reportType, data)
		return
	}

	// Create report payload
	payload := map[string]interface{}{
		"type":      reportType,
		"timestamp": time.Now(),
		"data":      data,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling report data: %v", err)
		return
	}

	// Send HTTP POST request
	resp, err := http.Post(rs.reportingURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending report: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("Error sending report: HTTP %d", resp.StatusCode)
		return
	}

	log.Printf("Report [%s] sent successfully", reportType)
}

// ReportIndexCreation reports index creation event
func (rs *ReportingService) ReportIndexCreation(indexName string, dimension int, metric string, ivfParams IVFParams) {
	event := map[string]interface{}{
		"event":      "index_created",
		"index_name": indexName,
		"dimension":  dimension,
		"metric":     metric,
		"ivf_params": ivfParams,
		"timestamp":  time.Now(),
	}

	rs.sendReport("index_event", event)
}

// ReportIndexDeletion reports index deletion event
func (rs *ReportingService) ReportIndexDeletion(indexName string) {
	event := map[string]interface{}{
		"event":      "index_deleted",
		"index_name": indexName,
		"timestamp":  time.Now(),
	}

	rs.sendReport("index_event", event)
}

// ReportQueryPerformance reports query performance metrics
func (rs *ReportingService) ReportQueryPerformance(indexName string, qps float64, avgLatency, p95Latency, p99Latency float64) {
	stats := QPSStats{
		IndexName:     indexName,
		QPS:           qps,
		AvgLatency:    avgLatency,
		P95Latency:    p95Latency,
		P99Latency:    p99Latency,
		LastUpdatedAt: time.Now(),
	}

	rs.sendReport("query_performance", stats)
}