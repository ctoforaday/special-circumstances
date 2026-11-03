package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// corroboration builds a supporting corroboration of one claim from one source.
func corroboration(label, url, claim string, outcome recordpb.SourceOutcome) *recordpb.Verify {
	v := &recordpb.Verify{CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
		Url:        proto.String(url),
		Title:      proto.String("a source red found"),
		Claim:      proto.String(claim),
		Outcome:    &outcome,
		Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
		Text:       proto.String("what the source says, in red's words"),
	}
	if label != "" {
		v.Label = proto.String(label)
	}
	return v
}

// ONE SOURCE MAY CORROBORATE MANY CLAIMS. This is the collision the label exists to end.
//
// `keyFields` is walked first-match, and before `label` was on Verify the first match for a
// corroboration was `url` — so the key was THE SOURCE. A lens that found one good source
// bearing on three claims recorded the first and had the other two refused as duplicates, which
// is the ordinary case rather than an exotic one: a strong source usually bears on several
// claims. `blue cite` never had this problem because it keys on its minted label.
func TestOneSourceCorroboratesManyClaims(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-evidence"}
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	const url = "https://example.org/one-good-source"
	for i, claim := range []string{"the first claim it bears on", "the second claim it bears on"} {
		v := corroboration(NewID("citation"), url, claim, recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
		if _, err := Append(id, v); err != nil {
			t.Fatalf("corroboration %d of the same source was refused: %v\n"+
				"One source bearing on several claims is the ordinary case; keyed on the URL, only the first could ever record.", i+1, err)
		}
	}
	b, err := FamilyOf(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var got int
	for _, e := range b.Events {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok && v.GetUrl() == url {
			got++
		}
	}
	if got != 2 {
		t.Errorf("%d corroborations of one source on the record, want 2 — the second was dropped, not refused, which is worse", got)
	}
}

// A LABELLED CORROBORATION IS A FOOTNOTE LIKE ANY OTHER.
//
// A human reader cares that the text has appropriate references, not which team inserted them.
// `citationid.go` used to state the opposite as a property — "Red's `lens cite` carries no label
// and is EXCLUDED" — and the consequence was that red's independent corroboration reached no
// reader of the document at all.
func TestASupportingCorroborationJoinsTheBibliography(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-evidence"}
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	label := NewID("citation")
	if _, err := Append(id, corroboration(label, "https://example.org/red-found-this", "blue's claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}
	// And a REFUTING one, which must NOT become a footnote: a source that contradicts the
	// sentence is not a reference backing it, and spliced into the bibliography it reads as
	// support. It carries no label at all. Written as an archived reading is, with no gap carrying it:
	// the footnote rule reads the reading alone.
	Migrating = true
	_, err := Append(id, corroboration("", "https://example.org/contradicts", "blue's other claim", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES))
	Migrating = false
	if err != nil {
		t.Fatal(err)
	}

	srcs, err := CitedSources(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var found, refuting bool
	for _, s := range srcs {
		if s.Label == label {
			found = true
			if s.URL != "https://example.org/red-found-this" {
				t.Errorf("the corroboration's url did not travel: %+v", s)
			}
		}
		if strings.Contains(s.URL, "contradicts") {
			refuting = true
		}
	}
	if !found {
		t.Error("a supporting corroboration is absent from the bibliography source set — red's independent confirmation reaches no reader of the document, which is the defect this change exists to fix")
	}
	if refuting {
		t.Error("a REFUTING corroboration reached the bibliography — a source that contradicts the sentence would render as a reference backing it")
	}

	// The lockdown's EXPECTED set must agree, or red's anchor reads as one blue dropped.
	labels, err := CitationLabels(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var expected bool
	for _, l := range labels {
		if l == label {
			expected = true
		}
	}
	if !expected {
		t.Error("the corroboration's label is missing from CitationLabels — the blue-report lockdown compares that set against the anchors actually present, so red's spliced anchor would report as a dropped citation")
	}
}

// A CORROBORATION THAT FINDS THE SOURCE AGAINST ITS SENTENCE IS A GAP, and the gap holds the PASS.
//
// A supporting corroboration reaches the reader as a footnote. A `refutes` reading backs no sentence
// and is not spliced; what it found reaches the board as the gap its own act mints, of a class its
// outcome admits — and `claim_contradicts_source` is material at every grade. Written with no gap, the
// reading is refused: the defect it found would reach no board.
func TestACorroborationThatFindsAgainstHoldsThePassThroughItsGap(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().
		register("red-chair").register(evLens).seed()
	lens := Identity{Run: run, SeatID: evLens}
	const claim = "The claim stands before."
	reading := func() *recordpb.Verify {
		return corroboration("", "https://example.org/says-otherwise", claim, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)
	}
	if _, err := Append(lens, reading()); err == nil || !strings.Contains(err.Error(), "names the gap that carries it") {
		t.Fatalf("a refuting corroboration written with no gap = %v, want it refused naming its gap", err)
	}
	if err := passRefusalOver(t, run, BlockerMaterialGap); err != nil {
		t.Fatalf("fixture: a material gap holds the PASS before the reading: %v", err)
	}
	evs, err := AppendReading(lens, reading(), readingGap("G-f0000001"), &recordpb.Anchor{Id: proto.String("G-f0000001"), Location: proto.String(claim)})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := recordpb.BodyAs[*recordpb.Verify](evs[len(evs)-1])
	if len(evs) != 3 || !ok || v.GetCarriedBy() != "G-f0000001" || !v.GetCarrierMinted() {
		t.Fatalf("the reading's act wrote %d event(s), the last %+v — want its Mint, its Anchor and the reading carried by the gap its act minted", len(evs), v)
	}
	if err := passRefusalOver(t, run, BlockerMaterialGap); err == nil || !strings.Contains(err.Error(), "G-f0000001") {
		t.Errorf("the reading's gap does not hold the PASS: %v", err)
	}

	// AND A SUPPORTING CORROBORATION NEVER BLOCKS, and carries no gap: it is a reference backing the
	// claim. One naming a gap is refused.
	backing := corroboration(NewID("citation"), "https://example.org/agrees", "another claim entirely", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
	backing.CarriedBy, backing.CarrierMinted = proto.String("G-f0000001"), proto.Bool(false)
	if _, err := Append(lens, backing); err == nil || !strings.Contains(err.Error(), "carried_by and carrier_minted are set by the tool") {
		t.Errorf("a backing corroboration naming a gap = %v, want it refused", err)
	}
	if _, err := Append(lens, corroboration(NewID("citation"), "https://example.org/agrees", "another claim entirely", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}
	if err := passRefusalOver(t, run, BlockerFailedCheck); err != nil {
		t.Errorf("a SUPPORTING corroboration holds the PASS: %v", err)
	}
}

// A SECOND READING OF A SENTENCE AN OPEN GAP OF ITS CLASS HOLDS NAMES THAT GAP. Its act's Mint meets
// the one-gap rule inside the same write; the Mint and its Anchor are dropped and the reading lands
// alone, carried by the standing gap — not minted by it. A reading of another class mints its own.
func TestASecondReadingOfAHeldSentenceNamesItsGap(t *testing.T) {
	const thirdLens = "red-lens-adversary"
	run := newStage(t).cast(evLens, outsideLens, thirdLens, "red-chair", "blue-respond", "judge").ingestCited().
		register("red-chair").register(evLens).register(outsideLens).register(thirdLens).seed()
	const claim = "The claim stands before."
	read := func(seat, gap string, class recordpb.GapClass) []*Event {
		t.Helper()
		m := readingGap(gap)
		m.Class = class.Enum()
		evs, err := AppendReading(Identity{Run: run, SeatID: seat}, corroboration("", "https://example.org/"+seat, claim, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES),
			m, &recordpb.Anchor{Id: proto.String(gap), Location: proto.String(claim)})
		if err != nil {
			t.Fatal(err)
		}
		return evs
	}
	read(evLens, "G-f0000001", recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
	evs := read(outsideLens, "G-f0000002", recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)
	v, ok := recordpb.BodyAs[*recordpb.Verify](evs[len(evs)-1])
	if len(evs) != 1 || !ok || v.GetCarriedBy() != "G-f0000001" || v.GetCarrierMinted() {
		t.Fatalf("the second reading wrote %d event(s), the last %+v — want the reading alone, carried by G-f0000001 and not minting it", len(evs), v)
	}
	if evs := read(thirdLens, "G-f0000003", recordpb.GapClass_GAP_CLASS_WITHIN_SOURCE_CONDITION_MISATTRIBUTION); len(evs) != 3 {
		t.Errorf("a reading of another admitted class wrote %d event(s), want its own Mint, Anchor and reading", len(evs))
	}
}

// THE PROTECTED-ANCHOR SET IS ONE SET, and red's citations are in it.
//
// Two detectors guard the same property — the hookgate lockdown (via CitationLabels) and the
// scorecard's unbacked_citations (via CitationLabelsOf). The scorecard used to build its
// EXPECTED set with its own loop over `Cite` events, so when red's corroborations gained labels
// and started splicing anchors, blue dropping a RED anchor would be caught by one and missed by
// the other. Two detectors for one protection, disagreeing, and neither saying so.
//
// The gain is worth stating as well as the fix: red's citation anchors are IMMORTAL now, the
// same as blue's. Before this they were spliced into the report and protected by nothing.
func TestRedsCitationAnchorsAreProtectedLikeBlues(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-evidence"}
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	label := NewID("citation")
	if _, err := Append(id, corroboration(label, "https://example.org/red", "a claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}

	byRunDir, err := CitationLabels(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	b, err := FamilyOf(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	byBoard := CitationLabelsOf(b.Events)

	if len(byRunDir) != 1 || byRunDir[0] != label {
		t.Errorf("CitationLabels = %v, want just %q — this is the lockdown's EXPECTED set, so a label missing here is an anchor nothing protects", byRunDir, label)
	}
	if strings.Join(byBoard, ",") != strings.Join(byRunDir, ",") {
		t.Errorf("the two readers of the protected-anchor set disagree:\n  CitationLabels   = %v\n  CitationLabelsOf = %v\n"+
			"One guards the PostToolUse lockdown and the other the unbacked_citations detector. Disagreeing, they protect different anchors and neither reports the gap.", byRunDir, byBoard)
	}
}

// A REFUTING READING SPLICES NO FOOTNOTE. The evidence view lists both of a lens's corroborations;
// the supporting one carries the footnote it minted, and the one that finds the source against its
// sentence carries none — what it found is on the board as its gap.
func TestARefutingCorroborationCarriesNoFootnote(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().
		register("red-chair").register(evLens).seed()
	id := Identity{Run: run, SeatID: evLens}
	const claim = "The claim stands before."
	if _, err := AppendReading(id, corroboration("", "https://example.org/says-no", claim, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES),
		readingGap("G-f0000001"), &recordpb.Anchor{Id: proto.String("G-f0000001"), Location: proto.String(claim)}); err != nil {
		t.Fatal(err)
	}
	label := NewID("citation")
	if _, err := Append(id, corroboration(label, "https://example.org/agrees", "a supported claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}
	b, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	var supported, refuted EvidenceVerificationJSON
	for _, v := range EvidenceJSONOf(b.Events, b.At).Independent {
		if v.FindsAgainst() {
			refuted = v
		} else {
			supported = v
		}
	}
	if supported.Label != label {
		t.Errorf("the supporting corroboration does not carry the footnote it minted: %+v", supported)
	}
	if refuted.Claim != claim || refuted.Label != "" {
		t.Errorf("the refuting reading is not listed, or carries a footnote label, so it spliced one after all: %+v", refuted)
	}
}

// THE TWO GUARDS A MUTATION SWEEP FOUND UNTESTED, asserted as the pairs they are.
//
// Both survived a sweep with `||`/`&&` flipped, which means each half of each conjunction was
// only ever satisfied together with the other. That is the shape this branch kept finding in
// production code and it was sitting in code written the same day.
func TestTheCorroborationGuardsHoldEachHalfSeparately(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-evidence"}
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	label := NewID("citation")
	if _, err := Append(id, corroboration(label, "https://example.org/s", "the claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}

	// ExistingCorroboration bails on EITHER half being empty, not only on both. A blank url
	// or a blank claim cannot identify an act, so scanning for one would match on the other alone
	// — returning some unrelated corroboration's label as "this one's retry".
	for _, tc := range []struct{ name, url, claim string }{
		{"no url", "", "the claim"},
		{"no claim", "https://example.org/s", ""},
		{"neither", "", ""},
	} {
		got, err := ExistingCorroboration(mustRun(t, runDir), "red-lens-evidence", tc.url, tc.claim)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != nil {
			t.Errorf("%s: returned %q — an incomplete pair identifies no act, so it must match nothing rather than match on the half it has", tc.name, got.GetLabel())
		}
	}
	// And the complete pair DOES find it, or the guard above would pass by never matching at all.
	if got, err := ExistingCorroboration(mustRun(t, runDir), "red-lens-evidence", "https://example.org/s", "the claim"); err != nil || got.GetLabel() != label {
		t.Errorf("the complete pair returned %q (err %v), want %q — the negative cases above prove nothing if the positive never matches", got.GetLabel(), err, label)
	}
}
