package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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
	// The board, so the matrix stays a scan surface.
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
	// THE STAMP IS A FIELD. It carries the word and the clause naming how the run ended, and
	// nothing else — a fact a reader can skim, badge or grep has to be one token. The argument
	// that used to sit inline is verdictGloss, asserted directly below.
	ceiling := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING)}
	if s := verdictStamp(ceiling); s != "**Verdict:** CEILING-TERMINATED" {
		t.Errorf("the verdict field must be the word alone: %q", s)
	}
	if g := verdictGloss(ceiling); !strings.Contains(g, "CEILING-TERMINATED") || !strings.Contains(g, "never audited by a red pass") || !strings.Contains(g, "travels OUT of the run") {
		t.Errorf("the CEILING gloss must name the re-audit debt and not read as a failure: %q", g)
	}
	halted := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED)}
	if s := verdictStamp(halted); s != "**Verdict:** HALTED" {
		t.Errorf("the verdict field must be the word alone: %q", s)
	}
	if g := verdictGloss(halted); !strings.Contains(g, "HALTED") || !strings.Contains(g, "halt opinion") {
		t.Errorf("the HALTED gloss must point at the recorded halt opinion: %q", g)
	}
	// A missing outcome is flagged by the STAMP, and assembly does not repeat it into
	// "Read this first" — the gloss exists to carry an argument the field cannot hold, and
	// there is no argument here.
	if g := verdictGloss(nil); !strings.Contains(g, "no terminal outcome recorded") {
		t.Errorf("a missing outcome must still be answerable from the gloss: %q", g)
	}
	deadlock := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED), Ended: proto.String("deadlock")}
	if s := verdictStamp(deadlock); !strings.Contains(s, "UNVERIFIED by judged deadlock") {
		t.Errorf("deadlock reason not stamped: %q", s)
	}
	exhausted := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED), Ended: proto.String("ceiling")}
	if s := verdictStamp(exhausted); !strings.Contains(s, "UNVERIFIED by safety ceiling") {
		t.Errorf("exhausted reason not stamped: %q", s)
	}
}

