// Package graph renders a run's ACTUAL behaviour — replayed from its event record — as a
// graph, so the shape of what happened (and the holes in it) can be seen rather than grepped.
// It is the per-run counterpart to the static design map: the seats that really ran, and each
// gap's real lifecycle. Read-only; it emits mermaid (renders natively in markdown/artifacts/
// GitHub) or DOT (for graphviz-grade layout on a dense run).
package graph

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

var nonID = regexp.MustCompile(`[^a-zA-Z0-9]`)

func nodeID(prefix, s string) string { return prefix + "_" + nonID.ReplaceAllString(s, "_") }

// perGap holds the dialectic tally the graph annotates each gap with — the numbers that make a
// hole visible (a challenge with no ruling, a gap closed with no argument on the record).
type perGap struct {
	closings, motionsFiled, motionsRuled, dispositions int
}

// tallyByGap takes the BOARD, not raw events, because a ruling cannot be attributed to a gap
// from an event alone: `motion-rule` carries motion_id, never gap_id — the ask holds the gap and
// the answer holds only the ask's id. record.Motions performs that join, so counting rulings off
// the raw stream would bucket every one of them under the empty key and report zero rulings
// forever, which is the shape of the defect this counter was ported to escape.
func tallyByGap(b *record.Board) map[string]*perGap {
	m := map[string]*perGap{}
	get := func(id string) *perGap {
		if id == "" {
			return &perGap{}
		}
		if m[id] == nil {
			m[id] = &perGap{}
		}
		return m[id]
	}
	for _, e := range b.Events {
		// THE BODY IS THE TYPE. A closing carries its OWN gap_id field, so the id is read off
		// the message that holds it rather than off a key that might sit on anything — and an
		// event with no body has no gap to attribute, which is not the same fact as an event
		// whose gap_id is empty.
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		if f, ok := body.(*recordpb.Closing); ok {
			get(f.GetGapId()).closings++
		}
	}
	// A GRADE MOTION IS THIS GAP'S CHALLENGE, AND A DOCKET RULING IS ITS DISPOSITION. Both come
	// off the same join for the same reason: the gap rides the FILING, and neither the ruling
	// event nor the raw stream can attribute one. These counters read `dispute`/`dispute-respond`
	// until the motion collapse retired both types, after which the unanswered-challenge detector
	// below could only ever see zero — and reported "no unanswered challenges" and "no challenges"
	// as the same picture.
	//
	// TWO AXES, AND THEY ARE NOT THE SAME QUESTION. `motionsFiled`/`motionsRuled` ask WAS THIS
	// CHALLENGE ANSWERED — the unanswered-challenge detector below reads them. `dispositions`
	// asks DID THE BENCH SETTLE THE GAP, which only a docket ruling does: counting a grade or
	// petition ruling there would credit a gap with a disposition nobody wrote.
	//
	// A DOCKET MOTION COUNTS ON BOTH, and it used to count on neither until ruled. That made the
	// one shape this detector exists for invisible: a gap escalated to the bench and never
	// answered scored motionsFiled=0, so `motionsFiled > 0 && motionsRuled == 0` could not fire,
	// and the node rendered as an ordinary open gap. seat-command-triggers.md sells the docket
	// motion as what makes "a gap could reach the bench and get no ruling with nothing able to
	// notice" noticeable — and the detector for exactly that shape could not see it.
	//
	// ONLY THE SUBJECTS THAT NAME A GAP. A petition and an inquiry carry no gap id, and `get("")`
	// would mint a phantom node keyed on the empty string.
	for _, m := range record.Motions(b) {
		switch m.Subject {
		case "grade", "docket":
			if m.GapID == "" {
				continue
			}
			get(m.GapID).motionsFiled++
			if m.Ruled() {
				get(m.GapID).motionsRuled++
				if m.Subject == "docket" {
					get(m.GapID).dispositions++
				}
			}
		}
	}
	return m
}

// Mermaid returns a markdown document with two fenced mermaid diagrams: the run's seat-by-round
// flow, and the gap lifecycle. Markdown so it renders as-is in an artifact, a PR, or a phone.
func Mermaid(b *record.Board) string {
	var out strings.Builder
	out.WriteString("# Run graph — actual behaviour from the record\n\n")
	out.WriteString("## Seat flow by round\n\n```mermaid\n")
	out.WriteString(seatFlowMermaid(b))
	out.WriteString("```\n\n## Gap lifecycle\n\n```mermaid\n")
	out.WriteString(gapFlowMermaid(b))
	out.WriteString("```\n")
	return out.String()
}

// seatFlowMermaid groups seats into round subgraphs, each seat labelled with its event tally —
// so a round where a seat emitted nothing it should have (an empty debate, a skipped ruling)
// shows as a thin node.
func seatFlowMermaid(b *record.Board) string {
	type seat struct {
		id     string
		tally  map[string]int
		nEvent int
	}
	byRound := map[int]map[string]*seat{}
	var rounds []int
	for _, e := range b.Events {
		round, seatID := int(e.GetRound()), e.GetSeatId()
		if byRound[round] == nil {
			byRound[round] = map[string]*seat{}
			rounds = append(rounds, round)
		}
		s := byRound[round][seatID]
		if s == nil {
			s = &seat{id: seatID, tally: map[string]int{}}
			byRound[round][seatID] = s
		}
		// The tally is keyed on the event type's SCHEMA SPELLING (`motion_rule`, not
		// `motion-rule`): the type is an enum value now, and recordpb.Word is the one place
		// that word is derived. The zero renders as "", which is what an event with no type
		// produced before.
		s.tally[recordpb.Word(e.GetType())]++
		s.nEvent++
	}
	sort.Ints(rounds)

	var b2 strings.Builder
	b2.WriteString("flowchart TD\n")
	for _, r := range rounds {
		b2.WriteString(fmt.Sprintf("  subgraph rnd%d[\"round %d\"]\n    direction TB\n", r, r))
		var ids []string
		for id := range byRound[r] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			s := byRound[r][id]
			b2.WriteString(fmt.Sprintf("    %s[\"%s<br/>%s\"]\n", nodeID("s", fmt.Sprintf("r%d_%s", r, id)), id, tallyLabel(s.tally)))
		}
		b2.WriteString("  end\n")
	}
	for i := 1; i < len(rounds); i++ {
		b2.WriteString(fmt.Sprintf("  rnd%d --> rnd%d\n", rounds[i-1], rounds[i]))
	}
	return b2.String()
}

