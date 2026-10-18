package secrets

import (
	"math/rand"
	"strings"
	"testing"
)

// WHY THESE FIXTURES ARE GENERATED, AND KEYED BY CLASS.
//
// Keyed by class: statement coverage cannot see a dead row, because Scan loops over Patterns
// and so every row "executes". A row nobody's fixture names was deletable with the suite green
// (gh[osur]_ and sk-proj- both were). TestEveryPatternHasAFormat makes a row without a format
// fail here instead.
//
// Generated: a hand-written fixture is one sample, and one sample drawn from a convenient
// subset of the alphabet hides a class that is narrower than the vendor's. An Anthropic class
// without `_`, tested by one alphanumeric fixture, misses 56% of real base64url keys with the
// suite green. So each format is written down as the vendor publishes it — literal parts and
// random parts over the WHOLE alphabet — and many seeded samples are drawn from it. Narrow a
// class below its vendor's alphabet and samples fail.
//
// Every generated value is random filler in a real shape; none is a live credential.

// Alphabets, as the vendors publish them.
const (
	upper     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower     = "abcdefghijklmnopqrstuvwxyz"
	digit     = "0123456789"
	base62    = upper + lower + digit
	base64url = base62 + "_-"
	base64std = base62 + "+/"
	base32    = upper + "234567"
	hexLower  = "0123456789abcdef"
	upperNum  = upper + digit
)

// part is one span of a format: a literal, or between min and max random characters drawn
// from alpha.
type part struct {
	lit      string
	alpha    string
	min, max int
}

func lit(s string) part                   { return part{lit: s} }
func rnd(alpha string, min, max int) part { return part{alpha: alpha, min: min, max: max} }
func fixed(alpha string, n int) part      { return part{alpha: alpha, min: n, max: n} }

type format struct {
	name  string
	parts []part
}

func (f format) draw(r *rand.Rand) string {
	var b strings.Builder
	for _, p := range f.parts {
		if p.alpha == "" {
			b.WriteString(p.lit)
			continue
		}
		n := p.min + r.Intn(p.max-p.min+1)
		for i := 0; i < n; i++ {
			b.WriteByte(p.alpha[r.Intn(len(p.alpha))])
		}
	}
	return b.String()
}

// armor is a private-key header followed by key material; the header is what the row
// matches, and the body shows the match does not depend on what follows it.
func armor(header string) []part {
	return []part{lit(header + "\n"), fixed(base64std, 64), lit("\n")}
}

