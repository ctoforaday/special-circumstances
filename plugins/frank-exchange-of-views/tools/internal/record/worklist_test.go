package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// chairBoard is a board at head 2 whose one cast lens has sat twice and found nothing, so it is
// retired at the head, the lens condition of the PASS gate holds, and each test adds only the
// condition it is about. The gaps are minted by outsideLens — a seat not in the cast — so no mint
// makes the cast lens active again.
func chairBoard(t *testing.T, lensSat bool) *stage {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
	if lensSat {
		b.dispatch(2, evLens).register(evLens).dispatch(2, evLens).register(evLens)
	}
	return b
}

const outsideLens = "red-lens-logic"

func cmMint(gap string, cm recordpb.ClassMaterial, severity string, supersedes ...string) *recordpb.Mint {
	sev, _ := GradeOf(severity)
	return &recordpb.Mint{GapId: proto.String(gap), Class: proto.String("x"), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		ClassMaterial: cm.Enum(), Supersedes: supersedes}
}

// openChairSitting opens the chair's sitting after the parties sat and closes its log channel, so
// neither the owed register nor the log holds `complete` false for a reason unrelated to the gate.
func openChairSitting(b *stage) *stage {
	b.register("red-chair")
	return b.add("red-chair", &recordpb.Log{Text: proto.String("nothing blocked"),
		Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})
}

var passGate = &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}

func blockingItems(s SittingJSON) []string {
	var out []string
	for _, it := range s.Open {
		if it.Blocks {
			out = append(out, it.What)
		}
	}
	return out
}

func itemNaming(s SittingJSON, sub string) (Item, bool) {
	for _, it := range s.Open {
		if strings.Contains(it.What, sub) {
			return it, true
		}
	}
	return Item{}, false
}

// An open gap of a `never` class, graded high, does not hold PASS: the chair's list names it with
// blocks:false, the PASS appends, and the sitting is complete.
func TestChairWorkListOverOpenNeverClassGap(t *testing.T) {
	b := chairBoard(t, true)
	b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
	run := openChairSitting(b).seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("a PASS over an open never-class gap was refused: %v", err)
	}
	s := sittingOfRunT(t, run, "chair", "red-chair")
	it, ok := itemNaming(s, "gap G1 is open and not material (its class is never material)")
	if !ok || it.Blocks {
		t.Errorf("the never-class gap must be listed, not blocking: %+v in %v", it, s.Open)
	}
	if !s.Complete {
		t.Errorf("sitting.complete is false after the PASS the gate admitted: %v", blockingItems(s))
	}
}

// `complete` IS THE GATE'S ANSWER. For each class default either side of the floor, and for a
// stranded ancestor, the sitting is complete exactly when the PASS appends, and each open row's
// `material` is the view's.
func TestChairWorkListAgreesWithPassGate(t *testing.T) {
	for _, c := range []struct {
		name         string
		build        func(b *stage)
		wantMaterial bool
		wantAppends  bool
	}{
		{"always-low", func(b *stage) { b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, "low")) }, true, false},
		{"never-high", func(b *stage) { b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high")) }, false, true},
		{"by_grade-low", func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
		}, false, true},
		{"by_grade-medium", func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "medium"))
		}, true, false},
		{"stranded by_grade-low ancestor", func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.add(outsideLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low", "G1"))
		}, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := chairBoard(t, true)
			c.build(b)
			run := openChairSitting(b).seed()
			_, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate)
			appended := err == nil
			if appended != c.wantAppends {
				t.Fatalf("PASS appended=%v, want %v (%v)", appended, c.wantAppends, err)
			}
			s := sittingOfRunT(t, run, "chair", "red-chair")
			if s.Complete != appended {
				t.Errorf("sitting.complete=%v but the PASS appended=%v — the list and the gate disagree: %v", s.Complete, appended, blockingItems(s))
			}
			for _, g := range mustWorkJSONT(t, run).Open {
				if g.ID == "G1" && g.Material != c.wantMaterial {
					t.Errorf("open[G1].material=%v, want %v", g.Material, c.wantMaterial)
				}
			}
			if strings.HasPrefix(c.name, "stranded") {
				it, ok := itemNaming(s, "gap G1 is open and superseded by G2")
				if !ok || !it.Blocks {
					t.Errorf("the stranded ancestor must block: %v", s.Open)
				}
				if _, listed := itemNaming(s, "gap G1 is open and not material"); listed {
					t.Errorf("a stranded ancestor was listed as not holding PASS: %v", s.Open)
				}
				if _, offered := itemNaming(s, "gap G1 is open and you have not closed it"); !offered {
					t.Errorf("the docket offer is missing on the stranded ancestor: %v", s.Open)
				}
			}
		})
	}
}

