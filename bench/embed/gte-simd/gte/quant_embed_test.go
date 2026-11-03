package gte

import (
	"os"
	"testing"
)

func TestEmbedQ4Similarities(t *testing.T) {
	modelPath := os.Getenv("GTE_Q4_MODEL_PATH")
	if modelPath == "" {
		modelPath = "gte-small-q4.gtemodel"
	}
	if _, err := os.Stat(modelPath); err != nil {
		t.Skipf("Q4 model file not found: %s", modelPath)
	}

	m, err := LoadQ4(modelPath)
	if err != nil {
		t.Fatalf("LoadQ4: %v", err)
	}
	defer m.Close()

	texts := []string{
		"The cat sat on the mat",
		"A kitten rested on the rug",
		"Stock prices rose sharply today",
	}

	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(text)
		if err != nil {
			t.Fatalf("Embed(%q): %v", text, err)
		}
		embeddings[i] = emb
	}

	// Similar sentences should have higher similarity
	sim12, _ := CosineSimilarity(embeddings[0], embeddings[1])
	sim13, _ := CosineSimilarity(embeddings[0], embeddings[2])
	sim23, _ := CosineSimilarity(embeddings[1], embeddings[2])

	t.Logf("cat/kitten similarity:   %.4f", sim12)
	t.Logf("cat/stocks similarity:   %.4f", sim13)
	t.Logf("kitten/stocks similarity: %.4f", sim23)

	if sim12 < sim13 {
		t.Error("expected cat/kitten more similar than cat/stocks")
	}
	if sim12 < sim23 {
		t.Error("expected cat/kitten more similar than kitten/stocks")
	}
	if sim12 < 0.5 {
		t.Errorf("cat/kitten similarity too low: %.4f", sim12)
	}
}

func BenchmarkEmbedQ4(b *testing.B) {
	modelPath := os.Getenv("GTE_Q4_MODEL_PATH")
	if modelPath == "" {
		modelPath = "gte-small-q4.gtemodel"
	}
	if _, err := os.Stat(modelPath); err != nil {
		b.Skipf("Q4 model file not found: %s", modelPath)
	}

	m, err := LoadQ4(modelPath)
	if err != nil {
		b.Fatalf("LoadQ4: %v", err)
	}
	defer m.Close()

	buf := make([]float32, m.HiddenSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.EmbedTo("The quick brown fox jumps over the lazy dog", buf)
	}
}
