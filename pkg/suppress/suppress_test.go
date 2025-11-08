package suppress

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuppressionChecker_NewChecker(t *testing.T) {
	checker := NewChecker()

	require.NotNil(t, checker, "NewSuppressionChecker returned nil")
	require.NotNil(t, checker.suppressions, "Expected suppressions map to be initialized")
}

func TestSuppressionChecker_ParseComment(t *testing.T) {
	tests := []struct {
		name           string
		comment        string
		expectedType   SuppressionType
		expectedReason string
		expectParsed   bool
	}{
		{
			name:           "nolint basic",
			comment:        "//nolint:unusedfunc",
			expectedType:   SuppressionNolint,
			expectedReason: "",
			expectParsed:   true,
		},
		{
			name:           "nolint with reason",
			comment:        "//nolint:unusedfunc // this is required for interface compliance",
			expectedType:   SuppressionNolint,
			expectedReason: "this is required for interface compliance",
			expectParsed:   true,
		},
		{
			name:           "lint ignore basic",
			comment:        "//lint:ignore unusedfunc reason here",
			expectedType:   SuppressionLintIgnore,
			expectedReason: "reason here",
			expectParsed:   true,
		},
		{
			name:           "lint ignore multi-word reason",
			comment:        "//lint:ignore unusedfunc needed for backward compatibility",
			expectedType:   SuppressionLintIgnore,
			expectedReason: "needed for backward compatibility",
			expectParsed:   true,
		},
		{
			name:           "generic nolint",
			comment:        "//nolint",
			expectedType:   SuppressionNolint,
			expectedReason: "",
			expectParsed:   true,
		},
		{
			name:           "nolint with multiple rules",
			comment:        "//nolint:unusedfunc,deadcode",
			expectedType:   SuppressionNolint,
			expectedReason: "",
			expectParsed:   true,
		},
		{
			name:           "unrelated comment",
			comment:        "// regular comment",
			expectedType:   SuppressionNolint,
			expectedReason: "",
			expectParsed:   false,
		},
		{
			name:           "nolint different rule",
			comment:        "//nolint:deadcode",
			expectedType:   SuppressionNolint,
			expectedReason: "",
			expectParsed:   false,
		},
		{
			name:           "malformed lint ignore",
			comment:        "//lint:ignore",
			expectedType:   SuppressionLintIgnore,
			expectedReason: "",
			expectParsed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewChecker()
			comment := &ast.Comment{Text: tt.comment}
			suppression := checker.parseComment(comment)

			if tt.expectParsed {
				require.NotNil(t, suppression, "Expected suppression to be parsed, got nil")
				require.Equal(t, tt.expectedType, suppression.Type)
				require.Equal(t, tt.expectedReason, suppression.Reason)
			} else {
				require.Nil(t, suppression, "Expected no suppression, got %v", suppression)
			}
		})
	}
}

func TestSuppressionChecker_Load(t *testing.T) {
	tests := []struct {
		name          string
		sourceCode    string
		expectedCount int
	}{
		{
			name: "method with nolint suppression",
			sourceCode: `package test

type TestStruct struct{}

//nolint:unusedfunc
func (t *TestStruct) UnusedMethod() {
	// This method is suppressed.
}

func (t *TestStruct) UsedMethod() {
	// This method is not suppressed.
}`,
			expectedCount: 1,
		},
		{
			name: "function with lint:ignore suppression",
			sourceCode: `package test

//lint:ignore unusedfunc required for interface
func InterfaceFunction() {
	// This function is required for interface compliance.
}`,
			expectedCount: 1,
		},
		{
			name: "multiple functions with different suppressions",
			sourceCode: `package test

//nolint:unusedfunc
func Function1() {}

//lint:ignore unusedfunc legacy support
func Function2() {}

func Function3() {
	// No suppression.
}`,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.sourceCode, parser.ParseComments)
			require.NoError(t, err, "Failed to parse source")

			checker := NewChecker()
			err = checker.Load(fset, []*ast.File{file})
			require.NoError(t, err, "Failed to load suppressions")

			suppressions := checker.getAllSuppressions()
			require.Len(t, suppressions, tt.expectedCount, "Expected %d suppressions", tt.expectedCount)
		})
	}
}

