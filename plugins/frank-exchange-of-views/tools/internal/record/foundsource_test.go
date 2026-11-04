package record

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A READING OF A SOURCE THE LENS FOUND (a Verify with no anchor) mints the gap blue answers with a
// cite: a backing source mints `unsourced_claim` on the sentence the report cites nothing for, a
// refuting one `claim_contradicts_source`, and `absent` nothing. No reading places a citation.

// uncitedSentences are the sentences ingestUncited's report holds with no citation behind them.
var uncitedSentences = []string{
	"The first claim stands.", "The second claim stands.", "The third claim stands.",
	"The fourth claim stands.", "The fifth claim stands.", "The sixth claim stands.",
}

// ingestUncited ingests a report whose sentences cite nothing, then ingestCited's sentence, which
// carries citation C-1.
func (b *stage) ingestUncited() *stage {
	return b.add("harness", &recordpb.ReportPart{Text: proto.String("# report\n\n" + strings.Join(uncitedSentences, " ") +
		" The claim stands before " + anchor.Token("C-1") + ".\n")})
}

// foundReading is a lens's reading of url, a source it found, against the report's sentence claim.
func foundReading(url, claim string, o recordpb.SourceOutcome) *recordpb.Verify {
	v := &recordpb.Verify{CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
		Url: proto.String(url), Title: proto.String("a source the lens found"), Claim: proto.String(claim),
		Outcome: o.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("what the source says")}
	if o == recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE {
		v.Bridge = recordpb.BridgeKind_BRIDGE_KIND_SUMMARY.Enum()
	}
	return v
}

// foundGap is the Mint of a found-source reading's gap on claim, of the first class v admits.
func foundGap(gap, claim string, v *recordpb.Verify) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(gap), Class: ReadingGapClasses(v)[0].Enum(), Location: proto.String(claim),
		Problem: proto.String("the report cites nothing for this sentence"), AcceptanceCheck: proto.String("a citation backs the sentence"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}
}

// readFound writes v as its act writes it: with its gap's Mint and Anchor where it mints, alone where
// it does not. It returns the reading as stored and the error.
func readFound(t *testing.T, id Identity, v *recordpb.Verify, gap string) (*recordpb.Verify, error) {
	t.Helper()
	var evs []*Event
	var err error
	if len(ReadingGapClasses(v)) > 0 {
		evs, err = AppendReading(id, v, foundGap(gap, v.GetClaim(), v), &recordpb.Anchor{Id: proto.String(gap), Location: proto.String(v.GetClaim())})
	} else {
		var ev *Event
		if ev, err = Append(id, v); ev != nil {
			evs = []*Event{ev}
		}
	}
	if err != nil {
		return nil, err
	}
	stored, ok := recordpb.BodyAs[*recordpb.Verify](evs[len(evs)-1])
	if !ok {
		t.Fatalf("the act's last event is %v, not the reading", evs[len(evs)-1])
	}
	return stored, nil
}

