package report

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// NO RAW ANCHOR REACHES A READER, IN ANY DOCUMENT OF THE SET.
//
// Each of the three anchor classes has a DEFINED FATE at assembly, and they are three different
// fates decided in three different places:
//
//	<!--fx:f-…-->     STRIPPED           (StripFindingMarkers, docs.go)
//	<!--cite:c-…-->   -> [^N]  + bibliography   (weaveCitations, assemble.go)
//	<!--proof:p-…--> -> [^PN] + definitions    (weaveProofRefs, proofs.go)
//
// What was tested was each fate ALONE: FuzzWeaveCitations drives weaveCitations over one string,
// markers_test drives StripFindingMarkers over one string, and TestNoDocumentInTheSetShipsA
// DanglingFootnote asks the set a different question — whether every footnote REFERENCE is
// defined. A raw anchor is not a footnote reference, so that scan cannot see one.
//
// The gap that leaves is the shape #590 was filed about, one level over. docOrder carries SEVEN
// documents and the weaves are applied PER DOCUMENT. A document whose composition path misses one
// of them — an eighth document added to the set, a section routed around the weave — ships the raw
// token. And an HTML comment is INVISIBLE in rendered markdown, so the failure is silent in
// exactly the way the assembled report's dangling `[^P1]` was loud: the reader sees prose with a
// hole where the evidence layer should be, and nothing anywhere says so.
//
// This asks the question of the ARTIFACT rather than of the three functions: after Assemble, does
// any document in the run carry a token of any class?
func TestNoDocumentInTheSetShipsARawAnchor(t *testing.T) {
	runDir := newRun(t)
	const sha = "feed0000face2222"
	seedProof(t, runDir, sha, "console.log('interleavings:', 0);", "interleavings: 0\n")

	// ALL THREE CLASSES, IN BLUE'S AUDITED REPORT — which is the surface lifted VERBATIM, so it
	// is the one where a missed weave would show up first.
	blue := strings.Join([]string{
		"# Whether the eviction path is safe — research report",
		"",
		"## TL;DR",
		"The eviction path is safe<!--cite:c-1-->, and the interleaving was model-checked<!--proof:p-1-->.",
		"",
		"## Analysis",
		"Red's objection to the read lock was withdrawn<!--fx:f-L1-F1--> after the counter-example<!--cite:c-1-->.",
		"",
	}, "\n")
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	add := func(seatID string, body proto.Message) {
		t.Helper()
		run := runtest.Open(t, runDir)
		id := record.Identity{Run: run, SeatID: seatID, Round: record.RoundIn(run)(seatID)}
		if !seen[seatID] {
			if _, _, err := record.RegisterSeat(id, ""); err != nil {
				t.Fatalf("register %s: %v", seatID, err)
			}
			seen[seatID] = true
		}
		if _, err := record.Append(id, body); err != nil {
			t.Fatalf("append %s/%T: %v", seatID, body, err)
		}
	}

	add("blue-synthesize", &recordpb.Cite{
		Label: proto.String("c-1"), Url: proto.String("https://ex/eviction"),
		Sha256: proto.String("deadbeef"), Title: proto.String("Eviction Under Contention"),
		AccessDate: proto.String("2026-09-05"),
	})
	add("blue-synthesize", &recordpb.Proof{
		ProofId: proto.String("p-1"), ProofSha: proto.String(sha),
		ProofBasis: proto.String("reproducible"), Script: proto.String("interleave.js"),
		Text: proto.String("the model check settles the interleaving"),
	})
	// ANCHORS IN RECORD-DERIVED TEXT TOO, not only in blue's lifted sections. A seat's own prose
	// carries tokens into the findings and transcript sections, which are composed from the event
	// log rather than copied — a different code path to the same page, and the reason
	// StripFindingMarkers runs over the FINAL output rather than over blue's content alone.
	add("red-lens-r1-L1", &recordpb.Finding{
		Label: proto.String("L1-F1"), Location: proto.String("§Analysis"),
		Text: proto.String("the read lock is dropped before evict<!--cite:c-1-->"),
	})
	add("red-merge-r1", &recordpb.Mint{
		GapId: proto.String("R1-1"), Problem: proto.String("eviction races the reader<!--fx:f-L1-F1-->"),
		Location: proto.String("cache.go:88"), Class: proto.String("correctness"),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
		AcceptanceCheck: proto.String("race the eviction under -race"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		RequiredFix:     proto.String("take the read lock in evict"),
	})
	add("blue-respond-r1", &recordpb.Position{
		Text: proto.String("the interleaving is model-checked<!--proof:p-1--> and R1-1 does not stand"),
	})
	add("judge-terminal", &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
		Prose:   proto.String("the round ceiling arrived before red could pass the final revision"),
	})

	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	// EVERY *.md IN THE RUN, not reportdoc.Files(). A document dropped from docOrder but still
	// written would be exactly the carrier this test exists to find, and asking the list would
	// hide it — the list is the thing under suspicion.
	anchorTok := regexp.MustCompile(`<!--(fx|cite|proof):[^>]*-->`)
	entries, err := os.ReadDir(runDir)
	if err != nil {
		t.Fatal(err)
	}
	scanned, withAnchors := 0, 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		body, rerr := os.ReadFile(filepath.Join(runDir, e.Name()))
		if rerr != nil {
			t.Fatal(rerr)
		}
		scanned++
		if raw := anchorTok.FindAllString(string(body), -1); len(raw) > 0 {
			withAnchors++
			t.Errorf("%s ships %d raw anchor token(s): %s\n"+
				"Each class has a fate at assembly — fx is stripped, cite becomes [^N], proof becomes [^PN] — "+
				"and a token that survives means this document's composition path missed one. An HTML comment "+
				"renders as nothing, so a reader sees prose with a hole where the evidence layer should be.",
				e.Name(), len(raw), strings.Join(raw, ", "))
		}
	}

	// THE PLAUSIBLE-ZERO GUARD, and this test needs two of them.
	//
	// The scan reports zero raw anchors when assembly is healthy AND when it wrote nothing at all,
	// and those are the same bytes. So: documents must exist, and — the one that actually bites —
	// the fixture's anchors must have REACHED them. A fixture whose tokens all landed in sections
	// the assembler happens to omit would prove nothing while reading green forever.
	if scanned == 0 {
		t.Fatal("no .md document was written — the scan above was vacuous, not clean")
	}
	report, err := os.ReadFile(filepath.Join(runDir, FileReport))
	if err != nil {
		t.Fatalf("read %s: %v", FileReport, err)
	}
	for _, want := range []struct{ what, probe string }{
		{"a citation woven to a footnote reference", "[^1]"},
		{"a proof woven to a footnote reference", "[^P1]"},
		{"the sentence a stripped finding-marker was attached to", "Red's objection to the read lock was withdrawn"},
	} {
		if !strings.Contains(string(report), want.probe) {
			t.Errorf("%s never reached %s (looked for %q) — the fixture's anchors did not travel, "+
				"so a green scan says nothing about the weaves:\n%s",
				want.what, FileReport, want.probe, truncateForLog(string(report)))
		}
	}
}

func truncateForLog(s string) string {
	const max = 1500
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n…(truncated)"
}
