package fuzz

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A lens at its cap reads a second source for a sentence whose first reading a sibling's gap held;
// the sibling closes that gap between the drive's read of the board and the act, so the reading mints
// and the cap refuses it. The drive tallies that refusal rather than failing the run on it; with the
// gap still standing, the reading is held by it and admitted.
func TestASecondSourceRefusedForBudgetOnceItsHoldingGapClosedIsTallied(t *testing.T) {
	awaitsPhase1e(t)
	const (
		sentence = "A § fuzz sentence to anchor findings."
		lens     = "red-lens-logic"
		sibling  = "red-lens-adversary"
		class    = "claim_contradicts_source"
	)
	for _, closed := range []bool{true, false} {
		t.Run(map[bool]string{true: "closed in between", false: "still standing"}[closed], func(t *testing.T) {
			bin := buildBinary(t)
			runDir := recordtest.TmpRun(t)
			if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"),
				[]byte(`{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","lanes":1,"k":1,"kMax":2,"mintBudget":1,"convergenceFraction":0.25,"maxEpochs":10}`)); err != nil {
				t.Fatal(err)
			}
			stageRun, err := record.NewRun(runDir)
			if err != nil {
				t.Fatal(err)
			}
			cast, lanes := record.CastFor(1)
			if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
				t.Fatal(err)
			}
			r := newRunner(bin, runDir, newLockedRand(7))
			axes := func(outcome string, c *cmd) *cmd {
				return c.set("--as", outcome).set("--confidence", "high").set("--reason", "fuzz: what the source actually says")
			}
			if err := r.stagePart(lanePartText("blue-lane-1")); err != nil {
				t.Fatal(err)
			}
			mint := func(seat, quote, class string) string {
				t.Helper()
				out, ok := r.admit("--json", "mint", "--seat-id", seat, "--quote", quote, "--class", class,
					"--problem", "fuzz: the sentence overstates the trend", "--check-kind", "document", "--check", "acc",
					"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
				var env struct {
					Result struct {
						GapID string `json:"gap_id"`
					} `json:"result"`
				}
				if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
					t.Fatalf("%s's mint: %s (%s)", seat, out, r.actErr)
				}
				r.noteMint(seat, env.Result.GapID)
				return env.Result.GapID
			}
			r.sit("lens", sibling)
			// THE SIBLING'S GAP IS THE ONE ITS OWN READING MINTS: a claim_contradicts_source mint stands on
			// the minter's reading of the sentence (record.readingMintedClasses), so the sibling reads a
			// source it found against the sentence and its act mints the gap — a `document` gap, so that
			// its close asks no live-answer demand.
			sib, _, _, _ := r.foundSourceReading(sibling, "refutes", sentence, "https://fuzz.invalid/"+sibling+"/held", axes)
			setArg(sib.args, "--class", class)
			setArg(sib.args, "--check-kind", "document")
			sibOut, sibOK := sib.admit()
			sv := r.checkReading(sibling, sibOut, sib.args)
			if !sibOK || sv == nil {
				t.Fatalf("%s's reading was refused: %s (%s)", sibling, sibOut, r.actErr)
			}
			g := ""
			r.noteMint(sibling, g)
			r.sit("lens", lens)
			// THE LENS AT ITS CAP: its own mints on the other sentence, a class each.
			for _, c := range []string{"spec_underspecification", "unsourced_claim", "invalid_inference", "false_universal"} {
				if left, ok := r.mintsLeft(lens); !ok {
					t.Fatalf("%s's budget is not measured", lens)
				} else if left == 0 {
					break
				}
				mint(lens, repeatedSentenceSeeds[0], c)
			}
			if left, _ := r.mintsLeft(lens); left != 0 {
				t.Fatalf("%s has %d mints left; the fixture needs it at its cap", lens, left)
			}
			first, _, _, _ := r.foundSourceReading(lens, "refutes", sentence, "https://fuzz.invalid/"+lens+"/first", axes)
			setArg(first.args, "--class", class)
			out, ok := first.admit()
			if v := r.checkReading(lens, out, first.args); !ok || v == nil {
				t.Fatalf("at the cap, the first reading was not held by %s: %s (%s)", g, out, r.actErr)
			}
			if closed {
				r.betweenSecondSourceReadAndAct = func(id string) {
					r.admit("close", "--seat-id", sibling, "--id", id, "--as", "repaired", "--reason", "fuzz close as repaired",
						"--verified-by", sibling, "--verified-with", "fuzz", "--verified-against", "rec")
					r.betweenSecondSourceReadAndAct = nil
				}
			}
			before := eitherCount(t, "a second found-source reading whose holding gap closed in between, at the cap")
			r.readAnotherSource(lens, "refutes", sentence, class, g, axes)
			if r.actErr != "" {
				t.Fatalf("the drive failed the run: %s", r.actErr)
			}
			tallied := eitherCount(t, "a second found-source reading whose holding gap closed in between, at the cap") - before
			if want := map[bool]int{true: 1, false: 0}[closed]; tallied != want {
				t.Errorf("refusals tallied = %d, want %d", tallied, want)
			}
		})
	}
}

// eitherCount is how many acts the either tally holds refused for why.
func eitherCount(t *testing.T, why string) int {
	t.Helper()
	eitherMu.Lock()
	defer eitherMu.Unlock()
	n := 0
	for k, row := range eitherRow {
		if strings.HasSuffix(k, " — "+why) {
			n += row[1]
		}
	}
	return n
}

// The RUNSTAT line, which the job log keeps, names a refused act's cause: the envelope exec puts on
// the lines below the exit status is carried onto it.
func TestRunstatNamesTheRefusal(t *testing.T) {
	err := "an act the drive issued: refused, and the drive expects it admitted: verify --seat-id L: exit status 2\n" +
		`  {"verb":"verify","role":"lens","ok":false,"code":"validation","error":"L holds 1 gap(s) open and its mint budget admits no more."}`
	got := runstatErr(err)
	for _, want := range []string{"exit status 2", "refused (validation)", "its mint budget admits no more"} {
		if !strings.Contains(got, want) {
			t.Errorf("RUNSTAT err %q does not say %q", got, want)
		}
	}
	if strings.Contains(got, "\n") {
		t.Errorf("RUNSTAT err spans lines: %q", got)
	}
	if got := runstatErr("forced unsound computation: no plan readied"); got != "forced unsound computation: no plan readied" {
		t.Errorf("an error with no envelope = %q, want its first line", got)
	}
}
