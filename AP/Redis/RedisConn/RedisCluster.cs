using Microsoft.AspNetCore.DataProtection.KeyManagement;
using Microsoft.Extensions.Options;
using StackExchange.Redis;
using static StackExchange.Redis.Role;

namespace AP.Redis.RedisConn;

public class RedisCluster : IRedisConn
{
    private readonly ConnectionMultiplexer _cluster;
    public string MasterEndpoint { get; set; } = "RedisCluster not set";
    public string SlaveEndpoint { get; set; } = "RedisCluster not set";
    public RedisCluster(IConfiguration config)
    {
        try
        {
            var section = config.GetSection("Redis:RedisCluster");
            var nodes = section.GetSection("Nodes").Get<string[]>() ?? [];

            var clusterConfig = new ConfigurationOptions
            {
                AbortOnConnectFail = false,
                ConnectRetry = 5,
                SyncTimeout = 5000,
                CommandMap = CommandMap.Create(new HashSet<string>
                    // 禁用不支援 cluster 的命令
                    { "INFO", "CONFIG", "CLUSTER", "PING" }, available: false),
                AllowAdmin = true,
                TieBreaker = "",
            };
            foreach (var s in nodes)
            {
                var (host, port) = EndpointUtil.ParseEndpoint(s);
                clusterConfig.EndPoints.Add(host, port);
            }
            _cluster = ConnectionMultiplexer.Connect(clusterConfig);
        }
        catch (Exception ex)
        {
            throw ex;
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
    public bool FillCluster()
    {
        try
        {
            var db = _cluster.GetDatabase();
            for (int i = 0; i < 1000000; i++)
            {
                string key = $"{i}:test:key:{i}:{Guid.NewGuid():N}"; //隨機 key（讓 slot 分散）
                string value = $"test:value:{i}";
                db.StringSet(key, value, flags: CommandFlags.DemandMaster);
            }
            return db.StringSet("FillClusterFinal", "FillClusterFinal", flags: CommandFlags.DemandMaster);
        }
        catch (Exception ex)
        {
            return false;
        }
    }
}
