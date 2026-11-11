package fetchcache

import (
	"bytes"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/corpus"
)

// requestBound is a quarter of the client's deadline: every request the client can send finishes
// inside it on the slowest target, so a slower runner still has three quarters in hand.
const requestBound = nativehelper.DefaultDeadline / 4

// TestEveryOpFitsTheDeadline sends every op in tessocr's op table its largest client request, the
// largest corpus page rendered at MaxRenderDPI and given whole, and times spawn to handshake reply
// first. Each must finish under a quarter of the deadline. An op in the table with no case here
// fails, so a new op cannot pass by being absent. It measures time, so it runs only on native-run's
// dedicated runners, where FEOV_NATIVE_MEASURE=1.
func TestEveryOpFitsTheDeadline(t *testing.T) {
	if os.Getenv("FEOV_NATIVE_MEASURE") != "1" {
		t.Skip("SKIPPED, NOT PASSED: this measures request time; set FEOV_NATIVE_MEASURE=1 (native-run does)")
	}
	start := time.Now()
	nativehelper.Require(t, "feov-ocr")
	if d := time.Since(start); d > requestBound {
		t.Errorf("spawn to handshake reply took %s, over %s", d, requestBound)
	} else {
		t.Logf("handshake: %s", d.Round(time.Millisecond))
	}

	page, w, h := largestCorpusPage(t)
	thr := tessocr.GridFor(MaxRenderDPI)
	en, err := tessocr.New()
	if err != nil {
		t.Fatal(err)
	}
	defer en.Close()

	cases := map[string]func() error{
		"tsv": func() error {
			_, err := en.PageTSV(page, tessocr.PSMAuto)
			return err
		},
		"rot_band": func() error {
			_, err := en.RotatedBand(page, 0, 0, w, h)
			return err
		},
		"grid_stats": func() error {
			_, err := tessocr.DetectGrid(page, thr)
			return err
		},
		"grid_lines": func() error {
			_, err := tessocr.GridLines(page, thr)
			return err
		},
		"repaired_rules": func() error {
			_, err := tessocr.RepairedRules(page, thr)
			return err
		},
		"ink": func() error {
			_, err := tessocr.InkFraction(page)
			return err
		},
	}
	for _, name := range tessocr.Ops {
		if strings.HasPrefix(name, "selftest-") {
			continue
		}
		run, ok := cases[name]
		if !ok {
			t.Errorf("op %s has no largest-request case: add one here", name)
			continue
		}
		start := time.Now()
		err := run()
		d := time.Since(start)
		switch {
		case err != nil:
			t.Errorf("%s on the largest page: %v", name, err)
		case d > requestBound:
			t.Errorf("%s on the largest page took %s, over %s", name, d.Round(time.Millisecond), requestBound)
		default:
			t.Logf("%s: %s", name, d.Round(time.Millisecond))
		}
	}
}

// largestCorpusPage renders every corpus page at MaxRenderDPI and returns the PNG with the most
// pixels, and its size.
func largestCorpusPage(t *testing.T) ([]byte, int, int) {
	t.Helper()
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	var best []byte
	bw, bh, slug := 0, 0, ""
	for _, c := range cases {
		run, sha, _, _ := renderCorpusPage(t, c)
		body, err := os.ReadFile(Path(run, sha))
		if err != nil {
			t.Fatal(err)
		}
		rd, err := RenderPages(run, sha, body, AtDPI(MaxRenderDPI))
		if err != nil {
			t.Fatalf("%s at %d DPI: %v", c.Slug, MaxRenderDPI, err)
		}
		r := rd.Renders[0]
		if r.WidthPx*r.HeightPx <= bw*bh {
			continue
		}
		b, err := os.ReadFile(PagePath(run, sha, 1))
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		best, bw, bh, slug = b, cfg.Width, cfg.Height, c.Slug
	}
	if best == nil {
		t.Fatal("the corpus holds no page")
	}
	t.Logf("largest corpus page at %d DPI: %s, %dx%d px, %d bytes of PNG", MaxRenderDPI, slug, bw, bh, len(best))
	return best, bw, bh
}
