package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// m23Archive is universe-m23: is 91 prime, one lane, 504 events migrated, outcome UNVERIFIED asserted by
// the bench (event 462) on a board with no material gap open.
const m23Archive = "2026-10-13_is-91-prime-m23.tar.gz"

// m23SourceSentence is the quote of m23's event 392, a reading of a source the evidence lens found,
// `--as refutes`. It is the SOURCE's sentence: no state of m23's report holds it.
const m23SourceSentence = "91 is not a prime number. It is a composite number."

// M23'S ARCHIVED READING OF A SOURCE THE LENS FOUND HOLDS NOTHING, AND THIS BINARY REFUSES IT LIVE.
//
// The lens read a source that contradicts the report and quoted the source, not the report. The
// archive records it (event 392), and the run ended UNVERIFIED by the bench's assertion. A reading
// holds nothing; the gap a lens mints on it holds the PASS. Migrate mints no gap for an archived
// reading (its grades would be fabricated), and a reading of a source the lens found names no
// anchor for the ledger to hold — so the migrated reading holds nothing. That loss is accepted and
// stated (gblock, 2026-10-30, GAP 2): the recorded outcome does not move, and the run's own stale
// areas still refuse a PASS at its end.
//
// LIVE, the same reading on the record as it stood before event 392 — with the gap flags a refuting
// reading now needs — is refused, because its quote is not the report's sentence, and records nothing.
func TestM23sArchivedCorroborationHoldsNothing(t *testing.T) {
	to := recordtest.TmpRun(t)
	res, err := migrate.Migrate(recordtest.RequireArchive(t, m23Archive), to, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate m23: %v", err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused migrating m23; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	dst := runtest.Open(t, to)
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(fam.Events) != 504 {
		t.Fatalf("the migrated m23 holds %d events, want 504 — the archive's 476, less its one coining, the 9 receipts blue filed, the 5 positions blue filed and the 7 inquiry reviews and rulings the chair recorded, and the 6 reproduces carried to a second computation on their script, and 44 admissions, one at the close of each sitting of a seat whose work arrives as items", len(fam.Events))
	}
	at := -1
	for i, e := range fam.Events {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok && v.GetClaim() == m23SourceSentence {
			at = i
		}
	}
	if at != 391 {
		t.Fatalf("the reading quoting the source's sentence is event %d of the migrated m23, want 392", at+1)
	}
	was, _ := recordpb.BodyAs[*recordpb.Verify](fam.Events[at])
	if was.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES || fam.Events[at].GetSeatId() != "red-lens-evidence" || was.GetAnchor() != "" {
		t.Fatalf("event 392 is %s by %s on anchor %q, want the evidence lens's refutes of a source it found", was.GetOutcome(), fam.Events[at].GetSeatId(), was.GetAnchor())
	}
	blockers, err := record.PassBlockers(dst)
	if err != nil {
		t.Fatal(err)
	}
	stale := 0
	for _, b := range blockers {
		switch {
		case b.Subject == "392" || b.Detail == m23SourceSentence:
			t.Errorf("the migrated m23 holds a blocker on the archived reading: %+v", b)
		case b.Kind == record.BlockerFailedCheck:
			t.Errorf("the migrated m23 holds a failed check (%s, %s) — no reading of it was of a citation", b.Subject, b.CheckState())
		case b.Kind == record.BlockerStaleArea:
			stale++
		}
	}
	if stale != 4 {
		t.Errorf("the migrated m23 holds %d stale_area blocker(s) at its end, want the chair's 4 — the PASS stays refused without the lost reading", stale)
	}
	var outcome *recordpb.Outcome
	for _, e := range fam.Events {
		if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
			outcome = o
		}
	}
	if outcome == nil || outcome.GetVerdict() != recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED {
		t.Errorf("the migrated m23's outcome is %v, want the UNVERIFIED its bench recorded", outcome.GetVerdict())
	}

	// LIVE: the record as it stood before event 392 — the migrated archive's first 391 events replayed into
	// a run of this binary's making, beside the archive's inputs — and the reading the lens made, as a
	// seat runs it now.
	dir := recordtest.RequireArchive(t, m23Archive)
	for _, f := range []string{"record.db", "record.db-wal", "record.db-shm"} {
		if err := os.Remove(filepath.Join(dir, "records", f)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	recordtest.Seed(t, dir, fam.Events[:at]...)
	counts := func() [3]int {
		return [3]int{countEvents(t, dir, recordpb.EventType_EVENT_TYPE_VERIFY), countEvents(t, dir, recordpb.EventType_EVENT_TYPE_MINT), countEvents(t, dir, recordpb.EventType_EVENT_TYPE_ANCHOR)}
	}
	before := counts()
	_, err = run(t, "verify", "--run", dir, "--seat-id", "red-lens-evidence",
		"--url", was.GetUrl(), "--title", was.GetTitle(), "--quote", m23SourceSentence,
		"--as", "refutes", "--confidence", recordpb.Word(was.GetConfidence()), "--reason", was.GetText())
	if err == nil {
		t.Fatal("this binary records m23's event 392: a refuting reading quoting the source's sentence")
	}
	for _, want := range []string{"--quote was not found in report.md", "REPORT's sentence", "the sentence it contradicts", "--reason"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal of m23's event 392 does not say %q: %v", want, err)
		}
	}
	if after := counts(); after != before {
		t.Errorf("the refused reading left verify/mint/anchor counts %v, were %v", after, before)
	}
}
