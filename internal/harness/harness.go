// Package harness provides testing utilities for the unusedfunc analyzer.
package harness

import (
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/stretchr/testify/require"

	"github.com/715d/unusedfunc/internal/analysis"
	"github.com/715d/unusedfunc/pkg/unusedfunc"
)

// BuildConfiguration represents a single build configuration to test.
type BuildConfiguration struct {
	// Name is a descriptive name for this configuration.
	Name string `yaml:"name"`

	// BuildTags are the build tags to use when loading packages.
	BuildTags []string `yaml:"build_tags"`

	// EnableCGo indicates whether CGo should be enabled.
	EnableCGo bool `yaml:"enable_cgo"`

	// GOOS sets the target operating system.
	GOOS string `yaml:"goos,omitempty"`

	// GOARCH sets the target architecture.
	GOARCH string `yaml:"goarch,omitempty"`

	// ExpectedUnused lists the functions expected to be reported as unused for this configuration.
	ExpectedUnused []ExpectedFunc `yaml:"expected_unused"`

	// ExpectedErrors lists any expected error messages for this configuration.
	ExpectedErrors []string `yaml:"expected_errors"`
}

// TestCase represents a single test scenario.
type TestCase struct {
	// Dir is the directory containing the test code.
	Dir string `yaml:"-"`

	// Repository contains optional git repository configuration for external testing.
	Repository *RepoConfig `yaml:"repository,omitempty"`

	// BuildConfigurations defines multiple build configurations to test.
	BuildConfigurations []BuildConfiguration `yaml:"build_configurations"`
}

// ExpectedFunc represents a function expected to be reported as unused.
type ExpectedFunc struct {
	// FuncName is the name of the function.
	FuncName string `yaml:"func"`

	// Reason describes why the function is unused.
	Reason string `yaml:"reason"`

	// File is the optional file path (relative to test dir)
	File string `yaml:"file,omitempty"`
}

// RepoConfig represents configuration for testing external repositories.
type RepoConfig struct {
	// URL is the git repository URL.
	URL string `yaml:"url"`

	// Ref is the git reference (commit, branch, or tag) to checkout.
	Ref string `yaml:"ref"`

	// Subdir is an optional subdirectory within the repository to test.
	Subdir string `yaml:"subdir,omitempty"`
}

// TestHarness manages test execution.
type TestHarness struct {
	// analyzer is the unusedfunc analyzer
	analyzer *unusedfunc.Analyzer

	// root is the root directory for test data
	root string
}

// NewHarness creates a new test harness.
func NewHarness(root string) *TestHarness {
	return &TestHarness{
		analyzer: unusedfunc.NewAnalyzer(unusedfunc.AnalyzerOptions{}),
		root:     root,
	}
}

// Run executes a test case with all its build configurations.
func (h *TestHarness) Run(t *testing.T, tc *TestCase) *TestResult {
	t.Helper()
	require.NotEmpty(t, tc.BuildConfigurations, "test case has no build configurations")

	var results []ConfigurationResult
	var allSuccess = true

	// Run each configuration.
	for _, cfg := range tc.BuildConfigurations {
		cfgResult := h.runConfiguration(t, tc, cfg)
		results = append(results, *cfgResult)
		if !cfgResult.Success {
			allSuccess = false
		}
	}

	// Create overall result message.
	var resultMsg string
	if allSuccess {
		resultMsg = fmt.Sprintf("All %d configurations passed", len(tc.BuildConfigurations))
	} else {
		failedCount := 0
		var msgs []string
		for _, cr := range results {
			if !cr.Success {
				failedCount++
				msgs = append(msgs, fmt.Sprintf("[%s] %s:\n  %s",
					cr.Configuration.Name, cr.Message, strings.Join(cr.Details, "\n")))
			}
		}
		resultMsg = fmt.Sprintf("%d/%d configurations failed:\n%s",
			failedCount, len(tc.BuildConfigurations), strings.Join(msgs, "\n"))
	}

	return &TestResult{
		TestCase:             tc,
		ConfigurationResults: results,
		Success:              allSuccess,
		Message:              resultMsg,
	}
}

// runConfiguration executes analysis for a single build configuration
func (h *TestHarness) runConfiguration(t *testing.T, tc *TestCase, cfg BuildConfiguration) *ConfigurationResult {
	t.Helper()
	loaderConfig := &LoaderConfig{
		BuildTags: cfg.BuildTags,
		EnableCGo: cfg.EnableCGo,
		GOOS:      cfg.GOOS,
		GOARCH:    cfg.GOARCH,
	}

	var pkgs []*packages.Package
	if tc.Repository != nil {
		// Load packages from repository.
		pkgs = LoadRepositoryPackages(t, tc.Repository, loaderConfig)
	} else {
		// Load packages from local directory.
		loaderConfig.Dir = filepath.Join(h.root, tc.Dir)
		pkgs = LoadPackages(t, loaderConfig)
	}

	// Run analysis.
	result, err := unusedfunc.NewAnalyzer(unusedfunc.AnalyzerOptions{}).Analyze(pkgs)
	if err != nil {
		// Check if this error was expected.
		for _, expectedErr := range cfg.ExpectedErrors {
			if strings.Contains(err.Error(), expectedErr) {
				return &ConfigurationResult{
					Configuration: cfg,
					Success:       true,
					Message:       fmt.Sprintf("Got expected error: %v", err),
				}
			}
		}
		require.NoError(t, err)
	}
	return h.validateConfigurationResults(cfg, result)
}

