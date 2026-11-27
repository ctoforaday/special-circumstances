package cli

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ACT STANDS AS FILED, AND WHAT ANSWERS A WRONG ONE IS ANOTHER ACT.
//
// No verb takes a correction. A second act the record refuses is refused with a tail that names the
// act answering the first, and each recording verb's help ends with the same words — both read from
// record's superseders. A row is a claim about the write path, so this file holds each row to it:
// every act a row's words cover is filed through the command line, at the time the row states, and
// the row as RENDERED — on the refusal and on the help page — is compared with what the write path
// then did. Nothing here compares the text with the function that produced it.

// roleLeaves visits every runnable command of every role's tree, with its role and its path.
func roleLeaves(visit func(role string, path []string, c *cobra.Command)) {
	roots := AllRoots()
	roles := make([]string, 0, len(roots))
	for r := range roots {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	var walk func(role string, path []string, c *cobra.Command)
	walk = func(role string, path []string, c *cobra.Command) {
		if !c.HasSubCommands() {
			visit(role, path, c)
			return
		}
		for _, sub := range c.Commands() {
			walk(role, append(append([]string{}, path...), sub.Name()), sub)
		}
	}
	for _, role := range roles {
		walk(role, nil, roots[role])
	}
}

// eventsOfWord is the run's events of one type.
func eventsOfWord(t *testing.T, runDir, word string) []*record.Event {
	t.Helper()
	run, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	typ, ok := recordpb.EventTypeOf(word)
	if !ok {
		t.Fatalf("%q is not an event type", word)
	}
	evs, _, err := record.EventsOf(run, typ)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

// drive files acts on one run and says what the write path did with each.
type drive struct {
	t   *testing.T
	run string
}

func (d drive) do(seatID string, args ...string) (string, error) {
	d.t.Helper()
	return runAt(d.t, ahead(args, "--run", d.run, "--seat-id", seatID)...)
}

// admitted files an act the write path must admit.
func (d drive) admitted(what, seatID string, args ...string) string {
	d.t.Helper()
	out, err := d.do(seatID, args...)
	if err != nil {
		d.t.Fatalf("%s: the write path refuses it, and a row or this drive says it is admitted: %v", what, err)
	}
	return out
}

// refused files an act the write path must refuse, and returns the refusal.
func (d drive) refused(what, seatID string, args ...string) string {
	d.t.Helper()
	_, err := d.do(seatID, args...)
	if err == nil {
		d.t.Fatalf("%s: the write path admits it, and a row or this drive says it is refused", what)
	}
	return oneLine(err.Error())
}

// nextSitting opens the seat's next sitting.
func (d drive) nextSitting(seatID string) {
	d.t.Helper()
	args := []string{"register"}
	if record.SeatOwesOccasion(seatID) {
		args = append(args, "--occasion", "docket")
	}
	d.admitted("register "+seatID, seatID, args...)
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// tail is what a refusal of a repeated act ends with when the table holds a row for it.
func tail(row string) string {
	return "The first stands. If it was wrong, say so in " + row + "; the record is append-only and both stay visible"
}

// endsWithRow holds a refusal to the row this file drives.
func (d drive) endsWithRow(refusal, row string) {
	d.t.Helper()
	if !strings.HasSuffix(refusal, tail(row)) {
		d.t.Fatalf("the refusal does not end with the row this test drives (%q) — a reworded row names acts nothing has filed; drive them here:\n%s", row, refusal)
	}
}

const (
	rowPosition     = "a new position at your next sitting (a sitting holds one, and this sitting's stands)"
	rowSpotCheck    = "a new spot-check at your next sitting (a sitting holds one, and this sitting's stands)"
	rowVerdict      = "a log entry: a PASS ends the debate, so no chair sitting follows it and no act answers it"
	rowClosing      = "a new closing on the gap at your next sitting (a sitting holds one per --id, and this sitting's stands)"
	rowRegrade      = "a new regrade at your next sitting, while the gap is open (a sitting holds one per --id, and this sitting's stands)"
	rowVerify       = "a new verification of the citation at your next sitting (a sitting holds one per --anchor, and this sitting's stands)"
	rowFoundSource  = "a new reading of that source for this sentence, at your next sitting — a reading of another source, or of this source for another sentence, is recorded now"
	rowClose        = "a new gap on the defect where it still stands in the report: a closed gap stays closed, and no mint names it"
	rowCarry        = "a log entry: a carry restates a closure the record holds, which stands as its lens made it, and a defect still in the report is a new gap of the lens that finds it"
	rowDocketRule   = "a declaration, which moves no gap: the gap comes before you again only on a new docketing of it"
	rowPetition     = "a declaration"
	rowPositionRule = "a declaration"
	rowInquiry      = "a new record of the inquiry's fate"
	rowReproduce    = "a new judgement of the same computation anchor — a seat's latest judgement of a sentence is its verdict on it"
	rowOutcome      = "a certification: the outcome stands once it is recorded, and after it the record takes no seat act but your register for the assemble occasion, your declaration, your certification, the assembly and a log entry"
)

const (
	quoteTwo   = "§2 the finding prose lands in a quoted sentence."
	quoteThree = "the parser accepts an empty body in this line."
)

func (d drive) closeG1(seatID string) {
	d.t.Helper()
	d.admitted("close G1", seatID, "close", "--id", "G1", "--as", "repaired", "--verified-by", "L1",
		"--verified-with", "read", "--verified-against", "blue/report.md", "--reason", "verified at the leaf")
}

// propose files a docket proposal on G1 after its one exchange and returns the motion id. A GAP
// THAT MOVED UNDER THE SEAT'S SITTING refuses the first filing and shows the move; the same filing
// then lands. The row is about what the write path admits, which is that second call.
func (d drive) propose(what, seatID string, act ...string) string {
	d.t.Helper()
	exchangeOn(d.t, d.run, "G1")
	args := append(append([]string{"motion", "docket", "file"}, act...), "--id", "G1", "--reason", "the bench should decide", "--json")
	out, err := d.do(seatID, args...)
	if err != nil {
		if !errors.As(err, new(*record.StaleError)) {
			d.t.Fatalf("%s: refused: %v", what, err)
		}
		out = d.admitted(what, seatID, args...)
	}
	return filedMotionID(d.t, out)
}

// landed files an act the write path must admit, once more where the first call met a gap that
// moved under the seat's sitting.
func (d drive) landed(what, seatID string, args ...string) string {
	d.t.Helper()
	out, err := d.do(seatID, args...)
	if err == nil {
		return out
	}
	if !errors.As(err, new(*record.StaleError)) {
		d.t.Fatalf("%s: refused: %v", what, err)
	}
	return d.admitted(what, seatID, args...)
}

// docketMotion proposes G1's close as repaired.
func (d drive) docketMotion(what, seatID string) string {
	d.t.Helper()
	return d.propose(what, seatID, "close", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
		"--verified-against", "blue/report.md")
}

// answerDrive is one row of the table, driven. helps names every help page that states the row, as
// "<role> <path>"; TestEveryHelpAnswerIsAdmitted fails a page whose answer no drive here covers.
type answerDrive struct {
	name  string
	row   string
	when  string // set where a page states two rows
	helps []string
	drive func(d drive)
}

func answerDrives() []answerDrive {
	// twice files a numbered act twice in one sitting: nothing refuses the second, so the row's
	// "a new <act>" is admitted at once.
	twice := func(seatID string, args ...string) func(d drive) {
		return func(d drive) {
			d.admitted("the first", seatID, ahead(args, "--reason", "the first")...)
			d.admitted("a new one, in the same sitting", seatID, ahead(args, "--reason", "this replaces my earlier one, which was wrong")...)
		}
	}
	oncePerSitting := func(row, seatID string, args ...string) func(d drive) {
		return func(d drive) {
			d.admitted("the first", seatID, ahead(args, "--reason", "the first")...)
			d.endsWithRow(d.refused("a second in the sitting", seatID, ahead(args, "--reason", "the second")...), row)
			d.nextSitting(seatID)
			d.admitted("the next sitting's", seatID, ahead(args, "--reason", "the first was wrong")...)
		}
	}
	return []answerDrive{
		{name: "position, the chair", row: rowPosition, helps: []string{"chair position"},
			drive: oncePerSitting(rowPosition, "red-chair", "position")},
		// A SPOT-CHECK STANDS ON A STALE AREA'S DELIVERED CHANGES: the run has one, and the next sitting's
		// follows a head move the first did not bear.
		{name: "spot-check", row: rowSpotCheck, helps: []string{"chair spot-check"}, drive: func(d drive) {
			chairOwesAStaleArea(d.t, d.run, lensSeat)
			d.admitted("the first", "red-chair", "spot-check", "--reason", "the first")
			d.endsWithRow(d.refused("a second in the sitting", "red-chair", "spot-check", "--reason", "the second"), rowSpotCheck)
			moveReportHead(d.t, d.run)
			d.nextSitting("red-chair")
			d.admitted("the next sitting's", "red-chair", "spot-check", "--reason", "the first was wrong")
		}},
		{name: "verdict", row: rowVerdict, helps: []string{"chair verdict"}, drive: func(d drive) {
			// AFTER A PASS the write path still admits the next sitting's verdict; that no chair
			// sitting follows one is the engine's rule, held by the simulator
			// (tests/simulator debate.test.mjs, "a chair that records PASS ends the debate even if it
			// relayed parties").
			d.closeG1(lensSeat)
			// The lens sits once more for its full re-read: a PASS waits on both whole deliveries.
			lensRereads(d.t, d.run)
			d.nextSitting("red-chair")
			d.admitted("a PASS", "red-chair", "verdict")
			d.endsWithRow(d.refused("a second verdict after the PASS", "red-chair", "verdict"), rowVerdict)
			d.nextSitting("red-chair")
			d.admitted("the next sitting's verdict, after a PASS", "red-chair", "verdict")
		}},
		// A CLOSING ARGUES A DOCKETING, so each drive dockets G1 first.
		{name: "closing, the lens", row: rowClosing, helps: []string{"lens closing"}, drive: func(d drive) {
			d.docketMotion("docket G1", "blue-respond")
			oncePerSitting(rowClosing, lensSeat, "closing", "--id", "G1")(d)
		}},
		{name: "closing, blue", row: rowClosing, helps: []string{"blue closing"}, drive: func(d drive) {
			d.docketMotion("docket G1", "blue-respond")
			oncePerSitting(rowClosing, "blue-respond", "closing", "--id", "G1")(d)
		}},
		{name: "regrade", row: rowRegrade, helps: []string{"lens regrade"}, drive: func(d drive) {
			oncePerSitting(rowRegrade, lensSeat, "regrade", "--id", "G1", "--severity", "high")(d)
			// "while the gap is open": once it is closed the next sitting's regrade is refused.
			d.closeG1(lensSeat)
			d.nextSitting(lensSeat)
			d.refused("a regrade of a closed gap", lensSeat, "regrade", "--id", "G1", "--severity", "low", "--reason", "the first was wrong")
		}},
		{name: "verify", row: rowVerify, when: "for a verification of a citation", helps: []string{"lens verify"}, drive: func(d drive) {
			withFetcher(d.t, &fakeFetcher{resp: map[string][]byte{"https://src/v": []byte("<html>the source</html>")}})
			d.admitted("cite", "blue-respond", "cite", "--quote", quoteTwo, "--url", "https://src/v", "--title", "T")
			anchor := firstCiteEvent(d.t, d.run).GetLabel()
			verify := func(as string) []string {
				return []string{"verify", "--anchor", anchor, "--as", as, "--confidence", "high", "--reason", "read it"}
			}
			d.admitted("the first", lensSeat, verify("supports")...)
			d.endsWithRow(d.refused("a second on the citation in the sitting", lensSeat, verify("weak")...), rowVerify)
			d.nextSitting(lensSeat)
			d.admitted("the next sitting's verification", lensSeat, verify("weak")...)
		}},
		{name: "verify --url, a reading of a source the lens found", row: rowFoundSource,
			when: "for a reading of a source you found", helps: []string{"lens verify"}, drive: func(d drive) {
				withFetcher(d.t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>the source</html>"), "https://src/d": []byte("<html>another source</html>")}})
				read := func(url, quote, as string) []string {
					args := []string{"verify", "--url", url, "--title", "T", "--quote", quote, "--as", as, "--confidence", "high", "--reason", "it says otherwise"}
					return args
				}
				d.admitted("the first", lensSeat, read("https://src/c", quoteTwo, "refutes")...)
				d.endsWithRow(d.refused("a second of the source for the sentence in the sitting", lensSeat, read("https://src/c", quoteTwo, "refutes")...), rowFoundSource)
				// Whatever the second determination: the reading is keyed on its source and its sentence.
				d.endsWithRow(d.refused("a second with another determination", lensSeat, read("https://src/c", quoteTwo, "absent")...), rowFoundSource)
				d.admitted("a reading of another source for the sentence, now", lensSeat, read("https://src/d", quoteTwo, "absent")...)
				d.admitted("a reading of the source for another sentence, now", lensSeat, read("https://src/c", quoteThree, "absent")...)
				d.nextSitting(lensSeat)
				d.admitted("the next sitting's reading", lensSeat, read("https://src/c", quoteTwo, "absent")...)
			}},
		{name: "close, the lens", row: rowClose, helps: []string{"lens close"}, drive: func(d drive) {
			d.closeG1(lensSeat)
			// A fresh second close is refused by the open-gap guard, which names the chair's carry.
			if r := d.refused("a second close", lensSeat, "close", "--id", "G1", "--as", "not_a_defect", "--verified-by", "L1",
				"--verified-with", "read", "--verified-against", "blue/report.md", "--reason", "the class was wrong"); !strings.Contains(r, "carry") {
				d.t.Fatalf("the second close's refusal does not name the carry:\n%s", r)
			}
		}},
		{name: "close, the chair's carry", row: rowCarry, helps: []string{"chair carry"}, drive: func(d drive) {
			d.closeG1(lensSeat)
			d.nextSitting("red-chair")
			// actFixture's chair is in its second sitting (epoch 2) when the lens closes G1.
			carry := func(as string) []string {
				return []string{"carry", "--id", "G1", "--carried-from", "2", "--as", as, "--reason", "restating"}
			}
			// THE EPOCH IS PART OF THE CLAIM: a carry naming a sitting the gap was not closed in is refused.
			if r := d.refused("a carry naming another epoch", "red-chair", "carry", "--id", "G1", "--carried-from", "7", "--as", "repaired", "--reason", "restating"); !strings.Contains(r, "names no closure of gap") {
				d.t.Fatalf("a carry naming another epoch was refused for another reason:\n%s", r)
			}
			d.admitted("the carry", "red-chair", carry("repaired")...)
			d.endsWithRow(d.refused("a second carry in the sitting", "red-chair", carry("not_a_defect")...), rowCarry)
		}},
		{name: "a ruling on a docket motion", row: rowDocketRule, helps: []string{"bench motion docket rule"}, drive: func(d drive) {
			m := docketFile(d.t, d.run, "red-chair", "G1", "red cannot settle G1")
			k := docketNow(d.t, d.run, "G1", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
			conveneBench(d.t, d.run, "G1")
			rule := func(k, text string) []string {
				return []string{"motion", "docket", "rule", "--id", k, "--remand", "--principle", "thoroughness over speed",
					"--tension", "cost against certainty", "--review-flag", "none", "--settled", "nothing yet", "--direction", "a reproduction", "--reason", text}
			}
			d.admitted("the ruling", "judge", rule(k, "one more exchange")...)
			d.endsWithRow(d.refused("a second ruling", "judge", rule(k, "again")...), rowDocketRule)
			// "a declaration": the bench's is admitted in the sitting that ruled.
			before := d.admitted("the board as the ruling left it", "blue-respond", "show", "board")
			// THE REPLY SAYS WHERE IT GOES: the work list is how a holding reaches the other seats.
			if out := d.admitted("a declaration", "judge", "declare", "--reason", "my ruling on that docket motion was wrong, and this is how G1 is to be read"); !strings.Contains(out, "every seat's work list carries it") {
				d.t.Fatalf("the declaration's reply does not say the other seats' work lists carry it:\n%s", out)
			}
			// "which moves no gap": the board is as the ruling left it, the ruling still answers that
			// motion, and the bench's own docket motion is refused.
			if after := d.admitted("the board after the declaration", "blue-respond", "show", "board"); after != before || !strings.Contains(after, `"open": true`) {
				d.t.Fatalf("the declaration moved the board:\nbefore\n%s\nafter\n%s", before, after)
			}
			d.endsWithRow(d.refused("a second ruling after the declaration", "judge", rule(k, "again")...), rowDocketRule)
			if r := d.refused("the bench's own docket motion", "judge", "motion", "docket", "file", "regrade", "--id", "G1", "--severity", "low", "--reason", "the bench should decide again"); !strings.Contains(r, "is ruled by the bench seat") {
				d.t.Fatalf("the bench's docket motion is refused for another reason than its gavel:\n%s", r)
			}
			// "only on a new docketing of it": blue's proposal is admitted on the remanded gap, the
			// lens's on its own gap is refused (it takes the act itself), and the bench rules blue's
			// docketing in the sitting convened for it.
			if r := d.refused("the lens's proposal on its own gap", lensSeat, "motion", "docket", "file", "regrade", "--id", "G1", "--severity", "low", "--reason", "r"); !strings.Contains(r, "you are its assignee") {
				d.t.Fatalf("the lens's proposal on its own gap is refused for another reason:\n%s", r)
			}
			again := d.docketMotion("a new proposal on the ruled gap, by blue", "blue-respond")
			k2 := docketNow(d.t, d.run, "G1", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, again)
			conveneBench(d.t, d.run, "G1")
			d.admitted("the bench's ruling on the new docketing", "judge", rule(k2, "the gap again")...)
		}},
		{name: "a ruling on a petition", row: rowPetition, helps: []string{"bench motion petition rule"}, drive: func(d drive) {
			m := motionID(d.t, d.run, "motion", "petition", "file", "--seat-id", lensSeat,
				"--class", "safety", "--relief", "halt before the next exchange", "--reason", "a consent gate is missing")
			d.admitted("the ruling", "judge", "motion", "petition", "rule", "--id", m, "--as", "denied", "--reason", "no boundary is crossed")
			d.endsWithRow(d.refused("a second ruling", "judge", "motion", "petition", "rule", "--id", m, "--as", "granted", "--relief", "Restate the section as the bench orders.", "--binds", "blue", "--reason", "again"), rowPetition)
			d.admitted("a declaration", "judge", "declare", "--reason", "my ruling on the petition was wrong, and this is what holds")
		}},
		{name: "a ruling on a position", row: rowPositionRule, helps: []string{"bench motion position rule"}, drive: func(d drive) {
			id := heardPosition(d.t, d.run)
			d.admitted("the ruling", "judge", "motion", "position", "rule", "--id", id, "--as", "overruled", "--reason", "the objection does not hold")
			d.endsWithRow(d.refused("a second ruling", "judge", "motion", "position", "rule", "--id", id, "--as", "upheld",
				"--lens", lensSeat, "--direction", "audit the inference", "--reason", "again"), rowPositionRule)
			d.admitted("a declaration", "judge", "declare", "--reason", "my ruling on the position was wrong, and this is what holds")
		}},
		{name: "inquiry", row: rowInquiry, helps: []string{"blue inquiry propose", "blue inquiry record", "lane inquiry propose", "lane inquiry record"}, drive: func(d drive) {
			proposeQ1(d.t, d.run)
			for _, status := range []string{"underway", "deferred", "declined", "abandoned", "confirmed", "refuted", "proposed"} {
				d.admitted("a fate of "+status, "blue-respond", "inquiry", "record", "--id", "Q1", "--as", status, "--reason", "recorded as "+status)
			}
			d.admitted("register another blue seat", "blue-synthesize", "register")
			d.admitted("a fate recorded by another blue seat", "blue-synthesize", "inquiry", "record", "--id", "Q1", "--as", "underway", "--reason", "another seat records it")
		}},
		{name: "reproduce", row: rowReproduce, helps: []string{"lens reproduce"}, drive: func(d drive) {
			s := script(d.t, d.run, "p.js", "console.log(91 % 7)")
			d.admitted("compute", "blue-respond", "compute", "--quote", quoteThree, "--dir", s, "--reason", "the computation", "--", "node", "p.js")
			computationAnchor := lastBody(d.t, d.run, &recordpb.Computation{}).GetComputationId()
			twice(lensSeat, "reproduce", "--anchor", computationAnchor, "--as", "sound")(d)
		}},
		{name: "log", row: "a new log entry", helps: []string{"blue log", "lane log", "lens log", "chair log", "bench log"}, drive: func(d drive) {
			for _, s := range []string{"blue-respond", "blue-lane-1", lensSeat, "red-chair", "judge"} {
				twice(s, "log", "--type", "defect")(d)
			}
		}},
		{name: "computation", row: "a new computation", helps: []string{"blue compute"}, drive: func(d drive) {
			s := script(d.t, d.run, "p.js", "console.log(91 % 7)")
			twice("blue-respond", "compute", "--quote", quoteThree, "--dir", s, "--", "node", "p.js")(d)
		}},
		{name: "cite", row: "a new citation", helps: []string{"blue cite"}, drive: func(d drive) {
			withFetcher(d.t, &fakeFetcher{resp: map[string][]byte{
				"https://src/a": []byte("<html>a source on the finding</html>"),
				"https://src/b": []byte("<html>the source that was meant</html>")}})
			d.admitted("the first", "blue-respond", "cite", "--quote", quoteTwo, "--url", "https://src/a", "--title", "A", "--reason", "it states the finding")
			d.admitted("a new citation of the same sentence", "blue-respond", "cite", "--quote", quoteTwo, "--url", "https://src/b", "--title", "B",
				"--reason", "the first cited the wrong source")
			if n := len(eventsOfWord(d.t, d.run, "cite")); n != 2 {
				d.t.Fatalf("the record holds %d cites, want both", n)
			}
		}},
		{name: "declare", row: "a new declaration", helps: []string{"bench declare"}, drive: twice("judge", "declare")},
		{name: "certify", row: "a new certification", helps: []string{"bench certify"}, drive: twice("judge", "certify")},
		{name: "halt", row: "a new halt", helps: []string{"bench halt"}, drive: twice("judge", "halt")},
		{name: "outcome", row: rowOutcome, helps: []string{"bench outcome"}, drive: func(d drive) {
			// THE OUTCOME CLOSES THE RECORD TO SEAT ACTS (record.requireOpenToSeats): a second is refused,
			// and the certification the row names is admitted after it.
			d.admitted("the assembly sitting", "judge", "register", "--occasion", "assemble")
			d.admitted("the outcome", "judge", "outcome", "--as", "UNVERIFIED", "--reason", "the first")
			if r := d.refused("a second outcome", "judge", "outcome", "--as", "UNVERIFIED", "--reason", "this replaces my earlier one"); !strings.Contains(r, "the run has its outcome") {
				d.t.Fatalf("a second outcome is refused, but not as the run's end: %s", r)
			}
			d.admitted("the certification that answers it", "judge", "certify", "--reason", "the stamp's reason was wrong, and this is what a human should re-examine")
		}},
	}
}

// Every row is driven, member by member; a refusal that ends with words no drive covers fails.
func TestARefusedRepeatNamesAnActTheWritePathAdmits(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, a := range answerDrives() {
		t.Run(a.name, func(t *testing.T) {
			runDir := actFixture(t)
			if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
				t.Fatal(err)
			}
			a.drive(drive{t: t, run: runDir})
		})
	}
}

// The set of types that owe a row is derived from the record's key declarations, never listed: a
// type whose second act in a sitting the record refuses, recorded by some command, holds a row on
// that command's page.
func TestEveryRefusedRepeatHasARow(t *testing.T) {
	derived := map[string]bool{}
	roleLeaves(func(role string, path []string, c *cobra.Command) {
		word := seat.RecordType(c)
		typ, ok := recordpb.EventTypeOf(word)
		if !ok || !record.RefusesARepeat(typ) {
			return
		}
		derived[word] = true
		if len(seat.AnswersFor(c)) == 0 {
			t.Errorf("%s %s records a %s, whose repeat the record refuses, and its page states no answer", role, strings.Join(path, " "), word)
		}
	})
	// The derivation is not vacuous, and it reaches the types the plan's census named.
	for _, want := range []string{"position", "spot_check", "verdict", "closing", "regrade", "verify", "close"} {
		if !derived[want] {
			t.Errorf("the derivation no longer reaches %s (derived: %v) — the key declarations or the walk moved", want, derived)
		}
	}
}

// Each page's answer is one this file drives: the paragraph a verb's help ends with is compared,
// as rendered, with the rows whose drives name that page.
func TestEveryHelpAnswerIsAdmitted(t *testing.T) {
	want := map[string][]answerDrive{}
	for _, a := range answerDrives() {
		for _, h := range a.helps {
			want[h] = append(want[h], a)
		}
	}
	seen := map[string]bool{}
	roleLeaves(func(role string, path []string, c *cobra.Command) {
		page := role + " " + strings.Join(path, " ")
		help := oneLine(c.Long)
		i := strings.Index(help, seat.AnswerLead)
		rows := want[page]
		if i < 0 {
			if len(rows) > 0 {
				t.Errorf("%s: a drive names this page for %q, and the page states no answer", page, rows[0].row)
			}
			return
		}
		seen[page] = true
		got := help[i:]
		var expect string
		switch len(rows) {
		case 0:
			t.Errorf("%s states an answer no drive covers — drive every act its words name in answerDrives:\n%s", page, got)
			return
		case 1:
			expect = seat.AnswerLead + ", say so in " + rows[0].row + "."
		default:
			// A page that states two rows says which acts each is for; the table's order is the
			// key fields' order.
			sort.Slice(rows, func(i, j int) bool { return rows[i].when < rows[j].when })
			parts := make([]string, len(rows))
			for i, r := range rows {
				parts[i] = r.when + ", say so in " + r.row
			}
			expect = seat.AnswerLead + ": " + strings.Join(parts, "; ") + "."
		}
		if got != expect {
			t.Errorf("%s:\n  the page says  %s\n  the drives say %s", page, got, expect)
		}
	})
	for page := range want {
		if !seen[page] {
			t.Errorf("a drive names the page %q, and no such page states an answer", page)
		}
	}
}
