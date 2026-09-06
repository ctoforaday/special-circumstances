package record

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// evidenceBoard builds a board from raw events, the way the other view tests do — the projection
// is a pure function of the event list, so the test states the events and nothing else.
func evidenceBoard(events ...*Event) *Board {
	return &Board{Events: events}
}

// evidenceEvent takes a typed BODY. It used to take alternating key/value pairs, which meant a
// fixture could state a key the verb does not write and the projection would simply never see it —
// the miss reading as "the view does not surface this". Two of the fixtures below did exactly
// that: `reference` on a verify (retired when the verify verbs stopped writing it) and `sha256` on
// a proof, whose field is `proof_sha` and is the key the reproduce join is on.
func evidenceEvent(t *testing.T, round int, seat string, body proto.Message) *Event {
	t.Helper()
	return recordtest.Event(t, seat, round, body)
}

// THE ANCHOR RESOLVES. This is the whole reason the view exists: a seat holding the c- id it read
// in the report gets the url it needs to re-fetch the source.
func TestEvidence_CitationAnchorResolvesToItsSource(t *testing.T) {
	b := evidenceBoard(evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{
		Label: proto.String("c-29a72fe2"), Url: proto.String("https://example.org/p"),
		Title: proto.String("A Paper"), Sha256: proto.String("abc123"),
		Location: proto.String("Seven is prime."), AccessDate: proto.String("2026-08-12"),
	}))

	got := EvidenceJSONOf(b.Events)
	if len(got.Sources) != 1 {
		t.Fatalf("Sources = %d, want 1", len(got.Sources))
	}
	s := got.Sources[0]
	if s.Anchor != "c-29a72fe2" {
		t.Errorf("Anchor = %q, want the c- id exactly as it appears inside <!--cite:…-->", s.Anchor)
	}
	if s.URL != "https://example.org/p" || s.Sha256 != "abc123" {
		t.Errorf("url/sha = %q/%q — the fields red re-fetches with are missing", s.URL, s.Sha256)
	}
	if s.Location != "Seven is prime." {
		t.Errorf("Location = %q, want the sentence the anchor sits at", s.Location)
	}
}

// RED'S OWN `lens cite` CARRIES NO LABEL AND IS NOT A SOURCE. The discriminator is the same one
// #341 introduced; getting it wrong here would put red's audit volume into blue's citation list.
func TestEvidence_RedLensCiteIsNotASource(t *testing.T) {
	b := evidenceBoard(
		evidenceEvent(t, 1, "lens-r1", &recordpb.Cite{Text: proto.String("checked something")}),            // no label: red's
		evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("u")}), // blue's
	)
	got := EvidenceJSONOf(b.Events)
	if len(got.Sources) != 1 || got.Sources[0].Anchor != "c-1" {
		t.Fatalf("Sources = %+v, want only the labelled blue cite", got.Sources)
	}
}

// A PROOF CARRIES BOTH IDS. The seat reads `p-…` in the report; `lens reproduce --id` takes the
// sha256. The gap between those two is what made the verb unreachable from the document.
func TestEvidence_ProofCarriesAnchorAndTheShaReproduceTakes(t *testing.T) {
	b := evidenceBoard(evidenceEvent(t, 1, "blue-r1", &recordpb.Proof{
		ProofId: proto.String("p-deadbeef"), ProofSha: proto.String("f00d"),
		ProofBasis: proto.String("reproducible"), Script: proto.String("proofs/trial.py"),
		// exit 0 is the SUCCESS status, and it is set explicitly: an implicit-presence encoding
		// would drop exactly the case that says the proof worked.
		Exit: proto.Int32(0),
	}))

	got := EvidenceJSONOf(b.Events)
	if len(got.Proofs) != 1 {
		t.Fatalf("Proofs = %d, want 1", len(got.Proofs))
	}
	p := got.Proofs[0]
	if p.Anchor != "p-deadbeef" || p.Sha256 != "f00d" {
		t.Errorf("anchor/sha = %q/%q — a seat holding the anchor cannot reach `reproduce --id`", p.Anchor, p.Sha256)
	}
	if p.Verified != nil {
		t.Errorf("Verified = %+v, want nil — nobody re-ran it", p.Verified)
	}
	if got.Counts.ProofsUnverified != 1 {
		t.Errorf("ProofsUnverified = %d, want 1", got.Counts.ProofsUnverified)
	}
}

