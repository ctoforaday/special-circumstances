package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE LOCATOR SEES THROUGH THE ANNOTATION LAYER, EVERYWHERE IT COMPARES TWO SPANS.
//
// A gap's location, an edit's replaced span and a lens's prescribed text are all quotes of the
// report, and the report carries an invisible layer of anchors — `<!--fx:…-->`, `<!--cite:…-->`,
// `<!--proof:…-->`. Minting a gap PLACES one at the location it names, so the stored location and
// the same sentence as an edit later quotes it differ by bytes nobody typed. Any comparison of two
// such spans on RAW BYTES therefore stops matching the moment the sentence gains an anchor, and it
// stops matching SILENTLY: the answer is "these are different spans", which is also the honest
// answer for two genuinely different sentences.
//
// # What that cost, measured
//
// On universe-m11's 2026-09-26_is-91-prime run, gap G4's stored location was
// "Eight authoritative mathematical sources were consulted:" while the report said
// "Six authoritative mathematical sources were consulted<!--fx:f-dbd94684-->:" — the anchor sits
// between "consulted" and the colon, so neither string contains the other. Four `blue edit` acts
// named G4 in --answers and `gap_edit` attributed NONE of them, so the work list delivered that
// stale location beside `edited_since: []` — a field asserting nothing had changed. THREE OF FIVE
// located gaps in that run under-reported their edits the same way.
//
// The seats then re-derived the board and the report for themselves, which is the correct response
// to a freshness signal that cannot be trusted, and is most of what a re-audit sitting spent its
// calls on.
//
// # Why a table over the CLASS rather than a test per site
//
// Three sites implemented this comparison independently and all three were raw: the `gap_edit`
// view, CurrentLocation's fold, and EstoppelConflict. A test per site covers the sites someone
// thought of; this table says what the CLASS owes, and a fourth implementation of the same
// comparison fails here until it is added and made anchor-blind.
//
// anchortext is in the table as the REFERENCE: it already matches "the report minus its invisible
// annotation layer" and always did, which is what makes the other three deviations rather than an
// unsolved problem.
//
// TWO MEMBERS OF THE CLASS CANNOT BE REACHED BY A PURE ADAPTER and have record-backed tests below,
// named here so the class is enumerated in ONE place:
//   - the `gap_edit` view, which is SQL — TestGapEditAttributesAnEditToAnAnchoredLocation;
//   - EstoppelConflict, which folds events rather than comparing two arguments —
//     TestEstoppelSurvivesTheAnchorOnItsOwnPrescription.

// anchored is the same sentence as the report stores it once a finding has been minted against it.
const (
	plainSentence    = "Eight authoritative mathematical sources were consulted:"
	anchoredSentence = "Eight authoritative mathematical sources were consulted<!--fx:f-dbd94684-->:"
	otherSentence    = "Trial division identified 7 as a divisor."
)

// sameSpan is one site's answer to "are these two quotes the same span of the report".
type sameSpan struct {
	name string
	// eq reports whether the site treats a and b as the same span. Each adapter is the site's own
	// comparison, reached the way production reaches it.
	eq func(t *testing.T, a, b string) bool
}

func locatorSites() []sameSpan {
	return []sameSpan{
		{
			// THE REFERENCE. LocateSpan finds a quote in a report that carries the layer.
			name: "anchortext.LocateSpan",
			eq: func(t *testing.T, a, b string) bool {
				// b is the report as stored; a is the quote a seat types.
				start, end := anchortext.LocateSpan("intro\n\n"+b+"\n\ntail", a)
				return start >= 0 && end > start
			},
		},
		{
			name: "record.CurrentLocation",
			eq: func(t *testing.T, a, b string) bool {
				// An edit whose replaced span is b must relocate a gap minted at a.
				got := CurrentLocation(a, []GapEdit{{Old: b, New: "REPLACED"}})
				return got != a
			},
		},
	}
}

