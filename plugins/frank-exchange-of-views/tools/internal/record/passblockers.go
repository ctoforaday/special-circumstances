package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// WHAT HOLDS A PASS IS DECIDED HERE, ONCE.
//
// Four readers each worked it out for themselves — the verdict gate, the chair's work list,
// `dispatch next`'s pass_permitted and the FAIL convergence refusal — and they disagreed: the plan
// said pass_permitted over an unruled petition, a source reading nothing answered and a docket
// motion on a gap that had since closed, and the gate refused the PASS the chair then tried. On a converged
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
	BlockerMaterialGap      BlockerKind = "material_gap"
	BlockerRemandOwed       BlockerKind = "remand_owed"
	BlockerUnchecked        BlockerKind = "unchecked"
	BlockerFailedCheck      BlockerKind = "failed_check"
	BlockerNoReport         BlockerKind = "no_report"
	BlockerLensReady        BlockerKind = "lens_ready"
	BlockerStaleArea        BlockerKind = "stale_area"
	BlockerGradeMovement    BlockerKind = "grade_movement"
	BlockerClosingOwed      BlockerKind = "closing_owed"
	BlockerUnruledMotion    BlockerKind = "unruled_motion"
	BlockerUpheldPosition   BlockerKind = "upheld_position"
	BlockerOpenInquiry      BlockerKind = "open_inquiry"
	BlockerUncheckedInquiry BlockerKind = "unchecked_inquiry"
)

// Blocker is one thing on the board holding the PASS gate, and the seat whose act clears it.
type Blocker struct {
	Kind BlockerKind
	// Subject is what holds it: a gap id, an anchor, a lens seat, a motion id, an inquiry id — ""
	// for the kind that is about the run as a whole (no_report).
	Subject string
	// Owner is the seat whose act clears it — the gap's minter, the seat whose check failed, the
	// ready lens, the motion's gavel-holder, the chair for its own reads — and "" where no seat on
	// the record can.
	Owner string
	// Detail is what the wording needs beyond the subject: a motion's subject and gavel, a stale
	// area's pin, an unchecked or failed-check computation's sha256 ("" for a citation).
	Detail string
	// state is a failed check's state (CheckState); "" for every other kind.
	state CheckState
	// About is the gap an unruled motion is about, "" for a motion about no gap.
	About string
	// foldWhy is a ready lens's reason in the lens fold's own words, for the refusal.
	foldWhy string
	// motionSubject is an unruled motion's subject — which sitting dispatch convenes its gavel for.
	motionSubject string
	// since is the events.id the blocker arose at — a motion's filing, the edit that changed an
	// unchecked anchor's sentence (0 for one no edit has changed), the failed check, an inquiry's
	// latest fate — so dispatch can ask whether its owner has sat since, and an unchecked anchor's
	// wording can name its edit; 0 for the other kinds.
	since int64
	// placed is the events.id of the act that put an unchecked anchor on the record, which dispatch
	// readies its owner off (a cite or a computation moves no report head), and of an unchecked
	// inquiry's latest cause (InquiryJudgement.cause); 0 for every other blocker.
	placed int64
	// lastCheck is the events.id of an unchecked anchor's or inquiry's latest check, which does not
	// clear it, 0 when it has none.
	lastCheck int64
	// notes is the findings other lenses filed about an unchecked inquiry since its fate, each as
	// "<id> (<seat>)"; nil for every other blocker.
	notes []string
	// rerun is an unchecked computation's script state (unchecked.rerun), "" for every other blocker.
	rerun string
	// answeredBy is the gap that carries a failed check — for unread and unsound, the gap its seat
	// minted about the anchor; for found_against, the gap its reading names, followed through its
	// successors — "" while none does, and answeredAt that gap's mint's events.id.
	answeredBy string
	answeredAt int64
	// carriers is every gap that carries a failed check, answeredBy among them — for unread and
	// unsound, every gap its seat minted about the anchor from the check's sitting on; for
	// found_against, the gap its reading names and its whole successor tree — each a gap whose close
	// the record refuses while the check stands (requireFailedCheckAnswered); nil for every other kind.
	carriers []string
	// awaited is a closing blocker whose party has sat through no sitting opened since the
	// docketing (ClosingAwaited); false for a missing one, whose party sat and filed none.
	awaited bool
	// moved is a grade-movement blocker's gap's movement the bench has not ruled on, and totalMoved
	// the open gaps' total (gradeMovementOf).
	moved, totalMoved float64
}

