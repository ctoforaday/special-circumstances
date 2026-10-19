package migrate

import (
	"regexp"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE IDS A RECORD ONCE HELD ONLY IN PROSE BECOME FIELDS, ONCE, HERE.
//
// Before epoch 33 the gaps a declaration was about were read out of its holding's text, and the
// anchors a closure was verified with out of its --verified-with words, each by the id's shape at
// every read. Both are fields now — Declare.gap_ids and Close.anchors — refused at the write unless
// the record holds each id. A source whose schema has no table for the field predates it, and its
// ids are read from the same text the live reader read them from, once, keeping each the migrated
// record holds at that act's place; an id the prose names and the record does not hold stays in the
// prose and is no field, which is what the live reader's `not on the board` said of it. A source
// that has the table states the field, and its value is carried as written: prose is never read
// for an id again.

var (
	gapIDInProse    = regexp.MustCompile(`\b` + anchor.IDPattern("gap") + `\b`)
	anchorIDInProse = regexp.MustCompile(`\b` + anchor.IDPattern() + `\b`)
)

// listHolder is a source that can say whether its schema holds a list field — a table, empty or not.
type listHolder interface {
	HoldsList(word, field string) bool
}

// predates is whether src's schema has no table for word's list field: a shard-era source has none.
func predates(src Source, word, field string) bool {
	lh, ok := src.(listHolder)
	return !ok || !lh.HoldsList(word, field)
}

// proseRefs fills a body's id-list field from its prose when the source predates the field, with
// each id the record q holds, in order of first mention.
func proseRefs(src Source, q recordsql.Querier, body any) error {
	switch b := body.(type) {
	case *recordpb.Declare:
		if predates(src, "declare", "gap_ids") {
			ids, err := held(q, gapIDInProse.FindAllString(b.GetHolding(), -1))
			if err != nil {
				return err
			}
			b.GapIds = ids
		}
	case *recordpb.Close:
		if predates(src, "close", "anchors") {
			ids, err := held(q, anchorIDInProse.FindAllString(b.GetAnchorTool(), -1))
			if err != nil {
				return err
			}
			b.Anchors = ids
		}
	}
	return nil
}

// held is ids, deduplicated, keeping each an act on q's record carries — the write's own test
// (record.HoldsAnchor; a gap is held when a mint carries it).
func held(q recordsql.Querier, ids []string) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		ok, err := record.HoldsAnchor(q, id)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}
