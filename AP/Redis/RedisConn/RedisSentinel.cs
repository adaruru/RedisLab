using StackExchange.Redis;
using static StackExchange.Redis.Role;

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
            var sentinels = section.GetSection("Sentinel").Get<string[]>() ?? [];

            var sentinelConfig = new ConfigurationOptions
            {
                TieBreaker = "",
                CommandMap = CommandMap.Sentinel,
                AbortOnConnectFail = false
            };

            foreach (var s in sentinels)
            {
                var (host, port) = EndpointUtil.ParseEndpoint(s);
                sentinelConfig.EndPoints.Add(host, port);
            }
            var sentinelMux = ConnectionMultiplexer.Connect(sentinelConfig);

            var redisServiceConfig = new ConfigurationOptions()
            {
                ServiceName = masterName,
                AbortOnConnectFail = true,
                //Password = "",
            };

            var queryMaster = sentinelMux.GetSentinelMasterConnection(redisServiceConfig);

            MasterEndpoint = queryMaster.GetEndPoints()
                    .Select(ep => queryMaster.GetServer(ep))
                    .FirstOrDefault(s => !s.IsReplica)?.EndPoint.ToString() ?? "";

            _master = ConnectionMultiplexer.Connect(MasterEndpoint);

            var slaves = queryMaster.GetEndPoints()
                    .Select(ep => queryMaster.GetServer(ep))
                    .Where(s => s.IsReplica)
                    .Select(s => s.EndPoint.ToString());
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
