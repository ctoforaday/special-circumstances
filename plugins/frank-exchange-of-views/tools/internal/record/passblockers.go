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
	BlockerRemandOwed    BlockerKind = "remand_owed"
	BlockerContradiction BlockerKind = "contradiction"
	BlockerUnchecked     BlockerKind = "unchecked"
	BlockerFailedCheck   BlockerKind = "failed_check"
	BlockerNoReport      BlockerKind = "no_report"
	BlockerLensReady     BlockerKind = "lens_ready"
	BlockerStaleArea     BlockerKind = "stale_area"
	BlockerGradeMovement BlockerKind = "grade_movement"
	BlockerClosingOwed   BlockerKind = "closing_owed"
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
	// subject and gavel, a stale area's pin, an unchecked or failed-check proof's sha256 (""
	// for a citation), a contradiction's claim.
	Detail string
	// About is the gap an unruled motion is about, "" for a motion about no gap.
	About string
	// foldWhy is a ready lens's reason in the lens fold's own words, for the refusal.
	foldWhy string
	// motionSubject is an unruled motion's subject — which sitting dispatch convenes its gavel for.
	motionSubject string
	// since is the events.id the blocker arose at — a motion's filing, the contradicting verify,
	// the edit that changed an unchecked anchor's sentence (0 for one no edit has changed), the
	// failed check — so dispatch can ask whether its owner has sat since, and an unchecked anchor's
	// wording can name its edit; 0 for the other kinds.
	since int64
	// placed is the events.id of the act that put an unchecked anchor on the record, which dispatch
	// readies its owner off (a cite or a proof moves no report head); 0 for every other blocker.
	placed int64
	// lastCheck is the events.id of an unchecked anchor's latest check, which does not clear it, 0
	// when it has none.
	lastCheck int64
	// rerun is an unchecked proof's script state (unchecked.rerun), "" for every other blocker.
	rerun string
	// raisedBy is the finding that raised a contradiction, "" until one has, and raiser the seat
	// that recorded it.
	raisedBy, raiser string
	// answeredBy is the gap a failed check's seat minted about the anchor, "" until it has, and
	// answeredAt that mint's events.id.
	answeredBy string
	answeredAt int64
	// awaited is a closing blocker whose party has sat through no sitting opened since the
	// docketing (ClosingAwaited); false for a missing one, whose party sat and filed none.
	awaited bool
	// moved is a grade-movement blocker's gap's movement the bench has not ruled on, and totalMoved
	// the open gaps' total (gradeMovementOf).
	moved, totalMoved float64
}

// AnsweredBy is the gap a failed check's seat minted about the anchor, "" until it has — and for
// every other kind.
func (b Blocker) AnsweredBy() string { return b.answeredBy }

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
// contradictionAnswerBudget is what the answering gap costs its minter (requireMintWithinBudget).
const contradictionAnswerBudget = "The gap a lens mints from its own raising finding takes no mint of its budget: it is admitted at the cap"

// contradictionGapSubject is the subject the answering gap owes, as every gap does
// (requireMintSubject): the contradicted sentence while the report holds it, or its section once an
// edit removed it. Without it the --found-by mint is refused, and the lens loses the turn.
const contradictionGapSubject = "the mint names its subject as every gap does — `--quote` the claim's sentence while the report holds it, or `--about-kind section --about <its section>` once an edit has removed it"

const contradictionCitationAnswer = "A contradiction raised by reading a citation is answered as well by a retire of that citation, or once a lens has verified the citation's sentence as it now stands and every lens's latest verify of it backs it — the reading stops counting when its lens reads the citation's source again or an edit changes the sentence after its sitting opened, and another lens's backing verify while it counts is a disagreement that answers nothing"

// contradictionFact is a contradiction's name on every surface: the word and its id, which is the
// `--about` value a finding raises it by. A bare decimal would read as a count.
func contradictionFact(b Blocker) string { return "contradiction " + b.Subject }

