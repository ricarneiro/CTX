package cli

import (
	"context"
	"os"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ctx",
	Short: "Anti-tokens CLI for Claude Code",
	Long: `ctx reduces token consumption in Claude Code sessions by analyzing your
codebase locally and emitting dense markdown summaries that Claude consumes
instead of reading dozens of raw source files.

Each subcommand targets a specific stack:
  ctx git       — git log, status and diff summary
  ctx auto      — auto-detect project stack and emit context
  ctx csharp    — C# / .NET analysis via Roslyn
  ctx react     — React / TypeScript analysis

Output is always UTF-8 markdown on stdout, suitable for piping into Claude.`,
	SilenceUsage:  true,
	SilenceErrors: true, // plugins print their own errors to stderr
}

// Execute runs the root command. Called by cmd/ctx/main.go.
func Execute() {
	registerPluginCommands(rootCmd)
	// Set version after plugins are registered so versionString() sees them all.
	rootCmd.Version = versionString()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// newContext builds the core.Context injected into plugins at runtime.
func newContext() *core.Context {
	wd, _ := os.Getwd()
	return &core.Context{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		WorkDir: wd,
		Verbose: false,
		Ctx:     context.Background(),
	}
}
