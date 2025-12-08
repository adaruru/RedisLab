package redislib

// RedisMode 定義 Redis 部署模式
type RedisMode int

const (
	// RedisMasterSlaves 主從模式
	RedisMasterSlaves RedisMode = iota
	// RedisSentinel 哨兵模式
	RedisSentinel
	// RedisCluster 叢集模式
	RedisCluster
	// RedisRaft Raft 共識模式
	RedisRaft
)

// String 返回 RedisMode 的字串表示
func (m RedisMode) String() string {
	switch m {
	case RedisMasterSlaves:
		return "RedisMasterSlaves"
	case RedisSentinel:
		return "RedisSentinel"
	case RedisCluster:
		return "RedisCluster"
	case RedisRaft:
		return "RedisRaft"
	default:
		return "Unknown"
	}
}

// ParseRedisMode 從字串解析 RedisMode
func ParseRedisMode(s string) (RedisMode, error) {
	switch s {
	case "RedisMasterSlaves":
		return RedisMasterSlaves, nil
	case "RedisSentinel":
		return RedisSentinel, nil
	case "RedisCluster":
		return RedisCluster, nil
	case "RedisRaft":
		return RedisRaft, nil
	default:
		return -1, ErrInvalidRedisMode
	}
}
