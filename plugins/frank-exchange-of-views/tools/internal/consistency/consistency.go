// Package consistency is the cross-projection oracle: one record, many readers, zero tolerated
// disagreements.
//
// # Why this exists as a package rather than as one test
//
// The record is a single event log with at least ten independent readers — the Board replay, the
// board/work/findings JSON, five markdown renders, the telemetry series, the lineage graph, report
// assembly, capture's audits. Every reader is its own traversal of the same facts, and the bug
// class that architecture invites is PROJECTION DISAGREEMENT: two readers deriving different
// answers from one record, each locally plausible. It is not hypothetical — the measured
// 2026-08-22 incident was the board reporting open:0 while assembly reported a refuted source
// still cited, and the merge that landed the store found five more instances of the same class.
//
// Unit tests cannot see this class, because each one exercises a single reader against a fixture
// built to that reader's expectations. The oracle inverts that: it derives the ground truth ONCE,
// from the raw event walk — deliberately not through record.BoardState, whose reduction is itself
// a reader under test — and then holds every projection to it, and to each other.
//
// # What a violation is
//
// A violation is a DISAGREEMENT, not a verdict. When the replay says a gap was closed by the bench
// and the last closing event on the record is red's, the oracle reports both sides; which reader
// is wrong is the finding's job to establish. The one thing a violation always means is that two
// carriers of one fact have diverged, which is the thing this codebase keeps paying for.
package consistency

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/graph"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// anchorToken matches the three anchor classes as minted; group 1 is the id.
var anchorToken = regexp.MustCompile(`<!--(?:fx|cite|proof):([a-z]-[0-9a-f]+)-->`)

// gtGap is the ground truth for one gap, derived from the raw event walk.
//
// LAST-WRITER SEMANTICS THROUGHOUT: the record is append-only and ordered, so "who closed this
// gap" means the last closing event, "its fate" means that event's word, and "its grades" mean the
// mint's grades overlaid by every regrade in order. Any reader that answers differently is
// answering a different question, and the divergence is what the oracle reports.
type gtGap struct {
	mintRound      int
	open           bool
	everClosed     bool
	lastCloser     string // "red" (a close event) or "bench" (a closing opinion)
	lastClass      string // the word carried by the LAST closing event; "" if it carried none
	lastCloseRound int
	sev, lik, imp  recordpb.Grade
	cx             recordpb.Grade
	supersedes     []string
	carriedCount   int // opinions whose disposition defers rather than closes
}

type groundTruth struct {
	order         []string
	gaps          map[string]*gtGap
	findingLabels map[string]bool
	findingIDs    map[string]bool
	anchorEvIDs   map[string]bool
	proofIDs      map[string]bool
	citeEvents    int
	verifyEvents  int
	avenues       map[string]string // avenue id -> last status word
	observations  int
}

// closesSet is derived from the schema's own `closes` facet via record.ClosureClasses, so the
// oracle and the readers cannot disagree about the vocabulary itself — only about the record.
func closesSet() map[string]bool {
	out := map[string]bool{}
	for _, v := range record.ClosureClasses {
		out[v.Name] = true
	}
	return out
}

func walk(events []*record.Event) *groundTruth {
	gt := &groundTruth{
		gaps:          map[string]*gtGap{},
		findingLabels: map[string]bool{},
		findingIDs:    map[string]bool{},
		anchorEvIDs:   map[string]bool{},
		proofIDs:      map[string]bool{},
		avenues:       map[string]string{},
	}
	closes := closesSet()
	for _, e := range events {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch m := body.(type) {
		case *recordpb.Mint:
			id := m.GetGapId()
			if _, seen := gt.gaps[id]; !seen {
				gt.order = append(gt.order, id)
			}
			gt.gaps[id] = &gtGap{
				mintRound: int(e.GetRound()), open: true,
				sev: m.GetSeverity(), lik: m.GetLikelihood(), imp: m.GetImpact(), cx: m.GetComplexityCost(),
				supersedes: append([]string{}, m.GetSupersedes()...),
			}
		case *recordpb.Regrade:
			g := gt.gaps[m.GetGapId()]
			if g == nil {
				continue // the replay errors on this; the replay-agreement rule will say so
			}
			if m.Severity != nil {
				g.sev = m.GetSeverity()
			}
			if m.Likelihood != nil {
				g.lik = m.GetLikelihood()
			}
			if m.Impact != nil {
				g.imp = m.GetImpact()
			}
			if m.ComplexityCost != nil {
				g.cx = m.GetComplexityCost()
			}
		case *recordpb.Close:
			g := gt.gaps[m.GetGapId()]
			if g == nil {
				continue
			}
			g.open, g.everClosed = false, true
			g.lastCloser = "red"
			g.lastClass = recordpb.Word(m.GetClosureClass())
			g.lastCloseRound = int(e.GetRound())
		case *recordpb.Opinion:
			g := gt.gaps[m.GetGapId()]
			if g == nil {
				continue
			}
			w := recordpb.Word(m.GetDisposition())
			if !closes[w] {
				g.carriedCount++
				continue
			}
			g.open, g.everClosed = false, true
			g.lastCloser = "bench"
			g.lastClass = w
			g.lastCloseRound = int(e.GetRound())
		case *recordpb.Finding:
			gt.observations++
			if l := m.GetLabel(); l != "" {
				gt.findingLabels[l] = true
			}
			if id := m.GetFindingId(); id != "" {
				gt.findingIDs[id] = true
			}
		case *recordpb.Anchor:
			if id := m.GetId(); id != "" {
				gt.anchorEvIDs[id] = true
			}
		case *recordpb.Proof:
			if id := m.GetProofId(); id != "" {
				gt.proofIDs[id] = true
			}
		case *recordpb.Cite:
			_ = m
			gt.citeEvents++
		case *recordpb.Verify:
			_ = m
			gt.verifyEvents++
		case *recordpb.Avenue:
			if id := m.GetAvenueId(); id != "" {
				gt.avenues[id] = recordpb.Word(m.GetStatus())
			}
		}
	}
	return gt
}

