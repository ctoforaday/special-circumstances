package reportproj

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// The replay analog of the verbs' torn-anchor adoption: applying an insertion whose marker is
// ALREADY in the running text is a no-op, not a second splice. A marker can be named twice in the
// record across a crash-retry re-dispatch (cite adopts the orphan and re-appends under the same
// label in a NEW sitting, past the once-per-sitting guard), so replay must place it once. Driven on
// insertMut directly — the once-per-sitting guard blocks constructing the pair through the record,
// but the skip is what makes replay robust when the pair nonetheless exists.
func TestInsertMutSkipsAMarkerAlreadyPlaced(t *testing.T) {
	const location = "Demand grows steadily over the period"
	marker := anchor.Token("C-cafef00d")
	base := location + "."

	once, err := (insertMut{location: location, marker: marker}).apply(base)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if strings.Count(once, marker) != 1 {
		t.Fatalf("first insert did not place the marker exactly once:\n  %q", once)
	}
	twice, err := (insertMut{location: location, marker: marker}).apply(once)
	if err != nil {
		t.Fatalf("second insert: %v", err)
	}
	if twice != once {
		t.Errorf("re-applying an already-placed marker changed the text:\n  once  %q\n  twice %q", once, twice)
	}
}
