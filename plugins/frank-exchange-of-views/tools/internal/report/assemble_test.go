package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"google.golang.org/protobuf/proto"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func TestTheTextIsPartedVerbatimAndFenceAware(t *testing.T) {
	md := strings.Join([]string{
		"# report", "",
		"## The Catechism", "", "Q1: kept  as-is.  ", "", // trailing spaces preserved inside
		"```", "## Open questions", "not a heading — inside a fence", "```", "",
		"## Analysis", "", "the analysis.", "",
	}, "\n")
	got := part(md)
	if !strings.Contains(got.body, "Q1: kept  as-is.  \n") {
		t.Errorf("the body is not the text as written: %q", got.body)
	}
	// The fenced "## Open questions" does not cut the text: the lines after it are the body's.
	if !strings.Contains(got.body, "not a heading — inside a fence") || !strings.HasSuffix(got.body, "the analysis.") {
		t.Errorf("a ## inside a code fence wrongly cut the text: %q", got.body)
	}
	// A real "## Open questions" section does not exist (only the fenced one), so the row ships
	// NOTHING: no text is authored in its place, and no line about the run or its seats stands in
	// the reader's document for it (universe-m21 shipped two such lines).
	if got.close != "" {
		t.Errorf("a fenced heading was mistaken for a real one, or a missing section was given text: %q", got.close)
	}
	// ONE READING OF A HEADING, for the cut and for everything else that reads one: case, runs
	// of whitespace, an anchor on the heading line, the note reference the weave leaves where a
	// citation's stood and a trailing "(in full)" do not make a present section absent.
	for _, h := range []string{"## Open Questions", "##  open   questions", "## Open questions<!--fx:F-00000aaa-->", "## Open questions[^2]", "## Open questions (in full)"} {
		if got := part("intro\n" + h + "\n\nwhat remains.\n"); got.body != "intro" || got.close != h+"\n\nwhat remains." {
			t.Errorf("%q: the text is not cut at its closing section: body %q, close %q", h, got.body, got.close)
		}
	}
}

func TestTheTitleIsTakenVerbatimOrFlagged(t *testing.T) {
	if got := titleOr(part("intro\n# Whether X — research report\n## The Catechism\n").title); got != "# Whether X — research report" {
		t.Errorf("the title is not the line as written: %q", got)
	}
	if got := titleOr(part("no title here\n## Analysis\n```\n# a comment in code\n```\n").title); !strings.Contains(got, "not authored here") {
		t.Errorf("a missing title must be flagged, never authored: %q", got)
	}
}

func TestRiskMatrixFromBoard(t *testing.T) {
	longProblem := "Blue claims JSON float-loss above 2^53 causes H1 failure. Event logs serialize timestamps as ISO 8601 strings, so it does not manifest."
	bj := record.BoardJSON{Open: []record.GapJSON{
		{ID: "G1", Problem: "overclaims capture", Likelihood: "high", Impact: "medium", RequiredFix: "grep the sites"},
		{ID: "G2", Problem: longProblem, Likelihood: "low", Impact: "low", RequiredFix: "verify serialization"},
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
	// A missing outcome is STATE, in the same stamp shape as the others (gblock: "State only") —
	// never invented, and never narrated here: the missing act is run.md's to say.
	if s := outcomeStamp(nil); s != "**Outcome:** PENDING (no outcome on the record yet — the bench records it at the run's exit)" {
		t.Errorf("a missing outcome must be the state PENDING, in the stamp's shape: %q", s)
	}
	// THE STAMP IS A FIELD. It carries the word and the clause naming how the run ended, then — since
	// gblock's 2026-09-10 ruling — its basis as state in parentheses. The word stays the FIRST token, so
	// a reader who skims, badges or greps it still finds one. The argument
	// that used to sit inline is verdictGloss, asserted directly below.
	ceiling := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING)}
	if s := outcomeStamp(ceiling); s != "**Outcome:** CEILING-TERMINATED" {
		t.Errorf("the verdict field must be the word alone: %q", s)
	}
	if g := verdictGloss(ceiling); !strings.Contains(g, "CEILING-TERMINATED") || !strings.Contains(g, "NOT a judged failure to verify") || !strings.Contains(g, "travels OUT of the run") {
		t.Errorf("the CEILING gloss must name the re-audit debt and not read as a failure: %q", g)
	}
	// THE CAUSE IS THE RECORD'S: the gloss carries the derivation's reason for THIS ceiling and
	// asserts no other cause.
	epochLimit := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING), VerdictBasis: proto.String(record.VerdictDerived),
		VerdictWhy: proto.String("epoch limit 12 reached — the run's term; the parties ready at the chair's last sitting were not dispatched and PASS is not permitted")}
	if g := verdictGloss(epochLimit); !strings.Contains(g, "(epoch limit 12 reached") || strings.Contains(g, "every open material gap reached its limit") {
		t.Errorf("an epoch-limit ceiling's gloss must state the epoch limit and assert no gap limit: %q", g)
	}
	// THE BASIS RIDES THE STAMP AS STATE ("State on the stamp, prose to run.md"): the word first, the
	// basis in parentheses, and not one word of what it means — that is basisNote, in run.md.
	derived := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_VERIFIED), VerdictBasis: proto.String(record.VerdictDerived)}
	if s := outcomeStamp(derived); s != "**Outcome:** VERIFIED (derived from the record)" {
		t.Errorf("a derived verdict's stamp must carry its basis as state: %q", s)
	}
	asserted := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED), VerdictBasis: proto.String(record.VerdictAsserted)}
	if s := outcomeStamp(asserted); s != "**Outcome:** UNVERIFIED (asserted by the bench)" {
		t.Errorf("an asserted verdict's stamp must carry its basis as state: %q", s)
	}
	halted := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED)}
	if s := outcomeStamp(halted); s != "**Outcome:** HALTED" {
		t.Errorf("the verdict field must be the word alone: %q", s)
	}
	if g := verdictGloss(halted); !strings.Contains(g, "HALTED") || !strings.Contains(g, "halt opinion") {
		t.Errorf("the HALTED gloss must point at the recorded halt opinion: %q", g)
	}
	// A missing outcome is stated by the STAMP as PENDING, and assembly does not repeat it into
	// "Read this first" — the gloss exists to carry an argument the field cannot hold, and
	// there is no argument here.
	if g := verdictGloss(nil); !strings.Contains(g, "no outcome on the record yet") {
		t.Errorf("a missing outcome must still be answerable from the gloss: %q", g)
	}
	// UNVERIFIED is the word alone, and the bench's ACCOUNT of why is not in the research document
	// at all. `outcome --reason` is seat-authored prose written after red's last sitting; it is
	// envelope, and it renders with its siblings under the transcript's Bench disposition. What
	// stays here is the TOOL's own derivation reasoning, which no seat wrote.
	unverified := &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED), Prose: proto.String("the workflow stopped with nobody ready")}
	if s := outcomeStamp(unverified); s != "**Outcome:** UNVERIFIED" {
		t.Errorf("the verdict field must be the word alone: %q", s)
	}
	if g := verdictGloss(unverified); strings.Contains(g, "nobody ready") {
		t.Errorf("the bench's own prose reached the research document: %q", g)
	}
}

