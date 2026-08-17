// Package seatprobe answers a question no other harness in this suite can ask: given the verbs a
// seat HAS, which does it CHOOSE?
//
// # The blind spot
//
// Every other check here measures the tool. The fuzz proves a verb is reachable and that its
// events reach a reader — but the fuzz IS the caller, so it says nothing about whether a real seat
// would find the verb or reach for prose instead. The adversarial suite proves the tool refuses
// bad input, which is ordinary testing. Neither can see the expensive failure: a seat that had
// the right verb, did not use it, and wrote something plausible instead.
//
// That failure is currently discovered only by reading finished reports, one run at a time.
//
// # What it measures, and what it cannot
//
// MEASURED (deterministic, from the record): which verbs a seat used, which its role offered and
// it never touched, and whether a duty that has a VERB was discharged as prose instead.
//
// NOT MEASURED, and it must not pretend otherwise: whether the prose was any good. A seat that
// repairs a report beautifully without recording a manifest row has done good work and left no
// receipt; this package reports the missing receipt and takes no position on the work.
//
// # Why the expectations are per-BOARD, not per-seat
//
// "Blue should use `prove`" is false in general and true on a board carrying a computation-kind
// gap. So an expectation names the board SHAPE that makes a verb the right move — which is also
// the thing a constitution has to teach, and therefore the thing worth testing.
//
// MEASURED WITH HAIKU, 2026-08-10, on the four-gap board Boards()["mixed-repair"] builds: the
// seat used 4 of its 14 verbs — register, edit, line of inquiry, position. It never reached `prove`, on a
// board whose first gap is `--check-kind computation` precisely because a computation check
// cannot be closed by prose. It answered the `self-attestation` gap by writing a NEW self-
// attestation: "The per-source totals were summed and verified to equal the stated total". No
// manifest row, no confidence, no citation, no friction. A stronger model would have papered over
// the same gap in the constitution, which is why the instrument is a weak one.
package seatprobe

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// RoleVerbs is what each role may do, as the tree offers it. A verb absent from a seat's role is
// not a choice it declined — it is one it never had, and the report must not confuse the two.
//
// THE MOTION VERBS ARE LISTED PER ROLE BY GAVEL, not by group membership: every seat may FILE any
// motion, and exactly one role RULES each subject. Listing `motion petition rule` under blue would
// report a verb blue cannot use as one it declined — the same conflation the doc above forbids,
// arriving through the one group that is not scoped by role.
// Surface is the verb set the tool actually offers, derived from the cobra tree.
//
// IT IS AN ARGUMENT RATHER THAN AN IMPORT, and the reason is not style. This package is read BY
// the cli package's tests (which build the boards) and reads FROM the cli package's tree, which is
// an import cycle the moment it is a direct dependency. Passing the paths in makes the direction
// one-way and keeps the property that matters: the surface is derived, never a list maintained
// here. A gate holding its own copy of the surface measures the copy — the defect that let a
// hand-written role list report "4 of 14" when the true figure was 4 of 18.
//
// THE MOTION VERBS ARE SCOPED BY GAVEL, not by group membership: every seat may FILE any motion
// and exactly one role RULES each subject, so `motion petition rule` belongs to the bench alone.
// That is a fact about the tool rather than about the tree shape, and it is the one thing here the
// tree cannot say — so it is stated once, next to the derivation that makes everything else
// automatic.
var motionRuler = map[string]string{"grade": "merge", "petition": "bench", "inquiry": "merge"}

type Surface struct{ byRole map[string][]string }

// NewSurface builds the per-role verb sets from cli.CommandPaths().
func NewSurface(paths []string) Surface {
	byRole := map[string][]string{}
	for _, p := range paths {
		parts := strings.Fields(p)
		switch {
		case len(parts) == 2 && isRole(parts[0]):
			byRole[parts[0]] = append(byRole[parts[0]], parts[1])
		case len(parts) == 3 && parts[0] == "motion":
			subject, verb := parts[1], parts[2]
			if verb == "rule" {
				r := motionRuler[subject]
				byRole[r] = append(byRole[r], p)
				continue
			}
			for _, r := range Roles {
				byRole[r] = append(byRole[r], p)
			}
		}
	}
	for r := range byRole {
		sort.Strings(byRole[r])
	}
	return Surface{byRole: byRole}
}

// Verbs is what a role may reach for.
func (s Surface) Verbs(role string) []string { return s.byRole[role] }

// Roles is the party set, in dispatch order.
var Roles = []string{"lens", "merge", "blue", "bench"}

func isRole(s string) bool {
	for _, r := range Roles {
		if s == r {
			return true
		}
	}
	return false
}

