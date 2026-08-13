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
// # Both halves now carry their join
//
// They did not. `lens verify` recorded a free-text `reference` rather than the anchor of the
// citation it checked, so this view listed red's work beside blue's sources without connecting
// them — and any join would have been a string-similarity guess presented as a fact, which reads
// identically to a right one whether it is right or not. #382 gave `verify` an `--anchor`, so a
// source now carries the verifications OF THAT SOURCE, and `verified: []` means nobody has
// checked it — stated, not inferred from an absence.
//
// Red's INDEPENDENT checks (corroboration it found itself, which blue never cited) have no
// anchor to join on and never will; they are the `independent` array, and that is a different
// fact from an unverified citation rather than a missing one.
//
// PROOFS carry the same join through `reproduce`'s `proof_sha`, and a nil `verified` means
// nobody re-ran it. Every absence in this view is a stated field, because an unaudited thing and
// a clean thing are otherwise the same empty space.

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

	// Verified is every `lens verify` naming THIS anchor. An empty slice is the honest zero and
	// it is rendered, not omitted: "nobody has checked this source" is what red reads to decide
	// where to spend its next pass, and an absent key would leave that to be inferred.
	Verified []EvidenceVerificationJSON `json:"verified"`
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
	// Anchor is the citation adjudicated, empty on an independent check.
	Anchor string `json:"anchor,omitempty"`
	// Outcome is what the source DID for the claim, and it has a negative half: `refutes` and
	// `absent` are the values that used to have nowhere to go, so the strongest finding on this
	// axis left as prose and the assembly screen looked for a verdict no field could carry.
	Outcome string `json:"outcome,omitempty"`
	// Text is red's reading — required, because a verdict with nothing behind it is the
	// assertion the verb exists to replace.
	Text       string `json:"text,omitempty"`
	Reference  string `json:"reference,omitempty"`
	AccessDate string `json:"access_date,omitempty"`
	SeatID     string `json:"seat_id,omitempty"`
	Round      int    `json:"round"`
}

// Refuted reports whether this verification found AGAINST the claim — the two outcomes that
// mean the report is carrying a citation it should not. A named predicate rather than a
// string comparison at each call site: the assembly screen, the counts here and any future
// reader must agree on what "found against" means, and they agree by calling this.
func (v EvidenceVerificationJSON) Refuted() bool {
	return v.Outcome == "refutes" || v.Outcome == "absent"
}

// EvidenceJSON is the whole evidence layer of the report.
type EvidenceJSON struct {
	Sources []EvidenceSourceJSON `json:"sources"`
	Proofs  []EvidenceProofJSON  `json:"proofs"`
	// Independent is red's corroboration — sources it went and found, which blue never cited and
	// which therefore have no anchor. A separate array rather than anchorless rows mixed into the
	// sources, because "checked something blue did not cite" and "checked a citation" answer
	// different questions and only one of them is about the report's own backing.
	Independent []EvidenceVerificationJSON `json:"independent"`
	Counts      struct {
		Sources int `json:"sources"`
		Proofs  int `json:"proofs"`
		// ProofsUnverified is the count nobody re-ran. It is stated because the honest zero and
		// the unchecked case are otherwise the same empty space.
		ProofsUnverified int `json:"proofs_unverified"`
		// SourcesUnverified is its citation twin: cited, and nobody has checked it.
		SourcesUnverified int `json:"sources_unverified"`
		// SourcesRefuted counts the citations red found AGAINST — refuted or absent. It is the
		// number the assembly screen acts on, and the one the old ledger could not produce.
		SourcesRefuted int `json:"sources_refuted"`
		Verifications  int `json:"verifications"`
	} `json:"counts"`
}

// EvidenceJSONOf projects the record's evidence layer in event order.
func EvidenceJSONOf(b *Board) EvidenceJSON {
	out := EvidenceJSON{
		Sources:     []EvidenceSourceJSON{},
		Proofs:      []EvidenceProofJSON{},
		Independent: []EvidenceVerificationJSON{},
	}

	// Red's verifications, split by whether they name a citation. The anchored ones are indexed
	// so each source carries its own; the rest are corroboration and stand alone.
	byAnchor := map[string][]EvidenceVerificationJSON{}
	for _, e := range b.Events {
		if e.Type != "verify" {
			continue
		}
		v := EvidenceVerificationJSON{
			Claim:      e.Payload.Str("claim"),
			Anchor:     e.Payload.Str("anchor"),
			Outcome:    e.Payload.Str("outcome"),
			Text:       e.Payload.Str("text"),
			Reference:  e.Payload.Str("reference"),
			AccessDate: e.Payload.Str("access_date"),
			SeatID:     e.SeatID,
			Round:      e.Round,
		}
		out.Counts.Verifications++
		if v.Anchor == "" {
			out.Independent = append(out.Independent, v)
			continue
		}
		byAnchor[v.Anchor] = append(byAnchor[v.Anchor], v)
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
			// An empty slice, never nil: `"verified": []` is "nobody has checked this source",
			// which is a fact red acts on. A null would leave it to be inferred.
			checks := byAnchor[label]
			if checks == nil {
				checks = []EvidenceVerificationJSON{}
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
				Verified:   checks,
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
			// `verify` is handled in the indexing pass above — it has to be, because a source
			// must carry its verifications and a cite event may arrive after the check of it.
		}
	}

	out.Counts.Sources = len(out.Sources)
	out.Counts.Proofs = len(out.Proofs)
	for _, p := range out.Proofs {
		if p.Verified == nil {
			out.Counts.ProofsUnverified++
		}
	}
	for _, s := range out.Sources {
		if len(s.Verified) == 0 {
			out.Counts.SourcesUnverified++
			continue
		}
		for _, v := range s.Verified {
			if v.Refuted() {
				out.Counts.SourcesRefuted++
				break
			}
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
