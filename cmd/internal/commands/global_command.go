// This file defines GlobalCommand which is the parent of all other commands.
//
// If you need to add a new subcommand, do the following:
//
//   1) Create the new subcommand similar to
//      cmd/internal/commands/project_command.go if the subcommand
//      will have its own set of subcommands or similar to
//      cmd/internal/commands/project_list_command.go if the
//      subcommand will actually do something.
//
//   2) Add the resulting new options struct to the Options struct
//      below so the options can also be specified in the options.xml
//      file.
//
//   3) Add the new subcommand generator as demonstrated in
//      GlobalCommand.addSubcmdGenerators().

package commands

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/authinfo"
	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Options
////////////////////////////////////////////////////////////////////////

// Options is the top-level structure that holds all the options read
// from options.xml and from the command-line.  It is used by
// GlobalCommand when configuring its subcommands.  Each member of
// Options represents a subcommand that can be directly invoked by
// GlobalCommand.  For example, if a subcommand is invoked by another
// subcommand (e.g. "glcli project list"), the subcommand options
// (i.e., ProjectListOptions) will be present in their parent
// subcommand options (i.e., ProjectOptions) which in turn will be
// present in this data structure (i.e. Options).
type Options struct {

	// Name of the root XML element.
	XMLName xml.Name `xml:"options"`

	// Global Options
	GlobalOpts GlobalOptions `xml:"global-options"`

	// Options for the "project" command.
	ProjectOpts ProjectOptions `xml:"project-options"`
}

