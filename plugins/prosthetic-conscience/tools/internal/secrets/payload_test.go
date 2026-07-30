package secrets

import (
	"encoding/json"
	"strings"
	"testing"
)

// Every literal here is assembled from fragments, so this file never contains a
// scannable secret itself. That is not fastidiousness: sc-secrets-gate blocks
// any Bash payload carrying one, including the payload that would write a test
// OF the gate. Discovered by having exactly that happen, twice.
var (
	awsKey  = "AK" + "IA" + "IOSFODNN7EXAMPLE"
	dashes  = strings.Repeat("-", 5)
	pkBegin = dashes + "BEGIN RSA PRIVATE KEY" + dashes
	// The same header with its "R" written as a JSON \u escape (0x52 == 'R').
	// Decoded, this is byte-for-byte pkBegin — which the test asserts, so the
	// fixture cannot quietly stop exercising the escape it exists for.
	pkEscaped = dashes + `BEGIN ` + "\\" + `u0052SA PRIVATE KEY` + dashes
	// Likewise the AWS key with its leading "A" as A. Writing a literal "A"
	// here would silently make the "escaped" case a second copy of the literal
	// one — which is exactly what happened first time, and was only caught by
	// reverting the fix and noticing that this case still passed.
	awsEscaped = "\\" + `u0041` + awsKey[1:]
)

// mustDecodeTo fails the test unless raw is valid JSON whose "command" decodes
// to want. Every escaped fixture goes through this, so a fixture cannot quietly
// degrade into testing the literal it was meant to contrast with.
func mustDecodeTo(t *testing.T, raw, want string) {
	t.Helper()
	var d struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("fixture is not valid JSON: %v (%s)", err, raw)
	}
	if d.Command != want {
		t.Fatalf("fixture decodes to %q, want %q", d.Command, want)
	}
	if strings.Contains(raw, want) {
		t.Fatalf("fixture contains the target literally, so it is not testing an escape: %s", raw)
	}
}

// THE defect (#80 §0), measured against the shipped binary before this fix:
// an identical decoded secret evaded the gate purely by being \u-escaped on the
// wire. Every pattern is ASCII-alphanumeric, which a standard encoder emits
// literally — so the common case matched, and the gate looked correct.
//
//	{"command":"echo AKIA…"}        -> DENY
//	{"command":"echo AKIA…"}   -> ALLOW   <- same decoded string
func TestScanPayloadCatchesEscapedSecrets(t *testing.T) {
	cases := []struct{ name, raw string }{
		{"literal", `{"command":"echo ` + awsKey + `"}`},
		{"first letter escaped", `{"command":"echo ` + awsEscaped + `"}`},
		{"nested one level down", `{"outer":{"command":"` + awsKey + `"}}`},
		{"inside an array", `{"args":["safe","` + awsKey + `"]}`},
		{"as an object KEY", `{"` + awsKey + `":"value"}`},
		{"private key header, escaped", `{"command":"` + pkEscaped + `"}`},
	}
	// Both escaped fixtures must differ on the wire and agree once decoded, or
	// the cases above prove nothing. Checked here rather than trusted, because
	// the first draft of this file got it wrong and only reverting the fix
	// exposed that the "escaped" AWS case was a second copy of the literal one.
	mustDecodeTo(t, `{"command":"`+pkEscaped+`"}`, pkBegin)
	mustDecodeTo(t, `{"command":"`+awsEscaped+`"}`, awsKey)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ScanPayload([]byte(c.raw)); len(got) == 0 {
				t.Errorf("allowed a payload carrying a secret: %s", c.raw)
			}
		})
	}
}

// Malformed JSON must not become a bypass. "We could not parse it" is not "it is
// safe", so an unparseable payload still gets the raw scan.
func TestScanPayloadDoesNotAllowOnUnparseableInput(t *testing.T) {
	if got := ScanPayload([]byte(`{"command":"` + awsKey + `"`)); len(got) == 0 {
		t.Error("truncated JSON carrying a secret was allowed")
	}
	if got := ScanPayload([]byte("not json at all " + pkBegin)); len(got) == 0 {
		t.Error("non-JSON carrying a secret was allowed")
	}
}

