package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// TestAssembleEndToEnd drives a minimal but real run through the record API and asserts the
// assembled report.md combines the two ownership classes correctly: blue's audited sections
// are lifted verbatim, and the verdict, findings, inquiries and debate are composed from the
// event log. It is the driveable check the unit tests around each composer cannot give.
func TestAssembleEndToEnd(t *testing.T) {
	runDir := newRun(t)

	blue := strings.Join([]string{
		"# Whether the cache is coherent — research report",
		// Blue authors a stale verdict in the preamble — it cannot know the outcome (#79);
		// the embed must strip it rather than park it beside the tool's authoritative stamp.
		"**Verdict:** UNVERIFIED (Round 0)",
		"",
		"## TL;DR",
		"The cache is coherent under the documented invariants; one edge case is unproven.",
		"",
		"## The Catechism",
		"Q1: is it worth our time? Yes — the failure mode is silent corruption.",
		"",
		"## Technical foundations",
		"The invalidation protocol is single-writer.",
		"",
		"## Analysis",
		"From single-writer it follows that reads never observe a torn value.",
		"",
		"## Open questions",
		"Does the eviction path hold under a clock step?",
		"",
		// Blue OVER-AUTHORS a tool-owned section — it must be dropped from the embed
		// (fabrication), not duplicated below the tool's composed version.
		"## Risk Matrix",
		"blue's fabricated risk row that the tool must not echo.",
		"",
		// Blue hand-authors footnotes — DROPPED now (citations are tool-composed from the
		// cite events; a blue-authored bibliography is fabrication).
		"## Footnotes",
		"[^cache]: a citation blue tried to hand-author.",
		"",
	}, "\n")
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}

	// THROUGH THE PRODUCTION WRITE PATH, deliberately: this is the integration test, so it
	// registers and appends exactly as a seat does. The bodies are typed, so `add` takes one
	// instead of alternating key/value strings — and a body the record would refuse now fails
	// here rather than being written and read back as a state no run could reach.
	seen := map[string]bool{}
	add := func(seatID string, body proto.Message) {
		t.Helper()
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID, Round: record.RoundIn(runtest.Open(t, runDir))(seatID)}
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

	// Ingest the round-0 report so AssembleAll renders it (#709): the report is the record
	// projection now, and the marker tokens are already in `blue`, so the marker events below
	// replay as no-ops (skip-if-present).
	add("blue-synthesize", &recordpb.BaseIngest{Text: proto.String(blue)})

	// Red mints a gap; the parties take positions; blue records one pursued line of inquiry (an
	// expansion) and one abandoned line of inquiry (an alternative considered); the bench opines;
	// the run's terminal verdict is recorded.
	add("red-merge-r1", &recordpb.Mint{
		GapId: proto.String("R1-1"), Problem: proto.String("eviction races the reader"),
		Location: proto.String("cache.go:88"), Class: proto.String("correctness"),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
		AcceptanceCheck: proto.String("race the eviction under -race"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		RequiredFix:     proto.String("take the read lock in evict"),
	})
	add("red-merge-r1", &recordpb.Position{Text: proto.String("gap R1-1 stands until the race is shown impossible")})
	add("blue-respond-r1", &recordpb.Position{Text: proto.String("R1-1 is repaired by ordering the invalidation before the store")})
	add("blue-respond-r1", &recordpb.Avenue{
		AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED),
		Line: proto.String("model-check the two-writer interleaving"), Method: proto.String("TLA+"),
	})
	add("blue-respond-r1", &recordpb.Avenue{
		AvenueId: proto.String("Q2"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED),
		Line: proto.String("rewrite the cache lock-free"), Reason: proto.String("cost exceeds the benefit at this scale"),
	})
	add("judge-r1", &recordpb.Opinion{
		GapId: proto.String("R1-1"), Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED),
		Principle: proto.String("correctness"), Tension: proto.String("cost vs certainty"),
		ReviewFlag: proto.String("false"), Settled: proto.String("the claim as it stood may not be re-asserted"),
		Final:     proto.Bool(true),
		Rationale: proto.String("a model-check is owed before this closes"),
	})
	add("judge-terminal", &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
		Prose:   proto.String("the round ceiling arrived before red could pass the final revision"),
	})

	path, err := Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	read := func(name string) string {
		c, rerr := os.ReadFile(filepath.Join(runDir, name))
		if rerr != nil {
			t.Fatalf("the set is missing %s: %v", name, rerr)
		}
		return string(c)
	}
	docket, deb, index := read(FileDocket), read(FileDebate), read(FileIndex)

	for _, want := range []string{
		// Blue-authored, lifted verbatim.
		"# Whether the cache is coherent — research report",
		"## TL;DR\nThe cache is coherent under the documented invariants",
		"Q1: is it worth our time? Yes",
		"From single-writer it follows that reads never observe a torn value.",
		"Does the eviction path hold under a clock step?",
		// Composed from the record.
		"**Verdict:** CEILING-TERMINATED",
		"## Research areas",
		"model-check the two-writer interleaving",
		"## Alternatives considered",
		"rewrite the cache lock-free",
		"cost exceeds the benefit at this scale",
		// The reviewer-facing orientation, composed from the board.
		"## Read this first",
		"(R1-1)",
		// The link bar: the current document named, its siblings linked.
		"**Report** · [Board](docket.md)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report.md missing %q\n---\n%s", want, got)
		}
	}

	// THE SPLIT IS THE POINT: the process record is in the set, and NOT in the research
	// document. A reader who wants the transcript follows a link; one who wants the answer is
	// not handed 70% of a run's telemetry to scroll past.
	for _, want := range []string{"## The board", "R1-1 — eviction races the reader"} {
		if !strings.Contains(docket, want) {
			t.Errorf("docket.md missing %q\n---\n%s", want, docket)
		}
		if strings.Contains(got, want) {
			t.Errorf("report.md still carries the docket (%q) — the split did not happen", want)
		}
	}
	for _, want := range []string{"## The debate", "### RED — NO VERDICT RECORDED THIS ROUND\ngap R1-1 stands", "### BLUE\nR1-1 is repaired", "R1-1: carried"} {
		if !strings.Contains(deb, want) {
			t.Errorf("debate.md missing %q\n---\n%s", want, deb)
		}
		if strings.Contains(got, want) {
			t.Errorf("report.md still carries the debate (%q) — the split did not happen", want)
		}
	}
	// The index is the run directory's front door, and names every document it wrote.
	for _, want := range []string{"**Verdict**", "## The documents", "[Board](docket.md)", "[Debate](debate.md)"} {
		if !strings.Contains(index, want) {
			t.Errorf("README.md missing %q\n---\n%s", want, index)
		}
	}
	// A document with no content is not written at all — this run filed no motions.
	if _, serr := os.Stat(filepath.Join(runDir, FileJudgments)); serr == nil {
		t.Errorf("judgments.md was written for a run with no motions — an empty heading standing in for a document")
	}

	// Blue over-authored a Risk Matrix, a stale verdict, and hand-authored footnotes; none may
	// be echoed. The tool's "## Risk matrix" is authoritative; blue's fabricated row / stale
	// verdict / hand-authored citation must not appear anywhere. With no cite events on the
	// record, no "## Bibliography" is composed either.
	for _, forbidden := range []string{"blue's fabricated risk row", "UNVERIFIED (Round 0)", "a citation blue tried to hand-author", "## Bibliography"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("blue's fabricated/duplicated content leaked into the report (%q)\n---\n%s", forbidden, got)
		}
	}

	// Fate is not crossed: the pursued line of inquiry is an expansion, not an alternative.
	alt := got[strings.Index(got, "## Alternatives considered"):]
	alt = alt[:strings.Index(alt, "## Open questions")]
	if strings.Contains(alt, "model-check the two-writer") {
		t.Errorf("a pursued line of inquiry leaked into Alternatives considered:\n%s", alt)
	}
}

