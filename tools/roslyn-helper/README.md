# ctx-roslyn-helper

C# subprocess invoked by `ctx csharp` commands. Loads .NET solutions with
Roslyn and answers JSON-RPC queries over stdin/stdout.

## Protocol

One JSON object per line, UTF-8 without BOM.

**Request:** `{"id": 1, "method": "ping", "params": {}}`
**Success:** `{"id": 1, "result": {...}}`
**Error:**   `{"id": 1, "error": {"code": "E_NOT_FOUND", "message": "..."}}`

### Methods

| Method           | Params                    | Description                        |
|------------------|---------------------------|------------------------------------|
| `ping`           | `{}`                      | Health check, returns version      |
| `loadSolution`   | `{"path": "C:\\...\\x.sln"}` | Load solution into workspace    |
| `projectSummary` | `{}`                      | Summary of loaded solution         |
| `shutdown`       | `{}`                      | Exit cleanly (no response written) |

## Build

Requires: .NET 8+ SDK.

```sh
dotnet build src/RoslynHelper -c Release
```

## Publish

```sh
dotnet publish src/RoslynHelper -c Release -r win-x64 --self-contained false -o publish/
```

Output: `publish/ctx-roslyn-helper.exe`

## Manual test

```sh
cd publish
./ctx-roslyn-helper.exe
```

Then type line by line:

```
{"id": 1, "method": "ping", "params": {}}
{"id": 2, "method": "loadSolution", "params": {"path": "C:\\path\\to\\My.sln"}}
{"id": 3, "method": "projectSummary", "params": {}}
{"id": 99, "method": "shutdown", "params": {}}
```

## Notes

- Does NOT need to be self-contained. Requires .NET 8+ runtime on the target machine.
  Users of `ctx csharp` are .NET developers — the runtime is already there.
- MSBuild warnings logged to stderr are informational (missing SDK targets, etc.).
  Only `WorkspaceFailed` events with `Failure` kind indicate real problems.
- The helper is spawned once per `ctx` invocation and kept alive for all queries
  in that session. ctx Go manages the subprocess lifecycle.
