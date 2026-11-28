package fuzz

import (
	"maps"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE DRIVE'S MODEL OF THE BENCH'S GRADES FOLLOWS THE LINEAGE AT EVERY SUPERSEDING MINT, the found-source
// mint over a prescription among them (readOverPrescription). The record holds a successor to each axis
// the bench holds locked on its ancestor, and admits the successor only at the bench's grade: that
// mint states high on every axis, so it lands exactly where the bench holds the ancestor's axis at
// high — the one case in which the successor exists, locked on the record. The lens's drawn regrade of
// it then keeps the axis only where the model holds the lock too; held free, the regrade draws a grade
// that moves the axis and the record refuses it.
//
// The fixture is that case: a gap whose prescription blue applied verbatim, its likelihood moved to
// high by the bench at a hearing that remands it, and the minting lens's reading of a source it found
// against the applied sentence. The lens's draws after the hearing are the ones that reach the lock:
// the drawn regrade takes likelihood, at medium_high.
func TestASuccessorMintedOverAPrescriptionCarriesTheBenchsGradesInTheDrive(t *testing.T) {
	const (
		sentence   = "The cost is rising over time across the whole sampled corpus."
		prescribed = "climbing sharply"
		blue       = "blue-respond"
	)
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"), fuzzRunConfig(1, 1, 2, 10)); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: []string{liftLens, "red-chair", "blue-lane-1", "blue-synthesize", blue, "judge"}, LaneSeatIds: []string{"blue-lane-1"}}); err != nil {
		t.Fatal(err)
	}
	recordtest.SettleQuestion(t, stageRun.Dir()) // the question is not this fixture's subject
	// Every coin lands tails, and every drawn grade is high: the bench's move of likelihood is to high.
	r := newRunner(bin, runDir, &lockedRand{r: rand.New(steadySource(99))})
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}

	// THE LENS PRESCRIBES, at a likelihood the bench has a move to high from.
	r.chairPlans(t)
	r.dispatchSits("lens", liftLens)
	r.mustExec(t, "--json", "--seat-id", liftLens, "show", "report")
	r.mustExec(t, "--json", "mint", "--seat-id", liftLens, "--quote", "rising over time", "--new", prescribed,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", dirApply, "--severity", "high", "--impact", "high", "--likelihood", "medium", "--reason", "fuzz reason")
	r.dispatchReturns(liftLens, false)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	prescribing := open[0]
	r.noteMint(liftLens, prescribing)

	// BLUE APPLIES THE PRESCRIPTION VERBATIM, and puts the gap before the bench.
	r.dispatchSits("blue", blue)
	if _, ok := r.do("edit", blue).set("--answers", prescribing).bare("--accept").
		set("--reason", "fuzz: accepting red's prescribed fix on "+prescribing+" exactly as recorded").admit(); !ok {
		t.Fatalf("blue's accept of the prescription: %s", r.actErr)
	}
	r.noteVerbatimGap(prescribing, prescribed)
	r.mustExec(t, "motion", "docket", "file", "--seat-id", blue, "--id", prescribing, "--reason", "fuzz: the bench's to settle")
	r.dispatchReturns(blue, false)
	if !r.stillEstops(prescribing) {
		t.Fatalf("blue's accept of %s does not estop red on the applied sentence", prescribing)
	}

	// THE HEARING: the bench moves likelihood to high, which locks it, and remands the gap.
	for _, party := range []struct{ role, seat string }{{"blue", blue}, {"lens", liftLens}} {
		if len(r.closingsOwedBy(party.seat)) == 0 {
			continue
		}
		r.dispatchSits(party.role, party.seat)
		r.fileClosings(party.seat)
		r.dispatchReturns(party.seat, false)
	}
	r.chairPlans(t)
	r.dispatchSits("bench", "judge", "docket")
	r.benchMoves("judge", prescribing, "likelihood")
	r.benchDisposes("judge", prescribing, "remanded", "--direction", "fuzz: one more exchange", "--reason", "fuzz: remanded for "+prescribing)
	r.dispatchReturns("judge", false)
	if r.actErr != "" {
		t.Fatalf("the bench's hearing: %s", r.actErr)
	}
	want := map[string]string{"likelihood": "high"}
	if held := r.benchGrades(prescribing); !maps.Equal(held, want) {
		t.Fatalf("after the hearing the bench holds %v of %s, want %v", held, prescribing, want)
	}

	// THE LENS READS A SOURCE IT FOUND AGAINST THE APPLIED SENTENCE, and its gap on that reading
	// supersedes the prescribing gap. FROM HERE THE DRAWS REACH THE LOCK: 88 takes the second of the
	// drawn regrade's three axes, likelihood, and the fourth of the five grades, medium_high; every
	// coin still lands tails.
	r.rng = &lockedRand{r: rand.New(steadySource(88))}
	r.chairPlans(t)
	r.dispatchSits("lens", liftLens)
	r.mustExec(t, "--json", "--seat-id", liftLens, "show", "report")
	cur, err := record.RenderRun(r.run())
	if err != nil {
		t.Fatal(err)
	}
	applied := ""
	for _, line := range strings.Split(string(cur), "\n") {
		if strings.Contains(plainText(line), prescribed) {
			applied = line
		}
	}
	if applied == "" {
		t.Fatalf("the report holds no line carrying the applied text %q:\n%s", prescribed, cur)
	}
	r.readOverPrescription(liftLens, applied, prescribing)
	if r.actErr != "" {
		t.Fatalf("the reading over the prescription: %s", r.actErr)
	}
	open = r.openGaps()
	if len(open) != 1 || open[0] == prescribing {
		t.Fatalf("the board holds %v open after the reading over %s, want its one successor", open, prescribing)
	}
	successor := open[0]
	if r.minterOf(successor) != liftLens {
		t.Fatalf("the drive holds %q the minter of %s, want %s", r.minterOf(successor), successor, liftLens)
	}

	// THE MODEL HOLDS WHAT THE RECORD HOLDS, and the drawn regrade keeps the bench's axis.
	if held := r.benchGrades(successor); !maps.Equal(held, want) {
		t.Errorf("the drive's model holds %v of the bench's on successor %s, want %v: the record holds the successor to its ancestor's lock", held, successor, want)
	}
	r.drawnRegrade(liftLens, successor)
	if r.actErr != "" {
		t.Errorf("the lens's drawn regrade of %s: %s", successor, r.actErr)
	}
}
