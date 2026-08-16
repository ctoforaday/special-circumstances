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

// A BOARD CARRIES THE STATE ITS EXPECTATIONS NEED.
//
// MEASURED, AND IT IS WHY THIS EXISTS. The first probe run reported nineteen unmet expectations,
// and THREE of them were unmeetable by construction: `adjudicate` expected `motion grade rule`
// with no motion on the board, `lens-audit` expected `reproduce` with no proof recorded, and
// `sitting` expected `motion petition rule` with no petition filed. The seats were blamed in the
// report for verbs they had no way to reach.
//
// The coverage gate above passed all three, because it checks an expectation's VERB against the
// role's verb list and never asks whether the board's STATE makes that verb reachable. That is the
// same shape as a fixture that declares a cited claim and silently builds without it — a situation
// asserted rather than created — and it is worse here, because the failure surfaces as a finding
// about the SEAT.
//
// The rule: a verb that answers something can only be demanded on a board that carries the thing.
func TestEveryExpectationIsReachableOnItsBoard(t *testing.T) {
	// needs maps a verb to what the board must already carry for it to be possible at all.
	needs := map[string]struct {
		what string
		has  func(Board) bool
	}{
		"motion grade rule":     {"a filed grade motion", func(b Board) bool { return hasMotion(b, "grade", false) }},
		"motion petition rule":  {"a filed petition motion", func(b Board) bool { return hasMotion(b, "petition", false) }},
		"motion direction rule": {"a proposed avenue", func(b Board) bool { return len(b.Avenues) > 0 }},
		"motion grade appeal":   {"a RULED grade motion", func(b Board) bool { return hasMotion(b, "grade", true) }},
		"motion direction appeal": {"a RULED avenue", func(b Board) bool {
			for _, a := range b.Avenues {
				if a.Ruled != "" {
					return true
				}
			}
			return false
		}},
		"reproduce":   {"a recorded proof", func(b Board) bool { return len(b.Proofs) > 0 }},
		"regrade":     {"a gap whose grade can move", func(b Board) bool { return len(b.Gaps) > 0 }},
		"close":       {"an open gap", func(b Board) bool { return len(b.Gaps) > 0 }},
		"closing":     {"a gap to argue about", func(b Board) bool { return len(b.Gaps) > 0 }},
		"spot-check":  {"a CLOSED gap in the archive", func(b Board) bool { return anyClosed(b) }},
		"claim-index": {"at least one cited claim", func(b Board) bool { return len(b.Claims) > 0 }},
		"verify":      {"at least one cited claim", func(b Board) bool { return len(b.Claims) > 0 }},
		"retire":      {"a claim in the report to remove", func(b Board) bool { return len(b.Claims) > 0 }},
		"avenue":      {"nothing — a seat may always propose a line", func(b Board) bool { return true }},
	}

	for name, b := range Boards() {
		for _, e := range b.Expect {
			n, tracked := needs[e.Verb]
			if !tracked {
				continue // verbs with no precondition: edit, cite, prove, friction, mint, position …
			}
			if !n.has(b) {
				t.Errorf("board %q expects %q and does not carry %s.\n\n"+
					"The expectation cannot be met, so the report will record it as a MISS BY THE SEAT — which is a finding about the fixture wearing a finding about the constitution. Build the state, or drop the expectation.",
					name, e.Verb, n.what)
			}
		}
	}
}

func anyClosed(b Board) bool {
	for _, g := range b.Gaps {
		if g.Closed {
			return true
		}
	}
	return false
}

func hasMotion(b Board, subject string, ruled bool) bool {
	for _, m := range b.Motions {
		if m.Subject == subject && (!ruled || m.Ruled != "") {
			return true
		}
	}
	return false
}

// A RULING A SEAT CANNOT READ IS NOT A RULING.
//
// MEASURED 2026-08-16 BY ASKING, which is the only way this was ever going to surface. A blue seat
// on `docket` called `show motions` twice and `show debate` once, then reported that the ruling
// field "just says 'endorsed' or 'too-thin' or 'rejected'" and that red's reasoning must be
// somewhere it could not see. It was right: Build stamped every ruling "ruled <verdict> on the line
// as it was proposed".
//
// Boards demand `motion grade appeal` and `motion direction appeal` — acts that are judgements
// about the ARGUMENT behind a refusal. Scoring whether a seat appealed a verdict it could not read
// is scoring a coin flip, and no acting probe could ever report it, because a content-free reason
// and a considered one produce identical events.
func TestEveryRuledFixtureCarriesTheArgumentForItsRuling(t *testing.T) {
	for name, b := range Boards() {
		for i, a := range b.Avenues {
			if a.Ruled == "" {
				continue
			}
			if strings.TrimSpace(a.RuledWhy) == "" {
				t.Errorf("board %q avenue %d is ruled %q with no RuledWhy — a seat asked whether to appeal it is guessing", name, i+1, a.Ruled)
				continue
			}
			assertNotRestatement(t, name, "avenue", a.Ruled, a.RuledWhy)
		}
		for i, m := range b.Motions {
			if m.Ruled == "" {
				continue
			}
			if strings.TrimSpace(m.RuledWhy) == "" {
				t.Errorf("board %q motion %d is ruled %q with no RuledWhy — a seat asked whether to press it is guessing", name, i+1, m.Ruled)
				continue
			}
			assertNotRestatement(t, name, "motion", m.Ruled, m.RuledWhy)
		}
	}
}

// assertNotRestatement refuses the shape the boilerplate had: a "reason" whose whole content is the
// verdict it is supposedly explaining. An empty field is caught by the caller; this catches the
// field that LOOKS filled, which is the version that survives review.
func assertNotRestatement(t *testing.T, board, kind, verdict, why string) {
	t.Helper()
	trimmed := strings.TrimSpace(why)
	if len(trimmed) < 60 {
		t.Errorf("board %q %s: the argument for ruling %q is %d characters — too short to be an argument, and a reason that only restates the verdict is what a seat reported as unreadable",
			board, kind, verdict, len(trimmed))
	}
	// "ruled too-thin on the line as it was proposed" — the verdict, wrapped in filler.
	if stripped := strings.TrimSpace(strings.ReplaceAll(strings.ToLower(trimmed), strings.ToLower(verdict), "")); len(stripped) < 40 {
		t.Errorf("board %q %s: remove the verdict %q from its own reason and almost nothing is left — that is a restatement, not an argument", board, kind, verdict)
	}
}
