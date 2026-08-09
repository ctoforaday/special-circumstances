// Command validate-frontmatter fails loudly on frontmatter the loader would silently drop.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// THE FAILURE. Every agent, command and skill declares its capabilities in YAML
// frontmatter, and Claude Code's loader responds to a parse failure by dropping the WHOLE
// block — name, tools, skills, model, all of it — and loading the file anyway with empty
// metadata. Nothing raises. An agent that declared `tools: Read, Write, Bash` and
// `skills: [...]` runs with neither, behaving exactly like an agent whose declarations do
// not work.
//
// probe.md sat broken this way: an unquoted description containing "skills: " ended the
// scalar early and broke the block. `claude plugin validate` catches it, and nothing in CI
// ran `claude plugin validate` — the declaration was policy with no mechanism.
//
// # Why this now uses a real YAML parser
//
// The .mjs version was deliberately dependency-free and did NOT parse YAML: it pattern-
// matched the constructs that make an unquoted scalar mean something other than it looks
// like. That reasoning held for unquoted scalars and broke for everything else, because its
// escape hatch waved through any value starting with " ' [ { > | as "already quoted,
// flow-style, or a block scalar" WITHOUT checking that it terminates. Measured (#157 §2) —
// three genuine parse failures that a real parser rejects and it reported valid:
//
//	tools: [Read, Write            ParserError   -> "32 blocks valid", exit 0
//	description: "unclosed         ScannerError  -> "32 blocks valid", exit 0
//	tools: + a tab-indented item   ScannerError  -> "32 blocks valid", exit 0
//
// All three produce exactly the silent whole-block drop this guard exists to catch. So the
// parser is now the authority on "does this parse", which is the only question the loader
// actually asks.
//
// # What the parser still cannot see, and why the heuristic stays
//
// `description: does things # oops` is VALID YAML. It parses, and the value silently
// becomes "does things" — the field loads TRUNCATED with nothing to show for it. No parser
// will ever flag that, because nothing is wrong with it as YAML; it is wrong as an
// intention. That check is therefore kept, and it is the reason this is not simply
// `yaml.Unmarshal` in a loop.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
	"gopkg.in/yaml.v3"
)

// keyLine matches a top-level `key: value` in a frontmatter block.
var keyLine = regexp.MustCompile(`^([A-Za-z_][\w-]*):\s+(\S.*)$`)

// problem is one finding, pre-formatted for the operator.
type problem string

// check validates one file's frontmatter and returns findings plus whether a block was
// scanned at all.
func check(file, text string) (problems []problem, scanned bool) {
	// CRLF FIRST, AND THIS WAS A REAL MISS (#323). The prefix test was `---\n`, so a file
	// written with CRLF line endings starts `---\r\n`, fails the test, and is reported as
	// HAVING NO FRONTMATTER — silently skipped, not flagged. Claude Code's loader reads it
	// fine; only this guard could not.
	//
	// Measured 2026-08-08: an edit that rewrote two agent constitutions with CRLF dropped the
	// count from 36 to 34 and the gate EXITED 0. Two constitutions stopped being checked and
	// the run reported success. On Windows the trigger is ordinary — any editor or script that
	// writes CRLF removes a file from this gate without a word.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, false // no frontmatter is fine HERE; mustCarryFrontmatter decides where it is not
	}
	end := strings.Index(text[3:], "\n---")
	if end < 0 {
		return []problem{problem(fmt.Sprintf(
			"%s: frontmatter opens with --- and never closes; the loader reads the entire file as YAML.", file))}, false
	}
	block := text[4 : end+3]

	// The parser is the authority on whether the loader gets a block at all.
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(block), &doc); err != nil {
		return []problem{problem(fmt.Sprintf(
			"%s: frontmatter does not parse as YAML, so the loader drops EVERY field in it and loads the file anyway with empty metadata: %v",
			file, strings.ReplaceAll(strings.TrimSpace(err.Error()), "\n", " ")))}, true
	}

	// Then the thing no parser can see: valid YAML that does not mean what it says.
	for i, line := range strings.Split(block, "\n") {
		m := keyLine.FindStringSubmatch(line)
		if m == nil {
			continue // blank, comment, list item, or a block-scalar continuation
		}
		key, value := m[1], m[2]
		if strings.HasPrefix(value, `"`) || strings.HasPrefix(value, "'") {
			continue // an explicit quote makes " #" literal, so there is nothing to warn about
		}
		if strings.Contains(value, " #") {
			problems = append(problems, problem(fmt.Sprintf(
				"%s:%d: %s is an unquoted value containing \" #\" — YAML reads the rest as a comment, so the field loads truncated with nothing to show for it. Wrap the value in double quotes.",
				file, i+2, key)))
		}
	}
	return problems, true
}

// mustCarryFrontmatter says which files are BROKEN rather than merely plain when they have no
// frontmatter block: a plugin's agents and commands, whose whole contract is declared there.
//
// Everything else this guard scans — READMEs, skill prose, docs, the root documents — is
// ordinary markdown, and a missing block means nothing. Naming the mandatory set is what turns
// "the count went down" into a failure with a filename on it.
func mustCarryFrontmatter(file string) bool {
	slash := filepath.ToSlash(file)
	if !strings.HasPrefix(slash, "plugins/") {
		return false
	}
	return strings.Contains(slash, "/agents/") || strings.Contains(slash, "/commands/")
}

// markdownFiles lists the tracked .md files this guard owns: everything under plugins/, plus
// the repo-root documents. Tracked only — an untracked scratch file is the author's business.
func markdownFiles(root string) ([]string, error) {
	out, err := gitx.Run(root, "ls-files", "*.md", "**/*.md")
	if err != nil {
		return nil, err
	}
	var kept []string
	for _, f := range gitx.Lines(out) {
		slash := filepath.ToSlash(f)
		if strings.HasPrefix(slash, "plugins/") || !strings.Contains(slash, "/") {
			kept = append(kept, slash)
		}
	}
	sort.Strings(kept)
	return kept, nil
}

func main() {
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate-frontmatter:", err)
		os.Exit(1)
	}
	files, err := markdownFiles(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate-frontmatter:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "validate-frontmatter: matched no markdown at all — the listing is wrong, not the tree")
		os.Exit(1)
	}

	var problems []problem
	scanned := 0
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(root, f))
		if err != nil {
			problems = append(problems, problem(fmt.Sprintf("%s: %v", f, err)))
			continue
		}
		p, did := check(f, string(b))
		problems = append(problems, p...)
		if did {
			scanned++
			continue
		}
		// THE MISS IS LOUD WHERE A BLOCK IS MANDATORY. A file that stops being recognized is
		// not invalid — it is ABSENT, and absence was indistinguishable from a clean board:
		// the count moved and nothing failed. An agent or command with no frontmatter loads
		// with empty metadata and behaves like one whose declarations do not work, which is
		// the exact failure this guard exists to prevent.
		if mustCarryFrontmatter(f) {
			problems = append(problems, problem(fmt.Sprintf(
				"%s: no frontmatter block at all. An agent/command without one loads with EMPTY metadata — no name, no tools, no skills — and runs anyway. If this file is not an agent or command, it is in the wrong directory.", f)))
		}
	}

	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, string(p))
		}
		fmt.Fprintf(os.Stderr, "\n%d frontmatter problem(s). These do not fail loudly at runtime; that is why they are checked here.\n", len(problems))
		os.Exit(1)
	}
	fmt.Printf("%d frontmatter blocks valid\n", scanned)
}
