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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
//
// THE GAVEL IS READ FROM THE SCHEMA, NOT COPIED HERE. This was a hand-written
// map[string]string — the fourth copy of a fact the MotionSubject enum already carries as a
// `ruled_by` annotation, which cli/motion and record/refs.go both resolve through
// recordpb.SubjectRuler. A gate holding its own copy of the surface measures the copy, which is
// the sentence one paragraph up, and this file was doing it. A subject added to the schema with
// its gavel declared would have been filed under byRole[""] here — offered to no role, and read
// by the coverage gate as a verb nobody needs a board for.

type Surface struct{ byRole map[string][]string }

// NewSurface builds the per-role verb sets from cli.CommandPaths().
func NewSurface(paths []string) Surface {
	byRole := map[string][]string{}
	for _, p := range paths {
		parts := strings.Fields(p)
		switch {
		case len(parts) >= 2 && isRole(parts[0]):
			// A ROLE VERB IS NOT ALWAYS ONE WORD. The verbs that carried two contracts are
			// subgroups now — `blue line-of-inquiry propose`, `merge class new` — and a case
			// matching exactly two fields dropped them from the surface entirely. The coverage
			// gate then reported full coverage of a surface it could not see, and its inverse
			// called a real verb one the role does not offer.
			//
			// `show` stays collapsed to the group: its projections are the READ path, covered by
			// AlwaysTaken as one act, and expanding them would demand a board per projection.
			verb := strings.Join(parts[1:], " ")
			if parts[1] == "show" {
				verb = "show"
			}
			byRole[parts[0]] = append(byRole[parts[0]], verb)
		case len(parts) == 3 && parts[0] == "motion":
			subject, verb := parts[1], parts[2]
			if verb == "rule" {
				byRole[rulerOf(subject)] = append(byRole[rulerOf(subject)], p)
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
// Everything not listed here is its own name — `recordpb.Word`, which is the enum's own spelling.
//
// THE KEY IS THE ENUM, NOT ITS WORD, and that is the half the migration bought. A string key
// could name a type the schema does not have and nothing would refuse it: `gap_mint` sat in this
// table until the conversion, mapping an event type that appears NOWHERE else in the tree —
// neither written by any verb nor declared by the schema — to `mint`. A dead key in a lookup
// table returns nothing, which reads exactly like a type that simply never occurred.
//
// IT NOW CARRIES THE HYPHENS TOO, and that is not decoration. The values here are CLI VERB PATHS
// — they are compared against `Surface.Verbs`, which comes from the cobra tree — while the enum
// spells its words with underscores (`spot_check`). Before the schema landed, the event type was
// a string spelled the CLI's way, so the identity default happened to match; it no longer does,
// and every multiword verb would otherwise be counted under a name no expectation is written
// against while the real verb reported as NEVER REACHED. This table is where the two
// vocabularies are joined, and it is the only place they are.
//
// TWO ENTRIES ARE PRESERVED DEFECTS, carried forward unchanged rather than quietly repaired
// while converting: `class-new` and `line-of-inquiry` match no path the tree offers (`class new`
// takes a space, and an avenue's verb is `line-of-inquiry propose` or `line-of-inquiry move`,
// which the event type alone does not distinguish). They named nothing before this change and
// name nothing after it. Reported, not fixed here.
var verbOfEvent = map[recordpb.EventType]string{
	recordpb.EventType_EVENT_TYPE_ANCHOR:         "finding",
	recordpb.EventType_EVENT_TYPE_AVENUE:         "line-of-inquiry",
	recordpb.EventType_EVENT_TYPE_BLUE_EDIT:      "edit",
	recordpb.EventType_EVENT_TYPE_CLASS_NEW:      "class-new",
	recordpb.EventType_EVENT_TYPE_FRICTION_NONE:  "friction-none",
	recordpb.EventType_EVENT_TYPE_INQUIRY_REVIEW: "inquiry-support",
	recordpb.EventType_EVENT_TYPE_MANIFEST_ROW:   "manifest-row",
	recordpb.EventType_EVENT_TYPE_MOTION:         "motion file",
	recordpb.EventType_EVENT_TYPE_MOTION_APPEAL:  "motion appeal",
	recordpb.EventType_EVENT_TYPE_MOTION_RULE:    "motion rule",
	recordpb.EventType_EVENT_TYPE_PROOF:          "prove",
	recordpb.EventType_EVENT_TYPE_SPOT_CHECK:     "spot-check",
}

// motionSubject is the SUBGROUP a seat typed, read off whichever motion body the event carries.
//
// A type switch rather than a switch on `e.GetType()`: the three bodies are the only messages
// that have a subject at all, so asking the body cannot go stale against the enum.
func motionSubject(e *record.Event) string {
	if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
		return subjectVerb(m.GetSubject())
	}
	if m, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok {
		return subjectVerb(m.GetSubject())
	}
	if m, ok := recordpb.BodyAs[*recordpb.MotionAppeal](e); ok {
		return subjectVerb(m.GetSubject())
	}
	return ""
}

// subjectVerb spells a motion subject the way the COMMAND TREE spells it, which for one of the
// three is not the enum's word.
//
// `MOTION_SUBJECT_DIRECTION` is typed `motion inquiry rule` / `motion inquiry appeal` — the
// schema re-conceived the subject as a DIRECTION (it rules on a line of inquiry blue proposed,
// and `DirectionMotion` carries the avenue's own id) while the CLI subgroup, its help text and
// every board expectation still say `inquiry`. Rendering the enum's word here would report a
// verb no role offers, making both `motion inquiry` expectations unmeetable BY CONSTRUCTION —
// the failure mode this file's own comment warns about, arriving through the rename.
func subjectVerb(s recordpb.MotionSubject) string {
	if s == recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION {
		return "inquiry"
	}
	return recordpb.Word(s)
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
func Read(sf Surface, run record.Run, seatID string) (*Choices, error) {
	b, err := record.BoardState(run)
	if err != nil {
		return nil, err
	}
	c := &Choices{SeatID: seatID, Used: map[string]int{}}
	for _, e := range b.Events {
		if e.GetSeatId() != seatID {
			continue
		}
		// THE ROLE COMES OFF THE EVENT, which is what #348 put it there for. Deriving it from
		// the seat id would re-create the string-shape guess that change removed — and
		// record.RoleOf is a different concept entirely (the LENS index, L5/L6), so reaching for
		// it by name would have silently reported every party seat as roleless.
		if c.Role == "" {
			c.Role = e.GetRole()
		}
		verb, ok := verbOfEvent[e.GetType()]
		if !ok {
			verb = recordpb.Word(e.GetType())
		}
		// A MOTION'S VERB INCLUDES ITS SUBJECT, because that is the command a seat actually
		// types and the expectations are written in those words. Reporting every filing as
		// `motion file` would make `motion grade file` unmeetable BY CONSTRUCTION — an
		// expectation that can never be met is the same plausible zero one layer up, and I
		// wrote it into the first draft of this file.
		if subject := motionSubject(e); subject != "" && strings.HasPrefix(verb, "motion ") {
			verb = "motion " + subject + strings.TrimPrefix(verb, "motion")
		}
		c.Used[verb]++
		// THE BODY IS THE TYPE, so asking for a Friction body asks the same question
		// `e.Type == "friction"` did — and a FrictionNone event, which is a different fact, is
		// refused by the assertion rather than by a second string compare that could drift from
		// it. `text` is the field the schema gives the seat's own sentence; `reason` was the
		// payload key that carried it.
		if f, ok := recordpb.BodyAs[*recordpb.Friction](e); ok {
			c.Friction = append(c.Friction, f.GetText())
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
func Check(sf Surface, run record.Run, expect []Expectation) ([]Verdict, error) {
	out := make([]Verdict, 0, len(expect))
	for _, e := range expect {
		c, err := Read(sf, run, e.Seat)
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
func Report(sf Surface, run record.Run, seats []string, expect []Expectation, attempts map[string]map[string]int) (string, error) {
	var b strings.Builder
	b.WriteString("# Seat choice report\n\nWhich verbs each seat REACHED FOR, of those its role offers.\n\n")
	for _, s := range seats {
		c, err := Read(sf, run, s)
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

	verdicts, err := Check(sf, run, expect)
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

// rulerOf resolves the role holding a motion subject's gavel, from the schema.
//
// IT PANICS, AND THAT IS THE POINT. NewSurface returns a Surface and no error, and the surface it
// builds is what the coverage gate measures itself against — so an unresolvable subject that
// returned "" would file the verb under byRole[""], where no role offers it and no board is
// demanded for it. The gate would then pass, having measured a surface with a verb missing from
// it. A panic at surface construction fails every probe run at once instead, which is the same
// trade cli/motion.rulerFor makes at command construction and for the same stated reason: the
// failure belongs at startup, not at the moment someone tries to use the verb that went missing.
//
// The un-annotated case cannot be reached from a known subject — record/gavel_test.go's
// TestEveryMotionSubjectNamesItsRuler refuses any MotionSubject without a ruler at the descriptor
// — so this arm is defense in depth against a schema change that lands without it.
func rulerOf(subject string) string {
	subj, known := record.MotionSubjectEnum(subject)
	if !known {
		panic("seatprobe: motion subject " + subject + " is not in the MotionSubject enum — the CLI tree and the schema have diverged, and the surface this builds is what the coverage gate measures")
	}
	r, err := recordpb.SubjectRuler(subj)
	if err != nil {
		panic("seatprobe: " + err.Error())
	}
	return r
}
