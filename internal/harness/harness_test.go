package harness

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAll runs all integration tests.
func TestAll(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "get current file path")

	harnessDir := filepath.Dir(filename)
	testdataDir := filepath.Join(harnessDir, "..", "..", "testdata")

	testCases := discoverTestCases(t, testdataDir)
	require.NotEmpty(t, testCases, "no test cases found")

	if testing.Verbose() {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}

	for _, tc := range testCases {
		t.Run(tc.Dir, func(t *testing.T) {
			t.Parallel()

			for _, config := range tc.BuildConfigurations {
				if len(config.BuildTags) > 0 {
					t.Logf("[%s] Build tags: %v", config.Name, config.BuildTags)
				}
				if config.EnableCGo {
					t.Logf("[%s] CGo enabled", config.Name)
				}
			}

			result := NewHarness(testdataDir).Run(t, tc)
			if result.Skipped {
				t.Skipf("Test skipped: %s", result.Message)
				return
			}

			if !result.Success {
				t.Errorf("Test failed: %s", result.Message)
			}
		})
	}
}

func discoverTestCases(t *testing.T, root string) []*TestCase {
	t.Helper()

	// Read all directories in testdata.
	entries, err := os.ReadDir(root)
	require.NoError(t, err)

	var testCases []*TestCase
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip realworld tests when running in short mode.
		if strings.HasPrefix(entry.Name(), "realworld-") && testing.Short() {
			continue
		}

		dir := filepath.Join(root, entry.Name())

		// Check if this directory has an expected.yaml.
		if _, err := os.Stat(filepath.Join(dir, "expected.yaml")); err == nil {
			testCases = append(testCases, LoadTestCase(t, dir, root))
		}
	}

	return testCases
}
