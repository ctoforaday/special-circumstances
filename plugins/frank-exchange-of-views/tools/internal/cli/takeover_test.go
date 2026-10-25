package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// cliTestLead is the session the CLI fixtures' setup runs in.
const cliTestLead = "session-cli-test"

// runAs runs the built binary as the operator in session (none when ""), from cwd.
func runAs(t *testing.T, bin, cwd, session string, args ...string) setupRun {
	t.Helper()
	c := exec.Command(bin, append(args, "--seat-id", "operator")...)
	c.Dir = cwd
	env := []string{"CLAUDE_PROJECT_DIR=" + recordtest.TmpRun(t)}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, seatenv.SessionVar+"=") && !strings.HasPrefix(kv, "FEOV_") {
			env = append(env, kv)
		}
	}
	if session != "" {
		env = append(env, seatenv.SessionVar+"="+session)
	}
	c.Env = env
	var so, se strings.Builder
	c.Stdout, c.Stderr = &so, &se
	err := c.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		code = -1
	}
	return setupRun{stdout: so.String(), stderr: se.String(), code: code}
}

// A RESUME IN A NEW SESSION RE-BINDS EXPLICITLY, end to end through the binary: setup records its
// session as the lead; another session's re-setup is refused naming the take-over; a take-over naming
// the wrong lead, or none, is refused and moves nothing; one naming the lead moves it; and the verb
// refuses a caller with no session.
func TestTakeOverRebindsTheLeadOnlyFromTheSessionNamed(t *testing.T) {
	bin := buildSetupBinary(t)
	cwd := recordtest.TmpRun(t)
	gitInit(t, cwd)
	if err := os.WriteFile(filepath.Join(cwd, "real.md"), []byte("exists\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommit(t, cwd)
	runDir := filepath.Join(cwd, "research", "resume")
	setupArgs := []string{"setup", runDir, "--topic", "t", "--model", "haiku", "--judgment-model", "haiku"}

	if r := runAs(t, bin, cwd, cliTestLead, setupArgs...); r.code != 0 {
		t.Fatalf("setup: exit %d: %s", r.code, r.stderr)
	}
	lead := func() string {
		m, _ := runlive.RowOf(cwd, runDir)
		return m.LeadSession
	}
	if lead() != cliTestLead {
		t.Fatalf("the run's lead after setup: %q, want %q", lead(), cliTestLead)
	}
	const resumer = "session-resuming"
	if r := runAs(t, bin, cwd, resumer, setupArgs...); r.code != 2 || !strings.Contains(r.stderr, "take-over") {
		t.Fatalf("another session's re-setup: exit %d, want 2 naming take-over: %s", r.code, r.stderr)
	}
	if r := runAs(t, bin, cwd, resumer, "take-over", runDir, "--from-session", "not-the-lead"); r.code == 0 || !strings.Contains(r.stderr, cliTestLead) {
		t.Fatalf("a take-over naming the wrong lead: exit %d, want a refusal naming %s: %s%s", r.code, cliTestLead, r.stdout, r.stderr)
	}
	if lead() != cliTestLead {
		t.Fatalf("a refused take-over moved the lead to %q", lead())
	}
	if r := runAs(t, bin, cwd, "", "take-over", runDir, "--from-session", cliTestLead); r.code == 0 || !strings.Contains(r.stderr, seatenv.SessionVar) {
		t.Fatalf("a take-over with no session: exit %d, want a refusal naming %s: %s", r.code, seatenv.SessionVar, r.stderr)
	}
	if r := runAs(t, bin, cwd, resumer, "take-over", runDir); r.code == 0 || !strings.Contains(r.stderr, "requires --from-session") {
		t.Fatalf("a take-over naming no lead was accepted: %s", r.stdout)
	}
	if r := runAs(t, bin, cwd, resumer, "take-over", runDir, "--from-session", cliTestLead); r.code != 0 || !strings.Contains(r.stdout, "now led by the calling session") {
		t.Fatalf("the take-over: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if lead() != resumer {
		t.Fatalf("the run's lead after the take-over: %q, want %q", lead(), resumer)
	}
	// The old lead's re-setup is now the one refused.
	if r := runAs(t, bin, cwd, cliTestLead, setupArgs...); r.code != 2 || !strings.Contains(r.stderr, "LED BY ANOTHER SESSION") {
		t.Fatalf("the old lead's re-setup after the take-over: exit %d: %s", r.code, r.stderr)
	}
}

// A SECOND SESSION'S RUN IN A LIVE PROJECT IS REFUSED, and the refusal leads with another checkout:
// a take-over is never its first remedy, because taking over a running session's run moves that
// session's seats off it. A session that leads a live run cannot take over another, and a row that
// names no lead session is taken over with the --from-session the refusal prints.
func TestASecondSessionsSetupLeadsWithAnotherCheckout(t *testing.T) {
	bin := buildSetupBinary(t)
	cwd := recordtest.TmpRun(t)
	gitInit(t, cwd)
	if err := os.WriteFile(filepath.Join(cwd, "real.md"), []byte("exists\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommit(t, cwd)
	setupArgs := func(runDir string) []string {
		return []string{"setup", runDir, "--topic", "t", "--model", "haiku", "--judgment-model", "haiku"}
	}
	first := filepath.Join(cwd, "research", "first")
	if r := runAs(t, bin, cwd, cliTestLead, setupArgs(first)...); r.code != 0 {
		t.Fatalf("setup: exit %d: %s", r.code, r.stderr)
	}
	firstRemedy := func(stderr string) string {
		for _, l := range strings.Split(stderr, "\n") {
			if strings.HasPrefix(l, "    - ") {
				return l
			}
		}
		return ""
	}
	const second = "session-second"
	for _, runDir := range []string{filepath.Join(cwd, "research", "second"), first} {
		r := runAs(t, bin, cwd, second, setupArgs(runDir)...)
		if r.code != 2 {
			t.Fatalf("a second session's setup of %s in a live project: exit %d, want 2: %s", runDir, r.code, r.stderr)
		}
		if fr := firstRemedy(r.stderr); !strings.Contains(fr, "another checkout") || strings.Contains(fr, "take-over") {
			t.Errorf("the first remedy for %s must be another checkout, not a take-over: %q\n%s", runDir, fr, r.stderr)
		}
	}
	if _, err := os.Stat(filepath.Join(cwd, "research", "second")); !os.IsNotExist(err) {
		t.Errorf("the refused run was built: %v", err)
	}

	// A session already leading a live run takes over no other.
	marker := filepath.Join(cwd, ".claude", "run-live.json")
	b, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(cwd, "research", "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	two := strings.Replace(string(b), `"runs": [`, `"runs": [{"runDir": "research/other", "leadSession": ""},`, 1)
	if two == string(b) {
		t.Fatalf("the marker has no runs list to extend:\n%s", b)
	}
	if err := os.WriteFile(marker, []byte(two), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := runAs(t, bin, cwd, cliTestLead, "take-over", other, "--from-session", ""); r.code == 0 || !strings.Contains(r.stderr, "already leads the live run") {
		t.Fatalf("a take-over by a session leading another live run: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	// A row naming no lead session: the refusal names it as `--from-session ""`, and that form works.
	if r := runAs(t, bin, cwd, second, "take-over", other, "--from-session", "unknown"); r.code == 0 || !strings.Contains(r.stderr, `--from-session ""`) {
		t.Fatalf("a take-over naming a session the row does not: exit %d, want a refusal printing --from-session \"\": %s", r.code, r.stderr)
	}
	if r := runAs(t, bin, cwd, second, "take-over", other, "--from-session", ""); r.code != 0 {
		t.Fatalf("the printed take-over: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if m, _ := runlive.RowOf(cwd, other); m.LeadSession != second {
		t.Fatalf("the run's lead session after the take-over: %q, want %q", m.LeadSession, second)
	}
}
