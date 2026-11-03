// dotq4_arm64.s — NEON Q4 dot product kernel (plan9 syntax + WORD macros)
//
// func DotQ4(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
// func LinearQ4(y, x, w, bias unsafe.Pointer, seqLen, inDim, outDim int)
//
// NEON instructions not in Go's assembler, encoded as WORD macros:
//   USHLL  Vd.8H, Vn.8B, #0 (UXTL)  = 0x2F08A400 | (Vn<<5) | Vd
//   USHLL2 Vd.8H, Vn.16B, #0 (UXTL2) = 0x6F08A400 | (Vn<<5) | Vd
//   USHLL  Vd.4S, Vn.4H, #0 (UXTL)  = 0x2F10A400 | (Vn<<5) | Vd
//   USHLL2 Vd.4S, Vn.8H, #0 (UXTL2) = 0x6F10A400 | (Vn<<5) | Vd
//   UCVTF  Vd.4S, Vn.4S             = 0x6E21D800 | (Vn<<5) | Vd
//   FSUB   Vd.4S, Vn.4S, Vm.4S      = 0x4EA0D400 | (Vm<<16) | (Vn<<5) | Vd
//   FMUL   Vd.4S, Vn.4S, Vm.4S      = 0x6E20DC00 | (Vm<<16) | (Vn<<5) | Vd
//
#include "textflag.h"

// UXTL Vd.8H, Vn.8B  (zero-extend 8 bytes → 8 halfwords, lower half)
#define UXTL_8B_8H(vn, vd)   WORD $(0x2F08A400 | ((vn)<<5) | (vd))
// UXTL2 Vd.8H, Vn.16B (zero-extend 8 bytes → 8 halfwords, upper half)
#define UXTL2_16B_8H(vn, vd) WORD $(0x6F08A400 | ((vn)<<5) | (vd))
// UXTL Vd.4S, Vn.4H  (zero-extend 4 halfwords → 4 words, lower half)
#define UXTL_4H_4S(vn, vd)   WORD $(0x2F10A400 | ((vn)<<5) | (vd))
// UXTL2 Vd.4S, Vn.8H (zero-extend 4 halfwords → 4 words, upper half)
#define UXTL2_8H_4S(vn, vd)  WORD $(0x6F10A400 | ((vn)<<5) | (vd))
// UCVTF Vd.4S, Vn.4S (uint32 → float32)
#define UCVTF_4S(vn, vd)     WORD $(0x6E21D800 | ((vn)<<5) | (vd))
// FSUB Vd.4S, Vn.4S, Vm.4S
#define FSUB_4S(vm, vn, vd)   WORD $(0x4EA0D400 | ((vm)<<16) | ((vn)<<5) | (vd))
// FMUL Vd.4S, Vn.4S, Vm.4S
#define FMUL_4S(vm, vn, vd)   WORD $(0x6E20DC00 | ((vm)<<16) | ((vn)<<5) | (vd))

// Register assignments for DotQ4:
//   R0 = x, R1 = blocks, R2 = nBlocks
//   V0,V1 = accumulator (8 float32)
//   V20 = [8.0f x 4]
//   V21 = [0x0F x 16] mask
//   V30 = [scale x 4]
//   V2 = loaded packed bytes
//   V3 = low nibbles, V4 = high nibbles
//   V5 = extended halfwords (temp)
//   V6 = dequantized float32 (temp)
//   V7 = loaded x values

// func DotQ4(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
TEXT ·DotQ4(SB), NOSPLIT, $0-28
    MOVD    x+0(FP), R0
    MOVD    blocks+8(FP), R1
    MOVD    nBlocks+16(FP), R2

    // Zero accumulators
    VEOR    V0.B16, V0.B16, V0.B16
    VEOR    V1.B16, V1.B16, V1.B16

    // V20 = [8.0f, 8.0f, 8.0f, 8.0f]
    FMOVS   $8.0, F20
    VDUP    V20.S[0], V20.S4

    // V21 = [0x0F x 16]
    MOVD    $0x0F0F0F0F0F0F0F0F, R3
    VMOV    R3, V21.D[0]
    VMOV    R3, V21.D[1]

    CBZ     R2, dq4_done

dq4_block:
    // Load scale, broadcast to V30
    FMOVS   (R1), F30
    VDUP    V30.S[0], V30.S4

    // Load 16 packed bytes from R1+4
    ADD     $4, R1, R3
    VLD1    (R3), [V2.B16]

    // Low nibbles: V3 = V2 & 0x0F
    VAND    V21.B16, V2.B16, V3.B16
    // High nibbles: V4 = V2 >> 4  (VUSHR is in Go asm)
    VUSHR   $4, V2.B16, V4.B16

    // === Low nibbles elements 0-3 ===
    UXTL_8B_8H(3, 5)        // V5.8H = zero-extend V3.8B (lower 8 bytes → 8 halfwords)
    UXTL_4H_4S(5, 6)        // V6.4S = zero-extend V5.4H (lower 4 halfwords → 4 words)
    UCVTF_4S(6, 6)          // V6.4S = float32(V6.4S)
    FSUB_4S(20, 6, 6)       // V6.4S = V6 - 8.0
    FMUL_4S(30, 6, 6)       // V6.4S = V6 * scale
    VLD1.P  16(R0), [V7.S4] // V7 = x[0:4], R0 += 16
    VFMLA   V6.S4, V7.S4, V0.S4  // V0 += V7 * V6

    // === Low nibbles elements 4-7 ===
    UXTL2_8H_4S(5, 6)       // V6.4S = zero-extend V5.8H upper → 4 words
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // === Low nibbles elements 8-11 ===
    // Need upper 8 bytes of V3: use UXTL2
    UXTL2_16B_8H(3, 5)      // V5.8H = zero-extend V3 upper 8 bytes → 8 halfwords
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // === Low nibbles elements 12-15 ===
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // === High nibbles elements 16-19 ===
    UXTL_8B_8H(4, 5)
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // === High nibbles elements 20-23 ===
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // === High nibbles elements 24-27 ===
    UXTL2_16B_8H(4, 5)
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // === High nibbles elements 28-31 ===
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R0), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // Advance blocks: +20
    ADD     $20, R1, R1
    SUB     $1, R2, R2
    CBNZ    R2, dq4_block

