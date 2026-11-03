package gte

import (
	"math"
	"runtime"
	"strings"
	"testing"
)

func TestFastExpAccuracy(t *testing.T) {
	testCases := []float32{-10, -5, -2, -1, -0.5, 0, 0.5, 1, 2, 5, 10}
	for _, x := range testCases {
		got := fastExp(x)
		want := float32(math.Exp(float64(x)))
		rel := relErrF(got, want)
		if rel > 0.01 { // 1% tolerance
			t.Errorf("fastExp(%v)=%v, want=%v, relErr=%v", x, got, want, rel)
		}
	}
}

func TestFastTanhAccuracy(t *testing.T) {
	testCases := []float32{-5, -2, -1, -0.5, 0, 0.5, 1, 2, 5}
	for _, x := range testCases {
		got := fastTanh(x)
		want := float32(math.Tanh(float64(x)))
		abs := float32(math.Abs(float64(got - want)))
		if abs > 0.001 {
			t.Errorf("fastTanh(%v)=%v, want=%v, absErr=%v", x, got, want, abs)
		}
	}
}

func TestFastTanhSaturation(t *testing.T) {
	if fastTanh(-10) != -1 {
		t.Errorf("fastTanh(-10)=%v, want -1", fastTanh(-10))
	}
	if fastTanh(10) != 1 {
		t.Errorf("fastTanh(10)=%v, want 1", fastTanh(10))
	}
}

func TestFastInvSqrtAccuracy(t *testing.T) {
	testCases := []float32{0.01, 0.1, 0.5, 1, 2, 4, 10, 100, 10000}
	for _, x := range testCases {
		got := fastInvSqrt(x)
		want := float32(1.0 / math.Sqrt(float64(x)))
		rel := relErrF(got, want)
		if rel > 0.001 { // 0.1% tolerance (2 Newton steps)
			t.Errorf("fastInvSqrt(%v)=%v, want=%v, relErr=%v", x, got, want, rel)
		}
	}
}

func TestFastSqrtAccuracy(t *testing.T) {
	testCases := []float32{0, 0.01, 0.25, 1, 4, 9, 100}
	for _, x := range testCases {
		got := fastSqrt(x)
		want := float32(math.Sqrt(float64(x)))
		abs := float32(math.Abs(float64(got - want)))
		if abs > 0.01 {
			t.Errorf("fastSqrt(%v)=%v, want=%v, absErr=%v", x, got, want, abs)
		}
	}
}

func TestCosineSimilarityIdentical(t *testing.T) {
	a := []float32{1, 0, 0}
	sim, err := CosineSimilarity(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(sim)-1.0) > 1e-6 {
		t.Errorf("identical vectors: sim=%v, want 1.0", sim)
	}
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	sim, err := CosineSimilarity(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(sim)) > 1e-6 {
		t.Errorf("orthogonal vectors: sim=%v, want 0.0", sim)
	}
}

func TestCosineSimilarityDimensionMismatch(t *testing.T) {
	a := []float32{1, 2}
	b := []float32{1, 2, 3}
	_, err := CosineSimilarity(a, b)
	if err == nil {
		t.Error("expected error for dimension mismatch")
	}
}

func TestEmbedBatch(t *testing.T) {
	model := loadTestModel(t)
	texts := []string{"hello", "world", "test"}
	embeddings, err := model.EmbedBatch(texts)
	if err != nil {
		t.Fatalf("EmbedBatch: %v", err)
	}
	if len(embeddings) != 3 {
		t.Fatalf("got %d embeddings, want 3", len(embeddings))
	}
	for i, emb := range embeddings {
		if len(emb) != model.Dim() {
			t.Errorf("embedding %d: dim=%d, want %d", i, len(emb), model.Dim())
		}
		// Verify L2-normalized
		norm := float32(0)
		for _, v := range emb {
			norm += v * v
		}
		if math.Abs(float64(norm)-1.0) > 0.01 {
			t.Errorf("embedding %d: norm=%v, want ~1.0", i, norm)
		}
	}
	// Same text should give same embedding
	single, err := model.Embed("hello")
	if err != nil {
		t.Fatal(err)
	}
	sim, _ := CosineSimilarity(embeddings[0], single)
	if sim < 0.999 {
		t.Errorf("batch vs single: sim=%v, want ~1.0", sim)
	}
}

// TestEmbedLongTextParallelHeads embeds a text long enough for selfAttention to
// run its heads in goroutines. The result must not depend on how many of them
// run at once.
func TestEmbedLongTextParallelHeads(t *testing.T) {
	model := loadTestModel(t)
	text := strings.Repeat("the quick brown fox jumps over the lazy dog near the riverbank ", 20)

	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	serial, err := model.Embed(text)
	if err != nil {
		t.Fatal(err)
	}
	serial = append([]float32(nil), serial...)

	runtime.GOMAXPROCS(8)
	for i := 0; i < 3; i++ {
		parallel, err := model.Embed(text)
		if err != nil {
			t.Fatal(err)
		}
		sim, err := CosineSimilarity(serial, parallel)
		if err != nil {
			t.Fatal(err)
		}
		if sim < 0.9999 {
			t.Errorf("run %d: sim=%v between 1 and 8 GOMAXPROCS, want ~1.0", i, sim)
		}
	}
}

func TestEmbedEmpty(t *testing.T) {
	model := loadTestModel(t)
	emb, err := model.Embed("")
	if err != nil {
		t.Fatalf("Embed empty: %v", err)
	}
	if len(emb) != model.Dim() {
		t.Errorf("dim=%d, want %d", len(emb), model.Dim())
	}
}

func relErrF(got, want float32) float32 {
	if want == 0 {
		return float32(math.Abs(float64(got)))
	}
	return float32(math.Abs(float64(got-want)) / math.Abs(float64(want)))
}
