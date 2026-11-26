package migrate

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
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
	In                  map[string]int // old events per word, as read
	Out                 map[string]int // new events per word, as written
	Refusals            []migration.Refusal
	AcceptedLosses      map[string]string             // word -> the entry's stated reason
	SourceHash          string                        // sha256 over the event stream AS READ — the logical record
	IDs                 map[string]string             // archived id -> the id the migrated record carries
	Serialized          map[string]int                // archived instance seat -> events moved after instance 1 (serializeInstances)
	StatedFills         []migration.StatedFill        // values supplied where the source predates the field
	ClassTranslations   []migration.ClassTranslation  // the archived class slugs an authored row of gapclass.go carried, and onto what
	SupersededOutcomes  []migration.SupersededOutcome // the source outcomes a later one replaced, not written (outcomes.go)
	GapAnchors          migration.GapPlacement        // what bringing the run's gaps onto anchors did (F-c)
	ReportOrder         string                        // what became of the order of the report's text (order.go)
	ProseRewrites       []migration.ProseRewrite      // every archived id rewritten inside a seat's wording, in order
	DroppedPlanBlockers int                           // stored plans' archived reading blockers of a found source, which leave the plan (TranslateReadingBlockers)
	QuotelessReadings   []int64                       // source ids of the readings whose quote held only tokens the fold takes out, written as no event (EventSchema 80)
	SettledEdits        []int64                       // source ids of the edits whose change a finding's token kept from standing, written as no event (EventSchema 80)
	UnowedEngagements   []int64                       // source ids of the engagements whose sitting owes no successor once cross-seat links drop, written as no event (EventSchema 80)
	Dockets             migration.DocketResidue       // Translations 3 and 4's residue: the source ids written as no event (EventSchema 80)
	LostActs            []migration.Refusal           // archived acts the delivered-text rule refuses, written as no event (undeliveredLoss)

	computations []migratedComputation          // every computation written, and where its artifact is (computations.go)
	reruns       map[string][]string            // bundle -> the rerun_sha of each reproduce whose source names one
	derived      map[string][]compute.Execution // bundle -> each execution an archived reproduce that reproduced stands for (computations.go)

	parts []string // each part of the report as the source wrote it, in record order
}

// undeliveredLoss is the accepted loss of an archived act the delivered-text rule refuses.
const undeliveredLoss = "an archived close or spot-check stands only on report text delivered to its seat, and its sitting holds no recorded delivery; it is written as no event, each such act listed in lost_acts, and the run's stamped outcome is carried as history"

// dependentLossWhy is the accepted loss of an archived act refused only because it depends on a lost act.
const dependentLossWhy = "an archived act refused only because it depends on a gap a lost act took with it is that same loss; it is written as no event, each such act listed in lost_acts with the act it depended on"

// preFieldFillWhy is the reason a mint whose source predates class_material records for the
// stamp the live rule gives it.
const preFieldFillWhy = "the source predates class_material; the gap takes the material default of the value its archived slug translates to"

// restampWhy is the reason a mint whose archived class_material is not its class's material default
// today records for the stamp the live rule gives it. %s is the archived stamp.
const restampWhy = "the source stamped %s; the tool stamps a mint's class_material from its class's material default, and the gap takes that default"

