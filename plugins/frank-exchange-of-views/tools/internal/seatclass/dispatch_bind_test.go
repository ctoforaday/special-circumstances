package seatclass

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// TestDebateDispatchBindsToSeatClass binds the ONE seat->class map (SeatClass) to debate.js's
// actual dispatch. debate.js spreads `...bulk` or `...judgment` into each agent() call's opts;
// this reads debate.js SOURCE and asserts every dispatch spreads the tier its seat's class maps
// to. If a dispatch ever uses the wrong tier, or a new seat is added without a class, this fails
// here rather than on a live run. debate.js runs under goja and is not importable, so a
// source-lint is the honest binding — and it keeps the map single-source, now that Go owns it.
//
// This REPLACES the deleted tests/simulator/debate-dispatch.test.mjs (the JS SEAT_CLASS oracle
// was a second copy of this map; the copy is gone, the bind now runs from the owning side).
func TestDebateDispatchBindsToSeatClass(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read debate.js — the dispatch source this binds against: %v", err)
	}

	// Longest key first so `judge-petition`/`judge-terminal` win over the `judge` prefix.
	seatKeys := make([]string, 0, len(SeatClass))
	for k := range SeatClass {
		seatKeys = append(seatKeys, k)
	}
	sort.Slice(seatKeys, func(i, j int) bool { return len(seatKeys[i]) > len(seatKeys[j]) })

	stemOf := func(labelTemplate string) string {
		head := strings.TrimSpace(strings.SplitN(labelTemplate, "·", 2)[0]) // e.g. `red-lens-${role}-r${round}`
		for _, s := range seatKeys {
			if strings.HasPrefix(head, s) {
				return s
			}
		}
		return ""
	}

	// Each dispatch opts object opens `{ ...bulk, label: ` + "`…`" + ` }` or `{ ...judgment, … }`.
	// The label is either a template literal whose head names the seat, or `labelFor(<seat>)` —
	// the dispatch loop's per-seat sitting counter, whose argument is the seat id. The one
	// COMPUTED argument is the lens party's (`labelFor(p.seat_id)`), and that site is recognisable
	// by the dispatch table it spreads (`...LENS_DISPATCH[p.seat_id]`): every lens is `red-lens`.
	bt := "`"
	re := regexp.MustCompile(`\{\s*\.\.\.(bulk|judgment),\s*label:\s*(?:` + bt + `([^` + bt + `]+)` + bt + `|labelFor\(([^)]*)\)([^}]*))`)
	ms := re.FindAllStringSubmatch(string(src), -1)

	seen := map[string]bool{}
	for _, m := range ms {
		spread, labelTemplate := m[1], m[2]
		if labelTemplate == "" {
			arg := strings.Trim(strings.TrimSpace(m[3]), "'\"")
			if arg == "p.seat_id" {
				if !strings.Contains(m[4], "LENS_DISPATCH") {
					t.Errorf("dispatch labelFor(p.seat_id) at %q spreads no LENS_DISPATCH row — a computed seat id this test cannot read", m[0])
					continue
				}
				arg = "red-lens"
			}
			labelTemplate = arg
		}
		stem := stemOf(labelTemplate)
		if stem == "" {
			t.Errorf("dispatch label %q matched no known seat", labelTemplate)
			continue
		}
		seen[stem] = true
		if got := ClassOf(stem); got != spread {
			t.Errorf("seat %q dispatches with ...%s but SeatClass says %s", stem, spread, got)
		}
	}

	if n := len(ms); n < 10 {
		t.Errorf("expected at least 10 dispatch sites in debate.js, found %d", n)
	}

	// Completeness: no dead SeatClass key — every mapped seat is actually dispatched. Archived
	// runs' transcripts still carry the heads of seats that no longer exist (`Red merge, round N`);
	// they classify as `other`, which is the visible bucket, and the record they were migrated to
	// (plans/roundless.md §III.A.5) is where their identity lives now — not in this table.
	for s := range SeatClass {
		if !seen[s] {
			t.Errorf("SeatClass key %q is never dispatched in debate.js", s)
		}
	}
}
