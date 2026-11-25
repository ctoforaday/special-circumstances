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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// AT DISPATCH A SEAT IS HANDED THE HEAD OF ITS QUEUE, OR ONE LINE WHEN THAT HEAD IS REPORT TEXT, AND
// NEVER REPORT TEXT (constitution-gaps §III.3.2, ruling 2026-11-14). The SubagentStart hook spawns
// feov-sitting-write's opening end; it hands a seat its configuration names the head of its queue, as
// the record queues it — not the whole list, which `show work` still prints — and where that head is
// report text, the one line saying it arrives one item per tool call. A configuration that
// seats several seats is handed nothing. The writer is driven as the process the hook spawns, so what
// is asserted is what the shipped binary hands.
func TestTheOpeningEndHandsTheHeadOfTheQueueAndNeverReportText(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary")
	}
	const sentence = "Costs rose sharply in the first quarter, then fell."
	runDir := newRun(t)
	writeReport(t, runDir, "# Costs\n\nQ1 was hard. "+sentence+"\n")
	gap, err := mintQuote(t, runDir, "k1", "Costs rose sharply")
	if err != nil {
		t.Fatal(err)
	}
	// The chair's item is its ruling on blue's grade motion: a chair's queue holds no report text here.
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond", "--id", gap,
		"--dimension", "severity", "--proposed", "low", "--reason", "the consequence is bounded"); err != nil {
		t.Fatal(err)
	}
	open := func(agent, agentType string) string {
		t.Helper()
		cmd := exec.Command(testbuild.Binary(t, "feov-sitting-write"),
			"-run", runDir, "-phase", "open", "-agent-id", agent, "-agent-type", agentType)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("feov-sitting-write -phase open as %s: %v: %s", agentType, err, stderr.String())
		}
		return string(out)
	}
	head := func(seatID string) record.QueueEntry {
		t.Helper()
		w, err := record.WorkOfSeat(runtest.Open(t, runDir), seatclass.RoleOfSeat(seatID), seatID)
		if err != nil {
			t.Fatal(err)
		}
		q := record.QueueOf(w)
		if len(q) == 0 {
			t.Fatalf("%s's queue is empty, so the opening end has no head to hand", seatID)
		}
		return q[0]
	}

	// THE CHAIR'S HEAD IS NO REPORT TEXT: it is handed that item, under its standing, and not the list.
	if h := head("red-chair"); h.Kind == record.ItemReportText {
		t.Fatalf("the fixture's chair is owed report text, so its arm measures nothing: %+v", h)
	}
	chair := open("agent_chair_01", seatclass.Seats["red-chair"].AgentType)
	at := strings.Index(chair, "{")
	if at < 0 {
		t.Fatalf("the writer handed the chair no item:\n%s", chair)
	}
	var handed struct {
		Sitting map[string]json.RawMessage `json:"sitting"`
		Current *struct {
			Key string `json:"key"`
		} `json:"current"`
	}
	if err := json.Unmarshal([]byte(chair[at:]), &handed); err != nil {
		t.Fatalf("not a handed item (%v):\n%s", err, chair)
	}
	if want := head("red-chair").Key; handed.Current == nil || handed.Current.Key != want {
		t.Errorf("the chair was handed %+v; want the head of its queue, %s", handed.Current, want)
	}
	if _, whole := handed.Sitting["open"]; whole {
		t.Errorf("the chair was handed its whole list at dispatch; `show work` prints it, the opening end hands the head:\n%s", chair)
	}
	if strings.Contains(chair, sentence) {
		t.Errorf("the chair's dispatch carries report text:\n%s", chair)
	}

	// THE LENS'S HEAD IS REPORT TEXT: it is handed the one line, and none of the text.
	lensType := seatclass.Seats[lensSeat].AgentType
	lens := open("agent_lens_01", lensType)
	if h := head(lensSeat); h.Kind != record.ItemReportText {
		t.Fatalf("the fixture's lens is owed no report text at its new sitting, so its arm measures nothing: %+v", h)
	}
	if strings.TrimSpace(lens) == "" || strings.Contains(lens, "{") || strings.Contains(lens, sentence) || strings.Count(strings.TrimSpace(lens), "\n") != 0 {
		t.Errorf("the lens was handed more than the one line saying its report text arrives one item per call:\n%s", lens)
	}

	// A CONFIGURATION THAT SEATS SEVERAL SEATS IS HANDED NOTHING.
	if blue := open("agent_blue_01", "frank-exchange-of-views:blue-researcher"); blue != "" {
		t.Errorf("blue-researcher's dispatch was handed %q; a configuration that seats several gets nothing", blue)
	}
}

// THE SHIPPED HOOK WRITES THE READ, AS THE PROCESSES THE HARNESS SPAWNS. The PreToolUse binary spawns
// the writer beside it for a seat's call that returns an envelope, handing it the call's arguments; for
// the chair it prints the plan the record stored with the chair's notes, and `record_read` set to the
// record's fold over what the seat typed. An in-process test stubs that spawn, so this is what notices
// the writer's answer and the hook's reader parting.
func TestTheShippedHookWritesTheRecordsReadOverWhatTheSeatTyped(t *testing.T) {
	if testing.Short() {
		t.Skip("builds two binaries")
	}
	for _, v := range []string{seatenv.Var, seatenv.AgentVar, seatenv.TypeVar} {
		t.Setenv(v, "")
	}
	const lead, agent = "session-lead", "agent_read_01"
	cwd := t.TempDir()
	runDir := newRun(t)
	stageEvents(t, runDir)
	if _, err := runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "", lead); err != nil {
		t.Fatal(err)
	}
	seatType := "frank-exchange-of-views:red-chair"
	open := exec.Command(testbuild.Binary(t, "feov-sitting-write"), "-run", runDir, "-phase", "open", "-agent-id", agent, "-agent-type", seatType)
	if out, err := open.CombinedOutput(); err != nil {
		t.Fatalf("opening the sitting: %v\n%s", err, out)
	}
	filePetition(t, runDir)
	printed, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var plan struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal([]byte(printed), &plan); err != nil || plan.Key == "" {
		t.Fatalf("fixture: the dispatch printed no plan key (%v):\n%s", err, printed)
	}
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
				Key   string            `json:"key"`
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
	if doc.H.Updated.Key != plan.Key {
		t.Errorf("the call carries plan key %q, want the stored plan's %q:\n%s", doc.H.Updated.Key, plan.Key, out)
	}
	got := doc.H.Updated.Read
	if got.Halted || got.Event != want.Event || !slices.Equal(got.BenchOwed, want.BenchOwed) {
		t.Errorf("record_read = %+v, want the record's %+v, over the halt the seat typed", got, want)
	}
}
