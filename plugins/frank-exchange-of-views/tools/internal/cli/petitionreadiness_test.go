package cli

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE BENCH RULES EVERY PETITION AND FILES NONE (#1191) — the rule every gavel is under
// (TestNoGavelHolderFilesItsOwnSubject). It halts the run itself where it must; `file` is off its
// surface, and a bench that reaches for it is told why rather than that the verb does not exist.
func TestTheBenchFilesNoPetition(t *testing.T) {
	if h := help(t, "motion", "petition", "--help", "--seat-id", "judge"); strings.Contains(h, "\n  file ") {
		t.Errorf("the bench's petition group offers `file`:\n%s", h)
	}
	if h := help(t, "motion", "petition", "--help", "--seat-id", "red-lens-evidence"); !strings.Contains(h, "\n  file ") {
		t.Errorf("a lens lost its petition right:\n%s", h)
	}
	runDir := newRun(t)
	stageEvents(t, runDir)
	out, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "judge",
		"--class", "safety", "--relief", "stop", "--reason", "the bench petitioning itself")
	if err == nil || !strings.Contains(out+err.Error(), "a petition motion is ruled by the bench seat, so `file` is not on your surface") ||
		!strings.Contains(out+err.Error(), "`halt` is yours") {
		t.Errorf("the bench filed a petition, or was refused without the reason: %v\n%s", err, out)
	}
}

// A PETITION A LENS FILES ON THE RECORD CONVENES THE BENCH AT THE NEXT CHAIR SITTING (#1203). A lens
// returns prose, so the record is the only channel its petition has: the verb wrote the motion and
// `dispatch next` readied nobody, and the petition waited for an exit that might never rule it.
// Driven through the real verbs: the lens files, the chair's plan readies the bench for a PETITION
// sitting and relays the motion as a bench-owned blocker; the bench sits for it and rules; the next
// plan readies nobody for it.
func TestAPetitionALensFilesConvenesTheBench(t *testing.T) {
	runDir := newRun(t)
	stageEvents(t, runDir)
	out, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-lens-evidence", "--json",
		"--class", "safety", "--relief", "stop and review", "--reason", "the report instructs a hazardous procedure")
	if err != nil {
		t.Fatalf("petition file: %v\n%s", err, out)
	}
	var filed struct {
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &filed); err != nil || filed.Result.MotionID == "" {
		t.Fatalf("the filing names no motion id: %v\n%s", err, out)
	}
	id := filed.Result.MotionID

	out, err = run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	plan := planOf(t, out)
	i := slices.IndexFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "judge" })
	if i < 0 {
		t.Fatalf("the plan readies no bench for petition %s: parties %+v, why %q", id, plan.Parties, plan.Why)
	}
	if bench := plan.Parties[i]; !slices.Equal(bench.Occasions, []string{"petition"}) || len(bench.GapIDs) != 0 {
		t.Errorf("bench party = %+v, want occasions [petition] and no gaps", bench)
	}
	if !slices.Contains(plan.Blockers, record.PlanBlocker{Kind: record.BlockerUnruledMotion, Subject: id, Owner: "judge"}) {
		t.Errorf("plan.blockers = %+v, want %s owned by the bench", plan.Blockers, id)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted over an unheard petition")
	}
	rows, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_DISPATCH)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(rows, func(e *record.Event) bool {
		d, _ := recordpb.BodyAs[*recordpb.Dispatch](e)
		return d.GetSeatId() == "judge" && len(d.GetGapIds()) == 0
	}) {
		t.Errorf("the record holds no dispatch row for the bench — the readiness the plan printed is not on the record")
	}

	// The bench sits for it and rules it; the chair's next plan readies nobody for it.
	for _, args := range [][]string{
		{"register", "--seat-id", "judge", "--occasion", "petition"},
		{"motion", "petition", "rule", "--seat-id", "judge", "--id", id, "--as", "denied", "--reason", "the procedure is described, not instructed"},
		{"register", "--seat-id", "red-chair"},
	} {
		if out, err := run(t, append(args, "--run", runDir)...); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	out, err = run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	plan = planOf(t, out)
	if slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "judge" }) {
		t.Errorf("the bench is readied again after ruling %s: %+v", id, plan.Parties)
	}
	if slices.ContainsFunc(plan.Blockers, func(b record.PlanBlocker) bool { return b.Subject == id }) {
		t.Errorf("the ruled petition %s still blocks: %+v", id, plan.Blockers)
	}
}

