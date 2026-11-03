package gte

import (
	"math"
	"testing"
	"unsafe"

	"github.com/rcarmo/gte-go/gte/simd"
)

func TestQuantizeRoundtrip(t *testing.T) {
	// Create a known vector
	data := make([]float32, 64) // 2 blocks
	for i := range data {
		data[i] = float32(i-32) * 0.1
	}

	blocks := quantizeQ4(data)
	if len(blocks) != 2*BlockQ4Size {
		t.Fatalf("expected %d bytes, got %d", 2*BlockQ4Size, len(blocks))
	}

	// Dequantize and check
	out := dequantQ4(blocks, 64)
	maxErr := float32(0)
	for i := range data {
		err := data[i] - out[i]
		if err < 0 {
			err = -err
		}
		if err > maxErr {
			maxErr = err
		}
	}
	// Q4 error should be < scale/2 ≈ max_val/14
	maxVal := float32(3.2) // max(abs(data)) = 3.2
	expectedMaxErr := maxVal / 7.0 // scale = maxVal/7, quantization error ≤ scale/2
	if maxErr > expectedMaxErr {
		t.Errorf("max quantization error %.4f exceeds expected %.4f", maxErr, expectedMaxErr)
	}
	t.Logf("max quantization error: %.4f (expected < %.4f)", maxErr, expectedMaxErr)
}

func TestDotQ4Accuracy(t *testing.T) {
	// Create two vectors, compute FP32 dot, then Q4 dot
	n := 384 // GTE hidden size
	a := make([]float32, n)
	b := make([]float32, n)
	for i := range a {
		a[i] = float32(i%17-8) * 0.01
		b[i] = float32(i%13-6) * 0.02
	}

	// FP32 reference
	fp32Dot := float32(0)
	for i := range a {
		fp32Dot += a[i] * b[i]
	}

	// Quantize b, compute dot
	bQ4 := quantizeQ4(b)
	nBlocks := n / QK4_0
	q4Dot := simd.DotQ4(unsafe.Pointer(&a[0]), unsafe.Pointer(&bQ4[0]), nBlocks)

	relErr := float64(math.Abs(float64(q4Dot-fp32Dot)) / math.Abs(float64(fp32Dot)))
	t.Logf("FP32 dot: %.6f, Q4 dot: %.6f, rel error: %.4f%%", fp32Dot, q4Dot, relErr*100)

	if relErr > 0.15 { // 15% relative error is generous for Q4
		t.Errorf("Q4 dot product relative error too high: %.2f%%", relErr*100)
	}
}

func TestLinearQ4(t *testing.T) {
	inDim := 64   // must be multiple of 32
	outDim := 32  // must be multiple of 32
	seqLen := 2

	// Create input
	x := make([]float32, seqLen*inDim)
	for i := range x {
		x[i] = float32(i%19-9) * 0.01
	}

	// Create weight matrix [outDim, inDim] and quantize
	w := make([]float32, outDim*inDim)
	for i := range w {
		w[i] = float32(i%23-11) * 0.005
	}
	bias := make([]float32, outDim)
	for i := range bias {
		bias[i] = float32(i) * 0.01
	}

	// FP32 reference: Y = X · W^T + bias
	yRef := make([]float32, seqLen*outDim)
	for s := 0; s < seqLen; s++ {
		for o := 0; o < outDim; o++ {
			sum := float32(0)
			for k := 0; k < inDim; k++ {
				sum += x[s*inDim+k] * w[o*inDim+k]
			}
			yRef[s*outDim+o] = sum + bias[o]
		}
	}

	// Q4 computation
	wQ4 := make([]byte, 0, outDim*inDim/QK4_0*BlockQ4Size)
	for o := 0; o < outDim; o++ {
		row := w[o*inDim : o*inDim+inDim]
		wQ4 = append(wQ4, quantizeQ4(row)...)
	}

	yQ4 := make([]float32, seqLen*outDim)
	simd.LinearQ4(
		unsafe.Pointer(&yQ4[0]),
		unsafe.Pointer(&x[0]),
		unsafe.Pointer(&wQ4[0]),
		unsafe.Pointer(&bias[0]),
		seqLen, inDim, outDim,
	)

	// Compare
	maxRelErr := float64(0)
	for i := range yRef {
		if yRef[i] == 0 {
			continue
		}
		relErr := math.Abs(float64(yQ4[i]-yRef[i])) / math.Abs(float64(yRef[i]))
		if relErr > maxRelErr {
			maxRelErr = relErr
		}
	}
	t.Logf("LinearQ4 max relative error: %.4f%%", maxRelErr*100)
	if maxRelErr > 0.25 { // 25% for Q4 matmul with small test values
		t.Errorf("LinearQ4 relative error too high: %.2f%%", maxRelErr*100)
	}
}

func TestDotQ4Zero(t *testing.T) {
	// All zeros should produce zero dot
	n := 32
	a := make([]float32, n)
	bQ4 := quantizeQ4(make([]float32, n))
	result := simd.DotQ4(unsafe.Pointer(&a[0]), unsafe.Pointer(&bQ4[0]), 1)
	if result != 0 {
		t.Errorf("expected 0, got %f", result)
	}
}

func BenchmarkDotQ4_384(b *testing.B) {
	n := 384
	x := make([]float32, n)
	for i := range x {
		x[i] = float32(i%17-8) * 0.01
	}
	data := make([]float32, n)
	for i := range data {
		data[i] = float32(i%13-6) * 0.02
	}
	blocks := quantizeQ4(data)
	nBlocks := n / QK4_0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simd.DotQ4(unsafe.Pointer(&x[0]), unsafe.Pointer(&blocks[0]), nBlocks)
	}
}

