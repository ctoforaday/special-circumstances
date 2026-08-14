package blue

import (
	"strings"
	"testing"
)

// FuzzPlanEdit proves the core lockdown invariant: whenever `blue edit` accepts a span
// replacement, it PRESERVES every finding-marker. It splices a marker into a clean body,
// draws an arbitrary old span from the body, and asserts that a SUCCESSFUL planEdit never
// drops the marker (a marker-spanning or absent span must instead be rejected). A failure
// is a real hole in the span/marker interplay, not a bad input.
func FuzzPlanEdit(f *testing.F) {
	f.Add("The cost is rising over time now.", uint(10), uint(4), uint(12))
	f.Add("Values lose precision here in practice.", uint(0), uint(0), uint(6))
	f.Add("A short clause, with punctuation, inside.", uint(20), uint(2), uint(15))

	f.Fuzz(func(t *testing.T, body string, markerPos, oldStart, oldLen uint) {
		// Clean bodies only: no pre-existing annotations, no newlines (paragraph breaks the
		// matcher legitimately refuses), printable.
		if strings.Contains(body, "<!--") || strings.Contains(body, "[^") || strings.ContainsRune(body, '\n') || len(body) == 0 {
			return
		}
		for i := 0; i < len(body); i++ {
			if body[i] < 0x20 {
				return
			}
		}
		mp := int(markerPos) % (len(body) + 1)
		report := body[:mp] + "<!--fx:f-seed01-->" + body[mp:]

		os := int(oldStart) % len(body)
		ol := int(oldLen) % (len(body) - os + 1)
		old := body[os : os+ol]

		next, err := planEdit(report, old, "REPLACEMENT")
		if err != nil {
			return // a mis-quote or a marker-spanning span → reject is always acceptable
		}
		// SUCCESS ⇒ the marker survives — both as a set-membership id and as a literal token.
		if droppedMarker(report, next) != "" {
			t.Fatalf("planEdit succeeded but dropped a marker\n report=%q\n old=%q\n next=%q", report, old, next)
		}
		if !strings.Contains(next, "<!--fx:f-seed01-->") {
			t.Fatalf("marker token mangled\n report=%q\n old=%q\n next=%q", report, old, next)
		}
	})
}
