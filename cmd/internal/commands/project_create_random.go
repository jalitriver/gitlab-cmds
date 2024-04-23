// This file provides the implementation for the "project
// create-random" command which creates random projects en masse.

package commands

import (
	"flag"
	"fmt"

	"github.com/google/uuid"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// ProjectCreateRandomOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in main.go so that all of the
// options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in main.go to be lean, we
// factor out our options into their own data structure.
//

// ProjectCreateRandomOptions are the options needed by this command.
type ProjectCreateRandomOptions struct {

	// DryRun should cause the command to print what it would do
	// instead of actually doing it.
	DryRun bool

	// ParentGroup is the group where projects will be created.  The
	// parent group must already exist.
	ParentGroup string

	// ProjectBaseName is the base name all new project will have.
	// The full name for the project will include random characters
	// after the base name.
	ProjectBaseName string

	// ProjectCount is the number of projects to create.
	ProjectCount uint64
}

// Initialize initializes this ProjectCreateRandomOptions instance so
// it can be used with the "flag" package to parse the command-line
// arguments.
func (opts *ProjectCreateRandomOptions) Initialize(flags *flag.FlagSet) {

	// -n
	flags.BoolVar(
		&opts.DryRun, "n", opts.DryRun,
		"print what it would do instead of actually doing it")

	// --dry-run
	flags.BoolVar(&opts.DryRun, "dry-run", opts.DryRun,
		"print what it would do instead of actually doing it")

	// --parent-group
	flags.StringVar(&opts.ParentGroup, "parent-group", "",
		"parent group for new projects")

	// --project-base-name
	flags.StringVar(&opts.ProjectBaseName, "project-base-name", "",
		"base name for new projects")

	// --project-count
	flags.Uint64Var(&opts.ProjectCount, "project-count", 0,
		"number of new projects to create")
}

////////////////////////////////////////////////////////////////////////
// ProjectCreateRandomCommand
////////////////////////////////////////////////////////////////////////

// ProjectCreateRandomCommand implements the "project create-random"
// command which creates random project en masse.
type ProjectCreateRandomCommand struct {

	// Embed the Command members.
	GitlabCommand[ProjectCreateRandomOptions]
}

// NewProjectCreateRandomCommand returns a new and initialized
// ProjectCreateRandomCommand instance.
func NewProjectCreateRandomCommand(
	name string,
	opts *ProjectCreateRandomOptions,
	client *gitlab.Client,
) *ProjectCreateRandomCommand {

	// Create the new command.
	cmd := &ProjectCreateRandomCommand{
		GitlabCommand: GitlabCommand[ProjectCreateRandomOptions]{
			BasicCommand: BasicCommand[ProjectCreateRandomOptions]{
				commandName: name,
				flags:       flag.NewFlagSet(name, flag.ExitOnError),
				options:     opts,
			},
			client: client,
		},
	}

	// Initialize our command-line options.
	opts.Initialize(cmd.flags)

	return cmd
}

// CreateRandomProject creates a projects in the parent group
// specified by parentGroupID.  The parentGroup string is only use for
// logging.  The name of each project is a combination of the project
// base name and a UUID.  If dryRun is true, this function only prints
// what it would without actually doing it.
func CreateRandomProject(
	client *gitlab.Client,
	parentGroup *gitlab.Group,
	projectBaseName string,
	dryRun bool,
) error {

	// Create UUID and use it as the suffix for the new project name.
	suffix := uuid.NewString()
	relativePath := projectBaseName + "-" + suffix
	fullPath := parentGroup.FullPath + "/" + relativePath

	// Set up options for creating the project.
	opts := gitlab.CreateProjectOptions{
		NamespaceID:          gitlab.Ptr(parentGroup.ID),
		Path:                 gitlab.Ptr(relativePath),
		Description:          gitlab.Ptr("Test Project"),
		MergeRequestsEnabled: gitlab.Ptr(true),
		SnippetsEnabled:      gitlab.Ptr(true),
		Visibility:           gitlab.Ptr(gitlab.PublicVisibility),
	}

	// Create the project.
	fmt.Printf("- Creating project: %q ... ", fullPath)
	if !dryRun {
		_, _, err := client.Projects.CreateProject(&opts)
		if err != nil {
			return fmt.Errorf("CreateProject: %w", err)
		}
	}
	fmt.Printf("Done.\n")

	return nil
}

// CreateRandomProjects creates the specified number of projects in the
// parent group.  The name of each project is a combination of the
// project base name and a UUID.  If dryRun is true, this function
// only prints what it would without actually doing it.
func CreateRandomProjects(
	client *gitlab.Client,
	parentGroup string,
	projectBaseName string,
	projectCount uint64,
	dryRun bool,
) error {

	// Get the parent group ID.
	fmt.Printf("- Searching for ID for parent group %q ... ", parentGroup)
	g, err := gitlab_util.FindExactGroup(client.Groups, parentGroup)
	if err != nil {
		return err
	}
	fmt.Printf("Done.\n")

	// Create each project.
	for i := uint64(0); i < projectCount; i++ {
		err := CreateRandomProject(client, g, projectBaseName, dryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

// Run is the entry point for this command.
func (cmd *ProjectCreateRandomCommand) Run(args []string) error {
	var err error

	// Parse command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Validate the options.
	if cmd.options.ParentGroup == "" {
		return fmt.Errorf("invalid parent group: %q", cmd.options.ParentGroup)
	} else if cmd.options.ProjectBaseName == "" {
		return fmt.Errorf("invalid project base name: %q", cmd.options.ProjectBaseName)
	} else if cmd.options.ProjectCount == 0 {
		return fmt.Errorf("invalid project count: %v", cmd.options.ProjectCount)
	}

	// Create random projects.
	return CreateRandomProjects(
		cmd.client,
		cmd.options.ParentGroup,
		cmd.options.ProjectBaseName,
		cmd.options.ProjectCount,
		cmd.options.DryRun)
}
