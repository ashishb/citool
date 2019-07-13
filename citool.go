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

var testname = flag.String("testname",
	"",
	"Only consider test results for this testname. Analyze mode only.")

var testStatus = flag.String("testresult",
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

var vcsType = flag.String("vcsType",
	"github",
	"Name of the VCS system - See https://circleci.com/docs/api/#version-control-systems-vcs-type. Download mode only.")

var username = flag.String("username",
	"",
	"Optional username to filter downloads/analysis on",
)

var repositoryName = flag.String("reponame",
	"",
	"Optional repository name to filter downloads/analysis on")

var branchName = flag.String("branch",
	"",
	"Optional branch name to filter download/analysis on")

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
		downloadParams := DownloadParams{
			CircleToken:    circleCiToken,
			VcsType:        vcsType,
			Username:       username,
			RepositoryName: repositoryName,
			BranchName:     branchName,
			Start:          *downloadStartOffset,
			Limit:          *downloadLimit}
		DownloadCircleCIBuildResults(downloadParams)
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
	if !IsEmpty(username) {
		LogDebug("Filtering on username: " + *username)
	}
	if !IsEmpty(repositoryName) {
		LogDebug("Filtering on repository name: " + *repositoryName)
	}
	if !IsEmpty(branchName) {
		LogDebug("Filtering on branch: " + *branchName)
	}
	if !IsEmpty(testname) {
		LogDebug("Filtering on test name: " + *testname)
	}
	if !IsEmpty(testStatus) {
		LogDebug("Filtering on test result: " + *testStatus)
	}
	filter.ChooseInPlace(results, filterRule)
}

func filterRule(result CircleCiBuildResult) bool {
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
	if !IsEmpty(testname) {
		if result.Workflows.JobName != *testname {
			return false
		}
	}
	if !IsEmpty(testStatus) {
		if result.Status != *testStatus {
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
