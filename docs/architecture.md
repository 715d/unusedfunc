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
│               Integration Layer (pkg/analyzer)              │
│  • Workflow Orchestration   • Component Coordination       │
│  • Result Aggregation       • Deduplication Logic          │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                Analysis Engine (pkg/ssa)                    │
│  • SSA Construction         • Worklist Algorithm           │
│  • Call Graph Analysis      • Interface Resolution         │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│              Supporting Services (pkg/)                     │
│  • Assembly Scanner         • Runtime Directive Detector   │
│  • Suppression Checker      • Platform-specific Handling   │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                Data Model (internal/)                       │
│  • FuncInfo Types           • Result Structures            │
│  • Business Logic           • Domain Rules                 │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

#### 1. SSA-Based Analysis
- **Decision**: Use Static Single Assignment for analysis
- **Rationale**: Zero false positives through precise call graph
- **Trade-off**: Higher memory usage for accuracy guarantee
- **Implementation**: `ssa.InstantiateGenerics` mode for generics support

#### 2. Worklist Algorithm
- **Decision**: Worklist-based traversal over recursive
- **Rationale**: Better memory control and cycle handling
- **Complexity**: O(V+E) where V=functions, E=calls
- **Location**: `SSAAnalyzer.FindReachableMethods` in `pkg/ssa/analyzer.go`

#### 3. Panic Recovery for SSA
- **Decision**: Graceful handling of SSA builder panics
- **Rationale**: Compatibility with golang/go#71899
- **Implementation**: Defer/recover with conservative fallback

## Code Quality Standards

### Quality Enforcement
- 47 enabled linters via `.golangci.yaml`
- `make lint` runs comprehensive static analysis
- Security scanning via gosec
- Zero tolerance policy for quality gates
- Go 1.22+ features adopted (integer range syntax)
- Consistent naming (avoid stuttering in types)

### Error Handling Patterns
- Use `fmt.Errorf` with `%w` for wrapping
- Error chains via `errors.Is/As`
- Lowercase messages without "failed to" prefix
- Sparse custom error types

### Testing Requirements
- Unit tests for exported functions
- Edge case validation via testdata fixtures
- Coverage targets in CI/CD
- Performance benchmarks (<2min/100K LOC)

## Performance Architecture

### Target Metrics
- **Analysis Time**: <2 minutes for 100K LOC
- **Memory**: Linear O(LOC) scaling
- **CPU**: O(V+E) optimal complexity

### Optimization Strategies
- Interface implementation caching
- Direct call tracking with Set[*ssa.Function]
- Pre-computed compliance index
- Parallel package processing opportunity

### Current Bottlenecks
- SSA construction: 60-80% of time
- Reachability analysis: <20% of time
- Memory dominated by SSA representation

## Interface Compliance System

### ComplianceIndex Architecture
- Single source of truth for interface relationships
- O(1) lookup for type-interface mappings
- Handles embedded types and aliases
- Thread-safe concurrent access via xsync.Map

### Interface Conversion Normalization
SSA normalizes ALL implicit conversions to explicit instructions:
- Function arguments: `acceptWriter(mw)` → MakeInterface
- Assignments: `w = mw` → MakeInterface
- Returns: `return mw` → MakeInterface
- Type assertions: `w.(io.Writer)` → TypeAssert
- Channel sends: `ch <- mw` → MakeInterface before send
- Composite literals: `[]Writer{mw}` → MakeInterface for each element

**Key Insight**: SSA builder inserts MakeInterface BEFORE Store/Send/Call instructions, providing centralized handling. No special processing needed for Phi nodes (only merge already-converted values).

### Optimization: Direct Usage Mapping
- Pre-optimization: O(types × interfaces) complexity (52% CPU)
- Post-optimization: Track only actual conversions
- Two-pass analysis: MakeInterface then ChangeInterface
- Result: 50-60% CPU reduction, 83% benchmark improvement

## Entry Point Detection

### Recognized Entry Points
- `main()` functions
- `init()` functions
- `Test*/Benchmark*/Example*` functions
- Exported functions in non-main packages
- Functions with runtime directives
- Assembly-referenced functions

### Entry Point Scope Limitation
- Analyze only requested packages (not AllPackages)
- Prevents RTA panics from excessive entry points
- Maps packages to SSA packages by path matching

### Runtime Special Cases
- **SetFinalizer**: RTA automatically handles via address-taken function tracking
- **Address-taken functions**: Marked with `AddrTaken: true` in RTA results
- **Reflection patterns**: Functions passed to runtime APIs considered reachable

## Edge Case Handling

### Assembly Integration
- Uses `pkg.OtherFiles` for build constraint handling
- Scans `.s` files for TEXT/CALL directives
- Maps assembly symbols to Go functions via regex
- Conservative marking of assembly-related functions
- Cross-architecture support via GOARCH tags

### Build Constraints
- Respects GOOS/GOARCH tags
- Platform-specific analysis support
- Handles conditional compilation

### Suppression Comments
- Supports `//nolint:unusedfunc`
- Supports `//lint:ignore unusedfunc [reason]`
- Position-based suppression tracking
- Maps token.Pos to suppression reason
- Integrated via FuncInfo.ShouldReport() 
 


 


 


 


 


## Generics Architecture

### Canonical Function Mapping
- All instantiations map to template form
- Uses `f.Origin()` for generic templates
- Prevents false positives from instantiations
- Example: `Container[int].Add` → `Container[T].Add`