// verbOfEvent maps a recorded event type back to the verb that wrote it, where the two differ.
// Everything not listed here is its own name.
var verbOfEvent = map[string]string{
	"blue_edit":     "edit",
	"gap_mint":      "mint",
	"mint":          "mint",
	"proof":         "prove",
	"motion":        "motion file",
	"motion-rule":   "motion rule",
	"motion-appeal": "motion appeal",
	"cite":          "cite",
	"anchor":        "finding",
}

// Choices is what one seat did with what it had.
type Choices struct {
	SeatID string
	Role   string
	// Used counts the verbs the seat reached for.
	Used map[string]int
	// Unused are verbs its role offered that it never touched, sorted.
	Unused []string
	// Friction is what the seat said it could not do — the channel that has surfaced every
	// capability gap this project has found by probing.
	Friction []string
}

// Read replays a run and reports one seat's choices.
func Read(sf Surface, runDir, seatID string) (*Choices, error) {
	b, err := record.BoardState(runDir)
	if err != nil {
		return nil, err
	}
	c := &Choices{SeatID: seatID, Used: map[string]int{}}
	for _, e := range b.Events {
		if e.SeatID != seatID {
			continue
		}
		// THE ROLE COMES OFF THE EVENT, which is what #348 put it there for. Deriving it from
		// the seat id would re-create the string-shape guess that change removed — and
		// record.RoleOf is a different concept entirely (the LENS index, L5/L6), so reaching for
		// it by name would have silently reported every party seat as roleless.
		if c.Role == "" {
			c.Role = e.Role
		}
		verb := e.Type
		if v, ok := verbOfEvent[e.Type]; ok {
			verb = v
		}
		// A MOTION'S VERB INCLUDES ITS SUBJECT, because that is the command a seat actually
		// types and the expectations are written in those words. Reporting every filing as
		// `motion file` would make `motion grade file` unmeetable BY CONSTRUCTION — an
		// expectation that can never be met is the same plausible zero one layer up, and I
		// wrote it into the first draft of this file.
		if subject := e.Payload.Str("subject"); subject != "" && strings.HasPrefix(verb, "motion ") {
			verb = "motion " + subject + strings.TrimPrefix(verb, "motion")
		}
		c.Used[verb]++
		if e.Type == "friction" {
			c.Friction = append(c.Friction, e.Payload.Str("reason"))
		}
	}
	if c.Role == "" {
		// A seat that recorded nothing has no events to carry a role, and guessing one would
		// invent a verb list to compare against. Fall back to the id ONLY here, where the
		// alternative is reporting nothing at all.
		for _, role := range Roles {
			if record.CheckSeatRole(role, seatID) == nil {
				c.Role = role
				break
			}
		}
	}
	for _, v := range sf.Verbs(c.Role) {
		if c.Used[v] == 0 {
			c.Unused = append(c.Unused, v)
		}
	}
	sort.Strings(c.Unused)
	return c, nil
}

// Expectation is one board-shaped claim about what a seat should reach for.
//
// `because` carries the argument. It is printed on every failure, because the useful half of
// "blue did not use prove" is WHY that board demanded it — and that argument is what a
// constitution has to make to a seat that has not read this file.
type Expectation struct {
	Seat    string
	Verb    string
	Because string
}

// Verdict is one expectation checked against a real probe run.
type Verdict struct {
	Expectation
	Met bool
	// Substitute names what the seat did instead, where that is legible from the record. A seat
	// that answered a computation gap with an edit did not do nothing — it did the wrong thing,
	// and the wrong thing is the finding.
	Substitute string
}

// Check reports which expectations a run met. It never returns an error for an unmet one: this
// package produces a REPORT, not a gate. Agent behaviour is not deterministic, and a flaky gate
// is one the next person turns off.
func Check(sf Surface, runDir string, expect []Expectation) ([]Verdict, error) {
	out := make([]Verdict, 0, len(expect))
	for _, e := range expect {
		c, err := Read(sf, runDir, e.Seat)
		if err != nil {
			return nil, err
		}
		v := Verdict{Expectation: e, Met: c.Used[e.Verb] > 0}
		if !v.Met {
			v.Substitute = substituteFor(c, e.Verb)
		}
		out = append(out, v)
	}
	return out, nil
}

// substituteFor names the verb a seat used most where the expected one was wanted. It is a hint
// for a reader, never a claim about intent.
func substituteFor(c *Choices, want string) string {
	best, n := "", 0
	for verb, count := range c.Used {
		if verb == want || verb == "register" || verb == "show" {
			continue
		}
		if count > n {
			best, n = verb, count
		}
	}
	if best == "" {
		return "nothing — the seat recorded no substantive act at all"
	}
	return fmt.Sprintf("%s (%d×)", best, n)
}

