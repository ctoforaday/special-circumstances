package record

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EVERY REFERENCE IS CHECKED AGAINST THE THING IT REFERENCES.
//
// This file exists because an audit of the whole surface found TWELVE OF TWELVE
// cross-references unvalidated. Only `close --id` and `mint --supersedes` were checked;
// every other field naming another entity — a gap, a finding, a seat —
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
// The subject is named as the FLAG, never the verb, here and in validate's call of the same
// helper: one check answers several verbs, the renderer prints the one the seat typed, and
// "--id names gap R9-9, which no mint event created" is the whole sentence it needs.
func GapExists(runDir, flag, id string) error {
	db, err := recordOfDir(runDir)
	if err != nil {
		return err
	}
	return requireGap(db, id, flag)
}

// InquiryExists resolves an inquiry.
func InquiryExists(runDir, flag, id string) error {
	db, err := recordOfDir(runDir)
	if err != nil {
		return err
	}
	return requireInquiry(db, id, flag)
}

// recordOfDir is the record of the run in runDir, for a flag's check: the handle, or a nil one for a
// run that has recorded nothing (openRunForRead), which every reference check reads as no record.
func recordOfDir(runDir string) (recordsql.Querier, error) {
	run, err := OpenRun(runDir)
	if err != nil {
		return nil, err
	}
	return openRunForRead(run)
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
// CitationExists resolves a citation anchor, and its refusal ends with the CALLER's advice, because
// two verbs on two seats take a citation and want different words: a lens's `verify --anchor`
// names the citation it adjudicated, blue's `compute --cites` the method a computation applies. A
// refusal naming --cites and blue's `cite` to the lens that typed --anchor sends it to a flag its
// verb does not have.
func CitationExists(advice string) func(runDir, flag, label string) error {
	return func(runDir, flag, label string) error {
		db, err := recordOfDir(runDir)
		if err != nil {
			return err
		}
		return requireCitation(db, label, flag, advice)
	}
}

// ProofExists resolves a proof anchor: `reproduce --anchor` names the proof whose sentence it judges.
func ProofExists(runDir, flag, id string) error {
	db, err := recordOfDir(runDir)
	if err != nil {
		return err
	}
	_, err = requireProofAnchor(db, id, flag)
	return err
}

// requireProofAnchor is the proof at an anchor, by its script's sha — the one fact a check of that
// anchor derives from it. A proof anchor no Proof carries is refused by name.
func requireProofAnchor(q recordsql.Querier, id, flag string) (sha string, err error) {
	var s sql.NullString
	found, err := queryRowAt(q, []any{&s}, `SELECT "proof_sha" FROM "proof" WHERE "proof_id" = ? ORDER BY "event_id" LIMIT 1`, id)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("%s names %s, which no proof on this record carries — it takes the id inside a `%s` token in the report; `show evidence` lists every proof by its anchor", flag, id, anchor.Token("P-…"))
	}
	return s.String, nil
}

// requireAboutAnchor refuses an act about an anchor that names no citation or proof on the record:
// the anchor is checked at the write, as every other reference here is, so a gap about an anchor
// always reaches one.
func requireAboutAnchor(q recordsql.Querier, kind recordpb.AboutKind, ref string) error {
	if kind != recordpb.AboutKind_ABOUT_KIND_ANCHOR {
		return nil
	}
	found, err := queryRowAt(q, []any{new(int)}, `SELECT 1 FROM "proof" WHERE "proof_id" = ? UNION ALL SELECT 1 FROM "cite" WHERE "label" = ? UNION ALL SELECT 1 FROM "verify" WHERE "label" = ? AND COALESCE("claim", '') != '' LIMIT 1`, ref, ref, ref)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("--about names %s, which no citation or proof on this record carries — it takes the id inside a `%s` or `%s` token in the report; `show evidence` lists every citation and proof by its anchor", ref, anchor.Token("C-…"), anchor.Token("P-…"))
	}
	return nil
}

