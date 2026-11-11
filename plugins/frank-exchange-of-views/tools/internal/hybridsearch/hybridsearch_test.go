package hybridsearch

import (
	"database/sql"
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

// AN INDEX THAT FAILED REACHES THE SEARCH'S CALLER. Two items sharing an Index share an FTS5 rowid,
// so the index refuses the second; FuseRRF returns that error rather than a ranking that lost its
// lexical side.
func TestFuseRRFReturnsTheIndexError(t *testing.T) {
	items := []Item{
		{Index: 0, Text: "moderate and rising costs", Vector: []float32{1, 0}},
		{Index: 0, Text: "rising harbour fees", Vector: []float32{0, 1}},
	}
	if scored, err := FuseRRF(items, "rising", []float32{1, 0}, 10); err == nil {
		t.Fatalf("FuseRRF over items sharing an Index = %+v, nil; want the index's SQLite error", scored)
	}
}

// A MATCH THAT FAILED REACHES THE CALLER. A temporary table named chunks shadows the FTS5 table the
// index creates: the index succeeds, its rows go into the temporary table, and MATCH cannot read it.
func TestRankBM25ReturnsTheMatchError(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("CREATE TEMP TABLE chunks(text TEXT)"); err != nil {
		t.Fatal(err)
	}
	items := []Item{{Index: 0, Text: "moderate and rising costs"}}
	ranks, err := rankBM25In(db, items, cleanFTS5Query("rising"))
	var indexed int
	if qerr := db.QueryRow("SELECT count(*) FROM main.sqlite_master WHERE name = 'chunks'").Scan(&indexed); qerr != nil || indexed != 1 {
		t.Fatalf("the FTS5 table was not created (%d, %v); the test needs the index to succeed and the match to fail", indexed, qerr)
	}
	if err == nil {
		t.Errorf("rankBM25In whose MATCH fails = %v, nil; want the SQLite error", ranks)
	}
}

// RRF'S CONSTANT DECIDES THE ORDER. Item 0 is BM25's first and the vectors' fourth; item 1 is
// second on both. With k = 60 item 1 leads (2/62 > 1/61 + 1/64); a k as small as 1 would put item 0
// first (1/2 + 1/5 > 2/3).
func TestFuseRRFRanksWithKSixty(t *testing.T) {
	items := []Item{
		{Index: 0, Text: "harbour harbour harbour", Vector: []float32{0.7, 0.714}},
		{Index: 1, Text: "the harbour is one of many places costs rose in the decade before", Vector: []float32{0.9, 0.436}},
		{Index: 2, Text: "dredging", Vector: []float32{1, 0}},
		{Index: 3, Text: "tolls", Vector: []float32{0.8, 0.6}},
	}
	scored, err := FuseRRF(items, "harbour", []float32{1, 0}, 10)
	if err != nil {
		t.Fatal(err)
	}
	byIndex := map[int]ScoredItem{}
	for _, s := range scored {
		byIndex[s.Index] = s
	}
	if a, b := byIndex[0], byIndex[1]; a.BM25Rank != 1 || a.VecRank != 4 || b.BM25Rank != 2 || b.VecRank != 2 {
		t.Fatalf("ranks = item 0 %+v, item 1 %+v; the test needs BM25 1/2 and vectors 4/2", a, b)
	}
	if got := []int{scored[0].Index, scored[1].Index}; got[0] != 1 || got[1] != 0 {
		t.Errorf("top two = %v, want [1 0] (k = 60)", got)
	}
	if want := float32(2.0 / 62); scored[0].Score != want {
		t.Errorf("item 1's score = %v, want 2/(60+2) = %v", scored[0].Score, want)
	}
}
