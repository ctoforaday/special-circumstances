package blue

import (
	"cmp"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// retire: the ONLY way substance leaves the report.
//
// "Never subtract substance" was a PROSE-level rule, and it was doing two jobs at
// once. It stopped run 3's real failure — blue quietly dropping content under
// repair pressure — but it also forbade rewriting, so the report could only grow:
// 1178 to 1668 lines in a single run, and every audit seat paid to re-read all of
// it, every epoch.
//
// Splitting the two jobs: prose may now be compacted, merged and reorganized
// freely, because a claim can only LEAVE through this verb. (Merging two cited
// sentences carries both anchors, and the claim count is of attached citations, not
// sentences — so a merge moves the count not at all.) Deletion stops being
// something a rule forbids and becomes something the record shows — with what was
// removed, why, and what (if anything) replaced it.
//
// That is strictly stronger than the prose rule it replaces. The old rule could
// be broken silently by an edit; here `edit` refuses a replacement that drops an
// anchor, so a counted claim leaves only where a retire names the anchors it takes.
//
// AND IT IS HOW AN ANCHOR LEAVES. An edit may carry an anchor but never drop one,
// so cutting an anchored sentence leaves the anchor BARE — no prose before it in
// its sentence, backing nothing. Nothing could take that anchor out, so it stood forever: an
// orphan `[^N]` with a live bibliography entry, an empty `- ?` bullet. The retire
// now names the bare anchors the claim's removal left (computed here, from the
// record and the report, never typed by a seat), and replay takes them out at this
// event. Edit still PERFORMS the removal of prose; retire EXPLAINS it and closes it.
func newRetire() *cobra.Command {
	named := flags.IDList(flags.AnchorID)
	c := seat.Prose(seat.New("retire", retireResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		if _, err := s.Run(); err != nil {
			return nil, err
		}
		// --reason is the prose channel (Prose provides it); it lands under `reason`,
		// which validate requires — substance leaves the report only with its reason.
		why, err := seat.Reason(cmd)
		body := &recordpb.Retire{
			Claim:        proto.String(seat.Str(cmd, flags.Quote)),
			Reason:       proto.String(why),
			SupersededBy: proto.String(seat.Str(cmd, flags.New)),
			Answers:      proto.String(seat.Str(cmd, flags.Answers)),
		}
		if err != nil {
			return nil, err
		}

		// THE REMOVAL IS CHECKED, NOT TAKEN ON TRUST.
		//
		// This verb recorded whatever it was told. Nothing confirmed the claim had ever
		// been in the report, and nothing confirmed it had left — so "substance leaves only
		// through the retire verb", the rule this comment above calls "strictly stronger
		// than the prose rule it replaces", rested on the seat's word.
		//
		// A PHANTOM RETIRE IS WORSE THAN USELESS. The changelog lists every retirement as a
		// claim that left the report: one naming a claim that was never there is a removal
		// the record asserts and the report never had.
		claim := seat.Str(cmd, flags.Quote)
		// A CITATION RETIRED, ITS SENTENCE KEPT: --anchor alone names it, and the tool reads the sentence.
		if !cmd.Flags().Changed(flags.Quote) && len(named.Value()) > 0 {
			return retireCitation(s, body, named.Value())
		}
		// A BLANK CLAIM IS A MISSING ONE, refused before the guard reads it as text: every report
		// contains the empty string, so universe-m29's `retire --quote ""` was told `"" is still in
		// the report` and sent to edit out a claim it never named.
		if strings.TrimSpace(claim) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --quote: the claim being removed, verbatim as it stood in the report before the edit that took it out — or, to retire a citation and keep its sentence, --anchor alone naming the citation")
		}
		// ASKED OF THE REPORT AND THE EDITS THE RETIRE LANDS ON. Blue's own edits and a lens's
		// anchor commit in parallel with it, and a retire checked against the report before them can
		// name an anchor another write has since put back into prose, or explain the removal of a
		// claim that has re-entered the report (#1252).
		var stayed []keptAnchor
		guard := func(l retireLanding) error {
			basis := record.RemovalAsserted
			defer func() { body.RemovalBasis = proto.String(basis) }()
			md, rerr := l.Report()
			if rerr != nil {
				// AN UNREADABLE REPORT REFUSES THE RETIRE: a removal recorded because the report
				// could not be asked whether the claim left is a removal nobody checked.
				return feov.Wrap(feov.Internal, rerr, "the report could not be read from the record to establish whether %q has left it, so nothing is recorded: %v", claim, rerr)
			}
			if strings.Contains(md, claim) {
				return feov.Errorf(feov.Conflict,
					"%q is still in the report. Retiring is how a removal is EXPLAINED, not how it is performed — remove the text with `edit` first, then retire the claim to say why it went and what replaced it",
					claim)
			}
			// Absent now. Whether it was ever THERE is a different question, and the record
			// can answer it: a claim removed by a recorded edit appears in that edit's old
			// span, read as a seat quotes it (cutFrom). Absent from both is a retirement of
			// something nobody can show existed.
			spans, err := l.EditSpans()
			if err != nil {
				return err
			}
			if slices.ContainsFunc(spans, func(sp record.EditSpan) bool { return cutFrom(quoteCore(claim), sp) }) {
				basis = record.RemovalVerified
			}
			// THE ANCHORS THAT EXIT WITH THE CLAIM: those that stood on it in the edit that took it
			// out and stand bare now go with the reason; one carried onto surviving prose goes when named.
			exiting := anchorsExiting(md, claim, spans)
			if len(exiting) == 0 && len(named.Value()) == 0 {
				if err := fragmentRefusal(md, claim, spans); err != nil {
					return err
				}
			}
			// A MARKER INSIDE A SENTENCE IS NEVER BARE: when the claim was a clause, the rest of the
			// sentence still stands before its marker. The seat names it, and the tool checks the name
			// against the edit that left the marker alone. B9's G4 was two retires that took nothing
			// and said nothing, over a corroboration marker on a mid-sentence clause.
			left := leftByTheCut(claim, spans)
			for _, id := range named.Value() {
				if err := namedExits(md, claim, id, left, spans); err != nil {
					return err
				}
				if !slices.Contains(exiting, id) {
					exiting = append(exiting, id)
				}
			}
			body.Anchors = exiting
			stayed = stayedBehind(md, left, exiting)
			return nil
		}

		if _, err := record.Append(s.Identity(), body, record.CheckOnReport(func(l *record.Landing) error { return guard(l) })); err != nil {
			return nil, err
		}
		return retireResult{Claim: seat.Str(cmd, flags.Quote), Anchors: body.Anchors, Stayed: stayed}, nil
	}))

	// NOT DescQuote. Every other --quote is matched against the report; this one must be ABSENT
	// from it — the edit that took the claim out comes first, and the retire is refused while the
	// text still stands. What it is matched against is that edit's old span.
	//
	// SUPPLIED, NOT TYPED, WHEN A CITATION IS RETIRED AND ITS SENTENCE KEPT: the record's claim is then
	// the sentence the citation stood on, which the tool reads off the report.
	seat.Supplies(c, "claim", "the sentence the retired citation stood on, read off the report, when --anchor alone retires a citation and keeps its sentence")
	flags.Text(c, flags.Quote, "the claim being removed, verbatim as it stood in the report BEFORE the edit that took it out — a sub-span of what that edit cut. It must be ABSENT from the report now (edit it out first, then retire; the retire is refused while the text still stands). Found inside a recorded edit's old span, anchors read past, the removal is recorded as verified; found in none, the retire is still recorded, with its removal basis recorded as asserted rather than verified — a removal the record cannot show")
	flags.Text(c, flags.New, "the claim that replaces it, when one does")
	seat.RequiredUnless(c, flags.Quote, "unless --anchor alone retires a citation")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "with --anchor alone: the gap id this citation retire responds to — the provenance join key an edit carries, so the gap's change log lists the retire beside what red asked for; omit only for a retire that answers no gap. A claim retire takes none: the edit that cut the claim names the gap it answers")
	c.Flags().Var(named, flags.Anchor, "with --quote: an anchor id of any kind that stood on this claim and still stands on prose after the edit, when it backed only this claim (comma-separated) — it leaves with the claim. Alone, without --quote: the one citation to retire while its sentence stays in the report word for word")
	return c
}

