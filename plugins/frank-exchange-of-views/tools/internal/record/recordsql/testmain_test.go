package recordsql

import (
	"fmt"
	"os"
	"testing"
)

// The orphaned-handle check, inline rather than through recordtest.Main.
//
// recordtest imports THIS package, so this package's tests cannot import recordtest — the same
// cycle that left the eighth copy of the TmpRun body living here, calling CloseUnder unqualified.
// The cycle is real and the check is not: OrphanedHandles is declared right here, so the package
// that owns the cache is the one place that needs no helper to read it (#666).
func TestMain(m *testing.M) {
	code := m.Run()
	if orphans := OrphanedHandles(); len(orphans) > 0 {
		fmt.Fprintf(os.Stderr, "recordsql: %d record handle(s) outlived the directory holding them:\n", len(orphans))
		for _, p := range orphans {
			fmt.Fprintln(os.Stderr, " ", p)
		}
		fmt.Fprintln(os.Stderr, "\nA test took its run directory from t.TempDir() without releasing the cached handle.\n"+
			"On Linux the removal succeeds and the test passes; on Windows it fails the cleanup.\n"+
			"Release it with CloseUnder(dir) in a t.Cleanup — which is what recordtest.TmpRun does.")
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
