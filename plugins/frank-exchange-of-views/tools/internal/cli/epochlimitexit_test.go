package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// limitGap is the one gap the epoch-limit exits below turn on.
const limitGap = "G-00000001"

// stageRetiredLens seeds a run whose one lens has sat three times and retired at the head, with the
// chair opening its fourth sitting: nobody is ready but what a test adds. withGap leaves limitGap open and
// material, minted in the lens's first sitting.
func stageRetiredLens(t *testing.T, runDir string, withGap bool) {
	t.Helper()
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:limit:%d", seat, n), body)
	}
	lens := func() *record.Event {
		return at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-evidence")})
	}
	evs := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# report")}),
		at("red-chair", &recordpb.Register{}), lens(), at("red-lens-evidence", &recordpb.Register{}),
	}
	if withGap {
		evs = append(evs, at("red-lens-evidence", &recordpb.Mint{GapId: proto.String(limitGap), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}))
	}
	for range 2 { // two barren sittings since it last minted retire the lens
		evs = append(evs, at("red-chair", &recordpb.Register{}), lens(), at("red-lens-evidence", &recordpb.Register{}))
	}
	evs = append(evs, at("red-chair", &recordpb.Register{}))
	recordtest.Seed(t, runDir, evs...)
}

// setEpochLimit writes the run's epoch limit where setup records it.
func setEpochLimit(t *testing.T, runDir string, limit int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs", "run-config.json"), fmt.Appendf(nil, `{"maxEpochs":%d}`, limit), 0o644); err != nil {
		t.Fatal(err)
	}
}

// asSeat runs a seat's command and fails the test on a refusal.
func asSeat(t *testing.T, runDir, seatID string, args ...string) string {
	t.Helper()
	out, err := runAt(t, append(args, "--run", runDir, "--seat-id", seatID)...)
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", seatID, args, err, out)
	}
	return out
}

// ruleDocket is the bench's ruling of a docket motion, in the seat's sitting already open.
func ruleDocket(t *testing.T, runDir, motionID, as string) {
	t.Helper()
	asSeat(t, runDir, "judge", benchRuleArgs(motionID, as, "correctness")...)
}

// outcomeOf is the verdict, basis and why the record's outcome event carries, "" where it holds none.
func outcomeOf(t *testing.T, runDir string) (verdict, basis, why string) {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_OUTCOME)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs {
		if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
			verdict, basis, why = strings.ToUpper(recordpb.Word(o.GetVerdict())), o.GetVerdictBasis(), o.GetVerdictWhy()
		}
	}
	return verdict, basis, why
}

