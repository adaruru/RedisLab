using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public class RedisRaft : IRedisConn
{
    private readonly ConnectionMultiplexer _raft;
    public string MasterEndpoint { get; set; }
    public string SlaveEndpoint { get; set; }
    public RedisRaft(IConfiguration config)
    {
        var nodes = config.GetSection("Redis:RedisRaft:Nodes").Get<string[]>() ?? [];
        _raft = ConnectionMultiplexer.Connect(string.Join(",", nodes));
    }

    public async Task<string?> ReadAsync(string key)
    {
        var db = _raft.GetDatabase();
        return await db.StringGetAsync(key);

    }
    public Task<string?> GetRamdonCache(string key)
    {
        throw new NotImplementedException();
    }
    public async Task<bool> WriteAsync(string key, string value)
    {
        var db = _raft.GetDatabase();
        return await db.StringSetAsync(key, value);
    }
}
