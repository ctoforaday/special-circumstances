//go:build goexperiment.simd

package gte

import (
	"math"
	"math/rand"
	"simd"
	"testing"
)

func naiveGemm64(transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c0 []float32, ldc int) []float64 {
	out := make([]float64, m*ldc)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			s := 0.0
			for p := 0; p < k; p++ {
				var bv float32
				if transB {
					bv = b[j*ldb+p]
				} else {
					bv = b[p*ldb+j]
				}
				s += float64(a[i*lda+p]) * float64(bv)
			}
			out[i*ldc+j] = float64(alpha)*s + float64(beta)*float64(c0[i*ldc+j])
		}
	}
	return out
}

func TestPortableSgemmDifferential(t *testing.T) {
	t.Logf("lanes=%d VectorBitSize=%d Emulated=%v", pLanes(), simd.VectorBitSize(), simd.Emulated())
	rng := rand.New(rand.NewSource(1))
	shapes := [][3]int{
		{1, 1, 1}, {1, 7, 3}, {2, 3, 5}, {5, 15, 17}, {6, 16, 8}, {7, 17, 9},
		{13, 33, 31}, {31, 31, 31}, {32, 32, 32}, {37, 47, 255}, {61, 65, 257},
		{100, 384, 384}, {127, 1536, 384}, {200, 384, 1536}, {383, 1537, 385},
		{384, 1536, 384}, {384, 384, 1536}, {9, 200, 32}, {200, 32, 200}, {271, 32, 271},
	}
	worst := 0.0
	for _, s := range shapes {
		m, n, k := s[0], s[1], s[2]
		for _, transB := range []bool{true, false} {
			for _, ab := range [][2]float32{{1, 0}, {0.17, 1}, {1.3, 0.5}} {
				alpha, beta := ab[0], ab[1]
				lda, ldb, ldc := k+3, k+1, n+2
				if !transB {
					ldb = n + 5
				}
				a := make([]float32, m*lda)
				var b []float32
				if transB {
					b = make([]float32, n*ldb)
				} else {
					b = make([]float32, k*ldb)
				}
				c := make([]float32, m*ldc)
				for i := range a {
					a[i] = float32(rng.NormFloat64())
				}
				for i := range b {
					b[i] = float32(rng.NormFloat64())
				}
				for i := range c {
					c[i] = float32(rng.NormFloat64())
				}
				want := naiveGemm64(transB, m, n, k, alpha, a, lda, b, ldb, beta, c, ldc)
				// Scale for relative error: alpha*sum|a||b| + |beta c|, so cancellation does not inflate it.
				got := append([]float32(nil), c...)
				sgemm(false, transB, m, n, k, alpha, a, lda, b, ldb, beta, got, ldc)
				maxRel := 0.0
				for i := 0; i < m; i++ {
					for j := 0; j < n; j++ {
						sc := 0.0
						for p := 0; p < k; p++ {
							var bv float32
							if transB {
								bv = b[j*ldb+p]
							} else {
								bv = b[p*ldb+j]
							}
							sc += math.Abs(float64(a[i*lda+p]) * float64(bv))
						}
						sc = math.Abs(float64(alpha))*sc + math.Abs(float64(beta)*float64(c[i*ldc+j])) + 1e-30
						r := math.Abs(float64(got[i*ldc+j])-want[i*ldc+j]) / sc
						maxRel = max(maxRel, r)
					}
					// padding columns untouched
					for j := n; j < ldc; j++ {
						if got[i*ldc+j] != c[i*ldc+j] {
							t.Fatalf("m=%d n=%d k=%d transB=%v: wrote padding col %d", m, n, k, transB, j)
						}
					}
				}
				worst = max(worst, maxRel)
				if maxRel > 1e-5 {
					t.Errorf("m=%d n=%d k=%d transB=%v alpha=%v beta=%v: max rel err %.3g", m, n, k, transB, alpha, beta, maxRel)
				}
			}
		}
	}
	t.Logf("sgemm: %d shapes x {NT,NN} x 3 alpha/beta, worst relative error (|err| / sum|a*b|) = %.3g", len(shapes), worst)
}

