# Performance

`unusedfunc` loads syntax and type information, builds SSA for loaded packages and dependencies, then runs RTA from target-package roots. Cost depends on the package graph, generics, interface use, and build configuration. This project has no supported performance baseline.

## Profile

```bash
./build/unusedfunc --profile ./...
go tool pprof -top cpu.prof
go tool pprof -top -cum cpu.prof
go tool pprof -alloc_space -top mem.prof
go tool pprof -inuse_space -top mem.prof
```

`--profile` writes `cpu.prof` and `mem.prof` to the current directory.

## Change Rules

Measure the same package pattern, build tags, working directory, Go version, and source tree before and after a change. Record elapsed time, resource limits, and relevant profile output. Keep CPU allocation and retained-heap results separate. Confirm that analysis output stays unchanged.

Do not add batching, manual garbage collection, or concurrency from a guess. Start with a representative profile.
