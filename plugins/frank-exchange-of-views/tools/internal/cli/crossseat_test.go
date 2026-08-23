package cli

import (
	"strings"
	"testing"
)

// CROSS-SEAT INTERACTIONS, the second pass.
//
// integration_test.go established that the seats can see each other at all. This file
// covers the SEQUENCES — the multi-step exchanges the protocol is actually made of, where
// each step's meaning depends on what another seat did before it.
//
// The bar for what belongs here: if the interaction can be got wrong in a way that no
// single-verb test would notice, it belongs. The run's own defect — red's board blind to
// the bench's closures — was exactly that shape, and it shipped because every test drove
// one verb and asked about its own event.

// gapIsOpen reports which side of the closure index a gap sits on in red's projection.
// The section a gap lives in IS its status, so every board assertion goes through here
// rather than asking whether the id appears at all — "mentioned somewhere" is the weak
// assertion that let the original bench-closure defect pass.
func gapIsOpen(t *testing.T, runDir, id string) bool {
	t.Helper()
	ledger := readProjection(t, runDir, "ledger")
	cut := strings.Index(strings.ToLower(ledger), "closure index")
	if cut < 0 {
		t.Fatalf("the projection has no closure index; open and closed cannot be distinguished.\n%s", ledger)
	}

	// Membership is an ENTRY, not a substring. A closed gap names its successor on its
	// own closure row, so a naive strings.Contains reports the successor as living in
	// both sections at once — the same too-weak assertion that let the original
	// bench-closure defect through, reappearing in the helper written to catch it.
	// Open entries are "### <id> — ..." headings; closure rows begin with the id.
	open, closed := false, false
	for _, ln := range strings.Split(ledger[:cut], "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "### "+id+" ") {
			open = true
		}
	}
	for _, ln := range strings.Split(ledger[cut:], "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), id+" |") {
			closed = true
		}
	}
	if open == closed {
		t.Fatalf("gap %s is in %s section(s) of the board — it must be in exactly one.\n%s",
			id, map[bool]string{true: "BOTH", false: "NEITHER"}[open], ledger)
	}
	return open
}

// A CARRIED gap is not a closed one, and the distinction is the entire bench reform.
//
// The judge was measured behaving as "a router with robes" — carrying 76 of 77 gaps
// forward, which moved work without deciding anything. The reform made carrying a real
// ruling with a research direction attached, and the run that followed carried 1 of 9.
// If replay ever treats carried as closing, that metric silently reads perfect and the
// board loses every genuinely-open gap the judge touched.
func TestBenchCarriedLeavesTheGapOpenWhileClosedDoesNot(t *testing.T) {
	runDir := seatRun(t)
	carried := mintGap(t, runDir, "bench-carries-this", "carry-vs-close")
	closed := mintGap(t, runDir, "bench-closes-that", "carry-vs-close")

	for _, c := range []struct{ id, as, principle string }{
		{carried, "carried", "the repair is unverified at the leaf; it needs another round"},
		{closed, "repaired", "the repair discharges the defect and the anchor is checkable"},
	} {
		if _, err := run(t, "opinion", "--run", runDir, "--seat-id", "judge-r1",
			"--id", c.id, "--as", c.as, "--principle", c.principle,
			"--tension", "thoroughness against ceremony",
			"--review-flag", "no — the ruling is mechanical", "--settled", "the proposition this ruling bars", "--final",
			"--reason", "the ruling as reasoned"); err != nil {
			t.Fatalf("bench opinion %s: %v", c.as, err)
		}
	}

	if !gapIsOpen(t, runDir, carried) {
		t.Errorf("gap %s was CARRIED, not closed, and must still be open — a carry that closes turns the judge back into a router and reports the carried_share metric as perfect while the board loses the gap", carried)
	}
	if gapIsOpen(t, runDir, closed) {
		t.Errorf("gap %s was CLOSED by the bench and is still on the open board", closed)
	}
}

