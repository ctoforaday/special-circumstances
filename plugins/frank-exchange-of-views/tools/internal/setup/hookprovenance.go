package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WHICH HOOKS SERVED THIS RUN, recorded at setup because nothing else can say it (#751).
//
// A run already records which RECORD BINARY served it — `tool_version` on every register. It
// records nothing about the hooks, and the two come from different places: `setup` bakes the
// record binary from the WORKING TREE, while hooks are served from the version-gated INSTALL
// CACHE. Measured on 2026-08-23: a run carried tool_version 0.72.0, a version no release held
// until three days later, under hooks from a release tagged six days earlier that did not
// contain `feov-pretooluse` at all.
//
// The consequence is not that something broke. It is that a run whose hooks were six days stale
// is BYTE-IDENTICAL ON THE RECORD to one whose hooks were current, apart from inferring absence
// from a fallback field — so `agent_id: (none)` on 29 of 29 seats reads as evidence about the
// identity design when it is evidence about install lag. Two open issues (#512, #555) are
// downstream of exactly that misreading.
//
// # Why setup, and why not the hook itself
//
// The hook is the only party that knows its own version, and A HOOK THAT IS ABSENT WRITES
// NOTHING — so the fact cannot be self-reported by the thing whose absence is in question. That
// is the whole difficulty, and it is why this is observed from outside rather than announced
// from inside.
//
// This is the cheap half of #751. The stronger half is a liveness probe that asks a firing hook
// its version (#555), which distinguishes absent from stale from current where a filesystem
// stat cannot — a cache entry that exists says nothing about whether the harness invoked it.
// This says what was INSTALLED; it does not claim what RAN.
type HookProvenance struct {
	// State is the answer in one word, and it is never empty: "installed", "absent" (no cached
	// copy of this plugin at all), or "unreadable" (a cache that could not be inspected, or a cached
	// hooks.json whose registered binaries could not be read — Missing then has no entry for it).
	//
	// NOT omitempty, and that is the point rather than a style choice. An absent field and a
	// field saying "nothing was installed" are different facts — one means a binary that predates
	// this record, the other is a positive observation — and omitempty makes them the same bytes.
	// That is the defect this field exists to remove, reproduced inside the fix.
	State string `json:"state"`
	// Versions are the cached versions of this plugin found on disk, newest-sorting last. More
	// than one is normal and is itself the finding: the harness picks one and nothing here can
	// say which, so a run with two cached versions has an ambiguity worth seeing.
	Versions []string `json:"versions"`
	// Binaries lists the hook programs present in each cached version, keyed by version. A
	// version whose hooks.json registers a binary its bin/ does not hold is the bootstrap window
	// every hooks.json guard describes — the entry fires, the binary is missing, and the guard
	// hands it to fetch-bin.sh; until that fetch lands, the hook does nothing.
	Binaries map[string][]string `json:"binaries"`
	// Missing names the hook binaries a cached version REGISTERS and does not ship, per version.
	// Empty is a real answer and is recorded as such.
	Missing map[string][]string `json:"missing"`
	// Why carries the reason when State is not "installed", so a reader is never left inferring
	// one from an empty structure.
	Why string `json:"why"`
}

// hookBinRoot is how a hooks.json command names a binary of its plugin. The prosthetic
// conscience's own gate (hookinvocation) holds every hooks.json command to this form, so reading it
// out of a decoded command is reading an enforced contract rather than guessing at a string.
const hookBinRoot = "${CLAUDE_PLUGIN_ROOT}/bin/"

// hooksFile is the part of hooks.json that registers commands: per event, matcher groups, each a
// list of hooks.
type hooksFile struct {
	Hooks map[string][]struct {
		Hooks []struct {
			Command string `json:"command"`
		} `json:"hooks"`
	} `json:"hooks"`
}

// hookProvenanceAt inspects the plugin install cache under the given home directory.
//
// The path convention is Claude Code's, not ours — `.claude/plugins/cache/<marketplace>/<plugin>/
// <version>` — and prosthetic-conscience's doctor reads the same layout from its own module.
// Two readers of an EXTERNAL convention is not two copies of one of our facts; there is no record
// here that either could be bypassing, and the modules cannot share code across the boundary.
func hookProvenanceAt(home, plugin string) HookProvenance {
	p := HookProvenance{State: "absent", Versions: []string{}, Binaries: map[string][]string{}, Missing: map[string][]string{}}
	cache := filepath.Join(home, ".claude", "plugins", "cache")
	matches, err := filepath.Glob(filepath.Join(cache, "*", plugin, "*"))
	if err != nil {
		p.State, p.Why = "unreadable", "the plugin cache could not be searched: "+err.Error()
		return p
	}
	if len(matches) == 0 {
		p.Why = "no cached copy of " + plugin + " under " + cache +
			" — the hooks this run's seats meet are whatever the harness has loaded, and this binary cannot see it"
		return p
	}
	var unread []string
	for _, dir := range matches {
		fi, serr := os.Stat(dir)
		if serr != nil || !fi.IsDir() {
			continue
		}
		v := filepath.Base(dir)
		p.Versions = append(p.Versions, v)
		p.Binaries[v] = binariesIn(filepath.Join(dir, "bin"))
		missing, rerr := registeredButAbsent(dir, p.Binaries[v])
		if rerr != nil {
			// NOT AN EMPTY LIST. "Registers nothing it lacks" is the answer a stale install is
			// misread as; a hooks.json that could not be read has given no answer, and says so.
			unread = append(unread, v+": "+rerr.Error())
			continue
		}
		p.Missing[v] = missing
	}
	if len(p.Versions) == 0 {
		p.Why = "the cache holds entries for " + plugin + " that are not directories"
		p.State = "unreadable"
		return p
	}
	sort.Strings(p.Versions)
	if len(unread) > 0 {
		sort.Strings(unread)
		p.State = "unreadable"
		p.Why = "which hook binaries a cached version registers could not be read, so what it lacks is not measured — " + strings.Join(unread, "; ")
		return p
	}
	p.State = "installed"
	return p
}

// binariesIn lists executable names in a bin directory. A missing directory is the empty list
// rather than an error: a fresh cache version ships without binaries until the hooks' fetch (or
// `doctor --fix`) installs them, which is the documented bootstrap window and a real state to record.
func binariesIn(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	out := []string{}
	for _, e := range ents {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// registeredButAbsent is the gap that matters: hook entries this version DECLARES whose binary
// it does not carry. That is the window where the harness fires an entry, the guard finds no
// binary, and the run proceeds with no identity injection while the fetch runs — which is the
// 2026-08-23 shape exactly.
func registeredButAbsent(versionDir string, present []string) ([]string, error) {
	path := filepath.Join(versionDir, "hooks", "hooks.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var hf hooksFile
	if err := json.Unmarshal(b, &hf); err != nil {
		return nil, fmt.Errorf("%s does not decode: %w", path, err)
	}
	have := map[string]bool{}
	for _, n := range present {
		have[n] = true
		have[trimExe(n)] = true
	}
	seen := map[string]bool{}
	out := []string{}
	commands, named := 0, 0
	for _, groups := range hf.Hooks {
		for _, g := range groups {
			for _, h := range g.Hooks {
				commands++
				name, ok := binaryOf(h.Command)
				if !ok {
					continue
				}
				named++
				if have[name] || seen[name] {
					continue
				}
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	// A hooks.json whose commands name no binary in the one form is not a version that registers
	// none: every release has registered its hooks by that form, so this is a shape this reader
	// does not know, and an empty list would read as nothing missing.
	if commands > 0 && named == 0 {
		return nil, fmt.Errorf("%s registers %d command(s) and none names a binary as %sNAME", path, commands, hookBinRoot)
	}
	sort.Strings(out)
	return out, nil
}

// binaryOf is the binary a hook command runs by the one form, "${CLAUDE_PLUGIN_ROOT}/bin/<name>",
// where a name is lower-case letters, digits and hyphens; ok is false when it names none.
func binaryOf(command string) (string, bool) {
	i := strings.Index(command, hookBinRoot)
	if i < 0 {
		return "", false
	}
	rest := command[i+len(hookBinRoot):]
	end := strings.IndexFunc(rest, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-')
	})
	if end < 0 {
		end = len(rest)
	}
	return rest[:end], end > 0
}

func trimExe(n string) string {
	if len(n) > 4 && n[len(n)-4:] == ".exe" {
		return n[:len(n)-4]
	}
	return n
}

// homeDir is where the plugin cache lives. It reads $HOME rather than os.UserHomeDir so the
// suite's own HOME sandbox (testbuild) redirects it, which is what makes this testable at all —
// and what stops a test reading the developer's real plugin cache and reporting a different
// answer on every machine.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return h
}
