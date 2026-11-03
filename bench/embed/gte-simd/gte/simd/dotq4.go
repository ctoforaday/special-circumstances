package simd

import "unsafe"

// DotQ4 computes the dot product of a float32 vector x[0:nBlocks*32] with a
// Q4_0-quantized vector stored as nBlocks blocks.
//
// Block format (20 bytes per 32 elements):
//   - bytes[0:4]:  float32 scale (little-endian)
//   - bytes[4:20]: 16 packed uint8, each containing 2 nibbles
//     low nibble (qs[i]&0xF) → elements 0..15
//     high nibble (qs[i]>>4) → elements 16..31
//   - Dequantized value: scale * (nibble - 8)
//
//go:noescape
func DotQ4(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32

// LinearQ4 computes Y[seqLen, outDim] = X[seqLen, inDim] · W_q4^T + bias.
// W is stored as outDim rows of Q4 blocks (inDim/32 blocks per row, 20 bytes each).
// bias may be nil (pass zero pointer and 0 for hasBias).
//
//go:noescape
func LinearQ4(y, x, w, bias unsafe.Pointer, seqLen, inDim, outDim int)
