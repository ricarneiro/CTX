using System.Text.Json.Serialization;
using System.Xml.Linq;
using RoslynHelper.Workspace;

namespace RoslynHelper.Models;

public sealed record PackageRef(
    [property: JsonPropertyName("name")]    string Name,
    [property: JsonPropertyName("version")] string Version
);

public sealed record ProjectSummary(
    [property: JsonPropertyName("name")]               string           Name,
    [property: JsonPropertyName("path")]               string           Path,
    [property: JsonPropertyName("type")]               string           Type,
    [property: JsonPropertyName("targetFrameworks")]   string[]         TargetFrameworks,
    [property: JsonPropertyName("outputType")]         string           OutputType,
    [property: JsonPropertyName("rootNamespace")]      string           RootNamespace,
    [property: JsonPropertyName("documentCount")]      int              DocumentCount,
    [property: JsonPropertyName("projectReferences")]  List<string>     ProjectReferences,
    [property: JsonPropertyName("packageReferences")]  List<PackageRef> PackageReferences
);

public sealed record SolutionSummary(
    [property: JsonPropertyName("solutionPath")] string               SolutionPath,
    [property: JsonPropertyName("solutionName")] string               SolutionName,
    [property: JsonPropertyName("projects")]     List<ProjectSummary> Projects
);

/// <summary>
/// Builds a <see cref="ProjectSummary"/> from a <see cref="ProjectEntry"/>
/// by parsing the .csproj XML for MSBuild properties.
/// </summary>
internal static class ProjectSummaryBuilder
{
    public static ProjectSummary Build(ProjectEntry entry, string solutionDir, int documentCount)
    {
        var projPath = entry.ProjectPath;

        string[]         targetFrameworks = [];
        string           outputType       = "Library";
        string           rootNamespace    = entry.Name;
        List<PackageRef> packageRefs      = [];
        List<string>     projRefs         = [];

        if (File.Exists(projPath))
        {
            try
            {
                var doc = XDocument.Load(projPath);

                var tf  = doc.Descendants("TargetFramework").FirstOrDefault()?.Value?.Trim();
                var tfs = doc.Descendants("TargetFrameworks").FirstOrDefault()?.Value?.Trim();

                targetFrameworks = tfs is not null
                    ? tfs.Split(';', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries)
                    : tf is not null ? [tf] : [];

                outputType    = doc.Descendants("OutputType").FirstOrDefault()?.Value?.Trim() ?? "Library";
                rootNamespace = doc.Descendants("RootNamespace").FirstOrDefault()?.Value?.Trim() ?? entry.Name;

                packageRefs = doc.Descendants("PackageReference")
                    .Select(el => new PackageRef(
                        el.Attribute("Include")?.Value ?? string.Empty,
                        el.Attribute("Version")?.Value ?? el.Element("Version")?.Value ?? string.Empty))
                    .Where(pr => !string.IsNullOrEmpty(pr.Name))
                    .OrderBy(pr => pr.Name)
                    .ToList();

                projRefs = doc.Descendants("ProjectReference")
                    .Select(el =>
                    {
                        var include = el.Attribute("Include")?.Value ?? string.Empty;
                        return Path.GetFileNameWithoutExtension(include);
                    })
                    .Where(n => !string.IsNullOrEmpty(n))
                    .OrderBy(n => n)
                    .ToList();
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine($"[summary] warn: could not parse {projPath}: {ex.Message}");
            }
        }

        var type = outputType.Equals("Exe",    StringComparison.OrdinalIgnoreCase) ||
                   outputType.Equals("WinExe", StringComparison.OrdinalIgnoreCase)
            ? "exe" : "lib";

        return new ProjectSummary(
            entry.Name,
            entry.RelativePath,
            type,
            targetFrameworks,
            outputType,
            rootNamespace,
            documentCount,
            projRefs,
            packageRefs
        );
    }
}
