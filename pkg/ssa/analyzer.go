// Package ssa implements SSA-based precision analysis for method usage.
package ssa

import (
	"fmt"
	"slices"
	"strings"

	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/715d/unusedfunc/internal/rta"

	"github.com/715d/unusedfunc/internal/analysis"
)

const mainPkg = "main"

// Analyzer performs precise method usage analysis using SSA and call graphs.
type Analyzer struct {
	// program is the SSA program representation
	program *ssa.Program

	// ssaPkg is a map of package name to the SSA package representations
	ssaPkg map[string]*ssa.Package

	// packages are the packages being analyzed
	packages []*packages.Package

	// entryPoints contains all entry points for reachability analysis
	entryPoints []*ssa.Function

	// exportedTemplateObjects tracks public generic templates that RTA cannot use
	// as roots but whose bodies must be retained in normal mode.
	exportedTemplateObjects []types.Object

	// nameCache is used for computing canonical names
	nameCache *analysis.NameCache

	// strict mode: when true, exported functions are NOT automatically entry points
	strict bool
}

// NewAnalyzer creates a new SSA analyzer for the given packages.
func NewAnalyzer(pkgs []*packages.Package, strict bool) (*Analyzer, error) {
	// Filter out nil packages.
	validPkgs := make([]*packages.Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		validPkgs = append(validPkgs, pkg)
	}

	if len(validPkgs) == 0 {
		return nil, fmt.Errorf("no valid packages provided")
	}

	sa := &Analyzer{
		packages:  validPkgs,
		nameCache: analysis.NewNameCache(),
		strict:    strict,
	}

	if err := sa.buildSSAProgram(); err != nil {
		return nil, fmt.Errorf("build ssa program: %w", err)
	}

	return sa, nil
}

// AnalyzeFuncs performs SSA-based analysis to mark reachable functions as used.
func (sa *Analyzer) AnalyzeFuncs(funcs map[types.Object]*analysis.FuncInfo) error {
	// First, add functions with runtime directives as entry points.
	sa.addRuntimeDirectiveFunctions(funcs)

	// Add assembly-implemented functions as entry points if they're exported.
	// and add functions called from assembly to the initial worklist
	sa.addAssemblyRelatedFunctions(funcs)

	reachable, err := sa.findReachableMethods()
	if err != nil {
		return err
	}

	// Create a map for matching by package path and name.
	// This is needed because generic instantiations may create different types.Object.
	// instances for the same logical method
	reachableByName := make(Set[string], len(reachable))
	for obj := range reachable {
		// Skip objects without a package (built-in types, universe scope, etc.)
		if obj.Pkg() == nil {
			continue
		}
		// Skip anonymous functions (they have empty names)
		if obj.Name() == "" {
			continue
		}
		key := sa.nameCache.ComputeObjectName(obj)
		reachableByName[key] = struct{}{}
	}

	// Mark reachable methods as used.
	for obj, methodInfo := range funcs {

		if _, ok := reachable[obj]; ok {
			methodInfo.IsUsed = true
			continue
		}

		// Skip name-based matching for objects without a package (built-in types, universe scope, etc.)
		if obj.Pkg() == nil {
			continue
		}

		// Skip anonymous functions (they have empty names)
		if obj.Name() == "" {
			continue
		}

		// Try name-based match for objects with packages.
		key := sa.nameCache.ComputeObjectName(obj)
		if _, ok := reachableByName[key]; ok {
			methodInfo.IsUsed = true
		}

	}

	return nil
}

// buildSSAProgram constructs the SSA representation with generic instantiation
func (sa *Analyzer) buildSSAProgram() error {
	// Create SSA program with InstantiateGenerics mode for proper generic analysis.
	mode := ssa.InstantiateGenerics | ssa.BareInits

	var pkgs []*ssa.Package
	sa.program, pkgs = ssautil.AllPackages(sa.packages, mode)
	if sa.program != nil {
		sa.program.Build()
		sa.ssaPkg = make(map[string]*ssa.Package, len(pkgs))
		for _, pkg := range pkgs {
			sa.ssaPkg[pkg.Pkg.Path()] = pkg
		}
	}

	if sa.program == nil {
		return fmt.Errorf("SSA program construction failed")
	}

	// Identify entry points for reachability analysis.
	sa.findEntryPoints()
	return nil
}

