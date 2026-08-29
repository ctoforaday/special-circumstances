package record

import (
	"encoding/json"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
// verifications, and nothing of what blue cited — is subsumed here: anchored checks attach to the
// source they name, and red's own corroboration is the `independent` array.
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
	URL        string `json:"url"`
	Title      string `json:"title"`
	Sha256     string `json:"sha256"`
	AccessDate string `json:"access_date"`
	// Location is the sentence the anchor sits at — what the source is offered as backing for.
	//
	// It is also the ONLY copy of that sentence. A cite carried it twice — `claim` re-quoted
	// what `location` already held to place the anchor — and Cite reserves field 8 to keep the
	// second spelling from coming back. There is no `claim` on a source here because there is
	// no second span to render; `Verify.claim` is a different field and is still on a
	// verification below.
	Location string `json:"location"`
	SeatID   string `json:"seat_id"`
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
	Anchor string `json:"anchor"`
	Sha256 string `json:"sha256"`
	Basis  string `json:"basis"`
	// Cites is the citation anchor of the METHOD this computation applies, when blue named one —
	// the link between the two evidence classes, and the only one on the record.
	Cites string `json:"cites"`
	// Drift is WHAT MOVED when the tool's two runs of the same script disagreed, which makes the
	// result a measurement of a moving system rather than a proof. Empty means they agreed.
	// It was briefly reduced to a boolean while the sentence lived in the proof cache's
	// meta.json; that file is gone and `Proof.drift` carries the sentence itself.
	Drift  string `json:"drift"`
	SeatID string `json:"seat_id"`
	Round  int    `json:"round"`

	// THREE FIELDS ARE GONE FROM THIS ROW AND THAT IS A LOSS, NOT A TIDY-UP.
	//
	//   script, exit — the frozen key census rules both OFF the record: the script is a file in
	//   the proof cache and the exit status belongs to proof.Result, the in-process execution
	//   struct. Rendering `"exit": 0` from a field nothing can write would state success on
	//   every proof ever recorded.
	//
	//   location — `Proof` carries no span at all. A citation's anchor resolves to the sentence
	//   it backs; a proof's no longer does, which is half of what this view was built for. It is
	//   reported rather than rendered as an empty string, because "" and "nobody wrote one down"
	//   are the plausible zero this file exists to refuse.

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
	Note       string `json:"note"`
	SeatID     string `json:"seat_id"`
	Round      int    `json:"round"`
}

