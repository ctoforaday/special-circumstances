package gte

import (
	"errors"
	"math"
	"os"
	"testing"
)

func loadTestModel(t *testing.T) *Model {
	t.Helper()
	path := os.Getenv("GTE_MODEL_PATH")
	if path == "" {
		path = "gte-small.gtemodel"
	}
	m, err := Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("model file not found: %s", path)
		}
		t.Fatalf("load model: %v", err)
	}
	t.Cleanup(m.Close)
	return m
}

func approxEq(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestEmbedSimilarities(t *testing.T) {
	model := loadTestModel(t)
	checkEmbedSimilarities(t, model)
}

func TestEmbedSimilaritiesMmap(t *testing.T) {
	path := os.Getenv("GTE_MODEL_PATH")
	if path == "" {
		path = "gte-small.gtemodel"
	}
	model, err := LoadMmap(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("model file not found: %s", path)
		}
		t.Fatalf("load model: %v", err)
	}
	t.Cleanup(model.Close)
	checkEmbedSimilarities(t, model)
}

func checkEmbedSimilarities(t *testing.T, model *Model) {
	t.Helper()

	sentences := []string{
		"I love cats",
		"I love dogs",
		"The stock market crashed",
	}

	expected := [][]float64{
		{1.000, 0.898, 0.727},
		{0.898, 1.000, 0.722},
		{0.727, 0.722, 1.000},
	}
	tol := 0.02

	embeddings := make([][]float32, len(sentences))
	for i, s := range sentences {
		emb, err := model.Embed(s)
		if err != nil {
			t.Fatalf("embed %d: %v", i, err)
		}
		if len(emb) != model.Dim() {
			t.Fatalf("embedding dim mismatch: got %d want %d", len(emb), model.Dim())
		}
		embeddings[i] = emb
	}

	for i := range embeddings {
		for j := range embeddings {
			sim, err := CosineSimilarity(embeddings[i], embeddings[j])
			if err != nil {
				t.Fatalf("cosine %d,%d: %v", i, j, err)
			}
			if !approxEq(float64(sim), expected[i][j], tol) {
				t.Fatalf("sim[%d,%d]=%.3f want %.3f (+/-%0.3f)", i, j, sim, expected[i][j], tol)
			}
		}
	}
}
