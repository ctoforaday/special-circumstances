package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// A HELD DELIVERY EXITS hookgate.ExitDeliveryHeld, and only from the next phase: the hook reads that
// status as "another call is handing the item". Every other failure, and a held lock on any other
// phase, exits 1.
func TestExitStatusMarksAHeldDeliveryOnTheNextPhase(t *testing.T) {
	held := fmt.Errorf("%w (x.lock) past 5s", hookgate.ErrDeliveryHeld)
	for _, c := range []struct {
		phase sittingwrite.Phase
		err   error
		want  int
	}{
		{sittingwrite.Next, held, hookgate.ExitDeliveryHeld},
		{sittingwrite.Next, errors.New("the record is captured"), 1},
		{sittingwrite.Open, held, 1},
	} {
		if got := exitStatus(c.phase, c.err); got != c.want {
			t.Errorf("exitStatus(%s, %v) = %d, want %d", c.phase, c.err, got, c.want)
		}
	}
}
