package citool

// JobStatusFilterTypes enum values are the job status values used for filtering.
type JobStatusFilterTypes string

// Only for filtering
// https://circleci.com/docs/api/#recent-builds-across-all-projects
const (
	JobCompleted  JobStatusFilterTypes = "completed"
	JobSuccessful JobStatusFilterTypes = "successful"
	JobFailed     JobStatusFilterTypes = "failed"
	JobRunning    JobStatusFilterTypes = "running"
)

// GetJobStatusFilterOrFail converts the string value to enum type.
// Panics if the string value does not match any enum value.
func GetJobStatusFilterOrFail(status string) JobStatusFilterTypes {
	switch status {
	case string(JobCompleted):
		return JobCompleted
	case string(JobSuccessful):
		return JobSuccessful
	case string(JobFailed):
		return JobFailed
	case string(JobRunning):
		return JobRunning
	default:
		panic("Unexpected job status value: " + status)
	}
}

// JobStatusType represents the actual status of the job.
// These values are valid only for the status field inside the job result JSON structure
type JobStatusType string

// :retried, :canceled, :infrastructure_fail, :timedout, :not_run, :running,
// :failed, :queued, :scheduled, :not_running, :no_tests, :fixed, :success
const (
	JobStatusRetried            JobStatusType = "retried"
	JobStatusCanceled           JobStatusType = "canceled"
	JobStatusInfrastructureFail JobStatusType = "infrastructure_fail"
	JobStatusTimedOut           JobStatusType = "timedout"
	JobStatusNotRun             JobStatusType = "not_run"
	JobStatusRunning            JobStatusType = "running"
	JobStatusFailed             JobStatusType = "failed"
	JobStatusQueued             JobStatusType = "queued"
	JobStatusScheduled          JobStatusType = "scheduled"
	JobStatusNotRunning         JobStatusType = "not_running"
	JobStatusNoTests            JobStatusType = "no_tests"
	JobStatusFixed              JobStatusType = "fixed"
	JobStatusSuccess            JobStatusType = "success"
)
