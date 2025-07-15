using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public class RedisMasterSlave : IRedisConn
{
    private readonly ConnectionMultiplexer _master;
    private readonly ConnectionMultiplexer _slave;
    private readonly List<ConnectionMultiplexer> _slaves = new List<ConnectionMultiplexer>();
    private static readonly Random _random = new();
    public string MasterEndpoint { get; set; }
    public string SlaveEndpoint { get; set; }
    public RedisMasterSlave(IConfiguration config)
    {
        var section = config.GetSection("Redis:RedisMasterSlaves");
        MasterEndpoint = section.GetValue<string>("Master") ?? "";
        _master = ConnectionMultiplexer.Connect(section.GetValue<string>("Master")!);

        var slaves = section.GetSection("Slaves").Get<string[]>() ?? [];

        foreach (var slave in slaves)
        {
            _slaves.Add(ConnectionMultiplexer.Connect(slave));
        }
        _slave = ConnectionMultiplexer.Connect(slaves.FirstOrDefault() ?? "");
        SlaveEndpoint = slaves.FirstOrDefault() ?? "";
    }

    public async Task<string?> ReadAsync(string key)
    {
        //var db = _master.GetDatabase();
        var db = _slave.GetDatabase();
        var value = await db.StringGetAsync(key);
        if (value.HasValue)
            return value;
        return null;
    }

    public async Task<string?> GetRamdonCache(string key)
    {
        foreach (var slave in _slaves.OrderBy(_ => _random.Next()))
        {
            var db = slave.GetDatabase();
            var value = await db.StringGetAsync(key);
            if (value.HasValue)
                return value;
        }
        return null;
    }

    public async Task<bool> WriteAsync(string key, string value)
    {
        var db = _master.GetDatabase();
        return await db.StringSetAsync(key, value);
    }
}
