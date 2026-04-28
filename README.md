# ctx — anti-tokens CLI for Claude Code

> Analyze your codebase locally. Feed Claude dense summaries, not raw files.

**Status:** alpha — under active development. Interfaces will change.

## Why

Every token Claude reads costs money and burns context window. When working on
a large C# or React codebase, Claude often spends thousands of tokens just
reading files to build a mental model before doing any actual work.

`ctx` runs locally, analyzes your project with language-aware tools (Roslyn for
C#, tree-sitter for TypeScript), and emits compact markdown summaries that give
Claude everything it needs in a fraction of the tokens.

## Installation

```sh
go install github.com/ricarneiro/ctx/cmd/ctx@latest
```

Requires Go 1.22+. Binaries for common platforms will be published after the
MVP is validated.

## Usage

```sh
# Git context: recent commits, status, branch info
ctx git

# Auto-detect stack and emit project overview
ctx auto project

# C# project structure (requires .NET SDK)
ctx csharp project

# C# file outline: types, methods, signatures
ctx csharp outline src/MyService.cs

# List compilation errors
ctx csharp errors
```

All output is UTF-8 markdown on stdout. Pipe it where you need it:

```sh
ctx csharp project | pbcopy   # macOS
ctx csharp project | clip     # Windows
```

Or reference it in a `CLAUDE.md`:

```markdown
Run `ctx csharp project` to get the project overview before making changes.
```

## Architecture

`ctx` uses a plugin system. Each stack (`git`, `csharp`, `react`, `auto`) is a
plugin that implements the `core.Plugin` interface and registers itself via
`init()`.

```
core.Plugin interface
  Name() string
  Version() string
  ShortDescription() string
  Command(ctx *core.Context) *cobra.Command
```

In the MVP, plugins are compiled into the binary. Future plan: migrate to
subprocess dispatch (binaries named `ctx-csharp`, `ctx-react` in PATH), same
pattern as `kubectl` plugins.

C# analysis uses a separate helper process (`tools/roslyn-helper/`) written in
C# with Roslyn. The helper communicates with `ctx` via JSON-RPC over
stdin/stdout. This lets us use the best tool for the job without pulling a .NET
runtime into the Go binary.

See `docs/DECISIONS.md` for the full rationale behind each architectural choice.

## Development

### Prerequisites
- Go 1.26+
- .NET SDK 10.0+ (for Roslyn helper)
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

### Build
```sh
go build -o ctx.exe ./cmd/ctx
```

### Tests & lint
```sh
go vet ./...
golangci-lint run ./...
```

### Pre-commit hooks
```sh
git config core.hooksPath .githooks
# On Linux/macOS also: chmod +x .githooks/pre-commit
```

### Roslyn helper
```sh
cd tools/roslyn-helper
dotnet publish src/RoslynHelper -c Release -r win-x64 --self-contained false -o publish/
```

## Contributing

Contributions welcome. See `CONTRIBUTING.md` (coming soon) for guidelines.
Open an issue first if you plan a large change.

## License

MIT — see [LICENSE](LICENSE).
