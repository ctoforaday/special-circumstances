package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/servedmodel"
)

// WHAT THE RUN WAS ACTUALLY ANSWERED BY, read back.
//
// THE TRAJECTORY IS THE STATEMENT OF RECORD, and this reads it. The event opening a seat's sitting
// carries its AGENT ID, the agent's stop names its trajectory (TranscriptOf), and the model is
// resolved from that here, when a reader asks.
//
// It used to be a field written at register, and it was empty every time: register is the seat's
// FIRST act, a fresh trajectory opens with `user` turns, and the model appears on the first
// ASSISTANT turn. 45 registers in one run, 43 with an agent id, ZERO with a model. A field that
// has never once carried a value is not a measurement, and reading it back gave the run's tier
// report nothing to say while looking like it had looked.
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
	// NotMeasured says WHY nothing answered, and is set exactly when Served is empty. A bare ""
	// is the shape that let a run with no measurement read like a run that matched: absent and
	// agreeing are the same bytes unless the absence carries its reason.
	NotMeasured string `json:"notMeasured,omitempty"`
}

// Measured reports whether anything actually observed this seat's serving model.
func (s SeatModel) Measured() bool { return s.Served != "" }

// Substituted reports a substitution the harness itself declared.
func (s SeatModel) Substituted() bool { return s.Requested != "" && s.Requested != s.Served }

// SeatModels returns one row per seat that sat, in the order each first opened a sitting.
//
// A SITTING IS THE STORED ONE (f.At): the hook's bracket opens it where the configuration seats one
// seat, and a register opens it where no bracket did. The agent is the one on the event that opened
// it — the bracket's, or the register's — so a seat bound and seated by its bracket alone has a row.
//
// THE LATEST SITTING WINS: a re-dispatched seat opens a fresh sitting, and its opening names the
// agent of that sitting. A seat re-dispatched into a substituted environment must not be masked by
// its first, clean dispatch.
// IT READS NO CONFIGURATION: what answered a seat does not depend on what was configured, and a
// report that compares the rows with the configured tiers reads the run's configuration once, so
// the rows and the report's own statement of the configured tiers are one read, never two that can
// disagree.
func SeatModels(run Run, f Family) []SeatModel {
	order := []string{}
	bySeat := map[string]*SeatModel{}
	classErr := map[string]error{}
	var onRecord map[string]string // read once, on the first seat whose trajectory is not found
	for _, e := range f.Events {
		seat, opens := f.At.Opens(e)
		if !opens {
			continue
		}
		row := bySeat[seat]
		if row == nil {
			class, cerr := TierClassOfSeat(run, seat)
			row = &SeatModel{SeatID: seat, Class: class}
			bySeat[seat] = row
			order = append(order, seat)
			if cerr != nil {
				classErr[seat] = cerr
			}
		}
		// A TIER THAT COULD NOT BE ESTABLISHED IS NOT MEASURED, not "rides no tier": the
		// substitution compare needs the tier, and "" would hold the seat to none.
		if cerr := classErr[seat]; cerr != nil {
			row.Served, row.Requested, row.NotMeasured = "", "", "the tier this seat rides could not be established: "+cerr.Error()
			continue
		}
		agent := recordpb.AgentOpening(e)
		if agent == "" {
			row.Served, row.Requested, row.NotMeasured = "", "",
				"the event that opened this seat's sitting carries no agent id, so the seat's trajectory cannot be named"
			continue
		}
		// THE TRAJECTORY IS THE ONE THE AGENT'S STOP NAMES. A seat still sitting has no stop, so
		// the record names no trajectory for it yet: the bench assembling inside its own sitting
		// reads its own row NOT MEASURED, saying so, and capture re-assembles the documents after
		// every stop is on the record.
		path, err := TranscriptOf(f, agent)
		var obs servedmodel.Observation
		if err == nil {
			obs, err = servedmodel.Observe(path)
		}
		if err != nil {
			// THE RECORD'S OWN TURNS ANSWER WHERE THE TRAJECTORY CANNOT BE OPENED. The stop ingests
			// the seat's transcript into seat_turn, each row naming the model that served it, and a
			// reader on another box — the record copied off it, an archived run — has no trajectory
			// at the path the stop names. Smoke 07d reported 13 seats NOT MEASURED while all 647 of
			// their turn rows named claude-haiku-5-5. A declared substitution is the trajectory's
			// alone, so this answers the model and not the declaration.
			if onRecord == nil {
				onRecord = servedOnRecord(run)
			}
			m, ok := onRecord[agent]
			if !ok {
				row.Served, row.Requested, row.NotMeasured = "", "", err.Error()+"; and the record holds no turn of this agent"
				continue
			}
			obs = servedmodel.Observation{Served: m}
		}
		row.Served, row.Requested, row.NotMeasured = obs.Served, obs.Requested, ""
	}
	out := make([]SeatModel, 0, len(order))
	for _, s := range order {
		out = append(out, *bySeat[s])
	}
	return out
}

// servedOnRecord is the model each agent's FIRST recorded turn names, from seat_turn — the turn
// servedmodel.Read takes from a trajectory. A run with no seat_turn table, or a record that cannot
// be read, answers nothing, and each seat then states its trajectory's reason.
func servedOnRecord(run Run) map[string]string {
	out := map[string]string{}
	if run.Dir() == "" {
		return out
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return out
	}
	// One walk of the primary key in agent order. SQLite takes a bare column beside a lone min()
	// from the row that holds the minimum, so model is the agent's first served model.
	rows, err := db.Query(`SELECT "agent_id", "model", min("turn_idx") FROM "seat_turn"
	  WHERE "model" <> '' GROUP BY "agent_id"`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var agent, model string
		var first int64
		if rows.Scan(&agent, &model, &first) == nil {
			out[agent] = model
		}
	}
	return out
}
