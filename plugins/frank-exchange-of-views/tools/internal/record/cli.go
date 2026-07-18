package record

import (
	"fmt"
	"os"
	"strings"
)

// Args is the parsed flag set. Booleans and strings are kept apart because the
// oracle distinguishes them: `--severity` with no value yields `true`, which then
// fails the grade enum rather than being coerced to "true".
type Args struct {
	vals       map[string]string
	bools      map[string]bool
	Positional []string
}

func ParseArgs(argv []string) *Args {
	a := &Args{vals: map[string]string{}, bools: map[string]bool{}}
	for i := 0; i < len(argv); i++ {
		s := argv[i]
		if strings.HasPrefix(s, "--") {
			k := s[2:]
			if i+1 >= len(argv) || strings.HasPrefix(argv[i+1], "--") {
				a.bools[k] = true
			} else {
				a.vals[k] = argv[i+1]
				i++
			}
			continue
		}
		a.Positional = append(a.Positional, s)
	}
	return a
}

// Get returns the string value; ok is false when the flag was absent entirely.
// A boolean flag reports ok=true with an empty string, and IsBool tells them apart.
func (a *Args) Get(k string) (string, bool) {
	if v, ok := a.vals[k]; ok {
		return v, true
	}
	if a.bools[k] {
		return "", true
	}
	return "", false
}

func (a *Args) Str(k string) string    { return a.vals[k] }
func (a *Args) Has(k string) bool      { _, ok := a.Get(k); return ok }
func (a *Args) IsBool(k string) bool   { return a.bools[k] }
func (a *Args) HasValue(k string) bool { _, ok := a.vals[k]; return ok }

// Val returns what the oracle would place in a payload literal: the string when
// one was given, the boolean true when the flag was bare.
func (a *Args) Val(k string) any {
	if v, ok := a.vals[k]; ok {
		return v
	}
	if a.bools[k] {
		return true
	}
	return nil
}

// SetInto copies a flag into the payload only when present — the `undefined`
// drop rule that keeps absent flags out of the event entirely.
func (a *Args) SetInto(p *Payload, key, flag string) *Payload {
	if a.Has(flag) {
		p.Set(key, a.Val(flag))
	}
	return p
}

// Copy is the declarative form of SetInto for the common case: a verb listing the
// payload it builds, in order, as specs.
//
// A spec is "name" when the payload key and the flag agree, or "payload_key:flag"
// when they differ (the CLI spells flags in kebab-case, the record in snake_case).
// Absent flags are dropped, preserving the oracle's `undefined` semantics.
//
//	a.Copy(p, "gap_id:id", "severity", "likelihood", "complexity_cost:cx", "basis")
//
// This replaces the repeated SetInto chains the verb table would otherwise carry:
// the verb reads as its payload SHAPE rather than as a sequence of assignments,
// which is what makes a missing field visible on inspection.
func (a *Args) Copy(p *Payload, specs ...string) *Payload {
	for _, spec := range specs {
		key, flag, found := strings.Cut(spec, ":")
		if !found {
			flag = key
		}
		a.SetInto(p, key, flag)
	}
	return p
}

// List mirrors `list()`: comma-separated, trimmed, empties removed, ALWAYS an
// array (the oracle writes [] rather than dropping the key).
func List(v string) []string {
	out := []string{}
	for _, s := range strings.Split(v, ",") {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// PayloadText resolves --file (read whole) or --text, else "".
func PayloadText(a *Args) (string, error) {
	if f, ok := a.vals["file"]; ok {
		b, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if v, ok := a.vals["text"]; ok {
		return v, nil
	}
	if a.bools["text"] {
		return "true", nil // String(true), as the oracle would
	}
	return "", nil
}

// Verb is one entry in a seat's contract. The verb SET is the role boundary: a
// missing verb is a missing capability, enforced by topology rather than by
// asking a seat to behave.
type Verb struct {
	Name string
	Help string
	Fn   func(runDir, seatID string, a *Args) (string, error)
}

// RunVerb is the CLI entry shared by every seat role.
func RunVerb(toolName string, verbs []Verb, argv []string) int {
	if len(argv) == 0 || argv[0] == "--help" || argv[0] == "help" {
		printHelp(toolName, verbs)
		return 0
	}
	verb := argv[0]
	a := ParseArgs(argv[1:])

	runDir, ok := a.Get("run")
	if !ok || runDir == "" {
		fmt.Fprintf(os.Stderr, "%s: --run <runDir> is required\n", toolName)
		return 2
	}
	if verb == "render" {
		r, err := Render(runDir, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", toolName, err)
			return 2
		}
		extra := ""
		if r.Anomalies > 0 {
			extra = fmt.Sprintf(", %d anomalies", r.Anomalies)
		}
		fmt.Printf("%s: rendered to %s (%d open, %d closed%s)\n", toolName, r.Out, r.Open, r.Closed, extra)
		return 0
	}
	var h *Verb
	for i := range verbs {
		if verbs[i].Name == verb {
			h = &verbs[i]
			break
		}
	}
	if h == nil {
		fmt.Fprintf(os.Stderr, "%s: verb '%s' is outside this seat's role (available: %s, render)\n", toolName, verb, verbNames(verbs))
		return 2
	}
	seatID, ok := a.Get("seat-id")
	if !ok || seatID == "" {
		fmt.Fprintf(os.Stderr, "%s: --seat-id is required (the engine assigns it; it is in your prompt)\n", toolName)
		return 2
	}
	result, err := h.Fn(runDir, seatID, a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", toolName, err)
		return 2
	}
	// Render-on-mutation: projections stay current after every write.
	if verb != "register" {
		if _, rerr := Render(runDir, ""); rerr != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", toolName, rerr)
			return 2
		}
	}
	if result != "" {
		fmt.Println(result)
	}
	return 0
}

func verbNames(verbs []Verb) string {
	names := make([]string, len(verbs))
	for i, v := range verbs {
		names[i] = v.Name
	}
	return strings.Join(names, ", ")
}

func printHelp(toolName string, verbs []Verb) {
	fmt.Printf("%s — record contract (the verb set IS the role boundary)\nverbs: %s, render, help\n", toolName, verbNames(verbs))
	for _, v := range verbs {
		fmt.Printf("  %s: %s\n", v.Name, v.Help)
	}
	fmt.Printf("  render: read-only projection refresh (any seat may invoke)\n")
}
