package hybridsearch

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"unicode"

	_ "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

const (
	// RRFK is the standard Reciprocal Rank Fusion constant.
	RRFK = 60.0
	// SearchNotice is the mandatory notice stating that hybrid search results are cut and ranked.
	SearchNotice = "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote (or --match for a regex)."
)

// SearchScope says a ranked hybrid search's list is ranked and cut, and how much it ranked.
type SearchScope struct {
	RankedAndCut   bool   `json:"ranked_and_cut"`
	Returned       int    `json:"returned"`
	SearchedChunks int    `json:"searched_chunks"`
	Notice         string `json:"notice"`
}

// UnsearchedSource records a cached source that could not be searched.
type UnsearchedSource struct {
	Sha    string `json:"sha256"`
	URL    string `json:"url"`
	Reason string `json:"reason"`
}

// Item is an indexable text chunk with a dense vector representation.
type Item struct {
	Index  int
	Text   string
	Vector []float32
}

// ScoredItem is a chunk scored by Reciprocal Rank Fusion.
type ScoredItem struct {
	Index    int
	Score    float32
	BM25Rank int
	VecRank  int
}

// FuseRRF executes hybrid BM25 + dense vector search fused via RRF (k=60).
func FuseRRF(items []Item, query string, queryVec []float32, topK int) ([]ScoredItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	if topK <= 0 {
		topK = 10
	}

	// 1. BM25 ranking via in-memory SQLite FTS5
	bm25Ranks, err := rankBM25(items, query)
	if err != nil {
		return nil, fmt.Errorf("hybridsearch: bm25: %w", err)
	}

	// 2. Vector ranking via cosine similarity
	type vecScore struct {
		index int
		cos   float32
	}
	vecScores := make([]vecScore, len(items))
	for i, it := range items {
		vecScores[i] = vecScore{
			index: it.Index,
			cos:   vecembed.Cosine(queryVec, it.Vector),
		}
	}
	sort.Slice(vecScores, func(i, j int) bool {
		return vecScores[i].cos > vecScores[j].cos
	})
	vecRanks := make(map[int]int, len(vecScores))
	for r, vs := range vecScores {
		vecRanks[vs.index] = r + 1
	}

	// 3. RRF fusion (k=60)
	scored := make([]ScoredItem, len(items))
	for i, it := range items {
		var score float64
		bRank, ok := bm25Ranks[it.Index]
		if ok {
			score += 1.0 / (RRFK + float64(bRank))
		}
		vRank := vecRanks[it.Index]
		score += 1.0 / (RRFK + float64(vRank))

		scored[i] = ScoredItem{
			Index:    it.Index,
			Score:    float32(score),
			BM25Rank: bRank,
			VecRank:  vRank,
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		return scored[i].Index < scored[j].Index
	})

	if len(scored) > topK {
		scored = scored[:topK]
	}
	return scored, nil
}

// rankBM25 is each item's BM25 rank for query, from 1, over an in-memory FTS5 index of the items.
// A query with no alphanumeric token has no lexical side: (nil, nil). Every SQLite or FTS5 error is
// returned: an index that failed is no evidence that nothing matched, and a ranking that dropped its
// lexical side would still say it is hybrid.
func rankBM25(items []Item, query string) (map[int]int, error) {
	matchQuery := cleanFTS5Query(query)
	if matchQuery == "" {
		return nil, nil
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// One connection: an in-memory database is per connection, so the index and the query share it.
	db.SetMaxOpenConns(1)
	return rankBM25In(db, items, matchQuery)
}

// rankBM25In indexes items in db and ranks what matchQuery matches; either step's error is returned.
func rankBM25In(db *sql.DB, items []Item, matchQuery string) (map[int]int, error) {
	if err := indexBM25(db, items); err != nil {
		return nil, err
	}
	return matchBM25(db, matchQuery)
}

// indexBM25 builds the FTS5 table the lexical side ranks: one row per item, rowid its Index+1.
func indexBM25(db *sql.DB, items []Item) error {
	if _, err := db.Exec("CREATE VIRTUAL TABLE chunks USING fts5(text, tokenize='porter unicode61');"); err != nil {
		return err
	}
	stmt, err := db.Prepare("INSERT INTO chunks(rowid, text) VALUES(?, ?);")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, it := range items {
		if _, err := stmt.Exec(int64(it.Index+1), it.Text); err != nil {
			return err
		}
	}
	return nil
}

// matchBM25 runs matchQuery against the table indexBM25 built and ranks what it matches.
func matchBM25(db *sql.DB, matchQuery string) (map[int]int, error) {
	rows, err := db.Query("SELECT rowid FROM chunks WHERE chunks MATCH ? ORDER BY rank;", matchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ranks := make(map[int]int)
	for rank := 1; rows.Next(); rank++ {
		var rowid int64
		if err := rows.Scan(&rowid); err != nil {
			return nil, err
		}
		ranks[int(rowid-1)] = rank
	}
	return ranks, rows.Err()
}

// cleanFTS5Query is query's alphanumeric tokens, lowercased, each quoted as an FTS5 string and
// joined by OR, so no token is read as FTS5 syntax. "" when it has none.
func cleanFTS5Query(query string) string {
	var terms []string
	var cur strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(unicode.ToLower(r))
		} else {
			if cur.Len() > 0 {
				terms = append(terms, cur.String())
				cur.Reset()
			}
		}
	}
	if cur.Len() > 0 {
		terms = append(terms, cur.String())
	}
	if len(terms) == 0 {
		return ""
	}
	for i, t := range terms {
		terms[i] = `"` + t + `"`
	}
	return strings.Join(terms, " OR ")
}
