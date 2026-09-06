package cli

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE STATE GRAPH (#673): which event may legally follow which, for a gap, a motion and a line of
// inquiry — probed by EXECUTION.
//
// #535 split this out because the SURFACE graph — role → verb → event, gated since #669 — can tell
// you a verb is unreachable but not that an ENTITY can get stuck. "Dead end" means nothing without
// it: an entity state with no outgoing transition that is not terminal is a defect, and nothing
// asked the question.
//
// DERIVED, NOT MODELLED, which is #535's own rule. There is no transition table here to drift from
// the code. Each state is BUILT by running the real verbs against a real record, each act is then
// attempted for real, and the resulting state is read back off the record's own projections
// (record.Gaps, record.Motions, record.Inquiries). What the matrix reports is what the shipped
// write path did.
//
// SEAT SCOPE IS FACTORED OUT ON PURPOSE. `close` is the merge's verb and the bench is refused it —
// a fact about ROLE, which the surface graph already owns, and not about the entity's state. So
// each act is attempted under every registered seat and counts as possible if ANY seat is allowed
// it. The question here is only "can this entity get from here to there at all".

// probeAct is one attempted transition. `args` carries no --seat-id and no --run: the prober
// supplies both, once per seat it tries.
type probeAct struct {
	name string
	args []string
}

// entityProbe is one entity's lifecycle: how to build each state, how to read it back, and what
// may be attempted from it.
//
// GENERIC BECAUSE THREE COPIES IS THREE PLACES TO FIX. The gap, the motion and the line of inquiry
// have different verbs and different projections and the same question, and the first draft of this
// file answered it for the motion alone. A second and third copy would have drifted the moment the
// outcome vocabulary or the overwrite rule moved.
type entityProbe struct {
	// name labels the matrix rows; id is the entity the probe follows.
	name, id string
	// states are built in the order given, each in its own fresh run.
	states []string
	// buildTo returns a run directory holding this entity in the named state.
	buildTo func(t *testing.T, state string) string
	// read is the entity's current state, read off the record.
	read func(t *testing.T, runDir string) string
	// fields is the entity's content, keyed, so an act that rewrites it is visible field by field.
	fields func(t *testing.T, runDir string) map[string]string
	acts   []probeAct
	// accumulates names the fields that APPEND rather than replace — a history line, a regrade
	// list. If one of them grows, the earlier value is still on the projection and a reader can
	// see both, so a rewrite alongside it is a RECORD rather than a loss.
	//
	// THIS IS THE DIFFERENCE BETWEEN THE FINDING AND FOUR FALSE ONES. The first draft of this gate
	// flagged every rewrite of an already-set field and reported five: `regrade` changing a
	// severity, which is the verb's entire purpose, and a line of inquiry re-recording its own
	// fate, which debate.js calls out as legitimate in terms — "Re-recording `pursued` WITH what
	// it learned is a legitimate reaffirmation and settles the line for that round — do not read it
	// as neglect". Both keep the prior value: Gap.Regrades is a list and Inquiry.History is a line
	// per move. The double appeal keeps nothing, which is why it is the one that survived.
	accumulates []string
	// terminal names the states from which nothing more is expected, with the reason each is a
	// finish rather than a stall. THIS IS THE ONE DECLARED THING IN THE WHOLE FILE, and it has to
	// be: whether a state SHOULD have an exit is design intent, and no probe reads intent off code.
	terminal map[string]string
	// refusedForOtherReasons are acts whose refusal is NOT a fact about state — a precondition the
	// probe cannot satisfy in a single-round run, say. Listed with the reason so a refusal that
	// looks like a wall is known to be scaffolding.
	refusedForOtherReasons map[string]string
}

// probeSeats are the seats each act is attempted under. Every role that can touch an entity is
// here, because an act refused by one seat and allowed by another is allowed.
var probeSeats = []string{"red-lens-r1-L1", "red-merge-r1", "blue-respond-r1", "judge-r1"}

// probeReason is what an ATTEMPTED act says, as distinct from what the SETUP said.
//
// It is not decoration. The setup builds a state by running the same acts, so an act attempted with
// the setup's own wording writes back the bytes already there — and an overwrite becomes invisible,
// because the fields before and after are equal. That is this gate's own finding reporting itself
// as absent: measured, the second appeal on an already-appealed motion read as `inert` until the
// wording differed.
const probeReason = "PROBE: this wording exists only to make an overwrite visible"

// withProbeReason replaces every prose flag so a rewrite of the SAME field shows as a change.
func withProbeReason(args []string) []string {
	out := append([]string{}, args...)
	for i := 0; i < len(out)-1; i++ {
		switch out[i] {
		case "--reason", "--hypothesis", "--settled":
			out[i+1] = probeReason
		}
	}
	return out
}

// attempt runs one act under every probe seat and reports whether ANY was allowed.
func attempt(t *testing.T, runDir string, act probeAct) bool {
	t.Helper()
	for _, seat := range probeSeats {
		args := append(withProbeReason(act.args), "--seat-id", seat, "--run", runDir)
		if _, err := run(t, args...); err == nil {
			return true
		}
	}
	return false
}

// build runs an act as SETUP, keeping its own wording so probeReason stays distinctive.
func build(t *testing.T, runDir string, act probeAct) bool {
	t.Helper()
	for _, seat := range probeSeats {
		args := append(append([]string{}, act.args...), "--seat-id", seat, "--run", runDir)
		if _, err := run(t, args...); err == nil {
			return true
		}
	}
	return false
}

// stateEdge is one attempted transition and what came of it. Every attempt lands in exactly one
// outcome and the matrix prints them all: a gate showing only the ACCEPTED transitions could not
// tell "refused" from "never attempted", which are the two readings a silent no-match sits between.
type stateEdge struct{ entity, from, act, outcome string }

// probeEntity walks one entity's states, attempts every act from each, and returns the matrix plus
// what it found.
func probeEntity(t *testing.T, e entityProbe) (edges []stateEdge, deadEnds, overwrites []string, attempts int) {
	t.Helper()
	for _, from := range e.states {
		out := 0
		for _, act := range e.acts {
			attempts++
			// A FRESH BOARD PER ATTEMPT. An act that succeeds moves the entity, and the next act
			// would then be probed from somewhere else entirely.
			dir := e.buildTo(t, from)
			if got := e.read(t, dir); got != from {
				t.Fatalf("%s: built for state %q and the record reads %q — the probe would report transitions out of a state it is not in", e.name, from, got)
			}
			before := e.fields(t, dir)
			if !attempt(t, dir, act) {
				outcome := "refused"
				if why, stated := e.refusedForOtherReasons[act.name]; stated {
					outcome = "refused (not a state fact)"
					_ = why
				}
				edges = append(edges, stateEdge{e.name, from, act.name, outcome})
				continue
			}
			to := e.read(t, dir)
			after := e.fields(t, dir)
			if to != from {
				out++
				edges = append(edges, stateEdge{e.name, from, act.name, "-> " + to})
				continue
			}
			// AN OVERWRITE IS A REWRITE OF A FIELD THAT WAS ALREADY SET, WITH NOTHING KEEPING THE
			// OLD VALUE. Three distinctions, each load-bearing:
			//
			//   - a FIRST write into an empty field is an ordinary act that happens not to move
			//     the state — red ruling a proposed line of inquiry, say;
			//   - an append ALONGSIDE the rewrite means the prior value is still readable, so the
			//     act records rather than replaces;
			//   - only a replacement nothing preserves is the defect, because that is the one
			//     where a position stops being the answer with nothing saying so.
			grew := false
			for _, k := range e.accumulates {
				if before[k] != after[k] {
					grew = true
				}
			}
			var rewritten []string
			firstWrite := false
			for k, was := range before {
				now, ok := after[k]
				if !ok || now == was {
					continue
				}
				if was == "" {
					// A FIRST WRITE IS NOT NOTHING, and calling it `inert` was this matrix
					// naming the wrong thing with confidence — the failure mode #666 was
					// corrected for. `motion inquiry rule` on a proposed line sets `ruling` and
					// leaves the STATUS alone, which is the whole two-axis design: blue moves
					// the line, red rules it. Reported as inert, that read as a verb the record
					// ignores.
					firstWrite = true
					continue
				}
				rewritten = append(rewritten, fmt.Sprintf("%s: %q -> %q", k, was, now))
			}
			switch {
			case len(rewritten) > 0 && grew:
				edges = append(edges, stateEdge{e.name, from, act.name, "appends"})
				continue
			case len(rewritten) > 0:
				// falls through to the overwrite report below
			case firstWrite || grew:
				edges = append(edges, stateEdge{e.name, from, act.name, "records"})
				continue
			default:
				edges = append(edges, stateEdge{e.name, from, act.name, "inert"})
				continue
			}
			sort.Strings(rewritten)
			edges = append(edges, stateEdge{e.name, from, act.name, "OVERWRITES"})
			overwrites = append(overwrites, fmt.Sprintf("%s/%s --%s--> %s", e.name, from, act.name, strings.Join(rewritten, "; ")))
		}
		key := e.name + "/" + from
		if out == 0 {
			if _, declared := e.terminal[from]; !declared {
				deadEnds = append(deadEnds, key)
			}
		} else if reason, declared := e.terminal[from]; declared {
			t.Errorf("%s is declared TERMINAL but %d act(s) move the entity out of it — the declaration is stale and is now excusing a state that has exits.\ndeclared reason: %s", key, out, reason)
		}
	}
	return edges, deadEnds, overwrites, attempts
}

// ---- the gap ----

func gapMint() probeAct {
	return probeAct{"mint", []string{"mint", "--key", "k1", "--class", "self-attestation",
		"--problem", "p", "--fix", "f", "--check", "c", "--check-kind", "document",
		"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "low",
		"--reason", "the board needs something to argue about"}}
}

func gapProbe(t *testing.T) entityProbe {
	close := probeAct{"close", []string{"close", "--id", "R1-1", "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x",
		"--reason", "closed on the merits"}}
	// THE BENCH CLOSES TOO, and by a different verb writing a different message: `close` writes a
	// Close with a closure_class, `opinion` writes an Opinion with a disposition, and Gap carries
	// both because "closed" arrives as one of two shapes. A gap probe that drove only the merge's
	// verb would report a lifecycle with one exit where the record has two.
	opinionArgs := func(as string) []string {
		return []string{"opinion", "--id", "R1-1", "--as", as,
			"--principle", "correctness first", "--tension", "correctness vs economy",
			"--review-flag", "no", "--settled", "the grading stands as recorded",
			"--reopens-on", "a reproduction on a clean tree",
			"--reason-file", writeTemp(t, "the rationale")}
	}
	docketRuleArgs := func(as string) []string {
		return []string{"motion", "docket", "rule", "--id", "M1", "--as", as,
			"--principle", "correctness first", "--tension", "correctness vs economy",
			"--review-flag", "no", "--settled", "the grading stands as recorded",
			"--reopens-on", "a reproduction on a clean tree",
			"--reason-file", writeTemp(t, "the rationale")}
	}
	return entityProbe{
		name: "gap", id: "R1-1",
		states: []string{"unminted", "open", "closed"},
		acts: []probeAct{
			gapMint(),
			close,
			{"regrade", []string{"regrade", "--id", "R1-1", "--severity", "medium",
				"--reason", "the consequence is narrower than first graded"}},
			{"carry", []string{"carry", "--id", "R1-1", "--carried-from", "0", "--as", "repaired",
				"--reason", "carried from the prior round"}},
			{"opinion:carried", opinionArgs("carried")},
			{"opinion:not_a_defect", opinionArgs("not_a_defect")},
			// THE SAME TWO EXITS THROUGH THE NEW VERB, and both are driven because the pair is
			// the point: `carried` defers and leaves the gap OPEN, anything else ends it. A probe
			// that drove only the closing word would report a lifecycle whose defer state the
			// record can reach and the graph cannot see.
			{"motion docket file", []string{"motion", "docket", "file", "--id", "R1-1",
				"--reason", "contested and not mine to close"}},
			{"motion docket rule:carried", docketRuleArgs("carried")},
			{"motion docket rule:not_a_defect", docketRuleArgs("not_a_defect")},
		},
		buildTo: func(t *testing.T, state string) string {
			t.Helper()
			runDir := adversarialRun(t)
			if state == "unminted" {
				return runDir
			}
			if !build(t, runDir, gapMint()) {
				t.Fatalf("no seat could mint, so the gap probe would measure a board it never reached")
			}
			if state == "closed" && !build(t, runDir, close) {
				t.Fatalf("no seat could close, so state %q was never reached", state)
			}
			return runDir
		},
		read: func(t *testing.T, runDir string) string {
			t.Helper()
			g := gapOf(t, runDir, "R1-1")
			switch {
			case g == nil:
				return "unminted"
			case g.Open:
				return "open"
			default:
				return "closed"
			}
		},
		fields: func(t *testing.T, runDir string) map[string]string {
			t.Helper()
			g := gapOf(t, runDir, "R1-1")
			if g == nil {
				return map[string]string{}
			}
			var regrades []string
			for _, r := range g.Regrades {
				regrades = append(regrades, r.GetBasis())
			}
			// FINGERPRINT THE CARRIERS, NOT THE COLLAPSE. This read ClosureReason(), which is a
			// DERIVED accessor with a precedence rule — the bench's disposition wins over the
			// merge's closure_class — and the derivation made a FIRST write look like a
			// replacement: a bench opinion on a gap the merge had already closed reported as an
			// overwrite of "repaired" by "not_a_defect". Both are on the record, in their own
			// fields, exactly as Gap's own comment says ("a closure arrives as one of two
			// different messages"). Nothing is lost; a view collapsed them and the probe believed
			// the view.
			return map[string]string{
				"closure_class":     dispo(g.Closure.GetClosureClass()),
				"bench_disposition": dispo(g.BenchClosure.GetDisposition()),
				"severity":          grade(g.Severity),
				"likelihood":        grade(g.Likelihood),
				"impact":            grade(g.Impact),
				"complexity":        grade(g.ComplexityCost),
				// APPEND-ONLY, and listed in `accumulates`: a regrade REPLACES the grade and
				// KEEPS the argument for the change, so both gradings stay readable.
				"regrades": strings.Join(regrades, " | "),
			}
		},
		accumulates: []string{"regrades"},
		terminal: map[string]string{
			"closed": "A closed gap is the finish of the gap lifecycle: red closes on the merits or the bench disposes of it, and what comes after is a SUCCESSOR gap with its own id, minted with `--supersedes`, rather than a reopening of this one. The lineage is the exit, and it belongs to the successor.",
		},
		refusedForOtherReasons: map[string]string{
			"carry": "a carry restates a closure an EARLIER ROUND argued, and every probe here builds a single round-1 board — so its refusal is about the round, not about the gap's state. Driving it needs a two-round fixture, which is the fuzz's shape rather than this prober's.",
		},
	}
}

// dispo and grade render an enum as "" AT ITS ZERO, because the probe's "was this field already
// set" test compares against the empty string and a proto enum's zero stringifies as a NAME —
// `DISPOSITION_UNSPECIFIED`, not "". Left as its name, an unset field looked set, and the FIRST
// bench disposition on a gap reported as an overwrite of a value that was never there.
func dispo(d recordpb.Disposition) string {
	if d == recordpb.Disposition_DISPOSITION_UNSPECIFIED {
		return ""
	}
	return d.String()
}

func grade(g recordpb.Grade) string {
	if g == recordpb.Grade_GRADE_UNSPECIFIED {
		return ""
	}
	return g.String()
}

func gapOf(t *testing.T, runDir, id string) *record.Gap {
	t.Helper()
	rn, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatalf("opening the run: %v", err)
	}
	b, err := record.BoardState(rn)
	if err != nil {
		t.Fatalf("reading the board: %v", err)
	}
	return b.Gaps[id]
}

