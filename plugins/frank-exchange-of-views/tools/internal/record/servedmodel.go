package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/servedmodel"
)

// WHAT THE RUN WAS ACTUALLY ANSWERED BY, read back.
//
// THE TRAJECTORY IS THE STATEMENT OF RECORD, and this reads it. `register` records the seat's
// AGENT ID, which names the trajectory; the model is resolved from that here, when a reader asks.
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
	// Substitution, when the tier served differs from the tier this seat's class was configured
	// for. Detected on READ because that is when the serving model can be known at all.
	Substitution string `json:"substitution,omitempty"`
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
// SUBSTITUTION IS DETECTED HERE, NOT AT REGISTER. `run` supplies the configured tier per class;
// pass an empty Run to skip the comparison and report only what answered.
func SeatModels(run Run, f Family) []SeatModel {
	model, judgmentModel := "", ""
	if run.Dir() != "" {
		model, judgmentModel = modeltier.Config(run.Dir())
	}
	order := []string{}
	bySeat := map[string]*SeatModel{}
	for _, e := range f.Events {
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
		agent := r.GetAgentId()
		if agent == "" {
			row.Served, row.Requested, row.NotMeasured = "", "",
				"this register carries no agent id, so the seat's trajectory cannot be named"
			continue
		}
		obs, err := servedmodel.Observe(agent)
		if err != nil {
			row.Served, row.Requested, row.NotMeasured = "", "", err.Error()
			continue
		}
		row.Served, row.Requested, row.NotMeasured = obs.Served, obs.Requested, ""
		configured := model
		if row.Class == "judgment" {
			configured = judgmentModel
		}
		row.Substitution = TierSubstitution(run, configured, obs)
	}
	out := make([]SeatModel, 0, len(order))
	for _, s := range order {
		out = append(out, *bySeat[s])
	}
	return out
}

// TierSubstitution reports a seat answered by a tier other than the one its class was configured
// for, or "" when there is nothing to report.
//
// # Why it reports in BOTH directions
//
// The post-hoc audit graded a dearer-than-configured seat FAIL ("the fable trap") and a cheaper
// one WARN, and capture exited 2 only on a FAIL. The measured incident was CHEAPER —
// `claude-fable-5` configured, `claude-opus-4-8` served — so the one check that noticed it graded
// it a warning and capture exited 0. The framing was wrong: the question a research debate asks of
// its tier is not "did this cost more than budgeted" but "did the adversary run at the strength the
// run claims", and a weaker adversary silently substituted is the worse failure of the two.
//
// # Why it reports and no longer refuses
//
// It was a gate at `register`, which is the one moment the answer cannot be known: a fresh
// trajectory has no assistant turn yet, so the observation it judged was empty on all 45 registers
// of a full run. Nine tests passed over it the whole time, because every one supplied the
// observation the gate never received. A refusal that cannot fire is worse than none — it reads as
// enforcement to anyone auditing the surface — so the comparison moved to the read path, where the
// trajectory has an answer, and it states rather than blocks.
//
// WHAT THAT COSTS IS REAL AND IS NOT MADE UP HERE: nothing now stops a substituted sitting while it
// runs. Nothing did before either, but the honest way to get it back is a check on a seat's SECOND
// act — the first moment a model has answered — and that is a change to when every seat is measured
// rather than to where one number is read.
//
// # Why the override is not a flag
//
// An operator who knows the environment will not serve the configured tier may still want the run.
// That decision is theirs and is made ONCE, at setup, written to run-config as a field. It is
// deliberately not a flag on a seat's verb: a seat typing --allow-substitution would be the party
// under test excusing itself.
func TierSubstitution(run Run, configured string, o servedmodel.Observation) string {
	if o.Served == "" || configured == "" {
		return "" // NOT MEASURED, or a run that declared no tier for this class
	}
	want, got := modeltier.Of(configured), modeltier.Of(o.Served)
	if want == got && !o.Substituted() {
		return ""
	}
	declared := ""
	if o.Substituted() {
		declared = " The harness declared the substitution itself: " + o.Requested + " -> " + o.Served + "."
	}
	msg := "configured " + configured + " (" + want + " tier), served " + o.Served + " (" + got + " tier)." + declared
	if run.Dir() != "" && allowSubstitution(run) {
		return "ALLOWED BY THIS RUN'S CONFIG — " + msg
	}
	return msg
}
