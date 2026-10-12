// Command deadcode fails when the FEOV tools module holds a function nothing reaches.
//
// Dev tooling for this repository only. It runs golang.org/x/tools/cmd/deadcode — pinned as a
// tool dependency of this module, so CI and a developer's box analyse with the same version —
// over the module with `-test`, once per platform the module ships on, and counts the lines.
//
// THE FAIL COUNT IS THE LINE COUNT, AND THERE IS NO ALLOWLIST. deadcode itself exits 0 whatever
// it finds, so its exit status says only that the analysis ran; a function kept "for a reason"
// is kept by being called, and a list of excused names beside the gate would be the same
// hand-kept record the gate exists to replace.
//
// WHAT IT TAKES AS REACHABLE: every `main` package under the module — the shipped binaries
// under cmd/ and the generators under devcmd/ alike — and, with `-test`, every test
// executable. So a function only a test calls is alive here; this gate answers "does anything
// at all reach it", not "does the product".
//
// TWO PLATFORMS because the module holds build-tagged files: a function reached only from a
// Windows file is dead on Linux's list and alive on Windows', and one platform's list alone
// would report it — or miss its opposite.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// module is the one module the gate covers, relative to the repository root.
const module = "plugins/frank-exchange-of-views/tools"

// platforms are the GOOS values the module is analysed under.
var platforms = []string{"linux", "windows"}

func main() {
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "deadcode:", err)
		os.Exit(1)
	}
	tool, err := pinnedTool(filepath.Join(root, "scripts"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "deadcode:", err)
		os.Exit(1)
	}
	found := 0
	for _, goos := range platforms {
		lines, err := unreachable(tool, filepath.Join(root, module), goos)
		if err != nil {
			// NOT MEASURED IS NOT CLEAN: an analysis that could not load the module printed no
			// findings either.
			fmt.Fprintf(os.Stderr, "deadcode: GOOS=%s: the analysis did not run, so nothing was measured: %v\n", goos, err)
			os.Exit(1)
		}
		for _, l := range lines {
			fmt.Printf("GOOS=%s %s/%s\n", goos, module, l)
		}
		found += len(lines)
	}
	if found > 0 {
		fmt.Fprintf(os.Stderr, "deadcode: %d line(s) of unreachable functions in %s — delete each, or call it; there is no allowlist\n", found, module)
		os.Exit(1)
	}
	fmt.Printf("deadcode: %s holds no unreachable function (-test; GOOS %s)\n", module, strings.Join(platforms, ", "))
}

// pinnedTool builds the deadcode this module pins and returns the executable's path.
func pinnedTool(scripts string) (string, error) {
	cmd := exec.Command("go", "tool", "-n", "deadcode")
	cmd.Dir = scripts
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolving the pinned tool (`go tool -n deadcode` in scripts/): %v: %s", err, strings.TrimSpace(errOut.String()))
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", fmt.Errorf("`go tool -n deadcode` named no executable")
	}
	return path, nil
}

// unreachable is the tool's list for one module under one GOOS, a line per function.
func unreachable(tool, dir, goos string) ([]string, error) {
	cmd := exec.Command(tool, "-test", "./...")
	cmd.Dir = dir
	// CGO off, as a cross-compile has it: the list is then the same whichever box computes it.
	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED=0")
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(errOut.String()))
	}
	var lines []string
	for _, l := range strings.Split(out.String(), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}
