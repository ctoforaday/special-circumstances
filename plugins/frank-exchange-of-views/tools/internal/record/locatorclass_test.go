package record

import (
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
// # Why a table over the CLASS rather than a test per site
//
// A test per site covers the sites someone thought of; this table says what the CLASS owes, and a
// new implementation of the same comparison fails here until it is added and made anchor-blind.
// anchortext is in the table as the REFERENCE: it matches "the report minus its invisible
// annotation layer". EstoppelConflict folds events rather than comparing two arguments, so it has
// a record-backed test below — TestEstoppelSurvivesTheAnchorOnItsOwnPrescription.

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