### SSA Configuration for Generics
- Requires `ssa.InstantiateGenerics` mode
- Object-to-SSA stores canonical functions only
- Worklist uses template consolidation
- 3-10x memory improvement from deduplication

### Generic Naming Convention
- Methods: `*Container[T].MethodName`
- Functions: Standard naming unchanged
- Display uses template form for clarity
- Types.Named.Origin() for type mapping

## Testing Architecture

### Matrix Testing Framework
- **Decision**: Custom test harness with matrix-based validation
- **Rationale**: Comprehensive platform/configuration coverage needed
- **Implementation**: Multiple `BuildConfiguration` per test case
- **Coverage**: GOOS/GOARCH, build tags, CGo states

### TestHarness Structure
The test harness provides sophisticated validation across multiple build configurations:

```go
type TestHarness struct {
    analyzer     *analyzer.Analyzer
    testdataRoot string
    verbose      bool
}

type TestCase struct {
    Name                string
    Description         string
    Dir                 string
    BuildConfigurations []BuildConfiguration
    SkipReason          string
    RequiresSSA         bool
}

type BuildConfiguration struct {
    Name           string
    BuildTags      []string
    EnableCGo      bool
    GOOS           string
    GOARCH         string
    ExpectedUnused []ExpectedFunc
    ExpectedErrors []string
}
```

### Core Components
- **TestHarness**: Core test orchestration in `/testdata/harness/harness.go`
- **TestRunner**: Parallel execution in `/testdata/harness/runner.go`
- **Assertions**: Result validation in `/testdata/harness/assertions.go`
- **Loader**: Package loading in `/testdata/harness/loader.go`
- **Config**: Platform matrix in `/testdata/harness/config.go`

### Position Tolerance System
- **Decision**: ±5 line tolerance for position validation
- **Rationale**: Go toolchain version differences affect line numbers
- **Implementation**: Robust assertion engine with tolerance bounds
- **Benefit**: Stable tests across Go versions and code changes

### Test Category Architecture
The harness supports 11 distinct test scenario categories:

```
testdata/
├── simple/           # Basic unused function scenarios
├── interfaces/       # Interface dispatch and method resolution
├── generics/         # Generic type parameter handling
├── assembly/         # Assembly file integration testing
├── buildconstraints/ # Build tag conditional compilation
├── cgo/              # CGo integration scenarios
├── linkname/         # //go:linkname directive handling
├── runtime/          # Runtime directive processing
├── crosspackage/     # Multi-package dependency analysis
├── edgecases/        # Corner cases and unusual patterns
└── integration/      # Full end-to-end workflows
```

### Expected.yaml Format Specification
The harness uses YAML-based configuration for test expectations:

```yaml
name: "Test case descriptive name"
description: "What is being tested"
requires_ssa: true

build_configurations:
  - name: "default"
    expected_unused:
      - func: "UnusedFunction"
        package: "example.com/pkg"
        reason: "Never called"
        file: "file.go"    # Optional
        line: 42           # Optional
        
  - name: "with-tags"
    build_tags: ["debug", "test"]
    expected_unused:
      - func: "DebugOnlyFunction"
        package: "example.com/pkg"
```

### Matrix Testing Features
- **Build Configurations**: Multiple GOOS/GOARCH/build-tag combinations per test
- **Expected Results**: YAML-based expected unused function specifications
- **Platform Coverage**: Linux, macOS, Windows, FreeBSD support
- **Parallel Execution**: Concurrent test runs with timeout management

### Validation Features
**Function-Level Validation**:
- Specific unused functions identified correctly
- Package path verification
- Optional file path validation
- Optional line number verification (±5 line tolerance)

**Platform-Specific Testing**:
- Different results per build configuration
- Cross-platform compatibility testing
- Build tag conditional compilation
- CGo enabled/disabled scenarios

**Error Handling**:
- Expected error message validation
- Graceful failure reporting
- Clear diagnostic information

### Performance Targets
The test harness enforces these performance requirements:
- **Analysis Time**: <2 minutes for 100K LOC
- **Memory Usage**: <500MB for large codebases
- **Parallel Execution**: Configurable concurrency
- **Timeout Management**: Per-test timeouts

### Golden File Validation
- **Format**: YAML-based expected results
- **Granularity**: Function, package, file, line level
- **Matrix Support**: Per-configuration expectations
- **Determinism**: Sorted processing prevents flaky tests

## Cross-Module Analysis

### Go Ecosystem Convention
- Follows standard Go tool behavior
- Fails with "does not contain main module"
- Guides users to run from target directory
- 90% code reduction vs custom multi-module logic

## Unnamed Interface Detection

### Pre-SSA Scanning
- Scans TypesInfo before SSA construction
- Catches interfaces in unreachable code
- Common locations: parameters, returns, fields
- Prevents false positives for interface compliance

## Package Loading Architecture

### Load Mode Configuration
The tool uses a comprehensive load mode to gather all necessary information:
```go
const DefaultLoadMode = packages.NeedDeps |
    packages.NeedName |
    packages.NeedFiles |
    packages.NeedCompiledGoFiles |
    packages.NeedImports |
    packages.NeedTypes |
    packages.NeedTypesSizes |
    packages.NeedSyntax |
    packages.NeedTypesInfo |
    packages.NeedModule
```

### Package Loading Flow
Located in `cmd/unusedfunc/main.go`:
- Uses `packages.Load()` with context support for cancellation
- Always loads test files to detect usage patterns from tests
- Supports build tags via `-tags` flag passed to `packages.Config`
- Default pattern: `./...` if no arguments provided

