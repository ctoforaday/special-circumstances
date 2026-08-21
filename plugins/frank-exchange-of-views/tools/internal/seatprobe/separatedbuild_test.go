package seatprobe

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY BOARD MUST BUILD THE WAY cmd/seatprobe BUILDS IT: through a SUBPROCESS, with the record
// root in that subprocess's environment and NOT in this one.
//
// # Why this cannot be an in-process gate
//
// internal/cli's TestEveryProbeBoardStillBuilds drives the cobra root in the test's own process.
// That is cheap and it catches write paths moving under the fixtures, which is what it is for. It
// cannot catch anything about the RECORD ROOT, and the reason is structural rather than a matter of
// coverage: the probe hands the root to each tool subprocess in `cmd.Env`, so `Build` — running in
// the probe's own process — resolves the record directory WITHOUT it. In an in-process test there
// is one process, `t.Setenv` is visible to both halves, and the discrepancy the probe lives with
// cannot be represented at all.
//
// MEASURED, TWICE, IN ONE AFTERNOON, and neither was found by a test:
//
//	Making `--class` strict stopped 8 of 9 boards on their first MINT. The in-process gate stayed
//	green because a sweep had handed it a staged registry the probe does not stage.
//
//	Fixing that in the wrong place — staging before the first register, from a process carrying no
//	record root — created <runDir>/records before the tool could separate the run, and stopped 9 of
//	9 boards on their first REGISTER. The in-process gate stayed green again, and this time it was
//	green for a reason no amount of care would have changed: with the root in the test's own
//	environment, `Build` resolved to the separated directory too, so there was never a local
//	records/ to collide with.
//
// So this one pays for a binary. That is the price of testing the configuration that ships.
func TestEveryProbeBoardBuildsThroughASubprocessWithASeparatedRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary")
	}
	bin := filepath.Join(t.TempDir(), "feov-record")
	if out, err := exec.Command("go", "build", "-o", bin, "../../cmd/feov-record").CombinedOutput(); err != nil {
		t.Fatalf("build feov-record: %v\n%s", err, out)
	}

	boards := Boards()
	if len(boards) == 0 {
		t.Fatal("no probe boards at all — an empty set builds cleanly forever and reports nothing")
	}
	for name, b := range boards {
		b := b
		t.Run(name, func(t *testing.T) {
			runDir := t.TempDir()
			// THE ROOT GOES TO THE SUBPROCESS ONLY, exactly as cmd/seatprobe declares it. Setting
			// it with t.Setenv instead would put it in this process too, and this test would
			// become the in-process one it exists to be different from.
			recordRoot := t.TempDir()
			run := func(args ...string) (string, error) {
				c := exec.Command(bin, args...)
				c.Env = append(os.Environ(),
					record.RecordRootEnv+"="+recordRoot,
					"CLAUDE_PROJECT_DIR="+runDir)
				out, err := c.CombinedOutput()
				if err != nil {
					return string(out), fmt.Errorf("%s: %s", strings.Join(args[:min(3, len(args))], " "), strings.TrimSpace(string(out)))
				}
				return string(out), nil
			}

			if err := Build(runDir, b, run); err != nil {
				t.Fatalf("board %q does not build with a separated record: %v\n\n"+
					"This is the configuration cmd/seatprobe runs. A board that builds in-process and\n"+
					"not here means every real dispatch dies before a seat is handed anything, while\n"+
					"the in-process gate reports the fixtures healthy.", name, err)
			}
			// AND THE SEPARATION ACTUALLY HAPPENED. A board that built because the records stayed
			// local has not been separated at all, and the probe's whole design — a seat that
			// cannot read the board off disk — is off.
			if _, err := os.Stat(filepath.Join(runDir, "records")); !os.IsNotExist(err) {
				t.Errorf("board %q left records/ inside the run: the seat can read the board off disk, "+
					"so the separation the probe depends on did not happen", name)
			}
		})
	}
}
