package record

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"strings"

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
	// Pages are the PDF pages an OCR span was found on, from blue's cite; a corroboration has
	// none. Assembly prints them in the source's note, never in the Bibliography.
	Pages []int32
	// Corroborated marks a source red found itself (a supporting corroboration) rather than one
	// blue cited. The Bibliography prefers blue's title for a URL both name.
	Corroborated bool
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
	// own, and where they came from — blue's cite, or the reading a corroboration was made against.
	// Assembly prints the copy beside the url, because the url is what a reader follows and it is
	// not what was read.
	CopySubstitution recordpb.CopySubstitution
	CopyURL          string
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
			Pages:      b.GetPages(),
			WorkStatus: b.GetWorkStatus(),

			SourceCompleteness: b.GetSourceCompleteness(),
			CopySubstitution:   b.GetCopySubstitution(),
			CopyURL:            b.GetCopyUrl(),
		}, true
	case *recordpb.Verify:
		// No sha: the bibliography renders title + url + access date, never the hash, and the copy
		// red's reading was made against is the copy pair below. `location` is blue's
		// placement field; a corroboration's span is its `claim`.
		return Source{
			Label:        b.GetLabel(),
			URL:          b.GetUrl(),
			Title:        b.GetTitle(),
			AccessDate:   b.GetAccessDate(),
			Location:     b.GetClaim(),
			Corroborated: true,
			WorkStatus:   b.GetWorkStatus(),

			SourceCompleteness: b.GetSourceCompleteness(),
			CopySubstitution:   b.GetCopySubstitution(),
			CopyURL:            b.GetCopyUrl(),
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

// ExistingCorroboration returns the labelled corroboration this seat already recorded for the same
// source and the same claim, nil for none, so a crash-retried `lens corroborate` returns its anchor
// instead of splicing a second one — and the verb can hold the retry to the reading it repeats.
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
func ExistingCorroboration(run Run, seatID, url, claim string) (*recordpb.Verify, error) {
	if url == "" || claim == "" {
		return nil, nil
	}
	var label string
	if _, err := queryRow(run, []any{&label},
		`SELECT v."label" FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
		  WHERE e."seat_id" = ? AND v."url" = ? AND v."claim" = ? AND COALESCE(v."label", '') != ''
		  ORDER BY v."event_id" LIMIT 1`,
		seatID, url, claim); err != nil || label == "" {
		return nil, err
	}
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	for _, e := range m.Events {
		if v := e.GetVerify(); v != nil && v.GetLabel() == label {
			return v, nil
		}
	}
	return nil, fmt.Errorf("corroboration %s is in the record's verify table and in none of its events", label)
}

// contradictionsOf is every contradiction red recorded, in the order the record opened them, each
// with the finding that raised it and the gap that answered it ("" while each is unset). ids is
// evs's events.id, aligned with it.
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
// judged one. So the duty is REPORTED and red grades its own finding.
//
// THE RULE, PER CLAIM (gblock, 2026-10-06), folded in record order:
//
//   - A READING is a verify whose outcome contradicts (refutes | absent) a non-empty claim. The
//     first reading of a claim with none open OPENS a contradiction, whose ID is that reading's
//     events.id. A later reading of the same claim while it is open JOINS it: its seat becomes the
//     owner, and `since` moves to it, so dispatch readies the new owner. The id does not change.
//   - A FINDING RAISES the open contradiction it names by reference (`--about-kind contradiction
//     --about <id>`), and nothing else: a finding whose quote contains the claim raises nothing.
//     The raise covers every reading joined so far. A reading of the claim after the finding opens
//     a NEW contradiction with its own id: the raise was of what had been read.
//   - A GAP ANSWERS what a finding raised: a mint whose found_by names the raising finding.
//     A source that contradicts or is silent on a claim is a citation defect, and citation classes
//     are material whatever the lens judges (gblock, 2026-10-19: the class default wins), so the
//     finding alone — which holds no gate and obliges nobody — does not discharge it. The gap's
//     class then decides its materiality, as for every gap. A gap its RAISER mints this way takes
//     no mint of that lens's budget (ownAnswersAt).
//
// NAME IT OR IT IS NOT RAISED (gblock, 2026-10-23). The id names the reading, which an edit does
// not move, so the raise survives blue editing the sentence away. A quote is a placement in the
// report, not an act about a reading: the tool does not infer which contradiction a finding is
// about from what its quote happens to contain.
//
// THE REFERENCE IS ANY LENS'S. A stricter join (same seat) would refuse a contradiction one lens
// found and another raised, which is the collaboration the lens roles exist for.
//
// A reading that names a citation carries its anchor, and the anchor's standing answers it too —
// a retire, or a backing check after an edit (unansweredContradictions). That answer is the
// citation's state, not an act naming the contradiction, so it is not this fold's: a finding may
// still raise a contradiction its citation's standing answers.
func contradictionsOf(evs []*Event, ids []int64) []contradiction {
	var out []*contradiction
	open := map[string]*contradiction{} // claim -> its open contradiction
	raise := func(c *contradiction, by, seat string) {
		c.raisedBy, c.raiser = by, seat
		delete(open, c.claim)
	}
	for i, e := range evs {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Verify:
			claim := strings.TrimSpace(b.GetClaim())
			if !contradicts(b.GetOutcome()) || claim == "" {
				continue
			}
			if c := open[claim]; c != nil {
				c.reader, c.since = e.GetSeatId(), ids[i]
				continue
			}
			c := &contradiction{id: ids[i], claim: claim, opener: e.GetSeatId(), reader: e.GetSeatId(), anchor: b.GetAnchor(), since: ids[i]}
			open[claim] = c
			out = append(out, c)
		case *recordpb.Finding:
			if b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_CONTRADICTION {
				for _, c := range out {
					if c.raisedBy == "" && strconv.FormatInt(c.id, 10) == strings.TrimSpace(b.GetAboutRef()) {
						raise(c, b.GetId(), e.GetSeatId())
					}
				}
			}
		case *recordpb.Mint:
			for _, c := range out {
				if c.raisedBy != "" && c.answeredBy == "" && slices.Contains(b.GetFoundBy(), c.raisedBy) {
					c.answeredBy = b.GetGapId()
				}
			}
		}
	}
	res := make([]contradiction, len(out))
	for i, c := range out {
		res[i] = *c
	}
	return res
}

