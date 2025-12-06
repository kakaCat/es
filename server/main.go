package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"es-serverless-manager/internal/handler"
	"es-serverless-manager/internal/model"
	"es-serverless-manager/internal/service"
)

func main() {
	// Database Configuration
	// 数据库配置：从环境变量读取，提供默认值
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5433" // Default to mapped port for local dev
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "es_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "es_password_2025"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "es_metadata"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// 如果连接失败，记录严重错误并退出，因为 MetadataService 强依赖数据库
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate
	// 自动迁移数据库表结构
	err = db.AutoMigrate(
		&model.TenantContainer{},
		&model.IndexMetadata{},
		&model.TenantQuota{},
		&model.DeploymentStatus{},
		&model.Metrics{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// Initialize Core Services
	// 初始化核心服务：元数据服务
	metadataService := service.NewMetadataService(db)

	// Terraform Manager (assuming templates are in ./terraform/tenants or similar)
	// Terraform 管理器初始化，确保目录存在
	terraformDir := "./terraform/tenants"
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		log.Printf("Warning: Failed to create terraform directory: %v", err)
	}
	terraformManager := service.NewTerraformManager(terraformDir)

	// Background Services
	// 初始化后台服务：监控服务和自动扩缩容服务
	monitoringService := service.NewMonitoringService(metadataService)
	autoscalerService := service.NewAutoscalerService(metadataService)

	// ES Service
	// ES 服务配置
	esURL := os.Getenv("ES_URL")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}
	esService := service.NewESService(esURL)

	// Start Background Services
	// 启动后台服务
	log.Println("Starting monitoring service...")
	monitoringService.Start()

	log.Println("Starting autoscaler service...")
	autoscalerService.Start()

	// Ensure clean shutdown of background services
	// 注册延迟关闭函数，确保服务优雅停止
	defer func() {
		log.Println("Stopping monitoring service...")
		monitoringService.Stop()
		log.Println("Stopping autoscaler service...")
		autoscalerService.Stop()
	}()

	// Initialize Handlers
	// 初始化 HTTP 处理函数
	clusterHandler := handler.NewClusterHandler(metadataService, terraformManager)
	vectorHandler := handler.NewVectorHandler(esService)

	// Setup Router
	// 设置 Gin 路由
	r := gin.Default()

	// CORS Middleware
	// 配置跨域中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health Check
	// 健康检查接口
	r.GET("/health", handler.HandleHealth)

	// Cluster Routes
	// 集群管理相关路由
	clusters := r.Group("/clusters")
	{
		clusters.POST("", clusterHandler.CreateCluster)      // 创建集群
		clusters.GET("", clusterHandler.ListClusters)        // 获取集群列表
		clusters.DELETE("", clusterHandler.DeleteCluster)    // 删除集群
		clusters.POST("/scale", clusterHandler.ScaleCluster) // 扩缩容集群
	}

	// Vector Routes
	// 向量索引管理相关路由
	vectors := r.Group("/vectors")
	{
		vectors.POST("", vectorHandler.CreateVectorIndex)   // 创建向量索引
		vectors.GET("", vectorHandler.ListVectorIndexes)    // 获取索引列表
		vectors.DELETE("", vectorHandler.DeleteVectorIndex) // 删除索引

		// Operations on specific index
		// 特定索引的操作
		vectors.POST("/:index_name/doc", vectorHandler.IndexDocument)  // 插入文档
		vectors.POST("/:index_name/search", vectorHandler.Search)      // 搜索
		vectors.GET("/:index_name/stats", vectorHandler.GetIndexStats) // 获取统计信息
	}

	// Start Server
	// 启动 HTTP 服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
