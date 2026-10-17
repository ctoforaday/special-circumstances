package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A WORD TYPED UNDER THE WRONG GROUP IS ANSWERED WITH THE GROUP IT IS IN. The bench's constitution
// says to read "the telemetry series", and on the 2026-10-17 smoke the bench tried `show telemetry`
// first; the series is `inquest telemetry`. The unknown-flag refusal already names the verbs that
// take a flag; this is the same answer for a command, read from the asking seat's own tree — so a
// seat whose surface has no such command is told nothing it cannot run.
func TestShowNamesTheCommandAWordBelongsTo(t *testing.T) {
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))

	_, err := run(t, "show", "telemetry", "--run", runDir, "--seat-id", "judge")
	if err == nil || !strings.Contains(err.Error(), "`inquest telemetry` is the command by that name") {
		t.Errorf("the bench's `show telemetry` = %v, want it pointed at `inquest telemetry`", err)
	}
	_, err = run(t, "show", "telemetry", "--run", runDir, "--seat-id", "red-lens-logic")
	if err == nil || strings.Contains(err.Error(), "inquest") {
		t.Errorf("a lens, whose surface has no inquest, was told about it: %v", err)
	}
	_, err = run(t, "show", "nonsense", "--run", runDir, "--seat-id", "judge")
	if err == nil || strings.Contains(err.Error(), "by that name") {
		t.Errorf("a word naming no command anywhere got a pointer: %v", err)
	}
}
