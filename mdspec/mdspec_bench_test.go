package mdspec

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// BenchmarkSpecCheckWithConcurrency_Sequential benchmarks sequential execution
// with maxConcurrency=-1.
func BenchmarkSpecCheckWithConcurrency(b *testing.B) {
	testCases, expectedResults := prepareTestCasesMap(b, latestSpecFile)

	// Create a function that returns correct results
	//nolint:unparam // error is always nil in this benchmark
	correctFunc := func(markdown string) (string, error) {
		randomDelay(5, 10) // Simulate some processing delay

		return expectedResults[markdown], nil
	}

	b.Run("SequentialExecution", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			err := SpecCheckWithConcurrency("v0.13", correctFunc, -1)
			require.NoError(b, err)
		}

		b.ReportMetric(float64(len(testCases)), "testcases")
	})

	b.Run("AutoOptimizedExecution", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			err := SpecCheckWithConcurrency("v0.13", correctFunc, 0)
			require.NoError(b, err)
		}

		b.ReportMetric(float64(len(testCases)), "testcases")
	})
}

// BenchmarkSpecCheckWithConcurrency_CustomLimit benchmarks concurrent execution
// with various custom concurrency limits.
func BenchmarkSpecCheckWithConcurrency_CustomLimit(b *testing.B) {
	testCases, expectedResults := prepareTestCasesMap(b, latestSpecFile)

	// Create a function that returns correct results
	//nolint:unparam // error is always nil in this benchmark
	correctFunc := func(markdown string) (string, error) {
		randomDelay(5, 10) // Simulate some processing delay

		return expectedResults[markdown], nil
	}

	// Test various concurrency limits
	limits := []int{2, 4, 8, 16}

	for _, limit := range limits {
		b.Run(b.Name()+"_"+string(rune(limit+'0')), func(b *testing.B) {
			b.ResetTimer()

			for b.Loop() {
				err := SpecCheckWithConcurrency("v0.13", correctFunc, limit)
				require.NoError(b, err)
			}

			b.ReportMetric(float64(len(testCases)), "testcases")
		})
	}
}

// BenchmarkSpecCheck_DefaultBehavior benchmarks the default SpecCheck function
// which uses auto-optimized concurrency.
func BenchmarkSpecCheck_DefaultBehavior(b *testing.B) {
	testCases, expectedResults := prepareTestCasesMap(b, latestSpecFile)

	// Create a function that returns correct results
	correctFunc := func(markdown string) (string, error) {
		randomDelay(5, 10) // Simulate some processing delay

		return expectedResults[markdown], nil
	}

	b.ResetTimer()

	for b.Loop() {
		err := SpecCheck("v0.13", correctFunc)
		require.NoError(b, err)
	}

	b.ReportMetric(float64(len(testCases)), "testcases")
}

// ============================================================================
//  Helper Functions for Benchmarks
// ============================================================================

func randomDelay(minMicros, maxMicros int) {
	//nolint:gosec // weak random is acceptable for benchmarking purposes
	delay := minMicros + rand.IntN(maxMicros-minMicros+1)
	time.Sleep(time.Duration(delay) * time.Microsecond)
}
