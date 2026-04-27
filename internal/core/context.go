package core

import (
	"context"
	"io"
)

// Context is passed to plugins when their command is executed.
// Allows ctx to inject dependencies and configuration without coupling
// plugins to its internals. Mirrors the pattern used by kubectl plugins.
type Context struct {
	// Stdout is where plugins write their primary markdown output.
	Stdout io.Writer

	// Stderr is where plugins write errors and diagnostic messages.
	Stderr io.Writer

	// WorkDir is the directory where ctx was invoked.
	WorkDir string

	// Verbose enables extra diagnostic output when true.
	Verbose bool

	// Ctx carries deadlines and cancellation signals.
	Ctx context.Context
}