### Working Directory Constraints
- Package loading uses current working directory exclusively
- No dynamic module root detection or directory changes
- Patterns resolved relative to current directory only
- Single-module analysis (cross-module not supported)
- Tool must be run from target module's directory

## Core Data Structures

### FuncInfo - Central Function Metadata
The `FuncInfo` struct serves as the primary data container throughout analysis:

```go
type FuncInfo struct {
    // Core identification
    Object         types.Object      // types.Func from Go compiler
    IsUsed         bool             // Reachability result
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
}
```

**Lifecycle Phases**:
1. **Creation**: Basic metadata from `types.Object`
2. **Enhancement**: Assembly scanning, directive detection
3. **Analysis**: `IsUsed` updates from SSA reachability
4. **Suppression**: `IsSuppressed` updates from comments
5. **Reporting**: `ShouldReport()` evaluation

### SSA Analyzer State Management
The `SSAAnalyzer` maintains complex internal state:

```go
type Analyzer struct {
    // Immutable foundation (set once)
    program     *ssa.Program          // SSA representation
    callgraph   *callgraph.Graph      // Static call graph
    packages    []*packages.Package   // Source packages
    
    // Mutable analysis state (grows during analysis)
    objectToSSA      map[types.Object]*ssa.Function  // Key mapping
    entryPoints      []*ssa.Function                 // Reachability roots
    implementorCache map[string][]types.Object       // Interface implementations
    liveTypes        map[types.Type]bool             // Type instantiation tracking
}
```

## Business Logic and Analysis Rules

### Core Analysis Algorithm
The tool employs a **worklist-based reachability analysis**:

```
1. Build SSA program with InstantiateGenerics mode
2. Identify entry points (main, init, tests, exported functions)
3. Initialize worklist with all entry points
4. Process worklist until empty:
   - Pop function from worklist
   - Inspect all instructions for function references
   - Add newly discovered functions to worklist
   - Mark all discovered functions as reachable
5. Any function not marked as reachable is potentially unused
```

### Entry Point Detection Logic
**Automatic Entry Points**:
- `main()` functions in any package
- `init()` functions in any package
- Test functions: `Test*`, `Benchmark*`, `Example*`
- Exported functions in non-main packages (library APIs)
- Runtime reflection targets with common patterns

**Special Entry Points Added During Analysis**:
- Functions with runtime directives (`//go:nosplit`, `//go:noinline`, etc.)
- CGo exported functions (`//export` directives)
- Assembly-implemented exported functions
- Assembly-called functions via `CALL ·funcName(SB)`

### Function Usage Rules (ShouldReport Decision Tree)
```
UNUSED := !IsUsed && !IsSuppressed && !HasSpecialCharacteristics
```

**Always Report as Unused**:
1. Unexported unused functions: `!IsExported && !IsUsed`
2. Exported functions in internal packages: `IsExported && !IsUsed && IsInInternal`
3. Exported functions in main packages: `IsExported && !IsUsed && Package.Name == "main"`

**Never Report (Exclusions)**:
- Used functions: `IsUsed == true`
- Suppressed functions: `IsSuppressed == true`
- Functions with runtime directives, linknames, assembly implementations
- Functions called from assembly code or with CGo exports
- Exported functions in non-main, non-internal packages (public API)

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
    ↓ (FuncInfo creation)
map[types.Object]*FuncInfo
    ↓ (SSA construction)
SSA Program + Call Graph
    ↓ (Reachability analysis)
Updated FuncInfo.IsUsed flags
    ↓ (Suppression application)
Final FuncInfo states
    ↓ (ShouldReport() filtering)
[]UnusedFunction
    ↓ (Output formatting)
FinalResult
```

### Result Data Structures
```go
type UnusedFunction struct {
    Name       string           // Function name
    Position   token.Position   // Source location
    Reason     string          // "unexported" or "exported in internal"
    Suppressed bool            // Excluded from reporting
    Package    string          // Package path context
}

type FinalResult struct {
    UnusedFunctions []UnusedFunction  // Functions to report
    Stats           Stats             // Analysis metrics
}
```

## SSA Mode Configuration
```go
Mode: ssa.SanityCheckFunctions | ssa.InstantiateGenerics | ssa.BareInits
```

**Error Handling**: Graceful panic recovery for golang/go#71899 issues with generic instantiation. Falls back to analysis without generic specialization if `InstantiateGenerics` fails.

## Instruction Analysis Details

### Analyzed Instruction Types
1. **Static function calls**: `callCommon.Value` points to `*ssa.Function`
2. **Interface method dispatch**: `callCommon.IsInvoke() == true`
3. **Function values**: Functions referenced as operands (closures, callbacks)
4. **Type assertions**: `*ssa.TypeAssert` instructions
5. **Interface conversions**: `*ssa.MakeInterface` instructions
6. **Change interface**: `*ssa.ChangeInterface` for interface-to-interface
7. **Return statements**: Implicit interface conversions at returns

**Validated via**: `debug/analyze_ssa_conversions.go` tool confirms SSA normalization

### Interface Method Resolution Algorithm
```go
// For interface calls: receiver.Method()
1. Find interface type from receiver
2. For each method in objectToSSA mapping:
   - Check if method name matches
   - Check if receiver type implements interface
   - Check if receiver type is actually instantiated
   - If all true: mark method as reachable