// formats holds every published shape each row must catch.
var formats = map[string][]format{
	"AWS access key id": {
		{"AKIA long-term", []part{lit("AKIA"), fixed(base32, 16)}},
		{"ASIA STS temporary", []part{lit("ASIA"), fixed(base32, 16)}},
		{"ABIA STS service bearer", []part{lit("ABIA"), fixed(base32, 16)}},
		{"ACCA context-specific", []part{lit("ACCA"), fixed(base32, 16)}},
	},
	"GitHub personal access token": {
		{"ghp_", []part{lit("ghp_"), fixed(base62, 36)}},
	},
	"GitHub fine-grained token": {
		{"github_pat_", []part{lit("github_pat_"), fixed(base62, 22), lit("_"), fixed(base62, 59)}},
	},
	"GitHub app/oauth token": {
		{"gho_", []part{lit("gho_"), fixed(base62, 36)}},
		{"ghu_", []part{lit("ghu_"), fixed(base62, 36)}},
		{"ghs_", []part{lit("ghs_"), fixed(base62, 36)}},
		{"ghr_", []part{lit("ghr_"), fixed(base62, 36)}},
	},
	"Slack token": {
		{"xoxb bot", []part{lit("xoxb-"), rnd(digit, 10, 13), lit("-"), rnd(digit, 10, 13), lit("-"), fixed(base62, 24)}},
		{"xoxp user", []part{lit("xoxp-"), rnd(digit, 10, 13), lit("-"), rnd(digit, 10, 13), lit("-"), rnd(digit, 10, 13), lit("-"), fixed(hexLower, 32)}},
		{"xoxe.xoxp config", []part{lit("xoxe.xoxp-1-"), rnd(upperNum, 163, 166)}},
		{"xoxe rotating refresh", []part{lit("xoxe-1-"), fixed(upperNum, 146)}},
		{"xapp app-level", []part{lit("xapp-1-A"), rnd(upperNum, 9, 10), lit("-"), fixed(digit, 13), lit("-"), fixed(hexLower, 64)}},
		{"xoxs legacy", []part{lit("xoxs-"), rnd(digit, 1, 12), lit("-"), rnd(digit, 1, 12), lit("-"), rnd(digit, 1, 12), lit("-"), fixed(hexLower, 64)}},
		{"xoxo legacy", []part{lit("xoxo-"), rnd(digit, 1, 12), lit("-"), rnd(digit, 1, 12), lit("-"), rnd(digit, 1, 12), lit("-"), fixed(hexLower, 64)}},
	},
	"Slack webhook url": {
		{"services", []part{lit("https://hooks.slack.com/services/T"), rnd(upperNum, 8, 10), lit("/B"), rnd(upperNum, 8, 10), lit("/"), fixed(base62, 24)}},
		{"workflows", []part{lit("https://hooks.slack.com/workflows/T"), rnd(upperNum, 8, 10), lit("/A"), rnd(upperNum, 8, 10), lit("/"), fixed(digit, 18), lit("/"), fixed(base62, 24)}},
		{"triggers", []part{lit("https://hooks.slack.com/triggers/T"), rnd(upperNum, 8, 10), lit("/"), fixed(digit, 13), lit("/"), fixed(hexLower, 32)}},
	},
	"private key block": {
		{"PKCS#1 RSA", armor("-----BEGIN RSA PRIVATE KEY-----")},
		{"PKCS#8", armor("-----BEGIN PRIVATE KEY-----")},
		{"PKCS#8 encrypted", armor("-----BEGIN ENCRYPTED PRIVATE KEY-----")},
		{"SEC1 EC", armor("-----BEGIN EC PRIVATE KEY-----")},
		{"DSA", armor("-----BEGIN DSA PRIVATE KEY-----")},
		{"OpenSSH", armor("-----BEGIN OPENSSH PRIVATE KEY-----")},
		{"PGP", armor("-----BEGIN PGP PRIVATE KEY BLOCK-----")},
		{"SSH2", armor("---- BEGIN SSH2 ENCRYPTED PRIVATE KEY ----")},
	},
	"PuTTY private key file": {
		{"ppk v2", []part{lit("PuTTY-User-Key-File-2: ssh-rsa\nEncryption: none\nPublic-Lines: 1\n"), fixed(base64std, 64)}},
		{"ppk v3", []part{lit("PuTTY-User-Key-File-3: ssh-ed25519\nEncryption: none\nPublic-Lines: 2\n"), fixed(base64std, 64)}},
	},
	"Anthropic api key": {
		{"api03", []part{lit("sk-ant-api03-"), fixed(base64url, 93), lit("AA")}},
		{"admin01", []part{lit("sk-ant-admin01-"), fixed(base64url, 93), lit("AA")}},
		{"oat01", []part{lit("sk-ant-oat01-"), fixed(base64url, 93), lit("AA")}},
	},
	"OpenAI api key": {
		{"sk-proj", []part{lit("sk-proj-"), fixed(base64url, 74), lit("T3BlbkFJ"), fixed(base64url, 74)}},
		{"sk-proj short", []part{lit("sk-proj-"), fixed(base64url, 58), lit("T3BlbkFJ"), fixed(base64url, 58)}},
		{"sk-svcacct", []part{lit("sk-svcacct-"), fixed(base64url, 74), lit("T3BlbkFJ"), fixed(base64url, 74)}},
		{"sk-admin", []part{lit("sk-admin-"), fixed(base64url, 58), lit("T3BlbkFJ"), fixed(base64url, 58)}},
	},
	"OpenAI legacy api key": {
		{"sk- legacy", []part{lit("sk-"), fixed(base62, 20), lit("T3BlbkFJ"), fixed(base62, 20)}},
	},
}

