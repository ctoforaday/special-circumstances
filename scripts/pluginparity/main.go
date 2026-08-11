// Command plugin-list-parity checks the three places the plugin set is written down.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// The set of plugins is declared in three places that cannot import each other: the
// marketplace manifest (JSON), the bootstrap script (a bash array), and the pasteable
// snippet in docs/setup-script.md (a bash for-loop inside a fenced block). A setup script
// cannot read the manifest — it runs before any checkout exists in the general case — so the
// duplication is STRUCTURAL. This makes it loud instead.
//
// The failure it catches is silent by construction: an unnamed plugin is simply never
// installed, and a session missing one behaves exactly like a session predating it.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

var (
	// importLine matches CLAUDE.md's always-on imports: `@plugins/…/SKILL.md`.
	importLine = regexp.MustCompile(`(?m)^@(\S+\.md)\s*$`)
	// alwaysOn matches a skill whose frontmatter description declares it always-on.
	alwaysOn      = regexp.MustCompile(`(?m)^description:.*Always-on`)
	bootstrapList = regexp.MustCompile(`(?m)^PLUGINS=\(([^)]*)\)`)
	snippetList   = regexp.MustCompile(`(?m)^for p in ([^;]+); do$`)
	claimedCount  = regexp.MustCompile(`(\d+) at the time of writing`)
)

type manifest struct {
	Plugins []struct{ Name string } `json:"plugins"`
}

// fields splits a whitespace-separated list, which is how both bash forms spell it.
func fields(s string) []string { return strings.Fields(strings.TrimSpace(s)) }

func main() {
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "plugin-list-parity:", err)
		os.Exit(1)
	}
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			fmt.Fprintf(os.Stderr, "plugin-list-parity: cannot read %s: %v\n", rel, err)
			os.Exit(1)
		}
		return string(b)
	}

	var m manifest
	if err := json.Unmarshal([]byte(read(".claude-plugin/marketplace.json")), &m); err != nil {
		fmt.Fprintln(os.Stderr, "plugin-list-parity: marketplace.json does not parse:", err)
		os.Exit(1)
	}
	var declared []string
	for _, p := range m.Plugins {
		declared = append(declared, p.Name)
	}
	sort.Strings(declared)

	failed := false
	fail := func(format string, a ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
		failed = true
	}

	for _, problem := range committedBinaries(root) {
		fail("%s", problem)
	}

	// Each source, with the pattern that finds it. A pattern that stops matching is
	// reported as such rather than treated as an empty list: a guard that reads nothing
	// and says nothing is the failure mode these checks exist to prevent.
	sources := []struct {
		where, text string
		re          *regexp.Regexp
	}{
		{"scripts/bootstrap-plugins.sh (PLUGINS=…)", read("scripts/bootstrap-plugins.sh"), bootstrapList},
		{"docs/setup-script.md (for p in …)", read("docs/setup-script.md"), snippetList},
	}
	for _, s := range sources {
		mm := s.re.FindStringSubmatch(s.text)
		if mm == nil {
			fail("%s: could not find the plugin list at all — the pattern this check greps for has moved, so it has been passing without reading anything.", s.where)
			continue
		}
		have := fields(mm[1])
		sort.Strings(have)
		for _, p := range declared {
			if !contains(have, p) {
				fail("%s: never installs %s — declared in .claude-plugin/marketplace.json but not here.", s.where, p)
			}
		}
		for _, p := range have {
			if !contains(declared, p) {
				fail("%s: installs %s, which the marketplace does not offer — the install will fail at setup time.", s.where, p)
			}
		}
	}

	// The same failure one layer down. docs/setup-script.md tells the reader how many hook
	// binaries a cold bootstrap builds, so they can tell a complete run from a truncated
	// one — and nothing recomputed that number. It was written as 8, corrected to 9, and
	// went stale again the day gray-area gained a second command. A verification table that
	// quietly disagrees with reality is worse than one with no number: it fails a correct
	// run and passes an incomplete one.
	cmdDirs, err := filepath.Glob(filepath.Join(root, "plugins", "*", "tools", "cmd", "*"))
	if err != nil {
		fail("plugin-list-parity: cannot count hook binaries: %v", err)
	}
	dirs := 0
	for _, d := range cmdDirs {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() && !devOnly(d) {
			dirs++
		}
	}
	doc := read("docs/setup-script.md")
	cm := claimedCount.FindStringSubmatch(doc)
	switch {
	case cm == nil:
		fail("docs/setup-script.md: the hook-binary count this check reads is gone — the phrasing moved, so the check has been passing without reading anything.")
	default:
		claimed, _ := strconv.Atoi(cm[1])
		if claimed != dirs {
			fail("docs/setup-script.md says a cold bootstrap builds %d hook binaries; plugins/*/tools/cmd/ holds %d. Update the doc — a reader counting binaries against a stale number cannot tell a complete run from a broken one.", claimed, dirs)
		}
	}

	// The same failure one layer down again: requirements.json documents the hook binaries
	// a consumer should expect, and nothing recomputed it. It drifted twice — sc-strike-counter
	// was never added, and the two PostToolUse binaries outlived the merge that replaced them.
	// A manifest that quietly disagrees with the tree is worse than none: it is read as
	// authoritative.
	for _, mf := range manifestsOf(root, "requirements.json") {
		plugin := strings.Split(filepath.ToSlash(mf), "/")[1]
		listed, ok := hookBinariesIn(filepath.Join(root, mf))
		if !ok {
			continue // a plugin that documents no hook binaries makes no claim to check
		}
		onDisk := map[string]bool{}
		for _, d := range cmdDirs {
			if fi, err := os.Stat(d); err == nil && fi.IsDir() && !devOnly(d) &&
				strings.Contains(filepath.ToSlash(d), "plugins/"+plugin+"/") {
				onDisk[filepath.Base(d)] = true
			}
		}
		for _, n := range listed {
			if !onDisk[n] {
				fail("%s: documents hook binary %s, which has no directory under plugins/%s/tools/cmd/.", mf, n, plugin)
			}
		}
		for n := range onDisk {
			if !contains(listed, n) && !documentedException(n) {
				fail("%s: plugins/%s/tools/cmd/%s exists and is not documented in _hook_binaries — a consumer reading this manifest would not know to expect it.", mf, plugin, n)
			}
		}
	}

	for _, p := range alwaysOnProblems(root) {
		fail("%s", p)
	}

	if failed {
		os.Exit(1)
	}
	fmt.Printf("plugin lists agree: %s — %d hook binaries\n", strings.Join(declared, ", "), dirs)
}

