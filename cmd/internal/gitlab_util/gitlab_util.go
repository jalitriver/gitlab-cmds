// This file provides utility functions related to the Gitlab REST API.

package gitlab_util

import (
	"fmt"
	"regexp"

	"github.com/xanzy/go-gitlab"
)

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

// ForEachProjectInGroup iterates over the projects in a group and
// (recursively or not) calls the function f once for each project
// whose full path name matches the regular expression.  Any empty
// regular expression matches any string.  The function f must return
// true and no error to indicate that it wants to continue being
// called with the remaining projects.  If f returns an error, it will
// be forwarded to the caller as the error return value for this
// function.
func ForEachProjectInGroup(
	s *gitlab.GroupsService,
	group string,
	expr string,
	recursive bool,
	f func (group *gitlab.Group, project *gitlab.Project) (bool, error),
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