// ---- the motion ----

// motionProbe is the shape all three motion subjects share.
//
// `prelude` is what must exist before the motion can: the gap a grade motion argues over, the
// PROPOSAL a direction motion rules on. `order` maps each state to the act that reaches it, which
// is where the direction motion differs — it has no `file` verb at all, because its filing half is
// `line-of-inquiry propose` and the motion does not exist until red rules.
func motionProbe(subject, id string, prelude, acts []probeAct, states []string, order, terminal map[string]string) entityProbe {
	return entityProbe{
		name: "motion/" + subject, id: id,
		states: states,
		acts:   acts,
		buildTo: func(t *testing.T, state string) string {
			t.Helper()
			runDir := adversarialRun(t)
			for _, p := range prelude {
				if !build(t, runDir, p) {
					t.Fatalf("no seat could run the %q prelude a %s motion needs", p.name, subject)
				}
			}
			if state == states[0] {
				return runDir
			}
			for _, step := range states[1:] {
				var act *probeAct
				for i := range acts {
					if acts[i].name == order[step] {
						act = &acts[i]
					}
				}
				if act == nil {
					t.Fatalf("no %q act declared for subject %q, so state %q cannot be built", order[step], subject, state)
				}
				if !build(t, runDir, *act) {
					t.Fatalf("building %s/%s: no seat could run %q", subject, state, order[step])
				}
				if step == state {
					break
				}
			}
			return runDir
		},
		read: func(t *testing.T, runDir string) string {
			t.Helper()
			m := motionOf(t, runDir, id)
			switch {
			case m == nil:
				return states[0]
			case m.Appealed:
				return "appealed"
			case m.Ruled():
				return "ruled"
			default:
				return "filed"
			}
		},
		fields: func(t *testing.T, runDir string) map[string]string {
			t.Helper()
			m := motionOf(t, runDir, id)
			if m == nil {
				return map[string]string{}
			}
			return map[string]string{
				"basis": m.Basis, "relief": m.Relief, "ruling": m.Ruling,
				"opinion": m.Opinion, "appeal_reason": m.AppealReason,
			}
		},
		terminal: terminal,
	}
}

