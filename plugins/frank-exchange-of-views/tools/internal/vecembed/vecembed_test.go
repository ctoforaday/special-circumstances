package vecembed

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
)

func TestTokenizerSpecialTokens(t *testing.T) {
	tok := DefaultTokenizer()
	if tok.cls != 101 {
		t.Fatalf("tok.cls = %d, want 101", tok.cls)
	}
	if tok.sep != 102 {
		t.Fatalf("tok.sep = %d, want 102", tok.sep)
	}
	if tok.unk != 100 {
		t.Fatalf("tok.unk = %d, want 100", tok.unk)
	}
}

func TestTokenizerEncode(t *testing.T) {
	tok := DefaultTokenizer()
	tests := []struct {
		text string
		want []int64
	}{
		{
			text: "hello world",
			want: []int64{101, 7592, 2088, 102},
		},
		{
			text: "Special Circumstances",
			want: []int64{101, 2569, 6214, 102},
		},
		{
			text: "Hello, World! 123",
			want: []int64{101, 7592, 1010, 2088, 999, 13138, 102},
		},
	}

	for _, tc := range tests {
		got := tok.Encode(tc.text)
		if len(got) != len(tc.want) {
			t.Fatalf("Encode(%q) len = %d (%v), want %d (%v)", tc.text, len(got), got, len(tc.want), tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("Encode(%q)[%d] = %d, want %d", tc.text, i, got[i], tc.want[i])
			}
		}
	}
}

func TestTokenizerTruncation(t *testing.T) {
	tok := DefaultTokenizer()
	// Long text
	long := "word "
	for i := 0; i < 600; i++ {
		long += "word "
	}
	ids := tok.Encode(long)
	if len(ids) > 512 {
		t.Fatalf("len(ids) = %d, want <= 512", len(ids))
	}
	if ids[0] != tok.cls {
		t.Fatalf("ids[0] = %d, want [CLS] (%d)", ids[0], tok.cls)
	}
	if ids[len(ids)-1] != tok.sep {
		t.Fatalf("ids[last] = %d, want [SEP] (%d)", ids[len(ids)-1], tok.sep)
	}
}

// A WORD THAT CROSSES THE LAST ID IS CUT AT IT. 509 one-piece words fill ids 1-509; the next word's
// pieces run past id 510, the last before [SEP], and only its first piece is kept: 512 ids, never
// the 513 or more the helper would refuse.
func TestTokenizerTruncatesAWordThatCrossesTheLastID(t *testing.T) {
	tok := DefaultTokenizer()
	const crossing = "unaffordability"
	pieces := tok.wordpiece(crossing, nil)
	if len(pieces) < 2 {
		t.Fatalf("%q is %d wordpieces; the test needs a word of two or more", crossing, len(pieces))
	}
	ids := tok.Encode(strings.Repeat("word ", 509) + crossing)
	if len(ids) != tok.maxLen {
		t.Fatalf("len(ids) = %d, want %d", len(ids), tok.maxLen)
	}
	if ids[509] != tok.Encode("word")[1] || ids[510] != pieces[0] || ids[511] != tok.sep {
		t.Errorf("ids[509:] = %v, want [word %d [SEP]]", ids[509:], pieces[0])
	}
}

func TestTokenizerEdgeCases(t *testing.T) {
	tok := DefaultTokenizer()

	// Empty string
	got := tok.Encode("")
	if len(got) != 2 || got[0] != tok.cls || got[1] != tok.sep {
		t.Fatalf("Encode(\"\") = %v, want [%d, %d]", got, tok.cls, tok.sep)
	}

	// Whitespace only
	gotWS := tok.Encode("   \t\n  ")
	if len(gotWS) != 2 || gotWS[0] != tok.cls || gotWS[1] != tok.sep {
		t.Fatalf("Encode(ws) = %v, want [%d, %d]", gotWS, tok.cls, tok.sep)
	}

	// Accented normalization: café should tokenize to same tokens as cafe
	gotCafe := tok.Encode("cafe")
	gotCafeAccent := tok.Encode("café")
	if len(gotCafe) != len(gotCafeAccent) {
		t.Fatalf("Encode(cafe) len=%d != Encode(café) len=%d", len(gotCafe), len(gotCafeAccent))
	}
	for i := range gotCafe {
		if gotCafe[i] != gotCafeAccent[i] {
			t.Fatalf("token %d mismatch: cafe=%d, café=%d", i, gotCafe[i], gotCafeAccent[i])
		}
	}
}

