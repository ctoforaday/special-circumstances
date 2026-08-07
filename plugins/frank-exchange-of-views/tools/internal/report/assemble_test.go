package report

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func TestSectionCopiesVerbatimAndIsFenceAware(t *testing.T) {
	md := strings.Join([]string{
		"# report", "",
		"## The Catechism", "", "Q1: kept  as-is.  ", "", // trailing spaces preserved inside
		"```", "## Technical foundations", "not a heading — inside a fence", "```", "",
		"## Analysis", "", "the analysis.", "",
	}, "\n")

	cat := section(md, "The Catechism")
	if !strings.Contains(cat, "Q1: kept  as-is.") {
		t.Errorf("catechism not copied verbatim: %q", cat)
	}
	// The fenced "## Technical foundations" must NOT end the Catechism section.
	if !strings.Contains(cat, "not a heading — inside a fence") {
		t.Errorf("a ## inside a code fence wrongly ended the section: %q", cat)
	}
	// A real "## Technical foundations" section does not exist (only the fenced one), so it
	// is reported missing rather than authored.
	if got := section(md, "Technical foundations"); got != "" {
		t.Errorf("a fenced heading was mistaken for a real one: %q", got)
	}
	if !strings.Contains(sectionOr(md, "Technical foundations"), "not authored here") {
		t.Error("a missing section must be flagged, never authored")
	}
	// Case-insensitive: blue title-cased "## Open Questions" must still be lifted, not
	// declared absent against the template's lowercase "## Open questions".
	caseMd := "intro\n## Open Questions\n\nwhat remains.\n"
	if got := section(caseMd, "Open questions"); !strings.Contains(got, "what remains") {
		t.Errorf("case-insensitive heading match failed — a present section was declared absent: %q", got)
	}
}

func TestTitleLiftedOrFlagged(t *testing.T) {
	if got := titleOr("intro\n# Whether X — research report\n## The Catechism\n"); got != "# Whether X — research report" {
		t.Errorf("title not lifted verbatim: %q", got)
	}
	if got := titleOr("no title here\n## Analysis\n"); !strings.Contains(got, "not authored here") {
		t.Errorf("a missing title must be flagged, never authored: %q", got)
	}
}

func TestRiskMatrixFromBoard(t *testing.T) {
	longProblem := "Blue claims JSON float-loss above 2^53 causes H1 failure. Event logs serialize timestamps as ISO 8601 strings, so it does not manifest."
	bj := record.BoardJSON{Open: []record.GapJSON{
		{ID: "R1-1", Problem: "overclaims capture", Likelihood: "high", Impact: "medium", RequiredFix: "grep the sites"},
		{ID: "R1-2", Problem: longProblem, Likelihood: "low", Impact: "low", RequiredFix: "verify serialization"},
	}}
	m := riskMatrix(bj)
	if !strings.Contains(m, "| overclaims capture | high | medium | — | grep the sites |") {
		t.Errorf("short open gap row wrong:\n%s", m)
	}
	// A long problem is distilled to its FIRST SENTENCE in the cell; the full text lives in
	// Red team findings, so the matrix stays a scan surface.
	if !strings.Contains(m, "Blue claims JSON float-loss above 2^53 causes H1 failure.") || strings.Contains(m, "does not manifest") {
		t.Errorf("long problem not distilled to first sentence in the matrix cell:\n%s", m)
	}
	if !strings.Contains(m, "| — |") {
		t.Errorf("absent complexity grade should render as a dash:\n%s", m)
	}
	empty := riskMatrix(record.BoardJSON{})
	if !strings.Contains(empty, "no open gaps") {
		t.Errorf("empty board should say so:\n%s", empty)
	}
}

