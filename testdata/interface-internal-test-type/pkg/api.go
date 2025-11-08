package pkg

import "github.com/715d/unusedfunc/testdata/interface-internal-test-type/internal/hscan"

// ProcessData is an exported function that uses the internal scanner
// This makes the internal code reachable from non-internal packages
func ProcessData(data interface{}) error {
	// This calls into the internal package
	// In production, this would never be called with test types
	return hscan.ScanValue(data)
}

// Scanner is a type alias that re-exports the internal interface
// This mimics what go-redis does with type Scanner = hscan.Scanner
type Scanner = hscan.Scanner
