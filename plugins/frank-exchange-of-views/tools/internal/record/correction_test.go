package record

import (
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// Same-sitting correction at the record layer (plans/same-sitting-correction.md): who may correct
// what, when, and how much of it — and that a reader sees the corrected act struck, in its place.

func corrRun(t *testing.T) Run {
	t.Helper()
	return mustRun(t, newRun(t))
}

func sit(t *testing.T, run Run, seat string) Identity {
	t.Helper()
	id := Identity{Run: run, SeatID: seat}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	return id
}

func mustAppend(t *testing.T, id Identity, body proto.Message) *Event {
	t.Helper()
	ev, err := Append(id, body)
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

// correcting is id, re-run to correct the act keyed key. A fresh Correct per invocation, as the
// command layer builds one per process.
func correcting(id Identity, typ recordpb.EventType, key, why string) Identity {
	id.Correct = &Correct{Type: typ, Key: key, Why: why}
	return id
}

func seatLog(text string) *recordpb.Log {
	return &recordpb.Log{Text: proto.String(text), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
}

func toolLog(text string) *recordpb.Log {
	return &recordpb.Log{Text: proto.String(text), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}
}

func corrMint(id string) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(id), Class: proto.String("x"), Problem: proto.String("p " + id), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
}

func allEvents(t *testing.T, run Run) []*Event {
	t.Helper()
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	return m.Events
}

func mustRefuse(t *testing.T, err error, wants ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("accepted; want a refusal saying %q", wants)
	}
	for _, w := range wants {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
}

// THE CORRECTED ACT STAYS, STRUCK; THE REPLACEMENT TAKES ITS PLACE. Two events are appended in
// one write — the replacement, keyed on the chain's root, and the correction naming both.
func TestACorrectionAppendsAReplacementAndStrikesTheAct(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	bad := mustAppend(t, blue, &recordpb.Position{Text: proto.String("G2 is reproducible via ")})
	before := len(allEvents(t, run))

	rep, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_POSITION, bad.GetKey(), "the shell ate the method"),
		&recordpb.Position{Text: proto.String("G2 is reproducible via the recorded proof")})
	if err != nil {
		t.Fatal(err)
	}
	if want := bad.GetKey() + "~1"; rep.GetKey() != want {
		t.Errorf("replacement key %q, want %q (the root's key and the chain depth)", rep.GetKey(), want)
	}
	evs := allEvents(t, run)
	if got := len(evs) - before; got != 2 {
		t.Fatalf("a correction appended %d events, want 2 (the replacement and the correction)", got)
	}
	c, ok := recordpb.BodyAs[*recordpb.Correction](evs[len(evs)-1])
	if !ok {
		t.Fatalf("the last event is a %s, want the correction", recordpb.Word(evs[len(evs)-1].GetType()))
	}
	if c.GetCorrects() != bad.GetKey() || c.GetReplacement() != rep.GetKey() || c.GetWhy() != "the shell ate the method" {
		t.Errorf("correction = %v", c)
	}
	idx := StruckIndexOf(evs)
	if s, ok := idx.Of(bad.GetKey()); !ok || s.By != "blue-respond" || s.Replacement != rep.GetKey() {
		t.Errorf("the corrected act is not struck by its replacement: %+v %v", s, ok)
	}
	// The struck act is still on the record — append-only — and Live reads the replacement.
	var positions, live []string
	for _, e := range evs {
		if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
			positions = append(positions, p.GetText())
		}
	}
	for _, e := range Live(evs) {
		if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
			live = append(live, p.GetText())
		}
	}
	if len(positions) != 2 {
		t.Errorf("the record holds %d positions, want both the struck and its replacement", len(positions))
	}
	if len(live) != 1 || live[0] != "G2 is reproducible via the recorded proof" {
		t.Errorf("Live reads %q, want only the replacement", live)
	}
}

// A RETRY OF THE SAME CORRECTION WRITES NOTHING; A DIFFERENT ONE NAMES THE ACT THAT STANDS.
func TestACorrectionRetryIsIdempotentAndAChainKeysOnItsRoot(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	k := mustAppend(t, blue, seatLog("first, wrong")).GetKey()
	fix := func(key, text string) (*Event, error) {
		return Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, key, "typo"), seatLog(text))
	}
	r1, err := fix(k, "second, right")
	if err != nil {
		t.Fatal(err)
	}
	n := len(allEvents(t, run))
	again, err := fix(k, "second, right")
	if err != nil {
		t.Fatalf("the retried correction was refused: %v", err)
	}
	if again.GetKey() != r1.GetKey() || len(allEvents(t, run)) != n {
		t.Errorf("the retry wrote %d event(s) and answered %q; want 0 and %q", len(allEvents(t, run))-n, again.GetKey(), r1.GetKey())
	}
	_, err = fix(k, "a third wording")
	mustRefuse(t, err, "already corrected", r1.GetKey())

	r2, err := fix(r1.GetKey(), "a third wording")
	if err != nil {
		t.Fatal(err)
	}
	if want := k + "~2"; r2.GetKey() != want {
		t.Errorf("a correction of a correction is keyed %q, want %q", r2.GetKey(), want)
	}
	_, err = fix(k, "a fourth")
	mustRefuse(t, err, "name "+r2.GetKey())
}

