# Redis 架構說明

## 應用架構衍生背景

### Redis 單機瓶頸 

為了達到快速讀寫，開發者會將大量查詢頻繁的資料快取到 Redis， Session、熱門商品資訊、排行榜、通知、推播資料等。

隨著系統規模成長，開發者往往會把越來越多資料塞進 Redis，希望通通快取、全面加速。

### Redis 單機瓶頸 

![應用架構衍生背景 - visual selection (1)](.attach/.README/應用架構衍生背景 - visual selection (1).png)

### Redis 單機瓶頸 

1. 不具備資料持久性（Durability）

   [^A不具備資料持久性]: Redis 是以 RAM 為主要儲存媒介，天生為 volatile 設計，一旦機器異常重啟、當機或斷電，快取中的資料將立即遺失。雖然 Redis 提供 RDB（定時快照）與 AOF（操作記錄）機制以支援持久化，快照（RDB）存在備份週期落差，非實時，AOF 在高併發下可能影響效能與寫入延遲

   若在高流量交易環境中啟用持久化，反而削弱 Redis 作為快取層的高效能特性，造成系統效能瓶頸。因此，單機 Redis 無法被視為具備資料持久性的可靠儲存方案。

2. 單機 IO 壓力集中（I/O Bottleneck）

   [^B 單機 IO 壓力集中 ]: 在高併發微服務架構中，若多個應用服務共用同一個 Redis 節點，而 Redis 快取容量或效能未隨之擴張，將導致所有讀寫操作集中在單一執行緒上，形成 I/O 壓力瓶頸 