// THE ROW'S TAG IS THE FATE THE SEAT RECORDED, WHATEVER PATH LED TO IT. `abandoned` means tried and
// died, and a seat that abandons straight from `proposed` chose that word for an inquiry it ran.
// `confirmed` straight from `proposed` gets the plain tag too.
func TestAnAbandonedRowCarriesThePlainFateWhateverItsPath(t *testing.T) {
	board := famOf(t, []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Question: proto.String("sieve to the square root")}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED), Reason: proto.String("the bound is 9, so the sieve adds nothing")}),
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q2"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("check a computer-algebra system")}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q2"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED), Reason: proto.String("no such system is available here")}),
	})
	alt := inquirySection(t, board, "Alternatives considered")
	const fate = "\n  - **Fate:** abandoned. **Why:** "
	if !strings.Contains(alt, "- **Question:** sieve to the square root"+fate+"the bound is 9, so the sieve adds nothing") {
		t.Errorf("an inquiry underway and then abandoned keeps the plain fate:\n%s", alt)
	}
	if !strings.Contains(alt, "- **Question:** check a computer-algebra system"+fate+"no such system is available here") {
		t.Errorf("an inquiry abandoned straight from proposed carries the fate its seat recorded:\n%s", alt)
	}
	if strings.Contains(alt, "before pursuit") {
		t.Errorf("the row says something about the path that no seat recorded:\n%s", alt)
	}
}

// THE ROW SAYS WHAT WAS ASKED, WHAT WAS EXPECTED, WHAT BECAME OF IT, BY WHAT METHOD, AND WHAT WAS
// FOUND, each under its own label, so a reader cannot take the question or the hypothesis for a
// claim of the report's.
func TestAnInquiryRowLabelsItsQuestionHypothesisFateAndFinding(t *testing.T) {
	av := func(id string, st recordpb.InquiryStatus, line, hyp, reason string) *record.Event {
		return recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String(id), Status: st.Enum(),
			Question: proto.String(line), Hypothesis: proto.String(hyp), Reason: proto.String(reason)})
	}
	// THE METHOD IS THE LATEST FATE'S. Q2's earlier fate states one and its latest states none, so
	// its row prints none: a method is the account of the fate it was recorded with.
	board := famOf(t, []*record.Event{
		av("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, "whether one Fermat test settles it", "a passing test is no evidence of primality", ""),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"), Method: proto.String("a Fermat test in base 3"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Reason: proto.String("a passing test is evidence, not proof")}),
		av("Q2", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, "whether trial division settles it", "", ""),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q2"), SupersedesStatus: proto.String("1"), Method: proto.String("trial division to 7"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Reason: proto.String("2, 3 and 5 do not divide 91")}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q2"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED), Reason: proto.String("7 divides 91")}),
		av("Q3", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, "whether a sieve is faster", "it is", ""),
		av("Q4", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, "whether a lookup table exists", "one does", ""),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q4"), SupersedesStatus: proto.String("1"), Method: proto.String("Look 91 up in a published table of primes."),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED), Reason: proto.String("the bound is 9")}),
	})
	research := inquirySection(t, board, "Research areas")
	for name, want := range map[string]string{
		"a refuted inquiry prints its hypothesis as one, the method its fate states, and says it did not hold": "- **Question:** whether one Fermat test settles it\n" +
			"  - **Hypothesis:** a passing test is no evidence of primality\n" +
			"  - **Fate:** refuted. **Method:** a Fermat test in base 3. **Found:** a passing test is evidence, not proof",
		"an inquiry with no hypothesis prints no hypothesis line, and a latest fate with no method prints none": "- **Question:** whether trial division settles it\n" +
			"  - **Fate:** confirmed. **Found:** 7 divides 91",
		"an inquiry still proposed has found nothing and holds no reason": "- **Question:** whether a sieve is faster\n" +
			"  - **Hypothesis:** it is\n" +
			"  - **Fate:** proposed.",
	} {
		if !strings.Contains(research+"\n", want+"\n") {
			t.Errorf("%s; want the row\n%s\nin\n%s", name, want, research)
		}
	}
	if strings.Contains(research, "**whether") {
		t.Errorf("a question is printed in bold, as a claim of the report's would be:\n%s", research)
	}
	alt := inquirySection(t, board, "Alternatives considered")
	// A method that already ends a sentence is not given a second full stop.
	if want := "  - **Hypothesis:** one does\n  - **Fate:** declined. **Method:** Look 91 up in a published table of primes. **Why:** the bound is 9"; !strings.Contains(alt, want) {
		t.Errorf("an approach that did not run found nothing: its reason is why it stands there, after the method its fate states; want\n%s\nin\n%s", want, alt)
	}
}

// A ROW IS ONE LIST ITEM HOWEVER MANY PARAGRAPHS ITS PROSE HOLDS. The archived sleeper-service-plan
// and quadratic-formula runs hold lines written as two and three paragraphs; printed as written,
// the blank line ended the item, the later paragraphs stood outside the list, and the hypothesis
// and fate opened a list of their own. Markdown and the site read the same rule: a line indented
// under an item, with no blank line before it, continues the item.
func TestAnInquiryRowOfSeveralParagraphsStaysOneRow(t *testing.T) {
	board := famOf(t, []*record.Event{recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
		Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Method: proto.String("census"),
		Question: proto.String("APPROACH: census the plan.\n\nARGUMENT: ordering is scarce."), Hypothesis: proto.String("one exists\nand it is explicit"),
		Reason: proto.String("none exists.\n\nThe plan names no function.")})})
	got := inquirySection(t, board, "Research areas")
	want := "- **Question:** APPROACH: census the plan.  \n  ARGUMENT: ordering is scarce.\n" +
		"  - **Hypothesis:** one exists  \n    and it is explicit\n" +
		"  - **Fate:** refuted. **Method:** census. **Found:** none exists.  \n    The plan names no function."
	if !strings.Contains(got, want) {
		t.Fatalf("want the row\n%s\ngot\n%s", want, got)
	}
	html := renderOne(got, FileReport, anchors{}, map[string]int{}, nil).html
	if n := strings.Count(html, "<ul>"); n != 2 {
		t.Errorf("the row opens %d lists on the site; one row is one item holding one nested list:\n%s", n, html)
	}
	if strings.Contains(html, "<p>") {
		t.Errorf("a paragraph of the row stands outside its list on the site:\n%s", html)
	}
}

// EVERY FATE PRINTS AS THE WORD ITS SEAT FILED, WITH NO GLOSS. A gloss says more than the record
// holds — an `abandoned` whose reason is a refutation, an `underway` that is where
// an approach came to rest — so the word stands alone. Only an approach that ran reports what it
// FOUND. A fate the enum gains fails here until its label is decided.
func TestEveryInquiryFateIsSaidOnItsRow(t *testing.T) {
	label := map[string]string{
		"proposed": "Why", "underway": "Found", "confirmed": "Found", "refuted": "Found",
		"deferred": "Why", "declined": "Why", "abandoned": "Why",
	}
	for _, status := range record.InquiryStatusNames() {
		st := recordpb.InquiryStatus(recordpb.InquiryStatus_value["INQUIRY_STATUS_"+strings.ToUpper(status)])
		board := famOf(t, []*record.Event{recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
			Status: st.Enum(), Question: proto.String("whether a sieve is faster"), Reason: proto.String("the bound is 9")})})
		want := "- **Question:** whether a sieve is faster\n  - **Fate:** " + status + ". **" + label[status] + ":** the bound is 9"
		var got string
		for _, heading := range []string{"Research areas", "Future research directions", "Alternatives considered"} {
			if s := inquirySection(t, board, heading); strings.Contains(s, "**Question:**") {
				got = s
			}
		}
		if !strings.Contains(got+"\n", want+"\n") {
			t.Errorf("the fate %q: want the row\n%s\ngot\n%s", status, want, got)
		}
	}
}

