package citool

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// No one then 100 results can be downloaded in a single request.
const maxDownloadCircleCiLimit = 100
const maxFetchRetryCount = 5

type DownloadParams struct {
	CircleToken     *string
	VcsType         *string
	Username        *string
	RepositoryName  *string
	BranchName      *string
	Start           int
	Limit           int
	DownloadDirPath string
	TestStatus      *TestStatusTypes
}

func DownloadCircleCIBuildResults(params DownloadParams) {
	validate(params)

	start := params.Start
	end := start + params.Limit - 1

	for start <= end {
		numToDownload := end - start + 1
		if numToDownload > maxDownloadCircleCiLimit {
			numToDownload = maxDownloadCircleCiLimit
		}
		tmpDownloadParams := params
		tmpDownloadParams.Start = start
		tmpDownloadParams.Limit = numToDownload
		LogDebug(fmt.Sprintf("Downloading from %d to %d (both inclusive)", start, start+numToDownload-1))
		downloadCircleCIBuildResults(tmpDownloadParams)
		start += maxDownloadCircleCiLimit
	}
	LogDebug("Downloading finished")
}

func validate(params DownloadParams) {
	// Validate
	if IsEmpty(params.CircleToken) {
		panic("Circle CI token is empty")
	}
	if IsEmpty(params.VcsType) {
		panic("VCS name cannot be empty")
	}
	userNameProvided := !IsEmpty(params.Username)
	repositoryNameProvided := !IsEmpty(params.RepositoryName)
	if userNameProvided != repositoryNameProvided {
		panic(
			fmt.Sprintf("Only one of the username(\"%s\") or respository name(\"%s\") is provided",
				*params.Username,
				*params.RepositoryName))
	}
	branchNameProvided := !IsEmpty(params.BranchName)
	if branchNameProvided {
		if !userNameProvided || !repositoryNameProvided {
			panic(fmt.Sprintf("branchname(\"%s\") cannot be provided without username or respositry name", *params.BranchName))
		}
	}
	if params.Start < 0 {
		panic(fmt.Sprintf("start offset cannot be negative, it is %d", params.Start))
	}
	if params.Limit <= 0 {
		panic(fmt.Sprintf("limit must be > 0, it is %d", params.Limit))
	}
}

func IsEmpty(value *string) bool {
	return value == nil || len(*value) == 0
}

// Works - "https://circleci.com/api/v1.1/project/github/celo-org/celo-monpo?circle-token=${TOKEN}&limit=1&offset=5&filter=running&shallow=true"
// Fails - "https://circleci.com/api/v1.1/project/github/celo-org/celo-monorepo/tree/master?circle-token=25b7d2f03ad0a5c9cf2a2f4740211aaf3c4d59af&filter=running&limit=1&offset=5
func downloadCircleCIBuildResults(params DownloadParams) {
	var downloadUrl *url.URL
	if IsEmpty(params.Username) {
		downloadUrl = constructDownloadUrlForAllProjects(params)
	} else {
		downloadUrl = constructDownloadUrlForASpecificProject(params)
	}
	LogDebug(fmt.Sprintf("Downloading from %s", downloadUrl.String()))
	data, err := getBody(*downloadUrl)
	if err != nil {
		panic(fmt.Sprintf("Failed to download from %s, error: %s", downloadUrl, err))
	}
	outputFilename := getOutputFilename(params.DownloadDirPath, params.Start, params.Limit)
	err2 := writeToFile(outputFilename, data)
	if err2 == nil {
		LogDebug(fmt.Sprintf("Write %s", outputFilename))
	} else {
		LogDebug(fmt.Sprintf("Error while writing %s is %s", outputFilename, err2))
	}
}

// https://circleci.com/docs/api/#recent-builds-across-all-projects
func constructDownloadUrlForAllProjects(params DownloadParams) *url.URL {
	baseUrl := "https://circleci.com/api/v1.1/recent-builds"
	v := url.Values{}
	v.Set("circle-token", *params.CircleToken)
	v.Set("offset", strconv.Itoa(params.Start))
	v.Set("limit", strconv.Itoa(params.Limit))
	v.Set("shallow", "true")
	queryString := v.Encode()
	downloadUrlString := fmt.Sprintf("%s?%s", baseUrl, queryString)
	downloadUrl, err := url.Parse(downloadUrlString)
	if err != nil {
		panic("Failed parse url " + downloadUrlString)
	}
	return downloadUrl
}

// https://circleci.com/docs/api/#recent-builds-for-a-single-project
func constructDownloadUrlForASpecificProject(params DownloadParams) *url.URL {
	baseUrl := fmt.Sprintf(
		"https://circleci.com/api/v1.1/project/%s/%s/%s",
		url.PathEscape(*params.VcsType),
		url.PathEscape(*params.Username),
		url.PathEscape(*params.RepositoryName))
	if len(*params.BranchName) > 0 {
		baseUrl = fmt.Sprintf("%s/tree/%s", baseUrl, url.PathEscape(*params.BranchName))
	}
	v := url.Values{}
	v.Set("circle-token", *params.CircleToken)
	v.Set("offset", strconv.Itoa(params.Start))
	v.Set("limit", strconv.Itoa(params.Limit))
	v.Set("shallow", "true")
	if params.TestStatus != nil {
		v.Set("filter", string(*params.TestStatus))
	}
	queryString := v.Encode()
	downloadUrlString := fmt.Sprintf("%s?%s", baseUrl, queryString)
	downloadUrl, err := url.Parse(downloadUrlString)
	if err != nil {
		panic("Failed parse url " + downloadUrlString)
	}
	return downloadUrl
}

func writeToFile(filename string, contents []byte) error {
	// Create up to one parent if required
	maybeCreateDirectory(filepath.Dir(filename))
	return ioutil.WriteFile(filename, contents, 0644)
}

func maybeCreateDirectory(dirpath string) bool {
	if len(dirpath) == 0 || dirpath == ".." && dirpath == string(filepath.Separator) {
		return false
	}
	return os.Mkdir(dirpath, os.ModePerm) == nil
}

// TODO: make this more customizable
func getOutputFilename(downloadDirPath string, start int, limit int) string {
	return fmt.Sprintf(filepath.Join(downloadDirPath, "from-%d-to-%d.json"), start, start+limit-1)
}

func getBody(url url.URL) ([]byte, error) {
	urlString := url.String()
	var err error
	retryCount := 0
	for retryCount < maxFetchRetryCount {
		retryCount++
		time.Sleep(time.Duration((retryCount - 1) * 1000 * 1000 * 1000))
		client := &http.Client{}
		request, err1 := http.NewRequest("GET", url.String(), nil)
		if err1 != nil {
			panic(fmt.Sprintf("Failed to create get request \"%s\"\n", urlString))
		}
		// Or else the response is in some weird format.
		request.Header.Set("Accept", "application/json")
		response, err2 := client.Do(request)
		if err2 != nil {
			fmt.Printf("Failed to fetch on %d try: url \"%s\"\n", retryCount, urlString)
			err = err2
			continue
		}
		if response.StatusCode != 200 {
			panic(fmt.Sprintf("Failed to fetch %s, error: %s\n", urlString, response.Status))
		}
		bodyBytes, err3 := ioutil.ReadAll(response.Body)
		if err3 != nil {
			fmt.Printf("Failed to fetch on %d try: %s\n", retryCount, urlString)
			err = err3
			continue
		}
		return bodyBytes, nil
	}
	return nil, err
}
