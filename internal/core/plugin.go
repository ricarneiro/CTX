package core

import "github.com/spf13/cobra"

// Plugin represents a ctx module (a stack or utility).
// The interface is designed to support future migration from compile-time
// registration (via init()) to subprocess dispatch (ctx-<name> binaries in PATH,
// kubectl-plugin style). Adding subprocess support later should not require
// changing this interface.
type Plugin interface {
	// Name returns the plugin identifier (e.g. "csharp", "react", "git").
	// Used as the subcommand name: ctx <name> ...
	Name() string

	// Version returns the semantic version of the plugin.
	Version() string

	// ShortDescription returns a one-line description shown in ctx --help.
	ShortDescription() string

	// Command returns the root cobra.Command for this plugin, with all
	// subcommands already configured. ctx adds this as a subcommand of root.
	Command(ctx *Context) *cobra.Command
}
