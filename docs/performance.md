# Performance Guide

## Executive Summary

The unusedfunc tool achieves its target performance of <2 minutes for 100K LOC codebases through careful architectural design and strategic optimizations. Current analysis shows **SSA construction dominates 60-80% of execution time**, with significant opportunities for further parallel processing improvements.

## Performance Benchmarks and Metrics

### Current Performance Profile
| Project Size | Target Time | Actual Time |
|-------------|-------------|-------------|
| Small (<1K LOC) | <1s | <1s ✓ |
| Medium (10K LOC) | <30s | 10-30s ✓ |
| Large (100K LOC) | <2min | 1-2min ✓ |

### Key Performance Metrics
- **Performance target**: <2min for 100K LOC achieved
- **SSA construction**: Consumes 60-80% of CPU time
- **Memory scaling**: O(LOC) for SSA construction
- **CPU complexity**: O(V+E) where V=functions, E=calls

## Performance Hotspots Analysis

### CPU Profile Analysis (Critical Bottlenecks)
1. **scanForInterfaceUsage**: 0.85s (14.71% cum) - Sequential processing
2. **AnalyzeMethods**: 1.08s (18.69% cum) - Main orchestration
3. **buildSSAProgram.func1.2**: 0.50s (8.65% cum) - Already parallelized
4. **Runtime overhead**: 42.73% from memory management (madvise)

### Memory Profile Analysis  
- **NewScope**: 30.67% of allocations
- **Scope.insert**: 27.40% of allocations
- **Type processing**: Dominates memory usage

### Performance-Critical Code Paths

#### 1. SSA Analysis (`pkg/ssa/analyzer.go`)
- Entry: `Analyzer.AnalyzeMethods()` orchestrates analysis
- RTA for reachability - handles reflection, generics, interfaces
- Interface resolution with fixed-point iteration (2-3 passes)
- Direct call indexing with buffer reuse

#### 2. Package Loading
- Config: Comprehensive load mode with full type info
- Function: `loadPackages()` loads AST, types, dependencies
- Impact: Front-loads all analysis data

#### 3. Interface Resolution Loop (lines 297-352)
- Runs to fixed-point (typically 2-3, max 20 iterations)
- Builds bidirectional mappings
- Memory intensive for large codebases

## Memory Optimization Strategies

### Critical Memory Bottlenecks

1. **Package Loading Pattern**
   - Issue: Comprehensive `packages.DefaultLoadMode` loads ALL type information simultaneously
   - Impact: O(packages × functions) memory complexity
   - Root cause: Creates complete type scopes for every package at once

2. **SSA Construction Strategy**
   - Issue: `ssautil.AllPackages()` builds SSA for ALL packages upfront
   - Impact: O(packages × functions) memory allocation pattern

3. **Scope Iteration Anti-Pattern**
   - Issue: Triple scope iteration: `pkg.Types.Scope()` → `scope.Names()` → `scope.Lookup(name)` 
   - Impact: Repeated type object creation without reuse

### High-Impact Optimizations

#### Streaming Package Processing
```go
// Instead of loading all packages at once:
pkgs, err := packages.Load(pkgConfig, cfg.Packages...)

// Use batched processing:
for batch := range batchPackages(cfg.Packages, 10) {
    processBatch(batch)
    runtime.GC() // Release memory between batches
}
```

#### Pre-allocation with Capacity Hints
```go
finalizerFunctions := make(map[string]bool, 16)
concreteToInterfaces := make(map[string]Set[string], len(funcs)*2)
interfaceToConcretes := make(map[string]Set[string], len(funcs))
```

#### Memory-Efficient Set Pattern
```go
// Use empty struct for memory-efficient sets
type Set[T comparable] map[T]struct{}
```

#### Helper Function Pattern for Reduced Allocations
```go
// Avoid repeated inline function definitions
func newStringSet() (Set[string], bool) {
    return make(Set[string], 16), false
}

// Use with LoadOrCompute
interfaces, _ := sa.concreteToInterfaces.LoadOrCompute(concreteName, newStringSet)
```

## Concurrency Patterns and Parallel Processing

### Existing Concurrency
1. **SSA Package Building** (`pkg/ssa/analyzer.go:108-123`)
   - Uses goroutines with WaitGroup
   - Each package built independently
   - Includes panic recovery

### High-Impact Parallelization Opportunities

#### 1. scanForInterfaceUsage Parallelization (50-70% speedup)
- **Location**: pkg/ssa/ssa_analysis.go:954-1098
- **Problem**: Sequential processing of independent SSA functions
- **Solution**: Worker pool with concurrent maps (xsync.Map)
- **Risk**: Medium - requires careful map synchronization

#### 2. Method Marking Parallelization (40-60% speedup)
- **Location**: pkg/ssa/ssa_analysis.go:353-376
- **Problem**: Sequential marking despite independence
- **Solution**: errgroup.Group with NumCPU limit
- **Risk**: Low - embarrassingly parallel operations