func TestCosine(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	if cos := Cosine(a, b); cos != 0 {
		t.Fatalf("Cosine(orthogonal) = %f, want 0", cos)
	}

	c := []float32{0.6, 0.8}
	if cos := Cosine(c, c); math.Abs(float64(cos-1.0)) > 1e-5 {
		t.Fatalf("Cosine(identical) = %f, want 1.0", cos)
	}

	d := []float32{-0.6, -0.8}
	if cos := Cosine(c, d); math.Abs(float64(cos+1.0)) > 1e-5 {
		t.Fatalf("Cosine(opposite) = %f, want -1.0", cos)
	}

	// Edge cases
	if cos := Cosine(nil, nil); cos != 0 {
		t.Fatalf("Cosine(nil, nil) = %f, want 0", cos)
	}
	if cos := Cosine([]float32{1, 2}, []float32{1}); cos != 0 {
		t.Fatalf("Cosine(mismatched) = %f, want 0", cos)
	}
	if cos := Cosine([]float32{0, 0, 0}, []float32{0, 0, 0}); cos != 0 {
		t.Fatalf("Cosine(zeros) = %f, want 0", cos)
	}
}

func TestEmbedWithBuiltHelper(t *testing.T) {
	cl := nativehelper.Require(t, "feov-embed")
	emb := NewWithClient(cl)

	texts := []string{
		"quantum computing with qubits and superposition",
		"classical computer processors and silicon chips",
		"recipe for delicious baked chocolate cake and frosting",
	}

	vecs, err := emb.Embed(texts)
	if err != nil {
		t.Fatalf("Embed() failed: %v", err)
	}
	if len(vecs) != 3 {
		t.Fatalf("len(vecs) = %d, want 3", len(vecs))
	}
	for i, v := range vecs {
		if len(v) != Dim {
			t.Fatalf("vec %d dim = %d, want %d", i, len(v), Dim)
		}
		var normSq float32
		for _, x := range v {
			normSq += x * x
		}
		if math.Abs(float64(normSq-1.0)) > 1e-3 {
			t.Fatalf("vec %d norm^2 = %f, want ~1.0", i, normSq)
		}
	}

	// Quantum computing vs classical computing vs chocolate cake:
	cos01 := Cosine(vecs[0], vecs[1])
	cos02 := Cosine(vecs[0], vecs[2])
	if cos01 <= cos02 {
		t.Fatalf("expected quantum and classical computing to be closer than quantum and cake: cos(0,1)=%f, cos(0,2)=%f", cos01, cos02)
	}
}

// TestTokenizerParity400: the Go tokenizer gives Hugging Face thenlper/gte-small's ids, exactly, on
// every one of the 400 reference texts (300 of speed.jsonl, 100 of len200.jsonl), truncation at
// MaxIDs included.
func TestTokenizerParity400(t *testing.T) {
	tok := DefaultTokenizer()
	hfSpeed, hfLen200 := hfTokens(t)
	sets := []struct {
		name  string
		texts []string
		want  [][]int64
	}{
		{"speed.jsonl", speedTexts(t), hfSpeed},
		{"len200.jsonl", len200Texts(t), hfLen200},
	}
	same, total, truncated := 0, 0, 0
	for _, set := range sets {
		for i, text := range set.texts {
			total++
			got, want := tok.Encode(text), set.want[i]
			if len(want) == MaxIDs {
				truncated++
			}
			if !slices.Equal(got, want) {
				at := 0
				for at < len(got) && at < len(want) && got[at] == want[at] {
					at++
				}
				t.Errorf("%s text %d: %d ids, Hugging Face %d; first difference at id %d", set.name, i, len(got), len(want), at)
				continue
			}
			same++
		}
	}
	t.Logf("tokenizer parity: %d/%d texts give Hugging Face's ids exactly (%d truncated at %d)", same, total, truncated, MaxIDs)
	if total != 400 {
		t.Fatalf("compared %d texts, want 400", total)
	}
}

// TestModelIDIncludesTargetOSArch: the id a vector cache is keyed by names this target, so a cache
// written by another os or arch is another model's.
func TestModelIDIncludesTargetOSArch(t *testing.T) {
	t.Setenv(nativehelper.EnvDir, t.TempDir())
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	want := "+target@" + runtime.GOOS + "-" + runtime.GOARCH
	if id := e.ModelID(); !strings.HasSuffix(id, want) || !strings.Contains(id, nativehelper.HelperDigest(HelperName)) {
		t.Errorf("ModelID() = %q; want the helper digest and the suffix %q", id, want)
	}
	d := nativehelper.HelperDigest(HelperName)
	for _, other := range [][2]string{{"linux", "arm64"}, {"darwin", "arm64"}, {"windows", "amd64"}, {"linux", "amd64"}} {
		if other[0] == runtime.GOOS && other[1] == runtime.GOARCH {
			continue
		}
		if ModelIDFor(d, other[0], other[1]) == e.ModelID() {
			t.Errorf("%s-%s's model id equals this target's: %q", other[0], other[1], e.ModelID())
		}
	}
}

