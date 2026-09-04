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

// THE ROUND-LOOP TERMINATION MODEL (#535 step 4), checked by EXHAUSTIVE EXECUTION rather than
// by a model of it.
//
// #535 calls this "the only place formality pays", and it is right about the shape: the state
// space is small, the property is liveness, and there is a real defect in its history. A run in
// 2026-08-22 stamped UNVERIFIED with ZERO gaps outstanding, because the bench ruled the last two
// gaps in one sitting and the engine read its own cleared board as deadlock. The bench wrote the
// finding into its own certify: red owns PASS/FAIL, and the bench's docket-clearing had
// substituted for red's affirmative call.
//
// It is NOT checked with a hand-written model, and that is this repository's own rule rather
// than a preference — "generated on both sides or not built". A TLA+ or Go state machine of this
// loop would be a second carrier of the loop's semantics, and the moment either moved it would
// be checking a debate nobody runs. What is enumerated here is the SEAT BEHAVIOUR — what red
// verdicts, what the bench rules, whether it halts — and the real debate.js decides the rest.
// Every terminal fact asserted below is a value the shipped script computed and returned.
//
// WHAT THIS CANNOT SEE, said plainly. It drives the loop with stubbed seats, so it checks the
// engine's arithmetic and nothing about whether a real bench would rule the way a move says.
// And the alphabet is what the moves below cover: a behaviour no move expresses is a behaviour
// no schedule reaches, which is why the moves are named for the states they exist to produce and
// the count of them is reported.

// gapShape is the minimum a red gap needs for debate.js to grade, mass and lineage it. The
// engine reads severity/likelihood/impact through its MASS table; a missing one is scored 0 and
// would quietly move the detector arm.
func gapShape(id string) map[string]any {
	return map[string]any{
		"id": id, "severity": "major", "likelihood": "medium", "impact": "medium",
		"complexity_cost": "low", "supersedes": []any{},
	}
}

// move is one round's seat behaviour: what red does, and what the bench does if it sits.
//
// The bench sits only when the docket is contested, which needs a gap id red RE-RAISES from a
// prior round — so no move can force a sitting in round 1, and the enumeration finds that rather
// than assuming it.
type move struct {
	name string
	// redPass ends the round at red's break with a PASS.
	redPass bool
	// freshGaps mints a new gap id each round, so nothing is ever contested and the bench never
	// sits — the schedule that reaches the ceiling.
	freshGaps bool
	// extraGap adds a second gap the bench does NOT rule, so a deadlock ruling leaves the board
	// open.
	extraGap bool
	// benchClears rules every gap red raised, emptying the board.
	benchClears bool
	// benchDeadlock is the bench's deadlock flag.
	benchDeadlock bool
	// halt files a petition the bench answers with a halt.
	halt bool
}

// moves is the alphabet. Each is named for the terminal state it exists to reach, and the set is
// closed: a schedule is a sequence of these and nothing else.
var moves = []move{
	{name: "pass", redPass: true},
	{name: "fail-carry"},
	{name: "fail-fresh", freshGaps: true},
	{name: "fail-bench-clears", benchClears: true},
	{name: "fail-deadlock-board-open", extraGap: true, benchDeadlock: true},
	{name: "fail-deadlock-board-cleared", benchClears: true, benchDeadlock: true},
	{name: "halt", halt: true},
}

// board is the run's independent oracle: what is open, tracked by the harness rather than read
// back out of the engine's own control flow. debate.js asks the assemble seat for the count, so
// the seat answers from here — which is what makes the UNVERIFIED-with-nothing-open assertion a
// check on the ENGINE rather than on the engine agreeing with itself.
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

// roundOf reads the round stamp out of a seat id (red-merge-r2 -> 2). A seat id with no stamp is
// round 0 — the frontier, the lanes, the synthesis, assemble — and those seats take no move.
func roundOf(seatID string) int {
	i := strings.LastIndex(seatID, "-r")
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(seatID[i+2:])
	if err != nil {
		return 0
	}
	return n
}

