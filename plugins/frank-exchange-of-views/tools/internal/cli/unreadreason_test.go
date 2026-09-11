package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A VALUE ACCEPTED AND NEVER READ IS A SILENT DROP.
//
// `blue cite` registered --reason through seat.Prose and never read it: the seat's argument for a
// citation was accepted, the verb succeeded, and the record held nothing. `blue ingest` did the same.
// Neither was visible to any test, because a test asserts what a verb DOES with its input and
// nothing asserts that it used it.
//
// refuseAnUnreadReason is called by the harness after every successful invocation, so every test in
// this package that passes --reason to any verb now also asserts the verb read it. The prose channel
// is the one flag this can be asked of cheaply: every read goes through flags.Prose.Read. The other
// flags are read through several paths (flags.Value, GetBool, typed Var flags), and asking the same
// of them would mean instrumenting each — not done here.
func refuseAnUnreadReason(t *testing.T, root *cobra.Command, args []string) {
	t.Helper()
	c, _, err := root.Find(args)
	if err != nil || c == nil || !c.Flags().Changed(flags.Reason) {
		return
	}
	if p := flags.ProseOf(c); p != nil && !p.WasRead() {
		t.Errorf("`%s` accepted --%s and never read it: the seat's words are recorded nowhere, and the verb says it succeeded. "+
			"Record it on a field the body has, or stop registering the channel (seat.Prose)", c.CommandPath(), flags.Reason)
	}
}

// THE CITATION'S ARGUMENT REACHES THE RECORD, and is shown where red resolves a citation.
func TestCiteRecordsItsReasonAndTheEvidenceViewShowsIt(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": []byte("<html>the sky is blue</html>")}})

	why := "the source measures sky colour directly, which is the sentence's whole claim"
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", "The sky is blue and the grass is green.", "--url", "https://sky/1", "--title", "Sky Facts",
		"--reason", why); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if got := firstCiteEvent(t, runDir).GetText(); got != why {
		t.Fatalf("cite dropped its --reason: Cite.text = %q", got)
	}
	b, err := record.EvidenceJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatal(err)
	}
	if len(ev.Sources) != 1 || ev.Sources[0].Text != why {
		t.Errorf("the evidence view does not show the citation's argument:\n%s", b)
	}
	// NOT REPORT TEXT: the argument is about the citation, and the report is about the subject.
	if strings.Contains(readReport(t, runDir), "measures sky colour") {
		t.Error("the citation's argument leaked into the report")
	}
}

// No argument offered is recorded as absent, not as an empty argument.
func TestCiteWithoutAReasonRecordsNoText(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": []byte("<html>the sky is blue</html>")}})
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", "The sky is blue and the grass is green.", "--url", "https://sky/1", "--title", "Sky Facts"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if c := firstCiteEvent(t, runDir); c.Text != nil {
		t.Errorf("a cite with no --reason recorded text %q", c.GetText())
	}
}

// INGEST HAS NOWHERE TO PUT A REASON — BaseIngest carries the report and nothing else — so it does
// not accept one. A refusal at parse is the honest answer; accepting and dropping was the defect.
func TestIngestDoesNotAcceptAReason(t *testing.T) {
	runDir := newRun(t)
	// The FILE, not writeReport: that helper ingests, and a second ingest is refused for being
	// second — which would pass this test for the wrong reason.
	dir := filepath.Join(runDir, "blue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte("# Findings\n\nThe sky is blue.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize", "--reason", "freeze it")
	if err == nil {
		t.Fatalf("ingest accepted a --reason it has no field for:\n%s", out)
	}
	if !strings.Contains(err.Error()+out, "unknown flag: --reason") {
		t.Errorf("the refusal is not the parse refusal naming the flag: %v\n%s", err, out)
	}
	// Refused at parse, so nothing was frozen: the file is still there to ingest properly.
	if _, statErr := os.Stat(filepath.Join(dir, "report.md")); statErr != nil {
		t.Errorf("a refused ingest removed the report file: %v", statErr)
	}
}
