package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// 31. A READING IS ATTRIBUTED TO THE VERB THE SEAT TYPED: a reading is `verify` and mints nothing, and
// the gap the seat mints on it is the seat's `mint`.
func TestAReadingIsAttributedToTheVerbTheSeatTyped(t *testing.T) {
	b := readingRun(t)
	sf := seatprobe.NewSurface(CommandPaths())
	run := runtest.Open(t, b.runDir)
	read := func(seat string) *seatprobe.Choices {
		t.Helper()
		c, err := seatprobe.Read(sf, run, seat)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	cc := "claim_contradicts_source"
	b.mustRead(lensL, citeReading(b.c1, "refutes")...)
	// A second reading, so `verify` outnumbers the report read the fixture served and is the
	// substitute the probe names.
	b.mustRead(lensL, citeReading(b.c2, "supports")...)
	c := read(lensL)
	if c.Used["verify"] != 2 || c.Used["mint"] != 0 || !slices.Contains(c.Unused, "mint") {
		t.Errorf("L's refuting reading: used %v, unused %v; want verify 2 and mint unused", c.Used, c.Unused)
	}
	vs, err := seatprobe.Check(sf, run, []seatprobe.Expectation{{Seat: lensL, Verb: "mint", Because: "the test asks"}})
	if err != nil {
		t.Fatal(err)
	}
	if vs[0].Met || !strings.HasPrefix(vs[0].Substitute, "verify") {
		t.Errorf("an expectation of mint = %+v, want unmet with verify the substitute", vs[0])
	}
	if _, out, err := b.mintSource(lensL, cc, about(b.c1)...); err != nil {
		t.Fatalf("L's gap on its reading: %v\n%s", err, out)
	}
	if c := read(lensL); c.Used["mint"] != 1 || c.Used["verify"] != 2 {
		t.Errorf("L's mint on its reading: used %v, want mint 1 and verify 2", c.Used)
	}
	if slices.Contains(sf.Verbs("lens"), "corroborate") || slices.Contains(c.Unused, "corroborate") || c.Used["corroborate"] != 0 {
		t.Errorf("the probe's lens verbs hold corroborate: %v", sf.Verbs("lens"))
	}
}
