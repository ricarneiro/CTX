using System.Text.Json.Serialization;

namespace RoslynHelper.Models;

public sealed record OutlineResult(
    [property: JsonPropertyName("path")]           string                 Path,
    [property: JsonPropertyName("namespace")]      string                 Namespace,
    [property: JsonPropertyName("lineCount")]      int                    LineCount,
    [property: JsonPropertyName("usings")]         List<string>           Usings,
    [property: JsonPropertyName("types")]          List<OutlineTypeModel> Types,
    [property: JsonPropertyName("hasSyntaxErrors")] bool                  HasSyntaxErrors
);

public sealed record OutlineTypeModel(
    [property: JsonPropertyName("kind")]      string                   Kind,
    [property: JsonPropertyName("name")]      string                   Name,
    [property: JsonPropertyName("modifiers")] List<string>             Modifiers,
    [property: JsonPropertyName("baseTypes")] List<string>             BaseTypes,
    [property: JsonPropertyName("members")]   List<OutlineMemberModel> Members,
    [property: JsonPropertyName("nested")]    List<OutlineTypeModel>   Nested
);

/// <remarks>
/// <c>IsObsolete</c> is nullable so that <c>null</c> (not obsolete) is omitted
/// from JSON by the dispatcher's <c>WhenWritingNull</c> policy.
/// </remarks>
public sealed record OutlineMemberModel(
    [property: JsonPropertyName("kind")]       string       Kind,
    [property: JsonPropertyName("signature")]  string       Signature,
    [property: JsonPropertyName("modifiers")]  List<string> Modifiers,
    [property: JsonPropertyName("line")]       int          Line,
    [property: JsonPropertyName("isObsolete")] bool?        IsObsolete = null
);