func TestInquiriesSplitByFate(t *testing.T) {
	board := famOf(t, []*record.Event{
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Question: proto.String("profile the hot path"), Method: proto.String("bench")}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q2"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED), Question: proto.String("rewrite in Rust"), Reason: proto.String("cost exceeds benefit")}),
		recordtest.Event(t, "red-lens-r1", &recordpb.Inquiry{InquiryId: proto.String("Q3"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED), Question: proto.String("third-party audit"), Reason: proto.String("out of scope")}),
	})
	exp := inquirySection(t, board, "Research areas")
	if !strings.Contains(exp, "profile the hot path") || strings.Contains(exp, "rewrite in Rust") {
		t.Errorf("research areas must carry ONLY underway and proposed inquiries:\n%s", exp)
	}
	alt := inquirySection(t, board, "Alternatives considered")
	if !strings.Contains(alt, "rewrite in Rust") || !strings.Contains(alt, "cost exceeds benefit") {
		t.Errorf("an abandoned inquiry is an alternative considered, its reason the counter:\n%s", alt)
	}
	if !strings.Contains(alt, "third-party audit") {
		t.Errorf("a declined inquiry is also an alternative considered:\n%s", alt)
	}
	if strings.Contains(alt, "profile the hot path") {
		t.Errorf("an underway inquiry must not appear under alternatives:\n%s", alt)
	}
	// No inquiries of a fate → the heading stays, and the section states the absence in the subject's
	// terms, each section its own sentence.
	empty := record.NewFamily(nil, record.Merged{})
	for heading, want := range map[string]string{
		"Research areas":             "No approach was put forward or pursued.",
		"Future research directions": "No approach was kept for later research.",
		"Alternatives considered":    "No approach was weighed and set aside.",
	} {
		if got := inquirySection(t, empty, heading); !strings.Contains(got, "## "+heading) || !strings.Contains(got, want) || strings.Contains(got, "record") {
			t.Errorf("an empty %s must say %q, in the subject's terms: %q", heading, want, got)
		}
	}
}

// EVERY STATUS LANDS WHERE ITS FATE SAYS — three sections, and one deliberate exclusion.
//
// # It ranges over every status
//
// It ranges over InquiryStatusNames(), so each status is exercised.
//
// # The model this asserts: the lifecycle of a research topic through the report
//
//	underway             "Research areas" — a topic followed, and what it yielded
//	deferred             "Future research directions" — KEPT for a later run or a deeper context,
//	                     which is a decision. Filing it under "Alternatives considered" said the
//	                     opposite of its fate, and only the `[deferred]` tag on the row
//	                     contradicted the heading above it.
//	confirmed, refuted  "Research areas" — an approach that ran and answered, whichever way
//	declined, abandoned  Alternatives considered — weighed and never run, or started with no answer
//	proposed             "Research areas", beside `underway`, with `[proposed]` on its own row.
//	                     Three areas is the decision: an inquiry blue put forward IS an area this
//	                     run is researching, and red's per-round inquiry review is what stops it
//	                     sitting undecided — not a heading describing the omission.
//
// EVERY status lands in EXACTLY ONE section, so a sixth that matches no predicate fails here
// rather than vanishing the way `proposed` and `deferred` once did.
func TestEveryInquiryStatusLandsWhereItsFateSays(t *testing.T) {
	section := map[string]string{
		"underway":  "research",
		"confirmed": "research", // an inquiry followed to its end is one the research took
		"refuted":   "research", // a refutation is an answer: it stands beside the confirmations
		"proposed":  "research", // an undecided inquiry IS an area this run is researching
		"deferred":  "future",
		"declined":  "alternatives",
		"abandoned": "alternatives",
	}
	for _, status := range record.InquiryStatusNames() {
		st, ok := record.InquiryStatusOf(status)
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
		board := famOf(t, []*record.Event{
			recordtest.Event(t, "blue-r1", &recordpb.Inquiry{
				InquiryId: proto.String("Q1"), Status: st.Enum(), Question: proto.String("the only inquiry"),
			}),
		})
		in := map[string]bool{
			"research":     strings.Contains(inquirySection(t, board, "Research areas"), "the only inquiry"),
			"future":       strings.Contains(inquirySection(t, board, "Future research directions"), "the only inquiry"),
			"alternatives": strings.Contains(inquirySection(t, board, "Alternatives considered"), "the only inquiry"),
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
			t.Errorf("status %q rendered under %v, want exactly [%s] — an inquiry in two sections is an alternative to "+
				"itself, and one in NONE loses the seat's recorded prose entirely", status, got, want)
		case got[0] != want:
			t.Errorf("status %q rendered under %q, want %q", status, got[0], want)
		}
		// AND THE SENTENCE A SEAT READS SAYS THE SAME. InquirySections is printed on the pages of
		// the acts that write an inquiry, on the page that prints the report and in blue's prompts;
		// it names each heading and then the fates it holds, so a status belongs to the clause that
		// opens at its heading.
		heading := map[string]string{"research": "Research areas", "future": "Future research directions", "alternatives": "Alternatives considered"}
		said := ""
		for name, h := range heading {
			from := strings.Index(InquirySections, h)
			if from < 0 {
				t.Fatalf("InquirySections does not name %q: %s", h, InquirySections)
			}
			clause := InquirySections[from+len(h):]
			for _, other := range heading {
				if i := strings.Index(clause, other); i >= 0 {
					clause = clause[:i]
				}
			}
			if strings.Contains(clause, status) {
				said += name
			}
		}
		if said != want {
			t.Errorf("InquirySections puts %q under %q and the assembler under %q: %s", status, said, want, InquirySections)
		}
	}
}

// inquirySection renders the inquiry section under heading, through its row.
func inquirySection(t *testing.T, fam record.Family, heading string) string {
	t.Helper()
	return sectionBlock(t, heading, &assembly{fam: fam})
}

// sectionBlock is the section under heading as assembly places it, through its row.
func sectionBlock(t *testing.T, heading string, a *assembly) string {
	t.Helper()
	for _, s := range layout {
		if s.Heading == heading && s.compose != nil {
			return s.block(a)
		}
	}
	t.Fatalf("no composed section %q in the layout", heading)
	return ""
}

// EVERY SECTION THE TOOL COMPOSES IS NAMED TO BLUE AS THE TOOL'S, AND NOTHING ELSE IS. The
// prompts' list is read as the whole of it: on universe-m20 it named the risk matrix, the board and
// the debate, and blue wrote "## Alternatives considered" by hand. debate.js is a hand-authored
// script the workflow runs whole, with no import step and no generator writing into it, so its
// list cannot be generated from the table; it is held to it here, both ways, through the prompts
// the simulator suite renders and pins.
func TestBluesPromptsNameTheSectionsTheToolComposes(t *testing.T) {
	var named []string
	for _, s := range layout {
		if s.refused() && s.Line == "" {
			named = append(named, "`## "+s.Heading+"`")
		}
	}
	// The list whole, in the table's order: a heading the table lacks, or a row the list lacks,
	// is a different string.
	list := strings.Join(named, ", ")
	for _, seat := range []string{"blue-synthesize", "blue-respond"} {
		prompt := strings.ReplaceAll(renderedPrompt(t, seat), "\n", " ")
		if !strings.Contains(prompt, list) {
			t.Errorf("the %s prompt does not name the sections the tool composes as the table has them: %s", seat, list)
		}
		// AND NOTHING ELSE AS THE TOOL'S: a heading appended to the list leaves the table's list
		// inside it. Every `## ` heading the prompt names is a row of the table, the wrapper it
		// forbids beside them, or one of the report's own it tells blue to write.
		for _, m := range regexp.MustCompile("`## ([^`]+)`").FindAllStringSubmatch(prompt, -1) {
			placed, _ := textRows()
			row := slices.ContainsFunc(append(placed, layout...), func(s Section) bool { return s.Heading == m[1] && (s.refused() || s.Composer != Tool) })
			if !row && m[1] != "Blue team report" {
				t.Errorf("the %s prompt names `## %s`, which is no section the table has the tool refuse or holds a place in the report's text for", seat, m[1])
			}
		}
		for _, s := range layout {
			if s.Composer == Tool && s.Line != "" && !strings.Contains(prompt, "`"+s.Line+"`") {
				t.Errorf("the %s prompt does not name the %s line the tool composes", seat, s.Line)
			}
		}
	}
	// The inquiry rule reaches the seat that records an approach's fate in the table's own sentence.
	if seat := "blue-lane-1"; !strings.Contains(strings.ReplaceAll(renderedPrompt(t, seat), "\n", " "), InquirySections) {
		t.Errorf("the %s prompt does not state which fate lands under which heading as the table does (InquirySections): %s", seat, InquirySections)
	}
}

