package hybridsearch

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

var (
	// ErrCacheMiss indicates the cache file was absent, corrupt, or addressed to another model.
	ErrCacheMiss = errors.New("vector cache miss")
)

var cacheMagic = [4]byte{'F', 'V', 'E', 'C'}

const (
	cacheVersion uint16 = 1
	cacheDim     uint16 = 384
)

// CachedChunk is one chunk's coordinates and dense vector embedding.
type CachedChunk struct {
	Index     int
	StartByte int
	EndByte   int
	Page      int
	Vector    []float32
}

// ReadCache reads a .vec binary cache file, returning ErrCacheMiss if invalid or for another model.
func ReadCache(path string, wantModelID string) ([]CachedChunk, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	defer f.Close()

	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil || magic != cacheMagic {
		return nil, ErrCacheMiss
	}

	var version, dim uint16
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil || version != cacheVersion {
		return nil, ErrCacheMiss
	}
	if err := binary.Read(f, binary.LittleEndian, &dim); err != nil || dim != cacheDim {
		return nil, ErrCacheMiss
	}

	var count uint32
	if err := binary.Read(f, binary.LittleEndian, &count); err != nil {
		return nil, ErrCacheMiss
	}

	var modelIDLen uint16
	if err := binary.Read(f, binary.LittleEndian, &modelIDLen); err != nil {
		return nil, ErrCacheMiss
	}
	modelIDBytes := make([]byte, modelIDLen)
	if _, err := io.ReadFull(f, modelIDBytes); err != nil {
		return nil, ErrCacheMiss
	}
	if string(modelIDBytes) != wantModelID {
		return nil, ErrCacheMiss
	}

	chunks := make([]CachedChunk, count)
	vecBytes := make([]byte, int(cacheDim)*4)
	for i := 0; i < int(count); i++ {
		var start, end, page uint32
		if err := binary.Read(f, binary.LittleEndian, &start); err != nil {
			return nil, ErrCacheMiss
		}
		if err := binary.Read(f, binary.LittleEndian, &end); err != nil {
			return nil, ErrCacheMiss
		}
		if err := binary.Read(f, binary.LittleEndian, &page); err != nil {
			return nil, ErrCacheMiss
		}
		if _, err := io.ReadFull(f, vecBytes); err != nil {
			return nil, ErrCacheMiss
		}
		vec := make([]float32, cacheDim)
		for d := 0; d < int(cacheDim); d++ {
			u := binary.LittleEndian.Uint32(vecBytes[d*4 : (d+1)*4])
			vec[d] = math.Float32frombits(u)
		}
		chunks[i] = CachedChunk{
			Index:     i,
			StartByte: int(start),
			EndByte:   int(end),
			Page:      int(page),
			Vector:    vec,
		}
	}
	return chunks, nil
}

// WriteCache atomically writes chunks and their embeddings to a .vec cache file.
func WriteCache(path string, modelID string, chunks []CachedChunk) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	write := func() error {
		if _, err := f.Write(cacheMagic[:]); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, cacheVersion); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, cacheDim); err != nil {
			return err
		}
		count := uint32(len(chunks))
		if err := binary.Write(f, binary.LittleEndian, count); err != nil {
			return err
		}
		modelIDLen := uint16(len(modelID))
		if err := binary.Write(f, binary.LittleEndian, modelIDLen); err != nil {
			return err
		}
		if _, err := f.WriteString(modelID); err != nil {
			return err
		}
		vecBytes := make([]byte, int(cacheDim)*4)
		for _, c := range chunks {
			if err := binary.Write(f, binary.LittleEndian, uint32(c.StartByte)); err != nil {
				return err
			}
			if err := binary.Write(f, binary.LittleEndian, uint32(c.EndByte)); err != nil {
				return err
			}
			if err := binary.Write(f, binary.LittleEndian, uint32(c.Page)); err != nil {
				return err
			}
			for d := 0; d < int(cacheDim); d++ {
				binary.LittleEndian.PutUint32(vecBytes[d*4:(d+1)*4], math.Float32bits(c.Vector[d]))
			}
			if _, err := f.Write(vecBytes); err != nil {
				return err
			}
		}
		return f.Sync()
	}

	if err := write(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
