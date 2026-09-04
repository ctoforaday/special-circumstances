package record

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// debateSource is the dispatch this package's roster claims to track.
const debateSource = "../../../skills/research-protocol/scripts/debate.js"

// hole stands for an interpolated segment on both sides of the comparison — `${round}` in a
// debate.js template, `\d+` in a roster pattern. Using one sentinel for both is what lets a
// template and a pattern be compared as literal text with holes, rather than by running one
// against the other and hoping the sample chosen was representative.
const hole = "\x00"

// interpolation matches a template literal's `${...}` segment.
var interpolation = regexp.MustCompile(`\$\{[^}]*\}`)

// recordClauseArg captures the seat id handed to recordClause at each dispatch.
//
// THIS IS THE ANCHOR, AND THE LABEL IS NOT. `recordClause` interpolates `SEAT_ID: ${seatId}` into
// the seat's prompt, and the seat copies that string into `register --seat-id` — so its argument
// is the exact id requireDispatchableSeat has to admit. A dispatch LABEL is a different string for
// a different reader: `red-lens-${role}-r${round} · ${slug}` is what a dashboard shows, and it
// does not match the seat id's own shape at all. Binding the roster to labels would compare it
// against something no seat ever types.
var recordClauseArg = regexp.MustCompile(`recordClause\(([^)]*)\)`)

// skeletonOfTemplate reduces a JS string or template literal to literal text with holes.
//
// The bool is not decoration: an argument this cannot reduce must reach the caller as UNKNOWN
// rather than as an empty skeleton, because an empty skeleton would match nothing, report one
// tidy failure, and read like an ordinary mismatch instead of like a test that no longer
// understands its own input.
func skeletonOfTemplate(arg string) (string, bool) {
	s := strings.TrimSpace(arg)
	if len(s) < 2 {
		return "", false
	}
	q := s[0]
	if (q != '\'' && q != '"' && q != '`') || s[len(s)-1] != q {
		return "", false
	}
	s = s[1 : len(s)-1]
	s = interpolation.ReplaceAllString(s, hole)
	// Nothing quote-like may survive: a nested literal means the argument is an expression this
	// reduction cannot read, and guessing at it is how the gate goes quiet.
	if strings.ContainsAny(s, "'\"`$") {
		return "", false
	}
	return s, true
}

// skeletonOfPattern reduces a roster pattern to the same shape.
//
// IT REFUSES WHAT IT DOES NOT UNDERSTAND. `\d+` is the only construct the roster uses for an
// interpolated segment; any other metacharacter appearing in a future pattern means this
// reduction is no longer reading the pattern it thinks it is. Returning false there — rather
// than a best-effort skeleton — is the difference between this gate failing loudly the day the
// roster grows a construct it cannot handle, and it silently passing every id forever after.
func skeletonOfPattern(p string) (string, bool) {
	s := strings.TrimPrefix(strings.TrimSuffix(p, "$"), "^")
	s = strings.ReplaceAll(s, `\d+`, hole)
	if strings.ContainsAny(s, `\[]()*+?{}|.`) {
		return "", false
	}
	return s, true
}

