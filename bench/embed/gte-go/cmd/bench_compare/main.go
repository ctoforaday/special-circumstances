package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/rcarmo/gte-go/gte"
)

var texts = []string{
	"The quick brown fox jumps over the lazy dog",
	"Machine learning is transforming how we build software systems",
	"I love programming in Go because it compiles fast and runs efficiently",
	"The stock market crashed yesterday amid fears of a global recession",
	"Neural networks can approximate any continuous function given enough parameters",
	"A simple embedding model like GTE-small produces 384-dimensional vectors",
	"Proxmox Virtual Environment is an open-source server virtualization platform",
	"Memory-mapped I/O avoids copying data from kernel space to user space",
	"OpenBLAS provides optimized BLAS routines with AVX2 SIMD acceleration on x86",
	"Cats are wonderful creatures that have been domesticated for thousands of years",
	"The weather forecast predicts heavy rainfall across the eastern seaboard tomorrow",
	"Kubernetes orchestrates containerized workloads across a cluster of machines",
	"Hello",
	"This is a significantly longer piece of text that exercises the tokenizer with many subword tokens and various punctuation marks, numbers like 42 and 3.14, and even some unusual words like defenestration and sesquipedalian.",
	"日本語のテキスト",
}

type stats struct {
	name   string
	values []float64
}

func (s *stats) report() {
	sort.Float64s(s.values)
	n := len(s.values)
	mean := 0.0
	for _, v := range s.values {
		mean += v
	}
	mean /= float64(n)

	stddev := 0.0
	for _, v := range s.values {
		d := v - mean
		stddev += d * d
	}
	stddev = math.Sqrt(stddev / float64(n))

	median := s.values[n/2]
	p5 := s.values[int(float64(n)*0.05)]
	p95 := s.values[int(float64(n)*0.95)]
	min := s.values[0]
	max := s.values[n-1]

	fmt.Printf("  %-35s  mean=%6.3f  median=%6.3f  min=%6.3f  max=%6.3f  p5=%6.3f  p95=%6.3f  stddev=%5.3f ms\n",
		s.name, mean, median, min, max, p5, p95, stddev)
}

func bench(name string, model *gte.Model, samples int) stats {
	buf := make([]float32, model.Dim())
	s := stats{name: name}

	// Warmup: 5 full passes
	for i := 0; i < 5; i++ {
		for _, t := range texts {
			_ = model.EmbedTo(t, buf)
		}
	}

	for i := 0; i < samples; i++ {
		start := time.Now()
		for _, t := range texts {
			if err := model.EmbedTo(t, buf); err != nil {
				fmt.Fprintf(os.Stderr, "embed error: %v\n", err)
				os.Exit(1)
			}
		}
		elapsed := time.Since(start)
		avgMs := float64(elapsed) / float64(len(texts)) / float64(time.Millisecond)
		s.values = append(s.values, avgMs)
	}
	return s
}

func main() {
	path := os.Getenv("GTE_MODEL_PATH")
	if path == "" {
		path = "gte-small.gtemodel"
	}
	samples := 20

	fmt.Println("GTE-Go Phase 3 Comparison Benchmark")
	fmt.Println("====================================")
	fmt.Printf("Texts:   %d diverse sentences (1–44 words, inc. short/long/unicode)\n", len(texts))
	fmt.Printf("Samples: %d per variant\n", samples)
	fmt.Printf("Warmup:  5 full passes\n")
	fmt.Println()

	// --- Load models ---
	fmt.Print("Loading via Load()... ")
	t0 := time.Now()
	modelLoad, err := gte.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Load failed: %v\n", err)
		os.Exit(1)
	}
	loadTime := time.Since(t0)
	fmt.Printf("%.3fs\n", loadTime.Seconds())

	fmt.Print("Loading via LoadMmap()... ")
	t1 := time.Now()
	modelMmap, err := gte.LoadMmap(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadMmap failed: %v\n", err)
		os.Exit(1)
	}
	mmapTime := time.Since(t1)
	fmt.Printf("%.3fs\n", mmapTime.Seconds())
	fmt.Println()

	// --- Inference benchmarks ---
	// P3a = OpenBLAS + old tokenizer (alloc-heavy) — same inference path as P3b with Load
	// P3b = OpenBLAS + zero-alloc tokenizer — Load path
	// P3c = OpenBLAS + zero-alloc tokenizer — LoadMmap path
	// Since P3a and P3b only differ in alloc count (not speed), we label them together.

	fmt.Println("Inference latency (ms per embedding, average over all texts):")
	fmt.Println()

	s1 := bench("P3a/P3b: Load + OpenBLAS", modelLoad, samples)
	s2 := bench("P3c: LoadMmap + OpenBLAS", modelMmap, samples)

	// Also run Load a second time to check stability
	s3 := bench("P3a/P3b: Load + OpenBLAS (run 2)", modelLoad, samples)

	s1.report()
	s2.report()
	s3.report()

	fmt.Println()
	fmt.Printf("Load time:     Load=%.3fs  LoadMmap=%.3fs  (%.1f× faster)\n",
		loadTime.Seconds(), mmapTime.Seconds(), loadTime.Seconds()/mmapTime.Seconds())

	modelLoad.Close()
	modelMmap.Close()
}
