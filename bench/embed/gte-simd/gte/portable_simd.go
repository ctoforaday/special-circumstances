//go:build goexperiment.simd

package gte

// Portable-SIMD kernels: with GOEXPERIMENT=simd, linear, attention and softmax
// run on Go's vector-size-agnostic simd package on every GOARCH, with no
// assembly on the hot path.

import (
	"math"
	"simd"
	"sync"
)

const portableSimd = true

const (
	pMR = 6   // micro-tile rows
	pKC = 256 // k-block: keeps a packed B panel (pKC x NR floats) L1-resident
)

var pBufPool = sync.Pool{New: func() any { return new([]float32) }}

func pGetBuf(size int) *[]float32 {
	buf := pBufPool.Get().(*[]float32)
	if cap(*buf) < size {
		*buf = make([]float32, size)
	}
	*buf = (*buf)[:size]
	return buf
}

func pLanes() int {
	var z simd.Float32s
	return z.Len()
}

// portableSgemm: C = alpha*op(A)*op(B) + beta*C for the NN and NT cases.
func portableSgemm(transA, transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, beta float32, c []float32, ldc int) bool {
	if transA {
		return false
	}
	if m == 0 || n == 0 {
		return true
	}
	if beta == 0 {
		for i := 0; i < m; i++ {
			clear(c[i*ldc : i*ldc+n])
		}
	} else if beta != 1 {
		for i := 0; i < m; i++ {
			row := c[i*ldc : i*ldc+n]
			for j := range row {
				row[j] *= beta
			}
		}
	}
	if k == 0 {
		return true
	}
	gebpPortable(transB, m, n, k, alpha, a, lda, b, ldb, c, ldc)
	return true
}

// gebpPortable computes C += alpha * A * op(B). B is packed into column panels of
// NR = 2*lanes, k-blocked by pKC; a pMR x NR micro-kernel holds 12 accumulators.
func gebpPortable(transB bool, m, n, k int, alpha float32, a []float32, lda int, b []float32, ldb int, c []float32, ldc int) {
	L := pLanes()
	NR := 2 * L
	kcMax := min(k, pKC)
	buf := pGetBuf(kcMax * NR)
	defer pBufPool.Put(buf)
	bp := *buf
	var tmp [pMR * 32]float32 // NR <= 32 (512-bit vectors)

	for kk := 0; kk < k; kk += pKC {
		kc := min(pKC, k-kk)
		for jj := 0; jj < n; jj += NR {
			nr := min(NR, n-jj)
			if transB {
				packNT(b, ldb, jj, nr, kk, kc, NR, bp)
			} else {
				packNN(b, ldb, jj, nr, kk, kc, NR, bp)
			}
			pan := bp[:kc*NR]
			for ii := 0; ii < m; ii += pMR {
				mr := min(pMR, m-ii)
				var rows [pMR][]float32
				for i := 0; i < pMR; i++ {
					r := ii + min(i, mr-1) // tail rows repeat the last real row; results discarded
					rows[i] = a[r*lda+kk : r*lda+kk+kc]
				}
				if mr == pMR && nr == NR {
					microKernel(kc, &alpha, &rows, pan, c[ii*ldc+jj:], ldc, L)
					continue
				}
				t := tmp[:pMR*NR]
				clear(t)
				for i := 0; i < mr; i++ {
					copy(t[i*NR:i*NR+nr], c[(ii+i)*ldc+jj:(ii+i)*ldc+jj+nr])
				}
				microKernel(kc, &alpha, &rows, pan, t, NR, L)
				for i := 0; i < mr; i++ {
					copy(c[(ii+i)*ldc+jj:(ii+i)*ldc+jj+nr], t[i*NR:i*NR+nr])
				}
			}
		}
	}
}

// packNT: bp[p*NR+d] = B[jj+d, kk+p], zero for d >= nr.
func packNT(b []float32, ldb, jj, nr, kk, kc, NR int, bp []float32) {
	for d := 0; d < nr; d++ {
		row := b[(jj+d)*ldb+kk : (jj+d)*ldb+kk+kc]
		for p, v := range row {
			bp[p*NR+d] = v
		}
	}
	if nr < NR {
		for p := 0; p < kc; p++ {
			clear(bp[p*NR+nr : p*NR+NR])
		}
	}
}

