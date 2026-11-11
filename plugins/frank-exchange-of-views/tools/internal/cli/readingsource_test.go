package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// T12. A GAP NAMES THE READING IT STANDS ON THROUGH record.ReadingBehind, never a field: docket.md's
// block for a gap of a class a reading defines names the source read for the sentence — the cited url
// for a gap about a citation, the --url for a gap quoting the sentence a found source was read against
// — with the reader, the outcome and the reading's event; the report holds none of it.
func TestAGapNamesTheReadingItStandsOn(t *testing.T) {
	b := readingRun(t)
	cc := "claim_contradicts_source"
	type reading struct {
		seat string
		id   int64
	}
	read := func(seat string, args ...string) reading {
		b.mustRead(seat, args...)
		_, id := b.lastVerify()
		return reading{seat, id}
	}
	l := read(lensL, citeReading(b.c1, "refutes")...)
	g, out, err := b.mintSource(lensL, cc, about(b.c1)...)
	if err != nil {
		t.Fatalf("L's gap about C1: %v\n%s", err, out)
	}
	// N's reading of C1 refutes it too: N's gap about C1 is refused by the one-gap rule, and L's gap
	// discharges N's debt.
	read(lensN, citeReading(b.c1, "refutes")...)
	if _, out, err := b.mintSource(lensN, cc, about(b.c1)...); err == nil {
		t.Fatalf("N's second gap about C1 landed:\n%s", out)
	}
	m := read(lensM, urlReading(rdU4, rdFree1, "refutes")...)
	g2, out, err := b.mintSource(lensM, cc, "--quote", rdFree1)
	if err != nil {
		t.Fatalf("M's gap quoting the sentence: %v\n%s", err, out)
	}
	read(lensP, urlReading(rdU5, rdFree2, "supports")...)
	g3, err := b.plainMint(lensP, rdFree2, "unsourced_claim")
	if err != nil {
		t.Fatalf("P's unsourced_claim gap: %v", err)
	}
	read(lensK, citeReading(b.c2, "supports")...)

	raw, err := record.EvidenceJSONBytes(runtest.Open(t, b.runDir))
	if err != nil {
		t.Fatal(err)
	}
	var ev struct {
		Sources []struct {
			Verified []map[string]any `json:"verified"`
		} `json:"sources"`
		Independent []map[string]any `json:"independent"`
	}
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatal(err)
	}
	if n := len(ev.Independent); n != 2 {
		t.Errorf("show evidence lists %d independent readings, want M's and P's", n)
	}

	md := assembled(t, b.runDir)
	docket := readFile(t, filepath.Join(b.runDir, "docket.md"))
	line := func(url string, r reading, outcome string) string {
		return fmt.Sprintf("%s — %s's `%s` reading, event %d", url, r.seat, outcome, r.id)
	}
	for gap, lines := range map[string][]string{
		g:  {line(rdU1, l, "refutes")},
		g2: {line(rdU4, m, "refutes")},
	} {
		block := gapBlock(docket, gap)
		for _, w := range lines {
			if !strings.Contains(block, w) {
				t.Errorf("docket.md's block for %s does not name %q:\n%s", gap, w, block)
			}
			if strings.Contains(md, w) {
				t.Errorf("report.md names %q", w)
			}
		}
	}
	// A plain mint stands on no reading: its block names none.
	if block := gapBlock(docket, g3); strings.Contains(block, readForLabel) {
		t.Errorf("docket.md's block for the unsourced_claim gap %s names a reading:\n%s", g3, block)
	}
}

// readForLabel is docket.md's label for the line naming the reading a gap stands on.
const readForLabel = "read for the sentence"
