package record

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// CITATION IDENTITY MIRRORS FINDING IDENTITY — see findingid.go for the full argument.
//
// A citation, like a finding, is a TOOL-INSERTED invisible anchor (<!--cite:c-<hex>-->)
// that the seat never hand-writes and cannot invent. The id is random for the same
// reason a finding id is: an id you cannot guess is one you have to LOOK UP, so "which
// citation do you mean" is a read, not a memory exercise. Under the lockdown the anchor
// is immortal, so the set of cite events is a strict bijection with the anchors in the
// document — the id is the identity that pins both halves together.

// NewCitationID mints an unguessable citation id.
//
// The "c-" prefix keeps it legible in a transcript and distinguishes it at a glance from
// a finding's "f-" id — the two anchor classes share the lockdown's protection but are
// never confused in an error message.
func NewCitationID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return "c-" + hex.EncodeToString(b)
}

// Source is one cited source, drawn from a blue `cite` event — the composer input for the
// assembled bibliography and the hash a reader verifies the cached bytes against.
type Source struct {
	Label      string
	URL        string
	Sha256     string
	Title      string
	AccessDate string
	Location   string
}

// citedSource reads a citable source off an event body, if it is one.
//
// TWO EVENT TYPES CARRY A CITATION NOW, and that is the change. Blue's `cite` attaches a source
// to a claim it is AUTHORING; red's `corroborate` records a source it FOUND, for a claim blue
// made. #341 split them deliberately and this does not undo that — they stay separate acts with
// separate verbs and separate readers. What they share is a minted label, and a label is what
// makes a footnote.
//
// The comments here used to state red's exclusion as a property of the system: "Red's `lens cite`
// carries no label and is EXCLUDED, so this is exactly the tool-inserted citation set." That was
// accurate and it was the defect — red's independent corroboration reached no reader of the
// document, and the strongest thing an adversarial process produces (a claim confirmed by a
// source its author never chose) died in a projection. A human reader cares that the text has
// appropriate references, not which team inserted them.
//
// ONLY A SUPPORTING CORROBORATION CARRIES A LABEL. The verb withholds it for `refutes`, `absent`
// and `weak`: a source that CONTRADICTS the sentence is not a reference backing it, and rendered
// in the bibliography it would read as support. Those go to the board instead.
func citedSource(body proto.Message) (Source, bool) {
	switch b := body.(type) {
	case *recordpb.Cite:
		return Source{
			Label:      b.GetLabel(),
			URL:        b.GetUrl(),
			Sha256:     b.GetSha256(),
			Title:      b.GetTitle(),
			AccessDate: b.GetAccessDate(),
			Location:   b.GetLocation(),
		}, true
	case *recordpb.Verify:
		// No sha: red read the source itself rather than through the run cache, and the
		// bibliography renders title + url + access date, never the hash. `location` is blue's
		// placement field; a corroboration's span is its `claim`.
		return Source{
			Label:      b.GetLabel(),
			URL:        b.GetUrl(),
			Title:      b.GetTitle(),
			AccessDate: b.GetAccessDate(),
			Location:   b.GetClaim(),
		}, true
	}
	return Source{}, false
}

// CitedSources returns the distinct cited sources — one per blue `cite` label, in
// first-seen (label) order. It is the composer input for the assembled ## Bibliography and,
// via its labels, the EXPECTED set the unbacked_citations detector checks.
//
// IT IS BOTH TEAMS' CITATIONS NOW. This said "Red's `lens cite` carries no label and is excluded,
// so this is exactly the tool-inserted citation set" — the second half is still true and the
// first is the thing that changed: a supporting corroboration mints a label and splices an
// anchor, so it MUST be in this set or the blue-report lockdown reads red's anchor as a citation
// blue dropped. See citedSource for why red's evidence stopped being excluded.
func CitedSources(run Run) ([]Source, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []Source
	for i := range m.Events {
		body, ok := recordpb.Body(m.Events[i])
		if !ok {
			// No body at all, so nothing to cite. The old code reached the same answer by a
			// different route: Payload.Str on an absent key returned "", which the label
			// filter below dropped.
			continue
		}
		src, ok := citedSource(body)
		if !ok {
			continue
		}
		// The empty-label filter is UNCHANGED and still load-bearing: an event with no label
		// names no anchor, and CitationLabels — the EXPECTED set the lockdown compares against
		// the report — must not gain a phantom. What changed is WHICH events can carry one.
		if src.Label == "" || seen[src.Label] {
			continue
		}
		seen[src.Label] = true
		out = append(out, src)
	}
	return out, nil
}

