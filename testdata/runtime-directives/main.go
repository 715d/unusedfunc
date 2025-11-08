// Package runtime tests runtime-callable functions
package main

import (
	"runtime"
	"sync"
	"unsafe"
)

// Runtime-specific functions that should NOT be reported as unused

//go:nosplit
func runtimeNosplit() {
	// Functions with //go:nosplit are often used by runtime
	// This directive prevents stack splitting
}

//go:noinline
func runtimeNoWriteBarrier() {
	// Functions with //go:noinline are often runtime-related
	// This prevents inlining
}

//go:noinline
func runtimeNoInline() int {
	// While noinline doesn't necessarily mean runtime use,
	// it's often used with runtime functions
	return 42
}

//go:norace
func runtimeSystemStack() {
	// Functions excluded from race detection
	// Typically used by runtime
}

// Finalizer functions - called by runtime GC

type Resource struct {
	id   int
	data []byte
}

// SetupFinalizer sets up object finalization
func SetupFinalizer() {
	r := &Resource{id: 1, data: make([]byte, 1024)}
	runtime.SetFinalizer(r, resourceFinalizer)
}

// resourceFinalizer is called by runtime when object is GC'd
func resourceFinalizer(r *Resource) {
	println("Finalizing resource", r.id)
}

// Callback patterns used by runtime

var (
	// initTasks holds initialization functions
	initTasks []func()

	// callbacks for runtime events
	callbacks = struct {
		mu sync.Mutex
		gc []func()
	}{}
)

// Called via init chain
func init() {
	// Register runtime callbacks
	registerGCCallback(onGC)

	// Setup finalizers
	SetupFinalizer()
}

// registerGCCallback registers a GC callback
func registerGCCallback(f func()) {
	callbacks.mu.Lock()
	defer callbacks.mu.Unlock()
	callbacks.gc = append(callbacks.gc, f)
}

// onGC is called when GC runs (simulated)
func onGC() {
	println("GC callback executed")
}

// Race detector annotations

//go:norace
func runtimeNoRace() {
	// Function excluded from race detection
	// Often used in runtime-critical code
}

// Memory synchronization

//go:nocheckptr
func runtimeNoCheckPtr() unsafe.Pointer {
	// Disables pointer checks
	// Used in low-level runtime code
	return unsafe.Pointer(&struct{}{})
}

// Regular functions for comparison

// UsedNormally is called from main
func UsedNormally() {
	println("Normal usage")
}

// UnusedRegular should be reported as unused
func UnusedRegular() {
	println("This is unused")
}

// Suppressed function
//
//nolint:unusedfunc
func SuppressedRuntime() {
	println("Suppressed via nolint")
}

// Functions that look like runtime functions but aren't

// This has a runtime-like name but no directives
func runtimeLookingName() {
	println("Not really a runtime function")
}

// Invalid directive format
// go:nosplit (has space - invalid)
func invalidDirective() {
	println("Invalid directive format")
}

// Complex runtime scenarios

type runtimeHook struct {
	before func()
	after  func()
}

var hooks = map[string]*runtimeHook{
	"alloc": {
		before: beforeAlloc,
		after:  afterAlloc,
	},
}

// beforeAlloc is used via runtime hook
func beforeAlloc() {
	// Called before allocation
}

// afterAlloc is used via runtime hook
func afterAlloc() {
	// Called after allocation
}

// UnusedHook should be reported
func UnusedHook() {
	// Not registered in hooks
}

// Go scheduler interaction

//go:nocheckptr
func schedulerFunction() {
	// Functions with pointer check directives
}

// Stack management

//go:nosplit
//go:noinline
func stackManagement() {
	// Multiple directives often indicate runtime use
}

// Main function
func main() {
	UsedNormally()

	// Trigger some runtime behavior
	runtime.GC()
	runtime.Gosched()

	// Use functions that interact with runtime
	ptr := runtimeNoCheckPtr()
	_ = ptr
}
