# SIMD Assembly Reference

Architecture-specific SIMD kernels for matrix operations. All kernels are
in Go plan9 assembly format, compiled natively by the Go toolchain —
no CGo required.

## File index

| File | Arch | Kernels | Used for |
|---|---|---|---|
| `simd_amd64.s` | x86-64 AVX2+FMA | `Sdot`, `Saxpy` | Attention scores, vector ops |
| `simd_arm64.s` | ARM64 NEON | `Sdot`, `Saxpy` | Attention scores, vector ops |
| `sgemm_amd64.s` | x86-64 AVX2+FMA | `SgemmNT`, `SgemmNN` | Direct GEMM (small matrices) |
| `sgemm_arm64.s` | ARM64 NEON | `SgemmNT`¹, `SgemmNN` | Direct GEMM |
| `gebp_amd64.s` | x86-64 AVX2+FMA | 6×16 micro-kernel | GEBP NT for large m |
| `gebp_arm64.s` | ARM64 NEON | 4×16 micro-kernel | GEBP NT dispatch |
| `exp_amd64.s` | x86-64 AVX2+FMA | `ExpSub` | Softmax in the attention heads |
| `sgemm_blocked_amd64.s` | x86-64 AVX2+FMA | NT tile (2j) | Blocked FMA tile |
| `sgemm_blocked_arm64.s` | ARM64 NEON | NT tile (1j) | Blocked FMA tile |
| `simd_other.go` | Scalar | All | Fallback for other architectures |

¹ Known bug in direct `SgemmNT` on arm64; not used in inference dispatch.

## Register conventions

### amd64 (AVX2, 16 × 256-bit YMM)

- Y0–Y11: accumulators (up to 6×2 for GEBP)
- Y12–Y13: packed B data loads
- Y14: broadcast A / temp
- Y15: alpha broadcast

### arm64 (NEON, 32 × 128-bit V)

- V0–V15: accumulators (up to 4×4 for GEBP)
- V16–V19: B data loads
- V20: broadcast A
- V31: alpha scalar

## GEBP micro-kernel design

The GEBP (General Block Panel) approach packs B tiles into column-panel format,
converting the NT access pattern to NN for the micro-kernel:

1. **Pack**: transpose B[jj:jj+NR, 0:k] → Bp[k, NR]
2. **Micro-kernel**: for each k step, broadcast A[row,p] and FMA with Bp[p,0:NR]
3. **Store**: scale by alpha, add to C

This eliminates horizontal reductions per output element — the key bottleneck
in direct NT GEMM assembly.

## Build tags

- `amd64`: AVX2+FMA assembly
- `arm64`: NEON+VFMLA assembly
- `!amd64 && !arm64`: scalar Go fallback (panics for GEBP; gonum used instead)

## Testing

52 tests cover all kernels at GTE-small matrix sizes, including edge cases,
pre-filled C matrices, and cross-platform correctness verification.
The arm64 kernels are tested on real hardware (OrangePi 6 Plus, CIX P1 CD8160).
