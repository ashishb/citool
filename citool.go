package main

import (
	"flag"
	"fmt"
	"robpike.io/filter"
	"strings"
)

var mode = flag.String("mode",
	"analyze",
	"Mode - \"download\" or \"analyze\"")

var branchFilter = flag.String("only-branch",
	"",
	"Only consider test results from this branch. Analyze mode only.")

var testnameFilter = flag.String("only-testname",
	"",
	"Only consider test results for this testname. Analyze mode only.")

var testStatusFilter = flag.String("only-testresult",
	"",
	"Only consider test results with this completion status. Analyze mode only.")

var inputFiles = flag.String("input-files",
	"",
	"Comma-separated list of files containing downloaded test results from CircleCI. Analyze mode only.")

var circleCiToken = flag.String("circle-token",
	"",
	"Circle CI access token. Download mode only.")

var downloadStartOffset = flag.Int("offset",
	0,
	"Circle CI build results download start offset")

var downloadLimit = flag.Int("limit",
	0,
	"Circle CI build results download limit")

const debugMode = true

func main() {
	flag.Parse()

	if *mode == "analyze" {
		files := strings.Split(*inputFiles, ",")
		// Treat non-positional args as input files as well
		files = append(files, flag.Args()...)
		testResults := getCircleCiBuildResults(&files)
		filterData(&testResults)
		PrintTestStats(testResults)
	} else if *mode == "download" {
		DownloadCircleCIBuildResults(*circleCiToken, *downloadStartOffset, *downloadLimit)
	} else {
		panic(fmt.Sprintf("Unexpected mode \"%s\"", *mode))
	}
}

func getCircleCiBuildResults(files *[]string) []CircleCiBuildResult {
	data := make([]CircleCiBuildResult, 0)
	for _, file := range *files {
		LogDebug("Input file: " + file)
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
		LogDebug("Filtering on branch: " + *branchFilter)
	}
	if len(*testnameFilter) > 0 {
		LogDebug("Filtering on test name: " + *testnameFilter)
	}
	if len(*testStatusFilter) > 0 {
		LogDebug("Filtering on test result: " + *testStatusFilter)
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

func LogDebug(msg string) {
	if !debugMode {
		return
	}
	fmt.Println(msg)
}