func TestInquiriesSplitByFate(t *testing.T) {
	board := &record.Board{Events: []*record.Event{
		recordtest.Event(t, "blue-r1", 1, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Line: proto.String("profile the hot path"), Method: proto.String("bench")}),
		recordtest.Event(t, "blue-r1", 1, &recordpb.Avenue{AvenueId: proto.String("Q2"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED), Line: proto.String("rewrite in Rust"), Reason: proto.String("cost exceeds benefit")}),
		recordtest.Event(t, "red-lens-r1", 1, &recordpb.Avenue{AvenueId: proto.String("Q3"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_DECLINED), Line: proto.String("third-party audit"), Reason: proto.String("out of scope")}),
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
		st, ok := record.AvenueStatusOf(status)
		if !ok {
			t.Fatalf("the declared status %q does not resolve to a schema value — the vocabulary and "+
				"the enum have drifted apart", status)
		}
		want, known := section[status]
		if !known {
			t.Errorf("status %q is in the enum and this test has no expectation for it — decide which section it "+
				"belongs to (or that it belongs to none, like `proposed`) rather than letting it match a complement "+
				"by accident", status)
			continue
		}
		board := &record.Board{Events: []*record.Event{
			recordtest.Event(t, "blue-r1", 1, &recordpb.Avenue{
				AvenueId: proto.String("Q1"), Status: st.Enum(), Line: proto.String("the only line"),
			}),
		}}
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
	board := &record.Board{Events: []*record.Event{
		recordtest.Event(t, "blue-r0", 0, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Line: proto.String("rewrite the parser")}),
		recordtest.Event(t, "blue-r2", 2, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED), Line: proto.String("rewrite the parser"), Reason: proto.String("the grammar moved under it")}),
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
	board := &record.Board{Events: []*record.Event{
		recordtest.Event(t, "blue-r0", 0, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Line: proto.String("survey the adjacent literature")}),
		// The LIVE vocabulary: red rules a direction through `motion inquiry rule`, whose motion_id
		// IS the line's own id — the proposal is the filing, so there is no second identity. The
		// fixture used to write the retired `avenue-rule` type, which nothing has written since
		// the motion collapse and which no longer has a read arm.
		recordtest.Event(t, "red-merge-r1", 1, &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION),
			Opinion:  proto.String("a real question, not THIS run's"),
			Ruling:   &recordpb.MotionRule_Direction{Direction: recordpb.DirectionRuling_DIRECTION_RULING_OUT_OF_SCOPE},
		}),
		// Blue pursues it ANYWAY. `contests_ruling` was a payload key on the line; the defiance is
		// its own act now — `motion direction appeal` — so what the report reads is the appeal.
		recordtest.Event(t, "blue-r1", 1, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Line: proto.String("survey the adjacent literature")}),
		recordtest.Event(t, "blue-r1", 1, &recordpb.MotionAppeal{
			MotionId: proto.String("Q1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION),
			Reason:   proto.String("the adjacent literature is what the question turns on"),
		}),
	}}
	exp := inquiries(board, "Research areas", accepted)
	// `out_of_scope`, WITH THE UNDERSCORE. An older comment in inquiry.go claims the hyphen is the
	// live spelling and the underscore "a word no surface recognizes"; the vocabulary says
	// otherwise — DirectionRuling spells it with an underscore and InquiryRulings agrees, so the
	// hyphen is what no surface recognizes now.
	for _, want := range []string{"out_of_scope", "a real question, not THIS run's", "against red's"} {
		if !strings.Contains(exp, want) {
			t.Errorf("the reader must see the ruling AND that blue moved against it; missing %q:\n%s", want, exp)
		}
	}
}

func TestDebateTranscriptFromEvents(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "red-merge-r1", 1, &recordpb.Position{Text: proto.String("gap A stands")}),
		recordtest.Event(t, "blue-r1", 1, &recordpb.Position{Text: proto.String("gap A repaired")}),
		// The payload keys are the ones the VERBS write: dispute→evidence, dispute-respond→
		// response+rationale, petition-rule→opinion. The prior fixture set basis/as (what the
		// buggy reader looked for), which is how A1–A3 hid — the test encoded the bug.
		// THE BENCH'S DISPOSITION IS A DOCKET MOTION'S RULING, and the transcript's "R1-1: carried"
		// line is a JOIN across both events: the gap is on the filing, the word on the ruling.
		recordtest.Event(t, "red-merge-r1", 1, &recordpb.Motion{
			MotionId: proto.String("M2"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:    proto.String("red cannot settle R1-1"),
			Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("R1-1")}},
		}),
		recordtest.Event(t, "judge-r1", 1, &recordpb.MotionRule{
			MotionId: proto.String("M2"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion:  proto.String("needs a probe"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED),
				Principle:   proto.String("correctness"), Tension: proto.String("cost"),
				ReviewFlag: proto.String("false"),
				Settled:    proto.String("the claim as it stood may not be re-asserted"),
				ReopensOn:  proto.String("the probe"),
			}},
		}),
		// A petition is a MOTION now — the retired `petition` event type has no arm in the schema,
		// so the fixture could only ever have described a state nothing writes.
		recordtest.Event(t, "blue-r1", 1, &recordpb.Motion{
			MotionId: proto.String("M1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
			Basis:    proto.String("the instruction would require asserting what I believe false"),
			Relief:   proto.String("strike the demand from the docket"),
			Filing:   &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_INTEGRITY)}},
		}),
		recordtest.Event(t, "judge-terminal", 0, &recordpb.Halt{Opinion: proto.String("safety gate tripped")}),
		recordtest.Event(t, "judge-terminal", 0, &recordpb.Certify{Statement: proto.String("re-examine the cost model")}),
	}
	d := debate(&record.Board{Events: evs}, evs)
	for _, want := range []string{
		"### Round 1", "### RED — NO VERDICT RECORDED THIS ROUND\ngap A stands", "### BLUE\ngap A repaired",
		"R1-1: carried",
		// THE PETITION SECTION IS NOT HERE ANY MORE, and its absence is the point. It rendered
		// both sides of a petition off the retired `petition`/`petition-rule` types — a second
		// rendering of a dialectic that `## Motions` already shows with each ruling beside the ask
		// it answers. Two carriers of one fact; this was the one reading a vocabulary nothing
		// writes, so it showed nothing while looking complete.
		"### Bench disposition", "**HALT** — safety gate tripped", "**Certification** — re-examine the cost model",
	} {
		if !strings.Contains(d, want) {
			t.Errorf("debate transcript missing %q:\n%s", want, d)
		}
	}
	if empty := debate(&record.Board{}, nil); !strings.Contains(empty, "no debate on the record") {
		t.Errorf("empty debate should say so: %q", empty)
	}
}

