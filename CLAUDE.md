# unusedfunc Quick Reference

## 🎯 What It Does
Detects unused functions and methods in Go code with high precision:
- **Unexported functions**: Reports if unused anywhere
- **Exported functions in `/internal`**: Reports if unused (enforces internal convention)
- **Exported functions elsewhere**: Never reports (public API)

## 🔬 Analysis Engine
**Modified RTA (Rapid Type Analysis)** with 8 precision enhancements:
1. Pattern-based reflection (knows what `json.Marshal`, `fmt.Printf` actually call)
2. Precise interface conversions (only marks interface-required methods)
3. Context-aware analysis (tracks calling context)
4. Enhanced compliance (`*Interface → any` patterns like `errors.As`)
5. TypeAssert/ChangeInterface support
6. SetFinalizer detection
7. Generic template auto-tracking
8. 50+ known safe functions mapped

**Result**: 80% of projects (16/20) have zero false positives on 1.2M+ LOC

## 📁 Project Structure

```
unusedfunc/
├── cmd/unusedfunc/          # CLI application
├── pkg/
│   └── ssa/                 # SSA analyzer (integrates with RTA)
├── internal/
│   ├── rta/                 # Modified RTA implementation (core algorithm)
│   ├── funcinfo/            # Function metadata
│   └── testharness/         # Test framework
├── testdata/                # Comprehensive test cases
└── docs/
    ├── architecture.md      # Design decisions, RTA modifications, patterns
    ├── workflows.md         # Build commands, validation, debugging
    ├── performance.md       # Optimization strategies, profiling
    ├── context/             # Session notes (temporal)
    └── reference/
        ├── known-limitations.md   # Template calls, reflection patterns
        ├── rta-algorithm.md       # Modified RTA details (8 enhancements)
        └── validation-results.md  # 20 project test results
```

## 🛠️ Build Commands

```bash
# Build
make build        # → build/unusedfunc
go build -o build/unusedfunc ./cmd/unusedfunc

# Test
make test         # All tests with race detector and coverage
make lint         # golangci-lint (47 linters)

# Run
./build/unusedfunc ./...                    # Analyze current module
./build/unusedfunc --skip-generated ./...   # Skip generated files (default)
./build/unusedfunc -v -json ./...           # Verbose JSON output
```

## 📚 Key Documentation

### Start Here
- **README.md** - User guide, comparison with staticcheck, examples
- **docs/architecture.md** - Complete technical details, all 8 RTA modifications
- **docs/workflows.md** - Development workflows, validation procedures

### Deep Dives
- **docs/reference/rta-algorithm.md** - Modified RTA algorithm explained
- **docs/reference/known-limitations.md** - Template calls, reflection patterns
- **docs/reference/validation-results.md** - Real-world test results

### Quick Lookups
- **docs/performance.md** - Optimization strategies, profiling guide
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
6. Apply Business Rules (unexported + exported in /internal)
   ↓