// "Precision beats recall: a false block on legitimate work is a bug" is the
// package's own stated contract, so it needs a test. Decoding must not start
// blocking ordinary work.
func TestScanPayloadDoesNotBlockInnocentPayloads(t *testing.T) {
	for _, raw := range []string{
		`{"command":"go test ./..."}`,
		`{"command":"git commit -m \"AKIA is a prefix, not a key\""}`,
		`{"file_path":"/w/secrets.go","content":"package secrets"}`,
		`{"command":"echo ghp_tooshort"}`,
		`{}`,
		`null`,
		`[]`,
	} {
		if got := ScanPayload([]byte(raw)); len(got) != 0 {
			t.Errorf("false block on %s -> %v", raw, got)
		}
	}
}

// Decoded values are joined by a newline, not concatenated, so two innocent
// neighbours cannot fuse into something satisfying a \b-anchored pattern.
func TestAdjacentValuesDoNotFuseIntoAMatch(t *testing.T) {
	raw := `{"a":"AK","b":"IA` + awsKey[4:] + `"}`
	if got := ScanPayload([]byte(raw)); len(got) != 0 {
		t.Errorf("two harmless values fused into a match: %v", got)
	}
}

// The same payload must scan identically every time. Map iteration in Go is
// randomised, so walking object keys unsorted would make this flaky rather than
// wrong — the worst kind of security bug to chase.
func TestScanPayloadIsDeterministic(t *testing.T) {
	raw := []byte(`{"z":"x","m":"y","a":"` + awsKey + `","k":{"n":"q"}}`)
	first := strings.Join(ScanPayload(raw), ",")
	for range 50 {
		if got := strings.Join(ScanPayload(raw), ","); got != first {
			t.Fatalf("unstable result: %q then %q", first, got)
		}
	}
}

// EVERY declared class needs a positive case, or a regex could be deleted and
// the suite would stay green. #80 measured gh[osur]_ and sk-proj- at zero.
func TestEveryPatternClassHasAPositiveCase(t *testing.T) {
	samples := map[string]string{
		"AWS access key id":            awsKey,
		"GitHub personal access token": "ghp_" + strings.Repeat("a", 36),
		"GitHub fine-grained token":    "github_pat_" + strings.Repeat("b", 22),
		"GitHub app/oauth token":       "ghs_" + strings.Repeat("c", 36),
		"Slack token":                  "xoxb-" + strings.Repeat("1", 12),
		"private key block":            pkBegin,
		"Anthropic api key":            "sk-ant-" + strings.Repeat("d", 24),
		"OpenAI api key":               "sk-proj-" + strings.Repeat("e", 24),
	}
	for _, p := range Patterns {
		sample, ok := samples[p.Class]
		if !ok {
			t.Errorf("class %q has no sample — it could be deleted and this suite would stay green", p.Class)
			continue
		}
		got := Scan(sample)
		// The CLASS must be named, not merely "something matched": the class is
		// what reaches the operator in the deny reason.
		var named bool
		for _, c := range got {
			if c == p.Class {
				named = true
			}
		}
		if !named {
			t.Errorf("sample for %q matched %v instead of naming its own class", p.Class, got)
		}
	}
}

// Scope, pinned rather than assumed: the gate reads tool_input and nothing else.
// #80 notes a secret in a sibling field is unseen. That is a deliberate boundary
// — tool_input is what the tool receives — and it should fail loudly if someone
// widens or narrows it without deciding to.
func TestScopeIsToolInputOnly(t *testing.T) {
	// A secret OUTSIDE tool_input is not this function's business; ScanPayload is
	// handed tool_input alone by its caller.
	if got := ScanPayload([]byte(`{"command":"safe"}`)); len(got) != 0 {
		t.Errorf("unexpected match: %v", got)
	}
}
