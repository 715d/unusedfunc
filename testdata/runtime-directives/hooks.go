package main

import (
	"sync/atomic"
	"unsafe"
)

// Additional runtime-specific patterns

// Memory allocator hooks

//go:linkname mallocHook runtime.malloc
func mallocHook(size uintptr) unsafe.Pointer

//go:linkname freeHook runtime.free
func freeHook(ptr unsafe.Pointer)

// GC hooks and callbacks

type gcCallback struct {
	f    func()
	next *gcCallback
}

var gcCallbacks atomic.Value // *gcCallback

// Called by simulated runtime during GC
func runGCCallbacks() {
	if cbs := gcCallbacks.Load(); cbs != nil {
		for cb := cbs.(*gcCallback); cb != nil; cb = cb.next {
			cb.f()
		}
	}
}

// addGCCallback adds a GC callback
func addGCCallback(f func()) {
	new := &gcCallback{f: f}
	for {
		old := gcCallbacks.Load()
		if old != nil {
			new.next = old.(*gcCallback)
		}
		if gcCallbacks.CompareAndSwap(old, new) {
			break
		}
	}
}

// Functions registered as GC callbacks

func gcCallback1() {
	// Called during GC
}

func gcCallback2() {
	// Another GC callback
}

// Panic/recover hooks

//go:nosplit
func panicHook() {
	// Called during panic
}

//go:nosplit
func recoverHook() {
	// Called during recover
}

// Scheduler hooks

//go:norace
func scheduleHook() {
	// Called by scheduler
}

//go:nosplit
func preemptHook() {
	// Called during preemption
}

// Signal handlers

//go:nosplit
//go:noinline
func sighandler(sig uint32, info *struct{}, ctx unsafe.Pointer) {
	// Signal handler called by runtime
}

// CPU profiling hooks

var cpuProfiler struct {
	on bool
	fn func([]uintptr)
}

//go:noinline
func cpuProfilerHook(stk []uintptr) {
	if cpuProfiler.on && cpuProfiler.fn != nil {
		cpuProfiler.fn(stk)
	}
}

// profileCallback is registered for profiling
func profileCallback(stk []uintptr) {
	// Process stack trace
}

// Memory profiling

//go:noinline
func memprofHook(ptr unsafe.Pointer, size uintptr) {
	// Memory profiling hook
}

// Unused functions that should be reported

func unusedHook1() {
	// Not registered anywhere
}

func unusedHook2() {
	// Also not used
}

// init registers various hooks
func init() {
	// Register GC callbacks
	addGCCallback(gcCallback1)
	addGCCallback(gcCallback2)

	// Register profiler
	cpuProfiler.fn = profileCallback
	cpuProfiler.on = true

	// Note: In real runtime, these would be registered differently
	// This simulates the pattern
}

// Special runtime annotations

//go:cgo_import_dynamic libc_malloc malloc "libc.so.6"
//go:cgo_import_dynamic libc_free free "libc.so.6"

// Functions with specific calling conventions

//go:uintptrescapes
func uintptrEscapes(ptr uintptr) {
	// Tells compiler that ptr escapes
}

//go:notinheap
type notInHeap struct {
	data [1024]byte
}

// allocNotInHeap allocates non-heap memory
//
//go:nosplit
func allocNotInHeap() *notInHeap {
	return (*notInHeap)(unsafe.Pointer(uintptr(0)))
}