// CitationLabels returns the distinct citation labels EXPECTED in blue/report.md — the
// label of every blue `cite` event on the record, in first-seen order. It is the citation
// twin of AnchorIDs: the blue-report lockdown's PostToolUse backstop compares this to the
// anchors actually present to catch a dropped citation, and it is the EXPECTED set behind
// the unbacked_citations detector. Only blue cites carry a `label`; red's `lens cite` does
// not, so this is exactly the set of tool-inserted citation anchors that must be present.
func CitationLabels(run Run) ([]string, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	return citationLabelsOf(m.Events), nil
}

// CitationLabelsOf is CitationLabels over events already read — for a caller holding a board,
// which would otherwise re-derive the set inline and become a second copy of the rule.
//
// It exists because one already had. internal/scorecard built the unbacked_citations detector's
// EXPECTED set with its own loop over `Cite` events, so widening this function to include red's
// labelled corroborations left that detector on the old rule: blue dropping a RED citation
// anchor would be caught by the hookgate lockdown and MISSED by the scorecard, two detectors
// disagreeing about one protection. That is the duplication this schema exists to remove.
func CitationLabelsOf(events []*Event) []string { return citationLabelsOf(events) }

func citationLabelsOf(events []*Event) []string {
	seen := map[string]bool{}
	var out []string
	for i := range events {
		body, ok := recordpb.Body(events[i])
		if !ok {
			continue
		}
		if src, ok := citedSource(body); ok && src.Label != "" && !seen[src.Label] {
			seen[src.Label] = true
			out = append(out, src.Label)
		}
	}
	return out
}

// ExistingCiteByKey returns the label of a prior blue cite this seat recorded under the
// same --key, so a crash-retried `blue cite` returns its existing anchor label instead of
// minting a duplicate. Mirrors existingFindingByKey: the retry dedup is a short-circuit
// BEFORE the fetch and the marker insert, not a change to the event key (which stays the
// unique citation label). A blue cite carries a `label`; red's `lens cite` does not, so
// this scans only the blue side of the shared "cite" event type.
func ExistingCiteByKey(run Run, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	// key is non-empty (guarded above), so a NULL cite_key can never match it — the same
	// property the fold's absent-key zero value had, held by the comparison itself.
	var label sql.NullString
	if _, err := queryRow(run, []any{&label},
		`SELECT c."label" FROM "cite" c JOIN "events" e ON e."id" = c."event_id"
		  WHERE e."seat_id" = ? AND c."cite_key" = ? ORDER BY c."event_id" LIMIT 1`,
		seatID, key); err != nil {
		return "", err
	}
	return label.String, nil
}

// TWO KINDS OF `cite` EVENT SHARE ONE TYPE — and conflating them inflates red's audit metric.
//
// `lens cite` records RED VERIFYING a source (a citation-ledger row: claim / reference /
// confidence / access_date). `blue cite` records BLUE AUTHORING one (label / url / sha256 /
// title / location, plus an immortal anchor in the report). Only blue's carries `label`, which
// is what distinguishes them — the same discriminator CitedSources and CitationLabels use.
//
// Counting both as "citations" makes an AUDIT-VOLUME number grow when BLUE writes, which is the
// one thing the engine most needs to keep apart. Measured on the 2026-08-04 smoke: the board
// reported 10 where the truth was 3 authored + 7 verified — 43% inflation on a tile labelled
// "citations checked" on RED's board, and debate.js tells red to copy that figure into its own
// envelope. The error grows with every use of the citation axis, so it worsens precisely as
// #256 is adopted.
//
// This is the FIRST of a family: a single event type carrying two provenances that a reader must
// not confuse. Red's coming `fix_basis: verified | proposed` (does a required_fix name a remedy
// red CHECKED, or one it merely proposes?) is the same shape and belongs beside this — a
// discriminator on the record, read at the counting site, never a rendering-time guess.

// A `cite` event is BLUE authoring a citation; a `verify` event is RED recording a checked
// source. Two event types, because the distinction used to be INFERRED (#341):
//
//	func IsVerifiedCite(e Event) bool { return e.Type == "cite" && e.Payload.Str("label") == "" }
//
// The absence of a field decided which act an event was, so a blue cite written without a label
// counted as red's audit volume — a number red reads as how much work it did — with no error
// and no signal. Both helpers are deleted; readers switch on the type.

// NewProofID mints a proof anchor id. Same shape as a citation's, different class prefix:
// one immortal-anchor mechanism carrying three classes now (fx: a finding, cite: a source,
// proof: a computation).
func NewProofID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic("record: entropy unavailable: " + err.Error())
	}
	return "p-" + hex.EncodeToString(b)
}

