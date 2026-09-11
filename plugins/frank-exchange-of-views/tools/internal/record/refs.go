package record

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY REFERENCE IS CHECKED AGAINST THE THING IT REFERENCES.
//
// This file exists because an audit of the whole surface found TWELVE OF TWELVE
// cross-references unvalidated. Only `close --id` and `mint --supersedes` were checked;
// every other field naming another entity — a gap, an observation, a finding, a seat —
// was accepted on the seat's say-so.
//
// That is not a hypothetical. It is the mechanism behind the 2026-07-18 run's worst
// damage: eight judicial closures were `opinion --id` events ACCEPTED AT WRITE TIME and
// then silently DROPPED at replay, because replay checks what the write path did not. The
// board was wrong by six gaps for three epochs and nothing on the surface said so.
//
// The split is the defect. A reference that replay will refuse must be refused when it is
// written, while the seat is still there to fix it — an event accepted into the log and
// discarded on the way out is the worst of both: it looks recorded and it does nothing.
//
// COST, MEASURED BEFORE ENFORCING: every reference in the run resolves — 6 successors,
// 20 fold-into targets, all minted; and zero forward references within a shard, so no
// seat named something it had not yet created. This would have refused nothing.

// requireGap refuses a reference to a gap no mint created.
// THE EXPORTED FORMS, for flags that carry their own existence check.
//
// A typed flag declares what it is checked against (`flags.GapID().WithCheck(record.GapExists)`),
// and seat.Begin runs it once the run directory is resolved. These are THIN WRAPPERS over the
// same helpers `validate` uses — one implementation, two call sites — because validate remains
// the enforcer: it is the single write path every caller goes through, and a rule the CLI held
// alone would be one every other caller skips. Two enforcers drifting is the defect this
// codebase keeps finding; two callers of one enforcer is not.
//
// The subject is named as the FLAG, not the verb: this refusal arrives while the seat is looking
// at one command, so "--id names gap R9-9, which no mint event created" is the whole sentence it
// needs. validate's copy keeps the verb too, because there it can be reached by any caller.
func GapExists(runDir string, id string) error {
	run, err := OpenRun(runDir)
	if err != nil {
		return err
	}
	return requireGap(run, id, "the", "--id")
}

// InquiryExists resolves a inquiry.
func InquiryExists(runDir string, id string) error {
	run, err := OpenRun(runDir)
	if err != nil {
		return err
	}
	return requireInquiry(run, id, "the", "--id")
}

// THESE THREE KEEP A `runDir string`, AND IT IS NOT AN OVERSIGHT.
//
// They are flags.Checker values — `flags.GapID().WithCheck(record.GapExists)` — and
// flags.Checker cannot name record.Run, because `record` imports `flags` and the reverse edge
// is an import cycle. So the string survives at exactly the boundary where the flag machinery
// hands one over, and each resolves it immediately: OpenRun still refuses a path nobody
// dispatched, so the check gains the refusal even though the parameter did not.
//
// The alternative is real but larger than this change: move the flag-NAME constants `record`
// reaches for into a leaf package, which breaks the cycle and lets Checker take a Run.
// CitationExists resolves a citation anchor.
func CitationExists(runDir string, label string) error {
	run, err := OpenRun(runDir)
	if err != nil {
		return err
	}
	return requireCitation(run, label, "the", "--cites")
}

func requireGap(run Run, id, verb, flag string) error {
	if id == "" {
		return nil
	}
	ids, err := allGapIDs(run)
	if err != nil {
		return err
	}
	if !ids[id] {
		return fmt.Errorf("record: %s %s names gap %s, which no mint event created — a dangling reference is accepted here and DROPPED at replay, which is how eight judicial closures vanished from a board that went on reporting them open", verb, flag, id)
	}
	return nil
}

// requireGaps is the list form, naming the first that does not resolve.
func requireGaps(run Run, ids []string, verb, flag string) error {
	for _, id := range ids {
		if err := requireGap(run, id, verb, flag); err != nil {
			return err
		}
	}
	return nil
}

