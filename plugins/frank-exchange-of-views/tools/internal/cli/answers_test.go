package cli

import (
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

// No leaf of any seat's tree registers a correction flag, so a verb added later cannot regain one;
// and every recording verb, run with one, is refused by the parser as a flag it does not have.
func TestNoCommandRegistersACorrectionFlag(t *testing.T) {
	var seen int
	roleLeaves(func(role string, path []string, c *cobra.Command) {
		seen++
		for _, name := range []string{"corrects", "correction-why"} {
			if c.Flags().Lookup(name) != nil || c.InheritedFlags().Lookup(name) != nil {
				t.Errorf("%s %s registers --%s: no verb takes a correction — an act stands as filed", role, strings.Join(path, " "), name)
			}
		}
	})
	if seen < 100 {
		t.Fatalf("the walk visited %d commands; it has stopped seeing the surface", seen)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	rows := actRows()
	names := make([]string, 0, len(rows))
	for n := range rows {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		row := rows[name]
		t.Run(name, func(t *testing.T) {
			runDir := actFixture(t)
			var vars actVars
			if row.setup != nil {
				vars = row.setup(t, runDir)
			}
			for _, flag := range [][]string{{"--corrects", "blue-respond:log:#1"}, {"--correction-why", "a word was lost"}} {
				args := append(row.act(vars, "the act as filed"), flag...)
				_, err := runAt(t, append(args, "--run", runDir, "--seat-id", row.seat)...)
				if err == nil || !strings.Contains(err.Error(), "unknown flag: "+flag[0]) {
					t.Fatalf("%s %s: want the parser's refusal of an unknown flag, got %v", name, flag[0], err)
				}
			}
			if n := len(eventsOfWord(t, runDir, "correction")); n != 0 {
				t.Fatalf("the record holds %d correction(s) after refused invocations", n)
			}
		})
	}
}

// eventsOfWord is the run's events of one type.
func eventsOfWord(t *testing.T, runDir, word string) []*record.Event {
	t.Helper()
	run, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	typ, ok := eventTypeWord(word)
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
	return runAt(d.t, append(append([]string{}, args...), "--run", d.run, "--seat-id", seatID)...)
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

// endsWithoutARow holds a refusal for which no seat-held act answers the first.
func (d drive) endsWithoutARow(refusal string) {
	d.t.Helper()
	if strings.Contains(refusal, "If it was wrong") || strings.Contains(refusal, "say so in") {
		d.t.Fatalf("the refusal names an answering act, and this test has driven none for it:\n%s", refusal)
	}
}

const (
	rowPosition    = "a new position at your next sitting (a sitting holds one, and this sitting's stands)"
	rowSpotCheck   = "a new spot-check at your next sitting (a sitting holds one, and this sitting's stands)"
	rowVerdict     = "the verdict of your next sitting (a sitting holds one, and this sitting's stands); after a PASS no chair sitting follows and nothing answers it"
	rowClosing     = "a new closing on the gap at your next sitting (a sitting holds one per --id, and this sitting's stands)"
	rowManifest    = "a new manifest row for the gap at your next sitting (a sitting holds one per --id, and this sitting's stands)"
	rowRegrade     = "a new regrade at your next sitting, while the gap is open (a sitting holds one per --id, and this sitting's stands)"
	rowVerify      = "a new verification of the citation at your next sitting (a sitting holds one per --anchor, and this sitting's stands)"
	rowCorroborate = "a new corroboration of the source at your next sitting (a sitting holds one per --url, and this sitting's stands)"
	rowBacking     = "a verification of the citation this corroboration placed, at your next sitting"
	rowClose       = "a docket motion on the gap"
	rowGradeRule   = "a docket motion on the gap, or a new grade motion while the gap is open"
	rowDocketRule  = "a new docket motion on the gap"
	rowPetition    = "a declaration"
	rowAvenue      = "a move of the avenue"
	rowReproduce   = "a new re-run of the proof, whose reason says which earlier re-run it replaces"
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

func (d drive) gradeMotion(what, seatID, dimension string, admit bool) string {
	d.t.Helper()
	args := []string{"motion", "grade", "file", "--id", "G1", "--dimension", dimension, "--proposed", "high", "--reason", "the consequence is larger"}
	if admit {
		return d.admitted(what, seatID, args...)
	}
	return d.refused(what, seatID, args...)
}

func (d drive) docketMotion(what, seatID string) {
	d.t.Helper()
	d.admitted(what, seatID, "motion", "docket", "file", "--id", "G1", "--reason", "the bench should decide")
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
			d.admitted("the first", seatID, append(append([]string{}, args...), "--reason", "the first")...)
			d.admitted("a new one, in the same sitting", seatID, append(append([]string{}, args...), "--reason", "this replaces my earlier one, which was wrong")...)
		}
	}
	oncePerSitting := func(row, seatID string, args ...string) func(d drive) {
		return func(d drive) {
			d.admitted("the first", seatID, append(append([]string{}, args...), "--reason", "the first")...)
			d.endsWithRow(d.refused("a second in the sitting", seatID, append(append([]string{}, args...), "--reason", "the second")...), row)
			d.nextSitting(seatID)
			d.admitted("the next sitting's", seatID, append(append([]string{}, args...), "--reason", "the first was wrong")...)
		}
	}
	return []answerDrive{
		{name: "position, the chair", row: rowPosition, helps: []string{"chair position"},
			drive: oncePerSitting(rowPosition, "red-chair", "position")},
		{name: "position, blue-respond", row: rowPosition, helps: []string{"blue position"},
			drive: oncePerSitting(rowPosition, "blue-respond", "position")},
		{name: "spot-check", row: rowSpotCheck, helps: []string{"chair spot-check"},
			drive: oncePerSitting(rowSpotCheck, "red-chair", "spot-check", "--none")},
		{name: "verdict", row: rowVerdict, helps: []string{"chair verdict"}, drive: func(d drive) {
			d.admitted("a FAIL", "red-chair", "verdict", "--as", "FAIL")
			d.endsWithRow(d.refused("a second verdict in the sitting", "red-chair", "verdict", "--as", "FAIL"), rowVerdict)
			d.nextSitting("red-chair")
			d.admitted("the next sitting's verdict, after a FAIL", "red-chair", "verdict", "--as", "FAIL")
			// AFTER A PASS the write path still admits the next sitting's verdict; that no chair
			// sitting follows one is the engine's rule, held by the simulator
			// (tests/simulator debate.test.mjs, "a chair that records PASS ends the debate even if it
			// relayed parties").
			d.closeG1(lensSeat)
			d.nextSitting("red-chair")
			d.admitted("a PASS", "red-chair", "verdict", "--as", "PASS")
			d.endsWithRow(d.refused("a second verdict after the PASS", "red-chair", "verdict", "--as", "PASS"), rowVerdict)
			d.nextSitting("red-chair")
			d.admitted("the next sitting's verdict, after a PASS", "red-chair", "verdict", "--as", "PASS")
		}},
		{name: "closing, the chair", row: rowClosing, helps: []string{"chair closing"}, drive: func(d drive) {
			oncePerSitting(rowClosing, "red-chair", "closing", "--id", "G1")(d)
			d.closeG1(lensSeat)
			d.nextSitting("red-chair")
			d.admitted("the next sitting's closing, the gap closed", "red-chair", "closing", "--id", "G1", "--reason", "the first was wrong")
		}},
		{name: "closing, blue", row: rowClosing, helps: []string{"blue closing"}, drive: func(d drive) {
			oncePerSitting(rowClosing, "blue-respond", "closing", "--id", "G1")(d)
			d.closeG1(lensSeat)
			d.nextSitting("blue-respond")
			d.admitted("the next sitting's closing, the gap closed", "blue-respond", "closing", "--id", "G1", "--reason", "the first was wrong")
		}},
		{name: "manifest row", row: rowManifest, helps: []string{"blue manifest-row"}, drive: func(d drive) {
			oncePerSitting(rowManifest, "blue-respond", "manifest-row", "--id", "G1")(d)
			d.closeG1(lensSeat)
			d.nextSitting("blue-respond")
			d.admitted("the next sitting's row, the gap closed", "blue-respond", "manifest-row", "--id", "G1", "--reason", "the first was wrong")
		}},
		{name: "regrade", row: rowRegrade, helps: []string{"lens regrade"}, drive: func(d drive) {
			oncePerSitting(rowRegrade, lensSeat, "regrade", "--id", "G1", "--severity", "high")(d)
			// "while the gap is open": once it is closed the next sitting's regrade is refused.
			d.closeG1(lensSeat)
			d.nextSitting(lensSeat)
			d.refused("a regrade of a closed gap", lensSeat, "regrade", "--id", "G1", "--severity", "low", "--reason", "the first was wrong")
		}},
		{name: "verify", row: rowVerify, helps: []string{"lens verify"}, drive: func(d drive) {
			withFetcher(d.t, &fakeFetcher{resp: map[string][]byte{"https://src/v": []byte("<html>the source</html>")}})
			d.admitted("cite", "blue-respond", "cite", "--quote", quoteTwo, "--url", "https://src/v", "--title", "T")
			anchor := firstCiteEvent(d.t, d.run).GetLabel()
			verify := func(as string) []string {
				return []string{"verify", "--anchor", anchor, "--quote", quoteTwo, "--as", as, "--confidence", "high", "--reason", "read it"}
			}
			d.admitted("the first", lensSeat, verify("supports")...)
			d.endsWithRow(d.refused("a second on the citation in the sitting", lensSeat, verify("weak")...), rowVerify)
			d.nextSitting(lensSeat)
			d.admitted("the next sitting's verification", lensSeat, verify("weak")...)
		}},
		{name: "corroborate, a determination that does not back the claim", row: rowCorroborate,
			when: "for a determination that does not back the claim", helps: []string{"lens corroborate"}, drive: func(d drive) {
				withFetcher(d.t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>the source</html>")}})
				corroborate := func(as string) []string {
					return []string{"corroborate", "--url", "https://src/c", "--title", "T", "--quote", quoteTwo, "--as", as, "--confidence", "high", "--reason", "it says otherwise"}
				}
				d.admitted("the first", lensSeat, corroborate("refutes")...)
				d.endsWithRow(d.refused("a second on the source in the sitting", lensSeat, corroborate("refutes")...), rowCorroborate)
				// Whatever the second determination: the key is the source.
				d.endsWithRow(d.refused("a second with another determination", lensSeat, corroborate("absent")...), rowCorroborate)
				d.nextSitting(lensSeat)
				d.admitted("the next sitting's corroboration", lensSeat, corroborate("absent")...)
			}},
		{name: "corroborate, a determination that backs the claim", row: rowBacking,
			when: "for a determination that backs the claim, which placed a citation", helps: []string{"lens corroborate"}, drive: func(d drive) {
				withFetcher(d.t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>the source</html>")}})
				corroborate := []string{"corroborate", "--url", "https://src/c", "--title", "T", "--quote", quoteTwo, "--as", "supports", "--confidence", "high", "--reason", "it states it"}
				d.admitted("the first", lensSeat, corroborate...)
				placed := eventsOfWord(d.t, d.run, "verify")
				if len(placed) != 1 || placed[0].GetVerify().GetLabel() == "" {
					t := d.t
					t.Fatalf("a backing corroboration is keyed on the label the tool mints, and this one holds none: %v", placed)
				}
				label := placed[0].GetVerify().GetLabel()
				// A REPEAT IS NOT REFUSED: it answers idempotently and writes nothing, so this row
				// reaches a seat through the help page alone.
				if out := d.admitted("the same corroboration again", lensSeat, append(append([]string{}, corroborate...), "--json")...); !strings.Contains(out, `"idempotent":true`) {
					d.t.Fatalf("a repeated backing corroboration is not answered as idempotent:\n%s", out)
				}
				if n := len(eventsOfWord(d.t, d.run, "verify")); n != 1 {
					d.t.Fatalf("a repeated backing corroboration wrote %d verify events, want the one", n)
				}
				verify := []string{"verify", "--anchor", label, "--quote", quoteTwo, "--as", "weak", "--confidence", "low", "--reason", "on rereading it only gestures at it"}
				// "at your next sitting": this sitting's verification of that citation shares the key.
				d.endsWithRow(d.refused("a verification of the placed citation in the same sitting", lensSeat, verify...), rowVerify)
				d.nextSitting(lensSeat)
				d.admitted("the next sitting's verification of the placed citation", lensSeat, verify...)
			}},
		{name: "close, the lens", row: rowClose, helps: []string{"lens close"}, drive: func(d drive) {
			d.closeG1(lensSeat)
			// A fresh second close is refused by the open-gap guard, which names the chair's carry.
			if r := d.refused("a second close", lensSeat, "close", "--id", "G1", "--as", "not_a_defect", "--verified-by", "L1",
				"--verified-with", "read", "--verified-against", "blue/report.md", "--reason", "the class was wrong"); !strings.Contains(r, "carry") {
				d.t.Fatalf("the second close's refusal does not name the carry:\n%s", r)
			}
			d.docketMotion("a docket motion on the closed gap, by the lens", lensSeat)
			// The row may not say "a motion": a GRADE motion on a closed gap is refused.
			d.gradeMotion("a grade motion on the closed gap, by the lens", lensSeat, "impact", false)
		}},
		{name: "close, the chair's carry", row: rowClose, helps: []string{"chair carry"}, drive: func(d drive) {
			d.closeG1(lensSeat)
			d.nextSitting("red-chair")
			carry := func(as string) []string {
				return []string{"carry", "--id", "G1", "--carried-from", "1", "--as", as, "--reason", "restating"}
			}
			d.admitted("the carry", "red-chair", carry("repaired")...)
			d.endsWithRow(d.refused("a second carry in the sitting", "red-chair", carry("not_a_defect")...), rowClose)
			d.docketMotion("a docket motion on the closed gap, by the chair", "red-chair")
			d.gradeMotion("a grade motion on the closed gap, by the chair", "red-chair", "impact", false)
		}},
		{name: "a ruling on a grade motion", row: rowGradeRule, helps: []string{"chair motion grade rule"}, drive: func(d drive) {
			m := fileGradeMotion(d.t, d.run)
			rule := func(as string) []string {
				return []string{"motion", "grade", "rule", "--id", m, "--as", as, "--reason", "the evidence does not reach it"}
			}
			d.admitted("the ruling", "red-chair", rule("rejected")...)
			d.endsWithRow(d.refused("a second ruling", "red-chair", rule("accepted")...), rowGradeRule)
			d.docketMotion("the ruler's docket motion", "red-chair")
			d.gradeMotion("the ruler's new grade motion, the gap open", "red-chair", "impact", true)
			// THE ROW IS THE RULER'S. A seat that DISAGREES with the ruling holds the same two acts,
			// and the record refuses it neither: the filer, the minting lens and the bench each
			// put the gap before the bench, and each files a new grade motion on another axis.
			for i, seatID := range []string{"blue-respond", lensSeat, "judge"} {
				d.docketMotion("a docket motion on the gap, by "+seatID+", which disagrees with the ruling", seatID)
				d.gradeMotion("a new grade motion, the gap open, by "+seatID, seatID, []string{"likelihood", "complexity", "severity"}[i], true)
			}
			// "while the gap is open".
			d.closeG1(lensSeat)
			d.gradeMotion("the ruler's new grade motion, the gap closed", "red-chair", "likelihood", false)
			d.docketMotion("the ruler's docket motion, the gap closed", "red-chair")
			d.docketMotion("the filer's docket motion, the gap closed", "blue-respond")
		}},
		{name: "a ruling on a docket motion", row: rowDocketRule, helps: []string{"bench motion docket rule"}, drive: func(d drive) {
			m := docketFile(d.t, d.run, "red-chair", "G1", "red cannot settle G1")
			rule := func(text string) []string {
				return []string{"motion", "docket", "rule", "--id", m, "--as", "remanded", "--principle", "thoroughness over speed",
					"--tension", "cost against certainty", "--review-flag", "none", "--settled", "nothing yet", "--reopens-on", "a reproduction", "--reason", text}
			}
			d.admitted("the ruling", "judge", rule("one more exchange")...)
			d.endsWithRow(d.refused("a second ruling", "judge", rule("again")...), rowDocketRule)
			d.docketMotion("the bench's new docket motion", "judge")
		}},
		{name: "a ruling on a petition", row: rowPetition, helps: []string{"bench motion petition rule"}, drive: func(d drive) {
			m := motionID(d.t, d.run, "motion", "petition", "file", "--seat-id", lensSeat,
				"--class", "safety", "--relief", "halt before the next exchange", "--reason", "a consent gate is missing")
			d.admitted("the ruling", "judge", "motion", "petition", "rule", "--id", m, "--as", "denied", "--reason", "no boundary is crossed")
			d.endsWithRow(d.refused("a second ruling", "judge", "motion", "petition", "rule", "--id", m, "--as", "granted", "--reason", "again"), rowPetition)
			d.admitted("a declaration", "judge", "declare", "--reason", "my ruling on the petition was wrong, and this is what holds")
		}},
		// NO ROW: the chair holds no act that answers its own ruling on an avenue. A docket motion
		// names a gap and an avenue is not one, so the route a grade ruling has is refused here.
		{name: "a ruling on an avenue", drive: func(d drive) {
			proposeQ1(d.t, d.run)
			d.admitted("the ruling", "red-chair", "motion", "avenue", "rule", "--id", "Q1", "--as", "out_of_scope", "--reason", "not this question")
			d.endsWithoutARow(d.refused("a second ruling", "red-chair", "motion", "avenue", "rule", "--id", "Q1", "--as", "endorsed", "--reason", "again"))
			for _, seatID := range []string{"red-chair", "blue-respond"} {
				d.refused("a docket motion naming the avenue, by "+seatID, seatID, "motion", "docket", "file", "--id", "Q1", "--reason", "the bench should decide the avenue")
			}
			// THE RULING BINDS NO MOVE: what blue does about one it disagrees with is the avenue's
			// own row, in every status.
			for _, status := range []string{"pursued", "declined", "deferred"} {
				d.admitted("a move to "+status+" against the ruling", "blue-respond", "avenue", "move", "--id", "Q1", "--as", status, "--reason", "what the question turns on")
			}
		}},
		{name: "avenue", row: rowAvenue, helps: []string{"blue avenue propose", "blue avenue move"}, drive: func(d drive) {
			proposeQ1(d.t, d.run)
			for _, status := range []string{"pursued", "deferred", "declined", "abandoned", "concluded", "proposed"} {
				d.admitted("a move to "+status, "blue-respond", "avenue", "move", "--id", "Q1", "--as", status, "--reason", "moved to "+status)
			}
			d.admitted("register another blue seat", "blue-synthesize", "register")
			d.admitted("a move by another blue seat", "blue-synthesize", "avenue", "move", "--id", "Q1", "--as", "pursued", "--reason", "another seat moves it")
		}},
		{name: "reproduce", row: rowReproduce, helps: []string{"lens reproduce"}, drive: func(d drive) {
			s := script(d.t, d.run, "p.js", "console.log(91 % 7)")
			d.admitted("prove", "blue-respond", "prove", "--quote", quoteThree, "--script", s, "--reason", "the computation")
			sha := lastBody(d.t, d.run, &recordpb.Proof{}).GetProofSha()
			twice(lensSeat, "reproduce", "--id", sha, "--as", "sound")(d)
		}},
		{name: "avenue review", row: "a new review of the avenues", helps: []string{"chair avenue review"}, drive: twice("red-chair", "avenue", "review")},
		{name: "log", row: "a new log entry", helps: []string{"blue log", "lens log", "chair log", "bench log"}, drive: func(d drive) {
			for _, s := range []string{"blue-respond", lensSeat, "red-chair", "judge"} {
				twice(s, "log", "--type", "defect")(d)
			}
		}},
		{name: "proof", row: "a new proof", helps: []string{"blue prove"}, drive: func(d drive) {
			s := script(d.t, d.run, "p.js", "console.log(91 % 7)")
			twice("blue-respond", "prove", "--quote", quoteThree, "--script", s)(d)
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
		{name: "outcome", row: "a new outcome", helps: []string{"bench outcome"}, drive: twice("judge", "outcome", "--as", "UNVERIFIED")},
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
		typ, ok := eventTypeWord(word)
		if !ok || !record.RefusesARepeat(typ) {
			return
		}
		derived[word] = true
		if len(seat.AnswersFor(c)) == 0 {
			t.Errorf("%s %s records a %s, whose repeat the record refuses, and its page states no answer", role, strings.Join(path, " "), word)
		}
	})
	// The derivation is not vacuous, and it reaches the types the plan's census named.
	for _, want := range []string{"position", "spot_check", "verdict", "closing", "manifest_row", "regrade", "verify", "close"} {
		if !derived[want] {
			t.Errorf("the derivation no longer reaches %s (derived: %v) — the key declarations or the walk moved", want, derived)
		}
	}
}

func eventTypeWord(word string) (recordpb.EventType, bool) {
	vs := recordpb.EventType(0).Descriptor().Values()
	for i := 0; i < vs.Len(); i++ {
		t := recordpb.EventType(vs.Get(i).Number())
		if word != "" && recordpb.Word(t) == word {
			return t, true
		}
	}
	return 0, false
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

// A seat's invocation writes no correction: the one path that does is migrate's replay, and Append
// refuses an identity that carries one outside a migration.
func TestNoSeatIdentityCarriesACorrection(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := actFixture(t)
	must(t, runDir, "log", "--seat-id", "blue-respond", "--type", "defect", "--reason", "a lost word")
	run, err := record.OpenRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	logs := eventsOfWord(t, runDir, "log")
	id := record.Identity{Run: run, SeatID: "blue-respond",
		Correct: &record.Correct{Type: recordpb.EventType_EVENT_TYPE_LOG, Key: logs[len(logs)-1].GetKey(), Why: "a word was lost"}}
	body := &recordpb.Log{Text: strPtr("the word restored"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
	if _, err := record.Append(id, body); err == nil || !strings.Contains(err.Error(), "written by migrate") {
		t.Fatalf("Append admitted a correction outside a migration: %v", err)
	}
	if n := len(eventsOfWord(t, runDir, "correction")); n != 0 {
		t.Fatalf("the record holds %d correction(s)", n)
	}
	// Migrate's replay still writes the pair.
	record.Migrating = true
	_, err = record.Append(id, body)
	record.Migrating = false
	if err != nil {
		t.Fatalf("a migration's replay of a correction was refused: %v", err)
	}
	if n := len(eventsOfWord(t, runDir, "correction")); n != 1 {
		t.Fatalf("the replay wrote %d correction(s), want one", n)
	}
}

func strPtr(s string) *string { return &s }
