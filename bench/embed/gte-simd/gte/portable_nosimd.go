//go:build !goexperiment.simd

package gte

const portableSimd = false

func portableSgemm(transA, transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) bool {
	panic("portableSgemm: built without GOEXPERIMENT=simd")
}

func softmaxPortable(x []float32)        { panic("softmaxPortable: built without GOEXPERIMENT=simd") }
func dotPortable(x, y []float32) float32 { panic("dotPortable: built without GOEXPERIMENT=simd") }