// packNN: bp[p*NR+d] = B[kk+p, jj+d], zero for d >= nr.
func packNN(b []float32, ldb, jj, nr, kk, kc, NR int, bp []float32) {
	for p := 0; p < kc; p++ {
		dst := bp[p*NR : p*NR+NR]
		copy(dst, b[(kk+p)*ldb+jj:(kk+p)*ldb+jj+nr])
		if nr < NR {
			clear(dst[nr:])
		}
	}
}

// microKernel: C[0:6, 0:2L] += alpha * rows[0:6][0:k] . pan[k x 2L].
func microKernel(k int, alpha *float32, rows *[pMR][]float32, pan []float32, c []float32, ldc, L int) {
	NR := 2 * L
	var c00, c01, c10, c11, c20, c21, c30, c31, c40, c41, c50, c51 simd.Float32s
	a0 := rows[0][:k]
	a1 := rows[1][:len(a0)]
	a2 := rows[2][:len(a0)]
	a3 := rows[3][:len(a0)]
	a4 := rows[4][:len(a0)]
	a5 := rows[5][:len(a0)]
	pan = pan[:len(a0)*NR]
	for p := range a0 {
		bq := pan[p*NR : p*NR+NR]
		b0 := simd.LoadFloat32s(bq)
		b1 := simd.LoadFloat32s(bq[L:])
		x := simd.BroadcastFloat32s(a0[p])
		c00 = x.MulAdd(b0, c00)
		c01 = x.MulAdd(b1, c01)
		x = simd.BroadcastFloat32s(a1[p])
		c10 = x.MulAdd(b0, c10)
		c11 = x.MulAdd(b1, c11)
		x = simd.BroadcastFloat32s(a2[p])
		c20 = x.MulAdd(b0, c20)
		c21 = x.MulAdd(b1, c21)
		x = simd.BroadcastFloat32s(a3[p])
		c30 = x.MulAdd(b0, c30)
		c31 = x.MulAdd(b1, c31)
		x = simd.BroadcastFloat32s(a4[p])
		c40 = x.MulAdd(b0, c40)
		c41 = x.MulAdd(b1, c41)
		x = simd.BroadcastFloat32s(a5[p])
		c50 = x.MulAdd(b0, c50)
		c51 = x.MulAdd(b1, c51)
	}
	// alpha arrives by pointer: a float32 parameter live across the loop is
	// reloaded each iteration with a legacy-SSE MOVSS, which stalls next to
	// 256-bit registers.
	al := simd.BroadcastFloat32s(*alpha)
	pStore2(al, L, c[0*ldc:0*ldc+NR], c00, c01)
	pStore2(al, L, c[1*ldc:1*ldc+NR], c10, c11)
	pStore2(al, L, c[2*ldc:2*ldc+NR], c20, c21)
	pStore2(al, L, c[3*ldc:3*ldc+NR], c30, c31)
	pStore2(al, L, c[4*ldc:4*ldc+NR], c40, c41)
	pStore2(al, L, c[5*ldc:5*ldc+NR], c50, c51)
}

func pStore2(al simd.Float32s, L int, row []float32, v0, v1 simd.Float32s) {
	v0.MulAdd(al, simd.LoadFloat32s(row)).Store(row)
	r1 := row[L:]
	v1.MulAdd(al, simd.LoadFloat32s(r1)).Store(r1)
}

// dotPortable returns sum(x[i]*y[i]).
func dotPortable(x, y []float32) float32 {
	L := pLanes()
	y = y[:len(x)]
	var s0, s1 simd.Float32s
	i := 0
	for ; i+2*L <= len(x); i += 2 * L {
		s0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), s0)
		s1 = simd.LoadFloat32s(x[i+L:]).MulAdd(simd.LoadFloat32s(y[i+L:]), s1)
	}
	for ; i+L <= len(x); i += L {
		s0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), s0)
	}
	sum := hsum(s0.Add(s1), L)
	for ; i < len(x); i++ {
		sum += x[i] * y[i]
	}
	return sum
}

