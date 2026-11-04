package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// repeatedWords is what every verb's human line says when the call repeats an act already on the
// record and writes nothing (seat.Repeated), read from the one place it is worded.
var repeatedWords = strings.TrimPrefix(seat.Repeated(""), " — ")

// stating is args with flag's value replaced, for a retry that states one field otherwise.
func stating(args []string, flag, value string) []string {
	out := slices.Clone(args)
	out[slices.Index(out, flag)+1] = value
	return out
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
		"title":       {stating(base("https://x", "The measured latency is bounded."), "--title", "U"), `--title "T"`},
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
