// This file provides utility functions related to the Gitlab REST API.

package gitlab_util

import (
	"fmt"

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

// ForEachProjectInGroup calls the function f once for each project in
// the group.  If the function f must return true to indicate that it
// wants to continue being called with the remaining projects.
func ForEachProjectInGroup(
	s *gitlab.GroupsService,
	group string,
	f func (group *gitlab.Group, project *gitlab.Project) bool,
) error {

	// Find the group.
	g, err := FindExactGroup(s, group)
	if err != nil {
		return err
	}
	
	// Set up the options for ListGroupProjects().
	opts := gitlab.ListGroupProjectsOptions{}
	opts.IncludeSubGroups = gitlab.Ptr(true)
	opts.Page = 1

	// Iterate over each page of groups.
	for {

		// Get the next page of projects.
		ps, resp, err := s.ListGroupProjects(g.ID, &opts)
		if err != nil {
			return fmt.Errorf("ForEachProjectInGroup: %w\n", err)
		}

		// Invoke the callback.
		for _, p := range ps {
			if !f(g, p) {
				goto out
			}
		}

		// Check if done.
		if resp.NextPage == 0 {
			break
		}

		// Move to the next page.
		opts.Page = resp.NextPage
	}

out:

	return nil
}
