/*
Package mdspec provides a way to check if a given function complies with the
CommonMark specification using official test cases. It also provides a way to
list all available versions of the specification.
*/
package mdspec

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
)

// Embed JSON files under _spec into the binary.
//
//go:embed _specs/*.json
var specFiles embed.FS

var (
	nameDirSpecs     = "_specs"
	nameFileSpecList = "spec_list.json"
	prefixFileSpec   = "spec_"
)

const (
	// defaultConcurrency specifies the default number of concurrent goroutines
	// for test execution. A value of 0 uses runtime.GOMAXPROCS(0), whose behavior may
	// depend on the Go version and environment. See Go release notes for details.
	defaultConcurrency = 0
)

// Variables to be mocked/monkey-patched during testing.
var (
	// jsonUnmarshal is a copy of json.Unmarshal to ease testing.
	jsonUnmarshal = json.Unmarshal
)

// TestCase represents a single test case from the CommonMark specification.
type TestCase struct {
	Markdown   string `json:"markdown"`
	HTML       string `json:"html"`
	Section    string `json:"section"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	ExampleNum int    `json:"example"`
}

// ----------------------------------------------------------------------------
//  Public functions
// ----------------------------------------------------------------------------

// SpecCheck checks if `yourFunc“ complies with the specified CommonMark version
// specification using official test cases.
//
// Usage:
//
//	err := mdspec.SpecCheck("v1.14", myFunc)
func SpecCheck(specVersion string, yourFunc func(string) (string, error)) error {
	return SpecCheckWithConcurrency(specVersion, yourFunc, defaultConcurrency)
}

// SpecCheckWithConcurrency is the same as SpecCheck but allows specifying the maximum
// number of concurrent goroutines for spec test execution.
//
//   - If "maxConcurrency = -1", the tests will not run concurrently (runs sequentially).
//   - If "maxConcurrency = 0", it will automatically optimize the concurrency.
//
// If your function is lightning fast (< 5μs/call), running tests concurrently may not
// yield performance benefits due to overhead of preparing goroutines and context switching.
// In such cases, consider using "maxConcurrency = -1" to run tests sequentially.
func SpecCheckWithConcurrency(specVersion string, yourFunc func(string) (string, error), maxConcurrency int) error {
	const noConcurrency = -1

	if !isValidFormatVer(specVersion) {
		return errors.Errorf(
			"invalid spec version format: %s, it should be like 'v0.14'", specVersion)
	}

	nameFileSpec := fmt.Sprintf("%s%s.json", prefixFileSpec, specVersion)

	jsonSpec, err := loadFile(nameFileSpec)
	if err != nil {
		return errors.Wrap(err, "spec file not found: "+nameFileSpec)
	}

	var testCases []TestCase

	err = jsonUnmarshal(jsonSpec, &testCases)
	if err != nil {
		return errors.Wrap(err, "failed to parse list of supported spec versions")
	}

	if maxConcurrency == noConcurrency {
		for _, testCase := range testCases {
			err = runSingleTest(testCase, yourFunc)
			if err != nil {
				return errors.Wrap(err, "test failed")
			}
		}

		return nil
	}

	return runTestsConcurrently(testCases, yourFunc, maxConcurrency)
}

// ListVersion returns a list of all available versions of the specification.
func ListVersion() ([]string, error) {
	jsonList, err := loadFile(nameFileSpecList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read list of supported spec versions")
	}

	// Temporary struct to unmarshal the list of supported spec versions
	objList := []struct {
		Version       string `json:"version"`
		URL           string `json:"url"`
		DateEnactment string `json:"date"`
	}{}

	err = jsonUnmarshal(jsonList, &objList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse list of supported spec versions")
	}

	// Create list of supported spec versions
	result := make([]string, len(objList))

	for i, obj := range objList {
		result[i] = obj.Version
	}

	return result, nil
}

// ----------------------------------------------------------------------------
//  Private functions
// ----------------------------------------------------------------------------

// getNamesFile returns a list of all available file names in the embedded
// filesystem. Note that this function does not recurse into subdirectories.
func getNamesFile(dir string) ([]string, error) {
	if len(dir) == 0 {
		dir = "."
	}

	out := []string{}

	entries, err := specFiles.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read directory")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		out = append(out, filepath.ToSlash(filepath.Join(dir, entry.Name())))
	}

	return out, nil
}

// isValidFormatVer returns true if the given version input is a valid format.
func isValidFormatVer(verInput string) bool {
	if verInput == "latest" {
		return true
	}

	return semver.IsValid(verInput)
}

// loadFile returns the contents of the file with the given name from the embedded
// filesystem.
func loadFile(nameFile string) ([]byte, error) {
	// Load the list of supported spec versions
	pathFile := filepath.ToSlash(filepath.Join(nameDirSpecs, nameFile))

	jsonData, err := specFiles.ReadFile(pathFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	return jsonData, nil
}

// runSingleTest executes a single test case using the given function and
// returns an error if the test fails.
func runSingleTest(testCase TestCase, yourFunc func(string) (string, error)) error {
	nameTest := fmt.Sprintf("%d_%s", testCase.ExampleNum, testCase.Section)
	expect := testCase.HTML

	actual, err := yourFunc(testCase.Markdown)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(
			"error %s: the given function failed to parse markdown.\n"+
				"given markdown: %#v\nexpect HTML: %#v\nactual HTML: %#v",
			nameTest, testCase.Markdown, expect, actual,
		))
	}

	if expect != actual {
		return errors.Errorf(
			"error %s: the given function did not return the expected HTML result.\n"+
				"given markdown: %#v\nexpect HTML: %#v\nactual HTML: %#v",
			nameTest, testCase.Markdown, expect, actual,
		)
	}

	return nil
}

// runTestsConcurrently runs all test cases concurrently with the specified
// concurrency limit. If maxConcurrency is 0, it defaults to runtime.GOMAXPROCS(0),
// which in Go 1.25+ is automatically optimized for container environments.
func runTestsConcurrently(testCases []TestCase, yourFunc func(string) (string, error), maxConcurrency int) error {
	errGroup, ctx := errgroup.WithContext(context.Background())

	if maxConcurrency == 0 {
		maxConcurrency = runtime.GOMAXPROCS(0)
	}

	errGroup.SetLimit(maxConcurrency)

	for _, testCase := range testCases {
		errGroup.Go(func() error {
			return runSingleTest(testCase, yourFunc)
		})
	}

	return errors.Wrap(errGroup.Wait(), "failed to run tests concurrently")
}
