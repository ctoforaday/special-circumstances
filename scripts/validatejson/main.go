// Command validate-json fails loudly on a malformed tracked .json file.
//
// Dev tooling for this repository only. A broken plugin.json / marketplace.json /
// requirements.json / .mcp.json fails SILENTLY at load time — the loader shrugs and the
// plugin behaves as though the file said nothing — so the noise has to be made here.
//
// Tracked files only, via `git ls-files`: an untracked scratch .json is the author's
// business, and sweeping the whole tree would fail on a temp file nobody ships.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

func main() {
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate-json:", err)
		os.Exit(1)
	}
	out, err := gitx.Run(root, "ls-files", "*.json", "**/*.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate-json:", err)
		os.Exit(1)
	}
	files := gitx.Lines(out)
	if len(files) == 0 {
		// Zero files is not "all valid" — it means the listing broke, and reporting
		// success would be the flattering failure.
		fmt.Fprintln(os.Stderr, "validate-json: git ls-files matched no .json at all — the listing is wrong, not the tree")
		os.Exit(1)
	}

	bad := 0
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(root, f))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			bad++
			continue
		}
		var any any
		if err := json.Unmarshal(b, &any); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			bad++
		}
	}
	if bad > 0 {
		os.Exit(1)
	}
	fmt.Printf("%d JSON files valid\n", len(files))
}
