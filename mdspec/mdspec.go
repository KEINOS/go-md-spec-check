package mdspec

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
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

// Variables to be mocked/monkey-patched during testing.
var (
	// jsonUnmarshal is a copy of json.Unmarshal to ease testing.
	jsonUnmarshal = json.Unmarshal
)

// SpecCheck checks if `yourFunc“ complies with the specified CommonMark version
// specification using official test cases.
//
// Usage:
//
//	err := mdspec.SpecCheck("v1.14", myFunc)
func SpecCheck(specVersion string, yourFunc func(string) (string, error)) error {
	if !isValidFormatVer(specVersion) {
		return errors.Errorf(
			"invalid spec version format: %s, it should be like 'v0.14'", specVersion)
	}

	nameFileSpec := fmt.Sprintf("%s%s.json", prefixFileSpec, specVersion)

	jsonSpec, err := loadFile(nameFileSpec)
	if err != nil {
		return errors.Wrap(err, "spec file not found: "+nameFileSpec)
	}

	// Temporary struct to unmarshal the list of supported spec versions
	listTests := []struct {
		Markdown   string `json:"markdown"`
		HTML       string `json:"html"`
		Section    string `json:"section"`
		StartLine  int    `json:"start_line"`
		EndLine    int    `json:"end_line"`
		ExampleNum int    `json:"example"`
	}{}

	if err := jsonUnmarshal(jsonSpec, &listTests); err != nil {
		return errors.Wrap(err, "failed to parse list of supported spec versions")
	}

	for _, test := range listTests {
		nameTest := fmt.Sprintf("%d_%s", test.ExampleNum, test.Section)
		expect := test.HTML

		actual, err := yourFunc(test.Markdown)
		if err != nil {
			msgErr := fmt.Sprintf(
				"error %s: the given function failed to parse markdown.\n"+
					"given markdown: %#v\nexpect HTML: %#v\nactual HTML: %#v\n",
				nameTest, test.Markdown, expect, actual,
			)

			return errors.Wrap(err, msgErr)
		}

		if expect != actual {
			msgErr := fmt.Sprintf(
				"error %s: the given function did not return the expected HTML result.\n"+
					"given markdown: %#v\nexpect HTML: %#v\nactual HTML: %#v",
				nameTest, test.Markdown, expect, actual,
			)

			return errors.New(msgErr)
		}
	}

	return nil
}

// isValidFormatVer returns true if the given version input is a valid formtat.
func isValidFormatVer(verInput string) bool {
	if verInput == "latest" {
		return true
	}

	re := regexp.MustCompile("^v[0-9]+.[0-9]+$")

	return re.MatchString(verInput)
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

	if err := jsonUnmarshal(jsonList, &objList); err != nil {
		return nil, errors.Wrap(err, "failed to parse list of supported spec versions")
	}

	// Create list of supported spec versions
	result := make([]string, len(objList))

	for i, obj := range objList {
		result[i] = obj.Version
	}

	return result, nil
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
