package cli

import (
	"encoding/json"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
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

// THE SHIPPED HOOK WRITES THE READ, AS THE PROCESSES THE HARNESS SPAWNS. The PreToolUse binary spawns
// the writer beside it for a seat's call that returns an envelope; it prints the input it was given
// with `record_read` set to the record's fold, over what the seat typed. An in-process test stubs that
// spawn, so this is what notices the writer's answer and the hook's reader parting.
func TestTheShippedHookWritesTheRecordsReadOverWhatTheSeatTyped(t *testing.T) {
	if testing.Short() {
		t.Skip("builds two binaries")
	}
	for _, v := range []string{seatenv.Var, seatenv.VarWrapper, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	const lead, agent = "session-lead", "agent_read_01"
	cwd := t.TempDir()
	runDir := newRun(t)
	writeReport(t, runDir, "# Costs\n\nCosts rose.\n")
	if _, err := runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "", lead); err != nil {
		t.Fatal(err)
	}
	seatType := "frank-exchange-of-views:red-chair"
	open := exec.Command(testbuild.Binary(t, "feov-sitting-write"), "-run", runDir, "-phase", "open", "-agent-id", agent, "-agent-type", seatType)
	if out, err := open.CombinedOutput(); err != nil {
		t.Fatalf("opening the sitting: %v\n%s", err, out)
	}
	filePetition(t, runDir)
	want, err := record.ReadRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(want.BenchOwed) != 1 || want.Halted {
		t.Fatalf("fixture: the read = %+v, want a petition owed and no halt", want)
	}

	payload, err := json.Marshal(map[string]any{"tool_name": "StructuredOutput", "cwd": cwd, "session_id": lead,
		"agent_id": agent, "agent_type": seatType,
		"tool_input": map[string]any{"notes": "mine", "record_read": map[string]any{"event": 999999, "bench_owed": []string{}, "halted": true}}})
	if err != nil {
		t.Fatal(err)
	}
	hook := exec.Command(testbuild.Binary(t, "feov-pretooluse"))
	hook.Stdin = strings.NewReader(string(payload))
	var stderr strings.Builder
	hook.Stderr = &stderr
	out, err := hook.Output()
	if err != nil {
		t.Fatalf("the hook: %v: %s", err, stderr.String())
	}
	var doc struct {
		H struct {
			Decision string `json:"permissionDecision"`
			Updated  struct {
				Notes string            `json:"notes"`
				Read  record.RecordRead `json:"record_read"`
			} `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("the hook printed no decision (%v):\n%s", err, out)
	}
	if doc.H.Decision != "allow" || doc.H.Updated.Notes != "mine" {
		t.Errorf("decision %q, notes %q; want the seat's call allowed with its own field intact:\n%s", doc.H.Decision, doc.H.Updated.Notes, out)
	}
	got := doc.H.Updated.Read
	if got.Halted || got.Event != want.Event || !slices.Equal(got.BenchOwed, want.BenchOwed) {
		t.Errorf("record_read = %+v, want the record's %+v, over the halt the seat typed", got, want)
	}
}
