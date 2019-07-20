package citool

import (
	"robpike.io/filter"
)

// FilterParams are parameters used for filtering results at the time of analysis.
type FilterParams struct {
	Username       *string
	RepositoryName *string
	BranchName     *string
	JobName        *string
	JobStatus      *string
	Start          int
	Limit          int
}

// FilterData filters the results field in-place using filterParams.
func (filterParams FilterParams) FilterData(results *[]CircleCiBuildResult) {
	username := filterParams.Username
	repositoryName := filterParams.RepositoryName
	branchName := filterParams.BranchName
	jobname := filterParams.JobName
	jobStatus := filterParams.JobStatus

	if !IsEmpty(username) {
		LogDebug("Filtering on username: " + *username)
	}
	if !IsEmpty(repositoryName) {
		LogDebug("Filtering on repository name: " + *repositoryName)
	}
	if !IsEmpty(branchName) {
		LogDebug("Filtering on branch: " + *branchName)
	}
	if !IsEmpty(jobname) {
		LogDebug("Filtering on job name: " + *jobname)
	}
	if !IsEmpty(jobStatus) {
		LogDebug("Filtering on job result: " + *jobStatus)
	}

	filter.ChooseInPlace(results, filterParams.filterRule)
}

func (filterParams FilterParams) filterRule(result CircleCiBuildResult) bool {
	username := filterParams.Username
	repositoryName := filterParams.RepositoryName
	branchName := filterParams.BranchName
	jobname := filterParams.JobName
	jobStatus := filterParams.JobStatus

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
	if !IsEmpty(jobname) {
		if result.Workflows.JobName != *jobname {
			return false
		}
	}
	if !IsEmpty(jobStatus) {
		if string(result.Status) != *jobStatus {
			return false
		}
	}
	return true
}
