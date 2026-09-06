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
	v := &recordpb.Verify{
		Independent: proto.Bool(true),
		Url:         proto.String(url),
		Title:       proto.String("a source red found"),
		Claim:       proto.String(claim),
		Outcome:     &outcome,
		Confidence:  recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
		Text:        proto.String("what the source says, in red's words"),
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
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-r1-evidence", Round: 1}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	const url = "https://example.org/one-good-source"
	for i, claim := range []string{"the first claim it bears on", "the second claim it bears on"} {
		v := corroboration(NewCitationID(), url, claim, recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
		if _, err := Append(id, v); err != nil {
			t.Fatalf("corroboration %d of the same source was refused: %v\n"+
				"One source bearing on several claims is the ordinary case; keyed on the URL, only the first could ever record.", i+1, err)
		}
	}
	b, err := BoardState(mustRun(t, runDir))
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
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-r1-evidence", Round: 1}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	label := NewCitationID()
	if _, err := Append(id, corroboration(label, "https://example.org/red-found-this", "blue's claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}
	// And a REFUTING one, which must NOT become a footnote: a source that contradicts the
	// sentence is not a reference backing it, and spliced into the bibliography it reads as
	// support. It carries no label at all.
	if _, err := Append(id, corroboration("", "https://example.org/contradicts", "blue's other claim", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)); err != nil {
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

// A CONTRADICTION RED FOUND AND NEVER RAISED BLOCKS A PASS.
//
// A supporting corroboration reaches the reader as a footnote. `refutes` and `absent` are not
// references backing the sentence and are deliberately not spliced — which would leave them in
// the `evidence` projection alone, seen by red and by nobody else. That is the same defect one
// axis over: red's strongest finding on this axis, with no reader.
//
// The remedy is a FINDING and the tool does not write it: a lens structurally cannot mint, and
// writing the finding here would mean inventing its severity, likelihood and impact — three
// grades nobody chose, feeding the mass calculation that decides what a gap is worth.
func TestAContradictionNobodyRaisedBlocksThePass(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	lens := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-r1-evidence", Round: 1}
	merge := Identity{Run: mustRun(t, runDir), SeatID: "red-merge-r1", Round: 1}
	for _, id := range []Identity{lens, merge} {
		if _, _, err := RegisterSeat(id, ""); err != nil {
			t.Fatal(err)
		}
	}
	const claim = "the API returns 200 on every path"
	if _, err := Append(lens, corroboration("", "https://example.org/says-otherwise", claim, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)); err != nil {
		t.Fatal(err)
	}

	err := requirePassClosesAllGaps(mustRun(t, runDir))
	if err == nil {
		t.Fatal("a PASS was allowed over a contradiction red recorded and nobody raised")
	}
	if !strings.Contains(err.Error(), claim) {
		t.Errorf("the refusal does not name the claim, so red cannot tell which reading is unanswered: %v", err)
	}
	if !strings.Contains(err.Error(), "lens finding") {
		t.Errorf("the refusal does not name the act that clears it — a blocking message that does not say how to unblock is an invitation to guess: %v", err)
	}

	// A FINDING QUOTING THE SAME CLAIM CLEARS IT. Red grades its own finding; the tool only
	// reports that one is owed.
	if _, err := Append(lens, &recordpb.Finding{
		Label:      proto.String("L1-F1"),
		Location:   proto.String(claim),
		Text:       proto.String("the source says the opposite at page 9"),
		Severity:   recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH),
		Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}
	if err := requirePassClosesAllGaps(mustRun(t, runDir)); err != nil {
		t.Errorf("the contradiction was raised as a finding and the PASS is still blocked: %v", err)
	}

	// AND A SUPPORTING CORROBORATION NEVER BLOCKS. It is a reference backing the claim; there is
	// nothing to raise.
	if _, err := Append(lens, corroboration(NewCitationID(), "https://example.org/agrees", "another claim entirely", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}
	if err := requirePassClosesAllGaps(mustRun(t, runDir)); err != nil {
		t.Errorf("a SUPPORTING corroboration blocked a PASS — only a contradiction owes a finding: %v", err)
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
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-r1-evidence", Round: 1}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	label := NewCitationID()
	if _, err := Append(id, corroboration(label, "https://example.org/red", "a claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}

	byRunDir, err := CitationLabels(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	b, err := BoardState(mustRun(t, runDir))
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

// THE DUTY IS VISIBLE WHERE RED LOOKS, not only where the merge is blocked.
//
// A contradiction with no finding refuses a PASS — but that is the MERGE's refusal, at the end
// of the round, and a duty the owing seat never sees is one it cannot discharge. The evidence
// projection is where a lens reads its own axis, so the outstanding readings are named there
// too. The empty array is the point as much as the full one: without the field, "nothing
// outstanding" and "nobody checked" are the same absence.
func TestTheEvidenceViewNamesTheContradictionsStillOwed(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-r1-evidence", Round: 1}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	const claim = "the cache never evicts under load"
	if _, err := Append(id, corroboration("", "https://example.org/says-no", claim, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)); err != nil {
		t.Fatal(err)
	}
	label := NewCitationID()
	if _, err := Append(id, corroboration(label, "https://example.org/agrees", "a supported claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}

	b, err := BoardState(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	ev := EvidenceJSONOf(b.Events)
	if len(ev.UnansweredContradictions) != 1 || ev.UnansweredContradictions[0] != claim {
		t.Errorf("unanswered_contradictions = %v, want just %q — red cannot discharge a duty it cannot see", ev.UnansweredContradictions, claim)
	}
	// The supporting one carries the footnote it minted; the refuting one carries none, because
	// it spliced none.
	var supported, refuted EvidenceVerificationJSON
	for _, v := range ev.Independent {
		if v.Refuted() {
			refuted = v
		} else {
			supported = v
		}
	}
	if supported.Label != label {
		t.Errorf("the supporting corroboration does not carry the footnote it minted: %+v", supported)
	}
	if refuted.Label != "" {
		t.Errorf("a refuting reading carries a footnote label, so it spliced one after all: %+v", refuted)
	}

	// Raise it, and the array empties — an empty array, not a null.
	if _, err := Append(id, &recordpb.Finding{
		Label: proto.String("L1-F1"), Location: proto.String(claim), Text: proto.String("the source says otherwise"),
		Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH),
		Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
	}); err != nil {
		t.Fatal(err)
	}
	b, _ = BoardState(mustRun(t, runDir))
	if got := EvidenceJSONOf(b.Events).UnansweredContradictions; len(got) != 0 {
		t.Errorf("unanswered_contradictions = %v after the finding was raised, want empty", got)
	}
}

// THE TWO GUARDS A MUTATION SWEEP FOUND UNTESTED, asserted as the pairs they are.
//
// Both survived a sweep with `||`/`&&` flipped, which means each half of each conjunction was
// only ever satisfied together with the other. That is the shape this branch kept finding in
// production code and it was sitting in code written the same day.
func TestTheCorroborationGuardsHoldEachHalfSeparately(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-r1-evidence", Round: 1}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	label := NewCitationID()
	if _, err := Append(id, corroboration(label, "https://example.org/s", "the claim", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)); err != nil {
		t.Fatal(err)
	}

	// ExistingCorroborationLabel bails on EITHER half being empty, not only on both. A blank url
	// or a blank claim cannot identify an act, so scanning for one would match on the other alone
	// — returning some unrelated corroboration's label as "this one's retry".
	for _, tc := range []struct{ name, url, claim string }{
		{"no url", "", "the claim"},
		{"no claim", "https://example.org/s", ""},
		{"neither", "", ""},
	} {
		got, err := ExistingCorroborationLabel(mustRun(t, runDir), "red-lens-r1-evidence", tc.url, tc.claim)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != "" {
			t.Errorf("%s: returned %q — an incomplete pair identifies no act, so it must match nothing rather than match on the half it has", tc.name, got)
		}
	}
	// And the complete pair DOES find it, or the guard above would pass by never matching at all.
	if got, err := ExistingCorroborationLabel(mustRun(t, runDir), "red-lens-r1-evidence", "https://example.org/s", "the claim"); err != nil || got != label {
		t.Errorf("the complete pair returned %q (err %v), want %q — the negative cases above prove nothing if the positive never matches", got, err, label)
	}
}

// AN EMPTY REOPENED ID NEVER ENTERS THE PROTECTED SET.
//
// `citationLabelsOf` feeds the blue-report lockdown's EXPECTED anchors. An empty string entering
// it makes the lockdown expect an anchor that cannot exist in any document, so every subsequent
// check reports a dropped citation that was never there. The guard is `id != "" && !seen[id]`,
// and a sweep showed neither half was tested alone.
func TestAnEmptyReopenedIDNeverReachesTheProtectedSet(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "blue-respond-r1", Round: 1}
	if _, _, err := RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	// An edit whose reopened list carries a blank alongside a real id, and the real id twice.
	if _, err := Append(id, &recordpb.BlueEdit{
		Old: proto.String("before"), New: proto.String("after"), Text: proto.String("why"),
		Reopened: []string{"", "c-real", "c-real", ""},
	}); err != nil {
		t.Fatal(err)
	}
	b, err := BoardState(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	got := reopenedAnchors(b.Events)
	if len(got) != 1 || got[0] != "c-real" {
		t.Errorf("reopenedAnchors = %#v, want exactly [c-real] — a blank would make the lockdown expect an anchor no document can carry, and a duplicate would count one reopening twice", got)
	}
}