// EvidenceVerificationJSON is one `lens verify` — a claim red checked against a source.
type EvidenceVerificationJSON struct {
	Claim string `json:"claim"`
	// Anchor is the citation adjudicated, empty on an independent check.
	Anchor string `json:"anchor"`
	// Label is the footnote this reading MINTED, on a corroboration that supported the claim.
	// Empty on an anchored verify (it adjudicates a footnote blue already made) and on a reading
	// that did not back the claim — those splice nothing, by design.
	Label string `json:"label,omitempty"`
	// Outcome is what the source DID for the claim, and it has a negative half: `refutes` and
	// `absent` are the values that used to have nowhere to go, so the strongest finding on this
	// axis left as prose and the assembly screen looked for a verdict no field could carry.
	Outcome string `json:"outcome"`
	// Confidence is how sure red is of that outcome — orthogonal to it. `refutes` at low
	// confidence is a call for more evidence; at high confidence it is a finding to act on.
	Confidence string `json:"confidence"`
	// Text is red's reading — required, because a verdict with nothing behind it is the
	// assertion the verb exists to replace.
	Text string `json:"text"`
	// URL and Title name the source red read. They replace `reference`, a free-text third
	// spelling of a source on the one verb that reads one red found itself: the corroboration a
	// run rests on was identified less precisely than the citations it audits. A source is named
	// by url + title here as it is on a cite.
	URL        string `json:"url"`
	Title      string `json:"title"`
	AccessDate string `json:"access_date"`
	SeatID     string `json:"seat_id"`
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
	// Reopened are the anchors whose text has MOVED since they were placed — a citation whose
	// sentence blue has since rewritten. The reference stands and its referent changed, so a
	// verification of it is STALE rather than refuted; red re-reads these first, and a reader who
	// cannot tell stale from refuted either re-checks everything or trusts everything.
	Reopened []string `json:"reopened"`
	// UnansweredContradictions are the claims where red read a source that CONTRADICTS or does
	// not support the report, and no finding was ever raised about it.
	//
	// STATED HERE BECAUSE THIS IS WHERE RED LOOKS. The duty is enforced at the merge's PASS gate,
	// which is the right place to REFUSE — but a duty that only surfaces when someone else is
	// blocked at the end of the round is one the seat that owes it never sees. An empty array is
	// the honest "nothing outstanding"; without the field, nothing outstanding and nothing
	// checked are the same absence.
	UnansweredContradictions []string `json:"unanswered_contradictions"`
	Counts                   struct {
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
	if out.UnansweredContradictions = unansweredContradictions(b); out.UnansweredContradictions == nil {
		out.UnansweredContradictions = []string{}
	}
	if out.Reopened = reopenedAnchors(b); out.Reopened == nil {
		out.Reopened = []string{}
	}

	// Red's verifications, split by whether they name a citation. The anchored ones are indexed
	// so each source carries its own; the rest are corroboration and stand alone.
	byAnchor := map[string][]EvidenceVerificationJSON{}
	for _, e := range b.Events {
		// BodyAs returns false for BOTH no body and a body of another type. Neither is a
		// verification, and neither is rendered as a check with every field blank.
		vf, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok {
			continue
		}
		// OUTCOME AND CONFIDENCE ARE ENUMS NOW, AND THIS VIEW RENDERS THE SEAT'S WORD. Every
		// consumer — Refuted(), the assembly screen, a seat reading the JSON — speaks
		// `supports`/`refutes`/`high`, so the value's spelling is what crosses the boundary, not
		// `SOURCE_OUTCOME_REFUTES`. `Word` maps the UNSPECIFIED zero back to "" for the same
		// reason the old read did: the pre-migration record had no outcome word for an absent
		// outcome, and `unspecified` is a word no seat ever typed.
		//
		// THE HYPHEN CAVEAT IS RESOLVED. `supports-with-bridge` was spelled with a hyphen on the
		// seat-facing surfaces while `Word` derives `supports_with_bridge`, so the help offered a
		// word the write path refused. The declared set carries the schema's spelling now, and
		// enums_test asserts that every declared value is one the record can hold.
		v := EvidenceVerificationJSON{
			Claim:      vf.GetClaim(),
			Anchor:     vf.GetAnchor(),
			Label:      vf.GetLabel(),
			Outcome:    recordpb.Word(vf.GetOutcome()),
			Confidence: recordpb.Word(vf.GetConfidence()),
			Text:       vf.GetText(),
			URL:        vf.GetUrl(),
			Title:      vf.GetTitle(),
			AccessDate: vf.GetAccessDate(),
			SeatID:     e.GetSeatId(),
			Round:      int(e.GetRound()),
		}
		out.Counts.Verifications++
		// THE SPLIT IS STILL ON THE ANCHOR, not on `Verify.independent`, and the empty string is
		// deliberate. A verification indexed under "" would join no source and disappear from
		// both arrays; sending it to `independent` keeps it visible. The schema now carries an
		// explicit `independent` bool alongside — a second carrier this view does not read.
		if v.Anchor == "" {
			out.Independent = append(out.Independent, v)
			continue
		}
		byAnchor[v.Anchor] = append(byAnchor[v.Anchor], v)
	}

	// Red's re-runs, keyed by the proof sha they checked — the one join the record supports.
	reruns := map[string]*EvidenceReproductionJSON{}
	for _, e := range b.Events {
		r, ok := recordpb.BodyAs[*recordpb.Reproduce](e)
		if !ok {
			continue
		}
		reruns[r.GetProofSha()] = &EvidenceReproductionJSON{
			// Reproduced is COMPUTED and written on every re-run, so its absence and a recorded
			// false both mean "these bytes did not match" for a reader of this view — which is
			// what the old `Payload.Get` + bool assertion already resolved to.
			Reproduced: r.GetReproduced(),
			Sound:      r.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND,
			Note:       r.GetNote(),
			SeatID:     e.GetSeatId(),
			Round:      int(e.GetRound()),
		}
	}

	for _, e := range b.Events {
		// TWO ARMS, so this stays `Body` plus a type switch rather than two `BodyAs` passes:
		// one walk of the events, and an event is a cite or a proof or neither.
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch bd := body.(type) {
		case *recordpb.Cite:
			// A cite with no label names no anchor, so it is not a citation a reader can reach.
			//
			// THIS IS NO LONGER "the label is what tells blue from red". It was, and citationid.go
			// used the same discriminator; since red's supporting corroborations mint a label too,
			// the two are told apart by EVENT TYPE (a Cite here, a Verify below) rather than by
			// whether a label is present. The filter stays for its own reason, stated above.
			label := bd.GetLabel()
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
				URL:        bd.GetUrl(),
				Title:      bd.GetTitle(),
				Sha256:     bd.GetSha256(),
				AccessDate: bd.GetAccessDate(),
				Location:   bd.GetLocation(),
				SeatID:     e.GetSeatId(),
				Round:      int(e.GetRound()),
				Verified:   checks,
			})
		case *recordpb.Proof:
			// `proof_sha` is the proof's sha256 — the record spells it once, on the field the
			// `reproduce` join is keyed by, so the two halves of that join cannot be two names
			// for one number. The JSON keeps `sha256`, which is what `reproduce --id` takes.
			sha := bd.GetProofSha()
			out.Proofs = append(out.Proofs, EvidenceProofJSON{
				Anchor:   bd.GetProofId(),
				Sha256:   sha,
				Basis:    bd.GetProofBasis(),
				Cites:    bd.GetCites(),
				Drift:    bd.GetDrift(),
				SeatID:   e.GetSeatId(),
				Round:    int(e.GetRound()),
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
func EvidenceJSONBytes(run Run) ([]byte, error) {
	b, err := BoardState(run)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(EvidenceJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
