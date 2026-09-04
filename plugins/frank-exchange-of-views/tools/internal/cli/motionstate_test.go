package cli

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE STATE GRAPH FOR A MOTION (#673): which event may legally follow which, probed by EXECUTION.
//
// #535 split this out because the SURFACE graph — role → verb → event, gated since #669 — can tell
// you a verb is unreachable but not that an ENTITY can get stuck. "Dead end" means nothing without
// this: an entity state with no outgoing transition that is not terminal is a defect, and nothing
// asked the question.
//
// DERIVED, NOT MODELLED, which is #535's own rule. There is no transition table here to drift from
// the code: each state is BUILT by running the real verbs against a real record, each act is then
// attempted for real, and the resulting state is read back off record.Motions. What the matrix
// reports is what the shipped write path did.
//
// SEAT SCOPE IS FACTORED OUT ON PURPOSE. `motion grade rule` is the merge's verb and the bench is
// refused it — a fact about ROLE, which the surface graph already owns, and not about the motion's
// state. So each act is attempted under every registered seat and counts as possible if ANY seat is
// allowed it. The question here is only "can this motion get from here to there at all".

// motionState is a motion's position in its lifecycle, read off the record rather than tracked.
type motionState string

const (
	stateUnfiled  motionState = "unfiled"
	stateFiled    motionState = "filed"
	stateRuled    motionState = "ruled"
	stateAppealed motionState = "appealed"
)

// stateOf reads a motion's state from the board. The absence of the motion is a state too, and it
// is named rather than returned as an empty string that would compare equal to a bug.
func stateOf(t *testing.T, runDir, id string) motionState {
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
		if m.ID != id {
			continue
		}
		switch {
		case m.Appealed:
			return stateAppealed
		case m.Ruled():
			return stateRuled
		default:
			return stateFiled
		}
	}
	return stateUnfiled
}

// fingerprintOf renders a motion's whole content, so an act that leaves the STATE alone but
// rewrites the entity is visible. State is a coarse reading — `appealed` is `appealed` whichever
// argument it carries — and the difference between "accepted and inert" and "accepted and it
// quietly rewrote the record's answer" is exactly what a coarse reading loses.
func fingerprintOf(t *testing.T, runDir, id string) string {
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
		if m.ID != id {
			continue
		}
		var keys []string
		for k := range m.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var f strings.Builder
		fmt.Fprintf(&f, "subject=%s filer=%s round=%d basis=%q relief=%q ruling=%q rulingBy=%s rulingRound=%d opinion=%q appealed=%v appealReason=%q",
			m.Subject, m.Filer, m.Round, m.Basis, m.Relief, m.Ruling, m.RulingBy, m.RulingRound, m.Opinion, m.Appealed, m.AppealReason)
		for _, k := range keys {
			fmt.Fprintf(&f, " %s=%q", k, m.Fields[k])
		}
		return f.String()
	}
	return "(no such motion)"
}

// probeSeats are the seats each act is attempted under. Every role that can touch a motion is here,
// because an act refused by one seat and allowed by another is allowed.
var probeSeats = []string{"red-lens-r1-L1", "red-merge-r1", "blue-respond-r1", "judge-r1"}

// motionAct is one attempted transition. `args` carries no --seat-id and no --run: the prober
// supplies both, once per seat it tries.
type motionAct struct {
	name string
	args []string
}

// gradeActs and petitionActs are the acts for the two subjects a test can build end to end.
//
// The SET is checked against the surface walker below rather than trusted: cli.CommandRecords()
// knows every verb that records a motion event, and a verb missing from these tables fails
// TestTheMotionActsCoverEveryVerbThatRecordsAMotionEvent instead of going silently unprobed. That
// is the #666 lesson applied — the roster is derived, the payload is declared, and a missing
// declaration is loud.
func gradeActs() []motionAct {
	return []motionAct{
		{"file", []string{"motion", "grade", "file", "--id", "R1-1", "--dimension", "severity",
			"--proposed", "low", "--reason", "the consequence is bounded by the caller's own validation"}},
		{"rule", []string{"motion", "grade", "rule", "--id", "M1", "--as", "rejected",
			"--reason", "the evidence does not reach it"}},
		{"appeal", []string{"motion", "grade", "appeal", "--id", "M1",
			"--reason", "pressing it on new grounds"}},
	}
}

