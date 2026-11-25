package fuzz

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// KEPTBYDEBT READS THE LINE THE PLAN PRINTS, not a line written to match it. The tally's debt gate
// counts the plans whose fold keeps a lens past its barren limit from retiring, and it reads them off
// the plan's why by keptByDebt: when the fold's account of a lens's pin was reworded, the pattern kept
// the old words, matched nothing, and the gate read zero folds kept in a sweep whose drive could not
// tell it from a lens that never owed. The board is the record's own fixture for the rule
// (TestALensThatOwesIsNotRetired): a computation nobody judged, three barren sittings of its owner
// each with a complete delivery, and a head move.
func TestKeptByDebtReadsThePlansLine(t *testing.T) {
	line := ownerPlanLine(t, true)
	g := keptByDebt.FindStringSubmatch(line)
	if g == nil {
		t.Fatalf("keptByDebt does not read the plan's line for a lens kept from retiring by a debt:\n%s", line)
	}
	if b := barrenSittings.FindStringSubmatch(line); b == nil || b[1] != strconv.Itoa(retiresAfterSittings) {
		t.Errorf("barrenSittings reads %v from the line of a lens %d sittings barren:\n%s", b, retiresAfterSittings, line)
	}
	if n, _ := strconv.Atoi(g[1]); n != 3 || !owesAComputation(g[2], map[string]bool{"K-1": true}) {
		t.Errorf("keptByDebt read %s barren sitting(s) owing %q, want 3 owing unchecked K-1, from:\n%s", g[1], g[2], line)
	}
	if !ownerAtBarrenLimit(line) {
		t.Errorf("ownerAtBarrenLimit does not read the owner at its barren limit from:\n%s", line)
	}
}

// OWEDCOMPUTATIONSEED'S TRIGGER READS THE LINE A PLAN PRINTS FOR A RETIRED OWNER: the same board without the
// computation, so the owner, three sittings barren, owes nothing and retires. computeOwed proves its
// computation on this line; a matcher that missed it would leave the forced run proving nothing.
func TestOwnerAtBarrenLimitReadsARetiredOwner(t *testing.T) {
	line := ownerPlanLine(t, false)
	if !strings.HasPrefix(line, computationOwner+": retired (pin ") || !ownerAtBarrenLimit(line) {
		t.Errorf("ownerAtBarrenLimit does not read a retired owner from the plan's line:\n%s", line)
	}
	if keptByDebt.MatchString(line) {
		t.Errorf("keptByDebt reads a retired lens as kept from retiring:\n%s", line)
	}
}

// ownerPlanLine is the computation owner's line of the plan on the record's fixture for the rule
// (TestALensThatOwesIsNotRetired): three barren sittings of the owner each with a complete delivery,
// and a head move; with a computation nobody judged where computed is set.
func ownerPlanLine(t *testing.T, computed bool) string {
	t.Helper()
	const owner = computationOwner
	dir := recordtest.TmpRun(t)
	run, err := record.NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(run.Records(), 0o755); err != nil {
		t.Fatal(err)
	}
	var evs []*recordpb.Event
	add := func(seat string, body proto.Message) {
		evs = append(evs, recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, len(evs)+1), body))
	}
	add(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{owner, "red-chair", "blue-respond", "judge"}})
	add(record.HarnessSeat, &recordpb.ReportPart{Text: proto.String("# report")})
	add("red-chair", &recordpb.Register{})
	if computed {
		add("blue-respond", &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
			ComputationId: proto.String("K-1"), BundleSha: proto.String("s1")})
	} else {
		add("blue-respond", &recordpb.Register{})
	}
	for range 3 {
		add("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(owner)})
		add(owner, &recordpb.Register{})
		add(owner, &recordpb.ReportRead{Head: proto.Int64(2)})
	}
	add(record.HarnessSeat, &recordpb.ReportPart{Text: proto.String("# report")})
	add("red-chair", &recordpb.Register{})
	recordtest.Seed(t, dir, append(evs, recordtest.SettledQuestion(t)...)...)

	plan, err := record.PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range plan.Why {
		if strings.HasPrefix(w, owner+": ") {
			return w
		}
	}
	t.Fatalf("the plan names no line for %s: %q", owner, plan.Why)
	return ""
}

// OWEDCOMPUTATIONSEED'S RUN RECORDS NO COMPUTATION BUT computeOwed's: every other is refused before it runs
// and fails the run by name, whichever draw issued it.
func TestOwedComputationSeedRefusesEveryOtherComputation(t *testing.T) {
	r := &runner{forceOwedComputation: true}
	if _, err := r.exec("compute", "--seat-id", "blue-respond", "--quote", "§ fuzz sentence", "--dir", "d", "--", "node", "compute"); err == nil {
		t.Fatal("a computation outside computeOwed was not refused in owedComputationSeed's run")
	}
	if !strings.Contains(r.actErr, "drive defect") || !strings.Contains(r.actErr, "computeOwed") {
		t.Errorf("the refusal does not fail the run naming computeOwed: %q", r.actErr)
	}
	if isCompute([]string{"edit", "--seat-id", "blue-respond", "--", "compute"}) {
		t.Error("isCompute reads a computation's own argument after `--` as the verb")
	}
}
