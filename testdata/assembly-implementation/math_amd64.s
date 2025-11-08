// +build amd64

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
    RET


// func VectorAdd(dst, a, b []float64)
TEXT ·VectorAdd(SB), NOSPLIT, $0-72
    // Simple implementation - real SIMD would use vector instructions
    MOVQ dst_base+0(FP), DI
    MOVQ a_base+24(FP), SI
    MOVQ b_base+48(FP), DX
    MOVQ dst_len+8(FP), CX
    
    // Check if length is 0
    TESTQ CX, CX
    JZ done

loop:
    // Load values
    MOVSD (SI), X0
    MOVSD (DX), X1
    
    // Add
    ADDSD X1, X0
    
    // Store result
    MOVSD X0, (DI)
    
    // Advance pointers
    ADDQ $8, DI
    ADDQ $8, SI
    ADDQ $8, DX
    
    // Decrement counter and loop
    DECQ CX
    JNZ loop

done:
    RET

// This assembly function calls back to Go
TEXT ·asmCallsGo(SB), $24-0
    // Prepare to call helperFunc
    MOVQ $10, AX
    MOVQ AX, 0(SP)
    CALL ·helperFunc(SB)
    
    // Use the result
    MOVQ 8(SP), AX
    
    // Prepare data for callback
    MOVQ $16, AX
    MOVQ AX, 0(SP)  // len
    MOVQ AX, 8(SP)  // cap
    LEAQ 16(SP), AX
    MOVQ AX, 16(SP) // data pointer
    
    // Call asmCallback
    CALL ·asmCallback(SB)
    RET