// EVERY ROW SAYS WHO COMPOSES THE SECTION, FROM WHAT, AND HOW CONTENT REACHES IT. A refusal at
// the write prints the row's Path as the way in, so a row without one refuses and leaves the seat
// nowhere to take the content.
func TestEveryRowOfTheLayoutSaysHowContentReachesIt(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range layout {
		if s.File == "" || s.Heading == "" || s.Composer != Blue && s.Composer != Tool {
			t.Errorf("row %+v: a file, a heading and a composer are required", s)
		}
		if len(strings.Fields(s.From)) < 3 || len(strings.Fields(s.Path)) < 4 || !strings.HasSuffix(s.Path, ".") {
			t.Errorf("row %q: from %q, path %q — each row says what the section is composed from and, in whole sentences, how content reaches it", s.Heading, s.From, s.Path)
		}
		if key := normalizeHeading(s.Heading); seen[key] {
			t.Errorf("two rows read as the heading %q", s.Heading)
		} else {
			seen[key] = true
		}
		if s.fate != nil {
			if !strings.Contains(s.From, "the question it asks, its hypothesis, its fate with the method it states, and what was found or why it stands there") {
				t.Errorf("inquiry section %q: its row does not say what a row of it holds: from %q", s.Heading, s.From)
			}
			for _, f := range fatesOf(s) {
				if !strings.Contains(s.Path, "`"+f+"`") || !strings.Contains(s.From, f) {
					t.Errorf("inquiry section %q holds %q and its row does not say so: from %q, path %q", s.Heading, f, s.From, s.Path)
				}
			}
		}
	}
}

// WHICH HEADINGS THE REPORT'S TEXT MAY NOT CARRY IS A RULING, written out here so a row that
// changes sides is seen. Refused: every heading composed into report.md, where the report's own
// would stand beside the tool's, and each section of another document that is an account of the
// run — red's readings of the citations and blue's edits among them. Not refused: Proofs and
// Motions, composed into other documents under names a subject uses.
func TestTheHeadingsRefusedInTheReportsTextAreTheRuling(t *testing.T) {
	refused := []string{"Outcome", "Risk matrix", "Research areas", "Future research directions", "Alternatives considered",
		"Bibliography", "The board", "The debate", "Citations", "How this run was conducted", "The verdict's basis",
		"Log (what the run told the operator)", "Record verification", "Claims retired", "Superseded bench statements",
		"Edits to the report"}
	if got := RefusedHeadings(); !slices.Equal(got, refused) {
		t.Errorf("the refused headings are\n  %q\nwant\n  %q", got, refused)
	}
	own := map[string]string{"Proofs": FileEvidence, "Motions": FileJudgments}
	for _, s := range layout {
		if file, ok := own[s.Heading]; s.SubjectHeading != ok || ok && s.File != file {
			t.Errorf("row %q in %s: SubjectHeading is %v", s.Heading, s.File, s.SubjectHeading)
		}
		// In report.md the report's own section would stand beside the tool's, and a section of
		// the report's own has no refusal for the field to lift.
		if s.SubjectHeading && (s.File == FileReport || s.Composer != Tool) {
			t.Errorf("row %q is marked a subject's heading too, and it is composed by %s in %s", s.Heading, s.Composer, s.File)
		}
	}
	if want := "Assembly also composes Motions in judgments.md and Proofs in evidence.md, and those names are not refused: " + OwnInTheReport + "."; SubjectHeadingsNote != want {
		t.Errorf("the note on the help pages reads\n  %s\nwant\n  %s", SubjectHeadingsNote, want)
	}
}

// THE NOTE LABELS THE WRITE REFUSES ARE THE ONES THE WEAVE WRITES, every one and no other: a label
// the weave can write and the write lets through is a hand note that can meet the tool's, and a
// label the write refuses and the weave never writes is a footnote of the report's own turned away.
func TestTheRefusedNoteLabelsAreTheOnesTheWeaveWrites(t *testing.T) {
	for n := 1; n < 2000; n += 7 {
		for _, l := range []string{citationNote(n), proofNote(n)} {
			for _, doc := range []string{"A claim" + l + ".\n", "A claim.\n\n" + l + ": a note.\n", "## TL;DR\n\n| a" + l + " | b |\n"} {
				if got, s, found := WovenNoteIn("", doc); !found || got != l || s.Heading != bibliographyHeading {
					t.Errorf("WovenNoteIn(%q) = %q in %q, %v; the weave writes %s", doc, got, s.Heading, found, l)
				}
			}
		}
	}
	// The weave itself, over a citation and a proof: every label in what it returns is one.
	woven, _ := weaveProofRefs(weaveCitations("One<!--cite:C-00000001-->. Two<!--proof:P-00000001-->.\n", nil), nil)
	labels := regexp.MustCompile(`\[\^[^\]]*\]`).FindAllString(woven, -1)
	if len(labels) != 4 {
		t.Fatalf("the weave wrote %d label(s), want a reference and a definition for each anchor:\n%s", len(labels), woven)
	}
	for _, l := range labels {
		if got, _, found := WovenNoteIn("", l); !found || got != l {
			t.Errorf("the weave wrote %s, and the write does not refuse it", l)
		}
	}
	if got, _, _ := WovenNoteIn("", "A claim[^p3].\n"); got != "[^p3]" {
		t.Errorf("a proof's label in lower case reads %q; a renderer folds a label's case", got)
	}
	for _, doc := range []string{
		"A claim[^scope].\n\n[^scope]: an aside.\n",
		"A claim[^a1], another[^1a], a third[^P], a fourth[^Px1], a fifth[^note-2].\n",
		"The class `[^1]` and the class `[^P2]`.\n",
		"```\n[^1]: quoted output\n```\n",
		"~~~\nA claim[^P1].\n~~~\n",
		"Item [1], power 2^3, set [^0-9].\n",
	} {
		if l, _, found := WovenNoteIn("", doc); found {
			t.Errorf("WovenNoteIn(%q) found %q; the weave writes no such label", doc, l)
		}
	}
	if !strings.Contains(WovenNotesRule, notesRow().Path) || !strings.Contains(WovenNotesRule, NoteInCode) {
		t.Errorf("the rule on the help pages does not carry the Bibliography row's path and the place of a label that is no footnote: %s", WovenNotesRule)
	}
}

