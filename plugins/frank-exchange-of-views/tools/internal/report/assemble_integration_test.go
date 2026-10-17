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
// ship as written, and the verdict, findings, inquiries and debate are composed from the
// event log. It is the driveable check the unit tests around each composer cannot give.
func TestAssembleEndToEnd(t *testing.T) {
	runDir := newRun(t)

	blue := strings.Join([]string{
		"# Whether the cache is coherent — research report",
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
		"## Appendix: eviction traces",
		"Three traces of the eviction path.",
		"",
		"## Open questions",
		"Does the eviction path hold under a clock step?",
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
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID}
		if !seen[seatID] {
			// The bench owes an OCCASION and nobody else may pass one — ask the write path's own
			// question rather than keeping a list of which ids are the bench.
			occasion := ""
			if record.SeatOwesOccasion(seatID) {
				occasion = "docket"
			}
			if _, _, err := record.RegisterSeat(id, "", occasion); err != nil {
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

	// Red mints a gap; the parties take positions; blue records one underway inquiry (an
	// expansion) and one abandoned inquiry (an alternative considered); the bench opines;
	// the run's terminal verdict is recorded.
	add("red-chair", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		GapId: proto.String("G1"), Problem: proto.String("eviction races the reader"),
		Location: proto.String("cache.go:88"), Class: proto.String("correctness"),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
		AcceptanceCheck: proto.String("race the eviction under -race"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		RequiredFix:     proto.String("take the read lock in evict"),
	})
	add("red-chair", &recordpb.Position{Text: proto.String("gap G1 stands until the race is shown impossible")})
	add("blue-respond", &recordpb.Position{Text: proto.String("G1 is repaired by ordering the invalidation before the store")})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED),
		Question: proto.String("model-check the two-writer interleaving"),
	})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
		Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Method: proto.String("TLA+"),
	})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q2"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED),
		Question: proto.String("rewrite the cache lock-free"), Reason: proto.String("cost exceeds the benefit at this scale"),
	})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q3"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED),
		Question: proto.String("measure the eviction path under a stepped clock"), Reason: proto.String("needs a harness this question does not justify"),
	})
	// THE BENCH OPINES IN TWO ACTS, and through the production write path both are refusable:
	// red DOCKETS the gap it cannot settle, and the bench RULES on that filing. The gap rides
	// the filing, so the ruling alone would be a disposition of nothing.
	add("red-chair", &recordpb.Motion{
		MotionId: proto.String("M1"),
		Basis:    proto.String("red cannot settle G1 without the model-check"),
		Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
	})
	add("judge", &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Opinion:  proto.String("a model-check is owed before this closes"),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED),
			Principle:   proto.String("correctness"), Tension: proto.String("cost vs certainty"),
			ReviewFlag: proto.String("false"),
			Settled:    proto.String("the claim as it stood may not be re-asserted"),
			Direction:  proto.String("a model-check of the two-writer interleaving"),
		}},
	})
	add("judge", &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
		Prose:   proto.String("the round ceiling arrived before red could pass the final revision"),
	})

	a, err := Assemble(runtest.Open(t, runDir))
	path := a.Report
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
		// Blue-authored, shipped as written.
		"# Whether the cache is coherent — research report",
		"## TL;DR\nThe cache is coherent under the documented invariants",
		"Q1: is it worth our time? Yes",
		"From single-writer it follows that reads never observe a torn value.",
		"Does the eviction path hold under a clock step?",
		// Composed from the record.
		"**Outcome:** CEILING-TERMINATED",
		"## Research areas",
		"model-check the two-writer interleaving",
		"**Fate:** underway. **Method:** TLA+.",
		"## Alternatives considered",
		"rewrite the cache lock-free",
		"cost exceeds the benefit at this scale",
		// The link bar: the current document named, its siblings linked.
		"**Report** · [Board](docket.md)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report.md missing %q\n---\n%s", want, got)
		}
	}

	// AND THE BOARD'S OWN VOCABULARY STAYS OFF IT. "Read this first" used to open the research
	// document with ranked board rows and the bench's certification prose — process, in front of
	// a reader who came for the subject. The section is gone; a gap id appearing here again would
	// mean it came back by another route.
	for _, never := range []string{"## Read this first", "(G1)"} {
		if strings.Contains(got, never) {
			t.Errorf("report.md carries %q — the board's vocabulary is docket.md's, not the research document's", never)
		}
	}

	// THE SPLIT IS THE POINT: the process record is in the set, and NOT in the research
	// document. A reader who wants the transcript follows a link; one who wants the answer is
	// not handed 70% of a run's telemetry to scroll past.
	for _, want := range []string{"## The board", "G1 — eviction races the reader"} {
		if !strings.Contains(docket, want) {
			t.Errorf("docket.md missing %q\n---\n%s", want, docket)
		}
		if strings.Contains(got, want) {
			t.Errorf("report.md still carries the docket (%q) — the split did not happen", want)
		}
	}
	for _, want := range []string{"## The debate", "### RED — NO VERDICT RECORDED THIS EPOCH\ngap G1 stands", "### BLUE\nG1 is repaired", "G1: remanded"} {
		if !strings.Contains(deb, want) {
			t.Errorf("debate.md missing %q\n---\n%s", want, deb)
		}
		if strings.Contains(got, want) {
			t.Errorf("report.md still carries the debate (%q) — the split did not happen", want)
		}
	}
	// The index is the run directory's front door, and names every document it wrote.
	for _, want := range []string{"**Outcome**", "## The documents", "[Board](docket.md)", "[Debate](debate.md)"} {
		if !strings.Contains(index, want) {
			t.Errorf("README.md missing %q\n---\n%s", want, index)
		}
	}
	// THE BENCH'S DISPOSITION IS A MOTION NOW, so this run DOES file one and judgments.md is
	// written. That inverts what this assertion used to say — the old fixture wrote a `bench
	// opinion`, which was not a motion, so the judgments document was legitimately empty.
	judgments := read(FileJudgments)
	for _, want := range []string{"## Motions", "**M1** · docket G1", "**ruled remanded** by judge"} {
		if !strings.Contains(judgments, want) {
			t.Errorf("judgments.md missing %q — the bench's disposition is a docket motion, and the "+
				"section whose promise is \"an ask and its answer are one row\" must carry it\n---\n%s", want, judgments)
		}
	}

	// THE INQUIRIES SHIP. report.md files each inquiry by its fate and nothing about how it got
	// there; the path, the seat that last recorded its fate and red's ruling are debate, not report
	// text, so they must land in a document a reader of the archived set can open — or they are
	// data dead in the record. Q2 was recorded straight as `abandoned`, with no `underway` step, and
	// the report's tag is the fate its seat recorded.
	inquiryDoc := read(FileInquiries)
	for _, want := range []string{"## underway (1)", "## abandoned (1)", "Q1 model-check the two-writer interleaving", "(blue-respond)"} {
		if !strings.Contains(inquiryDoc, want) {
			t.Errorf("inquiries.md missing %q\n---\n%s", want, inquiryDoc)
		}
	}
	if !strings.Contains(got, "- **Question:** rewrite the cache lock-free\n  - **Fate:** abandoned. **Why:** cost exceeds the benefit at this scale") {
		t.Errorf("an inquiry abandoned with no underway step carries the plain fate in report.md:\n%s", got)
	}
	if !strings.Contains(index, "[Inquiries](inquiries.md)") {
		t.Errorf("README.md does not name inquiries.md — a document the index omits is one nobody opens\n---\n%s", index)
	}
	// The link bar is built over the documents actually written, so it reaches the new one.
	if !strings.Contains(got, "[Inquiries](inquiries.md)") {
		t.Errorf("report.md's link bar does not reach inquiries.md\n---\n%s", got)
	}
	// EVERY DOCUMENT WRITTEN IS IN THE SET THE OUTSIDE READERS ASK FOR. Files() — docOrder — is what
	// the capture screens, the archive and the fuzz oracle read, and what Write sweeps for stale
	// copies. A document composed but missing from it ships, is never cleaned up when a later
	// assembly has nothing for it, and is invisible to every check outside this package. Asked of
	// the directory, so the next document added is held too, not only this one.
	inSet := map[string]bool{}
	for _, f := range Files() {
		inSet[f] = true
	}
	written, _ := filepath.Glob(filepath.Join(runDir, "*.md"))
	for _, p := range written {
		if name := filepath.Base(p); name != FileIndex && !inSet[name] {
			t.Errorf("%s was written but is not in Files() — outside readers never see it and a stale copy is never removed", name)
		}
	}

	// A document with no content is still not written at all — this run ran no computations, so
	// the evidence document has no body. The property this pins has not moved; only the document
	// that demonstrates it, because judgments.md is no longer empty.
	if _, serr := os.Stat(filepath.Join(runDir, FileEvidence)); serr == nil {
		t.Errorf("evidence.md was written for a run with no proofs — an empty heading standing in for a document")
	}

	// With no cite events on the record, no "## Bibliography" is composed.
	if strings.Contains(got, "## Bibliography") {
		t.Errorf("a bibliography was composed for a run that cites nothing\n---\n%s", got)
	}

	// EVERY SECTION ASSEMBLY SHIPS HAS A ROW, AND SHIPS IN ITS ROW'S PLACE. Read off the assembled
	// documents: a "## " heading in any of them is a row's, or stands in the report's own text.
	// inquiries.md is not read: it is `show inquiries`' rendering, under headings that carry a count.
	rows := map[string]bool{}
	for _, s := range layout {
		rows[s.File+"\x00"+normalizeHeading(s.Heading)] = true
		for _, h := range s.first {
			rows[s.File+"\x00"+normalizeHeading(h)] = true
		}
	}
	shipped := map[string]bool{}
	for _, f := range Files() {
		b, err := os.ReadFile(filepath.Join(runDir, f))
		if err != nil || f == FileInquiries {
			continue
		}
		eachLine(string(b), func(ln, tr string, fenced bool) {
			key, ok := headingKey(tr)
			if !ok || fenced {
				return
			}
			shipped[f+"\x00"+key] = true
			if !rows[f+"\x00"+key] && !strings.Contains(blue, ln) {
				t.Errorf("%s ships %q and the layout has no row for it: the write does not refuse it, the layout view does not list it, and no help page names it", f, ln)
			}
		})
	}
	for _, s := range layout {
		// The fixture runs no computation, cites and retires nothing, certifies once, and its title
		// is short.
		if s.File == FileEvidence || s.File == FileChangelog || s.Heading == "Bibliography" || s.Line != "" || s.embeds {
			continue
		}
		if !shipped[s.File+"\x00"+normalizeHeading(s.Heading)] {
			t.Errorf("the fixture does not compose %q into %s, so the check above did not read it", s.Heading, s.File)
		}
	}
	// In the row's place: report.md's headings are the table's, in the table's order.
	var order, wantOrder []string
	eachLine(got, func(_, tr string, fenced bool) {
		if key, ok := headingKey(tr); ok && !fenced && rows[FileReport+"\x00"+key] {
			order = append(order, key)
		}
	})
	for _, s := range layout {
		if s.File == FileReport && s.Line == "" && s.Heading != "Bibliography" {
			for _, h := range s.first {
				wantOrder = append(wantOrder, normalizeHeading(h))
			}
			if !s.embeds {
				wantOrder = append(wantOrder, normalizeHeading(s.Heading))
			}
		}
	}
	if strings.Join(order, "|") != strings.Join(wantOrder, "|") {
		t.Errorf("report.md's sections ship as %v, and the table's order is %v", order, wantOrder)
	}
	// THE REPORT'S OWN SECTIONS ARE THE BODY: in the order written, under their author's headings
	// and no other, after the Catechism and before the first section the tool composes; and the
	// section that closes the text ships after the tool's.
	body := "## Technical foundations\nThe invalidation protocol is single-writer.\n\n## Analysis\nFrom single-writer it follows that reads never observe a torn value.\n\n## Appendix: eviction traces\nThree traces of the eviction path.\n\n## Risk matrix\n"
	if at := strings.Index(got, body); at < 0 || at < strings.Index(got, "## The Catechism") {
		t.Errorf("the report's own sections do not ship as the body, whole and in the order written, between the Catechism and the risk matrix:\n%s", got)
	}
	for _, not := range []string{"## Additional findings", "not authored", "blue/report.md"} {
		if strings.Contains(got, not) {
			t.Errorf("report.md carries %q: a heading or a line of the tool's about the report's own text", not)
		}
	}

	// Fate is not crossed: the underway inquiry is an expansion, not an alternative.
	alt := got[strings.Index(got, "## Alternatives considered"):]
	alt = alt[:strings.Index(alt, "## Open questions")]
	if strings.Contains(alt, "model-check the two-writer") {
		t.Errorf("an underway inquiry leaked into Alternatives considered:\n%s", alt)
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
		"The cache is coherent<!--cite:C-00000001-->, and the interleaving was model-checked<!--proof:P-00000001-->.",
		"",
		"## Analysis",
		"Single-writer invalidation is the load-bearing invariant<!--cite:C-00000001-->.",
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
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID}
		if !seen[seatID] {
			// The bench owes an OCCASION and nobody else may pass one — ask the write path's own
			// question rather than keeping a list of which ids are the bench.
			occasion := ""
			if record.SeatOwesOccasion(seatID) {
				occasion = "docket"
			}
			if _, _, err := record.RegisterSeat(id, "", occasion); err != nil {
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

	add("blue-synthesize", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(),
		Label: proto.String("C-00000001"), Url: proto.String("https://ex/coherence"),
		Sha256: proto.String("deadbeef"), Title: proto.String("Coherence Proof"),
		AccessDate: proto.String("2026-08-03"),
	})
	// A SECOND source, cited only from red's board — so its reference lands in docket.md and
	// nowhere else. This is the case a global weave gets wrong and a per-file weave gets right.
	add("blue-synthesize", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(),
		Label: proto.String("C-00000002"), Url: proto.String("https://ex/eviction"),
		Sha256: proto.String("beefcafe"), Title: proto.String("Eviction Under Contention"),
		AccessDate: proto.String("2026-08-04"),
	})
	add("blue-synthesize", &recordpb.Proof{
		ProofId: proto.String("P-00000001"), ProofSha: proto.String(sha),
		ProofBasis: proto.String("reproducible"), Script: proto.String("interleave.js"),
		Text: proto.String("the model check settles the race"),
	})
	add("red-chair", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		GapId: proto.String("G1"), Problem: proto.String("eviction races the reader<!--cite:C-00000002-->"),
		Location: proto.String("cache.go:88"), Class: proto.String("correctness"),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
		AcceptanceCheck: proto.String("race the eviction under -race"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		RequiredFix:     proto.String("take the read lock in evict"),
	})
	// And a proof anchored from the transcript, so debate.md must define P1 for itself too.
	add("blue-respond", &recordpb.Position{
		Text: proto.String("the interleaving is model-checked<!--proof:P-00000001--> and G1 does not stand"),
	})
	add("judge", &recordpb.Outcome{
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

// A REPORT THE RECORD HOLDS AND CANNOT REPLAY FAILS THE ASSEMBLY. Assembled as though blue wrote
// nothing, the set would ship without a word of the report, under a title that says there is
// none. A run with no report ingested is the other case, and assembles with none of blue's
// sections: a section nobody wrote ships nothing, and no line stands in the document for it.
func TestAssemblyRefusesAReportItCannotReplay(t *testing.T) {
	runDir := newRun(t)
	run := runtest.Open(t, runDir)
	id := record.Identity{Run: run, SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	docs, err := AssembleAll(run)
	if err != nil {
		t.Fatalf("a run with no report ingested does not assemble: %v", err)
	}
	for _, not := range []string{"TL;DR", "The Catechism", "Open questions", "not authored", "blue/report.md"} {
		if strings.Contains(docs[0].Body, not) {
			t.Errorf("a run with no report ingested ships %q in report.md: a line for a section nobody wrote\n%s", not, docs[0].Body)
		}
	}
	if !strings.Contains(docs[0].Body, "## Risk matrix") {
		t.Errorf("a run with no report ingested does not ship the sections the tool composes:\n%s", docs[0].Body)
	}

	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String("# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n")}); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(id, &recordpb.BlueEdit{EditKey: proto.String("e1"), Old: proto.String("a span the report never held"),
		New: proto.String("another"), Text: proto.String("a reason")}); err != nil {
		t.Fatal(err)
	}
	if docs, err := AssembleAll(run); err == nil || !strings.Contains(err.Error(), "the report's text") {
		t.Fatalf("assembly of a report that does not replay: err %v, %d document(s) — it ships as though blue wrote nothing", err, len(docs))
	}

	// A report the record cannot READ is the same failure: two bases is a record no replay starts from.
	twice := runtest.Open(t, newRun(t))
	id = record.Identity{Run: twice, SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String("# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n")}); err != nil {
			t.Fatal(err)
		}
	}
	if docs, err := AssembleAll(twice); err == nil || !strings.Contains(err.Error(), "the report's text") {
		t.Fatalf("assembly of a report the record cannot read: err %v, %d document(s) — it ships as though blue wrote nothing", err, len(docs))
	}
}