// A PETITION IS HEARD ALONE, AND THE ENGINE LEARNS OF A HALT FROM THE NEXT PLAN. The engine reads the
// record only through the chair's plan; the #1663 smoke's terminal judge returned "Not halted" in an
// envelope's halt field and the engine stamped HALTED over a record with no halt. The plan that
// convenes the bench for a petition readies nobody else, so the plan after its ruling — `halted` where
// the bench halted the run, and nobody dispatched — is read before any other party sits.
func TestAPetitionIsHeardAloneAndAHaltEndsDispatch(t *testing.T) {
	plan := func(runDir string) record.Plan {
		t.Helper()
		out, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
		if err != nil {
			t.Fatalf("dispatch next: %v\n%s", err, out)
		}
		return planOf(t, out)
	}
	act := func(runDir string, args ...string) {
		t.Helper()
		if out, err := run(t, append(args, "--run", runDir)...); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	file := func(runDir string) string {
		t.Helper()
		out, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-lens-evidence", "--json",
			"--class", "safety", "--relief", "stop and review", "--reason", "the report instructs a hazardous procedure")
		if err != nil {
			t.Fatalf("petition file: %v\n%s", err, out)
		}
		var filed struct {
			Result struct {
				MotionID string `json:"motion_id"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(out), &filed); err != nil || filed.Result.MotionID == "" {
			t.Fatalf("the filing names no motion id: %v\n%s", err, out)
		}
		return filed.Result.MotionID
	}
	readies := func(p record.Plan, seat string) bool {
		return slices.ContainsFunc(p.Parties, func(x record.Party) bool { return x.SeatID == seat })
	}

	for _, halts := range []bool{false, true} {
		runDir := newRun(t)
		stageEvents(t, runDir)
		if before := plan(runDir); !readies(before, "blue-respond") {
			t.Fatalf("precondition: the staged board readies blue on its open gap: %+v", before.Parties)
		}
		id := file(runDir)
		act(runDir, "register", "--seat-id", "red-chair")
		heard := plan(runDir)
		if len(heard.Parties) != 1 || heard.Parties[0].SeatID != "judge" || !slices.Equal(heard.Parties[0].Occasions, []string{"petition"}) {
			t.Fatalf("the plan convening the bench for %s readies %+v, want the bench alone, convened for the petition", id, heard.Parties)
		}
		if !slices.ContainsFunc(heard.Why, func(w string) bool { return strings.Contains(w, "the bench hears the petition alone this epoch") }) {
			t.Errorf("the plan does not say why the parties the board readies wait: %q", heard.Why)
		}
		if heard.Halted {
			t.Fatal("a plan with no halt on the record says halted")
		}
		act(runDir, "register", "--seat-id", "judge", "--occasion", "petition")
		act(runDir, "motion", "petition", "rule", "--seat-id", "judge", "--id", id, "--as", "denied", "--reason", "the procedure is described, not instructed")
		if halts {
			act(runDir, "halt", "--seat-id", "judge", "--reason", "continuing would itself be the harm")
		}
		act(runDir, "register", "--seat-id", "red-chair")
		after := plan(runDir)
		if after.Halted != halts {
			t.Errorf("halted = %v after a petition sitting that halted = %v", after.Halted, halts)
		}
		if halts && (len(after.Parties) != 0 || !slices.Contains(after.Why, "a halt is on the record: the bench ended the run, and nobody is dispatched")) {
			t.Errorf("a halted record readies %+v, why %q", after.Parties, after.Why)
		}
		if !halts && !readies(after, "blue-respond") {
			t.Errorf("after a denial with no halt the parties the board readies do not sit: %+v", after.Parties)
		}
	}
}
