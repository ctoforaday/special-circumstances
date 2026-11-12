package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// A SEAT CALL THAT NOTHING BINDS TO A RUN IS REFUSED, AND THE REFUSAL IS THE LAST WORD.
//
// The run reaches a seat's call by the hook's FEOV_RUN, else by the tool's own inference over the
// live runs the calling session leads. A call with neither — no hook fired, and the session leads no
// live run — is refused with "--run <runDir> is required", and the run that IS live in the project
// receives nothing: no third carrier stands behind those two. Driven through the built binary, as a
// seat runs it, so the environment is the only thing that can bind the run.
//
// The control half is what makes the refusal mean something: the same call from the run's lead
// session binds the run by inference, so the refusal is the missing carrier and not a broken run.
func TestASeatCallNothingBindsIsRefusedAndWritesNothing(t *testing.T) {
	bin := testbuild.Binary(t, "feov-record")
	proj := recordtest.TmpRun(t)
	runDir := filepath.Join(proj, "research", "live-run")
	if err := os.MkdirAll(filepath.Join(proj, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"),
		[]byte(`{"runs":[{"runDir":"research/live-run","leadSession":"`+cliTestLead+`"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	env := func(session string) []string {
		return []string{seatenv.Var + "=", "CLAUDE_PROJECT_DIR=" + proj, seatenv.SessionVar + "=" + session}
	}

	stdout, stderr, code := execForms(t, bin, env("a-session-that-leads-no-run"), "register", "--seat-id", "red-lens-evidence")
	if code == 0 {
		t.Fatalf("a call no carrier bound to a run was served:\n%s", stdout)
	}
	if !strings.Contains(stderr, "--run <runDir> is required") {
		t.Errorf("refused for another reason (exit %d): %s", code, stderr)
	}
	if entries, _ := os.ReadDir(runDir); len(entries) != 0 {
		t.Errorf("the refused call wrote into the live run: %v", entries)
	}

	if _, stderr, code := execForms(t, bin, env(cliTestLead), "register", "--seat-id", "red-lens-evidence"); code != 0 {
		t.Fatalf("control: the run's lead session did not bind the run by inference (exit %d): %s", code, stderr)
	}
	if entries, _ := os.ReadDir(runDir); len(entries) == 0 {
		t.Error("control: the lead session's register wrote nothing into the live run")
	}
}
