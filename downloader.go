package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// No one then 100 results can be downloaded in a single request.
const maxDownloadCircleCiLimit = 100
const maxFetchRetryCount = 5

func DownloadCircleCIBuildResults(circleToken string, start int, limit int) {
	validate(circleToken, start, limit)

	end := start + limit - 1

	for start <= end {
		numToDownload := end - start + 1
		if numToDownload > maxDownloadCircleCiLimit {
			numToDownload = maxDownloadCircleCiLimit
		}
		LogDebug(fmt.Sprintf("Downloading from %d to %d (both inclusive)", start, start+numToDownload-1))
		downloadCircleCIBuildResults(circleToken, start, numToDownload)
		start += maxDownloadCircleCiLimit
	}
	LogDebug("Downloading finished")
}

func validate(circleToken string, start int, limit int) {
	// Validate
	if len(circleToken) == 0 {
		panic("Circle CI token is empty")
	}
	if start < 0 {
		panic(fmt.Sprintf("start offset cannot be negative, it is %d", start))
	}
	if limit <= 0 {
		panic(fmt.Sprintf("limit must be > 0, it is %d", limit))
	}
}

func downloadCircleCIBuildResults(circleToken string, start int, limit int) {
	url := constructUrl(circleToken, start, limit)
	data, err := getBody(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to download from %s, error: %s", url, err))
	}
	outputFilename := getOutputFilename(start, limit)
	err2 := writeToFile(outputFilename, data)
	if err2 == nil {
		LogDebug(fmt.Sprintf("Write %s", outputFilename))
	} else {
		LogDebug(fmt.Sprintf("Error while writing %s is %s", outputFilename, err2))
	}
}

func constructUrl(circleToken string, start int, limit int) string {
	return fmt.Sprintf("https://circleci.com/api/v1.1/recent-builds?"+
		"offset=%d&"+
		"limit=%d&"+
		"shallow=true&"+
		"filter=completed&"+
		"circle-token=%s",
		start,
		limit,
		circleToken)
}

func writeToFile(filename string, contents []byte) error {
	return ioutil.WriteFile(filename, contents, 0644)
}

// TODO: make this more customizable
func getOutputFilename(start int, limit int) string {
	return fmt.Sprintf("./data/from-%d-to-%d.json", start, start+limit-1)
}

func getBody(url string) ([]byte, error) {
	var err error
	retryCount := 0
	for retryCount < maxFetchRetryCount {
		retryCount++
		time.Sleep(time.Duration((retryCount - 1) * 1000 * 1000 * 1000))
		client := &http.Client{}
		request, err1 := http.NewRequest("GET", url, nil)
		if err1 != nil {
			panic(fmt.Sprintf("Failed to create get request \"%s\"\n", url))
		}
		// Or else the response is in some weird format.
		request.Header.Set("Accept", "application/json")
		response, err2 := client.Do(request)
		if err2 != nil {
			fmt.Printf("Failed to fetch on %d try: %s\n", retryCount, url)
			err = err2
			continue
		}
		bodyBytes, err3 := ioutil.ReadAll(response.Body)
		if err3 != nil {
			fmt.Printf("Failed to fetch on %d try: %s\n", retryCount, url)
			err = err3
			continue
		}
		return bodyBytes, nil
	}
	return nil, err
}
