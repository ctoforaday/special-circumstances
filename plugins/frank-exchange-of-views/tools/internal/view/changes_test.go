package view

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)


// seedChanges lays down one minted gap and three edits: two answering it, one answering
// nothing (blue's own work — 6 of the smoke's 26 edits were exactly that shape).
func seedChanges(t *testing.T, runDir string) {
	t.Helper()
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
			GapId:           proto.String("R1-1"),
			Class:           proto.String("overclaim"),
			Problem:         proto.String("independence is overclaimed"),
			RequiredFix:     proto.String("acknowledge the shared definition"),
			AcceptanceCheck: proto.String("the section no longer claims independence"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
	})
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, "blue-respond-r1", 1, "blue-respond-r1:blue_edit:e1", &recordpb.BlueEdit{
			Answers: proto.String("R1-1"),
			Old:     proto.String("five independent approaches"),
			New:     proto.String("five approaches"),
			Text:    proto.String("drop the independence claim"),
		}),
		recordtest.At(t, "blue-respond-r1", 1, "blue-respond-r1:blue_edit:e2", &recordpb.BlueEdit{
			Answers: proto.String("R1-1"),
			Old:     proto.String("They agree."),
			New:     proto.String("They agree, sharing one definition of primality."),
			Text:    proto.String("name the shared definition"),
		}),
		// The unattributed edit: blue's own work, answering no gap. Six of the smoke's 26 edits
		// were this shape, which is why `answers` being ABSENT has to stay expressible.
		recordtest.At(t, "blue-respond-r1", 1, "blue-respond-r1:blue_edit:e3", &recordpb.BlueEdit{
			Old:  proto.String("cost ,"),
			New:  proto.String("cost,"),
			Text: proto.String("punctuation repair"),
		}),
	})
}

// The unscoped form is the navigation hint: every edit, attributed or not.
func TestChangesListsEveryEditAndSaysWhichAnswerNoGap(t *testing.T) {
	runDir := t.TempDir()
	seedChanges(t, runDir)

	out := md(t, runDir, "changes")
	for _, want := range []string{
		"drop the independence claim", "name the shared definition", "punctuation repair",
		"answers **R1-1**", "no gap", "3 recorded edit(s)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("unscoped changes is missing %q:\n%s", want, out)
		}
	}
}

// The scoped form is the COMPARISON #267 asks for: required_fix and the actual change,
// side by side, instead of red inferring from a re-read whether a gap was answered.
func TestChangesScopedPutsRequiredFixBesideTheEdits(t *testing.T) {
	runDir := t.TempDir()
	seedChanges(t, runDir)

	b, err := Markdown(runDir, "changes", "R1-1")
	if err != nil {
		t.Fatalf("scoped changes: %v", err)
	}
	out := string(b)
	for _, want := range []string{
		"acknowledge the shared definition",         // what red required
		"the section no longer claims independence", // the acceptance check
		"independence is overclaimed",               // the defect
		"- five independent approaches",             // blue's actual old span
		"+ five approaches",                         // and its replacement
		"Edits blue recorded against it (2)",        // both, and only both
	} {
		if !strings.Contains(out, want) {
			t.Errorf("scoped changes is missing %q:\n%s", want, out)
		}
	}
	// The unattributed edit belongs to the unscoped list, not to this gap's comparison.
	if strings.Contains(out, "punctuation repair") {
		t.Error("an edit answering no gap leaked into a gap-scoped comparison")
	}
}

// A gap red asked about and blue never answered is the interesting case, so it is STATED
// — including that an edit which omitted --answers would be invisible here.
func TestChangesScopedSaysNoneRatherThanRenderingEmpty(t *testing.T) {
	runDir := t.TempDir()
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
			GapId:           proto.String("R1-1"),
			Class:           proto.String("x"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("c"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
	})

	b, err := Markdown(runDir, "changes", "R1-1")
	if err != nil {
		t.Fatalf("scoped changes: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "NONE") {
		t.Errorf("a gap with no edits did not say so:\n%s", out)
	}
	if !strings.Contains(out, "--answers` is invisible here") {
		t.Errorf("the view overclaims: an unattributed edit is invisible to it and it must say so:\n%s", out)
	}
}

// An id no mint created gets a refusal, not an empty comparison — the same discipline
// requireGap applies at the write side.
func TestChangesScopedRefusesAnUnknownGap(t *testing.T) {
	runDir := t.TempDir()
	seedChanges(t, runDir)

	if _, err := Markdown(runDir, "changes", "R9-99"); err == nil {
		t.Fatal("a view scoped to a gap nobody minted rendered a comparison anyway")
	} else if !strings.Contains(err.Error(), "R9-99") {
		t.Errorf("the refusal must name the id: %v", err)
	}
}

// An empty stack is a FINDING on a run past round 0, not blank formatting.
func TestChangesSaysWhyItIsEmpty(t *testing.T) {
	out := md(t, t.TempDir(), "changes")
	if !strings.Contains(out, "itself a finding") {
		t.Errorf("an empty change stack rendered as if there were simply nothing to show:\n%s", out)
	}
}

// The size delta is in CHARACTERS. A multi-byte character must not read as an edit.
func TestDeltaCountsRunesNotBytes(t *testing.T) {
	if got := delta("e", "é"); got != "same length" {
		t.Errorf("delta(e -> é) = %q, want %q — byte counting makes an accent look like +1", got, "same length")
	}
	if got := delta("ab", "abcd"); got != "+2 chars" {
		t.Errorf("delta = %q, want +2 chars", got)
	}
	if got := delta("abcd", "ab"); got != "-2 chars" {
		t.Errorf("delta = %q, want -2 chars", got)
	}
}
