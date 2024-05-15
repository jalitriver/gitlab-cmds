// This is the main entry point for the program which is designed
// around a command-line interface that accepts subcommands similar to
// how "aws" and "git" work.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/commands"
)

var (
	version = "0.0.2"
)

func main() {
	var err error

	// Sanity check.
	if len(os.Args) < 1 {
		fmt.Fprintf(
			os.Stderr,
			"\n*** Error: invalid command-line arguments: %v\n\n",
			os.Args)
		os.Exit(1)
	}

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
