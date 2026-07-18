package record

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Payload is an INSERTION-ORDERED string-keyed map.
//
// Why not map[string]any: the mjs oracle builds payloads as JavaScript object
// literals, whose JSON.stringify order is insertion order. Go marshals maps with
// SORTED keys. The differential gate compares events structurally, so ordering
// alone would not fail it — but the RENDER side compares byte-identical, the
// telemetry render embeds a payload-shaped object (new_mint.by_severity), and a
// sorted-key divergence there is exactly the language-boundary drift class the
// R2g plan warns about. Preserving order costs one type and removes the whole
// question.
//
// Absence is meaningful and distinct from empty: mjs drops `undefined` keys at
// stringify time, so a flag the seat never passed must not appear in the event at
// all. Set is therefore only called for values that were actually supplied.
type Payload struct {
	keys []string
	m    map[string]any
}

func NewPayload() *Payload { return &Payload{m: map[string]any{}} }

// Set appends a key in first-write order; re-setting an existing key updates the
// value in place and keeps its original position (mirrors JS object semantics).
func (p *Payload) Set(k string, v any) *Payload {
	if p.m == nil {
		p.m = map[string]any{}
	}
	if _, seen := p.m[k]; !seen {
		p.keys = append(p.keys, k)
	}
	p.m[k] = v
	return p
}

// SetIf sets the key only when present is true — the `undefined` drop rule. Use
// it wherever the mjs source reads an optional arg straight into a literal.
func (p *Payload) SetIf(present bool, k string, v any) *Payload {
	if present {
		p.Set(k, v)
	}
	return p
}

func (p *Payload) Get(k string) (any, bool) {
	if p == nil || p.m == nil {
		return nil, false
	}
	v, ok := p.m[k]
	return v, ok
}

// Str returns the value as a string when it is one; a boolean flag (--severity
// with no argument) is deliberately NOT coerced, so it fails enum validation the
// same way `true` does in the oracle.
func (p *Payload) Str(k string) string {
	v, ok := p.Get(k)
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func (p *Payload) Has(k string) bool { _, ok := p.Get(k); return ok }

func (p *Payload) Keys() []string { return append([]string(nil), p.keys...) }

// StrList reads a value written by Set as []string.
func (p *Payload) StrList(k string) []string {
	v, ok := p.Get(k)
	if !ok {
		return nil
	}
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func (p Payload) MarshalJSON() ([]byte, error) {
	if len(p.keys) == 0 {
		return []byte("{}"), nil
	}
	var b bytes.Buffer
	b.WriteByte('{')
	for i, k := range p.keys {
		if i > 0 {
			b.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		b.Write(kb)
		b.WriteByte(':')
		vb, err := json.Marshal(p.m[k])
		if err != nil {
			return nil, fmt.Errorf("payload key %q: %w", k, err)
		}
		b.Write(vb)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func (p *Payload) UnmarshalJSON(data []byte) error {
	p.keys = nil
	p.m = map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // keep the oracle's numeric text verbatim; no float round-trip drift
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("payload: expected object, got %v", tok)
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		k, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("payload: non-string key %v", keyTok)
		}
		var v any
		if err := dec.Decode(&v); err != nil {
			return err
		}
		p.Set(k, v)
	}
	_, err = dec.Token() // closing brace
	return err
}
