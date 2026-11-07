// Command staticproof asserts the per-format staticness contract over binaries.
//
// Usage:
//
//	go run ./scripts/staticproof [-artifact <name>] <binary>
//	(or releaseassets -check)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/staticproof"
)

func main() {
	artifact := flag.String("artifact", "", "artifact name (e.g. feov-ocr, feov-embed, or Go command)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: staticproof [-artifact <name>] <binary>")
		os.Exit(1)
	}

	bin := flag.Arg(0)
	art := *artifact
	if art == "" {
		base := filepath.Base(bin)
		// strip extension and target suffix if present
		base = strings.TrimSuffix(base, ".exe")
		for _, target := range []string{"_linux_amd64", "_linux_arm64", "_windows_amd64", "_windows_arm64", "_darwin_arm64"} {
			base = strings.TrimSuffix(base, target)
		}
		art = base
	}

	if err := staticproof.Check(art, bin); err != nil {
		fmt.Fprintf(os.Stderr, "staticproof: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: static proof passed (%s)\n", bin, art)
}
