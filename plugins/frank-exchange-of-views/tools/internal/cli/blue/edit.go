package blue

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// edit: the ONLY write path to blue/report.md for a response seat.
//
// report.md is read-only to the response seat (the lockdown hook denies its raw Edit/Write
// and Bash writes); it changes the report EXCLUSIVELY through this verb, which replaces the
// exact current span --old with --new while PRESERVING both immortal anchor classes — red's
// invisible finding-markers ("<!--fx:...-->") AND blue's tool-managed citation anchors
// ("<!--cite:...-->"). A span that would drop or split an anchor of EITHER class is REJECTED
// (edit around it). Each applied edit appends a `blue_edit` event — an append-only diff-stack
// that replays onto the round-0 report to equal the current head.
//
// PROVENANCE: --answers names the gap this edit responds to. It is validated against the
// board like every other reference (refs.go), and it is what makes `required_fix` and the
// change blue actually made JOINABLE — the pair every #267 measurement reads. Naming a real
// gap id in --reason while leaving --answers empty is REFUSED: the convention it replaces
// held 19 times in 26 and that is exactly the reliability a key may not have.
//
// Crash-safety is EVENT-FIRST: the event is appended only AFTER validation succeeds (so a
// mis-quote never lands a phantom op), then the write is applied; a crash between leaves the
// event durable and the write reconciled idempotently on retry — no wedge, no phantom op.
func newEdit() *cobra.Command {
	c := seat.Prose(seat.New("edit", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		reason, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(reason) == "" {
			return nil, fmt.Errorf("blue edit requires --reason: why this change (the argument red re-audits against)")
		}
		oldStr := seat.Str(cmd, flags.Quote)
		newStr := seat.Str(cmd, flags.New)
		if oldStr == "" {
			return nil, fmt.Errorf("blue edit requires --quote: the EXACT current span to replace (matched across the invisible marker layer, like the Edit tool)")
		}
		if oldStr == newStr {
			return nil, fmt.Errorf("blue edit: --old and --new are identical — no change to make")
		}
		key := seat.Str(cmd, flags.Key)

		// Crash-retry: a committed blue_edit for this key means the op is already on the
		// stack — reconcile the write idempotently, do NOT append a second op.
		prior, err := record.ExistingBlueEditByKey(run, s.SeatID, key)
		if err != nil {
			return nil, err
		}
		if prior {
			// The op is already on the diff-stack, so the render already reflects it — there is
			// nothing to re-apply. Under report-as-record the edit IS the event; no file to reconcile.
			return editResult{Idempotent: true}, nil
		}

		// FRESH: validate against a consistent snapshot BEFORE committing the event, so a
		// mis-quote or a marker-spanning edit never lands a phantom stack op. The snapshot is the
		// render of the record so far — there is no file.
		peek, err := reportproj.RenderFromRecord(run)
		if err != nil {
			return nil, err
		}
		planned, err := validateEdit(peek, oldStr, newStr)
		if err != nil {
			return nil, err
		}

		// Event-first: commit the diff-stack op, then apply the write.
		body := &recordpb.BlueEdit{
			EditKey: proto.String(seat.Str(cmd, flags.Key)),
			Answers: proto.String(seat.Str(cmd, flags.Answers)),
			Old:     proto.String(oldStr),
			New:     proto.String(newStr),
			Text:    proto.String(reason),
			// WHAT THIS EDIT REOPENED. An anchor is never lost — that is enforced above and
			// holds — but one that SURVIVES onto rewritten prose is a citation backing a
			// sentence nobody read, and nothing said so. Text moves through this verb alone,
			// so the no-loss proof and the requires-review mark belong on this one channel.
			//
			// Computed from the SNAPSHOT the validation used, not from a re-read: the write
			// has not happened yet, and a second read could see a different document.
			Reopened: bluedoc.ReopenedAnchors(peek, planned),
		}
		// ESTOPPEL, RECORDED BY THE TOOL COMPARING BYTES (#267 stage 4).
		//
		// If blue applied red's own proposed text EXACTLY, there is nothing left for red to
		// complain about at this site BY CONSTRUCTION, not by good behaviour. The fact is
		// computed here — never claimed by either seat — and `merge mint` reads it to refuse
		// a fresh gap relitigating text red itself prescribed.
		//
		// Blue is not obliged to reach this state: a counter-edit simply does not set the
		// flag, and record.DeclineStats counts that as blue exercising its right to disagree.
		if gapID := seat.Str(cmd, flags.Answers); gapID != "" {
			verbatim, err := record.ProposalAppliedVerbatim(run, gapID, oldStr, newStr)
			if err != nil {
				return nil, err
			}
			if verbatim {
				body.AppliedVerbatim = proto.Bool(true)
			}
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return editResult{}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey)
	c.Flags().String(flags.Quote, "", flags.DescQuote+". It is matched ACROSS the invisible anchor layer, and rejected if it contains a finding-marker or a citation anchor")
	c.Flags().String(flags.New, "", "the text that span should become")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap id this edit responds to (R1-4) — the provenance join key; omit only for an edit that answers no gap")
	return c
}

// planEdit is the PURE core: it computes the new report from replacing the span of `old`
// with `new`.
//
// The three LEGALITY checks — present-and-unique, no word split, and anchors transit
// unchanged — live in internal/bluedoc, because `merge mint` now has to answer the same
// question about a concrete proposed fix before red may attach one. What stays here is the
// part only blue does: the SPLICE. Under report-as-record no file is written — the BlueEdit event
// IS the mutation and reportproj.Render replays this same splice. The validation peek reuses this
// via validateEdit. Fuzzed directly (edit_fuzz_test.go).
func planEdit(report, old, new string) (string, error) {
	start, end, err := bluedoc.LocateUniqueReplacing("blue edit", report, old)
	if err != nil {
		return "", err
	}
	// ANCHORS MAY TRANSIT AN EDIT — but never be created, destroyed or duplicated by one.
	//
	// This guard used to REJECT any span containing an anchor ("edit around it"). Combined with
	// the uniqueness guard that produces a DEADLOCK, demonstrated: when a word appears twice and
	// the only disambiguating context carries red's anchor, the minimal quote is refused as
	// ambiguous and the contextual quote is refused as anchor-spanning. The anchored occurrence —
	// the one red actually flagged — becomes uneditable, while the unanchored one edits fine. And
	// 71% of anchored quotes in the smoke had their anchor mid-span, so this is the common shape,
	// not a corner.
	if err := bluedoc.AnchorsTransitUnchanged("blue edit", report[start:end], new); err != nil {
		return "", err
	}
	next := reportproj.ApplySplice(report, start, end, new)
	if dropped := droppedMarker(report, next); dropped != "" {
		return "", fmt.Errorf("blue edit: internal error — this edit would drop %s (report unchanged)", anchor.Label(dropped))
	}
	return next, nil
}

// validateEdit rejects a mis-quote or a marker-spanning edit against a snapshot, WITHOUT
// mutating — so no event is recorded for an edit that cannot apply. It RETURNS the planned
// document so the caller can record what this edit reopens, computed from the same snapshot the
// validation used rather than from a re-read that may have moved.
func validateEdit(report, old, new string) (string, error) {
	return planEdit(report, old, new)
}

// droppedMarker returns an immortal-anchor id (finding OR citation) present in before but
// absent from after, or "". The union sweep is what makes the cite⟺anchor bijection hold:
// no raw edit can drop a citation any more than it can drop a finding.
func droppedMarker(before, after string) string {
	have := map[string]bool{}
	for _, id := range claimcount.ProtectedAnchorIDs(after) {
		have[id] = true
	}
	for _, id := range claimcount.ProtectedAnchorIDs(before) {
		if !have[id] {
			return id
		}
	}
	return ""
}

type editResult struct {
	Idempotent bool `json:"idempotent,omitempty"`
}

func (r editResult) Human() string {
	if r.Idempotent {
		return "blue edit (idempotent retry — the op is already on the diff-stack, no second op)"
	}
	return "blue edit recorded — diff-stack op appended, finding-markers preserved, report re-derived on read"
}
