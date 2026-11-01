package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A GRANTED PETITION'S ORDER IS ON THE RULING (gblock, 2026-10-30).
//
// The bench's relief travelled only on its sitting's envelope, which the engine copied into three
// prompts and defaulted to binding both parties where the bench named nobody — an order nothing on
// the record carried, which an engine restart dropped, and an addressee the tool made up. The order
// is the ruling's now: a grant states its relief and whom it binds, a denial orders nothing, and every
// seat's work list carries each grant's relief from the record (WorkJSON.Relief).
//
// AN ARCHIVED GRANT STATES NO RELIEF: the one on record (quadratic-formula M5) put its order in the
// opinion. Migration replays it as the bench recorded it, so the grant's requirement is off under
// Migrating, and the migrated run's work lists carry no relief for it.

// requirePetitionOrder refuses a petition ruling whose order does not match its word, and an order on
// any other subject's ruling.
func requirePetitionOrder(r *recordpb.MotionRule) error {
	ordered := r.Relief != nil || r.Binds != nil
	p, petition := r.GetRuling().(*recordpb.MotionRule_Petition)
	switch {
	case !petition && ordered:
		return fmt.Errorf("--relief and --binds are a granted petition's order, and a %s ruling orders nothing beyond its ruling", recordpb.Word(r.Subject()))
	case !petition:
		return nil
	case p.Petition == recordpb.PetitionRuling_PETITION_RULING_DENIED && ordered:
		return fmt.Errorf("--as denied orders nothing, so it takes neither --relief nor --binds: a denial leaves the seats where they were, and an order beside it would bind them to what you refused")
	case p.Petition == recordpb.PetitionRuling_PETITION_RULING_GRANTED && !Migrating &&
		(strings.TrimSpace(r.GetRelief()) == "" || r.GetBinds() == recordpb.RulingBinds_RULING_BINDS_UNSPECIFIED):
		return fmt.Errorf("--as granted requires --relief and --binds: --relief is the order the seats it binds must follow, written as an instruction they can act on without reading your opinion, and --binds names whom it binds (%s). Every seat's work list carries it from the ruling, so a grant without them orders nothing",
			strings.Join(Names(MotionFields["petition"]["binds"]), " | "))
	}
	return nil
}
