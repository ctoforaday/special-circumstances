package gte

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// LoadMmap loads a .gtemodel using memory-mapped I/O.
// Weights are accessed directly from the mapped file — no copy into Go heap.
// The returned Model holds a reference to the mapping; call Close() to unmap.
func LoadMmap(modelPath string) (*Model, error) {
	f, err := os.Open(modelPath)
	if err != nil {
		return nil, fmt.Errorf("open model: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat model: %w", err)
	}
	size := int(fi.Size())

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}

	m := &Model{}
	m.mmapData = data

	off := 0

	// Magic
	if size < 4 || string(data[0:4]) != fileMagic {
		syscall.Munmap(data)
		return nil, errors.New("invalid model magic")
	}
	off = 4

	// Config: 6 x uint32
	if size < off+24 {
		syscall.Munmap(data)
		return nil, errors.New("truncated config")
	}
	cfg := make([]int, 6)
	for i := 0; i < 6; i++ {
		cfg[i] = int(binary.LittleEndian.Uint32(data[off:]))
		off += 4
	}
	m.VocabSize = cfg[0]
	m.HiddenSize = cfg[1]
	m.NumLayers = cfg[2]
	m.NumHeads = cfg[3]
	m.Intermediate = cfg[4]
	m.MaxSeqLen = cfg[5]
	if m.HiddenSize%m.NumHeads != 0 {
		syscall.Munmap(data)
		return nil, fmt.Errorf("hidden_size %% num_heads != 0: %d %% %d", m.HiddenSize, m.NumHeads)
	}
	m.HeadDim = m.HiddenSize / m.NumHeads

	// Vocab: length-prefixed strings
	m.Vocab = make([]string, m.VocabSize)
	m.vocabMap = make(map[string]int, m.VocabSize)
	for i := 0; i < m.VocabSize; i++ {
		if off+2 > size {
			syscall.Munmap(data)
			return nil, fmt.Errorf("truncated vocab at %d", i)
		}
		length := int(binary.LittleEndian.Uint16(data[off:]))
		off += 2
		if off+length > size {
			syscall.Munmap(data)
			return nil, fmt.Errorf("truncated vocab string at %d", i)
		}
		m.Vocab[i] = string(data[off : off+length])
		m.vocabMap[m.Vocab[i]] = i
		off += length
	}

	// Helper to slice float32 from mmap'd data
	sliceF32 := func(count int) ([]float32, error) {
		nbytes := count * 4
		if off+nbytes > size {
			return nil, fmt.Errorf("truncated at offset %d, need %d bytes", off, nbytes)
		}
		ptr := unsafe.Pointer(&data[off])
		s := unsafe.Slice((*float32)(ptr), count)
		off += nbytes
		return s, nil
	}

	// Embeddings
	if m.TokenEmbeddings, err = sliceF32(m.VocabSize * m.HiddenSize); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("token embeddings: %w", err)
	}
	if m.PositionEmb, err = sliceF32(m.MaxSeqLen * m.HiddenSize); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("position embeddings: %w", err)
	}
	if m.TokenTypeEmb, err = sliceF32(2 * m.HiddenSize); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("token type embeddings: %w", err)
	}
	if m.EmbedLnWeight, err = sliceF32(m.HiddenSize); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("embed ln weight: %w", err)
	}
	if m.EmbedLnBias, err = sliceF32(m.HiddenSize); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("embed ln bias: %w", err)
	}

	// Layers
	m.Layers = make([]LayerWeights, m.NumLayers)
	h := m.HiddenSize
	inter := m.Intermediate
	for l := 0; l < m.NumLayers; l++ {
		lw := &m.Layers[l]
		weights := []struct {
			dst  *[]float32
			size int
		}{
			{&lw.QueryWeight, h * h}, {&lw.QueryBias, h},
			{&lw.KeyWeight, h * h}, {&lw.KeyBias, h},
			{&lw.ValueWeight, h * h}, {&lw.ValueBias, h},
			{&lw.AttnOutputWeight, h * h}, {&lw.AttnOutputBias, h},
			{&lw.AttnLnWeight, h}, {&lw.AttnLnBias, h},
			{&lw.FfnInterWeight, inter * h}, {&lw.FfnInterBias, inter},
			{&lw.FfnOutputWeight, h * inter}, {&lw.FfnOutputBias, h},
			{&lw.FfnLnWeight, h}, {&lw.FfnLnBias, h},
		}
		for _, w := range weights {
			if *w.dst, err = sliceF32(w.size); err != nil {
				syscall.Munmap(data)
				return nil, fmt.Errorf("layer %d: %w", l, err)
			}
		}
	}

	// Pooler
	if m.PoolerWeight, err = sliceF32(h * h); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("pooler weight: %w", err)
	}
	if m.PoolerBias, err = sliceF32(h); err != nil {
		syscall.Munmap(data)
		return nil, fmt.Errorf("pooler bias: %w", err)
	}

	if err := m.initBuffers(); err != nil {
		syscall.Munmap(data)
		return nil, err
	}
	m.fuseQKV()

	return m, nil
}
