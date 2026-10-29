package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// ONE GAP PER CLASS AND SUBJECT (ruled by gblock 2026-10-28). A mint is refused where an open gap of
// the same class stands on the same sentence or the same inquiry; the refusal names that gap, its
// minter and its problem, and points the lens at a finding about it.

const (
	oneGapSentence = "Costs rose sharply across the first quarter."
	oneGapOther    = "Volume recovered later in the year."
	oneGapReport   = "# Costs\n\nThe year opened badly. " + oneGapSentence + " " + oneGapOther + "\n"
)

// oneGapRun is oneGapReport frozen, with blue, the chair and the lens seated.
func oneGapRun(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, oneGapReport)
	registerBlue(t, runDir)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	return runDir
}

// mintOn mints as the lens with class and the subject args name, its problem problem.
func mintOn(t *testing.T, runDir, class, problem string, subject ...string) (string, string, error) {
	t.Helper()
	out, err := runMint(t, append([]string{"--run", runDir, "--seat-id", lensSeat, "--class", class,
		"--problem", problem, "--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"}, subject...)...)
	return gapID(out), out, err
}

// refusedNaming fails the test unless err is the one-gap refusal naming held, its minter and problem.
func refusedNaming(t *testing.T, err error, out, held, problem string) {
	t.Helper()
	if err == nil {
		t.Fatalf("a second gap of one class on one subject was admitted: %s", out)
	}
	for _, want := range []string{held, lensSeat, problem, "--about-kind gap --about " + held, "one gap"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q:\n%v", want, err)
		}
	}
}

