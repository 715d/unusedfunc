package ssa

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"

	"github.com/715d/unusedfunc/internal/analysis"
)

// createTestPackageForSSA creates a test package suitable for SSA analysis
func createTestPackageForSSA(name, path string) *packages.Package {
	pkg := &packages.Package{
		ID:      path,
		Name:    name,
		PkgPath: path,
		Types:   gotypes.NewPackage(path, name),
	}
	return pkg
}

// TestSSAAnalyzer_NewSSAAnalyzer tests SSA analyzer creation.
func TestSSAAnalyzer_NewSSAAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		setupPkgs   func() []*packages.Package
		expectError bool
		checkResult func(*testing.T, *Analyzer)
	}{
		{
			name: "valid packages",
			setupPkgs: func() []*packages.Package {
				return []*packages.Package{
					createTestPackageForSSA("main", "main"),
				}
			},
			expectError: false,
			checkResult: func(t *testing.T, analyzer *Analyzer) {
				require.NotNil(t, analyzer)
				require.NotNil(t, analyzer.program, "Expected SSA program to be initialized")
			},
		},
		{
			name: "empty package slice",
			setupPkgs: func() []*packages.Package {
				return []*packages.Package{}
			},
			expectError: true,
			checkResult: nil,
		},
		{
			name: "nil package in slice",
			setupPkgs: func() []*packages.Package {
				return []*packages.Package{nil}
			},
			expectError: true,
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := tt.setupPkgs()
			analyzer, err := NewAnalyzer(pkgs)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.checkResult != nil && !tt.expectError {
				tt.checkResult(t, analyzer)
			}
		})
	}
}

// TestSSAAnalyzer_BuildSSAProgram tests SSA program construction.
func TestSSAAnalyzer_BuildSSAProgram(t *testing.T) {
	tests := []struct {
		name         string
		setupPkgs    func() []*packages.Package
		expectError  bool
		checkProgram func(*testing.T, *ssa.Program)
	}{
		{
			name: "simple package",
			setupPkgs: func() []*packages.Package {
				pkg := createTestPackageForSSA("test", "test")
				// In a real scenario, these would be populated by go/packages.
				pkg.TypesSizes = gotypes.SizesFor("gc", "amd64")
				return []*packages.Package{pkg}
			},
			expectError: false,
			checkProgram: func(t *testing.T, program *ssa.Program) {
				require.NotNil(t, program, "Expected SSA program, got nil")
				// Note: Mode field is unexported in ssa.Program, so we can't check it directly.
				// The mode is set correctly in buildSSAProgram()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := tt.setupPkgs()
			analyzer, err := NewAnalyzer(pkgs)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.checkProgram != nil && !tt.expectError {
				tt.checkProgram(t, analyzer.program)
			}
		})
	}
}

// TestSSAAnalyzer_AnalyzeMethods tests method analysis.
func TestSSAAnalyzer_AnalyzeMethods(t *testing.T) {
	tests := []struct {
		name          string
		setupAnalyzer func() (*Analyzer, map[gotypes.Object]*analysis.FuncInfo)
		expectError   bool
		checkResults  func(*testing.T, map[gotypes.Object]*analysis.FuncInfo)
	}{
		{
			name: "empty methods map",
			setupAnalyzer: func() (*Analyzer, map[gotypes.Object]*analysis.FuncInfo) {
				analyzer := &Analyzer{}
				prog := ssa.NewProgram(nil, ssa.SanityCheckFunctions)
				analyzer.program = prog
				// Note: no call graph set, so should fail validation.
				return analyzer, make(map[gotypes.Object]*analysis.FuncInfo)
			},
			expectError:  true,
			checkResults: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, methods := tt.setupAnalyzer()

			err := analyzer.AnalyzeFuncs(methods)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.checkResults != nil && !tt.expectError {
				tt.checkResults(t, methods)
			}
		})
	}
}

