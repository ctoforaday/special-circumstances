# Optimization Report

This document details the systematic optimization of `gte-go` from the original
67.4ms per embedding down to 5.5ms (OpenBLAS) / 6.4ms (pure Go) on amd64, and
64ms on arm64.

## Approach

Three phases, each building on the previous:

1. **BLAS and fast math** — replace scalar loops with optimized matrix multiply
2. **Structural optimizations** — fuse operations, improve memory layout
3. **System-level** — SIMD assembly, mmap, zero-alloc tokenizer

## Phase 1 — BLAS and fast math (67.4 → 7.3ms)

### 1a – float32 fast math (67.4 → 63.4ms, 1.06×)

Replaced `math.Tanh`, `math.Exp`, and `math.Sqrt` (all float64 round-trips)
with float32 polynomial/Padé approximations in `gte/fastmath.go`:

- `fastTanh`: degree-7 Padé approximant, clamped at ±5
- `fastExp`: range-reduced polynomial with bit-manipulation reconstruction
- `fastInvSqrt`: Quake III bit trick + 2 Newton-Raphson steps

Modest 6% gain because `linear()` dominates.

### 1b – gonum BLAS (63.4 → 7.4ms, 9.1×)

Replaced the hand-rolled `linear()` dot-product loops with `blas32.Sgemm` from
gonum. This was the single largest improvement. Even gonum's pure-Go BLAS with
cache blocking massively outperforms scalar code.

### 1c – adaptive attention (7.4 → 7.3ms, 9.2×)

Short sequences (< 32 tokens) use a scalar attention path without goroutine
overhead. Longer sequences use BLAS for Q·K^T and attn·V.

## Phase 2 — Structural optimizations (7.3 → 6.3ms)

### 2a – fused QKV projection (7.3 → 6.6ms, 10.2×)

Concatenated Q/K/V weight matrices into a single `[3*hidden, hidden]` array at
load time. One Sgemm call replaces three.

### 2b – contiguous head layout (6.6 → 6.3ms, 10.7×)

Changed the QKV split to write directly into `[numHeads, seqLen, headDim]`
buffers. Eliminates strided access in attention inner loops.

### 2c – sigmoid GELU (abandoned)

Tested but the approximation error exceeded the model's cosine similarity
tolerance by 0.006.

## Phase 3 — System-level optimizations (6.3 → 5.5ms)

### 3a – OpenBLAS via CGo (6.3 → 5.5ms, 12.3×)

Direct CGo wrapper for `cblas_sgemm`, bypassing gonum entirely. Eliminated
99% of per-embedding allocations (gonum's `sgemmParallel` spawns goroutines).
OpenBLAS provides AVX2 SIMD acceleration.

### 3b – zero-alloc tokenizer

Refactored tokenizer to reuse pre-allocated buffers. Reduced remaining allocs
from 14 → 12, 1.4KB → 186B.

### 3c – mmap model loading

`LoadMmap()` maps the `.gtemodel` file directly. Weight slices point into the
mapped region — no copy into Go heap. 12× faster startup, ~127MB less resident
memory.

## Phase 4 — SIMD assembly

### AVX2+FMA kernels (amd64)

- `Sdot`: 16-wide dot product with FMA, 4-way unrolled
- `Saxpy`: 16-wide y += alpha*x with FMA
- `SgemmNN`: 32-wide C tiling, register-blocked k loop
- `SgemmNT`: 2j-at-a-time dot products, shared A loads
- GEBP 6×16 micro-kernel for large-m NT

### NEON kernels (arm64)

- `Sdot`, `Saxpy`: 8-wide with VFMLA
- `SgemmNN`: 16-wide C tiling
- `SgemmNT`: blocked tile with VFMLA, prefetch hints
- GEBP 4×16 micro-kernel for NT dispatch

### Architecture-specific dispatch

The NT matmul (used by all `linear()` calls) is the dominant bottleneck.
The optimal strategy differs per architecture:

| Path | amd64 | arm64 |
|---|---|---|
| NT, small m | gonum (its asm DotUnitary + cache blocking) | GEBP + NEON |
| NT, large m | GEBP + AVX2 | GEBP + NEON |
| NN | AVX2+FMA tiled kernel | NEON tiled kernel |

**Why gonum wins NT on amd64**: gonum's `f32.DotUnitary` uses hand-written x86
assembly with deep pipelining. Although it doesn't use FMA (VMULPS+VADDPS
instead of VFMADD231PS), its zero-overhead per-dot-product design beats our
FMA tile kernel which has horizontal-reduce overhead per output element.

