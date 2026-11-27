package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
// material, minted in the lens's first sitting, and given its one exchange there — blue engaged on it
// sits twice and answers nothing — so a proposal on it is admitted.
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
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
			at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{limitGap}}),
			at("blue-respond", &recordpb.Register{}), at("blue-respond", &recordpb.Register{}))
	}
	for range 2 { // it has been served the report at the head, so the head readies it for nothing further
		evs = append(evs, at("red-chair", &recordpb.Register{}), lens(), at("red-lens-evidence", &recordpb.Register{}), served())
	}
	// The question is split, covered and checked: what these runs test is the epoch limit.
	evs = append(evs, recordtest.SettledQuestion(t)...)
	evs = append(evs, at("red-chair", &recordpb.Register{}))
	recordtest.Seed(t, runDir, evs...)
}

// argue files both parties' closings on limitGap — blue's and its assignee's — so the plan
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

// passEpochsTo opens, from epoch 4, each epoch before limit's final sitting: the chair registers and
// plans, and the plan readies nobody.
func passEpochsTo(t *testing.T, runDir string, limit int) {
	t.Helper()
	for e := 5; e < limit; e++ {
		asSeat(t, runDir, "red-chair", "register")
		if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); len(plan.Parties) != 0 {
			t.Fatalf("the plan at epoch %d of %d = %+v, want nobody ready", e, limit, plan)
		}
	}
}

// ruleDocket is the bench's ruling of a docketing with answer (a flag, without its dashes), in the
// seat's sitting already open.
func ruleDocket(t *testing.T, runDir, docketing, answer string) {
	t.Helper()
	asSeat(t, runDir, "judge", benchRuleArgs(docketing, answer, "correctness")...)
}

// ruleProposalDocketing is the bench's ruling, for the author's side, of the PROPOSAL docketing holding
// the close motionID on limitGap, in the bench's sitting already open: the close lands not_a_defect.
func ruleProposalDocketing(t *testing.T, runDir, motionID string) {
	t.Helper()
	if d := lastDocketing(t, runDir, limitGap); d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL || d.GetMotionId() != motionID {
		t.Fatalf("the docketing on %s = %+v, want PROPOSAL holding %s", limitGap, d, motionID)
	}
	if err := benchRules(t, runDir, limitGap, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
		t.Fatalf("the bench's ruling for the author: %v", err)
	}
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
// chair's last sitting leaves only what the terminal bench hears — a proposal standing in its
// assignee's window, whose parties sit for their closings first, or one petition — so the engine ends
// the debate CEILING, convenes the terminal bench, and tells the assembly to record CEILING. The bench's
// terminal register dockets the proposal (L29) and its ruling leaves nobody ready; the outcome is
// still the limit's.
func TestTheEpochLimitOutcomeSurvivesTheBenchRulingAtTheExit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		withGap bool
		file    func(t *testing.T, runDir string) string
		// closers are the parties the final plan keeps, for their closings; owner the blocker's.
		closers []string
		owner   string
		rule    func(t *testing.T, runDir, motionID string)
		// limit is the run's epoch limit. Blue is held out of the last epoch that dispatches a lens and
		// files nothing there, so its petition is filed at epoch 4 under a limit of 6, and the epochs
		// between pass before the final sitting.
		limit int
	}{
		{"a proposal the bench disposes of at the exit", true,
			func(t *testing.T, runDir string) string {
				return proposeClose(t, runDir, "red-chair", limitGap, "not_a_defect", "not converging, and the bench's to settle")
			},
			[]string{"red-lens-evidence", "blue-respond"}, "red-lens-evidence",
			func(t *testing.T, runDir, motionID string) {
				ruleProposalDocketing(t, runDir, motionID) // the terminal register docketed it (L29)
			}, 5},
		{"a petition the bench denies", false, filePetition, nil, "judge", func(t *testing.T, runDir, motionID string) { denyPetition(t, runDir, motionID) }, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, tc.withGap)
			setEpochLimit(t, runDir, tc.limit)
			motionID := tc.file(t, runDir) // in epoch 4, before the final sitting
			passEpochsTo(t, runDir, tc.limit)
			asSeat(t, runDir, "red-chair", "register") // the final sitting
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			var seated []string
			for _, p := range plan.Parties {
				seated = append(seated, p.SeatID)
				if len(p.GapIDs) != 0 {
					t.Errorf("the final plan engages %s on %v; a closing party is engaged on no gap", p.SeatID, p.GapIDs)
				}
			}
			if !plan.EpochLimitReached || !plan.Ceiling || !slices.Equal(seated, tc.closers) {
				t.Fatalf("the chair's last plan = %+v, want the epoch limit reached with exactly the closing parties %v", plan, tc.closers)
			}
			held := false
			for _, b := range plan.Blockers {
				held = held || (b.Kind == record.BlockerUnruledMotion && b.Subject == motionID && b.Owner == tc.owner)
			}
			if !held {
				t.Fatalf("the chair's last plan lists no blocker for %s owned by %s, so no terminal sitting convenes: %+v", motionID, tc.owner, plan.Blockers)
			}
			if tc.withGap {
				argue(t, runDir)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			tc.rule(t, runDir, motionID)
			if after, err := record.PlanDispatch(runtest.Open(t, runDir)); err != nil || len(after.Parties) != 0 {
				t.Fatalf("after the ruling the plan readies %+v (%v); this test is about a ruling that leaves nobody ready", after.Parties, err)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
			asSeat(t, runDir, "judge", "outcome", "--as", "CEILING", "--reason", "the run's term ran out with the bench still to rule")
			verdict, basis, why := outcomeOf(t, runDir)
			if verdict != "CEILING" || basis != record.VerdictDerived || !strings.Contains(why, fmt.Sprintf("epoch limit %d reached", tc.limit)) {
				t.Errorf("the outcome on the record = (%q, %q, %q), want CEILING derived from the epoch limit", verdict, basis, why)
			}
		})
	}
}

