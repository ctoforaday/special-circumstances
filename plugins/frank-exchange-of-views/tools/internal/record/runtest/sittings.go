package runtest

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// SittingsOpened is how many sittings seat has opened on run's record: its events that open one, as
// the loaded window index reads them (record.WindowIndex.Opens) — a hook bracket naming the seat, or
// a register that neither joined its own bracket's sitting nor repairs one. 0 for a seat that never
// sat, and for a run with no record yet.
func SittingsOpened(t *testing.T, run record.Run, seat string) int {
	t.Helper()
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatalf("runtest: loading the record to count %s's sittings: %v", seat, err)
	}
	n := 0
	for _, e := range m.Events {
		if s, opens := m.At.Opens(e); opens && s == seat {
			n++
		}
	}
	return n
}
