package fetchcache

import (
	"net/url"
	"strings"
)

// SameResource reports whether two URLs name one resource under RFC 3986's syntax-based
// normalisation (§6.2.2): scheme and host compared without case, a port the scheme defaults to
// dropped, an empty path read as "/", a percent-encoded UNRESERVED character read as the character
// itself (`%7E` is `~`), the hex of every other escape compared without case, and the fragment —
// which no request carries — ignored. Nothing beyond the syntax is assumed: a query is compared
// as written once its escapes are normalised, and a trailing slash still makes a different path.
//
// IT IS WHY TWO SPELLINGS OF ONE URL ARE NOT A BOT WALL. The shared-body warning fires when one
// body answers two URLs, because two sources do not answer with one body; in smoke 07d it fired
// on `/%7Esdmiller/…` beside `/~sdmiller/…` — one lecture, spelled twice — and two seats spent a
// read each confirming the source was real.
func SameResource(a, b string) bool {
	if a == b {
		return true
	}
	na, oka := normalURL(a)
	nb, okb := normalURL(b)
	return oka && okb && na == nb
}

// normalURL is the RFC 3986 §6.2.2 form SameResource compares; false for a string that is not an
// absolute URL, which compares only as itself.
func normalURL(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", false
	}
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	if p := u.Port(); p != "" && !(scheme == "http" && p == "80") && !(scheme == "https" && p == "443") {
		host += ":" + p
	}
	path := normalEscapes(u.EscapedPath())
	if path == "" {
		path = "/"
	}
	out := scheme + "://" + host + path
	if u.RawQuery != "" || u.ForceQuery {
		out += "?" + normalEscapes(u.RawQuery)
	}
	return out, true
}

// normalEscapes decodes each percent-escape of an unreserved character (ALPHA, DIGIT, `-`, `.`,
// `_`, `~`) and upper-cases the hex of every other one, which RFC 3986 §6.2.2.1-2 make equivalent.
func normalEscapes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			c := unhex(s[i+1])<<4 | unhex(s[i+2])
			if isUnreserved(c) {
				b.WriteByte(c)
			} else {
				b.WriteByte('%')
				b.WriteString(strings.ToUpper(s[i+1 : i+3]))
			}
			i += 2
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isHex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	}
	return c - 'A' + 10
}

func isUnreserved(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~'
}
