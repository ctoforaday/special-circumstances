package vecembed

import (
	"math"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
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

func TestMissingHelperRefusesByName(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("NATIVE_HELPER_DIR", tmp)
	embedder, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer embedder.Close()

	_, err = embedder.Embed([]string{"test text"})
	if err == nil {
		t.Fatal("Embed() on absent helper succeeded, want error")
	}
	if !nativehelper.IsAbsent(err) {
		t.Fatalf("err is not absent: %v", err)
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

func TestEmbedSplitsAtTheBatch(t *testing.T) {
	cl := nativehelper.Require(t, "feov-embed")
	emb := NewWithClient(cl)

	// Create 35 texts (16 + 16 + 3)
	texts := make([]string, 35)
	for i := range texts {
		texts[i] = "sample technical text chunk for batch testing"
	}
	vecs, err := emb.Embed(texts)
	if err != nil {
		t.Fatalf("Embed() failed: %v", err)
	}
	if len(vecs) != 35 {
		t.Fatalf("len(vecs) = %d, want 35", len(vecs))
	}
}
