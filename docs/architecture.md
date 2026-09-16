# Architecture Decisions

This document captures key architectural decisions and patterns for the unusedfunc project.

## Core Architecture

### Layered Service Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Layer (cmd/)                         │
│  • Configuration Management  • Output Formatting            │
│  • Process Orchestration    • Error Code Handling          │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│              Integration Layer (pkg/unusedfunc)             │
│  • Workflow Orchestration   • Component Coordination       │
│  • Result Aggregation       • Deduplication Logic          │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                Analysis Engine (pkg/ssa)                    │
│  • SSA Construction         • Reachability Root Selection  │
│  • RTA Integration          • Reachable-Object Matching    │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│              Supporting Services (pkg/)                     │
│  • Assembly Scanner         • Runtime Directive Detector   │
│  • Suppression Checker      • Platform-specific Handling   │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│             Reporting Policy (internal/analysis)            │
│  • FuncInfo Types           • Result Structures            │
│  • Business Logic           • Domain Rules                 │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

#### 1. SSA-Based Analysis
- **Decision**: Use Static Single Assignment for analysis
- **Rationale**: Whole-program reachability resolves direct calls, dynamic calls, and interface dispatch from explicit roots
- **Trade-off**: Loading full type information and building SSA retains a substantial program representation
- **Implementation**: `ssa.InstantiateGenerics` mode for generics support

#### 2. Worklist Algorithm
- **Decision**: Worklist-based traversal over recursive
- **Rationale**: Fixed-point reachability handles recursive and cyclic call patterns
- **Location**: `internal/rta.Analyze`, called by `pkg/ssa.Analyzer.findReachableMethods`

`make lint` runs the configured `golangci-lint` checks and `make test` runs the Go test suite with the race detector. See [performance](performance.md) for the supported profiling workflow; this repository has no documented benchmark target.

## Entry Point Detection

### Recognized Entry Points
- `main()` and package `init()` functions in target packages
- Names beginning with `Test`, `Benchmark`, or `Example`
- In normal mode, exported functions and methods in non-`main`, non-`internal` packages
- Exact-name reflection candidates: `String`, `GoString`, `Error`, `Marshal`, `Unmarshal`, `Validate`, `Decode`, and `Encode`
- Functions with recognized runtime or CGo export directives, functions called from assembly, and exported assembly implementations outside `main`

### Entry Point Scope Limitation

SSA is built for all loaded packages and dependencies, but root discovery considers only target packages. In strict mode, public exported functions and methods are omitted from the initial roots.

### Runtime Special Cases
- **SetFinalizer**: RTA recognizes a direct `runtime.SetFinalizer` call and marks its function or closure argument reachable.
- **Address-taken functions**: `Result.Reachable` records an `AddrTaken` flag.
- **Reflection patterns**: The SSA layer roots only the exact candidate names above; it does not infer arbitrary runtime API targets.

## Edge Case Handling

### Assembly Integration
- Uses `pkg.OtherFiles` for build constraint handling
- Scans `.s` files for TEXT/CALL directives
- Maps assembly symbols to Go functions via regex
- Conservative marking of assembly-related functions
- Uses the package loader's build-selected source files

### Build Constraints
- Respects GOOS/GOARCH tags
- Platform-specific analysis support
- Handles conditional compilation

### Suppression Comments
- Supports `//nolint:unusedfunc`
- Supports `//lint:ignore unusedfunc [reason]`
- Position-based suppression tracking
- Applied through `analysis.FuncInfo.ShouldReport()`

## Generics Architecture

### Canonical Function Mapping

When a reachable SSA function has a distinct `Origin()`, RTA also records that generic template. `ReachableObjects` covers selected generic template methods with no SSA function. This prevents a reachable instantiation from leaving its template unmatched.

### SSA Configuration for Generics