// TestSSAAnalyzer_InterfaceMethodReachability tests that interface method implementations.
// are correctly marked as used when called through the interface
func TestSSAAnalyzer_InterfaceMethodReachability(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedUsed   []string // Method names that should be marked as used
		expectedUnused []string // Method names that should NOT be marked as used
	}{
		{
			name: "basic interface method call",
			code: `
package main

type Writer interface {
	Write([]byte) (int, error)
}

type FileWriter struct{}

func (fw *FileWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (fw *FileWriter) unusedMethod() {}

func main() {
	var w Writer = &FileWriter{}
	w.Write([]byte("hello"))
}
`,
			expectedUsed:   []string{"main", "Write"},
			expectedUnused: []string{"unusedMethod"},
		},
		{
			name: "interface method through function parameter",
			code: `
package main

// import removed for test

type Printer interface {
	Print(string)
}

type ConsolePrinter struct{}

func (cp *ConsolePrinter) Print(msg string) {
	// print implementation
}

func (cp *ConsolePrinter) Debug(msg string) {
	// debug implementation
}

func usePrinter(p Printer) {
	p.Print("hello")
}

func main() {
	cp := &ConsolePrinter{}
	usePrinter(cp)
}
`,
			expectedUsed:   []string{"main", "usePrinter", "Print"},
			expectedUnused: []string{"Debug"},
		},
		{
			name: "embedded interface methods",
			code: `
package main

type Reader interface {
	Read([]byte) (int, error)
}

type Writer interface {
	Write([]byte) (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Buffer struct{}

func (b *Buffer) Read(data []byte) (int, error) {
	return 0, nil
}

func (b *Buffer) Write(data []byte) (int, error) {
	return len(data), nil
}

func (b *Buffer) Clear() {}

func process(rw ReadWriter) {
	buf := make([]byte, 100)
	rw.Read(buf)
	rw.Write(buf)
}

func main() {
	b := &Buffer{}
	process(b)
}
`,
			expectedUsed:   []string{"main", "process", "Read", "Write"},
			expectedUnused: []string{"Clear"},
		},
		{
			name: "interface satisfied by embedded struct",
			code: `
package main

type Logger interface {
	Log(string)
}

type BaseLogger struct{}

func (bl *BaseLogger) Log(msg string) {}

type FileLogger struct {
	BaseLogger
}

func (fl *FileLogger) LogToFile(msg string) {}

func useLogger(l Logger) {
	l.Log("test")
}

func main() {
	fl := &FileLogger{}
	useLogger(fl)
}
`,
			expectedUsed:   []string{"main", "useLogger", "Log"},
			expectedUnused: []string{"LogToFile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test code.
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			pkg := &packages.Package{
				ID:         "test",
				Name:       "main",
				PkgPath:    "test",
				Syntax:     []*ast.File{file},
				Fset:       fset,
				TypesSizes: gotypes.SizesFor("gc", "amd64"),
			}

			// Type check.
			conf := gotypes.Config{Importer: importer.Default()}
			info := &gotypes.Info{
				Types:      make(map[ast.Expr]gotypes.TypeAndValue),
				Defs:       make(map[*ast.Ident]gotypes.Object),
				Uses:       make(map[*ast.Ident]gotypes.Object),
				Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
				Implicits:  make(map[ast.Node]gotypes.Object),
			}
			pkg.TypesInfo = info
			pkg.Types, err = conf.Check("test", fset, []*ast.File{file}, info)
			require.NoError(t, err)

			analyzer, err := NewAnalyzer([]*packages.Package{pkg})
			require.NoError(t, err)

			// Build a map of all methods.
			methods := make(map[gotypes.Object]*analysis.FuncInfo)
			for _, obj := range info.Defs {
				if obj == nil {
					continue
				}
				if fn, ok := obj.(*gotypes.Func); ok {
					// Skip interface methods - we only want concrete implementations.
					if sig, ok := fn.Type().(*gotypes.Signature); ok && sig.Recv() != nil {
						if _, isInterface := sig.Recv().Type().Underlying().(*gotypes.Interface); isInterface {
							continue
						}
					}
					methods[fn] = analysis.NewFuncInfo(fn, pkg, analyzer.nameCache)
				}
			}

			// Analyze methods.
			err = analyzer.AnalyzeFuncs(methods)
			require.NoError(t, err)

			// Debug: print what methods we found.
			t.Logf("Found %d methods:", len(methods))
			for obj, info := range methods {
				sig, _ := obj.Type().(*gotypes.Signature)
				recvStr := "no receiver"
				if sig != nil && sig.Recv() != nil {
					recvStr = sig.Recv().Type().String()
				}
				t.Logf("  - %s (%s): IsUsed=%v", obj.Name(), recvStr, info.IsUsed)
			}

			for _, methodName := range tt.expectedUsed {
				found := false
				for obj, info := range methods {
					if obj.Name() == methodName {
						assert.True(t, info.IsUsed, "Method %s should be marked as used", methodName)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
			}

			for _, methodName := range tt.expectedUnused {
				found := false
				allUnused := true
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							allUnused = false
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, allUnused, "All methods named %s should NOT be marked as used", methodName)
			}
		})
	}
}

