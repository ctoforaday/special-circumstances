package gte

import (
	"math"
	"math/rand"
	"testing"
)

func softmaxRow(n int, seed int64) []float32 {
	rng := rand.New(rand.NewSource(seed))
	row := make([]float32, n)
	for i := range row {
		row[i] = float32(rng.NormFloat64()) * 6
	}
	return row
}

// softmaxRef is softmax computed with math.Exp in float64.
func softmaxRef(x []float32) []float64 {
	maxVal := float64(x[0])
	for _, v := range x {
		maxVal = math.Max(maxVal, float64(v))
	}
	out := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		out[i] = math.Exp(float64(v) - maxVal)
		sum += out[i]
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

func TestSoftmaxMatchesReference(t *testing.T) {
	sizes := []int{}
	for n := 1; n <= 40; n++ {
		sizes = append(sizes, n)
	}
	sizes = append(sizes, 64, 128, 256, 270, 512)
	for _, n := range sizes {
		row := softmaxRow(n, int64(n))
		ref := softmaxRef(row)
		softmax(row)
		for i := range row {
			if rel := math.Abs(float64(row[i])-ref[i]) / math.Max(ref[i], 1e-30); rel > 1e-5 && ref[i] > 1e-30 {
				t.Fatalf("n=%d i=%d: got %v, want %v (rel %g)", n, i, row[i], ref[i], rel)
			}
		}
	}
}

// TestSoftmaxMaskedEntries covers the -10000 attention-mask score and the range
// edge of the exponential.
func TestSoftmaxMaskedEntries(t *testing.T) {
	row := []float32{0, -10000, 1, -87.9 + 1, -88 + 1, -88.5 + 1, -10000}
	ref := softmaxRef(row)
	softmax(row)
	for i := range row {
		if row[i] < 0 || math.Abs(float64(row[i])-ref[i]) > 1e-6 {
			t.Errorf("i=%d: got %v, want %v", i, row[i], ref[i])
		}
	}
}

func BenchmarkSoftmax(b *testing.B) {
	src := softmaxRow(270, 1)
	row := make([]float32, len(src))
	for i := 0; i < b.N; i++ {
		copy(row, src)
		softmax(row)
	}
}
