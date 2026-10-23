package cli

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A VERIFICATION'S CLAIM IS THE SENTENCE ITS ANCHOR MARKS, read by the tool. The lens names the
// citation; a verdict pinned to a sentence the citation does not mark cannot be expressed.

const (
	verifiedSentence = "The data shows the sky is blue and the grass is green."
	verifiedReport   = "# Findings\n\n" + verifiedSentence + "\n\nClosing paragraph stays.\n"
)

// citedClause is a run whose report cites a clause inside verifiedSentence, with the lens
// registered, and the citation's id.
func citedClause(t *testing.T) (string, string) {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, verifiedReport)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/v": []byte("<html>a source on the sky</html>")}})
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "the sky is blue", "--url", "https://sky/v", "--title", "Sky"); err != nil {
		t.Fatalf("cite: %v", err)
	}
	m := regexp.MustCompile(`<!--cite:(C-[0-9a-f]{8})-->`).FindStringSubmatch(readReport(t, runDir))
	if m == nil {
		t.Fatalf("no citation marker in the report:\n%s", readReport(t, runDir))
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	return runDir, m[1]
}

// printsClaim reports whether verify's own result line — not the duty list printed after it, which
// quotes every unanswered claim — names claim.
func printsClaim(out, claim string) bool {
	line, _, _ := strings.Cut(out, "\n")
	return strings.Contains(line, `"`+claim+`"`)
}

func verifyAs(t *testing.T, runDir, anchor, as string) (string, error) {
	t.Helper()
	return run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
		"--as", as, "--confidence", "high", "--reason", "read at the leaf")
}

func TestVerifyRecordsTheSentenceItsAnchorMarks(t *testing.T) {
	runDir, anchor := citedClause(t)
	before := readReport(t, runDir)
	out, err := verifyAs(t, runDir, anchor, "supports")
	if err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	// The whole sentence holding the marker, its anchors out — not the clause the cite quoted.
	if got := firstVerify(t, runDir).GetClaim(); got != verifiedSentence {
		t.Errorf("the verification's claim = %q, want the sentence its anchor marks, %q", got, verifiedSentence)
	}
	// The lens typed no claim, so the tool prints the one it read: it is what a finding quotes.
	if !printsClaim(out, verifiedSentence) {
		t.Errorf("verify did not print the claim it recorded:\n%s", out)
	}
	if after := readReport(t, runDir); after != before {
		t.Errorf("a verification changed the report:\n%s", after)
	}
}

func TestVerifyTakesNoQuote(t *testing.T) {
	runDir, anchor := citedClause(t)
	_, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor, "--quote", verifiedSentence,
		"--as", "supports", "--confidence", "high", "--reason", "r")
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --quote") {
		t.Fatalf("verify --quote was not refused as an unknown flag: %v", err)
	}
}

// unansweredContradictions is the run's unanswered contradictions as `show evidence` lists them.
func unansweredContradictions(t *testing.T, runDir string) []record.EvidenceContradictionJSON {
	t.Helper()
	b, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions
}

// unraisedClaims is the claim of each unanswered contradiction no finding has raised.
func unraisedClaims(t *testing.T, runDir string) []string {
	t.Helper()
	var claims []string
	for _, c := range unansweredContradictions(t, runDir) {
		if c.RaisedBy == "" {
			claims = append(claims, c.Claim)
		}
	}
	return claims
}

