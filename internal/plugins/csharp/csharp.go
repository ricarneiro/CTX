// Package csharp implements the ctx csharp plugin — .NET solution analysis via Roslyn helper.
package csharp

import (
	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&csharpPlugin{})
}

type csharpPlugin struct{}

func (c *csharpPlugin) Name() string             { return "csharp" }
func (c *csharpPlugin) Version() string          { return "0.1.0" }
func (c *csharpPlugin) ShortDescription() string { return "C# / .NET project analysis via Roslyn" }

func (c *csharpPlugin) Command(ctx *core.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csharp",
		Short: c.ShortDescription(),
		Long: `Analyze C# / .NET projects and emit compact markdown summaries
optimized for Claude Code consumption.

Requires the Roslyn helper (ctx-roslyn-helper) to be built.
See 'ctx csharp project --help' for details.`,
	}
	cmd.AddCommand(projectCmd(ctx))
	return cmd
}
