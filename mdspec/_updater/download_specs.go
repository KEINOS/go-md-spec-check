/*
This package downloads the test cases from the official spec repository.

It will download if the spec page ("https://spec.commonmark.org/") has not been
modified since the last check (the hash value is stored in the source code).
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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/zeebo/xxh3"
	"golang.org/x/mod/semver"
)

const (
	// FileMode600 is the file mode for files created by this program.
	FileMode600 = os.FileMode(0o600)
	// currentHash is the hash value of the spec page last checked.
	currentHash = "cbf6a478e79c8f79" // last checked on 2026-02-12
	urlSpecList = "https://spec.commonmark.org/"
	nameDirOut  = "_specs"
	// minVerSpec is the minimum supported version. Older versions than this are
	// not supported due to lack of official spec.json files.
	minVerSpec = "0.13"
)

type SpecInfo struct {
	Version       string `json:"version"`
	URL           string `json:"url"`
	DateEnactment string `json:"date"`
}

// ----------------------------------------------------------------------------
//  Core functions
// ----------------------------------------------------------------------------

func main() {
	body, err := requestGet(urlSpecList)
	ExitOnError(err)

	// Check if the official spec page has been modified.
	if !IsUpToDate(currentHash, body) {
		fmt.Println("[!] DOWNLOAD CANCELED:")
		fmt.Println("* The official spec page has been modified. The latest spec may not be up-to-date.")
		fmt.Println("* Please verify the changes and update the 'currentHash' value in the source code and re-run this program.")

		os.Exit(1)
	}

	fmt.Println("Spec page is as expected. Downloading spec files...")

	specList, err := extractSpecFileURLfromHTML(body)
	ExitOnError(err)

	for index, specInfo := range specList {
		fmt.Printf("- % 3d: %s, %s, %s\n", index+1, specInfo.URL, specInfo.DateEnactment, specInfo.Version)
	}

	// Download the files and print its status.
	for _, spec := range specList {
		fmt.Printf("Downloading %s ... ", spec.URL)

		nameFileOut := fmt.Sprintf("spec_%s.json", spec.Version)
		pathFileOut := filepath.Join("..", nameDirOut, nameFileOut)

		ExitOnError(DownloadFile(spec.URL, pathFileOut))

		fmt.Println("ok")
	}

	// Export the spec list to a JSON file.
	dataSpecList, err := json.MarshalIndent(specList, "", "  ")
	ExitOnError(err)

	pathSpecListOut := filepath.Join("..", nameDirOut, "spec_list.json")
	ExitOnError(os.WriteFile(pathSpecListOut, dataSpecList, FileMode600))
}

// IsUpToDate returns true if the given expectHash matches the hash of the given body.
//
// The hash algorithm used is xxHash3.
func IsUpToDate(expectHash string, body []byte) bool {
	// Calculate the hash of the latest spec page.
	latestHash := fmt.Sprintf("%x", xxh3.Hash(body))

	fmt.Println("-----------------------------------------------------------------------------------")
	fmt.Println("* Spec page URL:", urlSpecList)
	fmt.Println("* Expected hash:", expectHash)
	fmt.Println("* Actual hash  :", latestHash)
	fmt.Println("-----------------------------------------------------------------------------------")

	return expectHash == latestHash
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

// ----------------------------------------------------------------------------
//  Private/helper functions
// ----------------------------------------------------------------------------

func extractSpecFileURLfromHTML(inputHTML []byte) ([]SpecInfo, error) {
	const expDate = `\((\d{4}-\d{2}-\d{2})\)` // RFC3339 date without time

	datePattern := regexp.MustCompile(expDate)

	baseURL, err := url.Parse(urlSpecList)
	if err != nil {
		return nil, errors.Wrap(err, "invalid base url")
	}

	res := bytes.NewReader(inputHTML)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse html")
	}

	var specInfos []SpecInfo

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		version := strings.TrimSpace(s.Find("a").First().Text())
		if version == "" || !semver.IsValid("v"+version) {
			return
		}

		if semver.Compare("v"+version, "v"+minVerSpec) < 0 {
			return
		}

		dateMatch := datePattern.FindStringSubmatch(s.Text())
		if len(dateMatch) < 2 {
			return
		}

		var specHref string
		s.Find("a").EachWithBreak(func(j int, a *goquery.Selection) bool {
			href, ok := a.Attr("href")
			if !ok {
				return true
			}
			if strings.HasSuffix(href, "/spec.json") {
				specHref = href
				return false
			}
			return true
		})

		if specHref == "" {
			return
		}

		resolvedURL := baseURL.ResolveReference(&url.URL{Path: specHref}).String()

		specInfos = append(specInfos, SpecInfo{
			Version:       "v" + version,
			DateEnactment: dateMatch[1],
			URL:           resolvedURL,
		})
	})

	return specInfos, nil

	// var urls []string
	//
	// doc.Find("a").Each(func(i int, s *goquery.Selection) {
	// 	href, ok := s.Attr("href")
	// 	if !ok {
	// 		return
	// 	}
	// 	if strings.HasSuffix(href, ".json") {
	// 		urls = append(urls, href)
	// 	}
	// })
	//
	// return urls, nil
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
