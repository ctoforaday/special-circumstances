// Package secrets is the single shared definition of matchable secret patterns.
// Every consumer (sc-secrets-gate, future telemetry redaction, any scrubber)
// imports this — patterns are defined exactly once. No public SDK ships these
// canonically (gitleaks/trufflehog carry their own databases as tool config),
// so this package is our source of truth.
//
// Precision beats recall: a false block on legitimate work is a bug.
package secrets

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

// Pattern pairs a human-readable class name with a high-precision regex.
type Pattern struct {
	Class string
	Re    *regexp.Regexp
}

// Patterns is the shared high-precision set.
var Patterns = []Pattern{
	{"AWS access key id", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"GitHub personal access token", regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b`)},
	{"GitHub fine-grained token", regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{22,}\b`)},
	{"GitHub app/oauth token", regexp.MustCompile(`\bgh[osur]_[A-Za-z0-9]{36}\b`)},
	{"Slack token", regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`)},
	{"private key block", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"Anthropic api key", regexp.MustCompile(`\bsk-ant-[A-Za-z0-9-]{20,}\b`)},
	{"OpenAI api key", regexp.MustCompile(`\bsk-proj-[A-Za-z0-9_-]{20,}\b`)},
}

// Scan returns the classes of every secret pattern found in text.
func Scan(text string) []string {
	var found []string
	for _, p := range Patterns {
		if p.Re.MatchString(text) {
			found = append(found, p.Class)
		}
	}
	return found
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