// lensFinding files a finding graded on every axis, anchored by args, and returns its id.
func lensFinding(t *testing.T, runDir, seatID, key string, args ...string) string {
	t.Helper()
	out, err := run(t, append([]string{"--json", "finding", "--run", runDir, "--seat-id", seatID, "--key", key,
		"--reason", "the source says otherwise", "--severity", "high", "--likelihood", "high", "--impact", "medium"}, args...)...)
	if err != nil {
		t.Fatalf("finding %q: %v\n%s", args, err, out)
	}
	var env struct {
		Result struct {
			FindingID string `json:"finding_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.FindingID == "" {
		t.Fatalf("finding --json printed no finding_id: %v\n%s", err, out)
	}
	return env.Result.FindingID
}

// A REFUTING VERIFICATION IS RAISED BY THE FINDING THAT NAMES IT — a finding quoting the cited
// clause, or the whole sentence its anchor marks, raises nothing — and it holds a PASS until the
// gap minted from that finding answers it: a finding alone answers nothing.
func TestARefutingVerificationIsRaisedByTheFindingNamingIt(t *testing.T) {
	runDir, anchor := citedClause(t)
	if out, err := verifyAs(t, runDir, anchor, "refutes"); err != nil || !printsClaim(out, verifiedSentence) {
		t.Fatalf("verify did not record and print the claim it read: %v\n%s", err, out)
	}
	if got := unraisedClaims(t, runDir); len(got) != 1 || got[0] != verifiedSentence {
		t.Fatalf("unanswered contradictions = %q, want the sentence the anchor marks", got)
	}
	for i, quote := range []string{"the sky is blue", verifiedSentence} {
		lensFinding(t, runDir, lensSeat, "q"+strconv.Itoa(i), "--quote", quote)
		if got := unraisedClaims(t, runDir); len(got) != 1 {
			t.Errorf("unraised contradictions = %q after a finding quoting %q, want the sentence still unraised: a quote raises nothing", got, quote)
		}
	}
	raising := lensFinding(t, runDir, lensSeat, "k1", "--about-kind", "contradiction", "--about", unansweredContradictions(t, runDir)[0].ID)
	if got := unraisedClaims(t, runDir); len(got) != 0 {
		t.Errorf("unraised contradictions = %q after the finding naming it, want none", got)
	}
	// RAISED IS NOT ANSWERED: the PASS is held until the gap minted from the raising finding.
	if got := len(unansweredContradictions(t, runDir)); got != 1 {
		t.Fatalf("%d contradiction(s) held after the finding alone, want 1 — a source the report misstates is material by its class", got)
	}
	if out, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--found-by", raising,
		"--class", "citation_figure_misattribution", "--quote", verifiedSentence, "--problem", "the source says the sky is green, and the reader relies on the colour",
		"--fix", "say what the source says", "--check-kind", "source", "--check", "the sentence matches the source",
		"--severity", "high", "--likelihood", "high", "--impact", "medium"); err != nil {
		t.Fatalf("mint --found-by %s: %v\n%s", raising, err, out)
	}
	if got := len(unansweredContradictions(t, runDir)); got != 0 {
		t.Errorf("%d contradiction(s) held after the gap minted from the raising finding, want 0", got)
	}
}

// A CITATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED, and the refusal records nothing.
func TestVerifyOfACitationWithNoSentenceIsRefused(t *testing.T) {
	verifies := func(t *testing.T, runDir string) int {
		t.Helper()
		b, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		return record.EvidenceJSONOf(b.Events, b.At).Counts.Verifications
	}
	refused := func(t *testing.T, runDir, anchor string, wants ...string) {
		t.Helper()
		before := verifies(t, runDir)
		_, err := verifyAs(t, runDir, anchor, "refutes")
		if err == nil {
			t.Fatal("a verification of a citation the report holds no sentence for was recorded")
		}
		for _, want := range wants {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the refusal does not say %q: %v", want, err)
			}
		}
		if got := verifies(t, runDir); got != before {
			t.Errorf("the refusal left %d verification(s) on the record", got-before)
		}
	}
	// cutDown edits the sentence down to its marker: the anchor stands bare.
	cutDown := func(t *testing.T) (string, string) {
		t.Helper()
		runDir, anchor := citedClause(t)
		tok := "<!--cite:" + anchor + "-->"
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
			"--quote", "The data shows the sky is blue"+tok+" and the grass is green.", "--new", tok,
			"--reason", "the sentence cannot stand"); err != nil {
			t.Fatalf("edit the sentence down to its marker: %v", err)
		}
		return runDir, anchor
	}

	t.Run("bare", func(t *testing.T) {
		runDir, anchor := cutDown(t)
		refused(t, runDir, anchor, anchor, "stands bare", "nothing to verify", "retire")
	})
	t.Run("retired", func(t *testing.T) {
		runDir, anchor := cutDown(t)
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", verifiedSentence,
			"--anchor", anchor, "--reason", "refuted"); err != nil {
			t.Fatalf("retire: %v", err)
		}
		refused(t, runDir, anchor, anchor, "is not in the report", "retired", "never placed", "nothing to verify")
		// Nor does the lens's work list offer the verification the tool refuses.
		out, err := run(t, "show", "--run", runDir, "--seat-id", lensSeat)
		if err != nil || strings.Contains(out, "citation "+anchor+" is on the record and nobody has verified it") {
			t.Errorf("the lens is still offered a verification of the retired citation %s: %v\n%s", anchor, err, out)
		}
	})
	t.Run("unrendered", func(t *testing.T) {
		runDir, anchor := citedClause(t)
		// A marker placed at text the report does not hold: replaying it fails, so nothing renders.
		if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: blueSeat},
			&recordpb.Anchor{Id: proto.String("C-0badc0de"), Location: proto.String("text the report never held")}); err != nil {
			t.Fatal(err)
		}
		refused(t, runDir, anchor, "does not render", anchor, "anomalies", "`log`")
	})
}

// THE CLAIM KEEPS THE SENTENCE'S MARKUP: verify prints it as the report holds it, unescaped, and
// the contradiction it opens names that claim. A finding quoting it, in either spelling, raises
// nothing; the finding naming the contradiction's id raises it.
func TestARefutedClaimKeepsItsMarkup(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\n**Iron is \"magnetic\".** Copper is not.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://iron/v": []byte("<html>a source on iron</html>")}})
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", `Iron is "magnetic"`, "--url", "https://iron/v", "--title", "Iron"); err != nil {
		t.Fatalf("cite: %v", err)
	}
	anchor := regexp.MustCompile(`<!--cite:(C-[0-9a-f]{8})-->`).FindStringSubmatch(readReport(t, runDir))[1]
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	const printed = `**Iron is "magnetic".**`
	if out, err := verifyAs(t, runDir, anchor, "absent"); err != nil || !printsClaim(out, printed) {
		t.Fatalf("verify did not print the claim %q: %v\n%s", printed, err, out)
	}
	if got := unraisedClaims(t, runDir); len(got) != 1 || got[0] != printed {
		t.Fatalf("unraised contradictions = %q, want the claim %q", got, printed)
	}
	for i, quote := range []string{`Iron is "magnetic".`, printed} {
		lensFinding(t, runDir, lensSeat, "q"+strconv.Itoa(i), "--quote", quote)
		if got := unraisedClaims(t, runDir); len(got) != 1 {
			t.Errorf("after a finding at %q, %d contradiction(s) unraised, want 1: a quote raises nothing", quote, len(got))
		}
	}
	lensFinding(t, runDir, lensSeat, "k1", "--about-kind", "contradiction", "--about", unansweredContradictions(t, runDir)[0].ID)
	if got := unraisedClaims(t, runDir); len(got) != 0 {
		t.Errorf("after the finding naming it, contradictions %q unraised, want none", got)
	}
}

// A READING NAMES THE CONTRADICTION IT LANDED IN, read back from the record: the one that opens it
// prints its id, the one that joins it prints the same id and the seat that opened it, and a reading
// that backs the claim prints none. The id it prints is the one a finding raises it by.
func TestAVerifyNamesTheContradictionItOpensOrJoins(t *testing.T) {
	runDir, anchor := citedClause(t)
	type printed struct {
		Result struct {
			Contradiction *struct {
				ID       string `json:"id"`
				Opened   bool   `json:"opened"`
				OpenedBy string `json:"opened_by"`
			} `json:"contradiction"`
		} `json:"result"`
	}
	reading := func(seatID, as string) string {
		t.Helper()
		human, err := run(t, "verify", "--run", runDir, "--seat-id", seatID, "--anchor", anchor,
			"--as", as, "--confidence", "high", "--reason", "read at the leaf")
		if err != nil {
			t.Fatalf("%s verify --as %s: %v\n%s", seatID, as, err, human)
		}
		return human
	}
	asJSON := func(seatID, as string) printed {
		t.Helper()
		out, err := run(t, "--json", "verify", "--run", runDir, "--seat-id", seatID, "--anchor", anchor,
			"--as", as, "--confidence", "high", "--reason", "read at the leaf")
		if err != nil {
			t.Fatalf("%s verify --json --as %s: %v\n%s", seatID, as, err, out)
		}
		var p printed
		if err := json.Unmarshal([]byte(out), &p); err != nil {
			t.Fatalf("verify --json: %v\n%s", err, out)
		}
		return p
	}

	opened := asJSON(lensSeat, "refutes")
	listed := unansweredContradictions(t, runDir)
	if len(listed) != 1 {
		t.Fatalf("%d contradiction(s) on the record after a refuting reading, want 1", len(listed))
	}
	if c := opened.Result.Contradiction; c == nil || c.ID != listed[0].ID || !c.Opened || c.OpenedBy != lensSeat {
		t.Fatalf("the opening reading printed contradiction %+v, want id %s opened by %s", c, listed[0].ID, lensSeat)
	}

	const joiner = "red-lens-logic"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", joiner); err != nil {
		t.Fatal(err)
	}
	human := reading(joiner, "absent")
	if want := "This reading joined contradiction " + listed[0].ID + " (opened by " + lensSeat + ")"; !strings.Contains(human, want) {
		t.Errorf("the joining reading did not print %q:\n%s", want, human)
	}
	if got := unansweredContradictions(t, runDir); len(got) != 1 || got[0].ID != listed[0].ID {
		t.Fatalf("contradictions after the joining reading = %+v, want the one it joined", got)
	}
	// Read after the join, the opening reading still names the contradiction it OPENED: the fold is
	// read through the reading, so a later reading that moved the contradiction on does not hide it.
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if _, ok := recordpb.BodyAs[*recordpb.Verify](e); ok {
			c, err := record.ContradictionOfReading(runtest.Open(t, runDir), e)
			if err != nil || c == nil || !c.Opened || strconv.FormatInt(c.ID, 10) != listed[0].ID {
				t.Errorf("the opening reading, read after the join, names %+v (%v), want contradiction %s opened", c, err, listed[0].ID)
			}
			break
		}
	}

	const backer = "red-lens-computation"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", backer); err != nil {
		t.Fatal(err)
	}
	if p := asJSON(backer, "supports"); p.Result.Contradiction != nil {
		t.Errorf("a supporting reading printed contradiction %+v, want none", p.Result.Contradiction)
	}

	// The printed id is the one a finding raises it by.
	lensFinding(t, runDir, lensSeat, "k1", "--about-kind", "contradiction", "--about", opened.Result.Contradiction.ID)
	if got := unraisedClaims(t, runDir); len(got) != 0 {
		t.Errorf("unraised contradictions = %q after the finding naming the printed id, want none", got)
	}
}

// The opening reading's human line names the id it opened, and a supporting one names no
// contradiction.
func TestAVerifyPrintsTheContradictionItOpens(t *testing.T) {
	runDir, anchor := citedClause(t)
	out, err := verifyAs(t, runDir, anchor, "refutes")
	if err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	line, _, _ := strings.Cut(out, "\n")
	if want := "This reading opened contradiction " + unansweredContradictions(t, runDir)[0].ID; !strings.Contains(line, want) {
		t.Errorf("the opening reading's line does not say %q:\n%s", want, out)
	}

	runDir, anchor = citedClause(t)
	out, err = verifyAs(t, runDir, anchor, "supports")
	if err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	if line, _, _ := strings.Cut(out, "\n"); strings.Contains(line, "contradiction") {
		t.Errorf("a supporting reading's line names a contradiction:\n%s", out)
	}
}
