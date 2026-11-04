package cli

import (
	"encoding/json"
	"regexp"
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

// verifyAs is lensSeat's reading of citation anchor with outcome as, given the gap flags where the
// outcome finds against the sentence (refutes, absent) — on a copy that is the work, such a reading
// mints its gap in the same act.
func verifyAs(t *testing.T, runDir, anchor, as string) (string, error) {
	t.Helper()
	args := []string{"verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
		"--as", as, "--confidence", "high", "--reason", "read at the leaf"}
	switch as {
	case "refutes":
		args = append(args, gapFlagArgs("claim_contradicts_source")...)
	case "absent":
		args = append(args, gapFlagArgs("citation_figure_misattribution")...)
	}
	return run(t, args...)
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

// A citation's reading takes no --quote: the sentence is the one its anchor marks.
func TestAVerifyOfACitationTakesNoQuote(t *testing.T) {
	runDir, anchor := citedClause(t)
	_, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor, "--quote", verifiedSentence,
		"--as", "supports", "--confidence", "high", "--reason", "r")
	if err == nil || !strings.Contains(err.Error(), "--quote") || !strings.Contains(err.Error(), "--anchor") {
		t.Fatalf("verify --anchor --quote was not refused naming both flags: %v", err)
	}
	if v := firstVerify(t, runDir); v != nil {
		t.Errorf("a refused verify recorded a verification: %v", v)
	}
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
	// THE BARE CITATION IS AN ARCHIVED ONE. `blue edit` refuses to leave a citation bare, so the
	// edit that did is appended to the record as a run from before that refusal holds it.
	t.Run("bare", func(t *testing.T) {
		runDir, anchor := citedClause(t)
		tok := "<!--cite:" + anchor + "-->"
		if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: blueSeat}, &recordpb.BlueEdit{
			Old:  proto.String("The data shows the sky is blue" + tok + " and the grass is green."),
			New:  proto.String(tok),
			Text: proto.String("the sentence cannot stand"),
		}); err != nil {
			t.Fatalf("the edit that left the citation bare: %v", err)
		}
		refused(t, runDir, anchor, anchor, "stands bare", "nothing to verify", "retire")
	})
	t.Run("retired", func(t *testing.T) {
		runDir, anchor := citedClause(t)
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat,
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

// THE CLAIM KEEPS THE SENTENCE'S MARKUP: verify prints it as the report holds it, unescaped, and the
// gap its reading mints stands on that claim.
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
	out, err := verifyAs(t, runDir, anchor, "absent")
	if err != nil || !printsClaim(out, printed) {
		t.Fatalf("verify did not print the claim %q: %v\n%s", printed, err, out)
	}
	if got := lastBody(t, runDir, &recordpb.Mint{}).GetLocation(); got != printed {
		t.Errorf("the reading's gap stands on %q, want the claim %q", got, printed)
	}
}

// gapFlagArgs are the gap flags a reading that mints is given: its class and the six other flags it
// requires.
func gapFlagArgs(class string) []string {
	return []string{"--class", class, "--problem", "the source says otherwise than the sentence",
		"--check-kind", "source", "--check", "the sentence states what the source says",
		"--severity", "high", "--likelihood", "high", "--impact", "medium"}
}
