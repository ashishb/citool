package main

import (
	"ci-analysis-tool/src/citool"
	"flag"
	"fmt"
	"strings"
)

var mode = flag.String("mode",
	"analyze",
	"Mode - \"download\" or \"analyze\" or \"version\"")

var jobname = flag.String("jobname",
	"",
	"Only consider job results for this jobname. Analyze mode only.")

var jobStatus = flag.String("jobstatus",
	"",
	"Only consider job results with this completion status. Analyze mode only.")

var inputFiles = flag.String("input-files",
	"",
	"Comma-separated list of files containing downloaded job results from CircleCI. Analyze mode only.")

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

var printJobSuccessRate = flag.Bool("print-success-rate",
	true,
	"Print per-job aggregated success rate. Analyze mode only.")

var printJobDuration = flag.Bool("print-duration",
	true,
	"Print per-job average duration. Analyze mode only.")

var printJobDurationTimeSeries = flag.Bool("print-duration-graph",
	true,
	"Print per-job duration time series graph (yes, a graph). Analyze mode only.")

var debugMode = flag.Bool("debug",
	false,
	"Set this to true to enable debug logging")

const versionString = "0.1.0"

func main() {
	flag.Parse()

	citool.SetDebugMode(*debugMode)
	if *mode == "analyze" {
		analyze()
	} else if *mode == "download" {
		download()
	} else if *mode == "version" {
		fmt.Printf("%s\n", versionString)
	} else {
		panic(fmt.Sprintf("Unexpected mode \"%s\"", *mode))
	}
}

func analyze() {
	files := strings.Split(*inputFiles, ",")
	// Treat non-positional args as input files as well
	files = append(files, flag.Args()...)
	jobResults := getCircleCiBuildResults(&files)
	filterParams := citool.FilterParams{
		Username:       username,
		RepositoryName: repositoryName,
		BranchName:     branchName,
		JobName:        jobname,
		JobStatus:      jobStatus}
	filterParams.FilterData(&jobResults)
	analyzeParams := citool.AnalyzeParams{
		PrintJobSuccessRate:         *printJobSuccessRate,
		PrintJobDurationInAggregate: *printJobDuration,
		PrintJobDurationTimeSeries:  *printJobDurationTimeSeries}
	citool.PrintJobStats(jobResults, analyzeParams)
}

func download() {
	var jobStatusType *citool.JobStatusFilterTypes = nil
	if !citool.IsEmpty(jobStatus) {
		tmp := citool.JobStatusFilterTypes(citool.GetJobStatusFilterOrFail(*jobStatus))
		jobStatusType = &tmp
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
		JobStatus:       jobStatusType}
	citool.DownloadCircleCIBuildResults(downloadParams)
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
	// To add an empty line after debug logging.
	citool.LogDebug("")
	return data
}
