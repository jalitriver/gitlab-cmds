// Data structure used to collect options from config.xml and from the
// command line.

package internal

import (
	"flag"
)

// CommonOptions holds the command-line options and options in the
// options.xml that are need by most of our commands.
type CommonOptions struct {
	
	// AuthFileName is an alternative file name for auth.xml.
	AuthFileName string `xml:"auth-file-name"`
	
	// BaseURL is the base URL for connecting to Gitlab REST endpoints.
	BaseURL string `xml:"base-url"`

	// OptionsFileName is an alternative file name for options.xml.
	// Note that the user can only change this option on the command
	// line, not in the options.xml file (because it leads to circular
	// logic having the user specify the location of the options.xml
	// file in the options.xml file).
	OptionsFileName string `xml:"optons-file-name,omitempty"`
}

// Initialize initializes this CommonOptions instance for use the the
// "flag" package in order to parse command-line options.
func (opts *CommonOptions) Initialize() {

	// Set default values that differ from the zero defaults.
	if opts.AuthFileName == "" {
		opts.AuthFileName = "auth.xml"
	}
	if opts.BaseURL == "" {
		opts.BaseURL = "https://gitlab.com/"
	}
	if opts.OptionsFileName == "" {
		opts.OptionsFileName = "options.xml"
	}

	// --auth
	flag.StringVar(
		&opts.AuthFileName,
		"auth",
		opts.AuthFileName,
		"name of XML file with authentication information")

	// --base-url
	flag.StringVar(
		&opts.BaseURL,
		"base-url",
		opts.BaseURL,
		"base URL for Gitlab REST endpoints which should not include " +
			"the \"api/v4\" suffix")

	// --options
	flag.StringVar(
		&opts.OptionsFileName,
		"options",
		opts.OptionsFileName,
		"name of XML file with default options")
}

