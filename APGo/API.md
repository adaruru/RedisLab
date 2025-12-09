# APGo API 使用文件

## 概述

APGo 提供 RESTful API 端點來操作 Redis 快取，支援多種 Redis 部署模式。

## 基礎 URL

```
http://localhost:8080
```

## 端點列表

### 1. 健康檢查

檢查 API 服務與 Redis 連線狀態。

**端點**: `GET /health`

**回應範例**:
```json
{
  "status": "healthy",
  "service": "APGo Redis API",
  "version": "1.0.0",
  "redis_mode": "connected",
  "master_endpoint": "127.0.0.1:6379",
  "slave_endpoint": "127.0.0.1:6380"
}
```

---

### 2. 讀取快取

從 Redis 讀取指定 key 的值（從 Slave/Replica 讀取）。

**端點**: `GET /cache`

**Query 參數**:
- `key` (必填): 快取鍵名

**請求範例**:
```bash
curl "http://localhost:8080/cache?key=user:123"
```

**成功回應** (200 OK):
```json
{
  "key": "user:123",
  "value": "John Doe",
  "message": "value: John Doe",
  "read_from": "127.0.0.1:6380"
}
```

**失敗回應** (404 Not Found):
```json
{
  "error": "key not found",
  "key": "user:999",
  "message": "key 'user:999' not found",
  "read_from": "127.0.0.1:6380"
}
```

**錯誤回應** (400 Bad Request):
```json
{
  "error": "key is required",
  "message": "請提供 key 參數"
}
```

---

### 3. 更新快取

寫入資料到 Redis（寫入 Master）。

**端點**: `POST /cache`

**Content-Type**: `application/json`

**Request Body**:
```json
{
  "key": "user:123",
  "value": "John Doe"
}
```

**請求範例**:
```bash
curl -X POST http://localhost:8080/cache \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:123",
    "value": "John Doe"
  }'
```

**成功回應** (200 OK):
```json
{
  "key": "user:123",
  "value": "John Doe",
  "message": "key 'user:123', value 'John Doe' well saved",
  "written_to": "127.0.0.1:6379"
}
```

**失敗回應** (400 Bad Request):
```json
{
  "error": "invalid request",
  "message": "Key: 'CacheRequest.Key' Error:Field validation for 'Key' failed on the 'required' tag"
}
```

**錯誤回應** (500 Internal Server Error):
```json
{
  "error": "write failed",
  "key": "user:123",
  "value": "John Doe",
  "message": "connection refused",
  "written_to": "127.0.0.1:6379"
}
```

---

### 4. 填充 Cluster 測試資料

批次填充測試資料到 Redis Cluster（僅 Cluster 模式支援）。

**端點**: `GET /fillcluster`

**說明**: 自動填充 100 筆測試資料，用於測試 hash slot 分配。

**請求範例**:
```bash
curl http://localhost:8080/fillcluster
```

**成功回應** (200 OK) - Cluster 模式:
```json
{
  "message": "Successfully filled 100 test records to cluster",
  "count": 100,
  "mode": "RedisCluster"
}
```

**失敗回應** (400 Bad Request) - 非 Cluster 模式:
```json
{
  "error": "unsupported mode",
  "message": "FillCluster only supports RedisCluster mode",
  "mode": "*redis.RedisMasterSlave"
}
```

---

## 使用範例

### 完整工作流程

```bash
# 1. 檢查服務健康狀態
curl http://localhost:8080/health

# 2. 寫入資料
curl -X POST http://localhost:8080/cache \
  -H "Content-Type: application/json" \
  -d '{"key":"product:001","value":"Laptop"}'

# 3. 讀取資料
curl "http://localhost:8080/cache?key=product:001"

# 4. 填充測試資料（僅 Cluster 模式）
curl http://localhost:8080/fillcluster
```

### PowerShell 範例

```powershell
# 寫入
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/cache" `
  -ContentType "application/json" `
  -Body '{"key":"test","value":"hello"}'

# 讀取
Invoke-RestMethod -Uri "http://localhost:8080/cache?key=test"
```

---

## 錯誤碼說明

| HTTP 狀態碼 | 說明 |
|------------|------|
| 200 | 請求成功 |
| 400 | 請求參數錯誤或不支援的操作 |
| 404 | 找不到指定的 key |
| 500 | 伺服器內部錯誤或 Redis 操作失敗 |

---

## 讀寫分離說明

APGo 實作了 Redis 的讀寫分離機制：

- **讀取操作** (`GET /cache`): 從 Slave/Replica 節點讀取，減輕 Master 負擔
- **寫入操作** (`POST /cache`): 寫入到 Master 節點，確保資料一致性

每個回應都包含 `read_from` 或 `written_to` 欄位，顯示實際操作的 Redis 端點位址。

---

## 不同 Redis 模式的行為

### Master-Slave 模式
- 讀取：從指定的 Slave 節點
- 寫入：到 Master 節點
- FillCluster：不支援

### Sentinel 模式
- 讀取：從 Sentinel 管理的 Slave 節點
- 寫入：到 Sentinel 管理的 Master 節點（自動故障轉移）
- FillCluster：不支援

### Cluster 模式
- 讀取：根據 hash slot 自動路由到對應的 Slave
- 寫入：根據 hash slot 自動路由到對應的 Master
- FillCluster：**支援**，填充 100 筆資料測試 slot 分配

### Raft 模式
- 讀取：從 Raft cluster 讀取（強一致性）
- 寫入：透過 Raft 共識寫入（強一致性）
- FillCluster：不支援

---

## 測試建議

1. **開發環境測試**:
   ```bash
   # 啟動 Master-Slave
   cd redis-master-slave
   docker-compose -f docker-compose-ap-go.yml up -d
   
   # 執行 APGo
   cd ../APGo
   GO_ENV=master-slave go run cmd/main.go
   ```

2. **Cluster 測試**:
   ```bash
   # 啟動 Cluster
   cd redis-cluster
   docker-compose -f docker-compose-ap-go.yml up -d
   
   # 執行 APGo
   cd ../APGo
   GO_ENV=cluster go run cmd/main.go
   
   # 測試填充功能
   curl http://localhost:8080/fillcluster
   ```

3. **壓力測試**:
   使用 Apache Bench 或 k6 進行壓力測試：
   ```bash
   # 安裝 Apache Bench
   # 測試讀取效能
   ab -n 1000 -c 10 "http://localhost:8080/cache?key=test"
   
   # 測試寫入效能
   ab -n 1000 -c 10 -p data.json -T application/json http://localhost:8080/cache
   ```

---

## 常見問題

### Q: 為什麼讀取和寫入的端點位址不同？

A: APGo 實作了讀寫分離，讀取從 Slave 節點，寫入到 Master 節點，這是 Redis 高可用架構的標準做法。

### Q: FillCluster 為什麼只支援 Cluster 模式？

A: FillCluster 用於測試 hash slot 分配，這是 Cluster 模式特有的功能。其他模式使用主從複製，不需要此功能。

### Q: 如何切換不同的 Redis 模式？

A: 設定環境變數 `GO_ENV`：
- `master-slave`: 主從模式
- `sentinel`: 哨兵模式
- `cluster`: 叢集模式
- `raft`: Raft 共識模式

---

## 授權

MIT License
