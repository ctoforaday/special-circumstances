package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE M31 SEQUENCE THROUGH THE VERBS: a seat's first act registers it silently, then the seat runs
// `register` itself. The fixture's own register (no agent id) opened sitting 1 and the silent one
// opened sitting 2, so each explicit register is refused naming dispatch 2, and writes nothing.
func TestRegisterAfterTheFirstActRegisteredItIsRefused(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_lens_first_act_registered")
	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--class", "safety", "--relief", "r", "--reason", "x"); err != nil {
		t.Fatalf("the first act: %v", err)
	}
	events := func() (all, registers int) {
		t.Helper()
		fam, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range fam.Events {
			switch {
			case e.GetType() == recordpb.EventType_EVENT_TYPE_LOG:
				// a refusal is logged; it is not the register
			case e.GetSeatId() == "red-lens-evidence" && e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER:
				registers++
				all++
			default:
				all++
			}
		}
		return all, registers
	}
	before, regsBefore := events()
	for range 2 {
		_, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence")
		if err == nil || feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), "already registered in this sitting (dispatch 2, register at event") {
			t.Errorf("the explicit register after an auto-register = %v; want it refused naming dispatch 2", err)
		}
	}
	if after, regs := events(); after != before || regs != regsBefore {
		t.Errorf("the explicit registers wrote %d event(s) (%d -> %d registers)", after-before, regsBefore, regs)
	}
}

// A LENS'S WHOLE READ RECORDS NOTHING, SO ITS REGISTER AFTER IT IS ITS REGISTER; ITS DELIVERY IS THEN
// ITS FIRST RECORD, AND A SECOND REGISTER IS REFUSED NAMING IT. Measured on the #1565 smoke: four lenses
// read the report whole, ran `register`, were refused naming only the dispatch, and blamed the dispatch
// bracket. A read writes nothing — no report_read and no register — so the register that follows it
// opens the sitting; once the writer has handed the delivery's last item, the report_read it wrote is
// the first record a repeated register's refusal names.
func TestARegisterAfterAWholeReadRegistersAndARepeatNamesTheDelivery(t *testing.T) {
	runDir := seatRunReport(t, servedReport)
	const agent = "agent_lens_read_first"
	t.Setenv(seatenv.AgentVar, agent)
	ofLens := func() (registers int, registered, read int64) {
		t.Helper()
		m, err := record.MergedEvents(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		ids := m.At.IDs(m.Events)
		for i, e := range m.Events {
			if e.GetSeatId() != lensSeat {
				continue
			}
			switch e.GetType() {
			case recordpb.EventType_EVENT_TYPE_REGISTER:
				registers++
				registered = ids[i]
			case recordpb.EventType_EVENT_TYPE_REPORT_READ:
				read = ids[i]
			}
		}
		return registers, registered, read
	}
	before, _, _ := ofLens()
	if _, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatalf("the whole read: %v", err)
	}
	if n, _, read := ofLens(); n != before || read != 0 {
		t.Fatalf("the whole read recorded %d register(s) and a read at event %d; want neither", n-before, read)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatalf("the register after the whole read was refused: %v", err)
	}
	_, registered, _ := ofLens()

	if n := writerHands(t, runDir, agent, lensSeat); n == 0 {
		t.Fatal("the writer handed the lens no item, so the refusal below measures nothing")
	}
	_, _, read := ofLens()
	if read <= registered {
		t.Fatalf("the delivery was recorded at event %d; want a report_read after the register at event %d", read, registered)
	}
	_, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat)
	want := fmt.Sprintf("(dispatch 2, register at event %d)", registered)
	named := fmt.Sprintf("your report_read at event %d is your first record after it", read)
	if err == nil || feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), named) {
		t.Errorf("the register after the delivery = %v; want a conflict naming %q and %q", err, want, named)
	}
}