// ReadingContradiction is the contradiction one recorded reading opened or joined: its id, whether
// the reading opened it, and the seat whose reading opened it.
type ReadingContradiction struct {
	ID     int64
	Opened bool
	Opener string
}

// ContradictionOfReading is the contradiction reading — a verify event as its write returned it —
// opened or joined, read from the record's fold (contradictionsOf) through that reading; nil where
// the reading opened and joined none (it backs the claim, or could not read the source). The fold
// stops at the reading so a later seat's reading of the same claim, which joins the contradiction
// and moves its `since`, does not hide which contradiction this one landed in. The reading is found
// by its key, which the record holds unique.
func ContradictionOfReading(run Run, reading *Event) (*ReadingContradiction, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	var at int64
	if found, err := queryRowAt(db, []any{&at}, `SELECT "id" FROM "events" WHERE "key" = ?`, reading.GetKey()); err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("the reading keyed %q is not on the record, so which contradiction it opened or joined cannot be read", reading.GetKey())
	}
	evs, win, err := eventsOfAt(db, contradictionActs...)
	if err != nil {
		return nil, err
	}
	ids := win.IDs(evs)
	n := 0
	for n < len(ids) && ids[n] <= at {
		n++
	}
	for _, c := range contradictionsOf(evs[:n], ids[:n]) {
		if c.since == at {
			return &ReadingContradiction{ID: c.id, Opened: c.id == at, Opener: c.opener}, nil
		}
	}
	return nil, nil
}

