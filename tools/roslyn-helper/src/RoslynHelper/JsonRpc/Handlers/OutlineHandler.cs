using System.Text.Json;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using RoslynHelper.Models;

namespace RoslynHelper.JsonRpc.Handlers;

/// <summary>
/// Parses a single .cs file and returns its structural outline:
/// namespaces, types, method signatures (no bodies), properties, fields, events.
/// Does NOT require a solution to be loaded — works on a single file.
/// </summary>
public sealed class OutlineHandler : IHandler
{
    public async Task<object> HandleAsync(JsonElement? @params)
    {
        if (@params is not { } p)
            throw new KnownException("E_INVALID_PARAMS", "params required for outline");

        if (!p.TryGetProperty("path", out var pathEl) || pathEl.ValueKind != JsonValueKind.String)
            throw new KnownException("E_INVALID_PARAMS", "params.path (string) is required");

        var path = pathEl.GetString()!;
        if (!File.Exists(path))
            throw new KnownException("E_NOT_FOUND", $"file not found: {path}");

        var source = await File.ReadAllTextAsync(path, System.Text.Encoding.UTF8);
        var tree   = CSharpSyntaxTree.ParseText(source, path: path);
        var root   = (CompilationUnitSyntax)await tree.GetRootAsync();

        bool hasSyntaxErrors = tree.GetDiagnostics()
            .Any(d => d.Severity == DiagnosticSeverity.Error);

        int lineCount = source.Split('\n').Length;

        // Top-level usings
        var usings = CollectUsings(root.Usings);
        string ns   = "";
        var types   = new List<OutlineTypeModel>();
        bool hasTopLevel = false;

        foreach (var member in root.Members)
        {
            switch (member)
            {
                case NamespaceDeclarationSyntax blockNs:
                    ns = blockNs.Name.ToString();
                    usings.AddRange(CollectUsings(blockNs.Usings));
                    foreach (var m in blockNs.Members)
                        CollectType(m, types);
                    break;

                case FileScopedNamespaceDeclarationSyntax fileScopedNs:
                    ns = fileScopedNs.Name.ToString();
                    usings.AddRange(CollectUsings(fileScopedNs.Usings));
                    foreach (var m in fileScopedNs.Members)
                        CollectType(m, types);
                    break;

                case GlobalStatementSyntax:
                    hasTopLevel = true;
                    break;

                default:
                    CollectType(member, types);
                    break;
            }
        }

        // Synthetic Program type for top-level statement files
        if (hasTopLevel)
        {
            types.Insert(0, new OutlineTypeModel(
                "class", "Program (top-level program)", [], [],
                [new OutlineMemberModel("method", "static void Main(string[] args)", ["static"], 1, false)],
                []));
        }

        return new OutlineResult(path, ns, lineCount, usings.Distinct().ToList(), types, hasSyntaxErrors);
    }

    // ─── Usings ─────────────────────────────────────────────────────────────

    private static List<string> CollectUsings(SyntaxList<UsingDirectiveSyntax> usings) =>
        usings
            .Where(u => u.Alias is null)
            .Select(u => u.Name?.ToString() ?? "")
            .Where(s => !string.IsNullOrEmpty(s))
            .ToList();

    // ─── Type dispatch ───────────────────────────────────────────────────────

    private static void CollectType(MemberDeclarationSyntax member, List<OutlineTypeModel> target)
    {
        switch (member)
        {
            case ClassDeclarationSyntax cls:
                target.Add(ExtractTypeDecl("class", cls.Identifier.Text,
                    cls.TypeParameterList, null, cls.Modifiers, cls.BaseList, cls.Members));
                break;

            case InterfaceDeclarationSyntax iface:
                target.Add(ExtractTypeDecl("interface", iface.Identifier.Text,
                    iface.TypeParameterList, null, iface.Modifiers, iface.BaseList, iface.Members));
                break;

            case StructDeclarationSyntax str:
                target.Add(ExtractTypeDecl("struct", str.Identifier.Text,
                    str.TypeParameterList, null, str.Modifiers, str.BaseList, str.Members));
                break;

            case RecordDeclarationSyntax rec:
            {
                var recKind = rec.ClassOrStructKeyword.IsKind(SyntaxKind.StructKeyword)
                    ? "record struct" : "record";
                target.Add(ExtractTypeDecl(recKind, rec.Identifier.Text,
                    rec.TypeParameterList, rec.ParameterList, rec.Modifiers, rec.BaseList, rec.Members));
                break;
            }

            case EnumDeclarationSyntax en:
                target.Add(ExtractEnum(en));
                break;
        }
    }

    // ─── Type extraction ─────────────────────────────────────────────────────

