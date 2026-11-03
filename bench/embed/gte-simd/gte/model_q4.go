package gte

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

const fileMagicQ4 = "GTE4"

// ModelQ4 holds a model with Q4-quantized weight matrices.
// Biases, LayerNorm, and attention scores remain FP32.
type ModelQ4 struct {
	VocabSize    int
	HiddenSize   int
	NumLayers    int
	NumHeads     int
	Intermediate int
	MaxSeqLen    int
	HeadDim      int

	Vocab    []string
	vocabMap map[string]int

	// Embeddings stored as Q4 blocks (looked up by index, dequantized per call)
	TokenEmbQ4   []byte // [VocabSize * blocksPerHidden * BlockQ4Size]
	PositionEmbQ4 []byte // [MaxSeqLen * blocksPerHidden * BlockQ4Size]
	TokenTypeEmbQ4 []byte // [2 * blocksPerHidden * BlockQ4Size]
	EmbedLnWeight []float32
	EmbedLnBias   []float32

	Layers       []LayerWeightsQ4
	PoolerWeightQ4 []byte
	PoolerBias   []float32

	// Working buffers (same as FP32 model)
	hiddenStates []float32
	attnScores   []float32
	qkvProj      []float32
	qHeadBuf     []float32
	kHeadBuf     []float32
	vHeadBuf     []float32
	cHeadBuf     []float32
	attnOutput   []float32
	ffnHidden    []float32
	tempHidden   []float32
	tokenBuf     []int
	attnMaskBuf  []bool
	basicBuf     []string
	wpBuf        []byte
	lowerBuf     []byte
}

// LayerWeightsQ4 holds quantized weights for one transformer block.
type LayerWeightsQ4 struct {
	// Q4 weight matrices (stored as raw blocks)
	QueryWeightQ4      []byte
	KeyWeightQ4        []byte
	ValueWeightQ4      []byte
	AttnOutputWeightQ4 []byte
	FfnInterWeightQ4   []byte
	FfnOutputWeightQ4  []byte

	// Biases remain FP32
	QueryBias      []float32
	KeyBias        []float32
	ValueBias      []float32
	AttnOutputBias []float32
	FfnInterBias   []float32
	FfnOutputBias  []float32

	// LayerNorm params remain FP32
	AttnLnWeight []float32
	AttnLnBias   []float32
	FfnLnWeight  []float32
	FfnLnBias    []float32

	// Fused QKV Q4: [3*hidden rows, blocksPerHidden * BlockQ4Size bytes per row]
	QKVWeightQ4 []byte
	QKVBias     []float32
}

