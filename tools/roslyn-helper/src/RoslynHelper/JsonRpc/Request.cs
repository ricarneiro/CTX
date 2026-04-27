using System.Text.Json;
using System.Text.Json.Serialization;

namespace RoslynHelper.JsonRpc;

/// <summary>Incoming JSON-RPC request from ctx (Go).</summary>
public sealed record Request(
    [property: JsonPropertyName("id")]     int          Id,
    [property: JsonPropertyName("method")] string       Method,
    [property: JsonPropertyName("params")] JsonElement? Params
);
