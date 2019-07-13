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
	Status      JobStatusType         `json:"status"`
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

// Aggregated job information
type AggregateJobInfo struct {
	JobName            string
	Frequency          int
	CumulativeDuration time.Duration
	SuccessCount       int32
	FailureCount       int32
}

func PrintJobStats(results []CircleCiBuildResult) {
	fmt.Printf("Number of job results: %d\n", len(results))
	aggregateJobInfo := make(map[string]*AggregateJobInfo, 0)
	for _, result := range results {
		jobName := result.Workflows.JobName
		status := result.Status
		// Only filter on success/failure for now
		if status != "success" && status != "failed" {
			continue
		}
		duration := getJobDuration(result)
		existingAggregateJobInfo, present := aggregateJobInfo[jobName]
		if !present {
			existingAggregateJobInfo = &AggregateJobInfo{jobName, 0, time.Duration(0), 0, 0}
		}
		existingAggregateJobInfo.Frequency += 1
		existingAggregateJobInfo.CumulativeDuration += duration
		aggregateJobInfo[jobName] = existingAggregateJobInfo
		aggregateJobInfo[jobName].Frequency = aggregateJobInfo[jobName].Frequency + 1
		if status == JobStatusSuccess {
			existingAggregateJobInfo.SuccessCount += 1
		} else if status == JobStatusFailed {
			existingAggregateJobInfo.FailureCount += 1
		} else {
			panic("Unexpected status: " + status)
		}
	}

	values := make([]*AggregateJobInfo, 0, len(aggregateJobInfo))
	for _, v := range aggregateJobInfo {
		values = append(values, v)
	}

	printJobSuccessRate(values)
	fmt.Println("")
	printJobDuration(values)
	fmt.Println("")
	printTimeSeriesData(results)
}

func printJobDuration(aggregateJobInfo []*AggregateJobInfo) {
	sort.Slice(aggregateJobInfo, func(i, j int) bool {
		averageDuration1 := time.Duration(aggregateJobInfo[i].CumulativeDuration.Nanoseconds() / int64(aggregateJobInfo[i].Frequency))
		averageDuration2 := time.Duration(aggregateJobInfo[j].CumulativeDuration.Nanoseconds() / int64(aggregateJobInfo[j].Frequency))
		// Slowest job first
		return averageDuration1 > averageDuration2
	})
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "Job name\tAverage job duration")
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "----------\t--------------------")
	for _, v := range aggregateJobInfo {
		averageDuration := time.Duration(v.CumulativeDuration.Nanoseconds() / int64(v.Frequency))
		// We don't need accuracy below one second.
		averageDuration = averageDuration.Round(time.Second)
		//noinspection GoUnhandledErrorResult
		fmt.Fprintln(writer, fmt.Sprintf("%s\t%v", v.JobName, averageDuration))
	}
	//noinspection GoUnhandledErrorResult
	writer.Flush()
}

func printJobSuccessRate(aggregateJobInfo []*AggregateJobInfo) {
	sort.Slice(aggregateJobInfo, func(i, j int) bool {
		// Highest failure rate first.
		return (aggregateJobInfo[i].SuccessCount*aggregateJobInfo[j].FailureCount <
			aggregateJobInfo[j].SuccessCount*aggregateJobInfo[i].FailureCount)
	})

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "Job name\tSuccess Rate")
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "----------\t-----------")
	for _, v := range aggregateJobInfo {
		successRate := int32(0)
		if v.SuccessCount+v.FailureCount > 0 {
			successRate = (100 * v.SuccessCount) / (v.SuccessCount + v.FailureCount)
		}
		//noinspection GoUnhandledErrorResult
		fmt.Fprintln(writer, fmt.Sprintf("%s\t%d/%d (%d%%)",
			v.JobName, v.SuccessCount, v.SuccessCount+v.FailureCount, successRate))
	}
	//noinspection GoUnhandledErrorResult
	writer.Flush()
}

type StartTimeAndDurationPair struct {
	StartTime time.Time     // in what units?
	Duration  time.Duration // in nanoseconds
}

func printTimeSeriesData(results []CircleCiBuildResult) {
	jobDurationsInSeconds := make(map[string][]StartTimeAndDurationPair, 0)
	for _, result := range results {
		// Only consider successful jobs to avoid skew due to failed job which might fail early on.
		if result.Status != JobStatusSuccess {
			continue
		}
		jobName := result.Workflows.JobName
		startTime := result.StartTime
		duration := getJobDuration(result)
		startTimeAndDurationPair := StartTimeAndDurationPair{
			StartTime: getTime(startTime),
			Duration:  duration}
		jobDurationsInSeconds[jobName] = append(
			jobDurationsInSeconds[jobName], startTimeAndDurationPair)
	}
	for key, value := range jobDurationsInSeconds {
		// Sort
		sort.Slice(value, func(i, j int) bool {
			// chronological order
			return value[i].StartTime.Sub(value[j].StartTime).Seconds() < 0
		})
		durations := make([]float64, len(value))
		for i, value := range value {
			durations[i] = value.Duration.Seconds()
		}
		fmt.Printf("\nJob name: %s (%d data points)\n\n", key, len(durations))
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

func getJobDuration(buildResult CircleCiBuildResult) time.Duration {
	startTime := getTime(buildResult.StartTime)
	endTime := getTime(buildResult.EndTime)
	return endTime.Sub(startTime)
}
