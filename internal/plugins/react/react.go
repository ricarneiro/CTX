// Package react implements the ctx react plugin.
// Full implementation: phase 4 (future prompts).
package react

import (
	"fmt"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&reactPlugin{})
}

type reactPlugin struct{}

func (r *reactPlugin) Name() string             { return "react" }
func (r *reactPlugin) Version() string          { return "0.0.1" }
func (r *reactPlugin) ShortDescription() string { return "React / TypeScript project analysis" }

func (r *reactPlugin) Command(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "react",
		Short: r.ShortDescription(),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(ctx.Stderr, "Not implemented yet — coming in a future prompt")
			return nil
		},
	}
}