// THE BASIS FIELDS EXIST BECAUSE A SEAT ASKED TO SELF-REPORT REPORTS THE FLATTERING VALUE.
// Each was derived rather than claimed, gated a write, and then reached the reader as nothing —
// so a verdict the record itself decided read as the same word as one the bench simply asserted.
func TestVerdictBasisReachesTheReader(t *testing.T) {
	derived := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_VERIFIED), VerdictBasis: proto.String(record.VerdictDerived)}
	if s := verdictGloss(derived); !strings.Contains(s, "derived from the record") {
		t.Errorf("a DERIVED verdict must say so — it is the difference between a mechanical result and a claim: %q", s)
	}
	asserted := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_VERIFIED), VerdictBasis: proto.String(record.VerdictAsserted)}
	s := verdictGloss(asserted)
	if !strings.Contains(s, "asserted by the bench") {
		t.Errorf("an ASSERTED verdict must say so, or it reads as a derived one: %q", s)
	}
	if strings.Contains(s, "derived from the record") {
		t.Errorf("an asserted verdict must not claim derivation: %q", s)
	}
	// An outcome that carries no basis says nothing rather than guessing, because guessing here
	// would invent the very distinction the field preserves.
	if s := verdictGloss(&recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_VERIFIED)}); strings.Contains(s, "basis") {
		t.Errorf("no recorded basis must produce no basis claim: %q", s)
	}
	// EVERY VERDICT BRANCH CARRIES IT. The first cut appended the note only to the default arm,
	// so CEILING and HALTED returned early and dropped it — and a ceiling termination IS derived
	// (rounds against the configured ceiling), and is the most common way a run ends.
	for _, verdict := range []recordpb.RunOutcome{
		recordpb.RunOutcome_RUN_OUTCOME_CEILING,
		recordpb.RunOutcome_RUN_OUTCOME_HALTED,
		recordpb.RunOutcome_RUN_OUTCOME_VERIFIED,
		recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED,
	} {
		o := &recordpb.Outcome{Verdict: verdict.Enum(), VerdictBasis: proto.String(record.VerdictDerived)}
		if s := verdictGloss(o); !strings.Contains(s, "derived from the record") {
			t.Errorf("%s dropped its basis — every terminal verdict says how it was reached: %q", recordpb.Word(verdict), s)
		}
	}
}

