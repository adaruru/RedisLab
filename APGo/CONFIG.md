# APGo 設定說明

## 設計理念：POC 專案的實用主義

這是一個**測試不同 Redis 架構的 POC 專案**，因此採用最直觀的設計：

**環境 = Redis 模式**

為什麼這樣設計？
1. ✅ **識別度高**：一看環境名就知道在測試哪種 Redis 架構
2. ✅ **配置清晰**：每個架構有自己完整的配置檔
3. ✅ **符合專案目的**：這個專案就是要測試不同 Redis 架構
4. ✅ **實用主義**：在實際生產環境中，Redis 架構一旦選定就不會改變

## 設定檔結構

### 開發環境（本地連線）
- [config.yaml](config.yaml) - 預設配置，連到遠端 Redis（192.168.1.91）

### Docker 測試環境（各 Redis 架構）
- [config.master-slave.yaml](config.master-slave.yaml) - 主從模式
- [config.sentinel.yaml](config.sentinel.yaml) - 哨兵模式
- [config.cluster.yaml](config.cluster.yaml) - 叢集模式
- [config.raft.yaml](config.raft.yaml) - Raft 模式

## 環境變數

每個 docker-compose 只測試一種架構，Redis 架構不會在運行時切換，識別度比靈活性更重要。

### GO_ENV（環境識別）

直接對應 Redis 架構模式：

```bash
# 主從模式
GO_ENV=master-slave

# 哨兵模式
GO_ENV=sentinel

# 叢集模式
GO_ENV=cluster

# Raft 模式
GO_ENV=raft
```

## Docker Compose 使用範例

### 主從模式測試
```yaml
# redis-master-slave/docker-compose-ap-go.yml
services:
  redis-master-slave-apgo:
    environment:
      - GO_ENV=master-slave  # 載入 config.master-slave.yaml
```

### Sentinel 模式測試
```yaml
# redis-sentinel/docker-compose-ap-go.yml
services:
  redis-sentinel-apgo:
    environment:
      - GO_ENV=sentinel  # 載入 config.sentinel.yaml
```

### Cluster 模式測試
```yaml
# redis-cluster/docker-compose-ap-go.yml
services:
  redis-cluster-apgo:
    environment:
      - GO_ENV=cluster  # 載入 config.cluster.yaml
```

### Raft 模式測試
```yaml
# redis-raft/docker-compose-ap-go.yml
services:
  redis-raft-apgo:
    environment:
      - GO_ENV=raft  # 載入 config.raft.yaml
```

## 配置檔範例

### config.yaml（開發環境）
```yaml
server:
  port: 8080
  mode: debug

redis:
  mode: RedisMasterSlaves

  master_slave:
    master: "192.168.1.91:6379"
    slaves:
      - "192.168.1.91:6380"
      - "192.168.1.91:6381"
```

### config.master-slave.yaml（Docker 主從）
```yaml
server:
  port: 8080
  mode: release

redis:
  mode: RedisMasterSlaves

  master_slave:
    master: "redis-master:6379"
    slaves:
      - "redis-slave1:6379"
      - "redis-slave2:6379"
```

## 設定載入邏輯

```
1. 讀取 config.yaml（基礎配置）
2. 如果有 GO_ENV，合併 config.{GO_ENV}.yaml
3. 環境變數覆蓋（如 APGO_SERVER_PORT）
```

範例：
```bash
GO_ENV=master-slave

載入順序：
1. config.yaml
2. config.master-slave.yaml（覆蓋）
3. 環境變數（最終覆蓋）
```

## 本地開發

```bash
# 使用預設配置（連到 192.168.1.91）
go run ./cmd/main.go

# 測試特定配置
GO_ENV=master-slave go run ./cmd/main.go
```

## 與 .NET Core 的對照

| .NET Core | Go（本專案） |
|-----------|--------------|
| `ASPNETCORE_ENVIRONMENT=RedisMasterSlaves` | `GO_ENV=master-slave` |
| `appsettings.RedisMasterSlaves.json` | `config.master-slave.yaml` |
| `appsettings.RedisSentinel.json` | `config.sentinel.yaml` |
| `appsettings.RedisCluster.json` | `config.cluster.yaml` |
| `appsettings.RedisRaft.json` | `config.raft.yaml` |

差異：
- ✅ 檔名更簡潔（kebab-case）
- ✅ 使用 YAML 而非 JSON
- ✅ 使用 snake_case 鍵名

## 命名慣例

### 環境名稱
使用 kebab-case（符合 Docker/K8s 慣例）：
- `master-slave` ✅（不是 MasterSlave 或 master_slave）
- `sentinel` ✅
- `cluster` ✅
- `raft` ✅

### 配置鍵名
使用 snake_case（符合 YAML/Go 慣例）：
- `master_slave` ✅（不是 masterSlave 或 MasterSlave）
- `master_name` ✅
- `sentinels` ✅

## 整合測試

### 測試策略

專案採用兩種測試方式：

1. **單元測試**：不需要 Redis 環境，測試參數驗證和介面實作
2. **整合測試**：需要實際 Redis 環境，測試完整的讀寫功能

### 單元測試（快速驗證）

```bash
cd APGo

# 執行所有測試
go test ./... -v

# 只測試 Redis 相關
go test ./internal/redis/... -v
```

**結果**：
- ✅ 參數驗證測試會執行
- ⏭️ 整合測試會被跳過（`t.Skip()`）

