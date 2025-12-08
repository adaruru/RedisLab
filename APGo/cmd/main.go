package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化 Gin 引擎
	router := gin.Default()

	// 設定基本路由
	setupRoutes(router)

	// 啟動服務器
	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes 設定所有路由
func setupRoutes(router *gin.Engine) {
	// 健康檢查端點
	router.GET("/health", healthCheck)

	// API 路由群組
	api := router.Group("/api")
	{
		// 預留給 cache controller 的路由
		api.GET("/cache", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"message": "Cache endpoints not yet implemented",
			})
		})
	}
}

// healthCheck 健康檢查處理器
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "APGo Redis API",
		"version": "1.0.0",
	})
}
