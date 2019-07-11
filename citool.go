package main

import (
	"flag"
	"fmt"
	"robpike.io/filter"
	"strings"
)

var branchFilter = flag.String("only-branch",
	"",
	"Only consider test results from this branch")

var testnameFilter = flag.String("only-testname",
	"",
	"Only consider test results for this testname")

var testStatusFilter = flag.String("only-testresult",
	"",
	"Only consider test results with this completion status")

var inputFiles = flag.String("input-files",
	"",
	"Comma-separated list of files containing downloaded test results from CircleCI")

const debug = true

func main() {
	flag.Parse()

	files := strings.Split(*inputFiles, ",")
	testResults := getCircleCiBuildResults(&files)
	filterData(&testResults)
	PrintTestStats(testResults)
}

func getCircleCiBuildResults(files *[]string) []CircleCiBuildResult {
	data := make([]CircleCiBuildResult, 0)
	for _, file := range *files {
		logDebug("Input file: " + file)
		// Ignore empty file names
		if len(file) == 0 {
			continue
		}
		tmp := GetJson(file)
		data = append(data, tmp...)
	}
	return data
}

func filterData(results *[]CircleCiBuildResult) {
	if len(*branchFilter) > 0 {
		logDebug("Filtering on branch: " + *branchFilter)
	}
	if len(*testnameFilter) > 0 {
		logDebug("Filtering on test name: " + *testnameFilter)
	}
	if len(*testStatusFilter) > 0 {
		logDebug("Filtering on test result: " + *testStatusFilter)
	}
	filter.ChooseInPlace(results, filterRule)
}

func filterRule(result CircleCiBuildResult) bool {
	if len(*branchFilter) > 0 {
		if result.Branch != *branchFilter {
			return false
		}
	}
	if len(*testnameFilter) > 0 {
		if result.Workflows.JobName != *testnameFilter {
			return false
		}
	}
	if len(*testStatusFilter) > 0 {
		if result.Status != *testStatusFilter {
			return false
		}
	}
	return true
}

//func filterByBranchName(results []CircleCiBuildResult, branchName string) []CircleCiBuildResult {
//	newResults := make([]CircleCiBuildResult, 0)
//	for _, result := range results {
//		if result.Branch == branchName {
//
//		}
//	}
//}

func logDebug(msg string) {
	if !debug {
		return
	}
	fmt.Println(msg)
}
