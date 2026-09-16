# `unusedfunc`

[![CI](https://github.com/715d/unusedfunc/workflows/CI/badge.svg)](https://github.com/715d/unusedfunc/actions)
[![codecov](https://codecov.io/gh/715d/unusedfunc/branch/main/graph/badge.svg)](https://codecov.io/gh/715d/unusedfunc)
[![Go Report Card](https://goreportcard.com/badge/github.com/715d/unusedfunc)](https://goreportcard.com/report/github.com/715d/unusedfunc)

A Go linter that identifies unused functions and methods with precise rules:
- **Unexported functions/methods**: Report if not used anywhere
- **Exported functions/methods**: Report if unused in `/internal` or `main` packages
- **Strict mode**: Report ALL unused exported functions (use when packages aren't imported externally)

## Quick Start

```bash
go install github.com/715d/unusedfunc/cmd/unusedfunc@latest
unusedfunc ./...
```

**First run?** You'll likely see reports for:
- Unexported helper functions that are no longer called
- Exported functions in `/internal` or `main` packages that aren't reachable

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

Building from source requires the Go version declared in [`go.mod`](go.mod) or a toolchain that can download it automatically.

**Other options:** [Binary releases](https://github.com/715d/unusedfunc/releases) | [Nix](docs/nix.md) | [From source](CONTRIBUTING.md#building)

## Usage

```bash
# Analyze current module
unusedfunc

# Analyze specific packages
unusedfunc ./pkg/...

# Verbose mode: adds statistics and debug logging to stderr
unusedfunc -v ./...

# JSON output (always includes statistics)
unusedfunc --json ./...

# Strict mode: report ALL unused exported functions (not just /internal)
unusedfunc --strict ./...

# Analyze a build-tag configuration
unusedfunc --build-tags integration ./...
```

Test files are always included, so test-only uses count as reachable. Run from the module you want to analyze; results cover the selected packages and build configuration, not every platform or caller outside that selection.

**Exit codes:** `0` for no findings, `1` for unused functions, `2` for errors. A successful JSON report contains `unused_functions`, `stats`, `version`, and `timestamp`; verbose logging goes to stderr.

## FAQ

### Why is my exported function being reported?

In normal mode, unused exports are reported in `/internal` and `main` packages. Internal packages are importable only within the tree rooted at the parent of `internal`; `main` packages are not importable as libraries. Analyze all relevant callers before removing a finding.

**In another package?** Exported functions in non-`main`, non-`internal` packages are never reported in normal mode, as they may be used by external code. Use `--strict` mode if you're certain your packages aren't imported externally.

### When should I use `--strict` mode?

Use `--strict` mode when:
- You're working on an **application** (not a library) where packages aren't imported externally
- You want to find unused exported functions across your entire codebase
- You're certain no external code imports your public packages

**Example scenario:** You have a web application where all packages are internal to the project. In strict mode, `unusedfunc` will report ALL unused exported functions, helping you clean up dead code that normal mode would skip.

**Warning:** Don't use `--strict` on libraries or modules that external code might import. It will report all unused exports as false positives.

### How does this compare to staticcheck's U1000?

| Feature | unusedfunc | unusedfunc --strict | staticcheck U1000 |
|---------|------------|---------------------|-------------------|
| **Exported Functions** | **Reports unused exports in `/internal` and `main` packages** | **Reports ALL unused exports** | Does not report unused exported package-level functions |
| **Use Case** | Libraries + apps following `/internal` convention | Applications with no external imports | General-purpose static analysis |
| **Philosophy** | Opinionated: enforces `/internal` package conventions | Aggressive: treats all exports as potentially unused | Conservative: avoids false positives |
| **Suppression** | `//nolint:unusedfunc` or `//lint:ignore unusedfunc` | `//nolint:unusedfunc` or `//lint:ignore unusedfunc` | `//lint:ignore U1000 <reason>` |

**When to use `unusedfunc`:**
- Your codebase uses `/internal` packages to organize implementation details
- You want to enforce that internal exports are actually used
- You need reachability analysis for interface/generic code

**When to use `staticcheck`:**
- You want comprehensive static analysis beyond just unused functions
- Your codebase doesn't follow the `/internal` convention
- You prefer a battle-tested, widely-adopted tool suite

### Why isn't this a golangci-lint plugin?

`unusedfunc` loads the selected packages and dependencies together, builds SSA, and computes reachability across them. It is provided as a standalone command, not a golangci-lint plugin.

**Run it separately:** Add `unusedfunc` as a dedicated CI step alongside golangci-lint, similar to how you'd run benchmarks or integration tests.

## Handling False Positives

**Use suppression comments** for code called via reflection or templates. Place the comment immediately before the declaration or on the same line; file-wide suppression is not supported:

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

**Generated-code limitation:** `--skip-generated` defaults to true, but currently only filters declarations used for runtime-directive detection. Generated functions still participate in analysis and can be reported; the flag is not a workaround for generated-code findings.

**Full reference:** [docs/reference/known-limitations.md](docs/reference/known-limitations.md) — reflection patterns, template limitations, workarounds, and examples.

## How It Works

`unusedfunc` builds SSA (Static Single Assignment) and computes reachability with modified RTA; it does not retain a call graph. Roots include main, init, test-name prefixes, selected runtime hooks, and public exports in normal mode. Internal and main-package exports are not automatically roots.

**Why SSA?** It represents:
- Interface method calls (which concrete type implements the interface?)
- Generic function instantiations (which type parameters are used?)
- Function values passed as arguments

Reflection heuristics and dynamic uses can still produce false positives or retain unused code; review findings before deleting functions.

**Technical details:** [Architecture docs](docs/architecture.md) | [RTA algorithm](docs/reference/rta-algorithm.md)

## Contributing

Contributions are welcome! Before submitting changes, please run `make test && make lint` and review our test cases in `testdata/`.

For detailed guidance on development setup, code standards, and release process, please see our [contributing guide](CONTRIBUTING.md).