dq4_done:
    // Horizontal sum V0 + V1 → scalar
    WORD    $(0x4E21D400)        // FADD V0.4S, V0.4S, V1.4S  → V0
    // Sum 4 lanes of V0
    VMOV    V0.S[0], R3
    VMOV    V0.S[1], R4
    VMOV    V0.S[2], R5
    VMOV    V0.S[3], R6
    FMOVS   R3, F10
    FMOVS   R4, F11
    FMOVS   R5, F12
    FMOVS   R6, F13
    FADDS   F11, F10, F10
    FADDS   F12, F10, F10
    FADDS   F13, F10, F10

    FMOVS   F10, ret+24(FP)
    RET

// func LinearQ4(y, x, w, bias unsafe.Pointer, seqLen, inDim, outDim int)
TEXT ·LinearQ4(SB), NOSPLIT, $0-56
    MOVD    y+0(FP), R0
    MOVD    x+8(FP), R1          // x base
    MOVD    w+16(FP), R2         // w base
    MOVD    bias+24(FP), R3      // bias (may be 0)
    MOVD    seqLen+32(FP), R4
    MOVD    inDim+40(FP), R5
    MOVD    outDim+48(FP), R6

    // nBlocks = inDim / 32
    LSR     $5, R5, R7           // R7 = nBlocks
    // rowBytes = nBlocks * 20
    MOVD    $20, R8
    MUL     R7, R8, R8           // R8 = rowBytes

    // V21 = [0x0F x 16] mask
    MOVD    $0x0F0F0F0F0F0F0F0F, R9
    VMOV    R9, V21.D[0]
    VMOV    R9, V21.D[1]
    // V20 = [8.0f x 4]
    FMOVS   $8.0, F20
    VDUP    V20.S[0], V20.S4

    CBZ     R4, lq4_done

lq4_seq:
    MOVD    R6, R10              // outDim counter
    MOVD    R2, R11              // current w row
    MOVD    R0, R12              // current y output

lq4_out:
    MOVD    R1, R13              // x_row
    MOVD    R11, R14             // w_row
    MOVD    R7, R15              // block counter

    VEOR    V0.B16, V0.B16, V0.B16
    VEOR    V1.B16, V1.B16, V1.B16

lq4_blk:
    FMOVS   (R14), F30
    VDUP    V30.S[0], V30.S4
    ADD     $4, R14, R16
    VLD1    (R16), [V2.B16]
    VAND    V21.B16, V2.B16, V3.B16
    VUSHR   $4, V2.B16, V4.B16

    // Low 0-3
    UXTL_8B_8H(3, 5)
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // Low 4-7
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // Low 8-11
    UXTL2_16B_8H(3, 5)
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // Low 12-15
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // High 16-19
    UXTL_8B_8H(4, 5)
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // High 20-23
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    // High 24-27
    UXTL2_16B_8H(4, 5)
    UXTL_4H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V0.S4

    // High 28-31
    UXTL2_8H_4S(5, 6)
    UCVTF_4S(6, 6)
    FSUB_4S(20, 6, 6)
    FMUL_4S(30, 6, 6)
    VLD1.P  16(R13), [V7.S4]
    VFMLA   V6.S4, V7.S4, V1.S4

    ADD     $20, R14, R14
    SUB     $1, R15, R15
    CBNZ    R15, lq4_blk

    // Horizontal sum
    WORD    $(0x4E21D400)        // FADD V0.4S, V0.4S, V1.4S → V0
    VMOV    V0.S[0], R16
    VMOV    V0.S[1], R17
    VMOV    V0.S[2], R19
    VMOV    V0.S[3], R20
    FMOVS   R16, F10
    FMOVS   R17, F11
    FMOVS   R19, F12
    FMOVS   R20, F13
    FADDS   F11, F10, F10
    FADDS   F12, F10, F10
    FADDS   F13, F10, F10

    // Add bias
    CBZ     R3, lq4_nobias
    SUB     R10, R6, R16
    LSL     $2, R16, R16
    ADD     R3, R16, R16
    FMOVS   (R16), F11
    FADDS   F11, F10, F10

lq4_nobias:
    FMOVS   F10, (R12)
    ADD     $4, R12, R12
    ADD     R8, R11, R11
    SUB     $1, R10, R10
    CBNZ    R10, lq4_out

    // Next seq position
    LSL     $2, R5, R16
    ADD     R16, R1, R1
    LSL     $2, R6, R16
    ADD     R16, R0, R0
    SUB     $1, R4, R4
    CBNZ    R4, lq4_seq

lq4_done:
    RET
