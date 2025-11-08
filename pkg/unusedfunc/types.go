// Package unusedfunc provides unused function/method analysis.
package unusedfunc

import "go/token"

// UnusedFunction represents a function that should be reported as unused.
type UnusedFunction struct {
	Name       string         `json:"name"`
	Position   token.Position `json:"position"`
	Reason     string         `json:"reason"`
	Suppressed bool           `json:"suppressed"`
	Package    string         `json:"package"`
}
