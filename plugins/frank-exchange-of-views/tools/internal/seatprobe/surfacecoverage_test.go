package seatprobe

import (
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// EVERY VERB A SEAT HAS IS DEMANDED BY SOME BOARD.
//
// This is the property the whole package exists to hold, and it is the one a verb-by-verb reading
// of the boards cannot give you: not "the boards look thorough" but "there is no verb in the tool
// that no scenario ever makes the right move".
//
// WHY IT IS CHECKED AGAINST THE TREE. The verb list comes from cobra via RoleVerbs, so a verb
// added tomorrow fails HERE on the day it is added rather than on the day someone notices no
// probe ever exercised it. A gate holding its own copy of the surface measures the copy — which is
// the defect that let `motion` sit outside the choice report's denominator for a day, and the one
// that let a hand-written role list report "4 of 14" when the true figure was 4 of 18.
//
// WHAT IT DOES NOT CLAIM. Coverage of the SURFACE is not coverage of the JUDGEMENT. A board that
// demands `close` proves a situation exists where closing is right; whether a seat closes for good
// reasons is a question only a human reading the run can answer, and the report this package
// produces is addressed to that human rather than to CI.
func TestEveryVerbHasABoardThatDemandsIt(t *testing.T) {
	demanded := map[string]map[string]bool{} // role -> verb -> demanded
	for name, b := range Boards() {
		role := roleOfSeat(b.Seat)
		if role == "" {
			t.Errorf("board %q names seat %q, whose role is not one of lens/merge/blue/bench", name, b.Seat)
			continue
		}
		if demanded[role] == nil {
			demanded[role] = map[string]bool{}
		}
		for _, e := range b.Expect {
			if roleOfSeat(e.Seat) != role {
				t.Errorf("board %q is a %s sitting but expects %q of %s — a board demands verbs of the seat it is FOR, or the coverage it claims is for a seat that was never dispatched", name, role, e.Verb, e.Seat)
				continue
			}
			demanded[role][e.Verb] = true
		}
	}

	for _, role := range Roles {
		var missing []string
		for _, verb := range surface().Verbs(role) {
			if demanded[role][verb] || AlwaysTaken[verb] != "" || NoSituation[role+" "+verb] != "" {
				continue
			}
			missing = append(missing, verb)
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			t.Errorf("%s has %d verb(s) no board makes the right move: %s\n\n"+
				"Each is a capability the tool offers and no scenario ever asks for — so nothing here would notice if a seat NEVER used it, which is the exact blind spot this package exists to close. Build a board where it is the accounted answer, or add it to AlwaysTaken WITH its reason.",
				role, len(missing), strings.Join(missing, ", "))
		}
	}
}

// AND NO BOARD DEMANDS A VERB THE SEAT DOES NOT HAVE.
//
// The inverse, and it is not symmetry for its own sake: an expectation naming a verb outside the
// seat's role can never be met, so it would report as an eternal UNMET and train its reader to
// discount the whole section.
func TestNoBoardDemandsAVerbTheSeatCannotReach(t *testing.T) {
	for name, b := range Boards() {
		role := roleOfSeat(b.Seat)
		have := map[string]bool{}
		for _, v := range surface().Verbs(role) {
			have[v] = true
		}
		for _, e := range b.Expect {
			if !have[e.Verb] {
				t.Errorf("board %q expects %s to use %q, which the %s role does not offer — an expectation that cannot be met reports UNMET forever and teaches its reader to skim",
					name, e.Seat, e.Verb, role)
			}
		}
	}
}

// EVERY BOARD IS A SITUATION, NOT A CHECKLIST.
//
// The failure this guards is the one that makes surface coverage worthless: satisfying the gate
// above by bolting every remaining verb onto one board. A seat dropped into a situation that wants
// nine unrelated acts is not choosing — it is working through a list — and a probe against it
// measures compliance rather than judgement.
func TestBoardsAreCoherentSittings(t *testing.T) {
	for name, b := range Boards() {
		if strings.TrimSpace(b.Report) == "" {
			t.Errorf("board %q has no report under audit — a seat with nothing to read cannot act on anything", name)
		}
		if len(b.Expect) == 0 {
			t.Errorf("board %q demands nothing", name)
		}
		if len(b.Expect) > 10 {
			t.Errorf("board %q demands %d verbs. That is a checklist, not a sitting: a seat facing it is working through a list and the probe measures compliance rather than judgement. Split it.", name, len(b.Expect))
		}
		for _, e := range b.Expect {
			if strings.TrimSpace(e.Because) == "" {
				t.Errorf("board %q expects %q with no argument for why — and `Because` is what a constitution would have to SAY for a seat to see it, which makes it the actionable half of every miss", name, e.Verb)
			}
		}
		for _, g := range b.Gaps {
			if g.Baits == "" || g.Why == "" {
				t.Errorf("board %q gap %q does not say what it baits or why — then it is scenery, and scenery is what makes a board look thorough", name, g.Key)
			}
			// A computation-kind gap that baits anything but `prove` contradicts the write path:
			// such a check closes ONLY on a proof, so no other verb can discharge it.
			if g.CheckKind == "computation" && g.Baits != "prove" {
				t.Errorf("board %q gap %q is check-kind computation but baits %q — a computation check closes only on a proof", name, g.Key, g.Baits)
			}
		}
	}
}

// AND THE EXEMPTIONS ARE REAL VERBS.
//
// An AlwaysTaken entry for a verb that no longer exists is a hole in the coverage claim that reads
// exactly like coverage — the same stale-exemption shape the enum gates already police.
func TestAlwaysTakenNamesRealVerbs(t *testing.T) {
	live := map[string]bool{}
	pair := map[string]bool{}
	for _, role := range Roles {
		for _, v := range surface().Verbs(role) {
			live[v] = true
			pair[role+" "+v] = true
		}
	}
	for key, why := range NoSituation {
		if !pair[key] {
			t.Errorf("NoSituation records %q, which is not a verb that role can reach — either the verb moved or the role did, and an entry for an unreachable pair is a coverage claim about nothing", key)
		}
		if strings.TrimSpace(why) == "" {
			t.Errorf("NoSituation records %q with no argument for why no sitting reaches it", key)
		}
	}
	for verb, why := range AlwaysTaken {
		if !live[verb] {
			t.Errorf("AlwaysTaken excuses %q, which is not a verb any role offers — remove it", verb)
		}
		if strings.TrimSpace(why) == "" {
			t.Errorf("AlwaysTaken excuses %q with no reason, which is an allowlist entry wearing a map's clothes", verb)
		}
	}
}

func roleOfSeat(seatID string) string {
	for _, role := range Roles {
		if seatRolePrefix(role, seatID) {
			return role
		}
	}
	return ""
}

// seatRolePrefix mirrors the record's seat-role binding closely enough for a test fixture: the
// authoritative check is record.CheckSeatRole, used by Read.
func seatRolePrefix(role, seatID string) bool {
	switch role {
	case "lens":
		return strings.HasPrefix(seatID, "red-lens")
	case "merge":
		return strings.HasPrefix(seatID, "red-merge")
	case "blue":
		return strings.HasPrefix(seatID, "blue")
	case "bench":
		return strings.HasPrefix(seatID, "judge")
	}
	return false
}

// surface is the tool's real verb set, read from the cobra tree.
//
// A TEST-ONLY import of the cli package, which is what keeps the dependency one-way: this package
// must not import cli (cli's own tests import THIS one to build the boards), and a test file may
// because it is compiled into the test binary rather than the package.
func surface() Surface { return NewSurface(cli.CommandPaths()) }
