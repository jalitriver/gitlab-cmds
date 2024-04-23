// cmd/glcli/main.go is the main entry point for the glcli program.
// It provides (mostly recursive) convenience functions for managing a
// Gitlab server.  It works similar to the "aws" CLI with subcommands.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/commands"
)

var (
	version = "0.0.1"
)

func main() {
	var err error

	// Find the base name for the executable.
	basename := filepath.Base(os.Args[0])

	// Create the GlobalCommand which is the parent of all other commands.
	globalCmd := commands.NewGlobalCommand(basename, version)

	// Invoke the global command.
	err = globalCmd.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n*** Error: %v\n\n", err)
		os.Exit(1)
	}
}
