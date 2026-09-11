package fuzz

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE DISPATCH LOOP'S TERMINATION, DRIVEN ON EVERY SCHEDULE (plans/roundless.md §III.B.1-B.2).
//
// There are no rounds: the chair sits, its envelope relays the plan `dispatch next` recorded, the
// loop dispatches the parties the plan names, and the debate ends when the plan is empty. So a
// schedule here is a sequence of CHAIR MOVES — what the record says at each chair sitting — and
// the oracle checks that every sequence settles to exactly one of the four terminal verdicts, for
// the right reason, after exactly as many chair sittings as it took, with nobody sitting after.

// move is what the chair's plan says at one sitting.
type move struct {
	name string
	// engage: the evidence lens (head moved) and blue on G1 are ready — an exchange.
	engage bool
	// fresh: the lens is ready and blue is engaged on a NEW gap this sitting mints.
	fresh bool
	// bench: G1 is at impasse; the plan dockets it and the bench sits. rules: the bench closes it.
	bench, rules bool
	// pass: nobody is ready and PASS is permitted; the chair records PASS → VERIFIED.
	pass bool
	// ceiling: nobody is ready, every open material gap at its limit and carried → CEILING.
	ceiling bool
	// stall: nobody is ready and neither PASS nor CEILING holds → UNVERIFIED.
	stall bool
	// halt: the chair petitions and the bench halts the run.
	halt bool
}

func (m move) terminal() bool { return m.pass || m.ceiling || m.stall || m.halt }

// noProgressEpochs is the oracle's statement of debate.js's no-progress valve (NO_PROGRESS_EPOCHS):
// that many identical plans in a row, and the loop stops UNVERIFIED at the sitting that repeats.
// It is the expectation the oracle holds the script to, so a script that changes the constant or
// drops the valve fails here by name rather than passing on a schedule that happens not to repeat.
const noProgressEpochs = 3

var moves = []move{
	{name: "engage", engage: true},
	{name: "fresh", fresh: true},
	{name: "bench-carries", bench: true},
	{name: "bench-rules", bench: true, rules: true},
	{name: "pass", pass: true},
	{name: "ceiling", ceiling: true},
	{name: "stall", stall: true},
	{name: "halt", halt: true},
}

// board is the oracle's own model of what is open: minted ids less ruled ones.
type board struct {
	raised map[string]bool
	ruled  map[string]bool
}

func (b *board) open() int {
	n := 0
	for id := range b.raised {
		if !b.ruled[id] {
			n++
		}
	}
	return n
}

// sittingOf reads the chair's sitting ordinal off its label (`red-chair #3 · slug`).
func sittingOf(label string) int {
	i := strings.Index(label, " #")
	if i < 0 {
		return 0
	}
	rest := label[i+2:]
	if j := strings.IndexByte(rest, ' '); j >= 0 {
		rest = rest[:j]
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0
	}
	return n
}

func party(seat string, gaps ...string) map[string]any {
	g := []any{}
	for _, x := range gaps {
		g = append(g, x)
	}
	return map[string]any{"seat_id": seat, "gap_ids": g}
}

func plan(parties []any, pass, ceiling bool, docket []any, why string) map[string]any {
	if docket == nil {
		docket = []any{}
	}
	return map[string]any{"head": 2, "parties": parties, "docket": docket, "pass_permitted": pass, "ceiling": ceiling, "why": []any{why}}
}