// TestSuppressionChecker_IsSuppressed tests suppression checking with reasons.
func TestSuppressionChecker_IsSuppressed(t *testing.T) {
	sourceCode := `package test

//nolint:unusedfunc
func SuppressedFunction() {}

func RegularFunction() {}

//lint:ignore unusedfunc test reason
func AnotherSuppressedFunction() {}

//nolint:unusedfunc
func Method1() {}

//lint:ignore unusedfunc this is a detailed reason
func Method2() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", sourceCode, parser.ParseComments)
	require.NoError(t, err, "Failed to parse source")

	checker := NewChecker()
	err = checker.Load(fset, []*ast.File{file})
	require.NoError(t, err, "Failed to load suppressions")

	functionPositions := make(map[string]token.Pos)
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			functionPositions[fn.Name.Name] = fn.Name.Pos()
		}
		return true
	})

	tests := []struct {
		funcName         string
		expectSuppressed bool
		expectReason     string
	}{
		{"SuppressedFunction", true, "suppressed"},
		{"RegularFunction", false, ""},
		{"AnotherSuppressedFunction", true, "test reason"},
		{"Method1", true, "suppressed"},
		{"Method2", true, "this is a detailed reason"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			pos := functionPositions[tt.funcName]
			require.NotEqual(t, token.NoPos, pos, "Could not find position for %s", tt.funcName)

			suppressed, reason := checker.IsSuppressed(pos)
			require.Equal(t, tt.expectSuppressed, suppressed, "Expected suppressed=%v for %s, got %v", tt.expectSuppressed, tt.funcName, suppressed)
			require.Equal(t, tt.expectReason, reason, "Expected reason %q for %s, got %q", tt.expectReason, tt.funcName, reason)
		})
	}

	// Test non-existent position.
	t.Run("InvalidPosition", func(t *testing.T) {
		suppressed, reason := checker.IsSuppressed(token.NoPos)
		require.False(t, suppressed, "Expected no suppression for invalid position")
		require.Empty(t, reason, "Expected empty reason for invalid position")
	})
}

// TestSuppressionChecker_Clear tests clearing suppressions.
func TestSuppressionChecker_Clear(t *testing.T) {
	sourceCode := `package test

//nolint:unusedfunc
func TestMethod() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", sourceCode, parser.ParseComments)
	require.NoError(t, err, "Failed to parse source")

	checker := NewChecker()
	err = checker.Load(fset, []*ast.File{file})
	require.NoError(t, err, "Failed to load suppressions")

	// Verify suppressions were loaded.
	suppressions := checker.getAllSuppressions()
	require.NotEmpty(t, suppressions, "Expected suppressions to be loaded")

	// Clear suppressions.
	checker.Clear()

	// Verify suppressions are cleared.
	suppressions = checker.getAllSuppressions()
	require.Empty(t, suppressions, "Expected suppressions to be cleared")
}

// TestSuppressionChecker_EdgeCases tests edge cases and error conditions.
func TestSuppressionChecker_EdgeCases(t *testing.T) {
	checker := NewChecker()

	// Test with nil file set.
	err := checker.Load(nil, []*ast.File{})
	require.Error(t, err, "Expected error with nil file set")

	// Test with nil files slice.
	fset := token.NewFileSet()
	err = checker.Load(fset, nil)
	require.Error(t, err, "Expected error with nil files slice")

	// Test IsSuppressed with invalid position.
	suppressed, _ := checker.IsSuppressed(token.NoPos)
	require.False(t, suppressed, "Expected no suppression for invalid position")

	// Test IsSuppressed with invalid position returns empty reason.
	_, reason := checker.IsSuppressed(token.NoPos)
	require.Empty(t, reason, "Expected empty reason for invalid position")
}

// TestSuppressionChecker_DirectSuppressionOnly tests that suppressions only apply to immediately following functions.
func TestSuppressionChecker_DirectSuppressionOnly(t *testing.T) {
	sourceCode := `package test

//nolint:unusedfunc
func SuppressedFunction() {}

// This function should NOT be suppressed by the above comment.
func NotSuppressedFunction() {}

//lint:ignore unusedfunc test
func AnotherSuppressedFunction() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", sourceCode, parser.ParseComments)
	require.NoError(t, err, "Failed to parse source")

	checker := NewChecker()
	err = checker.Load(fset, []*ast.File{file})
	require.NoError(t, err, "Failed to load suppressions")

	functionPositions := make(map[string]token.Pos)
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			functionPositions[fn.Name.Name] = fn.Name.Pos()
		}
		return true
	})

	// Verify only directly suppressed functions are suppressed.
	suppressed, _ := checker.IsSuppressed(functionPositions["SuppressedFunction"])
	require.True(t, suppressed, "Expected SuppressedFunction to be suppressed")

	suppressed, _ = checker.IsSuppressed(functionPositions["NotSuppressedFunction"])
	require.False(t, suppressed, "Expected NotSuppressedFunction to NOT be suppressed")

	suppressed, _ = checker.IsSuppressed(functionPositions["AnotherSuppressedFunction"])
	require.True(t, suppressed, "Expected AnotherSuppressedFunction to be suppressed")
}
