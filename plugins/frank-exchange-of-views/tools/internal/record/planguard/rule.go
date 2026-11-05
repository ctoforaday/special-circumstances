package planguard

import "strings"

// The constrained-scan shape (ConstrainedScan) is a statement that narrows what it reads and
// still walks the whole of a growing table — the signature of a MISSING INDEX rather than a
// design cost. The scan is read off the plan tree by heapScan, so "SCAN p" is judged as the table
// `p` is bound to (Scope) and a bare "SEARCH t" with no index counts as the full scan it is. The
// same min/max loop over a view or CTE SQLite computes as a subplan ("SEARCH v" beside
// "CO-ROUTINE v") counts when the subplan reads a growing table (minMaxOverSubplan).
//
// # Why not every scan over a growing table
//
// The first draft flagged all of them and was almost entirely false positives. Driven over the
// real read paths, every scan the record performs today falls into one of these, and not one is
// an index defect:
//
//	SELECT id, seat_id, ts, type, key FROM events ORDER BY id   -- the full replay
//	SELECT "event_id", "gap_id", … FROM "mint"                        -- a bulk detail read
//	SELECT (SELECT count(*) FROM "verify"), (SELECT count(*) FROM "cite")
//	SELECT "verdict" FROM "outcome" ORDER BY "event_id" DESC LIMIT 1
//	SCAN mint_supersedes USING INDEX sqlite_autoindex_mint_supersedes_1
//	SCAN events USING COVERING INDEX events_type
//
// Two distinctions fall out of that list, and both are read off the plan or the statement rather
// than kept in a list here:
//
//  1. "SCAN x USING [COVERING] INDEX y" is an INDEX TRAVERSAL, not a heap scan. SQLite says SCAN
//     because it visits every entry, but it does so in index order and often without touching the
//     table. Counting those made two thirds of the report noise.
//
//  2. AN UNCONSTRAINED READ OF A WHOLE TABLE MUST SCAN IT. `count(*)` over a table, a full
//     replay, a bulk read of every mint — no index can improve any of them, and an index is not
//     what they are missing. Their cost is real and it is a DESIGN question (it is what #684 F9
//     is about: routing count-only readers at the views instead of folding the record); it is not
//     the question this guard asks.
//
// What is left is the signature of a dropped index: a statement that CONSTRAINS a growing table
// and still walks all of it. The findings it reports today are in internal/record/testdata/
// shapes.golden, kind constrained-scan, and a fix removes its line.
//
// # The limit, stated rather than papered over
//
// The constraint test is read off the statement text, so a multi-table statement that constrains
// one table while legitimately scanning another would be reported. That is a LOUD false positive:
// someone reads it, and either finds a real defect or learns the rule needs sharpening. There is
// deliberately NO allowlist to silence one — an allowlist is the hand-kept list this package
// exists to avoid. A false positive sits in the generated baseline, visible, until the rule is
// sharpened or the statement is rewritten. `SELECT MIN("id") FROM "events"` inside a statement
// with a WHERE is one: its "SEARCH events" reads one row, and the rule cannot tell it from a walk
// (heapScan says why).
//
// There is a SECOND limit, and it is the one worth knowing before trusting this: dropping an
// index that only ever served an UNCONSTRAINED read is invisible here. `SELECT DISTINCT "type"
// FROM "events"` walks events_type as a covering index today; delete that index
// and the plan becomes a bare `SCAN events`, with no WHERE to make it a defect, and this stays
// green while the query got slower. That change is caught by internal/record/testdata/
// plans.golden, which records every planned statement's whole plan; the per-row and repeated-work
// shapes are read off the plan tree by Analyzer.Shapes against a committed baseline (Suite).
// This rule stays what it is: the missing-index signature on a constrained statement.

// constrains reports whether the statement narrows what it reads. A WHERE clause or a join
// predicate means the author asked for a subset — and a subset that still costs a full walk is
// what a missing index looks like.
//
// Word-boundary matched on the uppercased text: "WHERE" inside a quoted string or a column named
// `wherefore` must not make an unconstrained bulk read look constrained, because that direction
// manufactures a defect out of a correct query.
func constrains(statement string) bool {
	up := strings.ToUpper(statement)
	for _, kw := range []string{"WHERE", "JOIN"} {
		if containsWord(up, kw) {
			return true
		}
	}
	return false
}

func containsWord(haystack, word string) bool {
	for i := 0; ; {
		j := strings.Index(haystack[i:], word)
		if j < 0 {
			return false
		}
		start := i + j
		end := start + len(word)
		if !isWordByte(haystack, start-1) && !isWordByte(haystack, end) {
			return true
		}
		i = end
	}
}

func isWordByte(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}
