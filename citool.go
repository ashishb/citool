package main

import (
	"encoding/json"
	"fmt"
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
	BuildNumber int                   `json:"build_num"`
	Status      string                `json:"status"`
	EndTime     string                `json:"stop_time"`
	StartTime   string                `json:"start_time"`
	Workflows   CircleCiBuildWorkflow `json:"workflows"`
}

func getJson(filename string) []CircleCiBuildResult {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic("Unable to read file " + filename)
	}
	var circleCiBuildResults []CircleCiBuildResult
	err2 := json.Unmarshal(contents, &circleCiBuildResults)
	if err2 != nil {
		panic("Failed to extract JSON" + err2.Error())
	}
	return circleCiBuildResults
}

func main() {
	data1 := getJson("./3.json")
	data2 := getJson("./4.json")
	data := append(data1, data2...)
	printTestStats(data)
}

// Aggregated test information
type AggregateTestInfo struct {
	Frequency          int
	CumulativeDuration time.Duration
}

type TestNameAndDurationPair struct {
	TestName string
	Duration time.Duration
}

type TestNameAndDurationPairList []TestNameAndDurationPair

func printTestStats(results []CircleCiBuildResult) {
	aggregateTestInfo := make(map[string]*AggregateTestInfo, 0)
	for _, result := range results {
		testName := result.Workflows.JobName
		status := result.Status
		// Only filter on success for now
		if status != "success" {
			continue
		}
		duration := getTestDuration(result)
		existingAggregateTestInfo, present := aggregateTestInfo[testName]
		if !present {
			existingAggregateTestInfo = &AggregateTestInfo{0, time.Duration(0)}
		}
		existingAggregateTestInfo.Frequency += 1
		existingAggregateTestInfo.CumulativeDuration += duration
		aggregateTestInfo[testName] = existingAggregateTestInfo
		aggregateTestInfo[testName].Frequency = aggregateTestInfo[testName].Frequency + 1
	}

	testNameAndDurationList := make(TestNameAndDurationPairList, 0)
	for testName, testInfo := range aggregateTestInfo {
		averageDuration := time.Duration(testInfo.CumulativeDuration.Nanoseconds() / int64(testInfo.Frequency))
		testNameAndDurationList = append(
			testNameAndDurationList, TestNameAndDurationPair{testName, averageDuration})
	}
	sort.Slice(testNameAndDurationList, func(i, j int) bool {
		return testNameAndDurationList[i].Duration > testNameAndDurationList[j].Duration
	})

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "Test name\tAverage test duration")
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "----------\t--------------------")
	for _, testNameAndDuration := range testNameAndDurationList {
		// We don't need accuracy below one second.
		averageDuration := testNameAndDuration.Duration.Round(time.Second)
		//noinspection GoUnhandledErrorResult
		fmt.Fprintln(writer, fmt.Sprintf("%s\t%v", testNameAndDuration.TestName, averageDuration))
	}
	//noinspection GoUnhandledErrorResult
	writer.Flush()
}

func getTime(timeString string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339Nano, timeString)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse parsedTime \"%v\", error: %v", timeString, err))
	}
	return parsedTime
}

func getTestDuration(buildResult CircleCiBuildResult) time.Duration {
	startTime := getTime(buildResult.StartTime)
	endTime := getTime(buildResult.EndTime)
	return endTime.Sub(startTime)
}
