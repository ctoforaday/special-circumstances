// dotq4_int_amd64.s — AVX2 integer-MAC Q4 dot product
//
// func DotQ4Int(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
//
// Per block (32 elements):
//   1. Find absmax of x[0:32], compute inv_scale = 127/absmax
//   2. Quantize x to int8, offset to unsigned (+128)
//   3. Expand Q4 nibbles to signed int8 (nibble - 8)
//   4. VPMADDUBSW + VPMADDWD → 8 int32 partial sums
//   5. Hsum int32, correct for offset: dot -= 128 * sum(w_s8)
//   6. Accumulate: total += (w_scale * absmax / 127) * float(dot)
//
#include "textflag.h"

DATA di_ones16<>+0x00(SB)/4, $0x00010001
DATA di_ones16<>+0x04(SB)/4, $0x00010001
DATA di_ones16<>+0x08(SB)/4, $0x00010001
DATA di_ones16<>+0x0c(SB)/4, $0x00010001
DATA di_ones16<>+0x10(SB)/4, $0x00010001
DATA di_ones16<>+0x14(SB)/4, $0x00010001
DATA di_ones16<>+0x18(SB)/4, $0x00010001
DATA di_ones16<>+0x1c(SB)/4, $0x00010001
GLOBL di_ones16<>(SB), (RODATA+NOPTR), $32

DATA di_128<>+0x00(SB)/4, $0x80808080
DATA di_128<>+0x04(SB)/4, $0x80808080
DATA di_128<>+0x08(SB)/4, $0x80808080
DATA di_128<>+0x0c(SB)/4, $0x80808080
DATA di_128<>+0x10(SB)/4, $0x80808080
DATA di_128<>+0x14(SB)/4, $0x80808080
DATA di_128<>+0x18(SB)/4, $0x80808080
DATA di_128<>+0x1c(SB)/4, $0x80808080
GLOBL di_128<>(SB), (RODATA+NOPTR), $32

DATA di_8<>+0x00(SB)/4, $0x08080808
DATA di_8<>+0x04(SB)/4, $0x08080808
DATA di_8<>+0x08(SB)/4, $0x08080808
DATA di_8<>+0x0c(SB)/4, $0x08080808
DATA di_8<>+0x10(SB)/4, $0x08080808
DATA di_8<>+0x14(SB)/4, $0x08080808
DATA di_8<>+0x18(SB)/4, $0x08080808
DATA di_8<>+0x1c(SB)/4, $0x08080808
GLOBL di_8<>(SB), (RODATA+NOPTR), $32

DATA di_mask0f<>+0x00(SB)/4, $0x0F0F0F0F
DATA di_mask0f<>+0x04(SB)/4, $0x0F0F0F0F
DATA di_mask0f<>+0x08(SB)/4, $0x0F0F0F0F
DATA di_mask0f<>+0x0c(SB)/4, $0x0F0F0F0F
GLOBL di_mask0f<>(SB), (RODATA+NOPTR), $16

