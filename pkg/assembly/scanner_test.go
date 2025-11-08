package assembly

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanner_scanReader(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantImpl   map[string]struct{}
		wantCalled map[string]struct{}
	}{
		{
			name: "basic_assembly_file",
			input: `// +build amd64

#include "textflag.h"

// func Add(x, y int) int
TEXT ·Add(SB), NOSPLIT, $0-24
    MOVQ x+0(FP), AX
    MOVQ y+8(FP), BX
    ADDQ BX, AX
    MOVQ AX, ret+16(FP)
    RET

// func Subtract(x, y int) int
TEXT ·Subtract(SB), NOSPLIT, $0-24
    MOVQ x+0(FP), AX
    MOVQ y+8(FP), BX
    SUBQ BX, AX
    MOVQ AX, ret+16(FP)
    RET`,
			wantImpl: map[string]struct{}{
				"Add":      {},
				"Subtract": {},
			},
			wantCalled: map[string]struct{}{},
		},
		{
			name: "assembly_calling_go_functions",
			input: `TEXT ·asmCallsGo(SB), $24-0
    // Prepare to call helperFunc.
    MOVQ $10, AX
    MOVQ AX, 0(SP)
    CALL ·helperFunc(SB)
    
    // Use the result.
    MOVQ 8(SP), AX
    
    // Call asmCallback.
    CALL ·asmCallback(SB)
    RET`,
			wantImpl: map[string]struct{}{
				"asmCallsGo": {},
			},
			wantCalled: map[string]struct{}{
				"helperFunc":  {},
				"asmCallback": {},
			},
		},
		{
			name: "mixed_implementations_and_calls",
			input: `// Complex assembly file

TEXT ·VectorAdd(SB), NOSPLIT, $0-72
    // Implementation details...
    RET

TEXT ·MatrixMultiply(SB), NOSPLIT, $0
    // Call helper functions.
    CALL ·dotProduct(SB)
    CALL ·transpose(SB)
    RET

// Comment line.
TEXT ·helperRoutine(SB), $0
    CALL ·internalHelper(SB)
    RET`,
			wantImpl: map[string]struct{}{
				"VectorAdd":      {},
				"MatrixMultiply": {},
				"helperRoutine":  {},
			},
			wantCalled: map[string]struct{}{
				"dotProduct":     {},
				"transpose":      {},
				"internalHelper": {},
			},
		},
		{
			name: "ignore_comments",
			input: `// TEXT ·commentedOut(SB), NOSPLIT, $0
// This is just a comment.
// CALL ·notReally(SB)

TEXT ·realFunction(SB), $0
    // But this is real:
    CALL ·actualCall(SB)
    RET`,
			wantImpl: map[string]struct{}{
				"realFunction": {},
			},
			wantCalled: map[string]struct{}{
				"actualCall": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &Info{
				ImplementedFunctions: make(map[string]struct{}),
				CalledFunctions:      make(map[string]struct{}),
			}

			err := scanReader(strings.NewReader(tt.input), info)
			require.NoError(t, err)

			require.Equal(t, tt.wantImpl, info.ImplementedFunctions)
			require.Equal(t, tt.wantCalled, info.CalledFunctions)
		})
	}
}
