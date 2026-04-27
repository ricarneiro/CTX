using System.Text.Json;
using RoslynHelper.Models;
using RoslynHelper.Workspace;

namespace RoslynHelper.JsonRpc.Handlers;

/// <summary>
/// Returns a compact summary of the loaded solution: projects, target frameworks,
/// package references, and project-to-project dependencies.
/// Parses .csproj XML directly — no Roslyn workspace required.
/// </summary>
public sealed class ProjectSummaryHandler(WorkspaceManager workspace) : IHandler
{
    public Task<object> HandleAsync(JsonElement? @params)
    {
        var solution     = workspace.GetCurrentSolution();
        var solutionPath = workspace.SolutionPath!;
        var solutionName = Path.GetFileNameWithoutExtension(solutionPath);
        var solutionDir  = Path.GetDirectoryName(solutionPath) ?? string.Empty;

        // Document counts are tracked per-project in the WorkspaceManager load.
        // Re-compute here from file system for simplicity.
        var projects = solution.Projects
            .OrderBy(p => p.Name)
            .Select(p =>
            {
                var docCount = CountDocuments(p.ProjectPath);
                return ProjectSummaryBuilder.Build(p, solutionDir, docCount);
            })
            .ToList();

        return Task.FromResult<object>(new SolutionSummary(solutionPath, solutionName, projects));
    }

    private static int CountDocuments(string projPath)
    {
        var dir = Path.GetDirectoryName(projPath);
        if (dir is null || !Directory.Exists(dir)) return 0;
        try
        {
            return Directory.EnumerateFiles(dir, "*.cs",  SearchOption.AllDirectories).Count()
                 + Directory.EnumerateFiles(dir, "*.fs",  SearchOption.AllDirectories).Count()
                 + Directory.EnumerateFiles(dir, "*.vb",  SearchOption.AllDirectories).Count();
        }
        catch { return 0; }
    }
}
