package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHAT HOLDS A PASS IS DECIDED HERE, ONCE.
//
// Four readers each worked it out for themselves — the verdict gate, the chair's work list,
// `dispatch next`'s pass_permitted and the FAIL convergence refusal — and they disagreed: the plan
// said pass_permitted over an unruled petition, an unanswered contradiction and a docket motion on
// a gap that had since closed, and the gate refused the PASS the chair then tried. On a converged
// board with an unruled petition the PASS refusal said "issue `--as FAIL`" and the FAIL refusal
// said "or issue `--as PASS`", so the chair could record neither (#1202). Each copy had been
// realigned by hand before (B9, 6a67e0bf), and the next rule added reached only some of them.
//
// So every reader calls passBlockersOf, and words the blockers it returns from the kind table
// below: the gate's refusal teaches the verb, the work list names the item, the plan says why it
// is not permitted. A new reason a PASS is held is a row here, and TestEveryPassBlockerKind…
// fails until a fixture produces it and every surface agrees about it.

// BlockerKind is one reason the PASS gate refuses.
type BlockerKind string

const (
	BlockerStrandedGap   BlockerKind = "stranded_gap"
	BlockerMaterialGap   BlockerKind = "material_gap"
	BlockerContradiction BlockerKind = "contradiction"
	BlockerNoReport      BlockerKind = "no_report"
	BlockerLensReady     BlockerKind = "lens_ready"
	BlockerStaleArea     BlockerKind = "stale_area"
	BlockerUnruledMotion BlockerKind = "unruled_motion"
	BlockerAvenueReview  BlockerKind = "avenue_review"
)

// Blocker is one thing on the board holding the PASS gate, and the seat whose act clears it.
type Blocker struct {
	Kind BlockerKind
	// Subject is what holds it: a gap id, a claim, a lens seat, a motion id — "" for the kinds that
	// are about the run as a whole (no_report, avenue_review).
	Subject string
	// Owner is the seat whose act clears it — the gap's minter, the lens that read the
	// contradicting source, the ready lens, the motion's gavel-holder, the chair for its own
	// reads — and "" where no seat on the record can.
	Owner string
	// Detail is what the wording needs beyond the subject: a stranded gap's successor, a motion's
	// subject and gavel, a ready lens's reason, a stale area's pin.
	Detail string
	// About is the gap an unruled motion is about, "" for a motion about no gap.
	About string
	// foldWhy is a ready lens's reason in the retirement fold's own words, for the refusal.
	foldWhy string
	// motionSubject is an unruled motion's subject — which sitting dispatch convenes its gavel for.
	motionSubject string
	// since is the events.id the blocker arose at — a motion's filing, the contradicting verify —
	// so dispatch can ask whether its owner has sat since; 0 for the kinds dispatch does not ready.
	since int64
}

// Every reports whether the blocker holds every verdict, not only a PASS.
func (b Blocker) Every() bool { return kindOf(b.Kind).every }

// ChairOwned reports whether the chair clears it in its own sitting.
func (b Blocker) ChairOwned() bool { return b.Owner == chairSeat }

// blockerKind is one row of the table every surface words a blocker from.
type blockerKind struct {
	kind BlockerKind
	// every: it holds every verdict, not only a PASS. Under a migration only these refuse.
	every bool
	// gloss is the help page's phrase for the kind.
	gloss string
	// item is the chair's work-list sentence for one blocker.
	item func(Blocker) string
	// refusal is the gate's paragraph for every blocker of the kind, after "verdict … refused — ".
	refusal func([]Blocker) string
	// why is the plan's reason for one blocker, or nil where a readiness source already states it.
	why func(Blocker) string
}

// ownerPhrase names the seat that clears a blocker, as the subject of a sentence: "its minter,
// red-lens-logic," — or the fallback where the record names no seat.
func ownerPhrase(b Blocker, role, fallback string) string {
	if b.Owner == "" {
		return fallback
	}
	return "its " + role + ", " + b.Owner + ","
}

// ownerOr is the seat that clears a blocker, bare, or the fallback where the record names no seat —
// the same fallback the kind's other wordings use, so no surface prints an empty name.
func ownerOr(b Blocker, fallback string) string {
	if b.Owner == "" {
		return fallback
	}
	return b.Owner
}

// contradictionReader is who raises a contradiction when the record names no reader.
const contradictionReader = "a lens"

