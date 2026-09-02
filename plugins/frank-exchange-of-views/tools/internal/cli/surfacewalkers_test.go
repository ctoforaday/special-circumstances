package cli

import "testing"

// AN EMPTY SURFACE MAP IS A BROKEN WALKER, NOT A SMALL SURFACE — and nothing told the two apart.
//
// CommandFlags() walked the bare newRoot(), which with no seat is the OPERATOR's tree, and
// returned zero entries. Its only consumer — the fuzz's unreachedFlags gate — iterates that map,
// so an empty map made the loop body unreachable and the gate passed on every run without ever
// comparing a flag. The gate written to catch "a verb driven with half its flags never passed"
// was itself never passing a flag through its comparison, and its green said nothing (#654).
//
// The three walkers answer three questions about ONE tree, so they must see the same tree. This
// pins that: same paths, none empty. A walker that regresses to the operator-only root now fails
// here by name rather than going quiet downstream.
func TestTheSurfaceWalkersAllSeeTheSameTree(t *testing.T) {
	paths := CommandPaths()
	flags := CommandFlags()
	records := CommandRecords()
	refs := CommandReferences()

	// THE FLOOR FIRST. Every assertion below is vacuously true over an empty collection, which is
	// exactly how this defect survived: the comparison passed because there was nothing to
	// compare. Refuse the empty case explicitly before comparing anything.
	if len(paths) == 0 {
		t.Fatal("CommandPaths() is EMPTY — the walker is broken; every gate that diffs against it now reports a clean board")
	}
	if len(flags) == 0 {
		t.Fatal("CommandFlags() is EMPTY — unreachedFlags() iterates this map, so an empty one makes that gate assert nothing while staying green (#654)")
	}
	if len(records) == 0 {
		t.Fatal("CommandRecords() is EMPTY — the observed-graph report renders \"0 of 0 recording verbs reached\", which reads exactly like full coverage")
	}
	// CommandReferences fails in ONE direction: it type-asserts flag values, so a change to the
	// flag types stops the assertion matching and the map comes back empty — reported as "0
	// checked flag(s)", which reads like a surface where nothing is checked rather than a walker
	// that stopped looking. #654 one edge over.
	if len(refs) == 0 {
		t.Fatal("CommandReferences() is EMPTY — no flag was seen to carry an existence check, which is a broken derivation rather than an unchecked surface (#535 step 2)")
	}

	// CommandFlags is keyed on the same paths CommandPaths returns: same walk, same composition.
	if len(flags) != len(paths) {
		t.Errorf("CommandFlags has %d entries and CommandPaths has %d — the two walk the same tree and must agree; a divergence means one of them is scoped to a different root",
			len(flags), len(paths))
	}
	known := map[string]bool{}
	for _, p := range paths {
		known[p] = true
	}
	for p := range flags {
		if !known[p] {
			t.Errorf("CommandFlags names %q, which CommandPaths does not — the keys must join, or every gate that looks a path up in both silently misses", p)
		}
	}
	// References is a SUBSET too, and joins on the same key: a reference edge pointing at a path
	// CommandPaths does not know is an edge to nothing.
	for p := range refs {
		if !known[p] {
			t.Errorf("CommandReferences names %q, which CommandPaths does not — the reference edge joins on this key and would point at nothing", p)
		}
	}
	// Records is a SUBSET: most paths write no event, and that is a real answer rather than a miss.
	for p := range records {
		if !known[p] {
			t.Errorf("CommandRecords names %q, which CommandPaths does not — the observed graph joins on this key and would attribute the edge to nothing", p)
		}
	}
	if len(records) >= len(paths) {
		t.Errorf("CommandRecords has %d entries against %d paths — it should be a strict subset, since `show`, `graph` and `verify` record nothing",
			len(records), len(paths))
	}
}
