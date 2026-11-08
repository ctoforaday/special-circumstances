package hybridsearch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChunkSourceShort(t *testing.T) {
	text := "This is a short source text with just a few words."
	chunks := ChunkSource(text, nil)
	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1", len(chunks))
	}
	c := chunks[0]
	if c.Index != 0 {
		t.Fatalf("Index = %d, want 0", c.Index)
	}
	if c.StartByte != 0 || c.EndByte != len(text) {
		t.Fatalf("StartByte=%d, EndByte=%d, want 0, %d", c.StartByte, c.EndByte, len(text))
	}
	if c.Text != text {
		t.Fatalf("Text = %q, want %q", c.Text, text)
	}
	if c.Page != 0 {
		t.Fatalf("Page = %d, want 0", c.Page)
	}
}

func TestChunkSourceStrideAndPages(t *testing.T) {
	// Generate text with 400 words
	words := make([]string, 400)
	for i := range words {
		words[i] = "token"
	}
	text := strings.Join(words, " ")
	pageMap := []int{0, 500, 1500}

	chunks := ChunkSource(text, pageMap)
	// 400 words with chunkSize 200, stride 160:
	// Chunk 0: 0..200
	// Chunk 1: 160..360
	// Chunk 2: 320..400
	if len(chunks) != 3 {
		t.Fatalf("len(chunks) = %d, want 3", len(chunks))
	}
	for i, c := range chunks {
		if c.Index != i {
			t.Fatalf("chunk %d Index = %d", i, c.Index)
		}
		if text[c.StartByte:c.EndByte] != c.Text {
			t.Fatalf("chunk %d text mismatch with byte offsets", i)
		}
	}
}

func TestChunkReport(t *testing.T) {
	reportMD := `# Header 1

First paragraph with an anchor C-12345678.

Second paragraph without an anchor.

## Header 2

Third paragraph under header 2.
`
	chunks := ChunkReport(reportMD)
	if len(chunks) != 3 {
		t.Fatalf("len(chunks) = %d, want 3", len(chunks))
	}
	if chunks[0].Heading != "# Header 1" || chunks[0].Anchor != "C-12345678" {
		t.Fatalf("chunk 0 mismatch: heading=%q, anchor=%q", chunks[0].Heading, chunks[0].Anchor)
	}
	if chunks[1].Heading != "# Header 1" || chunks[1].Anchor != "" {
		t.Fatalf("chunk 1 mismatch: heading=%q, anchor=%q", chunks[1].Heading, chunks[1].Anchor)
	}
	if chunks[2].Heading != "## Header 2" || chunks[2].Text != "Third paragraph under header 2." {
		t.Fatalf("chunk 2 mismatch: heading=%q, text=%q", chunks[2].Heading, chunks[2].Text)
	}
}

func TestCacheReadWrite(t *testing.T) {
	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "abc123.vec")
	modelID := "test-model@v1"

	chunks := []CachedChunk{
		{
			Index:     0,
			StartByte: 10,
			EndByte:   100,
			Page:      1,
			Vector:    make([]float32, 384),
		},
		{
			Index:     1,
			StartByte: 90,
			EndByte:   200,
			Page:      2,
			Vector:    make([]float32, 384),
		},
	}
	chunks[0].Vector[0] = 0.5
	chunks[1].Vector[383] = 0.99

	if err := WriteCache(cachePath, modelID, chunks); err != nil {
		t.Fatalf("WriteCache failed: %v", err)
	}

	// Read with matching model
	got, err := ReadCache(cachePath, modelID)
	if err != nil {
		t.Fatalf("ReadCache failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].StartByte != 10 || got[0].Vector[0] != 0.5 {
		t.Fatalf("chunk 0 mismatch: %+v", got[0])
	}
	if got[1].EndByte != 200 || got[1].Vector[383] != 0.99 {
		t.Fatalf("chunk 1 mismatch: %+v", got[1])
	}

	// Read with different model returns ErrCacheMiss
	_, err = ReadCache(cachePath, "other-model@v2")
	if err != ErrCacheMiss {
		t.Fatalf("ReadCache(different model) = %v, want ErrCacheMiss", err)
	}

	// Non-existent file returns ErrCacheMiss
	_, err = ReadCache(filepath.Join(tmp, "nonexistent.vec"), modelID)
	if err != ErrCacheMiss {
		t.Fatalf("ReadCache(nonexistent) = %v, want ErrCacheMiss", err)
	}
}