func motionOf(t *testing.T, runDir, id string) *record.Motion {
	t.Helper()
	rn, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatalf("opening the run: %v", err)
	}
	b, err := record.BoardState(rn)
	if err != nil {
		t.Fatalf("reading the board: %v", err)
	}
	for _, m := range record.Motions(b) {
		if m.ID == id {
			return m
		}
	}
	return nil
}

func gradeMotionProbe() entityProbe {
	return motionProbe("grade", "M1", []probeAct{gapMint()}, []probeAct{
		{"file", []string{"motion", "grade", "file", "--id", "R1-1", "--dimension", "severity",
			"--proposed", "low", "--reason", "the consequence is bounded by the caller's own validation"}},
		{"rule", []string{"motion", "grade", "rule", "--id", "M1", "--as", "rejected",
			"--reason", "the evidence does not reach it"}},
		{"appeal", []string{"motion", "grade", "appeal", "--id", "M1",
			"--reason", "pressing it on new grounds"}},
	}, []string{"unfiled", "filed", "ruled", "appealed"},
		map[string]string{"filed": "file", "ruled": "rule", "appealed": "appeal"},
		map[string]string{
			"appealed": "BY DESIGN, and debate.js says so in blue's own prompt: \"appeal whether or not you go on to pursue the line, because the appeal is where your ARGUMENT is recorded and the fate is only what you decided to do about it.\" An appeal is a recorded dissent, not a request for a second ruling — record.RequireUnruledMotion refuses that in terms, and internal/report/motions.go renders the appeal as the filer's position.",
		})
}

