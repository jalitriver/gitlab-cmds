package main

import(
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"encoding/xml"

	"github.com/jalitriver/gitlab-cmds/cmd/internal"
	"github.com/xanzy/go-gitlab"
)

// Options holds the command-line options.
type Options struct {
	internal.CommonOptions
}

// Initialize initializes this Options instance by parsing the
// command-line arguments to find the location of options.xml file
// which might have been specified on the command-line. It then reads
// the options.xml file to initialize this Options instance.  Because
// command-line options take precedence over options in the
// options.xml file, it is necessary for the call to call flag.Parse()
// a second time.
func (opts *Options) Initialize() {

	// Inform the "flag" package where it should store the
	// command-line options.
	opts.CommonOptions.Initialize()

	// Parse the command-line options primarily looking for an
	// alternative location for the options.xml file which might have
	// been specified on the command line.
	flag.Parse()

	// Try to open the options.xml file.
	f, err := os.Open(opts.OptionsFileName)
	if err != nil {
		// Squash the error.  The user is not required to have an
		// options.xml file.
		return
	}
	defer f.Close()

	// Try to read the options.xml file.
	err = xml.NewDecoder(f).Decode(&opts)
	if err != nil {
		log.Fatalf(
			"unable to read options from %q: %v",
			opts.OptionsFileName,
			err)
	}
}

func main() {

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
	opts.Initialize()
	flag.Parse()

	// Load the authentication information from file.
	authInfo, err := internal.LoadAuthInfo(opts.AuthFileName)
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

	// Get the list of projects.
	listProjOpts := gitlab.ListProjectsOptions{}
	ps, _, err := client.Projects.ListProjects(&listProjOpts)
	if err != nil {
		log.Fatalf("ListProjects: %v\n", err)
	}

	// Print each project.
	for _, p := range ps {
		fmt.Printf("%v: %v\n", p.ID, p.WebURL)
	}
}
