package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// THE DOCUMENTS SHOW THE SEAT WORK THEY CLAIM TO SHOW (gblock, 2026-10-20). Each test assembles a
// run from the record and reads the document whose blurb promises the work: red's readings of the
// citations in evidence.md, how each gap closed in docket.md, blue's reason for each edit in
// CHANGELOG.md, and every record of an inquiry's path in inquiries.md. Each value is one a seat
// wrote, and each must reach the document whole.

// docOf assembles a run of the events and returns the named document's body ("" when the set
// carries no such document).
func docOf(t *testing.T, file string, evs ...*record.Event) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	docs, err := AssembleAll(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if d.File == file {
			return d.Body
		}
	}
	return ""
}

func mustContain(t *testing.T, doc, what string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(doc, w) {
			t.Errorf("%s does not carry %q:\n%s", what, w, doc)
		}
	}
}

// RED'S READING OF EACH CITATION STANDS BESIDE THE CITATION IT READ, with its outcome and its
// confidence; a corroboration stands as a source of its own beside its claim; a reading that placed
// no citation is listed after them; a source nobody read says so.
func TestEvidenceShowsEveryReadingOfEachCitation(t *testing.T) {
	const (
		reading  = "The infobox reads Divisors 1, 7, 13, 91, which is the sentence's claim."
		refuting = "The page says nothing about 91 at all; it lists primes below 50 only."
		corrob   = "The OEIS entry A000040 runs 89, 97 with no 91 between them."
	)
	cite := func(label, title, loc string) *record.Event {
		return recordtest.Event(t, "blue-synthesize", &recordpb.Cite{Label: proto.String(label), Url: proto.String("https://ex/" + label),
			Title: proto.String(title), AccessDate: proto.String("2026-10-18"), Location: proto.String(loc)})
	}
	doc := docOf(t, FileEvidence,
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		cite("C-00000001", "91 (number), Wikipedia", "The article lists the divisors of 91 as 1, 7, 13 and 91."),
		cite("C-00000002", "Unread source", "A second sentence nobody checked."),
		recordtest.Event(t, "red-lens-evidence", &recordpb.Verify{Anchor: proto.String("C-00000001"), Claim: proto.String("The article lists the divisors of 91 as 1, 7, 13 and 91."),
			Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH), Text: proto.String(reading)}),
		recordtest.Event(t, "red-lens-logic", &recordpb.Verify{Label: proto.String("C-00000003"), Claim: proto.String("No standard list of primes contains 91."),
			Url: proto.String("https://oeis.org/A000040"), Title: proto.String("A000040, OEIS"),
			Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_MEDIUM), Text: proto.String(corrob)}),
		recordtest.Event(t, "red-lens-adversary", &recordpb.Verify{Claim: proto.String("No standard list of primes contains 91."),
			Url: proto.String("https://ex/small-primes"), Title: proto.String("Small primes"),
			Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_LOW), Text: proto.String(refuting)}),
	)
	mustContain(t, doc, "evidence.md",
		"## Citations",
		"### 91 (number), Wikipedia\n\n- **source**: https://ex/C-00000001, accessed 2026-10-18\n- **backs**: The article lists the divisors of 91 as 1, 7, 13 and 91.\n- **read by `red-lens-evidence`** (epoch 1): **supports**, confidence high — "+reading+"\n",
		"### Unread source\n\n- **source**: https://ex/C-00000002, accessed 2026-10-18\n- **backs**: A second sentence nobody checked.\n- **not read against its sentence**",
		"### A000040, OEIS\n\n- **source**: https://oeis.org/A000040\n- **backs**: No standard list of primes contains 91.\n- **red's corroboration**, found by `red-lens-logic`: blue did not cite it\n- **read by `red-lens-logic`** (epoch 1): **supports**, confidence medium — "+corrob+"\n",
		"### Sources the report does not cite (1)",
		"- **Small primes** — https://ex/small-primes\n  - **read by `red-lens-adversary`** (epoch 1): **absent**, confidence low — "+refuting+"\n    - of the sentence: No standard list of primes contains 91.",
	)
	if strings.Count(doc, corrob) != 1 {
		t.Errorf("the corroboration's reading is listed %d times, want once:\n%s", strings.Count(doc, corrob), doc)
	}
}

