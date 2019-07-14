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

type AnalyzeParams struct {
	PrintJobSuccessRate         bool
	PrintJobDurationInAggregate bool
	PrintJobDurationTimeSeries  bool
}

func PrintJobStats(results []CircleCiBuildResult, params AnalyzeParams) {
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

	if params.PrintJobSuccessRate {
		printJobSuccessRate(values)
		fmt.Println("")
	}
	if params.PrintJobDurationInAggregate {
		printJobDuration(values)
		fmt.Println("")
	}
	if params.PrintJobDurationTimeSeries {
		printTimeSeriesData(results)
	}
}

func printJobDuration(aggregateJobInfo []*AggregateJobInfo) {
	sort.Slice(aggregateJobInfo, func(i, j int) bool {
		averageDuration1 := time.Duration(aggregateJobInfo[i].CumulativeDuration.Nanoseconds() / int64(aggregateJobInfo[i].Frequency))
		averageDuration2 := time.Duration(aggregateJobInfo[j].CumulativeDuration.Nanoseconds() / int64(aggregateJobInfo[j].Frequency))
		if averageDuration1 != averageDuration2 {
			// Slowest job first
			return averageDuration1 > averageDuration2
		}
		// If the jobs take same time then sort on the basis of name to have stable outcome
		return aggregateJobInfo[i].JobName > aggregateJobInfo[j].JobName
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
		comparison := (aggregateJobInfo[i].SuccessCount*aggregateJobInfo[j].FailureCount -
			aggregateJobInfo[j].SuccessCount*aggregateJobInfo[i].FailureCount)
		if comparison != 0 {
			return comparison < 0
		} else {
			// If the jobs have the same success rate then sort on the basis of name to have stable outcome
			return aggregateJobInfo[i].JobName > aggregateJobInfo[j].JobName
		}
	})

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "Job name\tSuccess Rate")
	//noinspection GoUnhandledErrorResult
	fmt.Fprintln(writer, "--------\t-----------")
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

const avgRate = 10
const graphHeight = 20 // lines
const graphWidth = 100 // characters

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
		// Get the durations from the chronologically sorted array.
		durations := make([]float64, len(value))
		for i, value := range value {
			durations[i] = value.Duration.Seconds()
		}
		durations = getMovingAverage(durations, avgRate)
		fmt.Printf("\nJob name: %s (%d data points)\n\n", key, len(durations))
		graph := asciigraph.Plot(durations,
			asciigraph.Height(graphHeight), asciigraph.Width(graphWidth))
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

func getMovingAverage(source []float64, numPoints int) []float64 {
	sourceNumOfPoints := len(source)
	if sourceNumOfPoints < numPoints {
		panic("This method should not have been called since we have less than " +
			string(numPoints) + " of points")
	}
	LogDebug(fmt.Sprintf("Size: %d", sourceNumOfPoints-(numPoints-1)))
	target := make([]float64, sourceNumOfPoints-(numPoints-1))
	for i := range target {
		target[i] = sum(source[i:i+numPoints]) / float64(numPoints)
		LogDebug(fmt.Sprintf("Average from %d to %d: %f", i, i+numPoints, target[i]))
	}
	return target
}

func sum(arr []float64) float64 {
	result := float64(0)
	for _, v := range arr {
		result += v
	}
	return result
}
