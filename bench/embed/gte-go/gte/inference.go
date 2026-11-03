package gte

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/rcarmo/gte-go/gte/simd"
	"gonum.org/v1/gonum/blas"
	blasImpl32 "gonum.org/v1/gonum/blas/blas32"
	"gonum.org/v1/gonum/blas/gonum"
)

var blasImpl gonum.Implementation

func init() {
	blasImpl32.Use(blasImpl)
}

func blasTrans(t bool) blas.Transpose {
	if t {
		return blas.Trans
	}
	return blas.NoTrans
}

// linear computes Y = X·W^T + bias using our zero-alloc sgemm for small m.
func linear(y, x, w, b []float32, seqLen, inDim, outDim int) {
	sgemm(false, true, seqLen, outDim, inDim, 1.0, x, inDim, w, inDim, 0.0, y, outDim)

	// Add bias
	if b != nil {
		for s := 0; s < seqLen; s++ {
			yRow := y[s*outDim:][:outDim]
			for o := 0; o < outDim; o++ {
				yRow[o] += b[o]
			}
		}
	}
}

func layerNorm(out, x, gamma, beta []float32, seqLen, hidden int) {
	for s := 0; s < seqLen; s++ {
		rowStart := s * hidden
		mean := float32(0)
		for i := 0; i < hidden; i++ {
			mean += x[rowStart+i]
		}
		mean /= float32(hidden)

		variance := float32(0)
		for i := 0; i < hidden; i++ {
			diff := x[rowStart+i] - mean
			variance += diff * diff
		}
		variance /= float32(hidden)
		stdInv := fastInvSqrt(variance + layerNormEps)

		for i := 0; i < hidden; i++ {
			out[rowStart+i] = gamma[i]*(x[rowStart+i]-mean)*stdInv + beta[i]
		}
	}
}

// residualLayerNorm fuses out = layerNorm(x + residual, gamma, beta)
// Eliminates one pass over the data vs separate add + layerNorm.
func residualLayerNorm(out, x, residual, gamma, beta []float32, seqLen, hidden int) {
	for s := 0; s < seqLen; s++ {
		rowStart := s * hidden
		// Fused: add residual and compute mean in one pass
		mean := float32(0)
		for i := 0; i < hidden; i++ {
			v := x[rowStart+i] + residual[rowStart+i]
			x[rowStart+i] = v // store sum for variance pass
			mean += v
		}
		mean /= float32(hidden)

		variance := float32(0)
		for i := 0; i < hidden; i++ {
			diff := x[rowStart+i] - mean
			variance += diff * diff
		}
		variance /= float32(hidden)
		stdInv := fastInvSqrt(variance + layerNormEps)

		for i := 0; i < hidden; i++ {
			out[rowStart+i] = gamma[i]*(x[rowStart+i]-mean)*stdInv + beta[i]
		}
	}
}

func gelu(x []float32) {
	const c = float32(0.7978845608) // sqrt(2/pi)
	for i := range x {
		v := x[i]
		x[i] = 0.5 * v * (1 + fastTanh(c*(v+0.044715*v*v*v)))
	}
}