// fix_basis reads `verified` only when red supplied an exact span AND replacement that the tool
// validated against the LIVE report — a forced re-read. A demand checked against the document
// read identically to one written from memory of what the document probably said.
func TestFixBasisAndTheConcreteProposalReachTheReader(t *testing.T) {
	verified := &recordpb.Mint{
		FixBasis: proto.String("verified"),
		Location: proto.String("the parser is linear"),
		FixNew:   proto.String("the parser is linear except on backtracking"),
	}
	s := fixProposal(verified)
	for _, want := range []string{"**verified**", "with the text in front of it", "the parser is linear except on backtracking"} {
		if !strings.Contains(s, want) {
			t.Errorf("a verified fix must show its checked replacement; missing %q:\n%s", want, s)
		}
	}
	if p := fixProposal(&recordpb.Mint{FixBasis: proto.String("proposed")}); !strings.Contains(p, "**proposed**") || !strings.Contains(p, "nothing checked this demand") {
		t.Errorf("a prose-only demand must say nothing checked it against the report: %q", p)
	}
	if p := fixProposal(&recordpb.Mint{}); p != "" {
		t.Errorf("no recorded basis must produce no claim: %q", p)
	}
	if p := fixProposal(nil); p != "" {
		t.Errorf("a nil mint must not panic or invent a basis: %q", p)
	}
}

// A PHANTOM RETIREMENT CANCELS REAL LOSS in the scorecard's additive-integrity detector, and
// only the basis distinguishes one from an honest round-0 rewrite.
func TestRemovalBasisReachesTheReader(t *testing.T) {
	v := withdrawnClaims([]*record.Event{recordtest.Event(t, "blue-r2", 0, &recordpb.Retire{Claim: proto.String("the parser is linear"), Reason: proto.String("r"), RemovalBasis: proto.String(record.RemovalVerified)})})
	if !strings.Contains(v, "**verified**") || !strings.Contains(v, "the record shows it leaving") {
		t.Errorf("a verified removal must say the record can show it:\n%s", v)
	}
	a := withdrawnClaims([]*record.Event{recordtest.Event(t, "blue-r0", 0, &recordpb.Retire{Claim: proto.String("the parser is linear"), Reason: proto.String("r"), RemovalBasis: proto.String(record.RemovalAsserted)})})
	if !strings.Contains(a, "**asserted**") || !strings.Contains(a, "nothing on the record shows it was ever present") {
		t.Errorf("an asserted removal must say the record cannot show it:\n%s", a)
	}
}

// AN UNANSWERED PETITION MUST BE LOUD. A petition is a seat's channel for an ethical, safety,
// integrity or constitutional objection, and the engine routes it to a bench sitting BEFORE the
// debate continues. A filing with no ruling means that sitting did not happen — and reporting
// nothing would make it indistinguishable from a run that had no petitions at all.
func TestAnUnansweredPetitionIsReported(t *testing.T) {
	// A PETITION IS A MOTION. This fixture used the retired `petition` type, so after the collapse
	// the detector counted zero filings and could not fire — the warning it exists to raise was
	// unreachable while this test went on passing.
	filed := []*record.Event{recordtest.Event(t, "red-merge-r1", 1, &recordpb.Motion{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION), Basis: proto.String("the demand would bury a hazard")})}
	d := debate(&record.Board{Events: filed}, filed)
	if !strings.Contains(d, "1 petition(s) received no ruling") {
		t.Errorf("a petition with no ruling must be reported, not silently absent:\n%s", d)
	}
	// The ruling joins by MOTION ID, which is the whole point of the collapse: `petition-rule`
	// carried only the petitioner, so two filings by one seat in one round could not be told
	// apart. record.Motions pairs the answer to its ask exactly.
	// The RULING ARM is what makes it answered: a motion-rule whose oneof is unset carried no
	// ruling at all, which is the state the unanswered detector is looking for.
	answered := append(filed, recordtest.Event(t, "judge-r1", 1, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
		Opinion:  proto.String("the hazard is graded, not buried"),
		Ruling:   &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED},
	}))
	if d := debate(&record.Board{Events: answered}, answered); strings.Contains(d, "received no ruling") {
		t.Errorf("an answered petition must not be reported as unanswered:\n%s", d)
	}
}

