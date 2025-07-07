
namespace AP.Redis;
public enum RedisMode
{
    RedisMasterSlaves = 0,
    RedisSentinel = 1,
    RedisCluster = 2,
    RedisRaft = 3
}