```

### Type Instantiation Tracking
- **Allocation tracking**: `*ssa.Alloc` instructions
- **Interface conversions**: `*ssa.MakeInterface` instructions  
- **Dereference operations**: `*ssa.UnOp` with `token.MUL`
- **Embedded type tracking**: Recursively track embedded struct fields

## State Management Patterns

### Design Principles
1. **Single Source of Truth**: FuncInfo centralizes all function metadata
2. **Immutable Core, Mutable Analysis**: SSA structures vs analysis state
3. **Position-Based Mapping**: `token.Pos` as consistent identifier
4. **Layered Enhancement**: Progressive FuncInfo enrichment
5. **Graceful Degradation**: Optional features (assembly) don't fail analysis

### Performance Optimizations
- **Interface implementor cache**: `implementorCache map[string][]types.Object`
- **Type instantiation cache**: `liveTypes map[types.Type]bool`
- **Object-to-SSA mapping**: `objectToSSA map[types.Object]*ssa.Function`
- **Worklist efficiency**: Functions processed only once via deduplication
- **Early termination**: Skip functions without bodies (except assembly-implemented)
- **Concurrent maps**: xsync.Map for `concreteToInterfaces`/`interfaceToConcretes` (50-70% lock reduction)
- **Atomic operations**: `LoadOrCompute()` pattern for thread-safe initialization

## Limitations

### Package Loading Limitations
- **Single Working Directory**: Cannot analyze packages outside current module
- **No Multi-Module Support**: Limited by current directory constraint
- **Pattern Resolution**: All patterns relative to current working directory
- **No Dynamic Module Detection**: Must run from correct directory

### Advanced Reflection Not Supported
The analyzer cannot detect dynamic function calls through reflection:
- Functions passed to `reflect.MakeFunc()`
- Dynamic method dispatch based on naming patterns
- Functions referenced in struct tags
- RPC-style method registration and calling
- Functions stored in registries or maps for dynamic invocation

### Known False Negatives
- Reflection-based function calls not matching common patterns
- Dynamic loading via `go:linkname` to external packages
- Build tag combinations not covered by current analysis run

**Workaround**: Use suppression comments (`//nolint:unusedfunc`) for these cases.

### Template Method Calls Not Supported

**Status**: Known and accepted limitation (matches industry standards)

#### Technical Explanation

Methods called exclusively from Go template files (`.tmpl`, `.gotmpl`, `.html`) are flagged as unused because template execution uses runtime reflection that is invisible to static analysis.

**Call Chain**:
```
template.Execute() → reflect.Value.MethodByName() → reflect.Value.Call() → YourMethod()
```

The SSA (Static Single Assignment) call graph construction only tracks statically determinable calls. The reflection-based method invocation used by the template engine creates a dynamic call chain that cannot be analyzed statically.

#### Why This Limitation Exists

1. **Runtime Resolution**: Template method names are resolved at runtime from template text strings
2. **Reflection-Based Invocation**: Methods are called via `reflect.Value.Call()`, not direct Go calls
3. **No Static Call Edges**: The SSA call graph has no edges from `template.Execute()` to template methods
4. **Same as All Tools**: This limitation exists in `staticcheck`, `deadcode`, and all major static analyzers

#### Industry Standard Behavior

All major Go static analysis tools have this limitation:

- **`staticcheck` (U1000)**: Requires manual `//lint:ignore U1000` suppression for template methods
- **`deadcode`**: Uses identical RTA algorithm, cannot detect template calls
- **`golangci-lint`**: Inherits limitation from underlying analyzers

This is not a bug in `unusedfunc`, but a fundamental constraint of static analysis when dealing with reflection-based invocation.

#### Example False Positive

```go
// batch.go
type TemplateContext struct {
    opts Options
}

// This method will be flagged as unused.
func (t *TemplateContext) Export() string {
    return t.opts.Compiled().Export
}

// template.gotmpl
{{ .Export }}  // Runtime reflection call - invisible to SSA
```