// planHold is the plan's verdict on a blocker, by its owner: the chair's own is cleared in its
// sitting and leaves pass_permitted standing, so its line must not say a PASS is not permitted;
// another seat's holds pass_permitted false, and the line says so.
func planHold(b Blocker, notPermitted string) string {
	if b.ChairOwned() {
		return "yours to clear before a PASS"
	}
	return notPermitted
}

func subjects(bs []Blocker) []string {
	out := make([]string, 0, len(bs))
	for _, b := range bs {
		out = append(out, b.Subject)
	}
	return out
}

// blockerKinds is the table, in the order the work list and the refusal state them.
var blockerKinds = []blockerKind{
	{
		kind: BlockerStrandedGap, every: true,
		gloss: "a superseded gap still open (this holds a FAIL too)",
		item: func(b Blocker) string {
			return fmt.Sprintf("gap %s is open and superseded by %s — every verdict is refused while it is; %s closes it naming its successor", b.Subject, b.Detail, ownerPhrase(b, "minter", "the lens that minted it"))
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, fmt.Sprintf("%s (superseded by %s)", b.Subject, b.Detail))
			}
			return fmt.Sprintf("%d superseded gap(s) are still OPEN: %s. Superseding is a promise to replace, and an ancestor left open is the same defect counted twice: the run then reports more open gaps than it has distinct ones. The lens that minted each closes it with the lens's `close --superseded-by` naming its replacement, or if it is genuinely still live, it was not superseded",
				len(bs), strings.Join(named, ", "))
		},
	},
	{
		kind:  BlockerMaterialGap,
		gloss: "an open material gap (material by its class, else graded medium or above)",
		item: func(b Blocker) string {
			return fmt.Sprintf("gap %s is open and material — PASS is refused while it is; %s closes it", b.Subject, ownerPhrase(b, "minter", "the lens that minted it"))
		},
		refusal: func(bs []Blocker) string {
			return fmt.Sprintf("%d material gap(s) still OPEN: %s. PASS requires every material gap resolved, and a gap is closed only by the lens that minted it, with the lens's `close --id <id> --as repaired|defect_accepted|not_a_defect|defect_owed_elsewhere`. PASS waits for those closures; `--as FAIL` does not",
				len(bs), strings.Join(subjects(bs), ", "))
		},
	},
	{
		kind:  BlockerContradiction,
		gloss: "a contradicting source no finding raises",
		item: func(b Blocker) string {
			return fmt.Sprintf("red read a source that contradicts or does not support the claim %q and no finding raises it — PASS is refused until one does; %s raises it", b.Subject, ownerPhrase(b, "reader", contradictionReader))
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				if b.Owner == "" {
					named = append(named, fmt.Sprintf("%q", b.Subject))
					continue
				}
				named = append(named, fmt.Sprintf("%q (read by %s)", b.Subject, b.Owner))
			}
			return fmt.Sprintf("red read a source that CONTRADICTS or does not support %d claim(s), and no finding was ever raised about them:\n  %s\n"+
				"Each is red's own reading that the report says something its source does not. The lens that read it raises it with the lens's `finding --quote \"<the claim>\" --reason \"<what the source actually says>\"`, graded on every axis, so it enters the board with the lifecycle, the blue duty and the gate every other defect has. "+
				"Read them with `show evidence`. A PASS claims the report is sound; these say otherwise on the record. Raising them is a lens's act; PASS waits for it, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("a source %s read contradicts or does not support %q and no finding raises it — %s", ownerOr(b, contradictionReader), b.Subject, planHold(b, "PASS is not permitted until one does"))
		},
	},
	{
		kind:  BlockerNoReport,
		gloss: "no ingested report",
		item:  func(Blocker) string { return "no report has been ingested — PASS is refused" },
		refusal: func([]Blocker) string {
			return "no report has been ingested or edited, so there is nothing a lens could have audited"
		},
	},
	{
		kind:  BlockerLensReady,
		gloss: "a ready lens",
		item: func(b Blocker) string {
			return fmt.Sprintf("lens %s is ready (%s) — PASS is refused while it is", b.Subject, b.Detail)
		},
		refusal: func(bs []Blocker) string {
			var why []string
			for _, b := range bs {
				why = append(why, b.foldWhy)
			}
			return fmt.Sprintf("%d cast lens(es) are ready: %s. A lens is ready while it is active (it minted fresh material within its last two sittings) or retired with its one re-arm owed; `dispatch next` readies them",
				len(bs), strings.Join(why, "; "))
		},
	},
	{
		kind:  BlockerStaleArea,
		gloss: "a stale area no spot-check this sitting names",
		item: func(b Blocker) string {
			return fmt.Sprintf("area %s is behind its pin %s — PASS is refused until a spot-check this sitting names it", b.Subject, b.Detail)
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, fmt.Sprintf("%s (pin %s)", b.Subject, b.Detail))
			}
			return fmt.Sprintf("the report changed behind %d lens(es) retired for good: %s. Read the changes since each pin against that area's duties, and name the areas in a spot-check this sitting (--areas) before a PASS; a defect you find there goes in that spot-check",
				len(bs), strings.Join(named, ", "))
		},
	},
	{
		kind:  BlockerUnruledMotion,
		gloss: "an unruled motion",
		item: func(b Blocker) string {
			return "motion " + b.Subject + " (" + b.Detail + ") was filed and never ruled — PASS is refused while it stands"
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, b.Subject+" ("+b.Detail+")")
			}
			return fmt.Sprintf("%d motion(s) filed and never ruled: %s. "+
				"Read what each one asks with `inquest motions` (its `basis` is the filer's argument, which your ruling answers), "+
				"then rule it with `motion <subject> rule --id <id> --as <verdict> --reason \"...\"` — IF THE GAVEL NAMED ABOVE IS YOURS. "+
				"Where it is not, the ruling is not yours to make and not yours to wait for silently: issue `--as FAIL` so the sitting ends on the record and the seat that holds it can answer. "+
				"A motion is answered before the debate moves on, so a PASS over an unanswered ask claims a settlement that did not happen",
				len(bs), strings.Join(named, ", "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s: motion (%s) filed and never ruled — %s", b.Subject, b.Detail, planHold(b, "PASS is not permitted while it stands"))
		},
	},
	// THE AVENUES ARE READ ONCE, EVERY EPOCH. One statement per epoch, not one per line: presence is
	// not the question, because the lines are generated onto the page from the record. What the
	// read owes is a judgement on whether the BODY delivered them, and where it did not, a gap. The
	// report is regenerated each epoch, so a review recorded before this epoch's edits answers a
	// question about a document that no longer exists.
	{
		kind:  BlockerAvenueReview,
		gloss: "no avenue review this epoch",
		item: func(Blocker) string {
			return "the report's account of its own research has not been read this epoch — PASS is refused until one `avenue review` says what the read found (and any shortfall is minted as a gap)"
		},
		refusal: func([]Blocker) string {
			return "this epoch has no avenue review. " +
				"READ THE REPORT ONCE (`show report`), list what the record claims this run investigated with " +
				"`show avenues`, and answer in one act: `avenue review --reason \"<what the report " +
				"says at those avenues>\"`. Where an avenue's research is thin, missing or unsupported by the text, " +
				"MINT A GAP for it — the shortfall is an ordinary defect and gets the ordinary lifecycle; this " +
				"event only records that the read happened, because an absent review reads exactly like a sound " +
				"one. A PASS claims the report is sound, and its account of what this run investigated is part " +
				"of the report; record the review, or issue `--as FAIL`"
		},
		why: func(b Blocker) string {
			return "this epoch has no avenue review — " + planHold(b, "PASS is not permitted until one is recorded")
		},
	},
}

