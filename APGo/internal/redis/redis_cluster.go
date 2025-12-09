package redis

import (
	"context"
	"fmt"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	goredis "github.com/redis/go-redis/v9"
)

// RedisCluster 實作 Cluster 模式的 Redis 連線
type RedisCluster struct {
	client *goredis.ClusterClient
	nodes  []string
}

// NewRedisCluster 建立新的 Cluster 模式 Redis 連線
func NewRedisCluster(nodes []string) (*RedisCluster, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("at least one cluster node is required")
	}

	// 使用 Cluster 客戶端
	client := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs: nodes,
	})

	// 測試連線
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to cluster: %w", err)
	}

	rc := &RedisCluster{
		client: client,
		nodes:  nodes,
	}

	return rc, nil
}

// ReadAsync 從 Cluster 讀取資料
func (r *RedisCluster) ReadAsync(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", redislib.ErrKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", redislib.ErrReadFailed, err)
	}
	return val, nil
}

// WriteAsync 寫入資料到 Cluster
func (r *RedisCluster) WriteAsync(ctx context.Context, key string, value string) (bool, error) {
	err := r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return false, fmt.Errorf("%w: %v", redislib.ErrWriteFailed, err)
	}
	return true, nil
}

// GetRandomCache 讀取資料（Cluster 會自動路由到正確節點）
func (r *RedisCluster) GetRandomCache(ctx context.Context, key string) (string, error) {
	return r.ReadAsync(ctx, key)
}

// GetMasterEndpoint 取得 Master 端點（Cluster 模式返回節點列表）
func (r *RedisCluster) GetMasterEndpoint() string {
	if len(r.nodes) == 0 {
		return "cluster:unknown"
	}
	// 返回第一個節點作為代表
	return fmt.Sprintf("cluster:%s", r.nodes[0])
}

// GetSlaveEndpoint 取得 Slave 端點（Cluster 模式返回節點數量）
func (r *RedisCluster) GetSlaveEndpoint() string {
	return fmt.Sprintf("cluster:%d-nodes", len(r.nodes))
}

// Close 關閉連線
func (r *RedisCluster) Close() error {
	return r.client.Close()
}

// FillCluster 填充測試資料到 Cluster（用於測試 hash slot 分配）
func (r *RedisCluster) FillCluster(ctx context.Context, count int) error {
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("cluster:test:key:%d", i)
		value := fmt.Sprintf("value-%d", i)

		if err := r.client.Set(ctx, key, value, 0).Err(); err != nil {
			return fmt.Errorf("failed to fill cluster at key %s: %w", key, err)
		}
	}
	return nil
}

// GetClusterInfo 取得 Cluster 資訊
func (r *RedisCluster) GetClusterInfo(ctx context.Context) (string, error) {
	// 從第一個節點取得 cluster info
	result, err := r.client.ClusterInfo(ctx).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster info: %w", err)
	}
	return result, nil
}

// GetClusterNodes 取得 Cluster 節點資訊
func (r *RedisCluster) GetClusterNodes(ctx context.Context) (string, error) {
	result, err := r.client.ClusterNodes(ctx).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster nodes: %w", err)
	}
	return result, nil
}
