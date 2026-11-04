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

// 29. A READING'S GAP NAMES ITS SOURCE (F12): each reading carrying a gap is on its evidence row as
// carried_by, and docket.md's block for the gap names the source, the reader, the outcome and the
// reading's event — how blue finds the source a lens read. The report holds none of it.
func TestAReadingsGapNamesItsSource(t *testing.T) {
	b := readingRun(t)
	cc := "claim_contradicts_source"
	type reading struct {
		seat string
		id   int64
	}
	read := func(seat string, args ...string) (string, reading) {
		r := b.mustRead(seat, args...)
		_, id := b.lastVerify()
		g := ""
		if r.Gap != nil {
			g = r.Gap.ID
		}
		return g, reading{seat, id}
	}
	g, l := read(lensL, citeReading(b.c1, "refutes", cc)...)
	gn, n := read(lensN, citeReading(b.c1, "refutes", cc)...)
	g2, m := read(lensM, urlReading(rdU4, rdFree1, "refutes", cc)...)
	g3, p := read(lensP, urlReading(rdU5, rdFree2, "supports", "unsourced_claim")...)
	gk, _ := read(lensK, citeReading(b.c2, "supports", "")...)
	if gn != g || g2 == "" || g3 == "" || gk != "" {
		t.Fatalf("gaps: L %s N %s M %s P %s K %q", g, gn, g2, g3, gk)
	}

	raw, err := record.EvidenceJSONBytes(runtest.Open(t, b.runDir))
	if err != nil {
		t.Fatal(err)
	}
	var ev struct {
		Sources []struct {
			Anchor   string           `json:"anchor"`
			Verified []map[string]any `json:"verified"`
		} `json:"sources"`
		Independent []map[string]any `json:"independent"`
	}
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatal(err)
	}
	carried := map[string]any{}
	for _, s := range ev.Sources {
		for _, v := range s.Verified {
			carried[s.Anchor+"/"+fmt.Sprint(v["seat_id"])] = v["carried_by"]
		}
	}
	for _, v := range ev.Independent {
		carried["independent/"+fmt.Sprint(v["seat_id"])] = v["carried_by"]
	}
	for k, want := range map[string]any{
		b.c1 + "/" + lensL: g, b.c1 + "/" + lensN: g,
		"independent/" + lensM: g2, "independent/" + lensP: g3,
	} {
		if carried[k] != want {
			t.Errorf("show evidence's row %s carries %v, want %v", k, carried[k], want)
		}
	}
	if v, ok := carried[b.c2+"/"+lensK]; !ok || v != nil {
		t.Errorf("K's backing row carries %v (present %v), want no carried_by key", v, ok)
	}

	md := assembled(t, b.runDir)
	docket := readFile(t, filepath.Join(b.runDir, "docket.md"))
	line := func(url string, r reading, outcome string) string {
		return fmt.Sprintf("%s — %s's `%s` reading, event %d", url, r.seat, outcome, r.id)
	}
	for gap, lines := range map[string][]string{
		g:  {line(rdU1, l, "refutes"), line(rdU1, n, "refutes")},
		g2: {line(rdU4, m, "refutes")},
		g3: {line(rdU5, p, "supports")},
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
}
