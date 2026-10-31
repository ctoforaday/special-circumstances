package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// repeatedWords is what every verb's human line says when the call repeats an act already on the
// record and writes nothing (seat.Repeated), read from the one place it is worded.
var repeatedWords = strings.TrimPrefix(seat.Repeated(""), " — ")

// corroborateArgs is a supports_with_bridge corroboration of the seeded sentence, with extra flags.
func corroborateArgs(runDir string, extra ...string) []string {
	return append([]string{"corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/bridged", "--title", "A Bridged Source",
		"--quote", corroborated, "--reason", "the source summarises this"}, extra...)
}

// A CORROBORATION RETRY IS THE SAME ACT: held to the write's rules and to the reading it repeats,
// and answered with the reading on the record, never the request.
func TestACorroborationRetryIsHeldToTheReadingItRepeats(t *testing.T) {
	runDir := corroborateRun(t)
	first := []string{"--as", "supports_with_bridge", "--bridge", "summary", "--confidence", "high"}
	if _, err := run(t, corroborateArgs(runDir, first...)...); err != nil {
		t.Fatalf("the first corroboration was refused: %v", err)
	}

	// A retry naming another title is a second statement of the act, and is refused naming both.
	before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_VERIFY)
	if _, err := run(t, append(corroborateArgs(runDir, first...), "--title", "Another Title")...); err == nil ||
		feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), `--title "A Bridged Source"`) ||
		!strings.Contains(err.Error(), `--title "Another Title"`) {
		t.Errorf("a retry naming another title = %v; want a conflict naming both titles", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_VERIFY); n != before {
		t.Errorf("the refused retry recorded %d verify event(s)", n-before)
	}

	// An identical retry is idempotent, and echoes the STORED row.
	out, err := run(t, append(corroborateArgs(runDir, first...), "--json")...)
	if err != nil {
		t.Fatalf("an identical retry was refused: %v", err)
	}
	var env struct {
		Result struct {
			Idempotent bool   `json:"idempotent"`
			Source     string `json:"source"`
			Bridge     string `json:"bridge"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("read the retry's result: %v\n%s", err, out)
	}
	if !env.Result.Idempotent || env.Result.Source != "A Bridged Source" || env.Result.Bridge != "summary" {
		t.Errorf("the retry answered %+v; want the stored row (idempotent, title %q, bridge summary)", env.Result, "A Bridged Source")
	}
	// THE HUMAN FORM SAYS IT TOO. It printed the stored verdict as a fresh reading — "verified:
	// supports_with_bridge" — while the record took nothing, so a seat reading it could not tell.
	bare, err := run(t, corroborateArgs(runDir, first...)...)
	if err != nil {
		t.Fatalf("an identical retry was refused: %v", err)
	}
	if !strings.Contains(bare, repeatedWords) || strings.Contains(bare, "verified:") {
		t.Errorf("the retry's human line does not say it wrote nothing:\n%s", bare)
	}

	// A retry the write would refuse is refused: no bridge on supports_with_bridge.
	if _, err := run(t, corroborateArgs(runDir, "--as", "supports_with_bridge", "--confidence", "high")...); err == nil ||
		!strings.Contains(err.Error(), "requires --bridge") {
		t.Errorf("a retry naming no bridge was not refused for it: %v", err)
	}

	// A retry reading the source differently is refused, naming the reading on the record.
	for name, args := range map[string][]string{
		"outcome":    {"--as", "supports", "--confidence", "high"},
		"bridge":     {"--as", "supports_with_bridge", "--bridge", "near_restatement", "--confidence", "high"},
		"confidence": {"--as", "supports_with_bridge", "--bridge", "summary", "--confidence", "low"},
	} {
		_, err := run(t, corroborateArgs(runDir, args...)...)
		if err == nil || feov.CodeOf(err) != string(feov.Conflict) ||
			!strings.Contains(err.Error(), "--as supports_with_bridge --bridge summary --confidence high") {
			t.Errorf("%s: a retry reading the source differently answered %v; want a conflict naming the stored reading", name, err)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_VERIFY); n != 1 {
		t.Errorf("the record holds %d verify events, want the first alone", n)
	}
}

// A PAGE ON A CORROBORATION IS REFUSED: a corroboration holds no pages, and an admitted --page
// would be dropped with nothing recorded.
func TestVerifyRefusesAPageOnACorroboration(t *testing.T) {
	runDir := corroborateRun(t)
	out, err := run(t, corroborateArgs(runDir, "--as", "supports", "--confidence", "high", "--json")...)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result struct {
			Label string `json:"label"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.Label == "" {
		t.Fatalf("no label in the corroboration's result: %v\n%s", err, out)
	}
	// ANOTHER LENS reads it: the corroboration is this sitting's verify of its label for its own seat.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-logic"); err != nil {
		t.Fatal(err)
	}
	verify := []string{"verify", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", env.Result.Label,
		"--as", "supports", "--confidence", "high", "--reason", "read again"}
	if _, err := run(t, append(verify, "--page", "1")...); err == nil || feov.CodeOf(err) != string(feov.Validation) {
		t.Errorf("verify --page on a corroboration answered %v; want a validation refusal", err)
	}
	if _, err := run(t, verify...); err != nil {
		t.Errorf("the same verify without --page was refused: %v", err)
	}
}

// A CITE KEY NAMES ONE CITATION, as a compute key names one program.
func TestACiteKeyNamesOneCitation(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe measured latency is bounded.\n\nThe queue drains in order.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://x": []byte("src"), "https://y": []byte("other")}})
	base := func(url, quote string, extra ...string) []string {
		return append([]string{"cite", "--run", runDir, "--seat-id", citeSeat, "--key", "C1",
			"--quote", quote, "--url", url, "--title", "T"}, extra...)
	}
	if _, err := run(t, base("https://x", "The measured latency is bounded.")...); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, base("https://x", "The measured latency is bounded.")...); err != nil {
		t.Errorf("an identical keyed retry was refused: %v", err)
	}
	for name, c := range map[string]struct {
		args  []string
		names string
	}{
		"url":         {base("https://y", "The measured latency is bounded."), `--url "https://x"`},
		"source-span": {base("https://x", "The measured latency is bounded.", "--source-span", "src"), "no --source-span"},
		"quote":       {base("https://x", "The queue drains in order."), `--quote "The measured latency is bounded."`},
		"title":       {base("https://x", "The measured latency is bounded.", "--title", "U"), `--title "T"`},
		"reason":      {base("https://x", "The measured latency is bounded.", "--reason", "it says so"), "no --reason"},
	} {
		if _, err := run(t, c.args...); err == nil || feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), c.names) {
			t.Errorf("%s: a keyed cite differing from the stored one answered %v; want a conflict naming %s", name, err, c.names)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE); n != 1 {
		t.Errorf("the record holds %d cite events, want the first alone", n)
	}
}
