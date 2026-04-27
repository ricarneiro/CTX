namespace RoslynHelper;

/// <summary>
/// An expected error with a standardized error code.
/// Throw from handlers to produce structured JSON-RPC error responses.
/// </summary>
public sealed class KnownException(string code, string message) : Exception(message)
{
    public string Code { get; } = code;
}
