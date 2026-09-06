package anchortext

import (
	"strings"
	"testing"
)

// FuzzLocateEnd proves the invariant "a quote of the semantic prose survives arbitrary
// annotation splicing." It takes a clean base sentence, splices an invisible finding
// marker and a footnote reference at fuzz-chosen positions (INSERT only — never deletes
// content), and asserts the un-annotated base still anchors, the offset lands at a
// content boundary (never inside an annotation), and a freshly inserted marker round-
// trips. Because splicing only inserts, the base content is always present in order, so
// a failure is a real matcher bug, not a bad input.
func FuzzLocateEnd(f *testing.F) {
	f.Add("The scheduler is preemptive and fair.", uint16(5))
	f.Add("Costs rise sharply as volume grows over time.", uint16(0))
	f.Add("Reconciliation catches accumulated drift.", uint16(12))
	f.Add("Values (e.g., 0.1, 0.2) lose precision.", uint16(30))

	f.Fuzz(func(t *testing.T, base string, pos uint16) {
		// Only fuzz clean prose. Pre-existing annotations would be skipped (not our test);
		// a newline could introduce a paragraph break the matcher legitimately refuses to
		// cross (covered deterministically in §V.3/§V.4); control chars are not report prose.
		if strings.Contains(base, "<!--") || strings.Contains(base, "[^") {
			return
		}
		if strings.ContainsRune(base, '\n') {
			return
		}
		for i := 0; i < len(base); i++ {
			if base[i] < 0x20 {
				return
			}
		}
		// Skip bases that normalize to nothing (pure punctuation/space) — locateEnd rejects
		// them by design. Pass the RAW base to locateEnd below, exactly as finding.go does;
		// locateEnd normalizes internally (pre-normalizing here would double-normalize).
		if normalizeQuote(base) == "" {
			return
		}

		// Splice a finding marker and a footnote at two distinct BASE offsets. Anchoring
		// both into the original base (never into the already-annotated string) keeps the
		// annotations well-formed and non-nested — the tool likewise only ever inserts at a
		// content boundary, never inside another marker.
		a := int(pos) % (len(base) + 1)
		b := (int(pos)*7 + 3) % (len(base) + 1)
		if a > b {
			a, b = b, a
		}
		annotated := base[:a] + "<!--fx:f-seed01-->" + base[a:b] + "[^fn1]" + base[b:]

		end := locateEnd(annotated, base)
		if end < 0 {
			t.Fatalf("base did not anchor through spliced annotations\n base=%q\n annotated=%q", base, annotated)
		}

		// The invariant: inserting a marker at the returned offset touches ONLY the
		// invisible layer. The offset may sit just before a trailing annotation (a legal
		// boundary), but inserting there must never split a marker/footnote — so stripping
		// the whole annotation layer recovers the same prose as before the insert.
		out := string(insertMarker([]byte(annotated), end, "<!--fx:f-abcdef-->"))
		if !strings.Contains(out, "<!--fx:f-abcdef-->") || reFindingMarker.FindStringSubmatch(out) == nil {
			t.Fatalf("inserted marker did not round-trip in %q", out)
		}
		if stripAnnotations(out) != stripAnnotations(annotated) {
			t.Fatalf("insert at %d split an annotation:\n got  %q\n want %q",
				end, stripAnnotations(out), stripAnnotations(annotated))
		}
	})
}
