package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE LITERAL IS HELD AGAINST THE PLUGIN THAT PUBLISHES IT.
//
// prosthetic-conscience owns `SC_FINAL_MESSAGE_CONTRACTED` and declines its Stop nudge when it is
// set; this harness is a launcher and sets it. The constant cannot be imported — it lives in
// another module's internal tree — so the two copies are held together by reading the published
// documentation, which is the only interface a launcher in any other repository has either.
//
// A drift here is SILENT by construction: a seat launched with a misspelled variable gets exactly
// today's behaviour, an injected nudge and a lost envelope, and nothing reports it.
func TestTheContractedMarkerMatchesWhatProstheticConsciencePublishes(t *testing.T) {
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	doc := filepath.Join(root, "plugins", "prosthetic-conscience", "hooks", "README.md")
	b, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	// A WORD BOUNDARY, not a substring: a rename that EXTENDS the name would satisfy Contains
	// while nothing set the variable the hook reads.
	named := regexp.MustCompile(`\b` + regexp.QuoteMeta(finalMessageContracted) + `\b`)
	if !named.MatchString(string(b)) {
		t.Errorf("%s does not name %q — either the marker was renamed and this launcher was left "+
			"behind, or the contract stopped being published; a seat launched on the wrong "+
			"spelling loses its envelope and nothing says so",
			doc, finalMessageContracted)
	}
}

// THE ENVIRONMENT THE SEAT PROCESS ACTUALLY GETS, not just the constant. A test of the spelling is
// not a test that anything sets it.
func TestSeatEnvCarriesTheMarkerAndTheRun(t *testing.T) {
	got := map[string]string{}
	for _, kv := range seatEnv("/run/dir", "agent-1") {
		if k, v, ok := strings.Cut(kv, "="); ok {
			got[k] = v // later entries win, which is what exec does
		}
	}
	for k, want := range map[string]string{
		finalMessageContracted: "1",
		"FEOV_RUN":             "/run/dir",
		"FEOV_AGENT_ID":        "agent-1",
	} {
		if got[k] != want {
			t.Errorf("seat environment %s=%q, want %q", k, got[k], want)
		}
	}
}
