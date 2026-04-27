// Package auto implements the ctx auto plugin — stack detection and routing.
package auto

import (
	"fmt"
	"strings"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	core.Register(&plugin{})
}

type plugin struct{}

func (p *plugin) Name() string             { return "auto" }
func (p *plugin) Version() string          { return "0.1.0" }
func (p *plugin) ShortDescription() string { return "Detect stack and route to appropriate plugin" }

func (p *plugin) Command(ctx *core.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto",
		Short: p.ShortDescription(),
		Long:  "Auto-detect the project stack and route to the appropriate ctx plugin.",
	}
	cmd.AddCommand(newDetectCmd(ctx))
	cmd.AddCommand(newProjectCmd(ctx))
	return cmd
}

// --- detect subcommand ---

func newDetectCmd(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "List detected stacks in the current directory",
		Long:  "Scan the current directory and list detected technology stacks without running any plugin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDetect(ctx)
		},
	}
}

func runDetect(ctx *core.Context) error {
	stacks, err := Detect(ctx.WorkDir)
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return fmt.Errorf("exit 1")
	}
	formatDetect(ctx, stacks)
	return nil
}

func formatDetect(ctx *core.Context, stacks []Stack) {
	w := ctx.Stdout
	fmt.Fprintln(w, "# Stack detection")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "**directory:** %s\n\n", ctx.WorkDir)

	if len(stacks) == 0 {
		fmt.Fprintln(w, "No known stack detected.")
		fmt.Fprintln(w)
		plugins := core.All()
		names := make([]string, len(plugins))
		for i, p := range plugins {
			names[i] = p.Name()
		}
		fmt.Fprintf(w, "Available plugins: %s\n", strings.Join(names, ", "))
		fmt.Fprintln(w, "Run `ctx <plugin> --help` to see what each plugin offers.")
		return
	}

	fmt.Fprintln(w, "## Detected stacks")
	for _, s := range stacks {
		fmt.Fprintf(w, "- **%s** (confidence: %s)\n", s.Name, s.Confidence)
		for _, e := range s.Evidence {
			fmt.Fprintf(w, "  - %s\n", e)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## Suggested commands")
	for _, s := range stacks {
		fmt.Fprintf(w, "- `ctx %s project`\n", s.Name)
	}
}

// --- project subcommand ---

func newProjectCmd(ctx *core.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "project",
		Short: "Detect stack and emit project context summary",
		Long:  "Auto-detect the project stack and run the `project` command of each matching plugin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProject(ctx)
		},
	}
}

func runProject(ctx *core.Context) error {
	stacks, err := Detect(ctx.WorkDir)
	if err != nil {
		fmt.Fprintln(ctx.Stderr, err.Error())
		return fmt.Errorf("exit 1")
	}

	if len(stacks) == 0 {
		formatDetect(ctx, stacks)
		return nil
	}

	first := true
	for _, stack := range stacks {
		if !first {
			fmt.Fprintln(ctx.Stdout, "---")
			fmt.Fprintln(ctx.Stdout)
		}
		first = false

		p := core.Get(stack.Name)
		if p == nil {
			fmt.Fprintf(ctx.Stdout,
				"# %s\n\nNo plugin registered for stack `%s`.\n\n",
				stack.Name, stack.Name)
			continue
		}

		pluginCmd := p.Command(ctx)

		// If the root command itself is a placeholder, the whole plugin is unimplemented.
		if pluginCmd.Annotations["placeholder"] == "true" {
			writePlaceholder(ctx, stack.Name)
			continue
		}

		sub := findSubcommand(pluginCmd, "project")
		if sub == nil {
			fmt.Fprintf(ctx.Stdout,
				"# %s\n\nThe `%s` plugin does not have a `project` command.\n\n",
				stack.Name, stack.Name)
			continue
		}

		if sub.Annotations["placeholder"] == "true" {
			writePlaceholder(ctx, stack.Name)
			continue
		}

		// Execute the project subcommand directly (no cobra routing overhead).
		if sub.RunE != nil {
			if err := sub.RunE(sub, []string{}); err != nil {
				fmt.Fprintf(ctx.Stderr, "error running %s project: %v\n", stack.Name, err)
			}
		} else if sub.Run != nil {
			sub.Run(sub, []string{})
		}
	}
	return nil
}

// findSubcommand returns the named subcommand of cmd, or nil if not found.
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, sub := range cmd.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}

func writePlaceholder(ctx *core.Context, name string) {
	fmt.Fprintf(ctx.Stdout,
		"# %s (placeholder)\n\nThe `%s` plugin detected this stack but its `project` command is not yet implemented.\nThis will be available in a future version of ctx.\n\n",
		name, name)
}
