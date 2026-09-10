package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE UNKNOWN-CLASS REFUSAL NAMES A VERB THE SEAT CAN TYPE, AND THIS RUNS IT.
//
// It is the more common of the two class refusals: a lens meets it the first time it reaches for a
// class the registry lacks, which is how the voice lens arrived at coining in the #861 rerun. It said
// `merge class new --class …` long after the role level was gone, and the fix to its sibling
// (`class new`'s confirmation and the duplicate-coin refusal, TestClassNewCoinsTheSlugInClass) left
// it behind — a grep of the tree read through a capped listing reported it absent. So the
// instruction is not merely asserted here: the verb it names is RUN, through the real tree, and the
// mint the seat was attempting then goes through. A refusal that names something the tree refuses
// fails this test, whatever the wording.
func TestTheUnknownClassRefusalNamesAVerbThatRuns(t *testing.T) {
	runDir := newRun(t)
	registerLensOnce(t, runDir)
	mintArgs := []string{"mint", "--run", runDir, "--seat-id", lensSeat,
		"--class", "never-registered", "--check-kind", "document", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "p"}

	// The field's own requirement text — record.proto's `why` on Mint.class — is the refusal a seat
	// meets when it passes --class EMPTY. (With the flag absent, cobra's required-flag check refuses
	// first and the record never speaks; an explicitly empty value gets past cobra to the record's
	// presence check, which renders the `why` verbatim.) It names the coining verb too.
	emptyClass := append([]string{}, mintArgs...)
	emptyClass[6] = ""
	if _, err := run(t, emptyClass...); err == nil || !strings.Contains(err.Error(), "`class new`") {
		t.Errorf("a mint with an empty --class must be refused naming the coining verb `class new`, got: %v", err)
	}

	_, err := run(t, mintArgs...)
	if err == nil {
		t.Fatal("a mint against a class neither registered nor coined was accepted")
	}
	const want = "`class new --class never-registered"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("the refusal must name the coining verb a seat can type (%s…):\n%v", want, err)
	}

	// THE INSTRUCTION, RUN AS WRITTEN — the verb and the flag the refusal named.
	if _, err := run(t, "class", "new", "--run", runDir, "--seat-id", lensSeat,
		"--class", "never-registered", "--definition", "d", "--neighbor", "x", "--distinguisher", "q"); err != nil {
		t.Fatalf("the verb the refusal named does not run: %v", err)
	}
	if _, err := run(t, mintArgs...); err != nil {
		t.Fatalf("after following the refusal's instruction, the mint still fails: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Mint{}).GetClass(); got != "never-registered" {
		t.Errorf("minted class = %q, want the slug the seat was told to coin", got)
	}
}
