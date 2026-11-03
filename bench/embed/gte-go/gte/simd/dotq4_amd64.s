// dotq4_amd64.s — AVX2+FMA Q4 dot product kernel
//
// func DotQ4(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
//
// Per block (32 elements, 20 bytes):
//   scale = float32 at blocks[0:4]
//   qs[0:16] packed nibbles
//   low nibble qs[i]&0xF → element i (0..15)
//   high nibble qs[i]>>4 → element i+16 (16..31)
//   dequant = scale * (nibble - 8)
//
// AVX2 approach: 4 groups of 8 floats per block.
//   VPMOVZXBD ymm, xmm[63:0] — zero-extends 8 bytes to 8 dwords
//   VCVTDQ2PS — int32 to float32
//   VFMADD231PS — fused multiply-add accumulate
//
#include "textflag.h"

DATA const_8<>+0x00(SB)/4, $0x41000000  // 8.0f
DATA const_8<>+0x04(SB)/4, $0x41000000
DATA const_8<>+0x08(SB)/4, $0x41000000
DATA const_8<>+0x0c(SB)/4, $0x41000000
DATA const_8<>+0x10(SB)/4, $0x41000000
DATA const_8<>+0x14(SB)/4, $0x41000000
DATA const_8<>+0x18(SB)/4, $0x41000000
DATA const_8<>+0x1c(SB)/4, $0x41000000
GLOBL const_8<>(SB), (RODATA+NOPTR), $32

DATA const_mask0f<>+0x00(SB)/1, $0x0f
DATA const_mask0f<>+0x01(SB)/1, $0x0f
DATA const_mask0f<>+0x02(SB)/1, $0x0f
DATA const_mask0f<>+0x03(SB)/1, $0x0f
DATA const_mask0f<>+0x04(SB)/1, $0x0f
DATA const_mask0f<>+0x05(SB)/1, $0x0f
DATA const_mask0f<>+0x06(SB)/1, $0x0f
DATA const_mask0f<>+0x07(SB)/1, $0x0f
DATA const_mask0f<>+0x08(SB)/1, $0x0f
DATA const_mask0f<>+0x09(SB)/1, $0x0f
DATA const_mask0f<>+0x0a(SB)/1, $0x0f
DATA const_mask0f<>+0x0b(SB)/1, $0x0f
DATA const_mask0f<>+0x0c(SB)/1, $0x0f
DATA const_mask0f<>+0x0d(SB)/1, $0x0f
DATA const_mask0f<>+0x0e(SB)/1, $0x0f
DATA const_mask0f<>+0x0f(SB)/1, $0x0f
GLOBL const_mask0f<>(SB), (RODATA+NOPTR), $16

// func DotQ4(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
TEXT ·DotQ4(SB), NOSPLIT, $0-28
    MOVQ    x+0(FP), SI          // SI = x pointer
    MOVQ    blocks+8(FP), DI     // DI = blocks pointer
    MOVQ    nBlocks+16(FP), CX   // CX = nBlocks

    // Zero accumulator
    VXORPS  Y0, Y0, Y0           // Y0 = sum accumulator

    // Load constants
    VMOVDQU const_8<>(SB), Y14   // Y14 = [8.0, 8.0, ...]
    VMOVDQU const_mask0f<>(SB), X15  // X15 = [0x0F x 16]

    TESTQ   CX, CX
    JZ      dotq4_done

dotq4_block:
    // Load scale and broadcast
    VBROADCASTSS (DI), Y1        // Y1 = scale broadcast

    // Load 16 packed bytes
    VMOVDQU 4(DI), X2            // X2 = qs[0:16]

    // Low nibbles: X3 = X2 & 0x0F
    VPAND   X15, X2, X3

    // High nibbles: X4 = X2 >> 4
    VPSRLW  $4, X2, X4
    VPAND   X15, X4, X4

    // === Process low nibbles: elements 0-7 (X3[0:8]) ===
    VPMOVZXBD X3, Y5             // Y5 = zero-extend first 8 bytes to 8 dwords
    VCVTDQ2PS Y5, Y5             // Y5 = float32
    VSUBPS  Y14, Y5, Y5          // Y5 -= 8.0
    VMULPS  Y1, Y5, Y5           // Y5 *= scale
    VFMADD231PS (SI), Y5, Y0     // Y0 += Y5 * x[0:8]

    // === Process low nibbles: elements 8-15 (X3[8:16]) ===
    VPSRLDQ $8, X3, X6           // shift X3 right 8 bytes
    VPMOVZXBD X6, Y6             // Y6 = zero-extend bytes 8-15 to dwords
    VCVTDQ2PS Y6, Y6
    VSUBPS  Y14, Y6, Y6
    VMULPS  Y1, Y6, Y6
    VFMADD231PS 32(SI), Y6, Y0   // Y0 += Y6 * x[8:16]

    // === Process high nibbles: elements 16-23 (X4[0:8]) ===
    VPMOVZXBD X4, Y7             // Y7 = zero-extend first 8 bytes to dwords
    VCVTDQ2PS Y7, Y7
    VSUBPS  Y14, Y7, Y7
    VMULPS  Y1, Y7, Y7
    VFMADD231PS 64(SI), Y7, Y0   // Y0 += Y7 * x[16:24]

    // === Process high nibbles: elements 24-31 (X4[8:16]) ===
    VPSRLDQ $8, X4, X8           // shift X4 right 8 bytes
    VPMOVZXBD X8, Y8             // Y8 = zero-extend bytes 8-15 to dwords
    VCVTDQ2PS Y8, Y8
    VSUBPS  Y14, Y8, Y8
    VMULPS  Y1, Y8, Y8
    VFMADD231PS 96(SI), Y8, Y0   // Y0 += Y8 * x[24:32]

    // Advance pointers
    ADDQ    $20, DI              // blocks += 20 (BlockQ4Size)
    ADDQ    $128, SI             // x += 32 floats = 128 bytes
    DECQ    CX
    JNZ     dotq4_block

