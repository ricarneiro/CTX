using System.Text.Json;
using RoslynHelper.Workspace;

namespace RoslynHelper.JsonRpc.Handlers;

/// <summary>
/// Loads a .sln file into the Roslyn workspace.
/// Idempotent — if the same path is already loaded, reloads it.
/// </summary>
public sealed class LoadSolutionHandler(WorkspaceManager workspace) : IHandler
{
    public async Task<object> HandleAsync(JsonElement? @params)
    {
        if (@params is not { } p)
            throw new KnownException("E_INVALID_REQUEST", "params required for loadSolution");

        if (!p.TryGetProperty("path", out var pathEl) || pathEl.ValueKind != JsonValueKind.String)
            throw new KnownException("E_INVALID_REQUEST", "params.path (string) is required");

        var path = pathEl.GetString()!;
        var (projectCount, documentCount) = await workspace.LoadAsync(path);
        return new { loaded = true, projectCount, documentCount };
    }
}