func TestFuseRRF(t *testing.T) {
	items := []Item{
		{
			Index:  0,
			Text:   "apple pie recipe with cinnamon and sugar",
			Vector: []float32{1, 0, 0},
		},
		{
			Index:  1,
			Text:   "quantum computing and quantum algorithms",
			Vector: []float32{0, 1, 0},
		},
		{
			Index:  2,
			Text:   "superposition in qubits and quantum mechanics",
			Vector: []float32{0, 0.9, 0.1},
		},
	}

	// Query has vector [0, 1, 0] (matches 1 and 2 closely) and text "quantum"
	query := "quantum computing"
	queryVec := []float32{0, 1, 0}

	scored, err := FuseRRF(items, query, queryVec, 10)
	if err != nil {
		t.Fatalf("FuseRRF failed: %v", err)
	}
	if len(scored) != 3 {
		t.Fatalf("len(scored) = %d, want 3", len(scored))
	}

	// Item 1 has top BM25 rank and top vector rank -> must be #1
	if scored[0].Index != 1 {
		t.Fatalf("top item = %d, want 1 (scored: %+v)", scored[0].Index, scored)
	}
	// Item 2 has second vector rank and second BM25 rank -> must be #2
	if scored[1].Index != 2 {
		t.Fatalf("second item = %d, want 2 (scored: %+v)", scored[1].Index, scored)
	}
	// Item 0 is apple pie -> must be #3
	if scored[2].Index != 0 {
		t.Fatalf("third item = %d, want 0", scored[2].Index)
	}
}

func TestFuseRRFEdgeCases(t *testing.T) {
	// Empty items
	res, err := FuseRRF(nil, "test", []float32{1, 0}, 5)
	if err != nil || len(res) != 0 {
		t.Fatalf("FuseRRF(nil) = %v, %v, want empty, nil", res, err)
	}

	items := []Item{
		{Index: 0, Text: "single item text", Vector: []float32{1, 0, 0}},
	}
	// topK <= 0 defaults to 10
	res, err = FuseRRF(items, "single", []float32{1, 0, 0}, 0)
	if err != nil || len(res) != 1 {
		t.Fatalf("FuseRRF(topK=0) = %v, %v, want 1 item (defaulted to 10), nil", res, err)
	}

	// topK = 1 on single item
	res, err = FuseRRF(items, "single", []float32{1, 0, 0}, 1)
	if err != nil || len(res) != 1 {
		t.Fatalf("FuseRRF(topK=1) = %v, %v, want 1 item, nil", res, err)
	}
	if res[0].Index != 0 {
		t.Fatalf("res[0].Index = %d, want 0", res[0].Index)
	}
}

func TestReadCacheCorrupted(t *testing.T) {
	tmp := t.TempDir()

	// 1. Corrupted magic
	badMagic := filepath.Join(tmp, "badmagic.vec")
	if err := os.WriteFile(badMagic, []byte("NOT_FVEC_HEADER_DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadCache(badMagic, "m@v1"); err != ErrCacheMiss {
		t.Fatalf("ReadCache(badMagic) = %v, want ErrCacheMiss", err)
	}

	// 2. Truncated header
	shortFile := filepath.Join(tmp, "short.vec")
	if err := os.WriteFile(shortFile, []byte("FV"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadCache(shortFile, "m@v1"); err != ErrCacheMiss {
		t.Fatalf("ReadCache(shortFile) = %v, want ErrCacheMiss", err)
	}
}
