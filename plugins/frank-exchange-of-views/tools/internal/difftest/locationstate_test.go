package difftest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY LOCATION STATE IS TAUGHT WHERE IT IS READ. The work and board help pages name each
// location_state value, read from its constant, and the engaged lens prompt carries the `gone`
// sentence word for word: a cut sentence is an answer, and a prompt that says only "an empty list
// means unanswered" would teach the opposite.
func TestWorkTeachesEveryLocationState(t *testing.T) {
	bin := buildBinary(t)
	for _, view := range []string{"work", "board"} {
		inv := capture(command(bin, "show", view, "--seat-id", "red-lens-evidence", "--help"))
		page := inv.stdout + inv.stderr
		for _, state := range []string{record.LocationMarked, record.LocationGone, record.LocationUnrendered} {
			if !strings.Contains(page, "`"+state+"`") {
				t.Errorf("the %s help does not name the location state %q:\n%s", view, state, page)
			}
		}
		if !strings.Contains(page, record.GoneTeaching) {
			t.Errorf("the %s help does not teach the gone state in its words", view)
		}
	}
	golden, err := os.ReadFile(filepath.Join(repoRoot(t), "plugins", "frank-exchange-of-views", "tests", "simulator", "testdata", "prompt-red-lens-evidence-engaged.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(golden), record.GoneTeaching) {
		t.Errorf("the engaged lens prompt does not carry the gone sentence word for word: %q", record.GoneTeaching)
	}
}
