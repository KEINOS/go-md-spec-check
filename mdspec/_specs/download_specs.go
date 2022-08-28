/*
This package downloads the test cases from the official spec repository.

Edit the "spec_list.json" file to add the specs you want to download.
*/
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

// FileMode600 is the file mode for files created by this program.
const FileMode600 = os.FileMode(0o600)

func main() {
	// Get the spec information from the embedded file system.
	listJSON, err := os.ReadFile("spec_list.json")
	ExitOnError(err)

	// Temporary struct to unmarshal the JSON.
	specList := []struct {
		Version       string `json:"version"`
		URL           string `json:"url"`
		DateEnactment string `json:"date"`
	}{}

	ExitOnError(json.Unmarshal(listJSON, &specList))

	// Download the files and print its status.
	for _, spec := range specList {
		//nolint:forbidigo // not an output for debugging
		fmt.Printf("Downloading %s ... ", spec.URL)

		nameFile := fmt.Sprintf("spec_%s.json", spec.Version)

		ExitOnError(DownloadFile(spec.URL, nameFile))

		//nolint:forbidigo // not an output for debugging
		fmt.Println("ok")
	}
}

// The requestGet is the actual function to GET request a file from the urlTarget.
func requestGet(urlTarget string) ([]byte, error) {
	urlParsed, err := url.Parse(urlTarget)
	if err != nil {
		return nil, errors.Wrap(err, "invalid url")
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		urlParsed.String(),
		&bytes.Buffer{},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download file")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to download file: %s", resp.Status)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return result, nil
}

// DownloadFile downloads a file from the urlTarget and saves it to pathOut.
func DownloadFile(urlTarget string, pathOut string) error {
	body, err := requestGet(urlTarget)
	if err != nil {
		return errors.Wrap(err, "failed to download file")
	}

	if err := os.WriteFile(pathOut, body, FileMode600); err != nil {
		return errors.Wrap(err, "failed to create file")
	}

	return nil
}

// ExitOnError exits the program if the error is not nil.
func ExitOnError(err error) {
	if err != nil {
		//nolint:forbidigo // not an output for debugging
		fmt.Println("error")

		log.Fatal(err)
	}
}