func TestVerdictStampFromOutcomeEvent(t *testing.T) {
	// A missing outcome is flagged, never invented.
	if s := verdictStamp(nil); !strings.Contains(s, "no terminal outcome recorded") {
		t.Errorf("missing outcome must be flagged: %q", s)
	}
	ceiling := record.NewPayload().Set("verdict", "CEILING")
	if s := verdictStamp(ceiling); !strings.Contains(s, "CEILING-TERMINATED") || !strings.Contains(s, "never audited by a red pass") || !strings.Contains(s, "travels OUT of the run") {
		t.Errorf("CEILING stamp must name the re-audit debt and not read as a failure: %q", s)
	}
	halted := record.NewPayload().Set("verdict", "HALTED")
	if s := verdictStamp(halted); !strings.Contains(s, "HALTED") || !strings.Contains(s, "Bench disposition") {
		t.Errorf("HALTED stamp must point at the recorded halt opinion: %q", s)
	}
	deadlock := record.NewPayload().Set("verdict", "UNVERIFIED").Set("deadlocked", true)
	if s := verdictStamp(deadlock); !strings.Contains(s, "UNVERIFIED by judged deadlock") {
		t.Errorf("deadlock reason not stamped: %q", s)
	}
	exhausted := record.NewPayload().Set("verdict", "UNVERIFIED").Set("exhausted", true)
	if s := verdictStamp(exhausted); !strings.Contains(s, "UNVERIFIED by safety ceiling") {
		t.Errorf("exhausted reason not stamped: %q", s)
	}
}

func TestAvenuesSplitByFate(t *testing.T) {
	evs := []record.Event{
		{Type: "avenue", SeatID: "blue-r1", Payload: record.NewPayload().Set("status", "pursued").Set("line", "profile the hot path").Set("method", "bench")},
		{Type: "avenue", SeatID: "blue-r1", Payload: record.NewPayload().Set("status", "abandoned").Set("line", "rewrite in Rust").Set("reason", "cost exceeds benefit")},
		{Type: "avenue", SeatID: "red-lens-r1", Payload: record.NewPayload().Set("status", "declined").Set("line", "third-party audit").Set("reason", "out of scope")},
	}
	exp := avenues(evs, "The expansions", accepted)
	if !strings.Contains(exp, "profile the hot path") || strings.Contains(exp, "rewrite in Rust") {
		t.Errorf("expansions must carry ONLY accepted (pursued) avenues:\n%s", exp)
	}
	alt := avenues(evs, "Alternatives considered", rejected)
	if !strings.Contains(alt, "rewrite in Rust") || !strings.Contains(alt, "cost exceeds benefit") {
		t.Errorf("a rejected avenue is an alternative considered, its reason the counter:\n%s", alt)
	}
	if !strings.Contains(alt, "third-party audit") {
		t.Errorf("a declined avenue is also an alternative considered:\n%s", alt)
	}
	if strings.Contains(alt, "profile the hot path") {
		t.Errorf("a pursued avenue must not appear under alternatives:\n%s", alt)
	}
	// No avenues of a fate → flagged, not blank.
	if none := avenues(nil, "The expansions", accepted); !strings.Contains(none, "none on the record") {
		t.Errorf("empty fate should say so: %q", none)
	}
}

func TestDebateTranscriptFromEvents(t *testing.T) {
	evs := []record.Event{
		{Round: 1, Type: "position", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("text", "gap A stands")},
		{Round: 1, Type: "position", SeatID: "blue-r1", Payload: record.NewPayload().Set("text", "gap A repaired")},
		// The payload keys are the ones the VERBS write: dispute→evidence, dispute-respond→
		// response+rationale, petition-rule→opinion. The prior fixture set basis/as (what the
		// buggy reader looked for), which is how A1–A3 hid — the test encoded the bug.
		{Round: 1, Type: "dispute", SeatID: "blue-r1", Payload: record.NewPayload().Set("gap_id", "R1-1").Set("dimension", "impact").Set("proposed", "low").Set("evidence", "trivial harm")},
		{Round: 1, Type: "dispute-respond", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("response", "rejected").Set("rationale", "harm compounds")},
		{Round: 1, Type: "opinion", SeatID: "judge-r1", Payload: record.NewPayload().Set("gap_id", "R1-1").Set("disposition", "carried").Set("principle", "correctness").Set("tension", "cost").Set("review_flag", "false").Set("rationale", "needs a probe")},
		{Round: 1, Type: "petition-rule", SeatID: "judge-petition", Payload: record.NewPayload().Set("petitioner", "blue").Set("ruling", "granted").Set("opinion", "relief warranted")},
		{Round: 0, Type: "halt", SeatID: "judge-terminal", Payload: record.NewPayload().Set("opinion", "safety gate tripped")},
		{Round: 0, Type: "certify", SeatID: "judge-terminal", Payload: record.NewPayload().Set("statement", "re-examine the cost model")},
	}
	d := debate(evs)
	for _, want := range []string{
		"### Round 1", "### RED\ngap A stands", "### BLUE\ngap A repaired",
		"disputes R1-1/impact → low: trivial harm", "answered (rejected): harm compounds",
		"R1-1: carried", "petition blue: granted", "relief warranted", // A3: petition prose now renders
		"### Bench disposition", "**HALT** — safety gate tripped", "**Certification** — re-examine the cost model",
	} {
		if !strings.Contains(d, want) {
			t.Errorf("debate transcript missing %q:\n%s", want, d)
		}
	}
	if empty := debate(nil); !strings.Contains(empty, "no debate on the record") {
		t.Errorf("empty debate should say so: %q", empty)
	}
}

