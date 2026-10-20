package record

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
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
	// Pages are the PDF pages an OCR quote was found on, from blue's cite; a corroboration has
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
		}, true
	case *recordpb.Verify:
		// No sha: red read the source itself rather than through the run cache, and the
		// bibliography renders title + url + access date, never the hash. `location` is blue's
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

// ExistingCiteByKey returns the label of a prior blue cite this seat recorded under the
// same --key, so a crash-retried `blue cite` returns its existing anchor label instead of
// minting a duplicate. Mirrors FindingByKey: the retry dedup is a short-circuit
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

// ExistingProofByKey gives `blue prove` crash-retry idempotency: a seat whose message died
// after the event landed re-runs the same key and gets the recorded sha back rather than
// executing the script a second time and splicing a second anchor.
func ExistingProofByKey(run Run, seatID, key string) (sha, id string, err error) {
	if key == "" {
		return "", "", nil
	}
	// The schema carries the one fact once, on Proof.proof_sha; reproduce writes the same value
	// and every joiner reads it back under this name.
	var s, i sql.NullString
	_, err = queryRow(run, []any{&s, &i},
		`SELECT p."proof_sha", p."proof_id" FROM "proof" p JOIN "events" e ON e."id" = p."event_id"
		  WHERE e."seat_id" = ? AND p."proof_key" = ? ORDER BY p."event_id" LIMIT 1`,
		seatID, key)
	return s.String, i.String, err
}

// Proof is one recorded computation, drawn from a blue `prove` event — the composer input
// for the assembled report's Proofs section and the handle red re-runs.
type Proof struct {
	Label  string // the proof anchor id
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

	// Rerun is the SCRIPT's standing re-run (#343): the latest `reproduce` of this sha that ran it,
	// shared by every proof on the script. Nil when nobody re-ran it — and that absence is itself
	// information the reader needs, so the report says so rather than omitting it.
	Rerun *ProofRerun
	// Soundness is THIS ANCHOR's sentence — sound | unsound | not_judged | reopened — and Judgements
	// are the standing verdicts it is read from, one per seat (proofverdict.go).
	Soundness  string
	Judgements []ProofJudgement
}

// ProofRerun is one `reproduce` that RAN a proof's script. The result is COMPUTED by the tool, not
// claimed by the seat — which is the whole reason re-running beats re-reading.
type ProofRerun struct {
	SeatID  string
	Epoch   int
	Sitting int
	Rerun   recordpb.Rerun
	// Recorded and Observed are the two outputs, kept only where the re-run differed.
	Recorded, Observed string
}

// ProofJudgement is one standing soundness verdict on a proof's anchor: red's JUDGEMENT, made by
// reading the script, of whether it establishes that sentence.
type ProofJudgement struct {
	SeatID string
	Epoch  int
	Sound  bool
	Note   string
}

// RecordedProofs returns every proof on the record, in event order, each with its script's re-run
// and its own sentence's verdict.
//
// The OUTPUT and the SCRIPT BODY are deliberately not here: they live on disk under
// <run>/proofs/<sha256>/ and can be large. The assembler reads them from the artifact so the
// report shows the exact bytes that ran, rather than a copy the record made of them.
func RecordedProofs(run Run) ([]Proof, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	ledger := anchorChecksOf(m.Events, m.At)
	reruns := scriptRerunsOf(m.Events)
	var out []Proof
	for _, e := range m.Events {
		pf, isProof := recordpb.BodyAs[*recordpb.Proof](e)
		if !isProof {
			continue
		}
		v := proofVerdictOf(ledger.at(pf.GetProofId()), reruns)
		p := Proof{
			Label: pf.GetProofId(),
			// Written as `sha256`, joined on as `proof_sha` — one fact, one field now.
			SHA:    pf.GetProofSha(),
			Basis:  pf.GetProofBasis(),
			Script: pf.GetScript(),
			Exit:   int(pf.GetExit()),
			Cites:  pf.GetCites(),
			// --reason lands on Proof.text, the message's only prose channel — the same
			// flag/field split recordpb/required.go records for Verify.text.
			Reason:    pf.GetText(),
			Drift:     pf.GetDrift(),
			Soundness: v.soundness,
		}
		if r := v.rerun; r != nil {
			w := m.At.Of(r.ev)
			p.Rerun = &ProofRerun{SeatID: r.ev.GetSeatId(), Epoch: w.Epoch, Sitting: w.Sitting, Rerun: r.rerun,
				Recorded: r.recorded, Observed: r.observed}
		}
		for _, c := range v.judgements {
			r, _ := recordpb.BodyAs[*recordpb.Reproduce](c.ev)
			p.Judgements = append(p.Judgements, ProofJudgement{SeatID: c.seat, Epoch: m.At.Of(c.ev).Epoch, Sound: c.holds, Note: r.GetNote()})
		}
		out = append(out, p)
	}
	return out, nil
}

