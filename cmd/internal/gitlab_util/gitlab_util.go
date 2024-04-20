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

// FindUniqueGroupID determines if the group specified by the search
// is unique.  If so, it returns the group ID; otherwise, it returns
// an error.
func FindUniqueGroupID(s *gitlab.GroupsService, group string) (int, error) {

	// Set the group search string.
	grpopts := gitlab.ListGroupsOptions{
		Search: gitlab.Ptr(group),
	}

	// Get at least one page of matching groups.
	groups, _, err := s.ListGroups(&grpopts)
	if err != nil {
		err = fmt.Errorf("FindUniqueGroupID: %w", err)
		return 0, err
	}

	// Make sure exactly one group was found.
	if len(groups) == 0 {
		fmt.Errorf("FindUniqueGroupID: could not find group: %v", group)
		return 0, err
	}
	if len(groups) > 1 {
		err := fmt.Errorf(
			"FindUniqueGroupID: found multiple matching groups: %v",
			GroupFullPaths(groups))
		return 0, err
	}

	return groups[0].ID, nil
}
