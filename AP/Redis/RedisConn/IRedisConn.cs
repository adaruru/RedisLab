using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public interface IRedisConn
{
    public string MasterEndpoint { get; set; }
    public string SlaveEndpoint { get; set; }
    Task<string?> ReadAsync(string key);
    Task<string?> GetRamdonCache(string key);
    Task<bool> WriteAsync(string key, string value);
}