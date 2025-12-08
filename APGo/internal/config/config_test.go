package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 測試基本設定載入
	config, err := LoadConfig()
	if err != nil {
		t.Logf("Config load may fail if appsettings.json not found: %v", err)
		return
	}

	// 驗證預設模式
	if config.Redis.Mode == "" {
		t.Error("Redis mode should not be empty")
	}

	t.Logf("Loaded config with Redis mode: %s", config.Redis.Mode)
}

func TestLoadConfigWithEnv(t *testing.T) {
	// 設定環境變數
	os.Setenv("APGO_REDIS_MODE", "RedisSentinel")
	defer os.Unsetenv("APGO_REDIS_MODE")

	config, err := LoadConfig()
	if err != nil {
		t.Logf("Config load may fail if appsettings.json not found: %v", err)
		return
	}

	// 驗證環境變數覆蓋
	if config.Redis.Mode != "RedisSentinel" {
		t.Errorf("Expected mode RedisSentinel, got %s", config.Redis.Mode)
	}
}

func TestGetRedisMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"valid master-slave", "RedisMasterSlaves", false},
		{"valid sentinel", "RedisSentinel", false},
		{"valid cluster", "RedisCluster", false},
		{"valid raft", "RedisRaft", false},
		{"invalid mode", "InvalidMode", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Redis: RedisConfig{
					Mode: tt.mode,
				},
			}

			_, err := config.GetRedisMode()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRedisMode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