// AnsweredBy is the gap that carries a failed check, "" while none does — and for every other kind.
func (b Blocker) AnsweredBy() string { return b.answeredBy }

// Carriers is every gap that carries a failed check, each one's close refused while the check
// stands; nil for every other kind.
func (b Blocker) Carriers() []string { return b.carriers }

// CheckState is a failed check's state, "" for every other kind.
func (b Blocker) CheckState() CheckState { return b.state }

// Since is the events.id a blocker arose at — for a failed check, the reading that failed it — and 0
// for the kinds that carry none.
func (b Blocker) Since() int64 { return b.since }

// owed reports whether a failed check's own seat owes its gap: an unread or unsound check no gap
// carries (failedCheck.owed).
func (b Blocker) owed() bool {
	return b.Kind == BlockerFailedCheck && b.state != CheckFoundAgainst && b.answeredBy == ""
}

// ChairOwned reports whether the chair clears it in its own sitting.
func (b Blocker) ChairOwned() bool { return b.Owner == chairSeat }

// blockerKind is one row of the table every surface words a blocker from.
type blockerKind struct {
	kind BlockerKind
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

// uncheckedFact is the one sentence every wording of an unchecked blocker states: which anchor, the
// edit that changed its sentence where one did, and its latest check — each as the record holds it,
// so a reader can go to each.
func uncheckedFact(b Blocker) string {
	switch {
	case b.since == 0 && b.Detail != "":
		return "computation " + b.Subject + " is on the record and nobody has judged whether its program establishes the sentence it backs"
	case b.since == 0:
		return "citation " + b.Subject + " is on the record and nobody has read its source against the sentence it backs"
	case b.lastCheck == 0:
		return fmt.Sprintf("%s's sentence changed at edit %d, and nobody has checked it", b.Subject, b.since)
	}
	return fmt.Sprintf("%s's sentence changed at edit %d and no sitting opened since has checked it (its last check is event %d)", b.Subject, b.since, b.lastCheck)
}

// uncheckedAct is the check an unchecked blocker is owed, in the order the act needs: a citation's
// source read against its sentence; a computation's program read, then compared against its bundle's
// re-run — executing the bundle where nobody has — then its sentence judged. An execution belongs to
// the bundle, so a bundle with a standing re-run is never executed again; the comparison belongs to
// this computation, so it is made against this computation's own recorded output whatever another
// computation on the bundle showed.
//
// THE COMPUTATION ITEM CARRIES THE ID THE VERB TAKES: the computation's anchor, which `reproduce --anchor` takes
// and the seat reads in the report. When the verb took the script's sha256 and the item named the
// anchor, three lenses on universe-m11 spent 22 computation-file reads and 14 filesystem hunts translating
// one into the other, for 12 re-runs. AND IT SAYS WHERE THE PROGRAM IS READ FIRST: an item that handed
// over the re-run's id let a lens record a verdict on a script it had not read (wave A, 2026-10-09:
// three placeholder verdicts, each of which stands).
func uncheckedAct(b Blocker) string {
	if b.Detail == "" {
		return "reading its source against the sentence as it now stands (`verify --anchor " + b.Subject + "`)"
	}
	read := "`show evidence --anchor " + b.Subject + "` prints the program and its recorded output"
	switch b.rerun {
	case RerunNone, "":
		return "READING its program and then RE-RUNNING it, never by reading alone: " + read + ", and `reproduce --anchor " + b.Subject + "` then executes its bundle, compares this computation's recorded output, and records your verdict on its sentence in one act"
	case RerunNotCompared:
		return "judging whether its program establishes THIS sentence — its bundle has a standing re-run, so `reproduce --anchor " + b.Subject + "` executes nothing, compares THIS computation's own output, exit and out/ against that re-run, and records your verdict on its sentence: " + read
	}
	showed := "did not reproduce (" + b.rerun + ")"
	if b.rerun == recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED) {
		showed = "REPRODUCED"
	}
	return "judging whether its program establishes THIS sentence — compared against its bundle's re-run, its own output " + showed + ", and one bundle can back several sentences, so a verdict on another is not a verdict on this one: " +
		read + ", and `reproduce --anchor " + b.Subject + "` records your verdict on its sentence, executing nothing once the bundle has a standing re-run"
}

// uncheckedOwed is the other sentence every wording of an unchecked blocker states: whose the check
// is. Its owner owes it (anchorOwnerOf) — dispatch readies that seat and its work list blocks on
// the anchor — and it is not the only seat whose check clears: the fold reads no seat.
func uncheckedOwed(b Blocker) string {
	return ownerOr(b, uncheckedOwner) + " owes that check and is the seat readied for it; another lens's check clears it as well"
}

