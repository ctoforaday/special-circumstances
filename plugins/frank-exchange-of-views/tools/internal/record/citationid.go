package record

import (
	"crypto/rand"
	"encoding/hex"

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

// CitedSources returns the distinct cited sources — one per blue `cite` label, in
// first-seen (label) order. It is the composer input for the assembled ## Bibliography and,
// via its labels, the EXPECTED set the unbacked_citations detector checks. Red's `lens cite`
// carries no label and is excluded, so this is exactly the tool-inserted citation set.
func CitedSources(runDir string) ([]Source, error) {
	m, err := MergedEvents(runDir)
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
		c, isCite := body.(*recordpb.Cite)
		if !isCite {
			continue
		}
		// The empty-label filter is UNCHANGED and still load-bearing. Under the old shape it
		// was the discriminator between blue's authored cite and red's `lens cite` (see the
		// note below); red now writes a `verify` event, which the type switch already excludes.
		// It stays because a cite with no label names no anchor, and CitationLabels — the
		// EXPECTED set the lockdown compares against the report — must not gain a phantom.
		label := c.GetLabel()
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, Source{
			Label:      label,
			URL:        c.GetUrl(),
			Sha256:     c.GetSha256(),
			Title:      c.GetTitle(),
			AccessDate: c.GetAccessDate(),
			Location:   c.GetLocation(),
		})
	}
	return out, nil
}

// CitationLabels returns the distinct citation labels EXPECTED in blue/report.md — the
// label of every blue `cite` event on the record, in first-seen order. It is the citation
// twin of AnchorIDs: the blue-report lockdown's PostToolUse backstop compares this to the
// anchors actually present to catch a dropped citation, and it is the EXPECTED set behind
// the unbacked_citations detector. Only blue cites carry a `label`; red's `lens cite` does
// not, so this is exactly the set of tool-inserted citation anchors that must be present.
func CitationLabels(runDir string) ([]string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for i := range m.Events {
		body, ok := recordpb.Body(m.Events[i])
		if !ok {
			continue
		}
		if c, isCite := body.(*recordpb.Cite); isCite {
			if id := c.GetLabel(); id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out, nil
}

// ExistingCiteByKey returns the label of a prior blue cite this seat recorded under the
// same --key, so a crash-retried `blue cite` returns its existing anchor label instead of
// minting a duplicate. Mirrors ExistingFindingByKey: the retry dedup is a short-circuit
// BEFORE the fetch and the marker insert, not a change to the event key (which stays the
// unique citation label). A blue cite carries a `label`; red's `lens cite` does not, so
// this scans only the blue side of the shared "cite" event type.
func ExistingCiteByKey(runDir, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for i := range m.Events {
		e := m.Events[i]
		if e.GetSeatId() != seatID {
			continue
		}
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		// key is non-empty (guarded above), so GetCiteKey()'s zero value for an ABSENT
		// cite_key can never match it — the old `Payload.Str("cite_key") == key` had the same
		// property for the same reason.
		if c, isCite := body.(*recordpb.Cite); isCite && c.GetCiteKey() == key {
			return c.GetLabel(), nil
		}
	}
	return "", nil
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
func ExistingProofByKey(runDir, seatID, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for i := range m.Events {
		e := m.Events[i]
		if e.GetSeatId() != seatID {
			continue
		}
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		// The proof's sha was written under the payload key `sha256` and is READ BACK under
		// `proof_sha` by every joiner — reproduce writes `proof_sha` for the same value, and
		// `lens reproduce --id` takes it. The schema carries the one fact once, on
		// Proof.proof_sha; this is that same value, not a rename of a different field.
		if p, isProof := body.(*recordpb.Proof); isProof && p.GetProofKey() == key {
			return p.GetProofSha(), nil
		}
	}
	return "", nil
}

// Proof is one recorded computation, drawn from a blue `prove` event — the composer input
// for the assembled report's Proofs section and the handle red re-runs.
type Proof struct {
	Label  string // the p-<hex> anchor id
	SHA    string
	Basis  string
	Cites  string // the METHOD citation this applies, when blue named one
	Reason string

	// Drift is WHETHER the two runs diverged, not the sentence describing how.
	//
	// The record carries five fields for a proof — answers, cites, drift, proof_key, text —
	// measured by reading the verb, and `script`, `exit`, `location` and `output` were excluded
	// deliberately. The script BODY and its exit code are artifacts, not facts about the debate:
	// they live in the proof cache, keyed by the sha256 this struct carries. A reader that wants
	// them joins on SHA rather than finding them copied into every event.
	Drift bool

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
func RecordedProofs(runDir string) ([]Proof, error) {
	m, err := MergedEvents(runDir)
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