// schedule drives one debate: a move per round, the real loop, and what it decided.
func runSchedule(t *testing.T, script string, sched []move) (debatejs.Outcome, *board) {
	t.Helper()
	b := &board{raised: map[string]bool{}, ruled: map[string]bool{}}
	moveFor := func(round int) move {
		if round >= 1 && round <= len(sched) {
			return sched[round-1]
		}
		// Past the schedule the parties simply keep failing on a standing gap. maxRounds is set
		// to len(sched) so this is unreachable, and it returns a defined move rather than a zero
		// value that would silently mean "carry".
		return move{name: "fail-carry"}
	}

	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "termination", "verdict": "FAIL", "citations_checked": 0,
			"gaps": []any{}, "petitions": []any{}, "friction": []any{}, "rulings": []any{},
			"closures": []any{}, "dispute_responses": []any{}, "deadlock": false,
			"resolutions": []any{}, "grade_disputes": []any{}, "holdings": []any{},
			"manifest": []any{}, "claim_count": 3,
			"saturation_reached": false, "round_record_appended": true,
		}
		r := roundOf(seatID)
		m := moveFor(r)
		switch {
		case strings.HasPrefix(seatID, "red-merge"):
			if m.redPass {
				e["verdict"] = "PASS"
				return e
			}
			ids := []string{"G1"}
			if m.freshGaps {
				ids = []string{fmt.Sprintf("F%d", r)}
			}
			if m.extraGap {
				ids = append(ids, "G2")
			}
			gaps := []any{}
			for _, id := range ids {
				b.raised[id] = true
				gaps = append(gaps, gapShape(id))
			}
			e["gaps"] = gaps
			e["manifest"] = []any{}
			if m.halt {
				e["petitions"] = []any{map[string]any{"class": "safety", "argument": "halt this run"}}
			}
		case strings.HasPrefix(seatID, "blue-respond"):
			// Blue must manifest a receipt per open gap or the engine throws (W2b). The
			// manifest is the board's open set, which is what a blue that answered every open
			// gap would return.
			man := []any{}
			for id := range b.raised {
				if !b.ruled[id] {
					man = append(man, id)
				}
			}
			e["manifest"] = man
		case strings.HasPrefix(seatID, "judge-petition"):
			e["rulings"] = []any{map[string]any{"class": "safety", "ruling": "denied", "opinion": "no"}}
			if m.halt {
				e["halt"] = map[string]any{"opinion": "the run is halted for the schedule under test"}
			}
		case strings.HasPrefix(seatID, "judge"):
			e["deadlock"] = m.benchDeadlock
			res := []any{}
			if m.benchClears {
				for id := range b.raised {
					if b.ruled[id] || id == "G2" {
						continue
					}
					b.ruled[id] = true
					res = append(res, map[string]any{
						"gap_id": id, "resolution": "not_a_defect",
						"settled": "settled", "reopens_on": "new evidence", "rationale": "ruled",
					})
				}
			}
			e["resolutions"] = res
		case seatID == "assemble":
			e["open_gaps"] = b.open()
		}
		return e
	}

	_, out, err := debatejs.CaptureRun(script, debatejs.Config{
		Topic: "termination", RunDir: t.TempDir(), BinDir: t.TempDir(), Lanes: 1,
		MaxRounds: len(sched), Model: "haiku", JudgmentModel: "haiku",
		Backend: backend, Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("schedule %s: %v", scheduleName(sched), err)
	}
	return out, b
}

func scheduleName(sched []move) string {
	parts := make([]string, len(sched))
	for i, m := range sched {
		parts[i] = m.name
	}
	return strings.Join(parts, " -> ")
}