// UNVERIFIED IS A STATED FIELD, NOT AN OMISSION — and it must survive the JSON, because a reader
// that cannot tell "nobody re-ran it" from "it reproduced" rates an unaudited proof as an audited
// one. `omitempty` on that field would erase exactly the distinction.
func TestEvidence_UnverifiedProofSaysSoInTheJSON(t *testing.T) {
	b := evidenceBoard(evidenceEvent(t, 1, "blue-r1", &recordpb.Proof{
		ProofId: proto.String("p-1"), ProofSha: proto.String("s1"),
	}))
	out, err := json.Marshal(EvidenceJSONOf(b.Events))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"verified":null`) {
		t.Errorf("the JSON omits the verified field on an un-re-run proof:\n%s\n\nAn absent field reads as a clean proof; that is the plausible zero this view exists to remove", out)
	}
}

// THE ONE JOIN THE RECORD SUPPORTS. `reproduce` carries proof_sha, so red's re-run attaches to
// the proof it checked — and the two axes stay apart, because a script that prints its conclusion
// reproduces forever.
func TestEvidence_ReproduceJoinsItsProofAndKeepsTheAxesApart(t *testing.T) {
	b := evidenceBoard(
		evidenceEvent(t, 1, "blue-r1", &recordpb.Proof{ProofId: proto.String("p-1"), ProofSha: proto.String("s1")}),
		evidenceEvent(t, 2, "lens-r2", &recordpb.Reproduce{
			ProofSha: proto.String("s1"), Reproduced: proto.Bool(true),
			Soundness: recordtest.P(recordpb.Soundness_SOUNDNESS_UNSOUND),
			Note:      proto.String("it prints its conclusion"),
		}),
	)
	got := EvidenceJSONOf(b.Events)
	v := got.Proofs[0].Verified
	if v == nil {
		t.Fatal("Verified = nil — the proof_sha join did not happen")
	}
	if !v.Reproduced {
		t.Error("Reproduced = false, want true (the tool computed it)")
	}
	if v.Sound {
		t.Error("Sound = true on a `--as unsound` re-run — the mechanical half was allowed to stand in for the judgement")
	}
	if got.Counts.ProofsUnverified != 0 {
		t.Errorf("ProofsUnverified = %d, want 0", got.Counts.ProofsUnverified)
	}
}

// A VERIFICATION ATTACHES TO THE CITATION IT NAMES. This is the join #382 made possible: red's
// verdict travels with the source it is about, so "has anyone checked this?" is a field.
func TestEvidence_VerificationAttachesToItsCitation(t *testing.T) {
	b := evidenceBoard(
		evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("https://example.org/p")}),
		evidenceEvent(t, 1, "lens-r1", &recordpb.Verify{
			Anchor: proto.String("c-1"), Claim: proto.String("seven is prime"),
			Url:        proto.String("example.org/p"),
			Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
			Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
			Text:       proto.String("the abstract says it"),
		}),
	)
	got := EvidenceJSONOf(b.Events)
	if len(got.Sources) != 1 || len(got.Sources[0].Verified) != 1 {
		t.Fatalf("Sources = %+v, want the citation carrying its one verification", got.Sources)
	}
	if v := got.Sources[0].Verified[0]; v.Outcome != "supports" || v.Text == "" {
		t.Errorf("verification = %+v, want the outcome AND the reading behind it", v)
	}
	if got.Counts.SourcesUnverified != 0 || got.Counts.Verifications != 1 {
		t.Errorf("counts: %d unverified / %d verifications, want 0/1",
			got.Counts.SourcesUnverified, got.Counts.Verifications)
	}
}

// AN INDEPENDENT CHECK HAS NO ANCHOR AND IS NOT A MISSING ONE. Corroboration red found itself
// is a different fact from an unverified citation, so it lands in its own array.
func TestEvidence_IndependentChecksStandApart(t *testing.T) {
	b := evidenceBoard(
		evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("https://example.org/p")}),
		evidenceEvent(t, 1, "lens-r1", &recordpb.Verify{
			Claim: proto.String("seven is prime"), Url: proto.String("a textbook I found"),
			Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
			Text:    proto.String("chapter 2"),
		}),
	)
	got := EvidenceJSONOf(b.Events)
	if len(got.Independent) != 1 {
		t.Fatalf("Independent = %+v, want red's anchorless check", got.Independent)
	}
	if len(got.Sources[0].Verified) != 0 || got.Counts.SourcesUnverified != 1 {
		t.Errorf("an independent check was credited to the citation — the cited source is still unchecked")
	}
}

// UNCHECKED IS AN EMPTY ARRAY IN THE JSON, not an absent key: "nobody has verified this source"
// is what red reads to decide where its next pass goes.
func TestEvidence_UncheckedSourceSaysSoInTheJSON(t *testing.T) {
	b := evidenceBoard(evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("u")}))
	out, err := json.Marshal(EvidenceJSONOf(b.Events))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"verified":[]`) {
		t.Errorf("an unverified source omits the verified key:\n%s", out)
	}
}