func kindOf(k BlockerKind) blockerKind {
	for _, r := range blockerKinds {
		if r.kind == k {
			return r
		}
	}
	panic("record: blocker kind " + string(k) + " has no row in blockerKinds")
}

// WorkItem is the chair's work-list sentence for the blocker.
func (b Blocker) WorkItem() string { return kindOf(b.Kind).item(b) }

// PassBlockerGlosses is each kind's help phrase, in table order — what `verdict --help` lists.
func PassBlockerGlosses() []string {
	out := make([]string, 0, len(blockerKinds))
	for _, r := range blockerKinds {
		out = append(out, r.gloss)
	}
	return out
}

// blockerGap is an open gap as the list reads it: the gap view's columns, off whichever reader
// the caller already holds (the work path's states, the plan's open gaps).
type blockerGap struct {
	id, mintedBy, supersededBy string
	material                   bool
}

func blockerGapsOfStates(gaps []WorkGapState) []blockerGap {
	var out []blockerGap
	for _, g := range gaps {
		if !g.Open {
			continue
		}
		sup := ""
		if g.Stranded {
			sup = g.SupersededBy
		}
		out = append(out, blockerGap{id: g.ID, mintedBy: g.MintedBy, supersededBy: sup, material: g.Material})
	}
	return out
}

