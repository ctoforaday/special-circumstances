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

// stageRetiredLens seeds a run whose one lens has sat three times at the head, with the
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
	served := func() *record.Event {
		return at("red-lens-evidence", &recordpb.ReportRead{Head: proto.Int64(2)})
	}
	evs := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.ReportPart{Text: proto.String("# report")}),
		at("red-chair", &recordpb.Register{}), lens(), at("red-lens-evidence", &recordpb.Register{}), served(),
	}
	if withGap {
		evs = append(evs, at("red-lens-evidence", &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(limitGap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}))
	}
	for range 2 { // it has been served the report at the head, so the head readies it for nothing further
		evs = append(evs, at("red-chair", &recordpb.Register{}), lens(), at("red-lens-evidence", &recordpb.Register{}), served())
	}
	evs = append(evs, at("red-chair", &recordpb.Register{}))
	recordtest.Seed(t, runDir, evs...)
}

// argue files both parties' closings on limitGap — blue's and its minting lens's — so the plan
// convenes the bench for its docketing: no party's closing is then awaited.
func argue(t *testing.T, runDir string) {
	t.Helper()
	for _, seat := range []string{"blue-respond", "red-lens-evidence"} {
		asSeat(t, runDir, seat, "closing", "--id", limitGap, "--reason", "the case of "+seat)
	}
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
		{"a petition the bench denies", false, filePetition, denyPetition},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, tc.withGap)
			setEpochLimit(t, runDir, 5)
			motionID := tc.file(t, runDir)             // in epoch 4, before the final sitting
			asSeat(t, runDir, "red-chair", "register") // epoch 5: the final sitting
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
			if verdict != "CEILING" || basis != record.VerdictDerived || !strings.Contains(why, "epoch limit 5 reached") {
				t.Errorf("the outcome on the record = (%q, %q, %q), want CEILING derived from the epoch limit", verdict, basis, why)
			}
		})
	}
}

// A MOTION FILED AT THE RUN'S FINAL SITTING IS REFUSED, AND NOTHING IS RECORDED, for each subject no
// seat hears in time to change how the run ends. No seat is dispatched after the sitting that opens
// the last epoch, so the chair rules no grade motion, and a docket motion the terminal bench rules
// ends the run CEILING however it is ruled. A petition is the exception and has its own test below. The plan says the sitting is
// final whoever the board readies (record's TestAPassPermittedAtTheLastEpochIsNotTheEpochLimit holds
// it where a PASS is permitted, which was m25's last sitting).
func TestAMotionFiledAtTheFinalSittingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name, filer string
		args        []string
	}{
		{"a docket motion", "red-chair", []string{"motion", "docket", "file", "--id", limitGap, "--reason", "the bench's holding reads this gap as open"}},
		{"a grade motion", "blue-respond", []string{"motion", "grade", "file", "--id", limitGap, "--dimension", "severity", "--proposed", "low", "--reason", "the defect changes no reader decision"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, true)
			setEpochLimit(t, runDir, 4)
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			if !plan.FinalSitting || plan.Epoch != 4 {
				t.Fatalf("the plan at epoch 4 of 4 = %+v, want epoch 4 and the final sitting", plan)
			}
			before := len(eventsOfType(t, runDir, recordpb.EventType_EVENT_TYPE_MOTION))
			out, err := runAt(t, append(tc.args, "--run", runDir, "--seat-id", tc.filer)...)
			if err == nil {
				t.Fatalf("%s filed at the final sitting was recorded: %s", tc.name, out)
			}
			for _, want := range []string{"is not heard in time", "epoch 4, the last under the run's epoch limit of 4", "Nothing was recorded", "records the verdict the board permits", "or a FAIL", "goes in a position", "a petition, which the bench hears at the exit"} {
				if !strings.Contains(out+err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v\n%s", want, err, out)
				}
			}
			if after := len(eventsOfType(t, runDir, recordpb.EventType_EVENT_TYPE_MOTION)); after != before {
				t.Errorf("the refused filing left %d motion(s) on the record, want %d", after, before)
			}

			// One epoch earlier the same filing is heard: the sitting is not final.
			runDir = newRun(t)
			stageRetiredLens(t, runDir, true)
			setEpochLimit(t, runDir, 5)
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); plan.FinalSitting || plan.Epoch != 4 {
				t.Fatalf("the plan at epoch 4 of 5 = %+v, want epoch 4 and not the final sitting", plan)
			}
			asSeat(t, runDir, tc.filer, tc.args...)
		})
	}
}