func petitionMotionProbe() entityProbe {
	return motionProbe("petition", "M1", nil, []probeAct{
		{"file", []string{"motion", "petition", "file", "--class", "safety",
			"--relief", "halt before the next round",
			"--reason", "continuing would require asserting a consent gate that does not exist"}},
		{"rule", []string{"motion", "petition", "rule", "--id", "M1", "--as", "granted",
			"--reason", "granted on the papers"}},
	}, []string{"unfiled", "filed", "ruled"},
		map[string]string{"filed": "file", "ruled": "rule"},
		map[string]string{
			"ruled": "A petition has no appeal and the surface says so: `motion petition appeal` does not exist, and the refusal is asserted in adversarial_test.go. A petition is heard BEFORE the debate continues, so there is nothing to escalate to.",
		})
}

// inquiryMotionProbe is the DIRECTION motion: the third subject, and the one shaped differently.
//
// It has no `file` verb. Its filing half is `line-of-inquiry propose`, which records an `avenue`
// rather than a motion event, and record.Motions only assembles the motion once a RULING arrives —
// so `unruled` here means "the proposal exists and nobody has ruled it", not "nothing exists".
// That asymmetry is why this subject was left unprobed in the first pass, and stating it was not
// the same as covering it.
func inquiryMotionProbe() entityProbe {
	return motionProbe("inquiry", "Q1", []probeAct{inquiryPropose()}, []probeAct{
		{"rule", []string{"motion", "inquiry", "rule", "--id", "Q1", "--as", "endorsed",
			"--reason", "worth this run's time"}},
		{"appeal", []string{"motion", "inquiry", "appeal", "--id", "Q1",
			"--reason", "pressing the ruling on new grounds"}},
	}, []string{"unruled", "ruled", "appealed"},
		map[string]string{"ruled": "rule", "appealed": "appeal"},
		map[string]string{
			"appealed": "The same finish as a grade motion's, for the same reason and from the same prompt: debate.js tells blue \"Red rules on your proposals and you may APPEAL a ruling — appeal whether or not you go on to pursue the line, because the appeal is where your ARGUMENT is recorded and the fate is only what you decided to do about it.\" The LINE itself stays movable; the motion's answer does not reopen.",
		})
}

