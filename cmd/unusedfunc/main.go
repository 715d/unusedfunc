// Package main implements the CLI driver for the unusedfunc linter.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"log/slog"
	"maps"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/715d/unusedfunc/internal/analysis"
	"github.com/715d/unusedfunc/pkg/unusedfunc"
)

// Config holds all command-line configuration options for the unusedfunc analyzer.
type Config struct {
	Packages      []string // the Go packages to analyze
	Verbose       bool     // enables detailed output and statistics
	JSON          bool     // enables JSON output format
	BuildTags     []string // build tags to use during package loading
	Profile       bool     // enables CPU and memory profiling
	SkipGenerated bool     // skip files with generated code markers
	Strict        bool     // report ALL unused exported functions (not just /internal)
}

const (
	exitUnusedFound = 1
	exitError       = 2
)

var (
	// Set via ldflags during build.
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

var cfg Config

func main() {
	var rootCmd = &cobra.Command{
		Use:   "unusedfunc [packages...]",
		Short: "Find unused functions in Go code",
		Long: `unusedfunc is a linter that identifies unused functions in Go code.

It reports:
- Unexported functions that are not used anywhere
- Exported functions that are unused within /internal packages
- With --strict: ALL unused exported functions (use when packages aren't imported externally)`,
		Example: `  unusedfunc ./...                    # Analyze all packages
  unusedfunc pkg1 pkg2               # Analyze specific packages
  unusedfunc -v ./internal           # Verbose output
  unusedfunc -json . > report.json   # JSON output to file
  unusedfunc --strict ./...          # Report ALL unused exports`,
		Args:               cobra.ArbitraryArgs,
		RunE:               runCommand,
		PersistentPreRunE:  setup,
		PersistentPostRunE: teardown,
		SilenceUsage:       true,
		SilenceErrors:      true,
		Version:            version,
	}

	// Set custom version template to include build info.
	rootCmd.SetVersionTemplate(fmt.Sprintf("unusedfunc version %s\n  commit: %s\n  built:  %s\n", version, gitCommit, buildTime))

	// Define flags.
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.BuildTags, "build-tags", []string{}, "Build tags to use during package loading")
	rootCmd.PersistentFlags().BoolVar(&cfg.Profile, "profile", false, "Enable CPU and memory profiling (writes cpu.prof and mem.prof to current directory)")
	rootCmd.PersistentFlags().BoolVar(&cfg.SkipGenerated, "skip-generated", true, "Skip files with generated code markers (e.g., '// Code generated')")
	rootCmd.PersistentFlags().BoolVar(&cfg.Strict, "strict", false, "Report ALL unused exported functions (not just those in /internal)")

	if err := rootCmd.Execute(); err != nil {
		_ = teardown(nil, nil)
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		var cErr codedError
		if errors.As(err, &cErr) {
			os.Exit(cErr.code)
		}
		os.Exit(exitError)
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		cfg.Packages = args
	} else {
		cfg.Packages = []string{"./..."}
	}

	slog.Info("starting unused function analysis", "packages", cfg.Packages)

	result, err := runAnalysis(cmd.Context(), &cfg)
	if err != nil {
		return errWithCode(fmt.Errorf("analyze: %w", err), exitError)
	}

	if err := writeResults(result, &cfg); err != nil {
		return errWithCode(fmt.Errorf("format results: %w", err), exitError)
	}

	if len(result.UnusedFunctions) > 0 {
		return errWithCode(nil, exitUnusedFound)
	}
	return nil
}

// Result represents the analysis output for a single package including
// all unused functions and execution statistics.
type Result struct {
	UnusedFunctions []unusedfunc.UnusedFunction `json:"unused_functions"`
	Stats           struct {
		TotalFunctions      int           `json:"total_functions"`
		UnusedFunctions     int           `json:"unused_functions"`
		SuppressedFunctions int           `json:"suppressed_functions"`
		AnalysisDuration    time.Duration `json:"analysis_duration"`
	} `json:"stats"`
}

func runAnalysis(ctx context.Context, cfg *Config) (*Result, error) {
	start := time.Now()

	slog.Info("loading packages", "packages", cfg.Packages)
	if len(cfg.BuildTags) > 0 {
		slog.Info("using build tags", "tags", cfg.BuildTags)
	}

	pkgs, err := unusedfunc.LoadPackages(ctx, unusedfunc.LoaderOptions{
		Packages:  cfg.Packages,
		BuildTags: cfg.BuildTags,
	})
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}
	slog.Info("loaded packages", "num", len(pkgs))

	slog.Info("running analysis")
	analyzer := unusedfunc.NewAnalyzer(unusedfunc.AnalyzerOptions{
		SkipGenerated: cfg.SkipGenerated,
		Strict:        cfg.Strict,
	})
	result, err := analyzer.Analyze(pkgs)
	if err != nil {
		return nil, fmt.Errorf("analyze packages: %w", err)
	}
	duration := time.Since(start)
	slog.Info("analysis completed", "dur", duration)

	return convertToResult(result, duration), nil
}

