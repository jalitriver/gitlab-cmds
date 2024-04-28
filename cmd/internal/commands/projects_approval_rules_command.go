// This file provides the implementation for the "projects approval-rules"
// command which provides project related subcommands.
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
//      ProjectsCommand.addSubcmds().

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
// ProjectsApprovalRulesOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// ProjectsApprovalRulesOptions are the options needed by this command.
type ProjectsApprovalRulesOptions struct {

	// Options for the "projects approval-rules list" command.
	ProjectsApprovalRulesListOpts ProjectsApprovalRulesListOptions `xml:"list-options"`

	// Options for the "projects approval-rules update" command.
	ProjectsApprovalRulesUpdateOpts ProjectsApprovalRulesUpdateOptions `xml:"update-options"`
}

// Initialize initializes this ProjectsApprovalRulesOptions instance so it can be
// used with the "flag" package to parse the command-line arguments.
func (opts *ProjectsApprovalRulesOptions) Initialize(flags *flag.FlagSet) {
	// empty
}

////////////////////////////////////////////////////////////////////////
// ProjectsApprovalRulesCommand
////////////////////////////////////////////////////////////////////////

// ProjectsApprovalRulesCommand provides subcommands for Gitlab project related
// maintenance.
type ProjectsApprovalRulesCommand struct {

	// Embed the Command members.
	ParentCommand[ProjectsApprovalRulesOptions]
}

// Usage prints the main usage message to the output writer.  If
// err is not nil, it will be printed before the main output.
func (cmd *ProjectsApprovalRulesCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] projects approval-rules [subcmd]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    Command for administering approval rules for Gitlab projects.\n")
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
func (cmd *ProjectsApprovalRulesCommand) addSubcmds(client *gitlab.Client) {
	cmd.subcmds["list"] = NewProjectsApprovalRulesListCommand(
		"list", &cmd.options.ProjectsApprovalRulesListOpts, client)
	cmd.subcmds["update"] = NewProjectsApprovalRulesUpdateCommand(
		"update", &cmd.options.ProjectsApprovalRulesUpdateOpts, client)
}

// NewProjectsApprovalRulesCommand returns a new, initialized
// ProjectsApprovalRulesCommand instance having the specified name.
func NewProjectsApprovalRulesCommand(
	name string,
	opts *ProjectsApprovalRulesOptions,
	client *gitlab.Client,
) *ProjectsApprovalRulesCommand {

	// Create the new command.
	cmd := &ProjectsApprovalRulesCommand{
		ParentCommand: ParentCommand[ProjectsApprovalRulesOptions]{
			BasicCommand: BasicCommand[ProjectsApprovalRulesOptions]{
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
func (cmd *ProjectsApprovalRulesCommand) Run(args []string) error {
	var err error

	// Parse command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Dispatch the subcommand specified by the remaining arguments.
	return cmd.DispatchSubcommand(cmd.flags.Args())
}
