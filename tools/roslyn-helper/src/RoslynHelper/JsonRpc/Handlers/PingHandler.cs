using System.Text.Json;

namespace RoslynHelper.JsonRpc.Handlers;

/// <summary>Health-check handler. Returns version string.</summary>
public sealed class PingHandler : IHandler
{
    public Task<object> HandleAsync(JsonElement? @params) =>
        Task.FromResult<object>(new { pong = true, version = "0.1.0" });
}