func tallyLabel(t map[string]int) string {
	var keys []string
	for k := range t {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if t[keys[i]] != t[keys[j]] {
			return t[keys[i]] > t[keys[j]]
		}
		return keys[i] < keys[j]
	})
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s×%d", k, t[k]))
	}
	return strings.Join(parts, " ")
}

// gapFlowMermaid renders each gap as a node coloured by terminal state, annotated with its
// dialectic tally, plus supersedes edges. A gap CLOSED with no closing and no disposition, or a
// gap with a dispute but zero dispute-responds, is a hole the annotation makes visible.
func gapFlowMermaid(b *record.Board) string {
	tally := tallyByGap(b)
	var out strings.Builder
	out.WriteString("flowchart LR\n")
	out.WriteString("  classDef open fill:#f7e0de,stroke:#c0453f,color:#161c26\n")
	out.WriteString("  classDef closed fill:#dcefe4,stroke:#2f855a,color:#161c26\n")
	out.WriteString("  classDef hole fill:#f6ead0,stroke:#b7791f,color:#161c26,stroke-width:2px\n")
	if len(b.GapOrder) == 0 {
		out.WriteString("  none[\"(no gaps minted)\"]\n")
		return out.String()
	}
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil {
			continue
		}
		pg := tally[id]
		if pg == nil {
			pg = &perGap{}
		}
		state, reason, class := "OPEN", "", "open"
		if !g.Open {
			state = "CLOSED"
			class = "closed"
			// ONE QUESTION, ONE ANSWER. This was `closure_class` else `disposition` — the
			// merge's word or the bench's — and the two now live on different messages
			// (`Close` and a docket motion's `DocketRuling`), so the fallback belongs to the
			// Gap rather than being
			// spelled out again here. `g.Closure != nil` would no longer serve: it means
			// "closed by a `close` event", so a bench closure would read as a torn one.
			reason = g.ClosureReason()
		}
		// A hole is a genuine defect, not merely a quiet gap: a dispute nobody answered, or a
		// TORN closure — closed with no recorded reason (no closure_class, no disposition). A
		// merge close that carries its closure_class is NOT a hole, even with no disposition.
		hole := (pg.motionsFiled > 0 && pg.motionsRuled == 0) || (!g.Open && reason == "")
		if hole {
			class = "hole"
		}
		label := fmt.Sprintf("%s · %s<br/>%s%s<br/>closing×%d motion×%d/%d disposition×%d",
			id, g.Mint.GetClass(), state, sep(reason), pg.closings, pg.motionsFiled, pg.motionsRuled, pg.dispositions)
		out.WriteString(fmt.Sprintf("  %s[\"%s\"]:::%s\n", nodeID("g", id), label, class))
	}
	// supersedes edges — a gap's lineage.
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Mint == nil {
			continue
		}
		for _, anc := range g.Mint.GetSupersedes() {
			out.WriteString(fmt.Sprintf("  %s -.supersedes.-> %s\n", nodeID("g", id), nodeID("g", anc)))
		}
	}
	return out.String()
}

func sep(reason string) string {
	if reason == "" {
		return ""
	}
	return ": " + reason
}

// Dot renders the gap lifecycle as a graphviz digraph — the escape hatch for a run dense enough
// that mermaid's layout struggles. Same gaps, states, and lineage; graphviz does the layout.
func Dot(b *record.Board) string {
	tally := tallyByGap(b)
	var out strings.Builder
	out.WriteString("digraph run {\n  rankdir=LR;\n  node [shape=box, style=\"rounded,filled\", fontname=\"monospace\"];\n")
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil {
			continue
		}
		pg := tally[id]
		if pg == nil {
			pg = &perGap{}
		}
		fill := "#f7e0de"
		state := "OPEN"
		reason := ""
		if !g.Open {
			fill, state = "#dcefe4", "CLOSED"
			// The merge's closure_class or the bench's disposition, asked once — see
			// gapFlowMermaid on why the `g.Closure != nil` guard cannot stand in for it.
			reason = g.ClosureReason()
		}
		if (pg.motionsFiled > 0 && pg.motionsRuled == 0) || (!g.Open && reason == "") {
			fill = "#f6ead0"
		}
		out.WriteString(fmt.Sprintf("  %q [label=%q, fillcolor=%q];\n", id, fmt.Sprintf("%s\\n%s\\nclose%d mot%d/%d disp%d", id, state, pg.closings, pg.motionsFiled, pg.motionsRuled, pg.dispositions), fill))
	}
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Mint == nil {
			continue
		}
		for _, anc := range g.Mint.GetSupersedes() {
			out.WriteString(fmt.Sprintf("  %q -> %q [style=dashed, label=\"supersedes\"];\n", id, anc))
		}
	}
	out.WriteString("}\n")
	return out.String()
}
