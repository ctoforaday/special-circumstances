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

// A LIVE RUN ANOTHER SESSION LEADS IS NO AMBIGUITY: its seats carry the other session's id, so a
// second session launches beside it in the same project.
func TestSetupLaunchesBesideAnotherSessionsLiveRun(t *testing.T) {
	cfg, _ := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	if err := os.MkdirAll(filepath.Join(cfg.Cwd, "research", "2026-08-01_running"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, "research/2026-08-01_running", nil, time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), "", "", "another-session"); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d beside another session's run, want 0:\n%s%s", code, out.String(), errb.String())
	}
	if rows := runlive.ReadRunLive(cfg.Cwd); len(rows) != 2 {
		t.Fatalf("rows: %+v, want both runs", rows)
	}
}

// A RE-SETUP IS ITS LEAD'S OWN. Another session re-running setup on a live run is refused, naming the
// lead and the take-over; the row's lead stands.
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
	for _, want := range []string{"LED BY ANOTHER SESSION", testLead, "take-over", "--from-session " + testLead} {
		if !strings.Contains(errb.String(), want) {
			t.Errorf("the refusal must carry %q:\n%s", want, errb.String())
		}
	}
	if m, _ := runlive.RowOf(cfg.Cwd, cfg.RunDir); m.LeadSession != testLead {
		t.Fatalf("the lead moved under a refusal: %q", m.LeadSession)
	}
}