// Check derives the ground truth for a run and holds every projection to it.
func Check(runDir string) ([]string, error) {
	m, err := record.MergedEvents(runDir)
	if err != nil {
		return nil, fmt.Errorf("consistency: reading the record: %w", err)
	}
	gt := walk(m.Events)

	var v []string
	add := func(rule, format string, args ...any) {
		v = append(v, rule+": "+fmt.Sprintf(format, args...))
	}

	// ---- the replay itself ----
	board, err := record.BoardState(runDir)
	if err != nil {
		// The walk tolerates what the replay refuses (a mutation on an unknown gap), so a record
		// the replay cannot read at all is itself the finding.
		return append(v, fmt.Sprintf("replay-refused: BoardState errored where the raw walk did not: %v", err)), nil
	}
	for id, g := range gt.gaps {
		rg := board.Gaps[id]
		if rg == nil {
			add("replay-agreement", "gap %s exists in the raw walk and not on the board", id)
			continue
		}
		if rg.Open != g.open {
			add("replay-agreement", "gap %s: raw walk open=%v, board open=%v", id, g.open, rg.Open)
		}
		if rg.HasClosed != g.everClosed {
			add("replay-agreement", "gap %s: raw walk everClosed=%v, board HasClosed=%v", id, g.everClosed, rg.HasClosed)
		}
		if g.everClosed && rg.ClosedRound != g.lastCloseRound {
			add("replay-agreement", "gap %s: last closing event is round %d, board ClosedRound=%d", id, g.lastCloseRound, rg.ClosedRound)
		}
		if g.everClosed && rg.ClosedByBench != (g.lastCloser == "bench") {
			add("closer-attribution", "gap %s: last closing event was %s's (round %d), board ClosedByBench=%v",
				id, g.lastCloser, g.lastCloseRound, rg.ClosedByBench)
		}
		if rg.Severity != g.sev || rg.Likelihood != g.lik || rg.Impact != g.imp || rg.ComplexityCost != g.cx {
			add("replay-agreement", "gap %s: grades diverge — walk (%s,%s,%s,%s) board (%s,%s,%s,%s)", id,
				recordpb.Word(g.sev), recordpb.Word(g.lik), recordpb.Word(g.imp), recordpb.Word(g.cx),
				recordpb.Word(rg.Severity), recordpb.Word(rg.Likelihood), recordpb.Word(rg.Impact), recordpb.Word(rg.ComplexityCost))
		}
	}
	for id := range board.Gaps {
		if gt.gaps[id] == nil {
			add("replay-agreement", "gap %s exists on the board and not in the raw walk", id)
		}
	}

	// ---- the JSON projections ----
	bj := record.BoardJSONOf(board)
	openGT, closedGT := 0, 0
	for _, g := range gt.gaps {
		if g.open {
			openGT++
		} else {
			closedGT++
		}
	}
	if bj.Counts.Open != openGT {
		add("counts", "counts.open=%d, raw walk says %d", bj.Counts.Open, openGT)
	}
	if bj.Counts.Closed != closedGT {
		add("counts", "counts.closed=%d, raw walk says %d", bj.Counts.Closed, closedGT)
	}
	benchGT := 0
	for _, g := range gt.gaps {
		if !g.open && g.lastCloser == "bench" {
			benchGT++
		}
	}
	if bj.Counts.ClosedByBench != benchGT {
		add("closer-attribution", "counts.closed_by_bench=%d; %d gaps' LAST closing event is the bench's", bj.Counts.ClosedByBench, benchGT)
	}
	if bj.Counts.TotalObservations != gt.observations {
		add("counts", "counts.total_observations=%d, %d finding events on the record", bj.Counts.TotalObservations, gt.observations)
	}
	// Citations = verify events, CitationsAuthored = cite events — the #341 split. The field
	// doc used to promise a different number (distinct cite events); implementing the doc is how
	// the oracle found the disagreement, and the doc is what moved.
	if bj.Counts.Citations != gt.verifyEvents {
		add("counts", "counts.citations=%d, %d verify events on the record", bj.Counts.Citations, gt.verifyEvents)
	}
	if bj.Counts.CitationsAuthored != gt.citeEvents {
		add("counts", "counts.citations_authored=%d, %d cite events on the record", bj.Counts.CitationsAuthored, gt.citeEvents)
	}
	if bj.Counts.Anomalies != len(bj.Anomalies) {
		add("counts", "counts.anomalies=%d but the anomalies list carries %d", bj.Counts.Anomalies, len(bj.Anomalies))
	}

	wj := record.WorkJSONOf(board)
	if got, want := idsOfWork(wj), openIDs(gt); !sameSet(got, want) {
		add("work-mirror", "work.open=%v, raw walk open=%v", got, want)
	}
	for _, ci := range wj.ClosedIndex {
		g := gt.gaps[ci.ID]
		if g == nil || g.open {
			add("work-mirror", "closed index carries %s, which the raw walk holds open or unknown", ci.ID)
			continue
		}
		if ci.ClosedBy != g.lastCloser {
			add("closer-attribution", "closed index %s: closed_by=%q, last closing event was %s's", ci.ID, ci.ClosedBy, g.lastCloser)
		}
		if ci.Fate != g.lastClass {
			add("closer-attribution", "closed index %s: fate=%q, the LAST closing event carried %q", ci.ID, ci.Fate, g.lastClass)
		}
	}

	fj := record.FindingsJSONOf(board)
	gotLabels := map[string]bool{}
	for _, f := range fj.Findings {
		gotLabels[f.Label] = true
	}
	if !sameSet(keys(gotLabels), keys(gt.findingLabels)) {
		add("findings", "findings projection labels %v, record carries %v", keys(gotLabels), keys(gt.findingLabels))
	}

	// ---- the markdown renders ----
	if ledger, err := view.Markdown(runDir, "ledger", ""); err != nil {
		add("ledger-md", "render failed: %v", err)
	} else {
		s := string(ledger)
		if m := regexp.MustCompile(`## OPEN GAPS \((\d+)\)`).FindStringSubmatch(s); m == nil {
			add("ledger-md", "no OPEN GAPS heading in the ledger")
		} else if m[1] != fmt.Sprint(openGT) {
			add("ledger-md", "ledger says OPEN GAPS (%s), raw walk says %d", m[1], openGT)
		}
		for id, g := range gt.gaps {
			if g.open && !strings.Contains(s, id) {
				add("ledger-md", "open gap %s missing from the ledger", id)
			}
			if !g.open && !strings.Contains(s, id) {
				add("ledger-md", "closed gap %s missing from the ledger's closure index", id)
			}
		}
	}
	if archive, err := view.Markdown(runDir, "archive", ""); err != nil {
		add("archive-md", "render failed: %v", err)
	} else {
		for id, g := range gt.gaps {
			if !g.open && !strings.Contains(string(archive), id) {
				add("archive-md", "closed gap %s missing from the archive", id)
			}
		}
	}
	if len(gt.avenues) > 0 {
		if inq, err := view.Markdown(runDir, "lines-of-inquiry", ""); err != nil {
			add("inquiry-md", "render failed: %v", err)
		} else {
			for id := range gt.avenues {
				if !strings.Contains(string(inq), id) {
					add("inquiry-md", "avenue %s missing from lines-of-inquiry", id)
				}
			}
		}
	}

	// ---- telemetry ----
	if rows, err := view.Telemetry(runDir); err != nil {
		add("telemetry", "derivation failed: %v", err)
	} else if len(rows) > 0 {
		last := rows[len(rows)-1]
		if oc, ok := numField(last, "open_count"); ok && oc != openGT {
			add("telemetry", "final round open_count=%d, raw walk says %d", oc, openGT)
		}
	}

	// ---- the lineage graph ----
	mmd := graph.Mermaid(board)
	for id, g := range gt.gaps {
		if !strings.Contains(mmd, `["`+id) && !strings.Contains(mmd, id) {
			add("graph", "gap %s has no node in the mermaid graph", id)
		}
		for _, anc := range g.supersedes {
			if gt.gaps[anc] == nil {
				add("lineage", "gap %s supersedes %s, which no mint created", id, anc)
			}
		}
	}
	if cyc := findCycle(gt); cyc != "" {
		add("lineage", "supersedes cycle: %s", cyc)
	}

	// ---- the anchor layer: report tokens and record events are a bijection ----
	//
	// Splice and append are TWO ACTS with no transaction over them: `blue cite` mutates
	// blue/report.md first and records the cite event second, so a crash between the two leaves
	// an anchor token no event backs — and the crash-retry mints a FRESH label and splices a
	// second anchor beside the orphan. Neither direction of the mismatch is legal in a settled
	// record, and this rule is what makes the torn state visible at all.
	if rep, rerr := os.ReadFile(filepath.Join(runDir, "blue", "report.md")); rerr == nil {
		inReport := map[string]bool{}
		for _, m := range anchorToken.FindAllStringSubmatch(string(rep), -1) {
			inReport[m[1]] = true
		}
		// Each class's recorded set is the union of the events entitled to name that id: a c-
		// label may come from blue's cite or from red's labelled verify (a corroboration); an f-
		// id from the finding or its anchor event; a p- id from the proof. Every splice precedes
		// its append, so BOTH directions of each bijection are invariants of a settled record.
		citeSet := map[string]bool{}
		if labels, lerr := record.CitationLabels(runDir); lerr != nil {
			add("anchor-record", "deriving expected citation labels: %v", lerr)
		} else {
			for _, l := range labels {
				citeSet[l] = true
			}
		}
		backed := func(id string) (bool, string) {
			switch {
			case strings.HasPrefix(id, "c-"):
				return citeSet[id], "cite/verify"
			case strings.HasPrefix(id, "f-"):
				return gt.findingIDs[id] || gt.anchorEvIDs[id], "finding/anchor"
			case strings.HasPrefix(id, "p-"):
				return gt.proofIDs[id], "proof"
			}
			return true, "" // an alien class is claimcount's problem, not a torn splice
		}
		for tok := range inReport {
			if ok, kind := backed(tok); !ok {
				add("anchor-record", "report.md carries anchor %s and no %s event names it — a torn splice, and a naive retry would mint a duplicate beside it", tok, kind)
			}
		}
		for l := range citeSet {
			if !inReport[l] {
				add("anchor-record", "the record carries citation label %s and report.md has no anchor for it — the citation will never render", l)
			}
		}
		for id := range gt.proofIDs {
			if !inReport[id] {
				add("anchor-record", "the record carries proof id %s and report.md has no anchor for it — the proof backs no sentence", id)
			}
		}
		for id := range gt.findingIDs {
			if !inReport[id] {
				add("anchor-record", "the record carries finding id %s and report.md has no marker for it", id)
			}
		}
	}
	// The finding and its anchor event are appended as a PAIR after the splice; a finding with no
	// anchor event is the crash window between the two appends, sealed by an idempotent retry
	// that never looked. Report-independent, so it runs even when report.md is gone.
	for id := range gt.findingIDs {
		if !gt.anchorEvIDs[id] {
			add("anchor-record", "finding %s has no anchor event — the immortal-marker detector never learned its marker exists", id)
		}
	}

	// ---- round attribution ----
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
			continue
		}
		if want, known := record.RoundOf(e.GetSeatId()); known && int(e.GetRound()) != want {
			add("round-attribution", "%s event by %s stamped round %d; the seat id says round %d",
				recordpb.Word(e.GetType()), e.GetSeatId(), e.GetRound(), want)
		}
	}

	sort.Strings(v)
	return v, nil
}

