//go:build noort

package main

const engineName = "noort-stub"

type engine struct{}

func newEngine(model string, threads int) (*engine, error) { return &engine{}, nil }

func (*engine) run(ids, mask []int64) ([]float32, error) {
	out := make([]float32, len(ids)*dim)
	for i := range ids {
		out[i*dim] = 1
	}
	return out, nil
}
