package capture

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

// EVERY SECTION cost.md ACTUALLY EMITS HAS A ROUTE INTO THE RUN DOCUMENT.
//
// The route table is checked against a RENDERED cost.md rather than against itself. A table that
// agreed only with a list in the same file would discriminate nothing — the defect being repaired
// is precisely that a section existed which the fold did not know about, and the fold reported
// success anyway.
//
// So this renders the real report and fails on any heading costRoutes cannot place. Adding a
// section to cost.Report without giving it a home fails HERE, at the moment it is added, rather
// than silently going missing from every run document thereafter.
func TestEveryCostSectionIsRouted(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	// Seat turns, so the measured section is one of the ones actually exercised here.
	if _, err := record.AppendSeatTurns(run, "AG", []seatturn.Turn{
		{Index: 0, TSMillis: 1000, Model: "claude-opus-4-1", Output: 3, Thinking: true},
		{Index: 1, TSMillis: 9000, Model: "claude-opus-4-1", Output: 40, Tool: true},
	}); err != nil {
		t.Fatal(err)
	}

	transcripts := t.TempDir()
	body := `{"agentId":"AG","timestamp":"2026-09-03T03:00:00.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"thinking"}],"usage":{"input_tokens":100,"output_tokens":3,"cache_read_input_tokens":900,"cache_creation_input_tokens":0}}}` + "\n"
	if err := os.WriteFile(filepath.Join(transcripts, "agent-AG.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var rendered bytes.Buffer
	if err := cost.Report(transcripts, run, &rendered); err != nil {
		t.Fatalf("rendering cost.md: %v", err)
	}

	sections := splitSections(rendered.String())
	if len(sections) == 0 {
		// AN EMPTY RENDER PASSES EVERY ASSERTION BELOW. A gate that measured nothing must say so.
		t.Fatal("cost.Report emitted no sections — this test measured nothing")
	}
	var seen, unrouted []string
	for _, s := range sections {
		seen = append(seen, s.Heading)
		if _, ok := costRoutes[s.Heading]; !ok {
			unrouted = append(unrouted, s.Heading)
		}
	}
	sort.Strings(seen)
	t.Logf("cost.md sections exercised here: %v", seen)
	for _, h := range unrouted {
		t.Errorf("cost.md emits a %q section that costRoutes has no home for — it would be written, "+
			"archived, and never reach the assembled run document. Add it to costRoutes.", h)
	}
}

// AND THE ROUTE TABLE CARRIES NO NAME THAT IS NEVER EMITTED.
//
// A stale route is the other half: it checks nothing while reading as coverage, which is how the
// previous fold's own test came to assert that Notes must NOT be carried — encoding the defect as
// a requirement. This cannot assert against one fixture (Tier check and Board telemetry need run
// config and minted gaps), so it reports rather than fails, and names what went unexercised.
func TestRouteTableHasNoObviouslyDeadEntries(t *testing.T) {
	if len(costRoutes) == 0 {
		t.Fatal("no routes at all")
	}
	dests := map[string]bool{}
	for _, dest := range costRoutes {
		dests[dest] = true
	}
	for _, want := range []string{"Cost", "Tier check", "Board telemetry"} {
		if !dests[want] {
			t.Errorf("no cost.md section routes to %q", want)
		}
	}
}
