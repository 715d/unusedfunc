# Known Limitations

`unusedfunc` proves static reachability. Suppress a declaration only after you confirm a dynamic use:

```go
//nolint:unusedfunc // used in template.gotmpl:15
func (t *TemplateContext) Export() string { return t.data }
```

`//lint:ignore unusedfunc reason` also works. Put the comment on the declaration line or the line before it.

## Dynamic Calls

The analyzer does not parse template text or resolve names from `reflect.Value.MethodByName`, registries, maps, or struct tags. Function values in maps and registries can still match dynamic calls by signature. It recognizes selected standard-library and common-library reflection contexts, but this is not value-flow analysis. Direct `reflect.ValueOf` and `reflect.TypeOf` calls receive special handling. If `reflect.Value.Call` is in the loaded program, address-taken functions can be retained conservatively.

## Generated Files

`--skip-generated` defaults to true. It only removes generated declarations from runtime-directive detection. Generated functions still load, enter SSA, and can report. Use a suppression for a confirmed dynamic use or narrow the package pattern.

## Assembly and Linkname

The scanner reads build-selected `.s` files for package-local direct `TEXT ·name(SB)` and `CALL ·name(SB)` forms. It excludes assembly implementations and roots direct assembly calls. Cross-package and indirect assembly references are not recognized.

A declaration-attached `//go:linkname` directive excludes that declaration from reporting. The analyzer does not model alias relationships or arbitrary file comments.

## Generic APIs

Normal mode retains public generic templates without a local instantiation by following typed-AST references and adding concrete helpers as roots. It does not follow general value flow through containers, fields, multi-result expressions, or generic call chains. Test a concrete instantiation or suppress a confirmed use.

## Build and Test Configurations

The loader always includes tests. Test, benchmark, and example declarations are roots. Analyze every target GOOS, GOARCH, CGo, and build-tag configuration; a use outside the selected configuration is not visible.
