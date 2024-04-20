package main

import(
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"encoding/xml"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/authinfo"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/common_options"
	"github.com/google/uuid"
	"github.com/xanzy/go-gitlab"
)

// Options holds the command-line options and values read from options.xml.
type Options struct {

	// Common Options
	common_options.CommonOptions

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

func main() {

	var err error

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

	// Load the authentication information from file.
	authInfo, err := authinfo.Load(opts.AuthFileName)
	if err != nil {
		log.Fatalf(
			"LoadAuthInfo: Unable to load authentication information " +
			"from file %v: %v", opts.AuthFileName, err)
	}

	// Create the Gitlab client based on the authentication
	// information provided by the user.
	client, err := authInfo.CreateGitlabClient(
		gitlab.WithBaseURL(opts.BaseURL))
	if err != nil {
		log.Fatalf("CreateGitlabClient: %v\n", err)
	}

	// DELETE ME!!!
	client = client

	// Create the projects.
	for i := uint64(0); i < opts.ProjectCount; i++ {
		fmt.Printf("uuid=%v\n", uuid.NewString())
	}
}
