package gitlab_util

import (
	"slices"
	"testing"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Stubs
////////////////////////////////////////////////////////////////////////

type GitlabProjectsServiceStub struct{}

func (s *GitlabProjectsServiceStub) GetProjectApprovalRules(
	pid interface{},
	opt *gitlab.GetProjectApprovalRulesListsOptions,
	options ...gitlab.RequestOptionFunc,
) ([]*gitlab.ProjectApprovalRule, *gitlab.Response, error) {

	resp := gitlab.Response{
		NextPage: 0,
	}

	rules := []*gitlab.ProjectApprovalRule{
		&gitlab.ProjectApprovalRule{
			ID:   1,
			Name: "Rule1",
			EligibleApprovers: []*gitlab.BasicUser{
				&gitlab.BasicUser{
					ID:       1,
					Username: "aberns",
				},
				&gitlab.BasicUser{
					ID:       2,
					Username: "bcrocket",
				},
			},
		},
		&gitlab.ProjectApprovalRule{
			ID:   2,
			Name: "Rule2",
			EligibleApprovers: []*gitlab.BasicUser{
				&gitlab.BasicUser{
					ID:       3,
					Username: "cdragun",
				},
				&gitlab.BasicUser{
					ID:       4,
					Username: "delliot",
				},
			},
		},
	}

	return rules, &resp, nil
}

////////////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////////////

func TestForEachApprovalRuleInProject(t *testing.T) {
	var err error
	service := GitlabProjectsServiceStub{}
	p := gitlab.Project{}
	var actual []string
	expected := []string{
		`0xcac460d19ffbb714       1  Rule1             ["aberns" "bcrocket"]`,
		`0x056daac148e9b0a1       2  Rule2             ["cdragun" "delliot"]`,
	}

	err = ForEachApprovalRuleInProject(
		&service, &p,
		func(rule *gitlab.ProjectApprovalRule) (bool, error) {
			actual = append(actual, ApprovalRuleToString(rule))
			return true, nil
		})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !slices.Equal(actual, expected) {
		t.Errorf("ForEachApprovalRuleInProject: expected=%v  actual=%v",
			expected, actual)

	}
}
