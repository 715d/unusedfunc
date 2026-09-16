# Performance Guide

## Executive Summary

`unusedfunc` loads full syntax and type information, builds SSA for loaded packages and dependencies, then runs RTA from target-package roots. Cost depends on the package graph, type information, generic instantiations, interface use, and selected build configuration. This repository has no supported size-based timing, memory, or benchmark baseline.

## Performance Hotspots Analysis

### Performance-Critical Code Paths

#### 1. Package Loading (`pkg/unusedfunc/loader.go`)
- `LoadPackages` requests dependencies, files, imports, types, syntax, and `TypesInfo`, and enables test variants.
- This front-loads the analysis input before SSA construction.

#### 2. SSA Analysis (`pkg/ssa/analyzer.go`)
- `Analyzer.buildSSAProgram` calls `ssautil.AllPackages` with `ssa.InstantiateGenerics | ssa.BareInits`, then waits for `Program.Build`, which builds packages concurrently.
- RTA reaches a fixed point over the resulting program.

#### 3. Function Collection and Assembly (`pkg/unusedfunc/analyzer.go`)
- `Analyzer.collectFunctions` processes packages concurrently, limited to `runtime.NumCPU()` workers.
- Assembly scanning is sequential.

These are implementation facts, not measurements. Profile a representative workload before treating a path as a bottleneck.

## Memory and Concurrency Constraints

SSA construction retains the loaded program, so reducing the input package set or selecting the appropriate build configuration can matter more than a local allocation change. Both SSA package building and function collection use concurrency; the collection worker limit does not control SSA's internal parallelism. Do not introduce batching, manual garbage collection, or additional parallelism based on this document alone; establish a profile and preserve deterministic results first.

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
# CPU time, inclusive call paths, and heap allocation/retention views
go tool pprof -top cpu.prof
go tool pprof -top -cum cpu.prof
go tool pprof -alloc_space -top mem.prof
go tool pprof -inuse_space -top mem.prof
```

## Safety Constraints and Validation

### Safety Requirements
- **Deterministic results**: Identical output regardless of execution order
- **No race conditions**: Use proven concurrent data structures
- **Correctness**: Maintain exact semantics
- **Memory safety**: Preserve nil checks and error handling

### Validation Strategy
1. Capture before/after profiles for the same package pattern, build tags, working directory, Go version, and source tree.
2. Keep CPU and heap observations separate: `-alloc_space` measures allocation volume, while `-inuse_space` measures retained heap.
3. Confirm that analyzer output is unchanged under the same workload.
4. Record the command, machine or CPU limit, elapsed time, and relevant `pprof` top output so the finding is reproducible.

## Quick Reference

- **Profiling setup**: `--profile` in `cmd/unusedfunc/main.go` writes `cpu.prof` and `mem.prof` in the current directory.
- **Function collection**: `pkg/unusedfunc.Analyzer.collectFunctions` is capped at `runtime.NumCPU()` workers.
- **SSA construction**: `pkg/ssa.Analyzer.buildSSAProgram` waits for concurrent package builds before analysis.