// uncheckedAfter is when a check clears an unchecked blocker: any check when no edit changed the
// sentence, else one from a sitting opened after the edit.
func uncheckedAfter(b Blocker) string {
	if b.since == 0 {
		return ""
	}
	return " from a sitting opened after the edit"
}

// uncheckedOwner is who owes an unchecked anchor's check when the record names no seat.
const uncheckedOwner = "the lens that owns it"

// failedFact is the one sentence every wording of a failed-check blocker states: which anchor, whose
// check failed and how, and where it is on the record.
func failedFact(b Blocker) string {
	switch b.state {
	case CheckUnsound:
		return fmt.Sprintf("%s judged the sentence computation %s backs unsound at event %d, and that judgement stands", ownerOr(b, failedChecker), b.Subject, b.since)
	case CheckFoundAgainst:
		return fmt.Sprintf("%s's reading of citation %s at event %d finds the source against the sentence (`refutes` or `absent`), and the report still carries that sentence", ownerOr(b, failedChecker), b.Subject, b.since)
	}
	return fmt.Sprintf("%s's check of citation %s at event %d read nothing of its source — it was unreachable, or the copy read was not the work — and no check of the sentence as it stands has read it", ownerOr(b, failedChecker), b.Subject, b.since)
}

// failedClears is what clears a failed-check blocker, stated by every wording of it.
func failedClears(b Blocker) string {
	switch b.state {
	case CheckUnsound:
		return "it clears when the computation is judged sound again — by the judging lens's own later judgement, or by a judgement from a sitting opened after blue's edit changed the sentence, since an edit leaves it unchecked — or when blue retires it; another lens's `sound` beside a standing `unsound` is a disagreement and clears nothing"
	case CheckFoundAgainst:
		return fmt.Sprintf("it clears when %s's own later reading of the sentence as it stands backs it, or when blue edits the sentence, which leaves it unchecked, or retires the citation; another lens's backing reading beside it is a disagreement and clears nothing. "+
			"The lens's close of the gap that carries it is refused while it stands; a closing ruling of the bench closes that gap, and the check still holds the PASS — no bench ruling clears it", ownerOr(b, failedChecker))
	}
	return "it clears when a lens reads the source — any result but `unreachable`, of a copy that is the work — or when blue retires the citation, or edits its sentence, which leaves it unchecked; the close of the gap that carries it is refused while it stands, and no bench ruling clears it"
}

// failedChecker is whose check failed when the record names no seat.
const failedChecker = "a lens"

// failedRemedy is blue's act on a failed check.
func failedRemedy(b Blocker) string {
	switch b.state {
	case CheckUnsound:
		return "fixes the computation or retires it"
	case CheckFoundAgainst:
		return "edits the sentence, or retires or replaces the citation"
	}
	return "retires the citation or cites a source a lens can read"
}

// failedOwed is whose act is owed on a failed-check blocker: for unread and unsound, the gap of the
// seat whose check failed until it is minted, and blue's remedy; for found_against, blue's remedy
// alone — its reader owes no mint, carried or not.
func failedOwed(b Blocker) string {
	switch {
	case b.owed():
		return fmt.Sprintf("%s owes its gap about the anchor (`mint --about-kind anchor --about %s`), which takes no mint of its budget, and blue %s", ownerOr(b, failedChecker), b.Subject, failedRemedy(b))
	case b.answeredBy == "":
		return fmt.Sprintf("no gap names it — blue %s", failedRemedy(b))
	}
	return fmt.Sprintf("gap %s carries it to blue, which %s", b.answeredBy, failedRemedy(b))
}

// blueFailedItem is blue-respond's work-list sentence for a failed-check blocker: the remedy it owes.
func blueFailedItem(b Blocker) string {
	if b.state == CheckFoundAgainst {
		return fmt.Sprintf("%s — PASS is refused while it stands; %s. Make the sentence say what the source says — an edit leaves the citation unchecked, owed a fresh check — or retire the citation, or replace it with a source that backs the sentence; %s", failedFact(b), failedOwed(b), failedClears(b))
	}
	carried := "the lens whose check failed has minted no gap about it yet"
	if b.answeredBy != "" {
		carried = "gap " + b.answeredBy + " states the defect and its fix"
	}
	if b.state == CheckUnsound {
		return fmt.Sprintf("%s — PASS is refused while it does, and %s. Fix it — an edit that makes its sentence say what the program shows, which leaves the computation unchecked for a fresh judgement, or a new computation with this one retired — or retire the claim it backs", failedFact(b), carried)
	}
	retire := "Retire the citation"
	if b.answeredBy != "" {
		retire = "Retire the citation, answering gap " + b.answeredBy + ","
	}
	return fmt.Sprintf("%s — PASS is refused while it stands, and %s. %s — its sentence may stay word for word, backed by whatever else backs it — or cite a source a lens can read and retire this one; an edit of its sentence leaves it unchecked, owed a fresh check", failedFact(b), carried, retire)
}

