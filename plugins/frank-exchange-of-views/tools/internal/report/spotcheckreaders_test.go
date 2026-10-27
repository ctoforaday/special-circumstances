package report

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// A SPOT-CHECK'S REASON REACHES EVERY READER, WHATEVER THE SPOT-CHECK NAMES.
//
// Over a converged board the chair records no verdict and states a defect it found in a stale area
// in that area's spot-check — which names the area and no closure. The board printed such a
// spot-check as "naming no closures" without its prose, and no transcript printed a spot-check at
// all, so the defect reached neither the report's reader nor the bench, whose account of the run's
// ending cites it. Each reader is asked here for both shapes: areas only, and areas with closures.
func TestASpotChecksReasonReachesTheReportAndTheTranscript(t *testing.T) {
	const (
		areasOnly = "the logic area's new paragraph contradicts [^3] the table"
		withIDs   = "G1's anchor still resolves"
	)
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		mintT(t, "red-lens-logic", "G1"),
		closeT(t, "red-lens-logic", "G1"),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{Areas: []string{"red-lens-logic"}, Reason: proto.String(areasOnly)}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{Areas: []string{"red-lens-logic"}, Ids: []string{"G1"}, Reason: proto.String(withIDs)}),
	)
	run := runtest.Open(t, dir)

	docs, err := shippedDocs(run)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]string{}
	for _, d := range docs {
		body[d.File] = d.Body
	}
	// THE REPORT'S DOCUMENTS ESCAPE A SEAT'S `[^`, so nothing a seat wrote opens a footnote.
	for _, file := range []string{FileDocket, FileDebate} {
		for _, want := range []string{
			"#2 (red-chair): read stale area(s) red-lens-logic — " + stdReaders().escapeNotes(areasOnly),
			"#3 (red-chair): read stale area(s) red-lens-logic; re-verified G1 of 1 archived closure(s) — " + withIDs,
		} {
			if !strings.Contains(body[file], want) {
				t.Errorf("%s does not print the spot-check %q:\n%s", file, want, body[file])
			}
		}
		if strings.Contains(body[file], " "+areasOnly) {
			t.Errorf("%s prints a seat's [^ unescaped:\n%s", file, body[file])
		}
	}

	// THE TRANSCRIPT THE BENCH READS, in both of its forms.
	md, err := view.Markdown(run, "debate", "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "### RED SPOT-CHECK\n#2 (red-chair): read stale area(s) red-lens-logic — " + areasOnly; !strings.Contains(string(md), want) {
		t.Errorf("the debate view does not print the areas-only spot-check %q:\n%s", want, md)
	}
	if !strings.Contains(string(md), withIDs) {
		t.Errorf("the debate view does not print the spot-check naming closures:\n%s", md)
	}
	raw, err := record.DebateJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var dj record.DebateJSON
	if err := json.Unmarshal(raw, &dj); err != nil {
		t.Fatal(err)
	}
	got := map[int][]string{}
	for _, ep := range dj.Epochs {
		got[ep.Epoch] = ep.SpotChecks
	}
	if s := got[2]; len(s) != 1 || !strings.Contains(s[0], "red-lens-logic") || !strings.HasSuffix(s[0], areasOnly) {
		t.Errorf("the debate's JSON epoch 2 spot_checks = %q, want the areas-only spot-check with its reason", s)
	}
	if s := got[3]; len(s) != 1 || !strings.Contains(s[0], "re-verified G1 of 1") || !strings.HasSuffix(s[0], withIDs) {
		t.Errorf("the debate's JSON epoch 3 spot_checks = %q, want the spot-check naming G1 with its reason", s)
	}
}
