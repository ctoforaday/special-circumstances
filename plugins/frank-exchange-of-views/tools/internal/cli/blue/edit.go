package blue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
// Crash-safety is EVENT-FIRST: the event is appended only AFTER validation succeeds (so a
// mis-quote never lands a phantom op), then the write is applied; a crash between leaves the
// event durable and the write reconciled idempotently on retry — no wedge, no phantom op.
func newEdit() *cobra.Command {
	c := seat.Prose(seat.New("edit",
		`replace an exact span in blue/report.md, preserving red's finding-markers: --key <your F1> --old "<exact current span>" --new "<replacement>" --reason "..."`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			reason, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(reason) == "" {
				return nil, fmt.Errorf("blue edit requires --reason: why this change (the argument red re-audits against)")
			}
			oldStr := seat.Str(cmd, flags.Old)
			newStr := seat.Str(cmd, flags.New)
			if oldStr == "" {
				return nil, fmt.Errorf("blue edit requires --old: the EXACT current span to replace (matched across the invisible marker layer, like the Edit tool)")
			}
			if oldStr == newStr {
				return nil, fmt.Errorf("blue edit: --old and --new are identical — no change to make")
			}
			key := seat.Str(cmd, flags.Key)

			// Crash-retry: a committed blue_edit for this key means the op is already on the
			// stack — reconcile the write idempotently, do NOT append a second op.
			prior, err := record.ExistingBlueEditByKey(s.RunDir, s.SeatID, key)
			if err != nil {
				return nil, err
			}
			if prior {
				if err := applyEdit(s.RunDir, oldStr, newStr); err != nil {
					return nil, err
				}
				return editResult{Idempotent: true}, nil
			}

			// FRESH: validate against a consistent snapshot BEFORE committing the event, so a
			// mis-quote or a marker-spanning edit never lands a phantom stack op.
			peek, err := record.ReadBlueReport(s.RunDir)
			if err != nil {
				return nil, err
			}
			if err := validateEdit(string(peek), oldStr, newStr); err != nil {
				return nil, err
			}

			// Event-first: commit the diff-stack op, then apply the write.
			p := record.NewPayload()
			seat.Set(cmd, p, "edit_key", flags.Key)
			p.Set("old", oldStr)
			p.Set("new", newStr)
			p.Set("text", reason)
			if _, err := record.Append(s.RunDir, s.SeatID, "blue_edit", p); err != nil {
				return nil, err
			}
			if err := applyEdit(s.RunDir, oldStr, newStr); err != nil {
				return nil, err
			}
			return editResult{}, nil
		}))

	c.Flags().String(flags.Key, "", "a stable local handle (your own F1, F2 …) making a retried edit idempotent")
	c.Flags().String(flags.Old, "", "REQUIRED — the EXACT current span to replace (matched across the invisible anchor layer; rejected if absent or if it contains a finding-marker or a citation anchor)")
	c.Flags().String(flags.New, "", "the replacement text")
	return c
}

// errMisQuote is the sentinel for "--old is not present" — applyEdit distinguishes it to
// drive the crash-reconcile branch (old gone but new already present ⇒ the write landed).
var errMisQuote = errors.New("blue edit: the --old text was not found in report.md — quote the EXACT current span you are replacing (whitespace and the invisible marker layer are ignored; everything else must match)")

