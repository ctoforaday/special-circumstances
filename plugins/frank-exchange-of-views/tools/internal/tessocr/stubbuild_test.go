package tessocr

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// The default-build contract as a test: a plain, no-tags `go build` of this package MUST
// succeed on any machine — no C stack, no cgo env, no CGO_ENABLED juggling — because the
// engine is opt-in behind `-tags tessocr` and everything else resolves to the stub.
// Running it from the tagged engine suite proves the configuration that suite never
// otherwise compiles; running it from the stub suite proves a clean checkout.
func TestDefaultBuildCompiles(t *testing.T) {
	goTool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if _, err := os.Stat(goTool); err != nil {
		// Fall back to PATH; a test binary can outlive its GOROOT.
		var lookErr error
		goTool, lookErr = exec.LookPath("go")
		if lookErr != nil {
			t.Skipf("no go tool found (GOROOT stat: %v, PATH: %v)", err, lookErr)
		}
	}
	cmd := exec.Command(goTool, "build", ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("plain `go build` failed — the stub no longer covers the default build:\n%s\n%v", out, err)
	}
}
