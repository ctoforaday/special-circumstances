package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHAT THE RUN WAS ACTUALLY ANSWERED BY, read back.
//
// `register` writes served_model (and requested_model, when the harness declared a swap) as
// fields on its event, because that is the first moment a seat's own trajectory names the model
// replying to it. This is the read half — the surface `capture` grades and a human reads.
//
// THE THREE STATES ARE KEPT APART, and collapsing them is the defect this whole thread is about:
// a seat that was served the configured tier, a seat that was served something else, and a seat
// whose serving model NOBODY MEASURED. The third is not the first. A run reported as matching its
// configuration when in fact nothing looked is exactly how ~$379 went to a tier that never ran.

// SeatModel is one seat's dispatch as the record has it.
type SeatModel struct {
	SeatID string `json:"seatId"`
	// Class is the tier class the engine dispatched this seat on (bulk|judgment), or "" for a
	// seat that rides no tier.
	Class string `json:"class"`
	// Served is the model that answered, or "" for NOT MEASURED.
	Served string `json:"served,omitempty"`
	// Requested is set only where the harness DECLARED a substitution, naming both ends.
	Requested string `json:"requested,omitempty"`
}

// Measured reports whether anything actually observed this seat's serving model.
func (s SeatModel) Measured() bool { return s.Served != "" }

// Substituted reports a substitution the harness itself declared.
func (s SeatModel) Substituted() bool { return s.Requested != "" && s.Requested != s.Served }

// SeatModels returns one row per seat that registered, in first-register order.
//
// THE LAST REGISTER WINS, matching SeatOfAgent: a re-dispatched seat writes a fresh register, and
// the latest one is the sitting that actually ran. A seat re-dispatched into a substituted
// environment must not be masked by its first, clean dispatch.
func SeatModels(b *Board) []SeatModel {
	order := []string{}
	bySeat := map[string]*SeatModel{}
	for _, e := range b.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			continue
		}
		seat := e.GetSeatId()
		row := bySeat[seat]
		if row == nil {
			row = &SeatModel{SeatID: seat, Class: TierClassOfSeat(seat)}
			bySeat[seat] = row
			order = append(order, seat)
		}
		r := e.GetRegister()
		row.Served, row.Requested = r.GetServedModel(), r.GetRequestedModel()
	}
	out := make([]SeatModel, 0, len(order))
	for _, s := range order {
		out = append(out, *bySeat[s])
	}
	return out
}