// EVERY REFUSAL, with the reason a seat reads.
func TestACorrectionIsRefused(t *testing.T) {
	t.Run("another seat's act", func(t *testing.T) {
		run := corrRun(t)
		red := sit(t, run, "red-chair")
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("blue's")).GetKey()
		_, err := Append(correcting(red, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), seatLog("red's"))
		mustRefuse(t, err, "only the seat that wrote an act may correct it")
	})
	t.Run("an earlier sitting's act", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("first sitting")).GetKey()
		sit(t, run, "blue-respond")
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), seatLog("second sitting"))
		mustRefuse(t, err, "earlier sitting", "sitting 1; this is sitting 2")
	})
	t.Run("relied on: another seat acted since", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("read by red")).GetKey()
		sit(t, run, "red-chair")
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), seatLog("too late"))
		mustRefuse(t, err, "another seat has acted since this log", "a register by red-chair")
	})
	t.Run("a tool-written log", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, toolLog("the tool refused a write")).GetKey()
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), toolLog("restated"))
		mustRefuse(t, err, "cannot be corrected", "the tool wrote it")
	})
	t.Run("an act that creates an identity: mint", func(t *testing.T) {
		run := corrRun(t)
		red := sit(t, run, "red-chair")
		k := mustAppend(t, red, corrMint("G1")).GetKey()
		m := corrMint("G1")
		m.Problem = proto.String("reworded")
		_, err := Append(correcting(red, recordpb.EventType_EVENT_TYPE_MINT, k, "w"), m)
		mustRefuse(t, err, "is a mint, which cannot be corrected")
	})
	t.Run("an act that creates an identity: motion", func(t *testing.T) {
		run := corrRun(t)
		red := sit(t, run, "red-chair")
		mustAppend(t, red, corrMint("G1"))
		motion := func(basis string) *recordpb.Motion {
			return &recordpb.Motion{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
				Basis: proto.String(basis), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}
		}
		k := mustAppend(t, red, motion("red cannot settle G1")).GetKey()
		_, err := Append(correcting(red, recordpb.EventType_EVENT_TYPE_MOTION, k, "w"), motion("reworded"))
		mustRefuse(t, err, "is a motion, which cannot be corrected")
	})
	t.Run("a replacement equal to its target", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("same")).GetKey()
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), seatLog("same"))
		mustRefuse(t, err, "this correction changes nothing")
	})
	t.Run("a PROSE correction that moves a frozen field", func(t *testing.T) {
		run := corrRun(t)
		judge := sit(t, run, "judge")
		k := mustAppend(t, judge, &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("stopped because  refused")}).GetKey()
		_, err := Append(correcting(judge, recordpb.EventType_EVENT_TYPE_OUTCOME, k, "w"),
			&recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_VERIFIED.Enum(), Prose: proto.String("stopped because the gate refused")})
		mustRefuse(t, err, "may change only the seat's own wording", "must stay as event")
	})
	t.Run("a FULL correction onto another label", func(t *testing.T) {
		run := corrRun(t)
		red := sit(t, run, "red-chair")
		mustAppend(t, red, corrMint("G1"))
		mustAppend(t, red, corrMint("G2"))
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("checked")}).GetKey()
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_MANIFEST_ROW, k, "w"),
			&recordpb.ManifestRow{GapId: proto.String("G2"), Row: proto.String("checked")})
		mustRefuse(t, err, "keeps the act's label", `"G1"`, `"G2"`)
	})
	t.Run("another type's key", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("a log")).GetKey()
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_POSITION, k, "w"), &recordpb.Position{Text: proto.String("a position")})
		mustRefuse(t, err, "is a log, and this command writes a position")
	})
	t.Run("a key nothing carries", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, "blue-respond:log:#9", "w"), seatLog("x"))
		mustRefuse(t, err, "no act on this record carries that key")
	})
	t.Run("no why", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("a")).GetKey()
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "  "), seatLog("b"))
		mustRefuse(t, err, "--correction-why")
	})
	t.Run("a correction appended on its own", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("a")).GetKey()
		_, err := Append(blue, &recordpb.Correction{Corrects: proto.String(k), Replacement: proto.String(k), Why: proto.String("w")})
		mustRefuse(t, err, "never on its own")
	})
	t.Run("a second body of the corrected type in one invocation", func(t *testing.T) {
		run := corrRun(t)
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("a")).GetKey()
		id := correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w")
		mustAppend(t, id, seatLog("b"))
		_, err := Append(id, seatLog("c"))
		mustRefuse(t, err, "already wrote its correction")
	})
}

