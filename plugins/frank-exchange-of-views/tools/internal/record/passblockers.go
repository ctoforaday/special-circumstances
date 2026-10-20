package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
	BlockerReopened      BlockerKind = "reopened"
	BlockerNoReport      BlockerKind = "no_report"
	BlockerLensReady     BlockerKind = "lens_ready"
	BlockerStaleArea     BlockerKind = "stale_area"
	BlockerUnruledMotion BlockerKind = "unruled_motion"
	BlockerInquiryReview BlockerKind = "inquiry_review"
)

// Blocker is one thing on the board holding the PASS gate, and the seat whose act clears it.
type Blocker struct {
	Kind BlockerKind
	// Subject is what holds it: a gap id, a contradiction's id, a lens seat, a motion id — "" for
	// the kinds that are about the run as a whole (no_report, inquiry_review).
	Subject string
	// Owner is the seat whose act clears it — the gap's minter, the lens that read the
	// contradicting source, the ready lens, the motion's gavel-holder, the chair for its own
	// reads — and "" where no seat on the record can.
	Owner string
	// Detail is what the wording needs beyond the subject: a stranded gap's successor, a motion's
	// subject and gavel, a stale area's pin, a reopened proof's sha256 (""
	// for a reopened citation), a contradiction's claim.
	Detail string
	// About is the gap an unruled motion is about, "" for a motion about no gap.
	About string
	// foldWhy is a ready lens's reason in the lens fold's own words, for the refusal.
	foldWhy string
	// motionSubject is an unruled motion's subject — which sitting dispatch convenes its gavel for.
	motionSubject string
	// since is the events.id the blocker arose at — a motion's filing, the contradicting verify,
	// the edit that reopened a citation or proof — so dispatch can ask whether its owner has sat
	// since, and a reopened anchor's wording can name its edit; 0 for the other kinds.
	since int64
	// lastCheck is the events.id of a reopened anchor's latest check, which does not clear it.
	lastCheck int64
	// raisedBy is the finding that raised a contradiction, "" until one has.
	raisedBy string
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
	// refusal is the gate's paragraph for every blocker of the kind, after the gate's head ("PASS refused — ").
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

// contradictionCitationAnswer is what else answers a contradiction whose reading checked a
// citation (unansweredContradictions), stated by the work item and the PASS refusal alike.
const contradictionCitationAnswer = "A contradiction raised by reading a citation is answered as well by a retire of that citation, or once a lens has verified the citation's sentence as it now stands and every lens's latest verify of it backs it — the reading stops counting when its lens reads the citation's source again or an edit changes the sentence after its sitting opened, and another lens's backing verify while it counts is a disagreement that answers nothing"

// contradictionFact is a contradiction's name on every surface: the word and its id, which is the
// `--about` value that answers it. A bare decimal would read as a count.
func contradictionFact(b Blocker) string { return "contradiction " + b.Subject }

// reopenedFact is the one sentence every wording of a reopened blocker states: which anchor, the
// edit that changed its sentence, and its latest check — all three as the record
// holds them, so a reader can go to each.
func reopenedFact(b Blocker) string {
	return fmt.Sprintf("%s's sentence changed at edit %d and no sitting opened since has checked it (its last check is event %d)", b.Subject, b.since, b.lastCheck)
}

// reopenedOwed is the other sentence every wording of a reopened blocker states: whose the check
// is. Its owner owes it (reopenedOwnerOf) — dispatch readies that seat and its work list blocks on
// the anchor — and it is not the only seat whose check clears: the fold reads no seat.
func reopenedOwed(b Blocker) string {
	return ownerOr(b, reopenedOwner) + " owes that check and is the seat readied for it; another lens's check from such a sitting clears it as well"
}

// reopenedOwner is who owes a reopened anchor's check when the record names no seat.
const reopenedOwner = "the lens that owns it"

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
		gloss: "a contradicting source no gap answers",
		item: func(b Blocker) string {
			if b.raisedBy != "" {
				return fmt.Sprintf("%s: red read a source that contradicts or does not support the claim %q; finding %s raised it, and no gap answers it — PASS is refused until one does; %s mints the gap with `mint --found-by %s`. A source the report misstates is a citation defect, so its gap is material by its class whatever reader decision you judge it to change. %s", contradictionFact(b), b.Detail, b.raisedBy, ownerPhrase(b, "reader", contradictionReader), b.raisedBy, contradictionCitationAnswer)
			}
			return fmt.Sprintf("%s: red read a source that contradicts or does not support the claim %q and it is unanswered — PASS is refused while it is; %s raises it with `finding --about-kind contradiction --about %s`, whether or not the sentence still stands, and answers it with the gap minted from that finding (`mint --found-by <the finding's id>`). A source the report misstates is a citation defect, so its gap is material by its class whatever reader decision you judge it to change. %s", contradictionFact(b), b.Detail, ownerPhrase(b, "reader", contradictionReader), b.Subject, contradictionCitationAnswer)
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				if b.Owner == "" {
					named = append(named, fmt.Sprintf("%s: %q", contradictionFact(b), b.Detail))
					continue
				}
				if b.raisedBy != "" {
					named = append(named, fmt.Sprintf("%s: %q (read by %s, raised by finding %s, no gap minted from it)", contradictionFact(b), b.Detail, b.Owner, b.raisedBy))
					continue
				}
				named = append(named, fmt.Sprintf("%s: %q (read by %s)", contradictionFact(b), b.Detail, b.Owner))
			}
			return fmt.Sprintf("red read a source that CONTRADICTS or does not support %d claim(s), and nothing has answered them:\n  %s\n"+
				"Each is red's own reading that the report says something its source does not. The lens that read it raises it with the lens's `finding --about-kind contradiction --about <its id> --reason \"<what the source actually says>\"`, graded on every axis — a finding whose `--quote` contains the claim raises it as well, while the report still holds the sentence — and answers it with the gap it mints from that finding (the lens's `mint --found-by <the finding's id>`), so it enters the board with the lifecycle, the blue duty and the gate every other defect has. A finding alone answers nothing: a source the report misstates is a citation defect, material by its class. "+contradictionCitationAnswer+". "+
				"Read them with `show evidence`. A PASS claims the report is sound; these say otherwise on the record. Raising them is a lens's act; PASS waits for it, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s: a source %s read contradicts or does not support %q and it is unanswered — %s", contradictionFact(b), ownerOr(b, contradictionReader), b.Detail, planHold(b, "PASS is not permitted until it is answered"))
		},
	},
	// A CHECK IS OF A SENTENCE, AND THE SENTENCE MOVED (reopened.go). The owner is the anchor's
	// (reopenedOwnerOf) whichever lens made the outdated check: a citation's is the evidence lens, a
	// proof's the lens that owns proofs. The owner owes the check, so it is not retired, and the
	// edit that reopened the anchor moved the head past its last sitting: dispatch readies it.
	//
	// THE OWNER IS WHO IS READIED, NOT WHO ALONE MAY CLEAR. Any lens's check from a sitting opened
	// after the edit clears the blocker (reopenedEvidence), so each wording says both: the check
	// that clears it, and the seat that owes it and is readied for it.
	{
		kind:  BlockerReopened,
		gloss: "a citation or proof unchecked since its sentence changed",
		item: func(b Blocker) string {
			act := "reading it as it now stands against its source"
			if b.Detail != "" {
				act = "judging whether its script establishes the sentence as it now stands (`reproduce --anchor " + b.Subject + "`, which re-runs nothing once the script has a re-run)"
			}
			return fmt.Sprintf("%s — PASS is refused until a lens checks it from a sitting opened after the edit, by %s. %s",
				reopenedFact(b), act, reopenedOwed(b))
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, reopenedFact(b)+" — "+ownerOr(b, reopenedOwner)+" owes the check and is the seat readied for it")
			}
			return fmt.Sprintf("%d citation(s) or proof(s) are unchecked since their sentence changed:\n  %s\n"+
				"Each check on the record is of a sentence the report no longer carries. `show evidence --anchor <id>` reads each as it now stands. "+
				"A `verify` of the citation (any result but `unreachable`) or a `reproduce` naming the proof's anchor, from a sitting opened after the edit, clears it, whichever lens records it; "+
				"PASS waits for the check, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s — %s", reopenedFact(b), planHold(b, "PASS is not permitted until a lens checks it from a sitting opened after the edit. "+reopenedOwed(b)))
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
			return fmt.Sprintf("lens %s is ready — PASS is refused while it is", b.Subject)
		},
		refusal: func(bs []Blocker) string {
			var why []string
			for _, b := range bs {
				why = append(why, b.foldWhy)
			}
			return fmt.Sprintf("%d cast lens(es) are ready: %s. A lens is ready when it has never sat or the report head has moved since its last sitting, until three sittings in a row with no fresh material mint retire it — and it is not retired while it owes: something new on a gap it minted since it last sat, a contradiction it read, a check it owns. A sitting stopped at the per-sitting call limit, or one that never closed, is not a sitting it sat: it moves no pin and counts toward no retirement; `dispatch next` readies them",
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
			return fmt.Sprintf("the report changed behind %d retired lens(es): %s. Read the changes since each pin against that area's duties, and name the areas in a spot-check this sitting (--areas) before a PASS; a defect you find there goes in that spot-check",
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
	// THE INQUIRIES ARE READ ONCE FOR EACH CHANGE TO WHAT THE READ READS. One statement per read, not
	// one per inquiry: presence is not the question, because the inquiries are generated onto the page
	// from the record. What the read owes is a judgement on whether the BODY delivered them, and where
	// it did not, a gap. A review recorded before an edit, or before an inquiry's proposal or fate,
	// answers a question about a document that no longer exists; one recorded after both still stands,
	// so re-reading an unchanged report owes nothing (InquiryReviewDueOf).
	{
		kind:  BlockerInquiryReview,
		gloss: "no inquiry review since the report or its inquiries last changed",
		item: func(Blocker) string {
			return "the report's account of its own research has not been read since the report or its inquiries last changed — PASS is refused until one `inquiry review` says what the read found"
		},
		// THE SHORTFALL IS A DEFECT THE CHAIR FOUND, and the chair mints nothing: this refusal told
		// it to "MINT A GAP" and to "issue `--as FAIL`" otherwise, and the chair reads it alone. It
		// has no `mint`, and the review is the chair's own item, so over a converged board the FAIL
		// is refused too (requireFailIsNotConvergent). Where a shortfall stops the chair, it weighs it
		// like any defect it finds; the stopping judgment in its constitution says where that goes.
		refusal: func([]Blocker) string {
			return "no inquiry review is recorded since the report or its inquiries last changed. " +
				"READ THE REPORT ONCE (`show report`), list what the record claims this run investigated with " +
				"`show inquiries`, and answer in one act: `inquiry review --reason \"<what the report " +
				"says at those inquiries>\"`. Where an inquiry's research is thin, missing or unsupported by the text, " +
				"that shortfall is a defect you found, weighed before your verdict like any other; this " +
				"event only records that the read happened, because an absent review reads exactly like a sound " +
				"one. A PASS claims the report is sound, and its account of what this run investigated is part " +
				"of the report; record the review before a PASS"
		},
		why: func(b Blocker) string {
			return "no inquiry review since the report or its inquiries last changed — " + planHold(b, "PASS is not permitted until one is recorded")
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
//
// TWO SETS OF ARMS. The stream alone answers eventBlockersOf's; the gap arms and the lens arms read
// the open gaps and the fresh-material set, which are SQL as it stands now.
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
	events := eventBlockersOf(evs, ids, win)
	out = append(out, events...)
	if lg := passLensGateOf(evs, ids, win, fresh, lensOwesOf(evs, ids, win, gaps, events)); lg.cast && lg.head != 0 {
		for _, f := range lg.ready() {
			out = append(out, Blocker{Kind: BlockerLensReady, Subject: f.seat, Owner: f.seat, foldWhy: f.why})
		}
		for _, a := range lg.uncovered() {
			out = append(out, Blocker{Kind: BlockerStaleArea, Subject: a.SeatID, Owner: chairSeat, Detail: fmt.Sprint(a.Pin)})
		}
	}
	slices.SortStableFunc(out, func(a, b Blocker) int { return kindRank(a.Kind) - kindRank(b.Kind) })
	return out
}

// kindRank is a kind's place in the table.
func kindRank(k BlockerKind) int {
	return slices.IndexFunc(blockerKinds, func(r blockerKind) bool { return r.kind == k })
}

// eventBlockersOf is the arms of the gate's list that the stream alone answers, in table order: a
// contradiction no gap answers, a citation or proof unchecked since its sentence changed, no
// ingested report, an unruled motion, no inquiry review since the report or its inquiries last
// changed. ids is evs's places, aligned with it.
//
// IT ANSWERS FOR THE STREAM IT IS HANDED. Asked of a prefix of the record, it says what held a PASS
// when the record ended there: no arm asks the index for a seat's newest sitting.
func eventBlockersOf(evs []*Event, ids []int64, win WindowIndex) []Blocker {
	var out []Blocker
	// ids IS evs's events.id, aligned with it — the place dispatchLedger and benchOccasionSittings read
	// sittings by, and the one a stored sitting is named by — so a blocker's `since` and the
	// sitting it is compared with are places in one sequence.
	for _, c := range unansweredContradictions(evs, win) {
		out = append(out, Blocker{Kind: BlockerContradiction, Subject: c.idWord(), Detail: c.claim, Owner: c.reader, since: c.since, raisedBy: c.raisedBy})
	}
	// A REOPENED ANCHOR HOLDS THE PASS ONLY WHERE THE CAST SEATS ITS OWNER. A blocker whose owner
	// cannot be dispatched holds the run UNVERIFIED with nobody to ask; `show evidence` lists the
	// anchor either way.
	cast := castOfEvents(evs)
	for _, r := range reopenedEvidence(evs, win) {
		if owner := reopenedOwnerOf(cast, r); owner != "" {
			out = append(out, Blocker{Kind: BlockerReopened, Subject: r.anchor, Owner: owner, Detail: r.sha, since: r.edit, lastCheck: r.lastCheck})
		}
	}
	if cast != nil && reportHeadOf(evs, ids) == 0 {
		out = append(out, Blocker{Kind: BlockerNoReport})
	}
	out = append(out, unruledMotionBlockers(evs, ids, win)...)
	if InquiryReviewDueOf(evs, win) {
		out = append(out, Blocker{Kind: BlockerInquiryReview, Owner: chairSeat})
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

// BlockerKinds is every kind of the table, in its order.
func BlockerKinds() []BlockerKind {
	out := make([]BlockerKind, 0, len(blockerKinds))
	for _, r := range blockerKinds {
		out = append(out, r.kind)
	}
	return out
}

// PassBlockers is passBlockersOf over the run, read as the gate reads it — the stream, the fresh
// material gaps and the open gaps on one snapshot (readSnapshot), as the dispatch plan reads them, so
// the gate and the plan cannot disagree about a blocker over a write landing between two reads. The
// fold runs after the transaction closes.
func PassBlockers(run Run) ([]Blocker, error) {
	var r planReads
	absent := false
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		if absent = noRecord(q); absent {
			return nil
		}
		var err error
		r, err = passBlockerReadsAt(q)
		return err
	}); err != nil || absent {
		return nil, err
	}
	return passBlockersOfReads(r), nil
}

// passBlockersAt is PassBlockers asked of q, the write path's transaction: the gate and the outcome
// read the blockers on the record their verdict lands on, so a mint, a motion or a finding another
// seat commits cannot fall between the blockers read and the verdict written.
func passBlockersAt(q recordsql.Querier) ([]Blocker, error) {
	if noRecord(q) {
		return nil, nil
	}
	r, err := passBlockerReadsAt(q)
	if err != nil {
		return nil, err
	}
	return passBlockersOfReads(r), nil
}

// passBlockerReadsAt is every answer the gate's list takes off the record: the stream, the fresh
// material gaps and the open gaps.
func passBlockerReadsAt(q recordsql.Querier) (planReads, error) {
	var r planReads
	var err error
	if r.evs, r.win, err = eventsAt(q); err != nil {
		return r, err
	}
	if r.fresh, err = freshMaterialOf(q); err != nil {
		return r, err
	}
	r.gaps, err = openGaps(q)
	return r, err
}

// passBlockersOfReads folds the gate's list from what passBlockerReadsAt read.
func passBlockersOfReads(r planReads) []Blocker {
	return passBlockersOf(r.evs, r.win.IDs(r.evs), r.win, blockerGapsOfOpen(r.gaps), r.fresh)
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
func requireNoBlockers(w writeView, verdict recordpb.Verdict, blockers []Blocker) error {
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
		return requireFailIsNotConvergent(w, blockers)
	}
	return nil
}

// requireBenchMotionsRuled refuses the run's outcome while a motion whose gavel is the bench's
// stands unruled. The terminal bench sitting exists to rule them, and it fires on the bench-owned
// blockers of the chair's last plan, which the record computes and capture holds to the record. A
// motion filed after that plan is on no plan, so the terminal sitting does not convene for it: this
// refusal is its backstop (#1202). Refused here, such a motion is loud at the one write path, and
// the bench recording the outcome holds the gavel, so it rules the motion and records the outcome
// after.
//
// A HALT IS EXEMPT. It is the bench's own decision to end the run, taken with the record in view;
// refusing it would force rulings after the bench has already stopped the run.
func requireBenchMotionsRuled(q recordsql.Querier) error {
	halted, err := recordHasAt(q, `SELECT 1 FROM "halt" LIMIT 1`)
	if err != nil || halted {
		return err
	}
	blockers, err := passBlockersAt(q)
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
	return fmt.Errorf("%d motion(s) whose gavel is the bench's stand unruled: %s. "+
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
	head := "PASS refused — "
	if every {
		head = "a verdict is refused — "
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
