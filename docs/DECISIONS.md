# Architecture Decision Records — ctx

Decisions are recorded here as they are made. Each entry has context (why we
needed to decide), decision (what we chose), and consequences (what it implies).

---

## 1. Language: Go (not C# or TypeScript)

**Context:** ctx is a CLI tool that runs locally on developer machines. It
needs to start fast (< 100ms cold start), produce a single distributable
binary, and work on macOS, Linux, and Windows without asking the user to
install a runtime.

**Decision:** Go.

**Consequences:**
- Single static binary, trivial cross-compilation (`GOOS=linux go build`).
- No runtime install required on target machines.
- Go's stdlib covers file I/O, JSON, HTTP, process spawning — no heavy deps.
- Cannot use Roslyn directly for C# analysis (Go has no decent C# parser).
  Handled by a separate helper process (see decision 4).

---

## 2. CLI framework: Cobra

**Context:** Need subcommands (`ctx git`, `ctx csharp outline`), flags, help
text, and shell completion. Rolling our own is wasteful.

**Decision:** `github.com/spf13/cobra`. Optionally `viper` for config in the
future, but not added yet.

**Consequences:**
- Cobra is the de facto standard (`kubectl`, `gh`, `hugo`, `helm` all use it).
- Well-understood by contributors. Good docs. Stable API.
- Shell completion (bash/zsh/fish/PowerShell) comes for free.

---

## 3. Plugin system: compile-time in MVP, subprocess later

**Context:** ctx needs to support multiple stacks. Plugins must be modular.
Two valid approaches: (a) compile all plugins into one binary, (b) dispatch
to external binaries (`ctx-csharp`, `ctx-react`) like kubectl.

**Decision:** MVP uses compile-time plugins registered via `init()`. The
`core.Plugin` interface is designed so subprocess plugins can implement it
later without changing callers.

**Consequences:**
- Simpler to build and ship in early stages.
- All plugins must be written in Go (or wrap external tools).
- Subprocess migration is planned but deferred until we need third-party
  plugins. At that point: scan PATH for `ctx-*` binaries, wrap them with a
  shim that implements `core.Plugin`.

---

## 4. C# analysis: separate Roslyn helper process

**Context:** Semantic C# analysis (types, method signatures, usages, errors)
requires a compiler-level parser. Roslyn is the only production-quality option.
Go has no usable C# parser.

**Decision:** A separate C# program (`tools/roslyn-helper/`) that loads
projects with Roslyn and answers queries sent as JSON-RPC messages over
stdin/stdout. ctx spawns it as a subprocess and communicates via pipes.

**Consequences:**
- Requires .NET SDK on the machine for C# commands (acceptable — users of
  `ctx csharp` are already .NET developers).
- Clean separation: Go handles CLI, orchestration, output formatting; C#
  handles semantic analysis.
- JSON-RPC over stdin/stdout is trivial to test in isolation.
- Performance: subprocess is spawned once per ctx invocation, amortized if we
  batch queries (future optimization).

---

## 5. Output format: UTF-8 markdown without BOM on stdout

**Context:** ctx output is consumed by Claude Code (via pipe or CLAUDE.md
instructions). Claude understands markdown well. BOM causes issues in some
contexts.

**Decision:** All plugin output is plain UTF-8 markdown, no BOM, on stdout.
Errors go to stderr. No ANSI colors (markdown already has structure).

**Consequences:**
- Output is directly pasteable into Claude conversations.
- Easy to redirect to file: `ctx csharp project > context.md`.
- Plugins must use `ctx.Stdout` / `ctx.Stderr`, never `fmt.Println` directly,
  so output destination can be overridden in tests.

---

## 6. Subcommand structure: ctx <stack> <command>

**Context:** Multiple stacks, each with multiple commands. Need a consistent,
scalable interface.

**Decision:** `ctx <stack> <command> [args] [flags]`. Examples:
- `ctx csharp project`
- `ctx csharp outline src/Foo.cs`
- `ctx git`
- `ctx auto project`

**Consequences:**
- Adding a new stack = adding a new top-level subcommand.
- Commands within a stack are subcommands of the stack command.
- Consistent with kubectl, gh, and other modern CLIs.

---

## 7. Module path: github.com/ricarneiro/ctx

**Context:** Source is hosted on a private Gitea instance
(`git.carneiro.ddnsfree.com`). Go module paths do not need to match the actual
git host. If we want `go install` from the public GitHub mirror later, the
module path must match.

**Decision:** `github.com/ricarneiro/ctx` from day one.

**Consequences:**
- Developers cloning from Gitea still use this module path — no friction.
- `go install github.com/ricarneiro/ctx/cmd/ctx@latest` works once we push
  to GitHub. No rename needed.
- If we never publish to GitHub, the module path is just an identifier and
  causes no problems.

---

## 8. License: MIT

**Context:** Want maximum adoption. No patent clause complexity.

**Decision:** MIT, copyright Ricardo Carneiro.

**Consequences:**
- Anyone can use, fork, embed, sell without restrictions.
- No copyleft. If this matters later, we can dual-license, but that's a
  future problem.

---

## 9. Commit convention: Conventional Commits

**Context:** Want automated CHANGELOG generation in the future. Need a
consistent commit format from day one.

**Decision:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`,
`refactor:`, `test:`). Enforced by convention, not by hook (yet).

**Consequences:**
- `git log --oneline` is readable.
- Can run `git-cliff` or `conventional-changelog` later to generate CHANGELOG.
- PRs should squash to a single conventional commit on merge.
