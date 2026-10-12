package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE MIGRATE VERB SAYS WHAT BECAME OF THE ORDER OF THE REPORT'S TEXT, to the operator: brought
// into the order it ships in, or kept as written because its acts do not replay over the
// reordered text.
func TestMigrateSaysWhatBecameOfTheReportsOrder(t *testing.T) {
	source := func(edit [2]string) string {
		runDir := newRun(t)
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}
		if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
			t.Fatalf("register: %v", err)
		}
		if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(orderHead + orderOwn1 + orderCat + orderOpen)}); err != nil {
			t.Fatal(err)
		}
		registerBlue(t, runDir)
		if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "M1", "--quote", edit[0], "--new", edit[1], "--reason", "a change"); err != nil {
			t.Fatalf("edit: %v\n%s", err, out)
		}
		return runDir
	}
	for want, edit := range map[string][2]string{
		"the report's text is brought into the order it ships in": {"91 is 7 x 13", "91 is 7 times 13"},
		// The quote runs from a section of the report's own into the Catechism's heading, and the
		// two are no longer neighbours once the Catechism stands in its place.
		"the report's text is OUT OF THE ORDER IT SHIPS IN and is kept as written": {
			"The quotient is 13<!--proof:P-0000000c-->.\n\n## The Catechism", "The quotient is thirteen<!--proof:P-0000000c-->.\n\n## The Catechism"},
	} {
		out, err := run(t, "migrate", "--seat-id", "operator", "--from", source(edit), "--to", recordtest.TmpRun(t))
		if err != nil {
			t.Fatalf("migrate: %v\n%s", err, out)
		}
		if !strings.Contains(out, want) {
			t.Errorf("migrate does not say %q:\n%s", want, out)
		}
	}
}
