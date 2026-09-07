package record

import (
	"sort"
	"strings"
	"testing"
)

// PAYLOAD KEYS SHARED BY MORE THAN ONE EVENT TYPE, held as a maintained record rather than as
// prose in an issue.
//
// WHY THIS EXISTS. #95 listed four such collisions in July as "hazards". By August one of them
// had caused a real defect and three had quietly stopped existing:
//
//	verdict     LIVE  — caused #410
//	class       GONE  — petition no longer uses the key
//	disposition GONE  — the `dispose` verb was retired
//	label       INERT — two writers, but all twelve readers are type-scoped
//
// A hand-written list that rots 3-of-4 in three weeks is worse than no list: it reads as an
// inventory and is a memory. So this one is COMPUTED from EnumFields and checked BOTH ways —
// a new collision must be classified by a person, and an entry that no longer collides must be
// deleted. It cannot drift in either direction without failing.
//
// SCOPE: enum-carrying keys only, and that is the point rather than a shortcut. The danger is
// a value compared against a VOCABULARY: when two types share a key with DISJOINT vocabularies,
// reading the wrong type is a type error Go cannot see, the comparison never matches, and the
// check reports "not applicable" forever instead of failing. Identifier keys like `label` carry
// no vocabulary, so a cross-type read yields a wrong value LOUDLY rather than a silent zero.
var knownEnumKeyCollisions = map[string]string{
	"verdict": "`verdict` events carry PASS|FAIL; `outcome` events carry the run's terminal " +
		"stamp VERIFIED|CEILING|HALTED|UNVERIFIED. DISJOINT, so a reader pointed at the wrong " +
		"type matches nothing and says so in words that read like a judgement. That is exactly " +
		"what happened: verify's pass-closes-all-gaps scanned `outcome` for PASS and reported " +
		"'gate not applicable' on every run ever recorded (#410). Any reader of this key MUST " +
		"name its event type in a constant asserted against enum(), the way passClosesAllGaps " +
		"does — see TestPassVerdictIsAWordItsEventTypeCanActuallyCarry.",
	"about_kind": "`finding` and `mint` declare it with the SAME three values, and that is the " +
		"reason it is safe rather than an excuse for it. The hazard this map guards is DISJOINT " +
		"vocabularies under one key, where a cross-type read matches nothing and reports the miss " +
		"as an answer; identical vocabularies have no such miss. They are identical on purpose: a " +
		"gap is where a finding ends up, and `merge mint`'s help used to send an omission back to " +
		"the borrowed quote `lens finding` had just stopped taking. " +
		"TestAFindingAndAGapAnchorToTheSameThings compares the two sets rather than restating " +
		"them, so a value added to one and not the other fails there — which is what keeps this " +
		"entry true.",
}

// sharedEnumKeys returns every payload key declared by more than one event type.
func sharedEnumKeys() map[string][]string {
	byKey := map[string][]string{}
	for typ, fields := range EnumFields {
		for _, f := range fields {
			byKey[f.Key] = append(byKey[f.Key], typ)
		}
	}
	for k, types := range byKey {
		if len(types) < 2 {
			delete(byKey, k)
			continue
		}
		sort.Strings(types)
	}
	return byKey
}

func TestEverySharedEnumKeyIsClassified(t *testing.T) {
	shared := sharedEnumKeys()
	for key, types := range shared {
		if _, ok := knownEnumKeyCollisions[key]; !ok {
			t.Errorf("payload key %q is now declared by %v and is CLASSIFIED NOWHERE.\n"+
				"Two event types sharing one key is how #410 happened: if the vocabularies are "+
				"disjoint a cross-type read can never match, and reports 'not applicable' "+
				"forever instead of failing. Add it to knownEnumKeyCollisions with the reason "+
				"it is safe, or give one of the two types its own key.", key, types)
		}
	}
}

// THE OTHER DIRECTION, which is the half #95's prose list did not have. An entry that no longer
// collides is a stale warning, and a list of stale warnings stops being read.
func TestNoClassifiedCollisionHasGoneAway(t *testing.T) {
	shared := sharedEnumKeys()
	for key := range knownEnumKeyCollisions {
		if _, still := shared[key]; !still {
			t.Errorf("payload key %q is classified as a collision but is no longer declared by "+
				"more than one event type. Delete the entry — three of #95's four listed "+
				"collisions had silently resolved themselves, which is why this check runs both "+
				"ways.", key)
		}
	}
}

// The vocabularies of a shared key must be checked for DISJOINTNESS explicitly, because that is
// what decides whether a wrong read is loud or silent. Overlapping vocabularies would make a
// cross-type read return a plausible value; disjoint ones make it return nothing at all.
func TestTheVerdictCollisionIsStillDisjoint(t *testing.T) {
	outcome, ok1 := enum("outcome", "verdict")
	verdict, ok2 := enum("verdict", "verdict")
	if !ok1 || !ok2 {
		t.Fatal("the verdict/outcome collision is classified but one side no longer declares the key")
	}
	for _, v := range verdict.Values {
		if outcome.allows(v.Name) {
			t.Errorf("`verdict` and `outcome` now share the value %q. They were disjoint, which "+
				"is what made a cross-type read silent; if they overlap, the failure mode has "+
				"CHANGED and every reader of this key needs re-reading.", v.Name)
		}
	}
	// And the specific word the #67 gate switches on must belong to exactly one of them.
	if outcome.allows("PASS") || !verdict.allows("PASS") {
		t.Errorf("PASS must be a `verdict` word and not an `outcome` word; that asymmetry is the " +
			"whole content of #410")
	}
}

// The classification must say something. An entry with no reason is a name on a list, which is
// the shape this file exists to replace.
func TestEveryClassificationCarriesItsReason(t *testing.T) {
	for key, why := range knownEnumKeyCollisions {
		if len(strings.TrimSpace(why)) < 40 {
			t.Errorf("collision %q is classified with no usable reason (%q). The next reader "+
				"needs to know WHY it is safe and what a reader of this key must do.", key, why)
		}
	}
}
