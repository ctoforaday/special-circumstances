//go:build !amd64

package simd

const hasExpSub = false

// ExpSub is not available on this architecture.
func ExpSub(x []float32, sub float32) float32 {
	panic("ExpSub: no assembly for this architecture")
}
