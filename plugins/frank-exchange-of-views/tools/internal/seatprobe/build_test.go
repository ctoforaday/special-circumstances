package seatprobe

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
	t.Setenv("FEOV_AGENT_ID", "probe-red-chair")
	var registers []string
	seen := map[string]string{}
	calls := 0
	exec := func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "register" {
			registers = append(registers, strings.Join(args, " "))
			// Whatever handle is live at the moment of the call is what the record would bind.
			seen[strings.Join(args, " ")] = os.Getenv("FEOV_AGENT_ID")
		}
		// The build stages every later act through the id the tool REPORTS, so a stand-in tool
		// reports one per mint — distinct, and in the shape the build reads.
		calls++
		switch {
		case len(args) > 0 && args[0] == "mint":
			return fmt.Sprintf("minted G-%08x", calls), nil
		case len(args) > 1 && args[0] == "avenue" && args[1] == "propose":
			return fmt.Sprintf("avenue Q-%08x recorded", calls), nil
		case len(args) > 2 && args[0] == "motion" && args[2] == "file":
			return fmt.Sprintf("motion M-%08x filed", calls), nil
		}
		return "", nil
	}
	runDir := recordtest.TmpRun(t)
	if err := Build(runtest.Open(t, runDir), Boards()["arithmetic"], exec); err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(registers) == 0 {
		t.Fatal("the build registered no seats — staging needs them, so this fixture proves nothing")
	}
	// The handle must be whatever the CALLER had, never one the build set per seat. The test sets
	// one deliberately: the build must not overwrite it with a seat-derived handle, because that
	// is the shape that bound each seat to the identity it was about to be dispatched under.
	for cmd, handle := range seen {
		if handle != "probe-red-chair" {
			t.Errorf("the build changed the agent handle for %q (saw %q) — it is binding the seats it stages", cmd, handle)
		}
	}
}

// THE BUILD READS THE ID THE TOOL MINTED, in the shape the tool mints.
//
// Every act the build stages after a mint names the minted id, so a reader that matches nothing
// returns "" and the board fails to build — or, where the id is optional, stages an unruled line
// that reads as a fixture choice. Each row is the confirmation its verb prints; none of these ids
// is matched by a letter-and-digits pattern, and a kind reads only its own letter.
func TestMintedReadsTheIDOfItsKindFromTheVerbsConfirmation(t *testing.T) {
	for _, c := range []struct{ kind, out, want string }{
		{"avenue", "avenue Q-7cdcd115 recorded (proposed): is the 42% figure reproducible", "Q-7cdcd115"},
		{"gap", "minted G-0badf00d\n  found_by: F-7cdcd115 (evidence)", "G-0badf00d"},
		{"motion", "motion M-00c0ffee filed (docket)", "M-00c0ffee"},
		// A kind reads only its own letter: the gap a motion names is not the motion's id.
		{"motion", "closing filed for G-0badf00d", ""},
		{"avenue", "minted G-0badf00d", ""},
		// Exactly eight hex: a shorter or longer run is not an id.
		{"avenue", "avenue Q-7cdcd11 recorded", ""},
		{"avenue", "avenue Q-7cdcd1155 recorded", ""},
		{"avenue", "avenue Q1 recorded", ""},
	} {
		if got := minted(c.kind, c.out); got != c.want {
			t.Errorf("minted(%q, %q) = %q, want %q", c.kind, c.out, got, c.want)
		}
	}
}