func petitionActs() []motionAct {
	return []motionAct{
		{"file", []string{"motion", "petition", "file", "--class", "safety",
			"--relief", "halt before the next round",
			"--reason", "continuing would require asserting a consent gate that does not exist"}},
		{"rule", []string{"motion", "petition", "rule", "--id", "M1", "--as", "granted",
			"--reason", "granted on the papers"}},
	}
}

// buildTo walks a fresh run to the named state and returns the run directory.
func buildTo(t *testing.T, want motionState, subject string, acts []motionAct) string {
	t.Helper()
	runDir := adversarialRun(t)
	if subject == "grade" {
		mint := []string{"mint", "--seat-id", "red-merge-r1", "--run", runDir,
			"--key", "k1", "--class", "self-attestation", "--problem", "p", "--fix", "f",
			"--check", "c", "--check-kind", "document", "--severity", "high", "--likelihood", "high",
			"--impact", "high", "--complexity", "low", "--reason", "the board needs something to argue about"}
		if _, err := run(t, mint...); err != nil {
			t.Fatalf("minting the gap the grade motion argues over: %v", err)
		}
	}
	order := []motionState{stateFiled, stateRuled, stateAppealed}
	for _, step := range order {
		if step == stateFiled && want == stateUnfiled {
			break
		}
		name := map[motionState]string{stateFiled: "file", stateRuled: "rule", stateAppealed: "appeal"}[step]
		var act *motionAct
		for i := range acts {
			if acts[i].name == name {
				act = &acts[i]
			}
		}
		if act == nil {
			t.Fatalf("no %q act declared for subject %q, so state %q cannot be built", name, subject, want)
		}
		if !build(t, runDir, *act) {
			t.Fatalf("building %s/%s: no seat could run %q, so the probe would measure a board it never reached", subject, want, name)
		}
		if step == want {
			break
		}
	}
	return runDir
}

// probeReason is what an ATTEMPTED act says, as distinct from what the SETUP said.
//
// It is not decoration. The setup builds a state by running the same acts, so an act attempted
// with the setup's own wording writes back the bytes already there — and an overwrite becomes
// invisible, because the fingerprint before and after are equal. That is the shape this gate is
// here to catch reporting itself as absent: measured, the second appeal on an already-appealed
// motion read as `inert` until the wording differed.
const probeReason = "PROBE: this wording exists only to make an overwrite visible"

// withProbeReason replaces the act's --reason so a rewrite of the SAME field shows as a change.
func withProbeReason(args []string) []string {
	out := append([]string{}, args...)
	for i := 0; i < len(out)-1; i++ {
		if out[i] == "--reason" {
			out[i+1] = probeReason
		}
	}
	return out
}

