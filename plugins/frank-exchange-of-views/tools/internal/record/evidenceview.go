package record

import (
	"encoding/json"
)

// THE EVIDENCE VIEW — what backs the report, and what has been checked of it.
//
// # The anchor a seat could not resolve
//
// The report carries three classes of tool-inserted anchor: `<!--fx:f-…-->` a finding,
// `<!--cite:c-…-->` a source, `<!--proof:p-…-->` a computation. A seat reading a sentence sees
// the token. Only the FIRST could ever be resolved — `show findings` is keyed by the f-label —
// and the other two were opaque.
//
// Measured, in a seat's own words when asked what it could do next: "blocked by missing
// information (what does `c-29a72fe2` point to?)". The tool holds that citation's url, title,
// sha256, access date and the cached bytes themselves, and handed the seat an id.
//
// It is worse than an inconvenience for RED, whose terminal duty is verification at the leaf.
// The constitution tells red to re-read the exact bytes blue cited (`fetch --url <the cited
// url>`) — with no path from the anchor in front of it to that url. And `lens reproduce --id
// <sha256>` ASSERTED the way through: "every proof is listed in the report beside the sentence
// it backs". The report lists no sha. It carries `p-<hex>`.
//
// This is the class that keeps paying out: the tool computes something, discards it, and asks
// the seat to reconstruct it.
//
// # Why one view and not two
//
// A citation and a proof are the same act at different strengths — blue offers evidence for a
// sentence, anchors it, and red checks it. Splitting them into `citations` and `proofs` would
// have grown the surface by two to answer one question ("what does this token point to"), and
// left a seat holding a `p-` id to guess which view took it. `citation-ledger` — red's
// verifications, and nothing of what blue cited — is subsumed here as the `verifications` array.
//
// # Why red's verifications are NOT joined to blue's sources
//
// They cannot be, and pretending otherwise would be the exact defect this view repairs. `lens
// verify` records a free-text `reference`, not the anchor of the citation it checked, so any
// join would be a string-similarity guess presented as a fact — and a wrong join reads
// identically to a right one. They are listed apart, and the absence of the join is stated
// rather than papered over. Filed as its own defect; the fix is a join key on `verify`, not a
// cleverer reader.
//
// PROOFS DO carry the join, because `reproduce` records `proof_sha` — so a proof's verification
// is attached to it, and a nil `verified` means nobody re-ran it. That absence is information a
// seat acts on, so it is a field rather than an omission.

// EvidenceSourceJSON is one source blue cited, keyed by the anchor token in the report.
type EvidenceSourceJSON struct {
	// Anchor is the c-<hex> exactly as it appears inside `<!--cite:…-->`. It is the lookup key
	// a seat holds when it reads the report, which is the whole reason this view exists.
	Anchor     string `json:"anchor"`
	URL        string `json:"url,omitempty"`
	Title      string `json:"title,omitempty"`
	Sha256     string `json:"sha256,omitempty"`
	AccessDate string `json:"access_date,omitempty"`
	// Location is the sentence the anchor sits at — what the source is offered as backing for.
	Location string `json:"location,omitempty"`
	Claim    string `json:"claim,omitempty"`
	SeatID   string `json:"seat_id,omitempty"`
	Round    int    `json:"round"`
}

// EvidenceProofJSON is one computation blue recorded, keyed by its anchor, WITH red's re-run.
type EvidenceProofJSON struct {
	// Anchor is the p-<hex> in the report; Sha256 is what `lens reproduce --id` takes. Both are
	// here because the seat holds the first and the verb wants the second, and that gap is what
	// made `reproduce` unreachable from the document.
	Anchor   string `json:"anchor"`
	Sha256   string `json:"sha256,omitempty"`
	Basis    string `json:"basis,omitempty"`
	Script   string `json:"script,omitempty"`
	Exit     int    `json:"exit"`
	Location string `json:"location,omitempty"`
	// Cites is the citation anchor of the METHOD this computation applies, when blue named one —
	// the link between the two evidence classes, and the only one on the record.
	Cites  string `json:"cites,omitempty"`
	Drift  string `json:"drift,omitempty"`
	SeatID string `json:"seat_id,omitempty"`
	Round  int    `json:"round"`

	// Verified is red's re-run, or nil when nobody re-ran it. Nil is not "failed" and not
	// "clean": it is unchecked, and a reader that could not tell those apart would rate an
	// unaudited proof as an audited one.
	Verified *EvidenceReproductionJSON `json:"verified"`
}

// EvidenceReproductionJSON is one `reproduce` event — the two axes kept apart.
type EvidenceReproductionJSON struct {
	// Reproduced is COMPUTED by the tool (bytes re-run and compared). Sound is RED'S JUDGEMENT
	// from reading the script. A script that prints its conclusion reproduces forever, so the
	// mechanical half alone establishes nothing and the two are never collapsed into one word.
	Reproduced bool   `json:"reproduced"`
	Sound      bool   `json:"sound"`
	Note       string `json:"note,omitempty"`
	SeatID     string `json:"seat_id,omitempty"`
	Round      int    `json:"round"`
}