// ---- the line of inquiry ----
//
// Its lifecycle has TWO AXES and only one of them is the state. Blue MOVES a line between
// `proposed`, `pursued`, `declined`, `abandoned` and `deferred`; red RULES it `endorsed`,
// `out_of_scope` or `too_thin`. The status is the state — it is what the line IS — and the ruling
// is content, which is why a first ruling reads here as an ordinary act that does not move the
// state rather than as a transition.

func inquiryPropose() probeAct {
	return probeAct{"propose", []string{"line-of-inquiry", "propose",
		"--reason", "a line worth taking", "--hypothesis", "it would settle the open question"}}
}

func inquiryProbe() entityProbe {
	moveTo := func(status string) probeAct {
		return probeAct{"move:" + status, []string{"line-of-inquiry", "move", "--id", "Q1",
			"--as", status, "--reason", "what became of it"}}
	}
	return entityProbe{
		name: "inquiry", id: "Q1",
		// Every status the vocabulary offers is built and probed, because a fate nobody can leave
		// is exactly what this gate is for and `deferred` and `abandoned` are the two most likely
		// to be one.
		states: []string{"unproposed", "proposed", "pursued", "declined", "abandoned", "deferred"},
		acts: []probeAct{
			inquiryPropose(),
			moveTo("pursued"), moveTo("declined"), moveTo("abandoned"), moveTo("deferred"),
			{"rule", []string{"motion", "inquiry", "rule", "--id", "Q1", "--as", "endorsed",
				"--reason", "worth this run's time"}},
			{"appeal", []string{"motion", "inquiry", "appeal", "--id", "Q1",
				"--reason", "pressing the ruling on new grounds"}},
		},
		buildTo: func(t *testing.T, state string) string {
			t.Helper()
			runDir := adversarialRun(t)
			if state == "unproposed" {
				return runDir
			}
			if !build(t, runDir, inquiryPropose()) {
				t.Fatalf("no seat could propose a line of inquiry")
			}
			if state != "proposed" {
				if !build(t, runDir, moveTo(state)) {
					t.Fatalf("no seat could move a line of inquiry to %q", state)
				}
			}
			return runDir
		},
		read: func(t *testing.T, runDir string) string {
			t.Helper()
			q := inquiryOf(t, runDir, "Q1")
			if q == nil {
				return "unproposed"
			}
			return q.Status
		},
		fields: func(t *testing.T, runDir string) map[string]string {
			t.Helper()
			q := inquiryOf(t, runDir, "Q1")
			if q == nil {
				return map[string]string{}
			}
			return map[string]string{
				"line": q.Line, "hypothesis": q.Hypothesis, "reason": q.Reason,
				"ruling": q.Ruling, "ruling_why": q.RulingWhy,
				// APPEND-ONLY, and listed in `accumulates`: History is a line per move, so a
				// re-recorded fate keeps every earlier one beside it.
				"history": strings.Join(q.History, " | "),
			}
		},
		accumulates: []string{"history"},
		// NOTHING IS DECLARED TERMINAL HERE, and that is a claim rather than an omission: a line
		// of inquiry is a LIVING RECORD — debate.js tells blue "every round, revisit what is still
		// open and say what became of it" — so every fate must remain movable. A fate that could
		// not be left would be a line blue is forbidden to revisit, which is the opposite of what
		// the lifecycle was built for.
		terminal: map[string]string{},
	}
}