// THE REFUSAL READS CODE AS BOTH SHIPPED READINGS DO (#1580). A label in an indented code block, or in
// a code span of two backticks running over a line break, is code to them and no footnote; a label on
// the line after a comment is the comment's paragraph to them — report.html has no HTML grammar, and
// the shipped .md escapes the comment — and is a footnote there, though a reader with the grammar ends
// the comment's block at its line and reads the indented line as code.
func TestWovenNoteInReadsCodeAsTheShippedReadingsDo(t *testing.T) {
	for _, doc := range []string{"Para.\n\n    grep '[^1]'\n", "A ``class\n[^P2]`` here.\n"} {
		if l, _, found := WovenNoteIn("", doc); found {
			t.Errorf("WovenNoteIn(%q) found %q in code", doc, l)
		}
	}
	if l, _, found := WovenNoteIn("", "x\n\n<!-- c -->\n    see [^1]\n"); !found || l != "[^1]" {
		t.Errorf("a label both shipped readings take as prose is not refused: %q, %v", l, found)
	}
}

// THE REFUSAL READS A HEADING AS THE ASSEMBLER DOES — the same key, the same fences — and finds
// only what the tool composes.
func TestComposedInFindsWhatTheToolComposes(t *testing.T) {
	for _, tc := range []struct{ doc, line, heading string }{
		{"# t\n\n## TL;DR\nx\n\n## Alternatives  Considered\ny\n", "## Alternatives  Considered", "Alternatives considered"},
		{"## risk matrix<!--fx:F-00000aaa-->\n", "## risk matrix<!--fx:F-00000aaa-->", "Risk matrix"},
		{"  ## The Board (in full)\n", "## The Board (in full)", "The board"},
		{"# t\n**outcome:** VERIFIED\n\n## TL;DR\nx\n", "**outcome:** VERIFIED", "Outcome"},
	} {
		line, s, found := ComposedIn("", tc.doc)
		if !found || line != tc.line || s.Heading != tc.heading {
			t.Errorf("ComposedIn(%q) = %q, %q, %v; want %q in %q", tc.doc, line, s.Heading, found, tc.line, tc.heading)
		}
	}
	for _, s := range layout {
		_, got, found := ComposedIn("", "## TL;DR\nx\n\n## "+s.Heading+"\ny\n")
		if found != s.refused() || found && got.Heading != s.Heading {
			t.Errorf("a report carrying `## %s`: found %v (%q), and the table says its composer is %s, a subject's heading too: %v", s.Heading, found, got.Heading, s.Composer, s.SubjectHeading)
		}
	}
	for _, doc := range []string{
		"# t\n\n## TL;DR\nx\n\n## Limits of the answer\ny\n",
		"## Analysis\n\n### Alternatives considered\ny\n",
		"## Analysis\n\n```\n## Risk matrix\n```\n",
		"## Alternatives considered and rejected\n",
		"## Analysis\n\n**Outcome:** the test passes.\n",
		// No heading stands above the report's own sections, so none is the tool's to refuse there.
		"## Additional findings\n",
		"## Technical foundations\nx\n\n## Analysis\ny\n\n## " + bodyHeading + "\nz\n",
	} {
		if line, _, found := ComposedIn("", doc); found {
			t.Errorf("ComposedIn(%q) found %q; that text claims no section the tool composes", doc, line)
		}
	}
}

// A RECORDED INQUIRY IS RENDERED ONCE. Reading raw events would render an inquiry underway in
// epoch 0 and abandoned in epoch 2 under BOTH headings — as an expansion and as an alternative to
// itself.
func TestARecordedInquiryIsRenderedOnce(t *testing.T) {
	board := famOf(t, []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Question: proto.String("rewrite the parser")}),
		// Two chair sittings put the fate in epoch 2 — the epoch is the number of the chair's
		// stored sittings at or before the row, not read off the seat id.
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "blue-r2", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED), Question: proto.String("rewrite the parser"), Reason: proto.String("the grammar moved under it")}),
	})
	exp := inquirySection(t, board, "Research areas")
	alt := inquirySection(t, board, "Alternatives considered")
	if strings.Contains(exp, "rewrite the parser") {
		t.Errorf("an inquiry ABANDONED in epoch 2 is not an expansion — its latest status decides:\n%s", exp)
	}
	if !strings.Contains(alt, "rewrite the parser") || !strings.Contains(alt, "the grammar moved under it") {
		t.Errorf("the abandoned inquiry must carry its current reason:\n%s", alt)
	}
	// The substance came from the CREATION event and the reason from the event recording its
	// fate. The PATH between them is the debate, not the subject: it renders in inquiries.md, never
	// in the report.
	if strings.Contains(alt, "e0 underway") || strings.Contains(alt, "history:") {
		t.Errorf("the status history is reconstructable from the ledger and is not report text:\n%s", alt)
	}
}

// RED'S RULING IS THE DEBATE, NOT THE SUBJECT. The report is research prose and tool
// markers, and both the ruling and the proposing seat's id are on the ledger. The ruling's opinion
// renders in judgments.md, and the ruling beside the inquiry's path in inquiries.md
// (view.TestAnInquiryRulingRendersBesideTheFateRecordedAgainstIt).
func TestInquiryRulingAndSeatStayOutOfTheReport(t *testing.T) {
	board := famOf(t, []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("survey the adjacent literature")}),
		// The LIVE vocabulary: red rules an inquiry through `motion inquiry rule`, whose motion_id
		// IS the inquiry's own id — the proposal is the filing, so there is no second identity.
		recordtest.Event(t, "red-chair", &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Opinion:  proto.String("a real question, not THIS run's"),
			Ruling:   &recordpb.MotionRule_Inquiry{Inquiry: recordpb.InquiryRuling_INQUIRY_RULING_OUT_OF_SCOPE},
		}),
		// Blue pursues it ANYWAY: the ruling binds no fate, and the fate is what blue records about it.
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Question: proto.String("survey the adjacent literature")}),
	})
	exp := inquirySection(t, board, "Research areas")
	if !strings.Contains(exp, "- **Question:** survey the adjacent literature\n") {
		t.Fatalf("the inquiry itself must still reach the report:\n%s", exp)
	}
	for _, leak := range []string{"out_of_scope", "a real question, not THIS run's", "against red's", "ruled", "(blue-r"} {
		if strings.Contains(exp, leak) {
			t.Errorf("the debate over an inquiry leaked into the report as %q:\n%s", leak, exp)
		}
	}
}

