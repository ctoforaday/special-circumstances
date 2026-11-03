package gte

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/rcarmo/gte-go/gte/simd"
)

// selfAttentionQ4 runs one transformer block's self-attention with Q4 weights.
func (m *ModelQ4) selfAttentionQ4(layer *LayerWeightsQ4, seqLen int, attnMask []bool) {
	hidden := m.HiddenSize
	heads := m.NumHeads
	headDim := m.HeadDim

	// QKV projection: [seqLen, hidden] → [seqLen, 3*hidden]
	linearQ4(m.qkvProj, m.hiddenStates, layer.QKVWeightQ4, layer.QKVBias, seqLen, hidden, 3*hidden)

	// De-interleave QKV into per-head layout
	for s := 0; s < seqLen; s++ {
		src := m.qkvProj[s*3*hidden:]
		for h := 0; h < heads; h++ {
			hoff := h * seqLen * headDim
			dst := hoff + s*headDim
			srcH := h * headDim
			copy(m.qHeadBuf[dst:dst+headDim], src[srcH:srcH+headDim])
			copy(m.kHeadBuf[dst:dst+headDim], src[hidden+srcH:hidden+srcH+headDim])
			copy(m.vHeadBuf[dst:dst+headDim], src[2*hidden+srcH:2*hidden+srcH+headDim])
		}
	}

	scale := fastInvSqrt(float32(headDim))

	// Attention: Q·K^T (FP32 × FP32, both already dequantized from projection)
	for h := 0; h < heads; h++ {
		off := h * seqLen * headDim
		qBuf := m.qHeadBuf[off : off+seqLen*headDim]
		kBuf := m.kHeadBuf[off : off+seqLen*headDim]
		vBuf := m.vHeadBuf[off : off+seqLen*headDim]
		cBuf := m.cHeadBuf[off : off+seqLen*headDim]

		scoreOff := h * seqLen * seqLen
		scores := m.attnScores[scoreOff : scoreOff+seqLen*seqLen]

		// scores = Q · K^T * scale (FP32 sgemm)
		sgemm(false, true, seqLen, seqLen, headDim, scale, qBuf, headDim, kBuf, headDim, 0.0, scores, seqLen)

		for i := 0; i < seqLen; i++ {
			row := scores[i*seqLen : i*seqLen+seqLen]
			if attnMask != nil {
				for j := 0; j < seqLen; j++ {
					if !attnMask[j] {
						row[j] = -10000
					}
				}
			}
			softmax(row)
		}

		// context = scores · V (FP32 sgemm)
		sgemm(false, false, seqLen, headDim, seqLen, 1.0, scores, seqLen, vBuf, headDim, 0.0, cBuf, headDim)

		// Re-interleave
		for s := 0; s < seqLen; s++ {
			copy(m.attnOutput[s*hidden+h*headDim:s*hidden+h*headDim+headDim], cBuf[s*headDim:s*headDim+headDim])
		}
	}

	// Output projection (Q4) + residual LayerNorm
	linearQ4(m.tempHidden, m.attnOutput, layer.AttnOutputWeightQ4, layer.AttnOutputBias, seqLen, hidden, hidden)
	residualLayerNorm(m.hiddenStates, m.tempHidden, m.hiddenStates, layer.AttnLnWeight, layer.AttnLnBias, seqLen, hidden)
}

// feedForwardQ4 runs one transformer block's FFN with Q4 weights.
func (m *ModelQ4) feedForwardQ4(layer *LayerWeightsQ4, seqLen int) {
	hidden := m.HiddenSize
	inter := m.Intermediate

	linearQ4(m.ffnHidden, m.hiddenStates, layer.FfnInterWeightQ4, layer.FfnInterBias, seqLen, hidden, inter)
	gelu(m.ffnHidden[:seqLen*inter])

	linearQ4(m.tempHidden, m.ffnHidden, layer.FfnOutputWeightQ4, layer.FfnOutputBias, seqLen, inter, hidden)
	residualLayerNorm(m.hiddenStates, m.tempHidden, m.hiddenStates, layer.FfnLnWeight, layer.FfnLnBias, seqLen, hidden)
}

