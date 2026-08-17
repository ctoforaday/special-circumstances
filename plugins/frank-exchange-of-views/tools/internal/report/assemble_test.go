package report

import (
	"sort"
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

func TestInquiriesSplitByFate(t *testing.T) {
	board := &record.Board{Events: []record.Event{
		{Type: "line-of-inquiry", SeatID: "blue-r1", Payload: record.NewPayload().Set("inquiry_id", "Q1").Set("status", "pursued").Set("line", "profile the hot path").Set("method", "bench")},
		{Type: "line-of-inquiry", SeatID: "blue-r1", Payload: record.NewPayload().Set("inquiry_id", "Q2").Set("status", "abandoned").Set("line", "rewrite in Rust").Set("reason", "cost exceeds benefit")},
		{Type: "line-of-inquiry", SeatID: "red-lens-r1", Payload: record.NewPayload().Set("inquiry_id", "Q3").Set("status", "declined").Set("line", "third-party audit").Set("reason", "out of scope")},
	}}
	exp := inquiries(board, "Research areas", accepted)
	if !strings.Contains(exp, "profile the hot path") || strings.Contains(exp, "rewrite in Rust") {
		t.Errorf("research areas must carry ONLY pursued and proposed inquiries:\n%s", exp)
	}
	alt := inquiries(board, "Alternatives considered", rejected)
	if !strings.Contains(alt, "rewrite in Rust") || !strings.Contains(alt, "cost exceeds benefit") {
		t.Errorf("a rejected line of inquiry is an alternative considered, its reason the counter:\n%s", alt)
	}
	if !strings.Contains(alt, "third-party audit") {
		t.Errorf("a declined line of inquiry is also an alternative considered:\n%s", alt)
	}
	if strings.Contains(alt, "profile the hot path") {
		t.Errorf("a pursued line of inquiry must not appear under alternatives:\n%s", alt)
	}
	// No inquiries of a fate → flagged, not blank.
	if none := inquiries(&record.Board{}, "Research areas", accepted); !strings.Contains(none, "none on the record") {
		t.Errorf("empty fate should say so: %q", none)
	}
}

// EVERY STATUS LANDS WHERE ITS FATE SAYS — three sections, and one deliberate exclusion.
//
// # This test never looked at a status
//
// It ranged over `record.InquiryStatuses`, which is `[]EnumValue`, and passed the STRUCT to
// `Set("status", status)`. `Payload.Str` returns "" for a non-string, so `a.Status` was the empty
// string on all five iterations. With `rejected` defined as the complement of `accepted`, "" fell
// into alternatives every time and `in != inAlt` held — so it passed, five times, having exercised
// no status at all. A test whose entire subject is "every status reaches a section" had never seen
// one. Found 2026-08-16 while adding the third section. It ranges over InquiryStatusNames() now.
//
// # The model this asserts: the lifecycle of a research topic through the report
//
//	pursued              "Research areas" — a topic followed, and what it yielded
//	deferred             "Future research directions" — KEPT for a later run or a deeper context,
//	                     which is a decision. Filing it under "Alternatives considered" said the
//	                     opposite of its fate, and only the `[deferred]` tag on the row
//	                     contradicted the heading above it.
//	declined, abandoned  Alternatives considered — weighed and not taken, or tried and died
//	proposed             "Research areas", beside `pursued`, with `[proposed]` on its own row. A
//	                     first cut excluded it from every section (TestFuzzDebate refused: a seat's
//	                     recorded reasoning must reach the reader) and a second gave it a fourth
//	                     section. Three areas is the decision: a line blue put forward IS an area
//	                     this run is researching, and red's per-round support verdict is what stops
//	                     it sitting undecided — not a heading describing the omission.
//
// EVERY status lands in EXACTLY ONE section, so a sixth that matches no predicate fails here
// rather than vanishing the way `proposed` and `deferred` once did.
func TestEveryInquiryStatusLandsWhereItsFateSays(t *testing.T) {
	section := map[string]string{
		"pursued":   "research",
		"proposed":  "research", // an undecided line IS an area this run is researching
		"deferred":  "future",
		"declined":  "alternatives",
		"abandoned": "alternatives",
	}
	for _, status := range record.InquiryStatusNames() {
		want, known := section[status]
		if !known {
			t.Errorf("status %q is in the enum and this test has no expectation for it — decide which section it "+
				"belongs to (or that it belongs to none, like `proposed`) rather than letting it match a complement "+
				"by accident", status)
			continue
		}
		board := &record.Board{Events: []record.Event{{Type: "line-of-inquiry", SeatID: "blue-r1",
			Payload: record.NewPayload().Set("inquiry_id", "Q1").Set("status", status).Set("line", "the only line")}}}
		in := map[string]bool{
			"research":     strings.Contains(inquiries(board, "Research areas", accepted), "the only line"),
			"future":       strings.Contains(inquiries(board, "Future research directions", deferred), "the only line"),
			"alternatives": strings.Contains(inquiries(board, "Alternatives considered", rejected), "the only line"),
		}
		var got []string
		for name, present := range in {
			if present {
				got = append(got, name)
			}
		}
		sort.Strings(got)
		switch {
		case len(got) != 1:
			t.Errorf("status %q rendered under %v, want exactly [%s] — a line of inquiry in two sections is an alternative to "+
				"itself, and one in NONE loses the seat's recorded prose entirely (TestFuzzDebate's A1-A3 class "+
				"caught exactly that when `proposed` was excluded)", status, got, want)
		case got[0] != want:
			t.Errorf("status %q rendered under %q, want %q", status, got[0], want)
		}
	}
}

// A MOVED LINE OF INQUIRY IS ONE LINE. Reading raw events rendered a line pursued at r0 and abandoned
// at r2 under BOTH headings — as an expansion and as an alternative to itself.
func TestAMovedInquiryIsRenderedOnce(t *testing.T) {
	board := &record.Board{Events: []record.Event{
		{Round: 0, Type: "line-of-inquiry", SeatID: "blue-r0", Payload: record.NewPayload().Set("inquiry_id", "Q1").Set("status", "pursued").Set("line", "rewrite the parser")},
		{Round: 2, Type: "line-of-inquiry", SeatID: "blue-r2", Payload: record.NewPayload().Set("inquiry_id", "Q1").Set("status", "abandoned").Set("reason", "the grammar moved under it")},
	}}
	exp := inquiries(board, "Research areas", accepted)
	alt := inquiries(board, "Alternatives considered", rejected)
	if strings.Contains(exp, "rewrite the parser") {
		t.Errorf("a line of inquiry ABANDONED at r2 is not an expansion — its latest status decides:\n%s", exp)
	}
	if !strings.Contains(alt, "rewrite the parser") || !strings.Contains(alt, "the grammar moved under it") {
		t.Errorf("the abandoned line of inquiry must carry its current reason:\n%s", alt)
	}
	// The substance came from the CREATION event and the reason from the MOVE — the history is
	// the evidence of choosing, which is the whole point of giving a line of inquiry a lifecycle.
	if !strings.Contains(alt, "r0 pursued → r2 abandoned") {
		t.Errorf("the history that produced the status must be rendered:\n%s", alt)
	}
}

// RED'S RULING AND BLUE'S DEFIANCE OF IT ARE THE SUBSTANCE. Blue pursuing a line red ruled
// out-of-scope looked identical to blue pursuing one red endorsed.
func TestInquiryRulingAndContestReachTheReader(t *testing.T) {
	board := &record.Board{Events: []record.Event{
		{Round: 0, Type: "line-of-inquiry", SeatID: "blue-r0", Payload: record.NewPayload().Set("inquiry_id", "Q1").Set("status", "proposed").Set("line", "survey the adjacent literature")},
		// The LIVE vocabulary: red rules a direction through `motion inquiry rule`, whose motion_id
		// IS the line's own id — the proposal is the filing, so there is no second identity. The
		// fixture used to write the retired `avenue-rule` type, which nothing has written since
		// the motion collapse and which no longer has a read arm.
		{Round: 1, Type: "motion-rule", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("subject", "inquiry").Set("motion_id", "Q1").Set("ruling", "out-of-scope").Set("reason", "a real question, not THIS run's")},
		{Round: 1, Type: "line-of-inquiry", SeatID: "blue-r1", Payload: record.NewPayload().Set("inquiry_id", "Q1").Set("status", "pursued").Set("contests_ruling", "out-of-scope")},
	}}
	exp := inquiries(board, "Research areas", accepted)
	for _, want := range []string{"out-of-scope", "a real question, not THIS run's", "against red's"} {
		if !strings.Contains(exp, want) {
			t.Errorf("the reader must see the ruling AND that blue moved against it; missing %q:\n%s", want, exp)
		}
	}
}

func TestDebateTranscriptFromEvents(t *testing.T) {
	evs := []record.Event{
		{Round: 1, Type: "position", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("reason", "gap A stands")},
		{Round: 1, Type: "position", SeatID: "blue-r1", Payload: record.NewPayload().Set("reason", "gap A repaired")},
		// The payload keys are the ones the VERBS write: dispute→evidence, dispute-respond→
		// response+rationale, petition-rule→opinion. The prior fixture set basis/as (what the
		// buggy reader looked for), which is how A1–A3 hid — the test encoded the bug.
		{Round: 1, Type: "dispute", SeatID: "blue-r1", Payload: record.NewPayload().Set("gap_id", "R1-1").Set("dimension", "impact").Set("proposed", "low").Set("reason", "trivial harm")},
		{Round: 1, Type: "dispute-respond", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("response", "rejected").Set("reason", "harm compounds")},
		{Round: 1, Type: "opinion", SeatID: "judge-r1", Payload: record.NewPayload().Set("gap_id", "R1-1").Set("disposition", "carried").Set("principle", "correctness").Set("tension", "cost").Set("review_flag", "false").Set("reason", "needs a probe")},
		{Round: 1, Type: "petition", SeatID: "blue", Payload: record.NewPayload().Set("class", "integrity").Set("reason", "the instruction would require asserting what I believe false").Set("relief", "strike the demand from the docket")},
		{Round: 1, Type: "petition-rule", SeatID: "judge-petition", Payload: record.NewPayload().Set("petitioner", "blue").Set("ruling", "granted").Set("reason", "relief warranted")},
		{Round: 0, Type: "halt", SeatID: "judge-terminal", Payload: record.NewPayload().Set("reason", "safety gate tripped")},
		{Round: 0, Type: "certify", SeatID: "judge-terminal", Payload: record.NewPayload().Set("reason", "re-examine the cost model")},
	}
	d := debate(evs)
	for _, want := range []string{
		"### Round 1", "### RED\ngap A stands", "### BLUE\ngap A repaired",
		"disputes R1-1/impact → low: trivial harm", "answered (rejected): harm compounds",
		"R1-1: carried",
		// BOTH SIDES OF THE PETITION. The transcript used to render only the ruling, so the
		// reader got the bench's answer with no question attached — and the relief sought and
		// the basis argued are what the answer is an answer TO.
		"### Petitions",
		"**blue petitions the bench (integrity)**: the instruction would require asserting what I believe false",
		"relief sought: strike the demand from the docket",
		"ruled **granted** on blue's petition", "relief warranted",
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

// THE BASIS FIELDS EXIST BECAUSE A SEAT ASKED TO SELF-REPORT REPORTS THE FLATTERING VALUE.
// Each was derived rather than claimed, gated a write, and then reached the reader as nothing —
// so a verdict the record itself decided read as the same word as one the bench simply asserted.
func TestVerdictBasisReachesTheReader(t *testing.T) {
	derived := record.NewPayload().Set("verdict", "VERIFIED").Set("verdict_basis", record.VerdictDerived)
	if s := verdictStamp(derived); !strings.Contains(s, "derived from the record") {
		t.Errorf("a DERIVED verdict must say so — it is the difference between a mechanical result and a claim: %q", s)
	}
	asserted := record.NewPayload().Set("verdict", "VERIFIED").Set("verdict_basis", record.VerdictAsserted)
	s := verdictStamp(asserted)
	if !strings.Contains(s, "asserted by the bench") {
		t.Errorf("an ASSERTED verdict must say so, or it reads as a derived one: %q", s)
	}
	if strings.Contains(s, "derived from the record") {
		t.Errorf("an asserted verdict must not claim derivation: %q", s)
	}
	// A run recorded before the field existed carries no basis; it says nothing rather than
	// guessing, because guessing here would invent the very distinction the field preserves.
	if s := verdictStamp(record.NewPayload().Set("verdict", "VERIFIED")); strings.Contains(s, "basis") {
		t.Errorf("no recorded basis must produce no basis claim: %q", s)
	}
	// EVERY VERDICT BRANCH CARRIES IT. The first cut appended the note only to the default arm,
	// so CEILING and HALTED returned early and dropped it — and a ceiling termination IS derived
	// (rounds against the configured ceiling), and is the most common way a run ends.
	for _, verdict := range []string{"CEILING", "HALTED", "VERIFIED", "UNVERIFIED"} {
		p := record.NewPayload().Set("verdict", verdict).Set("verdict_basis", record.VerdictDerived)
		if s := verdictStamp(p); !strings.Contains(s, "derived from the record") {
			t.Errorf("%s dropped its basis — every terminal verdict says how it was reached: %q", verdict, s)
		}
	}
}

// fix_basis reads `verified` only when red supplied an exact span AND replacement that the tool
// validated against the LIVE report — a forced re-read. A demand checked against the document
// read identically to one written from memory of what the document probably said.
func TestFixBasisAndTheConcreteProposalReachTheReader(t *testing.T) {
	verified := record.NewPayload().Set("fix_basis", "verified").
		Set("fix_old", "the parser is linear").Set("fix_new", "the parser is linear except on backtracking")
	s := fixProposal(verified)
	for _, want := range []string{"**verified**", "with the text in front of it", "the parser is linear except on backtracking"} {
		if !strings.Contains(s, want) {
			t.Errorf("a verified fix must show its checked replacement; missing %q:\n%s", want, s)
		}
	}
	if p := fixProposal(record.NewPayload().Set("fix_basis", "proposed")); !strings.Contains(p, "**proposed**") || !strings.Contains(p, "nothing checked this demand") {
		t.Errorf("a prose-only demand must say nothing checked it against the report: %q", p)
	}
	if p := fixProposal(record.NewPayload()); p != "" {
		t.Errorf("no recorded basis must produce no claim: %q", p)
	}
	if p := fixProposal(nil); p != "" {
		t.Errorf("a nil mint must not panic or invent a basis: %q", p)
	}
}

// A PHANTOM RETIREMENT CANCELS REAL LOSS in the scorecard's additive-integrity detector, and
// only the basis distinguishes one from an honest round-0 rewrite.
func TestRemovalBasisReachesTheReader(t *testing.T) {
	v := withdrawnClaims([]record.Event{{Type: "retire", SeatID: "blue-r2", Payload: record.NewPayload().
		Set("claim", "c").Set("reason", "r").Set("removal_basis", record.RemovalVerified)}})
	if !strings.Contains(v, "**verified**") || !strings.Contains(v, "the record shows it leaving") {
		t.Errorf("a verified removal must say the record can show it:\n%s", v)
	}
	a := withdrawnClaims([]record.Event{{Type: "retire", SeatID: "blue-r0", Payload: record.NewPayload().
		Set("claim", "c").Set("reason", "r").Set("removal_basis", record.RemovalAsserted)}})
	if !strings.Contains(a, "**asserted**") || !strings.Contains(a, "nothing on the record shows it was ever present") {
		t.Errorf("an asserted removal must say the record cannot show it:\n%s", a)
	}
}

// AN UNANSWERED PETITION MUST BE LOUD. A petition is a seat's channel for an ethical, safety,
// integrity or constitutional objection, and the engine routes it to a bench sitting BEFORE the
// debate continues. A filing with no ruling means that sitting did not happen — and reporting
// nothing would make it indistinguishable from a run that had no petitions at all.
func TestAnUnansweredPetitionIsReported(t *testing.T) {
	filed := []record.Event{{Round: 1, Type: "petition", SeatID: "red-merge-r1",
		Payload: record.NewPayload().Set("class", "safety").Set("reason", "the demand would bury a hazard")}}
	d := debate(filed)
	if !strings.Contains(d, "1 petition(s) received no ruling") {
		t.Errorf("a petition with no ruling must be reported, not silently absent:\n%s", d)
	}
	answered := append(filed, record.Event{Round: 1, Type: "petition-rule", SeatID: "judge-r1",
		Payload: record.NewPayload().Set("petitioner", "red-merge-r1").Set("ruling", "denied").Set("reason", "the hazard is graded, not buried")})
	if d := debate(answered); strings.Contains(d, "received no ruling") {
		t.Errorf("an answered petition must not be reported as unanswered:\n%s", d)
	}
}

// A WITHDRAWN CLAIM IS PART OF WHAT THE DEBATE DECIDED. Substance leaves the report only through
// `retire`, which names the claim as it stood and why it went — and the reader saw none of it,
// making a report where a claim was argued and withdrawn identical to one where it was never made.
func TestWithdrawnClaimsReachTheReader(t *testing.T) {
	evs := []record.Event{{Round: 2, Type: "retire", SeatID: "blue-respond-r2", Payload: record.NewPayload().
		Set("claim", "the parser is O(n) in the input size").
		Set("reason", "refuted at the leaf — the inner scan is quadratic on backtracking").
		Set("superseded_by", "the parser is O(n) except on backtracking inputs")}}
	w := withdrawnClaims(evs)
	for _, want := range []string{"the parser is O(n) in the input size", "refuted at the leaf", "superseded by: the parser is O(n) except on backtracking inputs"} {
		if !strings.Contains(w, want) {
			t.Errorf("withdrawn claims missing %q:\n%s", want, w)
		}
	}
	if withdrawnClaims(nil) != "" {
		t.Error("a run that retired nothing omits the section rather than showing it empty")
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
		{Type: "certify", Payload: record.NewPayload().Set("reason", "re-examine the cost model before shipping")},
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
			{Type: "finding", SeatID: "red-lens-r1-L5", Payload: record.NewPayload().Set("label", "L5-F1").Set("reason", "minted — omit")},
			{Type: "finding", SeatID: "red-lens-r1-L6", Payload: record.NewPayload().Set("label", "L6-F2").Set("reason", "also minted — omit")},
			{Type: "finding", SeatID: "red-lens-r1-L5", Payload: record.NewPayload().Set("label", "L5-F3").Set("location", "§H1").Set("reason", "un-minted red reasoning kept for the record")},
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
		{Type: "friction", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("reason", "the --cx flag is missing from help")},
		{Type: "friction", SeatID: "blue-respond-r2", Payload: record.NewPayload().Set("reason", "manifest cap fights methodology gaps")},
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
		{Round: 1, Type: "revision", SeatID: "blue-respond-r1", Payload: record.NewPayload().Set("reason", "expanded the caching section; retired the stale figure")},
		{Round: 2, Type: "revision", SeatID: "blue-respond-r2", Payload: record.NewPayload().Set("reason", "addressed R2-1 in the analysis")},
		{Round: 1, Type: "position", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("reason", "not a revision")},
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
