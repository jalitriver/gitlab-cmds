package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/authinfo"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/global_options"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/xanzy/go-gitlab"
)

// Options holds the command-line options and values read from options.xml.
type Options struct {

	// Global Options
	global_options.GlobalOptions

	// Expr is the regular expression that filters the projects.
	Expr string

	// Group for which projects will be listed.
	Group string

	// Recursive controls whether the projects are listed recursively.
	Recursive bool
}

// Initialize initializes this Options instance by parsing the
// command-line arguments to find the location of options.xml file
// which might have been specified on the command-line. It then reads
// the options.xml file to initialize this Options instance.  Because
// command-line options take precedence over options in the
// options.xml file, it is necessary for the caller to call
// flag.Parse() a second time.
func (opts *Options) Initialize() error {

	// Inform the "flag" package where it should store the global
	// command-line options.
	opts.GlobalOptions.Initialize()

	// Inform the "flag" package where it should store the
	// command-specific options.
	flag.StringVar(&opts.Expr, "expr", "", "regular expression which filters the projects")
	flag.StringVar(&opts.Group, "group", "", "group for which projects will be listed")
	flag.BoolVar(&opts.Recursive, "r", false, "whether to recursively list projects")
	flag.BoolVar(&opts.Recursive, "recursive", false, "whether to recursively list projects")

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
		return nil, err
	}

	// Augment the options from the options.xml file with options from
	// the command-line arguments.
	flag.Parse()

	// Validate the options.
	if opts.Group == "" {
		return nil, fmt.Errorf("invalid group: %q", opts.Group)
	}

	return opts, nil
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

	// Print the projects.
	err = gitlab_util.ForEachProjectInGroup(
		client.Groups, opts.Group, opts.Expr, opts.Recursive,
		func(g *gitlab.Group, p *gitlab.Project) (bool, error) {
			fmt.Printf("%v: %v\n", p.ID, p.PathWithNamespace)
			return true, nil
		})
	if err != nil {
		goto out
	}

out:

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n*** Error: %v\n\n", err)
		os.Exit(1)
	}
}
