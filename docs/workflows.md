# Development Workflows

## Build and Check

```bash
make build       # build/unusedfunc
make test        # race-enabled tests and coverage
make lint        # golangci-lint
make fmt         # goimports
make check       # fmt and lint
```

The project declares development tools in `go.mod`. `make fmt` and `make lint` run them with `go tool`.

## Add a Fixture

Create `testdata/<case>/expected.yaml`. The harness discovers each immediate child directory with that file.

```yaml
build_configurations:
  - name: default
    expected_unused:
      - func: <module/package>.UnusedFunction
```

Add `build_tags`, `enable_cgo`, `goos`, `goarch`, `strict`, `expected_errors`, or `file` only when the case needs them. `make test` clones declared `realworld-*` fixtures. Use `go test -short ./...` when network access is unavailable.

## Debug a Finding

```bash
./build/unusedfunc -v ./...
./build/unusedfunc --build-tags integration ./...
```

Check the caller's root, the selected package pattern, and the build configuration. For templates, reflection, assembly, or build-specific uses, see [known limitations](reference/known-limitations.md). See [architecture](architecture.md) before you change root selection or reporting policy.

## Profile

```bash
./build/unusedfunc --profile ./...
go tool pprof -top cpu.prof
go tool pprof -inuse_space -top mem.prof
```

See [performance](performance.md) before you change concurrency or memory use.

## Release

```bash
make lint && make test
go mod tidy && go mod verify
go tool goreleaser check
```

Update documentation, then tag the release as `v*.*.*`.
