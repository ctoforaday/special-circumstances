package record

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// AN ID IS UNIQUE ACROSS THE WHOLE RECORD, NOT ONLY WITHIN ITS OWN KIND.
//
// The invariant this file exists to hold: **no id minted by one part of the system can ever be
// read as an id of a different kind.** Not "unlikely to be confused" — unable to be. A seat that
// passes the right id to the wrong verb, or the wrong id to the right verb, must get a refusal,
// never a silent join to a different entity.
//
// WHY IT MATTERS MORE AFTER #344. `motion <subject> rule --id` takes a motion id for a grade or a
// petition and an avenue id for a direction, chosen by SUBGROUP. If those two namespaces could
// ever produce the same string, the subgroup would decide which entity a caller meant — and the
// probe that produced this file found the tool already deciding one fact (the motion's subject)
// from the subgroup rather than the record, so this is not a hypothetical shape.
//
// The guarantee is bought with the LETTER, and letters are only a guarantee while they are
// distinct: every id is one shape, <LETTER>-<8 hex>, so the letter is ALL that tells two kinds
// apart. Nothing structural stops a future kind from choosing `M`; this test is that structure.
// It is deliberately written against the MINTER rather than against a list of strings, so a new id
// kind fails here on the day it is added rather than on the day it collides.

// idKind is one minted namespace: the kind the minter is asked for, and the shape it produces.
type idKind struct {
	name string
	// pattern must match every id this kind mints and NOTHING another kind mints. It is READ FROM
	// internal/flags where a flag takes this kind alone, and from the anchor kinds table flags
	// itself reads where none does — a finding and a proof are named on the general anchor flag —
	// never restated here: a matrix that exists to catch namespace drift cannot itself hold a
	// second copy of the namespace.
	pattern *regexp.Regexp
	// anchored says an anchor stands for this kind, so the general anchor flag takes its id.
	anchored bool
}

func kindShape(kind string) *regexp.Regexp {
	return regexp.MustCompile(`^` + anchor.IDPattern(kind) + `$`)
}

func idKinds() []idKind {
	return []idKind{
		{name: "gap", pattern: flags.GapID().Shape(), anchored: true},
		{name: "avenue", pattern: flags.AvenueID().Shape()},
		{name: "motion", pattern: flags.MotionID().Shape()},
		{name: "citation", pattern: flags.CitationAnchor().Shape(), anchored: true},
		{name: "finding", pattern: kindShape("finding"), anchored: true},
		{name: "proof", pattern: kindShape("proof"), anchored: true},
	}
}

// EVERY KIND'S SHAPE REJECTS EVERY OTHER KIND'S IDS.
//
// This is the whole invariant, stated as a matrix rather than as a promise. It mints a run of ids
// from each kind, records each one, and checks that no other kind's pattern accepts them — so a
// new kind that picks a colliding letter fails here, naming both sides of the collision.
func TestNoIDKindCanBeReadAsAnother(t *testing.T) {
	kinds := idKinds()
	if len(kinds) != len(idLetters) {
		t.Fatalf("%d id kinds in the matrix and %d the minter is asked for — the point is that EVERY kind is in it", len(kinds), len(idLetters))
	}
	for _, k := range kinds {
		if idLetters[k.name] == "" {
			t.Fatalf("the matrix holds %q, which is no kind the minter is asked for", k.name)
		}
	}

	// Several of each: the id is random, and one sample proves one sample.
	minted := map[string][]string{}
	for _, k := range kinds {
		runDir := newRun(t)
		if err := writeSeat(t, runDir); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 12; i++ {
			id := NewID(k.name)
			minted[k.name] = append(minted[k.name], id)
			// THE WRITE PATH TAKES WHAT THE MINTER MAKES: each id goes onto the record under the
			// event of its kind.
			if err := appendMintedFor(t, runDir, k.name, id); err != nil {
				t.Fatalf("%s: record %s: %v", k.name, id, err)
			}
		}
	}

	// 1. Every id matches its OWN kind.
	for _, k := range kinds {
		for _, id := range minted[k.name] {
			if !k.pattern.MatchString(id) {
				t.Errorf("%s minted %q, which its own pattern %s does not match — the pattern is the only statement of the namespace, and one that does not describe its minter guarantees nothing",
					k.name, id, k.pattern)
			}
		}
	}

	// 2. And NO id matches any OTHER kind. This is the invariant.
	for _, mine := range kinds {
		for _, theirs := range kinds {
			if mine.name == theirs.name {
				continue
			}
			for _, id := range minted[mine.name] {
				if theirs.pattern.MatchString(id) {
					t.Errorf("NAMESPACE COLLISION: %s minted %q and the %s pattern (%s) accepts it.\n"+
						"A command taking a %s id would join it to a %s, silently and at replay. The letters ARE the guarantee; one of these two kinds must change.",
						mine.name, id, theirs.name, theirs.pattern, theirs.name, mine.name)
				}
			}
		}
	}

	// 3. And no two kinds ever produce the same STRING, which is the property a reader actually
	// relies on and which the patterns only imply.
	owner := map[string]string{}
	for _, k := range kinds {
		for _, id := range minted[k.name] {
			if prev, seen := owner[id]; seen {
				t.Errorf("id %q is minted by BOTH %s and %s — one string, two entities, and every join on it is a coin flip", id, prev, k.name)
			}
			owner[id] = k.name
		}
	}

	// 4. THE GENERAL ANCHOR FLAG TAKES EXACTLY THE KINDS AN ANCHOR STANDS FOR. A finding has no flag
	// shape of its own: its id is an anchor id, and the anchor flag is what accepts it. The table
	// agrees — the kind it reads off the id is the kind that was minted, and "" where no anchor
	// stands for it.
	anchorShape := flags.AnchorID().Shape()
	for _, k := range kinds {
		for _, id := range minted[k.name] {
			if got := anchorShape.MatchString(id); got != k.anchored {
				t.Errorf("the anchor flag on the %s id %q: accepted=%v, want %v", k.name, id, got, k.anchored)
			}
			want := ""
			if k.anchored {
				want = k.name
			}
			if got := anchor.Kind(id); got != want {
				t.Errorf("anchor.Kind(%q) = %q for a minted %s, want %q", id, got, k.name, want)
			}
		}
	}
}

