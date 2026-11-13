package staticproof

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureSrc is a program large enough that its unstripped darwin link keeps several times
// machOMaxLocalSyms local symbols (4,838 under go1.27), and static on both targets.
const fixtureSrc = `package main

import (
	"fmt"
	"net/http"
)

func main() { fmt.Println(http.StatusOK) }
`

// fixture cross-compiles fixtureSrc for goos/arm64 with the given -ldflags: "-s -w" strips it,
// "-w" keeps only its symbols, "-s -w=false" only its debug information, "" both.
func fixture(t *testing.T, goos, ldflags string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{"main.go": fixtureSrc, "go.mod": "module fixture\n\ngo 1.25\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(dir, "bin")
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags", ldflags, "-o", out, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos, "GOARCH=arm64", "GOFLAGS=")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building the %s fixture: %v\n%s", goos, err, b)
	}
	return out
}

// AN UNSTRIPPED HELPER IS REFUSED BY ITS STRUCTURE, whatever its size: an unstripped darwin
// feov-ocr is 10,993,848 B, inside its 12 MiB bound, so the size alone passes it. A stripped
// link of the same program passes, so the refusal is the strip rule's and nothing else's. A link
// that keeps only its symbols, or only its debug information, is refused as well: each of the
// format's rules has a fixture that only it refuses.
func TestAnUnstrippedHelperIsRefusedByItsStructure(t *testing.T) {
	for _, goos := range []string{"linux", "darwin"} {
		t.Run(goos, func(t *testing.T) {
			stripped, unstripped := fixture(t, goos, "-s -w"), fixture(t, goos, "")
			symbolsOnly, debugOnly := fixture(t, goos, "-w"), fixture(t, goos, "-s -w=false")
			for helper := range helperMaxBytes {
				if err := Check(helper, stripped); err != nil {
					t.Errorf("Check(%s, stripped %s) = %v; want it passed", helper, goos, err)
				}
				for name, bin := range map[string]string{"unstripped": unstripped, "symbols-only": symbolsOnly, "debug-only": debugOnly} {
					err := Check(helper, bin)
					if err == nil || !strings.Contains(err.Error(), "without stripping") {
						t.Errorf("Check(%s, %s %s) = %v; want the strip rule's refusal", helper, name, goos, err)
					}
				}
			}
			// A Go command is stripped by its own build; the helper's rule is not its.
			if err := Check("feov-record", unstripped); err != nil {
				t.Errorf("Check(feov-record, unstripped %s) = %v; want it passed", goos, err)
			}
		})
	}
}