// A WITHDRAWN CLAIM IS PART OF WHAT THE DEBATE DECIDED. Substance leaves the report only through
// `retire`, which names the claim as it stood and why it went — and the reader saw none of it,
// making a report where a claim was argued and withdrawn identical to one where it was never made.
func TestWithdrawnClaimsReachTheReader(t *testing.T) {
	// The body carries the facts the assertions read. It was left EMPTY by the earlier conversion,
	// so the render had nothing to show and the test was asserting against a blank string.
	evs := []*record.Event{recordtest.Event(t, "blue-respond-r2", 2, &recordpb.Retire{
		Claim:        proto.String("the parser is O(n) in the input size"),
		Reason:       proto.String("refuted at the leaf"),
		SupersededBy: proto.String("the parser is O(n) except on backtracking inputs"),
	})}
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
		"## The Board", "blue cannot know red's findings.", "", // tool-owned — dropped
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
			"R1-1": {ID: "R1-1", Open: true, Severity: recordpb.Grade_GRADE_LOW, Impact: recordpb.Grade_GRADE_LOW, Likelihood: recordpb.Grade_GRADE_LOW,
				Mint: &recordpb.Mint{Problem: proto.String("a minor nit."), RequiredFix: proto.String("tidy it")}},
			"R1-2": {ID: "R1-2", Open: true, Severity: recordpb.Grade_GRADE_CERTAIN, Impact: recordpb.Grade_GRADE_HIGH, Likelihood: recordpb.Grade_GRADE_HIGH,
				Mint: &recordpb.Mint{Problem: proto.String("a load-bearing flaw."), RequiredFix: proto.String("fix the core")}},
			"R1-3": {ID: "R1-3", Open: false, Severity: recordpb.Grade_GRADE_HIGH, // closed — must not appear
				Mint: &recordpb.Mint{Problem: proto.String("already closed.")}},
		},
	}
	evs := []*record.Event{
		recordtest.Event(t, "", 0, &recordpb.Certify{Statement: proto.String("re-examine the cost model before shipping")}),
	}
	o := orientation(board, evs, "")
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
	empty := orientation(&record.Board{}, nil, "")
	if !strings.Contains(empty, "no open gaps remain") {
		t.Errorf("an empty board should say nothing is outstanding:\n%s", empty)
	}
}

func TestUnmintedFindingsSurfaced(t *testing.T) {
	board := &record.Board{
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Mint: &recordpb.Mint{FoundBy: []string{"L5-F1", "L6-F2"}}},
		},
		Events: []*record.Event{
			recordtest.Event(t, "red-lens-r1-L5", 0, &recordpb.Finding{Label: proto.String("L5-F1"), Text: proto.String("minted — omit")}),
			recordtest.Event(t, "red-lens-r1-L6", 0, &recordpb.Finding{Label: proto.String("L6-F2"), Text: proto.String("also minted — omit")}),
			recordtest.Event(t, "red-lens-r1-L5", 0, &recordpb.Finding{Label: proto.String("L5-F3"), Location: proto.String("§H1"), Text: proto.String("un-minted red reasoning kept for the record")}),
		},
	}
	got := boardSection(board)
	if !strings.Contains(got, "Lens findings credited by no gap (1)") {
		t.Errorf("exactly one un-minted finding should be surfaced:\n%s", got)
	}
	// THE SECTION STATES THE JOIN AND NOT A DELIBERATION. It used to say the merge "weighed but
	// did not mint" these, which nothing on the record supports — declining to mint writes no
	// event — and which is what made a silently dropped finding read as a considered decline
	// (#747: three dropped in one run, one of them a fabricated-quote allegation).
	if strings.Contains(got, "weighed but did not mint") {
		t.Error("the section asserts the merge weighed these findings; no event records deliberation, and a silent drop then reads as a decision")
	}
	if !strings.Contains(got, "is NOT recorded") {
		t.Errorf("the section does not say that whether these were considered is unrecorded:\n%s", got)
	}
	if !strings.Contains(got, "L5-F3") || !strings.Contains(got, "un-minted red reasoning kept") {
		t.Errorf("the un-minted finding's substance must be surfaced:\n%s", got)
	}
	if strings.Contains(got, "minted — omit") || strings.Contains(got, "also minted") {
		t.Errorf("a finding credited by a gap's found_by must NOT be re-listed:\n%s", got)
	}
}