// uncheckedFact is the one sentence every wording of an unchecked blocker states: which anchor, the
// edit that changed its sentence where one did, and its latest check — each as the record holds it,
// so a reader can go to each.
func uncheckedFact(b Blocker) string {
	switch {
	case b.since == 0 && b.Detail != "":
		return "proof " + b.Subject + " is on the record and nobody has judged whether its script establishes the sentence it backs"
	case b.since == 0:
		return "citation " + b.Subject + " is on the record and nobody has read its source against the sentence it backs"
	case b.lastCheck == 0:
		return fmt.Sprintf("%s's sentence changed at edit %d, and nobody has checked it", b.Subject, b.since)
	}
	return fmt.Sprintf("%s's sentence changed at edit %d and no sitting opened since has checked it (its last check is event %d)", b.Subject, b.since, b.lastCheck)
}

// uncheckedAct is the check an unchecked blocker is owed, in the order the act needs: a citation's
// source read against its sentence; a proof's script read and, where it has no re-run, re-run, then
// its sentence judged. A re-run belongs to the script, so a script with one is never re-run again,
// whatever it showed, and the act turns on that alone.
//
// THE PROOF ITEM CARRIES THE ID THE VERB TAKES: the proof's anchor, which `reproduce --anchor` takes
// and the seat reads in the report. When the verb took the script's sha256 and the item named the
// anchor, three lenses on universe-m11 spent 22 proof-file reads and 14 filesystem hunts translating
// one into the other, for 12 re-runs. AND IT SAYS WHERE THE SCRIPT IS READ FIRST: an item that handed
// over the re-run's id let a lens record a verdict on a script it had not read (wave A, 2026-10-09:
// three placeholder verdicts, each of which stands).
func uncheckedAct(b Blocker) string {
	if b.Detail == "" {
		return "reading its source against the sentence as it now stands (`verify --anchor " + b.Subject + "`)"
	}
	read := "`show evidence --anchor " + b.Subject + "` prints the script and its recorded output"
	if b.rerun == RerunNone || b.rerun == "" {
		return "READING its script and then RE-RUNNING it, never by reading alone: " + read + ", and `reproduce --anchor " + b.Subject + "` then re-runs it and records your verdict on its sentence in one act"
	}
	showed := "did not reproduce (" + b.rerun + ")"
	if b.rerun == recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED) {
		showed = "REPRODUCED"
	}
	return "judging whether its script establishes THIS sentence — the script was re-run and " + showed + ", and one script can back several sentences, so a verdict on another is not a verdict on this one: " +
		read + ", and `reproduce --anchor " + b.Subject + "` records your verdict on its sentence without re-running it"
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
	if b.Detail != "" {
		return fmt.Sprintf("%s judged the sentence proof %s backs unsound at event %d, and that judgement stands", ownerOr(b, failedChecker), b.Subject, b.since)
	}
	return fmt.Sprintf("%s's check of citation %s at event %d read nothing of its source — it was unreachable, or the copy read was not the work — and no check of the sentence as it stands has read it", ownerOr(b, failedChecker), b.Subject, b.since)
}

// failedClears is what clears a failed-check blocker, stated by every wording of it.
func failedClears(b Blocker) string {
	if b.Detail != "" {
		return "it clears when the proof is judged sound again — by the judging lens's own later judgement, or by a judgement from a sitting opened after blue's edit changed the sentence, since an edit leaves it unchecked — or when blue retires it; another lens's `sound` beside a standing `unsound` is a disagreement and clears nothing"
	}
	return "it clears when a lens reads the source — any result but `unreachable`, of a copy that is the work — or when blue retires the citation, or edits its sentence, which leaves it unchecked; the gap that carries it does not close while it stands, and no bench ruling clears it"
}

// failedChecker is whose check failed when the record names no seat.
const failedChecker = "a lens"

// failedRemedy is blue's act on a failed check.
func failedRemedy(b Blocker) string {
	if b.Detail != "" {
		return "fixes the proof or retires it"
	}
	return "retires the citation or cites a source a lens can read"
}

// failedOwed is whose act is owed on a failed-check blocker: the gap of the seat whose check failed
// until it is minted, and blue's remedy.
func failedOwed(b Blocker) string {
	if b.answeredBy == "" {
		return fmt.Sprintf("%s owes its gap about the anchor (`mint --about-kind anchor --about %s`), which takes no mint of its budget, and blue %s", ownerOr(b, failedChecker), b.Subject, failedRemedy(b))
	}
	return fmt.Sprintf("gap %s carries it to blue, which %s", b.answeredBy, failedRemedy(b))
}

