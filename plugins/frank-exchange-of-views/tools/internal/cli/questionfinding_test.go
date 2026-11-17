package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A FINDING ABOUT THE QUESTION REACHES THE SEAT THAT CAN ACT ON IT, as a finding about an inquiry does.
//
// About a part, it reaches the logic lens, on its work list's item for the part's check while that
// check is owed, so the check is not blind to it; blue's list carries none — a gap about the part is
// what reaches blue. About the question as a whole, it names a part of the question no part covers,
// and only blue splits one out: it reaches blue-respond, on its question item, until blue splits again.
func TestAFindingAboutAPartReachesTheLogicLensAndOneAboutTheQuestionReachesBlue(t *testing.T) {
	runDir := seatRun(t)
	logic := record.QuestionPartOwner()
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:question",
		&recordpb.Cast{SeatIds: []string{"red-lens-evidence", logic, "red-chair", "blue-synthesize", "blue-respond", "judge"}}))
	registerSeatOnce(t, runDir, logic)
	work := func(seat string) string {
		t.Helper()
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "--json", "work")
		if err != nil {
			t.Fatalf("%s's work list: %v", seat, err)
		}
		return out
	}
	fileAs := func(seat, key, reason string, about ...string) string {
		t.Helper()
		args := append([]string{"finding", "--seat-id", seat, "--key", key, "--reason", reason,
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--json", "--about-kind", "question"}, about...)
		var filed struct {
			Result struct {
				ID string `json:"finding_id"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(must(t, runDir, args...)), &filed); err != nil || filed.Result.ID == "" {
			t.Fatalf("no finding id on the filing's envelope: %v", err)
		}
		return filed.Result.ID
	}
	file := func(key, reason string, about ...string) string { return fileAs(lensSeat, key, reason, about...) }
	note := func(id string) string { return id + " (" + lensSeat + ")" }

	// UNSPLIT: the whole-question finding is on blue-respond's split item, and on no lens's list.
	early := file("FQ0", "the question asks about cost as well as accuracy")
	if w := work("blue-respond"); !strings.Contains(w, string(record.BlockerQuestionUnsplit)) || !strings.Contains(w, note(early)) {
		t.Errorf("blue-respond's split item does not carry the finding %s about the question as a whole:\n%s", early, w)
	}
	if w := work(logic); strings.Contains(w, early) {
		t.Errorf("a finding about the question as a whole is on %s's work list; only blue splits the question:\n%s", logic, w)
	}

	// SPLIT: a split answers the findings before it; one filed since is on the open part's item.
	part := splitPart(t, runDir)
	late := file("FQ1", "the split still leaves cost out")
	w := work("blue-respond")
	if strings.Contains(w, early) {
		t.Errorf("a finding about the question as a whole filed before the split is still on blue-respond's list:\n%s", w)
	}
	if !strings.Contains(w, part) || !strings.Contains(w, note(late)) {
		t.Errorf("blue-respond's open part item does not carry the finding %s filed since the split:\n%s", late, w)
	}
	if w := work("blue-synthesize"); strings.Contains(w, late) {
		t.Errorf("a finding about the question as a whole is on blue-synthesize's list; blue-respond is the seat that splits it again:\n%s", w)
	}
	// A finding about the part before its coverage is one the coverage answers: its check is not owed it.
	before := file("FQ1b", "the part has no account yet", "--about", part)

	// COVERED: the part's check is owed, and a finding about the part since its coverage is on the
	// logic lens's item for it — listed in a sitting opened after the coverage — and on no list of blue's.
	must(t, runDir, "question", "cover", "--seat-id", "blue-respond", "--id", part, "--as", "covered", "--reason", "the report answers it")
	about := file("FQ2", "the coverage rests on one section that answers half the part", "--about", part)
	// The owner's own finding is its own judgement in hand, as an inquiry's is: its item does not list it.
	own := fileAs(logic, "FQ3", "the owner's own reading of the coverage", "--about", part)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", logic); err != nil {
		t.Fatalf("register %s: %v", logic, err)
	}
	if w := work(logic); !strings.Contains(w, string(record.BlockerUncheckedPart)) || !strings.Contains(w, note(about)) {
		t.Errorf("%s's work list does not carry part %s's check with the finding %s about it:\n%s", logic, part, about, w)
	}
	if w := work(logic); strings.Contains(w, before) || strings.Contains(w, own) {
		t.Errorf("%s's item for part %s lists a finding filed before its coverage (%s) or its own (%s):\n%s", logic, part, before, own, w)
	}
	if w := work("blue-respond"); strings.Contains(w, about) {
		t.Errorf("a finding about a question part is on blue's work list; the gap about the part is what reaches blue:\n%s", w)
	}
}
