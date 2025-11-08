// Package unusedfunc provides unused function/method analysis.
package unusedfunc

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"
	"maps"
	goruntime "runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"

	"github.com/715d/unusedfunc/internal/analysis"
	"github.com/715d/unusedfunc/pkg/assembly"
	"github.com/715d/unusedfunc/pkg/runtime"
	"github.com/715d/unusedfunc/pkg/ssa"
	"github.com/715d/unusedfunc/pkg/suppress"
)

// AnalyzerOptions holds configuration options for the analyzer.
type AnalyzerOptions struct {
	SkipGenerated bool // Skip files with generated code markers.
}

// Analyzer orchestrates the method analysis process using SSA.
type Analyzer struct {
	suppressions *suppress.Checker
	nameCache    *analysis.NameCache
	opts         AnalyzerOptions
}

// NewAnalyzer creates a new analyzer with the given options.
func NewAnalyzer(opts AnalyzerOptions) *Analyzer {
	return &Analyzer{
		suppressions: suppress.NewChecker(),
		nameCache:    analysis.NewNameCache(),
		opts:         opts,
	}
}

// Analyze performs the unusedfunc analysis on the given packages.
func (a *Analyzer) Analyze(pkgs []*packages.Package) (map[types.Object]*analysis.FuncInfo, error) {
	// Validate input.
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages provided")
	}

	// Step 1: Load suppressions from all package files.
	if err := a.loadSuppressions(pkgs); err != nil {
		return nil, fmt.Errorf("failed to load suppressions: %w", err)
	}

	// Step 2: Scan assembly files for function implementations and calls.
	assemblyInfo := a.scanAssemblyFiles(pkgs)

	// Step 3: Create SSA analyzer and analyze all functions.
	ssaAnalyzer, err := ssa.NewAnalyzer(pkgs)
	if err != nil {
		return nil, fmt.Errorf("create SSA analyzer: %w", err)
	}

	// Step 4: Get all functions from packages.
	funcs := a.collectFunctions(pkgs, assemblyInfo)

	// Step 5: Run SSA analysis.
	if err := ssaAnalyzer.AnalyzeFuncs(funcs); err != nil {
		return nil, fmt.Errorf("SSA analysis failed: %w", err)
	}

	// Step 5: Check suppressions and mark suppressed functions.
	a.checkSuppressions(funcs)

	return funcs, nil
}

func (a *Analyzer) collectFunctions(pkgs []*packages.Package, assemblyInfo map[string]*assembly.Info) map[types.Object]*analysis.FuncInfo {
	// Lock-free concurrency pattern: pre-allocate results slice with exact size.
	// Each goroutine writes to its own index, eliminating need for locks/mutexes.
	// Safe under Go's memory model because:
	// 1. Slice elements are independent memory locations.
	// 2. Each goroutine has exclusive write access to its index.
	// 3. Main goroutine reads after all goroutines complete (sync via WaitGroup).
	results := make([]map[types.Object]*analysis.FuncInfo, len(pkgs))

	var wg errgroup.Group
	wg.SetLimit(goruntime.NumCPU())
	var total int64

	for idx, pkg := range pkgs {
		wg.Go(func() error {
			result := make(map[types.Object]*analysis.FuncInfo)

			// Filter out generated files if requested.
			filteredFiles := pkg.Syntax
			if a.opts.SkipGenerated {
				filteredFiles = nil
				for _, file := range pkg.Syntax {
					if file != nil && !a.isGeneratedFile(pkg.Fset, file) {
						filteredFiles = append(filteredFiles, file)
					}
				}
			}

			declMap := buildFuncDeclMapFromFiles(filteredFiles)

			scope := pkg.Types.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if fn, ok := obj.(*types.Func); ok {
					// Skip CGo-generated functions.
					if isCGoGeneratedFunction(fn.Name()) {
						continue
					}
					// Skip anonymous functions (they have empty names)
					if fn.Name() == "" {
						continue
					}
					funcInfo := analysis.NewFuncInfo(fn, pkg, a.nameCache)
					a.detectRuntimeDirectives(funcInfo, declMap)
					// Check if this function has assembly implementation or is called from assembly.
					if assemblyInfo[pkg.PkgPath] != nil {
						_, ok := assemblyInfo[pkg.PkgPath].ImplementedFunctions[fn.Name()]
						funcInfo.HasAssemblyImplementation = ok
						_, ok = assemblyInfo[pkg.PkgPath].CalledFunctions[fn.Name()]
						funcInfo.CalledFromAssembly = ok
					}
					result[fn] = funcInfo
				}
				// Also collect methods on named types.
				if tn, ok := obj.(*types.TypeName); ok {
					if named, ok := tn.Type().(*types.Named); ok {

						for i := range named.NumMethods() {
							method := named.Method(i)
							funcInfo := analysis.NewFuncInfo(method, pkg, a.nameCache)
							a.detectRuntimeDirectives(funcInfo, declMap)
							// Check if this method has assembly implementation or is called from assembly.
							if assemblyInfo[pkg.PkgPath] != nil {
								_, ok := assemblyInfo[pkg.PkgPath].ImplementedFunctions[method.Name()]
								funcInfo.HasAssemblyImplementation = ok
								_, ok = assemblyInfo[pkg.PkgPath].CalledFunctions[method.Name()]
								funcInfo.CalledFromAssembly = ok
							}
							result[method] = funcInfo
						}
					}
				}
			}

			results[idx] = result
			atomic.AddInt64(&total, int64(len(result)))
			return nil
		})
	}

	_ = wg.Wait()

	// Merge all results into final map.
	finalFuncs := make(map[types.Object]*analysis.FuncInfo, total)
	for _, pkgFuncs := range results {
		maps.Copy(finalFuncs, pkgFuncs)
	}
	return finalFuncs
}

