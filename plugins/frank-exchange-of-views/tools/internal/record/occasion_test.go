package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE REFUSAL RUNS BOTH WAYS, and the second direction is the one that makes this a fact.
//
// A field only the bench is asked for, and which nobody checks anyone else against, is a field that
// can be asserted anywhere and means nothing. So: the bench MUST name its occasion, and every other
// seat is refused for naming one — their id already answers it, and a second answer on the record is
// a fact with two authors.
func TestOnlyTheBenchCarriesAnOccasionAndItMustCarryOne(t *testing.T) {
	for _, c := range []struct {
		name, seat, occasion, wantErr string
	}{
		{"the bench names its sitting", "judge", "docket", ""},
		{"and each of the four is a word it may use", "judge", "assemble", ""},
		{"the bench must name one", "judge", "", "registers with no --occasion"},
		{"a word outside the vocabulary is refused", "judge", "hearing", "is not an occasion"},
		{"a lens may not name one", "red-lens-evidence", "docket", "only the bench carries one"},
		{"nor may the chair", "red-chair", "terminal", "only the bench carries one"},
		{"and a lens registering without one is the ordinary case", "red-lens-evidence", "", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := recordtest.TmpRun(t)
			id := Identity{Run: mustRun(t, runDir), SeatID: c.seat}
			_, _, err := RegisterSeat(id, "", c.occasion)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("register %s with occasion %q = %v, want accepted", c.seat, c.occasion, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("register %s with occasion %q was ACCEPTED; a field nothing refuses is a field nothing can trust", c.seat, c.occasion)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("the refusal does not say why:\n got: %v\nwant it to contain: %q", err, c.wantErr)
			}
		})
	}
}

// THE REFUSAL NAMES THE WHOLE VOCABULARY, because the seat reading it is mid-sitting and the
// alternative is another round trip to --help for a four-word list.
func TestAnOccasionRefusalNamesEveryWordTheBenchMayUse(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	_, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge"}, "", "")
	if err == nil {
		t.Fatal("a bench register with no occasion was accepted")
	}
	words := OccasionWords()
	if len(words) != 4 {
		t.Errorf("the vocabulary has %d words, want 4 — if a sitting kind was added or removed, this test and debate.js both owe it: %v", len(words), words)
	}
	for _, w := range words {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not offer %q, so a seat reading it cannot act on it: %v", w, err)
		}
	}
}

// AND THE WORD REACHES THE RECORD, not merely the validator. A field that refuses correctly and
// stores nothing reads downstream exactly like a question nobody was asked.
func TestTheOccasionIsOnTheRegisterEvent(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge"}, "", "assemble"); err != nil {
		t.Fatal(err)
	}
	m, err := MergedEvents(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, e := range m.Events {
		b, ok := recordpb.BodyAs[*recordpb.Register](e)
		if !ok {
			continue
		}
		found = true
		if got := recordpb.Word(b.GetOccasion()); got != "assemble" {
			t.Errorf("the register stored occasion %q, want assemble", got)
		}
	}
	if !found {
		t.Fatal("no register event on the record at all")
	}
}