**Why GEBP wins NT on arm64**: gonum has no NEON assembly for NT — it falls
back to scalar Go code. Our NEON GEBP micro-kernel is 1.6× faster.

## Key insight: the horizontal reduce problem

For NT GEMM, each output element C[i,j] = dot(A[i,:], B[j,:]) requires a
horizontal sum at the end of the dot product. On x86, this is 5 instructions
(VEXTRACTF128 + VADDPS + 2×VHADDPS). For 8064 output elements per sgemm call,
that's 40K overhead instructions — 10% of the total FMA work.

OpenBLAS avoids this by using the GEBP approach with B-packing, which converts
the NT dot products into NN-style fused multiply-accumulate without horizontal
reductions. This is the main reason OpenBLAS achieves 5.5ms vs our 6.4ms.

## Phase 4 — Q4 Quantization (model size optimization)

### Motivation

GTE-small has 33M parameters. At FP32, the model file is 63 MB. For edge
deployments (embedded devices, WASM, mobile), download size and flash storage
matter more than raw latency. Q4_0 quantization reduces the model to 20 MB
while preserving semantic accuracy.

### Q4_0 format

Block size: 32 elements. Each block:

```
┌──────────────┬──────────────────────────┐
│ scale (4B)   │ packed nibbles (16B)     │
│ float32      │ 32 × 4-bit values       │
└──────────────┴──────────────────────────┘
```

- **Compression**: 20 bytes per 32 floats (vs 128 FP32) = **6.4× reduction**
- **Quantization**: symmetric, `q = round(value / scale) + 8`, clamped to [0,15]
- **Dequantization**: `value = scale × (nibble - 8)`
- **What's quantized**: all weight matrices (Q/K/V, attention output, FFN up/down, pooler)
- **What stays FP32**: biases, LayerNorm params, attention scores, hidden states

### Model size

| Component | FP32 | Q4 |
|---|---|---|
| Weight matrices | 62.8 MB | 19.85 MB |
| Biases + LayerNorm | 0.23 MB | 0.23 MB |
| Vocab | 0.35 MB | 0.35 MB |
| **Total** | **63.4 MB** | **20.3 MB** |

### Accuracy

Q4 embeddings are nearly identical to FP32:

| Metric | Value |
|---|---|
| Same-text cosine (FP32 vs Q4) | **0.99** |
| Rank ordering | Preserved (no inversions) |
| Weak-pair bias | +0.015 (predictable, consistent) |

Example pair similarities:

| Pair | FP32 | Q4 |
|---|---|---|
| cat/kitten (strong) | 0.882 | 0.888 |
| cat/stocks (weak) | 0.695 | 0.715 |

### SIMD kernels

The Q4 dot product dequantizes on-the-fly during the FMA loop:

**amd64 (AVX2+FMA):**
```
per block (32 elements):
  VBROADCASTSS scale → Y1
  VMOVDQU 16 packed bytes → X2
  VPAND 0x0F → low nibbles
  VPSRLW 4 → high nibbles
  4× (VPMOVZXBD → VCVTDQ2PS → VSUBPS 8.0 → VMULPS scale → VFMADD231PS x)
```