// requireFindings refuses found_by attribution to a finding that does not exist.
//
// found_by is the credit chain from a lens's work to the board gap it earned, and it is
// read back by the capture-recapture estimate. An invented label inflates the count of
// lens-sourced gaps with a finding nobody made.
func requireFindings(run Run, labels []string, verb, flag string) error {
	if len(labels) == 0 {
		return nil
	}
	// One existence question per named label, in the caller's order, so the refusal names the
	// FIRST missing one exactly as the set-membership walk did. The list is bounded by what a
	// seat passes to --found-by, not by the record's size.
	for _, l := range labels {
		found, err := recordHas(run, `SELECT 1 FROM "finding" WHERE "label" = ? LIMIT 1`, l)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("record: %s %s names finding %s, which no lens recorded — attribution to a finding nobody made inflates the lens-sourced count the capture estimate reads", verb, flag, l)
		}
	}
	return nil
}

// requireSeat refuses attribution to a seat that never registered.
// requireCitation refuses a citation label that names no `blue cite` on the record.
//
// `blue prove --cites` names the METHOD a computation applies — "the source that says trial
// division decides primality". It was set into the payload and never checked, so a proof could
// cite a citation that does not exist and the assembled report would carry the link as though it
// meant something. That is the same hole `lens verify --anchor` had until 0.60.0, on the same
// axis, one verb over.
func requireCitation(run Run, label, verb, flag string) error {
	if label == "" {
		return nil
	}
	known, err := CitationLabels(run)
	if err != nil {
		return err
	}
	for _, k := range known {
		if k == label {
			return nil
		}
	}
	// "blue has cited" WAS TRUE AND IS NOT. The set now includes red's supporting
	// corroborations, which mint a label of their own — so a count described as blue's would
	// misstate what a seat is being compared against.
	return fmt.Errorf("record: %s %s=%s names no citation on the record — %d source(s) are cited (blue's, and red's corroborations), and `show evidence` lists them by anchor. Cite the method with `cite` first; a proof pointing at a citation that does not exist claims a provenance it does not have",
		verb, flag, label, len(known))
}

