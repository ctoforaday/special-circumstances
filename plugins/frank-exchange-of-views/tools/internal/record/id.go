package record

import (
	"crypto/rand"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// NewID mints the id of a new thing of a kind: the kind's letter, a hyphen and eight random hex.
//
// IT IS RANDOM SO THAT IT MUST BE LOOKED UP. A sequence can be written down without checking that
// it exists: on the 2026-07-18 run a chair read seven recorded acts and nine in prose and
// disposed of L6-F8..F16, which no event created. It is also minted with no count of what the
// record holds, so two seats minting at once cannot be handed one id. The columns an id is MINTED
// into are UNIQUE — mint.gap_id, motion.motion_id, cite.label,
// computation.computation_id, and anchor.id for every placed act — so a repeat there is a refused
// write. inquiry.inquiry_id is not: an inquiry writes one row per recorded fate under its id.
func NewID(kind string) string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return anchor.ID(kind, b)
}

// keyedAct is the event the keyed lookup q (seat, key -> event id) finds, with its body, nil for an
// empty key or none. The body is what a keyed retry is held to: the act it claims to repeat, as the
// record stores it.
func keyedAct(run Run, q, seatID, key string) (*recordpb.Event, error) {
	if key == "" {
		return nil, nil
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	var id int64
	if found, err := queryRowAt(db, []any{&id}, q, seatID, key); err != nil || !found {
		return nil, err
	}
	ev, err := recordsql.EventAt(db, id)
	if err == nil && ev == nil {
		err = fmt.Errorf("event %d holds the act keyed %q and was not there to read", id, key)
	}
	return ev, err
}
