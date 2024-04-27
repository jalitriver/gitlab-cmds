// This file provides the implementation for the command
// "projects approval-rules list" which lists approval rules in all
// projects recursively found in a group where the projects are
// selected by a regular expression.

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// ProjectsApprovalRulesListOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// ProjectsApprovalRulesListOptions are the options needed by this command.
type ProjectsApprovalRulesListOptions struct {

	// Expr is the regular expression that filters the projects.
	// Defaults to "".
	Expr string `xml:"expr"`

	// Group for which projects will be listed.  Defaults to "".
	Group string `xml:"group"`

	// Recursive controls whether the projects are listed recursively.
	// Defaults to false.
	Recursive bool `xml:"recursive"`
}

// Initialize initializes this ProjectsApprovalRulesListOptions
// instance so it can be used with the "flag" package to parse the
// command-line arguments.
func (opts *ProjectsApprovalRulesListOptions) Initialize(flags *flag.FlagSet) {

	// --expr
	flags.StringVar(&opts.Expr, "expr", opts.Expr,
		"regular expression that selects projects for which approval "+
			"rules will be listed")

	// --group
	flags.StringVar(&opts.Group, "group", opts.Group,
		"group to list")

	// -r
	flags.BoolVar(&opts.Recursive, "r", opts.Recursive,
		"whether to recursively find projects")

	// --recursive
	flags.BoolVar(&opts.Recursive, "recursive", opts.Recursive,
		"whether to recursively find projects")
}

////////////////////////////////////////////////////////////////////////
// ProjectsApprovalRulesListCommand
////////////////////////////////////////////////////////////////////////

// ProjectsApprovalRulesListCommand implements the command
// "projects approval-rules list" which lists approval rules in all
// projects recursively found in a group where the projects are
// selected by a regular expression.
type ProjectsApprovalRulesListCommand struct {

	// Embed the Command members.
	GitlabCommand[ProjectsApprovalRulesListOptions]
}

// Usage prints the usage message to the output writer.  If err is not
// nil, it will be printed before the main output.
func (cmd *ProjectsApprovalRulesListCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] projects approval-rules list [subcmd_options]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    List approval rules on projects found recursively.\n")
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

// NewProjectsApprovalRulesListCommand returns a new, initialized
// ProjectsApprovalRulesListCommand instance.
func NewProjectsApprovalRulesListCommand(
	name string,
	opts *ProjectsApprovalRulesListOptions,
	client *gitlab.Client,
) *ProjectsApprovalRulesListCommand {

	// Create the new command.
	cmd := &ProjectsApprovalRulesListCommand{
		GitlabCommand: GitlabCommand[ProjectsApprovalRulesListOptions]{
			BasicCommand: BasicCommand[ProjectsApprovalRulesListOptions]{
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

func ApprovalRuleToString(rule *gitlab.ProjectApprovalRule) string {
	var result strings.Builder

	// Add rule ID and name.
	result.WriteString(fmt.Sprintf("%v: %v: ", rule.ID, rule.Name))

	// Iterate over the eligable approvers.
	result.WriteString("[")
	for i := 0; i < len(rule.EligibleApprovers); i++ {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("(%v, %v)",
			rule.EligibleApprovers[i].ID,
			rule.EligibleApprovers[i].Username))
	}
	result.WriteString("]")

	return result.String()
}

// Run is the entry point for this command.
func (cmd *ProjectsApprovalRulesListCommand) Run(args []string) error {
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

	// Print each approval rule for each project.
	return gitlab_util.ForEachProjectInGroup(
		cmd.client.Groups,
		cmd.options.Group,
		cmd.options.Expr,
		cmd.options.Recursive,
		func(g *gitlab.Group, p *gitlab.Project) (bool, error) {
			fmt.Printf("%v: %v\n", p.ID, p.PathWithNamespace)
			gitlab_util.ForEachApprovalRuleInProject(
				cmd.client.Projects, p,
				func(rule *gitlab.ProjectApprovalRule) (bool, error) {
					fmt.Printf("    %v\n", ApprovalRuleToString(rule))
					return true, nil
				})
			return true, nil
		})
}