The SSA representation shows:
```
No call edge from template.Execute() to TemplateContext.Export()
↓
Export() appears unreachable
↓
Flagged as unused (false positive)
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

See @docs/reference/known-limitations.md for comprehensive limitation documentation.

#### Future Enhancement

Template parsing support could be added via an opt-in flag:

**Proposed Implementation**:
1. Parse `.tmpl`, `.gotmpl`, `.html` files
2. Extract `{{ .MethodName }}` action patterns
3. Mark referenced methods as entry points
4. Opt-in via `--include-templates` flag (default: false)

**Implementation Effort**: 4-8 hours
**Priority**: Low (workaround via suppression comments is sufficient)

## Reflection Handling

### Current State and Limitations

The unusedfunc analyzer has **limited reflection support** with conservative patterns for common use cases. The `reflectadvanced` test case in testdata demonstrates the current limitations and potential improvements.

### Supported Reflection Patterns

#### Basic Pattern Detection
The current `isPotentialReflectionTarget` function handles basic patterns:
- **fmt interfaces**: `String()`, `GoString()`, `Error()` methods
- **Encoding patterns**: `Marshal()`, `Unmarshal()` methods  
- **Common patterns**: `Validate()`, `Decode()`, `Encode()` methods

#### Conservative Approach
- Functions matching common reflection patterns are marked as potentially used
- Reduces false positives at the cost of some false negatives
- Simple string-based matching for method names

### Unsupported Reflection Patterns

#### Very Hard to Implement
1. **reflect.MakeFunc Patterns**
   ```go
   func addImpl(args []reflect.Value) []reflect.Value { ... }
   fn := reflect.MakeFunc(fnType, addImpl)  // addImpl usage undetected
   ```
   - Requires sophisticated data flow analysis
   - Track function arguments to `reflect.MakeFunc` calls
   - Currently beyond analyzer capabilities

#### Medium Difficulty Patterns  
2. **Method Dispatch via Reflection**
   ```go
   func (c *Calculator) OpAdd(a, b float64) float64 { ... }     // USED via pattern
   func (c *Calculator) Divide(a, b float64) float64 { ... }   // UNUSED (no "Op" prefix)
   
   if strings.HasPrefix(method.Name, "Op") {  // Pattern matching
       calc.operations[opName] = calcValue.Method(i)
   }
   ```
   - Analyzable through string operation analysis
   - Method name pattern detection

3. **Struct Tag Processing**  
   ```go
   type Model struct {
       ID   int    `handler:"processID"`     // processID function used via tag
       Age  int    `process:"false"`         // processAge unused (disabled)
   }
   ```
   - Requires struct tag parsing
   - Extract function names from tag values

4. **RPC-Style Method Registration**
   ```go
   // All exported methods automatically registered
   for i := 0; i < t.NumMethod(); i++ {
       method := t.Method(i)
       if method.PkgPath == "" {  // Only exported methods
           s.methods[method.Name] = v.Method(i)
       }
   }
   ```
   - Pattern: all exported methods of registered types are used
   - Type-based registration detection

#### Easy to Implement
5. **Function Registry Pattern**
   ```go
   func init() {
       RegisterFunc("process", processData)     // processData used via registration
       RegisterFunc("transform", transformData) // transformData used via registration  
   }
   ```
   - Track function arguments to registration functions
   - Simple call analysis for registry patterns

### Feasibility Analysis

#### Implementable Improvements (Recommended)
- **Function Registry Pattern**: Track `RegisterFunc` calls and similar patterns
- **RPC Registration**: Mark exported methods as used when type is registered
- **Basic Naming Patterns**: Analyze `strings.HasPrefix` usage for method dispatch
- **Struct Tag Parsing**: Extract function references from struct tag values

#### Expected Coverage Improvement
- **Current**: ~30% of reflection patterns handled
- **Phase 1 (Easy patterns)**: ~70-80% coverage  
- **Phase 2 (Medium patterns)**: ~90-95% coverage
- **Phase 3 (Hard patterns)**: ~100% coverage (significant implementation effort)

### Implementation Strategy

#### Phase 1: Basic Improvements
1. Remove conservative skipping of reflection test cases
2. Implement function registry pattern detection
3. Add RPC-style exported method handling
4. Expect some false negatives but much better coverage

#### Phase 2: Advanced Pattern Analysis  
1. Implement struct tag parsing for function references
2. Add method dispatch pattern analysis
3. Interface-based registration system support

#### Phase 3: Complete Coverage (Future)
1. Complex data flow analysis for `reflect.MakeFunc`
2. Dynamic call pattern inference
3. Comprehensive reflection usage detection

### Current Recommendation

**Remove conservative skipping** - The current approach of skipping reflection-heavy test cases is too conservative. The analyzer can handle most reflection patterns with moderate implementation effort, providing significant value while documenting remaining limitations clearly.

### Testing and Validation

The `testdata/reflection-and-embedding/` test case should be enabled with:
1. Implementation of basic reflection patterns
2. Clear documentation of unsupported cases  
3. Suppression comment recommendations for complex cases
4. Performance impact assessment for reflection analysis

This balanced approach provides practical value while maintaining realistic expectations about reflection analysis limitations.

## Interface Compliance System

Interface compliance tracking is a critical subsystem that determines which concrete types implement which interfaces, enabling accurate detection of used interface methods. This system prevents false positives when methods are only used through interface dispatch.

### ComplianceIndex Architecture
- **Location**: `pkg/ssa/compliance_index.go`
- **Builder**: `pkg/ssa/compliance_index_builder.go`
- **Purpose**: Pre-compute which types implement which interfaces
- **Performance**: O(1) lookup after O(n*m) build phase
- Single source of truth for interface relationships
- Handles embedded types and aliases
- Thread-safe concurrent access via xsync.Map

### Key Components

#### 1. Type Collection (HIGH Priority)
- Scans ALL packages (not just analysis targets)
- Collects concrete types and interfaces from entire dependency tree
- Handles embedded types and type aliases
- Essential for cross-package interface implementations

#### 2. Method Matching (HIGH Priority)
- Signature comparison for interface compliance
- Receiver type handling (value vs pointer receivers)
- Generic parameter substitution and normalization
- Exact matching required for Go interface semantics

#### 3. Interface Normalization (MEDIUM Priority)
- Template forms: `Comparable[T]` → normalized template
- Instantiated forms: `Comparable[int]` → specific instantiation
- Tracks both forms for complete coverage
- Handles generic interface instantiations properly

### Interface Conversion Normalization
SSA normalizes ALL implicit conversions to explicit instructions:
- Function arguments: `acceptWriter(mw)` → MakeInterface
- Assignments: `w = mw` → MakeInterface
- Returns: `return mw` → MakeInterface
- Type assertions: `w.(io.Writer)` → TypeAssert
- Channel sends: `ch <- mw` → MakeInterface before send
- Composite literals: `[]Writer{mw}` → MakeInterface for each element

**Key Insight**: SSA builder inserts MakeInterface BEFORE Store/Send/Call instructions, providing centralized handling. No special processing needed for Phi nodes (only merge already-converted values).

### Optimization: Direct Usage Mapping
- Pre-optimization: O(types × interfaces) complexity (52% CPU)
- Post-optimization: Track only actual conversions
- Two-pass analysis: MakeInterface then ChangeInterface
- Result: 50-60% CPU reduction, 83% benchmark improvement

### Common Problems & Solutions

#### Problem: Interface Methods Marked Unused
**Symptoms**: 
- Method implements interface but reported as unused
- Only occurs with interface dispatch calls
- Direct method calls work correctly

**Root Causes**:
1. **Type not collected from dependency**: Missing from compliance index
2. **Interface not normalized correctly**: Generic interface handling issues
3. **Method signature mismatch**: Subtle differences in signatures
4. **Cross-package interface**: Interface and implementation in different packages

**Solutions**:
- Ensure `AllPackages()` used for type collection, not just target packages
- Verify interface normalization handles generic types correctly
- Check method signatures match exactly (parameter names don't matter)
- Confirm both interface and implementing types are scanned

#### Problem: Generic Interface Compliance
**Symptoms**:
- `IntComparable.Compare` implements `Comparable[int]` interface
- Method still marked as unused despite proper implementation
- Only affects generic interfaces, regular interfaces work

**Root Cause**: Generic interface instances not properly tracked in compliance index

**Solution**:
1. Normalize generic interfaces to template form
2. Track instantiations separately from templates
3. Map both forms in compliance checking
4. Use `types.Named.Origin()` for canonical type mapping

#### Problem: Cross-Package Interfaces
**Symptoms**:
- Interface defined in package A
- Implementation in package B
- Method marked unused despite being used through interface

**Root Cause**: Incomplete package scanning missing dependency packages

**Solution**:
- Use comprehensive package collection including all dependencies
- Build compliance index from complete package set
- Ensure SSA construction includes all relevant packages

### Building Compliance Index

```go
// Located in pkg/ssa/compliance_index_builder.go:Build
func (cb *ComplianceIndexBuilder) Build() *ComplianceIndex {
    // 1. Collect all types from all packages
    for _, pkg := range cb.packages {
        cb.collectTypes(pkg)
    }
    
    // 2. For each concrete type, check against all interfaces
    for concreteType := range cb.concreteTypes {
        for interfaceType := range cb.interfaceTypes {
            if types.Implements(concreteType, interfaceType) {
                cb.recordImplementation(concreteType, interfaceType)
            }
        }
    }
    
    // 3. Handle generic instantiations
    cb.processGenericInstantiations()
    
    // 4. Build reverse mappings for efficient lookup
    return cb.buildIndex()
}
```

### Querying Compliance

```go
// Located in pkg/ssa/compliance_index.go:ConcreteImplements
func (ci *ComplianceIndex) ConcreteImplements(concrete types.Type, method string) []types.Type {
    // O(1) map lookup for implemented interfaces
    interfaces := ci.implementationMap[concrete]
    
    var matching []types.Type
    for _, iface := range interfaces {
        if ci.interfaceHasMethod(iface, method) {
            matching = append(matching, iface)
        }
    }
    
    return matching
}
```

### Interface Method Analysis - Advanced Patterns

#### Key Analysis Findings

**1. Perimeter Methods Correctly Unused (HIGH Impact)**
- `*Circle.Perimeter` and `*Rectangle.Perimeter` correctly detected as unused
- Although part of `Shape` interface, never called via interface dispatch
- Only `Area()` invoked through `CalculateShapeArea(shape Shape)`
- SSA correctly tracks that `Perimeter()` has no call sites

**2. DataProcessor Methods Not Reported (HIGH Impact)**
- `*DataProcessor.GetProcessedCount` and `*DataProcessor.IncrementErrors` not unused
- `DataProcessor` embedded in `FileProcessor` type
- Embedded methods become reachable through parent instantiation
- Correct behavior per Go embedding semantics

**3. SSA Analysis Correctness (MEDIUM Impact)**
- Tracks interface method calls through `trackActualInterfaceCalls()`
- Identifies which interface methods are actually invoked
- Handles embedded types correctly via method set propagation
- Distinguishes between interface declaration and actual usage

### Technical Implementation Details
**Interface Dispatch Tracking**:
1. Build SSA with interface types using `ssa.InstantiateGenerics`
2. Track method calls through interface values via `trackActualInterfaceCalls()`
3. Mark methods as used only if invoked through call sites
4. Handle embedded method sets according to Go language specification

### Performance Optimizations

**Pre-Computation Strategy**:
- Compliance index built once at startup
- All interface relationships computed upfront
- Runtime queries are O(1) map lookups
- No expensive computation during analysis

**Concurrent Safety**:
- Thread-safe implementation using `xsync.Map`
- Safe for concurrent read access during analysis
- Write operations only during initialization phase

**Memory Optimization**:
- Minimal memory overhead through efficient data structures
- Shared references to avoid duplication
- Lazy evaluation for rarely-used mappings

### Debugging Workflow

**Quick Diagnosis Steps**:
1. **Start here first** when methods implementing interfaces are marked unused
2. **Check the compliance index** - verify interface is being tracked
3. **Verify type collection** - ensure ALL packages are scanned, not just targets
4. **Test with simple cases** first (e.g., `io.Writer`, `fmt.Stringer`)

**Investigation Process**:
```
1. Check if interface is indexed:
   - Look in ComplianceIndex.methodToInterfaces map
   - Verify interface type is from scanned packages