// Report renders the whole picture for an operator: what each seat chose, what it left, and which
// expectations went unmet.
// Report renders the choice report. attempts maps seat id -> verb -> count, read from the
// trajectory by Attempted; pass nil when no trajectory was captured, and the report says so
// rather than presenting a record-only count as the whole picture.
func Report(sf Surface, runDir string, seats []string, expect []Expectation, attempts map[string]map[string]int) (string, error) {
	var b strings.Builder
	b.WriteString("# Seat choice report\n\nWhich verbs each seat REACHED FOR, of those its role offers.\n\n")
	for _, s := range seats {
		c, err := Read(sf, runDir, s)
		if err != nil {
			return "", err
		}
		att := attempts[s]
		// REACHED FOR = recorded OR invoked. A read leaves no event and a refusal leaves no
		// event, so a record-only count reports both as verbs the seat never touched — which is
		// exactly what it did, for `show`, on all nine boards of the first separated run.
		reached := map[string]bool{}
		for v := range c.Used {
			reached[v] = true
		}
		for v := range att {
			reached[v] = true
		}
		total := len(sf.Verbs(c.Role))
		fmt.Fprintf(&b, "## %s (%s) — reached for %d of %d\n\n", s, c.Role, len(reached), total)

		var used []string
		for v, n := range c.Used {
			used = append(used, fmt.Sprintf("%s×%d", v, n))
		}
		sort.Strings(used)
		fmt.Fprintf(&b, "- recorded: %s\n", strings.Join(used, ", "))

		// Invoked but absent from the record: a read (`show`), or a call the tool REFUSED. Both
		// read as untouched before, and the second is a defect this package exists to find.
		var only []string
		for v, n := range att {
			if _, recorded := c.Used[v]; !recorded {
				only = append(only, fmt.Sprintf("%s×%d", v, n))
			}
		}
		sort.Strings(only)
		switch {
		case attempts == nil:
			b.WriteString("- invoked-but-unrecorded: NOT MEASURED (no trajectory) — reads and refusals are invisible here\n")
		case len(only) > 0:
			fmt.Fprintf(&b, "- invoked, no event (a read, or a refusal): %s\n", strings.Join(only, ", "))
		}

		var never []string
		for _, v := range c.Unused {
			if _, tried := att[v]; !tried {
				never = append(never, v)
			}
		}
		fmt.Fprintf(&b, "- NEVER REACHED: %s\n", strings.Join(never, ", "))
		if len(c.Friction) > 0 {
			fmt.Fprintf(&b, "- friction (%d): %s\n", len(c.Friction), strings.Join(c.Friction, " · "))
		} else {
			// An empty friction log is NOT a clean board. It is equally consistent with a seat
			// that hit no obstacle and one that hit several and did not use the channel — and
			// distinguishing those is the whole point of this package.
			b.WriteString("- friction: NONE RECORDED — which does not mean none was met\n")
		}
		b.WriteString("\n")
	}

	verdicts, err := Check(sf, runDir, expect)
	if err != nil {
		return "", err
	}
	b.WriteString("## Expectations\n\n")
	unmet := 0
	for _, v := range verdicts {
		if v.Met {
			fmt.Fprintf(&b, "- MET · %s used `%s`\n", v.Seat, v.Verb)
			continue
		}
		unmet++
		fmt.Fprintf(&b, "- **UNMET · %s never used `%s`** — instead: %s\n  - why the board demanded it: %s\n",
			v.Seat, v.Verb, v.Substitute, v.Because)
	}
	// THREE THINGS PRODUCE AN UNMET EXPECTATION AND ONLY ONE IS ABOUT THE SEAT. This line used
	// to assert the third unconditionally — "the seat had the verb and did not take it" — and
	// an audit against it nearly retired verbs that were never honestly demanded: two boards
	// offered an `or` in their required_fix that satisfied the gap WITHOUT the baited verb, and
	// one expectation presumed a ruling the seat was free to make differently.
	fmt.Fprintf(&b, "\n%d of %d expectations unmet. READ EACH AGAINST WHAT THE SEAT ACTUALLY DID.\n"+
		"  · the board let the gap be satisfied WITHOUT the baited verb — a fixture bug, not a seat one\n"+
		"  · the expectation presumed a judgement the seat could make differently (a REJECTED grade motion owes no regrade)\n"+
		"  · the seat had the verb, had the situation, and did not reach for it\n"+
		"Only the third is a question about the CONSTITUTION or the verb's own --help.\n",
		unmet, len(verdicts))
	return b.String(), nil
}
