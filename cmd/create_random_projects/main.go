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

// GroupFullPaths returns just the full paths for the groups.
func GroupFullPaths(groups []*gitlab.Group) []string {
	result := make([]string, 0, len(groups))
	for _, group := range groups {
		result = append(result, group.FullPath)
	}
	return result
}

func main() {

	var err error
	var authInfo authinfo.AuthInfo
	var client *gitlab.Client
	var groups []*gitlab.Group
	var grpopts gitlab.ListGroupsOptions

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
	opts := new(Options)
	flag.BoolVar(&opts.DryRun, "dry-run", opts.DryRun, "print what it would do instead of actually doing it")
	flag.StringVar(&opts.ParentGroup, "parent-group", "", "parent group for new projects")
	flag.StringVar(&opts.ProjectBaseName, "project-base-name", "", "base name for new projects")
	flag.Uint64Var(&opts.ProjectCount, "project-count", 0, "number of new projects to create")
	err = opts.Initialize()
	if err == nil {
		flag.Parse()
		if opts.ParentGroup == "" {
			err = fmt.Errorf("invalid parent group: %q", opts.ParentGroup)
		} else if opts.ProjectBaseName == "" {
			err = fmt.Errorf("invalid project base name: %q", opts.ProjectBaseName)
		} else if opts.ProjectCount == 0 {
			err = fmt.Errorf("invalid project count: %v", opts.ProjectCount)
		}
	}
	if err != nil {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "%v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	//
	// Errors below here do not need to print the usage message.
	//

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

	// Search for the parent group ID.
	fmt.Printf("- Searching for ID for parent group %q ... ", opts.ParentGroup)
	grpopts = gitlab.ListGroupsOptions{
		Search: gitlab.Ptr(opts.ParentGroup),
	}
	groups, _, err = client.Groups.ListGroups(&grpopts)
	if err != nil {
		err = fmt.Errorf("ListGroups: %w", err)
		goto out
	}
	fmt.Printf("Done.\n")
	if len(groups) == 0 {
		err = fmt.Errorf("could not find group: %v", opts.ParentGroup)
		goto out
	}
	if len(groups) > 1 {
		err = fmt.Errorf("found multiple matching groups: %v", GroupFullPaths(groups))
		goto out
	}

	// Create the projects.
	for i := uint64(0); i < opts.ProjectCount; i++ {

		// Create UUID and use it as the suffix for the new project name.
		suffix := uuid.NewString()
		projname := opts.ProjectBaseName + "-" + suffix
		projpath := opts.ParentGroup + "/" + projname

		// Set up options for creating the project.
		projopts := gitlab.CreateProjectOptions{
			NamespaceID:          gitlab.Ptr(groups[0].ID),
			Path:                 gitlab.Ptr(projname),
			Description:          gitlab.Ptr("Test Project"),
			MergeRequestsEnabled: gitlab.Ptr(true),
			SnippetsEnabled:      gitlab.Ptr(true),
			Visibility:           gitlab.Ptr(gitlab.PublicVisibility),
		}

		// Create the project.
		fmt.Printf("- Creating project: %q ... ", projpath)
		if !opts.DryRun {
			_, _, err = client.Projects.CreateProject(&projopts)
			if err != nil {
				err = fmt.Errorf("CreateProject: %w", err)
				goto out
			}
		}
		fmt.Printf("Done.\n")
	}

out:

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n*** Error: %v\n\n", err)
		os.Exit(1)
	}
}
