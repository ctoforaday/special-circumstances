package feov

import (
	"encoding/json"
	"strings"
	"testing"
)

// THE ENVELOPE IS RECOGNISED BY WHAT IT IS, not by how it happens to begin. The failure hook used
// to match the bytes `{"ok":false`, which recognised any tool's failure object as the record
// tool's own refusal and did not recognise the record tool's handler refusals, whose first key
// was another.
func TestOnlyTheRefusalEnvelopeDecodesAsOne(t *testing.T) {
	own, err := json.Marshal(RefusalEnvelope{Verb: "close", Role: "lens", Code: string(Validation), Error: "unknown flag: --key"})
	if err != nil {
		t.Fatal(err)
	}
	if env, ok := DecodeRefusal(string(own)); !ok || env.Verb != "close" || env.Code != string(Validation) {
		t.Fatalf("the envelope as it is written is not recognised: %s", own)
	}
	if got := `{"` + strings.Join(RefusalKeys(), `":…,"`) + `":…}`; got != `{"verb":…,"role":…,"ok":…,"code":…,"error":…}` {
		t.Errorf("the envelope's keys, in the order written, are %s", got)
	}
	for name, line := range map[string]string{
		"another tool's failure object":       `{"ok":false}`,
		"an object with three of the keys":    `{"ok":false,"code":"validation","error":"x"}`,
		"the keys and one more":               `{"verb":"close","role":"lens","ok":false,"code":"validation","error":"x","help":"…"}`,
		"a success":                           `{"verb":"close","role":"lens","ok":true,"code":"","error":""}`,
		"a key of the wrong type":             `{"verb":"close","role":"lens","ok":"false","code":"validation","error":"x"}`,
		"the tool's human line":               `feov-record: close: unknown flag: --key`,
		"not an object":                       `[]`,
		"a success envelope, with its result": `{"verb":"close","role":"lens","ok":true,"result":{"message":"closed"}}`,
	} {
		if _, ok := DecodeRefusal(line); ok {
			t.Errorf("%s decoded as the record tool's refusal: %s", name, line)
		}
	}
}
