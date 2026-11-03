package gte

import (
	"unsafe"
)

// unsafeString converts a byte slice to a string without allocation.
// The string is only valid while the byte slice is not modified.
func unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func isPunctuation(b byte) bool {
	return (b >= 33 && b <= 47) || (b >= 58 && b <= 64) || (b >= 91 && b <= 96) || (b >= 123 && b <= 126)
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (m *Model) basicTokenize(text string) []string {
	return basicTokenizeWith(text, &m.basicBuf, &m.lowerBuf)
}

func basicTokenizeWith(text string, basicBuf *[]string, lowerBuf *[]byte) []string {
	tokens := (*basicBuf)[:0]
	i := 0
	for i < len(text) {
		for i < len(text) && isWhitespace(text[i]) {
			i++
		}
		if i >= len(text) {
			break
		}
		start := i
		if isPunctuation(text[i]) {
			i++
		} else {
			for i < len(text) && !isWhitespace(text[i]) && !isPunctuation(text[i]) {
				i++
			}
		}
		src := text[start:i]
		needsLower := false
		for j := 0; j < len(src); j++ {
			if src[j] >= 'A' && src[j] <= 'Z' {
				needsLower = true
				break
			}
		}
		if !needsLower {
			tokens = append(tokens, src)
		} else {
			// Reuse buffer for lowercase
			if cap(*lowerBuf) < len(src) {
				*lowerBuf = make([]byte, len(src)*2)
			}
			b := (*lowerBuf)[:len(src)]
			for j := 0; j < len(src); j++ {
				c := src[j]
				if c >= 'A' && c <= 'Z' {
					c += 32
				}
				b[j] = c
			}
			// Must copy to string since buffer will be reused
			tokens = append(tokens, string(b))
		}
	}
	*basicBuf = tokens
	return tokens
}

func (m *Model) wordpieceTokenize(word string, out []int) []int {
	return wordpieceTokenizeWith(word, out, m.vocabMap, &m.wpBuf)
}

func wordpieceTokenizeWith(word string, out []int, vocabMap map[string]int, wpBuf *[]byte) []int {
	if word == "" {
		return out
	}
	buf := (*wpBuf)[:0]
	start := 0
	for start < len(word) {
		end := len(word)
		found := -1
		for start < end {
			buf = buf[:0]
			if start > 0 {
				buf = append(buf, '#', '#')
			}
			buf = append(buf, word[start:end]...)
			// Lookup without allocating a string: use unsafe conversion.
			candidate := unsafeString(buf)
			if id, ok := vocabMap[candidate]; ok {
				found = id
				break
			}
			end--
		}
		if found < 0 {
			out = append(out, tokenUNK)
			start++
		} else {
			out = append(out, found)
			start = end
		}
	}
	*wpBuf = buf[:0]
	return out
}

func (m *Model) tokenize(text string) ([]int, []bool) {
	return tokenizeWith(text, m.vocabMap, m.MaxSeqLen, &m.tokenBuf, &m.attnMaskBuf, &m.basicBuf, &m.wpBuf, &m.lowerBuf)
}

// tokenizeWith is the shared tokenization logic usable by both FP32 and Q4 models.
func tokenizeWith(text string, vocabMap map[string]int, maxSeqLen int, tokenBuf *[]int, attnMaskBuf *[]bool, basicBuf *[]string, wpBuf *[]byte, lowerBuf *[]byte) ([]int, []bool) {
	basic := basicTokenizeWith(text, basicBuf, lowerBuf)

	if cap(*tokenBuf) < maxSeqLen {
		*tokenBuf = make([]int, 0, maxSeqLen)
	}
	if cap(*attnMaskBuf) < maxSeqLen {
		*attnMaskBuf = make([]bool, 0, maxSeqLen)
	}

	tokens := (*tokenBuf)[:0]
	attnMask := (*attnMaskBuf)[:0]

	tokens = append(tokens, tokenCLS)
	for _, tok := range basic {
		if len(tokens) >= maxSeqLen-1 {
			break
		}
		prevLen := len(tokens)
		tokens = wordpieceTokenizeWith(tok, tokens, vocabMap, wpBuf)
		if len(tokens) > maxSeqLen-1 {
			tokens = tokens[:prevLen]
			break
		}
	}
	if len(tokens) < maxSeqLen {
		tokens = append(tokens, tokenSEP)
	}
	attnMask = attnMask[:len(tokens)]
	for i := range attnMask {
		attnMask[i] = true
	}

	*tokenBuf = tokens
	*attnMaskBuf = attnMask
	return tokens, attnMask
}

// toLowerNoCopy lowercases ASCII bytes in-place and returns as string without allocation.
func toLowerNoCopy(b []byte) string {
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return unsafeString(b)
}