// findEntryPoints identifies all entry points for reachability analysis.
//
// KNOWN LIMITATION: Methods called exclusively from Go template files (.tmpl, .gotmpl, .html)
// are NOT detected as entry points because template execution uses runtime reflection that is
// invisible to static analysis. Template execution follows this call chain:
//
//	template.Execute() → reflect.Value.MethodByName() → reflect.Value.Call() → YourMethod()
//
// The SSA call graph cannot see through this reflection chain. This is an accepted limitation.
// shared by all major Go static analysis tools (staticcheck, deadcode, golangci-lint).
//
// Workaround: Use suppression comments for template methods:
//
//	//nolint:unusedfunc // used in template.gotmpl:15
//	func (t *TemplateContext) Export() string { return t.data }
//
// See docs/reference/known-limitations.md for comprehensive limitation documentation.
func (sa *Analyzer) findEntryPoints() {
	sa.entryPoints = make([]*ssa.Function, 0, 4)

	// Only consider packages we're actually analyzing (target packages), not dependencies.
	for _, origPkg := range sa.packages {
		// Skip dependency packages - only analyze packages from our main module.
		if !isTargetPackage(origPkg) {
			continue
		}

		pkg := sa.ssaPkg[origPkg.PkgPath]
		if pkg == nil {
			continue
		}

		if main := pkg.Func("main"); main != nil {
			sa.entryPoints = append(sa.entryPoints, main)
		}

		for _, member := range pkg.Members {
			if fn, ok := member.(*ssa.Function); ok && fn != nil {
				if fn.Name() == "init" {
					sa.entryPoints = append(sa.entryPoints, fn)
				}
			}
		}

		// Add exported functions (not methods) and test functions as entry points.
		// Only in main packages should exported functions be considered entry points.
		for _, member := range pkg.Members {
			if fn, ok := member.(*ssa.Function); ok && fn != nil {
				if sa.isTestFunction(fn) {
					sa.entryPoints = append(sa.entryPoints, fn)
				}

				// Add exported functions only from non-main packages.
				// In main packages, only main() and init() are entry points.
				// In strict mode: don't add any exported functions as entry points (check if actually used).
				// In normal mode: add non-internal exported functions as entry points (public API).
				if sa.isExportedFunction(fn) && pkg.Pkg.Name() != mainPkg {
					isInternal := sa.isInternalPackage(pkg.Pkg.Path())
					// Strict mode: never add (check all for usage).
					// Normal mode: add only non-internal (public API assumed used).
					shouldAdd := !sa.strict && !isInternal

					if shouldAdd {
						// Only add functions, not methods.
						if fn.Object() != nil {
							if sig, ok := fn.Object().Type().(*types.Signature); ok && sig.Recv() == nil {
								if function, ok := fn.Object().(*types.Func); ok && isGenericFunction(function) {
									sa.addExportedTemplateObject(function)
								} else {
									sa.entryPoints = append(sa.entryPoints, fn)
								}
							}
						}
					}
				}
			}
		}

		// Add exported methods as entry points for library packages.
		// This ensures that unexported methods called by exported methods are not marked as unused.
		// In strict mode, skip this entirely (check all methods for actual usage).
		if pkg.Pkg.Name() != mainPkg && !sa.strict && !sa.isInternalPackage(pkg.Pkg.Path()) {
			for _, member := range pkg.Members {
				if typ, ok := member.(*ssa.Type); ok && typ != nil {
					// Get the underlying types.Type.
					if namedType, ok := typ.Object().Type().(*types.Named); ok {
						// Get all methods for this type (including pointer receivers)
						mset := sa.program.MethodSets.MethodSet(namedType)
						for i := range mset.Len() {
							sel := mset.At(i)
							if sel.Obj().Exported() {
								if method, ok := sel.Obj().(*types.Func); ok && isGenericFunction(method) {
									sa.addExportedTemplateObject(method)
								} else if fn := sa.program.MethodValue(sel); fn != nil {
									sa.entryPoints = append(sa.entryPoints, fn)
								}
							}
						}

						// Also check pointer type methods.
						ptrType := types.NewPointer(namedType)
						ptrMset := sa.program.MethodSets.MethodSet(ptrType)
						for i := range ptrMset.Len() {
							sel := ptrMset.At(i)
							if sel.Obj().Exported() {
								if method, ok := sel.Obj().(*types.Func); ok && isGenericFunction(method) {
									sa.addExportedTemplateObject(method)
								} else if fn := sa.program.MethodValue(sel); fn != nil {
									if !slices.Contains(sa.entryPoints, fn) {
										sa.entryPoints = append(sa.entryPoints, fn)
									}
								}
							}
						}
					}
				}
			}
		}

		// Add functions that might be called via reflection or build tags.
		for _, member := range pkg.Members {
			if fn, ok := member.(*ssa.Function); ok && fn != nil {
				if sa.isPotentialReflectionTarget(fn) {
					sa.entryPoints = append(sa.entryPoints, fn)
				}
			}
		}
	}
}