// TestSSAAnalyzer_GenericMethodReachability tests that generic methods are correctly.
// marked as used when instantiated and called
func TestSSAAnalyzer_GenericMethodReachability(t *testing.T) {
	t.Skip("Skipping generic tests due to SSA builder limitations with generics")
	tests := []struct {
		name           string
		code           string
		expectedUsed   []string
		expectedUnused []string
	}{
		{
			name: "generic function instantiation",
			code: `
package main

func Map[T any](slice []T, fn func(T) T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func main() {
	nums := []int{1, 2, 3}
	doubled := Map(nums, func(x int) int { return x * 2 })
	_ = doubled
}
`,
			expectedUsed:   []string{"main", "Map"},
			expectedUnused: []string{"Filter"},
		},
		{
			name: "generic type with methods",
			code: `
package main

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}

func main() {
	s := &Stack[int]{}
	s.Push(42)
	_ = s.Pop()
}
`,
			expectedUsed:   []string{"main", "Push", "Pop"},
			expectedUnused: []string{"Size"},
		},
		{
			name: "generic interface constraint",
			code: `
package main

type Ordered interface {
	~int | ~float64 | ~string
}

func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {
	x := Min(10, 20)
	_ = x
}
`,
			expectedUsed:   []string{"main", "Min"},
			expectedUnused: []string{"Max"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test code.
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			pkg := &packages.Package{
				ID:         "test",
				Name:       "main",
				PkgPath:    "test",
				Syntax:     []*ast.File{file},
				Fset:       fset,
				TypesSizes: gotypes.SizesFor("gc", "amd64"),
			}

			// Type check.
			conf := gotypes.Config{Importer: importer.Default()}
			info := &gotypes.Info{
				Types:      make(map[ast.Expr]gotypes.TypeAndValue),
				Defs:       make(map[*ast.Ident]gotypes.Object),
				Uses:       make(map[*ast.Ident]gotypes.Object),
				Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
				Implicits:  make(map[ast.Node]gotypes.Object),
			}
			pkg.TypesInfo = info
			pkg.Types, err = conf.Check("test", fset, []*ast.File{file}, info)
			require.NoError(t, err)

			analyzer, err := NewAnalyzer([]*packages.Package{pkg})
			require.NoError(t, err)

			// Build methods map.
			methods := make(map[gotypes.Object]*analysis.FuncInfo)
			for _, obj := range info.Defs {
				if obj == nil {
					continue
				}
				if fn, ok := obj.(*gotypes.Func); ok {
					// Skip interface methods - we only want concrete implementations.
					if sig, ok := fn.Type().(*gotypes.Signature); ok && sig.Recv() != nil {
						if _, isInterface := sig.Recv().Type().Underlying().(*gotypes.Interface); isInterface {
							continue
						}
					}
					methods[fn] = analysis.NewFuncInfo(fn, pkg, analyzer.nameCache)
				}
			}

			// Analyze.
			err = analyzer.AnalyzeFuncs(methods)
			require.NoError(t, err)

			for _, methodName := range tt.expectedUsed {
				found := false
				atLeastOneUsed := false
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							atLeastOneUsed = true
							break
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, atLeastOneUsed, "At least one method %s should be marked as used", methodName)
			}

			for _, methodName := range tt.expectedUnused {
				found := false
				allUnused := true
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							allUnused = false
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, allUnused, "All methods named %s should NOT be marked as used", methodName)
			}
		})
	}
}

