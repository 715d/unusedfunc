# Modified RTA Reference

`internal/rta/rta.go` is a modified fork of `golang.org/x/tools/go/callgraph/rta`. It receives concrete SSA roots from `pkg/ssa` and returns reachable functions, generic objects without SSA functions, and runtime types. It uses a worklist and cross-product tables for address-taken function values and interface invokes. It does not retain a call graph.

## Project-Specific Behavior

| Area | Behavior |
|---|---|
| Interface conversion | A non-empty interface conversion retains only the required methods, including unexported marker methods. |
| Empty interface | Known JSON, fmt, XML, YAML, gob, binary, and SQL contexts retain selected methods. Direct `reflect.ValueOf` and `reflect.TypeOf` arguments retain exported methods. Other conversions retain exported methods conservatively. |
| Interface assertion | User-code assertions and interface-to-interface conversions conservatively scan compatible program types. Standard-library assertions use recorded runtime types. `errors.As` and `errors.AsType` marker patterns retain implementors. |
| Function values | Dynamic calls match address-taken functions by signature. `reflect.Value.Call` conservatively retains address-taken functions when it is in the program. |
| Finalizers | A direct `runtime.SetFinalizer` call retains a function or closure passed as its second argument. |
| Generics | A reachable instantiation retains its origin template. The SSA layer follows references from uninstantiated public templates and makes concrete helpers roots. |

The source owns the exact mapped APIs in `knownSafeFunctions`. The mapping considers the current function context, not whether an interface conversion reaches a particular mapped call. It does not trace values through stores, joins, or wrappers. Methods with their own type parameters are not retained for reflection.

## Interface Matching

RTA records runtime types from interface materialization and joins them with invoke sites. Fingerprints reject impossible matches before `types.Implements`. Type assertions and interface-to-interface conversions can scan compatible user types, so this behavior is conservative rather than flow-sensitive.

## Change Rules

Change this algorithm only with a focused fixture in `testdata/`. Preserve the reporting policy in `internal/analysis`; RTA only decides reachability. Read [known limitations](known-limitations.md) before you broaden dynamic-use rules, and [performance](../performance.md) before you change caching or traversal.
