package reportproj_test

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

// The marker verbs (cite/compute/finding) record their act and then an Anchor, and the projection
// re-places the marker from the Anchor alone. This is the fidelity guard for that path: a base with
// a citation, a computation and a gap replays to exactly what write-time InsertAnchor would have
// produced, markers in record order — and an act recorded with no Anchor places nothing.
func TestRenderRunReplaysMarkerInsertions(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "blue-respond", "red-lens-evidence"} {
		if _, _, err := record.RegisterSeat(ident(t, runDir, s), "", ""); err != nil {
			t.Fatalf("register %s: %v", s, err)
		}
	}
	base := "The estimate is stable across the range. Demand grows steadily over the period."
	if _, err := record.Append(ident(t, runDir, "blue-synthesize"), &recordpb.ReportPart{Text: proto.String(base)}); err != nil {
		t.Fatalf("append base: %v", err)
	}

	const citeID, computationID, gapID = "C-cafef00d", "K-12345678", "G-deadbeef"
	// cite anchors the first sentence, computation the first too (two markers, one sentence), the gap
	// the second. Recorded in this order; replay must honour it.
	if _, err := record.Append(ident(t, runDir, "blue-respond"), &recordpb.Cite{FetchedAt: proto.String("2026-11-12T00:00:00Z"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), SourceSpan: proto.String("s"), SpanStart: proto.Int32(0), SpanEnd: proto.Int32(1), SpanTextSha: proto.String("t"), SourceMatched: proto.String("s"), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
		Label: proto.String(citeID), Location: proto.String("The estimate is stable across the range"),
	}); err != nil {
		t.Fatalf("append cite: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "blue-respond"), &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
		ComputationId: proto.String(computationID), BundleSha: proto.String("sha-" + computationID), Location: proto.String("The estimate is stable across the range"),
	}); err != nil {
		t.Fatalf("append computation: %v", err)
	}
	if got, err := record.RenderRun(runtest.Open(t, runDir)); err != nil || got != base {
		t.Fatalf("a cite and a computation with no Anchor event rendered a marker: %q, %v", got, err)
	}
	for _, id := range []string{citeID, computationID} {
		if _, err := record.Append(ident(t, runDir, "blue-respond"), &recordpb.Anchor{
			Id: proto.String(id), Location: proto.String("The estimate is stable across the range"),
		}); err != nil {
			t.Fatalf("append the anchor of %s: %v", id, err)
		}
	}
	// Every placing act records BOTH events; only the Anchor carries the marker geometry.
	if _, err := record.Append(ident(t, runDir, "red-lens-evidence"), &recordpb.Mint{
		GapId: proto.String(gapID), Location: proto.String("Demand grows steadily over the period"),
		Problem: proto.String("the growth is asserted"), RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("the check runs"),
		Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity: recordpb.Grade_GRADE_LOW.Enum(), Likelihood: recordpb.Grade_GRADE_LOW.Enum(), Impact: recordpb.Grade_GRADE_LOW.Enum(),
	}); err != nil {
		t.Fatalf("append mint: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "red-lens-evidence"), &recordpb.Anchor{
		Id: proto.String(gapID), Location: proto.String("Demand grows steadily over the period"),
	}); err != nil {
		t.Fatalf("append anchor: %v", err)
	}

	got, err := record.RenderRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("RenderRun: %v", err)
	}

	want := base
	want = insertHelper(t, want, "The estimate is stable across the range", anchor.Token(citeID))
	want = insertHelper(t, want, "The estimate is stable across the range", anchor.Token(computationID))
	want = insertHelper(t, want, "Demand grows steadily over the period", anchor.Token(gapID))
	if got != want {
		t.Fatalf("marker replay drifted from write-time placement:\n  want %q\n  got  %q", want, got)
	}
	// AN ID IS PLACED ONCE ON A RUN: a second Anchor for it, from any seat, is refused by the record.
	if _, err := record.Append(ident(t, runDir, "red-lens-evidence"), &recordpb.Anchor{
		Id: proto.String(citeID), Location: proto.String("Demand grows steadily over the period"),
	}); err == nil || !strings.Contains(err.Error(), "UNIQUE constraint failed: anchor.id") {
		t.Errorf("a second anchor for %s was not refused as a duplicate id: %v", citeID, err)
	}
	// Every marker present, exactly once.
	for _, id := range []string{citeID, computationID, gapID} {
		if n := strings.Count(got, anchor.Token(id)); n != 1 {
			t.Errorf("marker %s appears %d times, want 1", anchor.Token(id), n)
		}
	}
}

// A quoted gap with NO paired Anchor inserts nothing: the marker geometry rides the Anchor, and
// replaying the mint too would double it.
func TestRenderRunDoesNotInsertFromAGapAlone(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, s := range []string{"blue-synthesize", "red-lens-evidence"} {
		if _, _, err := record.RegisterSeat(ident(t, runDir, s), "", ""); err != nil {
			t.Fatalf("register %s: %v", s, err)
		}
	}
	base := "Demand grows steadily over the period."
	if _, err := record.Append(ident(t, runDir, "blue-synthesize"), &recordpb.ReportPart{Text: proto.String(base)}); err != nil {
		t.Fatalf("append base: %v", err)
	}
	if _, err := record.Append(ident(t, runDir, "red-lens-evidence"), &recordpb.Mint{
		GapId: proto.String("G-deadbeef"), Location: proto.String("Demand grows steadily over the period"),
		Problem: proto.String("the growth is asserted"), RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("the check runs"),
		Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity: recordpb.Grade_GRADE_LOW.Enum(), Likelihood: recordpb.Grade_GRADE_LOW.Enum(), Impact: recordpb.Grade_GRADE_LOW.Enum(),
	}); err != nil {
		t.Fatalf("append mint: %v", err)
	}
	got, err := record.RenderRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("RenderRun: %v", err)
	}
	if got != base {
		t.Fatalf("a gap with no Anchor changed the report:\n  base %q\n  got  %q", base, got)
	}
}
