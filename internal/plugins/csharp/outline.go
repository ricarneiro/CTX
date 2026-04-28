package csharp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/ricarneiro/ctx/internal/plugins/csharp/helper"
	"github.com/spf13/cobra"
)

func outlineCmd(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "outline <file.cs>",
		Short: "Show structural outline of a C# file (no method bodies)",
		Long: `Parse a C# source file and emit its structural skeleton:
namespaces, types, method signatures, properties, fields, events.
Method bodies are omitted — reduces large files by 80–90%.

Does not require a loaded solution. Works on a single file.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOutline(ctx, args[0])
		},
	}
}

func runOutline(ctx *core.Context, file string) error {
	abs, err := resolveFilePath(ctx.WorkDir, file)
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}

	if !strings.HasSuffix(strings.ToLower(abs), ".cs") {
		fmt.Fprintf(ctx.Stderr, "not a C# file: %s\n", abs)
		return errExit
	}

	if _, statErr := os.Stat(abs); statErr != nil {
		fmt.Fprintf(ctx.Stderr, "file not found: %s\n", abs)
		return errExit
	}

	client, err := helper.NewClient()
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}
	defer client.Close()

	outline, err := client.Outline(abs)
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return errExit
	}

	return WriteOutline(ctx.Stdout, outline)
}

func resolveFilePath(workDir, file string) (string, error) {
	if filepath.IsAbs(file) {
		return filepath.Clean(file), nil
	}
	return filepath.Clean(filepath.Join(workDir, file)), nil
}
