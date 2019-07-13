package citool

import (
	"encoding/json"
	"fmt"
	"github.com/guptarohit/asciigraph"
	"io/ioutil"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

type CircleCiBuildWorkflow struct {
	JobName string `json:"job_name"`
}

type CircleCiBuildResult struct {
	Username    string                `json:"username"`
	Reponame    string                `json:"reponame"`
	Branch      string                `json:"branch"`
	BuildNumber int                   `json:"build_num"`
	Status      TestStatusType        `json:"status"`
	EndTime     string                `json:"stop_time"`
	StartTime   string                `json:"start_time"`
	Workflows   CircleCiBuildWorkflow `json:"workflows"`
}

func GetJson(filename string) []CircleCiBuildResult {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Unable to read file \"%s\"", filename))
	}
	var circleCiBuildResults []CircleCiBuildResult
	err2 := json.Unmarshal(contents, &circleCiBuildResults)
	if err2 != nil {
		panic("Failed to extract JSON" + err2.Error())
	}
	return circleCiBuildResults
}

// Aggregated test information
type AggregateTestInfo struct {
	TestName           string
	Frequency          int
	CumulativeDuration time.Duration
	SuccessCount       int32
	FailureCount       int32
}

func PrintTestStats(results []CircleCiBuildResult) {
	fmt.Printf("\nNumber of test results: %d\n", len(results))
	aggregateTestInfo := make(map[string]*AggregateTestInfo, 0)
	for _, result := range results {
		testName := result.Workflows.JobName
		status := result.Status
		// Only filter on success/failure for now
		if status != "success" && status != "failed" {
			continue
		}
		duration := getTestDuration(result)
		existingAggregateTestInfo, present := aggregateTestInfo[testName]
		if !present {
			existingAggregateTestInfo = &AggregateTestInfo{testName, 0, time.Duration(0), 0, 0}
		}
		existingAggregateTestInfo.Frequency += 1
		existingAggregateTestInfo.CumulativeDuration += duration
		aggregateTestInfo[testName] = existingAggregateTestInfo
		aggregateTestInfo[testName].Frequency = aggregateTestInfo[testName].Frequency + 1
		if status == TestStatusSuccess {
			existingAggregateTestInfo.SuccessCount += 1
		} else if status == TestStatusFailed {
			existingAggregateTestInfo.FailureCount += 1
		} else {
			panic("Unexpected status: " + status)
		}
	}

	values := make([]*AggregateTestInfo, 0, len(aggregateTestInfo))
	for _, v := range aggregateTestInfo {
		values = append(values, v)
	}

	printTestSuccessRate(values)
	fmt.Println("")
	printTestDuration(values)
	fmt.Println("")
	printTimeSeriesData(results)
}

func printTestDuration(aggregateTestInfo []*AggregateTestInfo) {
	sort.Slice(aggregateTestInfo, func(i, j int) bool {
		averageDuration1 := time.Duration(aggregateTestInfo[i].CumulativeDuration.Nanoseconds() / int64(aggregateTestInfo[i].Frequency))
		averageDuration2 := time.Duration(aggregateTestInfo[j].CumulativeDuration.Nanoseconds() / int64(aggregateTestInfo[j].Frequency))
		// Slowest test first
		return averageDuration1 > averageDuration2
	})
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "Test name\tAverage test duration")
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "----------\t--------------------")
	for _, v := range aggregateTestInfo {
		averageDuration := time.Duration(v.CumulativeDuration.Nanoseconds() / int64(v.Frequency))
		// We don't need accuracy below one second.
		averageDuration = averageDuration.Round(time.Second)
		//noinspection GoUnhandledErrorResult
		fmt.Fprintln(writer, fmt.Sprintf("%s\t%v", v.TestName, averageDuration))
	}
	//noinspection GoUnhandledErrorResult
	writer.Flush()
}

func printTestSuccessRate(aggregateTestInfo []*AggregateTestInfo) {
	sort.Slice(aggregateTestInfo, func(i, j int) bool {
		// Highest failure rate first.
		return (aggregateTestInfo[i].SuccessCount*aggregateTestInfo[j].FailureCount <
			aggregateTestInfo[j].SuccessCount*aggregateTestInfo[i].FailureCount)
	})

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "Test name\tSuccess Rate")
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "----------\t-----------")
	for _, v := range aggregateTestInfo {
		successRate := int32(0)
		if v.SuccessCount+v.FailureCount > 0 {
			successRate = (100 * v.SuccessCount) / (v.SuccessCount + v.FailureCount)
		}
		//noinspection GoUnhandledErrorResult
		fmt.Fprintln(writer, fmt.Sprintf("%s\t%d/%d (%d%%)",
			v.TestName, v.SuccessCount, v.SuccessCount+v.FailureCount, successRate))
	}
	//noinspection GoUnhandledErrorResult
	writer.Flush()
}

type StartTimeAndDurationPair struct {
	StartTime time.Time     // in what units?
	Duration  time.Duration // in nanoseconds
}

func printTimeSeriesData(results []CircleCiBuildResult) {
	testDurationsInSeconds := make(map[string][]StartTimeAndDurationPair, 0)
	for _, result := range results {
		// Only consider successful jobs to avoid skew due to failed job which might fail early on.
		if result.Status != TestStatusSuccess {
			continue
		}
		testName := result.Workflows.JobName
		startTime := result.StartTime
		duration := getTestDuration(result)
		startTimeAndDurationPair := StartTimeAndDurationPair{
			StartTime: getTime(startTime),
			Duration:  duration}
		testDurationsInSeconds[testName] = append(
			testDurationsInSeconds[testName], startTimeAndDurationPair)
	}
	for key, value := range testDurationsInSeconds {
		// Sort
		sort.Slice(value, func(i, j int) bool {
			// chronological order
			return value[i].StartTime.Sub(value[j].StartTime).Seconds() < 0
		})
		durations := make([]float64, len(value))
		for i, value := range value {
			durations[i] = value.Duration.Seconds()
		}
		fmt.Printf("\nTest name: %s (%d data points)\n\n", key, len(durations))
		graph := asciigraph.Plot(durations,
			asciigraph.Height(10), asciigraph.Width(100))
		fmt.Println(graph)
	}
}

func getTime(timeString string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339Nano, timeString)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse time \"%v\", error: %v", timeString, err))
	}
	return parsedTime
}

func getTestDuration(buildResult CircleCiBuildResult) time.Duration {
	startTime := getTime(buildResult.StartTime)
	endTime := getTime(buildResult.EndTime)
	return endTime.Sub(startTime)
}
