# Modified RTA (Rapid Type Analysis) Algorithm Reference

## Quick Reference
- **Source**: Modified fork of `golang.org/x/tools@v0.35.0/go/callgraph/rta/rta.go`
- **Current module dependency**: `golang.org/x/tools v0.50.0`
- **Location**: `/internal/rta/rta.go`
- **Purpose**: Reachability analysis for unused-function detection with 8 fork-specific changes
- **Algorithm**: Cross-product tabulation with pattern-based reflection handling
- **Entry Point**: `rta.Analyze(roots []*ssa.Function) *rta.Result`
- **Returns**: reachable functions, runtime types, and reachable objects

## Overview

This is a **heavily modified fork** of the upstream RTA algorithm for unused-function detection. The changes tailor reachability to known reflection contexts, interface materialization, finalizers, and generic templates.

### Fork-Specific Behavior

1. **Pattern-based reflection handling** - Uses listed method names in known function contexts
2. **Precise non-empty interface conversions** - Only marks interface-required methods
3. **Context-aware analysis** - Tracks calling context for smarter decisions
4. **Enhanced interface compliance** - Handles `*Interface → any` conversions
5. **TypeAssert support** - Conservative type assertion handling
6. **ChangeInterface support** - Interface-to-interface conversion tracking
7. **SetFinalizer detection** - Marks GC finalizer functions as reachable
8. **Generic template tracking** - Auto-tracks templates when instantiations are marked

## Core Algorithm (Modified)

### High-Level Flow
```
1. Initialize with the roots supplied by the caller
2. While worklist is not empty:
   a. Pop function from worklist
   b. Set as current function (for context)
   c. Analyze function instructions with enhanced logic:
      - Check if calls to known safe functions (knownSafeFunctions map)
      - Handle TypeAssert/ChangeInterface with fingerprint optimization
      - Detect SetFinalizer patterns
      - Track generic template relationships
   d. Update cross-product tables
   e. Add newly reachable functions to worklist
3. Return results (reachable functions + reachable objects)
```

## Key Data Structures (Modified)

### Result (Enhanced)
```go
type Result struct {
    // Reachable functions with address-taken flag
    Reachable map[*ssa.Function]struct{ AddrTaken bool }

    // Tracks generic templates and methods without SSA functions
    ReachableObjects map[types.Object]bool

    // Runtime types needed for interfaces/reflection
    RuntimeTypes typeutil.Map
}
```

### rta struct (Enhanced)
```go
type rta struct {
    result  *Result
    prog    *ssa.Program

    // Context tracking for precision
    currentFunction *ssa.Function

    worklist []*ssa.Function

    // Cross-product tables
    addrTakenFuncsBySig typeutil.Map
    dynCallSites        typeutil.Map
    invokeSites         typeutil.Map

    // Type information with fingerprints
    concreteTypes   typeutil.Map  // *concreteTypeInfo
    interfaceTypes  typeutil.Map  // *interfaceTypeInfo

    // Pre-computed implementation relationships
    interfaceToTypes map[*types.Interface][]types.Type
    typeToInterfaces map[types.Type][]*types.Interface

    // User type index for efficient scanning
    userTypesIndexBuilt bool
    userTypes           []types.Type
}
```

## Modification 1: Pattern-Based Reflection Handling

### Known Safe Functions Map
```go
var knownSafeFunctions = map[string][]string{
    // JSON encoding/decoding
    "encoding/json.Marshal":           {"MarshalJSON", "MarshalText"},
    "encoding/json.Unmarshal":         {"UnmarshalJSON", "UnmarshalText"},
    "(*encoding/json.Encoder).Encode": {"MarshalJSON", "MarshalText"},

    // fmt package
    "fmt.Printf":  {"String", "GoString", "Error", "Format"},
    "fmt.Sprintf": {"String", "GoString", "Error", "Format"},
    "fmt.Errorf":  {"String", "GoString", "Error", "Format"},

    // XML, YAML, binary encoding, SQL
    // ... additional mapped functions
}
```

### How It Works
When analyzing a call like:
```go
json.Marshal(user)
```

