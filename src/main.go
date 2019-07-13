package main

import (
	"ci-analysis-tool/src/citool"
	"flag"
	"fmt"
	"strings"
)

var mode = flag.String("mode",
	"analyze",
	"Mode - \"download\" or \"analyze\"")

var testname = flag.String("testname",
	"",
	"Only consider test results for this testname. Analyze mode only.")

var testStatus = flag.String("test-status",
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

var downloadDirPath = flag.String("download-dir",
	"./circleci_data",
	"Directory to download Circle CI data to")

func main() {
	flag.Parse()

	if *mode == "analyze" {
		files := strings.Split(*inputFiles, ",")
		// Treat non-positional args as input files as well
		files = append(files, flag.Args()...)
		testResults := getCircleCiBuildResults(&files)
		filterParams := citool.FilterParams{
			Username:       username,
			RepositoryName: repositoryName,
			BranchName:     branchName,
			TestName:       testname,
			TestStatus:     testStatus}
		filterParams.FilterData(&testResults)
		citool.PrintTestStats(testResults)
	} else if *mode == "download" {
		var testStatusType *citool.TestStatusTypes = nil
		if !citool.IsEmpty(testStatus) {
			tmp := citool.TestStatusTypes(citool.GetTestStatusOrFail(*testStatus))
			testStatusType = &tmp
		}
		downloadParams := citool.DownloadParams{
			CircleToken:     circleCiToken,
			VcsType:         vcsType,
			Username:        username,
			RepositoryName:  repositoryName,
			BranchName:      branchName,
			Start:           *downloadStartOffset,
			Limit:           *downloadLimit,
			DownloadDirPath: *downloadDirPath,
			TestStatus:     testStatusType}
		citool.DownloadCircleCIBuildResults(downloadParams)
	} else {
		panic(fmt.Sprintf("Unexpected mode \"%s\"", *mode))
	}
}

func getCircleCiBuildResults(files *[]string) []citool.CircleCiBuildResult {
	data := make([]citool.CircleCiBuildResult, 0)
	for _, file := range *files {
		citool.LogDebug("Input file: " + file)
		// Ignore empty file names
		if len(file) == 0 {
			continue
		}
		tmp := citool.GetJson(file)
		data = append(data, tmp...)
	}
	return data
}
