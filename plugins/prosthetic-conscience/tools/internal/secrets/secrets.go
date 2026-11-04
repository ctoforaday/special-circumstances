// Package secrets finds secrets in a tool payload with gitleaks' detector
// (github.com/zricethezav/gitleaks/v8, MIT) — its default configuration, as the gitleaks
// command line runs it, extended with the rules in gitleakstoml/gitleaks.toml. table.go holds
// that configuration as gitleaks translates it, written by devcmd/secretsgen.
//
// The default configuration is the vendor rules, their entropy thresholds, the allowlists (the
// AWS documentation key ending EXAMPLE, gitleaks' stopwords) and the keyword prefilter. The
// extension adds a rule only where a real key format is missed by the default; gitleaks.toml
// names the gap beside each one. Measured on 50,843 hook inputs from this machine's
// transcripts, the default configuration refused 3, all of them secrets tooling quoting its
// own fixtures.
//
// A finding reaches a caller as its gitleaks rule id and nothing else: the detector redacts
// every finding, its logger is discarded, and ScanPayload returns ids. The matched value never
// reaches a refusal, a log line or an error.
package secrets

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/logging"
)

// maxDecodeDepth is the gitleaks command line's default for --max-decode-depth: a secret
// base64-, hex- or percent-encoded up to five times over is decoded and found.
const maxDecodeDepth = 5

// detector builds a detector over the rules keep selects from the generated table.
//
// The table is gitleaks.toml as gitleaks translates it, so no call parses TOML, and a call
// compiles only the rules keep holds: compiling every rule is most of a hook call's cost.
func detector(keep func(*RuleSpec) bool) (*detect.Detector, error) {
	// gitleaks logs findings — the secret included — at debug and trace level, to stderr.
	// Nothing it logs is for this gate's caller.
	zerolog.SetGlobalLevel(zerolog.Disabled)
	logging.Logger = zerolog.Nop()

	cfg, err := table.Config(keep)
	if err != nil {
		return nil, err
	}
	return configure(detect.NewDetector(cfg)), nil
}

// configure sets the detector's options: the gitleaks command line's, except that an allow
// comment is ignored and every finding is redacted.
func configure(d *detect.Detector) *detect.Detector {
	d.MaxDecodeDepth = maxDecodeDepth
	// A `gitleaks:allow` comment silences a finding on its line. In a tool payload the agent
	// writes that line, so honouring it would let the payload exempt itself.
	d.IgnoreGitleaksAllow = true
	// Every finding comes back with Secret, Match and Line redacted.
	d.Redact = 100
	return d
}

// scan returns the rule id of every finding in text.
func scan(d *detect.Detector, text string) []string {
	var ids []string
	for _, f := range d.DetectString(text) {
		ids = append(ids, f.RuleID)
	}
	return ids
}

// ScanPayload returns the gitleaks rule ids of the secrets in a JSON tool payload, sorted and
// without repeats. An error means the detector could not be built; the payload was not scanned.
//
// The payload is scanned decoded as well as raw. JSON may write any character as \uXXXX, and
// the escaped form of a secret is not the secret: a gate reading only the wire bytes refuses
// `AKIA…` and allows the same key with its first letter written as A. So every string the
// payload holds — object keys included, since a key pasted as a key is still on the wire — is
// scanned, joined by newlines so two innocent neighbours cannot fuse into a match. The raw
// pass covers a payload that does not decode: an undecodable payload is still scanned, never
// waved through.
//
// The detector holds the rules whose keywords gitleaks' prefilter finds in either text, on any
// decoding pass, and every rule without keywords: the rules the full configuration's detector
// would run on these texts. The rest it would skip, so they are not compiled.
func ScanPayload(raw []byte) ([]string, error) {
	return scanPayload(raw, func(texts []string) (*detect.Detector, error) {
		return detector(table.prefiltered(texts...))
	})
}

// scanPayload is ScanPayload with the detector built by build, for the texts it will scan.
func scanPayload(raw []byte, build func(texts []string) (*detect.Detector, error)) ([]string, error) {
	texts := []string{string(raw)}
	var doc any
	if json.Unmarshal(raw, &doc) == nil {
		texts = append(texts, strings.Join(strs(doc, nil), "\n"))
	}
	d, err := build(texts)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, text := range texts {
		ids = append(ids, scan(d, text)...)
	}
	sort.Strings(ids)
	out := ids[:0]
	for i, id := range ids {
		if i == 0 || id != ids[i-1] {
			out = append(out, id)
		}
	}
	return out, nil
}

// strs walks decoded JSON and collects every string it holds, keys included, in a fixed order.
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
		sort.Strings(keys)
		for _, k := range keys {
			acc = append(acc, k)
			acc = strs(t[k], acc)
		}
	}
	return acc
}