When the current function contains a statically resolved call to a mapped function, a newly recorded type's empty-interface conversion retains the union of mapped method names and the common-method fallback: `MarshalJSON`, `UnmarshalJSON`, `MarshalText`, `UnmarshalText`, `String`, `GoString`, `Error`, and `Format`. The mapping is a function-context heuristic: it does not trace which conversion is an argument to the listed call. The map includes both JSON versions and JSON v2's streaming and text hooks.

A direct SSA use of an empty-interface conversion by `reflect.ValueOf` or `reflect.TypeOf` instead retains that type's exported methods, even if a mapped call appears in the same function or the type was recorded earlier. This exception does not follow values through stores, joins, or wrapper functions. Methods with their own type parameters are excluded from reflection retention because reflection cannot instantiate them.

## Modification 2: Precise Non-Empty Interface Conversions

### Problem Solved
```go
var w io.Writer = &bytes.Buffer{}
```

**This fork:**
- Marks only the methods required by the `io.Writer` interface (`Write`)
- Leaves other methods unmarked unless another reachability rule applies

### Implementation

`addRuntimeTypeForInterface(T types.Type, iface *types.Interface, skip bool)` records the runtime type and marks only methods required by `iface`. Its lookup uses the interface method's package, so unexported marker methods are included when they satisfy the interface.

## Modification 3: Context-Aware Analysis

### Current Function Tracking
```go
type rta struct {
    currentFunction *ssa.Function  // Track what's calling
    // ...
}

func (r *rta) visitFunc(f *ssa.Function) {
    r.currentFunction = f  // Set context
    // Analyze instructions with context available
}
```

### Usage
Enables detecting patterns like:
```go
// In current function
json.Marshal(x)  // Context: we're in a JSON context
```

The analyzer can make smarter decisions based on what function is doing the calling.

## Modification 4: Enhanced Interface Compliance (*Interface → any)

### Problem: errors.As Pattern
```go
type Validator interface {
    error
    isValidator() // Marker method
}

func check(err error) {
    var v Validator
    if errors.As(err, &v) { // Passes *Validator to any
        // Uses validator
    }
}
```

### Solution

`handleMakeInterface` recognizes a pointer to an interface converted to an empty interface and calls `markImplementorsMethodsReachable`. That helper scans program packages for implementors and retains interface-required methods, including unexported markers, whether or not the type is already recorded. This remains conservative rather than flow-sensitive.

`errors.AsType` has no caller-side pointer conversion. Its instantiated target assertion receives the same marker treatment. Internal `As` and `Unwrap` hook invokes are restricted to recorded runtime types; interface conversions replay invoke sites even when the type was first recorded structurally.

## Modification 5: TypeAssert Instruction Support

### Handling Type Assertions
```go
if w, ok := x.(io.Writer); ok {
    w.Write(data)
}
```

The `TypeAssert` instruction must ensure concrete types have required interface methods marked.

### Implementation with Fingerprinting

`handleTypeAssert` handles assertions from a non-empty interface to a concrete type directly. For a non-empty asserted interface it consults compatible program types: user-code assertions may build the user-type index, while standard-library assertions are limited to recorded runtime types. Required methods are marked through `markInterfaceMethodsReachable`. The `errors.AsType` target assertion is the exception described above; its internal hook assertions do not trigger a program-wide scan.

## Modification 6: ChangeInterface Instruction Support

### Interface-to-Interface Conversions
```go
func widen(rc io.ReadCloser) io.Reader {
    return rc
}
```

SSA represents this as `ChangeInterface` instruction.

### Implementation

`handleChangeInterface` obtains the non-empty target interface, finds program types that implement it, and marks its required methods reachable. The implementation does not restrict this scan to values proven to flow through the particular conversion.

### Fingerprint Optimization

`implements(cinfo *concreteTypeInfo, iinfo *interfaceTypeInfo)` rejects a candidate when the interface fingerprint is not a subset of the concrete fingerprint. It then calls `types.Implements`; the fingerprint is only a fast rejection filter.

## Modification 7: Runtime.SetFinalizer Detection

### Problem
```go
func NewResource() *Resource {
    r := &Resource{}
    runtime.SetFinalizer(r, (*Resource).cleanup)
    return r
}

func (r *Resource) cleanup() {}  // Called by GC
```

The `cleanup` method is reachable but not through normal call paths.

### Solution