// retireCitation retires one citation and keeps its sentence: the anchor leaves the report, and the
// sentence it stood on stays word for word, backed by whatever else backs it — another citation, or
// nothing, when it is prose a lens audits like any other.
//
// IT IS HOW BLUE ANSWERS A CITATION NOBODY CAN READ, AND HOW A CITED SENTENCE LEAVES (gblock,
// 2026-10-25). An edit may carry an anchor but never drop one, and the claim retire takes the
// anchors of a sentence an edit cut — so a citation whose source could not be read could leave only
// with its sentence, and re-citing a readable source left the dead one standing beside it (b6's two
// dead anchors on a sentence that stayed). And an edit may not leave a citation bare
// (refuseBaredCitations), so a cited sentence that is to go loses its citation here first, and is
// then prose an edit may cut.
//
// THE TOOL READS THE SENTENCE. The claim the record holds is the sentence holding the anchor in the
// report the retire lands on (bluedoc.SentenceAround, as `verify` reads it), and claim_kept says it
// stayed, so the report's retirements name the citation that left and never list the sentence as a
// claim that did.
func retireCitation(s seat.Context, body *recordpb.Retire, named []string) (seat.Result, error) {
	if len(named) != 1 {
		return nil, feov.Errorf(feov.Validation, "--anchor alone retires one citation and keeps its sentence, and this names %d: %s. Retire each citation in its own retire, or quote the claim with --quote to retire the claim with the anchors it held", len(named), strings.Join(named, ", "))
	}
	id := named[0]
	if k := anchor.Kind(id); k != "citation" {
		return nil, feov.Errorf(feov.Validation, "--anchor alone retires a citation and keeps its sentence, and %s is a %s anchor — it leaves the report with its claim: edit the claim out, then retire it with --quote", id, k)
	}
	tok := anchor.Token(id)
	guard := func(l *record.Landing) error {
		md, err := l.Report()
		if err != nil {
			return feov.Wrap(feov.Internal, err, "the report could not be read from the record to find the sentence citation %s stands on, so nothing is recorded: %v", id, err)
		}
		sentence, ok := bluedoc.SentenceAround(md, tok)
		switch {
		case !ok:
			return feov.Errorf(feov.Conflict, "the report holds no %s — citation %s is not on a sentence of the report, so there is no citation to retire from one", tok, id)
		case bareAnchors(md)[id]:
			return feov.Errorf(feov.Conflict, "citation %s stands bare in the report — the sentence it backed was cut, so there is no sentence to keep. It leaves with the claim it backed: retire that claim with --quote", id)
		}
		body.Claim = proto.String(sentence)
		body.Anchors = []string{id}
		body.ClaimKept = proto.Bool(true)
		return nil
	}
	if _, err := record.Append(s.Identity(), body, record.CheckOnReport(guard)); err != nil {
		return nil, err
	}
	return retireResult{Claim: body.GetClaim(), Anchors: body.Anchors, ClaimKept: true, Answers: body.GetAnswers()}, nil
}

