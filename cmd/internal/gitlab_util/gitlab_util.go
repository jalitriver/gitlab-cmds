// This file provides utility functions related to the Gitlab REST API.

package gitlab_util

import (
	"fmt"
	"hash/crc64"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Groups
////////////////////////////////////////////////////////////////////////

// GroupFullPaths returns just the full paths for the groups.
func GroupFullPaths(groups []*gitlab.Group) []string {
	result := make([]string, 0, len(groups))
	for _, group := range groups {
		result = append(result, group.FullPath)
	}
	return result
}

// FindExactGroup returns the ID of the group that exactly matches
// the search string.
func FindExactGroup(s *gitlab.GroupsService, group string) (*gitlab.Group, error) {

	// Set the group search string.
	opts := gitlab.ListGroupsOptions{}
	opts.Page = 1
	opts.Search = gitlab.Ptr(group)

	// Iterate over each page of groups.
	for {

		// Get a page of matching groups.
		gs, resp, err := s.ListGroups(&opts)
		if err != nil {
			err = fmt.Errorf("FindExactGroup: %w", err)
			return nil, err
		}

		// Check each group for an exact match.
		for _, g := range gs {
			if g.FullPath == group {
				return g, nil
			}
		}

		// Check if done.
		if resp.NextPage == 0 {
			break
		}

		// Move to the next page.
		opts.Page = resp.NextPage
	}

	// Could not find a matching group.
	err := fmt.Errorf(
		"FindExactGroup: could not find exact match for group: %q", group)
	return nil, err
}

////////////////////////////////////////////////////////////////////////
// Projects
////////////////////////////////////////////////////////////////////////

// ForEachProjectInGroup iterates over the projects in a group and
// recursively or not) calls the function f once for each project
// whose full path name matches the regular expression.  An empty
// regular expression matches any string.  The function f must return
// true and no error to indicate that it wants to continue being
// called with the remaining projects.  If f returns an error, it will
// be forwarded to the caller as the error return value for this
// function.  Prefer this function over GetAllProjects() to avoid the
// long delay to the user while waiting to collect all the projects.
func ForEachProjectInGroup(
	s *gitlab.GroupsService,
	group string,
	expr string,
	recursive bool,
	f func(group *gitlab.Group, project *gitlab.Project) (bool, error),
) error {

	// Find the group.
	g, err := FindExactGroup(s, group)
	if err != nil {
		return fmt.Errorf("ForEachProjectInGroup: %w", err)
	}

	// Compile the regexp.
	r, err := regexp.Compile(expr)
	if err != nil {
		return fmt.Errorf("ForEachProjectInGroup: %w", err)
	}

	// Set up the options for ListGroupProjects().
	opts := gitlab.ListGroupProjectsOptions{}
	opts.IncludeSubGroups = gitlab.Ptr(recursive)
	opts.Page = 1
	///opts.PerPage = 100

	// Iterate over each page of groups.
	for {

		// Get the next page of projects.
		ps, resp, err := s.ListGroupProjects(g.ID, &opts)
		if err != nil {
			return fmt.Errorf("ForEachProjectInGroup: %w\n", err)
		}

		// Invoke the callback if the full path to the project matches
		// the regular expression.
		for _, p := range ps {
			if r.MatchString(p.PathWithNamespace) {
				more, err := f(g, p)
				if err != nil {
					return err
				}
				if !more {
					return nil
				}
			}
		}

		// Check if done.
		if resp.NextPage == 0 {
			break
		}

		// Move to the next page.
		opts.Page = resp.NextPage
	}

	return nil
}

// GetAllProjects returns all the projects in a group recursively (or
// not) for each project whose full path name matches the regular
// expression.  An empty regular expression matches any string.
// Prefer ForEachProjectInGroup() over this function to avoid the long
// delay while waiting to collect all the projects.  The main reason
// to use this function is when deleting projects because Gitlab's
// paging gets confused because Gitlab's paging is relative to when
// you make the request for the next page, not when you made the
// request for the first page, and deleting projects necessarily
// changes the page on which some remaining projects appear.  This
// function is better to use when deleting projects because it
// collects all the projects up front allowing the caller to delete
// them with impunity because there will be no next page to get.
func GetAllProjects(
	s *gitlab.GroupsService,
	group string,
	expr string,
	recursive bool,
) ([]*gitlab.Project, error) {

	var result []*gitlab.Project

	// Callback function used to collect all of the projects.
	f := func(group *gitlab.Group, project *gitlab.Project) (bool, error) {
		result = append(result, project)
		return true, nil
	}

	// Collect all the projects.
	err := ForEachProjectInGroup(s, group, expr, recursive, f)
	if err != nil {
		return nil, fmt.Errorf("GetAllProjects: %w", err)
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////
// Approval Rules
////////////////////////////////////////////////////////////////////////

// ApprovalRuleToString converts the approval rule into a
// human-readable string.
func ApprovalRuleToString(rule *gitlab.ProjectApprovalRule) string {
	var usernames []string

	// Extract the usernames of the eligible approvers.
	for _, u := range rule.EligibleApprovers {
		usernames = append(usernames, u.Username)
	}

	// Sort the usernames.
	slices.Sort(usernames)

	// Get the string representation of the usernames.
	usernamesAsString := fmt.Sprintf("%q", usernames)

	// Calculate the CRC-64 checksum of the usernames string.
	cksum := crc64.Checksum(
		[]byte(usernamesAsString),
		crc64.MakeTable(crc64.ISO))

	// Add rule ID and name.
	return fmt.Sprintf("%#016x  %6d  %-16s  %s",
		cksum, rule.ID, rule.Name, usernamesAsString)
}

// updateApprovalRule updates the approval rule for the project to
// have the same values as before except with a new list of user IDs.
// This function is designed to be the callback for
// [ForEachApprovalRuleInProject()].
func UpdateApprovalRule(
	s *gitlab.ProjectsService,
	projectID int,
	rule *gitlab.ProjectApprovalRule,
	userIDs []int,
) error {
	var err error
	
	// Extract the existing group IDs.
	var groupIDs []int
	for _, group := range rule.Groups {
		groupIDs = append(groupIDs, group.ID)
	}

	// Extract the existing branch IDs.
	var branchIDs []int
	for _, branch := range rule.ProtectedBranches {
		branchIDs = append(branchIDs, branch.ID)
	}

	// Set update options.
	opts := gitlab.UpdateProjectLevelRuleOptions{
		Name: gitlab.Ptr(rule.Name),
		ApprovalsRequired: gitlab.Ptr(rule.ApprovalsRequired),
		UserIDs: &userIDs,
		GroupIDs: &groupIDs,
		ProtectedBranchIDs: &branchIDs,
		AppliesToAllProtectedBranches: gitlab.Ptr(rule.AppliesToAllProtectedBranches),
	}

	// Update the approval rule.
	_, _, err = s.UpdateProjectApprovalRule(projectID, rule.ID, &opts)
	return err
}

// ApprovalRulesGetter is an abstraction of GetProjectApprovalRules()
// in gitlab.ProjectsService which was added so
// ForEachApprovalRuleInProject() can be tested with requiring a paid
// Gitlab account because Gitlab CE (the free version of Gitlab) does
// not support approval rules.
type ApprovalRulesGetter interface {
	GetProjectApprovalRules(
		pid interface{},
		opt *gitlab.GetProjectApprovalRulesListsOptions,
		options ...gitlab.RequestOptionFunc,
	) ([]*gitlab.ProjectApprovalRule, *gitlab.Response, error)
}

// ForEachApprovalRuleInProject iterates over the approval rules in a
// project and calls the function f once for each approval rule.  The
// function f must return true and no error to indicate that it wants
// to continue being called with the remaining projects.  If f returns
// an error, it will be forwarded to the caller as the error return
// value for this function.
func ForEachApprovalRuleInProject(
	s ApprovalRulesGetter, /* was *gitlab.ProjectsService */
	p *gitlab.Project,
	f func(
		approvalRule *gitlab.ProjectApprovalRule,
	) (bool, error),
) error {

	// Set up the options for ListGroupProjects().
	opts := gitlab.GetProjectApprovalRulesListsOptions{}
	opts.Page = 1
	///opts.PerPage = 100

	// Iterate over each page of approval rules.
	for {

		// Get the next page of approval rules.
		rules, resp, err := s.GetProjectApprovalRules(p.ID, &opts)
		if err != nil {
			return fmt.Errorf("ForEachApprovalRuleInProject: %w\n", err)
		}

		// Invoke the callbacks.
		for _, rule := range rules {
			more, err := f(rule)
			if err != nil {
				return err
			}
			if !more {
				return nil
			}
		}

		// Check if done.
		if resp.NextPage == 0 {
			break
		}

		// Move to the next page.
		opts.Page = resp.NextPage
	}

	return nil
}

////////////////////////////////////////////////////////////////////////
// Users
////////////////////////////////////////////////////////////////////////

// FindExactUser search for the user and returns the user that exactly
// matches the search string.  The search string can be the user ID,
// name, username or e-mail address of the user.
func FindExactUser(
	s *gitlab.UsersService,
	user string,
	date time.Time,
) (*gitlab.User, error) {
	var err error
	var exactMatches []*gitlab.User
	var u *gitlab.User
	var userID int

	// If "user" is an integer, it is a user ID which requires
	// different processing.
	userID, err = strconv.Atoi(user)
	if err == nil {
		opts := gitlab.GetUsersOptions{}
		u, _, err = s.GetUser(userID, opts)
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	err = nil

	// Iterate over all the users that match the "user" string.
	err = ForEachUser(s, user, date, func(u *gitlab.User) (bool, error) {
		if u.Email == user || u.Username == user || u.Name == user {
			exactMatches = append(exactMatches, u)
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	if len(exactMatches) == 0 {
		return nil, fmt.Errorf("no match found for user: %q", user)
	}
	if len(exactMatches) > 1 {
		var usernames []string
		for _, exactMatch := range exactMatches {
			usernames = append(usernames, exactMatch.Username)
		}
		return nil, fmt.Errorf("multiple exact matches found: %q", usernames)
	}

	return exactMatches[0], nil

}

// ForEachUser iterates over users calling the function f once for
// each user matching the search string.  An empty search string
// matches all users.  The search string can be the name, username, or
// e-mail address of the user.  The function f must return true and no
// error to indicate that it wants to continue being called with the
// remaining users.  If f returns an error, it will be forwarded to
// the caller as the error return value for this function.
//
// Also see [FindExactUser()].
func ForEachUser(
	s *gitlab.UsersService,
	user string,
	date time.Time,
	f func(user *gitlab.User) (bool, error),
) error {

	// Set up the options for ListUsers().
	opts := gitlab.ListUsersOptions{}
	opts.CreatedAfter = &date
	if user != "" {
		opts.Search = &user
	}
	opts.Page = 1
	///opts.PerPage = 100

	// Iterate over each page of users.
	for {

		// Get the next page of users.
		users, resp, err := s.ListUsers(&opts)
		if err != nil {
			return fmt.Errorf("ForEachUser: %w\n", err)
		}

		// Invoke the callback for each user.
		for _, user := range users {
			more, err := f(user)
			if err != nil {
				return err
			}
			if !more {
				return nil
			}
		}

		// Check if done.
		if resp.NextPage == 0 {
			break
		}

		// Move to the next page.
		opts.Page = resp.NextPage
	}

	return nil
}
