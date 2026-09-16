# unusedfunc Quick Reference

## 🎯 What It Does
Detects unused functions and methods in Go code using reachability analysis:
- **Unexported functions**: Reports if unused anywhere
- **Exported functions in `/internal` or `main`**: Reports if unused
- **Exported functions elsewhere**: Never reports in normal mode (public API)
- **Strict mode (`--strict`)**: Reports ALL unused exports (use for applications, not libraries)

## 🔬 Analysis Engine
**Modified RTA (Rapid Type Analysis)** computes reachability without retaining a call graph. It handles interface conversions/assertions, direct `runtime.SetFinalizer` calls, and generic origins. Reflection handling uses function-context heuristics with a direct `reflect.ValueOf`/`TypeOf` argument exception, not general value-flow analysis.

Read **docs/reference/rta-algorithm.md** before changing reachability rules.

## 📁 Project Structure

```
unusedfunc/
├── cmd/unusedfunc/          # CLI application
├── pkg/
│   ├── unusedfunc/          # Package loading and orchestration
│   ├── ssa/                 # SSA analyzer (integrates with RTA)
│   ├── assembly/            # Assembly scanning
│   ├── runtime/             # Runtime directive detection
│   └── suppress/            # Suppression comments
├── internal/
│   ├── rta/                 # Modified RTA implementation (core algorithm)
│   ├── analysis/            # Function metadata and reporting policy
│   └── harness/             # Test framework
├── testdata/                # Comprehensive test cases
└── docs/
    ├── architecture.md      # Design decisions, RTA modifications, patterns
    ├── workflows.md         # Build commands, validation, debugging
    ├── performance.md       # Optimization strategies, profiling
    ├── context/             # Session notes (temporal)
    └── reference/
        ├── known-limitations.md   # Template calls, reflection patterns
        └── rta-algorithm.md       # Modified RTA details
```

## 🛠️ Build Commands

```bash
# Build
make build        # → build/unusedfunc
go build -o build/unusedfunc ./cmd/unusedfunc

# Test
make test         # All tests with race detector and coverage
make lint         # golangci-lint using .golangci.yaml

# Run
./build/unusedfunc ./...                    # Analyze current module
./build/unusedfunc --strict ./...           # Report ALL unused exports (not just /internal)
./build/unusedfunc -v --json ./...          # JSON output and verbose stderr logs
```

## 📚 Key Documentation

### Start Here
- **README.md** - User guide, comparison with staticcheck, examples
- **docs/architecture.md** - Read before changing analysis boundaries or root policy
- **docs/workflows.md** - Development workflows, validation procedures

### Deep Dives
- **docs/reference/rta-algorithm.md** - Modified RTA algorithm explained
- **docs/reference/known-limitations.md** - Template calls, reflection patterns