func BenchmarkLinearQ4_384x384(b *testing.B) {
	inDim := 384
	outDim := 384
	seqLen := 1

	x := make([]float32, seqLen*inDim)
	for i := range x {
		x[i] = float32(i%17-8) * 0.01
	}
	w := make([]float32, outDim*inDim)
	for i := range w {
		w[i] = float32(i%23-11) * 0.005
	}
	wQ4 := make([]byte, 0, outDim*inDim/QK4_0*BlockQ4Size)
	for o := 0; o < outDim; o++ {
		wQ4 = append(wQ4, quantizeQ4(w[o*inDim:o*inDim+inDim])...)
	}
	bias := make([]float32, outDim)
	y := make([]float32, seqLen*outDim)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simd.LinearQ4(
			unsafe.Pointer(&y[0]),
			unsafe.Pointer(&x[0]),
			unsafe.Pointer(&wQ4[0]),
			unsafe.Pointer(&bias[0]),
			seqLen, inDim, outDim,
		)
	}
}

func TestDotQ4IntAccuracy(t *testing.T) {
	n := 384
	a := make([]float32, n)
	b := make([]float32, n)
	for i := range a {
		a[i] = float32(i%17-8) * 0.01
		b[i] = float32(i%13-6) * 0.02
	}

	// FP32 reference
	fp32Dot := float32(0)
	for i := range a {
		fp32Dot += a[i] * b[i]
	}

	// Quantize b, compute dot with integer path
	bQ4 := quantizeQ4(b)
	nBlocks := n / QK4_0
	q4IntDot := simd.DotQ4Int(unsafe.Pointer(&a[0]), unsafe.Pointer(&bQ4[0]), nBlocks)

	relErr := float64(math.Abs(float64(q4IntDot-fp32Dot)) / math.Abs(float64(fp32Dot)))
	t.Logf("FP32 dot: %.6f, Q4Int dot: %.6f, rel error: %.4f%%", fp32Dot, q4IntDot, relErr*100)

	// Also compare with float DotQ4
	q4FloatDot := simd.DotQ4(unsafe.Pointer(&a[0]), unsafe.Pointer(&bQ4[0]), nBlocks)
	t.Logf("Q4Float dot: %.6f (for comparison)", q4FloatDot)

	if relErr > 0.20 { // 20% — generous for double quantization
		t.Errorf("Q4Int dot product relative error too high: %.2f%%", relErr*100)
	}
}

func BenchmarkDotQ4Int_384(b *testing.B) {
	n := 384
	x := make([]float32, n)
	for i := range x {
		x[i] = float32(i%17-8) * 0.01
	}
	data := make([]float32, n)
	for i := range data {
		data[i] = float32(i%13-6) * 0.02
	}
	blocks := quantizeQ4(data)
	nBlocks := n / QK4_0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simd.DotQ4Int(unsafe.Pointer(&x[0]), unsafe.Pointer(&blocks[0]), nBlocks)
	}
}

func BenchmarkLinearQ4Int_384x384(b *testing.B) {
	inDim := 384
	outDim := 384
	seqLen := 1

	x := make([]float32, seqLen*inDim)
	for i := range x { x[i] = float32(i%17-8) * 0.01 }
	w := make([]float32, outDim*inDim)
	for i := range w { w[i] = float32(i%23-11) * 0.005 }
	wQ4 := make([]byte, 0, outDim*inDim/QK4_0*BlockQ4Size)
	for o := 0; o < outDim; o++ {
		wQ4 = append(wQ4, quantizeQ4(w[o*inDim:o*inDim+inDim])...)
	}
	bias := make([]float32, outDim)
	y := make([]float32, seqLen*outDim)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		linearQ4Int(y, x, wQ4, bias, seqLen, inDim, outDim)
	}
}

func TestLinearQ4IntAccuracy(t *testing.T) {
	inDim := 64
	outDim := 32
	seqLen := 2

	x := make([]float32, seqLen*inDim)
	for i := range x { x[i] = float32(i%19-9) * 0.01 }
	w := make([]float32, outDim*inDim)
	for i := range w { w[i] = float32(i%23-11) * 0.005 }
	bias := make([]float32, outDim)
	for i := range bias { bias[i] = float32(i) * 0.01 }

	// FP32 reference
	yRef := make([]float32, seqLen*outDim)
	for s := 0; s < seqLen; s++ {
		for o := 0; o < outDim; o++ {
			sum := float32(0)
			for k := 0; k < inDim; k++ { sum += x[s*inDim+k] * w[o*inDim+k] }
			yRef[s*outDim+o] = sum + bias[o]
		}
	}

	wQ4 := make([]byte, 0)
	for o := 0; o < outDim; o++ { wQ4 = append(wQ4, quantizeQ4(w[o*inDim:o*inDim+inDim])...) }
	yQ4 := make([]float32, seqLen*outDim)
	linearQ4Int(yQ4, x, wQ4, bias, seqLen, inDim, outDim)

	maxRelErr := float64(0)
	for i := range yRef {
		if yRef[i] == 0 { continue }
		relErr := math.Abs(float64(yQ4[i]-yRef[i])) / math.Abs(float64(yRef[i]))
		if relErr > maxRelErr { maxRelErr = relErr }
	}
	t.Logf("LinearQ4Int max relative error: %.4f%%", maxRelErr*100)
	if maxRelErr > 0.30 {
		t.Errorf("error too high: %.2f%%", maxRelErr*100)
	}
}
