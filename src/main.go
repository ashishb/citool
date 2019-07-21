package main

import (
	"flag"
	"fmt"
	"github.com/ashishb/ci-analysis-tool/src/citool"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var mode = flag.String("mode",
	"",
	"Mode - \"download\" or \"analyze\"")

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
	defaultStartOffset,
	fmt.Sprintf("Circle CI build results download start offset (Default: %d)", defaultStartOffset))

var downloadLimit = flag.Int("limit",
	defaultDownloadLimit,
	fmt.Sprintf("Circle CI build results download limit (Default: %d)", defaultDownloadLimit))

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
	defaultDownloadDir,
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

var printJobSuccessTimeSeries = flag.Bool("print-success-graph",
	false,
	"Print per-job success graph (yes, a graph). Analyze mode only.")

var debugMode = flag.Bool("debug",
	false,
	"Set this to true to enable debug logging")

var version = flag.Bool("version",
	false,
	"Prints version of this tool")

const versionString = "0.1.0"
const defaultDownloadDir = "./circleci_data"
const defaultStartOffset = 0
const defaultDownloadLimit = 100

func main() {
	flag.Parse()

	citool.SetDebugMode(*debugMode)
	if *version {
		fmt.Printf("%s\n", versionString)
		os.Exit(0)
	} else if *mode == "analyze" {
		analyze()
	} else if *mode == "download" {
		download()
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func analyze() {
	files := getInputFiles()
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
		PrintJobDurationTimeSeries:  *printJobDurationTimeSeries,
		PrintJobSuccessTimeSeries:   *printJobSuccessTimeSeries}
	citool.PrintJobStats(jobResults, analyzeParams)
}

func download() {
	var jobStatusType *citool.JobStatusFilterTypes
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
	citool.DownloadCircleCIJobResults(downloadParams)
}

func getInputFiles() []string {
	files := make([]string, 0)
	if len(*inputFiles) > 0 {
		files = append(files, strings.Split(*inputFiles, ",")...)
	}
	if len(flag.Args()) > 0 {
		// Treat non-positional args as input files as well
		files = append(files, flag.Args()...)
	}
	// Get default files
	if len(files) == 0 && dirExists(defaultDownloadDir) {
		fileInfos, err := ioutil.ReadDir(defaultDownloadDir)
		if err != nil {
			panic(err)
		}
		for _, fileInfo := range fileInfos {
			if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".json") {
				filePath := filepath.Join(defaultDownloadDir, fileInfo.Name())
				citool.LogDebug(fmt.Sprintf("Found default file: %s", filePath))
				files = append(files, filePath)
			}
		}
	}
	citool.LogDebug(fmt.Sprintf("Files are %d\n", len(files)))

	if len(files) == 0 {
		fmt.Printf("No input files provided\n")
		os.Exit(1)
	}
	return files
}

func dirExists(dirname string) bool {
	fileInfo, err := os.Stat(dirname)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func getCircleCiBuildResults(files *[]string) []citool.CircleCiJobResult {
	data := make([]citool.CircleCiJobResult, 0)
	for _, file := range *files {
		citool.LogDebug(fmt.Sprintf("Input file: \"%s\"", file))
		// Ignore empty file names
		if len(file) == 0 {
			continue
		}
		tmp := citool.GetCircleCIJobResults(file)
		data = append(data, tmp...)
	}
	// To add an empty line after debug logging.
	citool.LogDebug("")
	return data
}
