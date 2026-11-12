package hybridsearch

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

func vec(at int, x float32) []float32 {
	v := make([]float32, vecembed.Dim)
	v[at] = x
	return v
}

// A CACHE FILE IS READ ONLY BY THE MODEL, HELPER, TARGET AND CHUNKER THAT WROTE IT. The model id
// carries the helper build and the target, so a cache carried from linux-amd64 to linux-arm64, or
// written before the chunker changed, is a miss — never vectors from another build or offsets from
// another cut.
func TestVecCacheRejectsChangedModelTargetOrChunkSpec(t *testing.T) {
	dir := t.TempDir()
	amd := vecembed.ModelIDFor("d1", "linux", "amd64")
	chunks := []CachedChunk{{StartByte: 10, EndByte: 100, Page: 1, Vector: vec(0, 0.5)}, {StartByte: 90, EndByte: 200, Page: 2, Vector: vec(383, 0.99)}}
	src := filepath.Join(dir, "t.vec")
	if err := os.WriteFile(src, EncodeVecCache(amd, SourceChunkSpec, chunks), 0o644); err != nil {
		t.Fatal(err)
	}
	paras := map[ParaSha][]float32{sha256.Sum256([]byte("one")): vec(1, 1), sha256.Sum256([]byte("two")): vec(2, 1)}
	rep := filepath.Join(dir, "report_vecs.bin")
	if err := os.WriteFile(rep, EncodeReportVecCache(amd, ReportChunkSpec, paras), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadVecCache(src, amd, SourceChunkSpec)
	if err != nil || len(got) != 2 || got[0].StartByte != 10 || got[0].Vector[0] != 0.5 || got[1].EndByte != 200 || got[1].Page != 2 || got[1].Vector[383] != 0.99 {
		t.Fatalf("ReadVecCache by its writer = %+v, %v; want the chunks written", got, err)
	}
	gotParas, err := ReadReportVecCache(rep, amd, ReportChunkSpec)
	if err != nil || len(gotParas) != 2 || gotParas[sha256.Sum256([]byte("two"))][2] != 1 {
		t.Fatalf("ReadReportVecCache by its writer = %d entries, %v; want both paragraphs", len(gotParas), err)
	}

	for _, tc := range []struct{ name, model, spec, reportSpec string }{
		{"another target", vecembed.ModelIDFor("d1", "linux", "arm64"), SourceChunkSpec, ReportChunkSpec},
		{"another helper build", vecembed.ModelIDFor("d2", "linux", "amd64"), SourceChunkSpec, ReportChunkSpec},
		{"another chunker", amd, "w256_s200", "para_v2"},
	} {
		if _, err := ReadVecCache(src, tc.model, tc.spec); !errors.Is(err, ErrCacheMiss) {
			t.Errorf("%s: ReadVecCache = %v, want ErrCacheMiss", tc.name, err)
		}
		if _, err := ReadReportVecCache(rep, tc.model, tc.reportSpec); !errors.Is(err, ErrCacheMiss) {
			t.Errorf("%s: ReadReportVecCache = %v, want ErrCacheMiss", tc.name, err)
		}
	}
	// One layout is never read as the other, and a cut or absent file is a miss.
	if _, err := ReadVecCache(rep, amd, ReportChunkSpec); !errors.Is(err, ErrCacheMiss) {
		t.Errorf("a report cache read as a source cache = %v, want ErrCacheMiss", err)
	}
	raw, _ := os.ReadFile(src)
	cut := filepath.Join(dir, "cut.vec")
	if err := os.WriteFile(cut, raw[:len(raw)-4], 0o644); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{cut, filepath.Join(dir, "absent.vec")} {
		if _, err := ReadVecCache(p, amd, SourceChunkSpec); !errors.Is(err, ErrCacheMiss) {
			t.Errorf("ReadVecCache(%s) = %v, want ErrCacheMiss", filepath.Base(p), err)
		}
	}
}
