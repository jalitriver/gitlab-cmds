// This file provides the implementation for the "project delete"
// command which optionally deletes projects recursively (or not)
// whose name matchs a regular expression.

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
// ProjectDeleteOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// ProjectDeleteOptions are the options needed by this command.
type ProjectDeleteOptions struct {

	// DryRun should cause the command to print what it would do
	// instead of actually doing it.
	DryRun bool `xml:"dry-run"`

	// Expr is the regular expression that filters the projects.
	Expr string `xml:"expr"`

	// Group for which projects will be listed.
	Group string `xml:"group"`

	// Recursive controls whether the projects are deleted recursively.
	Recursive bool `xml:"recursive"`
}

// Initialize initializes this ProjectDeleteOptions instance so it can be
// used with the "flag" package to parse the command-line arguments.
func (opts *ProjectDeleteOptions) Initialize(flags *flag.FlagSet) {

	// -n
	flag.BoolVar(&opts.DryRun, "n", opts.DryRun,
		"print what it would do instead of actually doing it")

	// --dry-run
	flag.BoolVar(&opts.DryRun, "dry-run", opts.DryRun,
		"print what it would do instead of actually doing it")

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
// ProjectDeleteCommand
////////////////////////////////////////////////////////////////////////

// ProjectDeleteCommand implements the "project delete" command which
// optionally recursively deletes projects in a group where the
// deleted projects are selected by a regular expression.
type ProjectDeleteCommand struct {

	// Embed the Command members.
	GitlabCommand[ProjectDeleteOptions]
}

// Usage prints the usage message to the output writer.  If err is not
// nil, it will be printed before the main output.
func (cmd *ProjectDeleteCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] project delete [subcmd_options]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    Deletes projects recursively.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Delete Options:\n")
	fmt.Fprintf(out, "\n")
	cmd.flags.SetOutput(out)
	cmd.flags.PrintDefaults()
	fmt.Fprintf(out, "\n")
	if out == os.Stderr {
		os.Exit(1)
	}
	os.Exit(0)
}

// NewProjectDeleteCommand returns a new and initialized ProjectDeleteCommand instance.
func NewProjectDeleteCommand(
	name string,
	opts *ProjectDeleteOptions,
	client *gitlab.Client,
) *ProjectDeleteCommand {

	// Create the new command.
	cmd := &ProjectDeleteCommand{
		GitlabCommand: GitlabCommand[ProjectDeleteOptions]{
			BasicCommand: BasicCommand[ProjectDeleteOptions]{
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

// DeleteProject deletes the project.  If dryRun is true, this
// function only prints what it would without actually doing it.
func DeleteProject(
	s *gitlab.ProjectsService,
	p *gitlab.Project,
	dryRun bool,
) error {
	fmt.Printf("- Deleting project: %q ... ", p.PathWithNamespace)
	if !dryRun {
		_, err := s.DeleteProject(p.ID)
		if err != nil {
			return fmt.Errorf("DeleteProject: %w", err)
		}
	}
	fmt.Printf("Done.\n")
	return nil
}

// DeleteProjects deletes all the projects in a group (recursively or
// not) for each project whose full path name matches the regular
// expression.  An empty regular expression matches any string.  If
// dryRun is true, this function only prints what it would without
// actually doing it.
func DeleteProjects(
	client *gitlab.Client,
	group string,
	expr string,
	recursive bool,
	dryRun bool,
) error {

	// Collect projects.
	fmt.Printf("- Collecting projects ... ")
	projects, err := gitlab_util.GetAllProjects(
		client.Groups, group, expr, recursive)
	if err != nil {
		return fmt.Errorf("DeleteProjects: %w", err)
	}
	fmt.Printf("Done.\n")

	// Delete projects.
	for _, project := range projects {
		err = DeleteProject(client.Projects, project, dryRun)
		if err != nil {
			return fmt.Errorf("DeleteProjects: %w", err)
		}
	}

	return nil
}

// Run is the entry point for this command.
func (cmd *ProjectDeleteCommand) Run(args []string) error {
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

	// Delete projects.
	return DeleteProjects(
		cmd.client,
		cmd.options.Group,
		cmd.options.Expr,
		cmd.options.Recursive,
		cmd.options.DryRun)
}