// leftByTheCut is every marker that stood on this claim in an edit that took it out (heldAt), with
// the position of the last such edit.
func leftByTheCut(claim string, spans []record.EditSpan) map[string]int {
	out := map[string]int{}
	for i, sp := range spans {
		held, _ := heldAt(claim, sp)
		for _, id := range held {
			out[id] = i
		}
	}
	return out
}

// heldAt is what an edit that took claim out of its old span carried of it: the anchors that stood on
// claim in that span, in the order the edit left them, bare or carried onto surviving prose — an edit
// may rewrite the sentences around them as freely as it may cut them. claim names them when it is the whole
// span or a run of the span's whole sentences (anchor.Sentences): a sentence is what an anchor marks,
// so one edit cutting two claims leaves each its own anchors to leave with. A claim lying inside the
// span as neither is a fragment of what the edit cut, and in is the text it was cut from — the
// sentences holding it, or the whole span where it crosses one.
func heldAt(claim string, sp record.EditSpan) (held []string, in []string) {
	want := quoteCore(claim)
	cut := claimcount.ProtectedAnchorIDs(sp.New)
	if len(cut) == 0 || !cutFrom(want, sp) {
		return nil, nil
	}
	if quoteCore(sp.Old) == want {
		return cut, nil
	}
	sents := anchor.Sentences(sp.Old)
	for i := range sents {
		for j := i; j < len(sents); j++ {
			seg := sp.Old[sents[i][0]:sents[j][1]]
			if quoteCore(seg) != want {
				continue
			}
			held = []string{}
			for _, id := range cut {
				if strings.Contains(seg, anchor.Token(id)) {
					held = append(held, id)
				}
			}
			return held, nil
		}
	}
	for _, s := range sents {
		if seg := sp.Old[s[0]:s[1]]; strings.Contains(claimcount.StripAnchors(seg), want) {
			in = append(in, seg)
		}
	}
	if len(in) == 0 {
		in = []string{sp.Old}
	}
	return nil, in
}

