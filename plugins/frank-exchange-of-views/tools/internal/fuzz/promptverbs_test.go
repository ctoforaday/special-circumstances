package fuzz

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// EVERY VERB A PROMPT NAMES MUST EXIST.
//
// MEASURED, and it nearly shipped: the merge prompt told red to run `merge rule-avenue`
// while the verb is `avenue-rule`. The rename landed in four Go files and missed the one
// surface an agent actually reads. Nothing caught it — the fuzzer drives verbs DIRECTLY, so
// it exercised the real verb and the prompt's dead name went unexercised behind a green
// sweep. It surfaced only because a human asked whether the installed engine matched source.
//
// This is the prompt-side twin of the command-path gate in trajectory_test.go: that one asks
// "is every verb driven", this one asks "does every verb we TELL a seat to run exist". A
// seat handed a nonexistent verb does not fail loudly — the role boundary answers "verb
// outside this seat's role" and the seat, per the friction footer, logs friction and works
// around it. The capability is simply lost for the run.
//
// The tree is the authority, as everywhere else: cli.CommandPaths() walks the real cobra
// command set rather than a list someone maintains alongside it.
var promptVerb = regexp.MustCompile(`feov-record"?\s+(lens|merge|blue|bench)\s+([a-z][a-z-]+)`)

// agentFacingFiles are the surfaces a seat reads: the orchestrator's prompts, the role
// constitutions, and the skill that carries the protocol.
func agentFacingFiles(t *testing.T) []string {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	var out []string
	for _, glob := range []string{
		filepath.Join(root, "skills", "research-protocol", "scripts", "*.js"),
		filepath.Join(root, "skills", "research-protocol", "*.md"),
		filepath.Join(root, "agents", "*.md"),
		filepath.Join(root, "commands", "*.md"),
	} {
		m, err := filepath.Glob(glob)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, m...)
	}
	if len(out) == 0 {
		t.Fatal("found no agent-facing files at all — a broken glob passes this test silently forever")
	}
	return out
}

func TestEveryVerbNamedInAPromptExists(t *testing.T) {
	real := map[string]bool{}
	for _, p := range cli.CommandPaths() {
		real[p] = true
	}

	type site struct{ path, verb string }
	var bad []site
	for _, path := range agentFacingFiles(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range promptVerb.FindAllStringSubmatch(string(b), -1) {
			role, verb := m[1], m[2]
			// `show` is reached with --view and needs no per-view check here; the view names
			// have their own single-source gate (viewNames drives the help).
			if !real[role+" "+verb] {
				bad = append(bad, site{filepath.Base(path), role + " " + verb})
			}
		}
	}
	sort.Slice(bad, func(i, j int) bool { return bad[i].verb < bad[j].verb })

	seen := map[string]bool{}
	var msgs []string
	for _, s := range bad {
		k := s.path + ": " + s.verb
		if seen[k] {
			continue
		}
		seen[k] = true
		msgs = append(msgs, k)
	}
	if len(msgs) > 0 {
		t.Errorf("%d verb(s) named in an agent-facing file do NOT exist in the command tree — a seat told to run one loses that capability for the whole run and merely logs friction:\n  %s",
			len(msgs), strings.Join(msgs, "\n  "))
	}
}
