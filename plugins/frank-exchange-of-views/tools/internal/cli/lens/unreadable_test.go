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

func TestAClosureWhoseDemandCannotBeReadIsRefused(t *testing.T) {
	err := requireEvidenceAnswered("G-1", func(string) (record.EvidenceDemand, error) { return record.EvidenceDemand{}, errUnreadable })
	wantRefusalNaming(t, err, "the live-answer demand on G-1")
}

// The healthy answers, so the refusal above is about the failure and not the fixture.
func TestTheDemandGuardOverAReadableRecord(t *testing.T) {
	demand := func(d record.EvidenceDemand) func(string) (record.EvidenceDemand, error) {
		return func(string) (record.EvidenceDemand, error) { return d, nil }
	}
	if err := requireEvidenceAnswered("G-1", demand(record.EvidenceDemand{CheckKind: "document"})); err != nil {
		t.Errorf("a document check was refused: %v", err)
	}
	for _, kind := range []string{"citation", "computation"} {
		if err := requireEvidenceAnswered("G-1", demand(record.EvidenceDemand{AnswerKind: kind, Answered: true})); err != nil {
			t.Errorf("an answered %s demand was refused: %v", kind, err)
		}
		if err := requireEvidenceAnswered("G-1", demand(record.EvidenceDemand{AnswerKind: kind})); err == nil {
			t.Errorf("an unanswered %s demand was admitted", kind)
		}
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
