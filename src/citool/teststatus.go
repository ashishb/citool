package citool

type TestStatusFilterTypes string

// Only for filtering
// https://circleci.com/docs/api/#recent-builds-across-all-projects
const (
	TestCompleted  TestStatusFilterTypes = "completed"
	TestSuccessful TestStatusFilterTypes = "successful"
	TestFailed     TestStatusFilterTypes = "failed"
	TestRunning    TestStatusFilterTypes = "running"
)

func GetTestStatusFilterOrFail(status string) TestStatusFilterTypes {
	switch status {
	case string(TestCompleted):
		return TestCompleted
	case string(TestSuccessful):
		return TestSuccessful
	case string(TestFailed):
		return TestFailed
	case string(TestRunning):
		return TestRunning
	default:
		panic("Unexpected test status value: " + status)
	}
}

type TestStatusType string

// Only for the status field inside the test result JSON structure
// :retried, :canceled, :infrastructure_fail, :timedout, :not_run, :running,
// :failed, :queued, :scheduled, :not_running, :no_tests, :fixed, :success
const (
	TestStatusRetried            TestStatusType = "retried"
	TestStatusCanceled           TestStatusType = "canceled"
	TestStatusInfrastructureFail TestStatusType = "infrastructure_fail"
	TestStatusTimedOut           TestStatusType = "timedout"
	TestStatusNotRun             TestStatusType = "not_run"
	TestStatusRunning            TestStatusType = "running"
	TestStatusFailed             TestStatusType = "failed"
	TestStatusQueued             TestStatusType = "queued"
	TestStatusScheduled          TestStatusType = "scheduled"
	TestStatusNotRunning         TestStatusType = "not_running"
	TestStatusNoTests            TestStatusType = "no_tests"
	TestStatusFixed              TestStatusType = "fixed"
	TestStatusSuccess            TestStatusType = "success"
)
