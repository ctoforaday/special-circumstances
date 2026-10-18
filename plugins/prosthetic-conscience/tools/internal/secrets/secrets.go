// Package secrets is the single shared definition of matchable secret patterns.
// Every consumer (sc-secrets-gate, future telemetry redaction, any scrubber)
// imports this — patterns are defined exactly once. No public SDK ships these
// canonically (gitleaks/trufflehog carry their own databases as tool config),
// so this package is our source of truth: a hand table, and the rows generated
// from gitleaks' ruleset that are accurate as bare regexes (see Patterns).
//
// Precision beats recall: a false block on legitimate work is a bug.
package secrets

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

// Pattern pairs a human-readable class name with a high-precision regex, and the literals
// every match of it contains.
//
// Literals are a prefilter: the regex runs only on text holding one of them. Go's regexp
// skips ahead by a literal prefix, and a \b-anchored pattern has none, so it steps through
// every byte of the payload. strings.Contains does the same skip far faster, and the
// prefilter cannot change a verdict as long as every match contains one of the literals —
// which patterns_test.go's generated fixtures enforce, since a sample whose literal is missing
// is a sample the row misses.
//
// Fold marks a row whose regex matches its literals case-insensitively. Its literals are
// lowercase, and the prefilter looks for them in the text folded by foldText.
type Pattern struct {
	Class    string
	Literals []string
	Fold     bool
	Re       *regexp.Regexp
}

// Patterns is the shared high-precision set, and the module's one table of secret patterns:
// the hand rows below, then gitleaksPatterns, which scripts/secretsgen writes from a pinned
// gitleaks ruleset. A pattern is the honest tool here: the payload is free text, and each row
// is a vendor format published so that scanners can find it. A refusal names the row's Class;
// for a generated row that is the gitleaks rule id.
var Patterns = append(handPatterns[:len(handPatterns):len(handPatterns)], gitleaksPatterns...)

