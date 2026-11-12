package vecembed

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

// These run against the built feov-embed through Require, on every native-run target, and build no
// fake helper: native-run's servercore container runs this package's test binary with no Go
// present. The three that measure run only where FEOV_NATIVE_MEASURE=1 (native-run's dedicated
// runners); each logs its numbers on one line that starts with its bar's name.

// requestBound is a quarter of the client's deadline: every request the client can send finishes
// inside it on the slowest target, so a slower runner still has three quarters in hand.
const requestBound = nativehelper.DefaultDeadline / 4

// The S5 and S6 bars (plans/vector-search.md §I.4, G2 and G3).
const (
	s5MeanCosine = 0.984
	s5MinCosine  = 0.966
	s6ChunksPerS = 5.0
	s6ColdQuery  = 1500 * time.Millisecond
)

// s6Targets are the targets S6 holds to its bars; the others report their speed.
var s6Targets = map[string]bool{"linux-amd64": true, "linux-arm64": true, "darwin-arm64": true}

func target() string { return runtime.GOOS + "-" + runtime.GOARCH }

func measured(t *testing.T) {
	t.Helper()
	if os.Getenv("FEOV_NATIVE_MEASURE") != "1" {
		t.Skip("SKIPPED, NOT PASSED: this measures the built helper; set FEOV_NATIVE_MEASURE=1 (native-run does)")
	}
}

// TestEmbedLoadsTheExportForItsArch: the helper's handshake reports the sha256 of the model it
// loaded, and the client accepts it only when it is this GOARCH's pin; the export installed beside
// the helper is that pin's file.
func TestEmbedLoadsTheExportForItsArch(t *testing.T) {
	requireEmbed(t)
	pins := nativehelper.HelperAssets(HelperName)
	if len(pins) != 1 {
		t.Fatalf("feov-embed carries %d asset pins for %s, want the one gte-small export", len(pins), runtime.GOARCH)
	}
	b, err := os.ReadFile(filepath.Join(os.Getenv(nativehelper.EnvDir), "gte-small.int8.ort"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != pins[0] {
		t.Errorf("the installed gte-small.int8.ort is %s; %s's pin is %s", got, runtime.GOARCH, pins[0])
	}
	t.Logf("export %s: handshake model_sha256 = pin %s", target(), pins[0])
}

// TestEmbedFidelityMeetsS5: the helper's embeddings of the 300 reference texts lie within S5 of
// FP32 gte-small's (ref_gte.npy): mean cosine at least 0.984, min at least 0.966.
func TestEmbedFidelityMeetsS5(t *testing.T) {
	measured(t)
	e := NewWithClient(requireEmbed(t))
	in, ref := speedTexts(t), refVectors(t)
	got, err := e.Embed(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(ref) {
		t.Fatalf("%d vectors for %d texts", len(got), len(ref))
	}
	sum, lo, at := 0.0, math.Inf(1), -1
	for i := range got {
		v := append([]float32(nil), got[i]...)
		normalise(v)
		c := float64(Cosine(v, ref[i]))
		sum += c
		if c < lo {
			lo, at = c, i
		}
	}
	mean := sum / float64(len(got))
	t.Logf("S5 target=%s n=%d mean=%.6f min=%.6f min_text=%d bar_mean=%.3f bar_min=%.3f %s", target(), len(got), mean, lo, at, s5MeanCosine, s5MinCosine, cpuDesc())
	if mean < s5MeanCosine {
		t.Errorf("mean cosine %.6f, under %.3f", mean, s5MeanCosine)
	}
	if lo < s5MinCosine {
		t.Errorf("min cosine %.6f (text %d), under %.3f", lo, at, s5MinCosine)
	}
}

// TestEmbedSpeedMeetsS6: a cold query (spawn, handshake with the model's hash and session, one
// embed) takes at most 1.5 s, and the 100 chunks of len200.jsonl embed at 5 chunks a second or
// better, on one thread. linux-amd64, linux-arm64 and darwin-arm64 are held to the bars; every
// other target reports its numbers and passes.
func TestEmbedSpeedMeetsS6(t *testing.T) {
	measured(t)
	chunks := len200Texts(t)
	start := time.Now()
	e := NewWithClient(requireEmbed(t))
	if _, err := e.Embed([]string{"which transit timing variations constrain the outer planet's mass"}); err != nil {
		t.Fatal(err)
	}
	cold := time.Since(start)
	start = time.Now()
	got, err := e.Embed(chunks)
	if err != nil {
		t.Fatal(err)
	}
	d := time.Since(start)
	if len(got) != len(chunks) {
		t.Fatalf("%d vectors for %d chunks", len(got), len(chunks))
	}
	rate := float64(len(chunks)) / d.Seconds()
	held := s6Targets[target()]
	t.Logf("S6 target=%s chunks=%d secs=%.2f chunks_per_s=%.2f cold_query_ms=%d held=%v bar_chunks_per_s=%.1f bar_cold_ms=%d %s",
		target(), len(chunks), d.Seconds(), rate, cold.Milliseconds(), held, s6ChunksPerS, s6ColdQuery.Milliseconds(), cpuDesc())
	if !held {
		return
	}
	if rate < s6ChunksPerS {
		t.Errorf("%.2f chunks/s, under %.1f", rate, s6ChunksPerS)
	}
	if cold > s6ColdQuery {
		t.Errorf("a cold query took %s, over %s", cold.Round(time.Millisecond), s6ColdQuery)
	}
}

// TestEveryEmbedOpFitsTheDeadline sends every op in this package's op table its largest client request,
// BatchSize texts of MaxIDs ids each, and times spawn to handshake reply (the model's hash and the
// session included) first. Each must finish under a quarter of the deadline. An op in the table
// with no case here fails, so a new op cannot pass by being absent. It measures time, so it runs
// only on native-run's dedicated runners, where FEOV_NATIVE_MEASURE=1.
func TestEveryEmbedOpFitsTheDeadline(t *testing.T) {
	measured(t)
	start := time.Now()
	c := requireEmbed(t)
	if d := time.Since(start); d > requestBound {
		t.Errorf("spawn to handshake reply took %s, over %s", d, requestBound)
	} else {
		t.Logf("deadline target=%s handshake_ms=%d bound=%s", target(), d.Milliseconds(), requestBound)
	}
	e := NewWithClient(c)

	cases := map[string]func() error{
		"embed": func() error {
			text := strings.Join(len200Texts(t)[:4], " ")
			if n := len(DefaultTokenizer().Encode(text)); n != MaxIDs {
				t.Fatalf("the largest request's text is %d ids, want %d", n, MaxIDs)
			}
			batch := make([]string, BatchSize)
			for i := range batch {
				batch[i] = text
			}
			_, err := e.Embed(batch)
			return err
		},
	}
	for _, op := range sortedOps() {
		name := Ops[op]
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
			t.Errorf("%s on its largest request: %v", name, err)
		case d > requestBound:
			t.Errorf("%s on its largest request took %s, over %s", name, d.Round(time.Millisecond), requestBound)
		default:
			t.Logf("deadline target=%s op=%s request=%dx%d ms=%d", target(), name, BatchSize, MaxIDs, d.Milliseconds())
		}
	}
}

func sortedOps() []nativehelper.Op {
	ops := make([]nativehelper.Op, 0, len(Ops))
	for op := range Ops {
		ops = append(ops, op)
	}
	slices.Sort(ops)
	return ops
}

// requireEmbed is a running feov-embed for the test, or the test's end (nativehelper.Require).
func requireEmbed(t *testing.T) *nativehelper.Client {
	t.Helper()
	return nativehelper.Require(t, HelperName)
}