func TestBlueEmbedDropsLiftedAndFabricated(t *testing.T) {
	blue := strings.Join([]string{
		"# A topic — research report",
		"**Verdict:** UNVERIFIED (Round 0)", // blue cannot author a verdict — must be stripped
		"",
		"## TL;DR", "lifted to the top.", "",
		"## Analysis", "also lifted.", "",
		"## Risk Matrix", "blue fabricated a risk matrix.", "", // tool-owned — dropped
		"## Red Team Findings (in full)", "blue cannot know red's findings.", "", // dropped
		"## Blue Team Report (in full)", "[to be filled]", "", // recursive stub — dropped
		"## Footnotes", "[^a]: a citation blue tried to author.", "", // DROPPED — citations are tool-composed now
		"## Appendix: raw benchmarks", "novel blue content.", "", // KEPT — genuinely additional
	}, "\n")
	got := blueEmbed(blue)

	for _, kept := range []string{"## Appendix: raw benchmarks", "novel blue content."} {
		if !strings.Contains(got, kept) {
			t.Errorf("blueEmbed dropped content it should keep (%q):\n%s", kept, got)
		}
	}
	for _, dropped := range []string{"lifted to the top", "also lifted", "blue fabricated", "blue cannot know", "[to be filled]", "**Verdict:**", "UNVERIFIED", "## Footnotes", "a citation blue tried to author"} {
		if strings.Contains(got, dropped) {
			t.Errorf("blueEmbed kept content it should drop (%q):\n%s", dropped, got)
		}
	}
	// A perfectly-scoped blue doc (only lifted sections) leaves nothing to embed.
	scoped := "# t\n\n## TL;DR\nx\n\n## Analysis\ny\n"
	if e := blueEmbed(scoped); e != "" {
		t.Errorf("a correctly-scoped blue doc should yield an empty embed, got:\n%s", e)
	}
}

func TestOrientationRanksAndPromotesBench(t *testing.T) {
	board := &record.Board{
		GapOrder: []string{"R1-1", "R1-2", "R1-3"},
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Open: true, Severity: "low", Impact: "low", Likelihood: "low",
				Mint: record.NewPayload().Set("problem", "a minor nit.").Set("required_fix", "tidy it")},
			"R1-2": {ID: "R1-2", Open: true, Severity: "certain", Impact: "high", Likelihood: "high",
				Mint: record.NewPayload().Set("problem", "a load-bearing flaw.").Set("required_fix", "fix the core")},
			"R1-3": {ID: "R1-3", Open: false, Severity: "high", // closed — must not appear
				Mint: record.NewPayload().Set("problem", "already closed.")},
		},
	}
	evs := []record.Event{
		{Type: "certify", Payload: record.NewPayload().Set("statement", "re-examine the cost model before shipping")},
	}
	o := orientation(board, evs)
	// The bench's certify is promoted to the top.
	if !strings.Contains(o, "re-examine the cost model before shipping") {
		t.Errorf("orientation must promote the bench's certify statement:\n%s", o)
	}
	// The load-bearing flaw (severity "certain", a top domain grade the old critical|high|medium|
	// low table sank to rank 0) ranks above the minor nit, and the closed gap is absent.
	ci := strings.Index(o, "a load-bearing flaw")
	ni := strings.Index(o, "a minor nit")
	if ci < 0 || ni < 0 || ci > ni {
		t.Errorf("open gaps must be ranked most-severe first:\n%s", o)
	}
	if strings.Contains(o, "already closed") {
		t.Errorf("a closed gap must not appear in Read this first:\n%s", o)
	}
	// Empty board says so, invents nothing.
	empty := orientation(&record.Board{}, nil)
	if !strings.Contains(empty, "no open gaps remain") {
		t.Errorf("an empty board should say nothing is outstanding:\n%s", empty)
	}
}