DATA di_absmask<>+0x00(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x04(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x08(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x0c(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x10(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x14(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x18(SB)/4, $0x7FFFFFFF
DATA di_absmask<>+0x1c(SB)/4, $0x7FFFFFFF
GLOBL di_absmask<>(SB), (RODATA+NOPTR), $32

DATA di_127f<>+0x00(SB)/4, $0x42FE0000
GLOBL di_127f<>(SB), (RODATA+NOPTR), $4

DATA di_inv127f<>+0x00(SB)/4, $0x3C010204
GLOBL di_inv127f<>(SB), (RODATA+NOPTR), $4

// func DotQ4Int(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
TEXT ·DotQ4Int(SB), NOSPLIT, $0-28
    MOVQ    x+0(FP), SI
    MOVQ    blocks+8(FP), DI
    MOVQ    nBlocks+16(FP), CX

    VXORPS  X0, X0, X0          // X0 = float accumulator

    VMOVDQU di_ones16<>(SB), Y14  // Y14 = ones for VPMADDWD
    VMOVDQU di_128<>(SB), Y13     // Y13 = 128 bytes for u8 offset
    VMOVDQU di_8<>(SB), Y12       // Y12 = 8 bytes for nibble centering
    VMOVDQU di_absmask<>(SB), Y11 // Y11 = abs mask
    VMOVDQU di_mask0f<>(SB), X10  // X10 = 0x0F mask (16 bytes)

    TESTQ   CX, CX
    JZ      dii_done

dii_block:
    // --- Find absmax of x[0:32] ---
    VMOVUPS (SI), Y1
    VMOVUPS 32(SI), Y2
    VMOVUPS 64(SI), Y3
    VMOVUPS 96(SI), Y4

    VANDPS  Y11, Y1, Y5
    VANDPS  Y11, Y2, Y6
    VMAXPS  Y6, Y5, Y5
    VANDPS  Y11, Y3, Y6
    VMAXPS  Y6, Y5, Y5
    VANDPS  Y11, Y4, Y6
    VMAXPS  Y6, Y5, Y5
    // Horizontal max Y5 → X5[0]
    VEXTRACTF128 $1, Y5, X6
    VMAXPS  X6, X5, X5
    MOVHLPS X5, X6
    MAXPS   X6, X5
    MOVSS   X5, X6
    SHUFPS  $0x55, X5, X5
    MAXSS   X6, X5              // X5[0] = absmax

    // Check zero
    XORPS   X6, X6
    UCOMISS X5, X6
    JE      dii_skip

    // --- inv_scale = 127 / absmax ---
    MOVSS   di_127f<>(SB), X6
    DIVSS   X5, X6              // X6 = 127/absmax
    VBROADCASTSS X6, Y6         // Y6 = inv_scale broadcast

    // --- Quantize x to int8: round(x * 127/absmax) ---
    VMULPS  Y6, Y1, Y1
    VMULPS  Y6, Y2, Y2
    VMULPS  Y6, Y3, Y3
    VMULPS  Y6, Y4, Y4
    VCVTPS2DQ Y1, Y1            // float → int32 (banker's round)
    VCVTPS2DQ Y2, Y2
    VCVTPS2DQ Y3, Y3
    VCVTPS2DQ Y4, Y4
    // Pack i32 → i16 → i8 (signed saturation)
    VPACKSSDW Y2, Y1, Y7        // Y7 = [Y1_lo16 | Y2_lo16 | Y1_hi16 | Y2_hi16]
    VPACKSSDW Y4, Y3, Y8        // Y8 = [Y3_lo16 | Y4_lo16 | Y3_hi16 | Y4_hi16]
    VPACKSSWB Y8, Y7, Y7        // Y7 = 32 × int8 (lane-interleaved)
    VPERMQ  $0xD8, Y7, Y7       // fix 64-bit lane order
    VPSHUFD $0xD8, Y7, Y7       // fix 32-bit group order within lanes

    // --- Offset to unsigned: x_u8 = x_i8 + 128 ---
    VPADDB  Y13, Y7, Y7         // Y7 = x_u8

    // --- Expand Q4 nibbles to signed int8 ---
    VMOVDQU 4(DI), X8           // X8 = 16 packed bytes
    VPAND   X10, X8, X9         // X9 = low nibbles (0-15), 16 bytes
    VPSRLW  $4, X8, X8
    VPAND   X10, X8, X8         // X8 = high nibbles (0-15), 16 bytes
    VINSERTI128 $1, X8, Y9, Y9  // Y9 = [low16 | high16] = 32 unsigned nibbles
    VPSUBB  Y12, Y9, Y9         // Y9 = w_s8 (32 signed, [-8,7])

    // --- Compute sum(w_s8) BEFORE hsum trashes X9 ---
    VEXTRACTI128 $1, Y9, X15    // high 16 bytes of w_s8
    // low 16 bytes of w_s8 = X9 (low 128 of Y9)
    VPMOVSXBW X9, Y1            // low 16 w_s8 → 16 int16
    VPMOVSXBW X15, Y2           // high 16 w_s8 → 16 int16
    VPADDW  Y2, Y1, Y1          // 16 int16
    VPMADDWD Y14, Y1, Y1        // 16×i16 → 8×i32
    VEXTRACTI128 $1, Y1, X2
    VPADDD  X2, X1, X1
    VPSHUFD $0x4E, X1, X2
    VPADDD  X2, X1, X1
    VPSHUFD $0xB1, X1, X2
    VPADDD  X2, X1, X1          // X1[0] = sum(w_s8)
    VPSLLD  $7, X1, X1          // X1[0] = 128 * sum(w_s8)
    // Save correction in X15
    MOVSS   X1, X15

    // --- Integer MAC: VPMADDUBSW + VPMADDWD ---
    VPMADDUBSW Y9, Y7, Y8       // 32×(u8×s8) → 16×i16
    VPMADDWD Y14, Y8, Y8        // 16×i16 → 8×i32 (sum pairs)

    // --- Hsum 8 int32 in Y8 ---
    VEXTRACTI128 $1, Y8, X9
    VPADDD  X9, X8, X8
    VPSHUFD $0x4E, X8, X9
    VPADDD  X9, X8, X8
    VPSHUFD $0xB1, X8, X9
    VPADDD  X9, X8, X8          // X8[0] = raw_dot

    // --- Apply correction ---
    VPSUBD  X15, X8, X8         // X8[0] = raw_dot - 128*sum(w_s8)

    // --- Descale: total += (w_scale * absmax / 127) * dot ---
    VCVTDQ2PS X8, X8            // int32 → float
    MOVSS   (DI), X9            // w_scale
    MULSS   X5, X9              // w_scale * absmax
    MULSS   di_inv127f<>(SB), X9 // * (1/127) = w_scale * absmax / 127
    MULSS   X9, X8              // X8 = block contribution
    ADDSS   X8, X0              // accumulate

dii_skip:
    ADDQ    $128, SI            // x += 32 floats
    ADDQ    $20, DI             // blocks += 20 bytes
    DECQ    CX
    JNZ     dii_block

dii_done:
    MOVSS   X0, ret+24(FP)
    VZEROUPPER
    RET
