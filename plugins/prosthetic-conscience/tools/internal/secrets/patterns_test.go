package secrets

import "testing"

// WHY THIS TEST IS WRITTEN AGAINST THE TABLE, NOT AGAINST A LIST OF STRINGS.
//
// internal/secrets reported 100.0% of statements while TWO of its eight patterns were
// dead weight. Measured by deleting one pattern line at a time and re-running
// `go test ./internal/secrets/ ./cmd/sc-secrets-gate/`:
//
//	AWS access key id             killed
//	GitHub personal access token  killed
//	GitHub fine-grained token     killed
//	GitHub app/oauth token        SURVIVED   <- deletable, CI green
//	Slack token                   killed
//	private key block             killed
//	Anthropic api key             killed
//	OpenAI api key                SURVIVED   <- deletable, CI green
//
// Nothing anywhere mentioned gho_, ghu_, ghs_, ghr_ or sk-proj. Scan() loops over
// Patterns, so every pattern is "executed" and statement coverage cannot tell the
// difference between a pattern that works and one that is never asked to match. This is
// the sharpest illustration in the repo that statement coverage is not the measurement
// we think it is — and it landed on the one gate whose contract is agent-guardrails:
// never send secrets outward.
//
// So the fixtures are keyed BY CLASS and the test iterates Patterns: a new pattern with
// no fixture fails here rather than shipping unverified. That is the property a list of
// hand-written cases cannot have — the previous test was a list, and it is how two
// patterns went unverified for their whole lives.
//
// Every value below is a syntactically valid, semantically dead example: the AWS key is
// Amazon's own documentation constant, and the rest are structural fillers.
type fixture struct {
	positive string // MUST match this pattern, and ONLY this pattern
	negative string // a near-miss that MUST match nothing at all
}

var fixtures = map[string]fixture{
	"AWS access key id": {
		positive: "AKIAIOSFODNN7EXAMPLE",
		negative: "AKIAIOSFODNN7EXAMPL", // 15 tail chars, not 16
	},
	"GitHub personal access token": {
		positive: "ghp_" + rep36,
		negative: "ghp_" + rep36[:35], // one short
	},
	"GitHub fine-grained token": {
		positive: "github_pat_" + rep36[:22],
		negative: "github_pat_" + rep36[:21], // under the 22 minimum
	},
	"GitHub app/oauth token": {
		positive: "gho_" + rep36,
		negative: "ghx_" + rep36, // x is outside [osur]
	},
	"Slack token": {
		positive: "xoxb-1234567890-abcdefghijk",
		negative: "xoxz-1234567890-abcdefghijk", // z is outside [baprs]
	},
	"private key block": {
		positive: "-----BEGIN RSA PRIVATE KEY-----",
		negative: "-----BEGIN CERTIFICATE-----",
	},
	"Anthropic api key": {
		positive: "sk-ant-api03-" + rep36[:20],
		negative: "sk-ant-",
	},
	"OpenAI api key": {
		positive: "sk-proj-" + rep36[:20],
		negative: "sk-proj-tooshort",
	},
}

// rep36 is EXACTLY 36 token-body characters — the length the GitHub patterns require.
// The count is load-bearing in both directions: the patterns are \b-anchored, so 37 fails
// to match just as 35 does, and the negatives below are built by shortening this.
const rep36 = "0123456789abcdefghijklmnopqrstuvwxyz"

func TestEveryPatternHasAFixture(t *testing.T) {
	for _, p := range Patterns {
		if _, ok := fixtures[p.Class]; !ok {
			t.Errorf("pattern %q has no fixture — add a positive and a negative to fixtures, or the pattern ships unverified (this is exactly how gh[osur]_ and sk-proj- stayed deletable at 100%% statement coverage)", p.Class)
		}
	}
	for class := range fixtures {
		found := false
		for _, p := range Patterns {
			if p.Class == class {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("fixture %q names no pattern in Patterns — a renamed class leaves a fixture testing nothing", class)
		}
	}
}

func TestEveryPatternMatchesItsPositive(t *testing.T) {
	for _, p := range Patterns {
		f, ok := fixtures[p.Class]
		if !ok {
			continue // reported by TestEveryPatternHasAFixture
		}
		t.Run(p.Class, func(t *testing.T) {
			// Asserted through Scan, and on the CLASS rather than on "something
			// matched": the old test only checked len(got) > 0, so any pattern could
			// be carried by another pattern's fixture.
			got := Scan("leaked: " + f.positive)
			if len(got) != 1 || got[0] != p.Class {
				t.Errorf("Scan(%q) = %v; want exactly [%q]", f.positive, got, p.Class)
			}
			if hit := Scan("clean: " + f.negative); len(hit) != 0 {
				t.Errorf("near-miss %q matched %v — precision beats recall here, a false block on legitimate work is a bug", f.negative, hit)
			}
		})
	}
}

// The gate's real input is a JSON payload, and its own history is a fix for scanning the
// wire bytes instead of the decoded values. Every pattern is therefore also proven
// through ScanPayload in the escaped form that once bypassed it.
func TestEveryPatternSurvivesJSONEscaping(t *testing.T) {
	for _, p := range Patterns {
		f, ok := fixtures[p.Class]
		if !ok {
			continue
		}
		t.Run(p.Class, func(t *testing.T) {
			var esc string
			for _, r := range f.positive {
				esc += escapeRune(r)
			}
			payload := []byte(`{"command":"echo ` + esc + `"}`)
			got := ScanPayload(payload)
			if len(got) == 0 {
				t.Fatalf("ScanPayload missed %s in \\uXXXX-escaped form — the sender's escaping choice must not decide the verdict", p.Class)
			}
			found := false
			for _, c := range got {
				if c == p.Class {
					found = true
				}
			}
			if !found {
				t.Errorf("ScanPayload(%s escaped) = %v; want it to include %q", p.Class, got, p.Class)
			}
		})
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
