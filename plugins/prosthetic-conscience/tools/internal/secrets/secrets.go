// Package secrets is the single shared definition of matchable secret patterns.
// Every consumer (sc-secrets-gate, future telemetry redaction, any scrubber)
// imports this — patterns are defined exactly once. No public SDK ships these
// canonically (gitleaks/trufflehog carry their own databases as tool config),
// so this package is our source of truth.
//
// Precision beats recall: a false block on legitimate work is a bug.
package secrets

import "regexp"

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
