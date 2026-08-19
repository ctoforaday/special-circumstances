package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"strings"
	"testing"
)

// THE DUTY NAMES A THING THE HELP MUST REACH.
//
// A duty says WHAT is owed and nothing about how — the help page is the only page that instructs.
// That holds only if every verb a duty points at is findable from the seat's own help. Motions
// were the exception: the group sits at the ROOT (any seat files, one role rules), so
// `merge --help` listed no motion verb and no pointer to one, while the work list says "motion M1
// was filed and never ruled — PASS is refused while it stands".

// EVERY SEAT CAN REACH THE MOTION GROUP, and now it is IN its tree rather than pointed at.
//
// The footer this test was written for said "MOTIONS ARE AT THE ROOT, not under your role" —
// which a bench seat read as an exemption rather than an address, and which stopped being true
// when the role level went away. The pointer is gone because the thing it pointed around is gone:
// `motion` is mounted in each seat's own surface, with `rule` only where the gavel is.
func TestEverySeatHasTheMotionGroupInItsOwnTree(t *testing.T) {
	for _, role := range []string{"blue", "lens", "merge", "bench"} {
		t.Run(role, func(t *testing.T) {
			root := NewRootFor(record.SampleSeatOf(role))
			c, _, err := root.Find([]string{"motion"})
			if err != nil || c.Name() != "motion" {
				t.Fatalf("the %s seat has no motion group in its own tree: %v", role, err)
			}
			// AND FILING IS EVERY SEAT'S. The asymmetry is which seat may RULE, never which may ask.
			g, _, err := root.Find([]string{"motion", "grade", "file"})
			if err != nil || g.Name() != "file" {
				t.Errorf("the %s seat cannot file a grade motion: %v", role, err)
			}
		})
	}
}

// THE SUBJECTS ARE REACHABLE FROM EVERY SEAT, and each carries at least one verb.
//
// This was "the pointer must point at something" — a footer naming a group that had been renamed
// or removed is worse than no footer. There is no pointer now; the group is in the tree. The claim
// that survives, and is the one that mattered, is that the group a seat lands in is not empty.
func TestEveryMotionSubjectCarriesAVerb(t *testing.T) {
	root := NewRootFor(record.SampleSeatOf("merge"))
	c, _, err := root.Find([]string{"motion"})
	if err != nil || c == nil || c.Name() != "motion" {
		t.Fatalf("the merge seat has no motion group: %v", err)
	}
	if !c.HasSubCommands() {
		t.Fatal("the motion group has no subjects — a seat sent here finds nothing to do")
	}
	for _, sub := range c.Commands() {
		if !sub.HasSubCommands() {
			t.Errorf("motion subject %q carries no verb at all", sub.Name())
		}
	}
}

// NO SEAT'S ROOT IS MISSING THE FRICTION FOOTER.
//
// This was about role GROUPS and the drift a hand-applied footer invites. The groups are gone and
// the drift risk moved with the footer: it now hangs on the seat's root, applied in one place in
// NewRootFor. A role added later that missed it would be a seat never told that a capability it
// cannot find is a finding about the tooling rather than something to work around — which is the
// regression this caught when the groups were deleted.
func TestNoSeatsRootIsMissingTheFrictionFooter(t *testing.T) {
	for _, role := range []string{"blue", "lens", "merge", "bench"} {
		root := NewRootFor(record.SampleSeatOf(role))
		if !strings.Contains(root.Long, "it does not exist for you") {
			t.Errorf("the %s seat's root help is missing the friction footer", role)
		}
		if !strings.Contains(root.Long, "'friction' verb") {
			t.Errorf("the %s seat's root help does not name the friction channel", role)
		}
	}
}
