package seatprobe

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"path/filepath"
	"strings"
	"testing"
)

// writeTrajectory fakes the shape the `claude` CLI emits: one JSON object per line, assistant
// turns carrying tool_use blocks.
func writeTrajectory(t *testing.T, commands ...string) string {
	t.Helper()
	var b strings.Builder
	for _, c := range commands {
		esc := strings.ReplaceAll(c, `\`, `\\`)
		esc = strings.ReplaceAll(esc, `"`, `\"`)
		b.WriteString(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"` + esc + `"}}]}}` + "\n")
	}
	p := filepath.Join(t.TempDir(), "trajectory.jsonl")
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAttemptedSeesWhatTheRecordCannot(t *testing.T) {
	sf := surface()
	p := writeTrajectory(t,
		`ls -la /run`,
		`"C:/tools/fxr.exe" blue show --run /run --view board`,
		`"C:/tools/fxr.exe" blue show --run /run --view findings`,
		`"C:/tools/fxr.exe" blue edit --old "x" --new "y" --reason "z"`,
		`grep -r show /run`,
	)
	got, err := Attempted(p, "fxr.exe", sf, "blue")
	if err != nil {
		t.Fatal(err)
	}
	// `show` is a READ. It records no event, so Read() can never see it and reported it as a
	// verb the seat never touched — on every board of the first separated run, while the seats
	// were calling it up to nine times each.
	if got["show"] != 2 {
		t.Errorf("show counted %d, want 2 — a read leaves no event, so the trajectory is the only witness", got["show"])
	}
	if got["edit"] != 1 {
		t.Errorf("edit counted %d, want 1", got["edit"])
	}
	// The seat's own shell work must not become phantom verbs: `ls` and `grep show` both
	// mention a verb name and invoke nothing.
	if len(got) != 2 {
		t.Errorf("counted %v — only invocations OF THE BINARY are verbs", got)
	}
}

func TestAttemptedTakesTheLongestVerbNotTheFirstToken(t *testing.T) {
	sf := surface()
	p := writeTrajectory(t,
		`"/x/fxr.exe" motion grade file --id R1-1 --dimension severity --proposed low --reason "r"`,
	)
	got, err := Attempted(p, "fxr.exe", sf, "blue")
	if err != nil {
		t.Fatal(err)
	}
	// Taking the first token files this under `motion`, and every expectation is written
	// against `motion grade file` — so the verb becomes unmeetable BY CONSTRUCTION. The same
	// shortcut in the fuzzer's trajectory reader reported all seven motion verbs as never
	// invoked while 35 motion events sat in the record.
	if got["motion grade file"] != 1 {
		t.Fatalf("counted %v, want motion grade file×1", got)
	}
}

func TestAttemptedIgnoresAnotherToolNamedLikeAVerb(t *testing.T) {
	sf := surface()
	p := writeTrajectory(t, `/usr/bin/other blue show --view board`)
	got, err := Attempted(p, "fxr.exe", sf, "blue")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("counted %v from a command that never ran the tool", got)
	}
}

// The report must keep the two sources apart. Collapsing them is how the first version claimed a
// seat NEVER REACHED a verb it had just called nine times.
func TestReportSeparatesRecordedFromInvoked(t *testing.T) {
	runDir := writeRun(t, []struct {
		seat, typ string
		payload   *record.Payload
	}{
		{seat: "blue-respond-r1", typ: "position", payload: record.NewPayload().Set("reason", "n")},
	})

	attempts := map[string]map[string]int{"blue-respond-r1": {"show": 9}}
	out, err := Report(surface(), runDir, []string{"blue-respond-r1"}, nil, attempts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "invoked, no event") || !strings.Contains(out, "show×9") {
		t.Errorf("a verb invoked but unrecorded is not surfaced:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "- NEVER REACHED:") && strings.Contains(line, "show") {
			t.Errorf("`show` is listed as never reached although the seat called it 9 times: %s", line)
		}
	}

	// And with no trajectory the report must SAY it did not measure, rather than presenting a
	// record-only count as the whole picture — the absent case and the honest zero again.
	out, err = Report(surface(), runDir, []string{"blue-respond-r1"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "NOT MEASURED") {
		t.Errorf("with no trajectory the report must declare reads and refusals unmeasured:\n%s", out)
	}
}

// FLAGS BEFORE THE VERB. A real probed seat wrote every call this way, and the first version of
// this counter stopped at the leading dash — reporting a seat with 22 tool calls as having
// reached for nothing. A flag's VALUE must not nominate a verb either, which is what the
// --seat-id and --reason cases pin.
func TestAttemptedFindsTheVerbAfterLeadingFlags(t *testing.T) {
	sf := surface()
	p := writeTrajectory(t,
		`"/s/fxr.exe" --run "." --seat-id red-lens-r1-L1 lens finding --key F1 --reason "r"`,
		`"/s/fxr.exe" --run "." --seat-id red-lens-r1-L1 lens show --view findings`,
		`"/s/fxr.exe" --run "." --seat-id red-lens-r1-L1 record lens finding --key F2 --reason "r"`,
	)
	got, err := Attempted(p, "fxr.exe", sf, "lens")
	if err != nil {
		t.Fatal(err)
	}
	if got["finding"] != 2 {
		t.Errorf("finding counted %d, want 2 (one of them behind a stray `record` token the seat typed): %v", got["finding"], got)
	}
	if got["show"] != 1 {
		t.Errorf("show counted %d, want 1: %v", got["show"], got)
	}
}

func TestAFlagValueNeverNominatesAVerb(t *testing.T) {
	sf := surface()
	// `close` and `verify` are real merge/lens verbs; here they are prose inside --reason.
	p := writeTrajectory(t, `"/s/fxr.exe" lens finding --key F1 --reason "close the gap and verify it"`)
	got, err := Attempted(p, "fxr.exe", sf, "lens")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got["finding"] != 1 {
		t.Fatalf("counted %v — prose inside a flag's value was read as verbs", got)
	}
}
