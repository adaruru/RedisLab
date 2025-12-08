// Package redislib 提供 Redis 連線的統一介面
//
// 此套件定義了 IRedisConn 介面，支援多種 Redis 部署模式：
//   - RedisMasterSlaves: 主從模式，讀寫分離
//   - RedisSentinel: 哨兵模式，自動故障轉移
//   - RedisCluster: 叢集模式，分散式儲存
//   - RedisRaft: Raft 共識模式，強一致性
//
// 使用範例：
//
//	var redis redislib.IRedisConn
//	// 根據模式建立對應的實作
//	redis = NewRedisMasterSlave(config)
//
//	// 寫入資料
//	ok, err := redis.WriteAsync(ctx, "key", "value")
//
//	// 讀取資料
//	value, err := redis.ReadAsync(ctx, "key")
package redislib
