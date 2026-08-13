package record

import (
	"encoding/json"
	"strings"
	"testing"
)

// evidenceBoard builds a board from raw events, the way the other view tests do — the projection
// is a pure function of the event list, so the test states the events and nothing else.
func evidenceBoard(events ...Event) *Board {
	return &Board{Events: events}
}

func evidenceEvent(t string, round int, seat string, kv ...any) Event {
	p := NewPayload()
	for i := 0; i+1 < len(kv); i += 2 {
		p.Set(kv[i].(string), kv[i+1])
	}
	return Event{Type: t, Round: round, SeatID: seat, Payload: p}
}

// THE ANCHOR RESOLVES. This is the whole reason the view exists: a seat holding the c- id it read
// in the report gets the url it needs to re-fetch the source.
func TestEvidence_CitationAnchorResolvesToItsSource(t *testing.T) {
	b := evidenceBoard(evidenceEvent("cite", 1, "blue-r1",
		"label", "c-29a72fe2", "url", "https://example.org/p", "title", "A Paper",
		"sha256", "abc123", "location", "Seven is prime.", "access_date", "2026-08-12"))

	got := EvidenceJSONOf(b)
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
		evidenceEvent("cite", 1, "lens-r1", "claim", "checked something"), // no label: red's
		evidenceEvent("cite", 1, "blue-r1", "label", "c-1", "url", "u"),   // blue's
	)
	got := EvidenceJSONOf(b)
	if len(got.Sources) != 1 || got.Sources[0].Anchor != "c-1" {
		t.Fatalf("Sources = %+v, want only the labelled blue cite", got.Sources)
	}
}

// A PROOF CARRIES BOTH IDS. The seat reads `p-…` in the report; `lens reproduce --id` takes the
// sha256. The gap between those two is what made the verb unreachable from the document.
func TestEvidence_ProofCarriesAnchorAndTheShaReproduceTakes(t *testing.T) {
	b := evidenceBoard(evidenceEvent("proof", 1, "blue-r1",
		"proof_id", "p-deadbeef", "sha256", "f00d", "proof_basis", "reproducible",
		"script", "proofs/trial.py", "exit", 0, "location", "Seven is prime."))

	got := EvidenceJSONOf(b)
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
	b := evidenceBoard(evidenceEvent("proof", 1, "blue-r1", "proof_id", "p-1", "sha256", "s1"))
	out, err := json.Marshal(EvidenceJSONOf(b))
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
		evidenceEvent("proof", 1, "blue-r1", "proof_id", "p-1", "sha256", "s1"),
		evidenceEvent("reproduce", 2, "lens-r2", "proof_sha", "s1", "reproduced", true,
			"soundness", "unsound", "note", "it prints its conclusion"),
	)
	got := EvidenceJSONOf(b)
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
		evidenceEvent("cite", 1, "blue-r1", "label", "c-1", "url", "https://example.org/p"),
		evidenceEvent("verify", 1, "lens-r1", "anchor", "c-1", "claim", "seven is prime",
			"reference", "example.org/p", "outcome", "supports", "text", "the abstract says it"),
	)
	got := EvidenceJSONOf(b)
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
		evidenceEvent("cite", 1, "blue-r1", "label", "c-1", "url", "https://example.org/p"),
		evidenceEvent("verify", 1, "lens-r1", "claim", "seven is prime",
			"reference", "a textbook I found", "outcome", "supports", "text", "chapter 2"),
	)
	got := EvidenceJSONOf(b)
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
	b := evidenceBoard(evidenceEvent("cite", 1, "blue-r1", "label", "c-1", "url", "u"))
	out, err := json.Marshal(EvidenceJSONOf(b))
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
	for _, outcome := range []string{"refutes", "absent"} {
		b := evidenceBoard(
			evidenceEvent("cite", 1, "blue-r1", "label", "c-1", "url", "u"),
			evidenceEvent("verify", 2, "lens-r2", "anchor", "c-1", "claim", "x",
				"outcome", outcome, "text", "read it"),
		)
		got := EvidenceJSONOf(b)
		if got.Counts.SourcesRefuted != 1 {
			t.Errorf("outcome %q: SourcesRefuted = %d, want 1 — a source red found against must be countable",
				outcome, got.Counts.SourcesRefuted)
		}
		if !got.Sources[0].Verified[0].Refuted() {
			t.Errorf("outcome %q: Refuted() = false; the screen and the counts must agree on what 'found against' means", outcome)
		}
	}
	// And the supporting half is NOT counted as found-against.
	b := evidenceBoard(
		evidenceEvent("cite", 1, "blue-r1", "label", "c-1", "url", "u"),
		evidenceEvent("verify", 2, "lens-r2", "anchor", "c-1", "claim", "x", "outcome", "weak", "text", "thin"),
	)
	if got := EvidenceJSONOf(b); got.Counts.SourcesRefuted != 0 {
		t.Errorf("`weak` counted as refuted — thin support is not contradiction, and conflating them turns a grading nuance into an assembly failure")
	}
}

// EMPTY IS EMPTY ARRAYS, NOT NULLS. A seat reading `"sources": null` has to know that means the
// same as `[]`; the arrays are initialised so a fresh run answers in the shape a full one does.
func TestEvidence_EmptyRunRendersArraysNotNulls(t *testing.T) {
	out, err := json.Marshal(EvidenceJSONOf(evidenceBoard()))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"sources":[]`, `"proofs":[]`, `"independent":[]`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("empty evidence view is missing %s:\n%s", want, out)
		}
	}
}
