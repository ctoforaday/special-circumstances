//go:build cgo

package gte

// #cgo LDFLAGS: -lopenblas
// #include <cblas.h>
import "C"
import "unsafe"

// openblasEnabled reports whether OpenBLAS is available at build time.
const openblasEnabled = true

// cblasSgemm calls OpenBLAS cblas_sgemm directly, zero Go allocations.
func cblasSgemm(transA, transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) {
	var ta, tb C.enum_CBLAS_TRANSPOSE
	if transA {
		ta = C.CblasTrans
	} else {
		ta = C.CblasNoTrans
	}
	if transB {
		tb = C.CblasTrans
	} else {
		tb = C.CblasNoTrans
	}
	C.cblas_sgemm(
		C.CblasRowMajor,
		ta, tb,
		C.blasint(m), C.blasint(n), C.blasint(k),
		C.float(alpha),
		(*C.float)(unsafe.Pointer(&a[0])), C.blasint(lda),
		(*C.float)(unsafe.Pointer(&b[0])), C.blasint(ldb),
		C.float(beta),
		(*C.float)(unsafe.Pointer(&c[0])), C.blasint(ldc),
	)
}