func contains(h []string, n string) bool {
	for _, x := range h {
		if x == n {
			return true
		}
	}
	return false
}

// documentedException names the cmd/ directories that are deliberately absent from a
// plugin's _hook_binaries list. sc-doctor is an operator command, not a hook; it is
// provisioned the same way and is called out in the manifest's own comment.
func documentedException(name string) bool { return name == "sc-doctor" }

// devOnly reports whether a cmd/ directory is a development harness rather than a shipped
// binary, by the marker the BUILD script reads for the same decision.
//
// The marker rather than a name list here, and this check is the reason why: it exists because
// the documented count went stale twice and the _hook_binaries manifest twice more. A list of
// exempt names living in this file would be a third hand-maintained copy of the same fact,
// kept by the same hands that let the first two drift.
func devOnly(cmdDir string) bool {
	_, err := os.Stat(filepath.Join(cmdDir, "DEV-ONLY"))
	return err == nil
}

// manifestsOf lists a filename under every plugin, repo-relative.
func manifestsOf(root, name string) []string {
	matches, _ := filepath.Glob(filepath.Join(root, "plugins", "*", name))
	var out []string
	for _, m := range matches {
		rel, _ := filepath.Rel(root, m)
		out = append(out, filepath.ToSlash(rel))
	}
	sort.Strings(out)
	return out
}

// hookBinariesIn reads the documented binary names, reporting whether the file makes the
// claim at all.
func hookBinariesIn(path string) ([]string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var d struct {
		HookBinaries struct {
			Binaries []struct{ Name string } `json:"binaries"`
		} `json:"_hook_binaries"`
	}
	if json.Unmarshal(b, &d) != nil || len(d.HookBinaries.Binaries) == 0 {
		return nil, false
	}
	var out []string
	for _, x := range d.HookBinaries.Binaries {
		out = append(out, x.Name)
	}
	return out, true
}