func TestDebateTranscriptFromEvents(t *testing.T) {
	evs := []*record.Event{
		mintT(t, "red-lens-logic", "G1"), // the gap the docket motion below names
		// The chair's register opens epoch 1; everything below sits in it.
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.Position{Text: proto.String("gap A stands")}),
		recordtest.Event(t, "blue-respond", &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("gap A repaired")}),
		// The payload keys are the ones the VERBS write: dispute→evidence, dispute-respond→
		// response+rationale, petition-rule→opinion. The prior fixture set basis/as (what the
		// buggy reader looked for), which is how A1–A3 hid — the test encoded the bug.
		// THE BENCH'S DISPOSITION IS A DOCKET MOTION'S RULING, and the transcript's "G1: remanded"
		// line is a JOIN across both events: the gap is on the filing, the word on the ruling.
		recordtest.Event(t, "red-chair", &recordpb.Motion{
			MotionId: proto.String("M2"),
			Basis:    proto.String("red cannot settle G1"),
			Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
		}),
		recordtest.Event(t, "judge", &recordpb.MotionRule{
			MotionId: proto.String("M2"),
			Opinion:  proto.String("needs a probe"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED),
				Principle:   proto.String("correctness"), Tension: proto.String("cost"),
				ReviewFlag: proto.String("false"),
				Settled:    proto.String("the claim as it stood may not be re-asserted"),
				Direction:  proto.String("the probe"),
			}},
		}),
		// A petition is a MOTION now — the retired `petition` event type has no arm in the schema,
		// so the fixture could only ever have described a state nothing writes.
		recordtest.Event(t, "blue-respond", &recordpb.Motion{
			MotionId: proto.String("M1"),
			Basis:    proto.String("the instruction would require asserting what I believe false"),
			Relief:   proto.String("strike the demand from the docket"),
			Filing:   &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_INTEGRITY)}},
		}),
		recordtest.Event(t, "judge", &recordpb.Halt{Opinion: proto.String("safety gate tripped")}),
		recordtest.Event(t, "judge", &recordpb.Certify{Statement: proto.String("re-examine the cost model")}),
	}
	d := debateT(t, evs)
	for _, want := range []string{
		"### Epoch 1", "### RED — NO VERDICT RECORDED THIS EPOCH\ngap A stands", "### BLUE CLOSING — G1\ngap A repaired",
		"G1: remanded",
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
	if empty := debate(record.NewFamily(nil, record.Merged{})); !strings.Contains(empty, "no debate on the record") {
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
	if p := fixProposal(&recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), FixBasis: proto.String("proposed")}); !strings.Contains(p, "**proposed**") || !strings.Contains(p, "nothing checked this demand") {
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
// only the basis distinguishes one from an honest epoch-0 rewrite.
func TestRemovalBasisReachesTheReader(t *testing.T) {
	v := withdrawnClaimsT(t, []*record.Event{
		recordtest.Event(t, "blue-r2", &recordpb.Register{}),
		recordtest.Event(t, "blue-r2", &recordpb.Retire{Claim: proto.String("the parser is linear"), Reason: proto.String("r"), RemovalBasis: proto.String(record.RemovalVerified)}),
	})
	if !strings.Contains(v, "**verified**") || !strings.Contains(v, "the record shows it leaving") {
		t.Errorf("a verified removal must say the record can show it:\n%s", v)
	}
	// The seat is named beside the row, so its SITTING rides with it — the rank of the seat's
	// stored sitting the row is in, not a number the seat stamped.
	if !strings.Contains(v, "(blue-r2 #1)") {
		t.Errorf("the retiring seat's sitting must render beside its name:\n%s", v)
	}
	a := withdrawnClaimsT(t, []*record.Event{recordtest.Event(t, "blue-r0", &recordpb.Retire{Claim: proto.String("the parser is linear"), Reason: proto.String("r"), RemovalBasis: proto.String(record.RemovalAsserted)})})
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
	filed := []*record.Event{recordtest.Event(t, "red-chair", &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("the demand would bury a hazard"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY)}}})}
	d := debateT(t, filed)
	if !strings.Contains(d, "1 petition(s) received no ruling") {
		t.Errorf("a petition with no ruling must be reported, not silently absent:\n%s", d)
	}
	// The ruling joins by MOTION ID, which is the whole point of the collapse: `petition-rule`
	// carried only the petitioner, so two filings by one seat in one round could not be told
	// apart. record.Motions pairs the answer to its ask exactly.
	// The RULING ARM is what makes it answered: a motion-rule whose oneof is unset carried no
	// ruling at all, which is the state the unanswered detector is looking for.
	answered := append(filed, recordtest.Event(t, "judge", &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Opinion:  proto.String("the hazard is graded, not buried"),
		Ruling:   &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED},
	}))
	if d := debateT(t, answered); strings.Contains(d, "received no ruling") {
		t.Errorf("an answered petition must not be reported as unanswered:\n%s", d)
	}
}

// A WITHDRAWN CLAIM IS PART OF WHAT THE DEBATE DECIDED. Substance leaves the report only through
// `retire`, which names the claim as it stood and why it went — and the reader saw none of it,
// making a report where a claim was argued and withdrawn identical to one where it was never made.
func TestWithdrawnClaimsReachTheReader(t *testing.T) {
	// The body carries the facts the assertions read. It was left EMPTY by the earlier conversion,
	// so the render had nothing to show and the test was asserting against a blank string.
	evs := []*record.Event{recordtest.Event(t, "blue-respond", &recordpb.Retire{
		Claim:        proto.String("the parser is O(n) in the input size"),
		Reason:       proto.String("refuted at the leaf"),
		SupersededBy: proto.String("the parser is O(n) except on backtracking inputs"),
	})}
	w := withdrawnClaimsT(t, evs)
	for _, want := range []string{"the parser is O(n) in the input size", "refuted at the leaf", "superseded by: the parser is O(n) except on backtracking inputs"} {
		if !strings.Contains(w, want) {
			t.Errorf("withdrawn claims missing %q:\n%s", want, w)
		}
	}
	if withdrawnClaims(record.NewFamily(nil, record.Merged{})) != "" {
		t.Error("a run that retired nothing omits the section rather than showing it empty")
	}
}

// EVERY LINE OF THE REPORT'S TEXT SHIPS, ONCE, AND NONE MOVES PAST ANOTHER. The text is parted in
// two at the section that closes it, with the title line taken out, and nothing is in neither
// part. universe-m20 lost a section, a proof and fifteen other anchors to a list of headings the
// composer left out, and universe-m21's Catechism shipped ahead of the sections it was written
// after; the composer keeps no list of sections to leave out or to move.
func TestEveryLineOfTheReportsTextShipsOnceWhereItWasWritten(t *testing.T) {
	lines := []string{
		"prose before the title.",
		"# A topic — research report",
		"**Verdict:** a line of the preamble",
		"```", "# a comment in a code block", "```",
		"",
		"## Technical foundations", "what is established.", "",
		// A section the write holds first, written late: a migrated record can hold one, and it
		// ships where it was written.
		"## The Catechism", "written after the argument.", "",
		"## TL;DR", "written later still.", "",
		// What the write refuses and a migrated record can still hold: kept whole, anchors and all.
		"## Alternatives considered", "an approach weighed<!--proof:P-00000001-->.", "",
		"## Risk Matrix", "a risk row.", "",
		"## Open questions", "the closing section.", "",
		"##  open questions (in full)", "a second, spelled another way.", "",
		"## Appendix: raw benchmarks", "a section after the closing one.", "",
	}
	got := part(strings.Join(lines, "\n"))
	if got.title != "# A topic — research report" {
		t.Errorf("title = %q", got.title)
	}
	title := slices.Index(lines, got.title)
	closing := slices.Index(lines, "## Open questions")
	body := strings.Join(append(slices.Clone(lines[:title]), lines[title+1:closing]...), "\n")
	if got.body != strings.TrimSpace(body) {
		t.Errorf("the body is not the text up to its closing section, as written:\n%s", got.body)
	}
	if got.close != strings.TrimSpace(strings.Join(lines[closing:], "\n")) {
		t.Errorf("the closing part is not the text from its closing section on, as written:\n%s", got.close)
	}
	// A report with no closing section is its body, and one that is only its closing section has none.
	if e := part("# t\n\n## TL;DR\nx\n\n## Appendix\ny\n"); e.close != "" || e.body != "## TL;DR\nx\n\n## Appendix\ny" {
		t.Errorf("a report with no closing section is not its body: %+v", e)
	}
	if e := part("# t\n\n## Open questions\ny\n"); e.body != "" || e.close != "## Open questions\ny" {
		t.Errorf("a report that is only its closing section: %+v", e)
	}
}

