package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A CLAIM THAT SAYS "MEASURED" MUST POINT AT THE MEASUREMENT.
//
// #591's third ask: a report sentence framed as measured or re-runnable must resolve to a proof
// id, or say it is unrecorded. The run that produced the issue shipped one that did neither —
// "measured, re-run at synthesis time, at 30 files / ~508KB", cited to a proof footnote, with no
// archived proof carrying the figure. It had been measured in a seat's context and died there.
//
// # This is the third detector on an axis that had two
//
// `dropped_finding_markers` asks whether a marker red anchored is still in the report.
// `unbacked_citations` asks the same of a citation. Both are EXPECTED-⊄-PRESENT checks over
// tool-managed anchors, and both live in the scorecard. The proof axis had the extractor —
// claimcount.ProofAnchorIDs, written when proof anchors joined the protected union so the
// hookgate lockdown would refuse a raw write that dropped one — and nothing that ever read it to
// ask the question. The protection existed without the observation.
//
// # Both directions, because they are different defects
//
// A proof ON THE RECORD whose anchor is ABSENT from the report is a computation this run
// performed that the document never points at: the evidence sits in the record and the cache,
// and the sentence it was run for now stands on its own authority. That is the ask, stated
// exactly.
//
// An anchor IN THE REPORT resolving to NO proof event is the mirror: a claim pointing at
// nothing. The assembler already notices this one — weaveProofs renders it as
// "_(unresolved proof … — no proof event on the record)_" — and then nothing reads that line,
// so the fact lives in prose in the artifact instead of in a check. Recovering it from that
// sentence would be the defect twice over; it is computed from the two sets here.
//
// # WHAT THIS DELIBERATELY DOES NOT DO, and the measurement that decided it
//
// The obvious reading of #591 is to scan the prose for measurement words — measured, re-run,
// computed — and flag the sentences carrying no anchor. Measured against both 2026-08-23
// reports, that heuristic finds 23 candidate units and 21 of them are ordinary English:
// "Measured against the external literature", "re-run its acceptance criterion", "a computed ρ
// at the documented default", "Attested Computation concept type". A detector at better than
// nine-in-ten false positives is not a weak detector, it is a line readers learn to skip, and a
// gate nobody reads is worse than an absent one because it looks like coverage.
//
// So the prose axis is not built, and this says so rather than shipping it quietly. What is
// checkable without a heuristic is the anchor↔record join, and that is what runs.

// ProofBackingAudit joins the proofs on the record against the proof anchors in blue's report.
func ProofBackingAudit(runDir string) Audit {
	// BLUE'S REPORT, NOT THE ASSEMBLED ONE, and the two are not interchangeable here. Assembly
	// WEAVES proof anchors into visible [^Pn] footnotes and the raw tokens are gone by design
	// (the thing FootnoteIntegrity judges). The anchors only exist to be joined in blue/report.md.
	blue := filepath.Join(runDir, "blue", "report.md")
	md, err := os.ReadFile(blue)
	if err != nil {
		return Audit{Check: "proof-backing", Verdict: "SKIP",
			Detail: fmt.Sprintf("no blue/report.md to read (%v) — proof anchors live there and are woven away at assembly, so there is nothing to join", err)}
	}
	proofs, err := record.RecordedProofs(runDir)
	if err != nil {
		return Audit{Check: "proof-backing", Verdict: "SKIP",
			Detail: fmt.Sprintf("the record could not be read (%v), so no proof could be joined to the report — NOT a run whose claims all resolve", err)}
	}
	anchors := claimcount.ProofAnchorIDs(string(md))

	// NOTHING ON EITHER SIDE IS NOTHING TO JOIN, and it is checked first: a run that ran no
	// proofs and anchors none is not a report missing its record, so it must not be told it is.
	if len(proofs) == 0 && len(anchors) == 0 {
		return Audit{Check: "proof-backing", Verdict: "SKIP", Detail: "this run recorded no proofs and its report anchors none"}
	}

	// A REPORT WITH NO RECORD BEHIND IT CANNOT BE JOINED, AND MUST NOT BE CONVICTED.
	//
	// MEASURED, on the first run against the real 2026-08-23 artifacts: both reported FAIL with
	// every anchor "resolving to nothing" — 6 and 2 of them — and every one of those anchors is
	// fine. The run directories in research/ carry the committed report and NOT records/, which
	// is gitignored, and the record layer reads an absent record directory as a legal EMPTY run
	// rather than an error (deliberately: a run that has recorded nothing is a real state).
	//
	// So zero proofs has two causes, and the audit was treating them as one. A record holding
	// events but no proof events is a real finding about this run. A record holding NO EVENTS AT
	// ALL is not a run this audit can speak about: every real sitting registers, so an event-less
	// record beside a written report means the two artifacts are not a pair, and convicting the
	// report on that is the false-positive shape this whole audit tier exists to refuse.
	board, berr := record.BoardState(runDir)
	if berr != nil || board == nil || len(board.Events) == 0 {
		return Audit{Check: "proof-backing", Verdict: "SKIP",
			Detail: fmt.Sprintf("blue/report.md anchors %d proof(s) and this run's record carries no events at all, so the two cannot be joined — every real sitting registers, so this is a report without its record rather than %d claims pointing at nothing",
				len(anchors), len(anchors))}
	}

	expected := make([]string, 0, len(proofs))
	onRecord := map[string]bool{}
	for _, p := range proofs {
		if p.Label == "" {
			continue
		}
		expected = append(expected, p.Label)
		onRecord[p.Label] = true
	}

	// EXPECTED ⊄ PRESENT, through the shared helper rather than a loop here — the mistake
	// CitationLabelsOf's own comment records is a detector that rebuilt the rule inline and then
	// stayed on the old one when the rule widened.
	unanchored := claimcount.MissingProofAnchorIDs(expected, string(md))

	// PRESENT ⊄ EXPECTED is the other direction, and it is a set difference rather than a rule,
	// so computing it from the same two sets does not make a second copy of anything.
	var unresolved []string
	for _, id := range anchors {
		if !onRecord[id] {
			unresolved = append(unresolved, id)
		}
	}
	sort.Strings(unanchored)
	sort.Strings(unresolved)

	detail := fmt.Sprintf("%d proof(s) on the record, %d anchored in blue/report.md", len(expected), len(anchors))
	var findings []string
	if len(unresolved) > 0 {
		findings = append(findings, fmt.Sprintf("%d anchor(s) resolve to NO proof event — a claim pointing at nothing: %s",
			len(unresolved), strings.Join(unresolved, ", ")))
	}
	if len(unanchored) > 0 {
		findings = append(findings, fmt.Sprintf("%d proof(s) this run RAN reach the report nowhere — the computation exists and the sentence it backs stands on its own authority: %s",
			len(unanchored), strings.Join(unanchored, ", ")))
	}
	if len(findings) == 0 {
		return Audit{Check: "proof-backing", Verdict: "PASS", Detail: detail + "; every proof is anchored and every anchor resolves"}
	}
	return Audit{Check: "proof-backing", Verdict: "FAIL", Detail: detail + "; " + strings.Join(findings, "; ")}
}