// The full grade dispute: blue contests, red accepts, red regrades. The regrade is a
// SEPARATE act from the acceptance, and the board must end up showing the moved grade —
// an accepted dispute that never moves the grade is agreement with no consequence.
func TestAcceptedDisputeIsFollowedByAGradeThatActuallyMoves(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "grade-moves", "dispute-to-regrade")

	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", id, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation"); err != nil {
		t.Fatalf("motion grade file: %v", err)
	}
	if _, err := run(t, "motion", "grade", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "M1", "--as", "accepted", "--reason", "the bound holds; regrading"); err != nil {
		t.Fatalf("motion grade rule: %v", err)
	}
	if _, err := run(t, "regrade", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--severity", "low",
		"--reason", "blue's dispute is accepted — the caller validates, so the blast radius is one call"); err != nil {
		t.Fatalf("red regrade: %v", err)
	}

	ev := lastOfType(t, runDir, "regrade")
	if got := ev.Payload.Str("severity"); got != "low" {
		t.Errorf("regrade recorded severity %q, want low — an accepted dispute that does not move the grade is a channel with no consequence", got)
	}
	if !payloadKeys(ev)["reason"] {
		t.Error("the regrade lost its basis; grade movement without a stated reason is the silent regrading this channel exists to prevent")
	}
}

// A petition crosses roles in both directions: a MERGE files it, the BENCH rules on it,
// and the relief is meant to bind the seats that come after. Both halves must be in the
// one record, or "heard before the debate continues" is unenforceable.
func TestPetitionCrossesFromMergeToBenchAndItsReliefIsRecorded(t *testing.T) {
	runDir := seatRun(t)

	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-merge-r1",
		"--class", "safety",
		"--reason", "continuing would require asserting a consent gate exists where it does not",
		"--relief", "halt and escalate to a human before the next round"); err != nil {
		t.Fatalf("motion petition file: %v", err)
	}
	if _, err := run(t, "motion", "petition", "rule", "--run", runDir, "--seat-id", "judge-r1",
		"--id", "M1", "--as", "granted",
		"--reason", "the relief binds the coming seats"); err != nil {
		t.Fatalf("motion petition rule: %v", err)
	}

	pet := lastOfType(t, runDir, "motion")
	if pet.Payload.Str("class") != "safety" || !payloadKeys(pet)["relief"] {
		t.Errorf("the petition lost its class or relief (payload %v) — relief that is not recorded cannot bind anybody", pet.Payload.Keys())
	}
	rule := lastOfType(t, runDir, "motion-rule")
	if got := rule.Payload.Str("ruling"); got != "granted" {
		t.Errorf("the ruling recorded %q, want granted", got)
	}
	// THE ATTRIBUTION IS THE ID NOW, and that is the substance of #312. The old check read
	// `petitioner` off the ruling — a field the ruler restated, which is why two petitions from
	// one seat in one round could not be told apart. The ruling names the MOTION, and the motion
	// names its filer, so the join is a fact rather than a restatement.
	if got := rule.Payload.Str("motion_id"); got != "M1" {
		t.Errorf("the ruling names motion %q, want M1 — a ruling that does not name its filing cannot be matched to it", got)
	}
	if got := pet.SeatID; got != "red-merge-r1" {
		t.Errorf("the motion was filed by %q, want the merge seat — the filer is on the filing, never restated on the answer", got)
	}
}

// Red reported that spot-check "cannot record an honestly-empty round". It always could:
// --none with --reason is exactly that path. The claim was unverified self-report, which
// is the friction-channel defect Gray Area exists to catch — so the refutation gets a test.
func TestSpotCheckCanRecordAnHonestlyEmptyArchive(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none", "--reason", "the archive was empty at round start; there was nothing to sample"); err != nil {
		t.Fatalf("an empty-archive spot-check must be recordable — red reported this was impossible and it was not: %v", err)
	}
	ev := lastOfType(t, runDir, "spot-check")
	if !payloadKeys(ev)["reason"] {
		t.Error("the empty spot-check lost its reason, which is the only thing distinguishing it from a skipped duty")
	}

	// And the duty cannot be discharged by asserting emptiness with no reason.
	if _, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none"); err == nil {
		t.Error("--none without --reason was accepted; an unexplained empty round is indistinguishable from a skipped one")
	}
}

