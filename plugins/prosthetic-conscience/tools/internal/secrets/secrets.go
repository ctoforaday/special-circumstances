// Package secrets is the single shared definition of matchable secret patterns.
// Every consumer (sc-secrets-gate, future telemetry redaction, any scrubber)
// imports this — patterns are defined exactly once. No public SDK ships these
// canonically (gitleaks/trufflehog carry their own databases as tool config),
// so this package is our source of truth; why it is not generated from
// gitleaks' ruleset is written on Patterns.
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
type Pattern struct {
	Class    string
	Literals []string
	Re       *regexp.Regexp
}

// Patterns is the shared high-precision set, and the module's one declared table of secret
// patterns. A pattern is the honest tool here: the payload is free text, and each row is a
// vendor format published so that scanners can find it.
//
// Every body class is the vendor's WHOLE alphabet, never a subset of it. A class narrower
// than the alphabet fails silently and often, because of the \b anchors: once a character
// outside the class falls inside the minimum length, \b has nowhere to anchor and the row
// matches nothing. An Anthropic class without `_` misses 56% of base64url keys that way.
// patterns_test.go builds its fixtures from each format's real alphabet, so a narrowed class
// fails the suite.
//
// The table is kept by hand rather than generated from gitleaks' ruleset. Measured against
// gitleaks v8.30.1 on 2026-10-18 (box at load ~60, so the times are relative):
//   - 221 regex rules, all of which compile under RE2. 130 are meaningful only behind an
//     entropy threshold and 9 behind allowlists (the AWS rule allowlists every id ending
//     EXAMPLE), so taken as bare regexes they refuse other things than gitleaks does; and
//     generic-api-key matched 42 of the 803 text files in this repository.
//   - On a 1 MiB payload of this repository's text the 221 rules took ~30 s to MatchString,
//     ~1.5 s behind gitleaks' own keyword prefilter. This table, which runs before every
//     watched call, takes ~40 ms on 1 MiB holding none of its literals and ~1.5 s on 1 MiB
//     with a literal on every line (BenchmarkScanPayload).
//   - Its vendor rows are narrower than a gate wants: an exact-length Anthropic body, and a
//     private-key rule that needs 64 characters of key after the header. Gitleaks scans
//     files and filters afterwards; this gate refuses on the match.
var Patterns = []Pattern{
	// AKIA long-term, ASIA STS temporary, ABIA STS service bearer, ACCA context-specific.
	// The body is base32 ([A-Z2-7]); the class is a superset of it.
	{"AWS access key id", []string{"AKIA", "ASIA", "ABIA", "ACCA"},
		regexp.MustCompile(`\b(?:AKIA|ASIA|ABIA|ACCA)[0-9A-Z]{16}\b`)},
	// GitHub tokens are a prefix + 36 base62 characters, the last six a CRC32 checksum.
	{"GitHub personal access token", []string{"ghp_"},
		regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b`)},
	// github_pat_ + 22 base62 + `_` + 59 base62.
	{"GitHub fine-grained token", []string{"github_pat_"},
		regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{22,}\b`)},
	{"GitHub app/oauth token", []string{"gho_", "ghs_", "ghu_", "ghr_"},
		regexp.MustCompile(`\bgh[osur]_[A-Za-z0-9]{36}\b`)},
	// xoxb bot, xoxp user, xoxa/xoxr workspace and refresh, xoxs/xoxo legacy, xoxe rotating
	// refresh (an xoxe.xoxp- config token carries an xoxp- this row also sees), xapp
	// app-level. Bodies are alphanumeric segments joined by `-`.
	{"Slack token", []string{"xox", "xapp-"},
		regexp.MustCompile(`\b(?:xox[abeoprs]|xapp)-[A-Za-z0-9-]{10,}\b`)},
	// hooks.slack.com/services/T…/B…/<secret>; workflows/ and triggers/ URLs post as well.
	{"Slack webhook url", []string{"hooks.slack.com/"},
		regexp.MustCompile(`\bhooks\.slack\.com/(?:services|workflows|triggers)/[A-Za-z0-9/]{20,}`)},
	// PEM and OpenSSH armor (five dashes), PGP armor (… PRIVATE KEY BLOCK-----), and SSH2
	// armor (four dashes and a space: ---- BEGIN SSH2 ENCRYPTED PRIVATE KEY ----).
	{"private key block", []string{"PRIVATE KEY"},
		regexp.MustCompile(`-{4,5} ?BEGIN [A-Z0-9 ]*PRIVATE KEY(?: BLOCK)? ?-{4,5}`)},
	// A PuTTY .ppk file opens with its version line; its Private-Lines follow.
	{"PuTTY private key file", []string{"PuTTY-User-Key-File-"},
		regexp.MustCompile(`\bPuTTY-User-Key-File-[0-9]+:`)},
	// sk-ant-api03-, sk-ant-admin01-, sk-ant-oat01- …: a base64url body.
	{"Anthropic api key", []string{"sk-ant-"},
		regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_-]{20,}\b`)},
	// Project, service-account and admin keys: a base64url body.
	{"OpenAI api key", []string{"sk-proj-", "sk-svcacct-", "sk-admin-"},
		regexp.MustCompile(`\bsk-(?:proj|svcacct|admin)-[A-Za-z0-9_-]{20,}\b`)},
	// The legacy user key: sk- + 20 base62 + T3BlbkFJ ("OpenAI" in base64) + 20 base62.
	{"OpenAI legacy api key", []string{"T3BlbkFJ"},
		regexp.MustCompile(`\bsk-[A-Za-z0-9]{20}T3BlbkFJ[A-Za-z0-9]{20}\b`)},
}

// Scan returns the classes of every secret pattern found in text.
func Scan(text string) []string {
	var found []string
	for _, p := range Patterns {
		if containsAny(text, p.Literals) && p.Re.MatchString(text) {
			found = append(found, p.Class)
		}
	}
	return found
}

func containsAny(text string, literals []string) bool {
	for _, l := range literals {
		if strings.Contains(text, l) {
			return true
		}
	}
	return false
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
