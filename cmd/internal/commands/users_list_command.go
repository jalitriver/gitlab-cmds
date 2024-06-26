// This file provides the implementation for the "users list" command
// which lists or searches for specific users so they can be used with
// other commands.

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jalitriver/gitlab-cmds/cmd/internal/date_arg"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/gitlab_util"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/string_slice"
	"github.com/jalitriver/gitlab-cmds/cmd/internal/xml_users"
	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// UsersListOptions
////////////////////////////////////////////////////////////////////////

//
// NOTE: We cannot put these options in the Command struct because the
// way it works is the options are (eventually) embedded in the single
// large "Options" data structure in global_command.go so that all of
// the options can be read from a single options.xml file.  Because we
// want the main "Options" data structure in global_command.go to be
// lean, we factor out our options into their own data structure.
//

// UsersListOptions are the options needed by this command.
type UsersListOptions struct {

	// CreatedDate is the date after which users must have been
	// created in order to be listed.
	CreatedAfter date_arg.DateArg `xml:"created-after"`

	// OutputFileName is the name of XML output file to which users
	// will be appended.  If empty, no XML output file is written, but
	// there will still be logging to the console.  If set to "-", XML
	// output will be written to os.Stdout.
	OutputFileName string `xml:"output-file-name"`

	// MatchSubstrings controls whether all substrings matches are
	// reported instead of only reporting exact matches.
	MatchSubstrings bool `xml:"match-substrings"`

	// Users (for the --users option)
	Users string_slice.StringSlice `xml:"users>user"`
}

// Initialize initializes this UsersListOptions instance so it can be
// used with the "flag" package to parse the command-line arguments.
func (opts *UsersListOptions) Initialize(flags *flag.FlagSet) {

	// --created-after
	flags.Var(&opts.CreatedAfter, "created-after",
		"date after which users not specified by user ID must have been "+
			"created to be listed the form of which is YYYY/MM/DD or "+
			"YYYY-MM-DD")

	// --match-substrings
	flags.BoolVar(&opts.MatchSubstrings, "match-substrings", opts.MatchSubstrings,
		"whether all substrings matches are reported instead of reporting "+
			"only exact matches")

	// -o
	flags.StringVar(&opts.OutputFileName, "o", opts.OutputFileName,
		"name of XML output file to which users will be appended")

	// --out
	flags.StringVar(&opts.OutputFileName, "out", opts.OutputFileName,
		"name of XML output file to which users will be appended")

	// --users
	flags.Var(&opts.Users, "users",
		"comma-separated list of user IDs, names, usernames, or "+
			"e-mail addresses")
}

////////////////////////////////////////////////////////////////////////
// UsersListCommand
////////////////////////////////////////////////////////////////////////

// UsersListCommand implements the "users list" command which lists
// (or looks up) specific users so they can be used with other
// commands.
type UsersListCommand struct {

	// Embed the Command members.
	GitlabCommand[UsersListOptions]
}

// Usage prints the usage message to the output writer.  If err is not
// nil, it will be printed before the main output.
func (cmd *UsersListCommand) Usage(out io.Writer, err error) {
	basename := filepath.Base(os.Args[0])
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out,
		"Usage: %s [global_options] users list [subcmd_options]\n",
		basename)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    List users matching search strings and optionally\n")
	fmt.Fprintf(out, "    save the list of users to file.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "    WARNING: At the time of writing, listing users by e-mail\n")
	fmt.Fprintf(out, "    address and the --created-after flag are not working\n")
	fmt.Fprintf(out, "    with Gitlab CE.\n")
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

// NewUsersListCommand returns a new, initialized
// UsersListCommand instance.
func NewUsersListCommand(
	name string,
	opts *UsersListOptions,
	client *gitlab.Client,
) *UsersListCommand {

	// Create the new command.
	cmd := &UsersListCommand{
		GitlabCommand: GitlabCommand[UsersListOptions]{
			BasicCommand: BasicCommand[UsersListOptions]{
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

// printUser prints the user.  If index is zero, the header is printed
// on the line above the user.
func printUser(index int, user *gitlab.User) error {

	// Print the header if necessary.
	if index == 0 {
		_, err := fmt.Printf("%8s  %-16s  %-24s  %-24s\n",
			"ID", "Username", "Name", "Email")
		if err != nil {
			return err
		}
	}

	// Print the user.
	_, err := fmt.Printf("%8d  %-16s  %-24s  %-24s\n",
		user.ID, user.Username, user.Name, user.Email)

	return err
}

// Run is the entry point for this command.
func (cmd *UsersListCommand) Run(args []string) error {
	var err error
	var found []*gitlab.User
	var users []*gitlab.User

	// Parse command-line arguments.
	err = cmd.flags.Parse(args)
	if err != nil {
		return err
	}

	// If users were specified, try to find exact matches for the
	// "user" search strings.  If an exact match is found, add them to
	// the "found" list so we can write them to file before exiting if
	// necessary.
	if len(cmd.options.Users) > 0 {
		for i, user := range cmd.options.Users {
			users, err = gitlab_util.FindUsers(
				cmd.client.Users,
				user,
				!cmd.options.MatchSubstrings,
				time.Time(cmd.options.CreatedAfter))
			if err != nil {
				return fmt.Errorf("unable to find user: %q\n", user)
			}
			found = append(found, users...)
			for j, u := range users {
				err = printUser(i+j, u)
				if err != nil {
					return err
				}
			}
		}
	}

	// If no users were specified, list all users.
	if len(cmd.options.Users) == 0 {
		i := 0
		err = gitlab_util.ForEachUser(
			cmd.client.Users,
			"", /* user */
			time.Time(cmd.options.CreatedAfter),
			func(u *gitlab.User) (bool, error) {
				found = append(found, u)
				i++
				return true, printUser(i-1, u)
			})
		if err != nil {
			return err
		}
	}

	// Save results to output file.
	if cmd.options.OutputFileName != "" {
		err = xml_users.WriteUsers(cmd.options.OutputFileName, found)
		if err != nil {
			return err
		}
	}

	return nil
}
