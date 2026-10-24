package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// duplicateScore is where the mint's own screen stops a mint and asks. CALIBRATED ON THE ONE
// MEASURED DUPLICATE, not chosen: on universe-m13, G7 and G8 — one semiprime-ordinal defect minted
// by two lenses ten seconds apart — scored 0.42; of the run's 55 gap pairs, three scored 0.40 or
// more. A lexical overlap cannot tell a duplicate from a neighbour (a distinct pair scored 0.47),
// so the screen does not decide: it refuses once and makes the seat say which it is.
const duplicateScore = 0.40

// requireNotAnOpenGap is `near-match` run BY THE MINT, at the write.
//
// The screen was a separate verb a lens runs first, and on universe-m13 it was run: at 01:53:56,
// four seconds before the other lens minted G7. G8 went on the board at 01:54:10 with G7 already
// there, and the lens that minted it saw G7 22 seconds later and moved on. A screen that answers
// before the board changes is a screen of a board that no longer exists; the one place the answer
// is current is here, where the mint is about to land.
//
// A match is answered, not overridden: --supersedes if the new gap replaces it, --distinct-from if
// the seat read it and it is a different defect — recorded on the mint, so the claim is auditable.
//
// THE SCREEN READS THE MINT'S OWN TRANSACTION (writeView). The m13 lens screened four seconds before
// the other lens's gap existed; a screen read before the write's transaction opens has the same hole
// narrowed to the instant between its read and its insert, and two lenses minting one defect at once
// both land through it.
func requireNotAnOpenGap(q recordsql.Querier, seatID string, m *recordpb.Mint) error {
	// THE SCREEN JUDGES WHAT A SEAT MAY DO NEXT; an archived mint is what a seat DID. A mint recorded
	// before the screen existed carries neither answer and can be given none by a replay, so under
	// Migrating the match below is not asked — only the distinction check, which is structure (a
	// reference to a gap nobody minted), stays on. b8 (2026-09-11) holds a report_voice mint scoring
	// 0.56 against an open gap of another class, and ten events that name it fall with it if this
	// asks (#1214). A replayed mint with no distinction to check leaves before the board is read:
	// FamilyOf is a full-record read, once per mint of the archive.
	if Migrating && len(m.GetDistinctFrom()) == 0 {
		return nil
	}
	// AN EMPTY BOARD IS AN EMPTY FAMILY, NOT AN ERROR: the first mint's ordinary state reads back
	// as no gaps. An error here is a record that could not be read, and a mint admitted over it is
	// a mint screened against no board at all — the read failure and the honest empty board must
	// not share an answer.
	f, err := familyAt(q)
	if err != nil {
		// CODED AND SAID AS THE SCREEN'S REFUSAL: the read's own sentence, bare, is about whatever
		// row it tripped on and carries no code, and a seat reading it cannot tell that its mint is
		// what was turned away. The renderer names the command; this says what the command could
		// not do. Conflict is the record contradicting itself, which is what a fold that cannot
		// complete is; the cause stays reachable through Unwrap.
		return feov.Wrap(feov.Conflict, err, "the board could not be read, so the screen cannot run: %v", err)
	}
	// A DISTINCTION IS FROM A GAP THAT EXISTS. The field is a seat's claim that it read the gap; a
	// claim about a gap the board has never held is a claim about nothing.
	for _, id := range m.GetDistinctFrom() {
		if f.Gap(id) == nil {
			return feov.Errorf(feov.Validation,
				"--distinct-from %s names no gap on the board; it answers the duplicate screen, so it names a gap the screen matched", id)
		}
	}
	if Migrating {
		return nil
	}
	answered := map[string]bool{}
	for _, id := range append(append([]string{}, m.GetSupersedes()...), m.GetDistinctFrom()...) {
		answered[id] = true
	}
	var hits, hitIDs []string
	for _, nm := range NearMatch(f, m.GetProblem(), m.GetLocation(), 0) {
		if nm.Status != "open" || nm.Score < duplicateScore || answered[nm.ID] {
			continue
		}
		problem := ""
		if g := f.Gap(nm.ID); g != nil && g.Mint != nil {
			problem = g.Mint.GetProblem()
			if len([]rune(problem)) > 200 {
				problem = string([]rune(problem)[:200]) + "…"
			}
		}
		var minter string
		if _, err := queryRowAt(q, []any{&minter}, `SELECT COALESCE("minted_by", '') FROM "gap" WHERE "gap_id" = ?`, nm.ID); err != nil {
			return err
		}
		hitIDs = append(hitIDs, nm.ID)
		hits = append(hits, fmt.Sprintf("%s (open, minted by %s, overlap %.2f): %s", nm.ID, minter, nm.Score, problem))
	}
	if len(hits) == 0 {
		return nil
	}
	// AN ANSWERING MINT THE SCREEN STOPS IS ANSWERED BY THE GAP IT MATCHED (gblock, 2026-10-24).
	// The lens owes the gap that answers what its finding raised; where the board already holds that
	// defect open, the answer is that gap, credited with the raising finding — no duplicate is minted.
	// On universe-m32 this refusal fell seven times on three raises whose defect one open gap held,
	// and six sittings ended owing an answer the refusal never named.
	raise, err := ownRaiseAt(q, seatID, m.GetFoundBy())
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's contradictions could not be read, so the screen cannot tell whether this mint answers one: %v", err)
	}
	if raise != nil {
		gap := "<id>"
		if len(hitIDs) == 1 {
			gap = hitIDs[0]
		}
		return feov.Errorf(feov.Validation,
			"this gap would answer contradiction %s, which your finding %s raised, and the board already holds an open gap this may be:\n  %s\n"+
				"Read each. If one IS this defect, do not mint it again: credit that gap with your finding — %s — and it answers the contradiction, with no mint taken. "+
				"If this gap replaces one, mint with --supersedes <id>. If you read it and this is a different defect, mint with --distinct-from <id> — "+
				"that is your claim, on the record, that they differ.",
			raise.idWord(), raise.raisedBy, strings.Join(hits, "\n  "), creditAct(gap, raise.raisedBy))
	}
	return feov.Errorf(feov.Validation,
		"the board already holds an open gap this may be:\n  %s\n"+
			"Read each. If one IS this defect, do not mint it again: file a finding about that gap, which reaches the seat that minted it. "+
			"If this gap replaces one, mint with --supersedes <id>. If you read it and this is a different defect, mint with --distinct-from <id> — "+
			"that is your claim, on the record, that they differ.",
		strings.Join(hits, "\n  "))
}