func runSchedule(t *testing.T, script string, sched []move) ([]debatejs.Dispatch, debatejs.Outcome, *board) {
	t.Helper()
	b := &board{raised: map[string]bool{}, ruled: map[string]bool{}}
	moveFor := func(sitting int) move {
		if sitting >= 1 && sitting <= len(sched) {
			return sched[sitting-1]
		}
		// PAST THE SCHEDULE the record's terms have run out: every open gap is at its limit. The
		// stand-in is the ceiling, which is what a run of stalled exchanges ends in.
		return move{name: "ceiling", ceiling: true}
	}
	var lastMove move
	var engaged []any // the gap ids the last plan engaged blue on — blue manifests exactly those
	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "termination", "petitions": []any{}, "log": []any{}, "rulings": []any{},
			"resolutions": []any{}, "holdings": []any{}, "manifest": []any{}, "claim_count": 3,
			"saturation_reached": false, "sitting_record_appended": true, "unruled_motions": 0,
		}
		switch {
		case seatID == "red-chair":
			m := moveFor(sittingOf(label))
			lastMove = m
			switch {
			case m.pass:
				e["plan"] = plan([]any{}, true, false, nil, "pass permitted")
				e["verdict"] = "PASS"
			case m.ceiling:
				e["plan"] = plan([]any{}, false, true, nil, "every open material gap is at its limit")
			case m.stall:
				e["plan"] = plan([]any{}, false, false, nil, "the bench sat and ruled nothing")
			case m.halt:
				engaged = []any{"G1"}
				e["plan"] = plan([]any{party("blue-respond", "G1")}, false, false, nil, "engaged")
				e["petitions"] = []any{map[string]any{"class": "safety", "argument": "halt this run"}}
			case m.bench:
				b.raised["G1"] = true
				e["plan"] = plan([]any{party("judge", "G1")}, false, false, []any{"G1"}, "G1 at impasse")
			case m.fresh:
				id := fmt.Sprintf("F%d", sittingOf(label))
				b.raised[id] = true
				engaged = []any{id}
				e["plan"] = plan([]any{party("red-lens-evidence"), party("blue-respond", id)}, false, false, nil, "fresh")
			default: // engage
				b.raised["G1"] = true
				engaged = []any{"G1"}
				e["plan"] = plan([]any{party("red-lens-evidence", "G1"), party("blue-respond", "G1")}, false, false, nil, "engaged")
			}
		case strings.HasPrefix(seatID, "red-lens-"):
			return debatejs.Envelope{"synopsis": "lens"}
		case seatID == "blue-respond":
			// Blue manifests what the PLAN engaged it on — a ruled gap the chair re-engages is still
			// blue's to answer this sitting; whether the record would ever engage it is the record's
			// question, and this test's is only that the loop settles on every schedule.
			e["manifest"] = engaged
		case strings.HasPrefix(seatID, "judge-petition"):
			e["rulings"] = []any{map[string]any{"class": "safety", "ruling": "denied", "opinion": "no"}}
			if lastMove.halt {
				e["halt"] = map[string]any{"opinion": "the run is halted for the schedule under test"}
			}
		case seatID == "judge":
			res := []any{}
			if lastMove.rules {
				b.ruled["G1"] = true
				res = append(res, map[string]any{"gap_id": "G1", "resolution": "not_a_defect", "settled": "settled", "reopens_on": "new evidence", "rationale": "ruled"})
			} else {
				res = append(res, map[string]any{"gap_id": "G1", "resolution": "carried", "rationale": "owed"})
			}
			e["resolutions"] = res
		case seatID == "assemble":
			e["open_gaps"] = b.open()
		}
		return e
	}
	ds, out, err := debatejs.CaptureRun(script, debatejs.Config{
		Topic: "termination", RunDir: t.TempDir(), BinDir: t.TempDir(), Lanes: 1,
		Model: "haiku", JudgmentModel: "haiku", Backend: backend, Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("schedule %s: %v", scheduleName(sched), err)
	}
	return ds, out, b
}

func scheduleName(sched []move) string {
	parts := make([]string, len(sched))
	for i, m := range sched {
		parts[i] = m.name
	}
	return strings.Join(parts, " > ")
}

