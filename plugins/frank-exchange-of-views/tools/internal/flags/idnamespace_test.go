package flags

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// CAN ONE --id CARRY EVERY ID? (gblock: "I'm ok for a single id to support all of the ids. That's
// good. We can prove no collision." — and "some IDs on one flag and others on another make no sense.")
//
// `--id` means a gap, a motion, an avenue and a proof digest depending on the verb, while `--anchor`
// and `--sha` carry more kinds under their own names (#1172). Consolidating them onto one flag that
// resolves by SHAPE is sound only if the shapes decide a value unambiguously — so this is the
// proof, and it is over the shapes the flag package declares rather than a second table written to
// match them.
//
// THE ANSWER IS "YES, BY THE LETTER". Every id is one shape, <LETTER>-<8 hex>, and the letter is the
// kind; a shape that takes several kinds (`anchor`) takes several LETTERS, and a shape that takes
// one (`citation-anchor`, `gap-id`) is a deliberate SUBSET of it, so a verb can refuse the wider
// kind. A resolver over these shapes therefore resolves MOST-SPECIFIC-FIRST; it cannot assume
// exactly one match.

// idShapes is every shaped id value the flag package offers, by the name its Type reports.
func idShapes() map[string]func() *ShapedValue {
	return map[string]func() *ShapedValue{
		"gap-id":          GapID,
		"motion-id":       MotionID,
		"avenue-id":       AvenueID,
		"anchor":          AnchorID,
		"citation-anchor": CitationAnchor,
		"sha256":          SHA,
	}
}

// minted is every kind of id the record mints, with the shapes that take it — ALL of them, so the
// matrix below is exact in both directions: a shape missing from a row must refuse the kind.
//
// A kind in two shapes is a SPECIALISATION, an allowance with a reason and never a tolerated
// overlap: a citation is an anchor and `lens verify` takes a source, so it must refuse a finding;
// a gap's anchor carries the gap's own id.
var minted = []struct {
	kind   string
	shapes []string
}{
	{"finding", []string{"anchor"}},
	{"citation", []string{"anchor", "citation-anchor"}},
	{"proof", []string{"anchor"}},
	{"gap", []string{"anchor", "gap-id"}},
	{"avenue", []string{"avenue-id"}},
	{"motion", []string{"motion-id"}},
}

// mint spells a fresh id of a kind exactly as the record's one minter does: the kind's row of the
// anchor table over four random bytes. `record.NewID` is that call and nothing else, and this
// package cannot import `record`, which imports it; internal/record's own matrix runs the minter.
func mint(t *testing.T, kind string) string {
	t.Helper()
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	return anchor.ID(kind, b)
}

func accepts(mk func() *ShapedValue, s string) bool { return mk().Set(s) == nil }

func TestOneIDFlagCanResolveEveryIDByShape(t *testing.T) {
	shapes := idShapes()
	used := map[string]bool{}
	for _, m := range minted {
		for _, name := range m.shapes {
			if shapes[name] == nil {
				t.Fatalf("%s names the shape %q and the flag package offers none", m.kind, name)
			}
			used[name] = true
		}
		// Several of each: the id is random, and one sample proves one sample.
		for i := 0; i < 32; i++ {
			id := mint(t, m.kind)
			for name, mk := range shapes {
				if got, want := accepts(mk, id), slices.Contains(m.shapes, name); got != want {
					t.Errorf("a %s id %q on a %s flag: accepted=%v, want %v — the letter is what tells the kinds apart, "+
						"and a shape that takes another kind's id joins it silently to the wrong entity", m.kind, id, name, got, want)
				}
			}
		}
	}
	// A digest is the one id that is not minted, and no id shape takes one.
	sum := sha256.Sum256([]byte("a stored proof run"))
	digest := hex.EncodeToString(sum[:])
	for name, mk := range shapes {
		if got, want := accepts(mk, digest), name == "sha256"; got != want {
			t.Errorf("a proof digest on a %s flag: accepted=%v, want %v", name, got, want)
		}
	}
	used["sha256"] = true
	// EVERY DECLARED SHAPE ACCEPTS SOMETHING THE TOOL MINTS. A shape that describes nothing cannot
	// collide with anything and would pass the matrix above entirely on refusals.
	for name := range shapes {
		if !used[name] {
			t.Errorf("no minted kind is accepted by the %s shape — the shape or the table is wrong", name)
		}
	}
}