// ProofAt is the recorded proof at one anchor, with its script's re-run and its sentence's verdict —
// what `reproduce --anchor` acts on. An anchor no proof carries is refused by name.
func ProofAt(run Run, anchorID string) (Proof, error) {
	proofs, err := RecordedProofs(run)
	if err != nil {
		return Proof{}, err
	}
	for _, p := range proofs {
		if p.Label == anchorID {
			return p, nil
		}
	}
	return Proof{}, fmt.Errorf("--anchor names %s, which no proof on this record carries (%d proof(s)) — it takes the id inside a proof's token in the report; `show evidence` lists every proof by its anchor", anchorID, len(proofs))
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

// contradictionsOf is every contradiction red recorded, in the order the record opened them, each
// with the finding that answered it ("" while it is open). ids is evs's events.id, aligned with it.
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
//   - A FINDING RAISES the open contradiction it names, by reference (`--about-kind contradiction
//     --about <id>`), or every open contradiction whose claim its quote contains. Either covers
//     every reading joined so far. A reading of the claim after the finding opens a NEW
//     contradiction with its own id: the raise was of what had been read.
//   - A GAP ANSWERS what a finding raised: a mint whose found_by credits the raising finding.
//     A source that contradicts or is silent on a claim is a citation defect, and citation classes
//     are material whatever the lens judges (gblock, 2026-10-19: the class default wins), so the
//     finding alone — which holds no gate and obliges nobody — does not discharge it. The gap's
//     class then decides its materiality, as for every gap.
//
// THE REFERENCE IS THE ANSWER THAT SURVIVES AN EDIT. A quote must be the report's text at the
// write, so once blue edits the contradicted sentence away no finding can quote it — the case that
// held a PASS for the rest of a run whose report blue had FIXED (plans/v2/01). The id names the
// reading, which an edit does not move.
//
// QUOTING THE CLAIM IS THE MATCHER'S QUESTION, NOT STRING EQUALITY. Both quotes are placed in the
// report by anchortext, which reads whitespace runs as one separator and drops trailing
// punctuation, so two spellings of one sentence place at one span — and a finding that quotes the
// contradicted clause with the rest of its sentence is about that clause. A finding on PART of
// the claim, or on another sentence, answers nothing: it does not quote what the source read.
//
// THE MATCH IS ANY LENS'S. A stricter join (same seat) would refuse a contradiction one lens found
// and another raised, which is the collaboration the lens roles exist for.
//
// A reading that names a citation carries its anchor, and the anchor's standing answers it too —
// a retire, or a backing check after an edit (unansweredContradictions). That answer is the
// citation's state, not an act naming the contradiction, so it is not this fold's: a finding may
// still raise a contradiction its citation's standing answers.
func contradictionsOf(evs []*Event, ids []int64) []contradiction {
	var out []*contradiction
	open := map[string]*contradiction{} // claim -> its open contradiction
	raise := func(c *contradiction, by string) {
		c.raisedBy = by
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
			c := &contradiction{id: ids[i], claim: claim, reader: e.GetSeatId(), anchor: b.GetAnchor(), since: ids[i]}
			open[claim] = c
			out = append(out, c)
		case *recordpb.Finding:
			if b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_CONTRADICTION {
				for _, c := range out {
					if c.raisedBy == "" && strconv.FormatInt(c.id, 10) == strings.TrimSpace(b.GetAboutRef()) {
						raise(c, b.GetId())
					}
				}
			}
			if loc := strings.TrimSpace(b.GetLocation()); loc != "" {
				for _, c := range out {
					if start, _ := anchortext.LocateSpan(loc, c.claim); c.raisedBy == "" && start >= 0 {
						raise(c, b.GetId())
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

// requireOpenContradiction refuses a finding's reference to a contradiction that is not open: a
// ref that is not an id, one no reading opened, or one a finding already answered — that refusal
// names the answering finding. It runs on replay as well: the reading is written before the
// finding that answers it, so a translated reference resolves and a mistranslated one is loud.
func requireOpenContradiction(q recordsql.Querier, ref string) error {
	ref = strings.TrimSpace(ref)
	if _, err := strconv.ParseInt(ref, 10, 64); err != nil {
		return fmt.Errorf("--about %q is not a contradiction's id: it is the number `show evidence` lists as the `id` of each of its unanswered_contradictions", ref)
	}
	evs, win, err := eventsOfAt(q, recordpb.EventType_EVENT_TYPE_VERIFY, recordpb.EventType_EVENT_TYPE_FINDING, recordpb.EventType_EVENT_TYPE_MINT)
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
// A corroboration's reading checked no citation (its anchor is ""), so only a finding answers it.
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
// it), the claim, the seat whose reading last joined it — the lens whose act raises it — and where
// that reading sits, which is when the lens came to owe it. anchor is the citation the opening
// reading checked, "" for a corroboration; raisedBy is the id of the finding that raised it, and
// answeredBy the id of the gap minted from that finding.
type contradiction struct {
	id                           int64
	claim, reader                string
	anchor, raisedBy, answeredBy string
	since                        int64
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