// TestSSAAnalyzer_TransitiveReachability tests that methods are marked as used.
// when they are transitively reachable (A calls B calls C)
func TestSSAAnalyzer_TransitiveReachability(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedUsed   []string
		expectedUnused []string
	}{
		{
			name: "simple transitive calls",
			code: `
package main

func a() {
	b()
}

func b() {
	c()
}

func c() {
	// end of chain
}

func d() {
	// not called
}

func main() {
	a()
}
`,
			expectedUsed:   []string{"main", "a", "b", "c"},
			expectedUnused: []string{"d"},
		},
		{
			name: "transitive method calls",
			code: `
package main

type Service struct{}

func (s *Service) Start() {
	s.initialize()
}

func (s *Service) initialize() {
	s.loadConfig()
}

func (s *Service) loadConfig() {
	// load config
}

func (s *Service) Stop() {
	// not called
}

func main() {
	s := &Service{}
	s.Start()
}
`,
			expectedUsed:   []string{"main", "Start", "initialize", "loadConfig"},
			expectedUnused: []string{"Stop"},
		},
		{
			name: "transitive interface calls",
			code: `
package main

type Handler interface {
	Handle()
}

type ChainHandler struct {
	next Handler
}

func (ch *ChainHandler) Handle() {
	ch.process()
	if ch.next != nil {
		ch.next.Handle()
	}
}

func (ch *ChainHandler) process() {
	ch.validate()
}

func (ch *ChainHandler) validate() {
	// validation logic
}

func (ch *ChainHandler) cleanup() {
	// not called
}

func main() {
	h := &ChainHandler{}
	h.Handle()
}
`,
			expectedUsed:   []string{"main", "Handle", "process", "validate"},
			expectedUnused: []string{"cleanup"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test code.
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			pkg := &packages.Package{
				ID:         "test",
				Name:       "main",
				PkgPath:    "test",
				Syntax:     []*ast.File{file},
				Fset:       fset,
				TypesSizes: gotypes.SizesFor("gc", "amd64"),
			}

			// Type check.
			conf := gotypes.Config{Importer: importer.Default()}
			info := &gotypes.Info{
				Types:      make(map[ast.Expr]gotypes.TypeAndValue),
				Defs:       make(map[*ast.Ident]gotypes.Object),
				Uses:       make(map[*ast.Ident]gotypes.Object),
				Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
				Implicits:  make(map[ast.Node]gotypes.Object),
			}
			pkg.TypesInfo = info
			pkg.Types, err = conf.Check("test", fset, []*ast.File{file}, info)
			require.NoError(t, err)

			analyzer, err := NewAnalyzer([]*packages.Package{pkg})
			require.NoError(t, err)

			// Build methods map.
			methods := make(map[gotypes.Object]*analysis.FuncInfo)
			for _, obj := range info.Defs {
				if obj == nil {
					continue
				}
				if fn, ok := obj.(*gotypes.Func); ok {
					// Skip interface methods - we only want concrete implementations.
					if sig, ok := fn.Type().(*gotypes.Signature); ok && sig.Recv() != nil {
						if _, isInterface := sig.Recv().Type().Underlying().(*gotypes.Interface); isInterface {
							continue
						}
					}
					methods[fn] = analysis.NewFuncInfo(fn, pkg, analyzer.nameCache)
				}
			}

			// Analyze.
			err = analyzer.AnalyzeFuncs(methods)
			require.NoError(t, err)

			for _, methodName := range tt.expectedUsed {
				found := false
				atLeastOneUsed := false
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							atLeastOneUsed = true
							break
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, atLeastOneUsed, "At least one method %s should be marked as used", methodName)
			}

			for _, methodName := range tt.expectedUnused {
				found := false
				allUnused := true
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							allUnused = false
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, allUnused, "All methods named %s should NOT be marked as used", methodName)
			}
		})
	}
}