// attempt runs one act under every probe seat and reports whether ANY was allowed.
func attempt(t *testing.T, runDir string, act motionAct) bool {
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
func build(t *testing.T, runDir string, act motionAct) bool {
	t.Helper()
	for _, seat := range probeSeats {
		args := append(append([]string{}, act.args...), "--seat-id", seat, "--run", runDir)
		if _, err := run(t, args...); err == nil {
			return true
		}
	}
	return false
}

// terminalStates are the states from which nothing more is expected, each with the reason it is a
// finish rather than a stall. THIS IS THE ONE DECLARED THING HERE, and it has to be: whether a
// state SHOULD have an exit is design intent, and no probe can read intent off the code. Being on
// this list means somebody looked.
var terminalStates = map[string]string{
	"grade/appealed": "BY DESIGN, and debate.js says so in blue's own prompt: \"appeal whether or not you go on to pursue the line, because the appeal is where your ARGUMENT is recorded and the fate is only what you decided to do about it.\" An appeal is a recorded dissent, not a request for a second ruling — record.RequireUnruledMotion refuses that in terms, and internal/report/motions.go renders the appeal as the filer's position. Nothing is owed an answer here.",
	"petition/ruled": "A petition has no appeal and the surface says so: `motion petition appeal` does not exist, and the refusal is asserted in adversarial_test.go. A petition is heard BEFORE the debate continues, so there is nothing to escalate to.",
}

// TestAMotionCannotReachAStateNothingCanLeave is the dead-end gate.
func TestAMotionCannotReachAStateNothingCanLeave(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))

	subjects := []struct {
		name  string
		acts  []motionAct
		build []motionState
	}{
		{"grade", gradeActs(), []motionState{stateUnfiled, stateFiled, stateRuled, stateAppealed}},
		{"petition", petitionActs(), []motionState{stateUnfiled, stateFiled, stateRuled}},
	}

	// Every attempt lands in exactly one of these, and the matrix prints all four. A gate that
	// showed only the accepted transitions could not tell "refused" from "never attempted", and
	// those are the two readings a silent no-match sits between.
	type edge struct{ subject, from, act, outcome string }
	var edges []edge
	var deadEnds, overwrites []string
	probed := 0

	for _, sub := range subjects {
		for _, from := range sub.build {
			out := 0
			for _, act := range sub.acts {
				probed++
				// A FRESH BOARD PER ATTEMPT. An act that succeeds moves the motion, and the
				// next act would then be probed from somewhere else entirely.
				dir := buildTo(t, from, sub.name, sub.acts)
				if got := stateOf(t, dir, "M1"); got != from {
					t.Fatalf("%s: built for state %q and the board reads %q — the probe would report transitions out of a state it is not in", sub.name, from, got)
				}
				before := fingerprintOf(t, dir, "M1")
				if !attempt(t, dir, act) {
					edges = append(edges, edge{sub.name, string(from), act.name, "refused"})
					continue
				}
				to := stateOf(t, dir, "M1")
				after := fingerprintOf(t, dir, "M1")
				switch {
				case to != from:
					out++
					edges = append(edges, edge{sub.name, string(from), act.name, "-> " + string(to)})
				case after != before:
					// ACCEPTED, THE STATE UNCHANGED, AND THE ENTITY REWRITTEN. This is the shape
					// RequireUnruledMotion exists to refuse, one verb over: both events stay on
					// the record and the replayed answer is whichever came last, so a position
					// stops being the answer without anything saying so.
					edges = append(edges, edge{sub.name, string(from), act.name, "OVERWRITES"})
					overwrites = append(overwrites, fmt.Sprintf("%s/%s --%s-->\n     before: %s\n     after:  %s",
						sub.name, from, act.name, before, after))
				default:
					edges = append(edges, edge{sub.name, string(from), act.name, "inert"})
				}
			}
			key := sub.name + "/" + string(from)
			if out == 0 {
				if _, declared := terminalStates[key]; !declared {
					deadEnds = append(deadEnds, key)
				}
			} else if reason, declared := terminalStates[key]; declared {
				t.Errorf("%s is declared TERMINAL but %d act(s) move a motion out of it — the declaration is stale and is now excusing a state that has exits.\ndeclared reason: %s", key, out, reason)
			}
		}
	}

	// THE FLOOR. Every assertion above is vacuously true over an empty probe, which is the shape
	// that lets a sweep pass by traversing nothing.
	if probed == 0 || len(edges) == 0 {
		t.Fatalf("the probe attempted %d act(s) and found %d edge(s) — it measured nothing, and an empty graph reads exactly like a graph with no dead ends", probed, len(edges))
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].subject != edges[j].subject {
			return edges[i].subject < edges[j].subject
		}
		if edges[i].from != edges[j].from {
			return edges[i].from < edges[j].from
		}
		return edges[i].act < edges[j].act
	})
	var b strings.Builder
	fmt.Fprintf(&b, "motion state graph (#673): %d act(s) attempted, %d outcome(s)\n", probed, len(edges))
	for _, e := range edges {
		fmt.Fprintf(&b, "  %-9s %-9s --%-7s %s\n", e.subject, e.from, e.act, e.outcome)
	}
	t.Log(b.String())

	for _, o := range overwrites {
		t.Errorf("SILENT OVERWRITE: an act was accepted, left the motion in the same state, and REWROTE it.\n\n  %s\n\n"+
			"Both events stay on the record and replay keeps the LAST, so the earlier position stops "+
			"being the answer with nothing saying so. record.RequireUnruledMotion refuses exactly "+
			"this for a second ruling, in those words; whatever act is named above has no such guard.", o)
	}
	for _, d := range deadEnds {
		t.Errorf("DEAD END: a motion in state %q has no act that moves it, and nothing declares that state a finish.\n\n"+
			"Either the state is terminal — say so in terminalStates with the reason, the way "+
			"grade/appealed and petition/ruled do — or an entity can reach a position the protocol "+
			"has no way out of, which is the defect this gate exists to find.", d)
	}
}

