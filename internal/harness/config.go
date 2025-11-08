// Package harness provides test harness infrastructure for validating the analyzer against real-world codebases.
package harness

// PlatformConfig represents a platform-specific test configuration.
type PlatformConfig struct {
	GOOS      string   `yaml:"goos"`
	GOARCH    string   `yaml:"goarch"`
	BuildTags []string `yaml:"build_tags"`
	Skip      bool     `yaml:"skip,omitempty"`
	Reason    string   `yaml:"reason,omitempty"`
}

// TestMatrix represents multiple platform configurations for testing.
type TestMatrix struct {
	Platforms []PlatformConfig `yaml:"platforms"`
}

// EnvironmentConfig captures environment settings for tests.
type EnvironmentConfig struct {
	CGOEnabled bool
	GoPath     string
	GoProxy    string
	GoSumDB    string
}
