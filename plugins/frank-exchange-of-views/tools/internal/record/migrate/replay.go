package migrate

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Options is the operator's say over a replay.
type Options struct {
	// AcceptLoss names the words whose documented untranslatability the operator accepts.
	// Acceptance is per word: it silences exactly that LossError and nothing else, and the
	// manifest records what was accepted.
	AcceptLoss map[string]bool
}

// Refusal is one event the replay could not land, with the words of whichever layer refused
// it — the registry, or the current write path's own validation. Refusals are the tool's
// OUTPUT: an old record may hold acts that were legal under the rules of their day, and
// each refusal class gets decided in review. There is no flag that skips validation.
type Refusal struct {
	OldID int64  `json:"old_id"`
	Word  string `json:"word"`
	Err   string `json:"error"`
}

// Result is what one replay did, counted where the manifest needs counts.
type Result struct {
	In             map[string]int // old events per word, as read
	Out            map[string]int // new events per word, as written
	Refusals       []Refusal
	AcceptedLosses map[string]string // word -> the entry's stated reason
	SourceHash     string            // sha256 over the event stream AS READ — the logical record
	GapIDs         map[string]string // archived gap id -> the id the migrated record carries (§III.A.5)
	Labels         map[string]string // archived finding label -> the label the migrated record carries
	Serialized     map[string]int    // archived instance seat -> events moved after instance 1 (serializeInstances)
}

// Replay re-drives an old record's events, in order, through the current write path into
// dst. Each act is validated by the same code every seat verb runs, and each event's stamp
// is the ORIGINAL stamp: record.Now is pinned per event, which is safe exactly because the
// migrator is its own process and nothing else writes dst.
//
// Register events go through Append like everything else, NOT through RegisterSeat — a
// deliberate deviation from the plan's first draft, with the reason on the record:
// RegisterSeat stamps ENGINE-OBSERVED facts from its own environment (tool_version,
// agent_type, run_via), and the migrator's environment is not the run's. The old register
// body already carries what was actually observed then, and replaying it verbatim is the
// only honest answer. Dispatch numbering survives because it derives from the key ordinal,
// which Append's own path re-derives.
func Replay(src Source, reg Registry, dst record.Run, opt Options) (*Result, error) {
	evs, err := src.Events()
	if err != nil {
		return nil, err
	}
	res := &Result{
		In:             map[string]int{},
		Out:            map[string]int{},
		AcceptedLosses: map[string]string{},
	}
	res.SourceHash, err = hashEvents(evs)
	if err != nil {
		return nil, err
	}

	was := record.Now
	defer func() { record.Now = was }()
	record.Migrating = true
	defer func() { record.Migrating = false }()

	evs, res.Serialized = serializeInstances(evs)
	rm := newRemap()
	// THE CAST IS SYNTHESIZED FIRST (plans/roundless.md §III.A.5): an archived run had no cast
	// event, and the live write path checks every register and dispatch against one. Its cast is
	// the seats that registered, translated — a fact the record already holds, restated as the
	// list the current schema reads it from.
	if len(evs) > 0 {
		// Under the archive's own clock: the cast precedes the first event and is stamped with its
		// time, so ORDER BY id and the timestamps tell one story.
		if ts, err := record.ParseStamp(evs[0].TS); err == nil {
			record.Now = func() time.Time { return ts }
		}
	}
	if wrote, err := writeSynthesizedCast(evs, rm, dst); err != nil {
		return nil, err
	} else if wrote {
		res.Out["cast"]++
	}
	// A CORRECTION IS REPLAYED AS THE PAIR IT WAS WRITTEN AS (plans/same-sitting-correction.md,
	// S10). On the source record the replacement sits immediately before its Correction — one
	// transaction wrote both — and the write path refuses either alone: a replacement re-driven as
	// an ordinary act lands on its target's key, and a Correction appended on its own is refused.
	// So the replacement is re-driven AS A CORRECTION of its target's MIGRATED key, which writes
	// both, and the source Correction is then passed over. Who and when are re-checked against the
	// replayed record in the original order, which is the order they held in.
	correctionWord := recordpb.Word(recordpb.EventType_EVENT_TYPE_CORRECTION)
	fixes := map[string]OldEvent{} // source replacement key -> the correction that names it
	for _, old := range evs {
		if old.Word == correctionWord {
			if r, _ := old.Fields["replacement"].(string); r != "" {
				fixes[r] = old
			}
		}
	}
	newKey := map[string]string{} // source key -> the key the migrated event carries
	paired := map[int64]bool{}    // source corrections already written beside their replacement
	for _, old := range evs {
		res.In[old.Word]++
		if old.Word == correctionWord {
			if !paired[old.ID] {
				// LOUD, NOT DROPPED: a correction whose replacement did not land would leave a
				// replacement nothing marks as one, or a strike with nothing in its place.
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word,
					Err: fmt.Sprintf("migrate: this correction names replacement %v, which did not land as a correction on the migrated record — it is absent from the source or was itself refused", old.Fields["replacement"])})
			}
			continue
		}
		seatID, err := rm.seat(old.SeatID)
		if err != nil {
			res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		bodies, err := reg.Translate(old, dst)
		if err != nil {
			var loss LossError
			if as(err, &loss) && opt.AcceptLoss[loss.Word] {
				res.AcceptedLosses[loss.Word] = loss.Reason
				continue
			}
			res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		ts, err := record.ParseStamp(old.TS)
		if err != nil {
			res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word,
				Err: fmt.Sprintf("unparseable ts %q: %v", old.TS, err)})
			continue
		}
		record.Now = func() time.Time { return ts }
		id := record.Identity{Run: dst, SeatID: seatID}
		fix, isReplacement := fixes[old.Key]
		if isReplacement && old.Key != "" {
			corrects, _ := fix.Fields["corrects"].(string)
			target, ok := newKey[corrects]
			if !ok || len(bodies) != 1 {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word,
					Err: fmt.Sprintf("migrate: this act replaced %s by a same-sitting correction, and that act did not land on the migrated record", corrects)})
				continue
			}
			typ, err := recordpb.SetBody(&recordpb.Event{}, bodies[0])
			if err != nil {
				return nil, err
			}
			why, _ := fix.Fields["why"].(string)
			id.Correct = &record.Correct{Type: typ, Key: target, Why: why}
		}
		for _, body := range bodies {
			// NO EPOCH IS CARRIED. The old row's epoch was recovered from its seat id by regex at
			// the time; the re-driven write computes the EPOCH from the chair registers already
			// replayed (events_w."epoch"), which is the same fact from the record rather than from
			// the name — plans/roundless.md §III.A.2. A migrated record's round therefore means
			// what a fresh record's does.
			rm.apply(body, seatID)
			ev, err := record.Append(id, body)
			if err != nil {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
				continue
			}
			res.Out[wordOf(ev)]++
			if _, seen := newKey[old.Key]; old.Key != "" && !seen {
				newKey[old.Key] = ev.GetKey()
			}
			if id.Correct != nil {
				paired[fix.ID] = true
				res.Out[correctionWord]++
			}
		}
	}
	res.GapIDs, res.Labels = rm.gaps, rm.labels
	return res, nil
}

