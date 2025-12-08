package redis

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	goredis "github.com/redis/go-redis/v9"
)

// RedisMasterSlave 實作主從模式的 Redis 連線
type RedisMasterSlave struct {
	master         *goredis.Client
	slave          *goredis.Client
	slaves         []*goredis.Client
	masterEndpoint string
	slaveEndpoint  string
}

// NewRedisMasterSlave 建立新的主從模式 Redis 連線
func NewRedisMasterSlave(master string, slaves []string) (*RedisMasterSlave, error) {
	if master == "" {
		return nil, fmt.Errorf("master endpoint is required")
	}

	// 連線到 Master
	masterClient := goredis.NewClient(&goredis.Options{
		Addr: master,
	})

	// 測試 Master 連線
	ctx := context.Background()
	if err := masterClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to master %s: %w", master, err)
	}

	rms := &RedisMasterSlave{
		master:         masterClient,
		masterEndpoint: master,
		slaves:         make([]*goredis.Client, 0, len(slaves)),
	}

	// 連線到所有 Slaves
	for _, slaveAddr := range slaves {
		slave := goredis.NewClient(&goredis.Options{
			Addr: slaveAddr,
		})

		// 測試 Slave 連線
		if err := slave.Ping(ctx).Err(); err != nil {
			// 如果 Slave 連線失敗，記錄錯誤但繼續
			fmt.Printf("Warning: failed to connect to slave %s: %v\n", slaveAddr, err)
			continue
		}

		rms.slaves = append(rms.slaves, slave)

		// 使用第一個成功連線的 Slave 作為預設 Slave
		if rms.slave == nil {
			rms.slave = slave
			rms.slaveEndpoint = slaveAddr
		}
	}

	// 如果沒有可用的 Slave，使用 Master 作為備用
	if rms.slave == nil {
		rms.slave = masterClient
		rms.slaveEndpoint = master
	}

	return rms, nil
}

// ReadAsync 從 Slave 讀取資料
func (r *RedisMasterSlave) ReadAsync(ctx context.Context, key string) (string, error) {
	val, err := r.slave.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", redislib.ErrKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", redislib.ErrReadFailed, err)
	}
	return val, nil
}

// WriteAsync 寫入資料到 Master
func (r *RedisMasterSlave) WriteAsync(ctx context.Context, key string, value string) (bool, error) {
	err := r.master.Set(ctx, key, value, 0).Err()
	if err != nil {
		return false, fmt.Errorf("%w: %v", redislib.ErrWriteFailed, err)
	}
	return true, nil
}

// GetRandomCache 隨機從一個 Slave 讀取資料
func (r *RedisMasterSlave) GetRandomCache(ctx context.Context, key string) (string, error) {
	if len(r.slaves) == 0 {
		// 如果沒有 Slave，從 Master 讀取
		return r.ReadAsync(ctx, key)
	}

	// 隨機打亂 slaves 順序
	indices := rand.Perm(len(r.slaves))

	// 嘗試從隨機順序的 Slave 讀取
	for _, idx := range indices {
		slave := r.slaves[idx]
		val, err := slave.Get(ctx, key).Result()
		if err == goredis.Nil {
			continue // Key 不存在，嘗試下一個 Slave
		}
		if err != nil {
			// 連線錯誤，嘗試下一個 Slave
			fmt.Printf("Warning: failed to read from slave: %v\n", err)
			continue
		}
		return val, nil
	}

	return "", redislib.ErrKeyNotFound
}

// GetMasterEndpoint 取得 Master 端點
func (r *RedisMasterSlave) GetMasterEndpoint() string {
	return r.masterEndpoint
}

// GetSlaveEndpoint 取得 Slave 端點
func (r *RedisMasterSlave) GetSlaveEndpoint() string {
	return r.slaveEndpoint
}

// Close 關閉所有連線
func (r *RedisMasterSlave) Close() error {
	var lastErr error

	// 關閉 Master
	if err := r.master.Close(); err != nil {
		lastErr = err
	}

	// 關閉所有 Slaves
	for _, slave := range r.slaves {
		if err := slave.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
