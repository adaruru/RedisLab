package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/AmandaChou/RedisLab/APGo/pkg/redislib"
	"github.com/spf13/viper"
)

// Config 應用程式設定
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

// ServerConfig 服務器設定
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// RedisConfig Redis 設定
type RedisConfig struct {
	Mode        string                 `mapstructure:"mode"`
	MasterSlave MasterSlaveConfig      `mapstructure:"master_slave"`
	Sentinel    SentinelConfig         `mapstructure:"sentinel"`
	Cluster     ClusterConfig          `mapstructure:"cluster"`
	Raft        RaftConfig             `mapstructure:"raft"`
}

// MasterSlaveConfig 主從模式設定
type MasterSlaveConfig struct {
	Description string   `mapstructure:"description"`
	Master      string   `mapstructure:"master"`
	Slaves      []string `mapstructure:"slaves"`
}

// SentinelConfig Sentinel 設定
type SentinelConfig struct {
	Description string   `mapstructure:"description"`
	MasterName  string   `mapstructure:"master_name"`
	Sentinels   []string `mapstructure:"sentinels"`
}

// ClusterConfig Cluster 設定
type ClusterConfig struct {
	Description string   `mapstructure:"description"`
	Nodes       []string `mapstructure:"nodes"`
}

// RaftConfig Raft 設定
type RaftConfig struct {
	Description string   `mapstructure:"description"`
	Nodes       []string `mapstructure:"nodes"`
}

// LoadConfig 載入設定檔
func LoadConfig() (*Config, error) {
	v := viper.New()

	// 設定檔案名稱和類型（Go 慣例使用 YAML）
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 設定搜尋路徑
	v.AddConfigPath(".")           // 當前目錄
	v.AddConfigPath("./config")    // config 子目錄
	v.AddConfigPath("/etc/apgo")   // 系統設定目錄

	// 支援環境變數覆蓋
	v.SetEnvPrefix("APGO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 讀取基本設定檔
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// 設定檔不存在，使用預設值
	}

	// 檢查環境特定設定檔
	env := getEnv()
	if env != "" {
		// 嘗試載入環境特定設定（例如 config.docker.yaml）
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := v.MergeInConfig(); err != nil {
			// 環境設定檔不存在不算錯誤
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to merge environment config: %w", err)
			}
		}
	}

	// 解析設定到結構
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 環境變數覆蓋 Redis Mode
	if envMode := os.Getenv("APGO_REDIS_MODE"); envMode != "" {
		config.Redis.Mode = envMode
	}

	// 設定預設值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}
	if config.Redis.Mode == "" {
		config.Redis.Mode = "RedisMasterSlaves"
	}

	return &config, nil
}

// getEnv 取得環境名稱
func getEnv() string {
	// 優先使用 GO_ENV（Go 慣例）
	if env := os.Getenv("GO_ENV"); env != "" {
		return strings.ToLower(env)
	}
	// 其次使用 APP_ENV
	if env := os.Getenv("APP_ENV"); env != "" {
		return strings.ToLower(env)
	}
	// 相容 .NET 的 ASPNETCORE_ENVIRONMENT
	if env := os.Getenv("ASPNETCORE_ENVIRONMENT"); env != "" {
		return strings.ToLower(env)
	}
	return ""
}

// GetRedisMode 取得 Redis 模式
func (c *Config) GetRedisMode() (redislib.RedisMode, error) {
	return redislib.ParseRedisMode(c.Redis.Mode)
}

// GetServerAddr 取得服務器監聽位址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}
