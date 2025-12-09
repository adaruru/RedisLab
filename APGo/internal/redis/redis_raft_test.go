package redis

import (
	"context"
	"testing"
	"time"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
)

// 驗證 RedisRaft 實作了 IRedisConn 介面
func TestRedisRaftImplementsInterface(t *testing.T) {
	var _ redislib.IRedisConn = (*RedisRaft)(nil)
}

func TestRedisRaft(t *testing.T) {
	// 這個測試需要實際的 Raft 環境
	t.Skip("需要實際的 Raft 環境才能執行")

	nodes := []string{
		"localhost:6379",
		"localhost:6380",
		"localhost:6381",
	}

	rr, err := NewRedisRaft(nodes)
	if err != nil {
		t.Fatalf("Failed to create RedisRaft: %v", err)
	}
	defer rr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 測試 Raft 資訊
	raftInfo, err := rr.GetRaftInfo(ctx)
	if err != nil {
		t.Fatalf("GetRaftInfo failed: %v", err)
	}
	t.Logf("Raft Info: %s", raftInfo)

	// 測試寫入（需要 Raft 共識）
	key := "test:raft:key"
	value := "test-value"

	success, err := rr.WriteAsync(ctx, key, value)
	if err != nil {
		t.Fatalf("WriteAsync failed: %v", err)
	}
	if !success {
		t.Fatal("WriteAsync returned false")
	}

	// 測試讀取（Strong Consistency）
	result, err := rr.ReadAsync(ctx, key)
	if err != nil {
		t.Fatalf("ReadAsync failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試 GetRandomCache
	result, err = rr.GetRandomCache(ctx, key)
	if err != nil {
		t.Fatalf("GetRandomCache failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試端點資訊
	masterEndpoint := rr.GetMasterEndpoint()
	if masterEndpoint == "" {
		t.Error("Master endpoint is empty")
	}
	t.Logf("Master (Leader) endpoint: %s", masterEndpoint)

	slaveEndpoint := rr.GetSlaveEndpoint()
	if slaveEndpoint == "" {
		t.Error("Slave endpoint is empty")
	}
	t.Logf("Slave (Followers) endpoint: %s", slaveEndpoint)

	// 測試節點資訊
	nodeInfo, err := rr.GetRaftNode(ctx)
	if err != nil {
		t.Fatalf("GetRaftNode failed: %v", err)
	}
	t.Logf("Raft Node Info: %s", nodeInfo)
}

func TestNewRedisRaft_InvalidParams(t *testing.T) {
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
			_, err := NewRedisRaft(tt.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisRaft() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
