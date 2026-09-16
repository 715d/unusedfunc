# Development Workflows

## Build and Test

### Standard Build
```bash
make build        # Build to build/unusedfunc
make test         # Run all tests with race detector and coverage
make lint         # Run golangci-lint
```

### Formatter and Fixture Checks

Development tools are declared with `tool` directives in `go.mod`. `make fmt` and `make lint` use `go tool`, which downloads and builds them with the project's Go toolchain. Editor and agent hooks should invoke the same tools from the module directory:

```bash
go tool goimports -d path/to/file.go
go tool golangci-lint run ./...
go tool goreleaser check
```

Tool dependencies share the application's module graph. After tool or dependency updates, run the tools and project tests together; golangci-lint's [upstream source-build guidance](https://golangci-lint.run/docs/welcome/install/local/#install-from-sources) warns that shared dependency upgrades can produce untested combinations.

`testdata` contains intentionally unused declarations and unusual language patterns. Production lint rules exclude these analyzer inputs; `make test` validates their exact expected findings through the harness. Build-sensitive fixtures must be checked with their declared configuration.

### Custom Entry Points
```bash
# Analyze specific packages
./build/unusedfunc ./internal/...

# Test files are always loaded and test uses participate in analysis.
./build/unusedfunc ./...

# Analyze with build tags
./build/unusedfunc --build-tags=integration ./...

# Whole program analysis
./build/unusedfunc ./...
```

Package loading always uses test variants, and functions whose names begin with `Test`, `Benchmark`, or `Example` are roots.

## Performance Tuning

See @docs/performance.md for profiling guidance and implementation constraints.

### GOGC Tuning
```bash
# Baseline target. Actual GC behavior also depends on live heap, roots, and any memory limit.
./build/unusedfunc ./...

# Try a larger target and measure the result for the workload.
GOGC=400 ./build/unusedfunc ./...

# Monitor GC behavior
GODEBUG=gctrace=1 GOGC=400 ./build/unusedfunc ./... 2>&1 | grep gc
```

**Tradeoff**: A larger target can use more memory and reduce collection frequency; measure elapsed time and memory before adopting it.

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
- [ ] Inspect every reported function for a static or dynamic use
- [ ] Add a focused regression case for confirmed analyzer behavior
- [ ] Record the command, revision, build tags, and Go version
- [ ] Profile representative workloads before making performance claims

### Common False Positive Patterns
1. **Template methods**: Methods called from `.gotmpl` files
2. **Reflection patterns**: `MethodByName()` dynamic dispatch
3. **Build tag conditionals**: Functions used under a different configuration

## Debugging Reachability Issues

See @docs/architecture.md#entry-point-detection for roots and reporting rules.

### Quick Diagnosis
**Function not marked as used despite being called:**
1. Check if caller is reachable from entry points
2. Verify interface conversion is tracked (MakeInterface/ChangeInterface)
3. Check the loaded build tags and target configuration
4. Check for function value assignment or a dynamic call

**False positive (function IS used):**
1. Template usage: Add suppression comment (see @docs/reference/known-limitations.md)
2. Reflection usage: Verify pattern is supported or add a suppression
3. Assembly call: Check build-selected `.s` files for package-local direct `CALL` instructions
4. Test-only: Test files are loaded automatically; check that the test package loaded successfully

### Enable Debug Logging
```bash
./build/unusedfunc -v ./...

# Verbose logs identify loading and analysis progress.
```

## Profiling

### Basic Profiling
```bash
./build/unusedfunc --profile ./...

# CPU profile
go tool pprof -http=:8080 cpu.prof

# Memory profile
go tool pprof -http=:8080 -sample_index=alloc_space mem.prof
```

## Test Harness Usage

See @docs/architecture.md#testing-architecture for test harness structure.

### Adding Test Cases
Create a `testdata/<case>/expected.yaml` with one or more build configurations:
```yaml
build_configurations:
  - name: "default"
    build_tags: []
    enable_cgo: false
    expected_unused:
      - func: "<module/package>.<function>"
        reason: "unexported function not used"
        file: "main.go" # optional suffix
    expected_errors: []
```

The harness discovers immediate `testdata/*/expected.yaml` directories. It compares reported function names and optional file suffixes.

## Release Checklist

- [ ] All tests pass: `make test`
- [ ] Linters pass: `make lint`
- [ ] Dependencies are clean and verified: `go mod tidy && go mod verify`
- [ ] Documentation updated
- [ ] Version tagged as `v*.*.*`

## Common Commands

```bash
# Quick development cycle
make build && ./build/unusedfunc ./internal/...

# Full validation
make lint && make test

# Performance check
make build && GOGC=400 ./build/unusedfunc --profile ./...

# Real-world test
cd /tmp && git clone --depth=1 URL && cd repo && unusedfunc ./...
```