7. Check Suppressions (//nolint:unusedfunc comments)
   ↓
8. Report Unused Functions
```

### Core Components

**Modified RTA** (`internal/rta/rta.go`)
- Based on golang.org/x/tools@v0.35.0 with heavy modifications
- 8 precision enhancements (see docs/reference/rta-algorithm.md)
- Pattern-based reflection via `knownSafeFunctions` map (50+ functions)
- Fingerprint optimization (96.9% rejection rate)
- Generic template auto-tracking

**SSA Analyzer** (`pkg/ssa/analyzer.go`)
- Integrates with modified RTA
- Entry point detection (main, init, tests, exports, reflection)
- Converts RTA results to `types.Object` set
- Handles assembly scanning, linkname detection

**Function Info** (`internal/funcinfo/`)
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
- Exported functions in non-main packages (library API)
- Functions with runtime directives (`//go:nosplit`, etc.)
- Assembly-implemented or assembly-called functions
- Functions with `//go:linkname` directives
- Functions with `//export` (CGo)
- Reflection targets (via `isPotentialReflectionTarget`)
- SetFinalizer callbacks (detected by RTA)

## 📊 Performance

**Targets**: <2 minutes per 100K LOC ✅ Achieved

**Actual Results** (from validation-results.md):
- Small (<20K LOC): ~1.2s average
- Medium (20-100K LOC): ~8.4s average
- Large (100-300K LOC): ~23.7s average

**Optimizations**:
- Method set reuse (eliminates 3.26GB allocations)
- Interface pre-computation (11x speedup: 523s → 47s on Kubernetes)
- Fingerprint fast-path (96.9% rejection rate)
- Parallel package processing (4-8x speedup)
- xsync.Map for concurrent maps (50-70% lock reduction)

## ✅ Validation Results

Tested on 20 popular Go projects (1.2M+ LOC total):

**Perfect (16/20 - 80% success rate)**:
- go-chi/chi, spf13/cobra, gin-gonic/gin, gorilla/mux, gorilla/websocket
- go-playground/validator, stretchr/testify, nats-io/nats.go, labstack/echo
- lib/pq, go-sql-driver/mysql, hashicorp/consul, etcd-io/etcd
- prometheus/prometheus, containerd/containerd, docker/cli

**Known Limitations (4/20)**:
- golang/protobuf - reflection MethodByName (676 false positives, use `--skip-generated`)
- grpc-go - reflection patterns (150 false positives, use `--skip-generated`)
- go-openapi/spec - schema validation reflection (89 false positives)
- gohugoio/hugo - template methods (43 false positives, use suppression comments)

## 🚨 Known Limitations

### 1. Template Method Calls (Industry Standard)
Methods called from `.gotmpl`, `.tmpl`, `.html` files are flagged as unused.

**Workaround**: Suppression comments
```go
//nolint:unusedfunc // used in template.gotmpl:42
func (t *Type) TemplateMethod() {}
```

### 2. reflect.MethodByName Patterns
Methods called via `reflect.Value.MethodByName("name")` may be flagged.

**Workaround**: Use `--skip-generated` (eliminates ~98% of cases)

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
--skip-generated    # Skip generated files (default: true)
--include-tests     # Include test files in analysis
-v, --verbose       # Show detailed progress
-json               # JSON output format
```

### Suppression Comments
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
4. Test-only? → Check if test files included

See **docs/workflows.md#debugging-reachability-issues** for detailed guide.

## 🎓 Understanding the Code

### Key Algorithms

**Modified RTA** (`internal/rta/rta.go:1-589`)
- Read lines 7-52 for modification summary
- Cross-product tabulation: address-taken × dynamic calls
- Fingerprint optimization for fast `implements()` checks
- Pattern-based reflection via `knownSafeFunctions` (lines 104-150)

**Analyzer Integration** (`pkg/ssa/analyzer.go`)
- `findEntryPoints()` - Detects all reachable roots
- `findReachableMethods()` - Calls `rta.Analyze()` and converts results
- `AnalyzeMethods()` - Main orchestration

**Business Rules** (`internal/funcinfo/funcinfo.go`)
- `ShouldReport()` - Decision tree for reporting
- `IsInInternalPackage()` - `/internal` detection

### Key Data Flow
```
types.Object → FuncInfo → SSA Entry Points → RTA → Reachable Set → Filter → Report
```

## 🔗 Related Tools

**Why not golangci-lint?**
- Requires whole-program SSA analysis (60-80% of time)
- Memory intensive (500MB+ for large codebases)
- Run as separate CI step like benchmarks

**vs staticcheck U1000**
- staticcheck: AST-based, never reports exports
- unusedfunc: RTA-based, reports exports in `/internal`

**vs deadcode**
- deadcode: Same RTA approach, more conservative (all exports live)
- unusedfunc: Opinionated about `/internal` convention

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

## 🎯 Success Metrics

- ✅ Zero false positives on static code (16/20 projects = 80%)
- ✅ Performance <2min per 100K LOC (achieved: ~24s/100K LOC average)
- ✅ Handles interfaces correctly (100% via RTA)
- ✅ Handles generics correctly (100% via template tracking)
- ✅ Respects suppression comments (100%)
- ✅ Clear limitations documented (template calls, MethodByName)

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
unusedfunc -json ./... > unused.json
```

## 📖 Further Reading

- **Precision**: docs/reference/rta-algorithm.md (8 modifications explained)
- **Validation**: docs/reference/validation-results.md (20 project results)
- **Limitations**: docs/reference/known-limitations.md (workarounds included)
- **Architecture**: docs/architecture.md (complete technical details)
- **Performance**: docs/performance.md (optimization strategies)
