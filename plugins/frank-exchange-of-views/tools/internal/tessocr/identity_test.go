package tessocr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
)

// TestIdentityNeverSpawns: Identity reads compiled constants only. With NATIVE_HELPER_DIR at a
// helper that logs every start it returns the compiled string and starts nothing, and with none it
// returns the same string, so the cache-reuse paths that call it never start a helper.
func TestIdentityNeverSpawns(t *testing.T) {
	dir := t.TempDir()
	nativehelpertest.Fake(t, dir, "feov-ocr", nativehelper.HelperDigest("feov-ocr"), "")
	starts := filepath.Join(t.TempDir(), "starts")
	t.Setenv("FAKEHELPER_START_LOG", starts)
	t.Setenv(nativehelper.EnvDir, dir)
	resetDefaultClient()
	t.Cleanup(resetDefaultClient)

	with := Identity()
	if _, err := os.Stat(starts); !os.IsNotExist(err) {
		b, _ := os.ReadFile(starts)
		t.Fatalf("Identity started the helper (start log: %q)", b)
	}
	t.Setenv(nativehelper.EnvDir, "none")
	if without := Identity(); without != with {
		t.Errorf("Identity is %q with a helper and %q without one", with, without)
	}

	// The control: the fake does log a start when the client spawns it.
	t.Setenv(nativehelper.EnvDir, dir)
	c, err := getClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(starts); err != nil {
		t.Fatalf("the fake logged no start when spawned, so the check above measures nothing: %v", err)
	}
}