`checkSetFinalizer(call *ssa.CallCommon)` recognizes a direct `runtime.SetFinalizer` call. It unwraps an optional `MakeInterface` around the second argument and marks either an `*ssa.Function` or an `*ssa.MakeClosure` function reachable and address-taken.

## Modification 8: Generic Template Tracking

### Problem: Generic Instantiations
```go
type Container[T any] struct { data T }
func (c *Container[T]) Add(item T) {}

func main() {
    c := &Container[int]{}  // Instantiation
    c.Add(42)              // Calls Container[int].Add
}
```

SSA creates `Container[int].Add` as separate function, but we want to track `Container[T].Add` template.

### Solution

`addReachable` calls `markGenericTemplateReachable` when it first discovers a function. When `f.Origin()` is a distinct template, the template is inserted into `Result.Reachable` without being added to the worklist. `Result.ReachableObjects` separately records generic methods that have no SSA function.

### Automatic Tracking

In normal mode, the SSA integration retains public generic functions and methods even without a local instantiation. Typed AST references preserve calls and function/method values; referenced templates are scanned recursively, while concrete helpers become RTA roots so their bodies receive ordinary dispatch analysis. Direct concrete-to-interface boundaries in template calls, assignments, and returns retain the required methods. Bodyless declarations are retained without traversal.

This fallback is not a replacement for instantiated SSA or value-flow analysis. See [generic API limitations](known-limitations.md#uninstantiated-generic-apis).

## Cross-Product Tabulation (Unchanged Core Logic)

The core cross-product tabulation is:

### Address-Taken Functions × Dynamic Calls
- `addrTakenFuncsBySig`: Groups functions by signature
- `dynCallSites`: Groups call sites by signature
- Cross-product computed incrementally

### Runtime Types × Interface Invokes
- `concreteTypes`: Concrete type information with fingerprints
- `interfaceTypes`: Interface type information with fingerprints
- `invokeSites`: Call sites grouped by interface
- Cross-product with optimization

## Performance Characteristics

The implementation uses a reusable worklist buffer, cached method sets, fingerprints before `types.Implements`, and lazily built type indexes. Their effect is workload-dependent; this repository provides no supported timing, allocation, coverage, or validation baseline. See [the performance guide](../performance.md) for the profiling workflow.

## Practical Usage

### Basic Usage
```go
func inspectReachability(roots []*ssa.Function, someFunction *ssa.Function, someMethod types.Object) {
    result := rta.Analyze(roots)
    if result == nil {
        return
    }

    if _, ok := result.Reachable[someFunction]; ok {
        fmt.Println("Function is reachable")
    }
    if result.ReachableObjects[someMethod] {
        fmt.Println("Generic template method is reachable")
    }
}
```

### Enhanced Features
```go
func inspectResult(result *rta.Result, fn *ssa.Function) {
    // Check if address-taken.
    if info, ok := result.Reachable[fn]; ok && info.AddrTaken {
        fmt.Println("Function's address is taken")
    }

    // Check generic templates.
    if obj := fn.Object(); obj != nil && result.ReachableObjects[obj] {
        fmt.Println("Template is reachable")
    }
}
```

## Integration with unusedfunc

### Analysis Pipeline

`pkg/ssa.Analyzer.findReachableMethods` excludes uninstantiated templates from RTA roots, adds concrete helpers discovered from public template bodies, then calls `rta.Analyze`. It combines the template objects with `Reachable` and `ReachableObjects`.

## Limitations

### Not Handled
1. **Template method calls** - Template text is not parsed.
2. **`reflect.Value.MethodByName`** - Method-name strings are not traced.
3. **Name-based registries and struct tags** - Method-name strings resolved reflectively do not establish static method edges. Actual function values stored in maps or registries can be matched to dynamic calls.

### Workarounds
Use suppression comments for known limitations:
```go
//nolint:unusedfunc // used in template.gotmpl:42
func (t *Type) TemplateMethod() {}

//nolint:unusedfunc // called via reflect.MethodByName
func (t *Type) ReflectionMethod() {}
```

## Summary

The fork adds reflection-context heuristics, narrower interface materialization, interface conversion handling, finalizer recognition, and generic template tracking to RTA's fixed-point reachability. Dynamic uses outside the supplied program or the documented heuristics still require a suppression.
