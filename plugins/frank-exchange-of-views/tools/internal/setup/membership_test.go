package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// SETUP RECORDS THE LAUNCHING SESSION AS THE RUN'S LEAD, and refuses a run with none: the hooks
// attribute an agent to a run only when its payload's session_id is the run's lead.
func TestSetupRecordsTheLaunchingSessionAndRefusesNone(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	cfg.Session = ""
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 2 || !strings.Contains(errb.String(), "NO LAUNCHING SESSION") {
		t.Fatalf("exit %d with no session, want 2 and the refusal:\n%s", code, errb.String())
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Fatalf("a run with no lead was built: %v", err)
	}
	cfg.Session = testLead
	out.Reset()
	errb.Reset()
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d:\n%s%s", code, out.String(), errb.String())
	}
	m, ok := runlive.RowOf(cfg.Cwd, cfg.RunDir)
	if !ok || m.LeadSession != testLead {
		t.Fatalf("the run's row: %+v (found %v), want lead %q", m, ok, testLead)
	}
}

// ONE LIVE RUN PER PROJECT, WHATEVER SESSION LEADS IT. The hooks tell two sessions' runs apart, but
// the runs still share the project's run directories and the memory capture writes, so a second
// session's run is refused, and its first remedy is another checkout — never a take-over, which is
// for a resume.
func TestSetupRefusesBesideAnotherSessionsLiveRun(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	if err := os.MkdirAll(filepath.Join(cfg.Cwd, "research", "2026-08-01_running"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, "research/2026-08-01_running", nil, time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), "", "", "another-session"); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("exit %d beside another session's run, want 2:\n%s%s", code, out.String(), errb.String())
	}
	for _, want := range []string{"ANOTHER RUN IS LIVE IN THIS PROJECT", "research/2026-08-01_running", "another-session"} {
		if !strings.Contains(errb.String(), want) {
			t.Errorf("the refusal must carry %q:\n%s", want, errb.String())
		}
	}
	if first := firstRemedy(errb.String()); !strings.Contains(first, "another checkout") {
		t.Errorf("the first remedy must be another checkout, got %q:\n%s", first, errb.String())
	}
	if strings.Contains(errb.String(), "take-over") {
		t.Errorf("a new run beside another session's is not a take-over:\n%s", errb.String())
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Errorf("the refused run was built: %v", err)
	}
	if rows := runlive.ReadRunLive(cfg.Cwd); len(rows) != 1 {
		t.Fatalf("rows: %+v, want only the live run's", rows)
	}
}

// firstRemedy is the first "    - " line of a refusal.
func firstRemedy(stderr string) string {
	for _, l := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(l, "    - ") {
			return l
		}
	}
	return ""
}

// A RE-SETUP IS ITS LEAD SESSION'S OWN. Another session re-running setup on a live run is refused,
// naming the run's lead session; the first remedy is another checkout, and the take-over is offered
// only for a resume whose lead session has ended. The row's lead session stands.
func TestSetupRefusesAnotherSessionsReSetupOfALiveRun(t *testing.T) {
	cfg, _ := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("first setup: exit %d:\n%s%s", code, out.String(), errb.String())
	}
	cfg.Session = "a-resuming-session"
	out.Reset()
	errb.Reset()
	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("exit %d for another session's re-setup, want 2:\n%s%s", code, out.String(), errb.String())
	}
	for _, want := range []string{"LED BY ANOTHER SESSION", testLead, "RESUMING", "has ended", "take-over", "--from-session " + testLead} {
		if !strings.Contains(errb.String(), want) {
			t.Errorf("the refusal must carry %q:\n%s", want, errb.String())
		}
	}
	if first := firstRemedy(errb.String()); !strings.Contains(first, "another checkout") {
		t.Errorf("the first remedy must be another checkout, got %q:\n%s", first, errb.String())
	}
	if m, _ := runlive.RowOf(cfg.Cwd, cfg.RunDir); m.LeadSession != testLead {
		t.Fatalf("the run's lead session moved under a refusal: %q", m.LeadSession)
	}
}

// A ROW THAT NAMES NO LEAD SESSION — written before setup recorded one — is refused for a re-setup
// like any other session's, and the take-over it prints is the form take-over accepts: --from-session "".
func TestSetupPrintsAWorkingTakeOverForARowNamingNoLeadSession(t *testing.T) {
	cfg, _ := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("first setup: exit %d:\n%s%s", code, out.String(), errb.String())
	}
	marker := filepath.Join(cfg.Cwd, ".claude", "run-live.json")
	b, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	stripped := strings.Replace(string(b), `"leadSession": "`+testLead+`"`, `"leadSession": ""`, 1)
	if stripped == string(b) {
		t.Fatalf("the marker names no lead session to strip:\n%s", b)
	}
	if err := os.WriteFile(marker, []byte(stripped), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errb.Reset()
	if code := Run(cfg, &out, &errb); code != 2 || !strings.Contains(errb.String(), `--from-session ""`) {
		t.Fatalf("exit %d, want 2 printing --from-session \"\":\n%s", code, errb.String())
	}
	if _, err := runlive.TakeOver(cfg.Cwd, cfg.RunDir, "", testLead); err != nil {
		t.Fatalf("the printed take-over is refused: %v", err)
	}
}
