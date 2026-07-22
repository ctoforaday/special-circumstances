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
	if _, err := run(t, "merge", "render", "--run", runDir, "--seat-id", "red-merge-r1"); err != nil {
		t.Fatalf("render: %v", err)
	}
	ledger := readProjection(t, runDir, "ledger.md")
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
		{closed, "closed", "the repair discharges the defect and the anchor is checkable"},
	} {
		if _, err := run(t, "bench", "opinion", "--run", runDir, "--seat-id", "judge-r1",
			"--id", c.id, "--as", c.as, "--principle", c.principle,
			"--tension", "thoroughness against ceremony",
			"--review-flag", "no — the ruling is mechanical",
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

	if _, err := run(t, "blue", "dispute", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", id, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation"); err != nil {
		t.Fatalf("blue dispute: %v", err)
	}
	if _, err := run(t, "merge", "dispute-respond", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "accepted", "--reason", "the bound holds; regrading"); err != nil {
		t.Fatalf("red dispute-respond: %v", err)
	}
	if _, err := run(t, "merge", "regrade", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--severity", "low",
		"--reason", "blue's dispute is accepted — the caller validates, so the blast radius is one call"); err != nil {
		t.Fatalf("red regrade: %v", err)
	}

	ev := lastOfType(t, runDir, "regrade")
	if got := ev.Payload.Str("severity"); got != "low" {
		t.Errorf("regrade recorded severity %q, want low — an accepted dispute that does not move the grade is a channel with no consequence", got)
	}
	if !payloadKeys(ev)["basis"] {
		t.Error("the regrade lost its basis; grade movement without a stated reason is the silent regrading this channel exists to prevent")
	}
}

// A lens finds, the merge mints citing it, the merge disposes the leftover observation.
// This is the whole red pipeline, and every step names an artifact a DIFFERENT seat made.
func TestLensWorkReachesTheBoardAndEveryObservationGetsAFate(t *testing.T) {
	runDir := seatRun(t)

	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-F1", "--location", "§3 \"the parser accepts an empty body\"",
		"--reason", "an empty body is accepted where the grammar requires one element",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("lens finding: %v", err)
	}
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note",
		"--reason", "the neighbouring function has the same shape but validates first"); err != nil {
		t.Fatalf("lens observe: %v", err)
	}

	// The mint cites the lens's finding by ITS label — a cross-seat reference.
	out, err := run(t, "merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "from-l1", "--class-new", "lens-to-board",
		"--definition", "d", "--neighbor", "n", "--distinguisher", "x",
		"--location", "§3 \"the parser accepts an empty body\"",
		"--problem", "the defect", "--fix", "the fix",
		"--check", "feed the parser an empty body and assert it is refused",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--cx", "low",
		"--found-by", "L1-F1")
	if err != nil {
		t.Fatalf("mint --found-by: %v", err)
	}
	if id := gapID(out); id == "" {
		t.Fatalf("mint returned no id: %q", out)
	}

	mint := lastOfType(t, runDir, "mint")
	if got := mint.Payload.StrList("found_by"); len(got) != 1 || got[0] != "L1-F1" {
		t.Errorf("the minted gap records found_by %v, want [L1-F1] — without it the lens's finding and the board gap are two unrelated records and no one can tell which lens earned the gap", got)
	}

	if _, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", "L1-O1", "--as", "declined",
		"--reason", "checked at the leaf; the neighbour validates first, so there is no defect"); err != nil {
		t.Fatalf("dispose: %v", err)
	}
	if got := lastOfType(t, runDir, "dispose").Payload.Str("observation"); got != "L1-O1" {
		t.Errorf("the disposal names observation %q, want L1-O1 — an observation with no fate is the lens's work quietly dropped", got)
	}
}