// A MOTION FILED AT THE RUN'S FINAL SITTING IS REFUSED, AND NOTHING IS RECORDED, for each subject no
// seat hears in time to change how the run ends. No seat is dispatched after the sitting that opens
// the last epoch, so a docket motion the terminal bench rules
// ends the run CEILING however it is ruled. A petition is the exception and has its own test below. The plan says the sitting is
// final whoever the board readies (record's TestAPassPermittedAtTheLastEpochIsNotTheEpochLimit holds
// it where a PASS is permitted, which was m25's last sitting).
func TestAMotionFiledAtTheFinalSittingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name, filer string
		args        []string
	}{
		{"a docket motion", "red-chair", append([]string{"motion", "docket", "file", "close", "--id", limitGap, "--reason", "the bench's holding reads this gap as open"}, closeFlags("not_a_defect")...)},
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
			for _, want := range []string{"is not heard in time", "epoch 4, the last under the run's epoch limit of 4", "Nothing was recorded", "records the PASS where the plan permits one, or no verdict", "goes in a position", "a petition, which the bench hears at the exit"} {
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
		"--relief", "strike the lens's last gap", "--reason", "the gap quotes a source nobody fetched", "--json")
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
		limit    int // 6 for blue's petition: blue files nothing in the last epoch that dispatches a lens
	}{
		{"a docketing the bench sat for and ruled nothing", true, "docket",
			func(t *testing.T, runDir string) string {
				return proposeClose(t, runDir, "red-chair", limitGap, "not_a_defect", "not converging, and the bench's to settle")
			},
			func(*testing.T, string) {},
			func(t *testing.T, runDir, motionID string) { ruleProposalDocketing(t, runDir, motionID) }, 5},
		{"a docketing left while a second proposal on its gap is refused in the sitting", true, "docket",
			func(t *testing.T, runDir string) string {
				return proposeClose(t, runDir, "red-chair", limitGap, "not_a_defect", "not converging, and the bench's to settle")
			},
			func(t *testing.T, runDir string) {
				// A DOCKETED GAP TAKES NO PROPOSAL (L38), so nothing reaches the bench's open sitting
				// beside the docketing it leaves: the chair's proposal is refused naming it, and records
				// nothing.
				d := lastDocketing(t, runDir, limitGap).GetDocketingId()
				before := motionsOnRecord(t, runDir)
				if out, err := runAt(t, closeProposalArgs(runDir, "red-chair", limitGap, "not_a_defect", "filed with the bench already sitting")...); err == nil ||
					!strings.Contains(err.Error()+out, "docketing "+d+" stands on it") || !strings.Contains(err.Error()+out, "no proposal is filed on a docketed gap") {
					t.Fatalf("a second proposal on %s while its docketing stands = %v %q, want refused naming %s", limitGap, err, out, d)
				}
				if after := motionsOnRecord(t, runDir); after != before {
					t.Fatalf("the refused proposal left %d motion(s), want %d", after, before)
				}
			},
			func(t *testing.T, runDir, motionID string) { ruleProposalDocketing(t, runDir, motionID) }, 5},
		{"a petition the bench heard and left", false, "petition", filePetition, func(*testing.T, string) {}, denyPetition, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, tc.withGap)
			setEpochLimit(t, runDir, tc.limit)
			motionID := tc.file(t, runDir)
			if tc.withGap {
				argue(t, runDir)
				windowEnds(t, runDir, "red-lens-evidence")
			}
			// A docket is on the plan; a petition is owed by the record's read and on no plan.
			wantOnPlan := 0
			if tc.withGap {
				wantOnPlan = 1
			}
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); len(plan.Parties) != wantOnPlan {
				t.Fatalf("the plan at epoch 4 of %d = %+v, want %d parties", tc.limit, plan, wantOnPlan)
			}
			asSeat(t, runDir, "judge", "register", "--occasion", tc.occasion) // the bench sits, and leaves the motion
			tc.sitting(t, runDir)
			passEpochsTo(t, runDir, tc.limit)
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
			motionID := proposeClose(t, runDir, "red-chair", limitGap, "not_a_defect", "not converging, and the bench's to settle")
			argue(t, runDir)
			windowEnds(t, runDir, "red-lens-evidence")
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); plan.EpochLimitReached || len(plan.Parties) != 1 || plan.Parties[0].SeatID != "judge" {
				t.Fatalf("the plan at the chair's fourth sitting = %+v, want the bench dispatched on the docket", plan)
			}

			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			ruleProposalDocketing(t, runDir, motionID)
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

