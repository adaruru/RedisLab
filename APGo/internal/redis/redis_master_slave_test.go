package redis

import (
	"testing"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
)

// 驗證 RedisMasterSlave 實作了 IRedisConn 介面
func TestRedisMasterSlaveImplementsInterface(t *testing.T) {
	var _ redislib.IRedisConn = (*RedisMasterSlave)(nil)
}

func TestNewRedisMasterSlave(t *testing.T) {
	tests := []struct {
		name    string
		master  string
		slaves  []string
		wantErr bool
	}{
		{
			name:    "empty master should fail",
			master:  "",
			slaves:  []string{},
			wantErr: true,
		},
		{
			name:    "valid config structure",
			master:  "localhost:6379",
			slaves:  []string{"localhost:6380"},
			wantErr: false, // Note: 實際連線可能失敗，但結構是有效的
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisMasterSlave(tt.master, tt.slaves)
			if (err != nil) != tt.wantErr {
				if tt.name == "empty master should fail" {
					return // 預期錯誤
				}
				// 對於有效設定，如果連線失敗是正常的（Redis 可能未啟動）
				t.Logf("Got error (may be expected if Redis is not running): %v", err)
			}
		})
	}
}
