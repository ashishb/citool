package citool

type TestStatusTypes string

// https://circleci.com/docs/api/#recent-builds-across-all-projects
const (
    TestCompleted  TestStatusTypes = "completed"
    TestSuccessful TestStatusTypes = "successful"
    TestFailed     TestStatusTypes = "failed"
    TestRunning    TestStatusTypes = "running"
)

func GetTestStatusOrFail(status string) TestStatusTypes {
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
