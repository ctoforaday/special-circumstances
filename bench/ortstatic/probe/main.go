// ortprobe: embed texts with gte-small through ONNX Runtime's C API (or the no-ORT stub, -tags noort),
// mean-pool over the attention mask, L2-normalize.
//
//	ortprobe <model> <input.jsonl> <output.bin> <threads>     (vocab: $VOCAB or vocab.txt beside the model)
//	ortprobe cos <output.bin> <ref.npy>                       cosine mean/p05/min vs a float32 (n,384) npy
//	ortprobe tok <input.jsonl>                                token ids, one JSON array per line
package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dim = 384

func readTexts(p string) []string {
	f, err := os.Open(p)
	must(err)
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1<<24), 1<<24)
	var ts []string
	for s.Scan() {
		var r struct {
			Text string `json:"text"`
		}
		must(json.Unmarshal(s.Bytes(), &r))
		ts = append(ts, r.Text)
	}
	return ts
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ortprobe:", err)
		os.Exit(1)
	}
}

func vocabFor(model string) string {
	if v := os.Getenv("VOCAB"); v != "" {
		return v
	}
	return filepath.Join(filepath.Dir(model), "vocab.txt")
}

func main() {
	if len(os.Args) == 4 && os.Args[1] == "cos" {
		cos(os.Args[2], os.Args[3])
		return
	}
	if len(os.Args) == 4 && os.Args[1] == "tok" {
		tk, err := LoadTokenizer(os.Args[3], 512)
		must(err)
		w := bufio.NewWriter(os.Stdout)
		for _, t := range readTexts(os.Args[2]) {
			b, _ := json.Marshal(tk.Encode(t))
			w.Write(b)
			w.WriteByte('\n')
		}
		w.Flush()
		return
	}
	if len(os.Args) != 5 {
		must(fmt.Errorf("usage: ortprobe <model> <input.jsonl> <output.bin> <threads> | cos <bin> <npy> | tok <jsonl> <vocab>"))
	}
	threads, err := strconv.Atoi(os.Args[4])
	must(err)
	tk, err := LoadTokenizer(vocabFor(os.Args[1]), 512)
	must(err)
	t0 := time.Now()
	eng, err := newEngine(os.Args[1], threads)
	must(err)
	load := time.Since(t0)
	ts := readTexts(os.Args[2])
	embed := func(t string, out []float32) []float32 {
		ids := tk.Encode(t)
		mask := make([]int64, len(ids))
		for i := range mask {
			mask[i] = 1
		}
		h, err := eng.run(ids, mask)
		must(err)
		v := make([]float64, dim)
		for i := range ids {
			for d := 0; d < dim; d++ {
				v[d] += float64(h[i*dim+d])
			}
		}
		var n float64
		for d := range v {
			v[d] /= float64(len(ids))
			n += v[d] * v[d]
		}
		n = math.Sqrt(n)
		for d := range v {
			if n > 0 {
				out = append(out, float32(v[d]/n))
			} else {
				out = append(out, 0)
			}
		}
		return out
	}
	// Warm-up on the first three texts, untimed, as bench/embed/ort/ort.py does.
	for i := 0; i < 3 && i < len(ts); i++ {
		embed(ts[i], nil)
	}
	out := make([]float32, 0, len(ts)*dim)
	t1 := time.Now()
	for _, t := range ts {
		out = embed(t, out)
	}
	d := time.Since(t1)
	f, err := os.Create(os.Args[3])
	must(err)
	bw := bufio.NewWriter(f)
	must(binary.Write(bw, binary.LittleEndian, out))
	must(bw.Flush())
	must(f.Close())
	fmt.Printf("load=%.2fs n=%d threads=%d secs=%.2f per_s=%.2f ms_each=%.1f dim=%d engine=%s\n",
		load.Seconds(), len(ts), threads, d.Seconds(), float64(len(ts))/d.Seconds(), d.Seconds()*1000/float64(len(ts)), dim, engineName)
}

func cos(binPath, npyPath string) {
	b, err := os.ReadFile(binPath)
	must(err)
	got := make([]float32, len(b)/4)
	must(binary.Read(strings.NewReader(string(b)), binary.LittleEndian, got))
	raw, err := os.ReadFile(npyPath)
	must(err)
	if len(raw) < 10 || string(raw[1:6]) != "NUMPY" {
		must(fmt.Errorf("%s: not an npy file", npyPath))
	}
	var hl, off int
	if raw[6] == 1 {
		hl, off = int(binary.LittleEndian.Uint16(raw[8:10])), 10
	} else {
		hl, off = int(binary.LittleEndian.Uint32(raw[8:12])), 12
	}
	hdr := string(raw[off : off+hl])
	if !strings.Contains(hdr, "'<f4'") || strings.Contains(hdr, "'fortran_order': True") {
		must(fmt.Errorf("%s: header %q is not little-endian float32 C-order", npyPath, hdr))
	}
	body := raw[off+hl:]
	ref := make([]float32, len(body)/4)
	must(binary.Read(strings.NewReader(string(body)), binary.LittleEndian, ref))
	if len(ref) != len(got) {
		must(fmt.Errorf("size mismatch: got %d floats, ref %d", len(got), len(ref)))
	}
	n := len(ref) / dim
	cs := make([]float64, n)
	var sum float64
	for i := 0; i < n; i++ {
		var dot, na, nb float64
		for d := 0; d < dim; d++ {
			a, r := float64(got[i*dim+d]), float64(ref[i*dim+d])
			dot += a * r
			na += a * a
			nb += r * r
		}
		cs[i] = dot / math.Sqrt(na*nb)
		sum += cs[i]
	}
	sort.Float64s(cs)
	fmt.Printf("fidelity n=%d mean=%.6f p05=%.6f min=%.6f\n", n, sum/float64(n), cs[n*5/100], cs[0])
}
