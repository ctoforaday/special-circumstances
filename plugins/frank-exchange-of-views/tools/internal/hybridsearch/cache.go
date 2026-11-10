package hybridsearch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"maps"
	"math"
	"os"
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

// A VECTOR IS ONLY AS GOOD AS THE MODEL, HELPER, TARGET AND CHUNKER THAT MADE IT, so a cache file
// states all four in its header and is read back only by a reader that wants all four. The model,
// helper and target are vecembed.ModelID; the chunker is a chunk spec (SourceChunkSpec,
// ReportChunkSpec). A file that disagrees on any of them is a miss, never a read: byte offsets from
// another chunker, or vectors from another build, would rank silently wrong.
//
// The owner of the run's cache directory (fetchcache) writes the files, through its replacing
// write, from the bytes Encode* returns here.

// ErrCacheMiss is a cache file that is absent, corrupt, or written by another model, helper, target
// or chunker.
var ErrCacheMiss = errors.New("vector cache miss")

// The magics name the two layouts. Each is followed by the model id and the chunk spec.
var (
	sourceMagic = [8]byte{'F', 'E', 'O', 'V', 'V', 'E', 'C', '2'}
	reportMagic = [8]byte{'F', 'E', 'O', 'V', 'R', 'P', 'T', '1'}
)

// CachedChunk is one source chunk's coordinates and dense vector.
type CachedChunk struct {
	Index     int
	StartByte int
	EndByte   int
	Page      int
	Vector    []float32
}

// ParaSha is the sha256 of a report paragraph's visible text: what its vector is cached under.
type ParaSha [32]byte

// EncodeVecCache is a source's chunk vectors as its .vec file holds them.
func EncodeVecCache(modelID, chunkSpec string, chunks []CachedChunk) []byte {
	var b bytes.Buffer
	writeHeader(&b, sourceMagic, modelID, chunkSpec, len(chunks))
	for _, c := range chunks {
		for _, v := range []int{c.StartByte, c.EndByte, c.Page} {
			_ = binary.Write(&b, binary.LittleEndian, uint32(v))
		}
		writeVector(&b, c.Vector)
	}
	return b.Bytes()
}

// ReadVecCache reads a source's .vec file, refusing — as ErrCacheMiss — a file whose magic, model id
// or chunk spec is not the one wanted.
func ReadVecCache(path, wantModelID, wantChunkSpec string) ([]CachedChunk, error) {
	r, count, err := openCache(path, sourceMagic, wantModelID, wantChunkSpec, 3*4+vecembed.Dim*4)
	if err != nil {
		return nil, err
	}
	chunks := make([]CachedChunk, count)
	for i := range chunks {
		var coords [3]uint32
		if err := binary.Read(r, binary.LittleEndian, &coords); err != nil {
			return nil, ErrCacheMiss
		}
		vec, err := readVector(r)
		if err != nil {
			return nil, err
		}
		chunks[i] = CachedChunk{Index: i, StartByte: int(coords[0]), EndByte: int(coords[1]), Page: int(coords[2]), Vector: vec}
	}
	if r.Len() != 0 {
		return nil, ErrCacheMiss
	}
	return chunks, nil
}

// EncodeReportVecCache is the report's paragraph vectors as report_vecs.bin holds them.
func EncodeReportVecCache(modelID, chunkSpec string, vecs map[ParaSha][]float32) []byte {
	var b bytes.Buffer
	writeHeader(&b, reportMagic, modelID, chunkSpec, len(vecs))
	// In hash order, so the same vectors are the same bytes.
	for _, sha := range slices.SortedFunc(maps.Keys(vecs), func(x, y ParaSha) int { return bytes.Compare(x[:], y[:]) }) {
		b.Write(sha[:])
		writeVector(&b, vecs[sha])
	}
	return b.Bytes()
}

// ReadReportVecCache reads report_vecs.bin, refusing — as ErrCacheMiss — a file whose magic, model
// id or chunk spec is not the one wanted.
func ReadReportVecCache(path, wantModelID, wantChunkSpec string) (map[ParaSha][]float32, error) {
	r, count, err := openCache(path, reportMagic, wantModelID, wantChunkSpec, len(ParaSha{})+vecembed.Dim*4)
	if err != nil {
		return nil, err
	}
	out := make(map[ParaSha][]float32, count)
	for range count {
		var sha ParaSha
		if _, err := io.ReadFull(r, sha[:]); err != nil {
			return nil, ErrCacheMiss
		}
		vec, err := readVector(r)
		if err != nil {
			return nil, err
		}
		out[sha] = vec
	}
	if r.Len() != 0 {
		return nil, ErrCacheMiss
	}
	return out, nil
}

func writeHeader(b *bytes.Buffer, magic [8]byte, modelID, chunkSpec string, count int) {
	b.Write(magic[:])
	for _, s := range []string{modelID, chunkSpec} {
		_ = binary.Write(b, binary.LittleEndian, uint16(len(s)))
		b.WriteString(s)
	}
	_ = binary.Write(b, binary.LittleEndian, uint32(count))
}

func writeVector(b *bytes.Buffer, v []float32) {
	var word [4]byte
	for d := range vecembed.Dim {
		var x float32
		if d < len(v) {
			x = v[d]
		}
		binary.LittleEndian.PutUint32(word[:], math.Float32bits(x))
		b.Write(word[:])
	}
}

// openCache reads the file at path and checks its header, returning the rest and the entry count.
// An absent file, another layout, another model id or chunk spec, or a count the rest cannot hold at
// entrySize bytes an entry is ErrCacheMiss; a file that exists and cannot be read is that error.
func openCache(path string, magic [8]byte, wantModelID, wantChunkSpec string, entrySize int) (*bytes.Reader, int, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, ErrCacheMiss
	}
	if err != nil {
		return nil, 0, err
	}
	r := bytes.NewReader(raw)
	var got [8]byte
	if _, err := io.ReadFull(r, got[:]); err != nil || got != magic {
		return nil, 0, ErrCacheMiss
	}
	for _, want := range []string{wantModelID, wantChunkSpec} {
		var n uint16
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return nil, 0, ErrCacheMiss
		}
		s := make([]byte, n)
		if _, err := io.ReadFull(r, s); err != nil || string(s) != want {
			return nil, 0, ErrCacheMiss
		}
	}
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil || int64(count)*int64(entrySize) != int64(r.Len()) {
		return nil, 0, ErrCacheMiss
	}
	return r, int(count), nil
}

func readVector(r *bytes.Reader) ([]float32, error) {
	var raw [vecembed.Dim * 4]byte
	if _, err := io.ReadFull(r, raw[:]); err != nil {
		return nil, ErrCacheMiss
	}
	vec := make([]float32, vecembed.Dim)
	for d := range vec {
		vec[d] = math.Float32frombits(binary.LittleEndian.Uint32(raw[d*4:]))
	}
	return vec, nil
}
