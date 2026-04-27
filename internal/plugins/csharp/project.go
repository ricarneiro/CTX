package csharp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/ricarneiro/ctx/internal/plugins/csharp/helper"
	"github.com/spf13/cobra"
)

// errExit is a sentinel used to trigger os.Exit(1) via the SilenceErrors flow.
// The real error message has already been printed to ctx.Stderr.
var errExit = fmt.Errorf("exit 1")

func projectCmd(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "project",
		Short: "Summarize the .NET solution in compact markdown",
		Long: `Analyze the .NET solution in the current directory and emit a compact
markdown summary suitable for Claude Code consumption.

Requires the Roslyn helper (ctx-roslyn-helper) to be built and accessible.
Set CTX_ROSLYN_HELPER to the exact path, or build it alongside ctx.exe.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProject(ctx)
		},
	}
}

func runProject(ctx *core.Context) error {
	slnPath, err := findSolution(ctx.WorkDir, ctx)
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}

	client, err := helper.NewClient()
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}
	defer client.Close()

	if _, err := client.LoadSolution(slnPath); err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}

	summary, err := client.ProjectSummary()
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}

	return WriteSummary(ctx.Stdout, summary)
}

// findSolution locates a .sln file in dir.
// Falls back to a single .csproj if no .sln exists.
func findSolution(dir string, ctx *core.Context) (string, error) {
	slns, err := filepath.Glob(filepath.Join(dir, "*.sln"))
	if err != nil {
		return "", fmt.Errorf("glob .sln: %w", err)
	}

	// Filter to files that actually exist
	slns = existingFiles(slns)

	if len(slns) > 0 {
		sort.Strings(slns)
		if len(slns) > 1 {
			fmt.Fprintf(ctx.Stderr, "warning: multiple .sln files found, using %s\n", filepath.Base(slns[0]))
		}
		return slns[0], nil
	}

	// Fallback: single .csproj
	csprojPaths, err := filepath.Glob(filepath.Join(dir, "*.csproj"))
	if err != nil {
		return "", fmt.Errorf("glob .csproj: %w", err)
	}
	csprojPaths = existingFiles(csprojPaths)

	switch len(csprojPaths) {
	case 0:
		return "", fmt.Errorf("no .sln or .csproj found in %s", dir)
	case 1:
		return csprojPaths[0], nil
	default:
		return "", fmt.Errorf(
			"no .sln found and multiple .csproj files exist in %s\n"+
				"Run ctx inside a specific project folder, or create a .sln:\n"+
				"  dotnet new sln --format sln\n"+
				"  dotnet sln add **/*.csproj",
			dir,
		)
	}
}

func existingFiles(paths []string) []string {
	out := paths[:0]
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		}
	}
	return out
}
