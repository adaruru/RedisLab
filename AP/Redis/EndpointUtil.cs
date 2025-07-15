
namespace AP.Redis;

public static class EndpointUtil
{
    public static (string, int) ParseEndpoint(string endpoint)
    {
        var parts = endpoint.Split(':');
        if (parts.Length != 2 || !int.TryParse(parts[1], out var port))
            throw new FormatException($"Invalid endpoint format: {endpoint}");
        return (parts[0], port);
    }
}