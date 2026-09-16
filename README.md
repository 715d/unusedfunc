# `unusedfunc`

[![CI](https://github.com/715d/unusedfunc/workflows/CI/badge.svg)](https://github.com/715d/unusedfunc/actions)
[![codecov](https://codecov.io/gh/715d/unusedfunc/branch/main/graph/badge.svg)](https://codecov.io/gh/715d/unusedfunc)
[![Go Report Card](https://goreportcard.com/badge/github.com/715d/unusedfunc)](https://goreportcard.com/report/github.com/715d/unusedfunc)

`unusedfunc` finds unreachable Go functions and methods with SSA-based reachability analysis.

| Declaration | Normal mode | `--strict` |
|---|---|---|
| Unexported | Report when unused | Report when unused |
| Exported in `/internal` or `main` | Report when unused | Report when unused |
| Other exported declaration | Keep as public API | Report when unused |

Use strict mode only when no external code imports the selected packages.

## Install

```bash
go install github.com/715d/unusedfunc/cmd/unusedfunc@latest
```

Building from source requires the Go version in [`go.mod`](go.mod). See [CONTRIBUTING.md](CONTRIBUTING.md#building) or [Nix usage](docs/nix.md).

## Use

Run the command from the module that you want to inspect:

```bash
unusedfunc ./...
unusedfunc ./internal/...
unusedfunc --strict ./...
unusedfunc --build-tags integration ./...
unusedfunc --json ./... > unused.json
unusedfunc -v ./...
```

Tests are always loaded, so test-only calls count as uses. Results apply only to the selected packages and build configuration. Exit status `0` means no findings, `1` means findings, and `2` means an error.

`--profile` writes `cpu.prof` and `mem.prof` to the current directory. See [performance](docs/performance.md).

## Dynamic Uses

The analyzer cannot prove every runtime call. Use a suppression for a confirmed template, reflection, assembly, or build-specific use:

```go
//nolint:unusedfunc // used in template.gotmpl:15
func (t *TemplateContext) Export() string { return t.data }
```

The suppression must be on the declaration line or the line before it. `//lint:ignore unusedfunc reason` also works. See [known limitations](docs/reference/known-limitations.md) for supported dynamic cases and workarounds.

## How It Works

The tool loads the selected packages, builds SSA, selects roots, and runs modified Rapid Type Analysis (RTA). It tracks direct calls, interface dispatch, generic origins, direct `runtime.SetFinalizer` callbacks, selected reflection patterns, runtime directives, and package-local assembly calls. It does not retain a call graph.

See [architecture](docs/architecture.md) for analysis boundaries and [the RTA reference](docs/reference/rta-algorithm.md) for the exact algorithm.

## Contribute

See [CONTRIBUTING.md](CONTRIBUTING.md). Run `make test && make lint` before you submit a change.
