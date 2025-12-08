# APGo - Go Gin Base API with Redis Support

這是一個基於 Go Gin 框架的 API 專案，實作了 redislib 來支援多種 Redis 部署模式。

## 專案目標

參考 [AP](../AP) 資料夾的 .NET Core 實作，建立一個 Go 版本的 Redis 操作 API，支援以下 Redis 模式：

- Redis Master-Slave
- Redis Sentinel
- Redis Cluster
- Redis Raft

## 實作步驟

### 步驟 1: 初始化 Go Module 和專案結構

- 初始化 Go module (`go mod init`)
- 建立專案目錄結構：
  - `cmd/` - 主程式入口
  - `internal/redis/` - Redis 連線實作
  - `internal/controller/` - API 控制器
  - `internal/config/` - 設定檔處理
  - `pkg/redislib/` - Redis 介面定義（可重用的套件）

1. cd APGo && go mod init
2. cd APGo && mkdir -p cmd internal/redis internal/controller internal/config pkg/redislib

目前專案結構：
APGo/
├── cmd/              (主程式入口)
├── internal/
│   ├── config/       (設定檔處理)
│   ├── controller/   (API 控制器)
│   └── redis/        (Redis 連線實作)
├── pkg/
│   └── redislib/     (Redis 介面定義)
├── go.mod
├── .gitignore
└── README.md
現在可以繼續執行步驟二：建立 Gin 基礎 API 框架。要繼續嗎？

### 步驟 2: 建立 Gin 基礎 API 框架

- 在 `cmd/main.go` 建立主程式
- 初始化 Gin 引擎
- 設定基本路由
- 實作健康檢查端點 (health check)

### 步驟 3: 實作 redislib 介面定義

- 定義 `IRedisConn` 介面，對應 C# 的 `IRedisConn`
  - `ReadAsync(key string) (string, error)`
  - `WriteAsync(key, value string) (bool, error)`
  - `GetRandomCache(key string) (string, error)`
  - `MasterEndpoint() string`
  - `SlaveEndpoint() string`
- 定義 `RedisMode` 列舉

### 步驟 4: 實作 RedisMasterSlave 連線

- 建立 `RedisMasterSlave` 結構，實作 `IRedisConn` 介面
- 使用 `go-redis/redis` 套件
- 實作讀寫分離邏輯（Master 寫入，Slave 讀取）
- 處理連線和錯誤

### 步驟 5: 實作 RedisSentinel 連線

- 建立 `RedisSentinel` 結構，實作 `IRedisConn` 介面
- 使用 Sentinel 模式連線
- 實作自動故障轉移支援
- 取得 Master/Slave 端點資訊

### 步驟 6: 實作 RedisCluster 連線

- 建立 `RedisCluster` 結構，實作 `IRedisConn` 介面
- 使用 Redis Cluster 模式
- 實作 `FillCluster()` 方法填充測試資料
- 處理叢集節點路由

### 步驟 7: 實作 RedisRaft 連線

- 建立 `RedisRaft` 結構，實作 `IRedisConn` 介面
- 使用 Redis Raft 模式連線
- 實作一致性讀寫

### 步驟 8: 建立 DI 容器和設定載入

- 實作設定檔讀取（支援 JSON/YAML）
- 建立依賴注入容器（可使用 `uber-go/dig` 或手動實作）
- 根據設定檔的 `Redis:Mode` 自動註冊對應的 Redis 實作
- 參考 C# 的 `RedisDI.AddRedisService` 方法

### 步驟 9: 實作 CacheController API 端點

- 建立 `CacheController` 控制器
- 實作以下端點：
  - `GET /cache?key=xxx` - 讀取快取
  - `POST /cache` - 更新快取（Request Body: `{key, value}`）
  - `GET /fillcluster` - 填充 Cluster 測試資料
- 返回對應的 Master/Slave 端點資訊

### 步驟 10: 建立設定檔範例

- 建立 `config.json` 或 `config.yaml`
- 提供各種 Redis 模式的設定範例：
  - Master-Slave 設定
  - Sentinel 設定
  - Cluster 設定
  - Raft 設定

### 步驟 11: 測試和文件

- 建立單元測試
- 建立整合測試
- 撰寫 API 使用文件
- 更新本 README 加入使用說明

## 技術棧

- **Web Framework**: Gin
- **Redis Client**: go-redis/redis v9
- **Config**: viper
- **DI**: uber-go/dig (或手動實作)

## 參考資料

- [AP (C# 實作)](../AP) - .NET Core 版本的參考實作
- [Redis 部署設定](../README.md) - 各種 Redis 模式的 Docker Compose 設定

## 專案狀態

🚧 開發中
