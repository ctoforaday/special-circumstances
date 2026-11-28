package cli

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/capture"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// impasseDocketed is boardGap at impasse under k-max 2 and the docketing the chair's plan writes there.
func impasseDocketed(t *testing.T) (runDir, docketing string) {
	t.Helper()
	runDir = newRun(t)
	impasseBoard(t, runDir, 2, 2)
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	d := lastDocketing(t, runDir, boardGap)
	if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
		t.Fatalf("the plan at impasse wrote %+v, want an impasse docketing", d)
	}
	return runDir, d.GetDocketingId()
}

// proposalInWindow is the chair's proposal of a close of boardGap after one exchange, standing in its
// assignee's window: the chair's next plan readies the assignee for it.
func proposalInWindow(t *testing.T) (runDir, motion string) {
	t.Helper()
	runDir = newRun(t)
	impasseBoard(t, runDir, 2, 1)
	motion = proposeClose(t, runDir, blueSeat, boardGap, "not_a_defect", "not a defect, and the lens will not close it")
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	return runDir, motion
}

// proposalDocketed is proposalInWindow with the window over: the plan's PROPOSAL docketing holds it.
func proposalDocketed(t *testing.T) (runDir, motion, docketing string) {
	t.Helper()
	runDir, motion = proposalInWindow(t)
	return runDir, motion, docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap).GetDocketingId()
}

// lensCloses is the assignee's close of boardGap.
func lensCloses() []string {
	return []string{"close", "--id", boardGap, "--as", "not_a_defect", "--verified-by", "red-lens-evidence", "--verified-with", "the report read",
		"--verified-against", "the report's sentence", "--reason", "it is not a defect"}
}

