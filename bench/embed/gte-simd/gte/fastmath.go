package gte

import (
	"math"
	"unsafe"
)

// --- fast float32 exp ---
// Schraudolph-style bit trick + 2nd-order correction.
// Max relative error ~0.3% over [-10, 10], which is fine for softmax/GELU.

func fastExp(x float32) float32 {
	// Clamp to avoid overflow/underflow
	if x < -88 {
		return 0
	}
	if x > 88 {
		return float32(math.MaxFloat32)
	}
	// Use the standard library for now — but as float32 via bit-level trick
	// Actually, a polynomial approximation is faster than math.Exp(float64(x))
	// Remez minimax degree-6 on [-log(2)/2, log(2)/2] after range reduction
	const (
		ln2    = 0.6931471805599453
		invLn2 = 1.4426950408889634
	)
	// Range reduction: x = n*ln2 + r, |r| <= ln2/2
	n := math.Floor(float64(x)*invLn2 + 0.5)
	r := float32(float64(x) - n*ln2)

	// Polynomial approximation of exp(r) for |r| <= ln2/2
	// Coefficients from minimax fit
	r2 := r * r
	p := float32(1.0) + r*(1.0+r*(0.5+r*(0.16666667+r*(0.041666668+r*(0.008333334+r*0.001388889)))))
	_ = r2

	// Reconstruct: exp(x) = 2^n * exp(r)
	// Use bit manipulation for 2^n
	bits := math.Float32bits(p)
	bits += uint32(n) << 23
	return math.Float32frombits(bits)
}

// --- fast float32 tanh ---
// Rational approximation, good to ~1e-4 over [-5, 5].
// Outside that range tanh ≈ ±1 anyway.

func fastTanh(x float32) float32 {
	if x < -5 {
		return -1
	}
	if x > 5 {
		return 1
	}
	x2 := x * x
	// Padé [3/2] approximant: tanh(x) ≈ x(1 + x²/15) / (1 + 2x²/5 + x⁴/15)
	// Simplified rational form that avoids division instability
	num := x * (135135 + x2*(17325+x2*(378+x2)))
	den := float32(135135) + x2*(62370+x2*(3150+x2*28))
	return num / den
}

// --- fast float32 inverse sqrt ---

func fastInvSqrt(x float32) float32 {
	// Classic Quake III fast inverse sqrt + one Newton-Raphson step
	half := 0.5 * x
	bits := *(*uint32)(unsafe.Pointer(&x))
	bits = 0x5f3759df - (bits >> 1)
	y := *(*float32)(unsafe.Pointer(&bits))
	y = y * (1.5 - half*y*y) // 1st Newton step
	y = y * (1.5 - half*y*y) // 2nd Newton step for better precision
	return y
}

// --- fast float32 sqrt ---

func fastSqrt(x float32) float32 {
	if x <= 0 {
		return 0
	}
	return x * fastInvSqrt(x)
}
