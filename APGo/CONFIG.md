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

### 為什麼不使用 APGO_REDIS_MODE 覆蓋？

❌ **不實用的做法**（理論上正確，實際上不需要）：
```yaml
# config.docker.yaml
environment:
  - GO_ENV=docker
  - APGO_REDIS_MODE=RedisMasterSlaves  # 需要看兩個地方
```

✅ **實用的做法**（專案目的明確）：
```yaml
# docker-compose-ap-go.yml
environment:
  - GO_ENV=master-slave  # 一看就懂
```

理由：
1. 這是 POC 專案，不是生產環境
2. 每個 docker-compose 只測試一種架構
3. Redis 架構不會在運行時切換
4. 識別度比靈活性更重要

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

## 總結

這個設計選擇是基於：

1. **專案性質**：這是測試 Redis 架構的 POC，不是通用應用
2. **實用主義**：識別度比理論上的靈活性更重要
3. **維護性**：每個環境有完整的配置，易於理解和修改
4. **符合目的**：環境名稱直接反映測試目標

在生產環境的真實專案中，可能會採用更「正統」的設計（環境與配置分離），但對於這個 POC 專案，當前設計是最合適的。