// TestSSAAnalyzer_TypeAssertionReachability tests that methods are correctly marked.
// as used when called after type assertions
func TestSSAAnalyzer_TypeAssertionReachability(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedUsed   []string
		expectedUnused []string
	}{
		{
			name: "type assertion with method call",
			code: `
package main

// import removed for test

type Shape interface {
	Area() float64
}

type Circle struct {
	radius float64
}

func (c *Circle) Area() float64 {
	return 3.14 * c.radius * c.radius
}

func (c *Circle) Circumference() float64 {
	return 2 * 3.14 * c.radius
}

func processShape(s Shape) {
	_ = s.Area()
	
	if circle, ok := s.(*Circle); ok {
		_ = circle.Circumference()
	}
}

func main() {
	c := &Circle{radius: 5}
	processShape(c)
}
`,
			expectedUsed:   []string{"main", "processShape", "Area", "Circumference"},
			expectedUnused: []string{},
		},
		{
			name: "type switch with method calls",
			code: `
package main

type Animal interface {
	Sound() string
}

type Dog struct{}

func (d *Dog) Sound() string {
	return "woof"
}

func (d *Dog) Fetch() {
	// fetch logic
}

type Cat struct{}

func (c *Cat) Sound() string {
	return "meow"
}

func (c *Cat) Scratch() {
	// scratch logic
}

func handleAnimal(a Animal) {
	switch v := a.(type) {
	case *Dog:
		v.Fetch()
	case *Cat:
		// Cat's Scratch is not called.
		_ = v
	}
}

func main() {
	d := &Dog{}
	handleAnimal(d)
}
`,
			expectedUsed:   []string{"main", "handleAnimal", "Sound", "Fetch"},
			expectedUnused: []string{"Scratch"},
		},
		{
			name: "nested type assertions",
			code: `
package main

type Reader interface {
	Read() string
}

type Writer interface {
	Write(string)
}

type File struct{}

func (f *File) Read() string {
	return "data"
}

func (f *File) Write(data string) {
	// write logic
}

func (f *File) Close() {
	// close logic
}

func process(r Reader) {
	data := r.Read()
	
	if w, ok := r.(Writer); ok {
		w.Write(data)
		
		if f, ok := w.(*File); ok {
			f.Close()
		}
	}
}

func main() {
	f := &File{}
	process(f)
}
`,
			expectedUsed:   []string{"main", "process", "Read", "Write", "Close"},
			expectedUnused: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test code.
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			pkg := &packages.Package{
				ID:         "test",
				Name:       "main",
				PkgPath:    "test",
				Syntax:     []*ast.File{file},
				Fset:       fset,
				TypesSizes: gotypes.SizesFor("gc", "amd64"),
			}

			// Type check.
			conf := gotypes.Config{Importer: importer.Default()}
			info := &gotypes.Info{
				Types:      make(map[ast.Expr]gotypes.TypeAndValue),
				Defs:       make(map[*ast.Ident]gotypes.Object),
				Uses:       make(map[*ast.Ident]gotypes.Object),
				Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
				Implicits:  make(map[ast.Node]gotypes.Object),
			}
			pkg.TypesInfo = info
			pkg.Types, err = conf.Check("test", fset, []*ast.File{file}, info)
			require.NoError(t, err)

			analyzer, err := NewAnalyzer([]*packages.Package{pkg})
			require.NoError(t, err)

			// Build methods map.
			methods := make(map[gotypes.Object]*analysis.FuncInfo)
			for _, obj := range info.Defs {
				if obj == nil {
					continue
				}
				if fn, ok := obj.(*gotypes.Func); ok {
					// Skip interface methods - we only want concrete implementations.
					if sig, ok := fn.Type().(*gotypes.Signature); ok && sig.Recv() != nil {
						if _, isInterface := sig.Recv().Type().Underlying().(*gotypes.Interface); isInterface {
							continue
						}
					}
					methods[fn] = analysis.NewFuncInfo(fn, pkg, analyzer.nameCache)
				}
			}

			// Analyze.
			err = analyzer.AnalyzeFuncs(methods)
			require.NoError(t, err)

			for _, methodName := range tt.expectedUsed {
				found := false
				atLeastOneUsed := false
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							atLeastOneUsed = true
							break
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, atLeastOneUsed, "At least one method %s should be marked as used", methodName)
			}

			for _, methodName := range tt.expectedUnused {
				found := false
				allUnused := true
				for obj, info := range methods {
					if obj.Name() == methodName {
						found = true
						if info.IsUsed {
							allUnused = false
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
				assert.True(t, allUnused, "All methods named %s should NOT be marked as used", methodName)
			}
		})
	}
}

