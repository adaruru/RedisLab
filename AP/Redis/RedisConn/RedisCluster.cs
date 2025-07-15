using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public class RedisCluster : IRedisConn
{
    private readonly ConnectionMultiplexer _cluster;
    public string MasterEndpoint { get; set; }
    public string SlaveEndpoint { get; set; }
    public RedisCluster(IConfiguration config)
    {
        var nodes = config.GetSection("Redis:RedisCluster:Nodes").Get<string[]>() ?? [];
        _cluster = ConnectionMultiplexer.Connect(string.Join(",", nodes));
    }

    public async Task<string?> GetCache(string key)
    {
        var db = _cluster.GetDatabase();
        return await db.StringGetAsync(key);
    }

    public Task<string?> GetRamdonCache(string key)
    {
        throw new NotImplementedException();
    }

    public async Task<bool> UpdateCache(string key, string value)
    {
        var db = _cluster.GetDatabase();
        return await db.StringSetAsync(key, value);
    }
    public bool FillCluster()
    {
        return true;
    }
}