func inquiryOf(t *testing.T, runDir, id string) *record.Inquiry {
	t.Helper()
	rn, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatalf("opening the run: %v", err)
	}
	b, err := record.BoardState(rn)
	if err != nil {
		t.Fatalf("reading the board: %v", err)
	}
	for _, q := range record.Inquiries(b) {
		if q.ID == id {
			return q
		}
	}
	return nil
}

// ---- coverage: is every verb that touches these entities actually probed? ----

// entityEvents are the record event types each probed entity can be touched by. Declared, because
// "which events belong to a gap" is a fact about the DOMAIN and is not written down anywhere the
// code can be asked for it; checked against the surface below, so the pairing cannot go stale
// silently.
var entityEvents = map[string][]string{
	// motion_rule is here as well as under "motion" because a DOCKET ruling closes a gap: the
	// bench's disposition is a motion ruling now, so the gap's third exit is an event the motion
	// entity also owns. One event, two entities, and saying so is what keeps both graphs honest.
	"gap":     {"mint", "close", "regrade", "opinion", "motion_rule"},
	"motion":  {"motion", "motion_rule", "motion_appeal"},
	"inquiry": {"avenue"},
}

// TestEveryVerbThatTouchesAProbedEntityIsProbed keeps the act tables honest.
//
// The tables are hand-written because a verb's FLAGS cannot be derived, but the SET of verbs can:
// cli.CommandRecords() maps every command path to the event it records. A verb added to the surface
// and not to a probe would go unprobed, and an unprobed verb is a transition this gate silently
// does not know about — the miss that reads exactly like a clean graph.
func TestEveryVerbThatTouchesAProbedEntityIsProbed(t *testing.T) {
	t.Parallel()
	want := map[string]bool{}
	for _, evs := range entityEvents {
		for _, ev := range evs {
			want[ev] = true
		}
	}

	var surface []string
	for path, ev := range CommandRecords() {
		if want[ev] {
			surface = append(surface, path)
		}
	}
	sort.Strings(surface)
	if len(surface) == 0 {
		t.Fatal("CommandRecords() reports NO verb recording any of these events — the walker is broken, and an empty surface would make the coverage check below vacuous")
	}

	probed := map[string]bool{}
	for _, p := range []entityProbe{gapProbe(t), gradeMotionProbe(), petitionMotionProbe(), inquiryMotionProbe(), inquiryProbe()} {
		for _, a := range p.acts {
			var path []string
			for _, w := range a.args {
				if strings.HasPrefix(w, "-") {
					break
				}
				path = append(path, w)
			}
			// The tree is per-role, so a probe invokes `close` where the surface walker names it
			// `merge close`. Both spellings are recorded; the suffix match below is what joins them.
			probed[strings.Join(path, " ")] = true
		}
	}

	// notProbed are surface verbs deliberately left out, each with the reason. A verb in NEITHER
	// list fails: that is what makes a newly-added one loud instead of silently unprobed.
	notProbed := map[string]string{
		"merge carry": "a carry restates a closure an EARLIER ROUND argued, and every probe here builds a single round-1 board, so it can only ever be refused for a reason that is not about the gap's state. Driving it needs a two-round fixture, which is the fuzz's shape rather than this prober's. It IS attempted and reported as `refused (not a state fact)` so the absence is visible in the matrix rather than only here.",
	}

	var missing []string
	for _, path := range surface {
		hit := false
		for p := range probed {
			if path == p || strings.HasSuffix(path, " "+p) {
				hit = true
			}
		}
		if hit {
			continue
		}
		if _, stated := notProbed[path]; stated {
			continue
		}
		missing = append(missing, path)
	}
	t.Logf("verbs touching a probed entity: %d · covered: %d · stated-unprobed: %d", len(surface), len(surface)-len(missing)-len(notProbed), len(notProbed))
	if len(missing) > 0 {
		t.Errorf("these verbs record an event that touches a probed entity and are neither probed nor stated as unprobed: %s\n\n"+
			"Add them to the entity's acts, or to notProbed WITH THE REASON. A verb in neither list "+
			"is a transition this gate does not know about, and its absence reads exactly like a "+
			"state graph with nothing to find.", strings.Join(missing, ", "))
	}
	for path := range notProbed {
		found := false
		for _, s := range surface {
			if s == path {
				found = true
			}
		}
		if !found {
			t.Errorf("%q is listed as stated-unprobed but records no such event any more — the reason is stale and the list is now excusing nothing", path)
		}
	}
}

