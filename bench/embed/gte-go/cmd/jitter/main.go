package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/rcarmo/gte-go/gte"
)

func main() {
	modelPath := flag.String("model", "gte-small.gtemodel", "model path")
	textsPath := flag.String("texts", "", "file with one text per line")
	n := flag.Int("n", 5000, "total embeddings")
	batchSize := flag.Int("batch", 0, "batch size (0=discrete)")
	warmup := flag.Int("warmup", 50, "warmup embeddings")
	outPath := flag.String("out", "jitter.csv", "output CSV path")
	flag.Parse()

	model, err := gte.Load(*modelPath)
	if err != nil {
		log.Fatalf("load: %v", err)
	}
	defer model.Close()

	var texts []string
	if *textsPath != "" {
		f, _ := os.Open(*textsPath)
		s := bufio.NewScanner(f)
		for s.Scan() {
			if t := s.Text(); t != "" {
				texts = append(texts, t)
			}
		}
		f.Close()
	}
	if len(texts) == 0 {
		texts = []string{"The quick brown fox jumps over the lazy dog"}
	}

	buf := make([]float32, model.Dim())
	for i := 0; i < *warmup; i++ {
		_ = model.EmbedTo(texts[i%len(texts)], buf)
	}

	out, _ := os.Create(*outPath)
	w := csv.NewWriter(out)

	if *batchSize <= 0 {
		// Discrete mode: one embed at a time
		w.Write([]string{"index", "ms"})
		for i := 0; i < *n; i++ {
			start := time.Now()
			_ = model.EmbedTo(texts[i%len(texts)], buf)
			ms := float64(time.Since(start).Microseconds()) / 1000.0
			w.Write([]string{strconv.Itoa(i), fmt.Sprintf("%.3f", ms)})
		}
	} else {
		// Batch mode: measure per-batch latency, report per-embed average
		w.Write([]string{"batch", "total_ms", "per_embed_ms"})
		nBatches := *n / *batchSize
		for b := 0; b < nBatches; b++ {
			batch := make([]string, *batchSize)
			for i := range batch {
				batch[i] = texts[(b*(*batchSize)+i)%len(texts)]
			}
			start := time.Now()
			for _, t := range batch {
				_ = model.EmbedTo(t, buf)
			}
			totalMs := float64(time.Since(start).Microseconds()) / 1000.0
			perMs := totalMs / float64(*batchSize)
			w.Write([]string{strconv.Itoa(b), fmt.Sprintf("%.3f", totalMs), fmt.Sprintf("%.3f", perMs)})
		}
	}

	w.Flush()
	out.Close()

	// Print summary
	fmt.Printf("wrote %s\n", *outPath)
}