// nearMisses MUST match nothing at all: precision beats recall, and a false block on
// legitimate work is a bug.
var nearMisses = map[string][]string{
	"AWS access key id":            {"AKIA" + strings.Repeat("Q", 15), "ASIAN", "AKIA" + strings.Repeat("Q", 17)},
	"GitHub personal access token": {"ghp_" + strings.Repeat("a", 35), "ghp_tooshort"},
	"GitHub fine-grained token":    {"github_pat_" + strings.Repeat("a", 21)},
	"GitHub app/oauth token":       {"ghx_" + strings.Repeat("a", 36)},
	"Slack token":                  {"xoxz-1234567890-abcdefghijk", "xapp-1-short"},
	"Slack webhook url":            {"https://hooks.slack.com/services/", "https://api.slack.com/apps/A0123456789/general"},
	"private key block":            {"-----BEGIN CERTIFICATE-----", "-----BEGIN PUBLIC KEY-----", "-----BEGIN PGP PUBLIC KEY BLOCK-----"},
	"PuTTY private key file":       {"PuTTY-User-Key-File: no version", "the PuTTY key file format"},
	"Anthropic api key":            {"sk-ant-", "sk-ant-api03-short"},
	"OpenAI api key":               {"sk-proj-tooshort", "sk-admin-tooshort"},
	"OpenAI legacy api key":        {"sk-" + strings.Repeat("a", 48)},
}

// contexts wrap a sample the way it reaches the gate: in a command, a header, quoted, at the
// edges of the string. Each keeps a non-word character on both sides, as real text does.
var contexts = []func(string) string{
	func(s string) string { return s },
	func(s string) string { return "export KEY=" + s + "\n" },
	func(s string) string { return `curl -H "Authorization: Bearer ` + s + `" https://x` },
	func(s string) string { return "'" + s + "'" },
	func(s string) string { return "(" + s + ");" },
	func(s string) string { return "see " + s + ". next" },
}

const samplesPerFormat = 2000

func TestEveryPatternHasAFormat(t *testing.T) {
	for _, p := range Patterns {
		if len(formats[p.Class]) == 0 {
			t.Errorf("pattern %q has no format — write down its published shape in formats, or the row ships unverified", p.Class)
		}
		if len(nearMisses[p.Class]) == 0 {
			t.Errorf("pattern %q has no near-miss — a row with no negative can widen to anything with the suite green", p.Class)
		}
	}
	for class := range formats {
		if !hasClass(class) {
			t.Errorf("formats names %q, which no pattern in Patterns carries — a renamed class leaves its formats testing nothing", class)
		}
	}
	for class := range nearMisses {
		if !hasClass(class) {
			t.Errorf("nearMisses names %q, which no pattern in Patterns carries", class)
		}
	}
}

func hasClass(class string) bool {
	for _, p := range Patterns {
		if p.Class == class {
			return true
		}
	}
	return false
}

// Every sample of every format must be caught by its own row and by no other. The class is
// asserted, not "something matched": otherwise one row's samples could carry another row.
func TestGeneratedKeysMatchTheirRow(t *testing.T) {
	for class, fs := range formats {
		for fi, f := range fs {
			t.Run(class+"/"+f.name, func(t *testing.T) {
				r := rand.New(rand.NewSource(int64(len(class)*1000 + fi)))
				missed := 0
				var example string
				for i := 0; i < samplesPerFormat; i++ {
					s := f.draw(r)
					got := Scan(contexts[i%len(contexts)](s))
					if len(got) != 1 || got[0] != class {
						if missed == 0 {
							example = s
						}
						missed++
					}
				}
				if missed > 0 {
					t.Errorf("%d of %d samples were not caught as exactly [%s]; first: %q", missed, samplesPerFormat, class, example)
				}
			})
		}
	}
}

