package record

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// AN ID IS UNIQUE ACROSS THE WHOLE RECORD, NOT ONLY WITHIN ITS OWN KIND.
//
// The invariant this file exists to hold: **no id minted by one part of the system can ever be
// read as an id of a different kind.** Not "unlikely to be confused" — unable to be. A seat that
// passes the right id to the wrong verb, or the wrong id to the right verb, must get a refusal,
// never a silent join to a different entity.
//
// WHY IT MATTERS MORE AFTER #344. `motion <subject> rule --id` takes an M-number for a grade or a
// petition and an A-number for a direction, chosen by SUBGROUP. If those two namespaces could
// ever produce the same string, the subgroup would decide which entity a caller meant — and the
// probe that produced this file found the tool already deciding one fact (the motion's subject)
// from the subgroup rather than the record, so this is not a hypothetical shape.
//
// The guarantee is bought with PREFIXES, and prefixes are only a guarantee while they are
// distinct. Nothing structural stops a future minter from choosing `M`; this test is that
// structure. It is deliberately written against the MINTERS rather than against a list of
// strings, so a new id kind fails here on the day it is added rather than on the day it collides.

// idKind is one minted namespace: what mints it, and the shape it produces.
type idKind struct {
	name string
	// pattern must match every id this kind mints and NOTHING another kind mints. It is READ FROM
	// internal/flags where the shape is declared, never restated here: this table used to carry
	// its own `^A\d+$` and siblings, and when the line-of-inquiry id moved A -> Q the copy stayed,
	// so the matrix reported the MINTER as wrong against a pattern nobody had updated. A matrix
	// that exists to catch namespace drift cannot itself hold a second copy of the namespace.
	// The finding label is the one exception below, with its reason.
	pattern *regexp.Regexp
	// mint produces the next id of this kind in a run directory.
	mint func(runDir string) (string, error)
}

func idKinds() []idKind {
	return []idKind{
		{
			name:    "gap",
			pattern: flags.GapID().Shape(),
			mint:    func(runDir string) (string, error) { return MintGapID(runDir, 1) },
		},
		{
			name:    "line-of-inquiry",
			pattern: flags.InquiryID().Shape(),
			mint:    MintInquiryID,
		},
		{
			name:    "motion",
			pattern: flags.MotionID().Shape(),
			mint:    MintMotionID,
		},
		{
			name: "finding",
			// NOT from flags: FindingLabel's shape is the id a LENS carries (`L2-F1`), while the
			// minter here is exercised with a seat id, so the two describe different halves of
			// the same vocabulary. Restated deliberately, which is what an exception looks like.
			pattern: regexp.MustCompile(`^[A-Za-z0-9]+-F\d+$`),
			mint:    func(runDir string) (string, error) { return NextFindingLabel(runDir, "red-lens-r1-L1") },
		},
	}
}

// EVERY KIND'S SHAPE REJECTS EVERY OTHER KIND'S IDS.
//
// This is the whole invariant, stated as a matrix rather than as a promise. It mints a run of ids
// from each kind and checks that no other kind's pattern accepts them — so a new minter that
// picks a colliding prefix fails here, naming both sides of the collision.
func TestNoIDKindCanBeReadAsAnother(t *testing.T) {
	kinds := idKinds()
	if len(kinds) < 4 {
		t.Fatalf("only %d id kinds — a matrix this small stops being a sweep, and the point is that EVERY minter is in it", len(kinds))
	}

	// Several of each, because a collision can hide at n=1: `A1` and `M1` differ, and a kind
	// that minted `A10` while another minted `A1` followed by `0` would not.
	minted := map[string][]string{}
	for _, k := range kinds {
		runDir := t.TempDir()
		if err := writeSeat(t, runDir); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 12; i++ {
			id, err := k.mint(runDir)
			if err != nil {
				t.Fatalf("%s: mint %d: %v", k.name, i, err)
			}
			minted[k.name] = append(minted[k.name], id)
			// Minting is derived from the RECORD, so the next id only advances once the
			// previous one is on it. Write the event the minter counts.
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
						"A command taking a %s id would join it to a %s, silently and at replay. The prefixes ARE the guarantee; one of these two kinds must change.",
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
}

// THE SHAPES ARE DISTINGUISHED BY MORE THAN THEIR FIRST CHARACTER.
//
// A prefix scheme that leans on one letter is one entity away from exhaustion, and the failure
// would arrive as a collision rather than as a naming argument. This states the current scheme so
// that adding a kind is a deliberate act: pick a letter no other kind uses, and this test is where
// you find out you cannot.
func TestEveryIDKindHasADistinctPrefixLetter(t *testing.T) {
	seen := map[string]string{}
	for _, k := range idKinds() {
		// The literal prefix a pattern pins, e.g. `^R\d+-\d+$` -> "R". A kind whose pattern
		// starts with a character class pins nothing and is reported as such.
		src := k.pattern.String()
		if !strings.HasPrefix(src, "^") || len(src) < 2 {
			t.Errorf("%s's pattern %q is not anchored — an unanchored id pattern matches inside another kind's id", k.name, src)
			continue
		}
		letter := string(src[1])
		if letter == "[" || letter == "(" || letter == `\` {
			// `finding` is the honest exception: its label is ROLE-scoped (`L1-F3`), so its
			// leading token is a seat name rather than a fixed letter. The `-F` infix is what
			// makes it unmistakable, and the matrix above is what proves it.
			continue
		}
		if prev, dup := seen[letter]; dup {
			t.Errorf("%s and %s both mint ids beginning %q — the first character is doing the work of telling them apart, and now it cannot", prev, k.name, letter)
		}
		seen[letter] = k.name
	}
}

// writeSeat registers the seat every minter counts events from.
func writeSeat(t *testing.T, runDir string) error {
	t.Helper()
	_, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")})
	return err
}

// appendMintedFor writes the event a minter counts, so the next mint advances.
func appendMintedFor(t *testing.T, runDir, kind, id string) error {
	t.Helper()
	switch kind {
	case "gap":
		_, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Mint{Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		return err
	case "line-of-inquiry":
		_, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Avenue{Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Line: proto.String("a line"), Reason: proto.String("r")})
		return err
	case "motion":
		_, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, "motion", NewPayload().
			Set("motion_id", id).Set("subject", "petition").Set("reason", "b").Set("class", "safety"))
		return err
	case "finding":
		_, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Finding{Location: proto.String("L"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
		return err
	}
	return fmt.Errorf("no recorder for id kind %q — add one, or the minter never advances and this test compares twelve copies of the same id", kind)
}
