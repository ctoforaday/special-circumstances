package record

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// TestConcurrentSeatsRace is `go test -race` over the real write and replay paths: six lens
// seats appending to the one record while each also replays the board, which is the live shape.
func TestConcurrentSeatsRace(t *testing.T) {
	runDir := newRun(t)
	const seats = 6
	const perSeat = 4

	// SIX REAL AREAS, because a lens seat is now named for what it audits and the roster refuses
	// an id no dispatch could produce. Six concurrent writers is what this test is about; the
	// areas are simply the six the engine has.
	areas := []string{"evidence", "logic", "dark-side", "voice", "computation", "adversary"}

	var wg sync.WaitGroup
	errs := make(chan error, seats*perSeat*3)
	for s := 1; s <= seats; s++ {
		wg.Add(1)
		go func(s int) {
			defer wg.Done()
			area := areas[s-1]
			seatID := "red-lens-" + area
			if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: seatID}, "", ""); err != nil {
				errs <- err
				return
			}
			for i := 0; i < perSeat; i++ {
				f := &recordpb.Finding{
					Label:      proto.String(fmt.Sprintf("%s-F%d", area, i)),
					Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
					Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
					Impact:     recordtest.P(recordpb.Grade_GRADE_HIGH),
					Text:       proto.String(strings.Repeat("finding prose ", 20)),
				}
				if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seatID}, f); err != nil {
					errs <- err
					continue
				}
				// A concurrent READER (the replay every projection now runs on demand)
				// racing the appenders: no write may be lost to a racing read.
				if _, err := FamilyOf(mustRun(t, runDir)); err != nil {
					errs <- err
				}
			}
		}(s)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent seat: %v", err)
	}

	// EVERY EVENT LANDED. This asked whether a shard lost a write to a racing render; it now asks
	// whether a TRANSACTION did, which is the same question against the mechanism that replaced
	// the one the test was written for — and a stronger one, because the writers share a file
	// rather than each owning their own.
	m, err := MergedEvents(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	findings := 0
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_FINDING {
			findings++
		}
	}
	if want := seats * perSeat; findings != want {
		t.Errorf("findings landed = %d, want %d", findings, want)
	}
}
