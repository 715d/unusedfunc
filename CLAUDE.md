# `unusedfunc` Contributor Guide

## Purpose

`unusedfunc` finds unreachable Go functions and methods.

| Declaration | Normal mode | Strict mode |
|---|---|---|
| Unexported | Report when unused | Report when unused |
| Exported in `/internal` or `main` | Report when unused | Report when unused |
| Other exported declaration | Keep as public API | Report when unused |

The analysis uses modified RTA. Read [docs/reference/rta-algorithm.md](docs/reference/rta-algorithm.md) before you change reachability rules.

## Key Paths

| Path | Purpose |
|---|---|
| `cmd/unusedfunc/` | CLI, output, and exit status |
| `pkg/unusedfunc/` | Package loading and analysis coordination |
| `pkg/ssa/` | SSA construction and root selection |
| `internal/rta/` | Modified RTA |
| `internal/analysis/` | Function metadata and reporting policy |
| `pkg/assembly/`, `pkg/runtime/`, `pkg/suppress/` | Dynamic-use metadata |
| `internal/harness/`, `testdata/` | Fixture test harness and cases |

## Commands

```bash
make build                 # build/unusedfunc
make test                  # race-enabled tests and coverage
make lint                  # golangci-lint
make fmt                   # goimports
make check                 # fmt and lint
```

`make test` runs the fixture harness. It clones declared `realworld-*` fixtures; use `go test -short ./...` to skip them. See [docs/workflows.md](docs/workflows.md) for fixture, debugging, and release steps.

## Analysis Rules

The analyzer loads test variants, builds SSA with generic instantiation, selects roots from target packages, and runs RTA. Roots include `main`, `init`, test, benchmark, and example functions. In normal mode, public exports in non-`main`, non-`internal` packages are roots. Strict mode omits these roots.

Runtime directives, CGo exports, selected assembly uses, and direct `runtime.SetFinalizer` callbacks have special handling. Suppression comments affect reporting only; they do not create roots.

Reflection and template uses are incomplete by design. `--skip-generated` does not remove generated functions from analysis or reporting. See [docs/reference/known-limitations.md](docs/reference/known-limitations.md) before you change these boundaries.

## Change Rules

- Add a focused `testdata/<case>/expected.yaml` fixture when analysis behavior changes.
- Use `log/slog` for diagnostics in non-test code. Write command results to stdout.
- Run `make lint && make test` before you finish.
- Read [docs/architecture.md](docs/architecture.md) before you change roots or reporting policy. Read [docs/performance.md](docs/performance.md) before you optimize analysis.
