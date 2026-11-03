package gte

import (
	"os"
	"testing"
	"time"
)

// BenchmarkEmbed measures single-text embedding latency.
// Set GTE_MODEL_PATH to point at a .gtemodel file; otherwise defaults to gte-small.gtemodel.
func BenchmarkEmbed(b *testing.B) {
	path := os.Getenv("GTE_MODEL_PATH")
	if path == "" {
		path = "gte-small.gtemodel"
	}
	model, err := Load(path)
	if err != nil {
		b.Skipf("model load failed: %v", err)
	}
	defer model.Close()
	benchEmbed(b, model)
}

func BenchmarkEmbedMmap(b *testing.B) {
	path := os.Getenv("GTE_MODEL_PATH")
	if path == "" {
		path = "gte-small.gtemodel"
	}
	model, err := LoadMmap(path)
	if err != nil {
		b.Skipf("mmap model load failed: %v", err)
	}
	defer model.Close()
	benchEmbed(b, model)
}

func benchEmbed(b *testing.B, model *Model) {
	buf := make([]float32, model.Dim())
	text := "The stock market crashed"

	b.ReportAllocs()
	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		if err := model.EmbedTo(text, buf); err != nil {
			b.Fatalf("embed failed: %v", err)
		}
	}
	elapsed := time.Since(start)
	avg := float64(elapsed) / float64(b.N) / float64(time.Millisecond)
	b.ReportMetric(avg, "ms/op_avg")
}

func BenchmarkEmbedBatchParallel(b *testing.B) {
	path := os.Getenv("GTE_MODEL_PATH")
	if path == "" {
		path = "gte-small.gtemodel"
	}
	model, err := Load(path)
	if err != nil {
		b.Skipf("model load failed: %v", err)
	}
	defer model.Close()

	texts := []string{
		"The quick brown fox jumps over the lazy dog",
		"Machine learning is transforming how we build software",
		"I love programming in Go because it compiles fast",
		"The stock market crashed yesterday amid recession fears",
		"Neural networks can approximate any continuous function",
		"A simple embedding model produces 384-dimensional vectors",
		"Memory-mapped I/O avoids copying data from kernel space",
		"Kubernetes orchestrates containerized workloads across clusters",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.EmbedBatchParallel(texts, 0)
		if err != nil {
			b.Fatalf("batch failed: %v", err)
		}
	}
}
