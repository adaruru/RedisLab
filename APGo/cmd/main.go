package main

import (
	"log"
	"net/http"

	"github.com/AmandaChou/RedisLab/APGo/internal/config"
	"github.com/AmandaChou/RedisLab/APGo/internal/controller"
	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	"github.com/gin-gonic/gin"
)

// redisConn 全域 Redis 連線（供 handler 使用）
var redisConn redislib.IRedisConn

func main() {
	// 載入設定
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 建立 Redis 連線（根據 config.yaml 的 redis.mode 自動選擇實作）
	redisConn, err = cfg.ConnectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisConn.Close()

	// 設定 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化 Gin 引擎
	router := gin.Default()

	// 設定基本路由
	setupRoutes(router, redisConn)

	// 啟動服務器
	log.Printf("Starting server on %s with Redis mode: %s",
		cfg.GetServerAddr(), cfg.Redis.Mode)
	if err := router.Run(cfg.GetServerAddr()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes 設定所有路由
func setupRoutes(router *gin.Engine, redisConn redislib.IRedisConn) {
	// 建立 CacheController
	cacheController := controller.NewCacheController(redisConn)

	// 健康檢查端點
	router.GET("/health", healthCheck)

	// Cache API 路由
	router.GET("/cache", cacheController.GetCache)
	router.POST("/cache", cacheController.UpdateCache)
	router.GET("/fillcluster", cacheController.FillCluster)
}

// healthCheck 健康檢查處理器
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":          "healthy",
		"service":         "APGo Redis API",
		"version":         "1.0.0",
		"redis_mode":      "connected",
		"master_endpoint": redisConn.GetMasterEndpoint(),
		"slave_endpoint":  redisConn.GetSlaveEndpoint(),
	})
}
