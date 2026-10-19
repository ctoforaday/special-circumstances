package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"testing"

	st "github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secrets/secretstest"
)

// ownTag marks every rule gitleaks.toml adds to gitleaks' default configuration.
const ownTag = "special-circumstances"

// gapFormats are, per rule gitleaks.toml adds, the real key formats the default configuration
// misses — the reason the rule exists. Each sample must be refused by its rule and by NO other:
// a sample another rule also refuses is not evidence the added rule is needed, and a sample no
// rule refuses is the gap reopened. So deleting a rule from gitleaks.toml fails here, and a
// default configuration that grows to cover the format fails here too, which is the signal to
// delete the rule.
var gapFormats = map[string][]st.Format{
	"anthropic-key-any-type": {
		{Name: "oauth token, a type the default has no rule for", Draw: func(r *rand.Rand) string {
			return "sk-" + "ant-oat01-" + st.S(r, st.Base64URL, 93) + "AA"
		}},
		{Name: "api key at a body length the default does not take", Draw: func(r *rand.Rand) string {
			return "sk-" + "ant-api03-" + st.S(r, st.Base64URL, 20+r.Intn(60))
		}},
	},
	"openai-key-any-length": {
		{Name: "project key ending in a dash", Draw: func(r *rand.Rand) string {
			return "sk-" + "proj-" + st.S(r, st.Base64URL, 74) + "T3Blbk" + "FJ" + st.S(r, st.Base64URL, 73) + "-"
		}},
		{Name: "service-account key without the marker at a fixed offset", Draw: func(r *rand.Rand) string {
			return "sk-" + "svcacct-" + st.S(r, st.Base64URL, 40+r.Intn(120))
		}},
	},
	"private-key-armor-body": {
		{Name: "openssh key without its END line", Draw: func(r *rand.Rand) string {
			return st.Dashes + "BEGIN OPENSSH PRIVATE KEY" + st.Dashes + "\n" + st.Lines(r, 2+r.Intn(3))
		}},
		{Name: "encrypted pem key with armor headers, without its END line", Draw: func(r *rand.Rand) string {
			return st.Dashes + "BEGIN RSA PRIVATE KEY" + st.Dashes + "\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-128-CBC," + st.S(r, "0123456789ABCDEF", 32) + "\n\n" + st.Lines(r, 2)
		}},
		{Name: "ssh2 key", Draw: func(r *rand.Rand) string {
			return "---- BEGIN SSH2 ENCRYPTED PRIVATE KEY ----\nComment: \"" + st.S(r, st.Lower, 8) + "\"\n" + st.Lines(r, 3) + "\n---- END SSH2 ENCRYPTED PRIVATE KEY ----"
		}},
	},
	"putty-private-key": {
		{Name: "ppk file", Draw: func(r *rand.Rand) string {
			return "PuTTY-User-Key-File-3: ssh-ed25519\nEncryption: none\nComment: " + st.S(r, st.Lower, 8) +
				"\nPublic-Lines: 2\n" + st.S(r, st.Base64, 64) + "\n" + st.S(r, st.Base64, 4) +
				"\nPrivate-Lines: 1\n" + st.S(r, st.Base64, 44) + "\nPrivate-MAC: " + st.S(r, st.HexLower, 64)
		}},
	},
}

// samples is how many keys each format draws.
const samples = 200

// contexts are the places a key sits in a payload.
var contexts = []func(string) string{
	func(k string) string { return k },
	func(k string) string { return "export KEY=" + k },
	func(k string) string { return "echo '" + k + "' > key.txt" },
	func(k string) string { return `{"api_key": "` + k + `"}` },
}

// curl puts a key where gitleaks' curl-auth-header rule also refuses it, so it is a context
// for vendor keys and not for gap samples, which must be refused by their own rule alone.
func curl(k string) string {
	return "curl -H \"Authorization: Bearer " + k + "\" https://api.example.com"
}

func mustScan(t *testing.T, command string) []string {
	t.Helper()
	raw, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatal(err)
	}
	ids, err := ScanPayload(raw)
	if err != nil {
		t.Fatalf("the detector did not build: %v", err)
	}
	return ids
}

