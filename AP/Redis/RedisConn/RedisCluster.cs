using StackExchange.Redis;

namespace AP.Redis.RedisConn;

public class RedisCluster : IRedisConn
{
    public string MasterEndpoint { get; set; } = "RedisCluster not set";
    public string SlaveEndpoint { get; set; } = "RedisCluster not set";
    private ConnectionMultiplexer _cluster;
    public ConfigurationOptions ClusterConfig { get; set; }
    public RedisCluster(IConfiguration config)
    {
        try
        {
            var section = config.GetSection("Redis:RedisCluster");
            var nodes = section.GetSection("Nodes").Get<string[]>() ?? [];

            ClusterConfig = new ConfigurationOptions
            {
                AbortOnConnectFail = false,
                ConnectRetry = 5,
                SyncTimeout = 5000,
                ConnectTimeout = 60000,
                AllowAdmin = true,
                DefaultVersion = new Version(6, 2, 19),
                TieBreaker = "",
            };
            foreach (var s in nodes)
            {
                var (host, port) = EndpointUtil.ParseEndpoint(s);
                ClusterConfig.EndPoints.Add(host, port);
            }
            _cluster = ConnectionMultiplexer.Connect(ClusterConfig);

            foreach (var ep in _cluster.GetEndPoints())
            {
                var server = _cluster.GetServer(ep);
                if (server.IsConnected) server.Ping();
            }

            _cluster.ConnectionFailed += (_, e) =>
            {
                Console.WriteLine($"[Redis] ConnectionFailed: {e.EndPoint}, {e.FailureType}, {e.Exception?.Message}");
            };

            _cluster.ConnectionRestored += (_, e) =>
            {
                Console.WriteLine($"[Redis] ConnectionRestored: {e.EndPoint}");
            };

            _cluster.ConfigurationChanged += (_, _) => Console.WriteLine("[Redis] ConfigurationChanged");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[RedisCluster] Initialization failed: {ex.Message}");
            throw;
        }
    }

    public async Task<string?> ReadAsync(string key)
    {
        var db = _cluster.GetDatabase();
        return await db.StringGetAsync(key, CommandFlags.PreferReplica);
    }

    public Task<string?> GetRamdonCache(string key)
    {
        throw new NotImplementedException();
    }

    public async Task<bool> WriteAsync(string key, string value)
    {
        var db = _cluster.GetDatabase();
        return await db.StringSetAsync(key, value, flags: CommandFlags.DemandMaster);
    }
    public async Task<bool> FillCluster()
    {

        for (int i = 0; i < 10000; i++)
        {
            try
            {
                var db = _cluster.GetDatabase();
                string key = $"test:key:{i}:{Guid.NewGuid():N}"; //隨機 key（讓 slot 分散）
                string value = $"test:value:{i}";
                await db.StringSetAsync(key, value, flags: CommandFlags.DemandMaster);
            }
            catch (Exception ex)
            {
                continue;
            }
        }
        var db2 = _cluster.GetDatabase();
        return await db2.StringSetAsync("FillClusterFinal", "FillClusterFinal", flags: CommandFlags.DemandMaster);
    }
}
