package gte

import (
	"math"
	"unsafe"

	"github.com/rcarmo/gte-go/gte/simd"
)

// linearQ4Int computes Y = X·W^T + bias where W is Q4-quantized,
// using integer MAC with amortized x-quantization.
//
// For each sequence position:
//   1. Quantize x to int8 ONCE (per block: find absmax, scale to [-127,127])
//   2. For each output row: integer dot of pre-quantized x with w nibbles
//   3. Descale: y[o] = Σ_blocks (x_scale[b] * w_scale[o,b] * int_dot[b])
//
// The key win: x quantization is amortized across all outDim rows.
func linearQ4Int(y, x []float32, w []byte, b []float32, seqLen, inDim, outDim int) {
	nBlocks := inDim / QK4_0
	rowBytes := nBlocks * BlockQ4Size

	// Pre-allocated scratch
	xI8 := make([]int8, nBlocks*QK4_0)
	xScales := make([]float32, nBlocks)

	for s := 0; s < seqLen; s++ {
		xRow := x[s*inDim:]
		yRow := y[s*outDim:]

		// Step 1: Quantize entire x row to int8 blocks
		for blk := 0; blk < nBlocks; blk++ {
			off := blk * QK4_0
			xBlock := xRow[off : off+QK4_0]

			absmax := float32(0)
			for _, v := range xBlock {
				av := v
				if av < 0 {
					av = -av
				}
				if av > absmax {
					absmax = av
				}
			}
			xScales[blk] = absmax / 127.0
			if absmax == 0 {
				for i := 0; i < QK4_0; i++ {
					xI8[off+i] = 0
				}
				continue
			}
			invScale := float32(127.0) / absmax
			for i := 0; i < QK4_0; i++ {
				q := int(math.Round(float64(xBlock[i] * invScale)))
				if q > 127 {
					q = 127
				} else if q < -127 {
					q = -127
				}
				xI8[off+i] = int8(q)
			}
		}

		// Step 2: For each output, integer dot with weight nibbles
		for o := 0; o < outDim; o++ {
			wRow := w[o*rowBytes:]
			sum := float32(0)

			for blk := 0; blk < nBlocks; blk++ {
				bOff := blk * BlockQ4Size
				bits := uint32(wRow[bOff]) | uint32(wRow[bOff+1])<<8 | uint32(wRow[bOff+2])<<16 | uint32(wRow[bOff+3])<<24
				wScale := *(*float32)(unsafe.Pointer(&bits))

				if xScales[blk] == 0 || wScale == 0 {
					continue
				}

				xOff := blk * QK4_0
				dot := int32(0)
				for i := 0; i < 16; i++ {
					nibLo := int32(wRow[bOff+4+i]&0x0F) - 8
					nibHi := int32(wRow[bOff+4+i]>>4) - 8
					dot += int32(xI8[xOff+i]) * nibLo
					dot += int32(xI8[xOff+16+i]) * nibHi
				}

				sum += xScales[blk] * wScale * float32(dot)
			}

			yRow[o] = sum
			if b != nil {
				yRow[o] += b[o]
			}
		}
	}
}

// Unused import guard
var _ = simd.DotQ4Int
