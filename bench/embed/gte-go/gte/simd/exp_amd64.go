package simd

const hasExpSub = true

// ExpSub sets x[i] = exp(x[i]-sub) and returns the sum of the results, using
// AVX2+FMA. len(x) must be a nonzero multiple of 8. Every x[i]-sub must be <= 88;
// results for arguments below -88 are 0.
func ExpSub(x []float32, sub float32) float32 {
	return expSubAVX2(&x[0], len(x), sub)
}

//go:noescape
func expSubAVX2(x *float32, n int, sub float32) float32
