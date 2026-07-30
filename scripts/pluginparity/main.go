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
	"fmt"
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
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
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
