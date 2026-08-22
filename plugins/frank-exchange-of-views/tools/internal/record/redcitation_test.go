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
	runDir := t.TempDir()
	id := Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: 1}
	if _, _, err := RegisterSeat(id); err != nil {
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
	b, err := BoardState(runDir)
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
	runDir := t.TempDir()
	id := Identity{RunDir: runDir, SeatID: "red-lens-r1-L1", Round: 1}
	if _, _, err := RegisterSeat(id); err != nil {
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

	srcs, err := CitedSources(runDir)
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
	labels, err := CitationLabels(runDir)
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
