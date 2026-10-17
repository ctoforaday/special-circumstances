package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE WORK LIST SAYS WHETHER BLUE ANSWERED, PER GAP (#1122).
//
// Four arms, and the third is the one the whole function exists for: "blue has not answered" and
// "blue is still sitting and may yet answer" are the same silence on the record, so a lens handed
// the first where the second is true argues a dispute that is still in progress.
//
// DRIVEN AS A TABLE OVER blueAnswers, with the record shapes built by the helpers the rest of this
// package uses. Each row asserts the arm it wants AND that no other arm's wording appears — an
// item that fires twice is as wrong as one that never fires.
func TestTheListSaysWhetherBlueAnsweredThisGap(t *testing.T) {
	const lens = "red-lens-evidence"
	for _, tc := range []struct {
		name      string
		answered  bool
		engaged   bool
		unresolve bool
		want      string // the marker for the arm that must fire, or "" for silence
		absent    []string
	}{
		{name: "answered", answered: true, engaged: true, want: "has been ANSWERED",
			absent: []string{"NOT MEASURED", "NO edit answers it"}},
		{name: "engaged and blue's sitting closed with no answer", engaged: true, want: "NO edit answers it",
			absent: []string{"NOT MEASURED", "has been ANSWERED"}},
		{name: "engaged and blue is still sitting", engaged: true, unresolve: true, want: "NOT MEASURED",
			absent: []string{"NO edit answers it", "has been ANSWERED"}},
		{name: "blue never engaged", want: "",
			absent: []string{"NOT MEASURED", "NO edit answers it", "has been ANSWERED"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(blueAnswersT(t,
				answerScenario(t, lens, tc.engaged, tc.answered, tc.unresolve),
				[]WorkGapState{{ID: "G1", Open: true}}, lens), "\n")
			if tc.want == "" {
				if got != "" {
					t.Fatalf("want silence, got:\n%s", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("the %s arm did not fire; got:\n%s", tc.name, got)
			}
			for _, no := range tc.absent {
				if strings.Contains(got, no) {
					t.Errorf("another arm fired too (%q):\n%s", no, got)
				}
			}
		})
	}
}

// AN ANSWER TO SOMEBODY ELSE'S GAP IS NOT AN ANSWER TO THIS ONE, and nor is an edit that replaced a
// span with itself. Both would read as "blue answered" under a looser join, and the second is the
// one exchangesOf already refuses to count as movement.
func TestOnlyASubstantiveEditOnThisGapCounts(t *testing.T) {
	const lens = "red-lens-evidence"
	gaps := []WorkGapState{{ID: "G1", Open: true}}

	noop := blueAnswersT(t, answerScenarioEdit(t, lens, "G1", "same", "same"), gaps, lens)
	if len(noop) != 1 || !strings.Contains(noop[0], "NO edit answers it") {
		t.Errorf("an edit that replaced a span with itself counted as an answer: %v", noop)
	}
	other := blueAnswersT(t, answerScenarioEdit(t, lens, "G2", "old", "new"), gaps, lens)
	if len(other) != 1 || !strings.Contains(other[0], "NO edit answers it") {
		t.Errorf("an edit answering another gap counted as an answer to this one: %v", other)
	}
}

// ONLY THE MINTING SEAT IS TOLD. A gap is its originator's from mint to close — only that seat
// re-audits and closes it — so this is news for one seat and noise for every other.
func TestOnlyTheMintingLensIsToldAboutItsGap(t *testing.T) {
	evs := answerScenario(t, "red-lens-evidence", true, true, false)
	gaps := []WorkGapState{{ID: "G1", Open: true}}
	if got := blueAnswersT(t, evs, gaps, "red-lens-logic"); len(got) != 0 {
		t.Errorf("a lens was told about a gap it did not mint: %v", got)
	}
}

// A CLOSED GAP IS NOT PENDING WORK. The work list is the OPEN work, and a closed gap's answer is
// already adjudicated.
func TestAClosedGapIsNotReported(t *testing.T) {
	const lens = "red-lens-evidence"
	evs := answerScenario(t, lens, true, false, false)
	if got := blueAnswersT(t, evs, []WorkGapState{{ID: "G1", Open: false}}, lens); len(got) != 0 {
		t.Errorf("a closed gap was reported as pending: %v", got)
	}
}

// answerScenario builds the record shape each arm needs: the lens's mint, and optionally blue
// dispatched onto the gap, blue's register, its answering edit, and its agent's return.
//
// THE EDIT SETS old AND new EXPLICITLY, so editAnswers is not reused. That helper leaves both empty,
// which makes old == new — and this predicate treats a span replaced by itself as an act and not an
// answer, matching exchangesOf. ManifestOwed asks a different question (was a manifest row owed) and
// counts any edit against the gap, so the two helpers are not interchangeable here.
func answerScenario(t *testing.T, lens string, engaged, answered, unresolved bool) []*Event {
	t.Helper()
	evs := []*Event{mintsGap(t, lens, "G1")}
	if !engaged {
		return evs
	}
	evs = append(evs, dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"))
	if answered {
		evs = append(evs, blueEditOf(t, "G1", "old", "new"))
	}
	if !unresolved {
		// The agent's return is what closes the sitting, and it must follow the edit — an act after
		// the stop is outside the sitting.
		evs = append(evs, agentStops(t, "blue-a"))
	}
	return evs
}

// answerScenarioEdit is answerScenario's engaged-and-closed case with the edit spelled out, for the
// two shapes that must NOT read as an answer.
func answerScenarioEdit(t *testing.T, lens, answers, old, new string) []*Event {
	t.Helper()
	return []*Event{
		mintsGap(t, lens, "G1"),
		dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
		blueEditOf(t, answers, old, new),
		agentStops(t, "blue-a"),
	}
}

func mintsGap(t *testing.T, lens, gap string) *Event {
	t.Helper()
	return recordtest.Event(t, lens, corrMint(gap))
}

func blueEditOf(t *testing.T, answers, old, new string) *Event {
	t.Helper()
	return recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{
		Answers: proto.String(answers), Old: proto.String(old), New: proto.String(new)})
}

// A GAP ANCHORED ON AN INQUIRY IS ANSWERED BY A FATE ON THAT INQUIRY. No edit can answer a gap that
// names no report text, so the list names the act that does: a fate blue records on the inquiry the
// gap is anchored on, while engaged on it. A fate on another inquiry answers nothing, and a sitting
// that recorded none is reported as missing the inquiry record — never as missing an edit.
func TestAGapAnchoredOnAnInquiryIsAnsweredByItsFate(t *testing.T) {
	const lens = "red-lens-dark-side"
	gaps := []WorkGapState{{ID: "G1", Open: true, AboutKind: "inquiry", AboutRef: "Q-1"}}
	scenario := func(fatedOn string) []*Event {
		evs := []*Event{
			recordtest.Event(t, "frontier", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q"),
				Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED)}),
			mintsGap(t, lens, "G1"), dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
		}
		if fatedOn != "" {
			evs = append(evs, recordtest.Event(t, "blue-respond", &recordpb.Inquiry{InquiryId: proto.String(fatedOn),
				Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED), Reason: proto.String("out of reach this run")}))
		}
		return append(evs, agentStops(t, "blue-a"))
	}
	for _, tc := range []struct {
		name, fatedOn, want string
		absent              []string
	}{
		{"blue recorded the inquiry's fate", "Q-1", "has been ANSWERED by an inquiry record — blue recorded the fate `deferred` on Q-1",
			[]string{"NO edit answers it", "NO inquiry record answers it"}},
		{"blue recorded another inquiry's fate", "Q-2", "NO inquiry record answers it — it is anchored on Q-1",
			[]string{"NO edit answers it", "has been ANSWERED"}},
		{"blue recorded no fate", "", "NO inquiry record answers it — it is anchored on Q-1",
			[]string{"NO edit answers it", "has been ANSWERED"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(blueAnswersT(t, scenario(tc.fatedOn), gaps, lens), "\n")
			if !strings.Contains(got, tc.want) {
				t.Errorf("want %q; got:\n%s", tc.want, got)
			}
			for _, no := range tc.absent {
				if strings.Contains(got, no) {
					t.Errorf("another arm fired too (%q):\n%s", no, got)
				}
			}
		})
	}
}