func (sa *Analyzer) isPotentialReflectionTarget(fn *ssa.Function) bool {
	// Functions that might be called via reflection should be considered entry points.
	// This is a conservative approach to avoid false positives.
	if fn.Object() == nil {
		return false
	}

	name := fn.Object().Name()

	// Common reflection targets.
	reflectionPatterns := []string{
		"String", "GoString", "Error", // fmt package interfaces
		"Marshal", "Unmarshal", // encoding packages
		"Validate", "Decode", "Encode", // common validation/serialization
	}
	return slices.Contains(reflectionPatterns, name)
}

// findReachableMethods returns a set of all methods reachable from entry points
// Uses Rapid Type Analysis (RTA) from the Go toolchain for proven correctness.
func (sa *Analyzer) findReachableMethods() (Set[types.Object], error) {
	if sa.program == nil {
		return nil, fmt.Errorf("SSA program not initialized")
	}

	if sa.entryPoints == nil {
		return nil, fmt.Errorf("entry points not initialized")
	}

	if len(sa.entryPoints) == 0 && len(sa.exportedTemplateObjects) == 0 {
		return nil, nil
	}

	var concreteEntryPoints []*ssa.Function
	for _, fn := range sa.entryPoints {
		if !sa.isGenericTemplate(fn) {
			concreteEntryPoints = append(concreteEntryPoints, fn)
		}
	}

	reachable := make(Set[types.Object])
	for _, obj := range sa.exportedTemplateObjects {
		if fn, ok := obj.(*types.Func); ok {
			reachable[fn] = struct{}{}
			sa.markTemplateFunctionReferences(fn, reachable, &concreteEntryPoints)
		}
	}

	if len(concreteEntryPoints) == 0 {
		return reachable, nil
	}

	result := rta.Analyze(concreteEntryPoints)
	if result == nil {
		return nil, fmt.Errorf("RTA analysis failed")
	}

	for fn := range result.Reachable {
		if fn != nil && fn.Object() != nil {
			reachable[fn.Object()] = struct{}{}
		}
	}
	for obj := range result.ReachableObjects {
		if obj != nil {
			reachable[obj] = struct{}{}
		}
	}

	return reachable, nil
}

// markTemplateFunctionReferences follows typed function references from an
// uninstantiated public generic template. Concrete references become RTA roots.
func (sa *Analyzer) markTemplateFunctionReferences(fn *types.Func, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if fn == nil || fn.Pkg() == nil {
		return
	}

	var targetPkg *packages.Package
	for _, pkg := range sa.packages {
		if pkg.Types == fn.Pkg() {
			targetPkg = pkg
			break
		}
	}
	if targetPkg == nil || targetPkg.TypesInfo == nil {
		return
	}

	for _, file := range targetPkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Pos() != fn.Pos() {
				return true
			}
			if funcDecl.Body == nil {
				return false
			}

			info := targetPkg.TypesInfo
			var visitBody func(*ast.BlockStmt, *types.Signature)
			visitBody = func(body *ast.BlockStmt, sig *types.Signature) {
				ast.Inspect(body, func(node ast.Node) bool {
					switch node := node.(type) {
					case *ast.FuncLit:
						literalSig, _ := info.TypeOf(node).(*types.Signature)
						visitBody(node.Body, literalSig)
						return false
					case *ast.Ident:
						callee, _ := info.Uses[node].(*types.Func)
						if callee != nil {
							sa.markTemplateFunctionReference(callee, reachable, concreteRoots)
						}
					case *ast.CallExpr:
						sa.markTemplateCallInterfaceArguments(node, info, reachable, concreteRoots)
					case *ast.AssignStmt:
						sa.markTemplateInterfaceAssignments(node.Lhs, node.Rhs, info, reachable, concreteRoots)
					case *ast.ValueSpec:
						sa.markTemplateInterfaceValueSpec(node, info, reachable, concreteRoots)
					case *ast.ReturnStmt:
						if sig != nil {
							sa.markTemplateInterfaceResults(node.Results, sig.Results(), info, reachable, concreteRoots)
						}
					}
					return true
				})
			}
			templateSig, _ := fn.Type().(*types.Signature)
			visitBody(funcDecl.Body, templateSig)
			return false
		})
	}
}

