package fetchcache

import (
	"mime"
	"net/url"
	"path"
	"strings"
	"unicode"
)

// maxLabelRunes bounds a filename label. A PDF `Title` is author-supplied and unbounded;
// the 2026-08-29 corpus turned up one 96-character title (IEEE 1012's, which repeats itself
// twice before naming the standard). The cap is on RUNES, not bytes, so a CJK title is
// truncated to a readable length rather than to a third of one.
const maxLabelRunes = 96

// Label picks the human-readable name for a cached document, per #629 D4: the document's own
// `Title`, then the `Content-Disposition` filename, then the tail of the URL.
//
// EVERY RUNG IS LOAD-BEARING, and that is measured rather than assumed. Across the four cited
// documents plus the wider corpus: three carried a meaningful `Title` ("Active Learning
// Literature Survey"), `little.pdf` carried NONE, and not one of them was served with a
// `Content-Disposition` header — so a Content-Disposition-first ordering would have fallen
// through every time. URL-tail-only is worse still: the Auer paper lives at
// `https://inria.hal.science/inria-00574987/document` and yields the filename "document".
//
// IT IS A LABEL, NEVER AN IDENTITY. The sha256 is the identity; nothing looks a document up by
// this string, and the cache file on disk is still named for its hash. That separation is why
// author-supplied text is safe to carry here at all — see sanitizeLabel for the other half.
func Label(docTitle, disposition, rawURL string) string {
	if s := sanitizeLabel(docTitle); s != "" {
		return s
	}
	if s := sanitizeLabel(dispositionFilename(disposition)); s != "" {
		return s
	}
	return sanitizeLabel(urlTail(rawURL))
}

// dispositionFilename reads the filename from a Content-Disposition header, preferring the
// RFC 5987 `filename*` form Go's mime package already decodes. A malformed header yields "",
// never an error a caller might be tempted to surface: the header is a courtesy, and its
// absence is the common case, not a fault.
func dispositionFilename(disposition string) string {
	if strings.TrimSpace(disposition) == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return ""
	}
	return params["filename"]
}

// urlTail is the last non-empty path segment of rawURL, with any query and fragment already
// dropped by the parse. An unparseable URL yields "" rather than a guess at its shape.
func urlTail(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return path.Base(strings.TrimRight(u.Path, "/"))
}

// sanitizeLabel makes untrusted author text safe to print and to type at a shell.
//
// THE INPUT IS HOSTILE UNTIL PROVEN OTHERWISE. A PDF `Title` is whatever an author put in the
// document metadata, and it reaches a seat's context and then, plausibly, a seat's `cp`
// command. So this strips every path separator (a title of "../../etc/passwd" becomes
// "etcpasswd"), every C0/C1 control character (a title carrying an ANSI escape or a newline
// could otherwise forge a line of the tool's own output), and leading dots (no dotfiles, no
// "..").
//
// It deliberately does NOT try to produce a UNIQUE name, because uniqueness is the sha's job.
// Two documents may legitimately carry the same label; nothing downstream cares, because
// nothing downstream keys on it.
func sanitizeLabel(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '/' || r == '\\' || r == ':':
			continue
		case r < 0x20 || (r >= 0x7f && r <= 0x9f):
			continue
		case unicode.IsSpace(r):
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	out = strings.TrimLeft(out, ". ")
	out = strings.TrimRight(out, " ")
	if r := []rune(out); len(r) > maxLabelRunes {
		out = strings.TrimRight(string(r[:maxLabelRunes]), " ")
	}
	return out
}

// MediaType is the bare media type of a Content-Type header — "application/pdf" from
// "application/pdf; charset=binary". Parameters are dropped because every consumer here
// switches on the type alone, and keeping the charset would make two spellings of one fact.
// An unparseable header yields "", which the index records as "not measured" rather than as
// a guess.
func MediaType(contentType string) string {
	if strings.TrimSpace(contentType) == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(mt))
}
