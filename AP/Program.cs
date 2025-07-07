using AP.Redis;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();
builder.Services.AddRedisService(builder.Configuration);
builder.Services.AddControllers();

var settingName = "appsettings";
builder.Configuration
    .AddJsonFile(path: $"{settingName}.json", optional: true)
    .AddJsonFile(path: $"{settingName}.{builder.Environment.EnvironmentName}.json", optional: true)
    .Build();
var app = builder.Build();

app.UseSwagger();
app.UseSwaggerUI();

app.Run();