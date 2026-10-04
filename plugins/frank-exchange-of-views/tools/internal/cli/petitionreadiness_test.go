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

// A PETITION IS A PARTY'S RIGHT, AND THE BENCH IS NOT A PARTY (#1191). The bench rules every
// petition and halts the run itself where it must; `file` is off its surface, and a bench that
// reaches for it is told why rather than that the verb does not exist.
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
	if err == nil || !strings.Contains(out+err.Error(), "the bench is not a party") {
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
	rows, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_DISPATCH)
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
