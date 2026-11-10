package hybridsearch

import (
	"database/sql"
	"slices"
	"strings"
	"testing"
)

// A PARAGRAPH IS RANKED BY THE WORDS A READER SEES. Every anchor kind's token comes out of the text
// BM25 indexes and the embedder reads, and every id it carried is kept, in order, the first also in
// Anchor.
func TestChunkReportStripsAnchorsAndExtractsAllTypes(t *testing.T) {
	report := "# Findings\n\n" +
		"Costs are rising.<!--cite:C-0a1b2c3d--> The trend holds<!--computation:K-0e0f1a2b-->.\n" +
		"It was contested.<!--fx:F-11223344--><!--gap:G-55667788-->\n\n" +
		"<!--gap:G-99aabbcc-->\n\n" +
		"## Plain\n\nNo anchors here.\n"
	chunks := ChunkReport(report)
	if len(chunks) != 2 {
		t.Fatalf("ChunkReport = %d chunks, want 2 (an anchor-only paragraph has no words to rank): %+v", len(chunks), chunks)
	}
	c := chunks[0]
	if want := "Costs are rising. The trend holds.\nIt was contested."; c.Text != want {
		t.Errorf("Text = %q, want %q", c.Text, want)
	}
	if strings.Contains(c.Text, "<!--") || strings.Contains(c.Text, "-->") {
		t.Errorf("Text still carries anchor markup: %q", c.Text)
	}
	if want := []string{"C-0a1b2c3d", "K-0e0f1a2b", "F-11223344", "G-55667788"}; !slices.Equal(c.Anchors, want) || c.Anchor != want[0] {
		t.Errorf("Anchors = %v, Anchor = %q; want %v, %q", c.Anchors, c.Anchor, want, want[0])
	}
	if c.Heading != "# Findings" || c.StartLine != 3 || c.EndLine != 4 {
		t.Errorf("chunk 0 at %q lines %d..%d, want # Findings 3..4", c.Heading, c.StartLine, c.EndLine)
	}
	if p := chunks[1]; p.Heading != "## Plain" || p.Text != "No anchors here." || p.Anchor != "" || p.Anchors != nil {
		t.Errorf("chunk 1 = %+v, want the plain paragraph with no anchors", p)
	}
	// What BM25 indexes is the visible text: an anchor's tag or id is no term.
	items := []Item{{Index: 0, Text: c.Text}, {Index: 1, Text: chunks[1].Text}}
	for _, q := range []string{"cite", "0a1b2c3d", "computation"} {
		ranks, err := rankBM25(items, q)
		if err != nil || len(ranks) != 0 {
			t.Errorf("rankBM25(%q) = %v, %v; want no match in the visible text", q, ranks, err)
		}
	}
}

// AN INDEX THAT FAILED IS NO EVIDENCE THAT NOTHING MATCHED. Every SQLite or FTS5 error reaches the
// caller; only a query with no token to match has an empty lexical side, and that is not an error.
func TestRankBM25PropagatesSQLiteErrors(t *testing.T) {
	open := func(t *testing.T) *sql.DB {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		t.Cleanup(func() { db.Close() })
		return db
	}
	items := []Item{{Index: 0, Text: "moderate and rising costs"}}

	// The MATCH fails: the table it reads is no FTS5 table.
	db := open(t)
	if _, err := db.Exec("CREATE TABLE chunks(text TEXT)"); err != nil {
		t.Fatal(err)
	}
	if ranks, err := matchBM25(db, cleanFTS5Query("rising")); err == nil {
		t.Errorf("matchBM25 on a table MATCH cannot read = %v, nil; want the SQLite error", ranks)
	}
	// The index fails: the table exists already.
	if err := indexBM25(db, items); err == nil {
		t.Error("indexBM25 over an existing table = nil; want the SQLite error")
	}
	// The index fails: the database is closed.
	closed := open(t)
	closed.Close()
	if err := indexBM25(closed, items); err == nil {
		t.Error("indexBM25 on a closed database = nil; want the error")
	}

	if ranks, err := rankBM25(items, " -- !! "); err != nil || ranks != nil {
		t.Errorf("rankBM25 with no token = %v, %v; want no lexical side and no error", ranks, err)
	}
	if ranks, err := rankBM25(items, `rising" OR NEAR(`); err != nil || ranks[0] != 1 {
		t.Errorf("rankBM25 with FTS5 syntax in the query = %v, %v; want its words matched as words", ranks, err)
	}
}
