package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// THE ARM THAT IS DISPATCHED, NOT THE ARM THAT IS RENDERED.
//
// internal/seatprobe.TestTheThreeArmsDiffer already asserts that Constitution() produces three
// different documents. It passed while the probe dispatched only two: armConstitution decided
// whether to CALL Constitution at all, and a short-circuit there returned the shipped file for
// `partial` — correct while `partial` meant the shipped file, silently wrong once it became a
// constructed treatment. Two arms, one treatment, and a between-arm difference of zero that would
// have been read as a finding about naming.
//
// So this tests the function the probe actually calls, on the constitutions it actually ships.
func TestEveryArmIsActuallyDispatchedDifferently(t *testing.T) {
	sf := seatprobe.NewSurface(cli.CommandPaths())
	agents := filepath.Join("..", "..", "..", "agents")
	src := filepath.Join(agents, "blue-researcher.md")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("constitutions not reachable from here: %v", err)
	}

	seen := map[string]seatprobe.Naming{}
	for _, arm := range seatprobe.NamingArms {
		p, err := armConstitution(src, t.TempDir(), sf, "blue", arm, false)
		if err != nil {
			t.Fatalf("%s: %v", arm, err)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s: %v", arm, err)
		}
		if other, dup := seen[string(b)]; dup {
			t.Errorf("arms %q and %q dispatch BYTE-IDENTICAL constitutions — the experiment would report no difference because there is none to report", other, arm)
		}
		seen[string(b)] = arm

		named := len(seatprobe.NamesSurviving(string(b), sf))
		switch arm {
		case seatprobe.NamingNone:
			if named > 0 {
				t.Errorf("the `none` arm dispatches a constitution naming %d verb(s) — the treatment did not happen", named)
			}
		case seatprobe.NamingPartial:
			if named == 0 {
				t.Errorf("the `partial` arm dispatches a constitution naming NO verbs — it is `none` wearing another label")
			}
			if !strings.Contains(string(b), "THE VERBS YOU WILL MOSTLY NEED") {
				t.Error("the `partial` arm's block is absent from what was dispatched")
			}
		case seatprobe.NamingComplete:
			if !strings.Contains(string(b), "YOUR COMPLETE VERB SURFACE") {
				t.Error("the `complete` arm's block is absent from what was dispatched")
			}
		}
	}
	if len(seen) != len(seatprobe.NamingArms) {
		t.Errorf("%d distinct constitutions for %d arms", len(seen), len(seatprobe.NamingArms))
	}
}

// THE SENTENCE THE REPORT PRINTS ABOUT THE ARM MUST MATCH THE ARM.
//
// namingTreatment exists to catch a manufactured null result: an arm whose treatment did not
// land, reported beside the finding it would otherwise fake. Its guards were written when every
// arm was subtractive, and it checked `before == 0` before it looked at the arm at all. The strip
// made `before == 0` true of every run and turned `partial` and `complete` additive — so the
// guard fired on all three, telling the reader that a null result in `complete` was a fixture
// artifact, on the one arm whose treatment was real.
//
// It is a report line rather than a computation, which is why nothing failed: the numbers behind
// it stayed correct while the sentence about them inverted. This pins the direction.
func TestTheNamingTreatmentLineReadsTheArmsDirection(t *testing.T) {
	sf := seatprobe.NewSurface(cli.CommandPaths())
	agents := filepath.Join("..", "..", "..", "agents")
	if _, err := os.Stat(filepath.Join(agents, "blue-researcher.md")); err != nil {
		t.Skipf("constitutions not reachable from here: %v", err)
	}
	for _, arm := range seatprobe.NamingArms {
		got := namingTreatment("blue", agents, sf, arm, false)
		if strings.Contains(got, "NOT MEASURED") {
			t.Fatalf("%s: %s", arm, got)
		}
		// The count is unconditional. A sentence interpreting the treatment is worth nothing if
		// it can be right about a number the reader never sees.
		if !strings.Contains(got, "under arm "+string(arm)) {
			t.Errorf("%s: the line reports no count: %s", arm, got)
		}
		removes := strings.Contains(got, "remove")
		if arm == seatprobe.NamingNone {
			if !removes {
				t.Errorf("`none` is the subtractive arm and its line never mentions removal: %s", got)
			}
			continue
		}
		if removes {
			t.Errorf("%s APPENDS a verb list; its line describes removal, which is the previous "+
				"experiment's direction printed over this one: %s", arm, got)
		}
		if strings.Contains(got, "ADDED NOTHING") {
			t.Errorf("%s: the arm's own treatment did not land: %s", arm, got)
		}
	}
}