// EvidenceVerificationJSON is one `lens verify` — a claim red checked against a source.
type EvidenceVerificationJSON struct {
	Claim string `json:"claim,omitempty"`
	// Reference is FREE TEXT, which is why nothing here is joined to a citation anchor.
	Reference  string `json:"reference,omitempty"`
	Trust      string `json:"trust,omitempty"`
	AccessDate string `json:"access_date,omitempty"`
	SeatID     string `json:"seat_id,omitempty"`
	Round      int    `json:"round"`
}

// EvidenceJSON is the whole evidence layer of the report.
type EvidenceJSON struct {
	Sources       []EvidenceSourceJSON       `json:"sources"`
	Proofs        []EvidenceProofJSON        `json:"proofs"`
	Verifications []EvidenceVerificationJSON `json:"verifications"`
	Counts        struct {
		Sources int `json:"sources"`
		Proofs  int `json:"proofs"`
		// ProofsUnverified is the count nobody re-ran. It is stated because the honest zero and
		// the unchecked case are otherwise the same empty space.
		ProofsUnverified int `json:"proofs_unverified"`
		Verifications    int `json:"verifications"`
	} `json:"counts"`
}

// EvidenceJSONOf projects the record's evidence layer in event order.
func EvidenceJSONOf(b *Board) EvidenceJSON {
	out := EvidenceJSON{
		Sources:       []EvidenceSourceJSON{},
		Proofs:        []EvidenceProofJSON{},
		Verifications: []EvidenceVerificationJSON{},
	}

	// Red's re-runs, keyed by the proof sha they checked — the one join the record supports.
	reruns := map[string]*EvidenceReproductionJSON{}
	for _, e := range b.Events {
		if e.Type != "reproduce" {
			continue
		}
		rep := false
		if v, ok := e.Payload.Get("reproduced"); ok {
			if got, isBool := v.(bool); isBool {
				rep = got
			}
		}
		reruns[e.Payload.Str("proof_sha")] = &EvidenceReproductionJSON{
			Reproduced: rep,
			Sound:      e.Payload.Str("soundness") == "sound",
			Note:       e.Payload.Str("note"),
			SeatID:     e.SeatID,
			Round:      e.Round,
		}
	}

	for _, e := range b.Events {
		switch e.Type {
		case "cite":
			// A blue cite carries a `label`; red's `lens cite` does not. That discriminator is
			// the same one CitedSources and CitationLabels use — see the #341 note in
			// citationid.go for what conflating the two costs.
			label := e.Payload.Str("label")
			if label == "" {
				continue
			}
			out.Sources = append(out.Sources, EvidenceSourceJSON{
				Anchor:     label,
				URL:        e.Payload.Str("url"),
				Title:      e.Payload.Str("title"),
				Sha256:     e.Payload.Str("sha256"),
				AccessDate: e.Payload.Str("access_date"),
				Location:   e.Payload.Str("location"),
				Claim:      e.Payload.Str("claim"),
				SeatID:     e.SeatID,
				Round:      e.Round,
			})
		case "proof":
			exit := 0
			if v, ok := e.Payload.Get("exit"); ok {
				switch n := v.(type) {
				case float64:
					exit = int(n)
				case int:
					exit = n
				}
			}
			sha := e.Payload.Str("sha256")
			out.Proofs = append(out.Proofs, EvidenceProofJSON{
				Anchor:   e.Payload.Str("proof_id"),
				Sha256:   sha,
				Basis:    e.Payload.Str("proof_basis"),
				Script:   e.Payload.Str("script"),
				Exit:     exit,
				Location: e.Payload.Str("location"),
				Cites:    e.Payload.Str("cites"),
				Drift:    e.Payload.Str("drift"),
				SeatID:   e.SeatID,
				Round:    e.Round,
				Verified: reruns[sha],
			})
		case "verify":
			out.Verifications = append(out.Verifications, EvidenceVerificationJSON{
				Claim:      e.Payload.Str("claim"),
				Reference:  e.Payload.Str("reference"),
				Trust:      e.Payload.Str("trust"),
				AccessDate: e.Payload.Str("access_date"),
				SeatID:     e.SeatID,
				Round:      e.Round,
			})
		}
	}

	out.Counts.Sources = len(out.Sources)
	out.Counts.Proofs = len(out.Proofs)
	out.Counts.Verifications = len(out.Verifications)
	for _, p := range out.Proofs {
		if p.Verified == nil {
			out.Counts.ProofsUnverified++
		}
	}
	return out
}

// EvidenceJSONBytes renders the evidence view as indented JSON.
func EvidenceJSONBytes(runDir string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(EvidenceJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