// A HEARING OWED AT THE FINAL SITTING IS THE EPOCH LIMIT. No seat is dispatched after that sitting, so
// the filing is heard by the bench at the exit, and the run ended because its term ran out: the plan
// names nobody and still says the limit was reached. The same sitting with nothing owed and nobody
// ready is not at the limit; a filing one epoch earlier is heard in its turn, and is no limit either.
func TestAHearingOwedAtTheFinalSittingReachesTheEpochLimit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int
		file  bool
		want  bool
	}{
		{"a petition at the final sitting", 5, true, true},
		{"nothing owed at the final sitting", 5, false, false},
		{"a petition before it", 6, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			stageRetiredLens(t, runDir, false)
			setEpochLimit(t, runDir, tc.limit)
			asSeat(t, runDir, "red-chair", "register")
			if tc.file {
				asSeat(t, runDir, "red-chair", "motion", "petition", "file", "--class", "integrity",
					"--relief", "strike the lens's last gap", "--reason", "the gap quotes a source nobody fetched")
			}
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			if len(plan.Parties) != 0 || plan.EpochLimitReached != tc.want {
				t.Fatalf("the plan at epoch %d of %d = %+v, want nobody on it and epoch_limit_reached %v", plan.Epoch, tc.limit, plan, tc.want)
			}
		})
	}
}
