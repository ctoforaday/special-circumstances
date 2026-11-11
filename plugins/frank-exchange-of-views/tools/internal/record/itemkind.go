package record

// WHAT AN ITEM IS ABOUT, AS FIELDS (one-item work list, gblock 2026-11-06).
//
// An item's sentence names its gap, anchor or motion, and nothing else did: a reader that wanted to
// know which gap an item was about matched the id out of `what`, and the release fuzz did exactly
// that (strings.Contains over the sentence). A lens handed its work one item at a time needs each
// item's identity as a key — to say which item is current, to tell an item that changed from one
// that is new, and to see one leave its list — and a key recovered from prose is the pattern
// facts-are-fields refuses. So every item carries its kind and its subject, set where the item is
// built, from the record fields the sentence is built from.

// ItemKind is what an item on a seat's work list is: a PASS blocker's kind where the item states
// one (BlockerKind), and otherwise one of the kinds below.
type ItemKind string

// The kinds of item that state no PASS blocker. Each names the item's condition, the way a
// BlockerKind does, and an item's kind is one of these or a BlockerKind — never a second word for a
// blocker.
const (
	// ItemAudit is a lens's audit of the report: its first read, its read since it was last served
	// the report whole, or the sitting's read once served (LastReadJSON.Kind says which).
	ItemAudit ItemKind = "audit"
	// ItemReaudit is a lens's own open gap that blue answered, or whose sentence an edit changed,
	// since the lens last acted on it — or that blue was engaged on and left unanswered (blueAnswers).
	ItemReaudit ItemKind = "reaudit"
	// ItemRegradeOwed is an accepted grade motion on a lens's own gap that no regrade has followed.
	ItemRegradeOwed ItemKind = "regrade_owed"
	// ItemFindingOnGap is another seat's finding anchored to a lens's own open gap.
	ItemFindingOnGap ItemKind = "finding_on_gap"
	// ItemNews is a closed gap of a lens's own whose answers have all been retired: the closure
	// stands, and the item says what a mint in answer would be.
	ItemNews ItemKind = "news"
	// ItemLog is the open log channel.
	ItemLog ItemKind = "log"
	// ItemLaw is whether law is staged for the run, on the lists of the two seats that rule.
	ItemLaw ItemKind = "law"
	// ItemRegisterOwed is the sitting a seat was dispatched for, or the chair's, not yet opened.
	ItemRegisterOwed ItemKind = "register_owed"
	// ItemVerdictOwed is the chair's terminal act, not yet recorded or owed after an overruled position.
	ItemVerdictOwed ItemKind = "verdict_owed"
	// ItemAwaitingEvidence is a gap owed a citation or computation placed as its answer
	// (WorkGapJSON.AwaitingEvidence).
	ItemAwaitingEvidence ItemKind = "awaiting_evidence"
	// ItemSuccessorOwed is blue-respond's engage on the open successors of gaps it found closed
	// (EngagedJSON.SuccessorsOwed).
	ItemSuccessorOwed ItemKind = "successor_owed"
	// ItemNotMaterial is an open gap that does not hold the PASS, on the chair's list.
	ItemNotMaterial ItemKind = "not_material"
	// ItemDocket is the chair's affordance to docket an open material gap.
	ItemDocket ItemKind = "docket"
	// ItemStaleInquiry is an inquiry no fate has revisited since it was declared.
	ItemStaleInquiry ItemKind = "stale_inquiry"
	// ItemKeptGrade is a gap whose lens kept its grade on re-audit, with what its check still finds
	// failing, on blue's list.
	ItemKeptGrade ItemKind = "kept_grade"
	// ItemComposedSource is a record whose text a composed section prints — an inquiry's latest fate,
	// an open gap's problem and fix — that the voice lens has not had since it was recorded
	// (composedsource.go).
	ItemComposedSource ItemKind = "composed_source"
	// ItemNothingOwed is the statement that a seat's list is complete, not empty.
	ItemNothingOwed ItemKind = "nothing_owed"
)

// ItemKinds is every kind an item can carry: the PASS blockers' kinds, in table order, then the
// kinds above.
func ItemKinds() []ItemKind {
	out := make([]ItemKind, 0, len(blockerKinds)+20)
	for _, r := range blockerKinds {
		out = append(out, ItemKind(r.kind))
	}
	return append(out, ItemAudit, ItemReaudit, ItemRegradeOwed, ItemFindingOnGap, ItemNews, ItemLog,
		ItemLaw, ItemRegisterOwed, ItemVerdictOwed, ItemAwaitingEvidence, ItemSuccessorOwed,
		ItemNotMaterial, ItemDocket, ItemStaleInquiry, ItemKeptGrade, ItemComposedSource, ItemNothingOwed)
}

// SubjectKind is what kind of thing an item is about.
type SubjectKind string

const (
	SubjectGap     SubjectKind = "gap"
	SubjectAnchor  SubjectKind = "anchor"
	SubjectInquiry SubjectKind = "inquiry"
	SubjectMotion  SubjectKind = "motion"
	// SubjectReport is the report as a whole; its id is the report head the item is about.
	SubjectReport SubjectKind = "report"
	// SubjectSeat is a lens seat (a ready lens), and SubjectArea a lens area (a stale area).
	SubjectSeat SubjectKind = "seat"
	SubjectArea SubjectKind = "area"
)

// ItemSubject is the one thing an item is about: a gap, an anchor, an inquiry, a motion, the report,
// a lens or an area, by its id. An item about nothing in particular — the log channel, the law,
// the register a seat owes — carries none.
type ItemSubject struct {
	Kind SubjectKind `json:"kind"`
	ID   string      `json:"id"`
}

// about is a subject of kind k with id, or nil where id is "": a blocker about the run as a whole
// (no_report) names no subject.
func about(k SubjectKind, id string) *ItemSubject {
	if id == "" && k != SubjectReport {
		return nil
	}
	return &ItemSubject{Kind: k, ID: id}
}

// gapItem is a non-blocking item of kind k about gap id.
func gapItem(k ItemKind, id, what string) Item {
	return Item{What: what, Kind: k, Subject: about(SubjectGap, id)}
}
