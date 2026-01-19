package integration

import (
	"os"
	"testing"
)

// requirePerfEnabled skips the test unless performance tests are explicitly enabled.
// Usage:
//
//	requirePerfEnabled(t)
func requirePerfEnabled(t *testing.T) {
	t.Helper()

	// Respect Go's standard short mode
	if testing.Short() {
		t.Skip("skipping performance test in -short mode")
	}

	// Only run when explicitly enabled
	if os.Getenv("RUN_PERF_TESTS") != "1" {
		t.Skip("skipping performance test (set RUN_PERF_TESTS=1 to enable)")
	}
}