// handPatterns are the rows kept by hand, where they are stronger than gitleaks' own rule for
// the vendor; secretsgen's handCovered names each gitleaks rule set aside for one, and why.
//
// Every body class is the vendor's WHOLE alphabet, never a subset of it. A class narrower
// than the alphabet fails silently and often, because of the \b anchors: once a character
// outside the class falls inside the minimum length, \b has nowhere to anchor and the row
// matches nothing. An Anthropic class without `_` misses 56% of base64url keys that way.
// patterns_test.go builds its fixtures from each format's real alphabet, so a narrowed class
// fails the suite.
//
// Most of gitleaks' rules are not taken as rows. Measured against gitleaks v8.30.1 on
// 2026-10-18: 130 of its 222 rules are meaningful only behind an entropy threshold and 9
// behind allowlists (the AWS rule allowlists every id ending EXAMPLE), so taken as bare
// regexes they refuse other things than gitleaks does; and generic-api-key matched 42 of the
// 803 text files in this repository. The hand rows are also broader than gitleaks' rule for
// the same vendor where a gate needs it: gitleaks' Anthropic rule needs an exact-length body,
// and its private-key rule needs 64 characters of key after the header.
var handPatterns = []Pattern{
	// AKIA long-term, ASIA STS temporary, ABIA STS service bearer, ACCA context-specific.
	// The body is base32 ([A-Z2-7]); the class is a superset of it.
	{Class: "AWS access key id", Literals: []string{"AKIA", "ASIA", "ABIA", "ACCA"},
		Re: regexp.MustCompile(`\b(?:AKIA|ASIA|ABIA|ACCA)[0-9A-Z]{16}\b`)},
	// GitHub tokens are a prefix + 36 base62 characters, the last six a CRC32 checksum.
	{Class: "GitHub personal access token", Literals: []string{"ghp_"},
		Re: regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b`)},
	// github_pat_ + 22 base62 + `_` + 59 base62.
	{Class: "GitHub fine-grained token", Literals: []string{"github_pat_"},
		Re: regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{22,}\b`)},
	{Class: "GitHub app/oauth token", Literals: []string{"gho_", "ghs_", "ghu_", "ghr_"},
		Re: regexp.MustCompile(`\bgh[osur]_[A-Za-z0-9]{36}\b`)},
	// xoxb bot, xoxp user, xoxa/xoxr workspace and refresh, xoxs/xoxo legacy, xoxe rotating
	// refresh (an xoxe.xoxp- config token carries an xoxp- this row also sees), xapp
	// app-level. Bodies are alphanumeric segments joined by `-`.
	{Class: "Slack token", Literals: []string{"xox", "xapp-"},
		Re: regexp.MustCompile(`\b(?:xox[abeoprs]|xapp)-[A-Za-z0-9-]{10,}\b`)},
	// hooks.slack.com/services/T…/B…/<secret>; workflows/ and triggers/ URLs post as well.
	{Class: "Slack webhook url", Literals: []string{"hooks.slack.com/"},
		Re: regexp.MustCompile(`\bhooks\.slack\.com/(?:services|workflows|triggers)/[A-Za-z0-9/]{20,}`)},
	// PEM and OpenSSH armor (five dashes), PGP armor (… PRIVATE KEY BLOCK-----), and SSH2
	// armor (four dashes and a space: ---- BEGIN SSH2 ENCRYPTED PRIVATE KEY ----).
	{Class: "private key block", Literals: []string{"PRIVATE KEY"},
		Re: regexp.MustCompile(`-{4,5} ?BEGIN [A-Z0-9 ]*PRIVATE KEY(?: BLOCK)? ?-{4,5}`)},
	// A PuTTY .ppk file opens with its version line; its Private-Lines follow.
	{Class: "PuTTY private key file", Literals: []string{"PuTTY-User-Key-File-"},
		Re: regexp.MustCompile(`\bPuTTY-User-Key-File-[0-9]+:`)},
	// sk-ant-api03-, sk-ant-admin01-, sk-ant-oat01- …: a base64url body.
	{Class: "Anthropic api key", Literals: []string{"sk-ant-"},
		Re: regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_-]{20,}\b`)},
	// Project, service-account and admin keys: a base64url body.
	{Class: "OpenAI api key", Literals: []string{"sk-proj-", "sk-svcacct-", "sk-admin-"},
		Re: regexp.MustCompile(`\bsk-(?:proj|svcacct|admin)-[A-Za-z0-9_-]{20,}\b`)},
	// The legacy user key: sk- + 20 base62 + T3BlbkFJ ("OpenAI" in base64) + 20 base62.
	{Class: "OpenAI legacy api key", Literals: []string{"T3BlbkFJ"},
		Re: regexp.MustCompile(`\bsk-[A-Za-z0-9]{20}T3BlbkFJ[A-Za-z0-9]{20}\b`)},
}

// Scan returns the classes of every secret pattern found in text.
func Scan(text string) []string {
	var found []string
	var folded *string // built once, for the first Fold literal searched
	// Each distinct literal is searched for at most once: rows share them (gitleaks' three
	// New Relic rules have the same three), and on 1 MiB a search is the scan's main cost.
	present := make([]int8, len(prefilter.lits)) // 0 not yet searched, 1 present, -1 absent
	has := func(li int) bool {
		if present[li] == 0 {
			l, in := prefilter.lits[li], text
			if l.fold {
				if folded == nil {
					f := foldText(text)
					folded = &f
				}
				in = *folded
			}
			present[li] = -1
			if strings.Contains(in, l.s) {
				present[li] = 1
			}
		}
		return present[li] == 1
	}
	for i, p := range Patterns {
		for _, li := range prefilter.rows[i] {
			if has(li) {
				if p.Re.MatchString(text) {
					found = append(found, p.Class)
				}
				break
			}
		}
	}
	return found
}

// literal is one distinct prefilter string; fold marks one searched for in folded text.
type literal struct {
	s    string
	fold bool
}

// prefilter holds every distinct literal of Patterns, and each row's literals as indices
// into it, in Patterns' order.
var prefilter = buildPrefilter(Patterns)

func buildPrefilter(ps []Pattern) (pf struct {
	lits []literal
	rows [][]int
}) {
	index := map[literal]int{}
	for _, p := range ps {
		var row []int
		for _, s := range p.Literals {
			l := literal{s, p.Fold}
			li, ok := index[l]
			if !ok {
				li = len(pf.lits)
				index[l] = li
				pf.lits = append(pf.lits, l)
			}
			row = append(row, li)
		}
		pf.rows = append(pf.rows, row)
	}
	return pf
}

// foldText maps every character a case-insensitive regex treats as an ASCII letter onto that
// lowercase letter: A-Z, and the two non-ASCII runes in an ASCII letter's case-fold orbit,
// U+017F (long s, UTF-8 C5 BF) and U+212A (Kelvin sign, E2 84 AA). A (?i) literal matches
// exactly the strings that fold to its lowercase form, so the prefilter on folded text cannot
// change a verdict. strings.ToLower alone would not do: it leaves U+017F as it is. Bytes, not
// runes: a valid UTF-8 sequence cannot be misread at a byte offset inside another, and this
// runs over every payload a Fold row sees.
func foldText(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case 'A' <= c && c <= 'Z':
			c += 'a' - 'A'
		case c == 0xC5 && i+1 < len(text) && text[i+1] == 0xBF:
			c, i = 's', i+1
		case c == 0xE2 && i+2 < len(text) && text[i+1] == 0x84 && text[i+2] == 0xAA:
			c, i = 'k', i+2
		}
		b.WriteByte(c)
	}
	return b.String()
}

// ScanPayload scans a JSON tool payload for secrets — decoded, not as wire bytes.
//
// WHY THIS EXISTS. sc-secrets-gate originally scanned the raw encoded JSON:
// Scan(string(in.ToolInput)). Every pattern here is ASCII-alphanumeric and a
// standard encoder writes those literally, so the common case matched and the
// gate looked correct. It is not. JSON may encode any character as \uXXXX, and
// the escaped form of an identical secret does not match a pattern written
// against the decoded text. Measured against the shipped binary:
//
//	{"command":"echo AKIA…EXAMPLE"}        -> DENY
//	{"command":"echo AKIA…EXAMPLE"}   -> ALLOW   (same decoded string)
//	{"command":"-----BEGIN RSA PRIVATE…"}  -> DENY
//	{"command":"-----BEGIN RSA PRI…"} -> ALLOW   (same decoded string)
//
// A gate whose result depends on the sender's escaping choices is not a gate.
// So: decode first, scan every string value, and scan the raw form too — the
// raw pass costs nothing and still catches anything the walk cannot reach.
//
// Undecodable input falls back to the raw scan rather than allowing: malformed
// JSON is the caller's problem, but it must not become a bypass.
func ScanPayload(raw []byte) []string {
	seen := map[string]bool{}
	var out []string
	add := func(classes []string) {
		for _, c := range classes {
			if !seen[c] {
				seen[c] = true
				out = append(out, c)
			}
		}
	}

	add(Scan(string(raw)))

	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return out // raw-only; a payload we cannot parse still gets scanned
	}
	// Values are joined by newline rather than concatenated: a newline is a
	// non-word character, so it cannot fuse two innocent values into something
	// that satisfies a \b-anchored pattern.
	add(Scan(strings.Join(strs(doc, nil), "\n")))
	return out
}

// strs walks decoded JSON and collects every string it contains, keys included —
// a credential pasted as an object key is still a credential on the wire.
func strs(v any, acc []string) []string {
	switch t := v.(type) {
	case string:
		acc = append(acc, t)
	case []any:
		for _, e := range t {
			acc = strs(e, acc)
		}
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys) // deterministic: the same payload must scan identically
		for _, k := range keys {
			acc = append(acc, k)
			acc = strs(t[k], acc)
		}
	}
	return acc
}