func idsOfWork(wj record.WorkJSON) []string {
	var out []string
	for _, g := range wj.Open {
		out = append(out, g.ID)
	}
	return out
}

func openIDs(gt *groundTruth) []string {
	var out []string
	for id, g := range gt.gaps {
		if g.open {
			out = append(out, id)
		}
	}
	return out
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sameSet(a, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func numField(row map[string]any, key string) (int, bool) {
	switch n := row[key].(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	}
	return 0, false
}

// findCycle walks the supersedes edges; the write path makes a cycle unconstructible (an ancestor
// must already exist), so a hit here means the record was built by something other than the tool.
func findCycle(gt *groundTruth) string {
	const white, grey, black = 0, 1, 2
	color := map[string]int{}
	var path []string
	var visit func(string) string
	visit = func(id string) string {
		color[id] = grey
		path = append(path, id)
		g := gt.gaps[id]
		if g != nil {
			for _, anc := range g.supersedes {
				switch color[anc] {
				case grey:
					return strings.Join(append(path, anc), " -> ")
				case white:
					if c := visit(anc); c != "" {
						return c
					}
				}
			}
		}
		color[id] = black
		path = path[:len(path)-1]
		return ""
	}
	for id := range gt.gaps {
		if color[id] == white {
			if c := visit(id); c != "" {
				return c
			}
		}
	}
	return ""
}
