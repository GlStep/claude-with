package main

import (
	"os"

	"github.com/glstep/claude-with/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
