package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	"github.com/gin-gonic/gin"
)

// MockRedisConn 模擬 Redis 連線（用於測試）
type MockRedisConn struct {
	readFunc   func(ctx context.Context, key string) (string, error)
	writeFunc  func(ctx context.Context, key, value string) (bool, error)
	masterAddr string
	slaveAddr  string
}

func (m *MockRedisConn) ReadAsync(ctx context.Context, key string) (string, error) {
	if m.readFunc != nil {
		return m.readFunc(ctx, key)
	}
	return "", redislib.ErrKeyNotFound
}

func (m *MockRedisConn) WriteAsync(ctx context.Context, key, value string) (bool, error) {
	if m.writeFunc != nil {
		return m.writeFunc(ctx, key, value)
	}
	return true, nil
}

func (m *MockRedisConn) GetRandomCache(ctx context.Context, key string) (string, error) {
	return m.ReadAsync(ctx, key)
}

func (m *MockRedisConn) GetMasterEndpoint() string {
	if m.masterAddr != "" {
		return m.masterAddr
	}
	return "127.0.0.1:6379"
}

func (m *MockRedisConn) GetSlaveEndpoint() string {
	if m.slaveAddr != "" {
		return m.slaveAddr
	}
	return "127.0.0.1:6380"
}

func (m *MockRedisConn) Close() error {
	return nil
}

func TestGetCache_Success(t *testing.T) {
	// 設定模擬的 Redis 連線
	mockConn := &MockRedisConn{
		readFunc: func(ctx context.Context, key string) (string, error) {
			if key == "test-key" {
				return "test-value", nil
			}
			return "", redislib.ErrKeyNotFound
		},
	}

	// 建立控制器
	controller := NewCacheController(mockConn)

	// 設定 Gin 測試模式
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/cache", controller.GetCache)

	// 建立測試請求
	req, _ := http.NewRequest("GET", "/cache?key=test-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 驗證回應
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["key"] != "test-key" {
		t.Errorf("Expected key 'test-key', got %v", response["key"])
	}

	if response["value"] != "test-value" {
		t.Errorf("Expected value 'test-value', got %v", response["value"])
	}
}

func TestGetCache_KeyNotFound(t *testing.T) {
	mockConn := &MockRedisConn{
		readFunc: func(ctx context.Context, key string) (string, error) {
			return "", redislib.ErrKeyNotFound
		},
	}

	controller := NewCacheController(mockConn)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/cache", controller.GetCache)

	req, _ := http.NewRequest("GET", "/cache?key=non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetCache_MissingKey(t *testing.T) {
	mockConn := &MockRedisConn{}
	controller := NewCacheController(mockConn)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/cache", controller.GetCache)

	req, _ := http.NewRequest("GET", "/cache", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateCache_Success(t *testing.T) {
	mockConn := &MockRedisConn{
		writeFunc: func(ctx context.Context, key, value string) (bool, error) {
			return true, nil
		},
	}

	controller := NewCacheController(mockConn)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/cache", controller.UpdateCache)

	reqBody := CacheRequest{
		Key:   "test-key",
		Value: "test-value",
	}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/cache", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["key"] != "test-key" {
		t.Errorf("Expected key 'test-key', got %v", response["key"])
	}
}

func TestUpdateCache_InvalidRequest(t *testing.T) {
	mockConn := &MockRedisConn{}
	controller := NewCacheController(mockConn)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/cache", controller.UpdateCache)

	// 發送無效的 JSON
	req, _ := http.NewRequest("POST", "/cache", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
