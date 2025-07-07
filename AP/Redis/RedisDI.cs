using AP.Redis.RedisConn;

namespace AP.Redis;
public static class RedisServiceRegistration
{
    public static IServiceCollection AddRedisService(this IServiceCollection services, IConfiguration config)
    {
        var modeStr = config.GetValue<string>("Redis:Mode") ?? "RedisMasterSlaves";
        if (!Enum.TryParse<RedisMode>(modeStr, ignoreCase: true, out var mode))
            throw new Exception($"Unsupported Redis mode: {modeStr}");

        switch (mode)
        {
            case RedisMode.RedisMasterSlaves:
                services.AddSingleton<IRedisConn, RedisMasterSlave>();
                break;
            case RedisMode.RedisSentinel:
                services.AddSingleton<IRedisConn, RedisSentinel>();
                break;
            case RedisMode.RedisCluster:
                services.AddSingleton<IRedisConn, RedisCluster>();
                break;
            case RedisMode.RedisRaft:
                services.AddSingleton<IRedisConn, RedisRaft>();
                break;
            default:
                throw new Exception($"Unhandled Redis mode: {mode}");
        }
        return services;
    }
}