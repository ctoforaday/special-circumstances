package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/rcarmo/gte-go/gte"
)

func main() {
	modelPath := flag.String("model-path", "gte-small.gtemodel", "Path to .gtemodel file")
	iterations := flag.Int("iterations", 100, "Number of benchmark iterations")
	warmup := flag.Int("warmup", 5, "Warmup iterations to run before timing")
	cpuProfile := flag.String("cpuprofile", "", "Write CPU profile to file")
	memProfile := flag.String("memprofile", "", "Write heap profile to file")
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

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatalf("could not create cpu profile: %v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start cpu profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("Loading model from %s...\n", *modelPath)
	start := time.Now()
	model, err := gte.Load(*modelPath)
	if err != nil {
		log.Fatalf("failed to load model: %v", err)
	}
	defer model.Close()
	fmt.Printf("Model loaded in %.3f s\n", time.Since(start).Seconds())
	fmt.Printf("Embedding dimension: %d, max seq len: %d\n", model.Dim(), model.MaxLen())

	buf := make([]float32, model.Dim())

	// Warmup
	if *warmup > 0 {
		fmt.Printf("Warmup: %d iterations...\n", *warmup)
		for i := 0; i < *warmup; i++ {
			if err := model.EmbedTo(sentences[i%len(sentences)], buf); err != nil {
				log.Fatalf("warmup embed failed: %v", err)
			}
		}
	}

	fmt.Printf("Benchmark: %d iterations over %d sentences...\n", *iterations, len(sentences))
	totalCalls := *iterations * len(sentences)
	var sink float32
	benchStart := time.Now()
	for i := 0; i < *iterations; i++ {
		for _, s := range sentences {
			if err := model.EmbedTo(s, buf); err != nil {
				log.Fatalf("embed failed: %v", err)
			}
			if len(buf) > 0 {
				sink += buf[0]
			}
		}
	}
	elapsed := time.Since(benchStart)
	avg := float64(elapsed) / float64(totalCalls) / float64(time.Millisecond)
	throughput := float64(totalCalls) / elapsed.Seconds()

	fmt.Printf("Done. Total calls: %d\n", totalCalls)
	fmt.Printf("Total time: %.3f s\n", elapsed.Seconds())
	fmt.Printf("Average per embedding: %.3f ms (computed: total_time/total_calls)\n", avg)
	fmt.Printf("Throughput: %.1f embeddings/s (computed: total_calls/total_time)\n", throughput)

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatalf("could not create mem profile: %v", err)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write mem profile: %v", err)
		}
		_ = f.Close()
	}

	// Prevent compiler from optimizing away work
	if sink == 42 {
		fmt.Println("sink", sink)
	}
}
