package capture

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY FIELD OF THE RELAYED PLAN IS CLASSIFIED HERE, ONCE, BY WHAT STANDS BEHIND IT. The engine
// reads the record only through the chair's copy of the plan, so a field it branches on with
// nothing behind it is a field the chair can alter unseen.
//
// A GUARD, NOT A GENERATED TABLE, and the reason is that each field's comparison is different code:
// a head against a row's pin, a docket against the motions standing at a sitting's end, an exchange
// count against a fold. What can be generated is: the field list is record.Plan's own, walked by
// reflection (TestEveryRelayedPlanFieldIsClassified fails on a field with no row here and on a row
// naming no field), and the notes the audit prints render from this table. What the label claims is
// tested from both ends (TestThePlanKeyReadsOnlyComparedFields): the engine's no-progress valve is
// sensitive to no field outside `compared`, and altering a `compared` field in a faithful relay
// produces a departure naming it.
type relayClass string

const (
	// relayCompared: DispatchParityAudit holds the relayed value to the record.
	relayCompared relayClass = "compared"
	// relayBackstopped: not compared by field; a named gate stands on the branch the engine takes.
	relayBackstopped relayClass = "backstopped"
	// relayResidue: the engine acts on it and nothing holds it. The reason states the open case.
	relayResidue relayClass = "residue"
	// relayNotActedOn: the engine takes no branch on it.
	relayNotActedOn relayClass = "not acted on"
)

type relayField struct {
	class relayClass
	why   string
}

// relayedPlanFields is keyed by GO FIELD PATH off record.Plan — `Plan.Parties[].SeatID` and
// `Plan.StaleAreas[].SeatID` share a JSON name and nothing else.
var relayedPlanFields = map[string]relayField{
	"Plan.Head":                   {relayCompared, "each paired relay, against its dispatch rows' pin"},
	"Plan.Parties[].SeatID":       {relayCompared, "each paired relay, against its dispatch rows"},
	"Plan.Parties[].GapIDs":       {relayCompared, "each paired relay, against the party's dispatch row"},
	"Plan.Parties[].Occasions":    {relayCompared, "each paired relay, against the party's dispatch row"},
	"Plan.Docket":                 {relayCompared, "each paired relay, against the chair's dockets of that sitting standing at its end (record.PlanFacts.Docketed)"},
	"Plan.Exchanges[].GapID":      {relayCompared, "each paired relay, against the gaps its dispatch rows engage"},
	"Plan.Exchanges[].Count":      {relayCompared, "each paired relay, against the exchange fold at the chair sitting's end (record.PlanFacts.ExchangesOn)"},
	"Plan.Exchanges[].Unresolved": {relayCompared, "each paired relay, against the exchange fold at the chair sitting's end (record.PlanFacts.ExchangesOn)"},
	"Plan.PassPermitted":          {relayCompared, "false on each paired relay: the record permits a PASS only with nobody dispatched. On a plan with no parties it reaches the engine's log line and return value only"},
	"Plan.Ceiling":                {relayCompared, "false on each paired relay: the record is at its ceiling only with nobody dispatched. On a plan with no parties the engine states CEILING from it, and the bench's `outcome` derives the word from the record and refuses one it cannot derive"},
	"Plan.MaxEpochs":              {relayCompared, "every relayed plan, against the run's terms"},
	"Plan.Epoch":                  {relayNotActedOn, "type-checked by the engine and never read after: the chair reads it"},
	"Plan.FinalSitting":           {relayNotActedOn, "type-checked by the engine and never read after: the chair reads it, and the record refuses a motion filed at the final sitting whatever the relay says"},
	"Plan.Blockers[].Kind":        {relayCompared, "every relayed plan, at its chair sitting's end, for the kinds record.RelayBlockerKinds names; the other kinds read SQL as it stands now or move by the chair's own review, and the engine's valve keys on none of them"},
	"Plan.Blockers[].Subject":     {relayCompared, "with its kind"},
	"Plan.Blockers[].Owner":       {relayCompared, "with its kind"},
	"Plan.EpochLimitReached":      {relayBackstopped, "the engine stops on it before its valve reads the plan. Relayed true with parties, nobody is dispatched and every party of that dispatch FAILs as never having sat; with none, the bench's `outcome` derives CEILING from the record and refuses one it cannot derive. The departure names the parties, never this field"},
	"Plan.RemandOwed[].GapID":     {relayResidue, "names the gaps blue is told a remand's duty on; no dispatch row records it"},
	"Plan.RemandOwed[].Direction": {relayResidue, "the research direction blue is handed; no dispatch row records it"},
	"Plan.Why":                    {relayNotActedOn, "prose: it reaches the account of an UNVERIFIED end and the workflow's return value"},
	"Plan.StaleAreas[].SeatID":    {relayNotActedOn, "type-checked by the engine and never read after"},
	"Plan.StaleAreas[].Pin":       {relayNotActedOn, "type-checked by the engine and never read after"},
}

// planField is one field of record.Plan as the relay carries it: its Go path, which keys
// relayedPlanFields, and its JSON path, which is what a relay and a departure name.
type planField struct{ goPath, jsonPath string }

// planFields is every leaf of record.PlanFields — the record's one walk of record.Plan, which the
// engine script's plan schema is generated from as well — in declaration order, each under its
// path. A list of structs is its fields' paths, never a leaf of its own.
func planFields() []planField {
	var out []planField
	var walk func(fs []record.PlanField, goPath, jsonPath string)
	walk = func(fs []record.PlanField, goPath, jsonPath string) {
		for _, f := range fs {
			g, j := goPath+"."+f.Name, strings.TrimPrefix(jsonPath+"."+f.JSON, ".")
			if f.Of != nil {
				walk(f.Items, g+"[]", j+"[]")
				continue
			}
			out = append(out, planField{g, j})
		}
	}
	walk(record.PlanFields(), "Plan", "")
	return out
}

// comparedPlanFields is the JSON path of every field classified compared, in the plan's own order,
// with the blocker kinds a relay is held on: the one statement of what the audit compares that its
// detail prints.
func comparedPlanFields() string {
	var names []string
	for _, f := range planFields() {
		if relayedPlanFields[f.goPath].class == relayCompared {
			names = append(names, f.jsonPath)
		}
	}
	var kinds []string
	for _, k := range record.RelayBlockerKinds() {
		kinds = append(kinds, string(k))
	}
	return strings.Join(names, ", ") + " (blockers of kind " + strings.Join(kinds, ", ") + ")"
}