// EVERY GATE REFUSAL HAS ONE BLOCKING ITEM. For each refusal the Gate case reaches, a board where
// only that refusal holds: the chair's list carries exactly one blocking item, naming it, and the
// PASS is refused at validation. The chair's FAIL is on the record so the terminal-act duty is
// discharged. The FAIL-only convergence refusal has no row: the list claims nothing about a FAIL,
// and the last row holds that nothing blocks where no PASS refusal holds.
func TestChairWorkListStatesEveryGateRefusal(t *testing.T) {
	for _, c := range []struct {
		name    string
		lensSat bool
		ingest  bool
		build   func(b *stage)
		want    string // "" = no refusal holds
	}{
		{"stranded", true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.add(outsideLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low", "G1"))
		}, "superseded by G2"},
		{"material", true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "medium"))
		}, "gap G1 is open and material"},
		{"unruled motion", true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
			b.docketMotion("red-chair", "M1", "G1")
		}, "motion M1"},
		{"no inquiry review", true, true, func(b *stage) {
			b.add("blue-synthesize", &recordpb.Avenue{AvenueId: proto.String("Q1"), Line: proto.String("a line"),
				Hypothesis: proto.String("h"), Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum()})
		}, "account of its own research"},
		{"unraised contradiction", true, true, func(b *stage) {
			b.add(evLens, &recordpb.Verify{Claim: proto.String("the sky is green"), Url: proto.String("https://example.org/sky"),
				Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
				Text: proto.String("it says blue")})
		}, "the sky is green"},
		{"cast lens ready", false, true, func(*stage) {}, "lens red-lens-evidence is ready (active)"},
		{"stale area uncovered", true, true, func(b *stage) {
			// retired at head 2; the head moves; its re-arm sitting is barren (retired for good at
			// that pin); the head moves again, past it.
			b.ingest()
			b.dispatch(int64(b.n), evLens).register(evLens)
			b.ingest()
		}, "area red-lens-evidence is behind its pin"},
		{"no report ingested", false, false, func(*stage) {}, "no report has been ingested"},
		{"none holds", true, true, func(b *stage) {
			b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "low"))
		}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge")
			if c.ingest {
				b.ingest()
			}
			b.register("red-chair")
			if c.lensSat {
				b.dispatch(2, evLens).register(evLens).dispatch(2, evLens).register(evLens)
			}
			c.build(b)
			openChairSitting(b)
			b.add("red-chair", &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)})
			run := b.seed()
			s := sittingOfRunT(t, run, "chair", "red-chair")
			blocking := blockingItems(s)
			verr := validate(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
			if c.want == "" {
				if len(blocking) != 0 || !s.Complete || verr != nil {
					t.Fatalf("no refusal holds, yet blocking=%v complete=%v gate=%v", blocking, s.Complete, verr)
				}
				return
			}
			if verr == nil {
				t.Fatalf("fixture: the gate admits a PASS, so %q is not the refusal holding", c.want)
			}
			if len(blocking) != 1 || !strings.Contains(blocking[0], c.want) {
				t.Errorf("want exactly one blocking item naming %q, got %v (gate: %v)", c.want, blocking, verr)
			}
			if s.Complete {
				t.Errorf("sitting.complete is true while the gate refuses: %v", verr)
			}
		})
	}
}

// THE DOCKET OFFER STANDS ONLY ON A GAP THAT HOLDS PASS: docketing one that does not would create
// the block the gap itself does not. A stranded gap is held as material, so it keeps the offer.
func TestChairDocketAffordanceOnlyOnMaterialGaps(t *testing.T) {
	gaps := []WorkGapState{
		{ID: "NEVER", Open: true, ClassMaterial: "never"},
		{ID: "LOW", Open: true, ClassMaterial: "by_grade", Severity: "low"},
		{ID: "MAT", Open: true, Material: true},
		{ID: "CARRIED", Open: true, Material: true, AwaitingDocket: true},
		{ID: "STRANDED", Open: true, Stranded: true, SupersededBy: "MAT"},
	}
	open := availableOf(nil, gaps, "chair", "red-chair")
	// THE OFFER IS THE VERB, NOT THE GAP'S NAME. A remanded gap is named by a row that says the
	// opposite — it returns only if docketed again — so a predicate that matched the id alone
	// read that row as an offer and could not fail.
	items := func(id string) []string {
		var out []string
		for _, it := range open {
			if strings.Contains(it.What, "gap "+id+" ") {
				out = append(out, it.What)
			}
		}
		return out
	}
	// EVERY item naming the gap is read, not the first: the remanded row is added before the
	// offer would be, so a first-match predicate hides an offer standing behind it.
	anyItem := func(id, want string) bool {
		for _, w := range items(id) {
			if strings.Contains(w, want) {
				return true
			}
		}
		return false
	}
	offered := func(id string) bool { return anyItem(id, "motion docket file") }
	if !anyItem("CARRIED", "BENCH REMANDED") {
		t.Errorf("a carried gap gets the remanded row in place of the offer, got %q", items("CARRIED"))
	}
	for id, want := range map[string]bool{"NEVER": false, "LOW": false, "MAT": true, "CARRIED": false, "STRANDED": true} {
		if offered(id) != want {
			t.Errorf("%s: docket offer present=%v, want %v (%v)", id, offered(id), want, hows(open))
		}
	}
}
