// Package auto implements the ctx auto plugin.
// Full implementation: prompt 2.
package auto

import (
	"fmt"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&autoPlugin{})
}

type autoPlugin struct{}

func (a *autoPlugin) Name() string             { return "auto" }
func (a *autoPlugin) Version() string          { return "0.0.1" }
func (a *autoPlugin) ShortDescription() string { return "Auto-detect project stack and emit context" }

func (a *autoPlugin) Command(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "auto",
		Short: a.ShortDescription(),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(ctx.Stderr, "Not implemented yet — coming in prompt 2")
			return nil
		},
	}
}
