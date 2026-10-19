package secrets

import (
	"encoding/json"
	"strings"
	"testing"

	st "github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secrets/secretstest"
)

// Every key here is drawn at test time (secretstest), so this file holds no scannable value.
var (
	awsKey = st.Key("aws access key id")
	pkKey  = st.Key("rsa private key")
	// The private key with the "R" of its header written as a JSON \u escape (0x52 == 'R').
	// Decoded, this is byte-for-byte pkKey — which the test asserts, so the fixture cannot
	// quietly stop exercising the escape it exists for.
	pkEscaped = strings.Replace(jsonText(pkKey), "BEGIN RSA", "BEGIN "+"\\"+"u0052SA", 1)
	// Likewise the AWS key with its leading "A" escaped.
	awsEscaped = "\\" + `u0041` + awsKey[1:]
)

// jsonText is s as it sits inside a JSON string literal, without the quotes.
func jsonText(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

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

// An identical decoded secret must not evade the gate by being \u-escaped on the wire: a
// standard encoder writes a key's characters literally, so a gate reading only the wire bytes
// looks correct on the common case.
func TestScanPayloadCatchesEscapedSecrets(t *testing.T) {
	cases := []struct{ name, raw string }{
		{"literal", `{"command":"echo ` + awsKey + `"}`},
		{"first letter escaped", `{"command":"echo ` + awsEscaped + `"}`},
		{"nested one level down", `{"outer":{"command":"` + awsKey + `"}}`},
		{"inside an array", `{"args":["safe","` + awsKey + `"]}`},
		{"as an object KEY", `{"` + awsKey + `":"value"}`},
		{"private key, escaped", `{"command":"` + pkEscaped + `"}`},
	}
	// Both escaped fixtures must differ on the wire and agree once decoded, or
	// the cases above prove nothing.
	mustDecodeTo(t, `{"command":"`+pkEscaped+`"}`, pkKey)
	mustDecodeTo(t, `{"command":"`+awsEscaped+`"}`, awsKey)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, _ := ScanPayload([]byte(c.raw)); len(got) == 0 {
				t.Errorf("allowed a payload carrying a secret: %s", c.raw)
			}
		})
	}
}

// Malformed JSON must not become a bypass. "We could not parse it" is not "it is
// safe", so an unparseable payload still gets the raw scan.
func TestScanPayloadDoesNotAllowOnUnparseableInput(t *testing.T) {
	if got, _ := ScanPayload([]byte(`{"command":"` + awsKey + `"`)); len(got) == 0 {
		t.Error("truncated JSON carrying a secret was allowed")
	}
	if got, _ := ScanPayload([]byte("not json at all " + pkKey)); len(got) == 0 {
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
		if got, _ := ScanPayload([]byte(raw)); len(got) != 0 {
			t.Errorf("false block on %s -> %v", raw, got)
		}
	}
}

// Decoded values are joined by a newline, not concatenated, so two innocent
// neighbours cannot fuse into something satisfying a \b-anchored rule.
func TestAdjacentValuesDoNotFuseIntoAMatch(t *testing.T) {
	raw := `{"a":"AK","b":"IA` + awsKey[4:] + `"}`
	if got, _ := ScanPayload([]byte(raw)); len(got) != 0 {
		t.Errorf("two harmless values fused into a match: %v", got)
	}
}

// The same payload must scan identically every time. Map iteration in Go is
// randomised, so walking object keys unsorted would make this flaky rather than
// wrong — the worst kind of security bug to chase.
func TestScanPayloadIsDeterministic(t *testing.T) {
	raw := []byte(`{"z":"x","m":"y","a":"` + awsKey + `","k":{"n":"q"}}`)
	ids, _ := ScanPayload(raw)
	first := strings.Join(ids, ",")
	for range 50 {
		again, _ := ScanPayload(raw)
		if got := strings.Join(again, ","); got != first {
			t.Fatalf("unstable result: %q then %q", first, got)
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
	if got, _ := ScanPayload([]byte(`{"command":"safe"}`)); len(got) != 0 {
		t.Errorf("unexpected match: %v", got)
	}
}
