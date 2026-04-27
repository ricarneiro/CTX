package cli

import (
	"fmt"
	"strings"

	"github.com/ricarneiro/ctx/internal/core"
	"github.com/spf13/cobra"
)

// Version, Commit, and BuildDate are injected at build time via ldflags:
//
//	go build -ldflags "-X github.com/ricarneiro/ctx/internal/cli.Version=1.0.0 \
//	                   -X github.com/ricarneiro/ctx/internal/cli.Commit=abc1234 \
//	                   -X github.com/ricarneiro/ctx/internal/cli.BuildDate=2025-01-01"
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.SetVersionTemplate("{{ .Version }}\n")
	// rootCmd.Version is set lazily in Execute() after all plugins are
	// registered, so that versionString() can read the full plugin list.
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print ctx version and compiled plugins",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionString())
	},
}

// versionString builds the full version line including plugin list.
func versionString() string {
	line1 := fmt.Sprintf("ctx %s (commit %s, built %s)", Version, Commit, BuildDate)

	plugins := core.All()
	if len(plugins) == 0 {
		return line1 + "\nplugins: none"
	}

	parts := make([]string, len(plugins))
	for i, p := range plugins {
		parts[i] = fmt.Sprintf("%s@%s", p.Name(), p.Version())
	}
	return line1 + "\nplugins: " + strings.Join(parts, ", ")
}
