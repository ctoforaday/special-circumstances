package vecembed

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
)

// The reference data (testdata/README.md says where each file came from). Embedded, so the test
// binary that native-run's servercore container runs carries it.
var (
	//go:embed testdata/speed.jsonl
	speedJSONL []byte
	//go:embed testdata/len200.jsonl
	len200JSONL []byte
	//go:embed testdata/ref_gte.npy
	refGTENpy []byte
	//go:embed testdata/hf_tokens_400.json
	hfTokensJSON []byte
)

// texts reads a JSON-lines file of {"text": …} objects, in line order.
func texts(t testing.TB, name string, b []byte) []string {
	t.Helper()
	var out []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var row struct {
			Text *string `json:"text"`
		}
		if err := json.Unmarshal(line, &row); err != nil || row.Text == nil {
			t.Fatalf("%s line %d is not a {\"text\": …} object: %v", name, len(out)+1, err)
		}
		out = append(out, *row.Text)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return out
}

func speedTexts(t testing.TB) []string {
	t.Helper()
	s := texts(t, "speed.jsonl", speedJSONL)
	if len(s) != 300 {
		t.Fatalf("speed.jsonl holds %d texts, want 300", len(s))
	}
	return s
}

func len200Texts(t testing.TB) []string {
	t.Helper()
	s := texts(t, "len200.jsonl", len200JSONL)
	if len(s) != 100 {
		t.Fatalf("len200.jsonl holds %d texts, want 100", len(s))
	}
	return s
}

// hfTokens is the Hugging Face tokenizer's ids for speed.jsonl and len200.jsonl.
func hfTokens(t testing.TB) (speed, len200 [][]int64) {
	t.Helper()
	var doc struct {
		Tokenizer string    `json:"tokenizer"`
		Speed     [][]int64 `json:"speed"`
		Len200    [][]int64 `json:"len200"`
	}
	if err := json.Unmarshal(hfTokensJSON, &doc); err != nil {
		t.Fatalf("hf_tokens_400.json: %v", err)
	}
	if doc.Tokenizer != "thenlper/gte-small" || len(doc.Speed) != 300 || len(doc.Len200) != 100 {
		t.Fatalf("hf_tokens_400.json holds tokenizer %q with %d speed and %d len200 rows; want thenlper/gte-small, 300 and 100",
			doc.Tokenizer, len(doc.Speed), len(doc.Len200))
	}
	return doc.Speed, doc.Len200
}

// refVectors reads ref_gte.npy: a little-endian float32 (300, Dim) array in C order, each row
// L2-normalised here.
func refVectors(t testing.TB) [][]float32 {
	t.Helper()
	rows, err := readNpyF32(refGTENpy, 300, Dim)
	if err != nil {
		t.Fatalf("ref_gte.npy: %v", err)
	}
	for _, r := range rows {
		normalise(r)
	}
	return rows
}

// readNpyF32 reads a version 1 .npy file holding a '<f4' C-order array of shape (n, d), and refuses
// any other.
func readNpyF32(b []byte, n, d int) ([][]float32, error) {
	if len(b) < 10 || string(b[:6]) != "\x93NUMPY" || b[6] != 1 {
		return nil, fmt.Errorf("not a version 1 .npy file")
	}
	hl := int(binary.LittleEndian.Uint16(b[8:10]))
	if len(b) < 10+hl {
		return nil, fmt.Errorf("header of %d bytes runs past the file", hl)
	}
	header := string(b[10 : 10+hl])
	for _, want := range []string{"'descr': '<f4'", "'fortran_order': False", fmt.Sprintf("'shape': (%d, %d)", n, d)} {
		if !strings.Contains(header, want) {
			return nil, fmt.Errorf("header %q lacks %s", strings.TrimSpace(header), want)
		}
	}
	data := b[10+hl:]
	if len(data) != 4*n*d {
		return nil, fmt.Errorf("data is %d bytes, want %d", len(data), 4*n*d)
	}
	rows := make([][]float32, n)
	for i := range rows {
		rows[i] = make([]float32, d)
		for j := range rows[i] {
			rows[i][j] = math.Float32frombits(binary.LittleEndian.Uint32(data[4*(i*d+j):]))
		}
	}
	return rows, nil
}

func normalise(v []float32) {
	var n float64
	for _, x := range v {
		n += float64(x) * float64(x)
	}
	n = math.Sqrt(n)
	if n == 0 {
		return
	}
	for i := range v {
		v[i] = float32(float64(v[i]) / n)
	}
}