### Quick Lookups
- **docs/performance.md** - Read before profiling or optimizing analysis
- **docs/context/** - Working notes from development sessions

## 🏗️ Architecture Overview

### Analysis Pipeline
```
1. Load Packages (packages.Load with full type info)
   ↓
2. Build SSA Program (ssa.InstantiateGenerics mode)
   ↓
3. Find Entry Points (main, init, tests, exported, reflection targets)
   ↓
4. Run Modified RTA (internal/rta/rta.go)
   ↓
5. Extract Reachable Functions (from RTA result)
   ↓
6. Mark Suppressions (//nolint:unusedfunc comments)
   ↓
7. Apply Reporting Policy (unexported + exports in internal/main; all exports in strict mode)
   ↓
8. Report Unused Functions
```

### Core Components

**Modified RTA** (`internal/rta/rta.go`)
- Fork-specific behavior is described in docs/reference/rta-algorithm.md
- Pattern-based reflection via `knownSafeFunctions`
- Fingerprints reject impossible interface matches before `types.Implements`
- Generic template auto-tracking

**SSA Analyzer** (`pkg/ssa/analyzer.go`)
- Integrates with modified RTA
- Entry point detection (main, init, tests, exports, reflection)
- Converts RTA results to `types.Object` set
- Consumes runtime and assembly metadata collected by `pkg/unusedfunc`

**Function Info** (`internal/analysis/`)
- Metadata per function (exported, internal, suppressed)
- Business logic: `ShouldReport()` determines if unused

**CLI** (`cmd/unusedfunc/`)
- Cobra-based command-line interface
- Progress reporting, JSON output, exit codes

## 🎯 Entry Points Detected

The analyzer considers these as entry points (always reachable):
- `main()` functions
- `init()` functions
- `Test*`, `Benchmark*`, `Example*` functions
- Exported functions and methods in non-main, non-internal packages, only in normal mode
- Functions with runtime directives (`//go:nosplit`, etc.)
- Functions called from assembly, and exported assembly implementations outside main
- Functions with recognized declaration-attached `//go:linkname` directives
- Functions with `//export` (CGo)
- Exact-name package-level reflection candidates (via `isPotentialReflectionTarget`)

RTA additionally discovers direct SetFinalizer callbacks while visiting reachable calls. Suppression comments filter reporting; they do not create roots.

## 🚨 Known Limitations

### 1. Template Method Calls
Methods called only from `.gotmpl`, `.tmpl`, or `.html` files may be reported because template text is not analyzed.

**Workaround**: Suppression comments
```go
//nolint:unusedfunc // used in template.gotmpl:42
func (t *Type) TemplateMethod() {}
```

### 2. reflect.MethodByName Patterns
Methods called via `reflect.Value.MethodByName("name")` may be flagged.

**Workaround**: Add a suppression documenting the dynamic use.

**Why**: Requires data flow analysis to track string constants through reflection calls.

### 3. Custom Reflection Patterns
Non-standard reflection patterns not in `knownSafeFunctions` map.

**Workaround**: Suppression comments
```go
//nolint:unusedfunc // called via reflect in customHandler
func (t *Type) CustomMethod() {}
```

See **docs/reference/known-limitations.md** for complete list and workarounds.

## 🔧 Configuration

### Command-Line Flags
```bash
--strict            # Check exports beyond internal/main packages too
--build-tags        # Select build tags (comma-separated)
--profile           # Write cpu.prof and mem.prof in the working directory
-v, --verbose       # Enable diagnostic logging on stderr
--json              # JSON output, always including statistics
```

Tests are always loaded. `--skip-generated` defaults to true but only filters declarations used to detect runtime directives; it does not exclude generated functions from analysis or reporting. See **docs/reference/known-limitations.md** before advising on generated-code findings.

**When to use `--strict`:**
- Application code where packages aren't imported externally
- Want to find ALL unused exported functions, not just those in `/internal`
- Cleaning up dead code in a self-contained project

**Warning:** Don't use on libraries - it will report all unused public API as false positives.

### Suppression Comments
Place the comment immediately before the declaration or on the same line. File-wide suppression is not supported.

```go
// Format 1: nolint style
//nolint:unusedfunc
func example() {}

// Format 2: lint:ignore style
//lint:ignore unusedfunc reason for suppression
func example() {}
```

## 🐛 Debugging

### Function Not Marked as Used?
1. Check entry points: Is it reachable from main/init/tests/exports?
2. Check interface tracking: Is type converted to interface?
3. Check generic tracking: Is template properly tracked?
4. Enable verbose: `./build/unusedfunc -v ./...`

### False Positive?
1. Template usage? → Add suppression with template file reference
2. Reflection usage? → Check if pattern is in `knownSafeFunctions`
3. Assembly call? → Verify `.s` file parsing worked
4. Test-only? → Check whether the selected patterns and build configuration include the caller

See **docs/workflows.md#debugging-reachability-issues** for detailed guide.

## 🎓 Understanding the Code

### Key Algorithms

**Modified RTA** (`internal/rta/rta.go`)
- Cross-product tabulation: address-taken × dynamic calls
- Fingerprint optimization for fast `implements()` checks
- Pattern-based reflection via `knownSafeFunctions`

**Analyzer Integration** (`pkg/ssa/analyzer.go`)
- `findEntryPoints()` - Detects all reachable roots
- `findReachableMethods()` - Calls `rta.Analyze()` and converts results
- `AnalyzeFuncs()` - Marks collected functions using reachability

**Business Rules** (`internal/analysis/func_info.go`)
- `ShouldReport()` - Decision tree for reporting
- `IsInInternalPackage()` - `/internal` detection

### Key Data Flow
```
types.Object → FuncInfo → SSA Entry Points → RTA → Reachable Set → Filter → Report
```

## 🔗 Related Tools

`unusedfunc` is a standalone command. Run it as a separate CI step alongside other linters; its roots and reporting policy are tailored to unused exports in `/internal` and `main`.

## 📝 Contributing

1. Read **docs/architecture.md** for design decisions
2. Check **docs/workflows.md** for development workflows
3. Add test cases to `testdata/` with `expected.yaml`
4. Run `make lint && make test`
5. Validate on real projects (see docs/workflows.md#real-world-validation)

## 💻 Coding Standards

### Logging
- **ALWAYS use `log/slog`** for all logging in non-test code
- Use structured logging with key-value pairs: `slog.Info("message", "key", value)`
- Log levels:
  - `slog.Debug()` - Detailed diagnostic information (only visible with `-v`)
  - `slog.Info()` - General informational messages
  - `slog.Warn()` - Warning conditions that should be noted
- **DO NOT use**:
  - ❌ `fmt.Fprintf(os.Stderr, ...)` for warnings/errors
  - ❌ `log.Printf()` or standard library `log` package
  - ❌ `fmt.Println()` for logging (only for program output to stdout)

**Example**:
```go
// ✅ CORRECT - structured logging
slog.Warn("scanning assembly files", "package", pkg.PkgPath, "error", err)
slog.Info("loaded packages", "num", len(pkgs))

// ❌ WRONG - don't use fmt for logging
fmt.Fprintf(os.Stderr, "warning: failed to load %s: %v\n", name, err)
```

### Go Style
- Follow standard Go conventions (gofmt, goimports)
- Early returns to reduce nesting
- Comments explain "why" not "what"
- Run `make lint` before committing

## 🚀 Quick Start

```bash
# Install
go install github.com/715d/unusedfunc/cmd/unusedfunc@latest

# Or build from source
git clone https://github.com/715d/unusedfunc.git
cd unusedfunc
make build

# Analyze your project
cd /path/to/your/project
unusedfunc ./...

# With verbose output
unusedfunc -v ./...

# JSON output for CI integration
unusedfunc --json ./... > unused.json
```

## 📖 Further Reading

- **Precision**: docs/reference/rta-algorithm.md (8 modifications explained)
- **Limitations**: docs/reference/known-limitations.md (workarounds included)
- **Architecture**: docs/architecture.md (complete technical details)
- **Performance**: docs/performance.md (profiling and verification)