// namedExits checks a marker the seat named against what the cut left: it stands in the report, it
// stood on this claim in an edit that took the claim out, and no LATER edit put it back into prose —
// a marker re-attached to text backs that text now, and leaves only with it.
func namedExits(md, claim, id string, left map[string]int, spans []record.EditSpan) error {
	tok := anchor.Token(id)
	if !strings.Contains(md, tok) {
		return feov.Errorf(feov.Validation, "--anchor %s: the report holds no %s — name a marker that stands where the claim was", id, tok)
	}
	i, ok := left[id]
	if !ok {
		return feov.Errorf(feov.Validation, "--anchor %s: %s", id, notHeld(claim, tok, spans))
	}
	for _, sp := range spans[i+1:] {
		if strings.Contains(sp.New, tok) && !slices.Contains(claimcount.BareAnchorIDs(sp.New), id) {
			return feov.Errorf(feov.Validation, "--anchor %s: a later edit put %s back into prose — it backs that text now, and leaves only with it", id, tok)
		}
	}
	return nil
}

// notHeld says why the anchor tok is not one this claim's cut left (heldAt), from the edits: where
// tok stood when an edit took this claim out with the text around it, that the claim is a fragment
// of a sentence such an edit cut, or that no recorded edit took the claim out.
func notHeld(claim, tok string, spans []record.EditSpan) string {
	var frag []string
	for _, sp := range spans {
		held, in := heldAt(claim, sp)
		if held == nil {
			for _, seg := range in {
				frag = append(frag, quoted(flatClaim(seg)))
			}
			continue
		}
		on := "on no sentence of it"
		for _, s := range anchor.Sentences(sp.Old) {
			if seg := sp.Old[s[0]:s[1]]; strings.Contains(seg, tok) {
				on = "on " + quoted(flatClaim(seg))
			}
		}
		these := "no anchor stood on this claim there"
		if len(held) > 0 {
			these = "the anchors that stood on this claim there are " + strings.Join(held, ", ")
		}
		return fmt.Sprintf("%s did not stand on this claim when an edit took it out with the text around it — it stood %s, and %s. "+
			"Name an anchor this claim held, or retire the claim %s stood on", tok, on, these, tok)
	}
	if len(frag) > 0 {
		return fmt.Sprintf("this claim is part of a sentence an edit cut, and a retire takes the anchors of whole sentences of what the edit cut. Quote the sentence as it stood: %s", strings.Join(frag, " or "))
	}
	return fmt.Sprintf("no recorded edit took this claim out, so %s did not stand on it there — `edit` the claim out first (the edit carries its anchors), then retire it naming the one that backed only it", tok)
}