// THE KINDS ARE TOLD APART BY ONE LETTER, AND EACH KIND HOLDS ITS OWN.
//
// Every id has the one shape, so the first character is doing all the work of telling the kinds
// apart. This states the scheme so that adding a kind is a deliberate act: pick a letter no other
// kind uses, and this test is where you find out you cannot. It reads the letter off what the
// MINTER makes, then swaps every other kind's letter onto the same eight hex: the pattern must
// refuse each, so a pattern that ignored the letter — or took two — fails whatever the hex is.
func TestEveryIDKindHasADistinctPrefixLetter(t *testing.T) {
	seen := map[string]string{}
	ids := map[string]string{}
	for _, k := range idKinds() {
		src := k.pattern.String()
		if !strings.HasPrefix(src, "^") || !strings.HasSuffix(src, "$") {
			t.Errorf("%s's pattern %q is not anchored at both ends — an unanchored id pattern matches inside another kind's id", k.name, src)
		}
		id := NewID(k.name)
		if len(id) != 10 || id[1] != '-' {
			t.Fatalf("%s minted %q, which is not a letter, a hyphen and eight hex", k.name, id)
		}
		letter := id[:1]
		if want := idLetters[k.name]; letter != want {
			t.Errorf("%s mints ids beginning %q, want %q", k.name, letter, want)
		}
		if prev, dup := seen[letter]; dup {
			t.Errorf("%s and %s both mint ids beginning %q — the first character is doing the work of telling them apart, and now it cannot", prev, k.name, letter)
		}
		seen[letter] = k.name
		ids[k.name] = id
	}
	for _, k := range idKinds() {
		for other, id := range ids {
			swapped := id[:1] + ids[k.name][1:]
			if got, want := k.pattern.MatchString(swapped), other == k.name; got != want {
				t.Errorf("the %s pattern on %q (a %s letter over a %s id's hex): matched=%v, want %v", k.name, swapped, other, k.name, got, want)
			}
		}
		// And the letter is a LETTER of this kind in this case only: the lower-case spelling is no id.
		if lower := strings.ToLower(ids[k.name][:1]) + ids[k.name][1:]; k.pattern.MatchString(lower) {
			t.Errorf("the %s pattern accepts %q", k.name, lower)
		}
	}
}

// writeSeat registers the seat the ids are recorded under.
func writeSeat(t *testing.T, runDir string) error {
	t.Helper()
	_, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, "", "")
	return err
}

// appendMintedFor writes the event that carries an id of a kind.
func appendMintedFor(t *testing.T, runDir, kind, id string) error {
	t.Helper()
	switch kind {
	case "gap":
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), AcceptanceCheck: proto.String("the check runs"), Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		return err
	case "avenue":
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Avenue{AvenueId: proto.String(id), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Line: proto.String("a line"), Reason: proto.String("r")})
		return err
	case "motion":
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Motion{
			MotionId: proto.String(id),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
			Basis:    proto.String("b"),
			Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
				Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY),
			}},
		})
		return err
	case "finding":
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Finding{Id: proto.String(id), Location: proto.String("L"), Text: proto.String("t"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		return err
	case "citation":
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, corroboration(id, "https://example.org/"+id, "the claim "+id+" bears on", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS))
		return err
	case "proof":
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Proof{ProofId: proto.String(id), Script: proto.String("s.py")})
		return err
	}
	return fmt.Errorf("no recorder for id kind %q — add one, or the write path is never shown an id of this kind", kind)
}
