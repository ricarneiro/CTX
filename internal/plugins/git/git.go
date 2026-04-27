// Package git implements the ctx git plugin.
// Full implementation: prompt 1.
package git

import (
	"fmt"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&gitPlugin{})
}

type gitPlugin struct{}

func (g *gitPlugin) Name() string             { return "git" }
func (g *gitPlugin) Version() string          { return "0.0.1" }
func (g *gitPlugin) ShortDescription() string { return "Git repository summary for Claude" }

func (g *gitPlugin) Command(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "git",
		Short: g.ShortDescription(),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(ctx.Stderr, "Not implemented yet — coming in prompt 1")
			return nil
		},
	}
}