// wordOf spells a written event's type the way the record does.
func wordOf(ev *record.Event) string { return recordpb.Word(ev.GetType()) }

// as is errors.As without the reflection footgun for the one concrete type this uses.
func as(err error, target *LossError) bool {
	l, ok := err.(LossError)
	if ok {
		*target = l
	}
	return ok
}

// hashEvents is the LOGICAL source hash: sha256 over a canonical text form of the event
// stream as read. File-byte hashes cannot be it — the same record hashes differently across
// WAL checkpoint states — so the hash covers what the replay actually consumed.
func hashEvents(evs []OldEvent) (string, error) {
	h := newStreamHasher()
	for _, ev := range evs {
		h.event(ev)
	}
	return h.sum(), nil
}

// sortedKeys keeps the canonical form canonical: map walks are randomized, hashes are not.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// writeSynthesizedCast writes the Cast an archived run never had, under the harness, from its
// register events after seat translation. Petition sittings are admitted by shape and left out.
//
// A record that already carries a cast is not an archived run: its own cast replays in order
// like every other event. Synthesizing one as well made the original a SECOND cast, which the
// write path refuses — so migrating any roundless-era record exited non-zero on a refusal of
// its own membership.
func writeSynthesizedCast(evs []OldEvent, rm *remap, dst record.Run) (bool, error) {
	for _, old := range evs {
		if old.Word == "cast" {
			return false, nil
		}
	}
	seen := map[string]bool{}
	var cast []string
	for _, old := range evs {
		if old.Word != "register" {
			continue
		}
		s, err := rm.seat(old.SeatID)
		if err != nil {
			continue // the refusal lands on the event itself, in the loop
		}
		if strings.HasPrefix(s, "judge-petition-") || seen[s] {
			continue
		}
		seen[s] = true
		cast = append(cast, s)
	}
	if len(cast) == 0 {
		return false, nil // nothing registered: nothing to admit, and nothing will register
	}
	sort.Strings(cast)
	_, err := record.Append(record.Identity{Run: dst, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast})
	return err == nil, err
}
