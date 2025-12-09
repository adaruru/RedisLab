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

### 步驟 2: 建立 Gin 基礎 API 框架

- 在 `cmd/main.go` 建立主程式
- 初始化 Gin 引擎
  - 安裝: cd APGo && go get -u github.com/gin-gonic/gin
  - 測試: cd APGo && go build -o bin/apgo ./cmd/main.go
- 設定基本路由
- 實作健康檢查端點 (health check)
- Makefile 來簡化後續的建置和執行

### 步驟 3: 實作 redislib 介面定義

- 參考 [AP](../AP) 資料夾的 .NET Core 實作，建立一個 Go 版本的 Redis 操作 API，支援以下 Redis 模式
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
- 建立 `redis_master_slave_test.go` 測試檔案
  - 介面實作驗證測試 (`TestRedisMasterSlaveImplementsInterface`)
  - 完整整合測試 (`TestRedisMasterSlave`)，使用 `t.Skip()` 跳過
  - 參數驗證測試 (`TestNewRedisMasterSlave_InvalidParams`)
- **測試注意事項**：
  - 整合測試需要實際 Redis 環境，使用 `t.Skip()` 預設跳過
  - 參數驗證測試不需要 Redis，可直接執行
  - 執行整合測試時需手動註解掉 `t.Skip()`，參考 [CONFIG.md 整合測試步驟](CONFIG.md#整合測試)

#### 步驟 4.1 確認不同環境如何設定練線設定
- 參考 AP\appsettings.json 設定 "Redis": 且在 program 讀取設定，compose 階段指定環境變數
- 確認 go gin 如何配置環境變數
- 參考 redis-master-slave\docker-compose-ap-go.yml
- 編輯 redis-master-slave\docker-compose-ap-go.yml 新增啟動 compose

#### 步驟 4.2 修改環境變數檔案最佳實作
  - AP\appsettings.json
  - AP\appsettings.RedisMasterSlaves.json
  - 這是 .Net core convention，我確認了 config 來自 Viper 套件
  - 請修改成 Golang 最佳實作，或是 Viper 最佳實作

  1. 設定檔改為 Go/Viper 的最佳實作方式。Go 社群通常使用 YAML 格式，檔名使用 config.yaml 而非 appsettings.json。
  2. snake_case 命名
  3. 標準環境變數 GO_ENV 是 Go 社群標準
  4. 使用 mapstructure 而非 json/yaml
  5. 支援多個搜尋路徑、可放在系統設定目錄

#### 步驟 4.3 修改環境變數要具有測試鑑別度
 - docker-compose-ap-go.yml，使用 GO_ENV=docker 而非 ASPNETCORE_ENVIRONMENT=RedisMasterSlaves
 - 但我有多個實作要測試，四種架構四種環境，
   1. AP\appsettings.RedisCluster.json
   2. AP\appsettings.RedisMasterSlaves.json
   3. AP\appsettings.RedisSentinel.json
   4. AP\appsettings.RedisRaft.json
 - 你現在的 GO_ENV=docker 沒有任何識別度，確認我的考量有意義
 - 說明 GO_ENV=docker 沒有任何識別度，卻比較好的原因
 - 環境 = Redis 模式: 是因為這只是 Lib POC 專案
 - 現實環境，Redis 模式不會因環境的變化，環境變數覆蓋並不易於閱讀，在實際專案也不會有這個需求

  1. 這是 POC 專案：專門測試不同 Redis 架構
  2. 環境 = Redis 模式（實用主義）
  3. APGo/
├── config.yaml              # 開發環境（預設）
├── config.master-slave.yaml # Docker 主從模式
├── config.sentinel.yaml     # Docker 哨兵模式
├── config.cluster.yaml      # Docker 叢集模式
└── config.raft.yaml         # Docker Raft 模式
  4. 識別度高：環境名稱 = Redis 模式，一看就懂
  5. 生產環境一旦選定架構就固定了（不會今天用 Cluster，明天改 Sentinel）
  6. 環境變數覆蓋不易於閱讀
  7. GO_ENV=master-slave  # ✅ 清楚明確


### 步驟 5: 實作 RedisSentinel 連線

- 建立 `RedisSentinel` 結構，實作 `IRedisConn` 介面
  - 使用 Sentinel 模式連線 (`goredis.NewFailoverClient`)
  - 實作自動故障轉移支援
  - 取得 Master/Slave 端點資訊
  - 參考 redis-sentinel\docker-compose.yml
  - 編輯 redis-sentinel\docker-compose-ap-go.yml 新增啟動 compose
- 建立 `redis_sentinel_test.go` 測試檔案
  - 介面實作驗證測試 (`TestRedisSentinelImplementsInterface`)
  - 完整整合測試 (`TestRedisSentinel`)，使用 `t.Skip()` 跳過
  - 參數驗證測試 (`TestNewRedisSentinel_InvalidParams`)
- **測試注意事項**：
  - 整合測試需要 Sentinel 環境，使用 `t.Skip()` 預設跳過
  - Sentinel 需要較長啟動時間（約 20 秒）
  - 執行整合測試參考 [CONFIG.md Sentinel 測試步驟](CONFIG.md#2-sentinel-模式整合測試)

步驟 5 完成內容：

1. ✅ 建立 [redis_sentinel.go](vscode-webview://0nlu7ssdt85f5uhh8ljum9dikvvs8gsel4mc6uulua9pmps9lc22/APGo/internal/redis/redis_sentinel.go) - 實作 `RedisSentinel` 結構
   - 使用 `goredis.NewFailoverClient` 實作 Sentinel 模式
   - 自動故障轉移支援（Failover）
   - 透過 Sentinel 查詢 Master/Slave 端點資訊
   - 實作完整的 `IRedisConn` 介面
2. ✅ 建立 [redis_sentinel_test.go](vscode-webview://0nlu7ssdt85f5uhh8ljum9dikvvs8gsel4mc6uulua9pmps9lc22/APGo/internal/redis/redis_sentinel_test.go) - 單元測試
   - 測試連線、讀寫、端點資訊
   - 參數驗證測試
3. ✅ 建立 [docker-compose-ap-go.yml](vscode-webview://0nlu7ssdt85f5uhh8ljum9dikvvs8gsel4mc6uulua9pmps9lc22/redis-sentinel/docker-compose-ap-go.yml) - Docker Compose 設定
   - 使用 `GO_ENV=sentinel` 環境變數
   - 整合 3 個 Sentinel 節點
   - 整合 Master-Slave 架構
   - 網路設定完整

設計重點：

- 使用 `NewFailoverClient` 讓 go-redis 自動處理故障轉移
- 透過 Sentinel 客戶端查詢當前的 Master/Slave 位址
- 符合步驟 4.3 的設計理念：`GO_ENV=sentinel` 直接識別 Sentinel 模式

### 步驟 6: 實作 RedisCluster 連線

- 建立 `RedisCluster` 結構，實作 `IRedisConn` 介面
- 使用 Redis Cluster 模式 (`goredis.NewClusterClient`)
- 實作 `FillCluster()` 方法填充測試資料
- 處理叢集節點路由和 hash slot
- 參考 redis-cluster\docker-compose.yml
- 編輯 redis-cluster\docker-compose-ap-go.yml 新增啟動 compose
- 建立 `redis_cluster_test.go` 測試檔案
  - 介面實作驗證測試 (`TestRedisClusterImplementsInterface`)
  - 完整整合測試 (`TestRedisCluster`)，使用 `t.Skip()` 跳過
  - 參數驗證測試 (`TestNewRedisCluster_InvalidParams`)
- **測試注意事項**：
  - Cluster 模式需要初始化（`redis-cli --cluster create`）
  - 測試需確認 Cluster 狀態正常（`CLUSTER INFO`）
  - 注意 hash slot 分配和資料路由

### 步驟 7: 實作 RedisRaft 連線

- 建立 `RedisRaft` 結構，實作 `IRedisConn` 介面
- 使用 Redis Raft 模式連線
- 實作一致性讀寫（Strong Consistency）
- 處理 Leader 選舉和節點同步
- 參考 redis-raft\docker-compose.yml
- 編輯 redis-raft\docker-compose-ap-go.yml 新增啟動 compose
- 建立 `redis_raft_test.go` 測試檔案
  - 介面實作驗證測試 (`TestRedisRaftImplementsInterface`)
  - 完整整合測試 (`TestRedisRaft`)，使用 `t.Skip()` 跳過
  - 參數驗證測試 (`TestNewRedisRaft_InvalidParams`)
- **測試注意事項**：
  - Raft 模式需要 RedisRaft 模組支援
  - 測試需確認 Leader 選舉完成
  - 注意一致性保證和寫入延遲

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