func TestUnmintedFindingsSurfaced(t *testing.T) {
	board := &boardT{
		GapOrder: []string{"G1"},
		Gaps: map[string]*record.Gap{
			"G1": {ID: "G1", Mint: &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), FoundBy: []string{"F-f0000001", "F-f0000002"}}},
		},
		Events: []*record.Event{
			recordtest.Event(t, "red-lens-logic", &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("minted — omit")}),
			recordtest.Event(t, "red-lens-dark-side", &recordpb.Finding{Id: proto.String("F-f0000002"), Text: proto.String("also minted — omit")}),
			recordtest.Event(t, "red-lens-logic", &recordpb.Finding{Id: proto.String("F-f0000003"), Location: proto.String("§H1"), Text: proto.String("un-minted red reasoning kept for the record")}),
		},
	}
	got := boardSection(board.famT(t))
	if !strings.Contains(got, "Lens findings credited by no gap (1)") {
		t.Errorf("exactly one un-minted finding should be surfaced:\n%s", got)
	}
	// THE SECTION STATES THE JOIN AND NOT A DELIBERATION. It used to say the merge "weighed but
	// did not mint" these, which nothing on the record supports — declining to mint writes no
	// event — and which is what made a silently dropped finding read as a considered decline
	// (#747: three dropped in one run, one of them a fabricated-quote allegation).
	if strings.Contains(got, "weighed but did not mint") {
		t.Error("the section asserts the chair weighed these findings; no event records deliberation, and a silent drop then reads as a decision")
	}
	if !strings.Contains(got, "is NOT recorded") {
		t.Errorf("the section does not say that whether these were considered is unrecorded:\n%s", got)
	}
	// THE HEADING NAMES THE FINDING BY ITS ID AND ITS LENS'S AREA, exactly: an id alone does not say
	// which lens filed it.
	if !strings.Contains(got, "\n### F-f0000003 (logic) — §H1\n") || !strings.Contains(got, "un-minted red reasoning kept") {
		t.Errorf("the un-minted finding's substance must be surfaced under `### F-f0000003 (logic) — §H1`:\n%s", got)
	}
	// A credited finding's text renders ONCE, under the gap that claimed it (surfaced by) —
	// and never again in the credited-by-no-gap section. The fixture used to leave GapOrder
	// empty so the gap itself vanished from the render; a fold-built board cannot produce that
	// shape, so the assertion now splits the section instead.
	_, unminted, _ := strings.Cut(got, "Lens findings credited by no gap")
	if strings.Contains(unminted, "minted — omit") || strings.Contains(unminted, "also minted") {
		t.Errorf("a finding credited by a gap's found_by must NOT be re-listed in the unminted section:\n%s", got)
	}
}

// A MINTED FINDING'S EVIDENCE MUST BE REACHABLE FROM THE GAP THAT CLAIMED IT.
//
// The provenance line used to read `surfaced by: F-f0000001, F-f0000002` and nothing in the report defined
// those labels — unmintedFindings renders a finding only when NO gap claims it, so the instant
// the chair acted on a finding its leaf-level evidence left the document and the citation
// dangled. A fuzz run where every finding was minted put red's words nowhere at all.
//
// It is the wrong half to drop: `problem` is the chair's RESTATEMENT, and a reader can only see
// a restatement drift from its evidence with both in front of them.
func TestAMintedFindingsEvidenceIsQuotedUnderItsGap(t *testing.T) {
	board := &boardT{
		GapOrder: []string{"G1"},
		Gaps: map[string]*record.Gap{
			// OPEN here; the CLOSED case is covered below. Scoping provenance to open gaps was
			// the first answer and it was wrong: unmintedFindings skips anything a gap claimed,
			// so a run where every finding was minted and every gap closed printed red's words
			// nowhere at all. The fuzz found one.
			"G1": {ID: "G1", Open: true, Mint: &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"),
				Problem: proto.String("the chair's restatement"),
				FoundBy: []string{"F-f0000001", "F-0000dead"},
			}},
		},
		Events: []*record.Event{
			recordtest.Event(t, "red-lens-logic", &recordpb.Finding{
				Id: proto.String("F-f0000001"), Location: proto.String("§H2"),
				Text: proto.String("what red actually observed at the leaf"),
			}),
		},
	}
	got := boardSection(board.famT(t))
	if !strings.Contains(got, "what red actually observed at the leaf") {
		t.Errorf("the minted finding's own words are absent — the gap cites F-f0000001 and nothing defines it:\n%s", got)
	}
	if !strings.Contains(got, "§H2") {
		t.Errorf("the finding's location is absent, so a reader cannot go and check it:\n%s", got)
	}
	// THE LINE NAMES THE FINDING BY ITS ID AND ITS LENS'S AREA, exactly.
	if want := "\nsurfaced by:\n- F-f0000001 (logic) (§H2): what red actually observed at the leaf\n"; !strings.Contains(got, want) {
		t.Errorf("the surfaced-by line is not %q:\n%s", want, got)
	}
	// AND AN UNRESOLVABLE CITATION IS ITSELF WORTH SEEING. A found_by naming no finding on the
	// record must not vanish into a shorter list — that reads exactly like a gap with less
	// provenance, which is the plausible-zero shape.
	if !strings.Contains(got, "\n- F-0000dead: (no finding with this id is on the record)") {
		t.Errorf("a found_by id with no finding behind it was dropped silently:\n%s", got)
	}
	// It stays out of the un-minted section either way: the gap is where it is rendered.
	if strings.Contains(got, "Lens findings not raised to a gap") {
		t.Errorf("a claimed finding must not ALSO be listed as un-raised:\n%s", got)
	}

	// AND A CLOSED GAP CARRIES IT TOO. The closure says how the gap was settled; it does not
	// restate what was observed, and an audit of a closure needs both.
	closedBoard := &boardT{
		GapOrder: []string{"G1"},
		Gaps: map[string]*record.Gap{
			"G1": {ID: "G1", Open: false, Mint: &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"),
				Problem: proto.String("the chair's restatement"),
				FoundBy: []string{"F-f0000001"},
			}},
		},
		Events: board.Events,
	}
	if got := boardSection(closedBoard.famT(t)); !strings.Contains(got, "what red actually observed at the leaf") {
		t.Errorf("a CLOSED gap dropped the evidence it was minted from, and nothing else renders it:\n%s", got)
	}
}

