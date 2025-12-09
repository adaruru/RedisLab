package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
)

// 驗證 RedisCluster 實作了 IRedisConn 介面
func TestRedisClusterImplementsInterface(t *testing.T) {
	var _ redislib.IRedisConn = (*RedisCluster)(nil)
}

func TestRedisCluster(t *testing.T) {
	// 這個測試需要實際的 Cluster 環境
	t.Skip("需要實際的 Cluster 環境才能執行")

	nodes := []string{
		"localhost:7001",
		"localhost:7002",
		"localhost:7003",
		"localhost:7004",
		"localhost:7005",
		"localhost:7006",
	}

	rc, err := NewRedisCluster(nodes)
	if err != nil {
		t.Fatalf("Failed to create RedisCluster: %v", err)
	}
	defer rc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 測試 Cluster 資訊
	clusterInfo, err := rc.GetClusterInfo(ctx)
	if err != nil {
		t.Fatalf("GetClusterInfo failed: %v", err)
	}
	t.Logf("Cluster Info: %s", clusterInfo)

	// 測試寫入
	key := "test:cluster:key"
	value := "test-value"

	success, err := rc.WriteAsync(ctx, key, value)
	if err != nil {
		t.Fatalf("WriteAsync failed: %v", err)
	}
	if !success {
		t.Fatal("WriteAsync returned false")
	}

	// 測試讀取
	result, err := rc.ReadAsync(ctx, key)
	if err != nil {
		t.Fatalf("ReadAsync failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試 GetRandomCache
	result, err = rc.GetRandomCache(ctx, key)
	if err != nil {
		t.Fatalf("GetRandomCache failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試 FillCluster
	if err := rc.FillCluster(ctx, 10); err != nil {
		t.Fatalf("FillCluster failed: %v", err)
	}
	t.Log("FillCluster succeeded")

	// 驗證填充的資料
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("cluster:test:key:%d", i)
		expectedValue := fmt.Sprintf("value-%d", i)

		val, err := rc.ReadAsync(ctx, key)
		if err != nil {
			t.Errorf("Failed to read key %s: %v", key, err)
			continue
		}
		if val != expectedValue {
			t.Errorf("Key %s: expected %s, got %s", key, expectedValue, val)
		}
	}

	// 測試端點資訊
	masterEndpoint := rc.GetMasterEndpoint()
	if masterEndpoint == "" {
		t.Error("Master endpoint is empty")
	}
	t.Logf("Master endpoint: %s", masterEndpoint)

	slaveEndpoint := rc.GetSlaveEndpoint()
	if slaveEndpoint == "" {
		t.Error("Slave endpoint is empty")
	}
	t.Logf("Slave endpoint: %s", slaveEndpoint)
}

func TestNewRedisCluster_InvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []string
		wantErr bool
	}{
		{
			name:    "empty nodes",
			nodes:   []string{},
			wantErr: true,
		},
		{
			name:    "nil nodes",
			nodes:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisCluster(tt.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
