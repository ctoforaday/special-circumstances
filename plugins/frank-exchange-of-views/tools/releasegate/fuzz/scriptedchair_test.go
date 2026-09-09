package fuzz

// A SCRIPTED CHAIR FOR TESTS THAT HAVE NO RECORD.
//
// The dispatch loop reads who sits from the chair's relayed plan, and in production the plan is
// `dispatch next`'s — computed from the board. The prompt-delivery tests drive debate.js through
// debatejs.Capture with no record behind it, so nothing can compute a plan for them; before
// roundless the loop ran a fixed number of rounds and these tests rode along. Now the schedule
// is the TEST'S, stated in full: which seats sit in which epoch, and that the chair PASSes once
// the script runs out. A test that needs "a blue sitting after the bench sat" says so here rather
// than hoping the loop happens to arrange it.

// benchThenBlue is the schedule the delivery tests need: blue sits once with nothing ruled (so
// the rulings clause can be shown conditional), the bench rules, then both parties sit again with
// the ruling in effect, then the chair passes.
var benchThenBlue = []map[string]any{
	plan([]any{party("red-lens-evidence"), party("blue-respond", "G1", "G2")}, false, false, nil, "opening"),
	plan([]any{party("judge", "G1")}, false, false, []any{"G1"}, "G1 at impasse"),
	plan([]any{party("red-lens-evidence", "G1"), party("blue-respond", "G1")}, false, false, nil, "G1 ruled and open"),
}

// relayScriptedPlan puts the sitting's plan on a chair envelope: the n-th chair sitting relays
// plans[n-1], and a sitting past the end relays an empty PASS-permitted plan with a PASS verdict,
// which is how the loop ends. The sitting ordinal is read off the label debate.js assigns
// (`red-chair #n · slug`), the same way the termination gate reads it.
func relayScriptedPlan(e map[string]any, label string, plans []map[string]any) {
	n := sittingOf(label)
	if n >= 1 && n <= len(plans) {
		e["plan"] = plans[n-1]
		e["verdict"] = "FAIL"
	} else {
		e["plan"] = plan([]any{}, true, false, nil, "every lens sat against the head and nothing material is open")
		e["verdict"] = "PASS"
	}
	e["unruled_motions"] = 0
}