// THE ROSTER TRACKS THE ENGINE, AND UNTIL NOW NOTHING CHECKED THAT IT DID.
//
// roster.go has always claimed "every id below is verified against debate.js's own dispatch
// sites" and named this test. The test did not exist (#675). What existed was
// TestTheRosterAndTheRoleTableAgree, which reconciles seatShapes against roleSeats — two
// hand-written Go tables stating one vocabulary, neither of them the engine. A comment asserting
// a check, in the same paragraph that argues "a comment is not a check", is the exact shape this
// package exists to refuse.
//
// BOTH DIRECTIONS ARE FAILURES AND THEY ARE NOT SYMMETRIC:
//
//	under-admission  debate.js dispatches a seat no shape matches. requireDispatchableSeat
//	                 refuses it at `register`, which is the seat's FIRST act, so the sitting
//	                 cannot start at all — and the refusal names the seat, so it reads as the
//	                 seat's fault rather than as a roster that never learned about it.
//	over-admission   a shape outlives the dispatch that justified it and goes on admitting ids
//	                 no run can produce. Already paid for once in a sibling surface: `judge-r1`
//	                 sat on two seat-probe boards for the probe's whole life, valid by the
//	                 roster and never seated by the orchestrator
//	                 (seatprobe/fidelity_test.go).
func TestTheRosterMatchesWhatTheEngineActuallyDispatches(t *testing.T) {
	src, err := os.ReadFile(debateSource)
	if err != nil {
		t.Fatalf("cannot read debate.js — the dispatch this roster binds against: %v", err)
	}
	js := string(src)

	shapeSkeletons := map[string]seatShape{}
	for _, s := range seatShapes {
		sk, ok := skeletonOfPattern(s.re.String())
		if !ok {
			t.Fatalf("roster pattern %s uses a construct this bind does not understand; teach "+
				"skeletonOfPattern about it rather than leaving the shape unbound", s.re)
		}
		shapeSkeletons[sk] = s
	}

	// The one argument that is not a literal, exempted by NAME and then verified rather than
	// waved through. A petition sitting is named for the seat that petitioned, so its tail is
	// itself a seat id and it lives outside seatShapes on purpose (see petitionPrefix).
	const petitionVar = "seatID"

	dispatched := map[string]bool{}
	ms := recordClauseArg.FindAllStringSubmatch(js, -1)
	for _, m := range ms {
		arg := strings.TrimSpace(m[1])
		if arg == petitionVar {
			continue
		}
		sk, ok := skeletonOfTemplate(arg)
		if !ok {
			t.Errorf("recordClause(%s): this bind cannot reduce that argument to a seat id. If the "+
				"engine has a new way of composing one, teach this test — an unread argument is a "+
				"seat the roster is no longer checked against", arg)
			continue
		}
		if _, known := shapeSkeletons[sk]; !known {
			t.Errorf("debate.js dispatches seat id %q and no roster shape matches it — "+
				"requireDispatchableSeat will refuse that seat at `register`, its first act",
				strings.ReplaceAll(sk, hole, "<n>"))
			continue
		}
		dispatched[sk] = true
	}

	// A count floor, because every assertion above is inside a loop: a regex that stopped
	// matching would report a clean board.
	if len(ms) < 10 {
		t.Errorf("found %d recordClause dispatch sites in debate.js, expected at least 10 — "+
			"this bind is reading the wrong thing and passing on an empty set", len(ms))
	}

	for sk, s := range shapeSkeletons {
		if dispatched[sk] {
			continue
		}
		// The operator is the one shape with no dispatch, and that is what it IS: the identity a
		// human or a script acts under outside the debate. debate.js never seats it.
		if s.role == OperatorRole {
			continue
		}
		t.Errorf("roster shape %s is admitted at `register` and debate.js dispatches no seat that "+
			"matches it — either the dispatch was retired and the shape outlived it, or this bind "+
			"stopped seeing it", s.re)
	}
}

// THE PETITION PREFIX IS ONE FACT WITH TWO AUTHORS, so it is bound rather than trusted.
//
// Go composes nothing here — it CUTS the prefix and recurses on the tail — so the two sides can
// disagree silently: rename the sitting in debate.js and every petition seat is refused at
// register, with a message about an id the engine does not dispatch, for an id it just did.
func TestThePetitionPrefixMatchesTheOneDebateComposes(t *testing.T) {
	src, err := os.ReadFile(debateSource)
	if err != nil {
		t.Fatalf("cannot read debate.js: %v", err)
	}

	re := regexp.MustCompile("const petitionSeatID = \\([^)]*\\) => `([^`]*)`")
	m := re.FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("debate.js no longer defines petitionSeatID as a single template literal — the " +
			"petition sitting id is composed some other way now, and petitionPrefix is bound to nothing")
	}
	head := interpolation.Split(m[1], 2)[0]
	if head != petitionPrefix {
		t.Errorf("debate.js names petition sittings %q; petitionPrefix is %q, so every petition seat "+
			"is refused at register", m[1], petitionPrefix)
	}
}