// planEdit is the PURE core: it computes the new report from replacing the span of `old`
// with `new`, or an error if old is absent (errMisQuote) or its span contains an immortal
// anchor of either class (a finding marker or a citation anchor). The crash-reconcile and
// the lock live in applyEdit; the validation peek reuses this via validateEdit. Fuzzed
// directly (edit_fuzz_test.go).
func planEdit(report, old, new string) (string, error) {
	start, end, ambiguous := lens.LocateSpanUnique(report, old)
	if start < 0 {
		return "", errMisQuote
	}
	// AMBIGUITY IS REFUSED, NOT GUESSED. Taking the first of several matches silently edits a site
	// blue may not have meant — and blue is explicitly told to propagate corrections to every site
	// stating a claim, so repeated text is the expected shape of a real report.
	if ambiguous {
		return "", fmt.Errorf("blue edit: your --old span appears MORE THAN ONCE in report.md, so the edit target is ambiguous — quote more surrounding context to pick out the one site you mean (to change every site, make one edit per site)")
	}
	// The span may not split a WORD (see splice.go for why this is not a whitespace rule).
	if !spanBoundaryOK(report, start, end) {
		return "", fmt.Errorf("blue edit: your --old span starts or ends inside a word — quote whole words. Editing letters rather than language produces one-byte ops nobody can read in the diff-stack")
	}
	// ANCHORS MAY TRANSIT AN EDIT — but never be created, destroyed or duplicated by one.
	//
	// This guard used to REJECT any span containing an anchor ("edit around it"). Combined with the
	// uniqueness guard above that produces a DEADLOCK, demonstrated: when a word appears twice and
	// the only disambiguating context carries red's anchor, the minimal quote is refused as
	// ambiguous and the contextual quote is refused as anchor-spanning. The anchored occurrence —
	// the one red actually flagged — becomes uneditable, while the unanchored one edits fine. And
	// 71% of anchored quotes in the smoke had their anchor mid-span, so this is the common shape,
	// not a corner.
	//
	// The invariant worth keeping is not "blue never touches an anchor"; it is "an edit never
	// changes WHICH anchors exist". That is checkable HERE, in the tool, on the exact bytes: the
	// multiset of anchor ids in --old must equal that in --new. Blue merely carries them across.
	// This is strictly stronger than the old rule, which said nothing about --new at all (5 edits
	// in the smoke smuggled an anchor INTO --new and duplicated it).
	if err := anchorsTransitUnchanged(report[start:end], new); err != nil {
		return "", err
	}
	next := report[:start] + new + report[end:]
	// SPLICE HYGIENE (see splice.go): tidy the two seams this edit created before anything else
	// looks at the result. Deterministic, narrow, and silent — the alternative, measured, is blue
	// spending a third of a round repairing its own doubled punctuation by hand.
	next, _ = tidySeam(next, start+len(new)) // trailing seam first — the later offset is stable
	next, _ = tidySeam(next, start)
	if dropped := droppedMarker(report, next); dropped != "" {
		return "", fmt.Errorf("blue edit: internal error — this edit would drop %s (report unchanged)", anchorLabel(dropped))
	}
	return next, nil
}

// validateEdit rejects a mis-quote or a marker-spanning edit against a snapshot, WITHOUT
// mutating — so no event is recorded for an edit that cannot apply.
func validateEdit(report, old, new string) error {
	_, err := planEdit(report, old, new)
	return err
}

// applyEdit performs the span replacement under the blue-report flock. IDEMPOTENT: on a
// crash-retry where the write already landed (old gone, new present) it is a no-op.
func applyEdit(runDir, old, new string) error {
	return record.MutateBlueReport(runDir, func(cur []byte) ([]byte, error) {
		report := string(cur)
		next, err := planEdit(report, old, new)
		if err != nil {
			if errors.Is(err, errMisQuote) && strings.Contains(report, new) {
				return cur, nil // already applied on a prior attempt — no-op (reconcile)
			}
			return nil, err
		}
		return []byte(next), nil
	})
}

// spanMarker reports whether span contains an immortal anchor of EITHER class — a finding
// marker ("<!--fx:") or a citation anchor ("<!--cite:") — and names the first id. A bare
// prefix with no well-formed id still rejects, named generically: never allow a partial
// anchor of either class to be split.
func spanMarker(span string) (bool, string) {
	hasFinding := strings.Contains(span, "<!--fx:")
	hasCitation := strings.Contains(span, "<!--cite:")
	if !hasFinding && !hasCitation {
		return false, ""
	}
	if ids := claimcount.ProtectedAnchorIDs(span); len(ids) > 0 {
		return true, ids[0]
	}
	if hasFinding {
		return true, "a finding-marker"
	}
	return true, "a citation anchor"
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

// anchorLabel describes an immortal-anchor id for an error message, by its class prefix, so
// a seat is told which kind of anchor its edit would have disturbed. A generic name (from
// spanMarker's partial-prefix case) is passed through unchanged.
func anchorLabel(id string) string {
	switch {
	case strings.HasPrefix(id, "c-"):
		return "citation anchor " + id + " (citations are tool-managed — remove one with the tool, never a raw edit)"
	case strings.HasPrefix(id, "f-"):
		return "finding-marker " + id
	default:
		return id
	}
}

type editResult struct {
	Idempotent bool `json:"idempotent,omitempty"`
}

func (r editResult) Human() string {
	if r.Idempotent {
		return "blue edit (idempotent retry — reconciled, no second stack op)"
	}
	return "blue edit applied — report.md updated, finding-markers preserved, diff-stack op recorded"
}