func TestEveryLocatorComparisonSeesThroughTheAnnotationLayer(t *testing.T) {
	for _, site := range locatorSites() {
		t.Run(site.name, func(t *testing.T) {
			// NOT VACUOUS: the site must match a sentence against itself, or every assertion below
			// passes on a comparison that answers "same" to everything or "different" to everything.
			if !site.eq(t, plainSentence, plainSentence) {
				t.Fatalf("%s does not match a sentence against itself — the adapter is wrong, not the site", site.name)
			}
			if site.eq(t, plainSentence, otherSentence) {
				t.Fatalf("%s matches two different sentences, so it would pass the anchor case for the wrong reason", site.name)
			}
			// THE CLASS: the same sentence, one side carrying the anchor its own mint placed.
			if !site.eq(t, plainSentence, anchoredSentence) {
				t.Errorf("%s does NOT see %q and %q as the same span.\n"+
					"They differ only by a finding anchor, which minting places at the location the gap names — so this\n"+
					"comparison stops matching exactly when a gap is minted, and stops matching SILENTLY: 'different span'\n"+
					"is also the honest answer for two unrelated sentences. Compare through the annotation layer.",
					site.name, plainSentence, anchoredSentence)
			}
			// And the other way round, because either side may be the stored one.
			if !site.eq(t, anchoredSentence, plainSentence) {
				t.Errorf("%s sees the pair as the same span in one direction only — the stored side is not always the anchored one", site.name)
			}
		})
	}
}

