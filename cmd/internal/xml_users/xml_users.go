// This file is for reading and writing to the users.xml.  This is
// common code (especially reading from users.xml) that needs to be
// available for multiple subcommands.

package xml_users

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xanzy/go-gitlab"
)

// User list for the user.xml file.
type XmlUsers struct {
	XMLName xml.Name   `xml:"users"`
	Users   []*XmlUser `xml:"user"`
}

// User for the user.xml file.
type XmlUser struct {
	ID       int    `xml:"id"`
	Username string `xml:"username"`
	Email    string `xml:"email"`
	Name     string `xml:"name"`
}

// FromGitlabUser converts from gitlab.User to gilab_util.XmlUser by
// removing all the unnecessary user information.
func FromGitlabUser(glUser *gitlab.User) *XmlUser {
	return &XmlUser{
		ID:       glUser.ID,
		Username: glUser.Username,
		Email:    glUser.Email,
		Name:     glUser.Name,
	}
}

// FromGitlabUsers converts from gitlab.User slice to
// gilab_util.XmlUser slice by removing all the unnecessary user
// information.
func FromGitlabUsers(glUsers []*gitlab.User) []*XmlUser {
	var result []*XmlUser
	for _, glUser := range glUsers {
		result = append(result, FromGitlabUser(glUser))
	}
	return result
}

// ReadUsers reads the users from the XML file.
func ReadUsers(fname string) ([]*XmlUser, error) {
	var err error
	var fin *os.File

	// Sanity check.
	if fname == "" || fname == "-" {
		return nil, fmt.Errorf("invalid file name: %q", fname)
	}

	// Open the file.
	fin, err = os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	// Load the users from the XML file
	xmlUsers := XmlUsers{}
	err = xml.NewDecoder(fin).Decode(&xmlUsers)
	if err != nil {
		return nil, err
	}

	return xmlUsers.Users, nil
}

// CountUsers returns the set of users as a map from username to the
// number of times that user appears in the list.
func CountUsers(users []*XmlUser) map[string]int {
	result := make(map[string]int)
	for _, user := range users {
		result[user.Username]++
	}
	return result
}

// appendUsersFromFile appends the list of new users to the list of
// users from the XML file.  The newly looked up users will replace
// users having the same ID in the XML file.
func AppendUsersFromFile(
	fname string,
	newXmlUsers []*XmlUser,
) ([]*XmlUser,
	error,
) {
	var err error
	var result []*XmlUser
	var origXmlUsers []*XmlUser

	// Load the original list of XML users from the file.  If we get
	// an error like "no such file or directory" we will just return
	// the same slice that was passed in because there is no XML file
	// to merge.
	origXmlUsers, err = ReadUsers(fname)
	if err != nil {
		return newXmlUsers, nil
	}

	// Go does not have sets so we use a map to do a quick lookup to
	// determine if the user performed a lookup on an existing user.
	newXmlUsersCount := CountUsers(newXmlUsers)

	// Keep the same order as in the original file.
	for _, xmlUserOrig := range origXmlUsers {

		// Skip original users if they have the same user ID as one of
		// the new users.
		if newXmlUsersCount[xmlUserOrig.Username] > 0 {
			continue
		}

		// Add the original user to the result.
		result = append(result, xmlUserOrig)
	}

	// Append the new users.
	result = append(result, newXmlUsers...)

	return result, nil
}

// WriteUsers writes the users to the output file.  If the output file
// already exists, the users will be merged into the existing output
// file.
func WriteUsers(fname string, glUsers []*gitlab.User) error {
	var encoder *xml.Encoder
	var err error
	var fout *os.File
	var xmlUsers []*XmlUser
	var xmlUsersRoot XmlUsers

	// Sanity check.
	if fname == "" {
		return fmt.Errorf("invalid file name: %q", fname)
	}

	// Convert from gitlab.User to gilab_util.XmlUser.
	xmlUsers = FromGitlabUsers(glUsers)

	// Check for duplicate users.
	xmlUsersCount := CountUsers(xmlUsers)
	if len(xmlUsersCount) < len(xmlUsers) {
		var dups []string
		for username, count := range xmlUsersCount {
			if count > 1 {
				dups = append(dups, username)
			}
		}
		return fmt.Errorf("WriteUsers: duplicate users detected: %q", dups)
	}

	// Append users from the original file to xmlUsers so they are not
	// lost when the original file is overwritten.
	if fname != "-" {
		xmlUsers, err = AppendUsersFromFile(fname, xmlUsers)
		if err != nil {
			goto out
		}
	}

	// If fname is "-" use stdout; otherwise, create a temporary file
	// in the same directory as fname.
	if fname == "-" {
		fout = os.Stdout
	} else {
		fout, err = os.CreateTemp(filepath.Dir(fname), filepath.Base(fname))
		if err != nil {
			goto out
		}
		defer fout.Close()
	}

	// Write XML to the temporary output file.
	xmlUsersRoot = XmlUsers{Users: xmlUsers}
	encoder = xml.NewEncoder(fout)
	encoder.Indent("", "  ")
	err = encoder.Encode(xmlUsersRoot)
	if err != nil {
		goto out
	}
	fout.WriteString("\n")

	// Atomically move the XML file into place.
	if fname != "-" {
		err = os.Rename(fout.Name(), fname)
		if err != nil {
			goto out
		}
	}

out:

	if err != nil {
		// Remove the temporary file if an error occurs.
		if fname != "-" && fout != nil {
			os.Remove(fout.Name())
		}
		return err
	}

	return nil
}
