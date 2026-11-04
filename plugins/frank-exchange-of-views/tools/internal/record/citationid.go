package record

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// CITATION IDENTITY MIRRORS FINDING IDENTITY — see NewID for the full argument.
//
// A citation, like a finding, is a TOOL-INSERTED invisible anchor (a citation anchor)
// that the seat never hand-writes and cannot invent. The id is random for the same
// reason a finding id is: an id you cannot guess is one you have to LOOK UP, so "which
// citation do you mean" is a read, not a memory exercise. Under the lockdown the anchor
// leaves the document only with a retire that names it, so each cite event's anchor is in the
// document or on a retire — the id is the identity that pins both halves together.

// Source is one cited source, drawn from a blue `cite` event — the composer input for the
// assembled bibliography and the hash a reader verifies the cached bytes against.
type Source struct {
	Label      string
	URL        string
	Sha256     string
	Title      string
	AccessDate string
	Location   string
	// Pages are the PDF pages an OCR span was found on, from blue's cite. Assembly prints them in
	// the source's note, never in the Bibliography.
	Pages []int32
	// WorkStatus is what a maintained index says about the cited work — stamped by the tool at
	// the cite, never typed. It reaches the reader: assembly prints a retraction in the source's
	// note AND in its Bibliography line, because a reader who follows one may never see the
	// other, and a retracted source presented as ordinary is the one defect a sound pipeline
	// still publishes.
	WorkStatus recordpb.WorkStatus
	// SourceCompleteness is which part of the work this citation's copy was. Assembly marks a
	// citation that rests on the abstract alone, because an abstract is quoted exactly like a
	// study and nothing else in the document tells the reader which it was.
	SourceCompleteness recordpb.SourceCompleteness
	// CopySubstitution and CopyURL say whose bytes a citation read when they are not the url's
	// own, and where they came from. Assembly prints the copy beside the url, because the url is what a reader follows and it is
	// not what was read.
	CopySubstitution recordpb.CopySubstitution
	CopyURL          string
}

// citedSource reads a citable source off an event body, if it is one: blue's `cite`, the one act that
// places a citation. A lens's reading of a source it found places none — it mints the gap blue
// answers with a cite — so a source enters the report only by blue's cite.
func citedSource(body proto.Message) (Source, bool) {
	b, ok := body.(*recordpb.Cite)
	if !ok {
		return Source{}, false
	}
	return Source{
		Label:      b.GetLabel(),
		URL:        b.GetUrl(),
		Sha256:     b.GetSha256(),
		Title:      b.GetTitle(),
		AccessDate: b.GetAccessDate(),
		Location:   b.GetLocation(),
		Pages:      b.GetPages(),
		WorkStatus: b.GetWorkStatus(),

		SourceCompleteness: b.GetSourceCompleteness(),
		CopySubstitution:   b.GetCopySubstitution(),
		CopyURL:            b.GetCopyUrl(),
	}, true
}

// CitedSources returns the distinct cited sources — one per blue `cite` label, in
// first-seen (label) order. It is the composer input for the assembled ## Bibliography and,
// via its labels, the EXPECTED set the unbacked_citations detector checks.
func CitedSources(run Run) ([]Source, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []Source
	for _, e := range m.Events {
		body, ok := recordpb.Body(e)
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
		// The empty-label filter is load-bearing: an event with no label names no anchor, and
		// CitationLabels — the EXPECTED set the lockdown compares against the report — must not
		// gain a phantom.
		if src.Label == "" || seen[src.Label] {
			continue
		}
		seen[src.Label] = true
		out = append(out, src)
	}
	return out, nil
}

