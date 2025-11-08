// Package suppress implements comment-based suppression of linter findings.
package suppress

import (
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"regexp"
	"strings"
)

// Checker handles nolint and lint:ignore comment suppression.
type Checker struct {
	// suppressions maps position to suppression reason
	suppressions map[token.Pos]string

	// fset is the file set for position calculations
	fset *token.FileSet
}

// Suppression represents a parsed suppression directive.
type Suppression struct {
	Position token.Pos
	Reason   string
	Type     SuppressionType
}

// SuppressionType represents different types of suppression comments.
type SuppressionType int

const (
	// SuppressionNolint represents //nolint:unusedfunc comments.
	SuppressionNolint SuppressionType = iota

	// SuppressionLintIgnore represents //lint:ignore unusedfunc comments.
	SuppressionLintIgnore
)

// Suppression patterns for different comment styles.
var (
	// nolintPattern matches //nolint:unusedfunc comments
	nolintPattern = regexp.MustCompile(`//\s*nolint:unusedfunc(?:\s+//\s*(.+))?`)

	// lintIgnorePattern matches //lint:ignore unusedfunc comments
	lintIgnorePattern = regexp.MustCompile(`//\s*lint:ignore\s+unusedfunc(?:\s+(.+))?`)

	// genericNolintPattern matches //nolint comments without specific linter
	genericNolintPattern = regexp.MustCompile(`//\s*nolint(?:\s|$)`)

	// nolintWithMultipleRules matches nolint with multiple comma-separated rules
	nolintWithMultipleRules = regexp.MustCompile(`//\s*nolint:([^/]+)`)
)

// NewChecker creates a new suppression checker.
func NewChecker() *Checker {
	return &Checker{
		suppressions: make(map[token.Pos]string),
	}
}

// Load parses suppression comments from AST files.
func (sc *Checker) Load(fset *token.FileSet, files []*ast.File) error {
	if fset == nil {
		return fmt.Errorf("fset cannot be nil")
	}
	if files == nil {
		return fmt.Errorf("files cannot be nil")
	}
	sc.fset = fset

	for _, file := range files {
		suppressionsByLine := make(map[int]*Suppression)

		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				if suppression := sc.parseComment(comment); suppression != nil {
					pos := fset.Position(comment.Pos())
					suppressionsByLine[pos.Line] = suppression
				}
			}
		}

		// Second pass: find functions/methods and check if they have a suppression on the line before.
		ast.Inspect(file, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				// Use the function name position to match what types.Object.Pos() returns.
				funcPos := funcDecl.Name.Pos()
				funcPosInfo := fset.Position(funcPos)

				// Check if there's a suppression on the line immediately before this function.
				// or on the same line as the function (Go standard behavior).
				var suppression *Suppression
				var exists bool

				if suppression, exists = suppressionsByLine[funcPosInfo.Line-1]; !exists {
					suppression, exists = suppressionsByLine[funcPosInfo.Line]
				}

				if exists {
					reason := suppression.Reason
					if reason == "" {
						reason = "suppressed"
					}
					sc.suppressions[funcPos] = reason
				}
			}
			return true
		})
	}

	return nil
}

// parseComment parses a comment to check if it's a suppression directive.
func (sc *Checker) parseComment(comment *ast.Comment) *Suppression {
	text := comment.Text

	if matches := nolintPattern.FindStringSubmatch(text); matches != nil {
		reason := ""
		if len(matches) > 1 && matches[1] != "" {
			reason = strings.TrimSpace(matches[1])
		}
		return &Suppression{
			Position: comment.Pos(),
			Reason:   reason,
			Type:     SuppressionNolint,
		}
	}

	if matches := lintIgnorePattern.FindStringSubmatch(text); matches != nil {
		reason := ""
		if len(matches) > 1 && matches[1] != "" {
			reason = strings.TrimSpace(matches[1])
		}
		return &Suppression{
			Position: comment.Pos(),
			Reason:   reason,
			Type:     SuppressionLintIgnore,
		}
	}

	if genericNolintPattern.MatchString(text) {
		return &Suppression{
			Position: comment.Pos(),
			Reason:   "",
			Type:     SuppressionNolint,
		}
	}

	if matches := nolintWithMultipleRules.FindStringSubmatch(text); len(matches) > 1 {
		for rule := range strings.SplitSeq(matches[1], ",") {
			rule = strings.TrimSpace(rule)
			if rule == "unusedfunc" {
				// Extract reason if present.
				reason := ""
				if idx := strings.Index(text, "//"); idx > 0 {
					afterComment := text[idx+2:]
					if afterIdx := strings.Index(afterComment, "//"); afterIdx >= 0 {
						reason = strings.TrimSpace(afterComment[afterIdx+2:])
					}
				}
				return &Suppression{
					Position: comment.Pos(),
					Reason:   reason,
					Type:     SuppressionNolint,
				}
			}
		}
	}

	return nil
}

// IsSuppressed checks if a function at the given position is suppressed.
func (sc *Checker) IsSuppressed(pos token.Pos) (bool, string) {
	// Simple direct check - no complex nearby logic.
	if reason, exists := sc.suppressions[pos]; exists {
		return true, reason
	}
	return false, ""
}

// Clear clears all suppressions.
func (sc *Checker) Clear() {
	sc.suppressions = make(map[token.Pos]string)
}

func (sc *Checker) getAllSuppressions() map[token.Pos]string {
	result := make(map[token.Pos]string)
	maps.Copy(result, sc.suppressions)
	return result
}