2. Check type implementation:
   - Use types.Implements() manually to verify
   - Check both value and pointer receivers
   
3. Verify method signatures:
   - Exact match required for interface compliance
   - Parameter names don't matter, types must match exactly
   - Return types must match exactly
   
4. For generics:
   - Check both template and instantiated forms
   - Look for normalization issues in generic handling
   - Verify Origin() mapping is working correctly
```

**Performance Considerations**:
- **Compliance checking** is often the analysis bottleneck
- **Pre-computation is essential** for acceptable performance
- **Direct usage mapping** reduces compliance checks by 98%
- **Cache hit rate should be >90%** for optimal performance

**Key Files to Examine**:
When debugging interface compliance issues:
- `pkg/ssa/compliance_index.go` - Core compliance logic
- `pkg/ssa/compliance_index_builder.go` - Index building process
- `testdata/interface-compliance-only/` - Basic test cases
- `testdata/interface-dispatch-complex/` - Advanced scenarios
- `testdata/generic-complex-patterns/` - Generic interface tests

This system is fundamental to preventing false positives while maintaining zero false negatives for interface method usage detection.

## API Usage Examples and Contracts

### Main Analysis Pipeline

```go
// Example: Analyzing a codebase for unused functions
func analyzeProject() error {
    // 1. Load packages with full type information
    cfg := &packages.Config{
        Mode: packages.NeedDeps |
              packages.NeedName |
              packages.NeedFiles |
              packages.NeedCompiledGoFiles |
              packages.NeedImports |
              packages.NeedTypes |
              packages.NeedTypesSizes |
              packages.NeedSyntax |
              packages.NeedTypesInfo |
              packages.NeedModule,
    }
    
    pkgs, err := packages.Load(cfg, "./...")
    if err != nil {
        return fmt.Errorf("loading packages: %w", err)
    }
    
    // 2. Create and configure analyzer
    analyzer := pkg.NewAnalyzer()
    
    // 3. Run comprehensive analysis
    result, err := analyzer.Analyze(pkgs)
    if err != nil {
        return fmt.Errorf("analysis failed: %w", err)
    }
    
    // 4. Process results
    for _, unused := range result.UnusedFunctions {
        fmt.Printf("%s: %s function is unused\n", 
            unused.Position, unused.Name)
    }
    
    return nil
}
```

### Custom Entry Point Detection

```go
// Example: Adding custom entry points for framework patterns
func addFrameworkEntryPoints(analyzer *ssa.Analyzer) {
    // Add HTTP handler patterns
    analyzer.AddEntryPattern("Handle*")
    
    // Add test initialization patterns
    analyzer.AddEntryPattern("init*")
    
    // Add reflection targets
    analyzer.AddReflectionTarget("Validate")
    analyzer.AddReflectionTarget("Marshal*")
}
```

### Suppression Management

```go
// Example: Managing suppressions programmatically
func manageSuppression(checker *suppress.SuppressionChecker) {
    // Check if function is suppressed
    if suppressed, reason := checker.IsSuppressed(pos); suppressed {
        fmt.Printf("Function suppressed: %s\n", reason)
    }
    
    // Get suppression metrics
    metrics := checker.GetSuppressionMetrics()
    fmt.Printf("Total suppressions: %d\n", metrics["total"])
}
```

### Extension Points

#### Custom Analyzers
Implement additional static analysis by:
1. Extending `FuncInfo` with new metadata fields
2. Adding scanning logic in `analyzer.Analyze()`
3. Modifying reporting logic in `FuncInfo.ShouldReport()`

#### Custom Entry Points
Add new entry point detection in `ssa.findEntryPoints()`:
1. Pattern-based function name matching
2. Annotation-based detection
3. External configuration support

#### Output Formats
Extend CLI output by:
1. Adding new format flags
2. Implementing format-specific serialization
3. Supporting structured output formats (XML, CSV, etc.)

---

## RTA Integration Strategy

### Architecture Evolution
The analyzer evolved from manual SSA traversal (~600 lines) to leveraging RTA's built-in call graph capabilities. This eliminated redundant code while improving precision.

**Key Insight**: RTA distinguishes between:
- **Interface dispatch**: Methods actually called through interfaces (tracked in call graph)
- **Interface compliance**: Methods required for interface satisfaction (marked via MakeInterface/ChangeInterface/TypeAssert)

**Refactoring Benefits (2025-08-31)**:
- Deleted 3 compliance index files (~730 lines) + 4 analysis functions (~200 lines)
- Net reduction: ~600+ lines (75% of SSA analysis code)
- Final size: analyzer.go reduced from ~700 to 474 lines
- Memory: +15-20% for call graph (acceptable tradeoff)
- Performance: Similar or faster, improved precision

### Implementation Pattern
```go
// Leverage RTA's call graph for precise reachability
result := rta.Analyze(entryPoints, true)  // buildCallGraph: true