// TestTheMotionActsCoverEveryVerbThatRecordsAMotionEvent keeps the probe's act tables honest.
//
// The tables are hand-written because a verb's FLAGS cannot be derived, but the SET of verbs can:
// cli.CommandRecords() maps every command path to the event it records, and the motion events are
// a closed set. A verb added to the surface and not to these tables would go unprobed, and an
// unprobed verb is a transition this gate silently does not know about — the miss that reads
// exactly like a clean graph.
func TestTheMotionActsCoverEveryVerbThatRecordsAMotionEvent(t *testing.T) {
	t.Parallel()
	motionEvents := map[string]bool{"motion": true, "motion_rule": true, "motion_appeal": true}

	var surface []string
	for path, ev := range CommandRecords() {
		if motionEvents[ev] {
			surface = append(surface, path)
		}
	}
	sort.Strings(surface)
	if len(surface) == 0 {
		t.Fatal("CommandRecords() reports NO verb recording a motion event — the walker is broken, and an empty surface would make the coverage check below vacuous")
	}

	probedPaths := map[string]bool{}
	for _, acts := range [][]motionAct{gradeActs(), petitionActs()} {
		for _, a := range acts {
			// The command path is the leading non-flag words.
			var path []string
			for _, w := range a.args {
				if strings.HasPrefix(w, "-") {
					break
				}
				path = append(path, w)
			}
			probedPaths[strings.Join(path, " ")] = true
		}
	}

	// The inquiry subject is NOT probed and that is stated rather than left to be noticed: its
	// filing verb is `blue line-of-inquiry propose`, which records `avenue` rather than a motion
	// event, so a direction motion is built through a different door than the other two. It is
	// listed here so the gap is a decision on the record instead of an absence.
	notProbed := map[string]string{
		"motion inquiry rule":   "a direction motion is FILED by `blue line-of-inquiry propose` (event `avenue`), not by a `motion … file` verb, so building one needs a different setup than grade and petition. Left for the inquiry lifecycle rather than bolted onto the motion prober.",
		"motion inquiry appeal": "as above — the same subject, the same different door.",
	}

	var missing []string
	for _, path := range surface {
		if probedPaths[path] {
			continue
		}
		if _, stated := notProbed[path]; stated {
			continue
		}
		missing = append(missing, path)
	}
	t.Logf("motion verbs on the surface: %d · probed: %d · stated-unprobed: %d", len(surface), len(probedPaths), len(notProbed))
	if len(missing) > 0 {
		t.Errorf("these verbs record a motion event and are neither probed nor stated as unprobed: %s\n\n"+
			"Add them to gradeActs/petitionActs, or to notProbed WITH THE REASON. A verb in neither "+
			"list is a transition this gate does not know about, and its absence reads exactly like "+
			"a state graph with nothing to find.", strings.Join(missing, ", "))
	}
	for path := range notProbed {
		found := false
		for _, s := range surface {
			if s == path {
				found = true
			}
		}
		if !found {
			t.Errorf("%q is listed as stated-unprobed but records no motion event any more — the reason is stale and the list is now excusing nothing", path)
		}
	}
}
