package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AmandaChou/RedisLab/APGo/internal/redis"
	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	"github.com/gin-gonic/gin"
)

// CacheController 快取控制器
type CacheController struct {
	redisConn redislib.IRedisConn
}

// NewCacheController 建立新的快取控制器
func NewCacheController(redisConn redislib.IRedisConn) *CacheController {
	return &CacheController{
		redisConn: redisConn,
	}
}

// CacheRequest 快取請求
type CacheRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// GetCache 讀取快取
// @Summary 讀取快取
// @Description 從 Redis 讀取指定 key 的值（從 Slave/Replica 讀取）
// @Tags Cache
// @Param key query string true "快取鍵"
// @Success 200 {object} map[string]interface{} "成功讀取"
// @Failure 404 {object} map[string]interface{} "找不到鍵"
// @Failure 500 {object} map[string]interface{} "讀取失敗"
// @Router /cache [get]
func (cc *CacheController) GetCache(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "key is required",
			"message": "請提供 key 參數",
		})
		return
	}

	ctx := context.Background()
	value, err := cc.redisConn.ReadAsync(ctx, key)
	if err != nil {
		if err == redislib.ErrKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":     "key not found",
				"key":       key,
				"message":   fmt.Sprintf("key '%s' not found", key),
				"read_from": cc.redisConn.GetSlaveEndpoint(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "read failed",
			"key":       key,
			"message":   err.Error(),
			"read_from": cc.redisConn.GetSlaveEndpoint(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":       key,
		"value":     value,
		"message":   fmt.Sprintf("value: %s", value),
		"read_from": cc.redisConn.GetSlaveEndpoint(),
	})
}

// UpdateCache 更新快取
// @Summary 更新快取
// @Description 寫入資料到 Redis（寫入 Master）
// @Tags Cache
// @Accept json
// @Produce json
// @Param request body CacheRequest true "快取請求"
// @Success 200 {object} map[string]interface{} "成功寫入"
// @Failure 400 {object} map[string]interface{} "請求參數錯誤"
// @Failure 500 {object} map[string]interface{} "寫入失敗"
// @Router /cache [post]
func (cc *CacheController) UpdateCache(c *gin.Context) {
	var req CacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx := context.Background()
	success, err := cc.redisConn.WriteAsync(ctx, req.Key, req.Value)
	if err != nil || !success {
		errMsg := "write failed"
		if err != nil {
			errMsg = err.Error()
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "write failed",
			"key":        req.Key,
			"value":      req.Value,
			"message":    errMsg,
			"written_to": cc.redisConn.GetMasterEndpoint(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":        req.Key,
		"value":      req.Value,
		"message":    fmt.Sprintf("key '%s', value '%s' well saved", req.Key, req.Value),
		"written_to": cc.redisConn.GetMasterEndpoint(),
	})
}

// FillCluster 填充 Cluster 測試資料
// @Summary 填充 Cluster 測試資料
// @Description 批次填充測試資料到 Redis Cluster（僅 Cluster 模式支援）
// @Tags Cache
// @Success 200 {object} map[string]interface{} "成功填充"
// @Failure 400 {object} map[string]interface{} "不支援的模式"
// @Failure 500 {object} map[string]interface{} "填充失敗"
// @Router /fillcluster [get]
func (cc *CacheController) FillCluster(c *gin.Context) {
	// 檢查是否為 Cluster 模式
	clusterConn, ok := cc.redisConn.(*redis.RedisCluster)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported mode",
			"message": "FillCluster only supports RedisCluster mode",
			"mode":    fmt.Sprintf("%T", cc.redisConn),
		})
		return
	}

	ctx := context.Background()
	count := 100 // 預設填充 100 筆資料

	if err := clusterConn.FillCluster(ctx, count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "fill failed",
			"message": err.Error(),
			"count":   count,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully filled %d test records to cluster", count),
		"count":   count,
		"mode":    "RedisCluster",
	})
}