// The generator must reach every character of every alphabet it draws from, or a narrowed
// fixture would let a narrowed class through.
func TestSamplesSpanTheirAlphabets(t *testing.T) {
	for class, fs := range formats {
		for _, f := range fs {
			for pi, p := range f.parts {
				if p.alpha == "" {
					continue
				}
				r := rand.New(rand.NewSource(int64(pi)))
				seen := map[byte]bool{}
				for i := 0; i < samplesPerFormat; i++ {
					for _, c := range []byte(format{parts: []part{p}}.draw(r)) {
						seen[c] = true
					}
				}
				for _, c := range []byte(p.alpha) {
					if !seen[c] {
						t.Errorf("%s/%s part %d never drew %q", class, f.name, pi, c)
					}
				}
			}
		}
	}
}

func TestNearMissesMatchNothing(t *testing.T) {
	for class, ms := range nearMisses {
		for _, m := range ms {
			if hit := Scan("clean: " + m + " end"); len(hit) != 0 {
				t.Errorf("%s near-miss %q matched %v — a false block on legitimate work is a bug", class, m, hit)
			}
		}
	}
}

// The gate's real input is a JSON payload, and its own history is a fix for scanning the
// wire bytes instead of the decoded values. Every format is therefore also proven through
// ScanPayload in the escaped form that once bypassed it.
func TestEveryFormatSurvivesJSONEscaping(t *testing.T) {
	for class, fs := range formats {
		for _, f := range fs {
			t.Run(class+"/"+f.name, func(t *testing.T) {
				s := f.draw(rand.New(rand.NewSource(1)))
				var esc strings.Builder
				for _, r := range s {
					esc.WriteString(escapeRune(r))
				}
				got := ScanPayload([]byte(`{"command":"echo ` + esc.String() + `"}`))
				for _, c := range got {
					if c == class {
						return
					}
				}
				t.Errorf("ScanPayload(%s, \\uXXXX-escaped) = %v; want it to include %q — the sender's escaping choice must not decide the verdict", f.name, got, class)
			})
		}
	}
}

// escapeRune writes an ASCII rune as \uXXXX, which is what a JSON encoder is permitted
// to do to any character and what a pattern written against decoded text cannot see.
func escapeRune(r rune) string {
	const hex = "0123456789abcdef"
	return string([]byte{
		'\\', 'u', '0', '0',
		hex[(byte(r)>>4)&0xf], hex[byte(r)&0xf],
	})
}

// BenchmarkScanPayload measures the gate on 1 MiB payloads with no secret in them, the worst
// case for a row, which then runs to the end of the input. "ordinary" holds none of the
// rows' literals, as most payloads do; "prefix-dense" puts a literal on every line, so the
// prefilter passes and the regexes run.
func BenchmarkScanPayload(b *testing.B) {
	for _, c := range []struct{ name, line string }{
		{"ordinary", `git -C /home/u/p log --oneline -n 20 && go test -count=1 ./internal/... # a token, a key, a secret\n`},
		{"prefix-dense", `git log -n 20 # AKIA ASIA sk-ant- ghp_ xox PRIVATE KEY hooks.slack.com/ T3BlbkFJ sk-proj-\n`},
	} {
		b.Run(c.name, func(b *testing.B) {
			var body strings.Builder
			for body.Len() < 1<<20 {
				body.WriteString(c.line)
			}
			payload := []byte(`{"command":"` + body.String() + `"}`)
			b.SetBytes(int64(len(payload)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got := ScanPayload(payload); len(got) != 0 {
					b.Fatalf("payload with no secret matched %v", got)
				}
			}
		})
	}
}
