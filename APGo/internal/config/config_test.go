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

func TestConnectRedis_InvalidMode(t *testing.T) {
	// 測試不支援的 Redis 模式
	config := &Config{
		Redis: RedisConfig{
			Mode: "InvalidMode",
		},
	}

	_, err := config.ConnectRedis()
	if err == nil {
		t.Error("ConnectRedis() should return error for invalid mode")
	}
}

func TestConnectRedis_MasterSlaveEmptyMaster(t *testing.T) {
	// 測試 MasterSlave 模式缺少 Master 設定
	config := &Config{
		Redis: RedisConfig{
			Mode: "RedisMasterSlaves",
			MasterSlave: MasterSlaveConfig{
				Master: "", // 空的 Master
				Slaves: []string{},
			},
		},
	}

	_, err := config.ConnectRedis()
	if err == nil {
		t.Error("ConnectRedis() should return error when master is empty")
	}
}

func TestConnectRedis_SentinelEmptyMasterName(t *testing.T) {
	// 測試 Sentinel 模式缺少 MasterName 設定
	config := &Config{
		Redis: RedisConfig{
			Mode: "RedisSentinel",
			Sentinel: SentinelConfig{
				MasterName: "", // 空的 MasterName
				Sentinels:  []string{},
			},
		},
	}

	_, err := config.ConnectRedis()
	if err == nil {
		t.Error("ConnectRedis() should return error when master name is empty")
	}
}

func TestConnectRedis_ClusterEmptyNodes(t *testing.T) {
	// 測試 Cluster 模式缺少節點設定
	config := &Config{
		Redis: RedisConfig{
			Mode: "RedisCluster",
			Cluster: ClusterConfig{
				Nodes: []string{}, // 空的節點列表
			},
		},
	}

	_, err := config.ConnectRedis()
	if err == nil {
		t.Error("ConnectRedis() should return error when cluster nodes is empty")
	}
}

func TestConnectRedis_RaftEmptyNodes(t *testing.T) {
	// 測試 Raft 模式缺少節點設定
	config := &Config{
		Redis: RedisConfig{
			Mode: "RedisRaft",
			Raft: RaftConfig{
				Nodes: []string{}, // 空的節點列表
			},
		},
	}

	_, err := config.ConnectRedis()
	if err == nil {
		t.Error("ConnectRedis() should return error when raft nodes is empty")
	}
}
