package main

// BERT uncased tokenizer (HF BertNormalizer + BertPreTokenizer + WordPiece), as used by thenlper/gte-small:
// clean text, pad CJK ideographs with spaces, strip accents (NFD, drop Mn), lowercase, split on whitespace
// and punctuation, greedy longest-match WordPiece with "##" continuations, [CLS] ... [SEP], truncated to maxLen.

import (
	"bufio"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type Tokenizer struct {
	vocab           map[string]int64
	cls, sep, unk   int64
	maxLen          int
	maxCharsPerWord int
}

func LoadTokenizer(vocabPath string, maxLen int) (*Tokenizer, error) {
	f, err := os.Open(vocabPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &Tokenizer{vocab: map[string]int64{}, maxLen: maxLen, maxCharsPerWord: 100}
	s := bufio.NewScanner(f)
	var i int64
	for s.Scan() {
		t.vocab[strings.TrimRight(s.Text(), "\r")] = i
		i++
	}
	t.cls, t.sep, t.unk = t.vocab["[CLS]"], t.vocab["[SEP]"], t.vocab["[UNK]"]
	return t, s.Err()
}

func isControl(r rune) bool {
	if r == '\t' || r == '\n' || r == '\r' {
		return false
	}
	return unicode.In(r, unicode.Cc, unicode.Cf, unicode.Co, unicode.Cs)
}

func isWS(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || unicode.Is(unicode.Zs, r)
}

func isPunct(r rune) bool {
	if (r >= 33 && r <= 47) || (r >= 58 && r <= 64) || (r >= 91 && r <= 96) || (r >= 123 && r <= 126) {
		return true
	}
	return unicode.IsPunct(r)
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) || (r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x2A700 && r <= 0x2B73F) || (r >= 0x2B740 && r <= 0x2B81F) || (r >= 0x2B820 && r <= 0x2CEAF) ||
		(r >= 0xF900 && r <= 0xFAFF) || (r >= 0x2F800 && r <= 0x2FA1F)
}

func (t *Tokenizer) normalize(text string) string {
	var b strings.Builder
	for _, r := range text {
		if r == 0 || r == 0xFFFD || isControl(r) {
			continue
		}
		if isWS(r) {
			b.WriteByte(' ')
			continue
		}
		if isCJK(r) {
			b.WriteByte(' ')
			b.WriteRune(r)
			b.WriteByte(' ')
			continue
		}
		b.WriteRune(r)
	}
	s := norm.NFD.String(b.String())
	var o strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		for _, l := range strings.ToLower(string(r)) {
			o.WriteRune(l)
		}
	}
	return o.String()
}

func (t *Tokenizer) words(s string) []string {
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = cur[:0]
		}
	}
	for _, r := range s {
		switch {
		case isWS(r):
			flush()
		case isPunct(r):
			flush()
			out = append(out, string(r))
		default:
			cur = append(cur, r)
		}
	}
	flush()
	return out
}

func (t *Tokenizer) wordpiece(w string, ids []int64) []int64 {
	rs := []rune(w)
	if len(rs) > t.maxCharsPerWord {
		return append(ids, t.unk)
	}
	start := len(ids)
	for i := 0; i < len(rs); {
		j := len(rs)
		found := int64(-1)
		for ; j > i; j-- {
			sub := string(rs[i:j])
			if i > 0 {
				sub = "##" + sub
			}
			if id, ok := t.vocab[sub]; ok {
				found = id
				break
			}
		}
		if found < 0 {
			return append(ids[:start], t.unk)
		}
		ids = append(ids, found)
		i = j
	}
	return ids
}

// Encode returns [CLS] tokens [SEP], truncated so the total is at most maxLen.
func (t *Tokenizer) Encode(text string) []int64 {
	ids := []int64{t.cls}
	for _, w := range t.words(t.normalize(text)) {
		ids = t.wordpiece(w, ids)
		if len(ids) >= t.maxLen-1 {
			break
		}
	}
	if len(ids) > t.maxLen-1 {
		ids = ids[:t.maxLen-1]
	}
	return append(ids, t.sep)
}