    private static OutlineTypeModel ExtractTypeDecl(
        string kind,
        string name,
        TypeParameterListSyntax? typeParams,
        ParameterListSyntax? recordParams,
        SyntaxTokenList modifiers,
        BaseListSyntax? baseList,
        SyntaxList<MemberDeclarationSyntax> members)
    {
        var mods      = modifiers.Select(m => m.Text).ToList();
        var baseTypes = baseList?.Types.Select(t => t.Type.ToString()).ToList() ?? [];
        var fullName  = name
                      + (typeParams?.ToString()  ?? "")
                      + (recordParams?.ToString() ?? "");

        var memberModels = new List<OutlineMemberModel>();
        var nested       = new List<OutlineTypeModel>();

        foreach (var member in members)
        {
            switch (member)
            {
                case MethodDeclarationSyntax m:
                    memberModels.Add(ExtractMethod(m));
                    break;
                case ConstructorDeclarationSyntax c:
                    memberModels.Add(ExtractConstructor(c));
                    break;
                case PropertyDeclarationSyntax prop:
                    memberModels.Add(ExtractProperty(prop));
                    break;
                case FieldDeclarationSyntax f:
                    memberModels.AddRange(ExtractFields(f));
                    break;
                case EventDeclarationSyntax e:
                    memberModels.Add(ExtractEventDecl(e));
                    break;
                case EventFieldDeclarationSyntax ef:
                    memberModels.AddRange(ExtractEventFields(ef));
                    break;
                case ClassDeclarationSyntax _:
                case InterfaceDeclarationSyntax _:
                case StructDeclarationSyntax _:
                case RecordDeclarationSyntax _:
                case EnumDeclarationSyntax _:
                    CollectType(member, nested);
                    break;
            }
        }

        return new OutlineTypeModel(kind, fullName, mods, baseTypes, memberModels, nested);
    }

    private static OutlineTypeModel ExtractEnum(EnumDeclarationSyntax en)
    {
        var mods    = en.Modifiers.Select(m => m.Text).ToList();
        var members = en.Members
            .Select(m => new OutlineMemberModel("enumValue", m.Identifier.Text, [], GetLine(m), false))
            .ToList();
        return new OutlineTypeModel("enum", en.Identifier.Text, mods, [], members, []);
    }

    // ─── Member extraction ───────────────────────────────────────────────────

    private static OutlineMemberModel ExtractMethod(MethodDeclarationSyntax m)
    {
        var typeParams = m.TypeParameterList?.ToString() ?? "";
        var sig        = $"{m.ReturnType} {m.Identifier.Text}{typeParams}{m.ParameterList}".Trim();
        var mods       = m.Modifiers.Select(x => x.Text).ToList();
        return new OutlineMemberModel("method", sig, mods, GetLine(m), HasObsolete(m.AttributeLists));
    }

    private static OutlineMemberModel ExtractConstructor(ConstructorDeclarationSyntax c)
    {
        var sig  = $"{c.Identifier.Text}{c.ParameterList}".Trim();
        var mods = c.Modifiers.Select(x => x.Text).ToList();
        return new OutlineMemberModel("constructor", sig, mods, GetLine(c), HasObsolete(c.AttributeLists));
    }

    private static OutlineMemberModel ExtractProperty(PropertyDeclarationSyntax p)
    {
        string accessors;
        if (p.AccessorList is { } al)
        {
            var parts = al.Accessors.Select(a =>
            {
                var accMods = a.Modifiers.Any() ? a.Modifiers.ToString() + " " : "";
                return accMods + a.Keyword.Text;
            });
            accessors = "{ " + string.Join("; ", parts) + "; }";
        }
        else
        {
            accessors = "=> ...";  // expression-bodied
        }

        var sig  = $"{p.Type} {p.Identifier.Text} {accessors}".Trim();
        var mods = p.Modifiers.Select(x => x.Text).ToList();
        return new OutlineMemberModel("property", sig, mods, GetLine(p), HasObsolete(p.AttributeLists));
    }

    private static IEnumerable<OutlineMemberModel> ExtractFields(FieldDeclarationSyntax f)
    {
        var mods     = f.Modifiers.Select(x => x.Text).ToList();
        var typeName = f.Declaration.Type.ToString();
        var isConst  = mods.Contains("const");
        var obs      = HasObsolete(f.AttributeLists);
        var line     = GetLine(f);

        foreach (var v in f.Declaration.Variables)
        {
            var sig = isConst && v.Initializer is { } init
                ? $"{typeName} {v.Identifier.Text} = {init.Value}"
                : $"{typeName} {v.Identifier.Text}";
            yield return new OutlineMemberModel("field", sig.Trim(), mods, line, obs);
        }
    }

    private static OutlineMemberModel ExtractEventDecl(EventDeclarationSyntax e)
    {
        var sig  = $"event {e.Type} {e.Identifier.Text}".Trim();
        var mods = e.Modifiers.Select(x => x.Text).ToList();
        return new OutlineMemberModel("event", sig, mods, GetLine(e), HasObsolete(e.AttributeLists));
    }

    private static IEnumerable<OutlineMemberModel> ExtractEventFields(EventFieldDeclarationSyntax ef)
    {
        var mods     = ef.Modifiers.Select(x => x.Text).ToList();
        var typeName = ef.Declaration.Type.ToString();
        var obs      = HasObsolete(ef.AttributeLists);
        var line     = GetLine(ef);

        foreach (var v in ef.Declaration.Variables)
        {
            yield return new OutlineMemberModel(
                "event", $"event {typeName} {v.Identifier.Text}".Trim(), mods, line, obs);
        }
    }

    // ─── Helpers ─────────────────────────────────────────────────────────────

    // Returns true if obsolete, null otherwise (null omitted by WhenWritingNull in dispatcher)
    private static bool? HasObsolete(SyntaxList<AttributeListSyntax> attrs)
    {
        var found = attrs.SelectMany(al => al.Attributes)
                         .Any(a => a.Name.ToString() is "Obsolete" or "ObsoleteAttribute");
        return found ? true : null;
    }

    private static int GetLine(SyntaxNode node) =>
        node.GetLocation().GetLineSpan().StartLinePosition.Line + 1;
}
