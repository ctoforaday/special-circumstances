package report

import (
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
	bj := record.BoardJSON{Open: []record.GapJSON{
		{ID: "R1-1", Problem: "overclaims capture", Likelihood: "high", Impact: "medium", RequiredFix: "grep the sites"},
		{ID: "R1-2", Problem: "cost model rough", Likelihood: "low", Impact: "low", Class: "risk_accepted", RequiredFix: "accepted: low blast radius"},
	}}
	m := riskMatrix(bj)
	if !strings.Contains(m, "| overclaims capture | high | medium | — | grep the sites |") {
		t.Errorf("open gap row wrong:\n%s", m)
	}
	if !strings.Contains(m, "risk_accepted — accepted: low blast radius") {
		t.Errorf("risk_accepted disposition not marked:\n%s", m)
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
		{Round: 1, Type: "dispute", SeatID: "blue-r1", Payload: record.NewPayload().Set("gap_id", "R1-1").Set("dimension", "impact").Set("proposed", "low").Set("basis", "trivial harm")},
		{Round: 1, Type: "dispute-respond", SeatID: "red-merge-r1", Payload: record.NewPayload().Set("as", "rejected").Set("basis", "harm compounds")},
		{Round: 1, Type: "opinion", SeatID: "judge-r1", Payload: record.NewPayload().Set("gap_id", "R1-1").Set("disposition", "carried").Set("principle", "correctness").Set("tension", "cost").Set("review_flag", "false").Set("rationale", "needs a probe")},
		{Round: 1, Type: "petition-rule", SeatID: "judge-petition", Payload: record.NewPayload().Set("petitioner", "blue").Set("ruling", "granted").Set("rationale", "relief warranted")},
		{Round: 0, Type: "halt", SeatID: "judge-terminal", Payload: record.NewPayload().Set("opinion", "safety gate tripped")},
		{Round: 0, Type: "certify", SeatID: "judge-terminal", Payload: record.NewPayload().Set("statement", "re-examine the cost model")},
	}
	d := debate(evs)
	for _, want := range []string{
		"### Round 1", "### RED\ngap A stands", "### BLUE\ngap A repaired",
		"disputes R1-1/impact → low: trivial harm", "answered (rejected): harm compounds",
		"R1-1: carried", "petition blue: granted",
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

func TestCellEscapesTableBreakers(t *testing.T) {
	if got := cell("a | b\nc"); strings.ContainsAny(got, "\n") || strings.Contains(got, " | ") {
		t.Errorf("cell did not neutralise a pipe/newline: %q", got)
	}
	if cell("   ") != "—" {
		t.Error("a blank cell should be a dash")
	}
}
