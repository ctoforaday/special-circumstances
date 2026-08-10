package record

import "fmt"

// AVENUES HAVE A LIFECYCLE NOW, BECAUSE THE UNIT IS THE CHOICE, NOT THE ENTRY.
//
// MEASURED, over 86 avenue events across six runs: ZERO lines were ever recorded twice and
// ZERO statuses ever changed. There was no id, no key, no update path — `avenue` was a
// one-shot append. 83 of the 86 landed in round 0.
//
// That makes the corpus's headline number mean something other than it appeared to. 68
// "pursued" is not 68 directions pursued to completion; it is 68 INTENTIONS DECLARED BEFORE
// ANY RESEARCH HAPPENED. If a direction died in round 2 there was no mechanism to say so,
// and the 21% rejection rate measured only what blue could rule out before starting — not
// how hard it looked.
//
// The goal is that blue finds several plausible directions, picks the best, and is SEEN TO
// HAVE DONE SO IN EVIDENCE. A one-shot append records the plan; it cannot record the
// choosing. So an avenue gets what a gap has — an id, a status that moves with a stated
// reason, and an adjudicator.

// AvenueStatuses are the states an avenue may hold. `proposed` is the new one: a direction
// blue has put forward and not yet resolved, which is the state the old shape could not
// express at all (everything had to be declared already-pursued or already-dead).
// `deferred` is the fate that had no name: a direction worth taking, and not by THIS run.
// It is not `declined` (judged not worth it) and not `abandoned` (tried, died) — it is kept,
// and it is the carrier for bootstrapping a later run. Deliberately a PROPOSAL for a human
// to select rather than a seed: a run that queues its own successor is a loop with no human
// in it.
var AvenueStatuses = []string{"proposed", "pursued", "declined", "abandoned", "deferred"}

// AvenueRulings are red's fates for a proposed direction. Red AUDITS and RULES; it never
// proposes one — directing research is what a gap's required_fix already does, and a second
// spelling for it is the aliasing this vocabulary exists to prevent.
var AvenueRulings = []string{"endorsed", "out-of-scope", "too-thin"}

// MintAvenueID assigns the next run-unique avenue id (A1, A2 …).
//
// Run-unique rather than round-scoped, unlike a gap: an avenue OUTLIVES the round that
// proposed it — that is the whole point of giving it a lifecycle — so a round-scoped id
// would have to be re-minted to survive, which is the bug this replaces.
func MintAvenueID(runDir string) (string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	n := 0
	for _, e := range m.Events {
		if e.Type == "avenue" && e.Payload.Str("avenue_id") != "" && e.Payload.Str("supersedes_status") == "" {
			n++
		}
	}
	return fmt.Sprintf("A%d", n+1), nil
}

// Avenue is one direction's state after replay: its latest status, with the history that
// produced it. The history is kept because "chose to abandon this at round 2, having
// pursued it at round 0" is the evidence of choosing, and only the sequence carries it.
type Avenue struct {
	ID         string
	Line       string
	Hypothesis string
	Method     string
	Status     string
	Reason     string   // the reason attached to the CURRENT status
	Round      int      // the round the current status was set in
	History    []string // "r0 pursued", "r2 abandoned" …
	SeatID     string   // who last moved it — attribution the one-line row has always carried
	Ruling     string   // red's fate, if ruled
	RulingWhy  string
	RuledRound int
	// Contests is the ruling blue moved AGAINST, recorded by `blue avenue` at the moment of
	// the move. Read from the field rather than re-derived from (status, ruling): the write
	// path already decided what counts as contesting, and a second derivation downstream is a
	// second definition that can disagree with it.
	Contests string
}

