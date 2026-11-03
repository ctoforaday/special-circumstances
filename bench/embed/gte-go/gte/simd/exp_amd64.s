#include "textflag.h"

// func expSubAVX2(x *float32, n int, sub float32) float32
//
// x[i] = exp(x[i]-sub), 8 at a time, returning the sum. Same reduction and
// polynomial as fastExp: d = n*ln2 + r with n the nearest integer to d/ln2,
// exp(r) by degree 6 Taylor, then n added to the exponent field. ln2 is split in
// two so the reduction stays accurate in float32. Arguments below -88 give 0.
//
// Y0 d, r   Y1 sub      Y2 1/ln2  Y3 ln2 high  Y4 ln2 low  Y5 -88
// Y6 1/720  Y7 1/120    Y8 1/24   Y9 1/6       Y10 1/2     Y11 1
// Y12 n     Y13 mask    Y14 p     Y15 sum

#define BROADCAST(bits, reg, xreg) \
    MOVL    $bits, AX; \
    VMOVD   AX, xreg; \
    VBROADCASTSS xreg, reg

TEXT ·expSubAVX2(SB), NOSPLIT, $0-28
    MOVQ    x+0(FP), DI
    MOVQ    n+8(FP), CX
    VBROADCASTSS sub+16(FP), Y1

    BROADCAST(0x3FB8AA3B, Y2, X2)
    BROADCAST(0x3F318000, Y3, X3)
    BROADCAST(0xB95E8083, Y4, X4)
    BROADCAST(0xC2B00000, Y5, X5)
    BROADCAST(0x3AB60B62, Y6, X6)
    BROADCAST(0x3C088889, Y7, X7)
    BROADCAST(0x3D2AAAAB, Y8, X8)
    BROADCAST(0x3E2AAAAB, Y9, X9)
    BROADCAST(0x3F000000, Y10, X10)
    BROADCAST(0x3F800000, Y11, X11)
    VXORPS  Y15, Y15, Y15

    SHRQ    $3, CX

loop:
    VMOVUPS (DI), Y0
    VSUBPS  Y1, Y0, Y0
    VCMPPS  $13, Y5, Y0, Y13            // d >= -88

    VMULPS  Y2, Y0, Y12
    VROUNDPS $0, Y12, Y12               // n
    VFNMADD231PS Y3, Y12, Y0            // r = d - n*high
    VFNMADD231PS Y4, Y12, Y0            //     - n*low

    VMOVAPS Y6, Y14
    VFMADD213PS Y7, Y0, Y14
    VFMADD213PS Y8, Y0, Y14
    VFMADD213PS Y9, Y0, Y14
    VFMADD213PS Y10, Y0, Y14
    VFMADD213PS Y11, Y0, Y14
    VFMADD213PS Y11, Y0, Y14

    VCVTTPS2DQ Y12, Y12
    VPSLLD  $23, Y12, Y12
    VPADDD  Y12, Y14, Y14
    VANDPS  Y13, Y14, Y14

    VMOVUPS Y14, (DI)
    VADDPS  Y14, Y15, Y15

    ADDQ    $32, DI
    DECQ    CX
    JNZ     loop

    VEXTRACTF128 $1, Y15, X0
    VADDPS  X0, X15, X15
    VHADDPS X15, X15, X15
    VHADDPS X15, X15, X15
    VMOVSS  X15, ret+24(FP)
    VZEROUPPER
    RET
