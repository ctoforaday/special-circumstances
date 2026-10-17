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

// m23Archive is universe-m23: is 91 prime, one lane, 482 events migrated, outcome UNVERIFIED asserted by
// the bench (event 480) on a board with no material gap open.
const m23Archive = "2026-10-13_is-91-prime-m23.tar.gz"

// m23SourceSentence is the quote of m23's event 373, a `corroborate --as refutes` by the evidence
// lens. It is the SOURCE's sentence: no state of m23's report holds it.
const m23SourceSentence = "91 is not a prime number. It is a composite number."

// M23 IS THE CASE: THIS BINARY REFUSES ITS EVENT 373, AND MIGRATES ITS RECORD WITH THE EVENT KEPT.
//
// The lens read a source that contradicts the report and quoted the source. The quote was held to
// the report only where a source backed the claim, so the corroboration was recorded, and the PASS
// blocker it left names a sentence no finding can quote. The lens filed three findings at it and
// logged the defect three times (events 376, 402, 443); the plan readied it once per contradiction;
// no verdict could be issued and the bench asserted UNVERIFIED.
//
// LIVE, the same call on the record as it stood before event 373 is refused and records nothing.
//
// MIGRATED, the event stands and so does its blocker. Migrate translates history and does not
// re-judge it: it replays acts through record.Append, and the placement of a quote is the verb's
// check (seat.Places), which a replay never reaches — as it reaches no finding's, cite's or mint's.
// The blocker is why this run's outcome is what it is: a migrated m23 without it would hold an
// asserted UNVERIFIED over a board that permits a PASS, and nothing on the record to say why. The
// run is over, so no seat is asked to clear it — a finding naming the contradiction by its id would
// have answered it, sentence or no sentence.
func TestM23sRefutingCorroborationIsRefusedLiveAndItsArchiveKeepsIt(t *testing.T) {
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
	if len(fam.Events) != 482 {
		t.Fatalf("the migrated m23 holds %d events, want 482 — the archive's 476 and the 6 reproduces carried to a second proof on their script", len(fam.Events))
	}
	at := -1
	for i, e := range fam.Events {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok && v.GetClaim() == m23SourceSentence {
			at = i
		}
	}
	if at != 372 {
		t.Fatalf("the corroboration quoting the source's sentence is event %d of the migrated m23, want 373", at+1)
	}
	was, _ := recordpb.BodyAs[*recordpb.Verify](fam.Events[at])
	if was.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES || fam.Events[at].GetSeatId() != "red-lens-evidence" {
		t.Fatalf("event 373 is %s by %s, want the evidence lens's refutes", was.GetOutcome(), fam.Events[at].GetSeatId())
	}
	contradictions := func(run record.Run) []record.Blocker {
		t.Helper()
		all, err := record.PassBlockers(run)
		if err != nil {
			t.Fatal(err)
		}
		var out []record.Blocker
		for _, b := range all {
			if b.Kind == record.BlockerContradiction {
				out = append(out, b)
			}
		}
		return out
	}
	held := contradictions(dst)
	if len(held) != 1 || held[0].Subject != "373" || held[0].Detail != m23SourceSentence || held[0].Owner != "red-lens-evidence" {
		t.Errorf("the migrated m23's contradiction blockers = %+v, want contradiction 373 on the source's sentence, owned by the evidence lens", held)
	}
	var outcome *recordpb.Outcome
	for _, e := range fam.Events {
		if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
			outcome = o
		}
	}
	if outcome == nil || outcome.GetVerdict() != recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED {
		t.Errorf("the migrated m23's outcome is %v, want the UNVERIFIED its bench recorded over that blocker", outcome.GetVerdict())
	}

	// LIVE: the record as it stood before event 373 — the migrated archive's first 372 events replayed into
	// a run of this binary's making, beside the archive's inputs and class registry — and the call
	// the lens made, flag for flag.
	dir := recordtest.RequireArchive(t, m23Archive)
	for _, f := range []string{"record.db", "record.db-wal", "record.db-shm"} {
		if err := os.Remove(filepath.Join(dir, "records", f)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	recordtest.Seed(t, dir, fam.Events[:at]...)
	_, err = run(t, "corroborate", "--run", dir, "--seat-id", "red-lens-evidence",
		"--url", was.GetUrl(), "--title", was.GetTitle(), "--quote", m23SourceSentence,
		"--as", "refutes", "--confidence", recordpb.Word(was.GetConfidence()), "--reason", was.GetText())
	if err == nil {
		t.Fatal("this binary records m23's event 373: a refuting corroboration quoting the source's sentence")
	}
	for _, want := range []string{"--quote was not found in report.md", "REPORT's sentence", "the sentence it contradicts", "--reason"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal of m23's event 373 does not say %q: %v", want, err)
		}
	}
	if left := contradictions(runtest.Open(t, dir)); len(left) != 0 {
		t.Errorf("the refused call left a contradiction on the record: %+v", left)
	}
}
