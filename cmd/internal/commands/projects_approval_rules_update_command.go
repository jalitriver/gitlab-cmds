// This file provides the implementation for the command "projects
// approval-rules update" which updates approval rules in all projects
// recursively found in a group where the projects are selected by a
// regular expression.

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/xml_users"
	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// ProjectsApprovalRulesUpdateOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// ProjectsApprovalRulesUpdateOptions are the options needed by this command.
type ProjectsApprovalRulesUpdateOptions struct {

	// ApproversFileName is the name of the XML file holding the list
	// of allowed approvers which should contain the output of the
	// "glmcds users list" command which is the serialization of an
	// [xml_users.XmlUsers] instance.
	ApproversFileName string `xml:"approvers-file-name"`

	// DryRun should cause the command to print what it would do
	// instead of actually doing it.  Defaults to false.
	DryRun bool `xml:"dry-run"`

	// Expr is the regular expression that filters the projects.
	// Defaults to "".
	Expr string `xml:"expr"`

	// Group for which projects will be updated.  Defaults to "".
	Group string `xml:"group"`

	// Recursive controls whether the projects are found recursively.
	// Defaults to false.
	Recursive bool `xml:"recursive"`
}

// Initialize initializes this ProjectsApprovalRulesUpdateOptions
// instance so it can be used with the "flag" package to parse the
// command-line arguments.
func (opts *ProjectsApprovalRulesUpdateOptions) Initialize(flags *flag.FlagSet) {

	// --approvers
	flags.StringVar(&opts.ApproversFileName, "approvers", opts.ApproversFileName,
		"name of the XML file holding the list of allowed approvers which "+
			"should contain the output of the \"glmcds users list\" command")

	// -n
	flags.BoolVar(
		&opts.DryRun, "n", opts.DryRun,
		"print what it would do instead of actually doing it")

	// --dry-run
	flags.BoolVar(&opts.DryRun, "dry-run", opts.DryRun,
		"print what it would do instead of actually doing it")

	// --expr
	flags.StringVar(&opts.Expr, "expr", opts.Expr,
		"regular expression that selects projects for which approval "+
			"rules will be updated")

	// --group
	flags.StringVar(&opts.Group, "group", opts.Group,
		"group to update")

	// -r
	flags.BoolVar(&opts.Recursive, "r", opts.Recursive,
		"whether to recursively find projects")

	// --recursive
	flags.BoolVar(&opts.Recursive, "recursive", opts.Recursive,
		"whether to recursively find projects")
}

////////////////////////////////////////////////////////////////////////
// ProjectsApprovalRulesUpdateCommand
////////////////////////////////////////////////////////////////////////

// ProjectsApprovalRulesUpdateCommand implements the command "projects
// approval-rules update" which updates approval rules in all projects
// recursively found in a group where the projects are selected by a
// regular expression.
type ProjectsApprovalRulesUpdateCommand struct {

	// Embed the Command members.
	GitlabCommand[ProjectsApprovalRulesUpdateOptions]
}

// Usage prints the usage message to the output writer.  If err is not
// nil, it will be printed before the main output.
func (cmd *ProjectsApprovalRulesUpdateCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] projects approval-rules update [subcmd_options]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    Update approval rules on projects found recursively.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Update Options:\n")
	fmt.Fprintf(out, "\n")
	cmd.flags.SetOutput(out)
	cmd.flags.PrintDefaults()
	fmt.Fprintf(out, "\n")
	if out == os.Stderr {
		os.Exit(1)
	}
	os.Exit(0)
}

// NewProjectsApprovalRulesUpdateCommand returns a new, initialized
// ProjectsApprovalRulesUpdateCommand instance.
func NewProjectsApprovalRulesUpdateCommand(
	name string,
	opts *ProjectsApprovalRulesUpdateOptions,
	client *gitlab.Client,
) *ProjectsApprovalRulesUpdateCommand {

	// Create the new command.
	cmd := &ProjectsApprovalRulesUpdateCommand{
		GitlabCommand: GitlabCommand[ProjectsApprovalRulesUpdateOptions]{
			BasicCommand: BasicCommand[ProjectsApprovalRulesUpdateOptions]{
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

// updateApprovalRule updates the approval rule for the project to
// have the same values as before except with a new list of user IDs.
// This function is designed to be the callback for
// [ForEachApprovalRuleInProject()].  The update actually happens only
// if dryRun is not set.
func updateApprovalRule(
	s *gitlab.ProjectsService,
	projectID int,
	rule *gitlab.ProjectApprovalRule,
	userIDs []int,
	dryRun bool,
) error {
	var err error
	fmt.Printf("    Updating rule %d (%q) ... ", rule.ID, rule.Name)
	if !dryRun {
		err = gitlab_util.UpdateApprovalRule(s, projectID, rule, userIDs)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Done.\n")
	return nil
}

// Run is the entry point for this command.
func (cmd *ProjectsApprovalRulesUpdateCommand) Run(args []string) error {
	var err error
	var approvers []*xml_users.XmlUser

	// Parse command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Validate the options.
	if cmd.options.ApproversFileName == "" {
		return fmt.Errorf("approvers file name not set")
	}
	if cmd.options.Group == "" {
		return fmt.Errorf("group not set")
	}

	// Load list of approvers.
	approvers, err = xml_users.ReadUsers(cmd.options.ApproversFileName)
	if err != nil {
		return nil
	}

	// Get the user IDs for the approvers.
	var approverIDs []int
	for _, approver := range approvers {
		approverIDs = append(approverIDs, approver.ID)
	}

	// Update each approval rule for each project.
	return gitlab_util.ForEachProjectInGroup(
		cmd.client.Groups,
		cmd.options.Group,
		cmd.options.Expr,
		cmd.options.Recursive,
		func(g *gitlab.Group, p *gitlab.Project) (bool, error) {
			fmt.Printf("%v\n", p.PathWithNamespace)
			return true, gitlab_util.ForEachApprovalRuleInProject(
				cmd.client.Projects,
				p,
				func(rule *gitlab.ProjectApprovalRule) (bool, error) {
					return true, updateApprovalRule(
						cmd.client.Projects,
						p.ID,
						rule,
						approverIDs,
						cmd.options.DryRun)
				})
		})
}
