package cli

import (
	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

// registerPluginCommands iterates the core registry and adds each plugin's
// root command as a subcommand of rootCmd. Called once from Execute().
func registerPluginCommands(rootCmd *cobra.Command) {
	ctx := newContext()
	for _, p := range core.All() {
		rootCmd.AddCommand(p.Command(ctx))
	}
}