// Extract reachable functions from call graph
reachable := make(map[types.Object]bool)
for fn := range result.CallGraph.Nodes {
    if fn.Object() != nil {
        reachable[fn.Object()] = true
    }
}

// Only manual addition: runtime.SetFinalizer tracking
// (GC callbacks aren't in normal call graph)
finalizerFuncs := findFinalizerFunctions(result.Reachable)
for fn := range finalizerFuncs {
    reachable[fn] = true
}
```

### Generic Handling
Generic instantiated methods have special characteristics:
- `fn.Package() == nil` (no package association)
- Names include type parameters: `(*Container[int]).Add[int]`
- Must track via `fn.Origin()` for template association
- **Never filter by package** when iterating call graph nodes

### Interface Conversion Normalization
SSA normalizes all interface conversions to three instruction types:
1. **MakeInterface**: Direct conversion `var w Writer = &Buffer{}`
2. **ChangeInterface**: Interface-to-interface `var rc io.ReadCloser = r`
3. **TypeAssert**: Runtime checks `w.(*Buffer)`

All three must be tracked to detect interface usage patterns.

### Fingerprint Optimization
Method set comparison is expensive. Use CRC32 fingerprinting for fast rejection:
```go
// Fast path: interface method bits must be subset of concrete type bits
if iface.fprint & ^concrete.fprint != 0 {
    return false  // 96.9% of checks rejected here
}
// Slow path: full types.Implements() check
return types.Implements(concrete.C, iface.I)
```

**Impact**: Rejects 96.9% of non-implementing types with single bitwise AND operation.

---

## Known Bugs

### Bug #1: Reflection MethodByName (High Priority)
**Status**: Documented, test case exists, not yet fixed
**Test**: `testdata/reflection-method-by-name/`
**Impact**: Methods called via `reflect.MethodByName("name")` incorrectly reported as unused
**Workaround**: `--skip-generated` flag (default: true) eliminates ~98% of cases

**Example**:
```go
type Handler struct{}
func (h *Handler) Process() {}  // Reported as unused ❌

