package report

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

// benchOccasion is the occasion a fixture registers a seat with: the bench owes one and no other
// seat may pass one, so this asks the write path's own question rather than keeping a list of which
// ids are the bench. `assemble` is the sitting these fixtures stand in for — the bench's acts in
// them are the outcome and what follows it, which the write takes only in the assembly sitting.
func benchOccasion(seatID string) string {
	if record.SeatOwesOccasion(seatID) {
		return "assemble"
	}
	return ""
}