// TestSSAAnalyzer_InitFunctionReachability tests that methods reachable from init functions.
// are correctly marked as used
func TestSSAAnalyzer_InitFunctionReachability(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedUsed   []string
		expectedUnused []string
	}{
		{
			name: "methods called from init",
			code: `
package main

var registry map[string]func()

type Initializer struct{}

func (i *Initializer) Setup() {
	i.registerHandlers()
}

func (i *Initializer) registerHandlers() {
	registry = make(map[string]func())
}

func (i *Initializer) Cleanup() {
	// not called
}

func init() {
	i := &Initializer{}
	i.Setup()
}

func main() {
	// main doesn't call anything
}
`,
			expectedUsed:   []string{"init", "main", "Setup", "registerHandlers"},
			expectedUnused: []string{"Cleanup"},
		},
		{
			name: "multiple init functions",
			code: `
package main

type Config struct{}

func (c *Config) Load() {
	c.validate()
}

func (c *Config) validate() {
	// validation
}

func (c *Config) Save() {
	// not called
}

func init() {
	c := &Config{}
	c.Load()
}

func init() {
	// Another init that doesn't call anything.
}

func main() {}
`,
			expectedUsed:   []string{"init", "main", "Load", "validate"},
			expectedUnused: []string{"Save"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test code.
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			pkg := &packages.Package{
				ID:         "test",
				Name:       "main",
				PkgPath:    "test",
				Syntax:     []*ast.File{file},
				Fset:       fset,
				TypesSizes: gotypes.SizesFor("gc", "amd64"),
			}

			// Type check.
			conf := gotypes.Config{Importer: importer.Default()}
			info := &gotypes.Info{
				Types:      make(map[ast.Expr]gotypes.TypeAndValue),
				Defs:       make(map[*ast.Ident]gotypes.Object),
				Uses:       make(map[*ast.Ident]gotypes.Object),
				Selections: make(map[*ast.SelectorExpr]*gotypes.Selection),
				Implicits:  make(map[ast.Node]gotypes.Object),
			}
			pkg.TypesInfo = info
			pkg.Types, err = conf.Check("test", fset, []*ast.File{file}, info)
			require.NoError(t, err)

			analyzer, err := NewAnalyzer([]*packages.Package{pkg})
			require.NoError(t, err)

			// Build methods map.
			methods := make(map[gotypes.Object]*analysis.FuncInfo)
			for _, obj := range info.Defs {
				if obj == nil {
					continue
				}
				if fn, ok := obj.(*gotypes.Func); ok {
					// Skip interface methods - we only want concrete implementations.
					if sig, ok := fn.Type().(*gotypes.Signature); ok && sig.Recv() != nil {
						if _, isInterface := sig.Recv().Type().Underlying().(*gotypes.Interface); isInterface {
							continue
						}
					}
					methods[fn] = analysis.NewFuncInfo(fn, pkg, analyzer.nameCache)
				}
			}

			// Analyze.
			err = analyzer.AnalyzeFuncs(methods)
			require.NoError(t, err)

			// Check results - handling multiple init functions.
			for _, methodName := range tt.expectedUsed {
				found := false
				for obj, info := range methods {
					if obj.Name() == methodName {
						assert.True(t, info.IsUsed, "Method %s should be marked as used", methodName)
						found = true
						// Don't break for init functions as there might be multiple.
						if methodName != "init" {
							break
						}
					}
				}
				assert.True(t, found, "Expected method %s not found", methodName)
			}

			for _, methodName := range tt.expectedUnused {
				for obj, info := range methods {
					if obj.Name() == methodName {
						assert.False(t, info.IsUsed, "Method %s should NOT be marked as used", methodName)
					}
				}
			}
		})
	}
}
