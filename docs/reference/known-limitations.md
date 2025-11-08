# Known Limitations

This document describes known limitations of the `unusedfunc` analyzer and provides workarounds for each case.

**Note**: The analyzer fully supports generics. Generic method calls are tracked correctly with zero false positives.

## Template Method Calls

**Status**: Known limitation (matches industry-standard tools)

### Description

Methods called exclusively from Go template files (`.tmpl`, `.gotmpl`, `.html`) are flagged as unused because template execution happens at runtime through reflection.

### Example

```go
// batch.go
type TemplateContext struct {
    data string
}

// This method is called from template but flagged as unused.
func (t *TemplateContext) Export() string {
    return t.data
}

// template.gotmpl
{{ .Export }}  // Runtime reflection call - not visible to static analysis
```

### Why This Happens

Template execution follows this call chain:
```
template.Execute() → reflect.Value.MethodByName() → reflect.Value.Call() → Export()
```

Static analysis tools (including `unusedfunc`, `staticcheck`, and `deadcode`) cannot see through reflection-based method invocation. The SSA (Static Single Assignment) call graph only tracks statically determinable calls.

### Workaround

Use suppression comments to mark template methods as intentionally used:

```go
//nolint:unusedfunc // used in template.gotmpl:15
func (t *TemplateContext) Export() string {
    return t.data
}
```

Or use the `lint:ignore` pattern:

```go
//lint:ignore unusedfunc used in batch-esm-runner.gotmpl
func (t *TemplateContext) Export() string {
    return t.data
}
```

### Industry Standard Behavior

This limitation exists in all major Go static analysis tools:

- **`staticcheck` (U1000)**: Documents this as known limitation, requires manual suppression
- **`deadcode`**: Uses RTA algorithm (same as `unusedfunc`), cannot see template calls
- **`golangci-lint`**: Inherits limitations from underlying analyzers

### Future Enhancement

Template parsing support may be added in a future version via an opt-in `--include-templates` flag that would:
- Scan `.tmpl`, `.gotmpl`, `.html` files
- Extract `{{ .MethodName }}` patterns
- Mark referenced methods as used

---

## Assembly Function Calls

**Status**: Partially handled

### Description

Functions called exclusively from assembly (`.s`) files may be flagged as unused if not properly detected.

### Example

```go
// function.go
func asmImplemented() int  // Implemented in assembly

// impl.s
TEXT ·asmImplemented(SB), $0-8
    MOVQ $42, ret+0(FP)
    RET
```

### Current Handling

The analyzer includes assembly file scanning:
- Detects `TEXT ·FunctionName(SB)` declarations
- Detects `CALL ·FunctionName(SB)` invocations
- Marks assembly-implemented and assembly-called functions as used

### Limitations

- Only detects direct calls in `.s` files from same package
- Cannot track cross-package assembly calls
- Does not handle indirect assembly references

### Workaround

For edge cases not detected, use suppression comments:

```go
//nolint:unusedfunc // called from assembly: impl_amd64.s:42
func lowLevelFunction() { ... }
```

---

## `//go:linkname` Directives

**Status**: Partially handled

### Description

Functions aliased via `//go:linkname` directives may not have their aliasing relationships fully tracked.

### Example

```go
//go:linkname internalFunc runtime.externalFunc
func internalFunc() { ... }
```

### Current Handling

The analyzer detects `//go:linkname` directives and marks annotated functions as used to avoid false positives.

### Limitations

The tool "does not currently understand the aliasing created by `//go:linkname` directives, so it will fail to recognize that calls to a linkname-annotated function with no body in fact dispatch to the function named in the annotation."

This is an acknowledged limitation of the underlying RTA (Rapid Type Analysis) algorithm.

### Workaround

Functions with `//go:linkname` are automatically marked as used, so workarounds are typically not needed. If false positives occur, use suppression comments.

---

## Reflection Patterns

**Status**: Conservative handling

### Description

Methods called via reflection may be flagged as unused unless they match known reflection patterns.

### Known Safe Patterns

The analyzer uses conservative pattern matching to avoid false positives for common reflection targets. Functions matching these names are automatically marked as used:

- `String`, `GoString`, `Error` (fmt package interfaces)
- `Marshal`, `Unmarshal` (encoding packages)
- `Validate`, `Decode`, `Encode` (common validation/serialization patterns)

### Limitations

Custom reflection patterns or less common reflection usage may not be detected.

### Workaround

```go
//nolint:unusedfunc // called via reflect.Value.MethodByName in customHandler
func (t *Type) CustomReflectionMethod() { ... }
```

---

## Test Code in Vendored Dependencies

**Status**: Expected behavior

### Description

Test helper functions in vendored dependencies (e.g., Hugo vendors `text/template` tests) may be flagged as unused.

### Why This Happens

Vendored code includes test files that are not executed in your project. Test helpers called only from vendored tests are correctly identified as unused in your codebase.

### Workaround

Use file-wide suppression for vendored test code:

```go
//lint:file-ignore unusedfunc Vendored test code from upstream
package template_test
```

Or exclude vendor directories via command-line flags if the analyzer supports it (check `--help`).

---

## Configuration-Dependent Code

**Status**: Expected behavior

### Description

Functions that are used only under specific build tags or configurations may appear unused when analyzing with different settings.

### Example

```go
//go:build linux

package platform

func linuxSpecificFunction() { ... }  // Unused on non-Linux builds
```

### Why This Happens

The analyzer sees only the code visible under the current build configuration. Functions conditional on different build tags are correctly identified as unused in the current context.

### Workaround

Run the analyzer once for each configuration of interest:

```bash
# Analyze Linux build
GOOS=linux unusedfunc ./...

# Analyze Windows build
GOOS=windows unusedfunc ./...
```

Or use suppression comments for platform-specific code:

```go
//nolint:unusedfunc // used on Linux builds
func linuxSpecificFunction() { ... }
```

---

## Suppression Comment Reference

The analyzer supports multiple suppression comment formats:

### Format 1: `nolint` Style
```go
//nolint:unusedfunc
func myFunction() { ... }

//nolint:unusedfunc // reason for suppression
func myFunction() { ... }
```

### Format 2: `lint:ignore` Style
```go
//lint:ignore unusedfunc reason for suppression
func myFunction() { ... }
```

### Best Practices

1. Always include a reason explaining why the function is used
2. Reference the location of use (file:line) when applicable
3. Keep suppression comments close to the function declaration
4. Use file-wide suppression sparingly (only for generated or vendored code)

---

## Comparison with Other Tools

| Tool | Algorithm | Template Support | Reflection Handling | Generics Support |
|------|-----------|------------------|---------------------|------------------|
| `unusedfunc` | RTA | No (documented limitation) | Known patterns | Yes |
| `deadcode` | RTA | No (documented limitation) | Conservative (all exported) | Yes |
| `staticcheck U1000` | AST-based | No (documented limitation) | Limited | Yes |
| `golangci-lint` | Various | Depends on linter | Depends on linter | Depends on linter |

All major Go static analysis tools share the template method limitation. The `unusedfunc` analyzer matches industry-standard behavior while providing clear documentation and workarounds.

---

## Getting Help

If you encounter a false positive not covered by these known limitations:

1. Check if it matches a documented limitation above
2. Try suppression comments as a workaround
3. File an issue with a minimal reproducible example
4. Include the function signature, usage location, and analyzer output

For questions about whether behavior is a bug or expected limitation, consult this document first.
