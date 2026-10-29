package seat

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type returnsProbe struct {
	ID string `json:"probe_id"`
}

func (returnsProbe) Human() string { return "probe" }

// A VERB'S PAGE STATES ITS --json SUCCESS FROM THE TYPE IT RETURNS, AND THE RESULT IS HELD TO IT. The
// smoke's voice lens read `.result.id` off `finding --json` and got null: the key is `finding_id`,
// and no write page said so (#1601 smoke, ev117). Returns writes the tree from the type; Emit refuses
// a handler that returns another, so the page cannot describe a shape the wire does not carry.
func TestAVerbsPageStatesTheResultItReturnsAndEmitHoldsItToIt(t *testing.T) {
	c := Returns(&cobra.Command{Use: "probe"}, returnsProbe{})
	if !strings.Contains(c.Long, "OUTPUT (JSON, with --json — the bare call is the human form): {verb,role,ok,result:{probe_id},standing:{") {
		t.Errorf("the page does not state the result's keys:\n%s", c.Long)
	}
	var out bytes.Buffer
	c.SetOut(&out)
	if err := Emit(c, returnsProbe{ID: "P1"}, nil); err != nil {
		t.Fatalf("the declared result was refused: %v", err)
	}
	if !strings.Contains(out.String(), "probe") {
		t.Errorf("the declared result did not print: %q", out.String())
	}
	if err := Emit(c, Msg{Message: "another shape"}, nil); err == nil || !strings.Contains(err.Error(), "its page states the --json result of a seat.returnsProbe") {
		t.Errorf("a result of another type = %v, want it refused naming the type the page states", err)
	}
}