// CiteByLabel returns the blue cite a citation label names, or nil when no cite carries it.
func CiteByLabel(run Run, label string) (*recordpb.Cite, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	var out *recordpb.Cite
	for _, e := range m.Events {
		if c := e.GetCite(); c != nil && c.GetLabel() == label {
			out = c
		}
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
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	return citationLabelsAt(db)
}

// citationLabelsAt is CitationLabels asked of q — the write path's transaction, or the run's handle.
func citationLabelsAt(q recordsql.Querier) ([]string, error) {
	evs, _, err := eventsAt(q)
	if err != nil {
		return nil, err
	}
	return citationLabelsOf(evs), nil
}

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

// ExistingCiteByKey returns the prior blue cite this seat recorded under the same --key, nil for
// none, so a crash-retried `blue cite` returns its existing anchor label instead of minting a
// duplicate — and the verb can hold the retry to the cite it claims to repeat. Mirrors FindingByKey: the retry dedup is a short-circuit
// BEFORE the fetch and the marker insert, not a change to the event key (which stays the
// unique citation label). A blue cite carries a `label`; red's `lens cite` does not, so
// this scans only the blue side of the shared "cite" event type.
func ExistingCiteByKey(run Run, seatID, key string) (*recordpb.Cite, error) {
	if key == "" {
		return nil, nil
	}
	// key is non-empty (guarded above), so a NULL cite_key can never match it — the same
	// property the fold's absent-key zero value had, held by the comparison itself.
	var label sql.NullString
	if _, err := queryRow(run, []any{&label},
		`SELECT c."label" FROM "cite" c JOIN "events" e ON e."id" = c."event_id"
		  WHERE e."seat_id" = ? AND c."cite_key" = ? ORDER BY c."event_id" LIMIT 1`,
		seatID, key); err != nil || label.String == "" {
		return nil, err
	}
	return CiteByLabel(run, label.String)
}

// TWO KINDS OF `cite` EVENT SHARE ONE TYPE — and conflating them inflates red's audit metric.
//
// `lens cite` records RED VERIFYING a source (a citation-ledger row: claim / reference /
// confidence / access_date). `blue cite` records BLUE AUTHORING one (label / url / sha256 /
// title / location, plus an anchor in the report). Only blue's carries `label`, which
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

// ExistingComputationByKey gives `blue compute` crash-retry idempotency: the computation this seat
// recorded under key, or nil. A seat whose message died after the event landed re-runs the same key
// and is held to the computation it repeats rather than executing the bundle a second time and
// splicing a second anchor.
func ExistingComputationByKey(run Run, seatID, key string) (*recordpb.Computation, error) {
	if key == "" {
		return nil, nil
	}
	var id sql.NullString
	if _, err := queryRow(run, []any{&id},
		`SELECT p."computation_id" FROM "computation" p JOIN "events" e ON e."id" = p."event_id"
		  WHERE e."seat_id" = ? AND p."key" = ? ORDER BY p."event_id" LIMIT 1`,
		seatID, key); err != nil || id.String == "" {
		return nil, err
	}
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	for _, e := range m.Events {
		if p := e.GetComputation(); p != nil && p.GetComputationId() == id.String {
			return p, nil
		}
	}
	return nil, nil
}

// Computation is one recorded computation, drawn from a blue `compute` event — the composer input
// for the assembled report's Computations section and the handle red re-runs.
type Computation struct {
	Label  string // the computation anchor id
	SHA    string // the bundle's
	Basis  string // the Basis word
	Argv   []string
	Exit   int
	Cites  string // the METHOD citation this applies, when blue named one
	Reason string
	// ToolchainPath and ToolchainSha256 name argv[0] as it resolved where blue ran it: "" for a
	// bundle file, whose bytes are in the bundle.
	ToolchainPath, ToolchainSha256 string

	// Drift is WHAT MOVED between the two runs, and Invariant what the seat stated holds across them;
	// both are present exactly when the basis is variant.
	Drift, Invariant string

	// Rerun is THIS ACT's comparison (#343): its latest `reproduce`, which set its own recorded
	// output, exit and out/ against an execution of the bundle. Nil when nobody compared it — and
	// that absence is information the reader needs, so the report says so rather than omitting it.
	Rerun *ComputationRerun
	// Executed is whether any checker executed the bundle: an act with no comparison on an executed
	// bundle is not compared, never not re-run.
	Executed bool
	// Standing is the bundle's standing re-run — the execution a later reproduce of any act on the
	// bundle compares against instead of executing — nil where no reproduce that ran it names its
	// execution.
	Standing *StandingRerun
	// Soundness is THIS ANCHOR's sentence — sound | unsound | not_judged | reopened — and Judgements
	// are the standing verdicts it is read from, one per seat (computationverdict.go).
	Soundness  string
	Judgements []ComputationJudgement
}

// ComputationRerun is one act's comparison: the `reproduce` that set the act's recorded result
// against an execution of its bundle. The result is COMPUTED by the tool, not claimed by the seat —
// which is the whole reason re-running beats re-reading.
type ComputationRerun struct {
	SeatID  string
	Epoch   int
	Sitting int
	Rerun   recordpb.Rerun
	// Drift is what differs between the act's recorded result and the execution, present where it
	// differed.
	Drift string
	// Ran is whether this reproduce executed the bundle itself.
	Ran bool
	// RerunSha names the execution compared against, rerun/<RerunSha>/ in the store; "" on a
	// reproduce migrated from a record that kept no raw re-run.
	RerunSha string
}

// StandingRerun is a bundle's standing re-run: the latest `reproduce` that executed it and named the
// execution.
type StandingRerun struct {
	SeatID   string
	RerunSha string
	Exit     int
	// Rerun is the standing reproduce's comparison of its own act; calls_the_tool says the execution
	// called the tool, so no act has an output of it to compare.
	Rerun recordpb.Rerun
}

// ComputationJudgement is one standing soundness verdict on a computation's anchor: red's JUDGEMENT,
// made by reading the program, of whether it establishes that sentence.
type ComputationJudgement struct {
	SeatID string
	Epoch  int
	Sound  bool
	Note   string
}

// RecordedComputations returns every computation on the record, in event order, each with its own
// comparison, its bundle's standing re-run and its own sentence's verdict.
//
// The OUTPUT and the BUNDLE are deliberately not here: they live on disk under
// <run>/computations/<bundle_sha>/ and can be large. The assembler reads them from the store so the
// report shows the exact bytes that ran, rather than a copy the record made of them.
func RecordedComputations(run Run) ([]Computation, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	ledger := anchorChecksOf(m.Events, m.At)
	reruns, acts := bundleRerunsOf(m.Events), actVerdictsOf(m.Events)
	var out []Computation
	for _, e := range m.Events {
		pf, isComputation := recordpb.BodyAs[*recordpb.Computation](e)
		if !isComputation {
			continue
		}
		v := computationVerdictOf(pf.GetComputationId(), pf.GetBundleSha(), ledger.at(pf.GetComputationId()), reruns, acts)
		p := Computation{
			Label: pf.GetComputationId(),
			SHA:   pf.GetBundleSha(),
			Basis: recordpb.Word(pf.GetBasis()),
			Argv:  pf.GetArgv(),
			Exit:  int(pf.GetExit()),
			Cites: pf.GetCites(),
			// --reason lands on Computation.text, the message's only prose channel — the same
			// flag/field split recordpb/required.go records for Verify.text.
			Reason:          pf.GetText(),
			ToolchainPath:   pf.GetToolchainPath(),
			ToolchainSha256: pf.GetToolchainSha256(),
			Drift:           pf.GetDrift(),
			Invariant:       pf.GetInvariant(),
			Soundness:       v.soundness,
		}
		if r := v.compared(); r != nil {
			w := m.At.Of(v.comparison)
			p.Rerun = &ComputationRerun{SeatID: v.comparison.GetSeatId(), Epoch: w.Epoch, Sitting: w.Sitting, Rerun: r.GetRerun(),
				Drift: r.GetDrift(), Ran: r.GetRan(), RerunSha: r.GetRerunSha()}
		}
		if b := v.bundle; b != nil {
			p.Executed = true
			if b.standing != nil {
				r, _ := recordpb.BodyAs[*recordpb.Reproduce](b.standing)
				p.Standing = &StandingRerun{SeatID: b.standing.GetSeatId(), RerunSha: r.GetRerunSha(), Exit: int(r.GetRerunExit()), Rerun: r.GetRerun()}
			}
		}
		for _, c := range v.judgements {
			r, _ := recordpb.BodyAs[*recordpb.Reproduce](c.ev)
			p.Judgements = append(p.Judgements, ComputationJudgement{SeatID: c.seat, Epoch: m.At.Of(c.ev).Epoch, Sound: c.holds, Note: r.GetNote()})
		}
		out = append(out, p)
	}
	return out, nil
}

// ComputationAt is the recorded computation at one anchor, with its comparison, its bundle's standing
// re-run and its sentence's verdict — what `reproduce --anchor` acts on. An anchor no computation
// carries is refused by name.
func ComputationAt(run Run, anchorID string) (Computation, error) {
	computations, err := RecordedComputations(run)
	if err != nil {
		return Computation{}, err
	}
	for _, p := range computations {
		if p.Label == anchorID {
			return p, nil
		}
	}
	return Computation{}, fmt.Errorf("--anchor names %s, which no computation on this record carries (%d computation(s)) — it takes the id inside a computation's token in the report; `show evidence` lists every computation by its anchor", anchorID, len(computations))
}