func hsum(v simd.Float32s, L int) float32 {
	var t [16]float32
	v.Store(t[:])
	s := float32(0)
	for _, e := range t[:L] {
		s += e
	}
	return s
}

// expSubPortable sets x[i] = exp(x[i]-sub) and returns the sum; len(x) must be a
// multiple of the lane count. Same reduction and degree-6 polynomial as
// fastExp; arguments below -88 give 0.
func expSubPortable(x []float32, sub float32) float32 {
	L := pLanes()
	vsub := simd.BroadcastFloat32s(sub)
	invLn2 := simd.BroadcastFloat32s(math.Float32frombits(0x3FB8AA3B))
	negHi := simd.BroadcastFloat32s(-math.Float32frombits(0x3F318000))
	negLo := simd.BroadcastFloat32s(-math.Float32frombits(0xB95E8083))
	lo := simd.BroadcastFloat32s(-88)
	magic := simd.BroadcastFloat32s(12582912) // 1.5*2^23: round to nearest even
	k6 := simd.BroadcastFloat32s(math.Float32frombits(0x3AB60B62))
	k5 := simd.BroadcastFloat32s(math.Float32frombits(0x3C088889))
	k4 := simd.BroadcastFloat32s(math.Float32frombits(0x3D2AAAAB))
	k3 := simd.BroadcastFloat32s(math.Float32frombits(0x3E2AAAAB))
	k2 := simd.BroadcastFloat32s(0.5)
	one := simd.BroadcastFloat32s(1)
	var acc simd.Float32s
	for i := 0; i+L <= len(x); i += L {
		d := simd.LoadFloat32s(x[i:]).Sub(vsub)
		ok := d.GreaterEqual(lo)
		d = d.Max(lo) // keeps n in int32 range; masked below
		n := d.MulAdd(invLn2, magic).Sub(magic)
		r := n.MulAdd(negHi, d)
		r = n.MulAdd(negLo, r)
		p := k6.MulAdd(r, k5)
		p = p.MulAdd(r, k4)
		p = p.MulAdd(r, k3)
		p = p.MulAdd(r, k2)
		p = p.MulAdd(r, one)
		p = p.MulAdd(r, one)
		e := p.ToBits().Add(n.ConvertToInt32().ToBits().ShiftAllLeft(23)).BitsToFloat32().Masked(ok)
		e.Store(x[i:])
		acc = acc.Add(e)
	}
	return hsum(acc, L)
}

// softmaxPortable: in-place softmax, every step vectorized; the tail is padded
// to a whole vector with values that exponentiate to 0.
func softmaxPortable(x []float32) {
	L := pLanes()
	n := len(x)
	body := n - n%L
	maxVal := x[0]
	if body > 0 {
		mv := simd.LoadFloat32s(x)
		for i := L; i < body; i += L {
			mv = mv.Max(simd.LoadFloat32s(x[i:]))
		}
		var t [16]float32
		mv.Store(t[:])
		for _, e := range t[:L] {
			maxVal = max(maxVal, e)
		}
	}
	for _, v := range x[body:] {
		maxVal = max(maxVal, v)
	}
	sum := float32(0)
	if body > 0 {
		sum = expSubPortable(x[:body], maxVal)
	}
	if body < n {
		var t [16]float32
		for i := range t[:L] {
			t[i] = maxVal - 1000
		}
		copy(t[:], x[body:])
		sum += expSubPortable(t[:L], maxVal)
		copy(x[body:], t[:n-body])
	}
	inv := simd.BroadcastFloat32s(1 / sum)
	for i := 0; i < body; i += L {
		simd.LoadFloat32s(x[i:]).Mul(inv).Store(x[i:])
	}
	s := 1 / sum
	for i := body; i < n; i++ {
		x[i] *= s
	}
}