func main() {
    h := &Handler{}
    reflect.ValueOf(h).MethodByName("Process").Call(nil)  // Actually called
}
```

**Why**: SSA doesn't track string constant arguments to connect MethodByName calls to actual methods.

**Workaround**: Use `--skip-generated` flag (eliminates ~98% of cases) or suppression comments.

**Real-World**: golang/protobuf (676 false positives), grpc-go (150 false positives)

### Bug #2: Interface Marker Methods (FIXED 2025-11-06)
**Status**: ✅ FIXED
**Test**: `testdata/interface-empty-marker-methods/` - PASSING
**Fix**: Changed `addRuntimeTypeForInterface()` to pass `method.Pkg()` instead of `nil` to `mset.Lookup()` to find unexported methods, and removed exported-only filter

**Example**:
```go
type Validator interface {
    isValidator()  // Marker method - never called directly
}

type EmailValidator struct{}
func (*EmailValidator) isValidator() {}  // Reported as unused ❌

func Process(v Validator) {
    // Uses interface but never calls isValidator()
}
```

**Why**: Method never invoked via dynamic dispatch, only used for type constraint satisfaction.

**Workaround**: Use `--skip-generated` flag or suppression comments.

**Real-World**: golang/protobuf (29 false positives), affects sum types and branded type patterns.

---

## Performance Optimizations

### Method Set Reuse (IMPLEMENTED 2025-09-04)
**Problem**: Creating new method sets repeatedly causes 3.26GB allocations.
**Solution**: Reuse `program.MethodSets.MethodSet(T)` instead of `types.NewMethodSet(T)`.
**Implementation**:
```go
// ❌ WRONG - Creates new method set
mset := types.NewMethodSet(T)

// ✅ CORRECT - Reuses cached method set
mset := analyzer.program.MethodSets.MethodSet(T)
```
**Impact**: Eliminates 3.26GB allocations on large codebases.

### Interface Implementation Pre-computation (IMPLEMENTED 2025-09-06)
**Problem**: O(n×m) complexity checking all types against all interfaces.
**Solution**: Build ComplianceIndex lazily on first use, cache results.
**Implementation**:
```go
// Build implementation index lazily
func (idx *ComplianceIndex) getImplementers(iface *types.Interface) []types.Type {
    if idx.cache == nil {
        idx.buildIndex()  // One-time O(n×m) cost
    }
    return idx.cache[iface]
}
```
**Impact**: 11x speedup (523s → 47s on Kubernetes).

### Fingerprint Fast-Path
**Problem**: `types.Implements()` is expensive, called millions of times.
**Solution**: CRC32 bitfield rejection before expensive check:
```go
func implements(concrete, iface *TypeInfo) bool {
    // Fast rejection: interface bits must be subset
    if iface.fprint & ^concrete.fprint != 0 {
        return false  // 96.9% of checks rejected here
    }
    // Slow path: full validation
    return types.Implements(concrete.C, iface.I)
}
```
**Impact**: Rejects 96.9% of type checks with bitwise AND operation.

### Parallel Package Processing (IMPLEMENTED 2025-09-04)
**Problem**: Sequential package analysis underutilizes CPU.
**Solution**: Use `slices.Chunk` with goroutines for parallel processing.
**Implementation**:
```go
chunks := slices.Chunk(packages, runtime.GOMAXPROCS(0))
var wg sync.WaitGroup
for _, chunk := range chunks {
    wg.Add(1)
    go func(pkgs []*packages.Package) {
        defer wg.Done()
        for _, pkg := range pkgs {
            analyzer.analyzePackage(pkg)
        }
    }(chunk)
}
wg.Wait()
```
**Impact**: 4-8x speedup on multi-core systems.

### GOGC Tuning
**Problem**: Frequent GC pauses on large codebases (Kubernetes, Hugo).
**Recommendation**: Set `GOGC=400` for large codebases to reduce GC pause frequency.
**Tradeoff**: Higher memory usage for faster analysis.
**Usage**: See @docs/workflows.md#gogc-tuning for operational details.
