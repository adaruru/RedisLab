package redis

import (
	"context"
	"testing"
	"time"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
)

// 驗證 RedisMasterSlave 實作了 IRedisConn 介面
func TestRedisMasterSlaveImplementsInterface(t *testing.T) {
	var _ redislib.IRedisConn = (*RedisMasterSlave)(nil)
}

func TestRedisMasterSlave(t *testing.T) {
	// 這個測試需要實際的 Master-Slave 環境
	t.Skip("需要實際的 Master-Slave 環境才能執行")

	master := "localhost:6379"
	slaves := []string{"localhost:6380", "localhost:6381"}

	rms, err := NewRedisMasterSlave(master, slaves)
	if err != nil {
		t.Fatalf("Failed to create RedisMasterSlave: %v", err)
	}
	defer rms.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 測試寫入
	key := "test:master-slave:key"
	value := "test-value"

	success, err := rms.WriteAsync(ctx, key, value)
	if err != nil {
		t.Fatalf("WriteAsync failed: %v", err)
	}
	if !success {
		t.Fatal("WriteAsync returned false")
	}

	// 測試讀取
	result, err := rms.ReadAsync(ctx, key)
	if err != nil {
		t.Fatalf("ReadAsync failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試 GetRandomCache
	result, err = rms.GetRandomCache(ctx, key)
	if err != nil {
		t.Fatalf("GetRandomCache failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試端點資訊
	masterEndpoint := rms.GetMasterEndpoint()
	if masterEndpoint == "" {
		t.Error("Master endpoint is empty")
	}
	t.Logf("Master endpoint: %s", masterEndpoint)

	slaveEndpoint := rms.GetSlaveEndpoint()
	if slaveEndpoint == "" {
		t.Error("Slave endpoint is empty")
	}
	t.Logf("Slave endpoint: %s", slaveEndpoint)
}

func TestNewRedisMasterSlave_InvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		master  string
		slaves  []string
		wantErr bool
	}{
		{
			name:    "empty master",
			master:  "",
			slaves:  []string{"localhost:6380"},
			wantErr: true,
		},
		{
			name:    "empty slaves",
			master:  "localhost:6379",
			slaves:  []string{},
			wantErr: false, // Note: 沒有 slave 也可以，會使用 master 作為備用
		},
		{
			name:    "nil slaves",
			master:  "localhost:6379",
			slaves:  nil,
			wantErr: false, // Note: nil slaves 也可以，會使用 master 作為備用
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisMasterSlave(tt.master, tt.slaves)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisMasterSlave() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
