package reportproj

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// insertHelper mirrors write-time placement so the test's oracle is the SAME transform the verbs
// apply — a marker spliced at the quote's end via anchortext.InsertAnchor.
func insertHelper(t *testing.T, text, location, marker string) string {
	t.Helper()
	next, err := anchortext.InsertAnchor([]byte(text), location, marker)
	if err != nil {
		t.Fatalf("InsertAnchor(%q, %q): %v", location, marker, err)
	}
	return string(next)
}

// The marker verbs (cite/prove/finding) no longer write report.md — they record an event, and the
// projection re-places their marker. This is the fidelity guard for that path: a base with a
// citation, a proof, and a finding-anchor replays to exactly what write-time InsertAnchor would
// have produced, markers in record order.
func TestRenderFromRecordReplaysMarkerInsertions(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "blue-respond-r1", "red-lens-r1-evidence"} {
		if _, _, err := record.RegisterSeat(ident(t, runDir, s), ""); err != nil {
			t.Fatalf("register %s: %v", s, err)
		}
	}
	base := "The estimate is stable across the range. Demand grows steadily over the period."
	if _, err := record.Append(ident(t, runDir, "blue-synthesize"), &recordpb.BaseIngest{Text: proto.String(base)}); err != nil {
		t.Fatalf("append base: %v", err)
	}

	const citeID, proofID, findID = "c-cafef00d", "p-12345678", "f-deadbeef"
	// cite anchors the first sentence, proof the first too (two markers, one sentence), finding
	// the second. Recorded in this order; replay must honour it.
	if _, err := record.Append(ident(t, runDir, "blue-respond-r1"), &recordpb.Cite{
		Label: proto.String(citeID), Location: proto.String("The estimate is stable across the range"),
	}); err != nil {
		t.Fatalf("append cite: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "blue-respond-r1"), &recordpb.Proof{
		ProofId: proto.String(proofID), Location: proto.String("The estimate is stable across the range"),
	}); err != nil {
		t.Fatalf("append proof: %v", err)
	}
	// A finding records BOTH events; only the Anchor carries the marker geometry.
	if _, err := record.Append(ident(t, runDir, "red-lens-r1-evidence"), &recordpb.Finding{
		Label: proto.String("L1-F1"), FindingId: proto.String(findID), Location: proto.String("Demand grows steadily over the period"),
	}); err != nil {
		t.Fatalf("append finding: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "red-lens-r1-evidence"), &recordpb.Anchor{
		Id: proto.String(findID), Location: proto.String("Demand grows steadily over the period"),
	}); err != nil {
		t.Fatalf("append anchor: %v", err)
	}

	got, err := RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("RenderFromRecord: %v", err)
	}

	want := base
	want = insertHelper(t, want, "The estimate is stable across the range", anchor.Token(citeID))
	want = insertHelper(t, want, "The estimate is stable across the range", anchor.Token(proofID))
	want = insertHelper(t, want, "Demand grows steadily over the period", anchor.Token(findID))
	if got != want {
		t.Fatalf("marker replay drifted from write-time placement:\n  want %q\n  got  %q", want, got)
	}
	// Every marker present, exactly once.
	for _, id := range []string{citeID, proofID, findID} {
		if n := strings.Count(got, anchor.Token(id)); n != 1 {
			t.Errorf("marker %s appears %d times, want 1", anchor.Token(id), n)
		}
	}
}

// A Finding event with NO paired Anchor inserts nothing: the marker geometry rides the Anchor, and
// replaying the Finding too would double it. This guards the deliberate skip of Finding in replay.
func TestRenderFromRecordDoesNotInsertFromFindingAlone(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "red-lens-r1-evidence"} {
		if _, _, err := record.RegisterSeat(ident(t, runDir, s), ""); err != nil {
			t.Fatalf("register %s: %v", s, err)
		}
	}
	base := "Demand grows steadily over the period."
	if _, err := record.Append(ident(t, runDir, "blue-synthesize"), &recordpb.BaseIngest{Text: proto.String(base)}); err != nil {
		t.Fatalf("append base: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "red-lens-r1-evidence"), &recordpb.Finding{
		Label: proto.String("L1-F1"), FindingId: proto.String("f-deadbeef"), Location: proto.String("Demand grows steadily over the period"),
	}); err != nil {
		t.Fatalf("append finding: %v", err)
	}
	got, err := RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("RenderFromRecord: %v", err)
	}
	if got != base {
		t.Fatalf("a Finding with no Anchor changed the report:\n  base %q\n  got  %q", base, got)
	}
}

// The replay analog of the verbs' torn-anchor adoption: applying an insertion whose marker is
// ALREADY in the running text is a no-op, not a second splice. A marker can be named twice in the
// record across a crash-retry re-dispatch (cite adopts the orphan and re-appends under the same
// label in a NEW sitting, past the once-per-sitting guard), so replay must place it once. Driven on
// insertMut directly — the once-per-sitting guard blocks constructing the pair through the record,
// but the skip is what makes replay robust when the pair nonetheless exists.
func TestInsertMutSkipsAMarkerAlreadyPlaced(t *testing.T) {
	const location = "Demand grows steadily over the period"
	marker := anchor.Token("c-cafef00d")
	base := location + "."

	once, err := (insertMut{location: location, marker: marker}).apply(base)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if strings.Count(once, marker) != 1 {
		t.Fatalf("first insert did not place the marker exactly once:\n  %q", once)
	}
	twice, err := (insertMut{location: location, marker: marker}).apply(once)
	if err != nil {
		t.Fatalf("second insert: %v", err)
	}
	if twice != once {
		t.Errorf("re-applying an already-placed marker changed the text:\n  once  %q\n  twice %q", once, twice)
	}
}