// A RUN THAT ENDS AT ITS EPOCH LIMIT RECORDS CEILING, WHATEVER THE BENCH RULES AT THE EXIT. The
// chair's last sitting leaves only the bench ready — one docketed gap, or one petition — so the
// engine ends the debate CEILING, convenes the terminal bench for the motion, and tells the assembly
// to record CEILING. The ruling leaves nobody ready; the outcome is still the limit's.
func TestTheEpochLimitOutcomeSurvivesTheBenchRulingAtTheExit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		withGap bool
		file    func(t *testing.T, runDir string) string
		rule    func(t *testing.T, runDir, motionID string)
	}{
		{"a docketed gap the bench disposes of", true,
			func(t *testing.T, runDir string) string {
				return docketFile(t, runDir, "blue-respond", limitGap, "not converging, and the bench's to settle")
			},
			func(t *testing.T, runDir, motionID string) { ruleDocket(t, runDir, motionID, "repaired") }},
		{"a petition the bench denies", false,
			func(t *testing.T, runDir string) string {
				out := asSeat(t, runDir, "blue-respond", "motion", "petition", "file", "--class", "integrity",
					"--relief", "strike the lens's last finding", "--reason", "the finding quotes a source nobody fetched", "--json")
				var env struct {
					Result struct {
						MotionID string `json:"motion_id"`
					} `json:"result"`
				}
				if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.MotionID == "" {
					t.Fatalf("the petition filing printed no motion id (%v): %s", err, out)
				}
				return env.Result.MotionID
			},
			func(t *testing.T, runDir, motionID string) {
				asSeat(t, runDir, "judge", "motion", "petition", "rule", "--id", motionID, "--as", "denied", "--reason", "the source is in the cache")
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, tc.withGap)
			setEpochLimit(t, runDir, 4)
			motionID := tc.file(t, runDir)
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			if !plan.EpochLimitReached || !plan.Ceiling || len(plan.Parties) != 0 {
				t.Fatalf("the chair's last plan = %+v, want the epoch limit reached with the bench undispatched", plan)
			}
			benchOwned := false
			for _, b := range plan.Blockers {
				benchOwned = benchOwned || (b.Kind == record.BlockerUnruledMotion && b.Subject == motionID && b.Owner == "judge")
			}
			if !benchOwned {
				t.Fatalf("the chair's last plan lists no bench-owned blocker for %s, so no terminal sitting convenes: %+v", motionID, plan.Blockers)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			tc.rule(t, runDir, motionID)
			if after, err := record.PlanDispatch(runtest.Open(t, runDir)); err != nil || len(after.Parties) != 0 {
				t.Fatalf("after the ruling the plan readies %+v (%v); this test is about a ruling that leaves nobody ready", after.Parties, err)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
			asSeat(t, runDir, "judge", "outcome", "--as", "CEILING", "--reason", "the run's term ran out with the bench still to rule")
			verdict, basis, why := outcomeOf(t, runDir)
			if verdict != "CEILING" || basis != record.VerdictDerived || !strings.Contains(why, "epoch limit 4 reached") {
				t.Errorf("the outcome on the record = (%q, %q, %q), want CEILING derived from the epoch limit", verdict, basis, why)
			}
		})
	}
}

// A MOTION THE BENCH HAD ALREADY SAT FOR READIED NOBODY AT THE LAST SITTING, so the run was not at
// its epoch limit there, and ruling it at the exit does not put it there: the record derives no
// terminal state, CEILING is refused, and UNVERIFIED is the word the bench may assert.
func TestAMotionTheBenchLeftUnruledIsNotTheEpochLimit(t *testing.T) {
	runDir := newRun(t)
	stageRetiredLens(t, runDir, true)
	setEpochLimit(t, runDir, 5)
	motionID := docketFile(t, runDir, "blue-respond", limitGap, "not converging, and the bench's to settle")
	if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); len(plan.Parties) != 1 || plan.Parties[0].SeatID != "judge" {
		t.Fatalf("the plan at epoch 4 of 5 = %+v, want the bench dispatched on the docket", plan)
	}
	asSeat(t, runDir, "judge", "register", "--occasion", "docket") // the bench sits, and rules nothing
	asSeat(t, runDir, "red-chair", "register")
	plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
	if plan.EpochLimitReached || plan.Ceiling || len(plan.Parties) != 0 {
		t.Fatalf("the chair's last plan = %+v, want nobody ready and no epoch-limit ceiling: the bench has sat for the docket", plan)
	}

	asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
	ruleDocket(t, runDir, motionID, "repaired")
	asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
	if out, err := runAt(t, "outcome", "--as", "CEILING", "--reason", "the term ran out", "--run", runDir, "--seat-id", "judge"); err == nil {
		t.Fatalf("CEILING was recorded over a run whose last sitting readied nobody: %s", out)
	}
	asSeat(t, runDir, "judge", "outcome", "--as", "UNVERIFIED", "--reason", "nobody was ready and neither PASS nor the ceiling held")
	if verdict, basis, _ := outcomeOf(t, runDir); verdict != "UNVERIFIED" || basis != record.VerdictAsserted {
		t.Errorf("the outcome on the record = (%q, %q), want UNVERIFIED asserted", verdict, basis)
	}
}