func TestUnmintedFindingsSurfaced(t *testing.T) {
	board := &record.Board{
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Mint: record.NewPayload().Set("found_by", []string{"L5-F1", "L6-F2"})},
		},
		Events: []record.Event{
			{Type: "finding", SeatID: "red-lens-r1-L5", Payload: record.NewPayload().Set("label", "L5-F1").Set("text", "minted — omit")},
			{Type: "finding", SeatID: "red-lens-r1-L6", Payload: record.NewPayload().Set("label", "L6-F2").Set("text", "also minted — omit")},
			{Type: "finding", SeatID: "red-lens-r1-L5", Payload: record.NewPayload().Set("label", "L5-F3").Set("location", "§H1").Set("text", "un-minted red reasoning kept for the record")},
		},
	}
	got := redFindings(board)
	if !strings.Contains(got, "Lens findings not raised to a gap (1)") {
		t.Errorf("exactly one un-minted finding should be surfaced:\n%s", got)
	}
	if !strings.Contains(got, "L5-F3") || !strings.Contains(got, "un-minted red reasoning kept") {
		t.Errorf("the un-minted finding's substance must be surfaced:\n%s", got)
	}
	if strings.Contains(got, "minted — omit") || strings.Contains(got, "also minted") {
		t.Errorf("a finding credited by a gap's found_by must NOT be re-listed:\n%s", got)
	}
}

func TestFrictionRendered(t *testing.T) {
	evs := []record.Event{
		{Type: "friction", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("text", "the --cx flag is missing from help")},
		{Type: "friction", SeatID: "blue-respond-r2", Payload: record.NewPayload().Set("text", "manifest cap fights methodology gaps")},
		{Type: "mint", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("problem", "not friction")},
	}
	f := frictionLog(evs)
	for _, want := range []string{"Friction (tooling gaps", "**red-merge-r1**: the --cx flag is missing", "**blue-respond-r2**: manifest cap fights"} {
		if !strings.Contains(f, want) {
			t.Errorf("friction log missing %q:\n%s", want, f)
		}
	}
	if empty := frictionLog(nil); empty != "" {
		t.Errorf("no friction events should render nothing, got: %q", empty)
	}
}

func TestConfidenceSelfAssessmentRenders(t *testing.T) {
	evs := []record.Event{
		{Round: 0, Type: "confidence", SeatID: "blue-synthesize", Payload: record.NewPayload().Set("label", "C1: throughput doubles").Set("grade", "high")},
		{Round: 1, Type: "confidence", SeatID: "blue-respond-r1", Payload: record.NewPayload().Set("label", "C2: latency budget holds").Set("grade", "low")},
		// A confidence event with no claim label contributes no row (nothing to target).
		{Round: 1, Type: "confidence", SeatID: "blue-respond-r1", Payload: record.NewPayload().Set("grade", "high")},
		{Round: 1, Type: "mint", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("problem", "not a confidence event")},
	}
	got := confidenceSelfAssessment(evs)
	for _, want := range []string{
		"## Blue's confidence self-assessment",
		"not red's audit", "does not feed the risk matrix", // the non-authoritative disclaimer
		"C1: throughput doubles", "high", "r0",
		"C2: latency budget holds", "low", "r1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("confidence self-assessment missing %q:\n%s", want, got)
		}
	}
	// Two labeled claims → exactly two table rows (the label-less event is dropped).
	if n := strings.Count(got, "| r"); n != 2 {
		t.Errorf("expected 2 confidence rows, got %d:\n%s", n, got)
	}
	if empty := confidenceSelfAssessment(nil); empty != "" {
		t.Errorf("no confidence events should render nothing, got: %q", empty)
	}
}