// Blue retires a claim; the claim leaves ONLY through this verb. Capture compares the
// claim_count fall against the retire events, so a retire that loses its reason or its
// successor breaks the accounting that detects claims vanishing quietly.
func TestRetiredClaimCarriesItsReasonAndSuccessor(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "the API returns 200 on a malformed body",
		"--reason", "refuted at the leaf — it returns 400, verified against the handler",
		"--new", "the API returns 400 on a malformed body"); err != nil {
		t.Fatalf("blue retire: %v", err)
	}
	ev := lastOfType(t, runDir, "retire")
	for _, want := range []string{"claim", "reason", "superseded_by"} {
		if !payloadKeys(ev)[want] {
			t.Errorf("the retirement lost %s (payload %v) — an unaccounted claim drop is the detector hit this verb exists to make impossible", want, ev.Payload.Keys())
		}
	}
}

// Two lenses write CONCURRENTLY into their own shards and the merge must read both. This
// is the append-only design's central physical claim; if shard merging were lossy, the
// board would silently under-count findings in exactly the runs that worked hardest.
func TestConcurrentLensShardsBothReachTheMerge(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L2"); err != nil {
		t.Fatalf("register L2: %v", err)
	}
	for _, l := range []struct{ seat, label string }{
		{"red-lens-r1-L1", "L1-F1"}, // local --key F1 → tool assigns L1-F1 (role-prefixed)
		{"red-lens-r1-L2", "L2-F1"}, // and L2-F1
	} {
		if _, err := run(t, "finding", "--run", runDir, "--seat-id", l.seat,
			"--key", "F1", "--quote", "§1", "--reason", "a finding",
			"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
			t.Fatalf("finding %s: %v", l.label, err)
		}
	}

	seen := map[string]bool{}
	for _, e := range events(t, runDir) {
		if e.Type == "finding" {
			seen[e.Payload.Str("label")] = true
		}
	}
	if !seen["L1-F1"] || !seen["L2-F1"] {
		t.Errorf("the merge sees %v, want both lens shards — a lossy shard merge under-counts findings precisely in the rounds with the most lenses", seen)
	}
}

// A closure with a SUCCESSOR: the check passes here, but the residue moves to a named
// gap. Losing the successor turns a partial repair into a clean one on the board.
func TestClosureWithSuccessorNamesWhereTheResidueWent(t *testing.T) {
	runDir := seatRun(t)
	first := mintGap(t, runDir, "partial-repair", "residue-carrying")
	next := mintGap(t, runDir, "the-residue", "residue-carrying")

	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", first, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/parser",
		"--superseded-by", next,
		"--reason", "the named site is repaired; the same class survives at the sibling site"); err != nil {
		t.Fatalf("close with successor: %v", err)
	}

	ev := lastOfType(t, runDir, "close")
	if got := ev.Payload.Str("successor"); got != next {
		t.Errorf("the closure records successor %q, want %s — a partial repair whose residue is unnamed reads as a complete one", got, next)
	}
	if gapIsOpen(t, runDir, first) {
		t.Errorf("gap %s was closed and is still open on the board", first)
	}
	if !gapIsOpen(t, runDir, next) {
		t.Errorf("the successor %s must remain OPEN — it is where the unrepaired residue lives", next)
	}
}

// The bench halts. A halt is deliberately its own verb rather than a value of --as, so it
// can never be reached by a typo in an enum, and it must be findable in the record by any
// seat that reads it.
func TestBenchHaltIsItsOwnActAndIsVisibleInTheRecord(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "halt", "--run", runDir, "--seat-id", "judge-r1",
		"--reason", "continuing would compromise the consent gate"); err != nil {
		t.Fatalf("bench halt: %v", err)
	}
	if got := lastOfType(t, runDir, "halt").SeatID; got != "judge-r1" {
		t.Errorf("the halt is attributed to %q, want judge-r1 — an unattributed halt cannot be reviewed", got)
	}

	// A halt must NOT be reachable as a disposition on a ruling: that is the typo path
	// the separate verb exists to close off.
	id := mintGap(t, runDir, "not-haltable", "halt-is-its-own-verb")
	if _, err := run(t, "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", id, "--as", "halt", "--principle", "p", "--tension", "t",
		"--review-flag", "no", "--settled", "the proposition this ruling bars", "--final", "--reason", "attempting to halt via a disposition"); err == nil {
		t.Error("`opinion --as halt` was accepted; ending the run must not be reachable by a mistyped disposition")
	}
}