// REFUTES AND ABSENT ARE THE FINDING HALF, and they are counted — that count is what the
// assembly screen acts on, and what the old high|medium|low enum could never produce.
func TestEvidence_RefutedCitationsAreCounted(t *testing.T) {
	for _, outcome := range []recordpb.SourceOutcome{
		recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES,
		recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT,
	} {
		b := evidenceBoard(
			evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("u")}),
			evidenceEvent(t, 2, "lens-r2", &recordpb.Verify{
				Anchor: proto.String("c-1"), Claim: proto.String("x"),
				Outcome: outcome.Enum(), Text: proto.String("read it"),
			}),
		)
		got := EvidenceJSONOf(b.Events)
		if got.Counts.SourcesRefuted != 1 {
			t.Errorf("outcome %q: SourcesRefuted = %d, want 1 — a source red found against must be countable",
				recordpb.Word(outcome), got.Counts.SourcesRefuted)
		}
		if !got.Sources[0].Verified[0].Refuted() {
			t.Errorf("outcome %q: Refuted() = false; the screen and the counts must agree on what 'found against' means", outcome)
		}
	}
	// And the supporting half is NOT counted as found-against.
	b := evidenceBoard(
		evidenceEvent(t, 1, "blue-r1", &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("u")}),
		evidenceEvent(t, 2, "lens-r2", &recordpb.Verify{
			Anchor: proto.String("c-1"), Claim: proto.String("x"),
			Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK), Text: proto.String("thin"),
		}),
	)
	if got := EvidenceJSONOf(b.Events); got.Counts.SourcesRefuted != 0 {
		t.Errorf("`weak` counted as refuted — thin support is not contradiction, and conflating them turns a grading nuance into an assembly failure")
	}
}

// EMPTY IS EMPTY ARRAYS, NOT NULLS. A seat reading `"sources": null` has to know that means the
// same as `[]`; the arrays are initialised so a fresh run answers in the shape a full one does.
func TestEvidence_EmptyRunRendersArraysNotNulls(t *testing.T) {
	out, err := json.Marshal(EvidenceJSONOf(evidenceBoard().Events))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"sources":[]`, `"proofs":[]`, `"independent":[]`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("empty evidence view is missing %s:\n%s", want, out)
		}
	}
}
