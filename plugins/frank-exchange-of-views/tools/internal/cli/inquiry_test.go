package cli

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE UNIT IS THE CHOICE, NOT THE ENTRY. These pin the inquiry's lifecycle: proposed with a
// question and a hypothesis, then given a fate — and given another — as the research moves.

func inquirySeat(t *testing.T, runDir string) string {
	t.Helper()
	const seat = "blue-respond"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil {
		t.Fatal(err)
	}
	return seat
}

// A proposal gets a tool-assigned id and starts undecided: blue owes no fate before it has one.
func TestInquiryProposalIsAssignedAnIDAndStartsProposed(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)

	out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "trial division by hand", "--hypothesis", "if 7 has no divisor in 2..6 it is prime")
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	if !strings.Contains(out, handleText(t, runDir, "Q1")) {
		t.Errorf("the tool must assign the id, not the seat: %q", out)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if got := a.GetStatus(); got != recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED {
		t.Errorf("status = %q, want proposed — a fresh proposal has no fate yet", recordpb.Word(got))
	}
	if a.GetHypothesis() == "" {
		t.Error("the hypothesis was not recorded, so a later abandonment has nothing to be judged against")
	}
}

// RECORDING A FATE IS THE POINT: an inquiry that dies mid-run says so, and keeps its question.
func TestInquiryRecordKeepsItsSubstance(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir) // the path below names r1; the chair sitting is what makes it epoch 1
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "survey primality libraries", "--hypothesis", "implementations disagree at small n"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "abandoned", "--reason", "every implementation agrees at n=7; the hypothesis is dead"); err != nil {
		t.Fatalf("record: %v", err)
	}

	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{handleText(t, runDir, "Q1"), "survey primality libraries", "implementations disagree", "the hypothesis is dead", "e1 proposed -> e1 abandoned"} {
		if !strings.Contains(out, want) {
			t.Errorf("the projection lost %q — the PATH is the evidence of choosing:\n%s", want, out)
		}
	}
}

// A fate recorded with nothing said is an unfalsifiable status.
func TestInquiryRecordRequiresWhatChanged(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "a question"); err != nil {
		t.Fatal(err)
	}
	_, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat, "--id", "Q1", "--as", "abandoned")
	if err == nil {
		t.Fatal("an inquiry slid to abandoned with no stated reason")
	}
	if !strings.Contains(err.Error(), "reason") {
		t.Errorf("the refusal must name what is missing: %v", err)
	}
}

// A dangling reference is refused at the write, like every other (refs.go).
func TestInquiryRecordRefusesAnUnknownID(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	_, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q9", "--as", "underway", "--reason", "why")
	if err == nil {
		t.Fatal("a fate recorded against an inquiry nobody proposed was accepted")
	}
	if !strings.Contains(err.Error(), handleText(t, runDir, "Q9")) {
		t.Errorf("the refusal must name the dangling id: %v", err)
	}
}

// THE AMBIGUITY IS STRUCTURAL, not policed. Proposing and recording a fate are two verbs, and
// `propose` simply has no --id to pass.
func TestProposeHasNoIDToConfuseTheRecordWith(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := runAt(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--question", "a question"); err == nil {
		t.Fatal("`propose --id` was accepted; the two contracts are still reachable through one shape")
	}
}

