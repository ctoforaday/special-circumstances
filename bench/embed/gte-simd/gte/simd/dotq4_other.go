//go:build !amd64 && !arm64

package simd

import "unsafe"

// DotQ4 scalar fallback.
func DotQ4(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32 {
	xp := (*[1 << 30]float32)(x)
	bp := (*[1 << 30]byte)(blocks)
	sum := float32(0)

	for b := 0; b < nBlocks; b++ {
		bOff := b * 20
		// Read scale (little-endian)
		bits := uint32(bp[bOff]) | uint32(bp[bOff+1])<<8 | uint32(bp[bOff+2])<<16 | uint32(bp[bOff+3])<<24
		scale := *(*float32)(unsafe.Pointer(&bits))
		xOff := b * 32

		for i := 0; i < 16; i++ {
			q := float32(int(bp[bOff+4+i]&0x0F) - 8)
			sum += xp[xOff+i] * scale * q
		}
		for i := 0; i < 16; i++ {
			q := float32(int(bp[bOff+4+i]>>4) - 8)
			sum += xp[xOff+16+i] * scale * q
		}
	}
	return sum
}

// LinearQ4 scalar fallback.
func LinearQ4(y, x, w, bias unsafe.Pointer, seqLen, inDim, outDim int) {
	yp := (*[1 << 30]float32)(y)
	xp := (*[1 << 30]float32)(x)
	bp := (*[1 << 30]byte)(w)
	var biasP *[1 << 30]float32
	if bias != nil {
		biasP = (*[1 << 30]float32)(bias)
	}

	nBlocks := inDim / 32
	rowBytes := nBlocks * 20

	for s := 0; s < seqLen; s++ {
		for o := 0; o < outDim; o++ {
			xPtr := unsafe.Pointer(&xp[s*inDim])
			wPtr := unsafe.Pointer(&bp[o*rowBytes])
			dot := DotQ4(xPtr, wPtr, nBlocks)
			if biasP != nil {
				dot += biasP[o]
			}
			yp[s*outDim+o] = dot
		}
	}
}