// Replay re-drives an old record's events, in order, through the replay writer (record.Replay) into
// dst. Each act is held to the shape every writer is held to, and to none of the live write's act
// judgements: an archived act is what a seat did. Each event's stamp is the ORIGINAL stamp, handed
// to the writer with the act.
//
// Register events go through Append like everything else, NOT through RegisterSeat — a
// deliberate deviation from the plan's first draft, with the reason on the record:
// RegisterSeat stamps ENGINE-OBSERVED facts from its own environment (tool_version,
// agent_type, run_via), and the migrator's environment is not the run's. The old register
// body already carries what was actually observed then, and replaying it verbatim is the
// only honest answer. Dispatch numbering survives because it derives from the key ordinal,
// which Append's own path re-derives.
//
// The whole replay is one transaction (record.Replay): every act is written inside a savepoint of
// it, and committed once at the end. Replay holds no source run, so it derives no execution for an
// archived reproduce (deriveExecution); Migrate does.
func Replay(src Source, reg Registry, dst record.Run, opt Options) (*Result, error) {
	batch, err := record.BeginReplay(dst)
	if err != nil {
		return nil, err
	}
	defer func() { _ = batch.Rollback() }()
	res, err := replay(src, "", reg, batch, dst, opt, false)
	if err != nil {
		return nil, err
	}
	if err := batch.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

// replay writes src's events through batch, the replay writer its caller began on dst, holds and
// commits, each act stamped with its source event's time. With inOrder set, each part of the report
// is written in the order its text ships in (order.go). fromDir is the source run, whose stored
// outputs an archived reproduce that reproduced is derived from (derivedRerun); "" where the caller
// holds no source run, and then none is derived.
//
// dst IS AN ADDRESS, NEVER A HANDLE. While the batch is open, the batch's transaction holds the
// record's one connection, and anything that takes dst and opens the record waits on it forever. So
// dst is used here only as the Run of the Identity each act is written under; every read of the
// destination goes through batch.Querier().
//
// An act the write refuses is a Refusal, and the replay goes on. A failure of the batch itself
// (record.ErrReplayFailed) is asked of the batch once, at the end — whichever append it landed on —
// and returned: nothing the batch wrote can be kept.
func replay(src Source, fromDir string, reg Registry, batch *record.Replay, dst record.Run, opt Options, inOrder bool) (*Result, error) {
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
	// A LABELLED CORROBORATION'S CITATION IS DROPPED (EventSchema 59): no reading places a citation.
	// Its label leaves every field (remap.dropped) and the Anchor that placed it is written as no
	// event; a reading that named it comes forward as a reading of the label's source. The labels and
	// their sources are read off the SOURCE stream: the migrated Verify has no label to be found by.
	labelled := labelSources(evs, rm)
	// A FINDING IS WRITTEN AS NO EVENT (EventSchema 80), and every name it carries leaves the id
	// lists and the token fields of the record (remap.dropped). The names are read off the SOURCE
	// stream: no translated body carries them.
	rm.dropFindings(findingNames(evs))
	// AN IN-ACT MINT OF A CITED READING COMES FORWARD ABOUT ITS CITATION (EventSchema 71): the
	// citation its carrying reading named is the subject the live mint names, read off the source
	// stream, since the migrated Verify names no gap.
	cited := citedInActMints(evs)
	// AN ARCHIVED DOCKET IS A DOCKETING, OR NO EVENT (EventSchema 80, docketing.go): which docket the
	// dispatch filed is read off the source stream's order, which no translated body carries.
	dockets, err := docketsOf(evs, reg, rm)
	if err != nil {
		return nil, err
	}
	reopened := map[string]bool{} // gaps whose close a dropped bench act took away (reopenedSpotCheck)
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
	//
	// Under the archive's own clock: the cast precedes the first event and is stamped with its time,
	// so ORDER BY id and the timestamps tell one story.
	castAt := record.Now()
	if len(evs) > 0 {
		if ts, err := record.ParseStamp(evs[0].TS); err == nil {
			castAt = ts
		}
	}
	if wrote, err := writeSynthesizedCast(evs, rm, batch, dst, castAt); err != nil {
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
			rm.apply(body)
		}
	}
	// The learning pass writes nothing, so what it rewrote is not what the record holds.
	rm.prose = nil
	place := newPlacer(batch.Querier(), evs, rm.ids)
	readDropped := readsDropped(evs, rm.seat)
	admits := predatesAdmission(source) // its closed sittings are admitted at their close
	// THE EVENTS.ID A DISPATCH PINS IS TRANSLATED AS AN ID IS. The cast written first, an anchor
	// placed after an act and a source event that writes nothing each move every later row, so a
	// pin carried as written names whatever the migrated record holds at that number. A pin names
	// the report head its party audits, which the source wrote before the dispatch; the migrated
	// pin is the id of the event that source head was written as.
	newID := map[int64]int64{} // source events.id -> the events.id of the first event written from it
	var wrote int64
	// A PIN THAT NAMED AN ACT LEFT OUT BY DESIGN names the record as it stood before it: the last
	// event written, which is what a read at that act saw less the act. A dispatch's pin, a stale
	// refusal's read point, a report read's head and a stored plan's ids that named such an act each
	// translate to it, so the act's absence moves no read point past an act the seat had not seen.
	// The acts left out this way are an act on a superseded gap (actOnSupersededGap), a finding
	// (EventSchema 80) and a quoteless reading; a settled edit is headBefore's. An act the write REFUSED is not one of
	// them, nor is a dropped Anchor, which no archived pin names: a pin naming either is refused,
	// below, since the record never held what it read.
	pinBefore := func(oldID int64) error {
		if _, seen := newID[oldID]; seen {
			return nil
		}
		var last int64
		if err := batch.Querier().QueryRow(`SELECT COALESCE(MAX("id"), 0) FROM "events"`).Scan(&last); err != nil {
			return fmt.Errorf("reading the last event written before source event %d: %w", oldID, err)
		}
		newID[oldID] = last
		return nil
	}
	// headBefore translates a pin naming a settled edit to the report head it left standing: the
	// latest edit or part written, which renders the report that edit's change already stood in.
	headBefore := func(oldID int64) error {
		var head int64
		if err := batch.Querier().QueryRow(`SELECT COALESCE(MAX("id"), 0) FROM "events" WHERE "type" IN ('blue_edit', 'report_part')`).Scan(&head); err != nil {
			return fmt.Errorf("reading the report head before source event %d: %w", oldID, err)
		}
		newID[oldID] = head
		return nil
	}
	computationAnchors := map[string][]string{}  // bundle_sha -> the anchors of the computations written on it, in order
	bundleOf := map[string]string{}              // archived script sha -> the bundle_sha its computation was written with
	standing := map[string]*recordpb.Reproduce{} // bundle_sha -> the latest reproduce written that ran it
	res.reruns = map[string][]string{}
	res.derived = map[string][]compute.Execution{}
	computationOf := map[string]migratedComputation{} // computation_id -> the computation written with it

	proposalMethod := map[string]string{} // inquiry id -> the method its proposal carried (before epoch 30)
	classMints := map[string]int{}        // archived class slug an authored row translates -> the mints written with it
	lost := lostGaps{}                    // each gap a lost act took with it (lostacts.go)
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
		if readDropped[old.ID] {
			continue // a second serve in its seat's sitting, or another seat's (readsDropped)
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
		if src.Word == "finding" {
			if err := pinBefore(old.ID); err != nil {
				return nil, err
			}
			continue
		}
		// A BENCH OPINION, AND A DOCKET NO DISPATCH FILED WITH ITS RULING, IS WRITTEN AS NO EVENT
		// (docketing.go); the gap each closed reads open where nothing else closed it.
		if src.Word == "opinion" {
			res.Dockets.Opinions = append(res.Dockets.Opinions, old.ID)
			if g, ok := src.Fields["gap_id"].(string); ok {
				reopened[rm.respelled(g)] = true
			}
			if err := pinBefore(old.ID); err != nil {
				return nil, err
			}
			continue
		}
		var drop docketDrop
		if bodies, drop = dockets.translate(bodies); drop != keptDocket {
			if drop == supersededDocket {
				res.Dockets.Superseded = append(res.Dockets.Superseded, old.ID)
			} else {
				res.Dockets.PartyDockets = append(res.Dockets.PartyDockets, old.ID)
			}
			if r, ok := src.Fields["motion_id"].(string); ok && src.Word == "motion_rule" && drop == partyDocket {
				reopened[rm.respelled(dockets.gapOf[r])] = true
			}
			if err := pinBefore(old.ID); err != nil {
				return nil, err
			}
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
		if bodies, err = mintsAboutCitations(bodies, cited); err != nil {
			res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
			continue
		}
		ts, err := record.ParseStamp(old.TS)
		if err != nil {
			res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
				Err: fmt.Sprintf("unparseable ts %q: %v", old.TS, err)})
			continue
		}
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
			// replayed (events.epoch, stamped by the write), which is the same fact from the record rather than from
			// the name — plans/roundless.md §III.A.2. A migrated record's round therefore means
			// what a fresh record's does.
			// A RE-RUN'S DRIFT IS THE PROGRAM'S BYTES, which are not the debate's wording: an anchor
			// token a program printed is what it printed, so it is kept through the respelling.
			repro, _ := body.(*recordpb.Reproduce)
			var derived *compute.Execution // the execution this reproduce is written naming, to store
			sourceNamed := false           // whether the source names the execution this reproduce compared against
			// The bytes are copied, not the pointer: the respelling sets a field through the
			// pointer it already holds, so a kept pointer would read back the respelled bytes.
			var drift *string
			if repro != nil && repro.Drift != nil {
				drift = proto.String(repro.GetDrift())
			}
			rm.apply(body)
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
				sourceNamed = repro.GetRerunSha() != ""
				derived = deriveExecution(repro, fromDir, computationOf, standing)
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
			if m, ok := body.(*recordpb.Mint); ok {
				if err := dropCrossSeatLink(batch.Querier(), m, seatID); err != nil {
					return nil, err
				}
			}
			// AN ENGAGEMENT THAT OWES NONE IS WRITTEN AS NO EVENT, and counted (UnowedEngagements):
			// the successors it engaged were named by a cross-seat lineage link dropCrossSeatLink drops,
			// so the record names none for it. A pin naming it names the record before it.
			if d, ok := body.(*recordpb.Dispatch); ok {
				if none, err := record.EngagementOwesNone(batch.Querier(), seatID, d); err != nil {
					return nil, err
				} else if none {
					res.UnowedEngagements = append(res.UnowedEngagements, old.ID)
					if err := pinBefore(old.ID); err != nil {
						return nil, err
					}
					continue
				}
			}
			// THE BENCH SIDES WITH A STATED POSITION, AND NOTHING ELSE (EventSchema 80, docketing.go): a
			// ruling of a docketing a Close ended first, a regrade of the bench's own, and a spot-check of
			// a gap a dropped bench act left open, and an inquiry check such a gap alone holds wanting, are
			// written as no event, each stated.
			if ended, err := endedRuling(batch.Querier(), body); err != nil {
				return nil, err
			} else if ended {
				res.Dockets.EndedRulings = append(res.Dockets.EndedRulings, old.ID)
				if err := pinBefore(old.ID); err != nil {
					return nil, err
				}
				continue
			}
			if benchRegrade(body, seatID) {
				res.Dockets.BenchRegrades = append(res.Dockets.BenchRegrades, old.ID)
				if err := pinBefore(old.ID); err != nil {
					return nil, err
				}
				continue
			}
			if drop, err := reopenedSpotCheck(batch.Querier(), body, reopened); err != nil {
				return nil, err
			} else if drop {
				res.Dockets.SpotChecks = append(res.Dockets.SpotChecks, old.ID)
				if err := pinBefore(old.ID); err != nil {
					return nil, err
				}
				continue
			}
			if drop, err := reopenedInquiryCheck(batch.Querier(), body, reopened); err != nil {
				return nil, err
			} else if drop {
				res.Dockets.InquiryChecks = append(res.Dockets.InquiryChecks, old.ID)
				if err := pinBefore(old.ID); err != nil {
					return nil, err
				}
				continue
			}
			if k, ok := body.(*recordpb.Docketing); ok {
				if err := docketingCause(batch.Querier(), k); err != nil {
					return nil, err
				}
			}
			if b, ok := body.(*recordpb.ReportPart); ok {
				res.parts = append(res.parts, b.GetText())
				if inOrder {
					b.Text = proto.String(report.Ordered(b.GetText()))
				}
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
			// is translated as a pin is, and refused where the act it names wrote no event. A read
			// point at a finding is the last event before it (pinBefore): the acts after it are news
			// to the seat as they were, and the finding, which the record no longer holds, is none.
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
			// refuses an id it cannot place; TranslatePlanKinds first respells its archived blocker kinds,
			// and TranslateReadingBlockers carries an archived reading blocker to its reading debt.
			if pl, ok := body.(*recordpb.Plan); ok {
				if err := TranslatePlanVerdict(pl); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				if err := TranslatePlanParties(pl); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				if err := TranslatePlanKinds(pl); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				dropped, err := TranslateReadingBlockers(pl, newID, batch.Querier())
				if err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				res.DroppedPlanBlockers += dropped
				if err := TranslatePlanIDs(pl, newID); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
				if err := CarryPlanPass(pl, batch.Querier()); err != nil {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: err.Error()})
					continue
				}
			}
			// A DROPPED LABEL'S OR FINDING'S ANCHOR IS WRITTEN AS NO EVENT, and a reading of the citation it placed
			// becomes a reading of the label's source.
			if a, ok := body.(*recordpb.Anchor); ok && rm.dropped[a.GetId()] {
				continue
			}
			// A READING WHOSE QUOTE WAS ONLY TOKENS THE FOLD TAKES OUT IS WRITTEN AS NO EVENT. A seat
			// that quoted nothing but a finding's token and a dropped label's (quadratic-formula: a
			// re-check of a citation stranded in the title line) quoted no report text; with both
			// tokens gone its quote says nothing, and no sentence of the report is equivalent to it —
			// the labelled reading's claim would attach the reading to a sentence it never named. The
			// reading is the residue, counted (QuotelessReadings); a pin naming it names the record
			// before it.
			if v, ok := body.(*recordpb.Verify); ok && v.Claim != nil && strings.TrimSpace(v.GetClaim()) == "" {
				res.QuotelessReadings = append(res.QuotelessReadings, old.ID)
				if err := pinBefore(old.ID); err != nil {
					return nil, err
				}
				continue
			}
			if v, ok := body.(*recordpb.Verify); ok && rm.dropped[v.GetAnchor()] {
				ls, known := labelled[v.GetAnchor()]
				if !known {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word,
						Err: fmt.Sprintf("this reading names citation %s, a label no labelled reading in the source minted", v.GetAnchor())})
					continue
				}
				v.Anchor, v.Url, v.Title = nil, proto.String(ls.url), proto.String(ls.title)
			}
			// A COMPUTATION WHOSE METHOD CITED A DROPPED LABEL CITES NOTHING: the citation it named is
			// written as no event, and a cites naming it is refused at the write, taking the
			// computation and every act on it down in a cascade. The method's source is the label's,
			// which no citation of the migrated record carries.
			if c, ok := body.(*recordpb.Computation); ok && rm.dropped[c.GetCites()] {
				c.Cites = nil
			}
			// A MINT'S class_material IS THE TOOL'S STAMP, and the write refuses a mint that arrives
			// with one. The archived stamp is taken off and the live rule stamps the translated class's
			// material default at the write. A fill is stated where the migrated stamp is not the
			// source's: every mint that predates the field (which of those moved is computable only
			// against a table this binary does not hold), and every mint whose archived stamp differs
			// from its class's default today. Both the fill and the translation count are recorded
			// once the write lands: a refused mint stamped nothing.
			var translatedSlug string
			var fill *migration.StatedFill
			if m, ok := body.(*recordpb.Mint); ok {
				slug, _ := src.Fields["class"].(string)
				if archivedGapClasses[slug] != "" {
					translatedSlug = slug
				}
				archived := m.ClassMaterial
				m.ClassMaterial = nil
				if m.Class != nil {
					stamp := recordpb.MaterialDefault(m.GetClass())
					switch {
					case archived == nil:
						fill = &migration.StatedFill{Where: "mint " + m.GetGapId(), Slug: slug,
							Value: recordpb.Word(stamp), Why: preFieldFillWhy}
					case *archived != stamp:
						fill = &migration.StatedFill{Where: "mint " + m.GetGapId(), Slug: slug,
							Value: recordpb.Word(stamp), Why: fmt.Sprintf(restampWhy, recordpb.Word(*archived))}
					}
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
			// AN EDIT OF A STRAY A FINDING'S TOKEN LEFT, WHOSE CHANGE ALREADY STANDS, IS WRITTEN AS NO
			// EVENT. A bare finding token standing against a terminator kept the terminator outside an
			// edit's span (bluedoc.settleAbuttingAnchors), and a later edit removed the stray it left
			// (m26, twice). With the token gone the first edit takes the terminator, so no stray stands
			// in the migrated report: the later edit's quote names text it never held, or finds the
			// sentence's own terminator and would delete it with the anchor on it. It is the residue,
			// counted (SettledEdits); a pin naming it names the report head before it.
			if settledEdit(batch.Querier(), src, body) {
				res.SettledEdits = append(res.SettledEdits, old.ID)
				if err := headBefore(old.ID); err != nil {
					return nil, err
				}
				continue
			}
			// AN ARCHIVED ACT THE DELIVERED-TEXT RULE REFUSES IS WRITTEN AS NO EVENT (gblock
			// 2026-11-14): an archived close or spot-check whose sitting
			// holds no delivery of report text stood on text the record never recorded reaching its
			// seat. Nothing is synthesized to make it stand; it is an accepted loss, listed by act, and
			// a pin or read point that named it names the last event written before it. The run's
			// stamped outcome is carried as history.
			if err := record.DeliveredTextRefusal(batch.Querier(), id.SeatID, body); err != nil {
				var u *record.UndeliveredError
				if !errors.As(err, &u) {
					return nil, err
				}
				res.LostActs = append(res.LostActs, migration.Refusal{OldID: old.ID, Word: old.Word, Err: u.Error()})
				res.AcceptedLosses[old.Word] = undeliveredLoss
				if err := lost.noteLoss(batch.Querier(), body, old.ID, old.Word); err != nil {
					return nil, err
				}
				if _, seen := newID[old.ID]; !seen {
					var last int64
					if err := batch.Querier().QueryRow(`SELECT COALESCE(MAX("id"), 0) FROM "events"`).Scan(&last); err != nil {
						return nil, fmt.Errorf("reading the last event written before source event %d: %w", old.ID, err)
					}
					newID[old.ID] = last
				}
				continue
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
					if _, _, err := batch.Write(record.Identity{Run: dst, SeatID: record.HarnessSeat}, a, ts); err != nil {
						res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: "admitting the sitting this close ends: " + err.Error()})
						continue
					}
					res.Out[admissionWord]++
				}
			}
			ev, actID, err := batch.Write(id, body, ts)
			if err != nil {
				gap, from, dep, derr := lost.dependsOn(batch.Querier(), body)
				if derr != nil {
					return nil, derr
				}
				if !dep {
					res.Refusals = append(res.Refusals, migration.Refusal{OldID: old.ID, Word: old.Word, Err: refused(err)})
					continue
				}
				res.LostActs = append(res.LostActs, migration.Refusal{OldID: old.ID, Word: old.Word, Err: dependentLoss(gap, from, err)})
				res.AcceptedLosses[old.Word] = dependentLossWhy
				if err := lost.noteLoss(batch.Querier(), body, old.ID, old.Word); err != nil {
					return nil, err
				}
				if _, seen := newID[old.ID]; !seen {
					var last int64
					if err := batch.Querier().QueryRow(`SELECT COALESCE(MAX("id"), 0) FROM "events"`).Scan(&last); err != nil {
						return nil, fmt.Errorf("reading the last event written before source event %d: %w", old.ID, err)
					}
					newID[old.ID] = last
				}
				continue
			}
			res.Out[wordOf(ev)]++
			// THE ROWS THE WRITE PATH WROTE BESIDE THE ACT (record.writeTombstones): every event above the
			// act's own id is one, written in its savepoint.
			var rows int
			if err := batch.Querier().QueryRow(`SELECT count(*) FROM "events" WHERE "id" > ?`, actID).Scan(&rows); err != nil {
				return nil, fmt.Errorf("counting the tombstone rows written after source event %d: %w", old.ID, err)
			}
			if rows > 0 {
				res.Out[recordpb.Word(recordpb.EventType_EVENT_TYPE_TOMBSTONE)] += rows
			}
			if translatedSlug != "" {
				classMints[translatedSlug]++
			}
			if fill != nil {
				res.StatedFills = append(res.StatedFills, *fill)
			}
			if p, ok := body.(*recordpb.Computation); ok && p.GetBundleSha() != "" {
				computationAnchors[p.GetBundleSha()] = append(computationAnchors[p.GetBundleSha()], p.GetComputationId())
				mc := migratedComputation{id: p.GetComputationId(), bundle: p.GetBundleSha(), argv: p.GetArgv(), exit: int(p.GetExit())}
				if old, archived := archivedSha(src); archived {
					mc.oldSha = old
					bundleOf[old] = p.GetBundleSha()
					if out, inline := src.Fields["output"].(string); inline {
						mc.output = &out
					}
				}
				res.computations = append(res.computations, mc)
				computationOf[mc.id] = mc
			}
			if r, ok := body.(*recordpb.Reproduce); ok {
				if derived != nil {
					res.derived[r.GetBundleSha()] = append(res.derived[r.GetBundleSha()], *derived)
				}
				if r.GetRan() {
					standing[r.GetBundleSha()] = r
				}
				if sourceNamed {
					res.reruns[r.GetBundleSha()] = append(res.reruns[r.GetBundleSha()], r.GetRerunSha())
				}
			}
			if _, seen := newID[old.ID]; !seen {
				// THE ACT'S OWN ID, as the write returns it: after a placing or ending act the newest
				// event is the tombstone row the write path wrote beside it (record.writeTombstones),
				// which no pin, plan id or reading ref names.
				wrote = actID
				// Every link of a chain is the one act written here.
				for _, k := range links[1:] {
					newID[byKey[k].ID] = wrote
				}
				newID[old.ID] = wrote
			}
			if anchorAfter != nil {
				// The act landed whether or not its Anchor does, so what follows still pairs with it.
				if _, _, err := batch.Write(record.Identity{Run: dst, SeatID: seatID}, anchorAfter, ts); err != nil {
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
	res.QuotelessReadings, res.SettledEdits = statedIDs(res.QuotelessReadings), statedIDs(res.SettledEdits)
	res.UnowedEngagements = statedIDs(res.UnowedEngagements)
	d := &res.Dockets
	d.Opinions, d.PartyDockets, d.Superseded = statedIDs(d.Opinions), statedIDs(d.PartyDockets), statedIDs(d.Superseded)
	d.EndedRulings, d.BenchRegrades, d.SpotChecks, d.InquiryChecks = statedIDs(d.EndedRulings), statedIDs(d.BenchRegrades), statedIDs(d.SpotChecks), statedIDs(d.InquiryChecks)
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
// The cast is written through batch, under dst as its Identity's run, stamped at.
func writeSynthesizedCast(evs []OldEvent, rm *remap, batch *record.Replay, dst record.Run, at time.Time) (bool, error) {
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
	_, _, err := batch.Write(record.Identity{Run: dst, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast}, at)
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

// refusedArgument reports whether body is a closing or a position the write path refuses from
// seatID at this place (record.ClosingRefusal, record.PositionRefusal) — one replay writes as no
// event. A read that fails is returned, never taken for a refusal.
func refusedArgument(q recordsql.Querier, body proto.Message, seatID string) (bool, error) {
	var err error
	switch b := body.(type) {
	case *recordpb.Closing:
		err = record.ClosingRefusal(q, seatID, b)
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

// settledEdit reports whether body is an edit whose change the migrated report already holds, because
// the text it changed was a stray a finding's token left. It is one of two shapes, both measured on
// m26:
//   - its archived quote held a finding's token, that quote no longer locates in the report as the
//     migrated record renders it, and its replacement already locates there;
//   - it only deletes leading punctuation, and the punctuation its quote locates stands against an
//     anchor the edit does not name. A stray terminator stood after a finding's token; with the token
//     gone the quote finds the sentence's own terminator instead, and replaying it would delete that
//     and the anchor on it.
//
// An exact-span edit and a run with no report yet are never settled: their misses stay the write's
// and the render's to refuse.
func settledEdit(q recordsql.Querier, src OldEvent, body proto.Message) bool {
	b, ok := body.(*recordpb.BlueEdit)
	if !ok || b.GetExactSpan() {
		return false
	}
	text, err := record.RenderAt(q)
	if err != nil {
		return false
	}
	start, end, err := bluedoc.LocateUniqueReplacing(text, b.GetOld())
	if err == nil {
		return strayTerminator(b.GetOld(), b.GetNew(), text[start:end])
	}
	if archived, _ := src.Fields["old"].(string); !strings.Contains(archived, "<!--"+findingTag+":") {
		return false
	}
	_, _, err = bluedoc.LocateUniqueReplacing(text, b.GetNew())
	return err == nil
}

// strayTerminator reports whether an edit from old to new only deletes old's leading punctuation, and
// span — where old locates — holds an anchor neither old nor new names.
func strayTerminator(old, replacement, span string) bool {
	if replacement == old || strings.TrimLeft(old, anchortext.TrailingPunct) != replacement {
		return false
	}
	for _, id := range anchor.IDs(span) {
		if !strings.Contains(old, id) && !strings.Contains(replacement, id) {
			return true
		}
	}
	return false
}

// labelSource is a dropped label's source: the url and title its labelled reading read.
type labelSource struct{ url, title string }

// labelSources reads every citation label an archived reading of a source its lens found minted off
// the source stream, records each on rm to drop (remap.dropLabels), and returns each label's source,
// keyed by both spellings, for the readings that named the citation it placed.
func labelSources(evs []OldEvent, rm *remap) map[string]labelSource {
	out := map[string]labelSource{}
	var labels []string
	for _, old := range evs {
		l, _ := old.Fields["label"].(string)
		if old.Word != "verify" || l == "" {
			continue
		}
		url, _ := old.Fields["url"].(string)
		title, _ := old.Fields["title"].(string)
		labels = append(labels, l)
		out[l] = labelSource{url: url, title: title}
		out[rm.spelling("citation", l)] = out[l]
	}
	rm.dropLabels(labels)
	return out
}

// findingNames is every name an archived finding carries — its id, its archived id and its label —
// as the source spells each, read off the source stream.
func findingNames(evs []OldEvent) []string {
	var out []string
	for _, old := range evs {
		if old.Word != "finding" {
			continue
		}
		for _, col := range []string{"id", "finding_id", "label"} {
			if n, _ := old.Fields[col].(string); n != "" {
				out = append(out, n)
			}
		}
	}
	return out
}

// citedInActMints is each gap an archived reading of a CITATION minted in its own act, with that
// citation: the Verify whose carried_by names the gap and whose carrier_minted is set, naming an
// anchor (EventSchema 55–70). Read from the record's fields, never from adjacency, and keyed on the
// source's spelling of the gap: the Mint it translates is read before the remap respells it. A
// reading of a source the lens found names no anchor and is not here: its Mint already quotes the
// sentence, which is the live mint of that subject.
func citedInActMints(evs []OldEvent) map[string]string {
	out := map[string]string{}
	for _, old := range evs {
		g, _ := old.Fields["carried_by"].(string)
		a, _ := old.Fields["anchor"].(string)
		if old.Word == "verify" && g != "" && a != "" && truthy(old.Fields["carrier_minted"]) {
			out[g] = a
		}
	}
	return out
}

// mintsAboutCitations is mintAboutCitation over an act's bodies; a refusal is the act's.
func mintsAboutCitations(bodies []proto.Message, cited map[string]string) ([]proto.Message, error) {
	for _, b := range bodies {
		if m, ok := b.(*recordpb.Mint); ok {
			if err := mintAboutCitation(m, cited); err != nil {
				return nil, err
			}
		}
	}
	return bodies, nil
}

// mintAboutCitation is the one translation of the in-act model: a Mint whose own act's reading
// named a citation comes forward ABOUT that citation — about_kind anchor, about_ref the reading's
// anchor, no quote — the subject a live mint of its class names (record.readingMintedClasses). Its
// Anchor row stays: an Anchor is a report op, and the token it placed at the citation's sentence is
// what every later edit's `old` carries. Both values are the source's; nothing is filled. A Mint that
// already names a subject (about_kind set) is REFUSED, never overwritten: the archive then names two
// subjects for one gap, and the tool makes neither up.
func mintAboutCitation(m *recordpb.Mint, cited map[string]string) error {
	a, ok := cited[m.GetGapId()]
	if !ok {
		return nil
	}
	if m.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_UNSPECIFIED {
		return fmt.Errorf("gap %s was minted in the act of a reading of citation %s and is about %s %s already: the archive names two subjects for one gap, and the translation writes neither", m.GetGapId(), a, recordpb.Word(m.GetAboutKind()), m.GetAboutRef())
	}
	m.AboutKind, m.AboutRef, m.Location = recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), proto.String(a), nil
	return nil
}

// truthy reads a stored bool: SQLite's 0/1, or a bool.
func truthy(v any) bool {
	switch b := v.(type) {
	case int64:
		return b != 0
	case bool:
		return b
	}
	return false
}

// statedIDs is ids, or an empty list where it holds none: the manifest states a zero rather than
// leaving it absent.
func statedIDs(ids []int64) []int64 {
	if ids == nil {
		return []int64{}
	}
	return ids
}