func (sa *Analyzer) markTemplateCallInterfaceArguments(call *ast.CallExpr, info *types.Info, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if typeAndValue, ok := info.Types[call.Fun]; ok && typeAndValue.IsType() {
		if len(call.Args) == 1 {
			sa.markTemplateInterfaceImplementation(info.TypeOf(call.Args[0]), info.TypeOf(call), reachable, concreteRoots)
		}
		return
	}

	funType := info.TypeOf(call.Fun)
	if funType == nil {
		return
	}
	sig, ok := funType.Underlying().(*types.Signature)
	if !ok || sig.Params() == nil {
		return
	}

	for i, arg := range call.Args {
		paramIndex := i
		if paramIndex >= sig.Params().Len() {
			if !sig.Variadic() {
				break
			}
			paramIndex = sig.Params().Len() - 1
		}
		if paramIndex < 0 {
			break
		}

		paramType := sig.Params().At(paramIndex).Type()
		if sig.Variadic() && i >= sig.Params().Len()-1 && !call.Ellipsis.IsValid() {
			slice, ok := types.Unalias(paramType).Underlying().(*types.Slice)
			if !ok {
				continue
			}
			paramType = slice.Elem()
		}
		sa.markTemplateInterfaceImplementation(info.TypeOf(arg), paramType, reachable, concreteRoots)
	}
}

func (sa *Analyzer) markTemplateInterfaceAssignments(lhs, rhs []ast.Expr, info *types.Info, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if len(lhs) != len(rhs) {
		return
	}
	for i := range lhs {
		sa.markTemplateInterfaceImplementation(info.TypeOf(rhs[i]), info.TypeOf(lhs[i]), reachable, concreteRoots)
	}
}

func (sa *Analyzer) markTemplateInterfaceValueSpec(spec *ast.ValueSpec, info *types.Info, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if len(spec.Names) != len(spec.Values) {
		return
	}
	for i, name := range spec.Names {
		targetType := info.TypeOf(name)
		if targetType == nil && spec.Type != nil {
			targetType = info.TypeOf(spec.Type)
		}
		sa.markTemplateInterfaceImplementation(info.TypeOf(spec.Values[i]), targetType, reachable, concreteRoots)
	}
}

func (sa *Analyzer) markTemplateInterfaceResults(values []ast.Expr, results *types.Tuple, info *types.Info, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if len(values) != results.Len() {
		return
	}
	for i := range values {
		sa.markTemplateInterfaceImplementation(info.TypeOf(values[i]), results.At(i).Type(), reachable, concreteRoots)
	}
}

func (sa *Analyzer) markTemplateInterfaceImplementation(concreteType, interfaceType types.Type, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if concreteType == nil || interfaceType == nil {
		return
	}

	iface, ok := types.Unalias(interfaceType).Underlying().(*types.Interface)
	if !ok || iface.Empty() {
		return
	}
	if _, ok := types.Unalias(concreteType).Underlying().(*types.Interface); ok {
		return
	}
	iface.Complete()
	if !types.Implements(concreteType, iface) {
		return
	}

	methodSet := sa.program.MethodSets.MethodSet(concreteType)
	for i := range iface.NumMethods() {
		interfaceMethod := iface.Method(i)
		if sel := methodSet.Lookup(interfaceMethod.Pkg(), interfaceMethod.Name()); sel != nil {
			sa.markTemplateFunctionReference(sel.Obj().(*types.Func), reachable, concreteRoots)
		}
	}
}

func (sa *Analyzer) markTemplateFunctionReference(fn *types.Func, reachable Set[types.Object], concreteRoots *[]*ssa.Function) {
	if _, exists := reachable[fn]; exists {
		return
	}
	reachable[fn] = struct{}{}

	if isGenericFunction(fn) {
		sa.markTemplateFunctionReferences(fn, reachable, concreteRoots)
		return
	}

	ssaFn := sa.getSSAFunction(fn)
	if ssaFn != nil && !slices.Contains(*concreteRoots, ssaFn) {
		*concreteRoots = append(*concreteRoots, ssaFn)
	}
}