// blueFailedItem is blue-respond's work-list sentence for a failed-check blocker: the remedy it owes.
func blueFailedItem(b Blocker) string {
	carried := "the lens whose check failed has minted no gap about it yet"
	if b.answeredBy != "" {
		carried = "gap " + b.answeredBy + " states the defect and its fix"
	}
	if b.Detail != "" {
		return fmt.Sprintf("%s — PASS is refused while it does, and %s. Fix it — an edit that makes its sentence say what the script shows, which leaves the proof unchecked for a fresh judgement, or a new proof with this one retired — or retire the claim it backs", failedFact(b), carried)
	}
	retire := "Retire the citation"
	if b.answeredBy != "" {
		retire = "Retire the citation, answering gap " + b.answeredBy + ","
	}
	return fmt.Sprintf("%s — PASS is refused while it stands, and %s. %s — its sentence may stay word for word, backed by whatever else backs it — or cite a source a lens can read and retire this one; an edit of its sentence leaves it unchecked, owed a fresh check", failedFact(b), carried, retire)
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
	// A REMAND OF A GAP THAT IS NOT MATERIAL HOLDS THE PASS UNTIL ITS EXCHANGE HAS BEGUN (ruled by
	// gblock 2026-10-24). A material gap holds it as material_gap
	// whatever its remand; this row is the one a gap below material holds once the bench has sent it
	// back, so the plan that readies its minting lens and blue never reads as permitting a PASS over
	// the exchange it readies them for (routeOf).
	{
		kind:  BlockerRemandOwed,
		gloss: "a gap that is not material, remanded by the bench, whose one more exchange has not been had",
		item: func(b Blocker) string {
			return fmt.Sprintf("gap %s is not material and the BENCH REMANDED it — PASS is refused until the one more exchange the remand grants, between %s and blue, has been had — an exchange that begins when blue records an act on the gap; the dispatch readies them on the ruling's direction", b.Subject, ownerOr(b, "the lens that minted it"))
		},
		refusal: func(bs []Blocker) string {
			return fmt.Sprintf("%d gap(s) the bench REMANDED are owed the one more exchange each remand grants: %s. The bench's remand outranks materiality, so a gap that is not material holds the PASS until its minting lens and blue have had that exchange — one that begins when blue records an act on the gap — on the ruling's direction; the dispatch readies them. PASS waits for it; `--as FAIL` does not",
				len(bs), strings.Join(subjects(bs), ", "))
		},
	},
	{
		kind:  BlockerContradiction,
		gloss: "a contradicting source no gap answers",
		item: func(b Blocker) string {
			if b.raisedBy != "" {
				return fmt.Sprintf("%s: red read a source that contradicts or does not support the claim %q; finding %s raised it, and no gap answers it — PASS is refused until one does; %s mints the gap with `mint --found-by %s`, and %s. A source the report misstates is a citation defect, so its gap is material by its class whatever reader decision you judge it to change. %s. %s", contradictionFact(b), b.Detail, b.raisedBy, ownerPhrase(b, "reader", contradictionReader), b.raisedBy, contradictionGapSubject, contradictionAnswerBudget, contradictionCitationAnswer)
			}
			return fmt.Sprintf("%s: red read a source that contradicts or does not support the claim %q and it is unanswered — PASS is refused while it is; %s raises it with `finding --about-kind contradiction --about %s`, whether or not the sentence still stands, and answers it with the gap minted from that finding (`mint --found-by <the finding's id>`; %s). A source the report misstates is a citation defect, so its gap is material by its class whatever reader decision you judge it to change. %s. %s", contradictionFact(b), b.Detail, ownerPhrase(b, "reader", contradictionReader), b.Subject, contradictionGapSubject, contradictionAnswerBudget, contradictionCitationAnswer)
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
				"Each is red's own reading that the report says something its source does not. The lens that read it raises it with the lens's `finding --about-kind contradiction --about <its id> --reason \"<what the source actually says>\"`, graded on every axis — only a finding naming it raises it; one whose `--quote` contains the claim raises nothing — and answers it with the gap it mints from that finding (the lens's `mint --found-by <the finding's id>`; "+contradictionGapSubject+"), so it enters the board with the lifecycle, the blue duty and the gate every other defect has. A finding alone answers nothing: a source the report misstates is a citation defect, material by its class. "+contradictionCitationAnswer+". "+
				"Read them with `show evidence`. A PASS claims the report is sound; these say otherwise on the record. Raising them is a lens's act; PASS waits for it, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s: a source %s read contradicts or does not support %q and it is unanswered — %s", contradictionFact(b), ownerOr(b, contradictionReader), b.Detail, planHold(b, "PASS is not permitted until it is answered"))
		},
	},
	// A CITATION OR PROOF NOBODY HAS CHECKED AS IT STANDS HOLDS THE PASS (unchecked.go). The owner is
	// the anchor's (anchorOwnerOf) whichever lens made any earlier check: a citation's is the evidence
	// lens, a proof's the computation lens. The owner owes the check, so it is not retired, and the
	// act that placed or changed the anchor moved the head past its last sitting: dispatch readies it.
	//
	// THE OWNER IS WHO IS READIED, NOT WHO ALONE MAY CLEAR. Any lens's check clears the blocker, so
	// each wording says both: the check that clears it, and the seat that owes it and is readied.
	{
		kind:  BlockerUnchecked,
		gloss: "a citation or proof with no check of the sentence the report carries",
		item: func(b Blocker) string {
			return fmt.Sprintf("%s — PASS is refused until a lens checks it%s, by %s. %s",
				uncheckedFact(b), uncheckedAfter(b), uncheckedAct(b), uncheckedOwed(b))
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				named = append(named, uncheckedFact(b)+" — "+ownerOr(b, uncheckedOwner)+" owes the check and is the seat readied for it")
			}
			return fmt.Sprintf("%d citation(s) or proof(s) have no check of the sentence the report carries:\n  %s\n"+
				"A citation nobody has read against its sentence, or a proof nobody has judged, is a claim nothing on the record backs yet; a check of a sentence an edit has since replaced is of a sentence the report no longer carries. `show evidence --anchor <id>` reads each as it now stands. "+
				"A `verify` of the citation that reads its source (any result but `unreachable`, of a copy that is the work) or a `reproduce` naming the proof's anchor — from a sitting opened after the edit, where an edit changed the sentence — clears it, whichever lens records it; "+
				"PASS waits for the check, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			return fmt.Sprintf("%s — %s", uncheckedFact(b), planHold(b, "PASS is not permitted until a lens checks it"+uncheckedAfter(b)+". "+uncheckedOwed(b)))
		},
	},
	// A CHECK THAT FAILED AND STANDS HOLDS THE PASS (failedcheck.go): a proof judged unsound, a
	// citation whose source no check could read. Its owner is the seat whose check failed, which owes
	// the gap that carries it to blue; blue's remedy, or a check that holds, clears it.
	{
		kind:  BlockerFailedCheck,
		gloss: "a citation whose source no check could read, or a proof a lens's standing judgement calls unsound",
		item: func(b Blocker) string {
			return fmt.Sprintf("%s — PASS is refused while it stands; %s. %s", failedFact(b), failedClears(b), failedOwed(b))
		},
		refusal: func(bs []Blocker) string {
			var named []string
			for _, b := range bs {
				gap := "no gap carries it yet"
				if b.answeredBy != "" {
					gap = "gap " + b.answeredBy + " carries it"
				}
				how := "read nothing of its source"
				if b.Detail != "" {
					how = "judged unsound"
				}
				named = append(named, fmt.Sprintf("%s (%s by %s at event %d; %s)", b.Subject, how, ownerOr(b, failedChecker), b.since, gap))
			}
			return fmt.Sprintf("%d citation(s) or proof(s) stand on a failed check:\n  %s\n"+
				"A PASS would certify a sentence whose proof the record calls unsound, or whose cited source nobody could read. The lens whose check failed mints its gap about the anchor (the lens's `mint --about-kind anchor --about <its anchor>`, which takes no mint of its budget), and blue fixes or retires the proof, or retires the citation or cites a source a lens can read. "+
				"A proof clears when it is judged sound again or retired, a citation when a lens reads its source or blue retires it, and either when blue's edit changes its sentence, which leaves it unchecked; the gap does not close while the check stands, and no bench ruling clears it. "+
				"`show evidence --anchor <id>` reads each check. PASS waits for that, and `--as FAIL` does not",
				len(bs), strings.Join(named, "\n  "))
		},
		why: func(b Blocker) string {
			until := "a lens reads its source or blue retires the citation"
			if b.Detail != "" {
				until = "the proof is judged sound again or retired"
			}
			return fmt.Sprintf("%s — %s", failedFact(b), planHold(b, "PASS is not permitted until "+until))
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
			return fmt.Sprintf("%d cast lens(es) are ready: %s. A lens is ready when it has never sat or its pin — the head the tool last served it the whole report at, in a sitting it sat through — is behind the head, until three sittings in a row with no fresh material mint retire it — and it is not retired while it owes: something new on a gap it minted since it last sat, a contradiction it read, a check it owns, the gap about a check of its that failed. A sitting stopped at the per-sitting call limit, or one that never closed, is not a sitting it sat: a read in it moves no pin, and it counts toward no retirement; `dispatch next` readies them",
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
	// latestRemandOwed is the gap's latest remand with no exchange since its ruling in which blue
	// acted on the gap (latestRemandOwed).
	latestRemandOwed bool
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
		out = append(out, blockerGap{id: g.ID, mintedBy: g.MintedBy, supersededBy: sup, material: g.Material, latestRemandOwed: g.latestRemandOwed})
	}
	return out
}

// blockerGapsOfOpen is the plan's open gaps as the list reads them, with owed the gaps whose latest
// remand's exchange is owed (remandsOwedOf) — nil where the reader asks only each gap's id and
// minter, as the lens fold does (lensOwesOf, owedBlockersOf).
func blockerGapsOfOpen(gaps []openGap, owed map[string]bool) []blockerGap {
	out := make([]blockerGap, 0, len(gaps))
	for _, g := range gaps {
		out = append(out, blockerGap{id: g.id, mintedBy: g.mintedBy, supersededBy: g.supersededBy, material: g.material, latestRemandOwed: owed[g.id]})
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
	for _, g := range gaps {
		if g.supersededBy == "" && !g.material && g.latestRemandOwed {
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
// contradiction no gap answers, a citation or proof with no check of the sentence the report
// carries, a standing failed check of one, no
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
		out = append(out, Blocker{Kind: BlockerContradiction, Subject: c.idWord(), Detail: c.claim, Owner: c.reader, since: c.since, raisedBy: c.raisedBy, raiser: c.raiser})
	}
	for _, u := range uncheckedAnchors(evs, win) {
		out = append(out, Blocker{Kind: BlockerUnchecked, Subject: u.anchor, Owner: anchorOwnerOf(u.proof), Detail: u.sha, since: u.edit, placed: u.placed, lastCheck: u.lastCheck, rerun: u.rerun})
	}
	for _, f := range failedChecksOf(evs, win) {
		out = append(out, Blocker{Kind: BlockerFailedCheck, Subject: f.anchor, Owner: f.owner, Detail: f.sha, since: f.since, answeredBy: f.answeredBy, answeredAt: f.answeredAt})
	}
	if castOfEvents(evs) != nil && reportHeadOf(evs, ids) == 0 {
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

// requireNoBlockers is the verdict gate. A PASS is refused over every blocker; any verdict, and
// any verdict a migration replays, over the kinds that hold every verdict. A FAIL is then held to
// the convergence rule, which stands only while nothing another seat must act on holds
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
		return feov.Errorf(feov.Conflict, "%s%s", head, paras[0])
	}
	for i := range paras {
		paras[i] = fmt.Sprintf("%d. %s", i+1, paras[i])
	}
	return feov.Errorf(feov.Conflict, "%s%d things hold it, in %d kinds, each with what clears it:\n\n%s", head, len(holding), len(paras), strings.Join(paras, "\n\n"))
}
