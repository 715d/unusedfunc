package hscan

import (
	"fmt"
	"testing"
	"time"
)

// TimeRFC3339Nano mimics the type from go-redis that triggers the bug
// This type is ONLY defined in test files, never in production
type TimeRFC3339Nano struct {
	time.Time
}

// ScanRedis is an exported method that implements the Scanner interface
// This method is reported as unused by the tool, but it's actually used via interface dispatch
func (t *TimeRFC3339Nano) ScanRedis(s string) error {
	var err error
	t.Time, err = time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return fmt.Errorf("failed to parse time: %w", err)
	}
	return nil
}

// TestScanValue - This test doesn't actually call the production code
// with the test type, similar to how some go-redis tests work
func TestScanValue(t *testing.T) {
	// Create the test type but don't use it with production code
	// This simulates a case where the type exists but isn't directly tested
	testTime := &TimeRFC3339Nano{Time: time.Now()}

	// Just verify the type is created correctly
	if testTime.Time.IsZero() {
		t.Error("Time should not be zero")
	}

	// Note: We're NOT calling ScanValue(testTime) here
	// This means there's no direct path from test to the interface dispatch
}

// Additional test type to make the issue more obvious
type AnotherTestType struct {
	Value string
}

// ScanRedis implementation for another test type
func (a *AnotherTestType) ScanRedis(s string) error {
	a.Value = s
	return nil
}

func TestAnotherType(t *testing.T) {
	// Similarly, don't actually use this with production code
	another := &AnotherTestType{}

	// Just test the method directly
	err := another.ScanRedis("test")
	if err != nil {
		t.Errorf("ScanRedis failed: %v", err)
	}
	if another.Value != "test" {
		t.Errorf("expected test, got %s", another.Value)
	}
}
