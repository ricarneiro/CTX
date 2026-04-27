package main

import (
	"github.com/ricarneiro/ctx/internal/cli"
	_ "github.com/ricarneiro/ctx/internal/plugins"
)

func main() {
	cli.Execute()
}
