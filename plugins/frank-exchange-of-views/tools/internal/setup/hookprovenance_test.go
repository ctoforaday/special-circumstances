package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// cachePlugin writes a plugin cache entry: a version directory with the given bin/ contents and
// a hooks.json registering the given hook binaries.
func cachePlugin(t *testing.T, home, marketplace, plugin, version string, bins, registers []string) {
	t.Helper()
	dir := filepath.Join(home, ".claude", "plugins", "cache", marketplace, plugin, version)
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, b := range bins {
		if err := os.WriteFile(filepath.Join(dir, "bin", b), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if registers == nil {
		return
	}
	hooks := `{"hooks":{"PreToolUse":[{"hooks":[`
	for i, r := range registers {
		if i > 0 {
			hooks += ","
		}
		hooks += `{"type":"command","command":"B=\"${CLAUDE_PLUGIN_ROOT}/bin/` + r + `\"; exec \"$B\""}`
	}
	hooks += `]}]}}`
	if err := os.MkdirAll(filepath.Join(dir, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hooks", "hooks.json"), []byte(hooks), 0o644); err != nil {
		t.Fatal(err)
	}
}

// NO CACHED PLUGIN IS AN ANSWER, NOT AN ABSENCE — which is the whole point of #751.
//
// The defect this records was invisible because a run with stale hooks and a run with current
// hooks were byte-identical on the record. A field that went missing when there was nothing to
// report would reproduce that inside the fix: "no hooks installed" and "this binary predates the
// field" would be the same bytes again.
func TestNoCachedPluginIsRecordedAsAnObservationNotAnEmptyField(t *testing.T) {
	p := hookProvenanceAt(t.TempDir(), "frank-exchange-of-views")
	if p.State != "absent" {
		t.Errorf("state = %q, want absent", p.State)
	}
	if p.Why == "" {
		t.Error("an absent cache says nothing about WHY, so a reader has to guess whether it was searched")
	}
	// The zero value must survive a round trip as a positive statement.
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"state":"absent"`, `"why":`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("the marshalled provenance drops %s — an unstamped run and one that observed nothing become the same bytes: %s", want, b)
		}
	}
}

// THE BOOTSTRAP WINDOW IS THE FINDING, and it is what the 2026-08-23 run was sitting in: a cached
// version that REGISTERS a hook binary it does not ship. The entry fires, the guard finds no
// binary, and the run proceeds with no identity injection while the fetch runs.
func TestAVersionThatRegistersAHookItDoesNotShipIsReported(t *testing.T) {
	home := t.TempDir()
	cachePlugin(t, home, "sc", "frank-exchange-of-views", "1.58.0",
		[]string{"feov-record"}, []string{"feov-record", "feov-pretooluse"})

	p := hookProvenanceAt(home, "frank-exchange-of-views")
	if p.State != "installed" {
		t.Fatalf("state = %q, want installed", p.State)
	}
	if !slices.Contains(p.Versions, "1.58.0") {
		t.Errorf("versions = %v, want 1.58.0", p.Versions)
	}
	if got := p.Missing["1.58.0"]; !slices.Contains(got, "feov-pretooluse") {
		t.Errorf("missing = %v — a version registering feov-pretooluse without shipping it is exactly "+
			"the 2026-08-23 shape, and it is not reported", got)
	}
	// ANTI-VACUITY: a binary that IS shipped must not be reported missing, or the field says
	// "everything is missing" and means nothing.
	if got := p.Missing["1.58.0"]; slices.Contains(got, "feov-record") {
		t.Errorf("feov-record is present in bin/ and reported missing: %v", got)
	}
	if got := p.Binaries["1.58.0"]; !slices.Contains(got, "feov-record") {
		t.Errorf("binaries = %v, want feov-record", got)
	}
}

// TWO CACHED VERSIONS IS ITSELF THE FINDING. The harness picks one and nothing here can say
// which, so a run with two is ambiguous and a reader should see both rather than one.
func TestEveryCachedVersionIsRecordedBecauseNothingHereKnowsWhichRan(t *testing.T) {
	home := t.TempDir()
	cachePlugin(t, home, "sc", "frank-exchange-of-views", "1.58.0", []string{"feov-record"}, nil)
	cachePlugin(t, home, "sc", "frank-exchange-of-views", "1.63.0",
		[]string{"feov-record", "feov-pretooluse"}, nil)

	p := hookProvenanceAt(home, "frank-exchange-of-views")
	if len(p.Versions) != 2 {
		t.Fatalf("versions = %v, want both cached versions — recording one would state a choice this binary did not make", p.Versions)
	}
	if !slices.Contains(p.Binaries["1.63.0"], "feov-pretooluse") || slices.Contains(p.Binaries["1.58.0"], "feov-pretooluse") {
		t.Errorf("the per-version binary lists are wrong: %v", p.Binaries)
	}
}

// THE FIELD REACHES run-config.json, which is a different question from whether the observer
// works. The three tests above drive hookProvenanceAt directly; this one marshals the struct the
// setup path actually writes, so a field added to the observer and never wired into runConfig —
// or wired and then dropped in a merge — fails here rather than shipping as a fix nobody gets.
//
// It asserts the SHAPE reaches the file, not what a particular machine's cache holds: the value
// depends on the developer's install state, and a test that asserted a version would pass or fail
// on where it ran.
func TestRunConfigCarriesHookProvenance(t *testing.T) {
	rc := runConfig{Topic: "t", Hooks: hookProvenanceAt(t.TempDir(), "frank-exchange-of-views")}
	b, err := json.Marshal(rc)
	if err != nil {
		t.Fatal(err)
	}
	var back struct {
		Hooks *HookProvenance `json:"hooks"`
	}
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Hooks == nil {
		t.Fatal("run-config.json carries no `hooks` key — the provenance is observed and then dropped, " +
			"which leaves the record exactly as silent about the hook side as before")
	}
	if back.Hooks.State == "" {
		t.Error("the hook provenance round-trips with an empty state, so a reader cannot tell " +
			"'nothing was installed' from 'this binary did not look'")
	}
}

// A HOOKS.JSON THAT COULD NOT BE READ IS NOT "NOTHING MISSING". An empty Missing list is the
// answer a stale install is misread as, so a version whose hooks.json is absent, does not decode,
// or names its binaries in a form this reader does not know gives no list, and State and Why say
// the registered set was not measured.
func TestAnUnreadableHooksFileIsStatedNotAnEmptyList(t *testing.T) {
	for name, hooks := range map[string]string{
		"absent":          "",
		"does not decode": `{"hooks":`,
		"another form":    `{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"exec \"$CLAUDE_PLUGIN_ROOT/bin/feov-pretooluse\""}]}]}}`,
	} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			cachePlugin(t, home, "sc", "frank-exchange-of-views", "1.58.0", []string{"feov-record"}, nil)
			if hooks != "" {
				dir := filepath.Join(home, ".claude", "plugins", "cache", "sc", "frank-exchange-of-views", "1.58.0", "hooks")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "hooks.json"), []byte(hooks), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			p := hookProvenanceAt(home, "frank-exchange-of-views")
			if p.State != "unreadable" || !strings.Contains(p.Why, "1.58.0") {
				t.Fatalf("state = %q, why = %q; want unreadable, naming the version", p.State, p.Why)
			}
			if got, ok := p.Missing["1.58.0"]; ok {
				t.Errorf("missing[1.58.0] = %v — a list where none was read reads as nothing missing", got)
			}
		})
	}
}

// THE SHIPPED HOOKS.JSON IS READ AS THE SHIPPED FORM. A cached copy of this plugin's own
// hooks.json with an empty bin/ must list every binary it registers as missing: if the file's
// command form drifts from the one the reader knows, this fails here rather than at a run whose
// provenance then says nothing.
func TestTheShippedHooksFileRegistersItsBinaries(t *testing.T) {
	path, err := repotree.Plugin("hooks", "hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	shipped, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	cachePlugin(t, home, "sc", "frank-exchange-of-views", "9.9.9", nil, nil)
	dir := filepath.Join(home, ".claude", "plugins", "cache", "sc", "frank-exchange-of-views", "9.9.9", "hooks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hooks.json"), shipped, 0o644); err != nil {
		t.Fatal(err)
	}
	p := hookProvenanceAt(home, "frank-exchange-of-views")
	if p.State != "installed" {
		t.Fatalf("state = %q (%s), want installed", p.State, p.Why)
	}
	for _, want := range []string{"feov-pretooluse", "feov-subagentstart", "feov-subagentstop"} {
		if !slices.Contains(p.Missing["9.9.9"], want) {
			t.Errorf("missing = %v, want %s among the binaries the shipped hooks.json registers", p.Missing["9.9.9"], want)
		}
	}
}
