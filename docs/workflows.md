# Development Workflows

## Build and Test

### Standard Build
```bash
make build        # Build to build/unusedfunc
make test         # Run all tests with race detector and coverage
make lint         # Run golangci-lint (47 linters)
```

### Custom Entry Points
```bash
# Analyze specific packages
./build/unusedfunc ./internal/...

# Include test files in analysis
./build/unusedfunc --include-tests ./...

# Whole program analysis
./build/unusedfunc ./...
```

## Performance Tuning

See @docs/architecture.md#performance-optimizations for detailed optimization strategies.

### GOGC Tuning
```bash
# Default (GOGC=100): GC at 2x heap size
./build/unusedfunc ./...

# Large codebases: GC at 5x heap size (fewer pauses)
GOGC=400 ./build/unusedfunc ./...

# Monitor GC behavior
GODEBUG=gctrace=1 GOGC=400 ./build/unusedfunc ./... 2>&1 | grep gc
```

**Tradeoff**: Higher memory usage for faster analysis.

## Real-World Validation

### Test Against Popular Projects
```bash
# Clone test projects
mkdir -p /tmp/validation
cd /tmp/validation
git clone --depth=1 https://github.com/go-chi/chi
git clone --depth=1 https://github.com/redis/go-redis
git clone --depth=1 https://github.com/uber-go/zap

# Run analysis
cd chi && unusedfunc ./... > chi-results.txt
cd ../go-redis && unusedfunc ./... > go-redis-results.txt
cd ../zap && unusedfunc ./... > zap-results.txt

# Check for false positives (verify each reported function)
```

### Validation Checklist
- [ ] Zero false positives on static code
- [ ] Known limitations documented (templates, reflection)
- [ ] Performance <2min per 100K LOC
- [ ] Handles interfaces correctly
- [ ] Handles generics correctly
- [ ] Test coverage >90%

### Common False Positive Patterns
1. **Template methods**: Methods called from `.gotmpl` files
2. **Reflection patterns**: `MethodByName()` dynamic dispatch
5. **Build tag conditionals**: Functions used under specific tags

## Debugging Reachability Issues

See @docs/architecture.md#interface-compliance-system for detailed debugging workflow.

### Quick Diagnosis
**Function not marked as used despite being called:**
1. Check if caller is reachable from entry points
2. Verify interface conversion is tracked (MakeInterface/ChangeInterface)
3. Ensure generic instantiation is tracked (`fn.Origin()`)
4. Check for function value assignment

**False positive (function IS used):**
1. Template usage: Add suppression comment (see @docs/reference/known-limitations.md)
2. Reflection usage: Verify pattern is in common patterns
3. Assembly call: Check `.s` file parsing
4. Test-only: Verify test files are included

### Enable Debug Logging
```bash
./build/unusedfunc -v ./...

# Check what entry points were detected
# Look for: main(), init(), Test*(), exported functions
```

## Profiling

### Basic Profiling
```bash
# CPU profile
go tool pprof -http=:8080 cpu.prof

# Memory profile
go tool pprof -http=:8080 -sample_index=alloc_space mem.prof
```

## Test Harness Usage

See @docs/architecture.md#testing-architecture for detailed test harness structure.

### Adding Test Cases
Create testdata directory with expected.yaml:
```yaml
name: "Test case name"
description: "What is being tested"
expected_unused:
  - func: "<module/package>.<function>"
    reason: "unexported function not used"
```

## Release Checklist

- [ ] All tests pass: `make test`
- [ ] Linters pass: `make lint`
- [ ] Benchmarks stable: `make benchmark`
- [ ] Real-world validation: 0 false positives on 3+ projects
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version tagged: `git tag v1.x.x`

## Common Commands

```bash
# Quick development cycle
make build && ./build/unusedfunc ./internal/...

# Full validation
make lint && make test

# Performance check
make benchmark && GOGC=400 time ./build/unusedfunc ./...

# Real-world test
cd /tmp && git clone --depth=1 URL && cd repo && unusedfunc ./...
```