// THE RELIANCE RULE COUNTS SEATS ACTING, not the harness observing a span or the tool logging what
// it did. Both land after the target here and the correction is still admitted.
func TestRelianceIgnoresTheHarnessAndTheTool(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := sit(t, run, "red-chair") // registered BEFORE the target: that act is not reliance
	blue := sit(t, run, "blue-respond")
	k := mustAppend(t, blue, seatLog("a")).GetKey()
	recordtest.Seed(t, runDir, recordtest.At(t, HarnessSeat, "harness:sitting_open:x", &recordpb.SittingOpen{AgentId: proto.String("a1")}))
	mustAppend(t, red, toolLog("the tool logged a refusal"))
	if _, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), seatLog("b")); err != nil {
		t.Fatalf("a harness span and a tool log counted as another seat relying on the act: %v", err)
	}
}

// THE TIER TABLE IS THE OWNER'S RULING (plans/same-sitting-correction.md F5): FULL and PROSE are
// named; every other type is NONE — and a type added later is NONE until someone rules otherwise.
func TestTheCorrectionTiersAreTheOwnersRuling(t *testing.T) {
	full := map[string]bool{"manifest_row": true, "position": true, "closing": true, "revision": true, "log": true,
		"inquiry_review": true, "spot_check": true, "regrade": true}
	// cite and proof joined PROSE on gblock's 2026-09-11 ruling (#886): the title, the argument and
	// the proof note are report text a seat may need to re-word in the sitting that wrote it.
	prose := map[string]bool{"motion_rule": true, "motion_appeal": true, "declare": true, "certify": true, "close": true,
		"avenue": true, "reproduce": true, "outcome": true, "halt": true, "cite": true, "proof": true}
	ed := recordpb.EventType(0).Descriptor()
	for i := 0; i < ed.Values().Len(); i++ {
		typ := recordpb.EventType(ed.Values().Get(i).Number())
		w := recordpb.Word(typ)
		if w == "" {
			continue
		}
		want := recordpb.CorrectionTier_CORRECTION_TIER_NONE
		switch {
		case full[w]:
			want = recordpb.CorrectionTier_CORRECTION_TIER_FULL
		case prose[w]:
			want = recordpb.CorrectionTier_CORRECTION_TIER_PROSE
		}
		if got := recordpb.Tier(typ); got != want {
			t.Errorf("%s is tier %v, want %v", w, got, want)
		}
	}
}

// A CORRECTION TAKES ITS TARGET'S PLACE (F12): Live puts the replacement where the struck act was,
// so a last-wins fold cannot pick it over a later act by the same seat.
func TestLivePutsTheReplacementInTheStruckActsPlace(t *testing.T) {
	a := recordtest.At(t, "s", "s:log:#1", seatLog("a"))
	b := recordtest.At(t, "s", "s:log:#2", seatLog("b"))
	r := recordtest.At(t, "s", "s:log:#1~1", seatLog("a, corrected"))
	c := recordtest.At(t, "s", "s:correction:s:log:#1", &recordpb.Correction{
		Corrects: proto.String("s:log:#1"), Replacement: proto.String("s:log:#1~1"), Why: proto.String("w")})
	var got []string
	for _, e := range Live([]*Event{a, b, r, c}) {
		got = append(got, e.GetKey())
	}
	want := []string{"s:log:#1~1", "s:log:#2", "s:correction:s:log:#1"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("Live = %v, want %v", got, want)
	}
}

// THE SQL WINNERS READ THE SAME ORDER: a proposal corrected after the line was moved keeps its place,
// so line_of_inquiry reads the corrected wording AND the move's status — not the replacement's
// `proposed`, which MAX(event_id) would have picked.
func TestLineOfInquiryReadsACorrectedProposalInItsPlace(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	propose := func(line string) *recordpb.Avenue {
		return &recordpb.Avenue{AvenueId: proto.String("A1"), Line: proto.String(line), Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum()}
	}
	k := mustAppend(t, blue, propose("try the  method")).GetKey()
	mustAppend(t, blue, &recordpb.Avenue{AvenueId: proto.String("A1"), Status: recordpb.AvenueStatus_AVENUE_STATUS_PURSUED.Enum(), SupersedesStatus: proto.String("proposed")})
	if _, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_AVENUE, k, "the shell ate it"), propose("try the recorded method")); err != nil {
		t.Fatal(err)
	}
	var line, status string
	if _, err := queryRow(run, []any{&line, &status}, `SELECT "line", "status" FROM "line_of_inquiry" WHERE "avenue_id" = 'A1'`); err != nil {
		t.Fatal(err)
	}
	if line != "try the recorded method" || status != "pursued" {
		t.Errorf("line_of_inquiry = (%q, %q), want the corrected line and the move's status", line, status)
	}
	var n int
	if _, err := queryRow(run, []any{&n}, `SELECT count(*) FROM "line_of_inquiry"`); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("line_of_inquiry has %d rows, want 1 — a replacement proposal is not a second line", n)
	}
}

