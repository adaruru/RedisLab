using AP.Redis;
using AP.Redis.RedisConn;
using Microsoft.AspNetCore.Mvc;
namespace AP.Controllers;

/// <summary>
/// 登入：提供使用者登入系統.
/// </summary>
/// <seealso cref="BaseUI.MVC.BaseController" />
public class CacheController : ApiBaseController
{
    private readonly IConfiguration _config;
    private readonly IHostEnvironment _env;
    private readonly IRedisConn _redis;

    public CacheController(
        IConfiguration config,
        IHostEnvironment env,
        IRedisConn redis)
    {
        _config = config;
        _env = env;
        _redis = redis;
    }

    [HttpGet]
    public async Task<IActionResult> GetCache([FromQuery] string key)
    {
        var value = await _redis.ReadAsync(key);
        return value != null ? Ok($"value: {value}, and write ip is {_redis.MasterEndpoint}") : StatusCode(500, $"key{key},not found, and write ip is {_redis.MasterEndpoint}");
    }

    [HttpPost]
    public async Task<IActionResult> UpdateCache([FromBody] CacheRequest req)
    {
        var result = await _redis.WriteAsync(req.Key, req.Value);
        return result ? Ok($"key{req.Key},value{req.Value},well saved, and read ip is {_redis.SlaveEndpoint}") : StatusCode(500, $"Failed to update cache , and read ip is {_redis.SlaveEndpoint}");
    }

    [HttpGet]
    public async Task<IActionResult> FillCluster()
    {
        try
        {
            var modeStr = _config.GetValue<string>("Redis:Mode") ?? "RedisMasterSlaves";
            if (!Enum.TryParse<RedisMode>(modeStr, ignoreCase: true, out var mode))
                throw new Exception($"Unsupported Redis mode: {modeStr}");
            if (mode == RedisMode.RedisCluster)
            {
                var redis = new RedisCluster(_config);
                var result = await redis.FillCluster();
                if (result)
                {
                    return Ok($"Redis:Mode:{modeStr}，填充測試資料完成");
                }
                else return Ok($"填充測試失敗");

            }
            return Ok($"Redis:Mode:{modeStr}，不做填充測試");
        }
        catch (Exception ex)
        {
            return Ok($"Exception: {ex.ToString()}");
        }

    }
}
public class CacheRequest
{
    public string Key { get; set; } = "";
    public string Value { get; set; } = "";
}
