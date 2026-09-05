package planguard

import "strings"

// Defects is the subset of Findings that indicate a MISSING INDEX rather than a design cost.
//
// # Why not every scan over a growing table
//
// The first draft flagged all of them and was almost entirely false positives. Driven over the
// real read paths, every scan the record performs today falls into one of these, and not one is
// an index defect:
//
//	SELECT id, seat_id, round, ts, type, key FROM events ORDER BY id   -- the full replay
//	SELECT "event_id", "gap_id", … FROM "mint"                        -- a bulk detail read
//	SELECT (SELECT count(*) FROM "verify"), (SELECT count(*) FROM "cite")
//	SELECT "verdict" FROM "outcome" ORDER BY "event_id" DESC LIMIT 1
//	SCAN mint_supersedes USING INDEX sqlite_autoindex_mint_supersedes_1
//	SCAN events USING COVERING INDEX events_round
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
// and still walks all of it. Today that set is empty, which is the same answer #684 F8 reached by
// hand — the difference is that this one re-asks on every run.
//
// # The limit, stated rather than papered over
//
// The constraint test is read off the statement text, so a multi-table statement that constrains
// one table while legitimately scanning another would be reported. That is a LOUD false positive:
// someone reads it, and either finds a real defect or learns the rule needs sharpening. There is
// deliberately NO allowlist to silence one — an allowlist is the hand-kept list this package
// exists to avoid, and none has been earned, because the board is currently clean. The first
// genuine false positive is the moment to decide what shape the exception takes, with the case in
// hand rather than a slot waiting to be filled.
//
// There is a SECOND limit, and it is the one worth knowing before trusting this: dropping an
// index that only ever served an UNCONSTRAINED read is invisible here. `SELECT DISTINCT "round"
// FROM "events" ORDER BY "round"` walks events_round as a covering index today; delete that index
// and the plan becomes a bare `SCAN events`, with no WHERE to make it a defect, and this stays
// green while the query got slower. Catching that needs a different instrument — a recorded
// baseline of each statement's plan, failing when a plan gets WORSE — which is a plan-golden, not
// a rule. Named rather than papered over, because "the guard is green" must not be read as more
// than it says.
func Defects(findings []Finding) []Finding {
	var out []Finding
	for _, f := range findings {
		if indexTraversal(f.Detail) || !constrains(f.Statement) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// indexTraversal reports whether the plan reached the rows through an index. SQLite writes
// "SCAN t USING INDEX i" and "SCAN t USING COVERING INDEX i"; both walk index order, and the
// covering form never touches the table at all.
func indexTraversal(detail string) bool {
	return strings.Contains(strings.ToUpper(detail), " USING ")
}

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
