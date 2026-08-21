package record

import (
	"crypto/rand"
	"encoding/hex"
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
	for _, e := range m.Events {
		if e.Type != "cite" {
			continue
		}
		label := e.Payload.Str("label")
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, Source{
			Label:      label,
			URL:        e.Payload.Str("url"),
			Sha256:     e.Payload.Str("sha256"),
			Title:      e.Payload.Str("title"),
			AccessDate: e.Payload.Str("access_date"),
			Location:   e.Payload.Str("location"),
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
	for _, e := range m.Events {
		if e.Type == "cite" {
			if id := e.Payload.Str("label"); id != "" && !seen[id] {
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
	result, _, err := ExistingByKey(runDir, seatID, "cite", key)
	return result, err
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
	result, _, err := ExistingByKey(runDir, seatID, "proof", key)
	return result, err
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
	Drift  string

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
	for _, e := range m.Events {
		if e.Type != "reproduce" {
			continue
		}
		rep := false
		if v, ok := e.Payload.Get("reproduced"); ok {
			if b, isBool := v.(bool); isBool {
				rep = b
			}
		}
		verified[e.Payload.Str("proof_sha")] = &ProofVerification{
			SeatID: e.SeatID, Round: e.Round, Reproduced: rep,
			Sound: e.Payload.Str("soundness") == "sound", Note: e.Payload.Str("reason"),
			Recorded: e.Payload.Str("recorded_output"), Observed: e.Payload.Str("observed_output"),
		}
	}
	var out []Proof
	for _, e := range m.Events {
		if e.Type != "proof" {
			continue
		}
		exit := 0
		if v, ok := e.Payload.Get("exit"); ok {
			if f, isNum := v.(float64); isNum {
				exit = int(f)
			} else if i, isInt := v.(int); isInt {
				exit = i
			}
		}
		out = append(out, Proof{
			Label:  e.Payload.Str("proof_id"),
			SHA:    e.Payload.Str("sha256"),
			Basis:  e.Payload.Str("proof_basis"),
			Script: e.Payload.Str("script"),
			Exit:   exit,
			Cites:  e.Payload.Str("cites"),
			Reason: e.Payload.Str("reason"),
			Drift:  e.Payload.Str("drift"),
			// nil when nobody re-ran it, which the report states rather than omits.
			Verified: verified[e.Payload.Str("sha256")],
		})
	}
	return out, nil
}