// validateConfigurationResults compares actual results with expected for a specific build configuration
func (h *TestHarness) validateConfigurationResults(cfg BuildConfiguration, funcs map[types.Object]*analysis.FuncInfo) *ConfigurationResult {
	cfgResult := ConfigurationResult{
		Configuration: cfg,
		Funcs:         funcs,
	}

	// First validate the configuration has valid expected functions.
	if err := validateExpectedFunctions(cfg.ExpectedUnused); err != nil {
		cfgResult.Success = false
		cfgResult.Message = fmt.Sprintf("Invalid expected.yaml: %v", err)
		cfgResult.Details = []string{err.Error()}
		return &cfgResult
	}

	// Extract unused functions from result.
	var unusedFuncs []UnusedFunc
	for _, f := range funcs {
		if f.ShouldReport() {
			packageName := getDisplayPackageName(f.Package)
			unusedFuncs = append(unusedFuncs, UnusedFunc{
				Name:    f.Name,
				Package: packageName,
				File:    getRelativeFile(h.root, f.DeclarationPos, f),
			})
		}
	}

	// Compare with expected for this configuration.
	validateResults(&cfgResult, cfg.ExpectedUnused, unusedFuncs)
	return &cfgResult
}

// ConfigurationResult represents the result of running a single build configuration.
type ConfigurationResult struct {
	// Configuration is the build configuration that was run.
	Configuration BuildConfiguration

	// Funcs is the raw result from the analyzer.
	Funcs map[types.Object]*analysis.FuncInfo

	// Success indicates if this configuration passed.
	Success bool

	// Message provides a summary of the result for this configuration.
	Message string

	// Details provides detailed information about failures for this configuration.
	Details []string
}

// TestResult represents the result of running a test case.
type TestResult struct {
	// TestCase is the test case that was run.
	TestCase *TestCase

	// ConfigurationResults contains results for each build configuration.
	ConfigurationResults []ConfigurationResult

	// Success indicates if the test passed (all configurations passed)
	Success bool

	// Skipped indicates if the test was skipped.
	Skipped bool

	// Message provides a summary of the result.
	Message string
}

// UnusedFunc represents an unused function found by analysis.
type UnusedFunc struct {
	Name    string
	Package string
	File    string
}

// validateExpectedFunctions validates that expected functions have required fields
func validateExpectedFunctions(expected []ExpectedFunc) error {
	for i, exp := range expected {
		if strings.TrimSpace(exp.FuncName) == "" {
			return fmt.Errorf("expected function at index %d has empty or missing 'func' field", i)
		}
	}
	return nil
}

func validateResults(cfgResult *ConfigurationResult, expected []ExpectedFunc, actual []UnusedFunc) {
	expectedMap := make(map[string]ExpectedFunc)
	for _, e := range expected {
		expectedMap[e.FuncName] = e
	}

	actualMap := make(map[string]UnusedFunc)
	for _, a := range actual {
		actualMap[a.Name] = a
	}

	var details []string
	success := true

	// Check for missing expected functions.
	var missing []string
	for key, exp := range expectedMap {
		if _, found := actualMap[key]; !found {
			missing = append(missing, fmt.Sprintf("%s (%s)", exp.FuncName, exp.Reason))
			success = false
		}
	}

	// Check for unexpected functions.
	var unexpected []string
	for key, act := range actualMap {
		if _, found := expectedMap[key]; !found {
			unexpected = append(unexpected, act.Name)
			success = false
		}
	}

	// Sort for consistent output.
	sort.Strings(missing)
	sort.Strings(unexpected)

	// Build details.
	if len(missing) > 0 {
		for _, m := range missing {
			details = append(details, "Should have been marked unused: "+m)
		}
	}

	if len(unexpected) > 0 {
		for _, u := range unexpected {
			details = append(details, "Should have been marked used: "+u)
		}
	}

	for key, exp := range expectedMap {
		if act, found := actualMap[key]; found {
			if exp.File != "" && !strings.HasSuffix(act.File, exp.File) {
				details = append(details, fmt.Sprintf(
					"File mismatch for %s: expected file ending with %q, got %q",
					exp.FuncName, exp.File, act.File))
				success = false
			}
		}
	}

	var message string
	if success {
		message = fmt.Sprintf("All %d expected unused functions found", len(expected))
	} else {
		message = fmt.Sprintf("Test failed: %d missing, %d unexpected", len(missing), len(unexpected))
	}

	cfgResult.Success = success
	cfgResult.Message = message
	cfgResult.Details = details
}

// getRelativeFile extracts the relative file path from a position for a specific configuration
func getRelativeFile(root string, pos token.Pos, funcInfo *analysis.FuncInfo) string {
	if funcInfo.Package == nil || funcInfo.Package.Fset == nil {
		return ""
	}

	position := funcInfo.Package.Fset.Position(pos)
	if position.Filename == "" {
		return ""
	}

	// Get relative path from the testdata root.
	relPath, err := filepath.Rel(root, position.Filename)
	if err != nil {
		// If we can't get relative path, just return the base filename.
		return filepath.Base(position.Filename)
	}
	return relPath
}

// getDisplayPackageName returns the appropriate package name for display/comparison
// For main packages, it uses the module path if available, otherwise "main".
// For other packages, it uses the package path.
func getDisplayPackageName(pkg *packages.Package) string {
	const pkgMain = "main"
	if pkg == nil {
		return "unknown"
	}

	packageName := pkg.PkgPath
	if pkg.Name == pkgMain {
		// For main packages, only use module path if this is the root package.
		// (i.e., package path equals module path)
		if pkg.Module != nil && pkg.Module.Path != "" && pkg.PkgPath == pkg.Module.Path {
			if pkg.Module.Path != pkgMain {
				packageName = pkg.Module.Path
			} else {
				packageName = pkgMain
			}
		}
		// For main packages in subdirectories, keep the full package path.
	}
	return packageName
}
