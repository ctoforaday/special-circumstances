package gte

import (
	"unsafe"

	"github.com/rcarmo/gte-go/gte/simd"
)

// sgemm dispatches matrix multiplication to the best available backend.
//
// Priority:
// 1. OpenBLAS via CGo (CGO_ENABLED=1) — fastest, SIMD + threading
// 2. Go assembly SGEMM (amd64 AVX2+FMA, arm64 NEON) — zero alloc, no C deps
// 3. gonum pure-Go BLAS — fallback for other architectures
func sgemm(transA, transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) {
	if openblasEnabled {
		cblasSgemm(transA, transB, m, n, k, alpha, a, lda, b, ldb, beta, c, ldc)
		return
	}
	if portableSimd {
		if portableSgemm(transA, transB, m, n, k, alpha, a, lda, b, ldb, beta, c, ldc) {
			return
		}
	}
	// Handle beta scaling (assembly kernels assume beta=1, i.e. C += ...)
	if beta == 0 {
		for i := 0; i < m; i++ {
			row := c[i*ldc : i*ldc+n]
			for j := range row {
				row[j] = 0
			}
		}
	} else if beta != 1 {
		for i := 0; i < m; i++ {
			row := c[i*ldc : i*ldc+n]
			for j := range row {
				row[j] *= beta
			}
		}
	}
	// Use SIMD assembly SGEMM when available (amd64/arm64)
	if simd.HasSgemmAsm {
		if !transA && transB {
			if amd64 && m < gebpMinRows {
				// amd64, short input: VGATHERDPS needs no packing, which pays
				// for itself only over enough rows.
				simd.SgemmNTGather(m, n, k, alpha,
					unsafePtr(a), unsafePtr(b), unsafePtr(c),
					lda, ldb, ldc)
			} else {
				// GEBP: packs B once per panel, so the k loop runs two FMAs per
				// 16 columns without a gather. arm64 uses it at every size.
				simd.SgemmNTGebp(m, n, k, alpha,
					unsafePtr(a), unsafePtr(b), unsafePtr(c),
					lda, ldb, ldc)
			}
			return
		}
		// NN: tiled assembly kernel, zero allocs.
		if !transA && !transB {
			simd.SgemmNN(m, n, k, alpha,
				unsafePtr(a), unsafePtr(b), unsafePtr(c),
				lda, ldb, ldc)
			return
		}
	}
	// Fallback: gonum BLAS
	blasImpl.Sgemm(blasTrans(transA), blasTrans(transB), m, n, k, alpha, a, lda, b, ldb, beta, c, ldc)
}

// gebpMinRows is the row count (token count, in the model) from which amd64 uses
// GEBP for the NT case. It equals the seqLen below which selfAttention stays
// sequential; see BenchmarkSgemmNTKernels in gte/simd.
const gebpMinRows = 32

const blockSize = 64

// sgemmNTSimd: C = alpha * A * B^T + beta * C, using sdot for inner products.
// A is [m,k], B is [n,k] (row-major), both accessed row-wise.
func sgemmNTSimd(m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) {
	for i := 0; i < m; i++ {
		aRow := a[i*lda : i*lda+k]
		cRow := c[i*ldc : i*ldc+n]
		if beta == 0 {
			for j := range cRow {
				cRow[j] = 0
			}
		} else if beta != 1 {
			for j := range cRow {
				cRow[j] *= beta
			}
		}
		for j := 0; j < n; j++ {
			bRow := b[j*ldb : j*ldb+k]
			cRow[j] += alpha * simd.Sdot(aRow, bRow)
		}
	}
}

// sgemmNNSimd: C = alpha * A * B + beta * C, using SIMD saxpy for inner loop.
// A is [m,k], B is [k,n], both row-major.
func sgemmNNSimd(m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) {
	for i := 0; i < m; i++ {
		cRow := c[i*ldc : i*ldc+n]
		if beta == 0 {
			for j := range cRow {
				cRow[j] = 0
			}
		} else if beta != 1 {
			for j := range cRow {
				cRow[j] *= beta
			}
		}
	}
	for i := 0; i < m; i++ {
		cRow := c[i*ldc : i*ldc+n]
		for p := 0; p < k; p++ {
			aVal := alpha * a[i*lda+p]
			if aVal == 0 {
				continue
			}
			bRow := b[p*ldb : p*ldb+n]
			simd.Saxpy(aVal, bRow, cRow)
		}
	}
}

func unsafePtr(s []float32) unsafe.Pointer {
	return unsafe.Pointer(&s[0])
}