// A MINTED FINDING'S EVIDENCE MUST BE REACHABLE FROM THE GAP THAT CLAIMED IT.
//
// The provenance line used to read `surfaced by: L5-F1, L6-F2` and nothing in the report defined
// those labels — unmintedFindings renders a finding only when NO gap claims it, so the instant
// the merge acted on a finding its leaf-level evidence left the document and the citation
// dangled. A fuzz run where every finding was minted put red's words nowhere at all.
//
// It is the wrong half to drop: `problem` is the merge's RESTATEMENT, and a reader can only see
// a restatement drift from its evidence with both in front of them.
func TestAMintedFindingsEvidenceIsQuotedUnderItsGap(t *testing.T) {
	board := &record.Board{
		GapOrder: []string{"R1-1"},
		Gaps: map[string]*record.Gap{
			// OPEN here; the CLOSED case is covered below. Scoping provenance to open gaps was
			// the first answer and it was wrong: unmintedFindings skips anything a gap claimed,
			// so a run where every finding was minted and every gap closed printed red's words
			// nowhere at all. The fuzz found one.
			"R1-1": {ID: "R1-1", Open: true, Mint: &recordpb.Mint{
				Problem: proto.String("the merge's restatement"),
				FoundBy: []string{"L5-F1", "L9-F9"},
			}},
		},
		Events: []*record.Event{
			recordtest.Event(t, "red-lens-r1-L5", 0, &recordpb.Finding{
				Label:    proto.String("L5-F1"),
				Location: proto.String("§H2"),
				Text:     proto.String("what red actually observed at the leaf"),
			}),
		},
	}
	got := boardSection(board)
	if !strings.Contains(got, "what red actually observed at the leaf") {
		t.Errorf("the minted finding's own words are absent — the gap cites L5-F1 and nothing defines it:\n%s", got)
	}
	if !strings.Contains(got, "§H2") {
		t.Errorf("the finding's location is absent, so a reader cannot go and check it:\n%s", got)
	}
	// AND AN UNRESOLVABLE CITATION IS ITSELF WORTH SEEING. A found_by naming no finding on the
	// record must not vanish into a shorter list — that reads exactly like a gap with less
	// provenance, which is the plausible-zero shape.
	if !strings.Contains(got, "L9-F9") || !strings.Contains(got, "no finding with this label") {
		t.Errorf("a found_by label with no finding behind it was dropped silently:\n%s", got)
	}
	// It stays out of the un-minted section either way: the gap is where it is rendered.
	if strings.Contains(got, "Lens findings not raised to a gap") {
		t.Errorf("a claimed finding must not ALSO be listed as un-raised:\n%s", got)
	}

	// AND A CLOSED GAP CARRIES IT TOO. The closure says how the gap was settled; it does not
	// restate what was observed, and an audit of a closure needs both.
	closedBoard := &record.Board{
		GapOrder: []string{"R1-1"},
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Open: false, Mint: &recordpb.Mint{
				Problem: proto.String("the merge's restatement"),
				FoundBy: []string{"L5-F1"},
			}},
		},
		Events: board.Events,
	}
	if got := boardSection(closedBoard); !strings.Contains(got, "what red actually observed at the leaf") {
		t.Errorf("a CLOSED gap dropped the evidence it was minted from, and nothing else renders it:\n%s", got)
	}
}

