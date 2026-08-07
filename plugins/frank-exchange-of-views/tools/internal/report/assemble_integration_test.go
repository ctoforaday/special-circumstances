package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// TestAssembleEndToEnd drives a minimal but real run through the record API and asserts the
// assembled report.md combines the two ownership classes correctly: blue's audited sections
// are lifted verbatim, and the verdict, findings, avenues and debate are composed from the
// event log. It is the driveable check the unit tests around each composer cannot give.
func TestAssembleEndToEnd(t *testing.T) {
	runDir := t.TempDir()

	blue := strings.Join([]string{
		"# Whether the cache is coherent — research report",
		// Blue authors a stale verdict in the preamble — it cannot know the outcome (#79);
		// the embed must strip it rather than park it beside the tool's authoritative stamp.
		"**Verdict:** UNVERIFIED (Round 0)",
		"",
		"## TL;DR",
		"The cache is coherent under the documented invariants; one edge case is unproven.",
		"",
		"## The Catechism",
		"Q1: is it worth our time? Yes — the failure mode is silent corruption.",
		"",
		"## Technical foundations",
		"The invalidation protocol is single-writer.",
		"",
		"## Analysis",
		"From single-writer it follows that reads never observe a torn value.",
		"",
		"## Open questions",
		"Does the eviction path hold under a clock step?",
		"",
		// Blue OVER-AUTHORS a tool-owned section — it must be dropped from the embed
		// (fabrication), not duplicated below the tool's composed version.
		"## Risk Matrix",
		"blue's fabricated risk row that the tool must not echo.",
		"",
		// Blue hand-authors footnotes — DROPPED now (citations are tool-composed from the
		// cite events; a blue-authored bibliography is fabrication).
		"## Footnotes",
		"[^cache]: a citation blue tried to hand-author.",
		"",
	}, "\n")
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	add := func(seatID, typ string, kv ...string) {
		t.Helper()
		if !seen[seatID] {
			if _, _, err := record.RegisterSeat(runDir, seatID); err != nil {
				t.Fatalf("register %s: %v", seatID, err)
			}
			seen[seatID] = true
		}
		p := record.NewPayload()
		for i := 0; i+1 < len(kv); i += 2 {
			p.Set(kv[i], kv[i+1])
		}
		if _, err := record.Append(runDir, seatID, typ, p); err != nil {
			t.Fatalf("append %s/%s: %v", seatID, typ, err)
		}
	}

	// Red mints a gap; the parties take positions; blue records one pursued avenue (an
	// expansion) and one abandoned avenue (an alternative considered); the bench opines;
	// the run's terminal verdict is recorded.
	add("red-merge-r1", "mint", "gap_id", "R1-1", "problem", "eviction races the reader", "location", "cache.go:88",
		"class", "correctness", "likelihood", "medium", "impact", "high",
		"acceptance_check", "race the eviction under -race", "check_kind", "document", "required_fix", "take the read lock in evict")
	add("red-merge-r1", "position", "text", "gap R1-1 stands until the race is shown impossible")
	add("blue-r1", "position", "text", "R1-1 is repaired by ordering the invalidation before the store")
	add("blue-r1", "avenue", "avenue_id", "A1", "status", "pursued", "line", "model-check the two-writer interleaving", "method", "TLA+")
	add("blue-r1", "avenue", "avenue_id", "A2", "status", "abandoned", "line", "rewrite the cache lock-free", "reason", "cost exceeds the benefit at this scale")
	add("judge-r1", "opinion", "gap_id", "R1-1", "disposition", "carried", "principle", "correctness",
		"tension", "cost vs certainty", "review_flag", "false", "rationale", "a model-check is owed before this closes")
	add("judge-terminal", "outcome", "verdict", "CEILING")

	path, err := Assemble(runDir)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	for _, want := range []string{
		// Blue-authored, lifted verbatim.
		"# Whether the cache is coherent — research report",
		"## TL;DR\nThe cache is coherent under the documented invariants",
		"Q1: is it worth our time? Yes",
		"From single-writer it follows that reads never observe a torn value.",
		"Does the eviction path hold under a clock step?",
		// Composed from the record.
		"**Verdict:** CEILING-TERMINATED",
		"## The expansions",
		"model-check the two-writer interleaving",
		"## Alternatives considered",
		"rewrite the cache lock-free",
		"cost exceeds the benefit at this scale",
		"## Red team findings (in full)",
		"R1-1 — eviction races the reader",
		"## The debate",
		"### RED\ngap R1-1 stands",
		"### BLUE\nR1-1 is repaired",
		"R1-1: carried",
		// The reviewer-facing orientation, composed from the board.
		"## Read this first",
		"(R1-1)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("assembled report missing %q\n---\n%s", want, got)
		}
	}

	// Blue over-authored a Risk Matrix, a stale verdict, and hand-authored footnotes; none may
	// be echoed. The tool's "## Risk matrix" is authoritative; blue's fabricated row / stale
	// verdict / hand-authored citation must not appear anywhere. With no cite events on the
	// record, no "## Bibliography" is composed either.
	for _, forbidden := range []string{"blue's fabricated risk row", "UNVERIFIED (Round 0)", "a citation blue tried to hand-author", "## Bibliography"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("blue's fabricated/duplicated content leaked into the report (%q)\n---\n%s", forbidden, got)
		}
	}

	// Fate is not crossed: the pursued avenue is an expansion, not an alternative.
	alt := got[strings.Index(got, "## Alternatives considered"):]
	alt = alt[:strings.Index(alt, "## Open questions")]
	if strings.Contains(alt, "model-check the two-writer") {
		t.Errorf("a pursued avenue leaked into Alternatives considered:\n%s", alt)
	}
}
