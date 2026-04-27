using System.Text.Json.Serialization;

namespace RoslynHelper.JsonRpc;

/// <summary>Error detail embedded in a JSON-RPC error response.</summary>
public sealed record ErrorDetail(
    [property: JsonPropertyName("code")]    string  Code,
    [property: JsonPropertyName("message")] string  Message,
    [property: JsonPropertyName("data")]    object? Data = null
);
