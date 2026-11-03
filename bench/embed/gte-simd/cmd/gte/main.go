package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/rcarmo/gte-go/gte"
)

func main() {
	modelPath := flag.String("model-path", "gte-small.gtemodel", "Path to .gtemodel file")
	flag.Parse()

	sentences := flag.Args()
	if len(sentences) == 0 {
		sentences = []string{
			"The weather is lovely today.",
			"It's so sunny outside!",
			"He drove to the stadium.",
			"Machine learning is transforming industries.",
			"I love programming in Go.",
		}
	}

	start := time.Now()
	model, err := gte.Load(*modelPath)
	if err != nil {
		log.Fatalf("failed to load model: %v", err)
	}
	defer model.Close()

	fmt.Printf("Model loaded in %.2f s\n", time.Since(start).Seconds())
	fmt.Printf("Embedding dimension: %d\n", model.Dim())
	fmt.Printf("Max sequence length: %d\n\n", model.MaxLen())

	embeddings := make([][]float32, len(sentences))
	for i, s := range sentences {
		t0 := time.Now()
		emb, err := model.Embed(s)
		if err != nil {
			log.Fatalf("embed %d: %v", i, err)
		}
		embeddings[i] = emb
		fmt.Printf("S%d: \"%s\"\n", i+1, s)
		fmt.Printf("    Time: %.3f ms\n", float64(time.Since(t0).Microseconds())/1000)
		fmt.Printf("    Embedding preview: % .6f % .6f % .6f % .6f % .6f ...\n", emb[0], emb[1], emb[2], emb[3], emb[4])
		fmt.Println()
	}

	fmt.Println("Cosine similarity matrix:")
	fmt.Print("     ")
	for i := range sentences {
		fmt.Printf("  S%d   ", i+1)
	}
	fmt.Println()

	for i := range sentences {
		fmt.Printf("S%d: ", i+1)
		for j := range sentences {
			sim, _ := gte.CosineSimilarity(embeddings[i], embeddings[j])
			fmt.Printf(" %.3f ", sim)
		}
		fmt.Println()
	}
}
