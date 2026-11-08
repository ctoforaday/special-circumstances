package vecembed

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

const (
	// HelperName is the binary name of the embedding native helper.
	HelperName = "feov-embed"
	// OpEmbed is the op code for vector embedding.
	OpEmbed nativehelper.Op = 16
	// Dim is the embedding dimension of gte-small.
	Dim = 384
	// BatchSize bounds a single request sent to the helper.
	BatchSize = 16
)

// Embedder computes dense vector representations of texts.
type Embedder interface {
	Embed(texts []string) ([][]float32, error)
	ModelID() string
	Close() error
}

// ClientEmbedder drives the feov-embed native helper executable.
type ClientEmbedder struct {
	client  *nativehelper.Client
	tok     *Tokenizer
	modelID string
}

// New creates a new ClientEmbedder backed by the feov-embed native helper.
func New() (*ClientEmbedder, error) {
	client, err := nativehelper.New(HelperName)
	if err != nil {
		return nil, err
	}
	digest := nativehelper.HelperDigest(HelperName)
	return &ClientEmbedder{
		client:  client,
		tok:     DefaultTokenizer(),
		modelID: "gte-small-int8@" + digest,
	}, nil
}

// NewWithClient creates a ClientEmbedder from an already initialized nativehelper.Client.
func NewWithClient(client *nativehelper.Client) *ClientEmbedder {
	digest := nativehelper.HelperDigest(HelperName)
	return &ClientEmbedder{
		client:  client,
		tok:     DefaultTokenizer(),
		modelID: "gte-small-int8@" + digest,
	}
}

// Default returns the default Embedder backed by the feov-embed native helper.
func Default() (Embedder, error) {
	return New()
}

// ModelID returns the model identifier string.
func (c *ClientEmbedder) ModelID() string {
	return c.modelID
}

// Close terminates the helper child process.
func (c *ClientEmbedder) Close() error {
	return c.client.Close()
}

// Embed encodes texts into L2-normalized 384-dimensional float32 vectors.
// Long batches are automatically split into sub-batches of at most BatchSize (16).
func (c *ClientEmbedder) Embed(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	out := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += BatchSize {
		end := start + BatchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[start:end]
		fields := make([][]byte, 2*len(batch))
		for i, text := range batch {
			ids := c.tok.Encode(text)
			idsBytes := make([]byte, len(ids)*8)
			maskBytes := make([]byte, len(ids)*8)
			for j, id := range ids {
				binary.LittleEndian.PutUint64(idsBytes[j*8:(j+1)*8], uint64(id))
				binary.LittleEndian.PutUint64(maskBytes[j*8:(j+1)*8], 1)
			}
			fields[2*i] = idsBytes
			fields[2*i+1] = maskBytes
		}

		res, err := c.client.Do(OpEmbed, fields...)
		if err != nil {
			return nil, err
		}
		if len(res.Result) != len(batch) {
			return nil, fmt.Errorf("vecembed: helper returned %d fields, want %d", len(res.Result), len(batch))
		}
		for i, r := range res.Result {
			if len(r) != Dim*4 {
				return nil, fmt.Errorf("vecembed: text %d result has length %d bytes, want %d", start+i, len(r), Dim*4)
			}
			vec := make([]float32, Dim)
			for d := 0; d < Dim; d++ {
				u := binary.LittleEndian.Uint32(r[d*4 : (d+1)*4])
				vec[d] = math.Float32frombits(u)
			}
			out = append(out, vec)
		}
	}
	return out, nil
}

// Cosine computes the cosine similarity between two L2-normalized vectors.
func Cosine(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}
