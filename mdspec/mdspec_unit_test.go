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

const latestSpecFile = "spec_v0.13.json"

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

func TestSpecCheck_golden(t *testing.T) {
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

	// Dummy cheat function that returns the exact same HTML as the test case
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
	assert.Equal(t, len(testCases), int(executionCount.Load()),
		"all test cases should be executed exactly once")

	t.Logf("Successfully executed %d test cases concurrently", executionCount.Load())
}

// ----------------------------------------------------------------------------
//  SpecCheckWithConcurrency() tests
// ----------------------------------------------------------------------------

func TestSpecCheckWithConcurrency_sequential_execution(t *testing.T) {
	t.Parallel()

	testCases, expectedResults := prepareTestCasesMap(t, latestSpecFile)

	var (
		executionCount atomic.Int32
		maxConcurrent  atomic.Int32
		currentRunning atomic.Int32
	)

	// Function that tracks execution and should run sequentially
	trackingFunc := func(markdown string) (string, error) {
		current := currentRunning.Add(1)
		executionCount.Add(1)

		// Track max concurrent
		for {
			maxVal := maxConcurrent.Load()
			if current <= maxVal || maxConcurrent.CompareAndSwap(maxVal, current) {
				break
			}
		}

		// Small delay to ensure overlap would be detected if it happened
		time.Sleep(1 * time.Millisecond)

		currentRunning.Add(-1)

		result, ok := expectedResults[markdown]
		if !ok {
			return "", errors.New("unexpected markdown")
		}

		return result, nil
	}

	// Run with maxConcurrency=-1 (sequential)
	err := SpecCheckWithConcurrency("v0.13", trackingFunc, -1)
	require.NoError(t, err, "sequential execution should succeed")

	// Verify all tests were executed
	assert.Equal(t, len(testCases), int(executionCount.Load()),
		"all test cases should be executed")

	// Verify sequential execution (max concurrent should be 1)
	assert.Equal(t, int32(1), maxConcurrent.Load(),
		"sequential execution should have max concurrency of 1, got %d", maxConcurrent.Load())

	t.Logf("Sequential execution: %d test cases with max concurrency=%d",
		executionCount.Load(), maxConcurrent.Load())
}

func TestSpecCheckWithConcurrency_custom_concurrency(t *testing.T) {
	t.Parallel()

	testCases, expectedResults := prepareTestCasesMap(t, latestSpecFile)

	var (
		executionCount atomic.Int32
		maxConcurrent  atomic.Int32
		currentRunning atomic.Int32
	)

	trackingFunc := func(markdown string) (string, error) {
		current := currentRunning.Add(1)
		executionCount.Add(1)

		// Track max concurrent
		for {
			maxVal := maxConcurrent.Load()
			if current <= maxVal || maxConcurrent.CompareAndSwap(maxVal, current) {
				break
			}
		}

		// Delay to increase chance of concurrent execution
		time.Sleep(5 * time.Millisecond)

		currentRunning.Add(-1)

		result, ok := expectedResults[markdown]
		if !ok {
			return "", errors.New("unexpected markdown")
		}

		return result, nil
	}

	// Test with custom concurrency limit
	customLimit := 3
	err := SpecCheckWithConcurrency("v0.13", trackingFunc, customLimit)
	require.NoError(t, err, "custom concurrency execution should succeed")

	// Verify all tests were executed
	assert.Equal(t, len(testCases), int(executionCount.Load()),
		"all test cases should be executed")

	// Verify concurrency was limited to custom value
	assert.LessOrEqual(t, maxConcurrent.Load(), int32(customLimit),
		"max concurrency should not exceed custom limit of %d, got %d",
		customLimit, maxConcurrent.Load())

	assert.GreaterOrEqual(t, maxConcurrent.Load(), int32(2),
		"should have at least 2 concurrent executions with limit=%d, got %d",
		customLimit, maxConcurrent.Load())

	t.Logf("Custom concurrency (limit=%d): %d test cases with max observed=%d",
		customLimit, executionCount.Load(), maxConcurrent.Load())
}

func TestSpecCheckWithConcurrency_auto_optimization(t *testing.T) {
	t.Parallel()

	testCases, expectedResults := prepareTestCasesMap(t, latestSpecFile)

	var executionCount atomic.Int32

	trackingFunc := func(markdown string) (string, error) {
		executionCount.Add(1)

		result, ok := expectedResults[markdown]
		if !ok {
			return "", errors.New("unexpected markdown")
		}

		return result, nil
	}

	// Test with maxConcurrency=0 (auto-optimize)
	err := SpecCheckWithConcurrency("v0.13", trackingFunc, 0)
	require.NoError(t, err, "auto-optimized execution should succeed")

	// Verify all tests were executed
	assert.Equal(t, len(testCases), int(executionCount.Load()),
		"all test cases should be executed")

	t.Logf("Auto-optimized execution: %d test cases (GOMAXPROCS=%d)",
		executionCount.Load(), runtime.GOMAXPROCS(0))
}

func TestSpecCheckWithConcurrency_error_propagation(t *testing.T) {
	t.Parallel()

	errorFunc := func(_ string) (string, error) {
		return "", errors.New("intentional error")
	}

	// Test with sequential execution
	err := SpecCheckWithConcurrency("v0.13", errorFunc, -1)
	require.Error(t, err, "should propagate error in sequential mode")
	assert.Contains(t, err.Error(), "intentional error")

	// Test with concurrent execution
	err = SpecCheckWithConcurrency("v0.13", errorFunc, 2)
	require.Error(t, err, "should propagate error in concurrent mode")
	assert.Contains(t, err.Error(), "intentional error")
}

// ============================================================================
//  Helpers for tests
// ============================================================================

func convAsKey(s string) string {
	//nolint:gosec // use of md5 is intentional. not for cryptographic purposes
	h := md5.Sum([]byte(s))

	return hex.EncodeToString(h[:])
}

// prepareTestCasesMap loads test cases and creates a map for lookup.
//
//nolint:unparam // it always receives "spec_v0.13.json" for now
func prepareTestCasesMap(tb testing.TB, specFile string) ([]TestCase, map[string]string) {
	tb.Helper()

	jsonSpec, err := loadFile(specFile)
	require.NoError(tb, err, "failed to load spec file")

	var testCases []TestCase

	require.NoError(tb, jsonUnmarshal(jsonSpec, &testCases),
		"failed to unmarshal test cases",
	)

	expectedResults := make(map[string]string, len(testCases))
	for _, tc := range testCases {
		expectedResults[tc.Markdown] = tc.HTML
	}

	return testCases, expectedResults
}