func convertToResult(funcs map[types.Object]*analysis.FuncInfo, dur time.Duration) *Result {
	var r Result
	r.Stats.AnalysisDuration = dur

	sortedFuncs := slices.SortedFunc(maps.Values(funcs), func(a, b *analysis.FuncInfo) int {
		// Sort by package path, then by name.
		if a.Package != nil && b.Package != nil {
			if cmp := strings.Compare(a.Package.PkgPath, b.Package.PkgPath); cmp != 0 {
				return cmp
			}
		}
		return strings.Compare(a.Name, b.Name)
	})

	for _, f := range sortedFuncs {
		r.Stats.TotalFunctions++
		if f.IsSuppressed {
			r.Stats.SuppressedFunctions++
		}

		if f.ShouldReport() {
			pos := token.NoPos
			if f.DeclarationPos.IsValid() {
				pos = f.DeclarationPos
			}

			var position token.Position
			if f.Package != nil && f.Package.Fset != nil {
				position = f.Package.Fset.Position(pos)
			} else {
				position = token.Position{
					Filename: "unknown",
					Line:     0,
					Column:   0,
				}
			}

			var reason string
			switch {
			case !f.IsExported:
				reason = "unexported and unused"
			case f.IsInInternalPackage():
				reason = "exported in internal and unused"
			case f.Package != nil && f.Package.Name == "main":
				reason = "exported in main and unused"
			case f.Strict:
				reason = "exported and unused (strict mode)"
			}

			packagePath := ""
			if f.Package != nil {
				packagePath = f.Package.PkgPath
			}

			r.UnusedFunctions = append(r.UnusedFunctions, unusedfunc.UnusedFunction{
				Name:       f.Name,
				Position:   position,
				Reason:     reason,
				Suppressed: f.IsSuppressed,
				Package:    packagePath,
			})
			r.Stats.UnusedFunctions++
		}
	}

	return &r
}

func writeResults(result *Result, cfg *Config) error {
	var output string
	var err error

	if cfg.JSON {
		output, err = formatJSONOutput(result)
	} else {
		output = formatTextOutput(result, cfg)
	}

	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

func formatJSONOutput(result *Result) (string, error) {
	functions := make([]jFunction, 0, len(result.UnusedFunctions))
	for _, function := range result.UnusedFunctions {
		functions = append(functions, jFunction{
			Name:       function.Name,
			File:       function.Position.Filename,
			Line:       function.Position.Line,
			Column:     function.Position.Column,
			Reason:     function.Reason,
			Suppressed: function.Suppressed,
			Package:    function.Package,
		})
	}

	data, err := json.MarshalIndent(jOutput{
		UnusedFunctions: functions,
		Stats:           result.Stats,
		Version:         version,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling json output: %w", err)
	}
	return string(data), nil
}

func formatTextOutput(result *Result, cfg *Config) string {
	var output strings.Builder

	if cfg.Verbose {
		slog.Info("",
			"total_functions", result.Stats.TotalFunctions,
			"unused_functions", result.Stats.UnusedFunctions,
			"suppressed_functions", result.Stats.SuppressedFunctions,
			"analysis_duration", result.Stats.AnalysisDuration.String())
	}

	if len(result.UnusedFunctions) == 0 {
		slog.Info("no unused functions found")
		return output.String()
	}

	// Group functions by package for better organization.
	packageFunctions := make(map[string][]unusedfunc.UnusedFunction)
	for _, f := range result.UnusedFunctions {
		packageFunctions[f.Package] = append(packageFunctions[f.Package], f)
	}

	for pkg, functions := range packageFunctions {
		if len(packageFunctions) > 1 && cfg.Verbose {
			output.WriteString(fmt.Sprintf("\n%s:\n", pkg))
		}

		for _, f := range functions {
			// Format: filename:line:column functionName (reason)
			if !cfg.Verbose {
				// Compact format for non-verbose mode.
				output.WriteString(fmt.Sprintf("%s:%d:%d %s\n",
					f.Position.Filename, f.Position.Line, f.Position.Column, f.Name))
			} else {
				output.WriteString(fmt.Sprintf("  %s:%d:%d %s (%s)\n",
					f.Position.Filename, f.Position.Line, f.Position.Column, f.Name, f.Reason))
			}
		}
	}

	return output.String()
}

type jOutput struct {
	UnusedFunctions []jFunction `json:"unused_functions"`
	Stats           any         `json:"stats"`
	Version         string      `json:"version"`
	Timestamp       string      `json:"timestamp"`
}

type jFunction struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Reason     string `json:"reason"`
	Suppressed bool   `json:"suppressed"`
	Package    string `json:"package"`
}

var cpuProfile *os.File

func setup(_ *cobra.Command, _ []string) error {
	// Disable logger unless verbose flag is set.
	slog.SetDefault(slog.New(slog.DiscardHandler))
	if cfg.Verbose {
		opts := &slog.HandlerOptions{Level: slog.LevelDebug}
		var handler slog.Handler = slog.NewTextHandler(os.Stderr, opts)
		if cfg.JSON {
			handler = slog.NewJSONHandler(os.Stderr, opts)
		}
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}

	if !cfg.Profile {
		return nil
	}

	// Start CPU profiling.
	var err error
	cpuProfile, err = os.Create("cpu.prof")
	if err != nil {
		return fmt.Errorf("creating cpu.prof: %w", err)
	}
	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		_ = cpuProfile.Close()
		return fmt.Errorf("starting CPU profile: %w", err)
	}
	slog.Info("cpu profiling started", "file", "cpu.prof")
	return nil
}

func teardown(_ *cobra.Command, _ []string) error {
	if !cfg.Profile || cpuProfile == nil {
		return nil
	}

	// Stop CPU profiling and close file.
	pprof.StopCPUProfile()
	defer cpuProfile.Close()
	slog.Info("cpu profiling stopped", "file", "cpu.prof")

	// Write memory profile.
	memFile, err := os.Create("mem.prof")
	if err != nil {
		return fmt.Errorf("creating mem.prof: %w", err)
	}
	defer memFile.Close()
	runtime.GC() // Get up-to-date statistics
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		return fmt.Errorf("writing memory profile: %w", err)
	}
	slog.Info("memory profiling completed", "file", "mem.prof")
	return nil
}

func errWithCode(err error, code int) error {
	return &codedError{err: err, code: code}
}

type codedError struct {
	err  error
	code int
}

func (e codedError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}
