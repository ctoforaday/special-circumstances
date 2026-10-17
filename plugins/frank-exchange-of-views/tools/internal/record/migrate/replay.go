package migrate

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
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
	IDs            map[string]string // archived id or finding label -> the id the migrated record carries
	Serialized     map[string]int    // archived instance seat -> events moved after instance 1 (serializeInstances)
	StatedFills    []StatedFill      // values supplied where the source predates the field
	GapAnchors     GapPlacement      // what bringing the run's gaps onto anchors did (F-c)
	ReportOrder    string            // what became of the order of the report's text (order.go)

	base string // the report's base as the source wrote it
}

// byGradeFillWhy is the reason a class-material fill records, in one place so the staged-registry
// fill and the class_new fill say the same thing.
const byGradeFillWhy = "the source predates material_default and this binary's shipped class table does not hold the slug; by_grade keeps the class's gaps as material as their grade alone made them"

// shippedDefault is the material default a class whose record predates the field takes: the
// binary's shipped value for the slug, else by_grade — and fill reports that it was the latter.
func shippedDefault(slug string) (v recordpb.ClassMaterial, fill bool) {
	if v, ok := record.ShippedMaterialDefaults[slug]; ok {
		return v, false
	}
	return recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, true
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
//
// The whole replay is one batch (record.Batch): every act is written in one transaction, each inside
// a savepoint, and committed once at the end.
func Replay(src Source, reg Registry, dst record.Run, opt Options) (*Result, error) {
	was := record.Now
	defer func() { record.Now = was }()
	record.Migrating = true
	defer func() { record.Migrating = false }()
	batch, err := record.BeginBatch(dst)
	if err != nil {
		return nil, err
	}
	defer func() { _ = batch.Rollback() }()
	res, err := replay(src, reg, batch, dst, opt, false)
	if err != nil {
		return nil, err
	}
	if err := batch.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

// replay writes src's events through batch, which its caller began on dst, holds and commits, with
// record.Migrating set; it pins record.Now per event and restores it on return, so each pass — the
// reorder trial and the replay as written after it — starts from the clock its caller had. With
// inOrder set, the report's base is written in the order its text ships in (order.go).
//
// dst IS AN ADDRESS, NEVER A HANDLE. While the batch is open, the batch's transaction holds the
// record's one connection, and anything that takes dst and opens the record waits on it forever. So
// dst is used here only as the Run of the Identity each act is written under; every read of the
// destination goes through batch.Querier().
//
// An act the write refuses is a Refusal, and the replay goes on. A failure of the batch itself
// (record.ErrBatchFailed) is asked of the batch once, at the end — whichever append it landed on —
// and returned: nothing the batch wrote can be kept.
func replay(src Source, reg Registry, batch *record.Batch, dst record.Run, opt Options, inOrder bool) (*Result, error) {
	was := record.Now
	defer func() { record.Now = was }()
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

	evs, res.Serialized = serializeInstances(evs)
	rm := newRemap(res.SourceHash)
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
	if wrote, err := writeSynthesizedCast(evs, rm, batch, dst); err != nil {
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
	// THE ID TABLE IS FILLED BEFORE ANYTHING IS WRITTEN. A seat's prose names ids minted after it —
	// a lane's inquiry argues against one a sibling lane proposed later in the stream — so the
	// stream is translated once to learn every id it mints, and again, below, to write it. An id is
	// a function of the source alone, so both passes spell it the same.
	for _, old := range evs {
		bodies, _ := reg.Translate(old)
		for _, body := range bodies {
			rm.apply(body, old.Fields["label"])
		}
	}
	place := newPlacer(batch.Querier(), evs, rm.ids)
	newKey := map[string]string{} // source key -> the key the migrated event carries
	// THE EVENTS.ID A DISPATCH PINS IS TRANSLATED AS AN ID IS. The cast written first, an anchor
	// placed after an act and a source event that writes nothing each move every later row, so a
	// pin carried as written names whatever the migrated record holds at that number. A pin names
	// the report head its party audits, which the source wrote before the dispatch; the migrated
	// pin is the id of the event that source head was written as.
	newID := map[int64]int64{} // source events.id -> the events.id of the first event written from it
	var wrote int64
	paired := map[int64]bool{}    // source corrections already written beside their replacement
	for _, old := range evs {
		res.In[old.Word]++
		if old.Word == correctionWord {
			if !paired[old.ID] {
				// LOUD, NOT DROPPED: a correction whose replacement did not land would leave a
				// replacement nothing marks as one, or a strike with nothing in its place.
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word,
					Err: fmt.Sprintf("this correction names replacement %v, which did not land as a correction on the migrated record — it is absent from the source or was itself refused", old.Fields["replacement"])})
			}
			continue
		}
		seatID, err := rm.seat(old.SeatID)
		if err != nil {
			res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		bodies, err := reg.Translate(old)
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
					Err: fmt.Sprintf("this act replaced %s by a same-sitting correction, and that act did not land on the migrated record", corrects)})
				continue
			}
			typ, err := recordpb.SetBody(&recordpb.Event{}, bodies[0])
			if err != nil {
				return nil, err
			}
			why, _ := fix.Fields["why"].(string)
			id.Correct = &record.Correct{Type: typ, Key: target, Why: rm.respell("Correction.why", why)}
		}
		for _, body := range bodies {
			// NO EPOCH IS CARRIED. The archived row's epoch was recovered from its seat id by regex
			// when it was written; the re-driven row's EPOCH comes from the chair sittings already
			// replayed (events_w."epoch"), which is the same fact from the record rather than from
			// the name — plans/roundless.md §III.A.2. A migrated record's round therefore means
			// what a fresh record's does.
			rm.apply(body, old.Fields["label"])
			if b, ok := body.(*recordpb.BaseIngest); ok {
				res.base = b.GetText()
				if inOrder {
					b.Text = proto.String(report.Ordered(res.base))
				}
			}
			// A BENCH ROW THAT PREDATES ITS OCCASIONS convened the bench for its docket: before
			// epoch 19 the chair readied the bench only on a docketed gap, and a petition was heard
			// off an envelope with no row. The write refuses a bench row naming no occasion, so the
			// row says what it always meant; one with no gap was never written then, and the write
			// refuses it rather than this guessing.
			if d, ok := body.(*recordpb.Dispatch); ok && len(d.GetOccasions()) == 0 && record.SeatOwesOccasion(d.GetSeatId()) {
				d.Occasions = []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}
			}
			if d, ok := body.(*recordpb.Dispatch); ok && d.Pin != nil {
				to, landed := newID[d.GetPin()]
				if !landed {
					res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word,
						Err: fmt.Sprintf("this dispatch pins event %d, which wrote no event on the migrated record before it — the pin would name whatever event holds that number", d.GetPin())})
					continue
				}
				d.Pin = proto.Int64(to)
			}
			// A COINING THAT PREDATES THE MATERIAL DEFAULT takes the shipped value for its slug,
			// else by_grade as a stated fill. A value the source carries is kept as recorded — a
			// run written by this binary migrates without any of its coinings changing.
			if cn, ok := body.(*recordpb.ClassNew); ok && cn.MaterialDefault == nil {
				v, fill := shippedDefault(cn.GetSlug())
				cn.MaterialDefault = v.Enum()
				if fill {
					res.StatedFills = append(res.StatedFills, StatedFill{
						Where: fmt.Sprintf("class_new event %d", old.ID), Slug: cn.GetSlug(),
						Value: recordpb.Word(v), Why: byGradeFillWhy})
				}
			}
			anchorAfter, err := place.step(body, id.Correct != nil)
			if err != nil {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
				continue
			}
			ev, err := batch.Append(id, body)
			if err != nil {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
				continue
			}
			res.Out[wordOf(ev)]++
			if _, seen := newID[old.ID]; !seen {
				if err := batch.Querier().QueryRow(`SELECT MAX("id") FROM "events"`).Scan(&wrote); err != nil {
					return nil, fmt.Errorf("reading the id of the event written from source event %d: %w", old.ID, err)
				}
				newID[old.ID] = wrote
			}
			if anchorAfter != nil {
				// The act landed whether or not its Anchor does, so what follows still pairs with it.
				if _, err := batch.Append(record.Identity{Run: dst, SeatID: seatID}, anchorAfter); err != nil {
					res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
				} else {
					res.Out[recordpb.Word(recordpb.EventType_EVENT_TYPE_ANCHOR)]++
				}
			}
			if _, seen := newKey[old.Key]; old.Key != "" && !seen {
				newKey[old.Key] = ev.GetKey()
			}
			if id.Correct != nil {
				paired[fix.ID] = true
				res.Out[correctionWord]++
			}
		}
	}
	// THE BATCH'S FAILURE IS ASKED ONCE, HERE. After a savepoint fails, every later append answers
	// the same failure and each was recorded above as a refusal; none of it may stand.
	if err := batch.Err(); err != nil {
		return nil, err
	}
	res.IDs, res.GapAnchors = rm.ids, place.census
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
//
// The cast is written through batch, under dst as its Identity's run.
func writeSynthesizedCast(evs []OldEvent, rm *remap, batch *record.Batch, dst record.Run) (bool, error) {
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
	_, err := batch.Append(record.Identity{Run: dst, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast})
	return err == nil, err
}