// failedCheckRefusal is the gate's paragraph for the failed checks that hold it: each named, then
// what is owed and what clears it, for each state present and no other — a found-against check owes
// its reader nothing, so its paragraph names no mint, and an unread one's clear condition (any lens's
// reading of the work) is not a found-against one's (its own reader's).
func failedCheckRefusal(bs []Blocker) string {
	var named []string
	states := map[CheckState]bool{}
	for _, b := range bs {
		states[b.state] = true
		gap := "no gap carries it yet"
		switch {
		case b.answeredBy != "":
			gap = "gap " + b.answeredBy + " carries it"
		case b.state == CheckFoundAgainst:
			gap = "no gap names it"
		}
		how := map[CheckState]string{CheckUnread: "read nothing of its source", CheckUnsound: "judged unsound", CheckFoundAgainst: "found against"}[b.state]
		named = append(named, fmt.Sprintf("%s (%s by %s at event %d; %s)", b.Subject, how, ownerOr(b, failedChecker), b.since, gap))
	}
	var said []string
	if states[CheckUnread] || states[CheckUnsound] {
		said = append(said, "A PASS would certify a sentence whose computation the record calls unsound, or whose cited source nobody could read. The lens whose check failed mints its gap about the anchor (the lens's `mint --about-kind anchor --about <its anchor>`, which takes no mint of its budget), and blue fixes or retires the computation, or retires the citation or cites a source a lens can read. "+
			"A computation clears when it is judged sound again or retired, a citation when a lens reads its source or blue retires it, and either when blue's edit changes its sentence, which leaves it unchecked; the close of the gap that carries it is refused while the check stands, and no bench ruling clears it.")
	}
	if states[CheckFoundAgainst] {
		said = append(said, "A PASS would certify a sentence a lens's reading of its cited source finds against: the reading found the source against the sentence, and its reader owes no mint — the gap a reading mints in its act carries it, or no gap names it; blue edits the sentence, or retires or replaces the citation. "+
			"It clears when its reader's own later reading of the sentence as it stands backs it, or when blue edits the sentence, which leaves it unchecked, or retires the citation; another lens's backing reading beside it is a disagreement and clears nothing. "+
			"The close of the gap that carries it is refused while it stands; a closing ruling of the bench closes that gap, and the check still holds the PASS — no bench ruling clears it.")
	}
	return fmt.Sprintf("%d citation(s) or computation(s) stand on a failed check:\n  %s\n%s "+
		"`show evidence --anchor <id>` reads each check. PASS waits for that, and `--as FAIL` does not",
		len(bs), strings.Join(named, "\n  "), strings.Join(said, " "))
}

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
	// A REMAND OF A GAP THAT IS NOT MATERIAL HOLDS THE PASS UNTIL ITS EXCHANGE HAS BEGUN (ruled by
	// gblock 2026-10-24). A material gap holds it as material_gap
	// whatever its remand; this row is the one a gap below material holds once the bench has sent it
	// back, so the plan that readies blue for the exchange never reads as permitting a PASS over it
	// (routeOf).
	{
		kind:  BlockerRemandOwed,
		gloss: "a gap that is not material, remanded by the bench, whose one more exchange has not been had",
		item: func(b Blocker) string {
			return fmt.Sprintf("gap %s is not material and the BENCH REMANDED it — PASS is refused until the one more exchange the remand grants, between %s and blue, has been had — an exchange that begins when blue records an act on the gap; the dispatch readies them on the ruling's direction", b.Subject, ownerOr(b, "the lens that minted it"))
		},
		refusal: func(bs []Blocker) string {
			return fmt.Sprintf("%d gap(s) the bench REMANDED are owed the one more exchange each remand grants: %s. The bench's remand outranks materiality, so a gap that is not material holds the PASS until its minting lens and blue have had that exchange — one that begins when blue records an act on the gap — on the ruling's direction; the dispatch readies blue for it. PASS waits for it; `--as FAIL` does not",
				len(bs), strings.Join(subjects(bs), ", "))
		},
	},
	// A CITATION OR COMPUTATION NOBODY HAS CHECKED AS IT STANDS HOLDS THE PASS (unchecked.go). The owner is
	// the anchor's (anchorOwnerOf) whichever lens made any earlier check: a citation's is the evidence
	// lens, a computation's the computation lens. The owner owes the check, so it is not retired, and the
	// act that placed or changed the anchor moved the head past its last sitting: dispatch readies it.
	//
	// THE OWNER IS WHO IS READIED, NOT WHO ALONE MAY CLEAR. Any lens's check clears the blocker, so
	// each wording says both: the check that clears it, and the seat that owes it and is readied.
	{
		kind:  BlockerUnchecked,
		gloss: "a citation or computation with no check of the sentence the report carries",
		item: func(b Blocker) string {
			return fmt.Sprintf("%s — PASS is refused until a lens checks it%s, by %s. %s",
				uncheckedFact(b), uncheckedAfter(b), uncheckedAct(b), uncheckedOwed(b))
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, uncheckedFact(b)+" — "+ownerOr(b, uncheckedOwner)+" owes the check and is the seat readied for it")
			}
			return fmt.Sprintf("%d citation(s) or computation(s) have no check of the sentence the report carries:\n  %s\n"+
				"A citation nobody has read against its sentence, or a computation nobody has judged, is a claim nothing on the record backs yet; a check of a sentence an edit has since replaced is of a sentence the report no longer carries. `show evidence --anchor <id>` reads each as it now stands. "+
				"A `verify` of the citation that reads its source (any result but `unreachable`, of a copy that is the work) or a `reproduce` naming the computation's anchor — from a sitting opened after the edit, where an edit changed the sentence — clears it, whichever lens records it; "+
				"PASS waits for the check, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s — %s", uncheckedFact(b), planHold(b, "PASS is not permitted until a lens checks it"+uncheckedAfter(b)+". "+uncheckedOwed(b)))
		},
	},
	// A CHECK THAT FAILED AND STANDS HOLDS THE PASS (failedcheck.go): a computation judged unsound, a
	// citation whose source no check could read, a citation a lens's standing reading finds against.
	// Its owner is the seat whose check failed. For unread and unsound it owes the gap that carries
	// the check to blue; a found-against check's gap was minted by its reading, and its reader owes
	// nothing. Blue's remedy, or a check that holds, clears it.
	{
		kind:  BlockerFailedCheck,
		gloss: "a citation whose source no check could read, a citation a lens's standing reading finds against, or a computation a lens's standing judgement calls unsound",
		item: func(b Blocker) string {
			return fmt.Sprintf("%s — PASS is refused while it stands; %s. %s", failedFact(b), failedClears(b), failedOwed(b))
		},
		refusal: failedCheckRefusal,
		why: func(b Blocker) string {
			switch b.state {
			case CheckUnsound:
				return fmt.Sprintf("%s — %s", failedFact(b), planHold(b, "PASS is not permitted until the computation is judged sound again or retired"))
			case CheckFoundAgainst:
				return fmt.Sprintf("%s — %s; %s", failedFact(b), planHold(b, "PASS is not permitted until "+ownerOr(b, failedChecker)+"'s own later reading backs the sentence, or blue edits the sentence or retires the citation"), failedOwed(b))
			}
			return fmt.Sprintf("%s — %s", failedFact(b), planHold(b, "PASS is not permitted until a lens reads its source or blue retires the citation"))
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
			return fmt.Sprintf("%d cast lens(es) are ready: %s. A lens is ready when it has never sat or its pin — the head the tool last served it the whole report at, in a sitting it sat through — is behind the head, until three sittings in a row with no fresh material mint retire it — and it is not retired while it owes: something new on a gap it minted since it last sat, a check it owns, the gap about a check of its that read nothing or judged a computation unsound. A sitting stopped at the per-sitting call limit, or one that never closed, is not a sitting it sat: a read in it moves no pin, and it counts toward no retirement; `dispatch next` readies them",
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
	// GRADE MOVEMENT IN TOTAL IS THE BENCH'S (grademovement.go). The gap's movement is not the reason
	// alone: the board's total past the threshold is, so every wording states both. The dispatch dockets
	// the gap, and from then its docket motion holds the PASS as an unruled motion. The gap's own reason
	// in the plan states its readiness, so the kind has no plan line of its own.
	{
		kind:  BlockerGradeMovement,
		gloss: "grade movement the bench has not ruled on past its threshold in total",
		item: func(b Blocker) string {
			return gradeMovementFact(b) + " — PASS is refused until the bench rules on it; the dispatch dockets the gap for the bench"
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, fmt.Sprintf("%s: %.1f (%s)", b.Subject, b.moved, b.Detail))
			}
			return fmt.Sprintf("%d gap(s) carry grade movement the bench has not ruled on, and the open gaps' total of %.1f mass is past the %.1f at which the bench rules on it:\n  %s\n"+
				"A moved grade changes what the gate reads: a gap regraded below medium stops holding a PASS, and gaps regraded up keep a converged board out of the FAIL refusal, so grade movement in total goes before the bench before it stands. "+
				"`dispatch next` dockets each gap for the bench, and the bench's ruling on the gap's docket motion rules on its movement. PASS waits for those rulings, and `--as FAIL` does not",
				len(bs), bs[0].totalMoved, GradeMovementThreshold, strings.Join(named, "\n  "))
		},
	},
	// A DOCKETED GAP IS RULED ON ITS PARTIES' CLOSINGS (closingowed.go). The owner is the party — the
	// gap's minting lens, or blue-respond — and the docket motion holds the PASS as an unruled motion
	// as well, so every wording says the bench rules once the closings are in. The gap's own reason in
	// the plan states who is readied and whether the bench sits, so the kind has no plan line of its own.
	{
		kind:  BlockerClosingOwed,
		gloss: "a docketed gap's closing argument a party has not filed",
		item: func(b Blocker) string {
			state := "is not on the record — the bench rules the gap once each party has filed its closing, and waits for this one"
			if !b.awaited {
				state = "is not on the record, and " + ownerOr(b, "the party") + " has sat since the docketing without filing it — the bench rules on the record as it stands"
			}
			return fmt.Sprintf("gap %s is docketed for the bench (motion %s) and the closing argument of %s on it %s; PASS is refused until the bench rules", b.Subject, b.Detail, ownerOr(b, "a party"), state)
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, fmt.Sprintf("%s (%s, motion %s)", b.Subject, ownerOr(b, "a party"), b.Detail))
			}
			return fmt.Sprintf("%d closing argument(s) on docketed gaps are not filed: %s. Each party to a docketed gap — its minting lens and blue-respond — files its closing, and the bench rules the gap on them, the transcript and the artifacts. PASS waits for that ruling, and `--as FAIL` does not",
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
	// AN UPHELD POSITION NAMING A LENS HOLDS THE PASS UNTIL THAT LENS HAS AUDITED ITS AREA (gblock,
	// 2026-10-30; positionheard.go): the lens is served the report in a sitting opened after the
	// ruling. The owner is the lens, so the plan readies it; the subject is the position's id.
	{
		kind:  BlockerUpheldPosition,
		gloss: "an audit the bench named in upholding the chair's position, not yet had",
		item: func(b Blocker) string {
			return fmt.Sprintf("the bench upheld the chair's position %s and named %s's area: it owes an audit of it on the ruling's direction — %q — and has been served the report in no sitting since the ruling; PASS is refused until it has", b.Subject, ownerOr(b, "a lens"), b.Detail)
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, fmt.Sprintf("%s (%s, on %q)", b.Subject, ownerOr(b, "a lens"), b.Detail))
			}
			return fmt.Sprintf("%d position(s) the bench upheld name a lens whose audit of its area has not been had: %s. "+
				"The bench held that the PASS was rightly withheld and named the work the debate continues on; the lens is readied to audit its area on the ruling's direction, and a defect it finds enters the board as a gap. "+
				"PASS waits for that audit, and `--as FAIL` does not",
				len(bs), strings.Join(named, ", "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s: the bench upheld the chair's position, naming %s to audit its area on %q — %s", b.Subject, ownerOr(b, "a lens"), b.Detail, planHold(b, "PASS is not permitted until it has"))
		},
	},
	// AN OPEN INQUIRY HOLDS THE PASS UNTIL BLUE RECORDS A FATE THAT SETTLES IT (ruled by gblock
	// 2026-11-01; inquirycheck.go). Its owner is blue-respond, which dispatch readies while any
	// inquiry is open; the plan's line names the latest fate's event, so a re-recorded `underway`
	// moves it, and a run that cannot settle its inquiries reaches its epoch limit and ends CEILING.
	{
		kind:  BlockerOpenInquiry,
		gloss: "an inquiry still proposed or underway, with no fate that settles it",
		item: func(b Blocker) string {
			return openInquiryFact(b) + " — PASS is refused until blue records a fate that settles it: confirmed or refuted for an approach that ran, abandoned for one that gave no answer, declined or deferred with why"
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, fmt.Sprintf("%s (%s)", b.Subject, b.Detail))
			}
			return fmt.Sprintf("%d inquiry(ies) are still OPEN: %s. An open inquiry is a decision the report has not made, and a PASS would certify research the record says is unfinished. "+
				"Blue records a fate that settles each — confirmed or refuted for an approach that ran, abandoned for one that gave no answer, declined or deferred with why — and the logic lens then checks it. "+
				"PASS waits for those fates, and `--as FAIL` does not",
				len(bs), strings.Join(named, ", "))
		},
		why: func(b Blocker) string {
			return openInquiryFact(b) + " — " + planHold(b, "PASS is not permitted until blue records a fate that settles it")
		},
	},
	// A SETTLED INQUIRY HOLDS THE PASS UNTIL A CHECK STANDS ON IT (ruled by gblock 2026-11-01). Its
	// owner is the logic lens (inquiryOwner), which dispatch readies once per cause; a wanting inquiry
	// holds the PASS through its gap's material_gap row instead, and has no row here.
	{
		kind:  BlockerUncheckedInquiry,
		gloss: "a settled inquiry with no check of its latest fate",
		item: func(b Blocker) string {
			return uncheckedInquiryFact(b) + " — PASS is refused until one does. " + uncheckedInquiryAct(b)
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, uncheckedInquiryFact(b))
			}
			return fmt.Sprintf("%d settled inquiry(ies) have no check of their latest fate:\n  %s\n"+
				"A settled inquiry is the report's account of research the run did or set aside, and nothing on the record has judged that account. "+
				"The logic lens judges each: `inquiry check --id <id> --reason \"...\"` where its fate stands as the report's account of it, or the gap about it where it is wanting. "+
				"A fate recorded again, or a material gap minted about it, leaves its check stale. PASS waits for those checks, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return uncheckedInquiryFact(b) + " — " + planHold(b, "PASS is not permitted until "+ownerOr(b, inquiryOwner)+" checks it")
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
	id, mintedBy string
	material     bool
	// latestRemandOwed is the gap's latest remand with no exchange since its ruling in which blue
	// acted on the gap (latestRemandOwed).
	latestRemandOwed bool
	// inquiry is the inquiry the gap is about, "" for a gap about anything else: an open material
	// one holds the inquiry wanting (inquiryJudgementsOf).
	inquiry string
}

