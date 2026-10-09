package surface

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// EVERY BINARY THAT LINKS THE RECORD LINKS THE REPORT RENDERER.
//
// record places each quote gap in the report as it stands — the board, the work list and the mint
// budget all render it — through a renderer internal/reportproj REGISTERS at init, because record
// cannot import the package that imports it. Registration by linkage makes the answer a property
// of the binary: feov-sitting-write linked the record and not the renderer, so every work list the
// SubagentStart hook delivered read its quote gaps `unrendered` while feov-record read them
// `marked` off the same record.
//
// THE SET IS DERIVED, cmd/ and devcmd/ as they stand on disk, so a binary added later is asked the
// same question without anyone listing it.
//
// A GUARD AND NOT A GENERATED CARRIER, and why: what would remove the registration is record
// importing the renderer, which means reportproj giving up its three record-facing functions and
// the twenty-seven call sites that reach them.
func TestEveryBinaryThatLinksTheRecordLinksTheReportRenderer(t *testing.T) {
	dir, err := repotree.Plugin("tools")
	if err != nil {
		t.Fatalf("locating the tools module: %v", err)
	}
	asked := 0
	for _, parent := range []string{"cmd", "devcmd"} {
		entries, err := os.ReadDir(filepath.Join(dir, parent))
		if err != nil {
			t.Fatalf("listing %s: %v", parent, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			bin := "./" + parent + "/" + e.Name()
			deps := hookDeps(t, bin)
			if !slices.Contains(deps, modulePath+"/internal/record") {
				continue
			}
			asked++
			if !slices.Contains(deps, modulePath+"/internal/reportproj") {
				t.Errorf("%s links internal/record and not internal/reportproj: every view it renders reads "+
					"each quote gap `unrendered`, and a mint it carries is refused for want of a renderer. "+
					"Import internal/reportproj in the package that renders.", bin)
			}
		}
	}
	// feov-record and feov-sitting-write at least. Fewer means the walk measured nothing.
	if asked < 2 {
		t.Fatalf("only %d binary(ies) link the record — the set was not measured", asked)
	}
}
