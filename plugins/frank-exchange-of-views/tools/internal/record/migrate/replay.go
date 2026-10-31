package migrate

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// Options is the operator's say over a replay.
type Options struct {
	// AcceptLoss names the words whose documented untranslatability the operator accepts.
	// Acceptance is per word: it silences exactly that LossError and nothing else, and the
	// manifest records what was accepted.
	AcceptLoss map[string]bool
}

// Result is what one replay did, counted where the manifest needs counts.
type Result struct {
	In                 map[string]int // old events per word, as read
	Out                map[string]int // new events per word, as written
	Refusals           []migration.Refusal
	AcceptedLosses     map[string]string             // word -> the entry's stated reason
	SourceHash         string                        // sha256 over the event stream AS READ — the logical record
	IDs                map[string]string             // archived id or finding label -> the id the migrated record carries
	Serialized         map[string]int                // archived instance seat -> events moved after instance 1 (serializeInstances)
	StatedFills        []migration.StatedFill        // values supplied where the source predates the field
	ClassTranslations  []migration.ClassTranslation  // the archived class slugs an authored row of gapclass.go carried, and onto what
	SupersededOutcomes []migration.SupersededOutcome // the source outcomes a later one replaced, not written (outcomes.go)
	GapAnchors         migration.GapPlacement        // what bringing the run's gaps onto anchors did (F-c)
	ReportOrder        string                        // what became of the order of the report's text (order.go)
	ProseRewrites      []migration.ProseRewrite      // every archived id rewritten inside a seat's wording, in order

	computations []migratedComputation // every computation written, and where its artifact is (computations.go)
	reruns       map[string][]string   // bundle -> the rerun_sha of each reproduce written that names one

	parts []string // each part of the report as the source wrote it, in record order
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
// inOrder set, each part of the report is written in the order its text ships in (order.go).
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
	rm.archivedNotes = true
	if sq, ok := src.(*SQLiteSource); ok {
		if e, known := sq.SourceEpoch(); known && e >= computationEpoch {
			rm.archivedNotes = false
		}
	}
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
	// AN OUTCOME THE SOURCE FOLLOWS WITH ANOTHER IS NOT WRITTEN (outcomes.go), and a corrected one is
	// read as its chain's head, which is what its place writes.
	superseded := supersededOutcomes(evs, reg, replaces, func(old OldEvent) OldEvent {
		if _, corrected := corrects[old.Key]; corrected && old.Key != "" {
			if _, head, ok := chainOf(old.Key); ok {
				return head
			}
		}
		return old
	})
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
	admits := predatesAdmission(source) // its closed sittings are admitted at their close
	// THE EVENTS.ID A DISPATCH PINS IS TRANSLATED AS AN ID IS. The cast written first, an anchor
	// placed after an act and a source event that writes nothing each move every later row, so a
	// pin carried as written names whatever the migrated record holds at that number. A pin names
	// the report head its party audits, which the source wrote before the dispatch; the migrated
	// pin is the id of the event that source head was written as.
	newID := map[int64]int64{} // source events.id -> the events.id of the first event written from it
	var wrote int64
	computationAnchors := map[string][]string{}  // bundle_sha -> the anchors of the computations written on it, in order
	bundleOf := map[string]string{}              // archived script sha -> the bundle_sha its computation was written with
	standing := map[string]*recordpb.Reproduce{} // bundle_sha -> the latest reproduce written that ran it
	res.reruns = map[string][]string{}
	proposalMethod := map[string]string{} // inquiry id -> the method its proposal carried (before epoch 30)
	classMints := map[string]int{}        // archived class slug an authored row translates -> the mints written with it
	for _, old := range evs {
		res.In[old.Word]++
		// A Correction, and a replacement, are written with the chain's root.
		if old.Word == correctionWord || replaces[old.Key] {
			continue
		}
		if s, ok := superseded[old.ID]; ok {
			res.SupersededOutcomes = append(res.SupersededOutcomes, s)
			continue
		}
		// src is what this place writes: the act itself, or the head of the chain it begins.
		src, links, chained := old, []string{old.Key}, false
		if _, corrected := corrects[old.Key]; corrected && old.Key != "" {
			var head OldEvent
			var ok bool
			links, head, ok = chainOf(old.Key)
			if !ok {
				res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
					Err: fmt.Sprintf("this act begins a correction chain (%s) whose head the source does not hold — the chain's last replacement is absent, so nothing says which act stands", strings.Join(links, " -> "))})
				continue
			}
			src, chained = head, true
		}
		seatID, err := rm.seat(old.SeatID)
		if err != nil {
			res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		bodies, err := reg.Translate(src)
		if err != nil {
			var loss LossError
			if as(err, &loss) && opt.AcceptLoss[loss.Word] {
				res.AcceptedLosses[loss.Word] = loss.Reason
				continue
			}
			res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		// AN ARCHIVED REPRODUCE NAMES ITS SCRIPT BY THE SCRIPT'S SHA, and the record names a bundle:
		// the computation written on that script says which.
		for _, b := range bodies {
			if r, ok := b.(*recordpb.Reproduce); ok {
				if nb, mapped := bundleOf[r.GetBundleSha()]; mapped {
					r.BundleSha = proto.String(nb)
				}
			}
		}
		if bodies, err = reproduceAnchors(bodies, computationAnchors); err != nil {
			res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		ts, err := record.ParseStamp(old.TS)
		if err != nil {
			res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
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
			// A RE-RUN'S DRIFT IS THE PROGRAM'S BYTES, which are not the debate's wording: an anchor
			// token a program printed is what it printed, so it is kept through the respelling.
			repro, _ := body.(*recordpb.Reproduce)
			// The bytes are copied, not the pointer: the respelling sets a field through the
			// pointer it already holds, so a kept pointer would read back the respelled bytes.
			var drift *string
			if repro != nil && repro.Drift != nil {
				drift = proto.String(repro.GetDrift())
			}
			rm.apply(body, src.Fields["label"])
			if repro != nil {
				repro.Drift = drift
				// AN ACT THAT RAN NOTHING STANDS ON ITS BUNDLE'S RE-RUN and states that re-run's
				// comparison, which the archive left to the reader (`not_run`).
				if !repro.GetRan() && repro.Rerun == nil {
					st, ok := standing[repro.GetBundleSha()]
					if !ok {
						res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
							Err: fmt.Sprintf("this reproduce of %s ran nothing, and no reproduce before it ran its bundle — there is no re-run for it to stand on", repro.GetComputationId())})
						continue
					}
					repro.Rerun, repro.Drift = st.Rerun, st.Drift
				}
			}
			if q, ok := body.(*recordpb.Inquiry); ok {
				carryInquiryMethod(q, proposalMethod)
			}
			// AN ARGUMENT THE WRITE REFUSES IS WRITTEN AS NO EVENT (epoch 38). A closing is a
			// party's argument on a docketed gap and a position is the chair's; an archive holds the
			// chair's closings, blue's and the lanes' positions, and blue's closings on gaps no
			// docket motion stood on. Each is an argument its seat made under the rule of its day:
			// re-attributed, it is authorship the record never held, and converted, it asserts a
			// filing never made, so it is left out — the residue, stated with the change, that
			// run-archive/ holds byte for byte. Asked after rm.apply, so a shard-era gap label
			// reads as the gap it was remapped to.
			if drop, err := refusedArgument(batch.Querier(), body, seatID); err != nil {
				return nil, err
			} else if drop {
				continue
			}
			if b, ok := body.(*recordpb.ReportPart); ok {
				res.parts = append(res.parts, b.GetText())
				if inOrder {
					b.Text = proto.String(report.Ordered(b.GetText()))
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
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
						Err: fmt.Sprintf("this dispatch pins event %d, which wrote no event on the migrated record before it — the pin would name whatever event holds that number", d.GetPin())})
					continue
				}
				d.Pin = proto.Int64(to)
			}
			// A STALE REFUSAL'S READ POINT is an events.id too: the last act the refusal showed. It
			// is translated as a pin is, and refused where the act it names wrote no event.
			if st, ok := body.(*recordpb.Stale); ok && st.Through != nil {
				to, landed := newID[st.GetThrough()]
				if !landed {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
						Err: fmt.Sprintf("this stale refusal reads from event %d, which wrote no event on the migrated record before it — the read point would name whatever event holds that number", st.GetThrough())})
					continue
				}
				st.Through = proto.Int64(to)
			}
			// A REPORT READ'S HEAD is an events.id too — the lens's pin — translated as a dispatch's.
			if rr, ok := body.(*recordpb.ReportRead); ok && rr.Head != nil {
				to, landed := newID[rr.GetHead()]
				if !landed {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
						Err: fmt.Sprintf("this report read was served at event %d, which wrote no event on the migrated record before it — the pin would name whatever event holds that number", rr.GetHead())})
					continue
				}
				rr.Head = proto.Int64(to)
			}
			// A STORED PLAN NAMES EVENTS BY events.id inside bytes a relay is compared with, and its
			// key is over them. TranslatePlanIDs carries each through newID and re-keys the plan, and
			// refuses an id it cannot place; TranslatePlanKinds first respells its archived blocker kinds.
			if pl, ok := body.(*recordpb.Plan); ok {
				if err := TranslatePlanKinds(pl); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				if err := TranslatePlanIDs(pl, newID); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				if err := CarryPlanVerdict(pl, batch.Querier()); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
			}
			// A FINDING ABOUT A CONTRADICTION NAMES IT BY events.id — the reading that opened it —
			// and is translated as the pin is: through the id of the event that reading was written
			// as. An untranslated id names whatever the migrated record holds at that number.
			if f, ok := body.(*recordpb.Finding); ok && f.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_CONTRADICTION {
				to, err := contradictionRef(f.GetAboutRef(), newID)
				if err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				f.AboutRef = proto.String(to)
			}
			// A MINT THAT PREDATES class_material is stamped by the live rule — its translated
			// value's material default, at the write — and the fill is stated, every one of them:
			// which of them moved is computable only against a table this binary does not hold. A
			// stamp the source carries is kept as recorded. Both the fill and the translation count
			// are recorded once the write lands: a refused mint stamped nothing.
			var translatedSlug string
			var fill *migration.StatedFill
			if m, ok := body.(*recordpb.Mint); ok {
				slug, _ := src.Fields["class"].(string)
				if archivedGapClasses[slug] != "" {
					translatedSlug = slug
				}
				if m.ClassMaterial == nil && m.Class != nil {
					fill = &migration.StatedFill{Where: "mint " + m.GetGapId(), Slug: slug,
						Value: recordpb.Word(recordpb.MaterialDefault(m.GetClass())), Why: preFieldFillWhy}
				}
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
				res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: refused(err)})
				continue
			}
			// A SITTING CLOSED BEFORE ENVELOPES WERE RECORDED AS ADMITTED IS ADMITTED AT ITS CLOSE
			// (admissionBefore), so no past sitting reads as failed.
			if cl, ok := body.(*recordpb.SittingClose); ok && admits {
				a, err := admissionBefore(batch.Querier(), cl)
				if err != nil {
					return nil, err
				}
				if a != nil {
					if _, err := batch.Append(record.Identity{Run: dst, SeatID: record.HarnessSeat}, a); err != nil {
						res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: "admitting the sitting this close ends: " + err.Error()})
						continue
					}
					res.Out[admissionWord]++
				}
			}
			ev, err := batch.Append(id, body)
			if err != nil {
				res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: refused(err)})
				continue
			}
			res.Out[wordOf(ev)]++
			if translatedSlug != "" {
				classMints[translatedSlug]++
			}
			if fill != nil {
				res.StatedFills = append(res.StatedFills, *fill)
			}
			if p, ok := body.(*recordpb.Computation); ok && p.GetBundleSha() != "" {
				computationAnchors[p.GetBundleSha()] = append(computationAnchors[p.GetBundleSha()], p.GetComputationId())
				mc := migratedComputation{id: p.GetComputationId(), bundle: p.GetBundleSha(), argv: p.GetArgv()}
				if old, archived := archivedSha(src); archived {
					mc.oldSha = old
					bundleOf[old] = p.GetBundleSha()
					if out, inline := src.Fields["output"].(string); inline {
						mc.output = &out
					}
				}
				res.computations = append(res.computations, mc)
			}
			if r, ok := body.(*recordpb.Reproduce); ok {
				if r.GetRan() {
					standing[r.GetBundleSha()] = r
				}
				if r.GetRerunSha() != "" {
					res.reruns[r.GetBundleSha()] = append(res.reruns[r.GetBundleSha()], r.GetRerunSha())
				}
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
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
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
	res.ClassTranslations = classTranslations(classMints)
	if res.SupersededOutcomes == nil {
		res.SupersededOutcomes = []migration.SupersededOutcome{}
	}
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
// Every computation on that bundle written before it gets the reproduce's soundness and note: its
// verdict was the only one any of those sentences had. The re-run is the bundle's, so the first
// anchor carries it (`ran`), and each other anchor ran nothing and states that re-run's comparison.
// A computation on the bundle written after it gets nothing: it was never judged.
//
// A reproduce with no computation of its bundle before it has no sentence to carry its verdict to,
// and is refused by name. A reproduce that names its anchor passes as written.
func reproduceAnchors(bodies []proto.Message, computationAnchors map[string][]string) ([]proto.Message, error) {
	var out []proto.Message
	for _, b := range bodies {
		r, ok := b.(*recordpb.Reproduce)
		if !ok || r.ComputationId != nil {
			out = append(out, b)
			continue
		}
		anchors := computationAnchors[r.GetBundleSha()]
		if len(anchors) == 0 {
			return nil, fmt.Errorf("this reproduce re-ran script %s, and no proof on that script was written before it — there is no anchor to carry its verdict to", r.GetBundleSha())
		}
		for i, a := range anchors {
			c := proto.Clone(r).(*recordpb.Reproduce)
			c.ComputationId = proto.String(a)
			c.Ran = proto.Bool(i == 0)
			out = append(out, c)
		}
	}
	return out, nil
}

// archivedSha is the script sha an archived proof names, when src is one: a source written before
// epoch 49 records a proof, whose artifact is under proofs/<sha>/.
func archivedSha(src OldEvent) (string, bool) {
	if src.Word != "proof" {
		return "", false
	}
	for _, col := range []string{"proof_sha", "sha256"} {
		if v, ok := src.Fields[col].(string); ok && v != "" {
			return v, true
		}
	}
	return "", false
}

// contradictionRef translates a finding's contradiction reference — a source events.id — to the
// events.id the reading it names was written as. A reference the replay wrote nothing for is a
// refusal: left as it was, it would name whatever the migrated record holds at that number.
func contradictionRef(ref string, newID map[int64]int64) (string, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(ref), 10, 64)
	if err != nil {
		return "", fmt.Errorf("this finding raises contradiction %q, which is not an events.id", ref)
	}
	to, landed := newID[n]
	if !landed {
		return "", fmt.Errorf("this finding raises contradiction %d, whose reading wrote no event on the migrated record before it — the reference would name whatever event holds that number", n)
	}
	return strconv.FormatInt(to, 10), nil
}

// refusedArgument reports whether body is a closing or a position the write path refuses from
// seatID at this place (record.ClosingRefusal, record.PositionRefusal) — one replay writes as no
// event. A read that fails is returned, never taken for a refusal.
func refusedArgument(q recordsql.Querier, body proto.Message, seatID string) (bool, error) {
	var err error
	switch b := body.(type) {
	case *recordpb.Closing:
		err = record.ClosingRefusal(q, b.GetGapId(), seatID)
	case *recordpb.Position:
		err = record.PositionRefusal(seatID)
	}
	if err == nil {
		return false, nil
	}
	var refusal *feov.Error
	if !errors.As(err, &refusal) {
		return false, fmt.Errorf("reading whether the write refuses an archived %T: %w", body, err)
	}
	return true, nil
}
