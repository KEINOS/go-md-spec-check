package mdspec

import (
	//nolint:gosec // use of md5 is intentional. not for cryptographic purposes
	"crypto/md5"
	"encoding/hex"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  getNamesFile()
// ----------------------------------------------------------------------------

func Test_getNamesFile_blank_path(t *testing.T) {
	t.Parallel()

	_, err := getNamesFile("")

	require.NoError(t, err, "empty path should be treated as '.'")
}

func Test_getNamesFile_unknown_path(t *testing.T) {
	t.Parallel()

	_, err := getNamesFile("unknown")

	require.Error(t, err, "unexisting path should return an error")
}

// ----------------------------------------------------------------------------
//  isValidFormatVer()
// ----------------------------------------------------------------------------

func Test_isValidFormatVer(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input string
		want  bool
	}{
		// Valid cases
		{"latest", true},
		{"v0.14", true},
		{"v0.31.2", true},
		{"v1.14.0", true},
		{"v1", true},
		// Invalid cases
		{"0.14", false},
		{"version 1.14", false},
		{"vvvv1.14", false},
		{"v0.14\n", false},
	} {
		if test.want {
			require.True(t, isValidFormatVer(test.input), test.input+" should be valid")
		} else {
			require.False(t, isValidFormatVer(test.input), test.input+" should be invalid")
		}
	}
}

// ----------------------------------------------------------------------------
//  ListVersion()
// ----------------------------------------------------------------------------

func TestListVersion_contains_all(t *testing.T) {
	t.Parallel()

	listExpect, err := ListVersion()
	require.NoError(t, err)

	listActual, err := getNamesFile(nameDirSpecs)
	require.NoError(t, err)

	for _, vExpect := range listExpect {
		found := false

		for _, vActual := range listActual {
			if strings.Contains(vActual, nameFileSpecList) {
				continue
			}

			if strings.Contains(vActual, vExpect) {
				found = true

				break
			}
		}

		if !found {
			t.Fatal("missing spec file: ", vExpect)
		}
	}
}

//nolint:paralleltest // do not parallelize due to dependency on other tests
func TestListVersion_fail_to_unmarshal(t *testing.T) {
	// Backup and defer restore the file name
	oldJSONUnmarshal := jsonUnmarshal

	defer func() {
		jsonUnmarshal = oldJSONUnmarshal
	}()

	// Mock/monkey patch to force an error
	jsonUnmarshal = func([]byte, any) error {
		return errors.New("forced error")
	}

	listExpect, err := ListVersion()

	require.Error(t, err, "it should fail to unmarshal")
	require.Nil(t, listExpect, "it should be nil on error")
	assert.Contains(t, err.Error(), "forced error")
}

//nolint:paralleltest // do not parallelize due to dependency on other tests
func TestListVersion_non_existing_dir(t *testing.T) {
	// Backup and defer restore the file name
	oldNameFileSpecList := nameFileSpecList

	defer func() {
		nameFileSpecList = oldNameFileSpecList
	}()

	// Mock/monkey patch the file name temporarily
	nameFileSpecList = "unknown"

	listExpect, err := ListVersion()

	require.Error(t, err, "non existing directory should return an error")
	require.Nil(t, listExpect, "it should be nil on error")
}

// ----------------------------------------------------------------------------
//  SpecCheck()
// ----------------------------------------------------------------------------

func convAsKey(s string) string {
	//nolint:gosec // use of md5 is intentional. not for cryptographic purposes
	h := md5.Sum([]byte(s))

	return hex.EncodeToString(h[:])
}

func TestSpecCheck_goledn(t *testing.T) {
	t.Parallel()

	// Preparation for the dummy function
	jsonSpec, err := loadFile("spec_v0.30.json")
	require.NoError(t, err, "failed to load spec file")

	listTests := []struct {
		Markdown   string `json:"markdown"`
		HTML       string `json:"html"`
		Section    string `json:"section"`
		StartLine  int    `json:"start_line"`
		EndLine    int    `json:"end_line"`
		ExampleNum int    `json:"example"`
	}{}

	require.NoError(t, jsonUnmarshal(jsonSpec, &listTests),
		"failed to unmarshal list of supported spec versions",
	)

	listPairs := map[string]string{}

	for _, t := range listTests {
		key := convAsKey(t.Markdown)
		listPairs[key] = t.HTML
	}

	// Dummy cheet function that returns the exact same HTML as the test case
	// has.
	myDummyParser := func(md string) (string, error) {
		key := convAsKey(md)

		result, ok := listPairs[key]
		if !ok {
			return "", errors.New("not found")
		}

		return result, nil
	}

	require.NoError(t, SpecCheck("v0.30", myDummyParser),
		"it should not return an error")
}

func TestSpecCheck_bad_html(t *testing.T) {
	t.Parallel()

	myDummyFunc := func(string) (string, error) {
		return "<p>bad HTML</p>", nil
	}

	err := SpecCheck("v0.30", myDummyFunc)

	require.Error(t, err, "bad HTML should return an error")
	assert.Contains(t, err.Error(), "the given function did not return the expected HTML result")
}

