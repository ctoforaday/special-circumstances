package difftest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A PROGRAM THAT CALLS THE TOOL IS NEVER RECORDED AS REPRODUCING, driven through the shipped binary.
//
// On a smoke run a blue computation's script called the tool as blue. In every lens's re-run the call was
// refused for a seat mismatch, the script piped the refusal through `jq` into a 0, the output
// matched, and the computation was recorded as reproducing byte for byte while its check never ran. Here
// the program calls the binary as blue and prints the same 0 whether the call succeeded or was
// refused — the shape that turns any refusal into an ordinary-looking result. Under a checker's run the
// binary refuses every call and writes it to the marker the run names: `compute` refuses a program whose
// stored copy calls it, and the lens's `reproduce` of a program that first calls it on re-run records
// `calls_the_tool`, whatever the program printed.
func TestAProgramThatCallsTheToolIsNeverRecordedAsReproducing(t *testing.T) {
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	m := newMapper()
	prepareRun(t, bin, runDir, m, nil, nil)
	// THE PROGRAM STARTS THE BINARY WITHOUT A SHELL. A shell script spliced the binary's path into
	// its text, and on Windows that path is `C:\\…\\feov-record.exe`: the shell read each backslash
	// as an escape, the call never started, its error went to /dev/null with the refusal, and the
	// script printed its 0 for a reason that had nothing to do with the tool. The paths travel as
	// JSON strings here, and a call that cannot start says so in the output rather than as a 0.
	//
	// from is the execution, counted in a file beside the run, from which the program calls the tool:
	// 1 calls on every run, and 4 first on the run after `compute`'s three.
	quote := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	program := func(name string, from int) string {
		t.Helper()
		path := filepath.Join(runDir, name)
		body := `const fs = require("fs");
const { execFileSync } = require("child_process");
fs.appendFileSync(` + quote(name+".runs") + `, "x");
if (fs.readFileSync(` + quote(name+".runs") + `, "utf8").length >= ` + strconv.Itoa(from) + `) {
  try {
    execFileSync(` + quote(bin) + `, ["--run", ` + quote(runDir) + `, "--seat-id", "blue-respond", "show", "board", "--json"], { stdio: "ignore" });
  } catch (e) {
    if (typeof e.status !== "number") { console.log("the tool did not start: " + e.code); process.exit(0); }
  }
}
console.log("0");
`
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	do := func(verb string, args ...string) invocation {
		t.Helper()
		return runGo(bin, runDir, m, cmd{verb: verb, args: append([]string{"--run", "{RUN}"}, args...)})
	}
	var ev struct {
		Computations []struct {
			Anchor string `json:"anchor"`
		} `json:"computations"`
	}
	evidence := func() {
		t.Helper()
		show := do("show", "--seat-id", "red-lens-evidence", "evidence")
		if err := json.Unmarshal([]byte(show.stdout), &ev); err != nil {
			t.Fatalf("show evidence: %v\n%s", err, show.stdout)
		}
	}

	refused := do("compute", "--seat-id", "blue-respond", "--quote", "A claim sits under S2.", "--script", program("always.js", 1), "--reason", "reads the board", "--json")
	if refused.code == 0 || !strings.Contains(refused.stdout+refused.stderr, "calls feov-record") {
		t.Fatalf("compute of a program whose stored copy calls the tool: exit %d, want a refusal naming the call\nstdout: %s\nstderr: %s", refused.code, refused.stdout, refused.stderr)
	}
	if evidence(); len(ev.Computations) != 0 {
		t.Fatalf("the refused compute recorded %d computation(s)", len(ev.Computations))
	}

	comp := do("compute", "--seat-id", "blue-respond", "--quote", "A claim sits under S2.", "--script", program("later.js", 4), "--reason", "reads the board", "--json")
	if comp.code != 0 {
		t.Fatalf("compute of a program whose runs call nothing: exit %d\nstdout: %s\nstderr: %s", comp.code, comp.stdout, comp.stderr)
	}
	if evidence(); len(ev.Computations) != 1 {
		t.Fatalf("show evidence holds %d computation(s), want 1", len(ev.Computations))
	}
	anchor := ev.Computations[0].Anchor
	if r := do("register", "--seat-id", "red-lens-evidence"); r.code != 0 {
		t.Fatalf("register: %s", r.stderr)
	}
	sound := do("reproduce", "--seat-id", "red-lens-evidence", "--anchor", anchor, "--as", "sound", "--reason", "it reads the board")
	if sound.code == 0 || !strings.Contains(sound.stdout+sound.stderr, "calls feov-record") {
		t.Fatalf("a sound verdict over a program that calls the tool: exit %d\nstdout: %s\nstderr: %s", sound.code, sound.stdout, sound.stderr)
	}
	unsound := do("reproduce", "--seat-id", "red-lens-evidence", "--anchor", anchor, "--as", "unsound", "--reason", "it reads the board through the tool", "--json")
	if unsound.code != 0 || !strings.Contains(unsound.stdout, `"rerun":"calls_the_tool"`) || !strings.Contains(unsound.stdout, `"reproduced":false`) ||
		!strings.Contains(unsound.stdout, "--seat-id blue-respond show board --json") {
		t.Errorf("the re-run of a program that calls the tool: exit %d, want rerun calls_the_tool naming the call\nstdout: %s\nstderr: %s", unsound.code, unsound.stdout, unsound.stderr)
	}
}
