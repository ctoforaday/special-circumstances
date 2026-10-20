package migrate

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
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
	ProseRewrites  []ProseRewrite    // every archived id rewritten inside a seat's wording, in order

	base string // the report's base as the source wrote it
}

// preFieldFillWhy is the reason a mint whose source predates class_material records for the
// stamp the live rule gives it.
const preFieldFillWhy = "the source predates class_material; the gap takes the material default of the value its archived slug translates to"

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
	source := src // the loop below names each act's own source event src
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
	// A CORRECTION CHAIN IS WRITTEN ONCE, AS ITS HEAD, IN ITS ROOT'S PLACE. A source at epoch 26 or
	// earlier may hold a seat's same-sitting correction: a Correction naming the act it struck and
	// the replacement that took its place, the replacement written right beside it. The record holds
	// no correction, and every act stands as filed; the act the seat meant is the chain's HEAD, so
	// it is written where the chain began — at the root's place, under the root's stamp — as the
	// ordinary act it is, and every later link and every Correction is passed over. The root's place
	// is the one every reader of the source gave the chain (its standing order), so no act that
	// depends on it changes sides. A chain whose head translates to no event writes nothing.
	const correctionWord = "correction" // the word a source spells a Correction with
	corrects := map[string]string{}     // source key of a corrected act -> the key of its replacement
	replaces := map[string]bool{}       // source keys that replaced another act
	byKey := map[string]OldEvent{}
	for _, old := range evs {
		if old.Key != "" {
			byKey[old.Key] = old
		}
		if old.Word == correctionWord {
			c, _ := old.Fields["corrects"].(string)
			r, _ := old.Fields["replacement"].(string)
			if c != "" && r != "" {
				corrects[c], replaces[r] = r, true
			}
		}
	}
	// chainOf is the chain a root begins: its links in order, root first, and its head.
	chainOf := func(root string) (links []string, head OldEvent, ok bool) {
		for k := root; ; {
			links = append(links, k)
			next, corrected := corrects[k]
			if !corrected {
				head, ok = byKey[k]
				return links, head, ok
			}
			if len(links) > len(corrects) {
				return links, OldEvent{}, false
			}
			k = next
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
	// The learning pass writes nothing, so what it rewrote is not what the record holds.
	rm.prose = nil
	place := newPlacer(batch.Querier(), evs, rm.ids)
	// THE EVENTS.ID A DISPATCH PINS IS TRANSLATED AS AN ID IS. The cast written first, an anchor
	// placed after an act and a source event that writes nothing each move every later row, so a
	// pin carried as written names whatever the migrated record holds at that number. A pin names
	// the report head its party audits, which the source wrote before the dispatch; the migrated
	// pin is the id of the event that source head was written as.
	newID := map[int64]int64{} // source events.id -> the events.id of the first event written from it
	var wrote int64
	proofAnchors := map[string][]string{} // proof_sha -> the anchors of the proofs written on it, in order
	proposalMethod := map[string]string{} // inquiry id -> the method its proposal carried (before epoch 30)
	for _, old := range evs {
		res.In[old.Word]++
		// A Correction, and a replacement, are written with the chain's root.
		if old.Word == correctionWord || replaces[old.Key] {
			continue
		}
		// src is what this place writes: the act itself, or the head of the chain it begins.
		src, links, chained := old, []string{old.Key}, false
		if _, corrected := corrects[old.Key]; corrected && old.Key != "" {
			var head OldEvent
			var ok bool
			links, head, ok = chainOf(old.Key)
			if !ok {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word,
					Err: fmt.Sprintf("this act begins a correction chain (%s) whose head the source does not hold — the chain's last replacement is absent, so nothing says which act stands", strings.Join(links, " -> "))})
				continue
			}
			src, chained = head, true
		}
		seatID, err := rm.seat(old.SeatID)
		if err != nil {
			res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		bodies, err := reg.Translate(src)
		if err != nil {
			var loss LossError
			if as(err, &loss) && opt.AcceptLoss[loss.Word] {
				res.AcceptedLosses[loss.Word] = loss.Reason
				continue
			}
			res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		if bodies, err = reproduceAnchors(bodies, proofAnchors); err != nil {
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
		// refused words a refusal of an act written for a chain as the chain's.
		refused := func(err error) string {
			if chained {
				return fmt.Sprintf("the correction chain %s is written as its head at the root's place, and the write refuses it there: %v", strings.Join(links, " -> "), err)
			}
			return err.Error()
		}
		for _, body := range bodies {
			// NO EPOCH IS CARRIED. The archived row's epoch was recovered from its seat id by regex
			// when it was written; the re-driven row's EPOCH comes from the chair sittings already
			// replayed (events_w."epoch"), which is the same fact from the record rather than from
			// the name — plans/roundless.md §III.A.2. A migrated record's round therefore means
			// what a fresh record's does.
			rm.apply(body, src.Fields["label"])
			if q, ok := body.(*recordpb.Inquiry); ok {
				carryInquiryMethod(q, proposalMethod)
			}
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
			// A STORED PLAN NAMES EVENTS BY events.id inside bytes a relay is compared with, and its
			// key is over them. translatePlanIDs carries each through newID and re-keys the plan, and
			// refuses an id it cannot place.
			if pl, ok := body.(*recordpb.Plan); ok {
				if err := translatePlanIDs(pl, newID); err != nil {
					res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
			}
			// A FINDING ABOUT A CONTRADICTION NAMES IT BY events.id — the reading that opened it —
			// and is translated as the pin is: through the id of the event that reading was written
			// as. An untranslated id names whatever the migrated record holds at that number.
			if f, ok := body.(*recordpb.Finding); ok && f.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_CONTRADICTION {
				to, err := contradictionRef(f.GetAboutRef(), newID)
				if err != nil {
					res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				f.AboutRef = proto.String(to)
			}
			// A MINT THAT PREDATES class_material is stamped by the live rule — its translated
			// value's material default, at the write — and the fill is stated, every one of them:
			// which of them moved is computable only against a table this binary does not hold. A
			// stamp the source carries is kept as recorded.
			if m, ok := body.(*recordpb.Mint); ok && m.ClassMaterial == nil && m.Class != nil {
				slug, _ := old.Fields["class"].(string)
				res.StatedFills = append(res.StatedFills, StatedFill{
					Where: "mint " + m.GetGapId(), Slug: slug,
					Value: recordpb.Word(recordpb.MaterialDefault(m.GetClass())), Why: preFieldFillWhy})
			}
			if err := proseRefs(source, batch.Querier(), body); err != nil {
				return nil, err
			}
			// A CLOSURE IS NOT ITS OWN ANCHOR. Before the write refused it, a close could name the gap it
			// closed among the ids it was verified with (m30: three closures of seven). That id attests
			// nothing a reader could check, so it is not carried; the closure's other anchors are.
			if c, ok := body.(*recordpb.Close); ok {
				c.Anchors = slices.DeleteFunc(c.Anchors, func(a string) bool { return a == c.GetGapId() })
			}
			anchorAfter, err := place.step(body)
			if err != nil {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: refused(err)})
				continue
			}
			ev, err := batch.Append(id, body)
			if err != nil {
				res.Refusals = append(res.Refusals, Refusal{OldID: old.ID, Word: old.Word, Err: refused(err)})
				continue
			}
			res.Out[wordOf(ev)]++
			if p, ok := body.(*recordpb.Proof); ok && p.GetProofSha() != "" {
				proofAnchors[p.GetProofSha()] = append(proofAnchors[p.GetProofSha()], p.GetProofId())
			}
			if _, seen := newID[old.ID]; !seen {
				if err := batch.Querier().QueryRow(`SELECT MAX("id") FROM "events"`).Scan(&wrote); err != nil {
					return nil, fmt.Errorf("reading the id of the event written from source event %d: %w", old.ID, err)
				}
				// Every link of a chain is the one act written here.
				for _, k := range links[1:] {
					newID[byKey[k].ID] = wrote
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
		}
	}
	// THE BATCH'S FAILURE IS ASKED ONCE, HERE. After a savepoint fails, every later append answers
	// the same failure and each was recorded above as a refusal; none of it may stand.
	if err := batch.Err(); err != nil {
		return nil, err
	}
	res.IDs, res.GapAnchors, res.ProseRewrites = rm.ids, place.census, rm.prose
	return res, nil
}

// carryInquiryMethod moves an inquiry's method from its proposal to its fates. BEFORE EPOCH 30 THE
// METHOD WAS THE PROPOSAL'S, written before any research; from epoch 30 it is a fate's account and a
// proposal holds none (the write refuses one). So a proposal's method is taken off it and given to
// every fate recorded on that inquiry that states none, and the migrated report prints on each row
// the method the source printed. An inquiry with no fate beyond its proposal has nowhere to carry it:
// its method is dropped, and that row prints none — the residue of this translation, which
// TestTheProposalsMethodMovesToItsFates states. A source at epoch 30 or later holds no proposal
// method, and passes through unchanged.
//
// ONLY THE PROPOSAL'S METHOD IS CARRIED, never an earlier fate's. No write path before epoch 30 put
// a method on a fate (none of the 23 archived runs holds one), so the source's fold printed the
// proposal's; and at epoch 30 a fate that states none clears the method, so carrying an earlier
// fate's forward would rewrite what a newer source prints.
func carryInquiryMethod(q *recordpb.Inquiry, held map[string]string) {
	if q.SupersedesStatus == nil {
		if m := q.GetMethod(); m != "" {
			held[q.GetInquiryId()] = m
		}
		q.Method = nil
		return
	}
	if m, ok := held[q.GetInquiryId()]; ok && q.GetMethod() == "" {
		q.Method = proto.String(m)
	}
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

// reproduceAnchors carries a reproduce written before epoch 28 — keyed by its script's sha, with no
// anchor — to the anchors it judged (gblock, 2026-10-12: reproduced per sha, sound per anchor).
// Every proof on that script written before it gets the reproduce's soundness and note: its
// verdict was the only one any of those sentences had. The re-run is the script's, so the first
// anchor carries it with its outputs, and each other anchor is judged on that re-run (not_run).
// A proof on the script written after it gets nothing: it was never judged.
//
// A reproduce with no proof of its script before it has no sentence to carry its verdict to, and
// is refused by name. A reproduce that names its anchor passes as written.
func reproduceAnchors(bodies []proto.Message, proofAnchors map[string][]string) ([]proto.Message, error) {
	var out []proto.Message
	for _, b := range bodies {
		r, ok := b.(*recordpb.Reproduce)
		if !ok || r.ProofId != nil {
			out = append(out, b)
			continue
		}
		anchors := proofAnchors[r.GetProofSha()]
		if len(anchors) == 0 {
			return nil, fmt.Errorf("this reproduce re-ran script %s, and no proof on that script was written before it — there is no anchor to carry its verdict to", r.GetProofSha())
		}
		for i, a := range anchors {
			c := proto.Clone(r).(*recordpb.Reproduce)
			c.ProofId = proto.String(a)
			if i > 0 {
				c.Rerun = recordpb.Rerun_RERUN_NOT_RUN.Enum()
				c.RecordedOutput, c.ObservedOutput = nil, nil
			}
			out = append(out, c)
		}
	}
	return out, nil
}

// contradictionRef translates a finding's contradiction reference — a source events.id — to the
// events.id the reading it names was written as. A reference the replay wrote nothing for is a
// refusal: left as it was, it would name whatever the migrated record holds at that number.
func contradictionRef(ref string, newID map[int64]int64) (string, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(ref), 10, 64)
	if err != nil {
		return "", fmt.Errorf("this finding answers contradiction %q, which is not an events.id", ref)
	}
	to, landed := newID[n]
	if !landed {
		return "", fmt.Errorf("this finding answers contradiction %d, whose reading wrote no event on the migrated record before it — the reference would name whatever event holds that number", n)
	}
	return strconv.FormatInt(to, 10), nil
}
