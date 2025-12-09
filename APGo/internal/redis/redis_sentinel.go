package redis

import (
	"context"
	"fmt"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	goredis "github.com/redis/go-redis/v9"
)

// RedisSentinel 實作 Sentinel 模式的 Redis 連線
type RedisSentinel struct {
	client         *goredis.Client
	masterName     string
	sentinels      []string
	masterEndpoint string
	slaveEndpoint  string
}

// NewRedisSentinel 建立新的 Sentinel 模式 Redis 連線
func NewRedisSentinel(masterName string, sentinels []string) (*RedisSentinel, error) {
	if masterName == "" {
		return nil, fmt.Errorf("master name is required")
	}
	if len(sentinels) == 0 {
		return nil, fmt.Errorf("at least one sentinel is required")
	}

	// 使用 Sentinel 客戶端
	client := goredis.NewFailoverClient(&goredis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinels,
	})

	// 測試連線
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to sentinel: %w", err)
	}

	rs := &RedisSentinel{
		client:     client,
		masterName: masterName,
		sentinels:  sentinels,
	}

	// 取得當前 Master 和 Slave 端點
	if err := rs.updateEndpoints(ctx); err != nil {
		// 端點更新失敗不算致命錯誤
		fmt.Printf("Warning: failed to update endpoints: %v\n", err)
	}

	return rs, nil
}

// updateEndpoints 更新 Master 和 Slave 端點資訊
func (r *RedisSentinel) updateEndpoints(ctx context.Context) error {
	// 透過 Sentinel 查詢當前 Master
	sentinelClient := goredis.NewSentinelClient(&goredis.Options{
		Addr: r.sentinels[0],
	})
	defer sentinelClient.Close()

	// 取得 Master 位址
	masterAddr, err := sentinelClient.GetMasterAddrByName(ctx, r.masterName).Result()
	if err != nil {
		return fmt.Errorf("failed to get master address: %w", err)
	}
	if len(masterAddr) >= 2 {
		r.masterEndpoint = fmt.Sprintf("%s:%s", masterAddr[0], masterAddr[1])
	}

	// 取得 Slave 位址
	slaves, err := sentinelClient.Sentinels(ctx, r.masterName).Result()
	if err == nil && len(slaves) > 0 {
		// 使用第一個 Slave
		if slave, ok := slaves[0].(map[interface{}]interface{}); ok {
			if ip, ok := slave["ip"].(string); ok {
				if port, ok := slave["port"].(string); ok {
					r.slaveEndpoint = fmt.Sprintf("%s:%s", ip, port)
				}
			}
		}
	}

	// 如果沒有 Slave，使用 Master
	if r.slaveEndpoint == "" {
		r.slaveEndpoint = r.masterEndpoint
	}

	return nil
}

// ReadAsync 從 Redis 讀取資料
func (r *RedisSentinel) ReadAsync(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", redislib.ErrKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", redislib.ErrReadFailed, err)
	}
	return val, nil
}

// WriteAsync 寫入資料到 Redis
func (r *RedisSentinel) WriteAsync(ctx context.Context, key string, value string) (bool, error) {
	err := r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return false, fmt.Errorf("%w: %v", redislib.ErrWriteFailed, err)
	}
	return true, nil
}

// GetRandomCache 讀取資料（Sentinel 會自動路由）
func (r *RedisSentinel) GetRandomCache(ctx context.Context, key string) (string, error) {
	return r.ReadAsync(ctx, key)
}

// GetMasterEndpoint 取得 Master 端點
func (r *RedisSentinel) GetMasterEndpoint() string {
	if r.masterEndpoint == "" {
		return fmt.Sprintf("sentinel:%s", r.masterName)
	}
	return r.masterEndpoint
}

// GetSlaveEndpoint 取得 Slave 端點
func (r *RedisSentinel) GetSlaveEndpoint() string {
	if r.slaveEndpoint == "" {
		return r.GetMasterEndpoint()
	}
	return r.slaveEndpoint
}

// Close 關閉連線
func (r *RedisSentinel) Close() error {
	return r.client.Close()
}