// ComputeCitesAdvice is what a proof naming no citation is told, by the flag check and by validate's
// copy of the same refusal alike: blue's own act, in blue's words.
const ComputeCitesAdvice = "The method is cited first, with blue's `cite`; a proof pointing at a citation that does not exist claims a provenance it does not have"

// THE REFERENCE AND STATE CHECKS BELOW ASK q, which on the write path is the writing transaction
// (writeView): a gap that closes, or is minted, between a check and the insert it admits is the
// race #1252 names.
func requireGap(q recordsql.Querier, id, flag string) error {
	if id == "" {
		return nil
	}
	ids, err := allGapIDs(q)
	if err != nil {
		return err
	}
	if !ids[id] {
		return fmt.Errorf("%s names gap %s, which no mint event created — a dangling reference is accepted here and DROPPED at replay, which is how eight judicial closures vanished from a board that went on reporting them open", flag, id)
	}
	return nil
}

// requireGaps is the list form, naming the first that does not resolve.
func requireGaps(q recordsql.Querier, ids []string, flag string) error {
	for _, id := range ids {
		if err := requireGap(q, id, flag); err != nil {
			return err
		}
	}
	return nil
}

// requireClosureAnchors refuses a closure's anchor the record holds no act for, and the gap the
// closure closes.
//
// WHAT A CLOSURE RESTS ON: an anchor of the report — a citation, a proof, a finding, another gap — or
// an inquiry, whose fate is a record act a closure can rest on as surely as a citation's reading.
// The gap being closed is none of these: a closure is verified WITH something other than the gap,
// and naming the gap itself attests to nothing a reader could check.
func requireClosureAnchors(q recordsql.Querier, gapID string, ids []string, flag string) error {
	for _, id := range ids {
		if id == gapID {
			return fmt.Errorf("%s names %s, the gap this closes — an anchor is what you verified the closure WITH, and a gap is not evidence of its own repair. Name the citations, proofs, findings, other gaps or inquiries you checked, or none", flag, id)
		}
		found, err := HoldsAnchor(q, id)
		if err != nil {
			return err
		}
		if !found {
			if found, err = recordHasAt(q, `SELECT 1 FROM "inquiry" WHERE "inquiry_id" = ? LIMIT 1`, id); err != nil {
				return err
			}
		}
		if !found {
			return fmt.Errorf("%s names %s, which no act on this record placed — it takes the id inside an anchor token in the report (%s) or an inquiry's id; `show evidence`, `show findings`, `show board` and `show inquiries` resolve them", flag, id, anchor.Token("C-…"))
		}
	}
	return nil
}

// HoldsAnchor is whether an act on q's record carries the anchor id: an anchor id is the id its
// placing act minted — a finding's, a gap's, a citation's, a proof's or a corroboration's.
func HoldsAnchor(q recordsql.Querier, id string) (bool, error) {
	return queryRowAt(q, []any{new(int)},
		`SELECT 1 FROM (SELECT "id" FROM "finding" UNION ALL SELECT "gap_id" FROM "mint"
		   UNION ALL SELECT "label" FROM "cite" UNION ALL SELECT "proof_id" FROM "proof"
		   UNION ALL SELECT "label" FROM "verify") WHERE "id" = ? LIMIT 1`, id)
}

// requireFindings refuses found_by attribution to a finding that does not exist.
//
// found_by is the chain from a lens's work to the board gap it earned, and it is
// read back by the capture-recapture estimate. An invented id inflates the count of
// lens-sourced gaps with a finding nobody made.
func requireFindings(q recordsql.Querier, ids []string, flag string) error {
	if len(ids) == 0 {
		return nil
	}
	// One existence question per named finding, in the caller's order, so the refusal names the
	// FIRST missing one exactly as the set-membership walk did. The list is bounded by what a
	// seat passes to --found-by, not by the record's size.
	for _, l := range ids {
		found, err := recordHasAt(q, `SELECT 1 FROM "finding" WHERE "id" = ? LIMIT 1`, l)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("%s names finding %s, which no lens recorded — attribution to a finding nobody made inflates the lens-sourced count the capture estimate reads", flag, l)
		}
	}
	return nil
}