// alwaysOnProblems reports every disagreement between CLAUDE.md's @ imports and the skills
// that describe themselves as always-on. Separated from main so the failure it exists to
// catch is reproducible in a scratch tree.
func alwaysOnProblems(root string) []string {
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }
	// THE ALWAYS-ON IMPORT LIST. CLAUDE.md's @ imports decide which rules load in every
	// session of every consuming project, and nothing checked them. Deleting one line
	// silently disables a rule: the SKILL.md still exists, still passes every check, still
	// carries its trailers — and is never loaded again. Same shape as record.test.mjs never
	// running, and as the hookenv guard that kept passing after its glob stopped matching:
	// the failure is invisible because nothing is WRONG, only absent.
	//
	// Keyed on each skill's own "Always-on" description rather than a hardcoded list, so the
	// check cannot drift the way the list it guards would.
	//
	// NOTE ON WHAT IS *NOT* DONE HERE: #214 also proposed adding CLAUDE.md to rule-sweep's
	// protocol surfaces. That would gate the whole FILE — the repo guide, the structure
	// table, the dev notes — to protect one list inside it, which is `invariant-at-wrong-
	// level` from the registry: enforcement at a surface coarser than the thing it protects.
	// This check guards the list itself.
	claude := ""
	if b, err := os.ReadFile(filepath.Join(root, "CLAUDE.md")); err == nil {
		claude = string(b)
	}
	imports := map[string]bool{}
	for _, m := range importLine.FindAllStringSubmatch(claude, -1) {
		rel := filepath.ToSlash(m[1])
		imports[rel] = true
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			add("CLAUDE.md imports %s, which does not exist — that rule is silently not loaded.", rel)
		}
	}
	if len(imports) == 0 {
		add("CLAUDE.md declares no @ imports at all — either every always-on rule just stopped loading, or the pattern this check greps for has moved and it has been passing without reading anything.")
	}

	skills, _ := filepath.Glob(filepath.Join(root, "plugins", "*", "skills", "*", "SKILL.md"))
	sort.Strings(skills)
	for _, abs := range skills {
		rel, _ := filepath.Rel(root, abs)
		rel = filepath.ToSlash(rel)
		b, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		declared := alwaysOn.MatchString(string(b))
		switch {
		case declared && !imports[rel]:
			add("%s describes itself as Always-on and is NOT imported by CLAUDE.md — it does not load, in any session, and nothing else would say so.", rel)
		case !declared && imports[rel]:
			add("%s is imported by CLAUDE.md but does not describe itself as Always-on — the marker is what this check keys on, so an unmarked import makes the guard blind to its own removal.", rel)
		}
	}

	return problems
}

// executableMagic are the leading bytes of the formats a build on any platform this suite
// runs on produces. Matching on CONTENT rather than on a name or a path is the point: the
// accident this catches was a file called `feov-record` with no extension sitting in a
// directory full of legitimate source, and a name-based rule would have to be extended for
// every new command — which is the hand-maintained list that let it happen.
var executableMagic = [][]byte{
	{'M', 'Z'},               // PE (Windows .exe)
	{0x7f, 'E', 'L', 'F'},    // ELF (Linux)
	{0xcf, 0xfa, 0xed, 0xfe}, // Mach-O 64-bit little-endian (macOS)
	{0xca, 0xfe, 0xba, 0xbe}, // Mach-O universal
}

// committedBinaries reports every TRACKED file under plugins/ whose bytes are an executable.
//
// WHY CONTENT AND NOT .gitignore ALONE. .gitignore already said these are "never committed"
// and had done for a long time; it covered bin/ and tools/dist/ and not the path `go build`
// actually writes to. A policy with a hole is how a 6 MB feov-record reporting 0.17.0 stayed
// tracked through nineteen version bumps, read by nothing and rebuilt by no one. .gitignore
// stops the known names; this stops the class.
//
// Binaries reach an installing project through GitHub Releases (SHA256-verified) or a local
// `go build` via doctor --fix. Neither route goes through git, so a tracked executable is
// always an accident.
func committedBinaries(root string) []string {
	out, err := gitx.Run(root, "ls-files", "plugins")
	if err != nil {
		return []string{fmt.Sprintf("cannot list tracked plugin files, so the committed-binary check did not run: %v", err)}
	}
	var problems []string
	seen := 0
	for _, rel := range strings.Split(out, "\n") {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		seen++
		f, err := os.Open(filepath.Join(root, rel))
		if err != nil {
			continue // deleted-but-staged, or a symlink; not this check's business
		}
		head := make([]byte, 4)
		n, _ := io.ReadFull(f, head)
		f.Close()
		if n < 2 {
			continue
		}
		for _, magic := range executableMagic {
			if n >= len(magic) && bytes.Equal(head[:len(magic)], magic) {
				problems = append(problems, fmt.Sprintf(
					"%s is a COMMITTED EXECUTABLE. Binaries reach a consumer through GitHub Releases or a local build (doctor --fix), never through git — a tracked one goes stale in silence: the last was 19 versions behind, read by nothing, and rebuilt by no one. Delete it, and add its path to .gitignore so the next build does not re-add it.", rel))
				break
			}
		}
	}
	// A check that reads nothing must SAY so rather than report success — the same failure
	// mode the plugin-list guards above are written against.
	if seen == 0 {
		return []string{"the committed-binary check listed zero tracked files under plugins/, so it has been passing without reading anything."}
	}
	return problems
}