// Avenues replays the avenue events into current state, in proposal order.
func Avenues(b *Board) []*Avenue {
	byID := map[string]*Avenue{}
	var order []string
	for _, e := range b.Events {
		id := e.Payload.Str("avenue_id")
		if id == "" {
			// A direction motion carries the avenue's id under `motion_id`, because to the
			// motion machinery it IS the motion's id — the proposal is the filing, so there is
			// no second identity to mint. Keying only on `avenue_id` dropped every ruling made
			// through the new verb before it reached the switch below.
			if e.Type == "motion-rule" && e.Payload.Str("subject") == "direction" {
				id = e.Payload.Str("motion_id")
			}
			if id == "" {
				continue
			}
		}
		switch e.Type {
		case "avenue":
			a, ok := byID[id]
			if !ok {
				a = &Avenue{ID: id}
				byID[id] = a
				order = append(order, id)
			}
			// A creation carries the substance; a MOVE carries only the new status and why,
			// so the substance must not be blanked by it.
			if v := e.Payload.Str("line"); v != "" {
				a.Line = v
			}
			if v := e.Payload.Str("hypothesis"); v != "" {
				a.Hypothesis = v
			}
			if v := e.Payload.Str("method"); v != "" {
				a.Method = v
			}
			a.Status, a.Reason, a.Round, a.SeatID = e.Payload.Str("status"), e.Payload.Str("reason"), e.Round, e.SeatID
			a.Contests = e.Payload.Str("contests_ruling")
			a.History = append(a.History, fmt.Sprintf("r%d %s", e.Round, a.Status))
		case "avenue-rule":
			// The PRE-#344 spelling. Permanent: a stored record written under it must still
			// render its rulings (record/compat.go).
			a, ok := byID[id]
			if !ok {
				continue
			}
			a.Ruling, a.RulingWhy, a.RuledRound = e.Payload.Str("ruling"), e.Payload.Str("reason"), e.Round
		case "motion-rule":
			// THE CURRENT SPELLING, and reading it here is not optional.
			//
			// A direction motion joins on the avenue's own id, so `motion direction rule` writes
			// a motion-rule whose motion_id IS an A-number. Until this arm existed, a ruling
			// made through the new verb never reached `--view lines-of-inquiry` — the projection
			// blue reads to decide whether to pursue, comply or drop. The line simply stayed
			// "Awaiting a decision", which is what an unruled line looks like, so red's ruling
			// was indistinguishable from red not having sat.
			if e.Payload.Str("subject") != "direction" {
				continue
			}
			a, ok := byID[id]
			if !ok {
				continue
			}
			a.Ruling, a.RulingWhy, a.RuledRound = e.Payload.Str("ruling"), e.Payload.Str("opinion"), e.Round
		}
	}
	out := make([]*Avenue, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// requireAvenue refuses a reference to an avenue no proposal created — the same discipline
// every other cross-reference gets (refs.go), for the same reason: a dangling reference is
// accepted at write time and dropped at replay.
func RequireAvenueRef(runDir, id string) error {
	if id == "" {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		if e.Type == "avenue" && e.Payload.Str("avenue_id") == id {
			return nil
		}
	}
	return fmt.Errorf("record: --id names avenue %s, which no avenue event proposed — a dangling reference is accepted here and dropped at replay", id)
}

// StaleAvenues returns the avenues still sitting at `proposed` or `pursued` — the ones blue
// owes a decision on. It is what makes the revisit duty checkable rather than hoped for: the
// measured failure was not that blue chose badly, it is that nothing ever asked it to choose
// again once the round-0 plan was written.
func StaleAvenues(b *Board) []*Avenue {
	var out []*Avenue
	for _, a := range Avenues(b) {
		if a.Status == "proposed" || a.Status == "pursued" {
			out = append(out, a)
		}
	}
	return out
}

// AvenueRuling returns red's most recent ruling on an avenue, or "" if it never ruled.
//
// The ruling and the avenue's fate were both on the record and joined NOWHERE, so blue
// pursuing a line red called out-of-scope looked exactly like blue pursuing one red endorsed.
// Red's ruling is an argument rather than a command — blue may pursue anyway — but the
// disagreement should be a fact, not something a reader reconstructs from two lists.
func AvenueRuling(runDir, avenueID string) string {
	b, err := BoardState(runDir)
	if err != nil {
		return ""
	}
	// BOTH SPELLINGS, most recent wins. The events are ordered by timestamp across shards, so
	// "last one seen" is the latest ruling regardless of which vocabulary wrote it.
	ruling := ""
	for _, e := range b.Events {
		switch e.Type {
		case "avenue-rule":
			if e.Payload.Str("avenue_id") == avenueID {
				ruling = e.Payload.Str("ruling")
			}
		case "motion-rule":
			if e.Payload.Str("subject") == "direction" && e.Payload.Str("motion_id") == avenueID {
				ruling = e.Payload.Str("ruling")
			}
		}
	}
	return ruling
}
