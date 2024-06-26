// This file defines the common interfaces and structs used by commands.

package commands

import (
	"flag"
	"fmt"
	"slices"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Runner
////////////////////////////////////////////////////////////////////////

// Runner defines the interface for running commands.
type Runner interface {

	// Run runs the command as specified by its arguments.
	Run(args []string) error
}

////////////////////////////////////////////////////////////////////////
// BasicCommand
////////////////////////////////////////////////////////////////////////

// Command holds common data needed for each command.  The
// parameterized type T should be the Options struct for the command.
// For example, BasicCommand[FooOptions] configures this command to
// work with the options for the "foo" command.  Also see
// [GitlabCommand] and [ParentCommand].
type BasicCommand[T any] struct {

	// Name is the name of this command.
	name string

	// Flagset is used for parsing the command-line flags specific to
	// this command.
	flags *flag.FlagSet

	// Options are the options that control how the command runs.
	// Note that it is tempting to embed the options directly in this
	// struct or even to allocate the options on the heap.  However,
	// the way it works is the options are embedded in the single
	// large "Options" data structure in global_command.go so that all
	// of the options can be read from a single options.xml file.
	// Thus, this pointer is actually just a pointer into the large
	// "Options" data structure in global_command.go.
	options *T
}

////////////////////////////////////////////////////////////////////////
// GitlabCommand
////////////////////////////////////////////////////////////////////////

// GitlabCommand is a BasicCommand with a Gitlab communications
// client.  The parameterized type T should be the Options struct for
// the command.  For example, GitlabCommand[ProjectListOptions]
// configures this command to work with the options for the "project
// list" command.
type GitlabCommand[T any] struct {

	// Embed BasicCommand members.
	BasicCommand[T]

	// client is the Gitlab communications client
	client *gitlab.Client
}

////////////////////////////////////////////////////////////////////////
// ParentCommand
////////////////////////////////////////////////////////////////////////

// ParentCommand is a BasicCommand with a subcommand map that maps the
// name of subcommands to their Runner.  The parameterized type T
// should be the Options struct for the command.  For example,
// ParentCommand[ProjectOptions] configures this command to work with
// the options for the "project" command.
type ParentCommand[T any] struct {

	// Embed BasicCommand members.
	BasicCommand[T]

	// subcmds maps from command name to Runner for the command
	subcmds map[string]Runner
}

// DispatchSubcommand dispatches the subcommand specified by the name
// args[0] using the remaining arguments are arguments for the
// subcommand.
func (p *ParentCommand[T]) DispatchSubcommand(args []string) error {

	// Determine which subcommand the user specified.
	if len(args) < 1 {
		return fmt.Errorf("no subcommand specified")
	}
	subcmd := args[0]

	// Find the runner for the subcommand.
	runner, ok := p.subcmds[subcmd]
	if !ok {
		return fmt.Errorf("invalid subcommand: %s", subcmd)
	}

	// Run the subcommand.
	return runner.Run(args[1:])
}

// SortedCommandNames returns a slice that holds the sorted command names.
func (cmd *ParentCommand[T]) SortedCommandNames() []string {

	var result []string

	// Collect all the command names.
	for k := range cmd.subcmds {
		result = append(result, k)
	}

	// Sort the keys.
	slices.Sort(result)

	return result
}