// A CLOSED GAP SHOWS HOW IT CLOSED: red's argument for it, the fix it required, the check that
// would settle it, and the closing lens's attestation — what it found, who verified, with what and
// against what. An open gap carries red's argument too.
func TestTheBoardShowsHowEachGapClosed(t *testing.T) {
	mint := func(gap, reason string) *record.Event {
		return recordtest.Event(t, "red-lens-voice", &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("problem of " + gap),
			RequiredFix: proto.String("fix for " + gap), AcceptanceCheck: proto.String("check for " + gap), MintReason: proto.String(reason),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:  recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
	}
	doc := docOf(t, FileDocket,
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		mint("G-00000001", "The same run-narration pattern recurs in four sections."),
		mint("G-00000002", "The limit is stated in the audit's vocabulary."),
		recordtest.Event(t, "red-lens-voice", &recordpb.Close{GapId: proto.String("G-00000001"),
			ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("The clause the gap named is gone; its literal check passes."),
			AnchorSeat: proto.String("red-lens-voice"), AnchorTool: proto.String("show report --match 'cached'"), AnchorTarget: proto.String("report head 272, line 11")}),
		mint("G-00000003", "The bound is argued and not computed."),
		recordtest.Event(t, "red-chair", &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("impasse on G-00000003"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G-00000003")}}}),
		recordtest.Event(t, "judge", &recordpb.MotionRule{MotionId: proto.String("M1"), Opinion: proto.String("the repair answers the check"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED),
				Principle: proto.String("correctness"), Tension: proto.String("cost"), ReviewFlag: proto.String("false"), Settled: proto.String("the bound stands as computed"), ReopensOn: proto.String("a counterexample below the bound")}}}),
	)
	mustContain(t, doc, "docket.md",
		"- G-00000001 | repaired | problem of G-00000001 | successor -\n"+
			"  - minted because: The same run-narration pattern recurs in four sections.\n"+
			"  - required_fix: fix for G-00000001\n"+
			"  - acceptance_check: check for G-00000001\n"+
			"  - closed on: The clause the gap named is gone; its literal check passes. — verified by `red-lens-voice`, with show report --match 'cached', against report head 272, line 11",
		"acceptance_check: check for G-00000002  \nminted because: The limit is stated in the audit's vocabulary.",
		"  - acceptance_check: check for G-00000003\n  - closed by the bench's ruling: **defect_accepted** — its opinion is in debate.md",
	)
}

// EVERY EDIT BLUE MADE STANDS IN THE CHANGELOG WITH BLUE'S REASON for it, the gap it answers, and
// the span it replaced beside the text that replaced it; a reason repeated by the next edit of the
// sitting is said to be the same.
func TestTheChangelogShowsEveryEditWithItsReason(t *testing.T) {
	edit := func(answers, old, nw, why string) *record.Event {
		return recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String(answers), Old: proto.String(old), New: proto.String(nw), Text: proto.String(why)})
	}
	const why = "The sentence said floating point where the arithmetic is exact."
	doc := docOf(t, FileChangelog,
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "blue-respond", &recordpb.Register{}),
		edit("G-00000001", "not floating point", "exact integers", why),
		edit("", "a second span", "its replacement", why),
	)
	mustContain(t, doc, "CHANGELOG.md",
		"## Edits to the report",
		"### `blue-respond` #1 (epoch 1)\n\n- answers **G-00000001** — "+why+"\n\n  ```diff\n  - not floating point\n  + exact integers\n  ```\n",
		"- blue's own, answering no gap — the same reason as the edit above\n\n  ```diff\n  - a second span\n  + its replacement\n  ```",
	)
}

// EVERY RECORD OF AN INQUIRY'S PATH SHIPS WITH THE REASON AND THE METHOD IT GAVE, in inquiries.md;
// the view a seat reads keeps the path's statuses and the current fate's account.
func TestTheInquiriesShowEveryRecordOfThePath(t *testing.T) {
	rec := func(seat string, st recordpb.InquiryStatus, reason, method string) *record.Event {
		return recordtest.Event(t, seat, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(st), Reason: proto.String(reason), Method: proto.String(method)})
	}
	evs := []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("Does a list of primes contain 91?")}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		rec("blue-respond", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "The first clause holds on the cited list.", "Read A000040."),
		rec("blue-respond", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "The first clause holds on the cited list.", "Read A000040 and search for a listing."),
		rec("blue-respond", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, "The second clause needs a source search a later run can make.", "Search for a source that lists 91."),
	}
	doc, seatView := inquiriesDoc(t, evs...)
	mustContain(t, doc, "inquiries.md",
		"  - path: e0 proposed -> e1 underway -> e1 underway -> e1 deferred\n"+
			"    - e0 proposed (blue-r0)\n"+
			"    - e1 underway (blue-respond) — The first clause holds on the cited list.; method: Read A000040.\n"+
			"    - e1 underway (blue-respond) — the same reason as the record before it; method: Read A000040 and search for a listing.\n",
	)
	if strings.Contains(seatView, "Read A000040.") {
		t.Errorf("the view a seat reads carries the superseded records of the path:\n%s", seatView)
	}
	if want := view.InquiryBody(nil, record.WindowIndex{}, true); want != "" {
		t.Errorf("no inquiries render %q", want)
	}
}
