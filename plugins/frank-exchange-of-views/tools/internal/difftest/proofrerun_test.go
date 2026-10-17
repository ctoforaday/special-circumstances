package difftest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A PROOF WHOSE SCRIPT CALLS THE TOOL NEVER RECORDS "REPRODUCED", driven through the shipped binary.
//
// On a smoke run a blue proof's script called the tool as blue. In every lens's re-run the call was
// refused for a seat mismatch, the script piped the refusal through `jq` into a 0, the output
// matched, and the proof was recorded as reproducing byte for byte while its check never ran. Here
// the script calls the binary as blue and prints only whether the call SUCCEEDED as a number — the
// shape that turns any refusal into an ordinary-looking result. Under a checker's run the binary
// refuses every call and writes it to the marker the run names, so `prove` grades the script
// `not_portable` and the lens's `reproduce` records `calls_the_tool`, whatever the script printed.
func TestAProofThatCallsTheToolIsNeverRecordedAsReproducing(t *testing.T) {
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	m := newMapper()
	prepareRun(t, bin, runDir, m, nil)
	script := filepath.Join(runDir, "reads-the-board.sh")
	body := "#!/bin/sh\n" + bin + " --run " + runDir + " --seat-id blue-respond show board --json >/dev/null 2>&1; echo 0\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	do := func(verb string, args ...string) invocation {
		t.Helper()
		return runGo(bin, runDir, m, cmd{verb: verb, args: append([]string{"--run", "{RUN}"}, args...)})
	}
	prove := do("prove", "--seat-id", "blue-respond", "--quote", "A claim sits under S2.", "--script", script, "--reason", "reads the board", "--json")
	if prove.code != 0 || !strings.Contains(prove.stdout, "not_portable") || !strings.Contains(prove.stdout, "calls feov-record") {
		t.Fatalf("prove of a script that calls the tool: exit %d, want not_portable naming the call\nstdout: %s\nstderr: %s", prove.code, prove.stdout, prove.stderr)
	}
	var ev struct {
		Proofs []struct {
			Anchor string `json:"anchor"`
		} `json:"proofs"`
	}
	show := do("show", "--seat-id", "red-lens-evidence", "evidence")
	if err := json.Unmarshal([]byte(show.stdout), &ev); err != nil || len(ev.Proofs) != 1 {
		t.Fatalf("show evidence: %v\n%s", err, show.stdout)
	}
	anchor := ev.Proofs[0].Anchor
	if r := do("register", "--seat-id", "red-lens-evidence"); r.code != 0 {
		t.Fatalf("register: %s", r.stderr)
	}
	sound := do("reproduce", "--seat-id", "red-lens-evidence", "--anchor", anchor, "--as", "sound", "--reason", "it reads the board")
	if sound.code == 0 || !strings.Contains(sound.stdout+sound.stderr, "calls feov-record") {
		t.Fatalf("a sound verdict over a script that calls the tool: exit %d\nstdout: %s\nstderr: %s", sound.code, sound.stdout, sound.stderr)
	}
	unsound := do("reproduce", "--seat-id", "red-lens-evidence", "--anchor", anchor, "--as", "unsound", "--reason", "it reads the board through the tool", "--json")
	if unsound.code != 0 || !strings.Contains(unsound.stdout, `"rerun":"calls_the_tool"`) || !strings.Contains(unsound.stdout, `"reproduced":false`) ||
		!strings.Contains(unsound.stdout, "--seat-id blue-respond show board --json") {
		t.Errorf("the re-run of a script that calls the tool: exit %d, want rerun calls_the_tool naming the call\nstdout: %s\nstderr: %s", unsound.code, unsound.stdout, unsound.stderr)
	}
}
