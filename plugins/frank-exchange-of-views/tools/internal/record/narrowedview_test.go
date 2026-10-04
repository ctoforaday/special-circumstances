package record

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY NARROWED VIEW RENDERS WHAT THE WHOLE RECORD WOULD.
//
// A view that loads only the families it renders is right exactly when nothing outside them reaches
// its output, and a family its renderer reads that the declaration leaves out does not error — it
// renders as nothing, in the bytes the view uses for "none on the record". So each declared view is
// rendered twice on one run: from its own narrowed read, and from the whole record the loader reads
// entire. Any difference is a family the declaration is missing.
//
// The run holds the joins the views make across families, because a missing family shows only
// where an event of it changes the output: a verdict per epoch (debate), a finding that answers a
// contradicting check and an edit that reopens an anchor (evidence), a mint crediting a finding
// (findings), a docket ruling settling a gap (board, debate, motions), a check backing an open gap's
// anchor (board), and a ruled avenue (motions). It holds a CORRECTION of a correctable act in every
// view that declares one, because the narrowed read loads corrections without being asked (eventsOfAt)
// and a view whose struck act renders as standing differs from the whole record only where one exists.
func TestEveryNarrowedViewRendersWhatTheWholeRecordWould(t *testing.T) {
	dir := newRun(t)
	str := proto.String
	medium := recordtest.P(recordpb.Grade_GRADE_MEDIUM)
	recordtest.Seed(t, dir,
		// Epoch 1.
		recordtest.Event(t, "red-chair", &recordpb.Register{ToolVersion: str("test")}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Register{ToolVersion: str("test")}),
		recordtest.Event(t, HarnessSeat, &recordpb.SittingOpen{AgentId: str("blue-1"),
			AgentType: str("frank-exchange-of-views:blue-researcher"), SeatId: str("blue-respond")}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Finding{FindingId: str("f-1"), Label: str("logic-F1"),
			Location: str("claim A"), Text: str("the source refutes claim A"), Severity: medium}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Finding{FindingId: str("f-2"), Label: str("logic-F2"),
			Text: str("uncredited")}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Mint{GapId: str("G1"), Class: str("self-attestation"),
			Problem: str("p"), RequiredFix: str("f"), AcceptanceCheck: str("a"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: medium, Likelihood: medium, Impact: medium,
			FoundBy: []string{"logic-F1"}}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Mint{GapId: str("G2"), Class: str("self-attestation"),
			Problem: str("p2"), RequiredFix: str("f"), AcceptanceCheck: str("a"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: medium, Likelihood: medium, Impact: medium}),
		// An open gap standing on the anchor, so the board renders its backing.
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Mint{GapId: str("G3"), Class: str("self-attestation"),
			Problem: str("p3"), RequiredFix: str("f"), AcceptanceCheck: str("a"), Location: str("claim A <!--cite:c-0a0a0a0a-->"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: medium, Likelihood: medium, Impact: medium}),
		recordtest.Event(t, "blue-respond", &recordpb.Cite{Label: str("c-0a0a0a0a"), Url: str("https://example.org"),
			Title: str("t"), CiteKey: str("k1"), Location: str("claim A")}),
		// A contradiction the finding above answers, and one nothing answers.
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Verify{Url: str("https://example.org"), Anchor: str("c-0a0a0a0a"),
			Claim: str("claim A"), Title: str("t"), Text: str("says otherwise"),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum()}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Verify{Url: str("https://example.org/b"),
			Claim: str("claim B"), Title: str("t"), Text: str("not there"),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum()}),
		recordtest.Event(t, "blue-respond", &recordpb.Proof{ProofId: str("p-1"), ProofSha: str("abc"), ProofBasis: str("basis")}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Reproduce{ProofSha: str("abc"), Reproduced: proto.Bool(true),
			Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: str("ran")}),
		// An edit that reopened the anchor (its text leaves G3's location where it is).
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Old: str("elsewhere"), New: str("elsewhere, revised"),
			Text: str("why"), Reopened: []string{"c-0a0a0a0a"}}),
		recordtest.Event(t, "red-chair", &recordpb.Regrade{GapId: str("G2"), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: str("moved")}),
		recordtest.Event(t, "red-chair", &recordpb.Position{Text: str("red's position")}),
		recordtest.Event(t, "blue-respond", &recordpb.Position{Text: str("blue's position")}),
		recordtest.Event(t, "red-chair", &recordpb.Closing{GapId: str("G1"), Text: str("red closes on G1")}),
		recordtest.Event(t, "blue-respond", &recordpb.Avenue{AvenueId: str("Q1"),
			Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: str("a direction")}),
		// The chair rules the avenue, which makes it a direction motion.
		recordtest.At(t, "red-chair", "red-chair:motion_rule:#1", &recordpb.MotionRule{MotionId: str("Q1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_AVENUE), Opinion: str("worth the run's time"),
			Ruling: &recordpb.MotionRule_Avenue{Avenue: recordpb.AvenueRuling_AVENUE_RULING_OUT_OF_SCOPE}}),
		recordtest.Event(t, "red-chair", &recordpb.Log{Text: str("no verb renders a gap's lineage in one read"),
			Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		// The corrections: each struck act, its replacement, and the correction naming both.
		recordtest.At(t, "red-chair", "red-chair:motion_rule:#1~1", &recordpb.MotionRule{MotionId: str("Q1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_AVENUE), Opinion: str("out of scope, on a second reading"),
			Ruling: &recordpb.MotionRule_Avenue{Avenue: recordpb.AvenueRuling_AVENUE_RULING_OUT_OF_SCOPE}}),
		correctionOf(t, "red-chair", "red-chair:motion_rule:#1"),
		recordtest.At(t, "red-chair", "red-chair:regrade:#9", &recordpb.Regrade{GapId: str("G2"),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: str("misread")}),
		recordtest.At(t, "red-chair", "red-chair:regrade:#9~1", &recordpb.Regrade{GapId: str("G2"),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: str("read again")}),
		correctionOf(t, "red-chair", "red-chair:regrade:#9"),
		recordtest.At(t, "red-chair", "red-chair:position:#9", &recordpb.Position{Text: str("red's position, misworded")}),
		recordtest.At(t, "red-chair", "red-chair:position:#9~1", &recordpb.Position{Text: str("red's position, reworded")}),
		correctionOf(t, "red-chair", "red-chair:position:#9"),
		recordtest.At(t, "red-chair", "red-chair:log:#9", &recordpb.Log{Text: str("the tool  refused"),
			Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		recordtest.At(t, "red-chair", "red-chair:log:#9~1", &recordpb.Log{Text: str("the tool refused the cite"),
			Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		correctionOf(t, "red-chair", "red-chair:log:#9"),
		recordtest.At(t, "blue-respond", "blue-respond:proof:#9", &recordpb.Proof{ProofId: str("p-2"), ProofSha: str("def"), ProofBasis: str("basis, misstated")}),
		recordtest.At(t, "blue-respond", "blue-respond:proof:#9~1", &recordpb.Proof{ProofId: str("p-2"), ProofSha: str("def"), ProofBasis: str("basis, restated")}),
		correctionOf(t, "blue-respond", "blue-respond:proof:#9"),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()}),
		// Epoch 2: the chair sits again, dockets G2, the bench rules, red closes G1 and passes.
		recordtest.Event(t, "red-chair", &recordpb.Register{ToolVersion: str("test")}),
		recordtest.Event(t, "red-chair", &recordpb.Motion{MotionId: str("M1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET), Basis: str("red cannot settle G2"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: str("G2")}}}),
		recordtest.Event(t, "judge", &recordpb.MotionRule{MotionId: str("M1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET), Opinion: str("the bench's reasoning"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
				Principle:   str("pr"), Tension: str("tn"), ReviewFlag: str("rf"), Settled: str("st"), Final: proto.Bool(true)}}}),
		recordtest.Event(t, "blue-respond", &recordpb.MotionAppeal{MotionId: str("M1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET), Reason: str("the ruling reads past the argument")}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Close{GapId: str("G1"),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   str("L1"), AnchorTool: str("go test"), AnchorTarget: str("./x"), Prose: str("verified at the leaf")}),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()}),
	)
	run := mustRun(t, dir)
	whole, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}

	// THE RUN CARRIES EVERY FAMILY A VIEW DECLARES, so a view's parity below is measured on events
	// it actually loads. A declared family this run lacks is one whose omission nothing here sees.
	held := map[recordpb.EventType]bool{}
	typeOfKey := map[string]recordpb.EventType{}
	for _, e := range whole.Events {
		held[e.GetType()] = true
		if e.GetKey() != "" {
			typeOfKey[e.GetKey()] = e.GetType()
		}
	}
	corrected := map[recordpb.EventType]bool{}
	for _, e := range whole.Events {
		if c, ok := recordpb.BodyAs[*recordpb.Correction](e); ok {
			corrected[typeOfKey[c.GetCorrects()]] = true
		}
	}
	names := make([]string, 0, len(narrowedViews))
	for name := range narrowedViews {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		t.Fatal("no narrowed view is declared — the registry is empty, so this test measures nothing")
	}
	for _, name := range names {
		v := narrowedViews[name]
		t.Run(name, func(t *testing.T) {
			correctable, correctedHere := false, false
			for _, f := range v.declared() {
				if !held[f] {
					t.Errorf("the %s view declares %s and this run holds none — seed one, or its parity is unmeasured", name, recordpb.Word(f))
				}
				if recordpb.Tier(f) != recordpb.CorrectionTier_CORRECTION_TIER_NONE {
					correctable = true
					correctedHere = correctedHere || corrected[f]
				}
			}
			if correctable && !correctedHere {
				t.Errorf("the %s view declares a correctable family and this run corrects none of its acts — seed a correction, or whether its read loads corrections is unmeasured", name)
			}
			narrowed, err := v.ofAny(run)
			if err != nil {
				t.Fatal(err)
			}
			var entire any
			if err := readSnapshot(run, func(q recordsql.Querier) error {
				entire, err = v.renderAny(run, q, whole.Events, whole.At)
				return err
			}); err != nil {
				t.Fatal(err)
			}
			got, want := mustJSON(t, narrowed), mustJSON(t, entire)
			if got != want {
				t.Errorf("the %s view renders differently from its narrowed read than from the whole record — "+
					"its renderer reads a family its declaration does not load.\nnarrowed:\n%s\nwhole record:\n%s", name, got, want)
			}
		})
	}
}

// THE RECORDED VERDICT IS EACH EPOCH'S, on the run path. An empty verdict means none was recorded,
// so a debate read that did not load the gates printed that for every epoch of every run (#1243).
func TestDebateJSONBytesCarriesEachEpochsRecordedVerdict(t *testing.T) {
	dir := newRun(t)
	recordtest.Seed(t, dir,
		recordtest.Event(t, "red-chair", &recordpb.Register{ToolVersion: proto.String("test")}),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()}),
		recordtest.Event(t, "red-chair", &recordpb.Register{ToolVersion: proto.String("test")}),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()}),
	)
	b, err := DebateJSONBytes(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	var dj DebateJSON
	if err := json.Unmarshal(b, &dj); err != nil {
		t.Fatal(err)
	}
	got := map[int]string{}
	for _, e := range dj.Epochs {
		got[e.Epoch] = e.Verdict
	}
	for epoch, want := range map[int]string{1: "fail", 2: "pass"} {
		if got[epoch] != want {
			t.Errorf("epoch %d: verdict %q, the record holds %q\n%s", epoch, got[epoch], want, b)
		}
	}
}

// THE EVIDENCE VIEW READS THE ACTS ITS TWO LISTS ARE ABOUT, on the run path: a finding at the
// contradicted claim answers it, and an edit that reopened an anchor lists it. Without them every
// contradiction reads as unanswered and no anchor as moved — on every run.
func TestEvidenceJSONBytesReadsTheAnswersAndTheReopenings(t *testing.T) {
	dir := newRun(t)
	str := proto.String
	recordtest.Seed(t, dir,
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Verify{Url: str("https://example.org"),
			Claim: str("claim A"), Title: str("t"), Text: str("says otherwise"),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum()}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Verify{Url: str("https://example.org/b"),
			Claim: str("claim B"), Title: str("t"), Text: str("not there"),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum()}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Finding{FindingId: str("f-1"), Label: str("logic-F1"),
			Location: str("claim A"), Text: str("the source refutes claim A")}),
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Old: str("x"), New: str("y"), Text: str("why"),
			Reopened: []string{"c-1"}}),
	)
	b, err := EvidenceJSONBytes(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	var ej EvidenceJSON
	if err := json.Unmarshal(b, &ej); err != nil {
		t.Fatal(err)
	}
	if len(ej.UnansweredContradictions) != 1 || ej.UnansweredContradictions[0] != "claim B" {
		t.Errorf("unanswered_contradictions = %q, want only [claim B] — claim A has a finding at it", ej.UnansweredContradictions)
	}
	if len(ej.Reopened) != 1 || ej.Reopened[0] != "c-1" {
		t.Errorf("reopened = %q, want [c-1] — a recorded edit reopened it", ej.Reopened)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A NAME IS DECLARED ONCE. The registry is keyed by name, so a second declaration would replace the
// first and the parity test would walk one of the two.
func TestANarrowedViewNameIsDeclaredOnce(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a second declaration of \"board\" was accepted — the registry now holds one of the two")
		}
	}()
	declareNarrowedView("board", rendersEvents(FindingsJSONOf), recordpb.EventType_EVENT_TYPE_FINDING)
}

// NO PROJECTION READS AROUND ITS DECLARATION. EventsOf takes its families from the caller, so a
// projection calling it renders whatever it forgot to name, with nothing to say so — the shape #1243
// was. Tests outside the package use it to read the record back; nothing else in the module may.
func TestNoProjectionReadsAroundItsDeclaration(t *testing.T) {
	root := filepath.Join("..", "..") // the tools module
	var callers []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) && path != root {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		inRecord := f.Name.Name == "record"
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				if inRecord && fn.Name == "EventsOf" {
					callers = append(callers, path)
				}
			case *ast.SelectorExpr:
				if fn.Sel.Name == "EventsOf" {
					callers = append(callers, path)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(callers) > 0 {
		t.Errorf("EventsOf is called outside a test in %q — a projection declares its families with "+
			"declareNarrowedView, which the parity test holds to the whole record", callers)
	}
}

// THE BOARD IS READ OFF ONE SNAPSHOT. Its closures come from the events and its openness from the
// `gap` view; asked of the record at two moments, a close landing between them rendered a gap closed
// with no closure and epoch 0. A lens closes gaps while the board is read, and every closed gap on
// every board read carries its closure and the epoch it closed in.
func TestEveryClosedGapOnTheBoardCarriesItsClosureWhileALensCloses(t *testing.T) {
	dir := newRun(t)
	str := proto.String
	medium := recordtest.P(recordpb.Grade_GRADE_MEDIUM)
	const gaps = 25
	seed := []*recordpb.Event{
		recordtest.Event(t, "red-chair", &recordpb.Register{ToolVersion: str("test")}),
		recordtest.Event(t, "red-lens-r1-logic", &recordpb.Register{ToolVersion: str("test")}),
	}
	closes := make([]*recordpb.Event, 0, gaps)
	for i := 1; i <= gaps; i++ {
		id := fmt.Sprintf("G%d", i)
		seed = append(seed, recordtest.Event(t, "red-lens-r1-logic", &recordpb.Mint{GapId: str(id), Class: str("self-attestation"),
			Problem: str("p"), RequiredFix: str("f"), AcceptanceCheck: str("a"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: medium, Likelihood: medium, Impact: medium}))
		closes = append(closes, recordtest.Stamped(recordtest.Event(t, "red-lens-r1-logic", &recordpb.Close{GapId: str(id),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   str("L1"), AnchorTool: str("go test"), AnchorTarget: str("./x"), Prose: str("verified")}),
			"2026-01-01T01:00:00Z"))
	}
	recordtest.Seed(t, dir, seed...)
	db, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	run := mustRun(t, dir)

	done := make(chan error, 1)
	go func() {
		for _, ev := range closes {
			if _, err := recordsql.Insert(db, ev); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	for closing := true; closing; {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			closing = false
		default:
		}
		bj, err := BoardJSONOfRun(run)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range bj.Closed {
			if g.Closure == nil || g.ClosedEpoch == 0 {
				t.Fatalf("gap %s reads closed with closure %v and closed_epoch %d — its openness and its closing act came off two reads of the record",
					g.ID, g.Closure, g.ClosedEpoch)
			}
		}
	}
}