// A PETITION FILED AT THE RUN'S FINAL SITTING IS RECORDED AND HEARD AT THE EXIT. It is the channel for
// an objection that the run must not go on, and the terminal bench hears a petition standing there and
// may halt the run on it: refusing it would leave the last chair sitting no way to put such an
// objection before anyone. Until it is ruled it holds the PASS; denied, the run ends CEILING for its
// term; with a halt, HALTED.
func TestAPetitionFiledAtTheFinalSittingIsHeardAtTheExit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		ruling func(t *testing.T, runDir, motionID string)
		want   string
	}{
		{"denied", denyPetition, "CEILING"},
		{"denied, and the bench halts", func(t *testing.T, runDir, motionID string) {
			denyPetition(t, runDir, motionID)
			asSeat(t, runDir, "judge", "halt", "--reason", "continuing would compromise the consent gate")
		}, "HALTED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, false)
			setEpochLimit(t, runDir, 4)
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); !plan.FinalSitting || !plan.PassPermitted {
				t.Fatalf("the plan at epoch 4 of 4 = %+v, want the final sitting with a PASS permitted", plan)
			}
			out := asSeat(t, runDir, "red-chair", "motion", "petition", "file", "--class", "integrity",
				"--relief", "stop the run", "--reason", "the report rests on a source fetched without consent", "--json")
			var env struct {
				Result struct {
					MotionID string `json:"motion_id"`
				} `json:"result"`
			}
			if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.MotionID == "" {
				t.Fatalf("the petition filed at the final sitting printed no motion id (%v): %s", err, out)
			}
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			benchOwned := false
			for _, b := range plan.Blockers {
				benchOwned = benchOwned || (b.Kind == record.BlockerUnruledMotion && b.Subject == env.Result.MotionID && b.Owner == "judge")
			}
			if plan.PassPermitted || len(plan.Parties) != 0 || !benchOwned {
				t.Fatalf("the plan after the petition = %+v, want the PASS held, nobody dispatched, and the petition the bench's to rule at the exit", plan)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			tc.ruling(t, runDir, env.Result.MotionID)
			asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
			asSeat(t, runDir, "judge", "outcome", "--as", tc.want, "--reason", "the petition was heard at the exit")
			if verdict, basis, why := outcomeOf(t, runDir); verdict != tc.want || basis != record.VerdictDerived {
				t.Errorf("the outcome on the record = (%q, %q, %q), want %s derived from the record", verdict, basis, why, tc.want)
			}
		})
	}
}

// eventsOfType is the record's events of one type.
func eventsOfType(t *testing.T, runDir string, typ recordpb.EventType) []*record.Event {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), typ)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

// filePetition files blue's petition and returns its motion id.
func filePetition(t *testing.T, runDir string) string {
	t.Helper()
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
}

// denyPetition is the bench's ruling of a petition, in the seat's sitting already open.
func denyPetition(t *testing.T, runDir, motionID string) {
	t.Helper()
	asSeat(t, runDir, "judge", "motion", "petition", "rule", "--id", motionID, "--as", "denied", "--reason", "the source is in the cache")
}