func (sa *Analyzer) addExportedTemplateObject(obj types.Object) {
	if obj != nil && !slices.Contains(sa.exportedTemplateObjects, obj) {
		sa.exportedTemplateObjects = append(sa.exportedTemplateObjects, obj)
	}
}

func isGenericFunction(fn *types.Func) bool {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return false
	}
	if sig.TypeParams().Len() > 0 {
		return true
	}
	if sig.Recv() == nil {
		return false
	}

	recvType := sig.Recv().Type()
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}
	named, ok := recvType.(*types.Named)
	if !ok {
		return false
	}
	if named.TypeParams().Len() > 0 {
		return true
	}
	for i := range named.TypeArgs().Len() {
		if _, ok := named.TypeArgs().At(i).(*types.TypeParam); ok {
			return true
		}
	}
	return false
}

func (sa *Analyzer) isGenericTemplate(fn *ssa.Function) bool {
	if fn == nil || fn.TypeParams().Len() == 0 {
		return false
	}
	if fn.Origin() == nil {
		return true
	}

	// A generic receiver may be represented as an instantiation whose type
	// argument is still a type parameter, so RTA cannot traverse its body.
	for _, typeArg := range fn.TypeArgs() {
		if _, ok := typeArg.(*types.TypeParam); ok {
			return true
		}
	}
	return false
}

// addRuntimeDirectiveFunctions adds functions with runtime directives as entry points
func (sa *Analyzer) addRuntimeDirectiveFunctions(methods map[types.Object]*analysis.FuncInfo) {
	for obj, funcInfo := range methods {
		// If the function has runtime directives or CGo export, add it as an entry point.
		if funcInfo.HasRuntimeDirective || funcInfo.HasCGoExport {
			// Find the corresponding SSA function using on-demand lookup.
			if ssaFn := sa.getSSAFunction(obj); ssaFn != nil {
				if !slices.Contains(sa.entryPoints, ssaFn) {
					sa.entryPoints = append(sa.entryPoints, ssaFn)
				}
			}
		}
	}
}

// addAssemblyRelatedFunctions adds assembly-related functions to entry points
func (sa *Analyzer) addAssemblyRelatedFunctions(functions map[types.Object]*analysis.FuncInfo) {
	for obj, funcInfo := range functions {
		ssaFn := sa.getSSAFunction(obj)
		if ssaFn == nil {
			continue
		}

		// Add functions called from assembly as entry points.
		if funcInfo.CalledFromAssembly {
			if !slices.Contains(sa.entryPoints, ssaFn) {
				sa.entryPoints = append(sa.entryPoints, ssaFn)
			}
		}

		// Assembly-implemented exported functions should also be entry points.
		// in non-main packages (library APIs)
		if funcInfo.HasAssemblyImplementation && funcInfo.IsExported {
			if ssaFn.Package() != nil && ssaFn.Package().Pkg.Name() != mainPkg {
				if !slices.Contains(sa.entryPoints, ssaFn) {
					sa.entryPoints = append(sa.entryPoints, ssaFn)
				}
			}
		}
	}
}

// getSSAFunction provides on-demand lookup of SSA function for a types.Object
func (sa *Analyzer) getSSAFunction(obj types.Object) *ssa.Function {
	if fn, ok := obj.(*types.Func); ok {
		if ssaFn := sa.program.FuncValue(fn); ssaFn != nil {
			return ssaFn
		}
	}
	return nil
}

// isTestFunction checks if a function is a test function
func (sa *Analyzer) isTestFunction(fn *ssa.Function) bool {
	name := fn.Name()
	return strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Example")
}

// isExportedFunction checks if a function is exported
func (sa *Analyzer) isExportedFunction(fn *ssa.Function) bool {
	return fn.Object() != nil && fn.Object().Exported()
}

// isInternalPackage checks if a package path is an internal package.
func (sa *Analyzer) isInternalPackage(pkgPath string) bool {
	return strings.Contains(pkgPath, "/internal/") ||
		strings.HasSuffix(pkgPath, "/internal") ||
		strings.HasPrefix(pkgPath, "internal/") ||
		pkgPath == "internal"
}

type Set[T comparable] map[T]struct{}
