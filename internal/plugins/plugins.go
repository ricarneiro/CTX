// Package plugins imports all built-in plugins to trigger their init()
// registration with the core registry. Import this package with a blank
// identifier in cmd/ctx/main.go.
package plugins

import (
	_ "github.com/ricarneiro/ctx/internal/plugins/auto"
	_ "github.com/ricarneiro/ctx/internal/plugins/csharp"
	_ "github.com/ricarneiro/ctx/internal/plugins/git"
	_ "github.com/ricarneiro/ctx/internal/plugins/react"
)
