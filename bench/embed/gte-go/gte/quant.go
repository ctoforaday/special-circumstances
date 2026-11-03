package gte

// Q4_0 quantization: 4-bit symmetric, block size 32.
// Each block: float32 scale + 16 packed bytes (32 nibbles).
// Dequantized value: scale * (nibble - 8)
//
// Compression: 20 bytes per 32 floats (vs 128 FP32) = 6.4× reduction.

import (
	"unsafe"

	"github.com/rcarmo/gte-go/gte/simd"
)

const (
	QK4_0       = 32           // elements per quantization block
	BlockQ4Size = 4 + QK4_0/2 // 4 (scale) + 16 (packed nibbles) = 20 bytes
)

// dotQ4 computes the dot product of a float32 vector x with a Q4-quantized vector.
func dotQ4(x []float32, blocks []byte, numElements int) float32 {
	nBlocks := numElements / QK4_0
	return simd.DotQ4(unsafe.Pointer(&x[0]), unsafe.Pointer(&blocks[0]), nBlocks)
}

// linearQ4 computes Y = X·W^T + bias where W is Q4-quantized.
// Uses SIMD dequant-on-the-fly dot product.
func linearQ4(y, x []float32, w []byte, b []float32, seqLen, inDim, outDim int) {
	var biasPtr unsafe.Pointer
	if b != nil {
		biasPtr = unsafe.Pointer(&b[0])
	}
	simd.LinearQ4(
		unsafe.Pointer(&y[0]),
		unsafe.Pointer(&x[0]),
		unsafe.Pointer(&w[0]),
		biasPtr,
		seqLen, inDim, outDim,
	)
}

// dequantQ4 expands a Q4-quantized block array back to float32.
// Used for embeddings (looked up by index, not matmul'd).
func dequantQ4(blocks []byte, numElements int) []float32 {
	nBlocks := numElements / QK4_0
	out := make([]float32, numElements)
	dequantQ4Into(out, blocks, nBlocks)
	return out
}

// dequantQ4Into expands Q4 blocks into an existing float32 buffer (zero-alloc).
func dequantQ4Into(out []float32, blocks []byte, nBlocks int) {
	for b := 0; b < nBlocks; b++ {
		bOff := b * BlockQ4Size
		scale := readScaleLE(blocks[bOff:])
		qs := blocks[bOff+4 : bOff+4+16]
		oOff := b * QK4_0

		for i := 0; i < 16; i++ {
			out[oOff+i] = scale * float32(int(qs[i]&0x0F)-8)
		}
		for i := 0; i < 16; i++ {
			out[oOff+16+i] = scale * float32(int(qs[i]>>4)-8)
		}
	}
}

// quantizeQ4 quantizes a float32 slice to Q4_0 blocks.
// Length must be a multiple of QK4_0 (pad with zeros if needed).
func quantizeQ4(data []float32) []byte {
	nBlocks := len(data) / QK4_0
	out := make([]byte, nBlocks*BlockQ4Size)

	for b := 0; b < nBlocks; b++ {
		block := data[b*QK4_0 : b*QK4_0+QK4_0]
		bOff := b * BlockQ4Size

		// Find absmax
		amax := float32(0)
		for _, v := range block {
			if v < 0 && -v > amax {
				amax = -v
			} else if v > amax {
				amax = v
			}
		}
		scale := amax / 7.0
		writeScaleLE(out[bOff:], scale)

		var invScale float32
		if scale != 0 {
			invScale = 1.0 / scale
		}
		qs := out[bOff+4 : bOff+4+16]
		for i := 0; i < 16; i++ {
			lo := clampQ4(block[i] * invScale)
			hi := clampQ4(block[i+16] * invScale)
			qs[i] = lo | (hi << 4)
		}
	}
	return out
}

func clampQ4(v float32) uint8 {
	q := int(v + 8.5)
	if q < 0 {
		q = 0
	}
	if q > 15 {
		q = 15
	}
	return uint8(q)
}

func readScaleLE(b []byte) float32 {
	_ = b[3]
	bits := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return *(*float32)(unsafePtrUint32(&bits))
}

func writeScaleLE(b []byte, v float32) {
	bits := *(*uint32)(unsafePtrFloat32(&v))
	b[0] = byte(bits)
	b[1] = byte(bits >> 8)
	b[2] = byte(bits >> 16)
	b[3] = byte(bits >> 24)
}