// stayedBehind is every marker the cut left that still stands and is not leaving, said — so a
// retire that took nothing cannot read the same as one that took everything.
func stayedBehind(md string, left map[string]int, exiting []string) []keptAnchor {
	var ids []string
	for id := range left {
		if !slices.Contains(exiting, id) && strings.Contains(md, anchor.Token(id)) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	out := make([]keptAnchor, 0, len(ids))
	for _, id := range ids {
		out = append(out, keptAnchor{ID: id, Why: "prose stands before it in its sentence, so it is not counted bare — if it backed only the claim that left, name it with --anchor"})
	}
	return out
}

// anchorsExiting names the anchors that leave the report with a retired claim unnamed: those that
// stood on it in the edit that took it out (heldAt) and stand BARE in the report now — present, with
// no prose anywhere they appear. One carried onto surviving prose leaves only when the seat names it
// (namedExits); stayedBehind says it stayed.
//
// THE QUOTE IS THE WHOLE OF WHAT LEFT, OR WHOLE SENTENCES OF IT (heldAt). The claim is compared with
// the edit's entire old span and with each run of its whole sentences, never searched for inside
// them: a fragment ("sky") would otherwise sweep the anchor of the sentence it came from, record
// removal_basis=verified, and put the fragment in the report's withdrawn claims as if it were what
// left (fragmentRefusal refuses it instead). A sentence is the unit an anchor marks, so an edit that
// cuts two claims leaves each its own anchors: m25 cut two sentences carrying three finding anchors
// in one edit (ev247), retired each with its own reason (ev258, ev259), and with the whole-span
// comparison alone both took nothing and the three anchors stood bare until a third retire quoted
// the pair (ev367). The comparison is across the anchor layer, with surrounding whitespace, emphasis
// markers and trailing punctuation trimmed, because that is how a seat quotes it: the edit's old span
// carries the anchor token the seat copied from `show report`, and the retire's --quote often does
// not.
func anchorsExiting(md, claim string, spans []record.EditSpan) []string {
	bare := bareAnchors(md)
	var out []string
	for _, sp := range spans {
		held, _ := heldAt(claim, sp)
		for _, id := range held {
			if bare[id] && !slices.Contains(out, id) {
				out = append(out, id)
			}
		}
	}
	return out
}

// bareAnchors is the set of anchors standing bare in md.
func bareAnchors(md string) map[string]bool {
	bare := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(md) {
		bare[id] = true
	}
	return bare
}

// fragmentRefusal refuses a retire whose claim is a fragment of a sentence an edit cut (heldAt's in),
// where no edit took the claim out whole and that sentence's anchors stand bare: the retire would
// explain a removal and leave them standing, and the claim that left is the sentence.
func fragmentRefusal(md, claim string, spans []record.EditSpan) error {
	bare := bareAnchors(md)
	var ids, from []string
	for _, sp := range spans {
		held, in := heldAt(claim, sp)
		if held != nil || quoteCore(sp.Old) == quoteCore(claim) {
			return nil
		}
		for _, seg := range in {
			n := len(ids)
			for _, id := range anchor.IDs(seg) {
				if bare[id] && !slices.Contains(ids, id) {
					ids = append(ids, id)
				}
			}
			if s := quoted(flatClaim(seg)); len(ids) > n && !slices.Contains(from, s) {
				from = append(from, s)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return feov.Errorf(feov.Validation, "%s is part of what an edit cut, and the anchors it left there stand bare: %s. "+
		"A retire takes the anchors of the claim it quotes — the whole of what the edit cut, or whole sentences of it — and a fragment of a sentence takes none. "+
		"Quote the sentence it was part of as it stood: %s", quoted(claim), strings.Join(ids, ", "), strings.Join(from, " or "))
}

// cutFrom reports whether an edit's old span held the claim whose quoteCore is want, read across the
// anchor layer as a seat quotes it — an anchor flush before a full stop splits the raw text a seat
// copies. It is the one test of "an edit took this claim out": heldAt reads no edit it fails, and the
// retire's removal basis is verified by it.
func cutFrom(want string, sp record.EditSpan) bool {
	return want != "" && strings.Contains(claimcount.StripAnchors(sp.Old), want)
}

// flatClaim is s as a seat quotes it: anchors out, whitespace runs read as one space.
func flatClaim(s string) string {
	return strings.Join(strings.Fields(claimcount.StripAnchors(s)), " ")
}

// quoted is s between double quotes, unescaped, so a seat copying it back gets the report's bytes.
func quoted(s string) string { return `"` + s + `"` }

// quoteCore is a quote or span reduced to what it says: anchors out, the list marker or block-quote
// marker it opens with out — those lie outside every sentence (anchor.Sentences), and an edit that
// quotes a bullet with its "- " cuts the same claim a retire names without it — then surrounding
// whitespace and emphasis markers and trailing punctuation trimmed until nothing more comes off.
// Internal text is untouched — "the sky" and "sky" stay different claims.
func quoteCore(s string) string {
	s = claimcount.StripAnchors(s)
	if sp := anchor.Sentences(s); len(sp) > 0 {
		s = s[sp[0][0]:]
	}
	for {
		t := strings.TrimRight(strings.Trim(s, " \t\n*_"), anchortext.TrailingPunct)
		if t == s {
			return s
		}
		s = t
	}
}

// keptAnchor is an anchor that stood on the claim and that the retire left in the report, and why.
type keptAnchor struct {
	ID  string `json:"id"`
	Why string `json:"why"`
}

type retireResult struct {
	Claim   string       `json:"claim"`
	Anchors []string     `json:"anchors,omitempty"`
	Stayed  []keptAnchor `json:"stayed,omitempty"`
	// ClaimKept is a citation retired from a sentence that stays: Claim is that sentence.
	ClaimKept bool `json:"claim_kept,omitempty"`
	// Answers is the gap a citation retire answers, as recorded; empty when it answers none.
	Answers string `json:"answers,omitempty"`
}

func (r retireResult) Human() string {
	if r.ClaimKept {
		answering := ""
		if r.Answers != "" {
			answering = ", answering gap " + r.Answers
		}
		return "retired citation " + strings.Join(r.Anchors, ", ") + answering + "; its sentence stays in the report word for word, backed by whatever else backs it: " + r.Claim
	}
	out := "retired: " + r.Claim + "\nanchors out with it: " + cmp.Or(strings.Join(r.Anchors, ", "), "none")
	for _, k := range r.Stayed {
		out += "\nanchor stayed: " + k.ID + " — " + k.Why
	}
	return out
}

// retireLanding is what the retire guard reads of the record it lands on.
type retireLanding interface {
	Report() (string, error)
	EditSpans() ([]record.EditSpan, error)
}
