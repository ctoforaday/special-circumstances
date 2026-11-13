package peer

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ReadAgyTitle reads the Antigravity conversation's current title from its text-format protobuf
// annotation file (<annotationsDir>/<convID>.pbtxt). Missing or unreadable files return "" rather
// than a guess.
func ReadAgyTitle(annotationsDirs []string, convID string) string {
	if convID == "" || strings.ContainsAny(convID, `/\`) {
		return ""
	}
	for _, dir := range annotationsDirs {
		if dir == "" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, convID+".pbtxt"))
		if err != nil {
			continue
		}
		if title := parsePbtxtTitle(string(b)); title != "" {
			return title
		}
	}
	return ""
}

// DefaultAntigravityAnnotationsDirs returns the default Antigravity annotations directory under
// the user's home directory.
func DefaultAntigravityAnnotationsDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	return []string{filepath.Join(home, ".gemini", "antigravity-cli", "annotations")}
}

// parsePbtxtTitle extracts the top-level `title` string field from a protobuf text-format
// document. It scans tokens according to the prototext grammar (comments, identifiers, colons,
// nested message blocks `{...}` / `<...>`, and concatenated quoted strings with escapes).
func parsePbtxtTitle(src string) string {
	p := pbtxtParser{src: src}
	return p.topLevelTitle()
}

type pbtxtParser struct {
	src string
	pos int
}

func (p *pbtxtParser) skipWSAndComments() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' || c == ';' {
			p.pos++
			continue
		}
		if c == '#' {
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
			continue
		}
		break
	}
}

func (p *pbtxtParser) readIdent() string {
	p.skipWSAndComments()
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' || c == '/' {
			p.pos++
			continue
		}
		break
	}
	return p.src[start:p.pos]
}

func (p *pbtxtParser) topLevelTitle() string {
	for {
		p.skipWSAndComments()
		if p.pos >= len(p.src) {
			return ""
		}
		if p.src[p.pos] == '[' {
			// Extension name [pkg.field]
			p.skipBracketed('[', ']')
			p.skipOptionalColon()
			p.skipValue()
			continue
		}
		name := p.readIdent()
		if name == "" {
			p.pos++
			continue
		}
		p.skipOptionalColon()
		if name == "title" {
			if s, ok := p.readStringValue(); ok {
				return s
			}
		}
		p.skipValue()
	}
}

func (p *pbtxtParser) skipOptionalColon() {
	p.skipWSAndComments()
	if p.pos < len(p.src) && p.src[p.pos] == ':' {
		p.pos++
	}
}

func (p *pbtxtParser) skipValue() {
	p.skipWSAndComments()
	if p.pos >= len(p.src) {
		return
	}
	switch p.src[p.pos] {
	case '{':
		p.skipBlock('{', '}')
	case '<':
		p.skipBlock('<', '>')
	case '[':
		p.skipBlock('[', ']')
	case '"', '\'':
		_, _ = p.readStringValue()
	default:
		_ = p.readIdent()
	}
}

func (p *pbtxtParser) skipBracketed(openCh, closeCh byte) {
	if p.pos < len(p.src) && p.src[p.pos] == openCh {
		p.pos++
	}
	for p.pos < len(p.src) && p.src[p.pos] != closeCh {
		p.pos++
	}
	if p.pos < len(p.src) && p.src[p.pos] == closeCh {
		p.pos++
	}
}

func (p *pbtxtParser) skipBlock(openCh, closeCh byte) {
	depth := 0
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == '#' {
			p.skipWSAndComments()
			continue
		}
		if c == '"' || c == '\'' {
			_, _ = p.readSingleQuoted()
			continue
		}
		if c == openCh {
			depth++
		} else if c == closeCh {
			depth--
			p.pos++
			if depth <= 0 {
				return
			}
			continue
		}
		p.pos++
	}
}

func (p *pbtxtParser) readStringValue() (string, bool) {
	var b strings.Builder
	got := false
	for {
		p.skipWSAndComments()
		if p.pos >= len(p.src) || (p.src[p.pos] != '"' && p.src[p.pos] != '\'') {
			break
		}
		s, ok := p.readSingleQuoted()
		if !ok {
			return "", false
		}
		b.WriteString(s)
		got = true
	}
	return b.String(), got
}

func (p *pbtxtParser) readSingleQuoted() (string, bool) {
	quote := p.src[p.pos]
	p.pos++
	var b strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == quote {
			p.pos++
			return b.String(), true
		}
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.src) {
				return "", false
			}
			esc := p.src[p.pos]
			p.pos++
			switch esc {
			case 'a':
				b.WriteByte('\a')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'v':
				b.WriteByte('\v')
			case '\\', '"', '\'', '?':
				b.WriteByte(esc)
			case 'x':
				if p.pos+2 <= len(p.src) {
					if v, err := strconv.ParseUint(p.src[p.pos:p.pos+2], 16, 8); err == nil {
						b.WriteByte(byte(v))
						p.pos += 2
						continue
					}
				}
				return "", false
			case 'u':
				if p.pos+4 <= len(p.src) {
					if v, err := strconv.ParseUint(p.src[p.pos:p.pos+4], 16, 32); err == nil {
						b.WriteRune(rune(v))
						p.pos += 4
						continue
					}
				}
				return "", false
			case 'U':
				if p.pos+8 <= len(p.src) {
					if v, err := strconv.ParseUint(p.src[p.pos:p.pos+8], 16, 32); err == nil {
						b.WriteRune(rune(v))
						p.pos += 8
						continue
					}
				}
				return "", false
			default:
				if esc >= '0' && esc <= '7' {
					oct := string(esc)
					for i := 0; i < 2 && p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '7'; i++ {
						oct += string(p.src[p.pos])
						p.pos++
					}
					v, _ := strconv.ParseUint(oct, 8, 8)
					b.WriteByte(byte(v))
				} else {
					b.WriteByte(esc)
				}
			}
			continue
		}
		r, size := utf8.DecodeRuneInString(p.src[p.pos:])
		b.WriteRune(r)
		p.pos += size
	}
	return "", false
}