func TestLogSectionRendered(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "red-merge-r1", 0, &recordpb.Log{Text: proto.String("the --cx flag is missing from help"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		recordtest.Event(t, "blue-respond-r2", 0, &recordpb.Log{Text: proto.String("manifest cap fights methodology gaps"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		// A NOMINAL entry renders in its own section, not among the problems: an attestation is
		// not a complaint, and the split is by TYPE now rather than by message.
		recordtest.Event(t, "judge-r2", 0, &recordpb.Log{Text: proto.String("the surface met the work"), Type: recordpb.LogType_LOG_TYPE_NOMINAL.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		recordtest.Event(t, "red-merge-r1", 0, &recordpb.Mint{Problem: proto.String("not a log entry")}),
	}
	f := logSection(evs)
	// THE TYPE RENDERS BESIDE THE SEAT, which is the whole point of the channel: an operator
	// triages by reading the type, not by reading the prose to work out which kind it was.
	for _, want := range []string{
		"Log (what the run told the operator",
		"**red-merge-r1** (defect): the --cx flag is missing",
		"**blue-respond-r2** (defect): manifest cap fights",
		"**judge-r2**: the surface met the work",
	} {
		if !strings.Contains(f, want) {
			t.Errorf("log section missing %q:\n%s", want, f)
		}
	}
	if empty := logSection(nil); empty != "" {
		t.Errorf("no log events should render nothing, got: %q", empty)
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
	evs := []*record.Event{
		recordtest.Event(t, "blue-respond-r1", 1, &recordpb.Revision{Text: proto.String("expanded the caching section; retired the stale figure")}),
		recordtest.Event(t, "blue-respond-r2", 2, &recordpb.Revision{Text: proto.String("addressed R2-1 in the analysis")}),
		recordtest.Event(t, "red-merge-r1", 1, &recordpb.Position{Text: proto.String("not a revision")}),
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

// BLUE IS NOT CHARGED FOR THE BENCH'S CLOSURES.
//
// correctnessManifest listed every closed gap with no manifest row as "a repair nobody audited,
// including its author". Bench-disposed gaps were in that set, and blue never repaired them —
// the report accused one party of skipping an audit another party never owed it.
func TestTheBenchsOwnClosuresAreNotBluesUnauditedRepairs(t *testing.T) {
	board := &record.Board{
		GapOrder: []string{"R1-1", "R1-2"},
		Gaps: map[string]*record.Gap{
			// Blue repaired this one and filed no receipt: a real missing manifest row.
			"R1-1": {ID: "R1-1", HasClosed: true, Closure: &recordpb.Close{}},
			// The bench disposed of this one. Nobody repaired it, so no receipt is owed.
			"R1-2": {ID: "R1-2", HasClosed: true, BenchClosure: &recordpb.DocketRuling{}, ClosedByBench: true},
		},
	}
	got := correctnessManifest(board)
	if !strings.Contains(got, "R1-1") {
		t.Errorf("blue's own unaudited repair left the list:\n%s", got)
	}
	if strings.Contains(got, "R1-2") {
		t.Errorf("the report still charges blue for a gap the BENCH closed:\n%s", got)
	}
}

// THE ORDERING THAT MAKES ClosedByBench THE WRONG KEY, asserted because the cheap fix passes
// every other test in this file.
//
// Blue closes a gap and files no manifest row; the bench later rules on the same gap. The flag
// follows the LAST closing event, so ClosedByBench is true — while blue's `close` is still the
// closure and the receipt is still genuinely missing. Keyed on the flag, this row vanishes from
// the one section whose stated purpose is to report exactly this.
func TestABlueRepairTheBenchLaterRuledOnIsStillCharged(t *testing.T) {
	g := &record.Gap{ID: "R1-1", HasClosed: true, Closure: &recordpb.Close{}}
	g.BenchClosure = &recordpb.DocketRuling{}
	g.ClosedByBench = true

	got := correctnessManifest(&record.Board{GapOrder: []string{"R1-1"}, Gaps: map[string]*record.Gap{"R1-1": g}})
	if !strings.Contains(got, "R1-1") {
		t.Errorf("a missing receipt disappeared because the bench later ruled on the gap:\n%s", got)
	}
}