// T-B23: EVERY DOCKETING DUTY IS KEYED ON THE DOCKETING (§III.B13's duty table and the M16 readers it
// names), driven through the command tree and each read through the verb or view that reports it, on
// a docketing of cause impasse and of cause proposal. Fails if a duty reads a motion where it should
// read the docketing, or the reverse. A ruling is appended through the record's write where the
// bench's side flags (S3) are not mounted.
func TestEveryDocketingDutyIsKeyedOnTheDocketing(t *testing.T) {
	const lens = "red-lens-evidence"

	t.Run("closings owed: awaited, filed, missing, not owed", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		owed := func() []record.PlanBlocker {
			var out []record.PlanBlocker
			for _, b := range planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")).Blockers {
				if b.Kind == record.BlockerClosingOwed {
					out = append(out, b)
				}
			}
			return out
		}
		if got := owed(); len(got) != 2 {
			t.Fatalf("closings awaited on docketing %s = %+v, want the lens's and blue's", d, got)
		}
		asSeat(t, runDir, "blue-respond", "register")
		asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--reason", "blue's position")
		asSeat(t, runDir, lens, "register")
		asSeat(t, runDir, lens, "register") // sat since the docketing, and filed none
		asSeat(t, runDir, "red-chair", "register")
		if got := owed(); len(got) != 1 || got[0].Owner != lens {
			t.Errorf("with blue's closing filed, closings owed = %+v, want the lens's alone", got)
		}
		a := capture.RecordParityAudit(runtest.Open(t, runDir))
		if a.Verdict != "FAIL" || !strings.Contains(a.Detail, lens+"'s closing on "+boardGap+" (docketing "+d+") was not filed") {
			t.Errorf("record parity = %s %s, want the lens's closing on docketing %s missing", a.Verdict, a.Detail, d)
		}
		if strings.Contains(a.Detail, otherLens) {
			t.Errorf("a lens that is no party is held to a closing: %s", a.Detail)
		}
	})

	t.Run("closings owed: a proposal's, filed in its window, answer its docketing", func(t *testing.T) {
		runDir, m := proposalInWindow(t)
		for _, seat := range []string{"blue-respond", lens} {
			asSeat(t, runDir, seat, "register")
			asSeat(t, runDir, seat, "closing", "--id", boardGap, "--reason", "the position of "+seat)
		}
		asSeat(t, runDir, lens, "register") // the window is over
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if d := lastDocketing(t, runDir, boardGap); d.GetMotionId() != m {
			t.Fatalf("the docketing = %+v, want PROPOSAL holding %s", d, m)
		}
		for _, b := range plan.Blockers {
			if b.Kind == record.BlockerClosingOwed {
				t.Errorf("a closing filed in the window is awaited again on its docketing: %+v", b)
			}
		}
		if !slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "judge" }) {
			t.Errorf("plan = %+v, want the bench readied: no closing is awaited", plan)
		}
	})

	t.Run("a closing's admission", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 3, 1)
		asSeat(t, runDir, "blue-respond", "register")
		actRefusedOnRecord(t, runDir, "a closing on a gap nothing puts before the bench", "blue-respond", []string{"is not before the bench"},
			"closing", "--id", boardGap, "--reason", "early")
		for name, setup := range map[string]func(*testing.T) string{
			"docketing": func(t *testing.T) string { r, _ := impasseDocketed(t); return r },
			"proposal":  func(t *testing.T) string { r, _ := proposalInWindow(t); return r },
		} {
			r := setup(t)
			asSeat(t, r, "blue-respond", "register")
			if out, err := runAt(t, "closing", "--id", boardGap, "--reason", "blue's position", "--run", r, "--seat-id", "blue-respond"); err != nil {
				t.Errorf("a closing while a %s stands was refused: %v %s", name, err, out)
			}
		}
	})

	t.Run("no closing in the filer's sitting before its proposal (requireNoClosingThisSitting)", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		// ON A DOCKETED GAP the docketing's closings are owed already and a proposal adds none: the check
		// is keyed on a proposal on an undocketed gap only, so a closing then a proposal is refused by
		// L38, naming the docketing — not as a second closing the sitting could not file.
		asSeat(t, runDir, "blue-respond", "register")
		asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--reason", "blue's position")
		actRefusedNaming(t, runDir, "blue's proposal after its closing on a docketed gap", "blue-respond", []string{"docketing " + d},
			append([]string{"motion", "docket", "file", "close", "--id", boardGap, "--reason", "late"}, closeFlags("not_a_defect")...)...)
	})

	t.Run("docketHeld: the bench rules the standing docketing it is convened on, naming its proposal", func(t *testing.T) {
		runDir, m, d := proposalDocketed(t)
		conveneBench(t, runDir, boardGap)
		k := lastDocketing(t, runDir, boardGap)
		wrong := benchRuling(k, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)
		wrong.MotionId = nil
		if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, wrong); err == nil || !strings.Contains(err.Error(), m) {
			t.Errorf("a ruling of %s naming no motion = %v, want refused naming its proposal %s", d, err, m)
		}
		if err := benchRules(t, runDir, boardGap, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
			t.Errorf("the bench's ruling of %s, held in its sitting: %v", d, err)
		}
	})

	t.Run("benchSatFor: one bench sitting per docketing", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		for _, seat := range []string{"blue-respond", lens} {
			asSeat(t, runDir, seat, "register")
			asSeat(t, runDir, seat, "closing", "--id", boardGap, "--reason", "the position of "+seat)
		}
		asSeat(t, runDir, "red-chair", "register")
		if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); !slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "judge" }) {
			t.Fatalf("plan = %+v, want the bench readied on docketing %s", plan, d)
		}
		asSeat(t, runDir, "judge", "register", "--occasion", "docket") // sits, and rules nothing
		asSeat(t, runDir, "red-chair", "register")
		if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "judge" }) {
			t.Errorf("plan = %+v: the bench is readied again on docketing %s it has sat for", plan, d)
		}
	})

	t.Run("the PASS gate: docketed, owned by the bench; a proposal in its window, by its assignee", func(t *testing.T) {
		runDir, _ := impasseDocketed(t)
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if !slices.Contains(plan.Blockers, record.PlanBlocker{Kind: record.BlockerDocketed, Subject: boardGap, Owner: "judge"}) {
			t.Errorf("blockers = %+v, want the docketing on %s, the bench's", plan.Blockers, boardGap)
		}
		r2, m := proposalInWindow(t)
		plan = planOf(t, asSeat(t, r2, "red-chair", "dispatch", "next", "--json"))
		if !slices.Contains(plan.Blockers, record.PlanBlocker{Kind: record.BlockerUnruledMotion, Subject: m, Owner: lens}) {
			t.Errorf("blockers = %+v, want proposal %s in its window, its assignee's", plan.Blockers, m)
		}
		for _, b := range plan.Blockers {
			if b.Kind == record.BlockerDocketed {
				t.Errorf("a proposal in its window is a docketed blocker: %+v", b)
			}
		}
	})

	t.Run("requireBenchMotionsRuled at the assembly: a standing docketing holds the outcome", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
		if out, err := runAt(t, "outcome", "--as", "UNVERIFIED", "--reason", "stopped", "--run", runDir, "--seat-id", "judge"); err == nil {
			t.Fatalf("the outcome landed over docketing %s, standing unruled: %s", d, out)
		} else if !strings.Contains(err.Error()+out, "docketing "+d) {
			t.Errorf("the outcome's refusal does not name docketing %s: %v", d, err)
		}
		if err := benchRules(t, runDir, boardGap, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
			t.Fatalf("the assembly's ruling of %s: %v", d, err)
		}
		asSeat(t, runDir, "judge", "outcome", "--as", "UNVERIFIED", "--reason", "stopped, with the docketing ruled")
	})

	t.Run("the raiser's close of a docketed gap concedes; under a proposal in its window it is refused", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		asSeat(t, runDir, lens, "register")
		if out, err := runAt(t, append(lensCloses(), "--run", runDir, "--seat-id", lens)...); err != nil {
			t.Fatalf("the raiser's close of %s's gap, its concession: %v %s", d, err, out)
		}
		if _, sd := standingCols(t, runDir, boardGap); sd != "" {
			t.Errorf("gap_state reads docketing %q standing after the raiser's concession of %s", sd, d)
		}
		r2, m := proposalInWindow(t)
		asSeat(t, r2, lens, "register")
		actRefusedOnRecord(t, r2, "the assignee's close under a proposal", lens, []string{"proposal " + m}, lensCloses()...)
	})

	t.Run("a proposal's filing is refused while a docketing stands, naming it", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		actRefusedNaming(t, runDir, "the author's proposal on a docketed gap", "blue-respond", []string{"docketing " + d},
			append([]string{"motion", "docket", "file", "close", "--id", boardGap, "--reason", "a second ask"}, closeFlags("not_a_defect")...)...)
	})

	t.Run("the remand: the docketing's gap", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		conveneBench(t, runDir, boardGap)
		k := lastDocketing(t, runDir, boardGap)
		remand := &recordpb.MotionRule{Opinion: proto.String("one more exchange"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			DocketingId: proto.String(k.GetDocketingId()), Disposition: recordpb.Disposition_DISPOSITION_REMANDED.Enum(), Principle: proto.String("p"),
			Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("nothing yet"), Direction: proto.String("a reproduction on the shipped binary")}}}
		if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, remand); err != nil {
			t.Fatalf("the remand of %s: %v", d, err)
		}
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if !slices.ContainsFunc(plan.RemandOwed, func(r record.RemandOwed) bool { return r.GapID == boardGap }) {
			t.Errorf("remand_owed = %+v, want %s, the remanded docketing's gap", plan.RemandOwed, boardGap)
		}
	})

	t.Run("the grade movement window", func(t *testing.T) {
		t.Skip("not driven here: needs a board past the grade-movement threshold to write a GRADE_MOVEMENT docketing and its --keep/--restore answer (S3)")
	})

	t.Run("the gap and word of a docket ruling (DocketRulingGap): a ruling of a docketing with no proposal closes its gap", func(t *testing.T) {
		runDir, _ := impasseDocketed(t)
		conveneBench(t, runDir, boardGap)
		if err := benchRules(t, runDir, boardGap, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
			t.Fatal(err)
		}
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		for _, b := range plan.Blockers {
			if b.Subject == boardGap {
				t.Errorf("after the bench's closing ruling of its docketing, %s still holds the PASS: %+v", boardGap, b)
			}
		}
	})

	t.Run("the dispatch's docketing read: docketedThisSitting", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if !slices.Equal(plan.Docket, []string{boardGap}) {
			t.Errorf("asked again in the sitting that wrote docketing %s, the plan's docket = %v", d, plan.Docket)
		}
		if ds, _ := docketingsOn(t, runDir, boardGap); len(ds) != 1 {
			t.Errorf("asked again, the plan wrote another docketing: %+v", ds)
		}
	})

	t.Run("verdict.go: the chair's PASS is refused naming the standing docketing", func(t *testing.T) {
		runDir, d := impasseDocketed(t)
		out, err := runAt(t, "verdict", "--run", runDir, "--seat-id", "red-chair")
		if err == nil {
			t.Fatalf("a PASS landed over docketing %s: %s", d, out)
		}
		if !strings.Contains(err.Error()+out, d) && !strings.Contains(err.Error()+out, "docketed") {
			t.Errorf("the PASS's refusal names neither docketing %s nor a docketed blocker: %v", d, err)
		}
	})

	t.Run("recordread.go and positionheard.go's unruled_docket_filed", func(t *testing.T) {
		t.Skip("not driven here: the record-read's standing-docketing read and the position hearing's standing proposal need a terminal record read and a chair position over a standing proposal; T-B23's remaining M16 rows")
	})
}