// requireInquiry refuses a move against a line of inquiry nobody proposed.
//
// `blue line-of-inquiry --id` required only that an id be PRESENT. A move naming an unknown line of inquiry wrote
// a status change for a line of inquiry that was never opened — and the lines-of-inquiry view
// renders it, so the run shows a direction being abandoned that nothing ever proposed.
func requireInquiry(run Run, id, verb, flag string) error {
	if id == "" {
		return nil
	}
	// Inquiries(b) keys every line of inquiry on Avenue events carrying a non-empty avenue_id,
	// so membership is one existence question.
	found, err := recordHas(run, `SELECT 1 FROM "avenue" WHERE "avenue_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("record: %s %s=%s names no line of inquiry on the record — `show lines-of-inquiry` lists every one with its id and fate. Propose it first (`blue line of inquiry --line …`, which ASSIGNS the id); --id moves a line of inquiry that already exists",
		verb, flag, id)
}

func requireSeat(run Run, seatID, verb, flag string) error {
	if seatID == "" {
		return nil
	}
	found, err := recordHas(run, `SELECT 1 FROM "events" WHERE "seat_id" = ? LIMIT 1`, seatID)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("record: %s %s names seat %s, which has recorded nothing in this run — a ruling attributed to a seat that never sat cannot be matched to the petition it answers", verb, flag, seatID)
}

// ---- STATE, not just existence ----
//
// A reference can resolve and still be wrong: the thing exists, but it is not in a state
// where the act makes sense. Regrading a closed gap moves a number nobody will read;
// answering a dispute nobody filed records half an argument; re-closing a closed gap
// double-counts closure history and, per replay.go, corrupts the repair_regression
// denominator that the whole repair metric divides by.
//
// EVERY RULE BELOW WAS MEASURED AGAINST THE 2026-07-18 RUN FIRST, and one candidate was
// DROPPED because of it: "supersedes must name a closed gap" sounded right and would have
// refused NINE OF NINE real mints — superseding a still-open predecessor is the normal
// pattern, not an error. The rules that survived had zero violations in the run, so they
// cost nothing today and exist for the cheaper tier.

// gapState answers what a gap IS, not merely whether it exists. The gap view's "open" is read
// off the disposition vocabulary (a red close OR the earliest closing docket ruling), which is
// the same rule the fold applied — an unknown gap answers not-closed, as before: requireGap owns
// existence.
func gapState(run Run, id string) (closed bool, err error) {
	var open bool
	found, err := queryRow(run, []any{&open}, `SELECT "open" FROM "gap" WHERE "gap_id" = ?`, id)
	if err != nil || !found {
		return false, err
	}
	return !open, nil
}

// requireOpenGap refuses an act that only makes sense on a live gap.
func requireOpenGap(run Run, id, verb, flag, why string) error {
	if id == "" {
		return nil
	}
	closed, err := gapState(run, id)
	if err != nil {
		return err
	}
	if closed {
		return fmt.Errorf("record: %s %s names gap %s, which is already CLOSED — %s", verb, flag, id, why)
	}
	return nil
}

// requireClosedGap is the mirror: spot-check re-verifies the ARCHIVE, so naming a gap that
// was never closed is not a sample of anything.
func requireClosedGaps(run Run, ids []string, verb, flag string) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		closed, err := gapState(run, id)
		if err != nil {
			return err
		}
		if !closed {
			return fmt.Errorf("record: %s %s names gap %s, which is still OPEN — the spot-check samples ARCHIVED closures, and re-verifying something that was never closed discharges the duty without doing it", verb, flag, id)
		}
	}
	return nil
}

// requireSupersededAreClosed is a COMPLETION duty, checked at the seat's terminal act.
//
// INVESTIGATED RATHER THAN ASSUMED. "Superseding an open gap" happened 9 times in the
// 2026-07-18 run and I first read the frequency as proof it was intended. Tracing what each
// one actually did splits them in two:
//
//	7 are STRUCTURALLY REQUIRED. The protocol is mint-the-successor, then close the
//	  ancestor naming it — so the ancestor is necessarily still open at mint time, and it
//	  cannot be otherwise, because the closure has to name a successor that already exists.
//	2 are a DEFECT nobody caught. R3-1 superseded R2-1 and R2-5 and neither was ever
//	  closed, so all three finished OPEN on the board. The run reported 9 open gaps; 7
//	  were distinct defects and one was counted three times.
//
// That is why the rule is not "supersedes must name a closed gap" — that would refuse all
// 9, including the 7 the protocol demands. The duty is that a superseded ancestor must not
// still be open when the seat FINISHES: superseding is a promise to replace, and a promise
// kept open inflates every count the board reports.
//
// Checked at verdict because that is the seat's terminal act and the last moment it is
// still there to close them.
func requireSupersededAreClosed(run Run) error {
	// The gap view answers this whole: `stranded` is an open gap somebody promised to replace,
	// and superseded_by is the LAST gap that made the promise — the same last-writer answer the
	// fold's map produced. An ancestor named in supersedes but never minted has no gap row and
	// drops out, as it dropped out of the board lookup.
	db, err := openRunForRead(run)
	if err != nil {
		return err
	}
	if db == nil {
		return nil // no record yet: nothing superseded, nothing stranded
	}
	rows, err := db.Query(`SELECT "gap_id", "superseded_by" FROM "gap" WHERE "stranded"`)
	if err != nil {
		return fmt.Errorf("record: asking the record for stranded ancestors: %w", err)
	}
	defer rows.Close()
	var stranded []string
	for rows.Next() {
		var anc, successor string
		if err := rows.Scan(&anc, &successor); err != nil {
			return err
		}
		stranded = append(stranded, fmt.Sprintf("%s (superseded by %s)", anc, successor))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(stranded) == 0 {
		return nil
	}
	sort.Strings(stranded)
	return fmt.Errorf("record: verdict refused — %d superseded gap(s) are still OPEN: %s. Superseding is a promise to replace, and an ancestor left open is the same defect counted twice: the 2026-07-18 run finished reporting 9 open gaps of which 7 were distinct. Close each ancestor (--superseded-by names its replacement), or if it is genuinely still live, it was not superseded",
		len(stranded), strings.Join(stranded, ", "))
}

// requirePassClosesAllMaterialGaps refuses PASS while any open gap is MATERIAL — current severity
// at GRADE_MEDIUM or above (plans/roundless.md §III.B.2.1) — at the write path, so no verdict
// route can bypass it; requireSupersededAreClosed holds the lineage case. A FAIL is always
// allowed. The chair's work list names exactly these gaps (sitting.go), and dispatch's
// pass_permitted counts the same ones. Refusing over ANY open gap made "below material does not
// hold the gate" unreachable: a run minting one trifle per sitting could never pass. An open sub-material gap at PASS stays
// open on the board and the report lists it as open, below material, not certified against — not
// auto-disposed, not carried, not accepted; red's finding stays visible and the report says what
// it was not certified against.
func requirePassClosesAllMaterialGaps(run Run) error {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return err
	}
	rows, err := db.Query(`SELECT g."gap_id", COALESCE(gs."mass", 0.0) >= ?
	  FROM "gap" g LEFT JOIN "enum_grade" gs ON gs."value" = g."current_severity"
	  WHERE g."open" ORDER BY g."minted_event"`, material)
	if err != nil {
		return fmt.Errorf("record: asking the record for its open gaps: %w", err)
	}
	defer rows.Close()
	var open, trifles []string
	for rows.Next() {
		var id string
		var isMaterial bool
		if err := rows.Scan(&id, &isMaterial); err != nil {
			return err
		}
		if isMaterial {
			open = append(open, id)
		} else {
			trifles = append(trifles, id)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(open) != 0 {
		sort.Strings(open)
		return fmt.Errorf("record: verdict PASS refused — %d material gap(s) still OPEN: %s. PASS requires every material gap resolved through `close --id <id> --as repaired|defect_accepted|not_a_defect|defect_owed_elsewhere`; close them, or issue `--as FAIL`",
			len(open), strings.Join(open, ", "))
	}
	_ = trifles // below material: on the board, listed by the report, not holding the gate

	m, err := MergedEvents(run)
	if err != nil {
		return err
	}
	evs := m.Events
	var unruled []string
	for _, mo := range MotionsOf(evs) {
		if mo.Ruled() {
			continue
		}
		phrase, err := rulerPhrase(mo.Subject)
		if err != nil {
			return err
		}
		unruled = append(unruled, mo.ID+" ("+phrase+")")
	}
	if len(unruled) != 0 {
		sort.Strings(unruled)
		return fmt.Errorf("record: verdict PASS refused — %d motion(s) filed and never ruled: %s. "+
			"Read what each one asks with `show motions` (its `basis` is the filer's argument, which your ruling answers), "+
			"then rule it with `motion <subject> rule --id <id> --as <verdict> --reason \"...\"` — IF THE GAVEL NAMED ABOVE IS YOURS. "+
			"Where it is not, the ruling is not yours to make and not yours to wait for silently: issue `--as FAIL` so the sitting ends on the record and the seat that holds it can answer. "+
			"A motion is answered before the debate moves on, so a PASS over an unanswered ask claims a settlement that did not happen",
			len(unruled), strings.Join(unruled, ", "))
	}
	if InquiryReviewDueOf(evs) {
		return fmt.Errorf("record: verdict PASS refused — this epoch has no line-of-inquiry review. " +
			"READ THE REPORT ONCE (`show report`), list what the record claims this run investigated with " +
			"`show lines-of-inquiry`, and answer in one act: `inquiry-support --reason \"<what the report " +
			"says at those lines>\"`. Where a line's research is thin, missing or unsupported by the text, " +
			"MINT A GAP for it — the shortfall is an ordinary defect and gets the ordinary lifecycle; this " +
			"event only records that the read happened, because an absent review reads exactly like a sound " +
			"one. A PASS claims the report is sound, and its account of what this run investigated is part " +
			"of the report; record the review, or issue `--as FAIL`")
	}
	if open := unansweredContradictions(evs); len(open) > 0 {
		sort.Strings(open)
		return fmt.Errorf("record: verdict PASS refused — red read a source that CONTRADICTS or does not support %d claim(s), and no finding was ever raised about them:\n  %s\n"+
			"Each is red's own reading that the report says something its source does not. Raise it with `lens finding --quote \"<the claim>\" --reason \"<what the source actually says>\"` graded on every axis, so it enters the board with the lifecycle, the blue duty and the gate every other defect has. "+
			"Read them with `show evidence`. A PASS claims the report is sound; these say otherwise on the record. Raise them, or issue `--as FAIL`",
			len(open), strings.Join(open, "\n  "))
	}
	return nil
}