SSA is built with `ssa.InstantiateGenerics`. Uninstantiated public generic functions and methods cannot be ordinary RTA roots: a typed-AST fallback retains their references and supplies concrete helpers as roots before RTA runs. This preserves normal interface dispatch within those helpers without treating every generic template as executable. See the [RTA reference](reference/rta-algorithm.md#modification-8-generic-template-tracking) for the fallback's scope.

## Testing Architecture

### Matrix Testing Framework
- **Decision**: Custom test harness with matrix-based validation
- **Rationale**: Comprehensive platform/configuration coverage needed
- **Implementation**: Multiple `BuildConfiguration` per test case
- **Coverage**: GOOS/GOARCH, build tags, CGo states

### TestHarness Structure

`internal/harness` discovers `testdata/*/expected.yaml` and runs each case's build configurations. A configuration may set build tags, CGo, GOOS, GOARCH, and strict mode; it may also list expected errors and expected unused functions.

### Core Components
- **Harness**: `internal/harness/harness.go` loads packages, invokes `unusedfunc.Analyzer`, and compares results.
- **Loader**: `internal/harness/loader.go` applies each configuration.
- **Config**: `internal/harness/config.go` defines the platform configuration types.

### Expected.yaml Format Specification

```yaml
build_configurations:
  - name: "default"
    expected_unused:
      - func: "<module/package>.UnusedFunction"
        reason: "Never called"
        file: "file.go" # Optional suffix
```

Expected results are compared by fully qualified canonical function name, with an optional file-suffix check. The harness does not assert package names or line numbers.

## Package Loading Architecture

### Load Mode Configuration
The tool uses a comprehensive load mode to gather all necessary information:
```go
const defaultLoadMode = packages.NeedDeps |
    packages.NeedName |
    packages.NeedFiles |
    packages.NeedCompiledGoFiles |
    packages.NeedImports |
    packages.NeedTypes |
    packages.NeedSyntax |
    packages.NeedTypesInfo
```

### Package Loading Flow
Located in `pkg/unusedfunc/loader.go`:
- `LoadPackages` calls `packages.Load` with context support and always enables test variants.
- Build tags are passed as one `-tags` build flag.
- The default pattern is `./...`.

### Working Directory Constraints

`LoaderOptions.Dir` lets library callers set the directory used for package loading; otherwise `go/packages` uses the current working directory. The CLI passes no directory override. Package patterns and module behavior are therefore those of `go/packages` in that directory.

## Core Data Structures

### FuncInfo - Central Function Metadata
The `FuncInfo` struct serves as the primary data container throughout analysis:

```go
type FuncInfo struct {
    // Core identification
    Object         types.Object      // Function object from Go type information
    Name           string            // Display name, including generic parameters
    IsUsed         bool              // Reachability result
    IsExported     bool             // Visibility (uppercase name)
    IsInInternal   bool             // Package path analysis
    
    // Suppression handling
    IsSuppressed   bool             // Comment-based exclusion
    
    // Special directives and assembly
    HasLinkname              bool  // //go:linkname directive
    HasRuntimeDirective      bool  // //go:nosplit, //go:norace, etc.
    HasAssemblyImplementation bool  // Function has .s file implementation
    CalledFromAssembly       bool  // Called by assembly code
    HasCGoExport            bool  // //export for CGo
    
    // Source location
    DeclarationPos token.Pos        // Position in source
    Package        *packages.Package // Container package
    Strict         bool              // Strict reporting mode
}
```

**Lifecycle Phases**:
1. **Creation**: Basic metadata from `types.Object`
2. **Enhancement**: Assembly scanning, directive detection
3. **Analysis**: `IsUsed` updates from SSA reachability
4. **Suppression**: `IsSuppressed` updates from comments
5. **Reporting**: `ShouldReport()` evaluation

## Business Logic and Analysis Rules

### Core Analysis Algorithm
The tool builds SSA with generic instantiation, chooses policy roots, and passes concrete roots to RTA. RTA scans reachable functions to a fixed point; collected functions not matched to its reachable objects remain candidates for reporting.

### Entry Point Detection Logic
**Automatic Entry Points**:
- `main()` and package `init()` functions in target packages
- Names beginning with `Test`, `Benchmark`, or `Example`
- In normal mode, exported functions and methods in non-`main`, non-`internal` packages
- The exact reflection-candidate names listed in [Entry Point Detection](#entry-point-detection)

**Special Entry Points Added During Analysis**:
- Functions with recognized runtime directives or CGo export directives
- Exported assembly implementations outside `main`
- Functions named by package-local assembly `CALL ·name(SB)` directives

### Function Usage Rules (ShouldReport Decision Tree)
```
UNUSED := !IsUsed && !IsSuppressed && !HasSpecialCharacteristics
```

**Always Report as Unused**:
1. Unexported unused functions: `!IsExported && !IsUsed`
2. Exported functions in internal packages: `IsExported && !IsUsed && IsInInternal`
3. Exported functions in main packages: `IsExported && !IsUsed && Package.Name == "main"`
4. In strict mode, every unused exported function

**Never Report (Exclusions)**:
- Used functions: `IsUsed == true`
- Suppressed functions: `IsSuppressed == true`
- Functions with runtime directives, linknames, assembly implementations
- Functions called from assembly code or with CGo exports
- In normal mode, exported functions in non-main, non-internal packages (public API)

### Internal Package Detection Rules
```go
IsInInternal := 
    strings.Contains(pkgPath, "/internal/") ||
    strings.HasSuffix(pkgPath, "/internal") ||
    strings.HasPrefix(pkgPath, "internal/") ||
    pkgPath == "internal"
```

## Data Flow Pipeline

```
Source Code
    ↓ (packages.Load)
[]*packages.Package
    ↓ (suppression load and assembly scan)
    ↓ (SSA construction)
SSA Program
    ↓ (FuncInfo collection and RTA reachability)
Final FuncInfo states
    ↓ (ShouldReport() filtering)
[]UnusedFunction
    ↓ (output formatting)
CLI result and formatted output
```

### Result Data Structures
```go
type UnusedFunction struct {
    Name       string           // Function name
    Position   token.Position   // Source location
    Reason     string          // Reporting policy
    Suppressed bool            // Suppression state
    Package    string          // Package path context
}
```

## SSA Mode Configuration
```go
Mode: ssa.InstantiateGenerics | ssa.BareInits
```

RTA processes direct calls, dynamic function calls, interface invokes, `MakeInterface`, `TypeAssert`, and `ChangeInterface`; it also recognizes direct `runtime.SetFinalizer` calls. The RTA reference documents the resulting reachability constraints without duplicating its instruction-level implementation here.

## Limitations

### Package Loading Limitations

The loader delegates package-pattern and module resolution to `go/packages`. A single invocation analyzes only the packages selected by its pattern and build configuration.

### Advanced Reflection Not Supported

The analyzer has no string-flow analysis for `reflect.Value.MethodByName` and does not resolve template text, method-name strings in registries or maps, or struct tags to establish method edges. `reflect.Value.Call` conservatively retains address-taken functions when it is present in the SSA program. Use a nearby suppression for a known dynamic use.

### Template Method Calls Not Supported

#### Technical Explanation

Methods called exclusively from Go template files (`.tmpl`, `.gotmpl`, `.html`) are flagged as unused because template execution uses runtime reflection that is invisible to static analysis.

**Call Chain**:
```
template.Execute() → reflect.Value.MethodByName() → reflect.Value.Call() → YourMethod()
```

Template text resolves method names at runtime, so its invocation is not represented as an SSA edge to the selected method.

#### Example False Positive

```go
// batch.go
type TemplateContext struct {
    opts Options
}

// This method may be reported as unused when no public-API exclusion applies.
func (t *TemplateContext) Export() string {
    return t.opts.Compiled().Export
}

// template.gotmpl
{{ .Export }}  // Runtime reflection call - invisible to SSA
```

#### Workaround

Use suppression comments with clear documentation:

```go
//nolint:unusedfunc // used in batch-esm-runner.gotmpl:15
func (t *TemplateContext) Export() string {
    return t.opts.Compiled().Export
}
```

Or the `lint:ignore` format:

```go
//lint:ignore unusedfunc called from template rendering in Export()
func (t *TemplateContext) Export() string {
    return t.opts.Compiled().Export
}
```

See [known limitations](reference/known-limitations.md) for the current limitation guidance.

## Reflection Handling

`pkg/ssa.Analyzer.isPotentialReflectionTarget` roots package-level SSA functions with the exact names `String`, `GoString`, `Error`, `Marshal`, `Unmarshal`, `Validate`, `Decode`, and `Encode`; it does not inspect methods. `internal/rta.knownSafeFunctions` further narrows empty-interface handling when the current function contains a mapped JSON, fmt, XML, YAML, gob, binary, or `database/sql` call.

Mapped-call narrowing remains a function-context heuristic. Direct `reflect.ValueOf` and `reflect.TypeOf` arguments bypass that narrowing for the converted type; this is a local SSA-use check, not general value-flow analysis. Methods with their own type parameters are not callable through reflection and are excluded. `MethodByName` string resolution, template text, name-based registries, and struct tags remain outside the analyzer's proof of use. Actual function values stored in maps or registries can be tracked through RTA's address-taken-function and dynamic-call matching.

## Interface Compliance System

RTA records runtime types from interface materialization and joins them with interface invoke sites. For a non-empty interface conversion it marks only the required concrete methods, including unexported marker methods. Type assertions and interface-to-interface conversions may scan compatible program types, so this is conservative rather than flow-sensitive. The implementation caches method sets and uses fingerprints as a rejection filter before `types.Implements`.

## API Usage Examples and Contracts

```go
func analyze(ctx context.Context) (map[types.Object]*analysis.FuncInfo, error) {
    pkgs, err := unusedfunc.LoadPackages(ctx, unusedfunc.LoaderOptions{
        Packages: []string{"./..."},
    })
    if err != nil {
        return nil, err
    }

    funcs, err := unusedfunc.NewAnalyzer(unusedfunc.AnalyzerOptions{}).Analyze(pkgs)
    if err != nil {
        return nil, err
    }
    return funcs, nil
}
```

The returned map is keyed by `types.Object`; callers apply `analysis.FuncInfo.ShouldReport` to determine which entries to report. The CLI performs that conversion and output formatting in `cmd/unusedfunc/main.go`.

---

## RTA Integration Strategy

`pkg/ssa.Analyzer.findReachableMethods` filters generic templates from RTA roots, calls `rta.Analyze`, and combines `Result.Reachable` with `Result.ReachableObjects`. It then matches canonical object names to accommodate generic instantiations. See [the RTA reference](reference/rta-algorithm.md) for the fork-specific algorithm.
