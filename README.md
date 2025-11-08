# `unusedfunc`

[![CI](https://github.com/715d/unusedfunc/workflows/CI/badge.svg)](https://github.com/715d/unusedfunc/actions)
[![codecov](https://codecov.io/gh/715d/unusedfunc/branch/main/graph/badge.svg)](https://codecov.io/gh/715d/unusedfunc)
[![Go Report Card](https://goreportcard.com/badge/github.com/715d/unusedfunc)](https://goreportcard.com/report/github.com/715d/unusedfunc)

A Go linter that identifies unused functions and methods with precise rules:
- **Unexported functions/methods**: Report if not used anywhere
- **Exported functions/methods**: Report if unused AND within `/internal` packages

## Quick Start

```bash
go install github.com/715d/unusedfunc/cmd/unusedfunc@latest
unusedfunc ./...
```

**First run?** You'll likely see reports for:
- Unexported helper functions that are no longer called
- Exported functions in `/internal` packages that aren't used internally

See [Handling False Positives](#handling-false-positives) if you encounter reflection-based code.

## Key Feature: Internal Package Enforcement

Go's `/internal` package convention is a compiler-enforced boundary, not a suggestion. `unusedfunc` enforces it at the function level. Exported functions in `/internal` packages are treated as implementation details, not public API.

**Example:**
```go
// pkg/utils/helpers.go
package utils

func PublicHelper() {} // NOT reported (exported in public package - may be used externally)
func privateHelper() {} // REPORTED if unused (unexported)

// internal/utils/helpers.go
package utils

func PublicHelper() {} // REPORTED if unused (exported but in /internal - must be used internally)
func privateHelper() {} // REPORTED if unused (unexported)
```

This makes `unusedfunc` valuable for projects that use `/internal` packages to organize implementation details while allowing cross-package access within the module.

## When to Use This Tool

Use `unusedfunc` if:
- Your project follows Go's `/internal` package convention
- You want to ensure exported functions in internal packages are actually used
- You need accurate analysis with reasonable performance
- You want a dedicated tool for unused code detection without broader static analysis overhead

## Installation

```bash
go install github.com/715d/unusedfunc/cmd/unusedfunc@latest
```

**Other options:** [Binary releases](https://github.com/715d/unusedfunc/releases) | [Nix](docs/nix.md) | [From source](CONTRIBUTING.md#building)

## Usage

```bash
# Analyze current module
unusedfunc

# Analyze specific packages
unusedfunc ./pkg/...

# Verbose mode: adds statistics and debug logging to stderr
unusedfunc -v ./...

# JSON output (verbose adds 'stats' field to JSON structure)
unusedfunc -json -v ./...

# Include generated files in analysis
unusedfunc --skip-generated=false ./...
```

## FAQ

### Why is my exported function being reported?

If you see reports for exported functions, check if they're in an `/internal` package. Go's `/internal` convention means these functions are **not** public API — they're only accessible within your module. If they're unused internally, they should be removed or made unexported.

**Not in `/internal`?** Exported functions in public packages are never reported, as they may be used by external code.

### How does this compare to staticcheck's U1000?

| Feature | unusedfunc | staticcheck U1000 |
|---------|------------|-------------------|
| **Architecture** | SSA-based analysis with RTA algorithm | AST-based analysis |
| **Exported Functions** | **Reports unused exports in `/internal` packages** | Never reports unused exports |
| **Performance** | Whole-program analysis (scales with codebase) | File-level analysis (consistent overhead) |
| **Philosophy** | Opinionated: enforces `/internal` package conventions | Conservative: avoids false positives |
| **Suppression** | `//nolint:unusedfunc` or `//lint:ignore unusedfunc` | `//lint:ignore U1000 <reason>` |

**When to use `unusedfunc`:**
- Your codebase uses `/internal` packages to organize implementation details
- You want to enforce that internal exports are actually used
- You need precise call graph analysis for interface/generic code

**When to use `staticcheck`:**
- You want comprehensive static analysis beyond just unused functions
- You need maximum precision across all code quality checks
- Your codebase doesn't follow the `/internal` convention
- You prefer a battle-tested, widely-adopted tool suite

### Why isn't this a golangci-lint plugin?

`unusedfunc` requires whole-program SSA analysis to build accurate call graphs across your entire codebase. This architectural choice enables precise detection of unused exports in `/internal` packages, but requires different resource constraints than golangci-lint's file-level analyzers.

**Run it separately:** Add `unusedfunc` as a dedicated CI step alongside golangci-lint, similar to how you'd run benchmarks or integration tests.

## Handling False Positives

**Use suppression comments** for code called via reflection or templates:

```go
//nolint:unusedfunc
func CalledViaReflection() {}

//lint:ignore unusedfunc Called in template.gotmpl:15
func (t *TemplateContext) Export() string {
    return t.data
}
```

**Common patterns requiring suppression:**
- Methods called via `reflect.MethodByName("MethodName")`
- Template method calls (`.tmpl`, `.gotmpl`, `.html` files)
- Methods discovered by test frameworks
- Protobuf-generated code

**Generated code is skipped by default.** Use `--skip-generated=false` to analyze everything.

**Full reference:** [docs/reference/known-limitations.md](docs/reference/known-limitations.md) — reflection patterns, template limitations, workarounds, and examples.

## How It Works

`unusedfunc` uses SSA (Static Single Assignment) analysis to build a complete call graph of your codebase, then traces reachability from entry points (main, init, tests, exported functions).

**Why SSA?** Unlike AST-based tools, SSA analysis can accurately track:
- Interface method calls (which concrete type implements the interface?)
- Generic function instantiations (which type parameters are used?)
- Function values passed as arguments

This precision is why `unusedfunc` can confidently report unused exports in `/internal` packages.

**Technical details:** [Architecture docs](docs/architecture.md) | [RTA algorithm](docs/reference/rta-algorithm.md)

## Contributing

Contributions are welcome! Before submitting changes, please run `make test && make lint` and review our test cases in `testdata/`.

For detailed guidance on development setup, code standards, and release process, please see our [contributing guide](CONTRIBUTING.md).
