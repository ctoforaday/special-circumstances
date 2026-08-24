package seatprobe

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"os"
	"strings"
	"testing"
)

// THE BUILD MUST NOT BIND THE SEAT IT IS ABOUT TO DISPATCH.
//
// Staging a board needs registered seats, so these registers happen — but binding them to the
// handle the seat is later dispatched with hands the seat a binding it did not create, and
// `register` stops being its first act because the guard is already satisfied.
//
// MEASURED: in the 2026-08-20 run one seat in nine never called `register`, made 22 tool calls,
// and recorded events anyway. Production would have refused its first write. The probe could not
// see it, because the probe had already registered it — an instrument that satisfies the guard it
// measures reports an untested guard and a compliant seat identically.
func TestBuildDoesNotBindTheSeatsItStages(t *testing.T) {
	t.Setenv("FEOV_AGENT_ID", "probe-red-merge-r1")
	var registers []string
	seen := map[string]string{}
	exec := func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "register" {
			registers = append(registers, strings.Join(args, " "))
			// Whatever handle is live at the moment of the call is what the record would bind.
			seen[strings.Join(args, " ")] = os.Getenv("FEOV_AGENT_ID")
		}
		return "", nil
	}
	runDir := tmpRun(t)
	if err := Build(runDir, Boards()["arithmetic"], exec); err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(registers) == 0 {
		t.Fatal("the build registered no seats — staging needs them, so this fixture proves nothing")
	}
	// The handle must be whatever the CALLER had, never one the build set per seat. The test sets
	// one deliberately: the build must not overwrite it with a seat-derived handle, because that
	// is the shape that bound each seat to the identity it was about to be dispatched under.
	for cmd, handle := range seen {
		if handle != "probe-red-merge-r1" {
			t.Errorf("the build changed the agent handle for %q (saw %q) — it is binding the seats it stages", cmd, handle)
		}
	}
}

// tmpRun is t.TempDir() for a RUN, plus the release its record handle needs. recordsql caches a
// *sql.DB per database and production never closes one; a test process accumulates them, and on
// Windows an open file cannot be removed so `t.TempDir` cleanup fails.
func tmpRun(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() { _ = recordsql.CloseUnder(dir) })
	return dir
}