func TestSpecCheck_function_error(t *testing.T) {
	t.Parallel()

	myDummyFunc := func(string) (string, error) {
		return "", errors.New("something went wrong")
	}

	err := SpecCheck("v0.30", myDummyFunc)

	require.Error(t, err, "bad HTML should return an error")
	assert.Contains(t, err.Error(), "the given function failed to parse markdown")
	assert.Contains(t, err.Error(), "something went wrong")
}

//nolint:paralleltest // do not parallelize due to dependency on other tests
func TestSpecCheck_spec_version_error(t *testing.T) {
	// Backup and defer restore functions
	oldJSONUnmarshal := jsonUnmarshal

	defer func() {
		jsonUnmarshal = oldJSONUnmarshal
	}()

	// Mock/monkey patch functions to force an error
	jsonUnmarshal = func(_ []byte, _ any) error {
		return errors.New("forced error")
	}

	myDummyFunc := func(string) (string, error) {
		return "", nil
	}

	t.Run("invalid version format", func(t *testing.T) {
		err := SpecCheck("version Unknown", myDummyFunc)

		require.Error(t, err, "invalid version format should return an error")
		assert.Contains(t, err.Error(), "invalid spec version format")
		assert.Contains(t, err.Error(), "it should be like")
	})

	t.Run("unsupported spec version", func(t *testing.T) {
		err := SpecCheck("v0.1", myDummyFunc)

		require.Error(t, err, "unsupported spec version should not return an error")
		assert.Contains(t, err.Error(), "spec file not found")
		assert.Contains(t, err.Error(), "spec_v0.1.json")
	})

	t.Run("unsupported spec version", func(t *testing.T) {
		err := SpecCheck("v0.13", myDummyFunc)

		require.Error(t, err, "forced unmarshal error should return an error")
		assert.Contains(t, err.Error(), "failed to parse list of supported spec versions")
		assert.Contains(t, err.Error(), "forced error")
	})
}

// ----------------------------------------------------------------------------
//  Concurrency verification tests
// ----------------------------------------------------------------------------

func TestSpecCheck_runs_concurrently(t *testing.T) {
	t.Parallel()

	var (
		concurrentCount atomic.Int32
		maxConcurrent   atomic.Int32
	)

	// Create a function that tracks concurrent execution
	slowFunc := func(_ string) (string, error) {
		// Increment concurrent counter
		current := concurrentCount.Add(1)

		// Update max if this is higher
		for {
			maxVal := maxConcurrent.Load()
			if current <= maxVal || maxConcurrent.CompareAndSwap(maxVal, current) {
				break
			}
		}

		// Simulate some work to ensure goroutines overlap
		time.Sleep(10 * time.Millisecond)

		// Decrement on exit
		concurrentCount.Add(-1)

		// Return valid HTML for any input
		return "<p>test</p>\n", nil
	}

	// Run SpecCheck with a small spec version that has enough test cases
	err := SpecCheck("v0.13", slowFunc)

	// We expect an error because our function doesn't return correct HTML
	require.Error(t, err, "should fail due to incorrect HTML output")

	// Verify that at least 2 goroutines ran concurrently
	maxReached := maxConcurrent.Load()
	minExpected := int32(2)

	if runtime.GOMAXPROCS(0) > 1 {
		assert.GreaterOrEqual(t, maxReached, minExpected,
			"expected at least %d concurrent goroutines, got %d (GOMAXPROCS=%d)",
			minExpected, maxReached, runtime.GOMAXPROCS(0))
	} else {
		t.Logf("Skipping concurrency check: GOMAXPROCS=1")
	}

	t.Logf("Max concurrent goroutines observed: %d (GOMAXPROCS=%d)",
		maxReached, runtime.GOMAXPROCS(0))
}

func TestSpecCheck_concurrency_correctness(t *testing.T) {
	t.Parallel()

	// Prepare test data
	jsonSpec, err := loadFile("spec_v0.30.json")
	require.NoError(t, err, "failed to load spec file")

	var testCases []TestCase

	require.NoError(t, jsonUnmarshal(jsonSpec, &testCases),
		"failed to unmarshal test cases",
	)

	// Create a map for fast lookup
	expectedResults := make(map[string]string, len(testCases))
	for _, tc := range testCases {
		expectedResults[tc.Markdown] = tc.HTML
	}

	// Track all test cases that were executed
	var executionCount atomic.Int32

	// Function that returns correct results and tracks execution
	correctFunc := func(markdown string) (string, error) {
		executionCount.Add(1)

		result, ok := expectedResults[markdown]
		if !ok {
			return "", errors.New("unexpected markdown input")
		}

		return result, nil
	}

	// Run SpecCheck - should succeed with correct function
	err = SpecCheck("v0.30", correctFunc)
	require.NoError(t, err, "SpecCheck should succeed with correct function")

	// Verify all test cases were executed
	assert.Equal(t, int32(len(testCases)), executionCount.Load(),
		"all test cases should be executed exactly once")

	t.Logf("Successfully executed %d test cases concurrently", executionCount.Load())
}
