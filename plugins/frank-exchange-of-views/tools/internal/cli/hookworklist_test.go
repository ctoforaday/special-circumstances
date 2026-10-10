package cli

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// THE LIST A SEAT IS HANDED AND THE LIST IT ASKS FOR PLACE A GAP THE SAME WAY.
//
// The work list reaches a seat twice: delivered at dispatch by feov-sitting-write, through the
// SubagentStart hook, and fetched by the seat's own read. Both are record.WorkJSONBytes over one
// record, and they answered differently — nine of nine delivered lists read every quote gap
// `unrendered` while the same seat's read, a second later at the same head, read it `marked`. The
// renderer is REGISTERED by a linked package, so the answer was a property of the BINARY, and an
// in-process comparison cannot see it: this test binary links the renderer whichever way the
// writer is built. So the writer is driven as the process the hook spawns.
func TestTheDeliveredWorkListPlacesAGapWhereTheFetchedOneDoes(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary")
	}
	const base = "# Costs\n\nQ1 was hard. Costs rose sharply in the first quarter, then fell.\n"
	runDir := newRun(t)
	writeReport(t, runDir, base)
	id, err := mintQuote(t, runDir, "k1", "Costs rose sharply")
	if err != nil {
		t.Fatal(err)
	}

	states := func(t *testing.T, doc string) map[string]string {
		t.Helper()
		var w record.WorkJSON
		if err := json.Unmarshal([]byte(doc), &w); err != nil {
			t.Fatalf("not a work list (%v):\n%s", err, doc)
		}
		got := map[string]string{}
		for _, g := range w.Open {
			got[g.ID] = g.LocationState
		}
		return got
	}

	fetched, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatal(err)
	}
	want := states(t, fetched)
	if want[id] != record.LocationMarked {
		t.Fatalf("the fixture does not hold a marked gap, so equality below would compare nothing: %v", want)
	}

	cmd := exec.Command(testbuild.Binary(t, "feov-sitting-write"),
		"-run", runDir, "-phase", "open", "-agent-id", "agent_01",
		"-agent-type", "frank-exchange-of-views:"+lensSeat)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("feov-sitting-write: %v: %s", err, stderr.String())
	}
	at := strings.Index(string(out), "{")
	if at < 0 {
		t.Fatalf("the writer handed the seat no work list:\n%s", out)
	}
	got := states(t, string(out[at:]))
	if len(got) != len(want) {
		t.Fatalf("the delivered list holds %d gap(s) and the fetched one %d", len(got), len(want))
	}
	for gap, state := range want {
		if got[gap] != state {
			t.Errorf("gap %s is %q in the list the hook delivers and %q in the one the seat fetches — one record, one head",
				gap, got[gap], state)
		}
	}
}
