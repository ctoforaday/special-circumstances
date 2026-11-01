package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S ORDERS REACH EVERY SEAT FROM THE RECORD, ON THE WORK LIST (gblock, 2026-10-30).
//
// A granted petition binds the seats its ruling names, and the ruling carries the order
// (petitionorder.go). Every seat's work list carries every grant's relief in record order, each with
// whom it binds, so a seat reads the orders binding it on the list the SubagentStart hook hands it as
// it is dispatched — the same carrier the bench's declarations travel on. The list is every seat's
// because `both` binds every seat and a seat reading an order addressed to the other party knows what
// that party owes.

// ReliefJSON is one granted petition's order as the work list carries it.
type ReliefJSON struct {
	// Motion is the petition the order grants, and Petitioner the seat that filed it.
	Motion     string `json:"motion"`
	Petitioner string `json:"petitioner"`
	// Binds is whom the order binds: blue | red | both.
	Binds string `json:"binds"`
	// Relief is the order, as the bench recorded it on the ruling.
	Relief string `json:"relief"`
	// Epoch is the epoch the bench granted it in.
	Epoch int `json:"epoch"`
}

// reliefOf is every granted petition's order on evs, in the order the petitions were filed.
func reliefOf(evs []*Event, win WindowIndex) []ReliefJSON {
	out := []ReliefJSON{}
	for _, m := range MotionsOf(evs, win) {
		if m == nil || m.Subject != recordpb.Word(recordpb.MotionSubject_MOTION_SUBJECT_PETITION) ||
			m.Ruling != recordpb.Word(recordpb.PetitionRuling_PETITION_RULING_GRANTED) || m.RulingRelief == "" {
			continue
		}
		out = append(out, ReliefJSON{Motion: m.ID, Petitioner: m.Filer, Binds: m.RulingBinds, Relief: m.RulingRelief, Epoch: m.RulingEpoch})
	}
	return out
}
