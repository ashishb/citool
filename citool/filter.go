package citool

import (
	"robpike.io/filter"
)

type FilterParams struct {
	Username       *string
	RepositoryName *string
	BranchName     *string
	TestName       *string
	TestStatus     *string
	Start          int
	Limit          int
}

func (filterParams FilterParams) FilterData(results *[]CircleCiBuildResult) {
	username := filterParams.Username
	repositoryName := filterParams.RepositoryName
	branchName := filterParams.BranchName
	testname := filterParams.TestName
	testStatus := filterParams.TestStatus

	if !IsEmpty(username) {
		LogDebug("Filtering on username: " + *username)
	}
	if !IsEmpty(repositoryName) {
		LogDebug("Filtering on repository name: " + *repositoryName)
	}
	if !IsEmpty(branchName) {
		LogDebug("Filtering on branch: " + *branchName)
	}
	if !IsEmpty(testname) {
		LogDebug("Filtering on test name: " + *testname)
	}
	if !IsEmpty(testStatus) {
		LogDebug("Filtering on test result: " + *testStatus)
	}

	filter.ChooseInPlace(results, filterParams.filterRule)
}

func (filterParams FilterParams) filterRule(result CircleCiBuildResult) bool {
	username := filterParams.Username
	repositoryName := filterParams.RepositoryName
	branchName := filterParams.BranchName
	testname := filterParams.TestName
	testStatus := filterParams.TestStatus

	if !IsEmpty(username) {
		if result.Username != *username {
			return false
		}
	}
	if !IsEmpty(repositoryName) {
		if result.Reponame != *repositoryName {
			return false
		}
	}
	if !IsEmpty(branchName) {
		if result.Branch != *branchName {
			return false
		}
	}
	if !IsEmpty(testname) {
		if result.Workflows.JobName != *testname {
			return false
		}
	}
	if !IsEmpty(testStatus) {
		if result.Status != *testStatus {
			return false
		}
	}
	return true
}
