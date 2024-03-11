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

	"github.com/cespare/xxhash/v2"
	"github.com/pkg/errors"
)

const (
	// FileMode600 is the file mode for files created by this program.
	FileMode600 = os.FileMode(0o600)
	// currentHash is the hash value of the spec page last checked.
	currentHash = "584ce77e30ef9594" // last checked on 2024-03-11
	// urlSpecLatest is the URL of the spec page with the current latest spec.
	urlSpecLatest = "https://spec.commonmark.org/current/"
)

func main() {
	// Check if the official spec page has been modified.
	if !IsUpToDate(currentHash) {
		fmt.Println("Official spec page has been modified. The latest spec mey not be up to date.")

		os.Exit(1)
	}

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

// IsUpToDate returns true if the given hash matches with the hash value from the latest spec
// page https://spec.commonmark.org/current/.
//
// The hash algorithm used is xxHash3.
func IsUpToDate(currentHash string) bool {
	body, err := requestGet(urlSpecLatest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return false
	}

	latestHash := fmt.Sprintf("%x", xxhash.Sum64(body))

	if currentHash == latestHash {
		return true
	}

	fmt.Println("* Spec page URL:", urlSpecLatest)
	fmt.Println("* Current hash :", currentHash)
	fmt.Println("* Latest hash  :", latestHash)

	return false
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
