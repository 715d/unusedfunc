# Generic Method Calls Test Case

## Purpose

This test case reproduces a critical bug in the RTA (Rapid Type Analysis) where generic methods calling other generic methods are not properly tracked as used.

## Bug Description

**Symptom:** Unexported generic helper methods are reported as unused even though they are called by exported generic methods.

**Root Cause:** The RTA fails to track call chains within generic types when:
1. The package is a library (non-main)
2. There's no concrete instantiation in the analyzed package
3. Generic methods call other generic methods on the same type

## Example

```go
type processor[T any] struct { value T }

// Reported as UNUSED (bug!)
func (p *processor[T]) helper() T {
    return p.value
}

// Correctly marked as USED
func (p *processor[T]) Process() T {
    return p.helper()  // This call is not tracked!
}

// Entry point
func NewProcessor[T any](val T) *processor[T] {
    return &processor[T]{value: val}
}
```

## Current Behavior

Running `unusedfunc .` reports:
- ❌ `helper()` - FALSE POSITIVE (actually used by `Process()`)
- ❌ `validate()` - FALSE POSITIVE (actually used by `Process()`)
- ❌ `internalProcess()` - FALSE POSITIVE (actually used by `Add()`)
- ❌ `internalValidate()` - FALSE POSITIVE (actually used by `Add()`)
- ✅ `unusedHelper()` - TRUE POSITIVE (genuinely unused)

## Expected Behavior

After fix, only genuinely unused methods should be reported:
- ✅ `unusedHelper()` only

## Impact

This bug affects:
- **GORM:** 8 false positives (42% of all reports)
- Any codebase using generic helper methods
- Libraries with generic fluent APIs

## Related Files

- `docs/context/20251030-bug2-ACTUAL-root-cause.md` - Detailed investigation
- Real-world example: GORM's `generics.go` with `apply()`, `getInstance()`, etc.