// AND THE SAME RULE THROUGH THE REAL RECORD, because the `gap_edit` view is SQL and cannot be
// reached by a Go adapter. This is the m11 defect reproduced end to end: mint a gap at a sentence,
// have blue edit that sentence WITH the anchor the mint placed, and ask the record what moved.
func TestGapEditAttributesAnEditToAnAnchoredLocation(t *testing.T) {
	run := corrRun(t)
	lens := sit(t, run, "red-lens-computation")
	blue := sit(t, run, "blue-respond")

	mustAppend(t, lens, &recordpb.Mint{
		GapId:           proto.String("G1"),
		Class:           proto.String("unverified-composition"),
		Location:        proto.String(plainSentence),
		Problem:         proto.String("the count and the list disagree"),
		AcceptanceCheck: proto.String("the count matches the list"),
		RequiredFix:     proto.String("state the count the list supports"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity:        recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
	})
	// The edit quotes the sentence AS THE REPORT HOLDS IT — anchor included, which is what the
	// tool requires (an edit may not strand an anchor, so blue's quote carries it).
	mustAppend(t, blue, &recordpb.BlueEdit{
		Answers: proto.String("G1"),
		Old:     proto.String(anchoredSentence),
		New:     proto.String("Six authoritative mathematical sources were consulted<!--fx:f-dbd94684-->:"),
		Text:    proto.String("corrected the count to match the list"),
	})

	edits, err := GapEdits(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(edits["G1"]) == 0 {
		t.Fatalf("the record attributes NO edit to G1, though blue replaced the very sentence G1 names.\n"+
			"The stored location is %q and the edit replaced %q — the same sentence, differing by the finding\n"+
			"anchor G1's own mint placed there. A reader of `edited_since` is told nothing changed, which is the\n"+
			"same bytes as a gap nobody has touched.", plainSentence, anchoredSentence)
	}
	if loc := CurrentLocation(plainSentence, edits["G1"]); !strings.Contains(loc, "Six") {
		t.Errorf("the gap's location did not follow the edit: got %q, want the replacement text", loc)
	}
}

// ESTOPPEL MUST SURVIVE THE ANCHOR ON ITS OWN PRESCRIPTION.
//
// A lens may not open a clean slate against text blue applied verbatim at its instruction. The
// check compares the lens's new quote against the `fix_new` it prescribed — and the prescription is
// red's own prose, written before any anchor existed, while the quote is taken from the report,
// which anchors the sentence the moment the fix lands. Compare those on raw bytes and estoppel
// stops firing exactly when it is needed, reporting no conflict — the same answer as a lens that is
// auditing something else entirely.
func TestEstoppelSurvivesTheAnchorOnItsOwnPrescription(t *testing.T) {
	const prescribed = "Six authoritative mathematical sources were consulted:"
	run := corrRun(t)
	lens := sit(t, run, "red-lens-computation")
	blue := sit(t, run, "blue-respond")

	mustAppend(t, lens, &recordpb.Mint{
		GapId:           proto.String("G1"),
		Class:           proto.String("self-attestation"),
		Location:        proto.String(plainSentence),
		Problem:         proto.String("the count and the list disagree"),
		RequiredFix:     proto.String("state the count the list supports"),
		FixNew:          proto.String(prescribed),
		AcceptanceCheck: proto.String("the count matches the list"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity:        recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
	})
	// Blue takes the prescription as given, and the tool anchors the sentence it lands on.
	mustAppend(t, blue, &recordpb.BlueEdit{
		Answers:         proto.String("G1"),
		Old:             proto.String(plainSentence),
		New:             proto.String(prescribed),
		AppliedVerbatim: proto.Bool(true),
		Text:            proto.String("applied red's text as given"),
	})

	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	// The lens comes back and quotes the sentence AS THE REPORT NOW HOLDS IT.
	anchored := "Six authoritative mathematical sources were consulted<!--fx:f-dbd94684-->:"
	if id, _ := EstoppelConflict(fam, anchored); id == "" {
		t.Errorf("estoppel did not fire on a quote of red's own applied prescription.\n"+
			"prescribed: %q\nquoted    : %q\nThey differ only by the anchor the report carries, so the one rule that stops a lens\n"+
			"relitigating its own words is blind from the moment the words land.", prescribed, anchored)
	}
	// NOT VACUOUS: an unrelated quote must still be free.
	if id, _ := EstoppelConflict(fam, otherSentence); id != "" {
		t.Errorf("estoppel fired on unrelated text (%q -> %s), so the assertion above would pass for the wrong reason", otherSentence, id)
	}
}

// THE CHAIN IS FOLLOWED TO THE END, not one step of it.
//
// `gap_edit` matches every edit against the gap's MINTED location, so an edit overlapping only an
// INTERMEDIATE form is never selected. Measured on universe-m11 after the anchor fix: G4 was rewritten
// Eight -> Seven -> Six and its location read "Seven…" while the report read "Six…" — a stale location
// by a second mechanism, transitive relocation rather than a byte comparison.
//
// SO THE ORDERED WALK LIVES IN GO. Following a chain means choosing, at each step, the NEXT edit that
// overlaps the location as it now stands; in SQL that is a recursive CTE whose recursive term needs
// ORDER BY … LIMIT 1, which SQLite allows neither of, and a plain UNION ALL enumerates orderings
// instead. GapEdits does the walk and CurrentLocation replays what it chose, both through the one
// per-step rule (relocate), so attribution and relocation cannot disagree about what "this edit moved
// that text" means.
func TestARelocationFollowsTheWholeChain(t *testing.T) {
	const second = "Seven authoritative mathematical sources were consulted<!--fx:f-dbd94684-->:"
	const third = "Six authoritative mathematical sources were consulted<!--fx:f-dbd94684-->:"
	run := corrRun(t)
	lens := sit(t, run, "red-lens-computation")
	blue := sit(t, run, "blue-respond")

	mustAppend(t, lens, &recordpb.Mint{
		GapId:           proto.String("G1"),
		Class:           proto.String("self-attestation"),
		Location:        proto.String(plainSentence),
		Problem:         proto.String("the count and the list disagree"),
		RequiredFix:     proto.String("state the count the list supports"),
		AcceptanceCheck: proto.String("the count matches the list"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity:        recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
	})
	// The second edit overlaps only what the FIRST one produced, which is the case the minted-text
	// join cannot see.
	mustAppend(t, blue, &recordpb.BlueEdit{Answers: proto.String("G1"),
		Old: proto.String(anchoredSentence), New: proto.String(second), Text: proto.String("eight is wrong")})
	mustAppend(t, blue, &recordpb.BlueEdit{Answers: proto.String("G1"),
		Old: proto.String(second), New: proto.String(third), Text: proto.String("seven is wrong too")})

	edits, err := GapEdits(run)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(edits["G1"]), 2; got != want {
		t.Errorf("the record attributes %d edit(s) to G1, want %d — an edit that overlaps only an "+
			"intermediate form of the sentence is still an edit to that sentence", got, want)
	}
	loc := CurrentLocation(plainSentence, edits["G1"])
	if !strings.Contains(loc, "Six") {
		t.Errorf("the gap's location stopped short of the last edit: got %q, want the final text — a "+
			"location one rewrite behind the report is the stale pointer this whole mechanism exists to "+
			"prevent", loc)
	}

	// AND AN UNRELATED EDIT IS NOT SWEPT IN by the widened walk. The chain must follow the sentence,
	// not every edit after the mint.
	other := sit(t, run, "blue-respond")
	mustAppend(t, other, &recordpb.BlueEdit{Answers: proto.String("G1"),
		Old: proto.String(otherSentence), New: proto.String("Trial division found 7."), Text: proto.String("unrelated")})
	again, err := GapEdits(run)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(again["G1"]); got != 2 {
		t.Errorf("an edit to a different sentence was attributed to G1: %d edit(s) now, want 2", got)
	}
}
