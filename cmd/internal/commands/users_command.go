// This file provides the implementation for the "users" command
// which provides user related subcommands.
//
// If you need to add a new subcommand, do the following:
//
//   1) Create the new subcommand similar to
//      cmd/internal/commands/projects_command.go if the subcommand
//      will have its own set of subcommands or similar to
//      cmd/internal/commands/projects_list_command.go if the
//      subcommand will actually do something.
//
//   2) Add the resulting new options struct to the Options struct
//      below so the options can also be specified in the options.xml
//      file.
//
//   3) Add the new subcommand as demonstrated in
//      UsersCommand.addSubcmds().

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// UsersOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// UsersOptions are the options needed by this command.
type UsersOptions struct {
	UsersListOpts UsersListOptions `xml:"list-options"`
}

// Initialize initializes this UsersOptions instance so it can be
// used with the "flag" package to parse the command-line arguments.
func (opts *UsersOptions) Initialize(flags *flag.FlagSet) {
	// empty
}

////////////////////////////////////////////////////////////////////////
// UsersCommand
////////////////////////////////////////////////////////////////////////

// UsersCommand provides subcommands for Gitlab project related
// maintenance.
type UsersCommand struct {

	// Embed the Command members.
	ParentCommand[UsersOptions]
}

// Usage prints the main usage message to the output writer.  If
// err is not nil, it will be printed before the main output.
func (cmd *UsersCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] users [subcmd]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    Command for administering a Gitlab users.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Subcommands:\n")
	fmt.Fprintf(out, "\n")
	for _, subcmd := range cmd.SortedCommandNames() {
		fmt.Fprintf(out, "  %s\n", subcmd)
	}
	fmt.Fprintf(out, "\n")
	if out == os.Stderr {
		os.Exit(1)
	}
	os.Exit(0)
}

// addSubcmds adds the subcommands for this command.
func (cmd *UsersCommand) addSubcmds(client *gitlab.Client) {
	cmd.subcmds["list"] = NewUsersListCommand(
		"list", &cmd.options.UsersListOpts, client)
}

// NewUsersCommand returns a new, initialized UsersCommand
// instance having the specified name.
func NewUsersCommand(
	name string,
	opts *UsersOptions,
	client *gitlab.Client,
) *UsersCommand {

	// Create the new command.
	cmd := &UsersCommand{
		ParentCommand: ParentCommand[UsersOptions]{
			BasicCommand: BasicCommand[UsersOptions]{
				name:    name,
				flags:   flag.NewFlagSet(name, flag.ExitOnError),
				options: opts,
			},
			subcmds: make(map[string]Runner),
		},
	}

	// Set up the function that prints the global usage and exits.
	cmd.flags.Usage = func() { cmd.Usage(os.Stderr, nil) }

	// Initialize our command-line options.
	cmd.options.Initialize(cmd.flags)

	// Add the subcommands.
	cmd.addSubcmds(client)

	return cmd
}

// Run is the entry point for this command.
func (cmd *UsersCommand) Run(args []string) error {
	var err error

	// Parse command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Dispatch the subcommand specified by the remaining arguments.
	return cmd.DispatchSubcommand(cmd.flags.Args())
}
