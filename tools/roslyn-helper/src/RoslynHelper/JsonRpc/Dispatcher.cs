using System.Text.Encodings.Web;
using System.Text.Json;
using System.Text.Json.Serialization;
using RoslynHelper.JsonRpc.Handlers;
using RoslynHelper.Workspace;

namespace RoslynHelper.JsonRpc;

/// <summary>
/// Reads newline-delimited JSON requests from stdin, dispatches to registered
/// handlers, and writes responses to stdout. One request per line, one response
/// per line, UTF-8 without BOM.
/// </summary>
public sealed class Dispatcher
{
    private static readonly JsonSerializerOptions JsonOpts = new()
    {
        PropertyNamingPolicy   = JsonNamingPolicy.CamelCase,
        WriteIndented          = false,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
        // Allow literal single quotes and em dashes; Go's json.Unmarshal handles them fine.
        Encoder = JavaScriptEncoder.UnsafeRelaxedJsonEscaping,
    };

    private readonly Dictionary<string, IHandler> _handlers;

    public Dispatcher(WorkspaceManager workspace)
    {
        _handlers = new Dictionary<string, IHandler>(StringComparer.OrdinalIgnoreCase)
        {
            ["ping"]           = new PingHandler(),
            ["loadSolution"]   = new LoadSolutionHandler(workspace),
            ["projectSummary"] = new ProjectSummaryHandler(workspace),
            ["outline"]        = new OutlineHandler(),
        };
    }

    public async Task RunAsync(TextReader stdin, TextWriter stdout)
    {
        string? line;
        while ((line = await stdin.ReadLineAsync()) is not null)
        {
            if (string.IsNullOrWhiteSpace(line)) continue;

            var response = await DispatchAsync(line);
            if (response is null) return; // shutdown

            await stdout.WriteLineAsync(response);
            await stdout.FlushAsync();
        }
    }

    private async Task<string?> DispatchAsync(string line)
    {
        Request? req;
        try
        {
            req = JsonSerializer.Deserialize<Request>(line, JsonOpts);
            if (req is null || string.IsNullOrEmpty(req.Method))
                return Err(0, "E_INVALID_REQUEST", "request must have id and method");
        }
        catch (JsonException ex)
        {
            return Err(0, "E_INVALID_REQUEST", $"invalid JSON: {ex.Message}");
        }

        if (req.Method.Equals("shutdown", StringComparison.OrdinalIgnoreCase))
            return null; // signal caller to exit

        if (!_handlers.TryGetValue(req.Method, out var handler))
            return Err(req.Id, "E_UNKNOWN_METHOD", $"method '{req.Method}' not found");

        try
        {
            var result = await handler.HandleAsync(req.Params);
            return Ok(req.Id, result);
        }
        catch (KnownException kex)
        {
            return Err(req.Id, kex.Code, kex.Message);
        }
        catch (Exception ex)
        {
            return ErrWithData(req.Id, "E_INTERNAL", ex.Message, new { stackTrace = ex.StackTrace });
        }
    }

    private string Ok(int id, object result) =>
        JsonSerializer.Serialize(new { id, result }, JsonOpts);

    private string Err(int id, string code, string message) =>
        JsonSerializer.Serialize(new { id, error = new { code, message } }, JsonOpts);

    private string ErrWithData(int id, string code, string message, object data) =>
        JsonSerializer.Serialize(new { id, error = new { code, message, data } }, JsonOpts);
}