// transformerForwardQ4 runs the full forward pass with Q4 weights.
func (m *ModelQ4) transformerForwardQ4(tokenIDs []int, seqLen int, attnMask []bool) {
	hidden := m.HiddenSize
	blocksPerHidden := hidden / QK4_0
	rowBytes := blocksPerHidden * BlockQ4Size

	// Use tempHidden as scratch for embedding dequant (avoids allocation)
	posBuf := m.tempHidden[:hidden]

	for s := 0; s < seqLen; s++ {
		tokenID := tokenIDs[s]
		base := s * hidden
		// Dequantize embeddings on-the-fly and sum
		tokOff := tokenID * rowBytes
		posOff := s * rowBytes
		dequantQ4Into(m.hiddenStates[base:base+hidden], m.TokenEmbQ4[tokOff:tokOff+rowBytes], blocksPerHidden)
		// Add position embedding
		dequantQ4Into(posBuf, m.PositionEmbQ4[posOff:posOff+rowBytes], blocksPerHidden)
		for d := 0; d < hidden; d++ {
			m.hiddenStates[base+d] += posBuf[d]
		}
		// Add token type embedding (always type 0)
		dequantQ4Into(posBuf, m.TokenTypeEmbQ4[0:rowBytes], blocksPerHidden)
		for d := 0; d < hidden; d++ {
			m.hiddenStates[base+d] += posBuf[d]
		}
	}
	layerNorm(m.hiddenStates, m.hiddenStates, m.EmbedLnWeight, m.EmbedLnBias, seqLen, hidden)

	for l := 0; l < m.NumLayers; l++ {
		m.selfAttentionQ4(&m.Layers[l], seqLen, attnMask)
		m.feedForwardQ4(&m.Layers[l], seqLen)
	}
}

// tokenize reuses the shared tokenizer logic.
func (m *ModelQ4) tokenize(text string) ([]int, []bool) {
	return tokenizeWith(text, m.vocabMap, m.MaxSeqLen, &m.tokenBuf, &m.attnMaskBuf, &m.basicBuf, &m.wpBuf, &m.lowerBuf)
}

// EmbedTo writes a normalized embedding into out.
func (m *ModelQ4) EmbedTo(text string, out []float32) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if len(out) != m.HiddenSize {
		return fmt.Errorf("output buffer len %d != hidden size %d", len(out), m.HiddenSize)
	}

	tokenIDs, attnMask := m.tokenize(text)
	seqLen := len(tokenIDs)
	m.transformerForwardQ4(tokenIDs, seqLen, attnMask)

	meanPooling(out, m.hiddenStates, attnMask, seqLen, m.HiddenSize)
	l2Normalize(out)
	return nil
}

// Embed returns a normalized embedding for the provided text.
func (m *ModelQ4) Embed(text string) ([]float32, error) {
	output := make([]float32, m.HiddenSize)
	if err := m.EmbedTo(text, output); err != nil {
		return nil, err
	}
	return output, nil
}

// EmbedBatch embeds multiple texts sequentially.
func (m *ModelQ4) EmbedBatch(texts []string) ([][]float32, error) {
	if m == nil {
		return nil, errors.New("model is nil")
	}
	result := make([][]float32, len(texts))
	buf := make([]float32, m.HiddenSize)
	for i, t := range texts {
		if err := m.EmbedTo(t, buf); err != nil {
			return nil, fmt.Errorf("embed %d: %w", i, err)
		}
		out := make([]float32, m.HiddenSize)
		copy(out, buf)
		result[i] = out
	}
	return result, nil
}

// EmbedBatchParallel embeds concurrently with n workers.
func (m *ModelQ4) EmbedBatchParallel(texts []string, n int) ([][]float32, error) {
	if m == nil {
		return nil, errors.New("model is nil")
	}
	if len(texts) == 0 {
		return nil, nil
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	if n > len(texts) {
		n = len(texts)
	}

	workers := make([]*ModelQ4, n)
	for i := range workers {
		w := *m
		w.initBuffers()
		workers[i] = &w
	}

	result := make([][]float32, len(texts))
	errs := make([]error, len(texts))

	var wg sync.WaitGroup
	textCh := make(chan int, len(texts))
	for i := range texts {
		textCh <- i
	}
	close(textCh)

	for w := 0; w < n; w++ {
		wg.Add(1)
		go func(worker *ModelQ4) {
			defer wg.Done()
			buf := make([]float32, worker.HiddenSize)
			for idx := range textCh {
				if err := worker.EmbedTo(texts[idx], buf); err != nil {
					errs[idx] = err
					continue
				}
				out := make([]float32, worker.HiddenSize)
				copy(out, buf)
				result[idx] = out
			}
		}(workers[w])
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("embed %d: %w", i, err)
		}
	}
	return result, nil
}

// Unused import guard for simd package in non-SIMD builds.
var _ = simd.DotQ4