// A petition crosses roles in both directions: a LENS files it, the BENCH rules on it,
// and the relief is meant to bind the seats that come after. Both halves must be in the
// one record, or "heard before the debate continues" is unenforceable.
func TestPetitionCrossesFromLensToBenchAndItsReliefIsRecorded(t *testing.T) {
	runDir := seatRun(t)

	if _, err := run(t, "lens", "petition", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--petition-class", "safety",
		"--reason", "continuing would require asserting a consent gate exists where it does not",
		"--relief", "halt and escalate to a human before the next round"); err != nil {
		t.Fatalf("lens petition: %v", err)
	}
	if _, err := run(t, "bench", "petition-rule", "--run", runDir, "--seat-id", "judge-r1",
		"--petitioner", "red-lens-r1-L1", "--petition-class", "safety", "--as", "granted",
		"--reason", "the relief binds the coming seats"); err != nil {
		t.Fatalf("bench petition-rule: %v", err)
	}

	pet := lastOfType(t, runDir, "petition")
	if pet.Payload.Str("class") != "safety" || !payloadKeys(pet)["relief"] {
		t.Errorf("the petition lost its class or relief (payload %v) — relief that is not recorded cannot bind anybody", pet.Payload.Keys())
	}
	rule := lastOfType(t, runDir, "petition-rule")
	if got := rule.Payload.Str("ruling"); got != "granted" {
		t.Errorf("the ruling recorded %q, want granted", got)
	}
	if got := rule.Payload.Str("petitioner"); got != "red-lens-r1-L1" {
		t.Errorf("the ruling names petitioner %q, want the lens that filed it — an unattributed ruling cannot be matched to its petition", got)
	}
}

// Red reported that spot-check "cannot record an honestly-empty round". It always could:
// --none with --reason is exactly that path. The claim was unverified self-report, which
// is the friction-channel defect Gray Area exists to catch — so the refutation gets a test.
func TestSpotCheckCanRecordAnHonestlyEmptyArchive(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none", "--reason", "the archive was empty at round start; there was nothing to sample"); err != nil {
		t.Fatalf("an empty-archive spot-check must be recordable — red reported this was impossible and it was not: %v", err)
	}
	ev := lastOfType(t, runDir, "spot-check")
	if !payloadKeys(ev)["reason"] {
		t.Error("the empty spot-check lost its reason, which is the only thing distinguishing it from a skipped duty")
	}

	// And the duty cannot be discharged by asserting emptiness with no reason.
	if _, err := run(t, "merge", "spot-check", "--run", runDir, "--seat-id", "red-merge-r1",
		"--none"); err == nil {
		t.Error("--none without --reason was accepted; an unexplained empty round is indistinguishable from a skipped one")
	}
}

// Blue retires a claim; the claim leaves ONLY through this verb. Capture compares the
// claim_count fall against the retire events, so a retire that loses its reason or its
// successor breaks the accounting that detects claims vanishing quietly.
func TestRetiredClaimCarriesItsReasonAndSuccessor(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "blue", "retire", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--claim", "the API returns 200 on a malformed body",
		"--reason", "refuted at the leaf — it returns 400, verified against the handler",
		"--superseded-by", "the API returns 400 on a malformed body"); err != nil {
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
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", "red-lens-r1-L2"); err != nil {
		t.Fatalf("register L2: %v", err)
	}
	for _, l := range []struct{ seat, label string }{
		{"red-lens-r1-L1", "L1-F1"},
		{"red-lens-r1-L2", "L2-F1"},
	} {
		if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", l.seat,
			"--label", l.label, "--location", "§1", "--reason", "a finding",
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

	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", first, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./internal/parser",
		"--successor", next,
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
	if _, err := run(t, "bench", "halt", "--run", runDir, "--seat-id", "judge-r1",
		"--reason", "continuing would compromise the consent gate"); err != nil {
		t.Fatalf("bench halt: %v", err)
	}
	if got := lastOfType(t, runDir, "halt").SeatID; got != "judge-r1" {
		t.Errorf("the halt is attributed to %q, want judge-r1 — an unattributed halt cannot be reviewed", got)
	}

	// A halt must NOT be reachable as a disposition on a ruling: that is the typo path
	// the separate verb exists to close off.
	id := mintGap(t, runDir, "not-haltable", "halt-is-its-own-verb")
	if _, err := run(t, "bench", "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", id, "--as", "halt", "--principle", "p", "--tension", "t",
		"--review-flag", "no", "--reason", "attempting to halt via a disposition"); err == nil {
		t.Error("`opinion --as halt` was accepted; ending the run must not be reachable by a mistyped disposition")
	}
}