// LoadQ4 reads a Q4-quantized .gtemodel file (magic "GTE4").
func LoadQ4(modelPath string) (*ModelQ4, error) {
	f, err := os.Open(modelPath)
	if err != nil {
		return nil, fmt.Errorf("open model: %w", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if string(magic) != fileMagicQ4 {
		return nil, fmt.Errorf("invalid Q4 model magic: got %q, want %q", string(magic), fileMagicQ4)
	}

	cfg := make([]int, 6)
	for i := 0; i < 6; i++ {
		v, err := readUint32(r)
		if err != nil {
			return nil, fmt.Errorf("read config %d: %w", i, err)
		}
		cfg[i] = v
	}

	m := &ModelQ4{
		VocabSize:    cfg[0],
		HiddenSize:   cfg[1],
		NumLayers:    cfg[2],
		NumHeads:     cfg[3],
		Intermediate: cfg[4],
		MaxSeqLen:    cfg[5],
	}
	if m.HiddenSize%m.NumHeads != 0 {
		return nil, fmt.Errorf("hidden_size %% num_heads != 0: %d %% %d", m.HiddenSize, m.NumHeads)
	}
	if m.HiddenSize%QK4_0 != 0 {
		return nil, fmt.Errorf("hidden_size %d not divisible by block size %d", m.HiddenSize, QK4_0)
	}
	m.HeadDim = m.HiddenSize / m.NumHeads

	if err := m.readVocab(r); err != nil {
		return nil, err
	}
	if err := m.readEmbeddings(r); err != nil {
		return nil, err
	}
	if err := m.readLayers(r); err != nil {
		return nil, err
	}
	if err := m.readPooler(r); err != nil {
		return nil, err
	}
	m.initBuffers()
	m.fuseQKV()

	return m, nil
}

func (m *ModelQ4) readVocab(r *bufio.Reader) error {
	m.Vocab = make([]string, m.VocabSize)
	m.vocabMap = make(map[string]int, m.VocabSize)
	for i := 0; i < m.VocabSize; i++ {
		length, err := readUint16(r)
		if err != nil {
			return fmt.Errorf("read vocab len %d: %w", i, err)
		}
		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return fmt.Errorf("read vocab %d: %w", i, err)
		}
		m.Vocab[i] = string(buf)
		m.vocabMap[m.Vocab[i]] = i
	}
	return nil
}

func (m *ModelQ4) readEmbeddings(r *bufio.Reader) error {
	var err error
	blocksPerHidden := m.HiddenSize / QK4_0

	// Token embeddings: [VocabSize, hidden] quantized
	m.TokenEmbQ4, err = readBytes(r, m.VocabSize*blocksPerHidden*BlockQ4Size)
	if err != nil {
		return fmt.Errorf("token embeddings q4: %w", err)
	}
	// Position embeddings: [MaxSeqLen, hidden] quantized
	m.PositionEmbQ4, err = readBytes(r, m.MaxSeqLen*blocksPerHidden*BlockQ4Size)
	if err != nil {
		return fmt.Errorf("position embeddings q4: %w", err)
	}
	// Token type embeddings: [2, hidden] quantized
	m.TokenTypeEmbQ4, err = readBytes(r, 2*blocksPerHidden*BlockQ4Size)
	if err != nil {
		return fmt.Errorf("token type embeddings q4: %w", err)
	}
	// LayerNorm: FP32
	if m.EmbedLnWeight, err = readFloat32Slice(r, m.HiddenSize); err != nil {
		return fmt.Errorf("embed ln weight: %w", err)
	}
	if m.EmbedLnBias, err = readFloat32Slice(r, m.HiddenSize); err != nil {
		return fmt.Errorf("embed ln bias: %w", err)
	}
	return nil
}

func (m *ModelQ4) readLayers(r *bufio.Reader) error {
	h := m.HiddenSize
	inter := m.Intermediate
	blocksH := h / QK4_0
	blocksInter := inter / QK4_0

	m.Layers = make([]LayerWeightsQ4, m.NumLayers)
	for l := 0; l < m.NumLayers; l++ {
		lw := &m.Layers[l]
		var err error

		// Q/K/V weights: [hidden, hidden] quantized → hidden rows, each blocksH blocks
		if lw.QueryWeightQ4, err = readBytes(r, h*blocksH*BlockQ4Size); err != nil {
			return fmt.Errorf("layer %d query weight: %w", l, err)
		}
		if lw.QueryBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d query bias: %w", l, err)
		}
		if lw.KeyWeightQ4, err = readBytes(r, h*blocksH*BlockQ4Size); err != nil {
			return fmt.Errorf("layer %d key weight: %w", l, err)
		}
		if lw.KeyBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d key bias: %w", l, err)
		}
		if lw.ValueWeightQ4, err = readBytes(r, h*blocksH*BlockQ4Size); err != nil {
			return fmt.Errorf("layer %d value weight: %w", l, err)
		}
		if lw.ValueBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d value bias: %w", l, err)
		}
		if lw.AttnOutputWeightQ4, err = readBytes(r, h*blocksH*BlockQ4Size); err != nil {
			return fmt.Errorf("layer %d attn output weight: %w", l, err)
		}
		if lw.AttnOutputBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d attn output bias: %w", l, err)
		}
		if lw.AttnLnWeight, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d attn ln weight: %w", l, err)
		}
		if lw.AttnLnBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d attn ln bias: %w", l, err)
		}

		// FFN inter: [intermediate, hidden] quantized
		if lw.FfnInterWeightQ4, err = readBytes(r, inter*blocksH*BlockQ4Size); err != nil {
			return fmt.Errorf("layer %d ffn inter weight: %w", l, err)
		}
		if lw.FfnInterBias, err = readFloat32Slice(r, inter); err != nil {
			return fmt.Errorf("layer %d ffn inter bias: %w", l, err)
		}
		// FFN output: [hidden, intermediate] quantized
		if lw.FfnOutputWeightQ4, err = readBytes(r, h*blocksInter*BlockQ4Size); err != nil {
			return fmt.Errorf("layer %d ffn output weight: %w", l, err)
		}
		if lw.FfnOutputBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d ffn output bias: %w", l, err)
		}
		if lw.FfnLnWeight, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d ffn ln weight: %w", l, err)
		}
		if lw.FfnLnBias, err = readFloat32Slice(r, h); err != nil {
			return fmt.Errorf("layer %d ffn ln bias: %w", l, err)
		}
	}
	return nil
}

