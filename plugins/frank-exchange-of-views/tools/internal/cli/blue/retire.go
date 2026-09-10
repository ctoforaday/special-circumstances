package blue

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
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
// sentences carries both anchors, and claim_count counts attached citations, not
// sentences — so a merge moves the count not at all.) Deletion stops being
// something a rule forbids and becomes something the record shows — with what was
// removed, why, and what (if anything) replaced it.
//
// That is strictly stronger than the prose rule it replaces. The old rule could
// be broken silently by an edit; this one leaves a hole in the record that
// capture detects, because claim_count falling further than the retire events
// account for is arithmetic, not judgement.
//
// AND IT IS HOW AN ANCHOR LEAVES. An edit may carry an anchor but never drop one,
// so cutting an anchored sentence leaves the anchor BARE — alone in its segment,
// backing nothing. Nothing could take that anchor out, so it stood forever: an
// orphan `[^N]` with a live bibliography entry, an empty `- ?` bullet. The retire
// now names the bare anchors the claim's removal left (computed here, from the
// record and the report, never typed by a seat), and replay takes them out at this
// event. Edit still PERFORMS the removal of prose; retire EXPLAINS it and closes it.
func newRetire() *cobra.Command {
	c := seat.Prose(seat.New("retire", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// --reason is the prose channel (Prose provides it); it lands under `reason`,
		// which validate requires — substance leaves the report only with its reason.
		why, err := seat.Reason(cmd)
		body := &recordpb.Retire{
			Claim:        proto.String(seat.Str(cmd, flags.Quote)),
			Reason:       proto.String(why),
			SupersededBy: proto.String(seat.Str(cmd, flags.New)),
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
		// A PHANTOM RETIRE IS WORSE THAN USELESS. The scorecard's additive-integrity
		// detector computes unrecorded_claim_loss as (drop in claim_count) MINUS (retire
		// events): a retirement of a claim that was never there subtracts from the
		// accounted side and CANCELS REAL LOSS, blinding the one detector built to catch
		// silent deletion.
		claim := seat.Str(cmd, flags.Quote)
		basis := record.RemovalAsserted
		var kept []keptAnchor
		if md, rerr := reportproj.RenderFromRecord(run); rerr == nil {
			if strings.Contains(md, claim) {
				return nil, feov.Errorf(feov.Conflict,
					"blue retire: %q is still in the report. Retiring is how a removal is EXPLAINED, not how it is performed — remove the text with `blue edit` first, then retire the claim to say why it went and what replaced it",
					claim)
			}
			// Absent now. Whether it was ever THERE is a different question, and the record
			// can answer it: a claim removed by a recorded edit appears in that edit's old
			// span. Absent from both is a retirement of something nobody can show existed.
			if record.ClaimAppearsInAnEdit(run, claim) {
				basis = record.RemovalVerified
			}
			// THE ANCHORS THAT EXIT WITH THE CLAIM. An edit that took this claim out and left
			// only anchors in its place left those anchors bare; they go now, with the reason.
			spans, err := record.EditSpans(run)
			if err != nil {
				return nil, err
			}
			exiting := anchorsExiting(md, claim, spans)
			if len(exiting) > 0 {
				basis = record.RemovalVerified // the edit that left them bare shows the claim leaving
			}
			if body.Anchors, kept, err = takeable(run, exiting); err != nil {
				return nil, err
			}
		}
		body.RemovalBasis = proto.String(basis)

		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return retireResult{Claim: seat.Str(cmd, flags.Quote), Anchors: body.Anchors, Kept: kept}, nil
	}))

	c.Flags().String(flags.Quote, "", flags.DescQuote+" — the claim being removed, as it stood before you edited it out")
	c.Flags().String(flags.New, "", "the claim that replaces it, when one does — the same --quote/--new pair `edit` and `mint` take")
	return c
}

// anchorsExiting names the anchors that leave the report with a retired claim: those an edit
// left BARE when it took the claim out. Validated against the report as it stands — an anchor
// must be present AND bare (no prose anywhere it appears), so an anchor a later edit carried
// on into surviving prose is never named. Everything else about the retire is unchanged when
// this is empty.
//
// THE QUOTE MUST BE THE WHOLE OF WHAT LEFT. The claim is compared with the edit's entire old
// span, not searched for inside it: a fragment ("sky") would otherwise sweep the anchor of the
// sentence it came from, record removal_basis=verified, and put the fragment in the report's
// withdrawn claims as if it were what left. The comparison is across the anchor layer, with
// surrounding whitespace, emphasis markers and trailing punctuation trimmed, because that is how
// a seat quotes it: the edit's old span carries the anchor token the seat copied from `show
// report`, and the retire's --quote often does not.
func anchorsExiting(md, claim string, spans []record.EditSpan) []string {
	want := quoteCore(claim)
	if want == "" {
		return nil
	}
	bare := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(md) {
		bare[id] = true
	}
	seen := map[string]bool{}
	var out []string
	for _, sp := range spans {
		ids := claimcount.ProtectedAnchorIDs(sp.New)
		if len(ids) == 0 || claimcount.HasProse(sp.New) || quoteCore(sp.Old) != want {
			continue
		}
		for _, id := range ids {
			if bare[id] && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

// quoteCore is a quote or span reduced to what it says: anchors out, then surrounding
// whitespace and emphasis markers and trailing punctuation trimmed until nothing more comes off.
// Internal text is untouched — "the sky" and "sky" stay different claims.
func quoteCore(s string) string {
	s = claimcount.StripAnchors(s)
	for {
		t := strings.TrimRight(strings.Trim(s, " \t\n*_"), anchortext.TrailingPunct)
		if t == s {
			return s
		}
		s = t
	}
}

// takeable splits the bare anchors exiting with a claim into those this retire takes out and
// those it must leave, with why.
//
// OWNERSHIP. Cite and proof anchors are blue's; a finding marker is red's, and leaves only once
// red's lifecycle has closed on it (record.FindingMarkerHold).
//
// A run whose database predates the retire_anchors table is not handled here: the generic body
// walk reads and writes every list table of a Retire, so such a run cannot hold ANY retire this
// binary reads back, whatever this verb names. The write refuses with the cause
// (recordsql.olderSchema) rather than recording an event the next read would choke on.
func takeable(run record.Run, exiting []string) (take []string, kept []keptAnchor, err error) {
	for _, id := range exiting {
		if strings.HasPrefix(id, "f-") {
			why, err := record.FindingMarkerHold(run, id)
			if err != nil {
				return nil, nil, err
			}
			if why != "" {
				kept = append(kept, keptAnchor{ID: id, Why: why})
				continue
			}
		}
		take = append(take, id)
	}
	return take, kept, nil
}

// keptAnchor is a bare anchor the retire left in the report, and why.
type keptAnchor struct {
	ID  string `json:"id"`
	Why string `json:"why"`
}

type retireResult struct {
	Claim   string       `json:"claim"`
	Anchors []string     `json:"anchors,omitempty"`
	Kept    []keptAnchor `json:"kept,omitempty"`
}

func (r retireResult) Human() string {
	out := "retired: " + r.Claim
	if len(r.Anchors) > 0 {
		out += "\nanchors out with it: " + strings.Join(r.Anchors, ", ")
	}
	for _, k := range r.Kept {
		out += "\nanchor kept: " + k.ID + " — " + k.Why
	}
	return out
}