### 整合測試步驟

#### 1. Master-Slave 模式整合測試

```bash
# Step 1: 啟動 Master-Slave 環境
cd redis-master-slave
docker-compose -f docker-compose-ap-go.yml up -d

# Step 2: 等待服務完全啟動
sleep 15

# Step 3: 檢查服務狀態
docker-compose -f docker-compose-ap-go.yml ps

# Step 4: 驗證 Redis 連線
docker exec redis-master redis-cli ping
docker exec redis-slave1 redis-cli ping

# Step 5: 修改測試檔案（暫時移除 Skip）
# 編輯 APGo/internal/redis/redis_master_slave_test.go
# 在測試中註解掉 t.Skip() 行（如果有）

# Step 6: 執行整合測試
cd ../APGo
go test ./internal/redis/... -v -run TestRedisMasterSlave

# Step 7: 測試完成後還原
# 將 t.Skip() 註解還原

# Step 8: 停止環境
cd ../redis-master-slave
docker-compose -f docker-compose-ap-go.yml down
```

#### 2. Sentinel 模式整合測試

```bash
# Step 1: 啟動 Sentinel 環境
cd redis-sentinel
docker-compose -f docker-compose-ap-go.yml up -d

# Step 2: 等待服務完全啟動
sleep 20

# Step 3: 檢查 Sentinel 狀態
docker exec sentinel1 redis-cli -p 26379 SENTINEL masters
docker exec sentinel1 redis-cli -p 26379 SENTINEL slaves mymaster

# Step 4: 驗證 Master 連線
docker exec sentinel-master redis-cli ping

# Step 5: 修改測試檔案
# 編輯 APGo/internal/redis/redis_sentinel_test.go
# 註解掉第 11 行：
#   // t.Skip("需要實際的 Sentinel 環境才能執行")

# Step 6: 執行整合測試
cd ../APGo
go test ./internal/redis/... -v -run TestRedisSentinel

# Step 7: 測試完成後還原
# 將第 11 行的註解還原：
#   t.Skip("需要實際的 Sentinel 環境才能執行")

# Step 8: 停止環境
cd ../redis-sentinel
docker-compose -f docker-compose-ap-go.yml down
```

#### 3. Cluster 模式整合測試（待實作）

```bash
# Step 1: 啟動 Cluster 環境
cd redis-cluster
docker-compose -f docker-compose-ap-go.yml up -d

# Step 2: 等待並初始化 Cluster
sleep 15
# ... 執行 cluster create 命令

# Step 3: 執行測試
cd ../APGo
go test ./internal/redis/... -v -run TestRedisCluster

# Step 4: 停止環境
cd ../redis-cluster
docker-compose -f docker-compose-ap-go.yml down
```

#### 4. Raft 模式整合測試（待實作）

```bash
# Step 1: 啟動 Raft 環境
cd redis-raft
docker-compose -f docker-compose-ap-go.yml up -d

# Step 2: 執行測試
cd ../APGo
go test ./internal/redis/... -v -run TestRedisRaft

# Step 3: 停止環境
cd ../redis-raft
docker-compose -f docker-compose-ap-go.yml down
```

### 測試腳本自動化（建議）

可以建立測試腳本來自動化整合測試流程：

**APGo/test-integration.sh**:
```bash
#!/bin/bash

MODE=$1  # master-slave, sentinel, cluster, raft

if [ -z "$MODE" ]; then
  echo "Usage: ./test-integration.sh <mode>"
  echo "  mode: master-slave, sentinel, cluster, raft"
  exit 1
fi

echo "Starting integration test for $MODE mode..."

# 啟動環境
cd ../redis-$MODE
docker-compose -f docker-compose-ap-go.yml up -d

# 等待服務啟動
echo "Waiting for services to start..."
sleep 20

# 執行測試
cd ../APGo
echo "Running tests..."

case $MODE in
  master-slave)
    go test ./internal/redis/... -v -run TestRedisMasterSlave
    ;;
  sentinel)
    go test ./internal/redis/... -v -run TestRedisSentinel
    ;;
  cluster)
    go test ./internal/redis/... -v -run TestRedisCluster
    ;;
  raft)
    go test ./internal/redis/... -v -run TestRedisRaft
    ;;
esac

# 停止環境
echo "Cleaning up..."
cd ../redis-$MODE
docker-compose -f docker-compose-ap-go.yml down

echo "Integration test completed!"
```

使用方式：
```bash
cd APGo
chmod +x test-integration.sh

# 測試 Sentinel 模式
./test-integration.sh sentinel
```

### 測試檔案說明

測試檔案中使用 `t.Skip()` 的原因：

1. **開發階段**：不需要每次都啟動 Docker 環境
2. **CI/CD 整合**：單元測試可以快速執行，整合測試獨立執行
3. **明確意圖**：清楚標示哪些測試需要外部依賴

當需要執行整合測試時，手動註解掉 `t.Skip()` 即可。

## 總結

這個設計選擇是基於：

1. **專案性質**：這是測試 Redis 架構的 POC，不是通用應用
2. **實用主義**：識別度比理論上的靈活性更重要
3. **維護性**：每個環境有完整的配置，易於理解和修改
4. **符合目的**：環境名稱直接反映測試目標

在生產環境的真實專案中，可能會採用更「正統」的設計（環境與配置分離），但對於這個 POC 專案，當前設計是最合適的。
