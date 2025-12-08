package redislib

import "context"

// IRedisConn 定義 Redis 連線介面
// 對應 C# 的 IRedisConn 介面
type IRedisConn interface {
	// ReadAsync 從 Redis 讀取資料（通常從 Slave 讀取）
	ReadAsync(ctx context.Context, key string) (string, error)

	// WriteAsync 寫入資料到 Redis（通常寫入 Master）
	WriteAsync(ctx context.Context, key string, value string) (bool, error)

	// GetRandomCache 隨機取得快取資料
	GetRandomCache(ctx context.Context, key string) (string, error)

	// GetMasterEndpoint 取得 Master 端點資訊
	GetMasterEndpoint() string

	// GetSlaveEndpoint 取得 Slave 端點資訊
	GetSlaveEndpoint() string

	// Close 關閉連線
	Close() error
}
