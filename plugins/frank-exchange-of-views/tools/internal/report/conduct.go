package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// WHAT ANSWERED THIS RUN, IN THE RUN'S OWN REPORT.
//
// The last unclosed half of #589. Every bulk seat in both 2026-08-23 runs requested
// claude-fable-5 and was answered by claude-opus-4-8 on every turn, and run B's CERTIFIED report
// went out describing its own methodology wrong — "the pairing this run actually used — blue on
// `fable`, red on `sonnet`" — and then reasoned at length about same-vendor bias from that
// premise. The claim was not careless: it was the only model fact available to a seat, which
// reads the run's CONFIGURATION and cannot see what replied to it.
//
// register now measures the serving model at each seat's own first turn (record.SeatModels reads
// it back). This is the surface that puts it in front of the human the report is for.
//
// IT IS TOOL-OWNED, and that is the whole point. Blue is already forbidden from authoring
// tool-owned sections, so the one fact a seat provably cannot know about itself is composed from
// the record instead of asked for in a prompt. A prompt-carried model name would be a second copy
// of run-config — the same unmeasured premise, in a new place.
//
// AND NOT MEASURED IS ITS OWN ROW. A run where nothing looked must not render as a run that
// matched its configuration; that collapse is how $379 of spend went to a tier that never ran.
func conduct(board *record.Board) string {
	seats := record.SeatModels(board)
	if len(seats) == 0 {
		return ""
	}

	// requestedFor is keyed SERVED -> REQUESTED -> count, and the nesting is the point: the first
	// cut kept one flat requested map and asked it `t.requested[served]`, which is a lookup of a
	// served name in a map keyed by requested names. It answered 0 for every real substitution —
	// the exact case the section exists to report — and the table rendered a clean run.
	type tally struct {
		byServed     map[string]int
		requestedFor map[string]map[string]int
		unmeasured   int
		total        int
	}
	classes := map[string]*tally{}
	var order []string
	for _, s := range seats {
		class := s.Class
		if class == "" {
			// A seat riding no tier is still a seat that ran; naming it "—" keeps it visible
			// rather than dropping it out of the total and making the arithmetic unexplainable.
			class = "—"
		}
		t := classes[class]
		if t == nil {
			t = &tally{byServed: map[string]int{}, requestedFor: map[string]map[string]int{}}
			classes[class] = t
			order = append(order, class)
		}
		t.total++
		if !s.Measured() {
			t.unmeasured++
			continue
		}
		t.byServed[s.Served]++
		if s.Substituted() {
			if t.requestedFor[s.Served] == nil {
				t.requestedFor[s.Served] = map[string]int{}
			}
			t.requestedFor[s.Served][s.Requested]++
		}
	}
	sort.Strings(order)

	var b strings.Builder
	b.WriteString("## How this run was conducted\n\n")
	b.WriteString("_Composed from the record, never from the run's configuration: these are the models that ANSWERED, " +
		"measured at each seat's own first turn. A run can be configured for one tier and served by another, and a seat " +
		"cannot see which — it reads the request._\n\n")
	b.WriteString("| tier | what answered | seats |\n|---|---|---|\n")
	substituted := false
	for _, class := range order {
		t := classes[class]
		for _, served := range sortedKeys(t.byServed) {
			line := "`" + served + "`"
			if asked := t.requestedFor[served]; len(asked) > 0 {
				substituted = true
				line += " — **SUBSTITUTED**, " + substitutionNote(asked)
			}
			fmt.Fprintf(&b, "| %s | %s | %d of %d |\n", class, line, t.byServed[served], t.total)
		}
		if t.unmeasured > 0 {
			// NOT MEASURED, spelled out. The absent case and the healthy case are the same
			// bytes unless one of them says so.
			fmt.Fprintf(&b, "| %s | _NOT MEASURED — nothing observed what replied_ | %d of %d |\n", class, t.unmeasured, t.total)
		}
	}
	if substituted {
		b.WriteString("\n**A substitution changes the adversary's strength, which is a fact about this report's " +
			"evidentiary weight and not a footnote.** Where a tier was answered by a model other than the one " +
			"configured, any reasoning in this document about the models used — cross-vendor diversity, tier " +
			"pairing, judge independence — was written against the CONFIGURED pairing and does not describe " +
			"what ran.\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// substitutionNote names what was ASKED FOR, given the map of requested models for one served
// model. Plural because a re-dispatched seat can have asked for something else the second time,
// and collapsing that to one name would hide a run that changed tier mid-flight.
func substitutionNote(asked map[string]int) string {
	names := make([]string, 0, len(asked))
	for r := range asked {
		names = append(names, "`"+r+"`")
	}
	sort.Strings(names)
	return "configured as " + strings.Join(names, " / ")
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
