using System.Text.RegularExpressions;
using System.Xml.Linq;

namespace RoslynHelper.Workspace;

/// <summary>
/// Parses .sln and .csproj files directly without a Roslyn/MSBuild workspace.
/// This avoids tight version coupling between MSBuild NuGet packages and the
/// installed SDK. For MVP commands (projectSummary), structural parsing is sufficient.
/// Roslyn semantic APIs will be integrated in a later phase.
/// </summary>
public sealed class WorkspaceManager : IDisposable
{
    public void Dispose() { } // reserved for future Roslyn workspace disposal
    private SolutionData? _solution;
    private string?       _solutionPath;

    public string? SolutionPath => _solutionPath;

    public Task<(int projectCount, int documentCount)> LoadAsync(string path)
    {
        if (!File.Exists(path))
            throw new KnownException("E_NOT_FOUND", $"solution file not found: {path}");

        SolutionData sln;
        try
        {
            sln = SolutionParser.Parse(path);
        }
        catch (Exception ex)
        {
            throw new KnownException("E_LOAD_FAILED", $"failed to load solution: {ex.Message}");
        }

        _solution     = sln;
        _solutionPath = path;

        var documentCount = sln.Projects.Sum(p => CountDocuments(p.ProjectPath));
        return Task.FromResult((sln.Projects.Count, documentCount));
    }

    public SolutionData GetCurrentSolution()
    {
        if (_solution is null)
            throw new KnownException("E_NOT_FOUND", "no solution loaded — call loadSolution first");
        return _solution;
    }

    /// <summary>Counts .cs/.fs/.vb source files in the project directory.</summary>
    private static int CountDocuments(string projPath)
    {
        var dir = Path.GetDirectoryName(projPath);
        if (dir is null || !Directory.Exists(dir)) return 0;
        try
        {
            return Directory.EnumerateFiles(dir, "*.cs", SearchOption.AllDirectories).Count()
                 + Directory.EnumerateFiles(dir, "*.fs", SearchOption.AllDirectories).Count()
                 + Directory.EnumerateFiles(dir, "*.vb", SearchOption.AllDirectories).Count();
        }
        catch { return 0; }
    }
}

/// <summary>Lightweight data model for a loaded solution.</summary>
public sealed class SolutionData(string solutionPath, List<ProjectEntry> projects)
{
    public string SolutionPath { get; } = solutionPath;
    public List<ProjectEntry> Projects { get; } = projects;
}

/// <summary>One project entry from the .sln file.</summary>
public sealed class ProjectEntry(string name, string relativePath, string absolutePath)
{
    public string Name         { get; } = name;
    public string RelativePath { get; } = relativePath;
    public string ProjectPath  { get; } = absolutePath;
}

/// <summary>
/// Parses the classic Visual Studio .sln format to extract project entries.
/// Handles both SDK-style and legacy projects.
/// </summary>
internal static class SolutionParser
{
    // Matches: Project("{type-guid}") = "Name", "path\to\project.csproj", "{project-guid}"
    private static readonly Regex ProjectLine = new(
        @"Project\(""\{[^}]+\}""\)\s*=\s*""([^""]+)""\s*,\s*""([^""]+)""\s*,",
        RegexOptions.Compiled);

    private static readonly HashSet<string> ProjectExtensions = new(StringComparer.OrdinalIgnoreCase)
    {
        ".csproj", ".fsproj", ".vbproj", ".pyproj", ".vcxproj"
    };

    public static SolutionData Parse(string slnPath)
    {
        var slnDir  = Path.GetDirectoryName(slnPath) ?? string.Empty;
        var content = File.ReadAllText(slnPath, System.Text.Encoding.UTF8);
        var projects = new List<ProjectEntry>();

        foreach (Match m in ProjectLine.Matches(content))
        {
            var name    = m.Groups[1].Value;
            var relPath = m.Groups[2].Value.Replace('/', Path.DirectorySeparatorChar);
            var ext     = Path.GetExtension(relPath);

            // Skip solution folders (no file extension) and non-code projects.
            if (string.IsNullOrEmpty(ext) || !ProjectExtensions.Contains(ext))
                continue;

            var absPath = Path.GetFullPath(Path.Combine(slnDir, relPath));
            if (!File.Exists(absPath)) continue; // skip phantom entries

            projects.Add(new ProjectEntry(name, relPath.Replace('\\', '/'), absPath));
        }

        return new SolutionData(slnPath, projects);
    }
}
