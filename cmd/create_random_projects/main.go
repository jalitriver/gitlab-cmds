package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/authinfo"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/common_options"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/xanzy/go-gitlab"
)

// Options holds the command-line options and values read from options.xml.
type Options struct {

	// Common Options
	common_options.CommonOptions

	// DryRun should cause the command to print what it would do
	// instead of actually doing it.
	DryRun bool `xml:"-"`

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

// Initialize initializes this Options instance by parsing the
// command-line arguments to find the location of options.xml file
// which might have been specified on the command-line. It then reads
// the options.xml file to initialize this Options instance.  Because
// command-line options take precedence over options in the
// options.xml file, it is necessary for the caller to call
// flag.Parse() a second time.
func (opts *Options) Initialize() error {

	// Inform the "flag" package where it should store the common
	// command-line options.
	opts.CommonOptions.Initialize()

	// Inform the "flag" package where it should store the
	// command-line options specific to this command.
	flag.BoolVar(&opts.DryRun, "dry-run", opts.DryRun, "print what it would do instead of actually doing it")
	flag.StringVar(&opts.ParentGroup, "parent-group", "", "parent group for new projects")
	flag.StringVar(&opts.ProjectBaseName, "project-base-name", "", "base name for new projects")
	flag.Uint64Var(&opts.ProjectCount, "project-count", 0, "number of new projects to create")

	// Parse the command-line options primarily looking for an
	// alternative location for the options.xml file which might have
	// been specified on the command line.
	flag.Parse()

	// If you have any command-line options that accumulate, you need
	// reset them here; otherwise, those options will have duplicate
	// values when flag.Parse() is called the second time as explained
	// in the method-level comment (above).

	// Try to open the options.xml file.
	if opts.OptionsFileName != "" {
		f, err := os.Open(opts.OptionsFileName)
		if err != nil {
			return err
		}
		defer f.Close()

		// Try to read the options.xml file.
		err = xml.NewDecoder(f).Decode(&opts)
		if err != nil {
			return fmt.Errorf("%v: %w", opts.OptionsFileName, err)
		}
	}

	return nil
}

// ParseOptions uses the "flag" package to parse our command-line
// options and return the result.
func ParseOptions() (*Options, error) {

	// Initialize a new Options instance including reading default
	// options from the options.xml configuration file.
	opts := new(Options)
	err := opts.Initialize()
	if err != nil {
		return nil ,err
	}

	// Augment the options from the options.xml file with options from
	// the command-line arguments.
	flag.Parse()

	// Validate the options.
	if opts.ParentGroup == "" {
		return nil, fmt.Errorf("invalid parent group: %q", opts.ParentGroup)
	} else if opts.ProjectBaseName == "" {
		return nil, fmt.Errorf("invalid project base name: %q", opts.ProjectBaseName)
	} else if opts.ProjectCount == 0 {
		return nil, fmt.Errorf("invalid project count: %v", opts.ProjectCount)
	}

	return opts, nil
}

// CreateProject creates a projects in the parent group specified by
// parentGroupID.  The parentGroup string is only use for logging.
// The name of each project is a combination of the project base name
// and a UUID.  If dryRun is true, this function only prints what it
// would without actually doing it.
func CreateProject(
	client *gitlab.Client,
	parentGroupID int,
	parentGroup string,
	projectBaseName string,
	dryRun bool,
) error {

	// Create UUID and use it as the suffix for the new project name.
	suffix := uuid.NewString()
	relativePath := projectBaseName + "-" + suffix
	fullPath := parentGroup + "/" + relativePath

	// Set up options for creating the project.
	opts := gitlab.CreateProjectOptions{
		NamespaceID:          gitlab.Ptr(parentGroupID),
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

// CreateProjects creates the specified number of projects in the
// parent group.  The name of each project is a combination of the
// project base name and a UUID.  If dryRun is true, this function
// only prints what it would without actually doing it.
func CreateProjects(
	client *gitlab.Client,
	parentGroup string,
	projectBaseName string,
	projectCount uint64,
	dryRun bool,
) error {
	
	// Get the parent group ID.
	fmt.Printf("- Searching for ID for parent group %q ... ", parentGroup)
	parentGroupID, err :=
		gitlab_util.FindUniqueGroupID(client.Groups, parentGroup)
	if err != nil {
		return err
	}
	fmt.Printf("Done.\n")

	// Create each project.
	for i := uint64(0); i < projectCount; i++ {
		err := CreateProject(client, parentGroupID, parentGroup, projectBaseName, dryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {

	var client *gitlab.Client
	var authInfo authinfo.AuthInfo

	// Find the base name for the executable.
	basename := filepath.Base(os.Args[0])

	// Usage.
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "\n")
		fmt.Fprintf(out, "Usage: %s [options] [<file> ...]\n", basename)
		fmt.Fprintf(out, "\n")
		fmt.Fprintf(out, "    List Gitlab Projects\n")
		fmt.Fprintf(out, "\n")
		fmt.Fprintf(out, "Options:\n")
		fmt.Fprintf(out, "\n")
		flag.PrintDefaults()
	}

	// Parse command-line arguments.
	opts, err := ParseOptions()
	if err != nil {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "%v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Load the authentication information from file.
	authInfo, err = authinfo.Load(opts.AuthFileName)
	if err != nil {
		err = fmt.Errorf(
			"LoadAuthInfo: Unable to load authentication information "+
				"from file %v: %w\n", opts.AuthFileName, err)
		goto out
	}

	// Create the Gitlab client based on the authentication
	// information provided by the user.
	client, err = authInfo.CreateGitlabClient(
		gitlab.WithBaseURL(opts.BaseURL))
	if err != nil {
		err = fmt.Errorf("CreateGitlabClient: %w\n", err)
		goto out
	}

	// Create projects.
	err = CreateProjects(
		client,
		opts.ParentGroup,
		opts.ProjectBaseName,
		opts.ProjectCount,
		opts.DryRun)
	if err != nil {
		goto out
	}

out:

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n*** Error: %v\n\n", err)
		os.Exit(1)
	}
}