**arm64 (NEON, via WORD macros):**
```
per block (32 elements):
  VLD1 16 bytes
  VAND 0x0F / VUSHR 4 → nibbles
  8× (UXTL → UCVTF → FSUB 8.0 → FMUL scale → VFMLA x)
```

Microbenchmarks (single dot product, 384 dimensions):
- **amd64**: 193 ns/dot (12 blocks)
- **arm64**: verified correct on CIX P1 (all tests pass)

### Latency trade-off

| Platform | FP32 (ms) | Q4 (ms) | Q4/FP32 |
|---|---|---|---|
| amd64 i7-12700 | 10 | 103 | 10.3× slower |
| arm64 CIX P1 | 20 | 92 | 4.6× slower |

**Why Q4 is slower**: On both platforms, the FP32 model weights fit in L3 cache.
The FP32 SIMD kernels (VGATHERDPS on amd64, GEBP NEON on arm64) operate at
near-peak FMA throughput. Q4 adds per-block overhead:
- 1 scale broadcast
- 2 nibble extractions (AND + shift)
- 4 (amd64) or 8 (arm64) int→float conversions
- 4/8 subtract-8 operations
- 4/8 scale multiplications

This overhead exceeds the memory bandwidth savings because the working set
already fits in cache. Q4 would show wins on systems with:
- Small L3 (model doesn't fit)
- Low memory bandwidth (embedded SoCs)
- Multiple concurrent model instances (memory-constrained)

### When to use Q4

| Use case | Recommended format |
|---|---|
| Latency-critical serving | FP32 (10ms amd64, 20ms arm64) |
| Edge/IoT deployment | **Q4** (20MB, 1 alloc, correct rankings) |
| WASM/browser inference | **Q4** (smaller download, no SIMD needed) |
| Memory-constrained (many models) | **Q4** (3× less RAM per model) |
| Batch throughput | FP32 + OpenBLAS (5.5ms/embed) |

### Usage

```bash
# Convert model
python convert_model_q4.py models/gte-small gte-small-q4.gtemodel

# Use in Go
model, _ := gte.LoadQ4("gte-small-q4.gtemodel")
emb, _ := model.Embed("Hello world")
```

The API is identical to the FP32 model. `IsQ4Model(path)` detects the format
by reading the file magic (`GTE4` vs `GTE1`).

### Integer MAC analysis (VPMADDUBSW / SDOT)

We explored replacing the float dequant-dot pipeline with integer multiply-
accumulate instructions:

| Approach | Instructions/block | Per-block overhead |
|---|---|---|
| **Float DotQ4** (current) | 25 | None — float accumulator |
| **Integer DotQ4Int** | 20 | +4 (descale: cvt+3×mul per block) |
| **Integer DotQ4Int** (per-dot) | 36+ | +x-quant: 30 insn overhead |

**Why integer is only marginally better in theory (20 vs 25 insn/block):**

The float path's key advantage: it accumulates directly into a float32 register
across all 12 blocks per dot product, then does ONE horizontal sum at the end.
No per-block scaling needed because `dequant = scale × (nibble-8)` folds the
scale into the VMULPS that precedes VFMADD231PS.

The integer path MUST descale per block because `w_scale` differs per block
per output row. Each block requires: `int_dot → VCVTDQ2PS → MULSS × 2 → ADDSS`.
That's 4 extra float instructions that negate the VPMADDUBSW throughput gain.

**When integer WOULD win:**
- Fixed scale across all blocks (requires format change)
- Extremely bandwidth-limited systems where 6.4× less memory read matters
- VPDPBUSD (AVX-VNNI) reducing the MAC to 1 instruction per 32 elements
- Very long vectors where amortized x-quant becomes negligible

**Measured results:**
- DotQ4Int per-dot: 4130ns vs DotQ4: 177ns (23× slower, x-quant dominates)
- linearQ4Int amortized scalar: 237μs vs LinearQ4 SIMD: 71μs (3.3× slower)
- Full model: 265ms (int scalar) vs 103ms (SIMD float) vs 10ms (FP32)