// A MOTION THE BENCH HAD ALREADY SAT FOR READIED NOBODY AT THE LAST SITTING, so the run was not at
// its epoch limit there, and ruling it at the exit does not put it there: the record derives no
// terminal state, CEILING is refused, and UNVERIFIED is the word the bench may assert. A docket and
// a petition each have their own sitting; and a motion filed into the bench's open sitting and
// ruled there was ruled before the chair's last sitting, so it is no ruling of the exit either.
func TestAMotionTheBenchLeftUnruledIsNotTheEpochLimit(t *testing.T) {
	for _, tc := range []struct {
		name     string
		withGap  bool
		occasion string
		file     func(t *testing.T, runDir string) string
		sitting  func(t *testing.T, runDir string) // what the bench does in the sitting it leaves the motion in
		rule     func(t *testing.T, runDir, motionID string)
	}{
		{"a docket the bench sat for and ruled nothing", true, "docket",
			func(t *testing.T, runDir string) string {
				return docketFile(t, runDir, "blue-respond", limitGap, "not converging, and the bench's to settle")
			},
			func(*testing.T, string) {},
			func(t *testing.T, runDir, motionID string) { ruleDocket(t, runDir, motionID, "repaired") }},
		{"a docket left beside one filed into the sitting and ruled there", true, "docket",
			func(t *testing.T, runDir string) string {
				return docketFile(t, runDir, "blue-respond", limitGap, "not converging, and the bench's to settle")
			},
			func(t *testing.T, runDir string) {
				// Filed by a seat that filed no closing on the gap this sitting: both parties did
				// (argue), and a party's docket motion over its own closing is refused.
				ruleDocket(t, runDir, docketFile(t, runDir, "red-chair", limitGap, "filed with the bench already sitting"), "repaired")
			},
			func(t *testing.T, runDir, motionID string) { ruleDocket(t, runDir, motionID, "repaired") }},
		{"a petition the bench heard and left", false, "petition", filePetition, func(*testing.T, string) {}, denyPetition},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, tc.withGap)
			setEpochLimit(t, runDir, 5)
			motionID := tc.file(t, runDir)
			if tc.withGap {
				argue(t, runDir)
			}
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); len(plan.Parties) != 1 || plan.Parties[0].SeatID != "judge" {
				t.Fatalf("the plan at epoch 4 of 5 = %+v, want the bench alone dispatched", plan)
			}
			asSeat(t, runDir, "judge", "register", "--occasion", tc.occasion) // the bench sits, and leaves the motion
			tc.sitting(t, runDir)
			asSeat(t, runDir, "red-chair", "register")
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			if plan.EpochLimitReached || plan.Ceiling || len(plan.Parties) != 0 {
				t.Fatalf("the chair's last plan = %+v, want nobody ready and no epoch-limit ceiling: the bench has sat for the motion", plan)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			tc.rule(t, runDir, motionID)
			asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
			if out, err := runAt(t, "outcome", "--as", "CEILING", "--reason", "the term ran out", "--run", runDir, "--seat-id", "judge"); err == nil {
				t.Fatalf("CEILING was recorded over a run whose last sitting readied nobody: %s", out)
			}
			asSeat(t, runDir, "judge", "outcome", "--as", "UNVERIFIED", "--reason", "nobody was ready and neither PASS nor the ceiling held")
			if verdict, basis, _ := outcomeOf(t, runDir); verdict != "UNVERIFIED" || basis != record.VerdictAsserted {
				t.Errorf("the outcome on the record = (%q, %q), want UNVERIFIED asserted", verdict, basis)
			}
		})
	}
}

// A BENCH RULING AT THE EXIT IS THE EPOCH LIMIT'S ONLY IN THE RUN'S LAST EPOCH. The chair's plan
// readies the bench and dispatches it, the engine stops the debate before the bench sits — a plan
// that stopped moving, a seat that failed — and the terminal bench rules the motion. The run's term
// did not run out, so the record derives no terminal state: CEILING is refused and UNVERIFIED is
// the word the bench may assert. One run is short of its limit and one is held to none.
func TestABenchRulingAtAnExitBeforeTheLastEpochIsNotTheEpochLimit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int
	}{{"one sitting short of the limit", 5}, {"a run held to no limit", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, true)
			if tc.limit > 0 {
				setEpochLimit(t, runDir, tc.limit)
			}
			motionID := docketFile(t, runDir, "blue-respond", limitGap, "not converging, and the bench's to settle")
			argue(t, runDir)
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); plan.EpochLimitReached || len(plan.Parties) != 1 || plan.Parties[0].SeatID != "judge" {
				t.Fatalf("the plan at the chair's fourth sitting = %+v, want the bench dispatched on the docket", plan)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			ruleDocket(t, runDir, motionID, "repaired")
			asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
			if out, err := runAt(t, "outcome", "--as", "CEILING", "--reason", "the term ran out", "--run", runDir, "--seat-id", "judge"); err == nil {
				t.Fatalf("CEILING was recorded over a run that stopped before its last epoch: %s", out)
			}
			asSeat(t, runDir, "judge", "outcome", "--as", "UNVERIFIED", "--reason", "the debate stopped with the bench dispatched and unsat")
			if verdict, basis, _ := outcomeOf(t, runDir); verdict != "UNVERIFIED" || basis != record.VerdictAsserted {
				t.Errorf("the outcome on the record = (%q, %q), want UNVERIFIED asserted", verdict, basis)
			}
		})
	}
}