// THE LOGIC LENS JUDGES A SETTLED INQUIRY, AND NO OTHER SEAT DOES. Driven through the command line,
// end to end: the check is refused while the inquiry is open and while its fate is newer than the
// lens's sitting, recorded once the lens sits after the fate, refused to another lens, recorded once
// however often the same check is sent, refused as a second judgement while it stands, printed on the
// inquiry's row, and stale once blue records the fate again.
func TestTheLogicLensChecksASettledInquiry(t *testing.T) {
	runDir := newRun(t)
	blue := inquirySeat(t, runDir)
	logic := record.InquiryOwner()
	if logic != "red-lens-logic" {
		t.Fatalf("the inquiry's owner is %q; this test drives the logic lens", logic)
	}
	sit := func(seatID string) {
		t.Helper()
		if _, err := run(t, "register", "--run", runDir, "--seat-id", seatID); err != nil {
			t.Fatalf("register %s: %v", seatID, err)
		}
	}
	check := func(seatID, reason string, extra ...string) (string, error) {
		t.Helper()
		return runAt(t, append([]string{"inquiry", "check", "--run", runDir, "--seat-id", seatID, "--id", "Q1", "--reason", reason}, extra...)...)
	}
	refused := func(what, want string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s was recorded", what)
		}
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("%s was refused, but not for its reason — want %q in:\n%v", what, want, err)
		}
	}
	checks := func() int {
		t.Helper()
		return len(eventsOfWord(t, runDir, "inquiry_check"))
	}
	inquiry := func() record.InquiryJSON {
		t.Helper()
		out, err := run(t, "show", "--run", runDir, "--seat-id", logic, "inquiries", "--json")
		if err != nil {
			t.Fatal(err)
		}
		var got record.InquiriesJSON
		if err := json.Unmarshal([]byte(out), &got); err != nil || len(got.Inquiries) != 1 {
			t.Fatalf("show inquiries --json: %v\n%s", err, out)
		}
		return got.Inquiries[0]
	}

	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", blue,
		"--question", "whether trial division to 9 settles whether 91 is prime", "--hypothesis", "a factor below 10 exists"); err != nil {
		t.Fatal(err)
	}
	sit(logic)
	_, err := check(logic, "the fate stands")
	refused("a check of a proposed inquiry", "a fate that settles nothing", err)
	if got := inquiry().State; got != record.InquiryOpen {
		t.Errorf("a proposed inquiry reads %q, want %q", got, record.InquiryOpen)
	}

	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", blue, "--id", "Q1", "--as", "confirmed",
		"--reason", "7 divides 91", "--method", "trial division by every prime up to 9"); err != nil {
		t.Fatal(err)
	}
	if got := inquiry().State; got != record.InquiryUnchecked {
		t.Errorf("a settled inquiry no check stands on reads %q, want %q", got, record.InquiryUnchecked)
	}
	_, err = check(logic, "the fate stands")
	refused("a check from a sitting opened before the fate", "after your sitting opened", err)

	sit(logic)
	const reason = "the report's trial division names 7 as the factor, which is what the fate claims"
	out, err := check(logic, reason)
	if err != nil {
		t.Fatalf("the logic lens's check, in a sitting opened after the fate, was refused: %v", err)
	}
	if !strings.Contains(out, handleText(t, runDir, "Q1")+" checked") {
		t.Errorf("the check's reply does not name the inquiry it judged:\n%s", out)
	}

	sit("red-lens-evidence")
	_, err = check("red-lens-evidence", "the fate stands")
	refused("another lens's check", "the check of an inquiry is red-lens-logic's", err)

	out, err = check(logic, reason, "--json")
	if err != nil {
		t.Fatalf("the same check again was refused; it records nothing and says so: %v", err)
	}
	var env struct {
		Result struct {
			Repeated bool `json:"repeated"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || !env.Result.Repeated {
		t.Errorf("the same check again does not say it repeated one already on the record (%v):\n%s", err, out)
	}
	if n := checks(); n != 1 {
		t.Errorf("the record holds %d checks after the same check twice, want 1", n)
	}

	// A SECOND JUDGEMENT WHILE ONE STANDS is refused by its own rule, which names what makes the
	// inquiry owed a check again; no seat-held act answers the first.
	_, err = check(logic, "a different account of the same fate")
	refused("a second check while one stands", "a check of inquiry", err)
	if strings.Contains(err.Error(), "If it was wrong") {
		t.Errorf("the refusal of a second check names an answering act, and the record holds none:\n%v", err)
	}
	if n := checks(); n != 1 {
		t.Errorf("the record holds %d checks after a refused second one, want 1", n)
	}

	got := inquiry()
	if got.State != record.InquirySound || got.CheckedBy != logic || got.Check != reason {
		t.Errorf("show inquiries --json reads state %q, check %q by %q; want %q, %q by %q", got.State, got.Check, got.CheckedBy, record.InquirySound, reason, logic)
	}
	bare, err := run(t, "show", "--run", runDir, "--seat-id", logic, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(bare, "checked by "+logic) || !strings.Contains(bare, reason) {
		t.Errorf("show inquiries does not print the check that stands, with its seat and its reason:\n%s", bare)
	}

	// A FATE RECORDED AGAIN — the same fate included — leaves the check stale.
	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", blue, "--id", "Q1", "--as", "confirmed",
		"--reason", "7 divides 91, and 13 is its cofactor"); err != nil {
		t.Fatal(err)
	}
	if got := inquiry(); got.State != record.InquiryUnchecked || got.Check != "" || got.CheckedBy != "" {
		t.Errorf("after a new fate the inquiry reads state %q, check %q by %q; want %q and no check", got.State, got.Check, got.CheckedBy, record.InquiryUnchecked)
	}
}

// castTheLogicLens adds the logic lens to the run's cast and sits it: a run castRun built names
// only the evidence lens, and an inquiry's check is the logic lens's alone. The cast the record
// holds is the last written, so the new one carries every seat the old one did.
func castTheLogicLens(t *testing.T, runDir string) {
	t.Helper()
	cast, err := record.CastOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:logic",
		&recordpb.Cast{SeatIds: append(cast, record.InquiryOwner())}))
	if _, err := run(t, "register", "--run", runDir, "--seat-id", record.InquiryOwner()); err != nil {
		t.Fatalf("register %s: %v", record.InquiryOwner(), err)
	}
}

// BLUE HAS NO BOARD VERBS AND RED HAS NO PROPOSAL VERB. The role boundary is the engine.
func TestRedCannotProposeAnInquiry(t *testing.T) {
	runDir := newRun(t)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", "red-chair",
		"--question", "red's own inquiry"); err == nil {
		t.Fatal("red proposed an inquiry; directing research is what a gap's required_fix does")
	}
}

// The awaiting-a-decision block is what makes the revisit duty checkable rather than hoped for.
func TestOpenInquiriesAreSurfacedAsOwingADecision(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat, "--question", "still open"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Awaiting a decision") || !strings.Contains(out, handleText(t, runDir, "Q1")) {
		t.Errorf("an undecided inquiry was not surfaced as owing a decision:\n%s", out)
	}
}

// A PROPOSAL HOLDS A QUESTION AND NO REASON. The question is what the inquiry asks; a reason is the
// account of a fate, and a proposal has none yet. propose writes --question into `question` and
// leaves `reason` unset, and it does not take --reason at all.
func TestAProposalHoldsAQuestionAndNoReason(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	const question = "whether trial division to 9 settles whether 91 is prime"
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", question, "--hypothesis", "a factor below 10 exists"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if a.GetQuestion() != question {
		t.Errorf("question = %q, want %q", a.GetQuestion(), question)
	}
	if a.Reason != nil {
		t.Errorf("a proposal holds a reason (%q): its text is its question, and a reason is a fate's account", a.GetReason())
	}
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "another question", "--reason", "a reason"); err == nil {
		t.Error("propose took --reason; a proposal has a question and no reason")
	}
}

// THE QUESTION TAKES THE PROSE CHANNEL'S TRIM: a value written through a heredoc ends in a newline
// the seat did not mean as text, and the record holds the question without it.
func TestAProposedQuestionIsTrimmedOfItsTrailingNewline(t *testing.T) {
	runDir := newRun(t)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 is prime\n"); err != nil {
		t.Fatal(err)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_INQUIRY)
	a, _ := recordpb.BodyAs[*recordpb.Inquiry](ev)
	if got := a.GetQuestion(); got != "whether 91 is prime" {
		t.Errorf("question = %q, want it without the trailing newline", got)
	}
}

// THE METHOD IS THE FATE'S ACCOUNT, NOT THE PROPOSAL'S. propose takes no --method, and the write
// refuses a proposal carrying one, so nothing written before the research can print as what the
// research did. `inquiry record` carries it, and the latest record's is the one the inquiry holds —
// a record that gives none holds none, which is how a wrong method is corrected and how one that no
// longer applies goes.
func TestTheMethodIsTheLatestFatesAccount(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 is prime", "--method", "trial division"); err == nil {
		t.Fatal("propose took --method; a proposal is written before any research and holds no method")
	}
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 is prime", "--hypothesis", "a factor below 10 exists"); err != nil {
		t.Fatal(err)
	}
	_, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: seat}, &recordpb.Inquiry{
		InquiryId: proto.String(record.NewID("inquiry")), Question: proto.String("another question"),
		Method: proto.String("trial division"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
	if err == nil || !strings.Contains(err.Error(), "a proposal holds no method") {
		t.Fatalf("the write took a proposal carrying a method: %v", err)
	}
	method := func() (string, []record.InquiryStepJSON) {
		t.Helper()
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries", "--json")
		if err != nil {
			t.Fatal(err)
		}
		var got record.InquiriesJSON
		if err := json.Unmarshal([]byte(out), &got); err != nil || len(got.Inquiries) != 1 {
			t.Fatalf("show inquiries --json: %v\n%s", err, out)
		}
		return got.Inquiries[0].Method, got.Inquiries[0].History
	}
	for _, step := range []struct{ as, method, want string }{
		{"underway", "trial division by 2, 3 and 5", "trial division by 2, 3 and 5"},
		{"confirmed", "trial division by every prime up to 9", "trial division by every prime up to 9"},
		{"confirmed", "", ""},
		{"confirmed", "trial division by 2, 3, 5 and 7", "trial division by 2, 3, 5 and 7"},
	} {
		args := []string{"inquiry", "record", "--run", runDir, "--seat-id", seat, "--id", "Q1", "--as", step.as, "--reason", "7 divides 91"}
		if step.method != "" {
			args = append(args, "--method", step.method)
		}
		if out, err := runAt(t, args...); err != nil {
			t.Fatalf("record %s: %v\n%s", step.as, err, out)
		}
		if got, history := method(); got != step.want {
			t.Errorf("after recording %s with method %q the inquiry holds %q (path %v); the latest fate's method is the inquiry's", step.as, step.method, got, history)
		}
	}
	// EVERY RECORD OF ITS FATE CARRIES ITS OWN ACCOUNT on --json's history, the proposal first.
	_, history := method()
	var methods []string
	for _, s := range history {
		methods = append(methods, s.Method)
		if s.Status != "proposed" && s.Reason != "7 divides 91" {
			t.Errorf("history step %+v does not carry the reason its record gave", s)
		}
	}
	if want := []string{"", "trial division by 2, 3 and 5", "trial division by every prime up to 9", "", "trial division by 2, 3, 5 and 7"}; !slices.Equal(methods, want) {
		t.Errorf("--json history carries methods %q, want each record's own %q", methods, want)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries")
	if err != nil {
		t.Fatal(err)
	}
	// The bare view prints the latest method once, as it does the reason, and each superseded
	// record's own reason and method on its step under the path, so a check can read what each
	// fate rested on.
	for _, m := range []string{"trial division by 2, 3, 5 and 7", "trial division by 2, 3 and 5", "trial division by every prime up to 9"} {
		if n := strings.Count(out, m); n != 1 {
			t.Errorf("method %q prints %d times, want once:\n%s", m, n, out)
		}
	}
	if !regexp.MustCompile(`path: e\d+ proposed -> e\d+ underway -> e\d+ confirmed -> e\d+ confirmed -> e\d+ confirmed\n`).MatchString(out) {
		t.Errorf("the path is not the fates in order:\n%s", out)
	}
	if !regexp.MustCompile(`\n    - e\d+ underway \([^)]+\) — 7 divides 91; method: trial division by 2, 3 and 5\n`).MatchString(out) {
		t.Errorf("the superseded underway record does not print its reason and method on its step:\n%s", out)
	}
}

// THE INQUIRIES FILTER MATCHES WHAT THE VIEW PRINTS, decoded: a quote holding characters the JSON
// encoder escapes (& < > " and a line break) is found, a superseded fate's words still select the
// inquiry because its history is printed, and a key name selects nothing. Every key the page shows
// is present on an inquiry with nothing in it: `wanting` and `history` as [], `check` and
// `checked_by` as "".
func TestTheInquiriesFilterMatchesTheDecodedValuesAndEveryKeyIsPresent(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	seat := inquirySeat(t, runDir)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
		"--question", "whether 91 & 7 share a factor < 10", "--hypothesis", "7 divides 91"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "underway", "--reason", "the cached copy of the \"factor table\" lists it"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
		"--id", "Q1", "--as", "confirmed", "--reason", "7 x 13 = 91"); err != nil {
		t.Fatal(err)
	}
	selected := func(args ...string) int {
		t.Helper()
		out, err := run(t, append([]string{"show", "--run", runDir, "--seat-id", seat, "inquiries"}, args...)...)
		if err != nil {
			t.Fatalf("show inquiries %v: %v\n%s", args, err, out)
		}
		var got record.InquiriesJSON
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("show inquiries %v: %v\n%s", args, err, out)
		}
		return len(got.Inquiries)
	}
	for _, tc := range []struct {
		args []string
		want int
		why  string
	}{
		{[]string{"--quote", "91 & 7 share a factor < 10"}, 1, "a quote holding & and < is the question as printed"},
		{[]string{"--quote", `the "factor table"`}, 1, "a quote holding \" is a superseded fate's reason, and the history is printed"},
		{[]string{"--match", "^status$"}, 0, "a key name is not text the entry holds"},
		{[]string{"--match", "^seat_id$"}, 0, "a key name is not text the entry holds"},
		{[]string{"--quote", "confirmed"}, 1, "the fate is a value"},
	} {
		if got := selected(tc.args...); got != tc.want {
			t.Errorf("show inquiries %v selected %d, want %d: %s", tc.args, got, tc.want, tc.why)
		}
	}

	out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "inquiries", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Inquiries []map[string]json.RawMessage `json:"inquiries"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil || len(raw.Inquiries) != 1 {
		t.Fatalf("show inquiries --json: %v\n%s", err, out)
	}
	for key, want := range map[string]string{"wanting": "[]", "check": `""`, "checked_by": `""`} {
		if got := string(raw.Inquiries[0][key]); got != want {
			t.Errorf("%s is %q, want %s — the page shows the key, so it is always there\n%s", key, got, want, out)
		}
	}
	if h := string(raw.Inquiries[0]["history"]); !strings.HasPrefix(h, "[") {
		t.Errorf("history is %q, want a list\n%s", h, out)
	}
}
