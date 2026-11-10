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

// A READING OF A SOURCE THE LENS FOUND (a Verify with no anchor) records what the source did for the
// sentence and mints nothing: a gap on it is `mint`'s, standing on the reading. No reading places a
// citation.

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

// readFound writes v alone, as a reading is written, and returns the reading as stored and the error.
func readFound(t *testing.T, id Identity, v *recordpb.Verify) (*recordpb.Verify, error) {
	t.Helper()
	ev, err := Append(id, v)
	if err != nil {
		return nil, err
	}
	stored, ok := recordpb.BodyAs[*recordpb.Verify](ev)
	if !ok {
		t.Fatalf("the act's event is %v, not the reading", ev)
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

// A FOUND-SOURCE READING KEYS ON ITS SOURCE AND ITS SENTENCE (gblock, 2026-11-03; the sentence by
// the reading debt's predicate, 2026-11-10). One seat, one sitting: readings that differ in url or in
// sentence are each recorded, and an identical retry, whatever its outcome, is refused naming the
// stored reading's event. In the seat's next sitting the same source and sentence with another
// outcome is recorded.
func TestAFoundSourceReadingKeysOnSourceAndSentence(t *testing.T) {
	run := foundSourceBoard(t, evLens)
	lens := Identity{Run: run, SeatID: evLens}
	const url1, url2, url3 = "https://example.org/one", "https://example.org/two", "https://example.org/three"
	s1, s2 := uncitedSentences[0], uncitedSentences[1]
	supports := recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS
	absent := recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT

	if _, err := readFound(t, lens, foundReading(url1, s1, supports)); err != nil {
		t.Fatalf("(url1, s1): %v", err)
	}
	first := lastEventID(t, run)
	if _, err := readFound(t, lens, foundReading(url1, s2, supports)); err != nil {
		t.Fatalf("(url1, s2): %v; one source backs many sentences", err)
	}
	if _, err := readFound(t, lens, foundReading(url2, s1, supports)); err != nil {
		t.Fatalf("(url2, s1): %v; several sources are read against one sentence", err)
	}
	for _, s := range []string{s1, s2} {
		if _, err := readFound(t, lens, foundReading(url3, s, absent)); err != nil {
			t.Fatalf("an absent reading of url3 on %q: %v", s, err)
		}
	}
	held := verifyCount(t, run)
	for _, o := range []recordpb.SourceOutcome{supports, absent} {
		_, err := readFound(t, lens, foundReading(url1, s1, o))
		mustRefuse(t, err, "you recorded a reading of "+url1+" for this sentence", fmt.Sprintf("at event %d this sitting", first))
	}
	if n := verifyCount(t, run); n != held {
		t.Errorf("the refused retries recorded %d reading(s)", n-held)
	}

	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	weak := recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK
	if _, err := readFound(t, lens, foundReading(url1, s1, weak)); err != nil {
		t.Fatalf("(url1, s1) in the next sitting with another outcome: %v; want it recorded", err)
	}
}

// A REPEATED FOUND-SOURCE READING IN A LATER SITTING IS RECORDED: the same seat, sentence, source and
// outcome. A repeated finds-against reading replaces the seat's standing reading of the sentence, and
// the same debt stands.
func TestARepeatedFoundSourceReadingInALaterSittingIsRecorded(t *testing.T) {
	run := foundSourceBoard(t, evLens)
	lens := Identity{Run: run, SeatID: evLens}
	const url = "https://example.org/one"
	s := uncitedSentences[0]
	refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	if _, err := readFound(t, lens, foundReading(url, s, refutes)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := readFound(t, lens, foundReading(url, s, refutes)); err != nil {
		t.Fatalf("the repeat in a later sitting: %v, want it recorded", err)
	}
	at := lastEventID(t, run)
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var debts []Blocker
	for _, b := range bs {
		if b.Kind == BlockerReadingDebt {
			debts = append(debts, b)
		}
	}
	if len(debts) != 1 || debts[0].Owner != evLens || debts[0].Since() != at {
		t.Errorf("reading debts = %+v, want one, %s's, since the repeat at %d", debts, evLens, at)
	}
}

// NO READING PLACES A CITATION, AND A READING WRITES ONE EVENT (T2). After readings of both forms with
// every outcome — a reading of blue's citation C-1 by each of six lenses, and a reading of a found
// source on its own uncited sentence per outcome — the record holds one Verify per reading and no Mint
// or Anchor beyond the fixture's, the cited sources, the evidence view's sources and the
// protected-anchor set hold blue's cite alone, citedSource refuses every reading, and a reading has no
// label to carry.
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
		if _, err := Append(lens, cited); err != nil {
			t.Fatalf("%s's %s reading of C-1: %v", lenses[n], recordpb.Word(o), err)
		}
		if _, err := readFound(t, Identity{Run: run, SeatID: evLens}, foundReading(fmt.Sprintf("https://example.org/%d", n), uncitedSentences[n], o)); err != nil {
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
	mints, anchors := 0, 0
	for _, e := range b.Events {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_MINT:
			mints++
		case recordpb.EventType_EVENT_TYPE_ANCHOR:
			anchors++
		}
	}
	if mints != 0 || anchors != 0 {
		t.Errorf("the readings wrote %d mint(s) and %d anchor(s); a reading writes one event", mints, anchors)
	}
	if (&recordpb.Verify{}).ProtoReflect().Descriptor().Fields().ByName("label") != nil {
		t.Error("a reading carries a label field: a reading could place a citation")
	}
}
