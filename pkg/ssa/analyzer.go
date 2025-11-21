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

	// exportedTemplateObjects tracks exported generic template methods
	// that don't have SSA functions but should be treated as entry points
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
						// Only add if it's a function, not a method.
						if fn.Object() != nil {
							if sig, ok := fn.Object().Type().(*types.Signature); ok && sig.Recv() == nil {
								sa.entryPoints = append(sa.entryPoints, fn)
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
								// Get the SSA function for this method.
								if fn := sa.program.MethodValue(sel); fn != nil {
									sa.entryPoints = append(sa.entryPoints, fn)
								} else if sel.Obj() != nil {
									// Generic template method - no SSA function exists.
									// Mark as entry point by adding to analysis directly.
									sa.exportedTemplateObjects = append(sa.exportedTemplateObjects, sel.Obj())
								}
							}
						}

						// Also check pointer type methods.
						ptrType := types.NewPointer(namedType)
						ptrMset := sa.program.MethodSets.MethodSet(ptrType)
						for i := range ptrMset.Len() {
							sel := ptrMset.At(i)
							if sel.Obj().Exported() {
								if fn := sa.program.MethodValue(sel); fn != nil {
									if !slices.Contains(sa.entryPoints, fn) {
										sa.entryPoints = append(sa.entryPoints, fn)
									}
								} else if sel.Obj() != nil {
									// Generic template method - no SSA function exists.
									// Mark as entry point by adding to analysis directly.
									if !slices.Contains(sa.exportedTemplateObjects, sel.Obj()) {
										sa.exportedTemplateObjects = append(sa.exportedTemplateObjects, sel.Obj())
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

	if len(sa.entryPoints) == 0 {
		return nil, nil
	}

	// Filter out generic templates - RTA needs concrete instantiations, not templates with type parameters.
	var concreteEntryPoints []*ssa.Function
	for _, fn := range sa.entryPoints {
		// Generic filtering logic: keep non-generic functions and instantiated generics.
		// fn.TypeParams() == nil → non-generic function (keep)
		// fn.Origin() != nil → instantiated generic like Container[int].Clear (keep)
		// Both nil → uninstantiated template like Container[T].Clear (skip - not callable)
		// This prevents analyzing template code that isn't actually instantiated.
		if fn.TypeParams() == nil || fn.Origin() != nil {
			concreteEntryPoints = append(concreteEntryPoints, fn)
		}
	}

	if len(concreteEntryPoints) == 0 {
		return nil, nil
	}

	// Analyze with our fork of RTA which has been modified to be more precise.
	result := rta.Analyze(concreteEntryPoints)
	if result == nil {
		return nil, fmt.Errorf("RTA analysis failed")
	}

	// Extract reachable functions from the Reachable map directly.
	// This avoids the overhead of building the call graph.
	reachable := make(Set[types.Object])

	// The Reachable map contains exactly the reachable functions.
	// This includes functions marked reachable by RTA's analysis of:
	// - Direct calls
	// - Interface method dispatch
	// - Type assertions (TypeAssert)
	// - Interface conversions (MakeInterface, ChangeInterface)
	// - Runtime.SetFinalizer functions (called by GC)
	for fn := range result.Reachable {
		if fn != nil && fn.Object() != nil {
			reachable[fn.Object()] = struct{}{}
		}
	}

	// Also add objects that were tracked without SSA functions (generic template methods)
	for obj := range result.ReachableObjects {
		if obj != nil {
			reachable[obj] = struct{}{}
		}
	}

	// Mark exported template objects as reachable (they are entry points)
	for _, obj := range sa.exportedTemplateObjects {
		if obj != nil {
			reachable[obj] = struct{}{}
			// Analyze the template method body to find calls and mark callees as reachable.
			sa.markTemplateMethodCalls(obj, reachable)
		}
	}

	return reachable, nil
}

// markTemplateMethodCalls analyzes a generic template method using AST to find and mark
// actual method calls as reachable. This handles the case where SSA doesn't have the
// template body available for analysis.
func (sa *Analyzer) markTemplateMethodCalls(methodObj types.Object, reachable Set[types.Object]) {
	// Find the AST for this method.
	fn, ok := methodObj.(*types.Func)
	if !ok {
		return
	}

	sig := fn.Type().(*types.Signature)
	recv := sig.Recv()
	if recv == nil {
		return // Not a method.
	}

	pkg := methodObj.Pkg()
	if pkg == nil {
		return
	}

	var targetPkg *packages.Package
	for _, p := range sa.packages {
		if p.Types == pkg {
			targetPkg = p
			break
		}
	}
	if targetPkg == nil {
		return
	}

	methodName := fn.Name()
	recvType := recv.Type()
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}

	// Get receiver type name.
	var recvTypeName string
	if named, ok := recvType.(*types.Named); ok {
		recvTypeName = named.Obj().Name()
	} else {
		return
	}

	// Walk AST to find the method declaration and analyze its calls.
	for _, file := range targetPkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != methodName {
				return true
			}

			// Check if this is a method with the right receiver.
			if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
				return true
			}

			// Extract receiver type name from AST.
			var astRecvName string
			switch t := funcDecl.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					astRecvName = ident.Name
				} else if idx, ok := t.X.(*ast.IndexExpr); ok {
					if ident, ok := idx.X.(*ast.Ident); ok {
						astRecvName = ident.Name
					}
				} else if idx, ok := t.X.(*ast.IndexListExpr); ok {
					if ident, ok := idx.X.(*ast.Ident); ok {
						astRecvName = ident.Name
					}
				}
			case *ast.Ident:
				astRecvName = t.Name
			case *ast.IndexExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					astRecvName = ident.Name
				}
			case *ast.IndexListExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					astRecvName = ident.Name
				}
			}

			if astRecvName != recvTypeName {
				return true
			}

			// Found the method - now analyze calls in its body.
			ast.Inspect(funcDecl.Body, func(n2 ast.Node) bool {
				callExpr, ok := n2.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if this is a method call.
				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Look up the called method in TypesInfo.
				if calleeObj := targetPkg.TypesInfo.Uses[selExpr.Sel]; calleeObj != nil {
					if calleeFn, ok := calleeObj.(*types.Func); ok {
						// Mark as reachable and recurse.
						if _, exists := reachable[calleeFn]; !exists {
							reachable[calleeFn] = struct{}{}
							sa.markTemplateMethodCalls(calleeFn, reachable)
						}
					}
				}
				return true
			})

			return false // Found the method, stop searching
		})
	}
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