// ---- the gate ----

// TestNoEntityCanReachAStateNothingCanLeave is the dead-end and silent-overwrite gate.
func TestNoEntityCanReachAStateNothingCanLeave(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))

	probes := []entityProbe{gapProbe(t), gradeMotionProbe(), petitionMotionProbe(), inquiryMotionProbe(), inquiryProbe()}

	var edges []stateEdge
	var deadEnds, overwrites []string
	attempts := 0
	for _, p := range probes {
		e, d, o, n := probeEntity(t, p)
		edges = append(edges, e...)
		deadEnds = append(deadEnds, d...)
		overwrites = append(overwrites, o...)
		attempts += n
	}

	// THE FLOOR. Every assertion below is vacuously true over an empty probe, which is the shape
	// that lets a sweep pass by traversing nothing.
	if attempts == 0 || len(edges) == 0 {
		t.Fatalf("the probe attempted %d act(s) and found %d outcome(s) — it measured nothing, and an empty graph reads exactly like a graph with no dead ends", attempts, len(edges))
	}
	moved := 0
	for _, e := range edges {
		if strings.HasPrefix(e.outcome, "-> ") {
			moved++
		}
	}
	if moved == 0 {
		t.Fatal("NO act moved any entity — every outcome was refused, inert or an overwrite, which means the prober is not driving the verbs and its silence about dead ends is worthless")
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].entity != edges[j].entity {
			return edges[i].entity < edges[j].entity
		}
		if edges[i].from != edges[j].from {
			return edges[i].from < edges[j].from
		}
		return edges[i].act < edges[j].act
	})
	var b strings.Builder
	fmt.Fprintf(&b, "state graph (#673): %d entities · %d act(s) attempted · %d moved\n", len(probes), attempts, moved)
	// THE LEGEND IS NOT DECORATION. `inert` means "no change to the fields THIS PROBE READS",
	// which is a narrower claim than "the act did nothing" — a bench opinion disposing `carried`
	// writes an event the record keeps and the Gap projection does not carry, and reporting that
	// as "did nothing" would be the matrix asserting something it did not measure.
	b.WriteString("  -> X = moved · records = first write into an empty field · appends = an" +
		" append-only field grew · inert = no change to the fields this probe reads ·" +
		" refused = no seat was allowed it\n")
	for _, e := range edges {
		fmt.Fprintf(&b, "  %-16s %-11s --%-15s %s\n", e.entity, e.from, e.act, e.outcome)
	}
	t.Log(b.String())

	for _, o := range overwrites {
		t.Errorf("SILENT OVERWRITE: an act was accepted, left the entity in the same state, and REWROTE a field that already had a value.\n\n  %s\n\n"+
			"Both events stay on the record and replay keeps the LAST, so the earlier value stops "+
			"being the answer with nothing saying so. record.RequireUnruledMotion refuses exactly "+
			"this for a second ruling, in those words, and record.RequireUnappealedMotion for a "+
			"second appeal; whatever act is named above has no such guard.", o)
	}
	for _, d := range deadEnds {
		t.Errorf("DEAD END: an entity in state %q has no act that moves it, and nothing declares that state a finish.\n\n"+
			"Either the state is terminal — say so in the probe's `terminal` map with the reason, "+
			"the way gap/closed and motion/grade/appealed do — or an entity can reach a position "+
			"the protocol has no way out of, which is the defect this gate exists to find.", d)
	}
}