3. 不具備 HA (High Availability 高可用性）

   [^C不具備HA ]: 若單點故障，及服務中斷，沒有辦法自動切換到備援節點來提供不中斷服務 Auto Failover

4. 無法橫向擴展( Scalability )

   [^D無法橫向擴展]: Redis 為記憶體資料庫，機器 RAM 就是儲存上限，Redis 是單執行緒，CPU 核心吃滿後就無法再提升效能。想升級只能垂直擴充（加 RAM、CPU）。

## Redis 架構說明

參考：https://hackmd.io/@tienyulin/redis-master-slave-replication-sentinel-cluster

### 主從複製 

Master-Slave Replication 一主多從，解決 [^A不具備資料持久性] ，從 server 做到簡單備援。

主 server 單點寫入，從 server 讀取，做到讀寫分離，解決 [^B 單機 IO 壓力集中 ]。

#### 實作要點

#### Redis image 必須沒有 module

在容器環境中，Redis 不會掛載實體磁碟，因此同步過程必須依賴 `repl-diskless-sync` 以即時記憶體傳輸方式進行主從複製，避免寫入本地磁碟造成失敗。
若使用了 Redis module，同步時的 .rdb 中將包含模組的內部資料結構，slave 若未載入相同模組，將無法正確反序列化資料，導致複製失敗，會造成 .rdb 成功傳輸給 Slave，但 Slave 載入記憶體階段出錯，而不停重試同步 。
因此，image 必須使用無任何外掛 module 的純淨 Redis 映像檔，範例測試 `redis:6.2.19`，以確保複製過程的相容性與穩定性。

```powershell
#檢查 pull 的 redis image 有沒有 Module，必須是 (empty array)
redis-cli MODULE LIST
```

### 哨兵模式 

Sentinel 主從複製的基礎上，解決 [^A不具備資料持久性] 、 [^B 單機 IO 壓力集中 ]。

搭配 Sentinel「只做監控、不儲存資料」的執行緒，透過投票 vote quorum 主從切換，使服務自動切換備援，解決[^C不具備HA ]。

#### sentinel.conf

參考 https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/

#### SDOWN（主觀下線）

Subjectively Down， 當 sentinel1 監控到 master 下線，sentinel1  會標記 master SDOWN

#### ODOWN（客觀下線）

Objectively Down，若 sentinel1、sentinel2 都標記 master SDOWN， master 

#### epoch

是一個 64 bit unsigned int，初始化 Master / Slave 都是 0，在發送訊息時加 1，在收到訊息時如果對方的 epoch 高於自己，則更新 epoch 並加 1；
遇到有衝突，則選擇 epoch 較高的那一則訊息，Master 在發送 ping 時會夾帶 configEpoch 且在回 pong 時會夾帶所屬的 slot mapping

#### 實作要點

##### master 的連線與 sentinel 要分別實作

```c#
#取得 sentinel
var sentinelMux = ConnectionMultiplexer.Connect(sentinelConfig);
#透過 sentinel 取得 queryMaster 並不是真的 master
var queryMaster = sentinelMux.GetSentinelMasterConnection(redisServiceConfig);

 #這才是真的 master
 MasterEndpoint = queryMaster.GetEndPoints()
                    .Select(ep => queryMaster.GetServer(ep))
                    .FirstOrDefault(s => !s.IsReplica)?.EndPoint.ToString() ?? "";
 #這才是真的 slaves
 var slaves = queryMaster.GetEndPoints()
         .Select(ep => queryMaster.GetServer(ep))
         .Where(s => s.IsReplica)
         .Select(s => s.EndPoint.ToString());
```

##### Lib 實體要 AddScoped

AddScoped才可以在每一次請求都 renew redis connection，讓 failover 可以順利讀取，銜接接管新的 redis connection。

##### sentinel 與所有 nodes 的系統時間要一致

指令檢查是否一致，若不一致，會觸發 tilt mode exited

Redis Sentinel 的 tilt mode 是為了保護叢集，在系統時間出現「快跳 / 慢跳」的狀況時（如 NTP 同步時），它會暫停：1主節點故障偵測、2選主投票、3自動切換功能，一段時間內不進行判斷。

如果你的各個 sentinel/slave/master 時間落差太大，failover 判斷會錯誤，造成叢集不一致或「永遠在 wait」。

```powershell
# 檢查容器時間
"sentinel-master","sentinel-slave1","sentinel-slave2","sentinel1","sentinel2","sentinel3" | ForEach-Object { $_ + ": " + (docker exec $_ date) }
```



### 叢集模式

Cluster 多主多從，解決 [^A不具備資料持久性] 、 [^B 單機 IO 壓力集中 ]。

節點之間會自行運作 Gossip 協議，使服務自動切換備援，解決[^C不具備HA ]。

資料分片 (Sharding) 的技術，自動地將資料拆分到多台的伺服器上，以水平的擴充記憶體的容量，解決 [^D無法橫向擴展]。

#### 叢集建立流程

```mermaid
flowchart TB
    A(開始 6 node Redis Cluster 建立)
    B[分配 Hash Slots 到 Master 節點 16383 平分 3 個<br>Master1 slots ........0-..5460<br>Master2 slots ..5461-10922<br>Master3 slots 10923-16383]

    C[分配主從 加入 Replica node<br>node5:7005 複製於 node1:7001<br>node6:7006 複製於 node2:7002<br>node4:7004 複製於 node3:7003]
    
    D["檢查主從分配結果<br>3主(已分配 slots )<br>3從(未設定)"]
    E[node 都根據最終狀態自行寫入/更新自己的 nodes.conf 檔案]
    F[指派 Config Epoch 每個 node 獨立版本號]
    G[執行CLUSTER MEET messages 等待叢集形成並檢查狀態]
    G["檢查狀態<br>3主(已分配 slots )<br>3從(設定只能讀，0 slots)"]
    H[檢查slots coverage<br/>16384 slots 全部覆蓋]

    A --> B 
    B --> C
    C --> D
    
    D --> E
    E --> F
    F --> G
    G --> H

```



#### 資料分片 (Sharding) 

把整個 key space 拆成 16384 個 slot( HASH_SLOT )，自動分配到多個 master 節點

當資料變多、流量變大，新增節點以分攤壓力

#### Gossip 協議

Gossip protocol 讓節點彼此交換狀態資訊，實作分散式健康檢查與節點發現。

- 每個節點定期隨機挑幾個節點溝通，檢查其他節點是否存活
- 傳遞哪些節點掛了（FAIL）或恢復（OK）
- 節點間彼此同步彼此看到的狀態，分散式傳播某個節點的失效資訊（去中心化），不依賴單一 Master、Sentinel 
- bus-port 預設是 server 1+port，若 server 6379 那溝通的 bus-port 就是 16379 
- 每個 master 配一個或多個 slave，當 master 掛掉，slave 自動接管（Fail-Over）

#### 實作要點

##### Cluster Redis 必須是乾淨的

每一個 redis node 必須沒有任何資料，故 docker compose 重啟時，除了 container 要全部刪掉、Volumes 也要全部刪除

沒有刪除乾淨，就會在redis-cluster creator 看到這個錯誤

2025-07-15 17:37:29 [ERR] Node 192.168.1.91:7001 is not empty. Either the node already knows other nodes (check with CLUSTER NODES) or contains some key in database 0.

##### UI 可看到 Master 所有資料，但只有 Slot Owner 可以編輯自己的資料

範例 Update Cache  "xxxxxxx" node 1、node 2 node 3 

##### Lib 實體要 AddScoped

AddScoped才可以在每一次請求都 renew redis connection，讓 failover 可以順利讀取，銜接接管新的 redis connection。



#### Redis Cluster 的現況

1）節點自動發現
2）Slave → Master 的選舉機制( vote quorum )，具備集群容錯能力
3）熱線分片（Hot Resharding）：可在線進行資料重新分片
4）進群管理：支援 `CLUSTER xxx` 指令操作
5）基於設定檔（nodes-port.conf）的集群管理
6）ASK 轉向 / MOVED 轉向機制（請求重新導向）

### Redis 架構

#### Redis 同步連線架構

```mermaid
sequenceDiagram
    participant Slave
    participant Master
    
    rect rgb(26, 101, 188)
    Note over Slave,Master: 連線建立階段
    Slave->>Master: 建立 TCP 連接 (REDIS_REPL_CONNECT)
    Master-->>Slave: 發送 PING
    Slave-->>Master: 回應 PONG
    Slave->>Master: REPLCONF listening-port <port>
    Slave->>Master: REPLCONF capa eof capa psync2
    Slave->>Master: PSYNC <runid> <offset>
    end
    
    rect rgb(26, 101, 188)
    Note over Slave,Master: 同步階段  RDB/AOF
    alt if 初次增量或無法增量時<br>全量同步
        Master-->>Slave: +FULLRESYNC <runid> <offset>
        Master->>Master: BGSAVE 開始建立 RDB 快照
        Master->>Master: BGSAVE 完成
        Master-->>Slave: 傳送 RDB 快照 (REDIS_REPL_SEND_RDB)
        Slave->>Slave: 儲存並載入 RDB 至記憶體
        Master-->>Slave: 傳送積壓命令（增量資料）以完成同步
    else else: Slave 有合法 offset<br>Master 有 backlog<br>增量同步
        Master-->>Slave: +CONTINUE
        Master-->>Slave: 傳送 backlog 內的命令
    end
    end
        rect rgb(26, 101, 188)
    
    Note over Slave,Master: 持續同步階段
    Slave-->>Master: REPLCONF ACK <offset>
    Master-->>Slave: 持續寫入命令流（AOF 增量同步）
    end
```

參考: [Redis Replication 同步如何執行？](https://medium.com/jerrynotes/redis-replication-%E5%90%8C%E6%AD%A5%E5%A6%82%E4%BD%95%E5%9F%B7%E8%A1%8C-ef48675f6918)

#### Redis Sentinel 架構圖

```mermaid
sequenceDiagram
    participant Sentinel1
    participant Sentinel2
    participant Sentinel3
    participant Master
    participant Slave1
    participant Slave2

    loop 每秒 監控(Monitoring)
        Sentinel1->>Master: PING
        Sentinel2->>Master: PING
        Sentinel3->>Master: PING
    end

    alt Master 沒回應
        Sentinel1-->>Sentinel1: 標記 Master SDOWN
        Sentinel2-->>Sentinel2: 標記 Master SDOWN 
        Sentinel2-->>Sentinel1: 我覺得 Master SDOWN<br>Pub/Sub
        
        Sentinel3-->>Sentinel3: 標記 Master SDOWN
        Sentinel3-->>Sentinel1: 我也覺得 Master SDOWN<br>Pub/Sub
        
        Sentinel1-->>Sentinel1: 標記 Master ODOWN
    end

    alt 達到 quorum 門檻 Pub/Sub 投票
        Sentinel1->>Sentinel2: 我先標記 Master ODOWN<br>所以我發起投票<br>要不要投我做 Leader？
        Sentinel2->>Sentinel1: OK<br>Pub/Sub
        Sentinel3->>Sentinel1: OK<br>Pub/Sub
    end
    
    alt 投票完 主從重組命令
        Sentinel1-->>Slave1: 我是 Leader！
        Sentinel1-->>Slave2: 我是 Leader！
       
        Sentinel1->>Slave1: 你升級為新的 Master！
        Sentinel1->>Slave2: SLAVEOF Slave1
    end
```



1. Slave 主動與 Master 建立 TCP 連線，這一步稱作 `REDIS_REPL_CONNECT`。

   一旦連線建立，Master 會發送一個 `PING` 指令來確認連線的穩定性。

   Slave 收到後，會回應一個 `PONG`，表示連線正常。

2. Slave 會發送兩個重要的 `REPLCONF` 命令：

   REPLCONF listening-port <port>：告知 Master 本身的監聽埠號。

   REPLCONF capa eof capa psync2：宣告自己支援的同步功能，例如 EOF 結尾檢查，以及 PSYNC2 協議（這是 Redis 4.0 之後的標準）。

3. 同步啟動

   同步開始時，Slave 會發送 `PSYNC` 命令給 Master，格式如下：

   PSYNC <runid> <offset>

4. 如果是第一次同步，參數會是 `PSYNC ? -1`，代表沒有任何舊的同步紀錄。

   接著 Master 會根據狀況決定同步方式，有兩種可能, 全量同步（Full Resync）、增量同步（Partial Resync）

5. 全量同步（Full Resync）

   當 Slave 是第一次同步，或是 Master 的 backlog 無法滿足 offset 時，就會進入 **全量同步流程**：

   Master 回應：`+FULLRESYNC <runid> <offset>`，並紀錄同步資訊。

   Master 啟動 `BGSAVE`，在背景建立一份 RDB 快照。

   當 BGSAVE 完成後，Master 將 RDB 快照內容以 bulk string 傳送給 Slave。

   Slave 收到後，會將 RDB 檔案寫入本地，並載入到記憶體中，重建資料狀態。

6. 完成後，Master 開始發送 backlog 中的新增指令，也就是增量同步的開始。

7. 增量同步（Partial Resync）

   如果 Slave 有舊的 runid 與 offset，且 Master 的 backlog 足夠覆蓋，Master 會回應 `+CONTINUE`：

   表示不需要重新載入整個 RDB。

   Master 直接從 offset 處發送尚未同步的命令，實現「增量補同步」。

   命令流與持續同步

   當初始同步完成後，Slave 定期會發送：

   REPLCONF ACK <offset>

8. 告訴 Master 自己目前的同步進度。

9. Master 則會持續發送寫入命令（即 AOF 格式的增量命令流），確保主從資料一致。

整體來說，這個過程讓 Redis 能實現高效且彈性的資料複製，無論是做高可用架構，還是實現讀寫分離，都非常有幫助。

###  總結

1. 連線建立 → PING/PONG → REPLCONF
2. PSYNC 啟動同步
3. 視情況選擇全量或增量同步
4. 傳送資料後保持 offset 同步並持續傳送命令

這就是 Redis 主從複製的完整時序圖流程。

例如 RDB 與 AOF 混合使用策略，我們可以再深入討論。

架構細節:

1. 所有 redis 節點彼此互聯 (PING PONG) 機制, 內部使用二進制協議優化傳輸，提高溝通速度。

2. Master 主動發出 PING

3. 節點的 fail 是通過集群中超過半數的節點檢測失效時才生效

4. 客戶端與 redis 節點直連, 不需要中間層(沒有 main master、沒有 Sentinel)

   客戶端不需要連結集群所有節點, 連接集群中任何一個可用節點即可

   

(1)领着选举过程是集群中所有 master 参与,如果半数以上 master 节点与 master 节点通信超过(cluster-node-timeout),认为当前 master 节点挂掉.

 (2):什么时候整个集群不可用(cluster_state:fail),当集群不可用时,所有对集群的 操作做都不可用，收到((error) CLUSTERDOWN The cluster is down)错误 

a:如果集群任意 master 挂掉,且当前 master 没有 slave.集群进入 fail 状态,也 可以理解成进群的 slot 映射[0-16383]不完成时进入 fail 状态. 

b:如果进群超过半数以上 master 挂掉，无论是否有 slave 集群进入 fail 状态.

二.Redis 集群安装篇(centos5.8 X64 系统) 

（要让集群正常工作至少需要 3 个主节点，在这里我们要创建 6 个 redis 节点， 其中三个为主节点，三个为从节点，对应的 redis 节点的ip 和端口对应关系如下）

127.0.0.1:7000

127.0.0.1:7001

127.0.0.1:7002

127.0.0.1:7003

127.0.0.1:7004

127.0.0.1:7005

1：下载 redis。 官网下载 3.0.0 版本，之前 2.几的版本不支持集群模式 下载地址：http://download.redis.io/releases/redis-3.0.2.tar.gz

2：上传服务器，解压，编译 

```
tar -zxvf redis-3.0.2.tar.gz.tar.gz
mv redis-3.0.2.tar.gz.tar.gz redis3.0
cd /usr/local/redis3.0
make
make install
```

3：创建集群需要的目录 

```
mkdir -p /usr/local/cluster
cd /usr/local/cluster
mkdir 7000
mkdir 7001
mkdir 7002
mkdir 7003
mkdir 7004
mkdir 7005
```

4：修改配置文件 redis.conf

### Redis Cluster 執行步驟

#### 1. Modify Redis Configuration File

```
cp /usr/local/redis3.0/redis.conf /usr.local/cluster
vi redis.conf
```

Modify the following options in `redis.conf`:

```
port 7000
daemonize yes
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
appendonly yes
```

After modifying, copy the configuration file to each Redis instance directory:

```
cp /usr/local/cluster/redis.conf /usr/local/cluster/7000
cp /usr/local/cluster/redis.conf /usr/local/cluster/7001
cp /usr/local/cluster/redis.conf /usr/local/cluster/7002
cp /usr/local/cluster/redis.conf /usr/local/cluster/7003
cp /usr/local/cluster/redis.conf /usr/local/cluster/7004
cp /usr/local/cluster/redis.conf /usr/local/cluster/7005
```

Modify the `port` value in each directory-specific `redis.conf` file accordingly.

#### 2. Start Redis Instances

```
cd /usr/local/cluster/7000 && redis-server redis.conf
cd /usr/local/cluster/7001 && redis-server redis.conf
cd /usr/local/cluster/7002 && redis-server redis.conf
cd /usr/local/cluster/7003 && redis-server redis.conf
cd /usr/local/cluster/7004 && redis-server redis.conf
cd /usr/local/cluster/7005 && redis-server redis.conf
```

Verify Redis instances are running:

```
ps -ef | grep redis
```

Example output:

```
root 13703 1 0 10:03 ? 00:00:00 redis-server *:7000 [cluster]
root 14015 1 0 10:04 ? 00:00:00 redis-server *:7001 [cluster]
root 14133 1 0 10:04 ? 00:00:00 redis-server *:7002 [cluster]
root 14172 1 0 10:04 ? 00:00:00 redis-server *:7003 [cluster]
root 14187 1 0 10:04 ? 00:00:00 redis-server *:7004 [cluster]
root 14323 1 0 10:04 ? 00:00:00 redis-server *:7005 [cluster]
```

#### 3. Install  Redis Cluster Management( Ruby and Gems Redis 5.0 以前)

Redis 5.0 以後已內建 cluster 管理指令，無需額外安裝 Ruby 工具。

Redis Cluster management requires Ruby version 1.8.7 or above. Install Ruby and necessary dependencies:

```
rpm -ivh http://yum.puppetlabs.com/el/5/products/x86_64/puppetlabs-release-5-6.noarch.rpm
yum install ruby ruby-devel rubygems rpm-build
```

Check the installed Ruby version:

```
ruby -v
```

Example output:

```
ruby 1.8.7 (2013-06-27 patchlevel 374) [x86_64-linux]
```

Check if `rubygems` is installed:

```
rpm -qa | grep ruby
```

Example output:

```
ruby-rdoc-1.8.7.374-2.el5
ruby-1.8.7.374-2.el5
ruby-devel-1.8.7.374-2.el5
ruby-irb-1.8.7.374-2.el5
ruby-libs-1.8.7.374-2.el5
rubygems-1.3.7-1.el5
```

#### 4. Install Redis Ruby Interface

```
gem install redis
```

#### 5. Create Redis Cluster

Navigate to Redis source directory and run:

```
cd /usr/local/redis3.0/src
./redis-trib.rb create --replicas 1 127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005
```

#### 6. Troubleshooting Notes

##### Error 1: Ruby Not Found

```
/usr/bin/env: ruby: No such file or directory
```

Solution:

```
yum install ruby
```

##### Error 2: Missing rubygems

```
./redis-trib.rb:24:in `require': no such file to load -- rubygems (LoadError)
```

Solution:

```
yum install rubygems
```

##### Error 3: Missing Redis Ruby Interface

```
/usr/lib/ruby/site_ruby/1.8/rubygems/custom_require.rb:31:in `gem_original_require': no such file to load -- redis (LoadError)
```

Solution:

```
gem install redis
```

Once all dependencies are installed, retry the Redis cluster creation command from Step 5.



### redis vote quorum

夠用的，最終一致性，簡化投票邏輯

### redis-raft

強一致性架構、高可靠主從轉移

| 項目          | Redis Master-Slave                      | Redis Sentinel                                               | Redis Cluster                                       | Redis Raft                                                   |
| ------------- | --------------------------------------- | ------------------------------------------------------------ | --------------------------------------------------- | ------------------------------------------------------------ |
| 架構          | 1 * Master<br />n * Slaves              | m * Sentinels m > 3, odd<br />1 * Master<br />n * Slaves     | m * Masters m > 3, odd<br />m * n * Slaves          | 1 * Leader<br />n * Followers >2<br />1+n > 3, odd           |
| 核心機制      | 手動設定主從<br />Slave 被動接收 Master | vote quorum<br />Sentinel ping/pong Master                   | Gossip 協議<br />全員 ping/pong                     | Raft 協議<br />週期心跳、Leader 選舉、Log 複寫               |
| 資料一致性    | NA                                      | 最終一致性(弱AP)                                             | 最終一致性 (AP)                                     | ✅強一致性 (CP)                                               |
| 水平擴展      | 不支援                                  | 不支援                                                       | ✅ Slot 分片橫向擴展                                 | 不支援                                                       |
| 寫入限制      | 寫 Master                               | 寫 Master                                                    | 任一 Master 可寫                                    | 寫 Leader                                                    |
| 讀取行為      | 指定 Slave 讀                           | 經 Sentinels 查 Slave 讀                                     | 經 slot redirection查 Slave 讀                      | 讀 Leader                                                    |
| 自動 Failover | NA                                      | ✅自動 failover                                               | ✅自動 failover                                      | ✅內建選主                                                    |
| 主從關係      | 一主多從                                | 一主多從                                                     | 多主多從分片組                                      | 所有節點皆等價 leader 選舉                                   |
| 成熟度        | NA                                      | 高（穩定多年）                                               | 高（官方核心）                                      | 中（Redis Labs 模組）                                        |
| 資料分佈      | 主存資料、從同步                        | 單主架構，從節點為備援                                       | 資料依 slot 分片                                    | 所有節點存同一份資料                                         |
| 適用場景      | 備援                                    | 備援+夠用的 failover                                         | 巨量資料（百 GB ～ TB）<br />分布式快取、高容量負載 | 強一致設定中心、交易系統、帳務紀錄                           |
| 場景例子      |                                         | 中小型 B2C / SaaS 服務平台<br />遊戲伺服器後台<br />微服務架構中的 session 存儲/簡易 KV | 大型電商網站商品快取<br />名人的 tweet              | 交易撮合系統中的撮合狀態同步<br />IoT 即時控制系統:<br />工廠設備的溫度、壓力、運作狀態 |

1. Redis 讀取

   單點讀取(只讀一台 redis) + 回源模式( 回 db Read-through )

   設計上不論 Redis 有幾台，client  都只會挑一台讀取，若是 cache miss 走 Read-through，不會 retry 其他 redis

2. Redis Raft 架構

   Raft 最少需要 3 個節點（1 Leader+ 2 Followers）才能容忍 1 個 failover（即保證 quorum 過半投票）

   1+n > 3, odd, 奇數以避免選 Leader tie

3. Redis Cluster 架構

   m * Masters m > 3, odd 最少 3 個 Masters 節點形成 cluster，奇數以避免選 Masters tie

   m * n * Slaves，每個 Master 至少配置一個 Slave 容忍 1 個 failover

4. Redis Sentinel 架構

   至少 3 個 sentinel 才能投票出結果

5. Redis Sentinel 核心機制

   全員 Sentinel ping/pong 一個 Master 檢查存活

   Pub/Sub 則是讓 Sentinel 間需要知道彼此存在 + 狀態資訊

6. Redis Sentinel 弱AP: 

   雖然在形式上具備分區容忍（P）與可用性（A），但實際上可用性不穩定（弱 A）。
   切換 Master 需透過 Sentinel 投票，過程可能花費數秒，client 在此期間連線舊 Master 可能出現寫入失敗或 timeout。
   切換後若未強制關閉原 Master，則可能產生 split-brain，導致 client 寫入結果不確定（有時成功、有時 timeout、有時寫錯資料）。

7. Redis Sentinel 讀取

   client 透過 Sentinel 查詢 Slave 清單，由 client 決定讀哪台 Slave 。

8. Redis Cluster 讀取

   client  透過 slot 對應與 redirection 找到對應節點

   若啟用 `readonly` 模式，client 可自動讀該 slot 的 Slave 節點。





## 補充

### 架構演進

![Redis架构演进](.attach/.README/redis-architecture-evolve-17525729942973.gif)

來源: https://tech.lucumt.info/docs/db/redis/redis-architecture-evolve/

| 版本          | 發布時間 | 重大更新                                                     | 說明                                                         |
| ------------- | -------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| **Redis 5.0** | 2017     | • stream data type<br />• ZPOP/BPOP 新指令                   | 新增 Kafka 式訊息串流支援，改善內存碎片問題，強化排序集 pop API |
| **Redis 6.0** | 2020     | • ACL（使用者層級權限）<br />• TLS 加密支援<br />• I/O 多執行緒 | 把接收資料這一段變成多執行緒，讓它更快；但資料處理本身仍維持單執行緒。 |
| **Redis 7.0** | 2022     | • ACL v2（選擇性權限控制）<br />• Sharded Pub/Sub            | 資安升級、多執行緒優化                                       |
| **Redis 8.0** | 2025     | • 整合模組：JSON、TimeSeries 等<br />• 效能提升：命令快 +87%、倍增吞吐量、複製快 +18%<br />• 記憶體最佳化＋licensing 改為 AGPL | 全模組化列入核心，支持 AI use case，整體效能與記憶體使用大幅改善 |

### Redis 其他架構

- **Twemproxy**：Twitter 開源 Redis 分片代理，部署簡單，但不支援多 key、pipeline、Pub/Sub。適合靜態分片。
- **Codis**：支援 slot 遷移、Web 管理介面，比 Twemproxy 功能完整。適合中型分片場景。
- **Redis on Kubernetes**：以 StatefulSet + PVC 部署 Redis，可搭配 Sentinel/Cluster/Operator 實現高可用與自動化。
- **Redis Cloud / ElastiCache**：雲端託管 Redis，支援備援、自動 failover 與擴展。適合不想自行維運的團隊。

### CAP 理論

分散式系統設計的基礎原則指出：CAP 三者無法兼得。當網路分區發生（P 成立）時，系統無法同時滿足一致性（C）與可用性（A），因此在分散式架構中，必須在 C 與 A 之間擇一放棄。

Partition Tolerance (P): 分區容忍，即使部分節點間網路中斷，整個系統仍能繼續服務

Availability (A): 可用性，每個非故障的節點，都能及時回應請求（不會掛住、不回應）

Consistency (C): 一致性，所有節點對同一資料讀到的值是一致的（如同單機系統）

| 系統類型      | 分類 | 說明                                         |
| ------------- | ---- | -------------------------------------------- |
| Redis Raft    | CP   | 一致性為先，必要時拒絕寫入（放棄可用性）     |
| Redis Cluster | AP   | 分區時仍允許寫入，但不同節點資料可能不一致   |
| RDBMS         | CA   | 不容忍分區，集中處理寫入與查詢，傳統單體架構 |