// The non-authoritative invariant, made structural: blue's confidence surfaces in its OWN
// section but NEVER inside the risk matrix (which composes from red's gap board alone). This
// is the guard against a future edit wiring confidence into the graded surface — blue must
// not grade its own exam.
func TestConfidenceStaysOutOfTheRiskMatrix(t *testing.T) {
	runDir := t.TempDir()

	// A real risk: red mints an open gap.
	if _, _, err := record.RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	id, err := record.MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(runDir, "red-merge-r1", "mint", record.NewPayload().Set("gap_id", id).
		Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").Set("likelihood", "high").Set("impact", "high").
		Set("problem", "the actual risk row")); err != nil {
		t.Fatal(err)
	}
	// Blue self-grades a claim with a distinctive marker label.
	if _, _, err := record.RegisterSeat(runDir, "blue-respond-r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(runDir, "blue-respond-r1", "confidence", record.NewPayload().
		Set("label", "BLUE-SELF-GRADE-MARKER").Set("grade", "high")); err != nil {
		t.Fatal(err)
	}

	path, err := Assemble(runDir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	full := string(b)

	// The confidence surfaces in its own section...
	if !strings.Contains(full, "## Blue's confidence self-assessment") || !strings.Contains(full, "BLUE-SELF-GRADE-MARKER") {
		t.Fatalf("blue's confidence self-assessment missing from the report:\n%s", full)
	}
	// ...but the risk matrix carries the gap and NOT the confidence.
	matrix := section(full, "Risk matrix")
	if matrix == "" {
		t.Fatalf("no risk matrix section in the assembled report:\n%s", full)
	}
	if !strings.Contains(matrix, "the actual risk row") {
		t.Errorf("risk matrix lost the real gap:\n%s", matrix)
	}
	if strings.Contains(matrix, "BLUE-SELF-GRADE-MARKER") {
		t.Errorf("NON-AUTHORITATIVE VIOLATION: blue's confidence leaked into the risk matrix:\n%s", matrix)
	}
}

func TestCellEscapesTableBreakers(t *testing.T) {
	if got := cell("a | b\nc"); strings.ContainsAny(got, "\n") || strings.Contains(got, " | ") {
		t.Errorf("cell did not neutralise a pipe/newline: %q", got)
	}
	if cell("   ") != "—" {
		t.Error("a blank cell should be a dash")
	}
}

func TestRevisionHistoryFromEvents(t *testing.T) {
	evs := []record.Event{
		{Round: 1, Type: "revision", SeatID: "blue-respond-r1", Payload: record.NewPayload().Set("text", "expanded the caching section; retired the stale figure")},
		{Round: 2, Type: "revision", SeatID: "blue-respond-r2", Payload: record.NewPayload().Set("text", "addressed R2-1 in the analysis")},
		{Round: 1, Type: "position", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("text", "not a revision")},
	}
	got := revisionHistory(evs)
	if !strings.Contains(got, "## Report revision history") {
		t.Fatalf("missing heading:\n%s", got)
	}
	if !strings.Contains(got, "### Round 1 — blue-respond-r1") || !strings.Contains(got, "expanded the caching section") {
		t.Errorf("round-1 revision not rendered:\n%s", got)
	}
	if !strings.Contains(got, "### Round 2 — blue-respond-r2") {
		t.Errorf("round-2 revision not rendered:\n%s", got)
	}
	if strings.Contains(got, "not a revision") {
		t.Errorf("a non-revision event leaked into the revision history:\n%s", got)
	}
	if revisionHistory(nil) != "" {
		t.Error("no revisions must yield empty (section omitted), not a bare heading")
	}
}
