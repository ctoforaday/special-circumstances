package tessocr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The detector boundary set (plan §V.5, pinned by the Wave 0 300-DPI tune in §VI),
// tested against the MEASURED per-page counts rather than eleven checked-in page scans:
// testdata/grid300-sel151.txt is the tune's own record (SEL=151 over the full 80-page
// corpus), and the threshold function is what these fixtures exercise. One real image
// keeps the cgo pixel path honest end to end (TestDetectGridOnRealCrop).
//
// p0025 is pinned as FIRES-AND-DEGRADES-HONESTLY: it is the tune's standing false
// positive (boxed text, structurally inseparable from a table by line-pixel counts), and
// its assertion here is that the detector DOES fire — the fallback stating itself — not a
// clean rejection this feature set cannot deliver.
var boundaryPages = []struct {
	page  string
	table bool
	why   string
}{
	{"p0066", true, "h within 12% of threshold — nearest table on the h axis"},
	{"p0028", true, "corrected-truth one-row strip; v nearest after p0066"},
	{"p0044", true, "thinnest margin in the corpus: i=104, 4 px over i>=100"},
	{"p0063", true, "v-axis boundary table"},
	{"p0038", true, "i-axis boundary table"},
	{"p0062", true, "i-axis boundary table"},
	{"p0025", false, "FIRES-AND-DEGRADES-HONESTLY — see below; truth is non-table, verdict is table"},
	{"p0071", false, "closest clean rejection: fails only h"},
	{"p0016", false, "vertical rules only, v=60983 h=0 — the AND is load-bearing"},
	{"p0022", false, "vertical rules only"},
	{"p0080", false, "vertical rules only"},
	{"p0002", false, "sub-threshold on every axis"},
}

func TestGrid300BoundarySet(t *testing.T) {
	counts := readGridFixture(t)
	for _, bp := range boundaryPages {
		s, ok := counts[bp.page]
		if !ok {
			// A pinned page missing from the fixture is a broken record, not a clean
			// pass — the loud-miss rule.
			t.Fatalf("%s is pinned but absent from grid300-sel151.txt", bp.page)
		}
		want := bp.table
		if bp.page == "p0025" {
			want = true // the standing false positive fires, by design
		}
		if got := Grid300.Table(s); got != want {
			t.Errorf("%s (%s): Table(%+v) = %v, want %v", bp.page, bp.why, s, got, want)
		}
	}
}

// Both directions on the measured boundaries — the operator-flip-rich spots plan §V.5
// names. Each case moves ONE measured axis just across its threshold and demands the
// verdict move with it; a >= flipped to > (or a threshold typo) fails here before it
// costs a corpus re-run.
func TestGrid300ThresholdDirections(t *testing.T) {
	counts := readGridFixture(t)
	p0044 := counts["p0044"] // i=104, the 4-px margin
	if !Grid300.Table(p0044) {
		t.Fatalf("p0044 %+v must fire at the final tune", p0044)
	}
	under := p0044
	under.Intersections = Grid300.MinIntersections - 1
	if Grid300.Table(under) {
		t.Errorf("p0044 with i=%d must NOT fire — the intersection threshold is load-bearing", under.Intersections)
	}
	exact := p0044
	exact.Intersections = Grid300.MinIntersections
	if !Grid300.Table(exact) {
		t.Errorf("i exactly at the threshold must fire: the tune is >=, not >")
	}

	p0066 := counts["p0066"]
	under = p0066
	under.HPix = Grid300.MinHPix - 1
	if Grid300.Table(under) {
		t.Errorf("p0066 with h=%d must NOT fire — the h threshold is load-bearing", under.HPix)
	}
	under = p0066
	under.VPix = Grid300.MinVPix - 1
	if Grid300.Table(under) {
		t.Errorf("p0066 with v=%d must NOT fire — the v threshold is load-bearing", under.VPix)
	}

	// The conjunction, from the direction that defeats a single-axis detector: p0016's
	// massive v signal must not fire however large, because h is zero.
	p0016 := counts["p0016"]
	if p0016.VPix < Grid300.MinVPix || p0016.HPix != 0 {
		t.Fatalf("fixture drift: p0016 = %+v, expected the v-only page", p0016)
	}
	if Grid300.Table(p0016) {
		t.Errorf("p0016 fired on vertical rules alone — the h∧v conjunction broke")
	}
}

// The full-corpus tally of the final tune, recomputed from the fixture record: 35 pages
// fire — the 33 labeled tables, p0025 (the standing false positive), and p0015 (the
// unlabeled filter-blocked page the tune reads as a table) — and 45 do not. A threshold
// drift that spares every pinned boundary page still moves this total.
func TestGrid300CorpusTotals(t *testing.T) {
	counts := readGridFixture(t)
	if len(counts) != 80 {
		t.Fatalf("fixture carries %d pages, want 80", len(counts))
	}
	fired := 0
	for _, s := range counts {
		if Grid300.Table(s) {
			fired++
		}
	}
	if fired != 35 {
		t.Errorf("final tune fires on %d/80 pages, want 35 (33 tables + p0025 + p0015)", fired)
	}
}

// readGridFixture parses the Wave 0 detector record. The format is the tune's own output
// (`p0001.png hpix=… vpix=… inter=… ms=…`); any line that does not parse is an error,
// because a silently skipped line would shrink the corpus without failing anything.
func readGridFixture(t *testing.T) map[string]GridStats {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "grid300-sel151.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out := map[string]GridStats{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var page string
		var s GridStats
		var ms int
		if _, err := fmt.Sscanf(line, "%s hpix=%d vpix=%d inter=%d ms=%d",
			&page, &s.HPix, &s.VPix, &s.Intersections, &ms); err != nil {
			t.Fatalf("unparseable fixture line %q: %v", line, err)
		}
		out[strings.TrimSuffix(page, ".png")] = s
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
