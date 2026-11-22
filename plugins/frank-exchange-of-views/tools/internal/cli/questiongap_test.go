package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A GAP ABOUT THE QUESTION REACHES THE SEAT THAT CAN ACT ON IT, as a gap about an inquiry does.
//
// About a part, it is listed on the logic lens's item for the part's check while that check is owed,
// so the check is not blind to it. About the question as a whole, it names a part of the question no
// part covers, and only blue splits one out: it is listed on blue-respond's question item, until blue
// splits again. Each gap is also on its own seat's list and, material, on blue's; these lists are the
// question's items, and the assertions read those items alone.
func TestAGapAboutAPartReachesTheLogicLensAndOneAboutTheQuestionReachesBlue(t *testing.T) {
	runDir := seatRun(t)
	logic := record.QuestionPartOwner()
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:question",
		&recordpb.Cast{SeatIds: []string{"red-lens-evidence", logic, "red-chair", "blue-synthesize", "blue-respond", "judge"}}))
	registerSeatOnce(t, runDir, logic)
	// items is the What of every item of kind on seat's work list, joined.
	items := func(seat string, kind record.BlockerKind) string {
		t.Helper()
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "--json", "work")
		if err != nil {
			t.Fatalf("%s's work list: %v", seat, err)
		}
		var w struct {
			Sitting struct {
				Open []record.Item `json:"open"`
			} `json:"sitting"`
		}
		if err := json.Unmarshal([]byte(out), &w); err != nil {
			t.Fatalf("%s's work list is not JSON: %v\n%s", seat, err, out)
		}
		var whats []string
		for _, it := range w.Sitting.Open {
			if it.Kind == record.ItemKind(kind) {
				whats = append(whats, it.What)
			}
		}
		return strings.Join(whats, "\n")
	}
	// mintAs mints a gap about the question — the whole of it with no part, else the part — at a
	// grade below material where low, so a part it names stays unchecked rather than wanting.
	mintAs := func(seat, key, problem, grade, class string, about ...string) string {
		t.Helper()
		args := append([]string{"mint", "--seat-id", seat, "--key", key, "--class", class, "--problem", problem, "--fix", "f",
			"--check-kind", "document", "--check", "c", "--severity", grade, "--likelihood", grade, "--impact", grade, "--complexity", "low",
			"--reason", "r", "--json", "--about-kind", "question"}, about...)
		var minted struct {
			Result struct {
				ID string `json:"gap_id"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(must(t, runDir, args...)), &minted); err != nil || minted.Result.ID == "" {
			t.Fatalf("no gap id on the mint's envelope: %v", err)
		}
		return minted.Result.ID
	}
	whole := func(key, problem string) string {
		return mintAs(lensSeat, key, problem, "medium", "coverage_misstated")
	}
	note := func(id, seat string) string { return id + " (" + seat + ")" }

	// UNSPLIT: the whole-question gap is on blue-respond's split item, and on no lens's question item.
	early := whole("GQ0", "the question asks about cost as well as accuracy")
	if w := items("blue-respond", record.BlockerQuestionUnsplit); !strings.Contains(w, note(early, lensSeat)) {
		t.Errorf("blue-respond's split item does not carry the gap %s about the question as a whole:\n%s", early, w)
	}

	// SPLIT: a split answers the gaps before it; one minted since is on the open part's item.
	part := splitPart(t, runDir)
	late := whole("GQ1", "the split still leaves cost out")
	w := items("blue-respond", record.BlockerOpenPart)
	if strings.Contains(w, early) {
		t.Errorf("a gap about the question as a whole minted before the split is still on blue-respond's question item:\n%s", w)
	}
	if !strings.Contains(w, part) || !strings.Contains(w, note(late, lensSeat)) {
		t.Errorf("blue-respond's open part item does not carry the gap %s minted since the split:\n%s", late, w)
	}
	if w := items("blue-synthesize", record.BlockerOpenPart); strings.Contains(w, late) {
		t.Errorf("a gap about the question as a whole is on blue-synthesize's question item; blue-respond is the seat that splits it again:\n%s", w)
	}

	// COVERED: the part's check is owed, and every gap minted about the part while no check stands —
	// the owner's own among them, as an inquiry's are — is on the logic lens's item for it, listed in a
	// sitting opened after the coverage. Below material, so the part stays unchecked, not wanting.
	before := mintAs(lensSeat, "GQ1b", "the part has no account yet", "low", "spec_underspecification", "--about", part)
	must(t, runDir, "question", "cover", "--seat-id", "blue-respond", "--id", part, "--as", "covered", "--reason", "the report answers it")
	own := mintAs(logic, "GQ3", "the owner's own reading of the coverage", "low", "reader_channel_mismatch", "--about", part)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", logic); err != nil {
		t.Fatalf("register %s: %v", logic, err)
	}
	if w := items(logic, record.BlockerUncheckedPart); !strings.Contains(w, part) || !strings.Contains(w, note(before, lensSeat)) || !strings.Contains(w, note(own, logic)) {
		t.Errorf("%s's item for part %s's check does not carry the gaps %s and %s minted about it while it stood unchecked:\n%s", logic, part, before, own, w)
	}
	if w := items(logic, record.BlockerUncheckedPart); strings.Contains(w, early) || strings.Contains(w, late) {
		t.Errorf("%s's item for part %s lists a gap about the question as a whole:\n%s", logic, part, w)
	}

	// CHECKED, THEN QUESTIONED AGAIN: a gap minted after the check puts it in question and is listed;
	// the ones the check had in hand are not.
	must(t, runDir, "question", "check", "--seat-id", logic, "--id", part)
	after := mintAs(lensSeat, "GQ4", "a second reading finds the coverage thin", "low", "partial_control_coverage", "--about", part)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", logic); err != nil {
		t.Fatalf("register %s: %v", logic, err)
	}
	w = items(logic, record.BlockerUncheckedPart)
	if !strings.Contains(w, note(after, lensSeat)) {
		t.Errorf("%s's item for part %s's check does not carry the gap %s minted since its last check:\n%s", logic, part, after, w)
	}
	if strings.Contains(w, before) || strings.Contains(w, own) {
		t.Errorf("%s's item for part %s lists a gap its last check had in hand (%s or %s):\n%s", logic, part, before, own, w)
	}
}