// gapNamedIn returns the first gap id from the board that appears as a WHOLE TOKEN in
// prose, or "". It exists for one refusal: `blue edit` prose that names the gap the edit
// answers while `--answers` is empty (validate, case "blue_edit").
//
// It matches against the BOARD, never against a pattern. A regex for "gap-id-shaped" would
// have to guess the shape (G5, G11 today) and would fire on any prose that happens to
// look like one — a version number, a section reference, a matrix cell. Membership in the
// set of ids some mint actually created is exact, needs no shape at all, and costs the same
// read requireGap already does.
//
// Tokens break on anything outside [A-Za-z0-9-], so "G5: Quantify…" yields "G5" with
// its trailing colon shed, and a longer id is never matched by a shorter one's prefix.
func gapNamedIn(run Run, prose string) (string, error) {
	if strings.TrimSpace(prose) == "" {
		return "", nil
	}
	ids, err := allGapIDs(run)
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", nil
	}
	for _, tok := range strings.FieldsFunc(prose, func(r rune) bool {
		return !(r == '-' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	}) {
		if ids[tok] {
			return tok, nil
		}
	}
	return "", nil
}

// rulerPhrase names a motion's subject and the seat holding its gavel, for a message that is
// REFUSING or LISTING that motion.
//
// ONE PHRASE, TWO SURFACES, BECAUSE THEY DESCRIBE ONE BLOCKAGE. The refusal
// (requirePassClosesAllMaterialGaps) and the sitting view (SittingOf) both tell a seat that a motion is
// unruled, and only the refusal named who could rule it. A seat reading "motion M1 was filed and
// never ruled" on its work list and "M1 (petition, ruled by the bench seat)" from the gate is
// being told two different things about one fact, and sitting.go's own header says what that
// costs: "a seat told it was finished by one surface and refused by another learns to trust
// neither".
//
// THE UNKNOWN SUBJECT IS STATED, NEVER EMPTY. SittingOf returns no error and cannot propagate the
// second failure mode, so this function must not hand it a blank name to interpolate: the natural
// shape, `"ruled by the " + ruler + " seat"` with an empty ruler, renders `ruled by the  seat` —
// a miss that looks like a typo rather than a broken lookup. An unknown subject says so in words.
// The err return carries only the case a caller CAN discharge: a known subject whose enum value
// declares no ruler, which is a schema defect rather than a bad input, and which
// TestEveryMotionSubjectNamesItsRuler already refuses at the descriptor.
func rulerPhrase(subject string) (string, error) {
	subj, known := MotionSubjectEnum(subject)
	if !known {
		return subject + ", a subject this binary does not know — it cannot say who rules it", nil
	}
	ruler, err := recordpb.SubjectRuler(subj)
	if err != nil {
		return "", err
	}
	return subject + ", ruled by the " + ruler + " seat", nil
}