// THE SHAPE IS THE WHOLE VALUE AND EXACTLY EIGHT HEX. An id one character short or long, with its
// letter in the other case, with upper-case hex, or bare of its letter is refused by the shape that
// takes the id itself.
func TestAShapeRefusesANearMissOfItsOwnID(t *testing.T) {
	shapes := idShapes()
	for _, m := range minted {
		id := mint(t, m.kind)
		near := []string{
			id[:len(id)-1], id + "0", "x" + id, id + " " + id,
			strings.ToLower(id[:1]) + id[1:], id[2:], id[:1] + id[2:], id[:2] + "ABCDEF01",
		}
		for _, name := range m.shapes {
			if !accepts(shapes[name], id) {
				t.Fatalf("%s refuses the %s id %q", name, m.kind, id)
			}
			for _, bad := range near {
				if accepts(shapes[name], bad) {
					t.Errorf("%s accepts %q, a near miss of the %s id %q", name, bad, m.kind, id)
				}
			}
		}
	}
	// The retired spellings are no id of any kind.
	for _, old := range []string{"G7", "M1", "Q1", "f-7daddb6a", "c-a1b2c3d4", "p-99887766", "adversary-F2", "L5-F1"} {
		for name, mk := range shapes {
			if accepts(mk, old) {
				t.Errorf("%s accepts %q", name, old)
			}
		}
	}
}

// A SPECIALISATION IS ONE-WAY, and this is what makes it a specialisation rather than a collision:
// every citation anchor is an anchor, and not every anchor is a citation anchor. If the wide kind
// started accepting only what the narrow one does they would be the same shape under two names, which
// is the defect #1172 is about.
func TestASpecialisationNarrowsAndIsNotMutual(t *testing.T) {
	shapes := idShapes()
	narrowed := 0
	for _, m := range minted {
		if len(m.shapes) < 2 {
			continue
		}
		wide, narrow := shapes[m.shapes[0]], shapes[m.shapes[1]]
		narrowed++
		// A value of the WIDE kind that the narrow kind must refuse: a finding anchor is an anchor,
		// and neither a citation nor a gap.
		wideOnly := mint(t, "finding")
		if !accepts(wide, wideOnly) {
			t.Fatalf("%s refuses %q, so it is not the wider kind", m.shapes[0], wideOnly)
		}
		if accepts(narrow, wideOnly) {
			t.Errorf("%s accepted %q — the two kinds have become the same shape under two names, which is "+
				"exactly what a single --id is meant to remove", m.shapes[1], wideOnly)
		}
	}
	if narrowed != 2 {
		t.Errorf("%d specialisations, want 2 (citation-anchor and gap-id, each narrowing anchor)", narrowed)
	}
}

// THE REFUSAL NAMES THE KIND, which is what a single --id buys: a seat handed back "that is a motion
// id and this verb takes a gap id" goes to the right verb, where "invalid id" sends it to re-check its
// typing. The shapes being decidable is what makes the better message possible.
func TestAWrongKindIsRefusedByNamingItsKind(t *testing.T) {
	g := GapID()
	motion := mint(t, "motion")
	err := g.Set(motion)
	if err == nil {
		t.Fatal("a gap-id flag accepted a motion id")
	}
	if got := fmt.Sprint(err); !strings.Contains(got, motion) || !strings.Contains(got, "gap-id") || !strings.Contains(got, "G-") {
		t.Errorf("the refusal %q does not quote the value, name the kind wanted and show its letter", got)
	}
	// The flag's own type name is what pflag prints and what the refusal is built from, so a kind
	// that cannot name itself cannot produce the message this consolidation depends on.
	if g.Type() != "gap-id" {
		t.Errorf("the shape does not name its kind (%q), so no refusal can say what was wanted", g.Type())
	}
	for name, mk := range idShapes() {
		if got := mk().Type(); got != name {
			t.Errorf("the %s shape reports its type as %q", name, got)
		}
	}
}