// TestTheRoundLoopTerminatesConsistentlyOnEverySchedule enumerates every sequence of moves up to
// the round ceiling and asserts the terminal properties on what the shipped loop returned.
func TestTheRoundLoopTerminatesConsistentlyOnEverySchedule(t *testing.T) {
	releaseGate(t)
	if testing.Short() {
		t.Skip("drives debate.js once per schedule")
	}
	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatalf("locating the shipped debate.js: %v", err)
	}

	const rounds = 3
	var scheds [][]move
	var build func(prefix []move)
	build = func(prefix []move) {
		if len(prefix) == rounds {
			s := make([]move, rounds)
			copy(s, prefix)
			scheds = append(scheds, s)
			return
		}
		for _, m := range moves {
			build(append(prefix, m))
		}
	}
	build(nil)

	// THE FLOOR. Every assertion below is vacuously true over an empty schedule list, which is
	// the shape that let #654 survive: the comparison passed because there was nothing to
	// compare.
	if len(scheds) != len(moves)*len(moves)*len(moves) {
		t.Fatalf("the enumeration produced %d schedules over %d moves and %d rounds — it is not exhaustive", len(scheds), len(moves), rounds)
	}

	verdicts := map[string]int{}
	relief, ceilingOwed, emptyBoardUnverified := 0, 0, 0
	for _, sched := range scheds {
		out, b := runSchedule(t, script, sched)
		name := scheduleName(sched)
		verdicts[out.Verdict]++
		if out.BenchClearedBoard == "relief_granted" {
			relief++
		}
		if out.BenchClearedBoard == "ceiling_owed_red_a_sitting" {
			ceilingOwed++
		}
		if out.Verdict == "UNVERIFIED" && out.GapsOutstanding == 0 {
			emptyBoardUnverified++
		}

		// TOTALITY. A verdict outside the four the script composes is a stamp nothing
		// downstream classifies.
		switch out.Verdict {
		case "HALTED", "VERIFIED", "CEILING", "UNVERIFIED":
		default:
			t.Errorf("%s: verdict %q is not one of HALTED/VERIFIED/CEILING/UNVERIFIED", name, out.Verdict)
		}

		// RED OWNS PASS/FAIL. VERIFIED is reachable only through a round whose move was red's
		// PASS; nothing the bench does may produce it.
		if out.Verdict == "VERIFIED" {
			passed := false
			for i := 0; i < out.Rounds && i < len(sched); i++ {
				if sched[i].redPass {
					passed = true
				}
			}
			if !passed {
				t.Errorf("%s: VERIFIED after %d round(s) with no red PASS in the schedule — something other than red certified the run", name, out.Rounds)
			}
		}

		// THE HISTORIC DEFECT (2026-08-22). UNVERIFIED says the parties did not converge;
		// an empty board says they did. The two together are the self-contradictory stamp
		// the bench-cleared relief exists to prevent.
		//
		// ONE STATE IS EXEMPT AND IT IS NOT AN EXCUSE, it is the other half of the same
		// design: where the relief WAS granted, red was given its sitting against the cleared
		// docket and refused PASS anyway. That is irreducible disagreement — the engine's own
		// comment says so — and the empty board is then a fact about what red refused rather
		// than a contradiction. The exemption is keyed on the engine's own reported field, not
		// on the harness re-deriving when the relief would have fired: a second computation of
		// that fact is the drift this whole check refuses to introduce.
		if out.Verdict == "UNVERIFIED" && out.GapsOutstanding == 0 && out.BenchClearedBoard != "relief_granted" {
			t.Errorf("%s: UNVERIFIED with ZERO gaps outstanding after %d round(s) (deadlocked=%v) and NO bench-cleared relief granted — red never had a sitting against the empty board, which is the 2026-08-22 defect", name, out.Rounds, out.Deadlocked)
		}

		// A HALT DOMINATES. If the bench halted, no other terminator may claim the run.
		if out.Halted && out.Verdict != "HALTED" {
			t.Errorf("%s: the bench halted but the run stamped %s", name, out.Verdict)
		}
		if out.Verdict == "HALTED" && !out.Halted {
			t.Errorf("%s: stamped HALTED with halted=false", name)
		}

		// THE CEILING IS THE CEILING. CEILING may only be claimed by a run that actually
		// reached maxRounds without a PASS, a deadlock or a halt.
		if out.Verdict == "CEILING" && (out.Rounds < rounds || out.Deadlocked || out.Halted) {
			t.Errorf("%s: CEILING at round %d of %d (deadlocked=%v halted=%v)", name, out.Rounds, rounds, out.Deadlocked, out.Halted)
		}

		// TERMINATION. The loop is bounded by maxRounds and nothing may push past it — the
		// one-shot relief included, which is the whole reason it is one-shot.
		if out.Rounds > rounds {
			t.Errorf("%s: ran %d rounds against a ceiling of %d", name, out.Rounds, rounds)
		}
		if out.Rounds < 1 {
			t.Errorf("%s: settled after %d rounds — the debate never sat", name, out.Rounds)
		}

		// THE ENGINE AND THE BOARD AGREE. gaps_outstanding is the assemble seat's answer, and
		// the seat answers from the harness board — so a disagreement here is the engine
		// reporting a count nobody's record supports.
		if out.Verdict != "HALTED" && out.GapsOutstanding != b.open() {
			t.Errorf("%s: the run reports %d gap(s) outstanding, the board holds %d", name, out.GapsOutstanding, b.open())
		}
	}

	// The distribution is REPORTED, not asserted: it is the evidence that the alphabet reaches
	// more than one terminal state. An enumeration that only ever produced UNVERIFIED would pass
	// every assertion above and check nothing.
	var kinds []string
	for v, n := range verdicts {
		kinds = append(kinds, fmt.Sprintf("%s=%d", v, n))
	}
	sort.Strings(kinds)
	t.Logf("termination model (#535 step 4): %d schedules over %d moves x %d rounds · %s", len(scheds), len(moves), rounds, strings.Join(kinds, " "))
	if len(verdicts) < 2 {
		t.Errorf("every schedule ended in the same terminal state (%v) — the move alphabet does not exercise the loop's exits, and the assertions above are checking one path", kinds)
	}

	// THE EXEMPTION MUST NOT BE VACUOUS. The empty-board assertion above excuses runs where the
	// engine reports the bench-cleared relief was granted. If that field ever stopped being read
	// — renamed on the return object, moved, dropped — it would read false everywhere, the
	// exemption would excuse nothing, and the test would still be green as long as no schedule
	// reached the state at all. Both halves are therefore counted: the relief must fire
	// somewhere, and the state the exemption exists for must be reached somewhere.
	t.Logf("cleared board: relief_granted in %d schedule(s), ceiling_owed_red_a_sitting in %d · %d ended UNVERIFIED against an empty board", relief, ceilingOwed, emptyBoardUnverified)
	if relief == 0 {
		t.Error("NO schedule was granted the bench-cleared relief — either the alphabet cannot reach it or bench_cleared_board is no longer read off the run, and the empty-board exemption is excusing nothing")
	}
	if ceilingOwed == 0 {
		t.Error("NO schedule reached ceiling_owed_red_a_sitting — the arm added for the ceiling-blocked relief is unreached, so the fix it carries is untested")
	}
	if emptyBoardUnverified == 0 {
		t.Error("NO schedule ended UNVERIFIED against an empty board — the state the exemption was written for is unreachable, so the exemption is untested and the assertion it guards has nothing to distinguish")
	}
}
