//go:build goexperiment.simd

package main

import (
	"fmt"
	"simd"
)

func main() {
	var v simd.Float32s
	fmt.Printf("VectorBitSize=%d Emulated=%v Float32s.Len=%d\n", simd.VectorBitSize(), simd.Emulated(), v.Len())
}