// lastEventID is the events.id of the last event run holds.
func lastEventID(t *testing.T, run Run) int64 {
	t.Helper()
	var n int64
	if err := recordOf(t, run).QueryRow(`SELECT MAX("id") FROM "events"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// verifyCount is how many readings run holds.
func verifyCount(t *testing.T, run Run) int {
	t.Helper()
	var n int
	if err := recordOf(t, run).QueryRow(`SELECT count(*) FROM "verify"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func foundSourceBoard(t *testing.T, lenses ...string) Run {
	t.Helper()
	b := newStage(t).cast(append(lenses, "red-chair", "blue-respond", "judge")...).ingestUncited().cite("C-1").register("red-chair")
	for _, l := range lenses {
		b.register(l)
	}
	return b.seed()
}

// EVERY OUTCOME, EVERY COMPLETENESS, BOTH FORMS: the classes a reading may take are §III.1's. A reading
// of a citation takes its outcome's gap_classes where it finds against; a reading of a source the
// lens found takes found_source_gap_classes — a backing outcome `unsourced_claim`, `refutes`
// `claim_contradicts_source`, `absent` none — and neither form mints off a copy that is not the
// work. A backing found-source reading mints and does not find against.
func TestReadingGapClassesCoversBothForms(t *testing.T) {
	type C = recordpb.GapClass
	citation := map[recordpb.SourceOutcome][]C{
		recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES: {recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, recordpb.GapClass_GAP_CLASS_WITHIN_SOURCE_CONDITION_MISATTRIBUTION},
		recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT:  {recordpb.GapClass_GAP_CLASS_CITATION_FIGURE_MISATTRIBUTION, recordpb.GapClass_GAP_CLASS_WITHIN_SOURCE_CONDITION_MISATTRIBUTION},
	}
	backing := map[recordpb.SourceOutcome]bool{
		recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS:             true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE: true,
		recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK:                 true,
	}
	outcomes := recordpb.SourceOutcome(0).Descriptor().Values()
	completeness := recordpb.SourceCompleteness(0).Descriptor().Values()
	rows := 0
	for i := 0; i < outcomes.Len(); i++ {
		o := recordpb.SourceOutcome(outcomes.Get(i).Number())
		for j := 0; j < completeness.Len(); j++ {
			sc := recordpb.SourceCompleteness(completeness.Get(j).Number())
			reads := o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE && o != recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED &&
				sc != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
			for _, anchored := range []bool{true, false} {
				v := &recordpb.Verify{Outcome: o.Enum(), SourceCompleteness: sc.Enum()}
				var want []C
				if anchored {
					v.Anchor = proto.String("C-1")
					if reads {
						want = citation[o]
					}
				} else if reads {
					want = foundSource[o]
				}
				rows++
				if got := ReadingGapClasses(v); !slices.Equal(got, want) {
					t.Errorf("%s of a %s copy, anchored %v: classes %v, want %v", recordpb.Word(o), recordpb.Word(sc), anchored, got, want)
				}
				if !anchored && backing[o] && FindsAgainst(v) {
					t.Errorf("a backing found-source reading (%s, %s) finds against its sentence; its source backs it", recordpb.Word(o), recordpb.Word(sc))
				}
			}
		}
	}
	if rows < 4*outcomes.Len() {
		t.Fatalf("%d rows walked, too few to read both forms", rows)
	}
}

// A FOUND-SOURCE READING KEYS ON ITS SOURCE AND ITS SENTENCE (gblock, 2026-11-03). One seat, one
// sitting: readings that differ in url or in sentence are each recorded — the second url read against
// a sentence the first url's gap carries names that gap — and an identical retry, whatever its
// outcome, is refused naming the stored reading's event. In the seat's next sitting the same source
// and sentence with another outcome is recorded.
func TestAFoundSourceReadingKeysOnSourceAndSentence(t *testing.T) {
	run := foundSourceBoard(t, evLens)
	lens := Identity{Run: run, SeatID: evLens}
	const url1, url2, url3 = "https://example.org/one", "https://example.org/two", "https://example.org/three"
	s1, s2 := uncitedSentences[0], uncitedSentences[1]
	supports := recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS
	absent := recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT

	v, err := readFound(t, lens, foundReading(url1, s1, supports), "G-a0000001")
	if err != nil || v.GetCarriedBy() != "G-a0000001" || !v.GetCarrierMinted() {
		t.Fatalf("(url1, s1) = %+v, %v; want it to mint G-a0000001", v, err)
	}
	first := lastEventID(t, run)
	if v, err = readFound(t, lens, foundReading(url1, s2, supports), "G-a0000002"); err != nil || v.GetCarriedBy() != "G-a0000002" || !v.GetCarrierMinted() {
		t.Fatalf("(url1, s2) = %+v, %v; one source backs many sentences, so it records and mints G-a0000002", v, err)
	}
	if v, err = readFound(t, lens, foundReading(url2, s1, supports), "G-a0000003"); err != nil || v.GetCarriedBy() != "G-a0000001" || v.GetCarrierMinted() {
		t.Fatalf("(url2, s1) = %+v, %v; several sources are read against one sentence, so it records naming G-a0000001", v, err)
	}
	for _, s := range []string{s1, s2} {
		if _, err := readFound(t, lens, foundReading(url3, s, absent), ""); err != nil {
			t.Fatalf("an absent reading of url3 on %q: %v", s, err)
		}
	}
	held := verifyCount(t, run)
	for _, o := range []recordpb.SourceOutcome{supports, absent} {
		_, err := readFound(t, lens, foundReading(url1, s1, o), "G-a0000009")
		mustRefuse(t, err, "you recorded a reading of "+url1+" for this sentence", fmt.Sprintf("at event %d this sitting", first), "gap G-a0000001 carries it")
	}
	if n := verifyCount(t, run); n != held {
		t.Errorf("the refused retries recorded %d reading(s)", n-held)
	}

	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	weak := recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK
	if v, err := readFound(t, lens, foundReading(url1, s1, weak), "G-a0000004"); err != nil || v.GetCarriedBy() != "G-a0000001" {
		t.Fatalf("(url1, s1) in the next sitting with another outcome = %+v, %v; want it recorded, naming G-a0000001", v, err)
	}
}

// A REPEATED FOUND-SOURCE READING IS REFUSED, in a later sitting, while the gap carrying the first is
// open: the same seat, sentence, source and outcome. The refusal names the gap, and is worded for a
// reading that may back its sentence — it neither says the source was found against nor sends the
// seat to a finding.
func TestARepeatedFoundSourceReadingIsRefused(t *testing.T) {
	run := foundSourceBoard(t, evLens)
	lens := Identity{Run: run, SeatID: evLens}
	const url = "https://example.org/one"
	s := uncitedSentences[0]
	supports := recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS
	if _, err := readFound(t, lens, foundReading(url, s, supports), "G-a0000001"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	held := verifyCount(t, run)
	_, err := readFound(t, lens, foundReading(url, s, supports), "G-a0000002")
	mustRefuse(t, err, "your reading of "+url+" for this sentence", "gap G-a0000001", "nothing is recorded")
	if err != nil {
		for _, banned := range []string{"finding", "finds the source against"} {
			if strings.Contains(err.Error(), banned) {
				t.Errorf("the repeat refusal of a backing found-source reading says %q:\n%v", banned, err)
			}
		}
	}
	if n := verifyCount(t, run); n != held {
		t.Errorf("the refused repeat recorded %d reading(s)", n-held)
	}
}

// THE REPEAT RULE REACHES A BACKING FOUND-SOURCE READING, asked at the record layer itself: its early
// return keys on whether the reading mints, so a backing reading of a source the lens found, whose
// earlier twin's gap is open, meets the unanchored arm.
func TestTheRepeatRuleReachesABackingFoundSourceReading(t *testing.T) {
	run := foundSourceBoard(t, evLens)
	lens := Identity{Run: run, SeatID: evLens}
	v := foundReading("https://example.org/one", uncitedSentences[0], recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
	if _, err := readFound(t, lens, proto.Clone(v).(*recordpb.Verify), "G-a0000001"); err != nil {
		t.Fatal(err)
	}
	if err := requireNotARepeatReading(recordOf(t, run), evLens, v); err == nil || !strings.Contains(err.Error(), "G-a0000001") {
		t.Errorf("requireNotARepeatReading on a backing found-source twin = %v, want a refusal naming G-a0000001", err)
	}
}

// NO READING PLACES A CITATION. After readings of both forms with every outcome — a reading of blue's
// citation C-1 by each of six lenses, and a reading of a found source on its own uncited sentence per
// outcome — the cited sources, the evidence view's sources and the protected-anchor set hold blue's
// cite alone, citedSource refuses every reading, and a reading has no label to carry.
func TestNoReadingPlacesACitation(t *testing.T) {
	lenses := []string{evLens, outsideLens, "red-lens-adversary", "red-lens-dark-side", "red-lens-voice", "red-lens-architecture"}
	run := foundSourceBoard(t, lenses...)
	outcomes := recordpb.SourceOutcome(0).Descriptor().Values()
	n := 0
	for i := 0; i < outcomes.Len(); i++ {
		o := recordpb.SourceOutcome(outcomes.Get(i).Number())
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED {
			continue
		}
		lens := Identity{Run: run, SeatID: lenses[n]}
		cited := &recordpb.Verify{Anchor: proto.String("C-1"), Claim: proto.String("The claim stands before."), Outcome: o.Enum(),
			Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("read at the leaf"),
			CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()}
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE {
			cited.Bridge = recordpb.BridgeKind_BRIDGE_KIND_SUMMARY.Enum()
		}
		var err error
		if classes := ReadingGapClasses(cited); len(classes) > 0 {
			g := readingGap(fmt.Sprintf("G-c000000%d", n))
			g.Class = classes[0].Enum()
			_, err = AppendReading(lens, cited, g, &recordpb.Anchor{Id: g.GapId, SentenceOf: proto.String("C-1")})
		} else {
			_, err = Append(lens, cited)
		}
		if err != nil {
			t.Fatalf("%s's %s reading of C-1: %v", lenses[n], recordpb.Word(o), err)
		}
		if _, err := readFound(t, Identity{Run: run, SeatID: evLens}, foundReading(fmt.Sprintf("https://example.org/%d", n), uncitedSentences[n], o), fmt.Sprintf("G-f000000%d", n)); err != nil {
			t.Fatalf("a %s reading of a found source: %v", recordpb.Word(o), err)
		}
		n++
	}

	srcs, err := CitedSources(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Label != "C-1" {
		t.Errorf("cited sources = %+v, want blue's C-1 alone", srcs)
	}
	labels, err := CitationLabels(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(labels, []string{"C-1"}) {
		t.Errorf("the protected-anchor set = %v, want blue's C-1 alone", labels)
	}
	b, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	ev := EvidenceJSONOf(b.Events, b.At)
	if len(ev.Sources) != 1 || ev.Sources[0].Anchor != "C-1" {
		t.Errorf("the evidence view's sources = %+v, want blue's C-1 alone", ev.Sources)
	}
	readings := 0
	for _, e := range b.Events {
		body, _ := recordpb.Body(e)
		if _, ok := body.(*recordpb.Verify); !ok {
			continue
		}
		readings++
		if src, ok := citedSource(body); ok {
			t.Errorf("a reading is read as a cited source: %+v", src)
		}
	}
	if readings != 2*n {
		t.Errorf("%d readings on the record, want %d", readings, 2*n)
	}
	if (&recordpb.Verify{}).ProtoReflect().Descriptor().Fields().ByName("label") != nil {
		t.Error("a reading carries a label field: a reading could place a citation")
	}
}
