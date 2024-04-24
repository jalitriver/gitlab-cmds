// This file provides the implementation for the "projects list"
// command which optionally recursively lists projects in a group
// where the listed projects are selected by a regular expression.

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// ProjectsListOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// ProjectsListOptions are the options needed by this command.
type ProjectsListOptions struct {

	// Expr is the regular expression that filters the projects.
	// Defaults to "".
	Expr string `xml:"expr"`

	// Group for which projects will be listed.  Defaults to "".
	Group string `xml:"group"`

	// Recursive controls whether the projects are listed recursively.
	// Defaults to false.
	Recursive bool `xml:"recursive"`
}

// Initialize initializes this ProjectsListOptions instance so it can be
// used with the "flag" package to parse the command-line arguments.
func (opts *ProjectsListOptions) Initialize(flags *flag.FlagSet) {

	// --expr
	flags.StringVar(&opts.Expr, "expr", opts.Expr,
		"regular expression that selects projects to list")

	// --group
	flags.StringVar(&opts.Group, "group", opts.Group,
		"group to list")

	// -r
	flags.BoolVar(&opts.Recursive, "r", opts.Recursive,
		"whether to recursively list projects")

	// --recursive
	flags.BoolVar(&opts.Recursive, "recursive", opts.Recursive,
		"whether to recursively list projects")
}

////////////////////////////////////////////////////////////////////////
// ProjectsListCommand
////////////////////////////////////////////////////////////////////////

// ProjectsListCommand implements the "projects list" command which
// optionally recursively lists projects in a group where the listed
// projects are selected by a regular expression.
type ProjectsListCommand struct {

	// Embed the Command members.
	GitlabCommand[ProjectsListOptions]
}

// Usage prints the usage message to the output writer.  If err is not
// nil, it will be printed before the main output.
func (cmd *ProjectsListCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] projects list [subcmd_options]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    List projects recursively.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "List Options:\n")
	fmt.Fprintf(out, "\n")
	cmd.flags.SetOutput(out)
	cmd.flags.PrintDefaults()
	fmt.Fprintf(out, "\n")
	if out == os.Stderr {
		os.Exit(1)
	}
	os.Exit(0)
}

// NewProjectsListCommand returns a new and initialized ProjectsListCommand instance.
func NewProjectsListCommand(
	name string,
	opts *ProjectsListOptions,
	client *gitlab.Client,
) *ProjectsListCommand {

	// Create the new command.
	cmd := &ProjectsListCommand{
		GitlabCommand: GitlabCommand[ProjectsListOptions]{
			BasicCommand: BasicCommand[ProjectsListOptions]{
				name:    name,
				flags:   flag.NewFlagSet(name, flag.ExitOnError),
				options: opts,
			},
			client: client,
		},
	}

	// Set up the function that prints the global usage and exits.
	cmd.flags.Usage = func() { cmd.Usage(os.Stderr, nil) }

	// Initialize our command-line options.
	opts.Initialize(cmd.flags)

	return cmd
}

// Run is the entry point for this command.
func (cmd *ProjectsListCommand) Run(args []string) error {
	var err error

	// Parse command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Validate the options.
	if cmd.options.Group == "" {
		return fmt.Errorf("group not set")
	}

	// Print each project.
	return gitlab_util.ForEachProjectInGroup(
		cmd.client.Groups,
		cmd.options.Group,
		cmd.options.Expr,
		cmd.options.Recursive,
		func(g *gitlab.Group, p *gitlab.Project) (bool, error) {
			fmt.Printf("%v: %v\n", p.ID, p.PathWithNamespace)
			return true, nil
		})
}
