package gitlab_util

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Stubs
////////////////////////////////////////////////////////////////////////

type GitlabProjectsServiceStub struct {}

func (s *GitlabProjectsServiceStub)	GetProjectApprovalRules(
	pid interface{},
	opt *gitlab.GetProjectApprovalRulesListsOptions,
	options ...gitlab.RequestOptionFunc,
) ([]*gitlab.ProjectApprovalRule, *gitlab.Response, error) {

	resp := gitlab.Response {
		NextPage: 0,
		}

	rules := []*gitlab.ProjectApprovalRule{
		&gitlab.ProjectApprovalRule{
			ID: 1,
			Name: "Rule1",
			EligibleApprovers: []*gitlab.BasicUser{
				&gitlab.BasicUser{
					ID: 1,
					Username: "aberns",
				},
				&gitlab.BasicUser{
					ID: 2,
					Username: "bcrocket",
				},
			},
		},
		&gitlab.ProjectApprovalRule{
			ID: 2,
			Name: "Rule2",
			EligibleApprovers: []*gitlab.BasicUser{
				&gitlab.BasicUser{
					ID: 1,
					Username: "aberns",
				},
				&gitlab.BasicUser{
					ID: 2,
					Username: "bcrocket",
				},
			},
		},
	}

	return rules, &resp, nil
}

////////////////////////////////////////////////////////////////////////
// Functions
////////////////////////////////////////////////////////////////////////

func collectApprovalRules(rule *gitlab.ProjectApprovalRule) string {
	var result strings.Builder

	// Add rule ID and name.
	result.WriteString(fmt.Sprintf("%v: %v: ", rule.ID, rule.Name))

	// Iterate over the eligable approvers.
	result.WriteString("[")
	for i := 0; i < len(rule.EligibleApprovers); i++ {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("(%v, %v)",
			rule.EligibleApprovers[i].ID,
			rule.EligibleApprovers[i].Username))
	}
	result.WriteString("]")

	return result.String()
}

////////////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////////////

func TestForEachApprovalRuleInProject(t *testing.T) {
	service := GitlabProjectsServiceStub{}
	p := gitlab.Project{}
	var actual []string
	expected := []string{
		"1: Rule1: [(1, aberns), (2, bcrocket)]",
		"2: Rule2: [(1, aberns), (2, bcrocket)]",
	}
	
	ForEachApprovalRuleInProject(
		&service, &p,
		func (rule *gitlab.ProjectApprovalRule) (bool, error) {
			actual = append(actual, collectApprovalRules(rule))
			return true, nil
		})

	if !slices.Equal(actual, expected) {
		t.Errorf("ForEachApprovalRuleInProject: expected=%v  actual=%v",
			expected, actual)

	}
}
