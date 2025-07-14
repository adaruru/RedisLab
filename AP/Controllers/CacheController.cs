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
    private readonly IRedisConn _redis;

    public CacheController(
        IConfiguration config,
        IRedisConn redis)
    {
        _config = config;
        _redis = redis;
    }

    [HttpGet]
    public async Task<IActionResult> GetCache([FromQuery] string key)
    {
        var value = await _redis.GetCache(key);
        return value != null ? Ok($"value: {value}, and write ip is {_redis.MasterEndpoint}") : StatusCode(500, $"key{key},not found, and write ip is {_redis.MasterEndpoint}");
    }

    [HttpPost]
    public async Task<IActionResult> UpdateCache([FromBody] CacheRequest req)
    {
        var result = await _redis.UpdateCache(req.Key, req.Value);
        return result ? Ok($"key{req.Key},value{req.Value},well saved, and read ip is {_redis.SlaveEndpoint}") : StatusCode(500, $"Failed to update cache , and read ip is {_redis.SlaveEndpoint}");
    }
}
public class CacheRequest
{
    public string Key { get; set; } = "";
    public string Value { get; set; } = "";
}