// requireOpenContradiction refuses a finding's reference to a contradiction that is not open: a
// ref that is not an id, one no reading opened, one a finding already raised — that refusal names
// the raising finding and the mint that answers it — or one a gap already answered. It runs on
// replay as well: the reading is written before the finding that raises it, so a translated
// reference resolves and a mistranslated one is loud.
func requireOpenContradiction(q recordsql.Querier, ref string) error {
	ref = strings.TrimSpace(ref)
	if _, err := strconv.ParseInt(ref, 10, 64); err != nil {
		return fmt.Errorf("--about %q is not a contradiction's id: it is the number `show evidence` lists as the `id` of each of its unanswered_contradictions", ref)
	}
	evs, win, err := eventsOfAt(q, contradictionActs...)
	if err != nil {
		return err
	}
	open := 0
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		switch {
		case c.idWord() == ref && c.answeredBy != "":
			return fmt.Errorf("contradiction %s is already answered, by gap %s minted from finding %s: a source read against the claim after it opens a new one with its own id", ref, c.answeredBy, c.raisedBy)
		case c.idWord() == ref && c.raisedBy != "":
			return fmt.Errorf("contradiction %s is already raised, by finding %s: one finding raises a contradiction, and the gap minted from it answers it — mint it with `--found-by %s`", ref, c.raisedBy, c.raisedBy)
		case c.idWord() == ref:
			return nil
		case c.raisedBy == "":
			open++
		}
	}
	return fmt.Errorf("--about %s names no contradiction on this record (%d open) — a contradiction's id is the `id` `show evidence` lists under unanswered_contradictions", ref, open)
}

// unansweredContradictions is contradictionsOf less the ones answered: by a gap minted from the
// finding that raised it, or by the
// standing of the citation the reading checked (gblock, 2026-10-16). win supplies each event's
// events.id and sitting.
//
// THE CITATION'S STANDING ANSWERS IT in two cases, both read off the anchor ledger:
//
//   - blue RETIRES the anchor: the sentence left the report with its claim, and there is nothing
//     left for a finding to raise;
//   - the anchor HOLDS (anchorStanding.holds): a check that backs the sentence stands, and no
//     check that does not. The reading itself has stopped standing — an edit changed the sentence
//     after its sitting opened, or its own seat read the sentence again — and what the record now
//     says of the sentence is that it is backed.
//
// ANOTHER LENS'S BACKING CHECK DOES NOT ANSWER A READING THAT STANDS. It is a disagreement over
// the same sentence, and another seat's verdict never replaces a lens's own (anchorchecks.go): the
// contradiction stays the reader's to raise. Blue cannot dissolve it by rewording either: the
// reworded sentence is reopened, owed a check, and only a check that backs it answers the
// contradiction.
//
// A corroboration's reading checked no citation (its anchor is ""), so only the gap minted from
// the finding that raised it answers it.
func unansweredContradictions(evs []*Event, win WindowIndex) []contradiction {
	ledger := anchorChecksOf(evs, win)
	var out []contradiction
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		if c.answeredBy != "" {
			continue
		}
		if s := ledger.at(c.anchor); s != nil && (s.retired || s.holds()) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// contradiction is one contradiction red recorded: its id (the events.id of the reading that opened
// it), the claim, the seat whose reading opened it (opener), the seat whose reading last joined it — the lens whose act raises it — and where
// that reading sits, which is when the lens came to owe it. anchor is the citation the opening
// reading checked, "" for a corroboration; raisedBy is the id of the finding that raised it and
// raiser the seat that recorded that finding; answeredBy is the id of the gap minted from that
// finding.
type contradiction struct {
	id                           int64
	claim, opener, reader        string
	anchor, raisedBy, answeredBy string
	raiser                       string
	since                        int64
}

// contradictionActs is the event types contradictionsOf folds: the readings that open, the findings
// that raise, and the mints that answer. Every reader of the fold reads these, so a
// type that answers cannot be read by one reader and missed by another.
var contradictionActs = []recordpb.EventType{
	recordpb.EventType_EVENT_TYPE_VERIFY, recordpb.EventType_EVENT_TYPE_FINDING,
	recordpb.EventType_EVENT_TYPE_MINT,
}

// idWord is the contradiction's id as every surface spells it: decimal, the `--about` value.
func (c contradiction) idWord() string { return strconv.FormatInt(c.id, 10) }

// contradicts is the half of the outcome set that says the text is not supported. It is the
// complement of the set that earns a footnote, and the two are written as one pair so a value
// added to the enum cannot fall between them unnoticed.
func contradicts(o recordpb.SourceOutcome) bool {
	return o == recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES ||
		o == recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT
}
