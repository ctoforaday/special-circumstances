package record

import (
	"regexp"
	"strings"
	"testing"
)

// altSep is the separator BETWEEN alternative invocations. A bare "|" cannot be it: the same
// character separates enum values inside a single invocation (`--as PASS|FAIL`,
// `granted|denied`, `--severity|--likelihood`), and splitting on it reported "--likelihood" as a
// command a seat was told to type.
var altSep = regexp.MustCompile(`\s\|\s`)

// A DUTY'S `how` IS A COMMAND THE SEAT WILL TYPE, so it has to be typeable.
//
// Every one of them was roleless. Measured by running them verbatim against a real board:
// `friction --reason "x"` exited 2 (`friction` at the root is the OPERATOR's read, and it
// rejected --reason at parse time), and `revision --reason "y"` exited 2 ("revision" is not a
// top-level command). The duty channel is the one surface that tells a seat what it owes AND how
// to discharge it, and it was handing back commands that fail.
//
// The root teaches where a verb lives, so a seat recovers — at the cost of a turn, on the surface
// whose whole job is to spend none. Under a protocol where the constitution names no commands,
// this string is the seat's only instruction.

// firstTokens returns the leading word of each invocation in a `how`, split on the separators the
// duty strings use to offer alternatives.
func firstTokens(how string) []string {
	var out []string
	for _, part := range altSep.Split(how, -1) {
		for _, frag := range strings.Split(part, "   then   ") {
			frag = strings.TrimSpace(frag)
			// Prose alternatives are not invocations: "if the demand is wrong, say so in …" is
			// an instruction to argue, deliberately offered beside the verb so that disagreeing
			// does not look like non-compliance.
			if frag == "" || strings.HasPrefix(frag, "if ") || strings.HasPrefix(frag, "(") {
				continue
			}
			frag = strings.TrimPrefix(frag, "(")
			if f := strings.Fields(frag); len(f) > 0 {
				out = append(out, f[0])
			}
		}
	}
	return out
}

// checkAddresses asserts every invocation in every duty is one the named role can actually type.
func checkAddresses(t *testing.T, role string, duties []Duty) {
	t.Helper()
	for _, d := range duties {
		for _, tok := range firstTokens(d.How) {
			switch tok {
			case role:
				// `<role> <verb> …` — the form the CLI accepts.
			case "motion":
				// The motion group sits at the ROOT, filed by any seat and ruled by one, so a
				// role in front of it would be this same defect pointing the other way.
			default:
				t.Errorf("role %s is told to run %q, which begins with %q — a seat typing this reaches the root, not its verb.\n  duty: %s\n  how:  %s",
					role, d.How, tok, d.What, d.How)
			}
		}
	}
}

func TestEveryOutstandingDutyIsTypeable(t *testing.T) {
	// Boards shaped to fire each role's duties at once. The friction duty fires for every role
	// on any board, because no seat here has closed the channel.
	for _, tc := range []struct {
		role, seat string
		board      *Board
	}{
		{"blue", "blue-respond-r1", &Board{Gaps: map[string]*Gap{}}},
		{"merge", "red-merge-r1", &Board{
			GapOrder: []string{"R1-1"},
			Gaps:     map[string]*Gap{"R1-1": {ID: "R1-1", Open: true}},
			Events: []Event{
				{SeatID: "blue-respond-r1", Type: "motion", Payload: NewPayload().Set("motion_id", "M1").Set("subject", "grade")},
			},
		}},
		{"bench", "judge-r1", &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "blue-respond-r1", Type: "motion", Payload: NewPayload().Set("motion_id", "P1").Set("subject", "petition")},
		}}},
		{"lens", "red-lens-r1-L1", &Board{Gaps: map[string]*Gap{}}},
	} {
		t.Run(tc.role, func(t *testing.T) {
			s := SittingOf(tc.board, tc.role, tc.seat)
			if len(s.Outstanding) == 0 {
				t.Fatalf("no outstanding duties for %s — the fixture fires nothing, so this asserts nothing", tc.role)
			}
			checkAddresses(t, tc.role, s.Outstanding)
		})
	}
}

func TestEveryAffordanceIsTypeable(t *testing.T) {
	t.Setenv(DutyArmEnv, string(DutyAvailable))
	for _, tc := range []struct {
		role, seat string
		board      *Board
	}{
		{"blue", "blue-respond-r1", &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "blue-respond-r1", Type: "blue_edit", Payload: NewPayload().Set("answers", "R1-2")},
		}}},
		{"merge", "red-merge-r1", &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "red-merge-r1", Type: "motion-rule", Payload: NewPayload().
				Set("subject", "grade").Set("as", "accepted").Set("gap_id", "R1-1")},
		}}},
	} {
		t.Run(tc.role, func(t *testing.T) {
			got := AvailableOf(tc.board, tc.role, tc.seat)
			if len(got) == 0 {
				t.Fatalf("no affordances for %s — the fixture fires nothing, so this asserts nothing", tc.role)
			}
			checkAddresses(t, tc.role, got)
		})
	}
}

// THE GUARD ON THE GUARD. If firstTokens stopped finding invocations it would return nothing and
// every assertion above would pass over an empty set — the clean-board reading of a broken
// instrument, which is the defect this whole surface keeps producing.
func TestFirstTokensFindsTheInvocations(t *testing.T) {
	how := `blue friction --reason "x"  |  blue friction --none --reason "y"`
	got := firstTokens(how)
	if len(got) != 2 {
		t.Fatalf("found %d invocation(s) in a two-invocation how string: %v", len(got), got)
	}
	mixed := `merge show motions   then   motion grade rule --id M1 --as accepted --reason "..."`
	if got := firstTokens(mixed); len(got) != 2 || got[0] != "merge" || got[1] != "motion" {
		t.Errorf("a `then` chain resolved to %v, want [merge motion]", got)
	}
}
