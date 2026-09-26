package flags

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// CAN ONE --id CARRY EVERY ID? (gblock: "I'm ok for a single id to support all of the ids. That's
// good. We can prove no collision." — and "some IDs on one flag and others on another make no sense.")
//
// `--id` already meant a gap, a motion, an inquiry and a proof digest depending on the verb, while
// `--anchor` and `--sha` carried three more kinds under their own names (#1172). Consolidating them
// onto one flag that resolves by SHAPE is sound only if the shapes decide a value unambiguously — so
// this is the proof, and it is over the shapes the flag package ALREADY declares rather than a second
// table written to match them.
//
// THE ANSWER IS "YES, WITH ONE STATED EXCEPTION", and the exception is the reason this test exists
// rather than a comment claiming disjointness: `citation-anchor` is a deliberate SUBSET of `anchor`,
// because `lens verify` takes a citation and must refuse a finding or a proof anchor. A resolver over
// these shapes therefore resolves MOST-SPECIFIC-FIRST; it cannot assume exactly one match. A test
// asserting plain disjointness would fail on that pair and would be wrong to.

// idShapes is every shaped id value the flag package offers, paired with a sample its own minter
// would produce. The samples are what make this more than a regex crossword: a shape that describes
// nothing the tool mints cannot collide with anything and would pass a pattern-only check.
func idShapes(t *testing.T) []struct {
	kind   string
	v      *ShapedValue
	sample string
} {
	t.Helper()
	sum := sha256.Sum256([]byte("a stored proof run"))
	return []struct {
		kind   string
		v      *ShapedValue
		sample string
	}{
		{"gap-id", GapID(), "G7"},
		{"motion-id", MotionID(), "M1"},
		{"inquiry-id", InquiryID(), "Q1"},
		{"anchor", AnchorID(), "f-7daddb6a"},
		{"citation-anchor", CitationAnchor(), "c-a1b2c3d4"},
		{"finding-label", FindingLabel(), "adversary-F2"},
		{"sha256", SHA(), hex.EncodeToString(sum[:])},
	}
}

// specialisations are the pairs where one shape deliberately narrows another. Each is an ALLOWANCE
// with a reason, not a tolerated overlap: the narrow one exists so a verb can refuse the wider kind.
var specialisations = map[string]string{
	"citation-anchor": "anchor", // `lens verify` takes a source citation; f- is a finding and p- a proof
}

func TestOneIDFlagCanResolveEveryIDByShape(t *testing.T) {
	shapes := idShapes(t)

	// EVERY DECLARED SHAPE HAS A SAMPLE THAT IT ACCEPTS. Without this the matrix below could be
	// entirely zeros and read as perfect disjointness.
	for _, s := range shapes {
		if err := s.v.Set(s.sample); err != nil {
			t.Fatalf("%s rejects %q, which is the shape its own minter produces — the sample or the shape is wrong: %v", s.kind, s.sample, err)
		}
	}

	for _, a := range shapes {
		for _, b := range shapes {
			if a.kind == b.kind {
				continue
			}
			// Does a's sample satisfy b's shape?
			if err := b.v.Set(a.sample); err == nil {
				if specialisations[a.kind] == b.kind {
					continue // the narrow kind's value is of course a valid wide one
				}
				t.Errorf("a %s (%q) also satisfies %s — one --id cannot resolve by shape, so these two kinds "+
					"need either distinct shapes or an entry in `specialisations` saying which narrows which",
					a.kind, a.sample, b.kind)
			}
		}
	}
}

// A SPECIALISATION IS ONE-WAY, and this is what makes it a specialisation rather than a collision:
// every citation anchor is an anchor, and not every anchor is a citation anchor. If the wide kind
// started accepting only what the narrow one does they would be the same shape under two names, which
// is the defect #1172 is about.
func TestASpecialisationNarrowsAndIsNotMutual(t *testing.T) {
	for narrow, wide := range specialisations {
		var nv, wv *ShapedValue
		for _, s := range idShapes(t) {
			switch s.kind {
			case narrow:
				nv = s.v
			case wide:
				wv = s.v
			}
		}
		if nv == nil || wv == nil {
			t.Fatalf("specialisations names %q/%q and idShapes does not offer both", narrow, wide)
		}
		// A value of the WIDE kind that the narrow kind must refuse. f- is a finding anchor: an
		// anchor, and not a citation.
		const wideOnly = "f-7daddb6a"
		if err := wv.Set(wideOnly); err != nil {
			t.Fatalf("%s refuses %q, so it is not the wider kind: %v", wide, wideOnly, err)
		}
		if err := nv.Set(wideOnly); err == nil {
			t.Errorf("%s accepted %q — the two kinds have become the same shape under two names, which is "+
				"exactly what a single --id is meant to remove", narrow, wideOnly)
		}
	}
}

// THE REFUSAL NAMES THE KIND, which is what a single --id buys: a seat handed back "that is a motion
// id and this verb takes a gap id" goes to the right verb, where "invalid id" sends it to re-check its
// typing. The shapes being decidable is what makes the better message possible.
func TestAWrongKindIsRefusedByNamingItsKind(t *testing.T) {
	g := GapID()
	err := g.Set("M1")
	if err == nil {
		t.Fatal("a gap-id flag accepted a motion id")
	}
	if got := fmt.Sprint(err); got == "" {
		t.Fatal("the refusal is empty")
	}
	// The flag's own type name is what pflag prints and what the refusal is built from, so a kind
	// that cannot name itself cannot produce the message this consolidation depends on.
	if g.Type() != "gap-id" {
		t.Errorf("the shape does not name its kind (%q), so no refusal can say what was wanted", g.Type())
	}
}