func TestLogSectionRendered(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Log{Text: proto.String("the --cx flag is missing from help"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		recordtest.Event(t, "blue-respond", &recordpb.Log{Text: proto.String("manifest cap fights methodology gaps"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
		// A SEAT THAT SAT AND FILED NOTHING is named in its own section, not among the problems.
		// It asserts nothing: the harness bracket is the whole evidence, so the seat spends no call.
		recordtest.Event(t, record.HarnessSeat, &recordpb.SittingOpen{AgentId: proto.String("agent-judge-1"), AgentType: proto.String("frank-exchange-of-views:lead-judge"), SeatId: proto.String("judge")}),
		mintT(t, "red-lens-logic", "G1"), // not a log entry
	}
	f := logSectionT(t, evs)
	// THE TYPE RENDERS BESIDE THE SEAT, which is the whole point of the channel: an operator
	// triages by reading the type, not by reading the prose to work out which kind it was.
	for _, want := range []string{
		"Log (what the run told the operator",
		"**red-chair** (defect): the --cx flag is missing",
		"**blue-respond** (defect): manifest cap fights",
		"Seats that sat and filed no log entry",
		"**judge**",
	} {
		if !strings.Contains(f, want) {
			t.Errorf("log section missing %q:\n%s", want, f)
		}
	}
	if empty := logSection(record.NewFamily(nil, record.Merged{})); empty != "" {
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

// BLUE IS NOT CHARGED FOR THE BENCH'S CLOSURES.
//
// correctnessManifest listed every closed gap with no manifest row as "a repair nobody audited,
// including its author". Bench-disposed gaps were in that set, and blue never repaired them —
// the report accused one party of skipping an audit another party never owed it.
func TestTheBenchsOwnClosuresAreNotBluesUnauditedRepairs(t *testing.T) {
	board := &boardT{
		GapOrder: []string{"G1", "G2"},
		Gaps: map[string]*record.Gap{
			// Blue repaired this one and filed no receipt: a real missing manifest row.
			"G1": {ID: "G1", HasClosed: true, Closure: &recordpb.Close{}},
			// The bench disposed of this one. Nobody repaired it, so no receipt is owed.
			"G2": {ID: "G2", HasClosed: true, BenchClosure: &recordpb.DocketRuling{}, ClosedByBench: true},
		},
		// Blue sat on both while they were open and edited only G1; the bench's disposition of
		// G2 is not blue's repair.
		Events: blueSatOn(t, []string{"G1", "G2"}, "G1"),
	}
	got := correctnessManifest(board.famT(t))
	if !strings.Contains(got, "G1") {
		t.Errorf("blue's own unaudited repair left the list:\n%s", got)
	}
	if strings.Contains(got, "G2") {
		t.Errorf("the report still charges blue for a gap the BENCH closed:\n%s", got)
	}
}

// THE ORDERING THAT MAKES ClosedByBench THE WRONG KEY, asserted because the cheap fix passes
// every other test in this file.
//
// Blue closes a gap and files no manifest row; the bench later rules on the same gap. The bench's
// word stands, so BenchClosedLast is true — while blue's `close` is still the closure and the
// receipt is still genuinely missing. Keyed on the flag, this row vanishes from the one section
// whose stated purpose is to report exactly this.
func TestABlueRepairTheBenchLaterRuledOnIsStillCharged(t *testing.T) {
	g := &record.Gap{ID: "G1", HasClosed: true, Closure: &recordpb.Close{}}
	g.BenchClosure = &recordpb.DocketRuling{}
	g.BenchClosedLast = true

	got := correctnessManifest((&boardT{GapOrder: []string{"G1"}, Gaps: map[string]*record.Gap{"G1": g}, Events: blueSatOn(t, []string{"G1"}, "G1")}).famT(t))
	if !strings.Contains(got, "G1") {
		t.Errorf("a missing receipt disappeared because the bench later ruled on the gap:\n%s", got)
	}
}

// blueSatOn is the record of blue-respond dispatched onto `engaged`, registering while they were
// still open, and editing to answer `repaired` — the sitting in which a manifest row is owed.
func blueSatOn(t *testing.T, engaged []string, repaired ...string) []*record.Event {
	evs := []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: engaged}),
		recordtest.Event(t, "blue-respond", &recordpb.Register{}),
	}
	for _, g := range repaired {
		evs = append(evs, recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String(g)}))
	}
	return evs
}

// THE REPORT NEVER SAYS NOT MEASURED (#1002, ruling 3). blueSatOn's sitting has neither a later
// register nor a stop, so a reader while the run runs cannot close it — and whether a sitting could
// be closed is the process's state, which the scorecard carries. The report's section renders what
// blue owed and receipted, and nothing about the reading.
func TestTheManifestSectionCarriesNoNotMeasuredLine(t *testing.T) {
	g := &record.Gap{ID: "G1"}
	withEdit := correctnessManifest((&boardT{GapOrder: []string{"G1"}, Gaps: map[string]*record.Gap{"G1": g}, Events: blueSatOn(t, []string{"G1"}, "G1")}).famT(t))
	if !strings.Contains(withEdit, "**1 repaired gap(s) carry no manifest row (G1).**") || strings.Contains(withEdit, "NOT MEASURED") {
		t.Errorf("an unreceipted repair in a sitting nothing closed renders its line and no not-measured line:\n%s", withEdit)
	}
	if got := correctnessManifest((&boardT{GapOrder: []string{"G1"}, Gaps: map[string]*record.Gap{"G1": g}, Events: blueSatOn(t, []string{"G1"})}).famT(t)); got != "" {
		t.Errorf("a sitting nothing closed that owed no row renders no section, got:\n%s", got)
	}
}

// A GAP BLUE REBUTTED IS NOT AN UNAUDITED REPAIR. The constitutions owe one row per REPAIRED gap;
// a gap blue argued against without an edit answering it was not repaired, whatever later closed it.
func TestARebuttedGapIsNotChargedAsAnUnauditedRepair(t *testing.T) {
	board := &boardT{
		GapOrder: []string{"G1"},
		Gaps:     map[string]*record.Gap{"G1": {ID: "G1", HasClosed: true, Closure: &recordpb.Close{}}},
		Events:   blueSatOn(t, []string{"G1"}),
	}
	if got := correctnessManifest(board.famT(t)); strings.Contains(got, "G1") {
		t.Errorf("blue is charged a receipt for a gap it rebutted without an edit:\n%s", got)
	}
}

// A GAP ITS LENS CLOSED BEFORE BLUE SAT IS NOT BLUE'S UNAUDITED REPAIR (#868). The lenses sit
// first; one closed G1 as repaired; blue sat, found it closed and filed no row, correctly. The
// report named G1 as "a repair nobody audited" anyway, charging blue a row it never owed — the
// same question the engine got wrong when it aborted the run.
func TestAGapClosedBeforeBlueSatIsNotChargedToBlue(t *testing.T) {
	board := &boardT{
		GapOrder: []string{"G1", "G2"},
		Gaps: map[string]*record.Gap{
			"G1": {ID: "G1", HasClosed: true, Closure: &recordpb.Close{}},
			"G2": {ID: "G2", HasClosed: true, Closure: &recordpb.Close{}},
		},
		Events: []*record.Event{
			mintT(t, "red-lens-logic", "G1"), mintT(t, "red-lens-voice", "G2"),
			recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1", "G2"}}),
			closeT(t, "red-lens-logic", "G1"),
			recordtest.Event(t, "blue-respond", &recordpb.Register{}),
			// An edit naming G1 after its lens closed it is not a repair of an open gap.
			recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1")}),
			recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G2")}),
			closeT(t, "red-lens-voice", "G2"),
		},
	}
	got := correctnessManifest(board.famT(t))
	if strings.Contains(got, "G1") {
		t.Errorf("blue is charged a row for a gap its lens closed before blue sat:\n%s", got)
	}
	if !strings.Contains(got, "G2") {
		t.Errorf("G2 was open when blue sat and closed after — its missing row left the list:\n%s", got)
	}
}

// renderedPrompt is a seat's dispatch prompt as the simulator suite renders and pins it.
func renderedPrompt(t *testing.T, seat string) string {
	t.Helper()
	path, err := repotree.Plugin("tests", "simulator", "testdata", "prompt-"+seat+".golden")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the rendered %s prompt: %v", seat, err)
	}
	return string(b)
}

// THE OPEN GAP'S ROW NAMES ITS CLASS WORD. The row passes the class through `grade`, which takes
// `any` and prints `—` for anything but a string: a bare GapClass would print `class —`.
func TestTheOpenGapRowNamesItsClassWord(t *testing.T) {
	board := &boardT{GapOrder: []string{"G1"}, Gaps: map[string]*record.Gap{
		"G1": {ID: "G1", Open: true, Mint: &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), Class: recordpb.GapClass_GAP_CLASS_FALSE_UNIVERSAL.Enum(),
			Problem: proto.String("a universal claim a case contradicts")}},
	}}
	got := boardSection(board.fam())
	if !strings.Contains(got, "class false_universal") || strings.Contains(got, "class —") {
		t.Errorf("the open gap's row must carry `class false_universal`:\n%s", got)
	}
}
