package redis

import (
	"context"
	"fmt"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	goredis "github.com/redis/go-redis/v9"
)

// RedisRaft 實作 Raft 模式的 Redis 連線
type RedisRaft struct {
	client *goredis.Client
	nodes  []string
}

// NewRedisRaft 建立新的 Raft 模式 Redis 連線
func NewRedisRaft(nodes []string) (*RedisRaft, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("at least one raft node is required")
	}

	// 連線到第一個節點（通常是 Leader）
	// RedisRaft 會自動處理 Leader 重定向
	client := goredis.NewClient(&goredis.Options{
		Addr: nodes[0],
	})

	// 測試連線
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to raft node: %w", err)
	}

	rr := &RedisRaft{
		client: client,
		nodes:  nodes,
	}

	return rr, nil
}

// ReadAsync 從 Raft 讀取資料（Strong Consistency）
func (r *RedisRaft) ReadAsync(ctx context.Context, key string) (string, error) {
	// RedisRaft 提供強一致性讀取
	// 所有讀取都會經過 Raft 共識
	val, err := r.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", redislib.ErrKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", redislib.ErrReadFailed, err)
	}
	return val, nil
}

// WriteAsync 寫入資料到 Raft（Strong Consistency）
func (r *RedisRaft) WriteAsync(ctx context.Context, key string, value string) (bool, error) {
	// RedisRaft 的寫入會經過 Raft 共識
	// 需要多數節點確認才會成功
	err := r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return false, fmt.Errorf("%w: %v", redislib.ErrWriteFailed, err)
	}
	return true, nil
}

// GetRandomCache 讀取資料（Raft 保證強一致性）
func (r *RedisRaft) GetRandomCache(ctx context.Context, key string) (string, error) {
	return r.ReadAsync(ctx, key)
}

// GetMasterEndpoint 取得 Leader 端點
func (r *RedisRaft) GetMasterEndpoint() string {
	// 在 Raft 中，Leader 就是 Master
	if len(r.nodes) == 0 {
		return "raft:unknown"
	}
	return fmt.Sprintf("raft-leader:%s", r.nodes[0])
}

// GetSlaveEndpoint 取得 Follower 端點
func (r *RedisRaft) GetSlaveEndpoint() string {
	// Raft 中的 Follower 相當於 Slave
	return fmt.Sprintf("raft-followers:%d-nodes", len(r.nodes)-1)
}

// Close 關閉連線
func (r *RedisRaft) Close() error {
	return r.client.Close()
}

// GetRaftInfo 取得 Raft 集群資訊
func (r *RedisRaft) GetRaftInfo(ctx context.Context) (string, error) {
	// 使用 RAFT.INFO 命令取得集群狀態
	result, err := r.client.Do(ctx, "RAFT.INFO").Result()
	if err != nil {
		return "", fmt.Errorf("failed to get raft info: %w", err)
	}
	return fmt.Sprintf("%v", result), nil
}

// GetRaftNode 取得節點資訊
func (r *RedisRaft) GetRaftNode(ctx context.Context) (string, error) {
	result, err := r.client.Do(ctx, "RAFT.NODE").Result()
	if err != nil {
		return "", fmt.Errorf("failed to get raft node: %w", err)
	}
	return fmt.Sprintf("%v", result), nil
}