// Every schedule of three chair moves — the full cross product — settles to one of the four
// verdicts, for the reason its moves dictate, after exactly the chair sittings it took.
func TestTheDispatchLoopTerminatesConsistentlyOnEverySchedule(t *testing.T) {
	if testing.Short() {
		t.Skip("drives debate.js once per schedule")
	}
	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	const depth = 3
	var scheds [][]move
	var build func(prefix []move)
	build = func(prefix []move) {
		if len(prefix) == depth {
			scheds = append(scheds, append([]move{}, prefix...))
			return
		}
		for _, m := range moves {
			build(append(prefix, m))
		}
	}
	build(nil)
	if len(scheds) != len(moves)*len(moves)*len(moves) {
		t.Fatalf("the enumeration produced %d schedules over %d moves at depth %d — it is not exhaustive", len(scheds), len(moves), depth)
	}
	verdicts := map[string]int{}
	for _, sched := range scheds {
		ds, out, b := runSchedule(t, script, sched)
		name := scheduleName(sched)
		verdicts[out.Verdict]++

		// The terminating move is the first terminal one, or the ceiling stand-in past the schedule —
		// unless the no-progress valve stops the loop first: noProgressEpochs identical plans in a
		// row end it UNVERIFIED at the sitting that repeats. engage serves the same plan every
		// sitting, and so do both bench moves (the same bytes); fresh mints a new id each sitting.
		term, sittings := move{name: "ceiling", ceiling: true}, len(sched)+1
		prev, run := "", 0
		for i, m := range sched {
			if m.terminal() {
				term, sittings = m, i+1
				break
			}
			key := m.name
			switch {
			case m.bench:
				key = "bench"
			case m.fresh:
				key = fmt.Sprintf("fresh-%d", i)
			}
			if key == prev {
				run++
			} else {
				prev, run = key, 1
			}
			if run >= noProgressEpochs {
				term, sittings = move{name: "no-progress"}, i+1
				break
			}
		}
		want := map[bool]string{true: ""}[false]
		switch {
		case term.halt:
			want = "HALTED"
		case term.pass:
			want = "VERIFIED"
		case term.ceiling:
			want = "CEILING"
		default:
			want = "UNVERIFIED"
		}
		if out.Verdict != want {
			t.Errorf("%s: verdict %q, want %q (terminating move %s)", name, out.Verdict, want, term.name)
		}
		if out.Halted != term.halt {
			t.Errorf("%s: halted=%v, want %v", name, out.Halted, term.halt)
		}
		if out.Epochs != sittings {
			t.Errorf("%s: %d chair sitting(s), want %d — the loop ran on past the record's word, or stopped before it", name, out.Epochs, sittings)
		}
		if out.Verdict == "UNVERIFIED" && out.GapsOutstanding != b.open() {
			t.Errorf("%s: UNVERIFIED with %d gap(s) outstanding, the board holds %d open", name, out.GapsOutstanding, b.open())
		}
		// NOBODY SITS AFTER TERMINATION but the bookends: the last debate seat dispatched is the
		// terminating chair sitting (or the petition sitting that halted), then judge-terminal /
		// assemble only.
		lastChair := -1
		for i, d := range ds {
			if d.SeatID == "red-chair" {
				lastChair = i
			}
		}
		for _, d := range ds[lastChair+1:] {
			switch {
			case d.SeatID == "assemble", d.SeatID == "judge-terminal", strings.HasPrefix(d.SeatID, "judge-petition"):
			case term.halt && d.SeatID == "blue-respond":
				t.Errorf("%s: blue sat after the halt", name)
			default:
				t.Errorf("%s: %s sat after the terminating chair sitting", name, d.SeatID)
			}
		}
		// The parties the plan named sat, in role order, within each epoch.
		if sittings > 1 && sched[0].engage {
			order := []string{}
			for _, d := range ds {
				if d.SeatID == "red-lens-evidence" || d.SeatID == "blue-respond" {
					order = append(order, d.SeatID)
				}
				if len(order) == 2 {
					break
				}
			}
			if strings.Join(order, ",") != "red-lens-evidence,blue-respond" {
				t.Errorf("%s: the first exchange sat %v, want the lens before blue", name, order)
			}
		}
	}
	keys := make([]string, 0, len(verdicts))
	for k := range verdicts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range []string{"HALTED", "VERIFIED", "CEILING", "UNVERIFIED"} {
		if verdicts[k] == 0 {
			t.Errorf("no schedule reached %s — the enumeration does not exercise every terminal", k)
		}
	}
	t.Logf("verdicts over %d schedules: %v", len(scheds), verdicts)
}
