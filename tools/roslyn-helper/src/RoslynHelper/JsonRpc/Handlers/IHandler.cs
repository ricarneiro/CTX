using System.Text.Json;

namespace RoslynHelper.JsonRpc.Handlers;

/// <summary>Contract for a JSON-RPC method handler.</summary>
public interface IHandler
{
    /// <summary>Execute the handler and return the result object (serialized as the 'result' field).</summary>
    Task<object> HandleAsync(JsonElement? @params);
}