// ExistingProofByKey gives `blue prove` crash-retry idempotency: a seat whose message died
// after the event landed re-runs the same key and gets the recorded sha back rather than
// executing the script a second time and splicing a second anchor.
func ExistingProofByKey(run Run, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	// The schema carries the one fact once, on Proof.proof_sha; reproduce writes the same value
	// and every joiner reads it back under this name.
	var sha sql.NullString
	if _, err := queryRow(run, []any{&sha},
		`SELECT p."proof_sha" FROM "proof" p JOIN "events" e ON e."id" = p."event_id"
		  WHERE e."seat_id" = ? AND p."proof_key" = ? ORDER BY p."event_id" LIMIT 1`,
		seatID, key); err != nil {
		return "", err
	}
	return sha.String, nil
}

// Proof is one recorded computation, drawn from a blue `prove` event — the composer input
// for the assembled report's Proofs section and the handle red re-runs.
type Proof struct {
	Label  string // the p-<hex> anchor id
	SHA    string
	Basis  string
	Script string
	Exit   int
	Cites  string // the METHOD citation this applies, when blue named one
	Reason string

	// Drift is WHAT MOVED between the two runs — the sentence, not a bool. It lived in the proof
	// cache's meta.json while the report rendered it, so the record was the one party to the
	// exchange that could not say what happened. That file is gone and this is a field.
	Drift string

	// Verified is red's independent re-run (#343): whether the proof reproduced for the
	// auditor, and what red made of it. Nil when nobody re-ran it — and that absence is
	// itself information the reader needs, so the report says so rather than omitting it.
	Verified *ProofVerification
}

// ProofVerification is one `reproduce` event: red re-ran the script and compared bytes.
// `Reproduced` is COMPUTED by the tool, not claimed by the seat — which is the whole reason
// re-running beats re-reading.
type ProofVerification struct {
	SeatID     string
	Round      int
	Reproduced bool
	// Sound is red's JUDGEMENT, made by reading the script: does it establish the claim it is
	// anchored to? Reproducing measures determinism only.
	Sound    bool
	Note     string
	Recorded string // only on a mismatch
	Observed string
}

// RecordedProofs returns every proof on the record, in event order.
//
// The OUTPUT and the SCRIPT BODY are deliberately not here: they live on disk under
// <run>/proofs/<sha256>/ and can be large. The assembler reads them from the artifact so the
// report shows the exact bytes that ran, rather than a copy the record made of them.
func RecordedProofs(run Run) ([]Proof, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	// Red's re-runs, keyed by the proof they checked, so the join happens once here rather
	// than in every reader.
	verified := map[string]*ProofVerification{}
	for i := range m.Events {
		e := m.Events[i]
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		r, isReproduce := body.(*recordpb.Reproduce)
		if !isReproduce {
			continue
		}
		verified[r.GetProofSha()] = &ProofVerification{
			SeatID: e.GetSeatId(), Round: int(e.GetRound()),
			// `reproduced` is COMPUTED by the tool and always written; an absent one reads
			// false, which is what the old bool-assertion default did with a missing or
			// wrongly-typed key. ProofVerification.Reproduced is a plain bool and cannot
			// carry the absence, so the presence distinction stops here as it did before.
			Reproduced: r.GetReproduced(),
			// The seat types --as sound|unsound; the string compare against "sound" becomes
			// the enum it was standing in for. UNSPECIFIED (absent) is not sound, exactly as
			// an absent `soundness` key was not.
			Sound: r.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND,
			// --reason lands on Reproduce.note. The flag and the field are two vocabularies
			// (recordpb/required.go says so outright: a close stores `prose`, an opinion
			// `rationale`, and the flag for both is --reason); `note` is the only prose
			// channel this message has.
			Note:     r.GetNote(),
			Recorded: r.GetRecordedOutput(), Observed: r.GetObservedOutput(),
		}
	}
	// THREE FIELDS OF THIS PROJECTION HAVE NO FIELD ON `recordpb.Proof`, and they are left
	// UNCONVERTED rather than defaulted — see the report to the lead. `blue prove` writes
	// `script`, `exit` and `drift` onto the proof event (cli/blue/prove.go:85,86,90) and
	// report/proofs.go renders all three: the script's path names the code fence's language,
	// the exit code is a line of the Proofs section, and `drift` is describeDrift's SENTENCE
	// ("exit code moved 2 -> 3"), which the schema's `optional bool drift` cannot carry.
	// Zero-filling them would render a drifting proof as a clean one, in a section whose whole
	// subject is whether the computation held — the plausible zero this migration exists to
	// remove. The schema decision is the lead's; the compile error is the loud miss.
	var out []Proof
	for i := range m.Events {
		e := m.Events[i]
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		pf, isProof := body.(*recordpb.Proof)
		if !isProof {
			continue
		}
		out = append(out, Proof{
			Label: pf.GetProofId(),
			// Written as `sha256`, joined on as `proof_sha` — one fact, one field now.
			SHA:    pf.GetProofSha(),
			Basis:  pf.GetProofBasis(),
			Script: pf.GetScript(),
			Exit:   int(pf.GetExit()),
			Cites:  pf.GetCites(),
			// --reason lands on Proof.text, the message's only prose channel — the same
			// flag/field split recordpb/required.go records for Verify.text.
			Reason: pf.GetText(),
			Drift:  pf.GetDrift(),
			// nil when nobody re-ran it, which the report states rather than omits.
			Verified: verified[pf.GetProofSha()],
		})
	}
	return out, nil
}