// LoadFromXMLFile loads options from the XML file.
func (opts *Options) LoadFromXMLFile(fname string) error {

	// Try to open the options.xml file.
	f, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("LoadFromXMLFile: %w", err)
	}
	defer f.Close()

	// Try to read the options.xml file.
	err = xml.NewDecoder(f).Decode(opts)
	if err != nil {
		return fmt.Errorf("LoadFromXMLFile: %v: %w", fname, err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////
// GlobalOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure (above) so that all of the options
// can be read from a single options.xml file.  Because we want the
// main "Options" data structure (above) to be lean, we factor out our
// options into their own data structure.
//

// GlobalOptions are the options needed by this command.
type GlobalOptions struct {

	// AuthFileName is an alternative file name for auth.xml which
	// holds authentication information like an OAuth token or
	// personal access token.  Defaults to "auth.xml".
	AuthFileName string `xml:"auth-file-name"`

	// BaseURL is the base URL for connecting to Gitlab REST
	// endpoints.  It does not include the "api/v4" part.  Defaults to
	// "https://gitlab.com/".
	BaseURL string `xml:"base-url"`

	// Help is whether the user wants help.  Defaults to false.
	Help bool `xml:"help"`

	// OptionsFileName is an alternative file name for options.xml.
	// Note that the user can only change this option on the command
	// line, not in the options.xml file (because it leads to circular
	// logic having the user specify the location of the options.xml
	// file in the options.xml file).  Defaults to "options.xml".
	OptionsFileName string `xml:"-"`

	// Version is whether the user wants the version.  Defaults to false.
	Version bool `xml:"version"`
}

// Initialize initializes this GlobalOptions instance so it can be
// used with the "flag" package to parse the command-line arguments.
func (opts *GlobalOptions) Initialize(flags *flag.FlagSet) {

	// Set default values that differ from the zero defaults.
	opts.AuthFileName = "auth.xml"
	opts.BaseURL = "https://gitlab.com/"
	opts.OptionsFileName = "options.xml"

	// --auth
	flags.StringVar(&opts.AuthFileName, "auth", opts.AuthFileName,
		"name of XML file with authentication information")

	// --base-url
	flags.StringVar(&opts.BaseURL, "base-url", opts.BaseURL,
		"base URL for Gitlab REST endpoints which should not include "+
			"the \"api/v4\" suffix")

	// -h
	flags.BoolVar(&opts.Help, "h", opts.Help,
		"print help")

	// --help
	flags.BoolVar(&opts.Help, "help", opts.Help,
		"print help")

	// --options
	flags.StringVar(&opts.OptionsFileName, "options", opts.OptionsFileName,
		"name of XML file with default options")

	// -v
	flags.BoolVar(&opts.Version, "v", opts.Version,
		"print version")

	// --version
	flags.BoolVar(&opts.Version, "version", opts.Version,
		"print version")
}

////////////////////////////////////////////////////////////////////////
// GlobalCommand
////////////////////////////////////////////////////////////////////////

// GlobalCommand is used to parse the global command-line arguments
// and invoke the first subcommand.
type GlobalCommand struct {

	// Embed the ParentCommand members.
	ParentCommand[GlobalOptions]

	// allOpts is the master structure that holds all of the options
	// which can be read from options.xml or the command-line.  For
	// example, the GlobalOptions instance used by this program is at
	// allOpts.GlobalOpts.  These options all need to be in a single
	// data structure in order to easily use Go's XML parser.
	allOpts *Options

	// generators is a slice of functions that generate the runnable
	// subcommands.  (This has nothing to do with Python-style
	// generators.)  See the comments for addSubcmdGenerators().
	generators map[string]func() Runner

	// client is the Gitlab communication client
	client *gitlab.Client

	// version is the program version needed for the --version option.
	version string
}

// Usage prints the main usage message to the output writer.  If
// err is not nil, it will be printed before the main output.
func (cmd *GlobalCommand) Usage(out io.Writer, err error) {
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Usage: %s [global_options] subcmd [subcmd_options]\n", cmd.name)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    Commands for administering a Gitlab server.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Global Options:\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  Global options must precede the first subcommand.\n")
	fmt.Fprintf(out, "\n")
	cmd.flags.SetOutput(out)
	cmd.flags.PrintDefaults()
	fmt.Fprintf(out, "\n")

	fmt.Fprintf(out, "Subcommands:\n")
	fmt.Fprintf(out, "\n")

	// If the subcommands have not been populated yet, populate them
	// with nil Runners so we can at least print their names before
	// exiting.
	if len(cmd.subcmds) == 0 {
		for cmdName := range cmd.generators {
			cmd.subcmds[cmdName] = nil
		}
	}

	// Print the subcommand names.
	for _, subcmd := range cmd.SortedCommandNames() {
		fmt.Fprintf(out, "  %s\n", subcmd)
	}
	fmt.Fprintf(out, "\n")

	if out == os.Stderr {
		os.Exit(1)
	}
	os.Exit(0)
}

// AddSubcommandGenerators adds the subcommands generators for the
// global command.  A generator in this context is just a function
// that creates the subcommand Runnable.  The reason for this is that
// Usage() can be called very early before the subcommands can be
// instantiated, but the Usage() command needs to be able to list the
// subcommands.  So instead cmd.subcmds being populating with the
// final Runnable subcommand, it is populated early with the generator
// Runnable.  Thus, if Usage() is called early it will still have the
// complete list of subcommands to display.
func (cmd *GlobalCommand) addSubcmdGenerators() {
	cmd.generators["project"] = func() Runner {
		return NewProjectCommand(
			"project", &cmd.allOpts.ProjectOpts, cmd.client)
	}
}

// generateSubcmds generates the subcommands from the list of
// generators created by addSubcmdGenerators().  See the comments for
// addSubcmdGenerators().
func (cmd *GlobalCommand) generateSubcmds() {
	for cmdName, g := range cmd.generators {
		cmd.subcmds[cmdName] = g()
	}
}

// NewGlobalCommand returns a new and initialized GlobalCommand instance
// having the specified name.
func NewGlobalCommand(name string, version string) *GlobalCommand {

	// Create the master data structure which holds all the options.
	// These options all need to be in a single data structure in
	// order to easily use Go's XML parser.
	allOpts := new(Options)

	// Create the new command.
	cmd := &GlobalCommand{
		ParentCommand: ParentCommand[GlobalOptions]{
			BasicCommand: BasicCommand[GlobalOptions]{
				name:    name,
				flags:   flag.NewFlagSet(name, flag.ExitOnError),
				options: &allOpts.GlobalOpts,
			},
			subcmds: make(map[string]Runner),
		},
		allOpts:    allOpts,
		generators: make(map[string]func() Runner),
		version:    version,
	}

	// Set up the function that prints the global usage and exits.
	cmd.flags.Usage = func() { cmd.Usage(os.Stderr, nil) }

	// Initialize our command-line options.
	cmd.options.Initialize(cmd.flags)

	// Add the subcommand generators from which, at the appropriate
	// time, the subcommands will be created .
	cmd.addSubcmdGenerators()

	return cmd
}

// Run is the entry point for this command.
func (cmd *GlobalCommand) Run(args []string) error {
	var err error
	var authInfo authinfo.AuthInfo

	// This is the first time we parse the global command-line
	// options.  We are primarily looking for an alternative location
	// for the options.xml file which might have been specified on the
	// command line.  Note that we only need to parse the "global"
	// part of the command-line (i.e., the part that preceeds the
	// first subcommand).  Thus, we do not need to invoke the Parse()
	// methods of any of flag.FlagSet objects for the subcommands.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Print help and then exit if requested by the user.
	if cmd.options.Help {
		cmd.Usage(os.Stdout, nil)
		// not reached
	}

	// Print the version if requested by the user.
	if cmd.options.Version {
		fmt.Printf("%s v%s\n", cmd.name, cmd.version)
		return nil
	}

	// Load options from XML file.
	if cmd.options.OptionsFileName != "" {
		err = cmd.allOpts.LoadFromXMLFile(cmd.options.OptionsFileName)
		if err != nil {
			cmd.Usage(os.Stderr, err)
			// not reached
		}
	}

	// Load the authentication information from file.
	authInfo, err = authinfo.Load(cmd.options.AuthFileName)
	if err != nil {
		return fmt.Errorf(
			"LoadAuthInfo: Unable to load authentication information "+
				"from file %v: %w\n", cmd.options.AuthFileName, err)
	}

	// Create the Gitlab client based on the authentication
	// information provided by the user.
	cmd.client, err = authInfo.CreateGitlabClient(
		gitlab.WithBaseURL(cmd.options.BaseURL))
	if err != nil {
		return fmt.Errorf("CreateGitlabClient: %w\n", err)
	}

	// Generate the subcommands.
	cmd.generateSubcmds()

	// Override the options from the XML file with global options from
	// the command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// Dispatch the subcommand specified by the remaining arguments.
	return cmd.DispatchSubcommand(cmd.flags.Args())
}