#### 3. Package-Level Function Collection (Linear speedup)
- **Current**: Sequential iteration over packages
- **Opportunity**: Worker pool for packages
- **Implementation**: Each package's function collection is independent

### Worker Pool Template
```go
type workItem struct {
    pkg *packages.Package
    // or function, file, etc.
}

workers := runtime.NumCPU()
work := make(chan workItem, len(items))
results := make(chan resultType, len(items))

var wg sync.WaitGroup
for i := 0; i < workers; i++ {
    wg.Add(1)
    go worker(work, results, &wg)
}

// Send work
for _, item := range items {
    work <- item
}
close(work)

// Collect results
go func() {
    wg.Wait()
    close(results)
}()
```

### Synchronization Patterns
- **xsync.Map**: For high-contention concurrent maps
- **sync.OnceValue**: For lazy initialization  
- **errgroup**: For concurrent operations with error handling
- **Atomic operations**: For simple counters and flags

## SSA Analyzer Optimizations

### Concurrent Data Structure Migration Pattern
```go
// Before: map + mutex
concreteToInterfaces map[string]Set[string]
mu sync.RWMutex

// After: xsync.Map
concreteToInterfaces *xsync.Map[string, Set[string]]
```
- **Benefits**: 50-70% reduction in contention overhead
- **Results**: Eliminated mutex contention for map access

### Interface Compliance Optimization
- **Pattern**: Track actual interface usage instead of checking all possible implementations
- **Key insight**: Only check interfaces that concrete types are actually converted to
- **Results**:
  - 50-60% CPU reduction in interface checking
  - 98% reduction in `types.Implements` calls
  - More precise unused method detection

## Profiling and Bottleneck Analysis

### Profiling Setup
```bash
# Enable profiling
./build/unusedfunc --profile ./...

# Output files
cpu.prof  # CPU profile
mem.prof  # Heap profile
```

### Analysis Commands
```bash
# CPU analysis
go tool pprof cpu.prof

# Memory analysis  
go tool pprof mem.prof

# Benchmark execution
go test -bench=. ./pkg/...
```

### Existing Benchmarks
- `BenchmarkAnalyzer_Analyze`: Full analysis workflow
- `BenchmarkComputeTypeName`: Type name computation with/without cache
- `BenchmarkSuppressionChecker_Load/IsSuppressed`: Suppression checking

## Performance Improvements Implemented

### 1. xsync.Map Migration (Completed)
- **Impact**: Eliminated mutex contention for map access
- **Files**: `pkg/ssa/ssa_analysis.go`
- **Pattern**: LoadOrCompute for atomic operations

### 2. Interface Compliance Optimization (Completed)
- **Impact**: 50-60% CPU reduction
- **Precision**: More accurate unused method detection
- **Trade-off**: Slight memory overhead for tracking maps

### 3. Pointer Type Creation (Completed)  
- **Impact**: 5-10% improvement, 60MB allocation reduction
- **Pattern**: Move invariant operations outside loops

## Implementation Roadmap

### Phase 1: Pipeline Parallelization (2-4 weeks)
- **Target**: 25-40% performance gain
- Concurrent Assembly & Suppression Processing
- Package Loading Optimization

### Phase 2: SSA Construction Enhancement (4-6 weeks)
- **Target**: 40-60% performance gain  
- Smart Resource Management
- Dependency-Aware Building

### Phase 3: Advanced Optimizations (6-8 weeks)
- **Target**: 20-30% additional gain
- Streaming Analysis
- Incremental Analysis with caching

### Combined Impact Projection: 60-80% overall improvement

## Safety Constraints and Validation

### Safety Requirements
- **Deterministic results**: Identical output regardless of execution order
- **No race conditions**: Use proven concurrent data structures
- **Correctness**: Maintain exact semantics
- **Memory safety**: Preserve nil checks and error handling

### Validation Strategy
1. **Benchmark before/after**: Measure on large codebases (kubernetes, docker)
2. **Correctness verification**: Ensure analysis results remain identical  
3. **Performance monitoring**: Track specific metrics (NewScope/Scope.insert)
4. **Stress testing**: Verify graceful handling of memory-constrained environments

## Quick Reference

### Symbol Mappings
- **Performance hotspots** → `Analyzer.AnalyzeMethods` in `pkg/ssa/analyzer.go`
- **xsync implementation** → `concreteToInterfaces` field in `Analyzer` struct  
- **Worker pool example** → SSA package building at `pkg/ssa/analyzer.go:108-123`
- **Profiling setup** → `--profile` flag handling in `cmd/unusedfunc/main.go`

### Common Performance Tasks
- **Add profiling** → Implement `--profile` flag, defer pprof.StopCPUProfile()
- **Fix slow analysis** → Check `scanForInterfaceUsage`, consider parallelization
- **Reduce memory** → Use pre-allocation with capacity hints, implement batching
- **Debug contention** → Replace mutex+map with xsync.Map pattern

### Performance Query Tags
```
#performance #optimization #concurrency #benchmarking
#profiling #memory #scalability #parallelization
```