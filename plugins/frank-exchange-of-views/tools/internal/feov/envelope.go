package feov

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// RefusalEnvelope is the --json form of every refusal the record tool gives: the same five keys,
// in this order, whatever refused the call — a handler, the record, the flag parser, an identity
// check, or a command that is not on the surface.
//
// IT IS DECLARED HERE, in the leaf package, because it has two parties and they must not each keep
// a copy: the record tool writes it (cli/seat.RenderRefusal) and the failure hook recognises it
// (hookcmd), and the hook binary does not link the command tree. Both read this one declaration,
// so a key added here is written and recognised in the same change.
type RefusalEnvelope struct {
	// Verb is the resolved command's name; "" when no command resolved.
	Verb string `json:"verb"`
	// Role is the role whose surface the call ran on; "" where no seat's surface resolved — the
	// operator's, or a call that named no seat.
	Role string `json:"role"`
	// OK is false: this is a refusal.
	OK bool `json:"ok"`
	// Code is the error's own Code, or "error" for a failure nothing coded.
	Code string `json:"code"`
	// Error is the diagnosis: no role, no command path, no help page.
	Error string `json:"error"`
}

// RefusalKeys is the envelope's keys in the order they are written, read off the declaration: the
// help that tells a seat what a refusal looks like and the reader that recognises one both name
// the keys from here.
func RefusalKeys() []string {
	t := reflect.TypeOf(RefusalEnvelope{})
	keys := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		keys = append(keys, t.Field(i).Tag.Get("json"))
	}
	return keys
}

// DecodeRefusal reads one line as the record tool's refusal envelope, and reports whether it is
// one: an object holding exactly the envelope's keys, typed as declared, with ok false. A JSON
// object of another shape — another tool's `{"ok":false}`, a success envelope — is not.
func DecodeRefusal(line string) (RefusalEnvelope, bool) {
	var env RefusalEnvelope
	var have map[string]json.RawMessage
	keys := RefusalKeys()
	if json.Unmarshal([]byte(line), &have) != nil || len(have) != len(keys) {
		return env, false
	}
	for _, k := range keys {
		if _, ok := have[k]; !ok {
			return env, false
		}
	}
	dec := json.NewDecoder(bytes.NewReader([]byte(line)))
	dec.DisallowUnknownFields()
	if dec.Decode(&env) != nil || env.OK {
		return RefusalEnvelope{}, false
	}
	return env, true
}