func blockerGapsOfOpen(gaps []openGap) []blockerGap {
	out := make([]blockerGap, 0, len(gaps))
	for _, g := range gaps {
		out = append(out, blockerGap{id: g.id, mintedBy: g.mintedBy, supersededBy: g.supersededBy, material: g.material})
	}
	return out
}

// passBlockersOf is THE answer to "what holds a PASS", off the stream, its row ids, the open gaps
// and the fresh-material set. The order is the kind table's, and within a kind the record's.
func passBlockersOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap, fresh map[string]bool) []Blocker {
	var out []Blocker
	for _, g := range gaps {
		if g.supersededBy != "" {
			out = append(out, Blocker{Kind: BlockerStrandedGap, Subject: g.id, Owner: g.mintedBy, Detail: g.supersededBy})
		}
	}
	for _, g := range gaps {
		if g.supersededBy == "" && g.material {
			out = append(out, Blocker{Kind: BlockerMaterialGap, Subject: g.id, Owner: g.mintedBy})
		}
	}
	// ids IS ALIGNED WITH evs — every caller passes the events' ids or their stream positions, the
	// rule dispatchLedger and benchRegisters read places by — so a blocker's `since` and the
	// sitting it is compared with are places in one sequence.
	for _, c := range unansweredContradictionsBy(evs) {
		out = append(out, Blocker{Kind: BlockerContradiction, Subject: c.claim, Owner: c.reader, since: ids[c.at]})
	}
	if lg := passLensGateOf(evs, ids, win, fresh); lg.cast {
		if lg.head == 0 {
			out = append(out, Blocker{Kind: BlockerNoReport})
		} else {
			for _, f := range lg.ready() {
				out = append(out, Blocker{Kind: BlockerLensReady, Subject: f.seat, Owner: f.seat, Detail: readyReason(f), foldWhy: f.why})
			}
			for _, a := range lg.uncovered() {
				out = append(out, Blocker{Kind: BlockerStaleArea, Subject: a.SeatID, Owner: chairSeat, Detail: fmt.Sprint(a.Pin)})
			}
		}
	}
	out = append(out, unruledMotionBlockers(evs, ids, win)...)
	if AvenueReviewDueOf(evs, win) {
		out = append(out, Blocker{Kind: BlockerAvenueReview, Owner: chairSeat})
	}
	return out
}

// unruledMotionBlockers is the gate list's unruled-motion arm: every motion with no ruling, owned by
// the seat whose gavel its subject is. ids is evs's places, aligned with it.
func unruledMotionBlockers(evs []*Event, ids []int64, win WindowIndex) []Blocker {
	var out []Blocker
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Ruled() {
			continue
		}
		g, err := gavelOf(m.Subject)
		if err != nil {
			// A SCHEMA DEFECT MUST NOT SILENTLY SHORTEN THE LIST. The motion still holds the gate,
			// owned by no seat, and the wording says what could not be resolved.
			g = gavel{phrase: m.Subject + ", and this binary cannot say who rules it: " + err.Error()}
		}
		out = append(out, Blocker{Kind: BlockerUnruledMotion, Subject: m.ID, Owner: g.seat, Detail: g.phrase, About: m.GapID,
			motionSubject: m.Subject, since: m.filed})
	}
	return out
}

// MotionBlockersOf is the unruled-motion arm of the gate's list over a stream, as the plan relays
// it. Capture holds a relayed plan's motion blockers to it, read off the record as it stood at the
// end of the chair sitting that relayed the plan.
func MotionBlockersOf(evs []*Event, win WindowIndex) []PlanBlocker {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	var out []PlanBlocker
	for _, b := range unruledMotionBlockers(evs, seq, win) {
		out = append(out, PlanBlocker{Kind: b.Kind, Subject: b.Subject, Owner: b.Owner})
	}
	return out
}