func TestASecondGapOfOneClassOnOneSentenceIsRefusedNamingTheFirst(t *testing.T) {
	runDir := oneGapRun(t)
	first, out, err := mintOn(t, runDir, "spec_underspecification", "the first problem", "--quote", oneGapSentence)
	if err != nil {
		t.Fatalf("first mint: %v\n%s", err, out)
	}
	// The same sentence, whatever span of it the quote names.
	for _, quote := range []string{oneGapSentence, "Costs rose sharply"} {
		_, out, err := mintOn(t, runDir, "spec_underspecification", "a second problem", "--quote", quote)
		refusedNaming(t, err, out, first, "the first problem")
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 1 {
		t.Errorf("the record holds %d mints, want the first alone", n)
	}
	// The act the refusal points to is admitted.
	lensFinding(t, runDir, lensSeat, "about-first", "--about-kind", "gap", "--about", first)
}

func TestAnotherClassOnTheSameSentenceIsAdmitted(t *testing.T) {
	runDir := oneGapRun(t)
	if _, out, err := mintOn(t, runDir, "spec_underspecification", "the first problem", "--quote", oneGapSentence); err != nil {
		t.Fatalf("first mint: %v\n%s", err, out)
	}
	if _, out, err := mintOn(t, runDir, "invalid_inference", "another kind of defect", "--quote", oneGapSentence); err != nil {
		t.Fatalf("a gap of another class on the same sentence was refused: %v\n%s", err, out)
	}
}

func TestTheSameClassOnAnotherSentenceIsAdmitted(t *testing.T) {
	runDir := oneGapRun(t)
	if _, out, err := mintOn(t, runDir, "spec_underspecification", "the first problem", "--quote", oneGapSentence); err != nil {
		t.Fatalf("first mint: %v\n%s", err, out)
	}
	if _, out, err := mintOn(t, runDir, "spec_underspecification", "the first problem", "--quote", oneGapOther); err != nil {
		t.Fatalf("a gap of the same class on another sentence was refused: %v\n%s", err, out)
	}
}

// proposeInquiry records an inquiry as blue and returns its id.
func proposeInquiry(t *testing.T, runDir, question string) string {
	t.Helper()
	out, err := run(t, "--json", "inquiry", "propose", "--run", runDir, "--seat-id", blueSeat, "--question", question,
		"--hypothesis", "the record would show it")
	if err != nil {
		t.Fatalf("inquiry propose: %v\n%s", err, out)
	}
	var env struct {
		Result struct {
			InquiryID string `json:"inquiry_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.InquiryID == "" {
		t.Fatalf("inquiry propose --json printed no inquiry_id: %v\n%s", err, out)
	}
	return env.Result.InquiryID
}

func TestTheInquiryIsASubject(t *testing.T) {
	runDir := oneGapRun(t)
	q1 := proposeInquiry(t, runDir, "whether costs rose")
	q2 := proposeInquiry(t, runDir, "whether volume recovered")
	first, out, err := mintOn(t, runDir, "cross_section_contradiction", "the row contradicts the summary", "--about-kind", "inquiry", "--about", q1)
	if err != nil {
		t.Fatalf("first mint: %v\n%s", err, out)
	}
	_, out, err = mintOn(t, runDir, "cross_section_contradiction", "the row and the body disagree", "--about-kind", "inquiry", "--about", q1)
	refusedNaming(t, err, out, first, "the row contradicts the summary")
	if _, out, err := mintOn(t, runDir, "derivation_status_overclaim", "the fate overclaims", "--about-kind", "inquiry", "--about", q1); err != nil {
		t.Errorf("a gap of another class on the same inquiry was refused: %v\n%s", err, out)
	}
	if _, out, err := mintOn(t, runDir, "cross_section_contradiction", "the row contradicts the summary", "--about-kind", "inquiry", "--about", q2); err != nil {
		t.Errorf("a gap of the same class on another inquiry was refused: %v\n%s", err, out)
	}
}

func TestAClosedGapDoesNotBlock(t *testing.T) {
	runDir := oneGapRun(t)
	q := proposeInquiry(t, runDir, "whether costs rose")
	for _, subject := range [][]string{{"--quote", oneGapSentence}, {"--about-kind", "inquiry", "--about", q}} {
		first, out, err := mintOn(t, runDir, "spec_underspecification", "the first problem", subject...)
		if err != nil {
			t.Fatalf("first mint on %q: %v\n%s", subject, err, out)
		}
		if out, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", first, "--as", "repaired",
			"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x", "--reason", "the check passes"); err != nil {
			t.Fatalf("close: %v\n%s", err, out)
		}
		if _, out, err := mintOn(t, runDir, "spec_underspecification", "the defect recurred", subject...); err != nil {
			t.Errorf("a recurrence on %q after its gap closed was refused: %v\n%s", subject, err, out)
		}
	}
}

func TestASuccessorOfTheOpenGapIsAdmitted(t *testing.T) {
	runDir := oneGapRun(t)
	first, out, err := mintOn(t, runDir, "spec_underspecification", "the first problem", "--quote", oneGapSentence)
	if err != nil {
		t.Fatalf("first mint: %v\n%s", err, out)
	}
	if _, out, err := mintOn(t, runDir, "spec_underspecification", "the problem restated", "--quote", oneGapSentence, "--supersedes", first); err != nil {
		t.Fatalf("a successor naming the open gap in --supersedes was refused: %v\n%s", err, out)
	}
}

// The gap minted from the finding that raised a contradiction is the contradiction's one answer, so
// an open gap of its class on its sentence does not refuse it.
func TestAContradictionsAnsweringGapIsAdmitted(t *testing.T) {
	runDir, anchor := citedClause(t)
	registerChairOnce(t, runDir)
	if out, err := verifyAs(t, runDir, anchor, "refutes"); err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	const class = "citation_figure_misattribution"
	held, out, err := mintOn(t, runDir, class, "the sentence misstates its source", "--quote", verifiedSentence)
	if err != nil {
		t.Fatalf("the open gap: %v\n%s", err, out)
	}
	if _, out, err := mintOn(t, runDir, class, "the sentence misstates its source again", "--quote", verifiedSentence); err == nil {
		t.Fatalf("a second gap of the class on the sentence, answering nothing, was admitted: %s", out)
	}
	raising := lensFinding(t, runDir, lensSeat, "k1", "--about-kind", "contradiction", "--about", unansweredContradictions(t, runDir)[0].ID)
	answer, out, err := mintOn(t, runDir, class, "the source says the sky is green", "--quote", verifiedSentence, "--found-by", raising)
	if err != nil {
		t.Fatalf("the gap answering the contradiction was refused beside %s: %v\n%s", held, err, out)
	}
	if answer == "" || len(unansweredContradictions(t, runDir)) != 0 {
		t.Errorf("the answering mint %q left the contradiction unanswered", answer)
	}
}

// TWO CONCURRENT MINTS ARE ONE GAP AND ONE REFUSAL. Another lens's mint of the same class on the same
// sentence commits on its own connection at the instant this lens's write opens its transaction
// (runtest.Overtaken); the check runs under the write lock, so it reads the other gap.
func TestTwoConcurrentMintsAreOneGapAndOneRefusal(t *testing.T) {
	runDir := oneGapRun(t)
	const other, otherID = "red-lens-logic", "G-0badf00d"
	mint := &recordpb.Mint{GapId: proto.String(otherID), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Location: proto.String(oneGapSentence), Problem: proto.String("the concurrent problem"), RequiredFix: proto.String("f"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Supersedes: []string{}}
	acts := []*recordpb.Event{
		recordtest.At(t, other, other+":mint:1", mint),
		recordtest.At(t, other, other+":anchor:"+otherID, &recordpb.Anchor{Id: proto.String(otherID), Location: proto.String(oneGapSentence)}),
	}
	var out string
	err := runtest.Overtaken(t, otherSeat(t, runDir), acts, func() error {
		var err error
		_, out, err = mintOn(t, runDir, "spec_underspecification", "this lens's problem", "--quote", oneGapSentence)
		return err
	})
	if err == nil {
		t.Fatalf("two concurrent mints of one class on one sentence both landed: %s", out)
	}
	for _, want := range []string{otherID, other, "the concurrent problem"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q:\n%v", want, err)
		}
	}
	b, berr := record.BoardJSONOfRun(runtest.Open(t, runDir))
	if berr != nil {
		t.Fatal(berr)
	}
	if len(b.Open) != 1 || b.Open[0].ID != otherID {
		t.Errorf("the board holds %d open gap(s), want the concurrent one alone", len(b.Open))
	}
}

// THE GAP A FAILED CHECK OBLIGES IS NOT A SECOND GAP. A lens whose reading of a citation read nothing
// owes a gap about it, outside its budget; another lens's open gap of the same class on the cited
// sentence does not refuse it, since the gap is about the anchor, and the finding the refusal would
// point to does not answer the check. A quote gap of that class on that sentence is refused as ever.
func TestTheGapAFailedCheckObligesIsNotRefusedAsASecondGap(t *testing.T) {
	runDir, label := citedClause(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-logic"); err != nil {
		t.Fatal(err)
	}
	mint := func(seat string, subject ...string) (string, error) {
		return run(t, append([]string{"mint", "--run", runDir, "--seat-id", seat,
			"--class", "spec_underspecification", "--problem", "the source cannot be read",
			"--fix", "retire the citation or cite a readable source", "--check-kind", "document",
			"--check", "a lens reads the cited source", "--severity", "medium", "--likelihood", "medium", "--impact", "medium"}, subject...)...)
	}
	if out, err := mint("red-lens-logic", "--quote", "the sky is blue"); err != nil {
		t.Fatalf("the other lens's gap on the sentence: %v\n%s", err, out)
	}
	if _, err := verifyAs(t, runDir, label, "unreachable"); err != nil {
		t.Fatalf("verify --as unreachable: %v", err)
	}
	if out, err := mint(lensSeat, "--about-kind", "anchor", "--about", label); err != nil {
		t.Fatalf("the gap the failed check obliges was refused: %v\n%s", err, out)
	}
	if out, err := mint(lensSeat, "--quote", "the sky is blue"); err == nil || !strings.Contains(err.Error(), "Two defects of one class on one subject are one gap") {
		t.Fatalf("a quote gap of the held class on the cited sentence was not refused as a second gap: %v\n%s", err, out)
	}
}
