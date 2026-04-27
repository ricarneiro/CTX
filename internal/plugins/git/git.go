// Package git implements the ctx git plugin — compact git state summary.
package git

import (
	"fmt"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&plugin{})
}

type plugin struct{}

func (p *plugin) Name() string             { return "git" }
func (p *plugin) Version() string          { return "0.1.0" }
func (p *plugin) ShortDescription() string { return "Summarize git state in compact markdown" }

func (p *plugin) Command(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "git",
		Short: p.ShortDescription(),
		Long: `Emit a compact markdown summary of the current git repository:
branch, ahead/behind upstream, recent commits, working tree changes,
and a diff summary grouped by directory.

Output goes to stdout. Pipe it into Claude or save it to a file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx)
		},
	}
}

func run(ctx *core.Context) error {
	data, err := collect(ctx.WorkDir)
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return fmt.Errorf("exit 1") // non-nil → os.Exit(1); SilenceErrors suppresses print
	}
	formatOutput(ctx.Stdout, data)
	return nil
}
