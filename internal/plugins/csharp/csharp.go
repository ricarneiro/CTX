// Package csharp implements the ctx csharp plugin.
// Full implementation: prompts 4–6 (requires Roslyn helper from prompt 3).
package csharp

import (
	"fmt"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&csharpPlugin{})
}

type csharpPlugin struct{}

func (c *csharpPlugin) Name() string             { return "csharp" }
func (c *csharpPlugin) Version() string          { return "0.0.1" }
func (c *csharpPlugin) ShortDescription() string { return "C# / .NET project analysis via Roslyn" }

func (c *csharpPlugin) Command(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "csharp",
		Short: c.ShortDescription(),
		// placeholder=true tells ctx auto to show a placeholder message instead of
		// attempting to invoke this plugin's subcommands.
		Annotations: map[string]string{"placeholder": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(ctx.Stderr, "Not implemented yet — coming in prompt 4")
			return nil
		},
	}
}