func ownRules(t *testing.T) []string {
	t.Helper()
	d, err := Detector()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for id, r := range d.Config.Rules {
		if slices.Contains(r.Tags, ownTag) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

// Every rule gitleaks.toml adds names its gap here, and nothing here names a rule it does not
// add. The tag is read from the built detector, so a rule that failed to load is missing.
func TestEveryAddedRuleHasGapSamples(t *testing.T) {
	got := ownRules(t)
	var want []string
	for id := range gapFormats {
		want = append(want, id)
	}
	sort.Strings(want)
	if !slices.Equal(got, want) {
		t.Fatalf("rules added to the default configuration %v; rules with gap samples %v", got, want)
	}
}

// The default configuration is loaded too: the extension adds to it rather than replacing it.
func TestTheDefaultConfigurationIsExtendedNotReplaced(t *testing.T) {
	d, err := Detector()
	if err != nil {
		t.Fatal(err)
	}
	if n := len(d.Config.Rules) - len(gapFormats); n < 200 {
		t.Fatalf("only %d rules besides the gate's own: gitleaks' default configuration did not load", n)
	}
	for _, id := range []string{"aws-access-token", "generic-api-key", "private-key"} {
		if _, ok := d.Config.Rules[id]; !ok {
			t.Errorf("default rule %s is missing", id)
		}
	}
}

func TestGapSamplesAreRefusedByTheirRuleAlone(t *testing.T) {
	for rule, formats := range gapFormats {
		for _, f := range formats {
			t.Run(rule+"/"+f.Name, func(t *testing.T) {
				r := st.Rand(rule + "/" + f.Name)
				for i := range samples {
					key := f.Draw(r)
					ctx := contexts[i%len(contexts)]
					got := mustScan(t, ctx(key))
					if !slices.Equal(got, []string{rule}) {
						t.Fatalf("sample %d in context %d: refused by %v, want [%s] alone", i, i%len(contexts), got, rule)
					}
				}
			})
		}
	}
}

// A sample per vendor through gitleaks' own rules: the default configuration is what refuses
// them. A vendor key is reported under its gitleaks rule id.
func TestVendorKeysAreRefusedByTheirGitleaksRule(t *testing.T) {
	for _, f := range st.Vendors {
		t.Run(f.Name, func(t *testing.T) {
			r := st.Rand(f.Name)
			for i := range samples {
				ctx := append(contexts[:len(contexts):len(contexts)], curl)[i%(len(contexts)+1)]
				got := mustScan(t, ctx(f.Draw(r)))
				if !slices.Contains(got, f.Rule) {
					t.Fatalf("sample %d: refused by %v, want %s among them", i, got, f.Rule)
				}
			}
		})
	}
}

// What gitleaks' allowlists and entropy thresholds pass, the gate passes: documentation
// keys, placeholders, an armor header with no key under it.
func TestDocumentationKeysAndPlaceholdersAreNotRefused(t *testing.T) {
	for _, s := range []string{
		"AK" + "IA" + "IOSFODNN7" + "EXAMPLE",
		"gl" + "pat-" + strings.Repeat("x", 20),
		"sk" + "_live_" + strings.Repeat("0", 24),
		"ghp" + "_" + strings.Repeat("x", 36),
		"sk-" + "ant-api03-" + strings.Repeat("x", 40),
		"sk-" + "proj-" + strings.Repeat("X", 40),
		st.Dashes + "BEGIN RSA PRIVATE KEY" + st.Dashes,
		st.Dashes + "BEGIN RSA PRIVATE KEY" + st.Dashes + "\n...\n" + st.Dashes + "END RSA PRIVATE KEY" + st.Dashes,
		st.Dashes + "BEGIN OPENSSH PRIVATE KEY" + st.Dashes + "\n" + strings.Repeat("A", 140),
		"Private-Lines: 1\n" + strings.Repeat("A", 44),
		"git status && go test ./...",
		"how do github tokens work",
	} {
		if got := mustScan(t, "echo "+s); len(got) != 0 {
			t.Errorf("refused %q: %v", s, got)
		}
	}
}

// An added rule carries an entropy threshold, as gitleaks' own vendor rules do: its format
// filled with one repeated character is a placeholder, and passes.
func TestAddedRulesPassALowEntropyBody(t *testing.T) {
	for _, s := range []string{
		"sk-" + "ant-oat01-" + strings.Repeat("a", 95),
		"sk-" + "svcacct-" + strings.Repeat("_", 60),
		"---- BEGIN SSH2 ENCRYPTED PRIVATE KEY ----\n" + strings.Repeat("A", 64) + "\n" + strings.Repeat("A", 64),
	} {
		if got := mustScan(t, s); len(got) != 0 {
			t.Errorf("refused a body of one repeated character %q: %v", s, got)
		}
	}
}

// A `gitleaks:allow` comment is the payload's author exempting the payload. The gate ignores it.
func TestAnAllowCommentDoesNotExemptThePayload(t *testing.T) {
	key := st.Key("aws access key id")
	if got := mustScan(t, "echo "+key+" # gitleaks"+":allow"); !slices.Contains(got, "aws-access-token") {
		t.Fatalf("a gitleaks:allow comment exempted a key: %v", got)
	}
}

// gitleaks decodes base64 up to the command line's default depth, so an encoded key is found.
func TestAnEncodedKeyIsFound(t *testing.T) {
	key := st.Key("aws access key id")
	raw, _ := json.Marshal(map[string]string{"command": "echo " + base64.StdEncoding.EncodeToString([]byte(key))})
	got, err := ScanPayload(raw)
	if err != nil || !slices.Contains(got, "aws-access-token") {
		t.Fatalf("a base64-encoded key was not found: %v %v", got, err)
	}
}

// No finding the detector returns holds the value it matched: Secret, Match and Line are
// redacted before ScanPayload reads the rule id.
func TestFindingsCarryNoValue(t *testing.T) {
	d, err := Detector()
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, f := range st.Vendors {
		keys = append(keys, f.Draw(st.Rand(f.Name)))
	}
	for _, fs := range gapFormats {
		for _, f := range fs {
			keys = append(keys, f.Draw(st.Rand(f.Name)))
		}
	}
	for _, key := range keys {
		if len(windows(key)) == 0 {
			t.Fatalf("a sample has no drawn material to look for")
		}
		findings := d.DetectString("echo " + key + "\n")
		if len(findings) == 0 {
			t.Fatalf("no finding for a sample")
		}
		for _, f := range findings {
			for _, field := range []string{f.Secret, f.Match, f.Line} {
				for _, w := range windows(key) {
					if strings.Contains(field, w) {
						t.Fatalf("finding %s carries drawn key material", f.RuleID)
					}
				}
			}
		}
	}
}

// windows are 16-byte spans of every run of 20 or more key characters in s: the drawn
// material, without the vendor prefixes, armor words and armor header lines (`DEK-Info: …`,
// which carries a public IV) a finding may keep.
func windows(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, ": ") {
			continue
		}
		run := 0
		for i := 0; i <= len(line); i++ {
			if i < len(line) && strings.IndexByte(st.Base64URL+"+/", line[i]) >= 0 {
				run++
				continue
			}
			for j := i - run; run >= 20 && j+16 <= i; j += 8 {
				w := line[j : j+16]
				if strings.ContainsAny(w, st.Digit) || strings.ContainsAny(w, st.Lower) && strings.ContainsAny(w, st.Upper) {
					out = append(out, w)
				}
			}
			run = 0
		}
	}
	return out
}

// The sizes are the 99th percentile (6 KiB) and the maximum (56 KiB) of the Bash, WebFetch,
// WebSearch and MCP inputs in this box's transcripts on 2026-10-18, and 1 MiB.
func BenchmarkScanPayload(b *testing.B) {
	if _, err := Detector(); err != nil {
		b.Fatal(err)
	}
	line := `git -C /home/u/p log --oneline -n 20 && go test -count=1 ./internal/... # a token, a key, a secret\n`
	for _, size := range []int{6 << 10, 56 << 10, 1 << 20} {
		b.Run(fmt.Sprintf("%dKiB", size>>10), func(b *testing.B) {
			var body strings.Builder
			for body.Len() < size {
				body.WriteString(line)
			}
			payload := []byte(`{"command":"` + body.String() + `"}`)
			b.SetBytes(int64(len(payload)))
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if got, _ := ScanPayload(payload); len(got) != 0 {
					b.Fatalf("payload with no secret matched %v", got)
				}
			}
		})
	}
}
