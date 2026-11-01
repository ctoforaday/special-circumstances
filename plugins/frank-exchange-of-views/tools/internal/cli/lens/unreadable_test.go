package lens

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// AN UNREADABLE RECORD REFUSES THE ACT (gblock, 2026-10-23). These guards read the record after
// the verb has already read it whole, so a fault a test can plant in the record stops the verb
// before it reaches them; the read each guard makes is therefore injected here, and the CLI suite
// (unreadablerefuses_test.go) holds the reads a planted fault does reach.

var errUnreadable = errors.New("database is locked")

func wantRefusalNaming(t *testing.T, err error, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("admitted over a read that failed (%s)", what)
	}
	for _, want := range []string{"could not be read to establish " + what, "nothing is recorded", errUnreadable.Error()} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want %q in the refusal:\n%v", want, err)
		}
	}
	if !errors.Is(err, errUnreadable) {
		t.Errorf("the read error is not wrapped: %v", err)
	}
}

func TestAClosureWhoseCheckKindCannotBeReadIsRefused(t *testing.T) {
	err := requireComputationAnswered("G-1",
		func(string) (bool, error) { return false, errUnreadable },
		func(string) (bool, error) { return true, nil })
	wantRefusalNaming(t, err, "the check kind G-1 was minted with")
}

func TestAClosureWhoseComputationJoinCannotBeReadIsRefused(t *testing.T) {
	err := requireComputationAnswered("G-1",
		func(string) (bool, error) { return true, nil },
		func(string) (bool, error) { return false, errUnreadable })
	wantRefusalNaming(t, err, "whether a computation answers G-1")
}

// The healthy answers, so the refusals above are about the failure and not the fixture.
func TestAComputationGuardOverAReadableRecord(t *testing.T) {
	answered := func(ok bool) func(string) (bool, error) { return func(string) (bool, error) { return ok, nil } }
	if err := requireComputationAnswered("G-1", answered(false), answered(false)); err != nil {
		t.Errorf("a document check was refused: %v", err)
	}
	if err := requireComputationAnswered("G-1", answered(true), answered(true)); err != nil {
		t.Errorf("an answered computation check was refused: %v", err)
	}
	if err := requireComputationAnswered("G-1", answered(true), answered(false)); err == nil {
		t.Error("an unanswered computation check was admitted")
	}
}

type failingFamily struct{}

func (failingFamily) Family() (record.Family, error) { return record.Family{}, errUnreadable }

type emptyFamily struct{}

func (emptyFamily) Family() (record.Family, error) { return record.Family{}, nil }

func TestAMintWhoseEstoppelCannotBeAskedIsRefused(t *testing.T) {
	wantRefusalNaming(t, requireNotEstopped(failingFamily{}, "Seven is prime.", nil),
		"whether this location is text red prescribed")
	if err := requireNotEstopped(emptyFamily{}, "Seven is prime.", nil); err != nil {
		t.Errorf("a mint over a readable record with no prescription was refused: %v", err)
	}
}
