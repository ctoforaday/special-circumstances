package report

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

// benchOccasion is the occasion a fixture registers a seat with: the bench owes one and no other
// seat may pass one, so this asks the write path's own question rather than keeping a list of which
// ids are the bench. `docket` is the sitting these fixtures stand in for — the chair dispatched the
// bench onto a gap.
func benchOccasion(seatID string) string {
	if record.SeatOwesOccasion(seatID) {
		return "docket"
	}
	return ""
}