// loadSuppressions loads suppression comments from all files in the given packages
func (a *Analyzer) loadSuppressions(pkgs []*packages.Package) error {
	// Clear any existing suppressions.
	a.suppressions.Clear()

	// Collect all files from all packages.
	var allFiles []*ast.File
	var fset *token.FileSet

	for _, pkg := range pkgs {
		if pkg == nil {
			continue // Skip nil packages
		}
		if pkg.Fset != nil {
			fset = pkg.Fset
		}
		if pkg.Syntax != nil {
			for _, file := range pkg.Syntax {
				if file != nil {
					allFiles = append(allFiles, file)
				}
			}
		}
	}

	// Load suppressions from all files.
	if fset != nil && len(allFiles) > 0 {
		return a.suppressions.Load(fset, allFiles)
	}

	return nil
}

// checkSuppressions checks each function for suppression comments and marks them accordingly
func (a *Analyzer) checkSuppressions(funcs map[types.Object]*analysis.FuncInfo) {
	for _, funcInfo := range funcs {
		funcInfo.IsSuppressed, _ = a.suppressions.IsSuppressed(funcInfo.DeclarationPos)
	}
}

// funcDeclMap holds a pre-built map of function declarations per package
type funcDeclMap map[string]*ast.FuncDecl

// buildFuncDeclMapFromFiles builds a map of function name -> FuncDecl for quick lookup from a list of files.
func buildFuncDeclMapFromFiles(files []*ast.File) funcDeclMap {
	declMap := make(funcDeclMap)
	for _, file := range files {
		if file == nil {
			continue
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name != nil {
				declMap[fn.Name.Name] = fn
			}
		}
	}
	return declMap
}

// isGeneratedFile checks if a file contains generated code markers.
func (a *Analyzer) isGeneratedFile(fset *token.FileSet, file *ast.File) bool {
	if file == nil || fset == nil {
		return false
	}

	// Check for common generated file markers in comments.
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text
			// Check for standard generated file markers.
			if strings.Contains(text, "Code generated") ||
				strings.Contains(text, "DO NOT EDIT") ||
				strings.Contains(text, "autogenerated") ||
				strings.Contains(text, "AUTO-GENERATED") {
				return true
			}
		}
	}

	return false
}

// detectRuntimeDirectives examines the declMap to detect runtime directives on functions
func (a *Analyzer) detectRuntimeDirectives(funcInfo *analysis.FuncInfo, declMap funcDeclMap) {
	if funcInfo.Object == nil {
		return
	}

	// Check if it's a known runtime hook function first.
	if runtime.IsRuntimeHookFunction(funcInfo.Object.Name()) {
		funcInfo.HasRuntimeDirective = true
		return
	}

	funcName := funcInfo.Object.Name()

	// Direct lookup instead of AST walk.
	if fn, exists := declMap[funcName]; exists {
		directive := runtime.HasRuntimeDirective(fn)
		if directive.Valid {
			funcInfo.HasRuntimeDirective = true
			// Check if it's specifically a CGo export directive.
			if directive.Type == runtime.DirectiveCGoExport {
				funcInfo.HasCGoExport = true
			}
		}
	}
}

// scanAssemblyFiles scans all packages for assembly files and returns assembly information
func (a *Analyzer) scanAssemblyFiles(pkgs []*packages.Package) map[string]*assembly.Info {
	result := make(map[string]*assembly.Info)

	for _, pkg := range pkgs {

		info, err := assembly.ScanPackage(pkg)
		if err != nil {
			// Log but don't fail - assembly scanning is supplementary.
			slog.Warn("scanning assembly files", "package", pkg.PkgPath, "error", err)
			continue
		}

		if len(info.ImplementedFunctions) > 0 || len(info.CalledFunctions) > 0 {
			result[pkg.PkgPath] = info
		}
	}

	return result
}

// isCGoGeneratedFunction checks if a function name indicates it's generated by CGo
func isCGoGeneratedFunction(name string) bool {
	return strings.HasPrefix(name, "_Cgo_") || strings.HasPrefix(name, "_cgo_")
}