func TestPortableExpSoftmax(t *testing.T) {
	L := pLanes()
	// exp over [-87, 0] dense, against math.Exp. Below about -87.3 the result is
	// subnormal and the exponent-field reconstruction (shared with fastExp and the
	// amd64 assembly) loses it; the subnormal band is reported separately.
	worst, worstAt := 0.0, 0.0
	for start := -87.0; start < 0; start += 0.01 * float64(L) {
		x := make([]float32, L)
		for i := range x {
			x[i] = float32(start + 0.01*float64(i))
		}
		ref := make([]float64, L)
		for i := range x {
			ref[i] = math.Exp(float64(x[i]))
		}
		expSubPortable(x, 0)
		for i := range x {
			if e := math.Abs(float64(x[i])-ref[i]) / ref[i]; e > worst {
				worst, worstAt = e, float64(float32(start+0.01*float64(i)))
			}
		}
	}
	t.Logf("expSubPortable over [-87,0): max rel err vs math.Exp = %.3g at x=%.4f", worst, worstAt)
	for _, v := range []float32{-87.5, -87.9} {
		x := make([]float32, L)
		for i := range x {
			x[i] = v
		}
		expSubPortable(x, 0)
		t.Logf("subnormal band: exp(%v) = %g (math.Exp %g, fastExp %g)", v, x[0], math.Exp(float64(v)), fastExp(v))
	}
	if worst > 2e-6 {
		t.Errorf("exp max rel err %.3g", worst)
	}
	// below -88 gives 0
	x := make([]float32, L)
	for i := range x {
		x[i] = -10000
	}
	if s := expSubPortable(x, 0); s != 0 {
		t.Errorf("exp(-10000) sum = %v", s)
	}

	rng := rand.New(rand.NewSource(2))
	worstSM := 0.0
	for _, n := range []int{1, 2, 3, 5, 7, 8, 9, 15, 16, 17, 31, 33, 40, 127, 200, 271, 384, 512, 513} {
		for trial := 0; trial < 20; trial++ {
			x := make([]float32, n)
			for i := range x {
				x[i] = float32(rng.NormFloat64() * 8)
				if rng.Intn(10) == 0 {
					x[i] = -10000 // masked token
				}
			}
			mx := math.Inf(-1)
			for _, v := range x {
				mx = max(mx, float64(v))
			}
			ref := make([]float64, n)
			sum := 0.0
			for i, v := range x {
				ref[i] = math.Exp(float64(v) - mx)
				sum += ref[i]
			}
			softmax(x)
			for i := range x {
				r := ref[i] / sum
				e := math.Abs(float64(x[i]) - r)
				if r > 1e-30 {
					e /= r
				}
				worstSM = max(worstSM, e)
			}
		}
	}
	t.Logf("softmax: max rel err vs float64 math.Exp reference = %.3g", worstSM)
	if worstSM > 1e-5 {
		t.Errorf("softmax max rel err %.3g", worstSM)
	}
}

func TestPortableDot(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	worst := 0.0
	for n := 0; n <= 70; n++ {
		x, y := make([]float32, n), make([]float32, n)
		ref, sc := 0.0, 1e-30
		for i := range x {
			x[i], y[i] = float32(rng.NormFloat64()), float32(rng.NormFloat64())
			ref += float64(x[i]) * float64(y[i])
			sc += math.Abs(float64(x[i]) * float64(y[i]))
		}
		worst = max(worst, math.Abs(float64(sdot(x, y))-ref)/sc)
	}
	t.Logf("dot: lengths 0..70, worst relative error = %.3g", worst)
	if worst > 1e-6 {
		t.Errorf("dot max rel err %.3g", worst)
	}
}