// TestMissingOrMismatchedHelperRefusesByName: no feov-embed, one built from other sources, and one
// that loaded another gte-small export each refuse, naming the helper and what disagreed.
func TestMissingOrMismatchedHelperRefusesByName(t *testing.T) {
	pins := nativehelper.HelperAssets(HelperName)
	if len(pins) != 1 {
		t.Fatalf("feov-embed carries %d asset pins for %s, want the one gte-small export", len(pins), runtime.GOARCH)
	}
	cases := []struct {
		name string
		spec *nativehelpertest.Spec
		kind error
		says []string
	}{
		{"absent", nil, nativehelper.ErrAbsent, []string{"helper-absent: feov-embed"}},
		{"other sources", &nativehelpertest.Spec{Digest: "0000", Assets: pins}, nativehelper.ErrDigestMismatch,
			[]string{"helper-digest-mismatch: feov-embed", "source digest 0000"}},
		{"other export", &nativehelpertest.Spec{Digest: nativehelper.HelperDigest(HelperName), Assets: []string{"1111"}}, nativehelper.ErrDigestMismatch,
			[]string{"helper-digest-mismatch: feov-embed", "gte-small.int8.ort sha256 (" + runtime.GOARCH + ") 1111", pins[0]}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.spec != nil {
				nativehelpertest.Install(t, dir, HelperName, *tc.spec)
			}
			t.Setenv(nativehelper.EnvDir, dir)
			e, err := New()
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			_, err = e.Embed([]string{"a text"})
			if !errors.Is(err, tc.kind) {
				t.Fatalf("Embed: %v; want %v", err, tc.kind)
			}
			for _, s := range tc.says {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("the refusal %q does not say %q", err, s)
				}
			}
		})
	}
}

// TestEmbedSplitsAtTheBatch: 1,000 texts, some far over MaxIDs, go to the helper as ceil(1000/16)
// = 63 requests, none over BatchSize texts or MaxIDs ids in a text, and every text comes back once.
// The fake helper logs each request it receives.
func TestEmbedSplitsAtTheBatch(t *testing.T) {
	dir := t.TempDir()
	nativehelpertest.Install(t, dir, HelperName, nativehelpertest.Spec{
		Digest: nativehelper.HelperDigest(HelperName),
		Assets: nativehelper.HelperAssets(HelperName),
	})
	t.Setenv(nativehelper.EnvDir, dir)
	log := filepath.Join(t.TempDir(), "embed.log")
	t.Setenv("FAKEHELPER_EMBED_LOG", log)

	long := strings.Repeat("the measured interval between transits widens ", 150)
	in := make([]string, 1000)
	over := 0
	for i := range in {
		in[i] = fmt.Sprintf("text %d about orbital resonance", i)
		if i%37 == 0 {
			in[i] = long + in[i]
			over++
		}
	}
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	vecs, err := e.Embed(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != len(in) {
		t.Fatalf("Embed returned %d vectors for %d texts", len(vecs), len(in))
	}

	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("the fake helper logged no embed request: %v", err)
	}
	reqs := strings.Split(strings.TrimSpace(string(b)), "\n")
	if want := (len(in) + BatchSize - 1) / BatchSize; len(reqs) != want {
		t.Errorf("%d requests, want %d", len(reqs), want)
	}
	sent, capped := 0, 0
	for i, r := range reqs {
		counts := strings.Fields(r)
		if len(counts) > BatchSize {
			t.Errorf("request %d carries %d texts, over %d", i, len(counts), BatchSize)
		}
		for _, c := range counts {
			n, err := strconv.Atoi(c)
			if err != nil {
				t.Fatalf("request %d logged %q", i, r)
			}
			if n > MaxIDs {
				t.Errorf("request %d carries a text of %d ids, over %d", i, n, MaxIDs)
			}
			if n == MaxIDs {
				capped++
			}
		}
		sent += len(counts)
	}
	if sent != len(in) {
		t.Errorf("the requests carried %d texts, want %d", sent, len(in))
	}
	if capped != over {
		t.Errorf("%d texts went at the %d-id cap, want the %d long ones", capped, MaxIDs, over)
	}
}
