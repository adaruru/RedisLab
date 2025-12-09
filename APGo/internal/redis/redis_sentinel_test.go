package redis

import (
	"context"
	"testing"
	"time"
)

func TestRedisSentinel(t *testing.T) {
	// 這個測試需要實際的 Sentinel 環境
	t.Skip("需要實際的 Sentinel 環境才能執行")

	masterName := "mymaster"
	sentinels := []string{"localhost:26379", "localhost:26380", "localhost:26381"}

	rs, err := NewRedisSentinel(masterName, sentinels)
	if err != nil {
		t.Fatalf("Failed to create RedisSentinel: %v", err)
	}
	defer rs.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 測試寫入
	key := "test:sentinel:key"
	value := "test-value"

	success, err := rs.WriteAsync(ctx, key, value)
	if err != nil {
		t.Fatalf("WriteAsync failed: %v", err)
	}
	if !success {
		t.Fatal("WriteAsync returned false")
	}

	// 測試讀取
	result, err := rs.ReadAsync(ctx, key)
	if err != nil {
		t.Fatalf("ReadAsync failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試 GetRandomCache
	result, err = rs.GetRandomCache(ctx, key)
	if err != nil {
		t.Fatalf("GetRandomCache failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 測試端點資訊
	masterEndpoint := rs.GetMasterEndpoint()
	if masterEndpoint == "" {
		t.Error("Master endpoint is empty")
	}
	t.Logf("Master endpoint: %s", masterEndpoint)

	slaveEndpoint := rs.GetSlaveEndpoint()
	if slaveEndpoint == "" {
		t.Error("Slave endpoint is empty")
	}
	t.Logf("Slave endpoint: %s", slaveEndpoint)
}

func TestNewRedisSentinel_InvalidParams(t *testing.T) {
	tests := []struct {
		name       string
		masterName string
		sentinels  []string
		wantErr    bool
	}{
		{
			name:       "empty master name",
			masterName: "",
			sentinels:  []string{"localhost:26379"},
			wantErr:    true,
		},
		{
			name:       "empty sentinels",
			masterName: "mymaster",
			sentinels:  []string{},
			wantErr:    true,
		},
		{
			name:       "nil sentinels",
			masterName: "mymaster",
			sentinels:  nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisSentinel(tt.masterName, tt.sentinels)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisSentinel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