// EVERY DOCUMENT IN THE SET CLOSES ITS OWN FOOTNOTES — the plan's §V end-to-end check, driven.
//
// A footnote definition cannot cross a file boundary. The unit tests prove each weave closes
// what IT emits, over one string; only the assembled SET can show that the anchors actually
// distribute across files and that no document ships a reference it does not define. The
// assertion is deliberately paired with a NON-EMPTINESS check: a scan for dangling references
// over a set that carries none reports a clean board in exactly the bytes it would use for a
// real one, and that plausible zero is what this test exists to refuse.
func TestNoDocumentInTheSetShipsADanglingFootnote(t *testing.T) {
	runDir := newRun(t)
	const sha = "feed0000face1111"
	seedProof(t, runDir, sha, "console.log('races:', 0);", "races: 0\n")

	blue := strings.Join([]string{
		"# Whether the cache is coherent — research report",
		"",
		"## TL;DR",
		"The cache is coherent<!--cite:c-1-->, and the interleaving was model-checked<!--proof:p-1-->.",
		"",
		"## Analysis",
		"Single-writer invalidation is the load-bearing invariant<!--cite:c-1-->.",
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
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID, Round: record.RoundIn(runtest.Open(t, runDir))(seatID)}
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

	// Ingest the round-0 report so AssembleAll renders it (#709): the report is the record
	// projection now, and the marker tokens are already in `blue`, so the marker events below
	// replay as no-ops (skip-if-present).
	add("blue-synthesize", &recordpb.BaseIngest{Text: proto.String(blue)})

	add("blue-synthesize", &recordpb.Cite{
		Label: proto.String("c-1"), Url: proto.String("https://ex/coherence"),
		Sha256: proto.String("deadbeef"), Title: proto.String("Coherence Proof"),
		AccessDate: proto.String("2026-08-03"),
	})
	// A SECOND source, cited only from red's board — so its reference lands in docket.md and
	// nowhere else. This is the case a global weave gets wrong and a per-file weave gets right.
	add("blue-synthesize", &recordpb.Cite{
		Label: proto.String("c-2"), Url: proto.String("https://ex/eviction"),
		Sha256: proto.String("beefcafe"), Title: proto.String("Eviction Under Contention"),
		AccessDate: proto.String("2026-08-04"),
	})
	add("blue-synthesize", &recordpb.Proof{
		ProofId: proto.String("p-1"), ProofSha: proto.String(sha),
		ProofBasis: proto.String("reproducible"), Script: proto.String("interleave.js"),
		Text: proto.String("the model check settles the race"),
	})
	add("red-merge-r1", &recordpb.Mint{
		GapId: proto.String("R1-1"), Problem: proto.String("eviction races the reader<!--cite:c-2-->"),
		Location: proto.String("cache.go:88"), Class: proto.String("correctness"),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
		AcceptanceCheck: proto.String("race the eviction under -race"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		RequiredFix:     proto.String("take the read lock in evict"),
	})
	// And a proof anchored from the transcript, so debate.md must define P1 for itself too.
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

	// `\[\^label\]` is a REFERENCE; the same shape followed by a colon at line start is its
	// DEFINITION. One pattern, told apart by the colon, because RE2 has no lookahead.
	fn := regexp.MustCompile(`\[\^([^\]]+)\]:?`)
	entries, err := os.ReadDir(runDir)
	if err != nil {
		t.Fatal(err)
	}
	carrying := map[string]int{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		body, rerr := os.ReadFile(filepath.Join(runDir, e.Name()))
		if rerr != nil {
			t.Fatal(rerr)
		}
		refs, defs := map[string]bool{}, map[string]bool{}
		for _, m := range fn.FindAllStringSubmatch(string(body), -1) {
			if strings.HasSuffix(m[0], ":") {
				defs[m[1]] = true
				continue
			}
			refs[m[1]] = true
		}
		for label := range refs {
			if !defs[label] {
				t.Errorf("%s references [^%s] and does not define it — a footnote definition cannot cross a file boundary", e.Name(), label)
			}
		}
		carrying[e.Name()] = len(refs)
	}

	// THE PLAUSIBLE-ZERO GUARD. Anchors must actually have reached more than one document,
	// or the loop above proved nothing at all.
	if carrying[FileReport] == 0 {
		t.Errorf("no footnote reference reached %s — the scan above was vacuous:\n%v", FileReport, carrying)
	}
	spread := 0
	for _, n := range carrying {
		if n > 0 {
			spread++
		}
	}
	if spread < 2 {
		t.Errorf("footnote references reached only %d document(s) — the cross-file case this test exists for was never exercised:\n%v", spread, carrying)
	}
}