// PassBlockers is passBlockersOf over the run, read as the gate reads it.
func PassBlockers(run Run) ([]Blocker, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	evs, win, err := eventsAt(db)
	if err != nil {
		return nil, err
	}
	fresh, err := freshMaterialOf(db)
	if err != nil {
		return nil, err
	}
	gaps, err := openGaps(db)
	if err != nil {
		return nil, err
	}
	return passBlockersOf(evs, win.IDs(evs), win, blockerGapsOfOpen(gaps), fresh), nil
}

// requireNoBlockers is the verdict gate. A PASS is refused over every blocker; any verdict, and
// any verdict a migration replays, over the kinds that hold every verdict. A live FAIL is then
// held to the convergence rule, which stands only while nothing another seat must act on holds
// the PASS — so its "or issue `--as PASS`" is true apart from the chair's own items.
//
// NOT UNDER A MIGRATION for the PASS-only kinds (gblock, 2026-09-11). Migrate translates history;
// it does not re-judge an archived PASS under a rule that did not exist when the PASS was issued.
// The archived b7 and b9 runs' PASSes stand over gaps their classes now make material, and
// refusing them would drop a real event and call the loss a translation. The exemption is not
// silent: stampMigrationAdmission records those gaps on the PASS, and verify reads them there.
func requireNoBlockers(run Run, verdict recordpb.Verdict, blockers []Blocker) error {
	var holding []Blocker
	for _, b := range blockers {
		if b.Every() || (verdict == recordpb.Verdict_VERDICT_PASS && !Migrating) {
			holding = append(holding, b)
		}
	}
	if len(holding) > 0 {
		return blockerRefusal(holding)
	}
	if verdict == recordpb.Verdict_VERDICT_FAIL && !Migrating {
		return requireFailIsNotConvergent(run, blockers)
	}
	return nil
}

// requireBenchMotionsRuled refuses the run's outcome while a motion whose gavel is the bench's
// stands unruled. The terminal bench sitting exists to rule them, and it fires on a count the chair
// reports, which nothing audits: an under-report skipped the one sitting that could answer the
// motion, and the outcome recorded over it in silence (#1202). Refused here, the miss is loud at
// the one write path, and the bench recording the outcome holds the gavel, so it rules the motion
// and records the outcome after.
//
// A HALT IS EXEMPT. It is the bench's own decision to end the run, taken with the record in view;
// refusing it would force rulings after the bench has already stopped the run.
func requireBenchMotionsRuled(run Run) error {
	halted, err := recordHas(run, `SELECT 1 FROM "halt" LIMIT 1`)
	if err != nil || halted {
		return err
	}
	blockers, err := PassBlockers(run)
	if err != nil {
		return err
	}
	var named []string
	for _, b := range blockers {
		if b.Kind != BlockerUnruledMotion || b.Owner != benchSeat {
			continue
		}
		named = append(named, b.Subject+" ("+b.Detail+")")
	}
	if len(named) == 0 {
		return nil
	}
	return fmt.Errorf("record: outcome refused — %d motion(s) whose gavel is the bench's stand unruled: %s. "+
		"The outcome is the run's last act, and a run that ends over an ask the bench never answered records a settlement that did not happen. "+
		"You hold the gavel: read what each one asks with `inquest motions`, rule it with `motion <subject> rule --id <id> --as <verdict> --reason \"...\"`, then record the outcome",
		len(named), strings.Join(named, ", "))
}

// blockerRefusal words the holding blockers, every one, grouped by kind in table order with each
// kind's paragraph once.
func blockerRefusal(holding []Blocker) error {
	every := true
	for _, b := range holding {
		every = every && b.Every()
	}
	head := "record: verdict PASS refused — "
	if every {
		head = "record: verdict refused — "
	}
	var paras []string
	for _, r := range blockerKinds {
		var of []Blocker
		for _, b := range holding {
			if b.Kind == r.kind {
				of = append(of, b)
			}
		}
		if len(of) > 0 {
			paras = append(paras, r.refusal(of))
		}
	}
	if len(paras) == 1 {
		return fmt.Errorf("%s%s", head, paras[0])
	}
	for i := range paras {
		paras[i] = fmt.Sprintf("%d. %s", i+1, paras[i])
	}
	return fmt.Errorf("%s%d things hold it, in %d kinds, each with what clears it:\n\n%s", head, len(holding), len(paras), strings.Join(paras, "\n\n"))
}
