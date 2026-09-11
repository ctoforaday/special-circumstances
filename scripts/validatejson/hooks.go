package main

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"sort"
)

// A PLUGIN'S hooks/hooks.json HAS NO SCHEMA THE CLIENT PUBLISHES, and the client answers a key it
// does not know by IGNORING it and printing "unknown key … ignored" at the start of every session
// (#913). The hook still runs, so nothing else ever fails — every consumer's every session just
// opens with the noise.
//
// THE KEY SETS ARE KEPT HERE BY HAND, and that is a choice with a stated reason: the client's own
// validator does not check this file. `claude plugin validate --strict` passed all four plugins on
// 2026-09-11 (2.1.268) while they carried 15 `_comment` keys the runtime warned about. The sets
// are the fields https://code.claude.com/docs/en/hooks documents — widen them when the docs add a
// field. A documented key refused here costs a one-line fix; an undocumented one let through costs
// a warning in every session of every consumer.
var (
	hookTopKeys   = keySet("description", "hooks")
	hookGroupKeys = keySet("matcher", "hooks")
	hookEntryKeys = keySet(
		"type", "if", "timeout", "statusMessage", "once", // every hook
		"command", "args", "async", "asyncRewake", "shell", // type: command
		"url", "headers", "allowedEnvVars", // type: http
		"server", "tool", "input", // type: mcp_tool
		"prompt", "model", // type: prompt and agent
	)
)

func keySet(keys ...string) map[string]bool {
	s := make(map[string]bool, len(keys))
	for _, k := range keys {
		s[k] = true
	}
	return s
}

// isPluginHooks reports whether a tracked, slash-separated path is a plugin's hooks.json. path.Match,
// not filepath.Match: on Windows the latter's `*` also crosses `/`, and would accept a hooks.json
// nested one directory deeper than any plugin puts one.
func isPluginHooks(p string) bool {
	ok, _ := path.Match("plugins/*/hooks/hooks.json", filepath.ToSlash(p))
	return ok
}

// hookKeyProblems names every key in a plugin hooks.json that the hooks reference does not
// document, and every place the file is not the documented shape. Nil means clean.
func hookKeyProblems(b []byte) []string {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		return []string{"not a JSON object: " + err.Error()}
	}
	out := undocumented("the top level", top, hookTopKeys)
	raw, ok := top["hooks"]
	if !ok {
		return append(out, `no "hooks" object, so this file registers nothing`)
	}
	var events map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(raw, &events); err != nil {
		return append(out, `"hooks" is not an object of event → matcher groups: `+err.Error())
	}
	for _, ev := range sortedKeys(events) {
		for i, group := range events[ev] {
			at := fmt.Sprintf("hooks.%s[%d]", ev, i)
			out = append(out, undocumented(at, group, hookGroupKeys)...)
			var entries []map[string]json.RawMessage
			if err := json.Unmarshal(group["hooks"], &entries); err != nil {
				out = append(out, at+`.hooks is not an array of hooks: `+err.Error())
				continue
			}
			for j, e := range entries {
				out = append(out, undocumented(fmt.Sprintf("%s.hooks[%d]", at, j), e, hookEntryKeys)...)
			}
		}
	}
	return out
}

func undocumented[V any](at string, obj map[string]V, allowed map[string]bool) []string {
	var out []string
	for _, k := range sortedKeys(obj) {
		if !allowed[k] {
			out = append(out, fmt.Sprintf("%s: key %q is not one the hooks reference documents — the "+
				"client ignores it and warns at the start of every session (#913); put prose in "+
				"hooks/README.md", at, k))
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
