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

// SearchScope describes the scope and truncation of a ranked hybrid search.
type SearchScope struct {
	RankedAndCut      bool               `json:"ranked_and_cut"`
	Returned          int                `json:"returned"`
	SearchedChunks    int                `json:"searched_chunks"`
	SearchedSources   int                `json:"searched_sources,omitempty"`
	UnsearchedSources []UnsearchedSource `json:"unsearched_sources,omitempty"`
	Notice            string             `json:"notice"`
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

func rankBM25(items []Item, query string) (map[int]int, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if _, err := db.Exec("CREATE VIRTUAL TABLE chunks USING fts5(text, tokenize='porter unicode61');"); err != nil {
		return nil, err
	}

	stmt, err := db.Prepare("INSERT INTO chunks(rowid, text) VALUES(?, ?);")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for _, it := range items {
		if _, err := stmt.Exec(int64(it.Index+1), it.Text); err != nil {
			return nil, err
		}
	}

	matchQuery := cleanFTS5Query(query)
	if matchQuery == "" {
		return map[int]int{}, nil
	}

	rows, err := db.Query("SELECT rowid, bm25(chunks) FROM chunks WHERE chunks MATCH ? ORDER BY rank;", matchQuery)
	if err != nil {
		// FTS5 query syntax error or no match
		return map[int]int{}, nil
	}
	defer rows.Close()

	ranks := make(map[int]int)
	rank := 1
	for rows.Next() {
		var rowid int64
		var bm25 float64
		if err := rows.Scan(&rowid, &bm25); err != nil {
			return nil, err
		}
		ranks[int(rowid-1)] = rank
		rank++
	}
	return ranks, rows.Err()
}

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
	return strings.Join(terms, " OR ")
}
