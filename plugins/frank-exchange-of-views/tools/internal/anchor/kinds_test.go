package anchor

import (
	"encoding/hex"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// EVERY COLUMN OF THE KINDS TABLE IS READ. A row's meaning reaches its readers only through these
// functions, so each column is pinned per kind: deleting a row, or flipping a column, fails here.
func TestTheKindsTableSaysWhatEachKindMeans(t *testing.T) {
	for _, c := range []struct {
		id, kind, token string
		assembly        Assembly
		claim, backs    bool
	}{
		{"F-00001a2b", "finding", "<!--fx:F-00001a2b-->", Strip, false, false},
		{"C-00001a2b", "citation", "<!--cite:C-00001a2b-->", WeaveSource, true, true},
		{"P-00001a2b", "proof", "<!--proof:P-00001a2b-->", WeaveProof, false, true},
		{"G-00000003", "gap", "<!--gap:G-00000003-->", Strip, false, false},
	} {
		if got := Kind(c.id); got != c.kind {
			t.Errorf("Kind(%q) = %q, want %q", c.id, got, c.kind)
		}
		if got := Token(c.id); got != c.token {
			t.Errorf("Token(%q) = %q, want %q", c.id, got, c.token)
		}
		if got := AssemblyOf(c.id); got != c.assembly {
			t.Errorf("AssemblyOf(%q) = %v, want %v", c.id, got, c.assembly)
		}
		if got := CountsAsClaim(c.id); got != c.claim {
			t.Errorf("CountsAsClaim(%q) = %v, want %v", c.id, got, c.claim)
		}
		if got := Backs(c.id); got != c.backs {
			t.Errorf("Backs(%q) = %v, want %v", c.id, got, c.backs)
		}
	}
	if got, want := Kinds(), []string{"finding", "citation", "proof", "gap"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Kinds() = %v, want %v — the walk order every reader reports in", got, want)
	}
	// AN ID NO ROW CLAIMS HAS NO KIND — an avenue's, a motion's, a lower-case letter, a word. It is no
	// finding, backs nothing, counts as no claim, its label passes through unchanged, and the token
	// spelled from it is one no reader takes for an anchor.
	for _, id := range []string{"Q-00000003", "M-00000003", "G3", "f-00001a2b", "banana"} {
		if Kind(id) != "" || Backs(id) || CountsAsClaim(id) || AssemblyOf(id) != Strip || Label(id) != id {
			t.Errorf("an unclaimed id %q: Kind=%q Backs=%v CountsAsClaim=%v Assembly=%v Label=%q", id, Kind(id), Backs(id), CountsAsClaim(id), AssemblyOf(id), Label(id))
		}
		if tok := Token(id); len(IDs("x"+tok+"y")) != 0 || SkipRun(tok, 0) != 0 {
			t.Errorf("Token(%q) = %q reads back as an anchor token", id, tok)
		}
	}
	// ONE LIFECYCLE: every kind's label says the same thing after its noun. It names the anchor by its
	// TOKEN, never the bare id: a finding id in text stands beside its lens's area (S9), which a reader
	// of report text alone does not know, and the token is what a seat reproduces.
	for id, noun := range map[string]string{"F-00001a2b": "finding anchor", "C-00001a2b": "citation anchor", "P-00001a2b": "proof anchor", "G-00000003": "gap anchor"} {
		if got := Label(id); got != noun+" "+Token(id)+lifecycle {
			t.Errorf("Label(%s) = %q", id, got)
		}
	}
}

// The id pattern matches what the table spells and nothing else.
func TestIDPatternReadsTheTable(t *testing.T) {
	id := regexp.MustCompile(`^` + IDPattern() + `$`)
	for s, want := range map[string]bool{"F-00001a2b": true, "C-00000000": true, "P-000000ff": true, "G-00000003": true,
		"g-0000001a": false, "f-00001a2b": false, "C-": false, "C-0000XY00": false, "C-0000ABCD": false, "00001a2b": false,
		// EXACTLY eight hex: seven and nine are not an id.
		"C-0000000": false, "C-000000000": false,
		// The retired spellings, and the kinds no anchor stands for.
		"G3": false, "c-1a2b": false, "Q-00000003": false, "M-00000003": false, "X-00000003": false} {
		if got := id.MatchString(s); got != want {
			t.Errorf("IDPattern on %q = %v, want %v", s, got, want)
		}
	}
}

// idKinds is every kind the record mints an id for, with its letter and whether an anchor stands
// for it.
var idKinds = []struct {
	name, prefix string
	anchored     bool
}{
	{"finding", "F-", true}, {"citation", "C-", true}, {"proof", "P-", true}, {"gap", "G-", true},
	{"avenue", "Q-", false}, {"motion", "M-", false},
}

// ID AND IDPattern AGREE, for all six kinds: what ID spells for a kind is matched whole by that
// kind's pattern and by no other kind's, the unnamed pattern takes exactly the anchor kinds, and a
// kind's pattern refuses its own letter over seven or nine hex.
func TestIDAndIDPatternAgreeForEveryKind(t *testing.T) {
	whole := func(names ...string) *regexp.Regexp { return regexp.MustCompile(`^` + IDPattern(names...) + `$`) }
	seen := map[string]string{}
	for _, k := range idKinds {
		for _, b := range [][4]byte{{}, {0xde, 0xad, 0xbe, 0xef}, {0xff, 0xff, 0xff, 0xff}, {0, 0, 0, 1}} {
			id := ID(k.name, b)
			if want := k.prefix + hex.EncodeToString(b[:]); id != want {
				t.Errorf("ID(%q, %x) = %q, want %q", k.name, b, id, want)
			}
			if other, dup := seen[id]; dup {
				t.Errorf("ID spells %q for both %s and %s", id, other, k.name)
			}
			seen[id] = k.name
			for _, o := range idKinds {
				if got, want := whole(o.name).MatchString(id), o.name == k.name; got != want {
					t.Errorf("IDPattern(%q) on the %s id %q = %v, want %v", o.name, k.name, id, got, want)
				}
			}
			if got := whole().MatchString(id); got != k.anchored {
				t.Errorf("IDPattern() on the %s id %q = %v, want %v — the unnamed pattern is every ANCHOR kind", k.name, id, got, k.anchored)
			}
			if got := Kind(id) != ""; got != k.anchored {
				t.Errorf("Kind(%q) = %q, and an anchor stands for a %s: %v", id, Kind(id), k.name, k.anchored)
			}
			for _, bad := range []string{id[:len(id)-1], id + "0", strings.ToLower(id[:1]) + id[1:], id[:2] + strings.ToUpper(id[2:])} {
				if bad == id {
					continue // an all-digit id has no upper-case spelling
				}
				if whole(k.name).MatchString(bad) {
					t.Errorf("IDPattern(%q) accepts %q", k.name, bad)
				}
			}
		}
	}
	// Several names are the union of their kinds and nothing more.
	both := whole("gap", "motion")
	for id, want := range map[string]bool{"G-00000001": true, "M-00000001": true, "F-00000001": false, "Q-00000001": false} {
		if got := both.MatchString(id); got != want {
			t.Errorf(`IDPattern("gap","motion") on %q = %v, want %v`, id, got, want)
		}
	}
	// A kind neither table holds is the caller's defect.
	defer func() {
		if recover() == nil {
			t.Error(`ID("label", …) spelled an id for a kind no table holds`)
		}
	}()
	ID("label", [4]byte{})
}

// Replace and StripAssembled. The assembly strips a stripped kind's token whatever stands inside
// it — a seat may copy one into record prose with its id elided — and leaves every other kind.
func TestReplaceAndStripAssembled(t *testing.T) {
	in := "A<!--fx:F-00000001--> b<!--cite:C-00000002-->. C<!--proof:P-00000003--><!--gap:G-00000004-->."
	if got, want := Replace(in, func(_, id string) string { return "[" + id + "]" }), "A[F-00000001] b[C-00000002]. C[P-00000003][G-00000004]."; got != want {
		t.Errorf("Replace = %q, want %q", got, want)
	}
	if got := Replace("no anchors", func(string, string) string { return "x" }); got != "no anchors" {
		t.Errorf("Replace with no token = %q", got)
	}
	if got, want := StripAssembled("A<!--fx:F-00000001--> b<!--fx:…--><!--cite:C-00000002--><!--gap:G-00000001-->."), "A b<!--cite:C-00000002-->."; got != want {
		t.Errorf("StripAssembled = %q, want %q", got, want)
	}
}
