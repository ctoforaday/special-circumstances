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

// A REFUTING VERIFICATION IS RAISED BY A FINDING THAT QUOTES THE SENTENCE ITS ANCHOR MARKS — a
// finding at the cited clause alone does not raise it, one at the whole sentence does — and it
// holds a PASS until the gap minted from that finding answers it: a finding alone answers nothing.
func TestARefutingVerificationIsRaisedByAFindingAtItsSentence(t *testing.T) {
	runDir, anchor := citedClause(t)
	contradictions := func() []string {
		t.Helper()
		b, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		var claims []string
		for _, c := range record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions {
			if c.RaisedBy == "" {
				claims = append(claims, c.Claim)
			}
		}
		return claims
	}
	if out, err := verifyAs(t, runDir, anchor, "refutes"); err != nil || !printsClaim(out, verifiedSentence) {
		t.Fatalf("verify did not record and print the claim it read: %v\n%s", err, out)
	}
	if got := contradictions(); len(got) != 1 || got[0] != verifiedSentence {
		t.Fatalf("unanswered contradictions = %q, want the sentence the anchor marks", got)
	}
	finding := func(key, quote string) string {
		t.Helper()
		out, err := run(t, "--json", "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key,
			"--quote", quote, "--reason", "the source says the sky is green",
			"--severity", "high", "--likelihood", "high", "--impact", "medium")
		if err != nil {
			t.Fatal(err)
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
	finding("k1", "the sky is blue")
	if got := contradictions(); len(got) != 1 {
		t.Errorf("unraised contradictions = %q after a finding at the clause alone, want the sentence still unraised", got)
	}
	raising := finding("k2", verifiedSentence)
	if got := contradictions(); len(got) != 0 {
		t.Errorf("unraised contradictions = %q after a finding at the sentence, want none", got)
	}
	// RAISED IS NOT ANSWERED: the PASS is held until the gap minted from the raising finding.
	held := func() int {
		t.Helper()
		b, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		return len(record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions)
	}
	if got := held(); got != 1 {
		t.Fatalf("%d contradiction(s) held after the finding alone, want 1 — a source the report misstates is material by its class", got)
	}
	if out, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--found-by", raising,
		"--class", "self-attestation", "--quote", verifiedSentence, "--problem", "the source says the sky is green, and the reader relies on the colour",
		"--fix", "say what the source says", "--check-kind", "source", "--check", "the sentence matches the source",
		"--severity", "high", "--likelihood", "high", "--impact", "medium"); err != nil {
		t.Fatalf("mint --found-by %s: %v\n%s", raising, err, out)
	}
	if got := held(); got != 0 {
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

// THE CLAIM KEEPS THE SENTENCE'S MARKUP, so the finding that answers it quotes it as verify printed
// it, unescaped: the same words without their emphasis place in the report and answer nothing.
func TestARefutedClaimIsAnsweredAsPrinted(t *testing.T) {
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
	held := func() int {
		t.Helper()
		b, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, c := range record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions {
			if c.RaisedBy == "" {
				n++
			}
		}
		return n
	}
	for _, f := range []struct {
		quote string
		held  int
	}{{`Iron is "magnetic".`, 1}, {printed, 0}} {
		if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "k"+strconv.Itoa(f.held), "--quote", f.quote,
			"--reason", "the source says iron is not magnetic", "--severity", "high", "--likelihood", "high", "--impact", "medium"); err != nil {
			t.Fatalf("finding at %q: %v", f.quote, err)
		}
		if got := held(); got != f.held {
			t.Errorf("after a finding at %q, %d contradiction(s) unraised, want %d", f.quote, got, f.held)
		}
	}
}