// LAYER 1: a guard that refuses because the thing already happened passes when what happened IS
// the act being corrected — a closure's correction names a closed gap by definition. The guard
// still refuses an ordinary second closure.
func TestAClosureCorrectionPassesTheClosedGapGuard(t *testing.T) {
	run := corrRun(t)
	red := sit(t, run, "red-chair")
	mustAppend(t, red, corrMint("G1"))
	closure := func(prose string) *recordpb.Close {
		return &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String(prose)}
	}
	k := mustAppend(t, red, closure("fixed via ")).GetKey()
	r1, err := Append(correcting(red, recordpb.EventType_EVENT_TYPE_CLOSE, k, "w"), closure("fixed via the patch"))
	if err != nil {
		t.Fatalf("a closure's correction was refused because the gap is closed — by the act being corrected: %v", err)
	}
	if _, err := Append(correcting(red, recordpb.EventType_EVENT_TYPE_CLOSE, r1.GetKey(), "w"), closure("fixed via the patch, re-run")); err != nil {
		t.Fatalf("correcting the replacement was refused: %v", err)
	}
	c := closure("fixed via the patch")
	c.AnchorTool = proto.String("another tool")
	_, err = Append(correcting(red, recordpb.EventType_EVENT_TYPE_CLOSE, k+"~2", "w"), c)
	mustRefuse(t, err, "must stay as event")
	_, err = Append(red, closure("a second closure"))
	if err == nil {
		t.Fatal("an ordinary second closure of a closed gap was accepted")
	}
}

// F7 IS DECIDED UNDER THE WRITE LOCK. Lenses write while blue corrects; whenever the correction
// commits, no other seat's act may sit between the target and it. Run it under -race.
//
// The lenses start after a delay that grows with the round, so the rounds straddle the moment the
// correction's transaction begins; the last round runs the lenses only AFTER the correction has
// returned, so at least one commit is always checked rather than the test passing on refusals alone.
func TestCorrectionConcurrentNeverCommitsAfterAForeignAct(t *testing.T) {
	committed, refused := 0, 0
	delays := []time.Duration{0, time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond,
		8 * time.Millisecond, 16 * time.Millisecond, 32 * time.Millisecond, -1}
	for round, delay := range delays {
		run := corrRun(t)
		lenses := []string{"red-lens-logic", "red-lens-evidence", "red-lens-voice"}
		for _, l := range lenses {
			sit(t, run, l)
		}
		blue := sit(t, run, "blue-respond")
		k := mustAppend(t, blue, seatLog("target")).GetKey()
		start, corrected := make(chan struct{}), make(chan struct{})
		var wg sync.WaitGroup
		for _, l := range lenses {
			wg.Add(1)
			go func(seat string) {
				defer wg.Done()
				<-start
				if delay < 0 {
					<-corrected
				} else {
					time.Sleep(delay)
				}
				for i := 0; i < 3; i++ {
					_, _ = Append(Identity{Run: run, SeatID: seat}, seatLog("lens act"))
				}
			}(l)
		}
		close(start)
		_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k, "w"), seatLog("corrected"))
		close(corrected)
		wg.Wait()
		if err != nil {
			if !strings.Contains(err.Error(), "another seat has acted since") {
				t.Fatalf("round %d: refused for an unexpected reason: %v", round, err)
			}
			refused++
			continue
		}
		committed++
		evs := allEvents(t, run)
		at := -1
		for i, e := range evs {
			if e.GetKey() == k {
				at = i
			}
			if at < 0 || i == at {
				continue
			}
			if e.GetType() == recordpb.EventType_EVENT_TYPE_CORRECTION {
				break
			}
			if e.GetSeatId() != "blue-respond" {
				t.Fatalf("round %d: %s's %s sits between the target and a committed correction", round, e.GetSeatId(), recordpb.Word(e.GetType()))
			}
		}
	}
	t.Logf("%d committed, %d refused", committed, refused)
}
