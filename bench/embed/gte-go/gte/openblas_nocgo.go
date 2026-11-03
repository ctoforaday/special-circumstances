//go:build !cgo

package gte

const openblasEnabled = false

func cblasSgemm(transA, transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) {
	panic("cblasSgemm called without CGo build")
}