func softmax(x []float32) {
	maxVal := x[0]
	for _, v := range x[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	sum := float32(0)
	start := 0
	if simd.HasExpSub {
		start = len(x) &^ 7
		if start > 0 {
			sum = simd.ExpSub(x[:start], maxVal)
		}
	}
	for i := start; i < len(x); i++ {
		x[i] = fastExp(x[i] - maxVal)
		sum += x[i]
	}
	inv := 1 / sum
	for i := range x {
		x[i] *= inv
	}
}

func l2Normalize(x []float32) {
	norm := float32(0)
	for _, v := range x {
		norm += v * v
	}
	if norm == 0 {
		return
	}
	inv := fastInvSqrt(norm)
	for i := range x {
		x[i] *= inv
	}
}

func (m *Model) selfAttention(layer *LayerWeights, seqLen int, attnMask []bool) {
	hidden := m.HiddenSize
	heads := m.NumHeads
	headDim := m.HeadDim

	linear(m.qkvProj, m.hiddenStates, layer.QKVWeight, layer.QKVBias, seqLen, hidden, 3*hidden)
	// De-interleave QKV into contiguous per-head layout:
	//   qkvProj: [seqLen, 3*hidden] with hidden = heads*headDim
	//   qHeadBuf/kHeadBuf/vHeadBuf: [heads, seqLen, headDim] (contiguous per-head)
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

	// For short sequences, goroutine overhead exceeds parallelism benefit.
	if seqLen < 32 {
		for h := 0; h < heads; h++ {
			m.selfAttentionHeadScalar(h, seqLen, headDim, hidden, scale, attnMask)
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(heads)
		for h := 0; h < heads; h++ {
			h := h
			go func() {
				m.selfAttentionHeadBLAS(h, seqLen, headDim, hidden, scale, attnMask)
				wg.Done()
			}()
		}
		wg.Wait()
	}

	linear(m.tempHidden, m.attnOutput, layer.AttnOutputWeight, layer.AttnOutputBias, seqLen, hidden, hidden)
	residualLayerNorm(m.hiddenStates, m.tempHidden, m.hiddenStates, layer.AttnLnWeight, layer.AttnLnBias, seqLen, hidden)
}

// selfAttentionHeadScalar is the original scalar path, now using contiguous per-head buffers.
func (m *Model) selfAttentionHeadScalar(h, seqLen, headDim, hidden int, scale float32, attnMask []bool) {
	hoff := h * seqLen * headDim
	qBuf := m.qHeadBuf[hoff : hoff+seqLen*headDim]
	kBuf := m.kHeadBuf[hoff : hoff+seqLen*headDim]
	vBuf := m.vHeadBuf[hoff : hoff+seqLen*headDim]

	headOffset := h * seqLen * seqLen
	for i := 0; i < seqLen; i++ {
		rowOffset := headOffset + i*seqLen
		qi := qBuf[i*headDim : i*headDim+headDim]
		for j := 0; j < seqLen; j++ {
			kj := kBuf[j*headDim : j*headDim+headDim]
			score := scale * simd.Sdot(qi, kj)
			if attnMask != nil && !attnMask[j] {
				score = -10000
			}
			m.attnScores[rowOffset+j] = score
		}
		softmax(m.attnScores[rowOffset : rowOffset+seqLen])
	}

	cBuf := m.cHeadBuf[hoff : hoff+seqLen*headDim]
	for i := 0; i < seqLen; i++ {
		for d := 0; d < headDim; d++ {
			sum := float32(0)
			for j := 0; j < seqLen; j++ {
				sum += m.attnScores[headOffset+i*seqLen+j] * vBuf[j*headDim+d]
			}
			cBuf[i*headDim+d] = sum
		}
	}

	// Re-interleave: per-head [seqLen, headDim] -> [seqLen, hidden]
	for s := 0; s < seqLen; s++ {
		copy(m.attnOutput[s*hidden+h*headDim:s*hidden+h*headDim+headDim], cBuf[s*headDim:s*headDim+headDim])
	}
}

// selfAttentionHeadBLAS uses Sgemm for Q·K^T and attn·V, better for longer sequences.
// Head buffers are already in contiguous [seqLen, headDim] layout.
func (m *Model) selfAttentionHeadBLAS(h, seqLen, headDim, hidden int, scale float32, attnMask []bool) {
	off := h * seqLen * headDim
	qBuf := m.qHeadBuf[off : off+seqLen*headDim]
	kBuf := m.kHeadBuf[off : off+seqLen*headDim]
	vBuf := m.vHeadBuf[off : off+seqLen*headDim]
	cBuf := m.cHeadBuf[off : off+seqLen*headDim]

	// scores[seqLen, seqLen] = Q[seqLen, headDim] · K^T[headDim, seqLen]
	scoreOff := h * seqLen * seqLen
	scores := m.attnScores[scoreOff : scoreOff+seqLen*seqLen]
	sgemm(false, true, seqLen, seqLen, headDim, scale, qBuf, headDim, kBuf, headDim, 0.0, scores, seqLen)

	// Apply mask + softmax per row
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

	// context[seqLen, headDim] = scores[seqLen, seqLen] · V[seqLen, headDim]
	sgemm(false, false, seqLen, headDim, seqLen, 1.0, scores, seqLen, vBuf, headDim, 0.0, cBuf, headDim)

	// Re-interleave: per-head [seqLen, headDim] -> [seqLen, hidden]
	for s := 0; s < seqLen; s++ {
		srcOff := s * headDim
		dstOff := s*hidden + h*headDim
		copy(m.attnOutput[dstOff:dstOff+headDim], cBuf[srcOff:srcOff+headDim])
	}
}

func (m *Model) feedForward(layer *LayerWeights, seqLen int) {
	hidden := m.HiddenSize
	inter := m.Intermediate

	linear(m.ffnHidden, m.hiddenStates, layer.FfnInterWeight, layer.FfnInterBias, seqLen, hidden, inter)
	gelu(m.ffnHidden[:seqLen*inter])

	linear(m.tempHidden, m.ffnHidden, layer.FfnOutputWeight, layer.FfnOutputBias, seqLen, inter, hidden)
	residualLayerNorm(m.hiddenStates, m.tempHidden, m.hiddenStates, layer.FfnLnWeight, layer.FfnLnBias, seqLen, hidden)
}

func (m *Model) transformerForward(tokenIDs []int, seqLen int, attnMask []bool) {
	hidden := m.HiddenSize
	for s := 0; s < seqLen; s++ {
		tokenID := tokenIDs[s]
		base := s * hidden
		embOffset := tokenID * hidden
		posOffset := s * hidden
		for d := 0; d < hidden; d++ {
			m.hiddenStates[base+d] = m.TokenEmbeddings[embOffset+d] + m.PositionEmb[posOffset+d] + m.TokenTypeEmb[d]
		}
	}
	layerNorm(m.hiddenStates, m.hiddenStates, m.EmbedLnWeight, m.EmbedLnBias, seqLen, hidden)

	for l := 0; l < m.NumLayers; l++ {
		m.selfAttention(&m.Layers[l], seqLen, attnMask)
		m.feedForward(&m.Layers[l], seqLen)
	}
}

func meanPooling(output, hiddenStates []float32, attnMask []bool, seqLen, hidden int) {
	for i := range output {
		output[i] = 0
	}
	count := 0
	for s := 0; s < seqLen; s++ {
		if attnMask[s] {
			base := s * hidden
			for d := 0; d < hidden; d++ {
				output[d] += hiddenStates[base+d]
			}
			count++
		}
	}
	if count == 0 {
		return
	}
	inv := float32(1.0 / float64(count))
	for d := 0; d < hidden; d++ {
		output[d] *= inv
	}
}

// EmbedTo writes a normalized embedding into out (len must equal Dim()).
// The buffer is reused each call; callers should copy if they need to keep the result.
func (m *Model) EmbedTo(text string, out []float32) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if len(out) != m.HiddenSize {
		return fmt.Errorf("output buffer len %d != hidden size %d", len(out), m.HiddenSize)
	}

	tokenIDs, attnMask := m.tokenize(text)
	seqLen := len(tokenIDs)
	m.transformerForward(tokenIDs, seqLen, attnMask)

	meanPooling(out, m.hiddenStates, attnMask, seqLen, m.HiddenSize)
	l2Normalize(out)
	return nil
}

// Embed returns a normalized embedding for the provided text.
func (m *Model) Embed(text string) ([]float32, error) {
	output := make([]float32, m.HiddenSize)
	if err := m.EmbedTo(text, output); err != nil {
		return nil, err
	}
	return output, nil
}

// EmbedBatch embeds multiple texts sequentially and returns a matrix [n][dim].
func (m *Model) EmbedBatch(texts []string) ([][]float32, error) {
	if m == nil {
		return nil, errors.New("model is nil")
	}
	result := make([][]float32, len(texts))
	buf := make([]float32, m.HiddenSize)
	for i, t := range texts {
		if err := m.EmbedTo(t, buf); err != nil {
			return nil, fmt.Errorf("embed %d: %w", i, err)
		}
		copyBuf := make([]float32, m.HiddenSize)
		copy(copyBuf, buf)
		result[i] = copyBuf
	}
	return result, nil
}

// EmbedBatchParallel embeds multiple texts concurrently using n worker goroutines.
// Each worker has its own inference buffers but shares the model weights.
// n=0 uses runtime.NumCPU().
func (m *Model) EmbedBatchParallel(texts []string, n int) ([][]float32, error) {
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

	// Create worker models sharing weights but with independent buffers
	workers := make([]*Model, n)
	for i := range workers {
		w := *m // shallow copy: shares weight slices
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
		go func(worker *Model) {
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

// CosineSimilarity returns the cosine similarity for two normalized vectors.
func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("dimension mismatch: %d vs %d", len(a), len(b))
	}
	sum := float32(0)
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum, nil
}