dotq4_done:
    // Horizontal sum Y0 → scalar
    VEXTRACTF128 $1, Y0, X1
    VADDPS  X1, X0, X0
    MOVHLPS X0, X1
    ADDPS   X1, X0
    MOVSS   X0, X1
    SHUFPS  $0x55, X0, X0
    ADDSS   X1, X0

    MOVSS   X0, ret+24(FP)
    VZEROUPPER
    RET

// func LinearQ4(y, x, w, bias unsafe.Pointer, seqLen, inDim, outDim int)
TEXT ·LinearQ4(SB), NOSPLIT, $0-56
    MOVQ    y+0(FP), R8          // R8 = y
    MOVQ    x+8(FP), R9          // R9 = x
    MOVQ    w+16(FP), R10        // R10 = w
    MOVQ    bias+24(FP), R11     // R11 = bias (may be 0)
    MOVQ    seqLen+32(FP), R12   // R12 = seqLen
    MOVQ    inDim+40(FP), R13    // R13 = inDim
    MOVQ    outDim+48(FP), R14   // R14 = outDim

    // nBlocks = inDim / 32
    MOVQ    R13, R15
    SHRQ    $5, R15              // R15 = nBlocks

    // rowBytes = nBlocks * 20
    MOVQ    R15, BX
    IMULQ   $20, BX        // BX = rowBytes per weight row

    // Load constants
    VMOVDQU const_8<>(SB), Y14
    VMOVDQU const_mask0f<>(SB), X15

    TESTQ   R12, R12
    JZ      linq4_done

linq4_seq:
    // For each sequence position
    MOVQ    R14, CX              // CX = outDim counter
    MOVQ    R10, DI              // DI = current weight row pointer
    MOVQ    R8, DX               // DX = current y output pointer

linq4_out:
    // Compute dot product: DotQ4(x_row, w_row, nBlocks)
    MOVQ    R9, SI               // SI = x_row pointer
    MOVQ    DI, AX               // AX = w_row pointer
    MOVQ    R15, BP              // BP = nBlocks counter
    VXORPS  Y0, Y0, Y0          // Y0 = accumulator

linq4_block:
    VBROADCASTSS (AX), Y1
    VMOVDQU 4(AX), X2
    VPAND   X15, X2, X3
    VPSRLW  $4, X2, X4
    VPAND   X15, X4, X4

    VPMOVZXBD X3, Y5
    VCVTDQ2PS Y5, Y5
    VSUBPS  Y14, Y5, Y5
    VMULPS  Y1, Y5, Y5
    VFMADD231PS (SI), Y5, Y0

    VPSRLDQ $8, X3, X6
    VPMOVZXBD X6, Y6
    VCVTDQ2PS Y6, Y6
    VSUBPS  Y14, Y6, Y6
    VMULPS  Y1, Y6, Y6
    VFMADD231PS 32(SI), Y6, Y0

    VPMOVZXBD X4, Y7
    VCVTDQ2PS Y7, Y7
    VSUBPS  Y14, Y7, Y7
    VMULPS  Y1, Y7, Y7
    VFMADD231PS 64(SI), Y7, Y0

    VPSRLDQ $8, X4, X8
    VPMOVZXBD X8, Y8
    VCVTDQ2PS Y8, Y8
    VSUBPS  Y14, Y8, Y8
    VMULPS  Y1, Y8, Y8
    VFMADD231PS 96(SI), Y8, Y0

    ADDQ    $20, AX
    ADDQ    $128, SI
    DECQ    BP
    JNZ     linq4_block

    // Horizontal sum Y0 → X0
    VEXTRACTF128 $1, Y0, X1
    VADDPS  X1, X0, X0
    MOVHLPS X0, X1
    ADDPS   X1, X0
    MOVSS   X0, X1
    SHUFPS  $0x55, X0, X0
    ADDSS   X1, X0

    // Add bias if present
    TESTQ   R11, R11
    JZ      linq4_nobias
    // bias offset = (outDim - CX) * 4
    MOVQ    R14, BP
    SUBQ    CX, BP
    ADDSS   (R11)(BP*4), X0

linq4_nobias:
    MOVSS   X0, (DX)            // store y[o]
    ADDQ    $4, DX               // y++
    ADDQ    BX, DI               // w += rowBytes
    DECQ    CX
    JNZ     linq4_out

    // Advance to next sequence position
    MOVQ    R13, AX
    SHLQ    $2, AX               // AX = inDim * 4
    ADDQ    AX, R9               // x += inDim floats
    MOVQ    R14, AX
    SHLQ    $2, AX               // AX = outDim * 4
    ADDQ    AX, R8               // y_base += outDim floats
    DECQ    R12
    JNZ     linq4_seq

linq4_done:
    VZEROUPPER
    RET