// inquiryAbout is the inquiry a gap about kind and ref is about, "" where it is about anything else.
func inquiryAbout(kind, ref string) string {
	if kind != recordpb.Word(recordpb.AboutKind_ABOUT_KIND_INQUIRY) {
		return ""
	}
	return ref
}

func blockerGapsOfStates(gaps []WorkGapState) []blockerGap {
	var out []blockerGap
	for _, g := range gaps {
		if !g.Open {
			continue
		}
		out = append(out, blockerGap{id: g.ID, mintedBy: g.MintedBy, material: g.Material, latestRemandOwed: g.latestRemandOwed,
			inquiry: inquiryAbout(g.AboutKind, g.AboutRef)})
	}
	return out
}

// blockerGapsOfOpen is the plan's open gaps as the list reads them, with owed the gaps whose latest
// remand's exchange is owed (remandsOwedOf) — nil where the reader asks only each gap's id and
// minter, as the lens fold does (lensOwesOf, owedBlockersOf).
func blockerGapsOfOpen(gaps []openGap, owed map[string]bool) []blockerGap {
	out := make([]blockerGap, 0, len(gaps))
	for _, g := range gaps {
		out = append(out, blockerGap{id: g.id, mintedBy: g.mintedBy, material: g.material, latestRemandOwed: owed[g.id], inquiry: g.inquiry})
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
		if g.material {
			out = append(out, Blocker{Kind: BlockerMaterialGap, Subject: g.id, Owner: g.mintedBy})
		}
	}
	for _, g := range gaps {
		if !g.material && g.latestRemandOwed {
			out = append(out, Blocker{Kind: BlockerRemandOwed, Subject: g.id, Owner: g.mintedBy})
		}
	}
	events := owedBlockersOf(evs, ids, win, gaps)
	out = append(out, events...)
	out = append(out, gradeMovementOf(evs, ids, win, openSet(gaps)).blockers()...)
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
// citation or computation with no check of the sentence the report carries, a standing failed check
// of one, no ingested report, an unruled motion, an upheld position's audit not yet had. ids is
// evs's places, aligned with it; open is the gap view's open set on the same snapshot, which names
// the gap a found-against check is carried by.
//
// IT ANSWERS FOR THE STREAM IT IS HANDED. Asked of a prefix of the record, it says what held a PASS
// when the record ended there: no arm asks the index for a seat's newest sitting.
func eventBlockersOf(evs []*Event, ids []int64, win WindowIndex, open map[string]bool) []Blocker {
	var out []Blocker
	// ids IS evs's events.id, aligned with it — the place dispatchLedger and benchOccasionSittings read
	// sittings by, and the one a stored sitting is named by — so a blocker's `since` and the
	// sitting it is compared with are places in one sequence.
	for _, u := range uncheckedAnchors(evs, win) {
		out = append(out, Blocker{Kind: BlockerUnchecked, Subject: u.anchor, Owner: anchorOwnerOf(u.computation), Detail: u.sha, since: u.edit, placed: u.placed, lastCheck: u.lastCheck, rerun: u.rerun})
	}
	for _, f := range failedChecksOf(evs, win, open) {
		out = append(out, Blocker{Kind: BlockerFailedCheck, Subject: f.anchor, Owner: f.owner, Detail: f.sha, state: f.state, since: f.since, answeredBy: f.answeredBy, answeredAt: f.answeredAt, carriers: f.answers})
	}
	if castOfEvents(evs) != nil && reportHeadOf(evs, ids) == 0 {
		out = append(out, Blocker{Kind: BlockerNoReport})
	}
	out = append(out, unruledMotionBlockers(evs, ids, win)...)
	out = append(out, upheldPositionBlockers(evs, ids, win)...)
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
// material gaps, the open gaps and the bench's remands.
func passBlockerReadsAt(q recordsql.Querier) (planReads, error) {
	var r planReads
	var err error
	if r.evs, r.win, err = eventsAt(q); err != nil {
		return r, err
	}
	if r.fresh, err = freshMaterialOf(q); err != nil {
		return r, err
	}
	if r.gaps, err = openGaps(q); err != nil {
		return r, err
	}
	r.remands, err = remandRulingsOf(q)
	return r, err
}

// passBlockersOfReads folds the gate's list from what passBlockerReadsAt read.
func passBlockersOfReads(r planReads) []Blocker {
	ids := r.win.IDs(r.evs)
	return passBlockersOf(r.evs, ids, r.win, blockerGapsOfOpen(r.gaps, remandsOwedOf(r.evs, ids, r.win, r.remands)), r.fresh)
}

// requireNoBlockers is the verdict gate. A PASS is refused over every blocker. A FAIL is then held to
// the convergence rule, which stands only while nothing another seat must act on holds the PASS — so
// its "or issue `--as PASS`" is true apart from the chair's own items.
//
// AN ACT JUDGEMENT (judgeAct; gblock, 2026-10-24): an archived verdict is carried as history. The
// archived b7 and b9 runs' PASSes stand over gaps their classes now make material, and the replay
// writes them as the chair issued them.
func requireNoBlockers(w writeView, verdict recordpb.Verdict, blockers []Blocker) error {
	if verdict == recordpb.Verdict_VERDICT_PASS && len(blockers) > 0 {
		return blockerRefusal(blockers)
	}
	if verdict == recordpb.Verdict_VERDICT_FAIL {
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
//
// AN ACT JUDGEMENT (judgeAct): an archived run's positions were argued to a bench that was never
// convened for them, and its outcome is what the run did. The migrated record holds the position
// unruled.
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
// kind's paragraph once. It is a Conflict: the verdict contradicts the board, and a caller reading
// the refusal envelope tells it from a tool failure by the code, not the sentence.
func blockerRefusal(holding []Blocker) error {
	head := "PASS refused — "
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
		return feov.Errorf(feov.Conflict, "%s%s", head, paras[0])
	}
	for i := range paras {
		paras[i] = fmt.Sprintf("%d. %s", i+1, paras[i])
	}
	return feov.Errorf(feov.Conflict, "%s%d things hold it, in %d kinds, each with what clears it:\n\n%s", head, len(holding), len(paras), strings.Join(paras, "\n\n"))
}
