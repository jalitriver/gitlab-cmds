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

// FindExactGroupID returns the ID of the group that exactly matches
// the search string.
func FindExactGroupID(s *gitlab.GroupsService, group string) (int, error) {

	// Set the group search string.
	opts := gitlab.ListGroupsOptions{}
	opts.Page = 1
	opts.Search = gitlab.Ptr(group)

	// Iterate over each page of groups.
	for {

		// Get a page of matching groups.
		gs, resp, err := s.ListGroups(&opts)
		if err != nil {
			err = fmt.Errorf("FindExactGroupID: %w", err)
			return 0, err
		}

		// Check each group for an exact match.
		for _, g := range gs {
			if g.FullPath == group {
				return g.ID, nil
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
		"FindExactGroupID: could not find exact match for group: %q", group)
	return 0, err
}
