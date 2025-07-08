using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public class RedisSentinel : IRedisConn
{

    private readonly ConnectionMultiplexer _master;
    private readonly ConnectionMultiplexer _slave;
    private readonly List<ConnectionMultiplexer> _slaves = new List<ConnectionMultiplexer>();
    private static readonly Random _random = new();
    public string MasterEndpoint { get; set; }
    public string SlaveEndpoint { get; set; }

    public RedisSentinel(IConfiguration config)
    {
        try
        {
            var section = config.GetSection("Redis:RedisSentinel");
            var masterName = section.GetValue<string>("MasterName") ?? "mymaster";
            var nodes = section.GetSection("Nodes").Get<string[]>() ?? [];
            var sentinels = section.GetSection("Sentinel").Get<string[]>() ?? [];

            var sentinelConfig = new ConfigurationOptions
            {
                TieBreaker = "",
                CommandMap = CommandMap.Sentinel,
                AbortOnConnectFail = false,
                ServiceName = masterName
            };

            foreach (var s in sentinels)
                sentinelConfig.EndPoints.Add(s);
            _master = ConnectionMultiplexer.Connect(sentinelConfig);

            MasterEndpoint = _master.GetEndPoints().FirstOrDefault()?.ToString()??"master map fail";

            var slaves = nodes.Where(node => node != MasterEndpoint).ToList();
            foreach (var slave in slaves)
            {
                _slaves.Add(ConnectionMultiplexer.Connect(slave));
            }
            SlaveEndpoint = slaves.FirstOrDefault();
            _slave = ConnectionMultiplexer.Connect(slaves.FirstOrDefault() ?? "");
        }
        catch (Exception ex)
        {
            throw ex;
        }
    }

    public async Task<string?> GetCache(string key)
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

    public async Task<bool> UpdateCache(string key, string value)
    {
        var db = _master.GetDatabase();
        return await db.StringSetAsync(key, value);
    }
}