// ExistingCorroborationLabel returns the label this seat already minted for the same source and
// the same claim, so a crash-retried `lens corroborate` returns its anchor instead of splicing a
// second one.
//
// IT REPLACES AN IDEMPOTENCY THE LABEL TOOK AWAY. A corroboration used to key on its `url`, so a
// retry collided and was refused — the same mechanism that also capped one source at one claim
// per sitting, which is why the key moved to the minted label. But a minted label is fresh every
// call, so nothing collided any more and a retry wrote a DUPLICATE: the exact cost that made
// "drop url from keyFields" the wrong answer, arrived at by another route.
//
// Keyed on (source, claim) rather than on a seat-supplied `--key` like blue's cite, because the
// pair IS the act — corroborating one claim from one source twice is one corroboration, and a
// retry should not need the seat to have anticipated it.
func ExistingCorroborationLabel(run Run, seatID, url, claim string) (string, error) {
	if url == "" || claim == "" {
		return "", nil
	}
	var label string
	if _, err := queryRow(run, []any{&label},
		`SELECT v."label" FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
		  WHERE e."seat_id" = ? AND v."url" = ? AND v."claim" = ? AND COALESCE(v."label", '') != ''
		  ORDER BY v."event_id" LIMIT 1`,
		seatID, url, claim); err != nil {
		return "", err
	}
	return label, nil
}

// unansweredContradictions returns the claims where red read a source that CONTRADICTS the
// report and no finding was ever raised about it.
//
// THE NEGATIVE HALF OF THE CORROBORATION AXIS HAS TO REACH THE BOARD. A supporting corroboration
// becomes a footnote and reaches the reader that way; `refutes` and `absent` are not references
// backing the sentence and are deliberately NOT spliced — which leaves them landing only in the
// `evidence` projection, seen by red and by nobody else. That is the same defect one axis over:
// evidence on the record with no reader.
//
// THE TOOL DOES NOT DISCHARGE THE DUTY, and that is deliberate. Writing the finding here would
// mean inventing its severity, likelihood and impact — three grades nobody chose, feeding the
// mass calculation that decides what a gap is worth. A fabricated grade reads exactly like a
// judged one. So the duty is REPORTED and red grades its own finding, the way
// InquiryReviewDue reports a read that has not happened rather than pretending it did.
//
// The match is deliberately loose — any finding by any lens quoting the same claim answers it.
// A stricter join (same seat, same round) would refuse a contradiction one lens found and
// another raised, which is the collaboration the lens roles exist for.
func unansweredContradictions(evs []*Event) []string {
	answered := map[string]bool{}
	for _, e := range evs {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok {
			if loc := strings.TrimSpace(f.GetLocation()); loc != "" {
				answered[loc] = true
			}
		}
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range evs {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || !contradicts(v.GetOutcome()) {
			continue
		}
		claim := strings.TrimSpace(v.GetClaim())
		if claim == "" || answered[claim] || seen[claim] {
			continue
		}
		seen[claim] = true
		out = append(out, claim)
	}
	return out
}

// contradicts is the half of the outcome set that says the text is not supported. It is the
// complement of the set that earns a footnote, and the two are written as one pair so a value
// added to the enum cannot fall between them unnoticed.
func contradicts(o recordpb.SourceOutcome) bool {
	return o == recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES ||
		o == recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT
}

// reopenedAnchors returns the anchors whose text has moved since they were placed — every id any
// `blue edit` reported reopening, in first-seen order.
//
// THE REFERENCE STANDS; ITS REFERENT MOVED. That is deliberately not the same as the citation
// being wrong: a source still says what it says. What changed is the sentence it was placed
// against, so a verification of it is STALE rather than refuted, and a reader who cannot tell
// those apart will either re-check everything or trust everything.
//
// Read from the BlueEdit events rather than recomputed by diffing documents, because the
// documents are gone: `blue/report.md` holds only its current state, and the before-image an
// edit replaced exists nowhere else. The one channel that text moves through is the only place
// this fact can be captured at the moment it becomes true.
func reopenedAnchors(evs []*Event) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range evs {
		ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok {
			continue
		}
		for _, id := range ed.GetReopened() {
			if id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}