func (m *ModelQ4) readPooler(r *bufio.Reader) error {
	var err error
	blocksH := m.HiddenSize / QK4_0
	if m.PoolerWeightQ4, err = readBytes(r, m.HiddenSize*blocksH*BlockQ4Size); err != nil {
		return fmt.Errorf("pooler weight q4: %w", err)
	}
	if m.PoolerBias, err = readFloat32Slice(r, m.HiddenSize); err != nil {
		return fmt.Errorf("pooler bias: %w", err)
	}
	return nil
}

func (m *ModelQ4) fuseQKV() {
	h := m.HiddenSize
	blocksH := h / QK4_0
	rowBytes := blocksH * BlockQ4Size

	for l := range m.Layers {
		lw := &m.Layers[l]
		// Fused QKV weight: [3*h rows, rowBytes each]
		lw.QKVWeightQ4 = make([]byte, 3*h*rowBytes)
		copy(lw.QKVWeightQ4[0:h*rowBytes], lw.QueryWeightQ4)
		copy(lw.QKVWeightQ4[h*rowBytes:2*h*rowBytes], lw.KeyWeightQ4)
		copy(lw.QKVWeightQ4[2*h*rowBytes:3*h*rowBytes], lw.ValueWeightQ4)
		// Fused bias
		lw.QKVBias = make([]float32, 3*h)
		copy(lw.QKVBias[0:h], lw.QueryBias)
		copy(lw.QKVBias[h:2*h], lw.KeyBias)
		copy(lw.QKVBias[2*h:3*h], lw.ValueBias)
	}
}

func (m *ModelQ4) initBuffers() {
	maxSeq := m.MaxSeqLen
	hidden := m.HiddenSize
	inter := m.Intermediate
	heads := m.NumHeads

	m.hiddenStates = make([]float32, maxSeq*hidden)
	m.attnScores = make([]float32, heads*maxSeq*maxSeq)
	m.qkvProj = make([]float32, maxSeq*3*hidden)
	m.qHeadBuf = make([]float32, heads*maxSeq*m.HeadDim)
	m.kHeadBuf = make([]float32, heads*maxSeq*m.HeadDim)
	m.vHeadBuf = make([]float32, heads*maxSeq*m.HeadDim)
	m.cHeadBuf = make([]float32, heads*maxSeq*m.HeadDim)
	m.attnOutput = make([]float32, maxSeq*hidden)
	m.ffnHidden = make([]float32, maxSeq*inter)
	m.tempHidden = make([]float32, maxSeq*hidden)
}

// Dim returns the embedding dimension.
func (m *ModelQ4) Dim() int { return m.HiddenSize }

// Close clears references for GC.
func (m *ModelQ4) Close() { *m = ModelQ4{} }

func readBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// IsQ4Model checks if a model file uses Q4 quantization by reading its magic.
func IsQ4Model(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := io.ReadFull(f, magic); err != nil {
		return false, err
	}
	switch string(magic) {
	case fileMagicQ4:
		return true, nil
	case fileMagic:
		return false, nil
	default:
		return false, errors.New("unknown model format: " + string(magic))
	}
}
