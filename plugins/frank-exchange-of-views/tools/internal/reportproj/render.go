package reportproj

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE REPORT IS THE RECORD, NOT A FILE (#709).
//
// The current report is the frozen base with every text-mutating event replayed over it, in
// record order. Five verbs mutate the report, and each replays through the SAME transform it
// applied at write time, located by the SAME rule:
//
//   - blue edit       → a splice (old→new), located by bluedoc.LocateUniqueReplacing (spliceMut).
//   - blue cite       → a citation marker spliced at the quoted sentence (insertMut).
//   - blue prove      → a proof marker, likewise.
//   - lens finding    → a finding marker, recorded as an Anchor event (its Finding sibling is
//                       metadata and inserts nothing — replaying it too would double the marker).
//   - lens corroborate→ a citation marker at the corroborated claim, on the Verify event (only
//                       when it backs the claim and so carries a c- Label; a plain verify does not).
//
// Because replay reproduces the identical running text each verb saw, the locators resolve the
// identical offsets, so the bytes match what the file used to hold. Once the base is ingested and
// the file deleted, this is the ONLY way the report exists: there is no file to raw-write,
// chattr, or bypass, and a change that is not a recorded event cannot exist.
//
// Every failure here is LOUD, never a plausible zero: an op that will not locate is a record that
// no longer describes a real sequence over this base, and a silently skipped op would render a
// report that never existed.

// ApplySplice is the pure text transform blue edit and replay share: replace report[start:end]
// with new, then tidy the two seams the splice created. NO legality guards — those (anchors
// transit, no dropped marker, present-and-unique) are write-time validation that already passed
// when the edit was recorded; replay reproduces bytes, it does not re-authorise.
func ApplySplice(report string, start, end int, new string) string {
	next := report[:start] + new + report[end:]
	// SPLICE HYGIENE (see splice.go): tidy the two seams this edit created. Deterministic, narrow,
	// silent — and it MUST run in replay too, or the replayed text drifts from what the file held.
	next, _ = tidySeam(next, start+len(new)) // trailing seam first — the later offset is stable
	next, _ = tidySeam(next, start)
	return next
}

// Op is one recorded splice from the diff-stack: the (old, new) of a BlueEdit event. Render takes
// plain strings rather than *recordpb.BlueEdit so callers testing the edit path need not build the
// record schema — the record-aware path is RenderFromRecord.
type Op struct{ Old, New string }

// Render reproduces a report from a base and an ordered splice stack. It is the edit-only surface
// blue edit's fidelity tests drive against planEdit; RenderFromRecord composes the full event
// stream (splices AND marker insertions).
func Render(base string, ops []Op) (string, error) {
	text := base
	for i, op := range ops {
		next, err := (spliceMut{old: op.Old, new: op.New}).apply(text)
		if err != nil {
			return "", fmt.Errorf("render: replaying edit %d of %d: %w", i+1, len(ops), err)
		}
		text = next
	}
	return text, nil
}

// mutation is one recorded text change, applied to the running report during replay. The two
// shapes are a splice (blue edit) and a marker insertion (cite/prove/finding); both locate against
// the running text exactly as the verb did, and both are loud when they cannot.
type mutation interface {
	apply(text string) (string, error)
	describe() string
}

type spliceMut struct{ old, new string }

func (m spliceMut) apply(text string) (string, error) {
	start, end, err := bluedoc.LocateUniqueReplacing("render", text, m.old)
	if err != nil {
		return "", err
	}
	return ApplySplice(text, start, end, m.new), nil
}

func (m spliceMut) describe() string { return fmt.Sprintf("edit %q→%q", m.old, m.new) }

type insertMut struct{ location, marker string }

func (m insertMut) apply(text string) (string, error) {
	// TORN-ANCHOR ADOPTION, REPLAYED. The verbs adopt an orphan marker rather than splice a
	// rival; if two events ever name the same marker (a crash-retry that appended twice), replay
	// must place it ONCE. Keyed on the exact marker token, so two findings on one sentence — two
	// distinct tokens — both still insert.
	if strings.Contains(text, m.marker) {
		return text, nil
	}
	next, err := anchortext.InsertAnchor([]byte(text), m.location, m.marker)
	if err != nil {
		return "", err
	}
	return string(next), nil
}

func (m insertMut) describe() string { return fmt.Sprintf("insert %s at %q", m.marker, m.location) }

// RenderFromRecord is the record-aware render: it reads the run's frozen base (the one BaseIngest
// event) and every text-mutating event, and folds them over the base in record order. This is the
// ONLY way to obtain the current report once the file is deleted — the former readers of
// blue/report.md call this instead.
//
// Events arrive in insertion order (MergedEvents is ORDER BY id), which is mutation order, and the
// BaseIngest is written first (round 0, before any mutation), so a linear pass composes correctly.
// Two structural failures are loud, never a plausible zero: no base (nothing was ingested) and two
// bases (the write-once rule was broken).
func RenderFromRecord(run record.Run) (string, error) {
	m, err := record.MergedEvents(run)
	if err != nil {
		return "", err
	}
	base := ""
	haveBase := false
	var muts []mutation
	for _, e := range m.Events {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.BaseIngest:
			if haveBase {
				return "", fmt.Errorf("render: two BaseIngest events on this run — the base is written once and never rewritten")
			}
			base, haveBase = b.GetText(), true
		case *recordpb.BlueEdit:
			muts = append(muts, spliceMut{old: b.GetOld(), new: b.GetNew()})
		case *recordpb.Cite:
			muts = append(muts, insertMut{location: b.GetLocation(), marker: anchor.Token(b.GetLabel())})
		case *recordpb.Proof:
			muts = append(muts, insertMut{location: b.GetLocation(), marker: anchor.Token(b.GetProofId())})
		case *recordpb.Anchor:
			// The finding-marker record. Its paired Finding event carries the grades and inserts
			// nothing — replaying that too would splice the marker twice.
			muts = append(muts, insertMut{location: b.GetLocation(), marker: anchor.Token(b.GetId())})
		case *recordpb.Verify:
			// A red corroboration that BACKS a claim carries a c- Label and places a citation
			// marker at its Claim; a plain verify (adjudicating a blue anchor) has no label and
			// placed no marker.
			if b.GetLabel() != "" {
				muts = append(muts, insertMut{location: b.GetClaim(), marker: anchor.Token(b.GetLabel())})
			}
		}
	}
	if !haveBase {
		return "", fmt.Errorf("render: no base has been ingested for this run — there is nothing to render")
	}
	text := base
	for i, mut := range muts {
		next, err := mut.apply(text)
		if err != nil {
			return "", fmt.Errorf("render: replaying mutation %d of %d (%s): %w", i+1, len(muts), mut.describe(), err)
		}
		text = next
	}
	return text, nil
}

// BaseIngested reports whether this run already has a frozen base — the record-state test that
// makes `blue ingest` write-once: a second ingest is refused, never an overwrite. The answer is
// read from the events, not a marker file, so it cannot drift from the truth.
func BaseIngested(run record.Run) (bool, error) {
	m, err := record.MergedEvents(run)
	if err != nil {
		return false, err
	}
	for _, e := range m.Events {
		if body, ok := recordpb.Body(e); ok {
			if _, isBase := body.(*recordpb.BaseIngest); isBase {
				return true, nil
			}
		}
	}
	return false, nil
}
