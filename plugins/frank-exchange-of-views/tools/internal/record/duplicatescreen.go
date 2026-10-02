package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
func requireNotAnOpenGap(run Run, m *recordpb.Mint) error {
	f, err := FamilyOf(run)
	if err != nil {
		return nil // no board yet is the first mint's ordinary state, not a duplicate
	}
	// A DISTINCTION IS FROM A GAP THAT EXISTS. The field is a seat's claim that it read the gap; a
	// claim about a gap the board has never held is a claim about nothing.
	for _, id := range m.GetDistinctFrom() {
		if f.Gap(id) == nil {
			return feov.Errorf(feov.Validation,
				"record: mint refused — --distinct-from %s names no gap on the board; it answers the duplicate screen, so it names a gap the screen matched", id)
		}
	}
	// THE SCREEN JUDGES WHAT A SEAT MAY DO NEXT; an archived mint is what a seat DID. A mint recorded
	// before the screen existed carries neither answer and can be given none by a replay, so under
	// Migrating the match is not asked — the distinction check above is structure and stays on. b8
	// (2026-09-11) holds a report-voice mint scoring 0.56 against an open gap of another class, and
	// ten events that name it fall with it if this asks (#1214).
	if Migrating {
		return nil
	}
	answered := map[string]bool{}
	for _, id := range append(append([]string{}, m.GetSupersedes()...), m.GetDistinctFrom()...) {
		answered[id] = true
	}
	var hits []string
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
		if _, err := queryRow(run, []any{&minter}, `SELECT COALESCE("minted_by", '') FROM "gap" WHERE "gap_id" = ?`, nm.ID); err != nil {
			return err
		}
		hits = append(hits, fmt.Sprintf("%s (open, minted by %s, overlap %.2f): %s", nm.ID, minter, nm.Score, problem))
	}
	if len(hits) == 0 {
		return nil
	}
	return feov.Errorf(feov.Validation,
		"record: mint refused — the board already holds an open gap this may be:\n  %s\n"+
			"Read each. If one IS this defect, do not mint it again: file a finding about that gap, which reaches the seat that minted it. "+
			"If this gap replaces one, mint with --supersedes <id>. If you read it and this is a different defect, mint with --distinct-from <id> — "+
			"that is your claim, on the record, that they differ.",
		strings.Join(hits, "\n  "))
}
