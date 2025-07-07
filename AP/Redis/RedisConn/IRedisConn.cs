using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public interface IRedisConn
{
    Task<string?> GetCache(string key);
    Task<string?> GetRamdonCache(string key);
    Task<bool> UpdateCache(string key, string value);
}