// requireCitation refuses a citation label that names no `blue cite` on the record.
//
// `blue compute --cites` names the METHOD a computation applies — "the source that says trial
// division decides primality". It was set into the payload and never checked, so a proof could
// cite a citation that does not exist and the assembled report would carry the link as though it
// meant something. That is the same hole `lens verify --anchor` had until 0.60.0, on the same
// axis, one verb over.
func requireCitation(q recordsql.Querier, label, flag, advice string) error {
	if label == "" {
		return nil
	}
	known, err := citationLabelsAt(q)
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
	return fmt.Errorf("%s=%s names no citation on the record — %d source(s) are cited (blue's, and red's corroborations), and `show evidence` lists them by anchor. %s",
		flag, label, len(known), advice)
}

// requireInquiry refuses a fate recorded against an inquiry nobody proposed.
//
// An id that is merely PRESENT is not enough: a fate recorded on an unknown inquiry would write a
// status change for an inquiry that was never opened, and the inquiries view would render it, so
// the run would show an inquiry being abandoned that nothing ever proposed.
func requireInquiry(q recordsql.Querier, id, flag string) error {
	if id == "" {
		return nil
	}
	// InquiriesOf keys every inquiry on Inquiry events carrying a non-empty inquiry_id,
	// so membership is one existence question.
	found, err := recordHasAt(q, `SELECT 1 FROM "inquiry" WHERE "inquiry_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return fmt.Errorf("%s=%s names no inquiry on the record — `show inquiries` lists every one with its id and fate. Propose it first (`inquiry propose --question …`, which ASSIGNS the id); --id names an inquiry that already exists",
		flag, id)
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

// gapState answers what a gap IS, not merely whether it exists. The gap view's "open" means no
// closing act of either arm is on the record (closure_act: a red close, or a docket ruling whose
// disposition closes) — an unknown gap answers not-closed: requireGap owns existence.
func gapState(q recordsql.Querier, id string) (closed bool, err error) {
	var open bool
	found, err := queryRowAt(q, []any{&open}, `SELECT "open" FROM "gap" WHERE "gap_id" = ?`, id)
	if err != nil || !found {
		return false, err
	}
	return !open, nil
}

// requireOpenGap refuses an act that only makes sense on a live gap.
func requireOpenGap(q recordsql.Querier, id, flag, why string) error {
	if id == "" {
		return nil
	}
	closed, err := gapState(q, id)
	if err != nil {
		return err
	}
	if closed {
		return fmt.Errorf("%s names gap %s, which is already CLOSED — %s", flag, id, why)
	}
	return nil
}

// requireClosedGap is the mirror: spot-check re-verifies the ARCHIVE, so naming a gap that
// was never closed is not a sample of anything.
func requireClosedGaps(q recordsql.Querier, ids []string, flag string) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		closed, err := gapState(q, id)
		if err != nil {
			return err
		}
		if !closed {
			return fmt.Errorf("%s names gap %s, which is still OPEN — the spot-check samples ARCHIVED closures, and re-verifying something that was never closed discharges the duty without doing it", flag, id)
		}
	}
	return nil
}

// openMaterialGaps is the board's open MATERIAL gaps, sorted: the set the PASS gate refuses over
// and the set a migrated PASS records as admitted. One query, so the gate and the admission it
// is exempt from cannot disagree about which gaps they mean.
func openMaterialGaps(q recordsql.Querier) ([]string, error) {
	if noRecord(q) {
		return nil, nil
	}
	rows, err := q.Query(`SELECT g."gap_id" FROM "gap" g WHERE g."open" AND g."material" ORDER BY g."minted_event"`)
	if err != nil {
		return nil, fmt.Errorf("asking the record for its open gaps: %w", err)
	}
	defer rows.Close()
	var open []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		open = append(open, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(open)
	return open, nil
}

// stampMigrationAdmission writes, on a PASS a migration replays, the open gaps the exempt PASS
// gate would have refused over (Gate.migration_admitted_gap_ids). Outside Migrating any Gate
// carrying the field is refused: the fact is the migration's to state, and a live PASS that could
// state it could stand over any open gap with verify reporting nothing. Under Migrating a
// non-PASS carrying it is refused (a FAIL admits nothing), and a PASS whose source already carries
// it — a run this binary migrated before — must carry exactly what the board yields now: the
// gaps' class material is kept as recorded, so the two agree unless the record was altered.
func stampMigrationAdmission(q recordsql.Querier, g *recordpb.Gate) error {
	arrived := append([]string(nil), g.GetMigrationAdmittedGapIds()...)
	if !Migrating {
		if len(arrived) > 0 {
			return fmt.Errorf("it carries migration_admitted_gap_ids (%s), which only a migration writes: it records the open material gaps an archived PASS was admitted over. A live PASS closes its material gaps or is refused; record the verdict without it",
				strings.Join(arrived, ", "))
		}
		return nil
	}
	if g.GetVerdict() != recordpb.Verdict_VERDICT_PASS {
		if len(arrived) > 0 {
			return fmt.Errorf("the archived %s verdict carries migration_admitted_gap_ids (%s) — only a PASS is admitted over open gaps",
				recordpb.Word(g.GetVerdict()), strings.Join(arrived, ", "))
		}
		return nil
	}
	open, err := openMaterialGaps(q)
	if err != nil {
		return err
	}
	if len(arrived) > 0 {
		sort.Strings(arrived)
		if strings.Join(arrived, "\x00") != strings.Join(open, "\x00") {
			return fmt.Errorf("the archived PASS carries migration_admitted_gap_ids (%s), but the open material gaps at it are (%s) — the record was altered after it was written",
				strings.Join(arrived, ", "), strings.Join(open, ", "))
		}
	}
	g.MigrationAdmittedGapIds = open
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
func gapNamedIn(q recordsql.Querier, prose string) (string, error) {
	if strings.TrimSpace(prose) == "" {
		return "", nil
	}
	ids, err := allGapIDs(q)
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

// gavel is who rules a motion: the seat holding its gavel ("" for a subject this binary does not
// know), and the phrase a message that is REFUSING or LISTING the motion names it by.
type gavel struct {
	seat, phrase string
}

// gavelOf resolves a motion subject to its gavel: the schema's `ruled_by` role, through the role
// roster (gavelSeatOf) to the one seat that holds it.
//
// ONE PHRASE, EVERY SURFACE, BECAUSE THEY DESCRIBE ONE BLOCKAGE. The gate's refusal, the chair's
// and the bench's work lists and the outcome refusal all tell a seat that a motion is unruled, off
// one Blocker (passblockers.go). A seat reading "motion M1 was filed and never ruled" on its work
// list and "M1 (petition, ruled by the bench seat)" from the gate is being told two different
// things about one fact, and sitting.go's own header says what that costs: "a seat told it was
// finished by one surface and refused by another learns to trust neither".
//
// THE UNKNOWN SUBJECT IS STATED, NEVER EMPTY: the natural shape, `"ruled by the " + ruler + "
// seat"` with an empty ruler, renders `ruled by the  seat` — a miss that looks like a typo rather
// than a broken lookup. The err return carries the schema defects: a known subject whose enum value
// declares no ruler, or declares a role no single seat holds. TestEveryMotionSubjectNamesItsRuler
// refuses both at the descriptor.
func gavelOf(subject string) (gavel, error) {
	subj, known := MotionSubjectEnum(subject)
	if !known {
		return gavel{phrase: subject + ", a subject this binary does not know — it cannot say who rules it"}, nil
	}
	ruler, err := recordpb.SubjectRuler(subj)
	if err != nil {
		return gavel{}, err
	}
	seat, err := gavelSeatOf(ruler)
	if err != nil {
		return gavel{}, fmt.Errorf("motion subject %s: %w", subject, err)
	}
	return gavel{seat: seat, phrase: subject + ", ruled by the " + ruler + " seat"}, nil
}
