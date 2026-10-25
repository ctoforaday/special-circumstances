package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// A LOOK REGISTERS NO ONE, so the seat's own register after it goes through.
//
// universe-m33: a voice lens screened a candidate with `near-match`, which records nothing, and the
// read registered it as a side effect. Its explicit `register` two calls later was refused as a
// repeat, and it spent a log entry on the friction. Only an act that records registers a seat.
func TestALookRegistersNoOneAndTheSeatsRegisterGoesThrough(t *testing.T) {
	runDir := seatRun(t)
	const agent = "agent_that_looked_first"
	t.Setenv(seatenv.AgentVar, agent)

	if _, err := run(t, "near-match", "--run", runDir, "--seat-id", lensSeat, "--problem", "a candidate defect"); err != nil {
		t.Fatalf("an unregistered lens could not screen a candidate: %v", err)
	}
	if seatID, found, err := record.RegisteredSeatOfAgent(runtest.Open(t, runDir), agent); err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatalf("near-match records nothing and registered the agent as %q", seatID)
	}

	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatalf("the seat's register after a look was refused: %v", err)
	}
	// THE STANDING RULE IS UNTOUCHED: a second register in the sitting is still refused.
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err == nil ||
		!strings.Contains(err.Error(), "already registered in this sitting") {
		t.Fatalf("a repeat register in one sitting was not refused as a repeat: %v", err)
	}
}

// NO COMMAND THAT RECORDS NOTHING REGISTERS ITS SEAT — every surface walked, so a read added
// tomorrow is held to it without being listed.
//
// A command records nothing when, run, it appends no event but a register: a record type is not
// the whole answer, because a lens's whole `show report` records a report_read. Each runs with an
// agent handle and no register on the record, through its own RunE with only the run and the seat
// id parsed, so a read that needs more flags may refuse — what it may not do is leave a register
// behind.
func TestNoCommandThatRecordsNothingRegistersItsSeat(t *testing.T) {
	runDir := seatRun(t)
	roles := make([]string, 0, len(AllRoots()))
	for role := range AllRoots() {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	reads := 0
	for _, role := range roles {
		surface := seat.DispatchedAs(AllRoots()[role])
		for _, path := range runnableCommands(t, surface, runDir) {
			leaf, _, err := NewRootFor(surface).Find(path)
			if err != nil {
				t.Fatal(err)
			}
			if seat.RecordType(leaf) != "" {
				continue
			}
			// ONE AGENT PER COMMAND: a register left by one would bind the next and hide it.
			agent := "agent_looking_" + role + "_" + strings.Join(path, "_")
			t.Setenv(seatenv.AgentVar, agent)
			before := recordsBesidesRegisters(t, runDir)
			_, _ = callLeaf(t, surface, path, runDir)
			if recordsBesidesRegisters(t, runDir) != before {
				// IT RECORDED SOMETHING — a lens's whole `show report` is a report_read — so it is
				// an act that records, and registering its seat is the rule working.
				t.Setenv(seatenv.AgentVar, "")
				continue
			}
			reads++
			if seatID, found, err := record.RegisteredSeatOfAgent(runtest.Open(t, runDir), agent); err != nil {
				t.Fatal(err)
			} else if found {
				t.Errorf("%s/%s records nothing and registered the agent as %q", role, strings.Join(path, " "), seatID)
			}
			t.Setenv(seatenv.AgentVar, "")
		}
	}
	// A walk that stops reaching the reads measures nothing and reports green.
	if reads < 40 {
		t.Fatalf("walked %d command(s) that record nothing — the walk is not reaching the reads", reads)
	}
}

// recordsBesidesRegisters counts the run's events that are not registers.
func recordsBesidesRegisters(t *testing.T, runDir string) int {
	t.Helper()
	n := 0
	for _, e := range events(t, runDir) {
		if _, ok := recordpb.BodyAs[*recordpb.Register](e); !ok {
			n++
		}
	}
	return n
}
