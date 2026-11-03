//go:build amd64

package simd

import "unsafe"

// DotQ4Int computes a Q4 dot product using integer MAC instructions.
// On amd64: AVX2 VPMADDUBSW+VPMADDWD (dynamic x quantization to int8).
// On arm64: SDOT (signed 8×8→32 dot product).
//
// Per block: quantizes x to int8 on-the-fly, does integer dot with Q4 nibbles,
// then descales once. ~2 integer MAC instructions per 32 elements vs ~20 float
// instructions in DotQ4.
//
//go:noescape
func DotQ4Int(x unsafe.Pointer, blocks unsafe.Pointer, nBlocks int) float32
