//go:build !amd64

package simd

import (
	"math"
	"unsafe"
)

// DotQ4Int scalar fallback — dynamic quantization + integer MAC emulation.
func DotQ4Int(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32 {
	xp := (*[1 << 30]float32)(x)
	bp := (*[1 << 30]byte)(blocks)
	total := float32(0)

	for b := 0; b < nBlocks; b++ {
		bOff := b * 20
		xOff := b * 32

		// Read w_scale
		bits := uint32(bp[bOff]) | uint32(bp[bOff+1])<<8 | uint32(bp[bOff+2])<<16 | uint32(bp[bOff+3])<<24
		wScale := *(*float32)(unsafe.Pointer(&bits))

		// Find absmax of x block
		absmax := float32(0)
		for i := 0; i < 32; i++ {
			v := xp[xOff+i]
			if v < 0 {
				v = -v
			}
			if v > absmax {
				absmax = v
			}
		}
		if absmax == 0 {
			continue
		}

		// Quantize x to int8
		invScale := float32(127.0) / absmax
		xI8 := [32]int8{}
		for i := 0; i < 32; i++ {
			q := int(math.Round(float64(xp[xOff+i] * invScale)))
			if q > 127 {
				q = 127
			}
			if q < -127 {
				q = -127
			}
			xI8[i] = int8(q)
		}

		// Integer dot with nibbles (unsigned, no -8 subtraction)
		// Then correct by subtracting 8*sum(x_i8)
		dot := int32(0)
		sumX := int32(0)
		for i := 0; i < 16; i++ {
			nibLo := int32(bp[bOff+4+i] & 0x0F)
			nibHi := int32(bp[bOff+4+i] >> 4)
			dot += int32(xI8[i]) * nibLo
			dot += int32(xI8[16+i]) * nibHi
			sumX += int32(xI8[i]) + int32(xI8[16+i])
		}
		dot -= 8 * sumX // correction for nibble-8 centering

		// Descale
		total += (wScale * absmax / 127.0) * float32(dot)
	}
	return total
}
