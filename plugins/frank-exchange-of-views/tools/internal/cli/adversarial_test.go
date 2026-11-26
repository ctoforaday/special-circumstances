package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"strings"
	"testing"
)

// SCENARIOS A HOSTILE SEAT WOULD TRY, and every one here was found by BEING one.
//
// # Why this file exists and what it is not
//
// TestFuzzDebate proves a verb is REACHABLE and that its events REACH A READER. It cannot ask the
// question this file is made of — "what does a seat get if it tries the wrong thing?" — because a
// harness only ever drives the paths it was written to drive, and it was written by the same
// person who wrote the verb.
//
// So these were found by hand, driving the real binary as each seat and deliberately taking the
// awkward path: ruling under the wrong subgroup, ruling twice, appealing a ruling,
// passing a verdict over an unanswered ask. Six of the ten scenarios below were ACCEPTED by the
// tool when first tried, and four of those six reached the rendered report.
//
// The sharpest was a ruling of a PETITION motion under another subgroup. It was accepted: the subject
// came from the seat's POSITION IN THE COMMAND TREE rather than from the motion, so the caller
// chose both the gavel and the verdict vocabulary by picking a subgroup. The report rendered
// "petition (safety) … ruled accepted" — a verdict that is not in the petition set at all. That is
// facts-are-fields one level below where the collapse applied it: a fact recovered from position
// instead of read from the record.
//
// # Massively parallel by construction
//
// Every scenario gets its own recordtest.TmpRun(t) and its own run directory, and `run` builds a fresh
// cobra root with its own output buffer per call. There is no shared state to serialise on — the
// record's locks are per run directory, and nothing here touches process globals (no env, no
// chdir). So every scenario declares t.Parallel() and the suite scales with cores.
//
// # How to add one
//
// Be a seat and try the thing you suspect. If the tool accepts it, decide whether it SHOULD — and
// if it should not, fix the write path first and add the scenario second. A scenario asserting
// today's behaviour because that is what happens is not a test, it is a screenshot.

// seatStep is one CLI invocation in a scenario's build-up.
type seatStep []string

// adversarialCase is one hostile move against a built-up board.
//
// `guards` is not decoration: it is what this scenario would catch if the check regressed, and it
// is the only part a reader cannot reconstruct from the args.
type adversarialCase struct {
	name string
	// setup runs in order and MUST all succeed — a scenario whose setup silently failed asserts
	// its refusal against an empty board, which passes for the wrong reason.
	setup []seatStep
	// proposed is the docket proposal filed once setup has run and gap G1 has had its one exchange
	// (exchangeOn): a proposal is admitted from there, and no CLI step stages an exchange.
	proposed []seatStep
	// act is the hostile move.
	act seatStep
	// refused is what the message must contain. Empty means the act must be ACCEPTED.
	refused string
	guards  string
}

func adversarialCases() []adversarialCase {
	// The board every motion scenario argues over: one gap, minted by the lens.
	mint := seatStep{"mint", "--about-kind", "section", "--about", "Findings", "--seat-id", "red-lens-evidence",
		"--key", "k1", "--class", "spec_underspecification", "--problem", "p",
		"--fix", "f", "--check", "c", "--check-kind", "document",
		"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "low",
		"--reason", "the board needs something to argue about"}
	fileDocket := append(seatStep{"motion", "docket", "file", "close", "--seat-id", "blue-respond",
		"--id", "G1", "--reason", "the consequence is bounded by the caller's own validation; the bench should decide the gap"},
		closeFlags("not_a_defect")...)
	filePetition := seatStep{"motion", "petition", "file", "--seat-id", "red-lens-evidence",
		"--class", "safety", "--relief", "halt before the next round",
		"--reason", "continuing would require asserting a consent gate that does not exist"}
	propose := seatStep{"inquiry", "propose", "--seat-id", "blue-respond",
		"--question", "a question worth taking up", "--hypothesis", "it would settle the open question"}

	return []adversarialCase{
		{
			name:    "a petition cannot be ruled through the docket subgroup",
			setup:   []seatStep{filePetition},
			act:     seatStep{"motion", "docket", "rule", "--seat-id", "judge", "--id", "M1", "--for-blue", "--reason", "ruling a petition as if it were a docket motion"},
			refused: "a docketing id is D-",
			guards: "THE ONE THAT WAS ACCEPTED AND REACHED THE REPORT. The subject came from the " +
				"command tree, so the caller chose the gavel AND the answer set. A docket ruling " +
				"names a docketing, so a petition's id is refused by its shape before any answer is read.",
		},
		{
			name:    "the wrong subgroup is named before the verb's own refusals",
			setup:   []seatStep{filePetition},
			act:     seatStep{"motion", "docket", "rule", "--seat-id", "judge", "--id", "M1", "--for-blue", "--reason", ""},
			refused: "a docketing id is D-",
			guards: "ORDER IS THE MESSAGE. The verb refuses an unreasoned ruling itself, before the " +
				"write; asked first, that would send the chair to write a reason for a ruling it " +
				"cannot make. The subject is established before anything phrased in its terms.",
		},
		{
			name:     "a docket motion cannot be ruled through the petition subgroup",
			setup:    []seatStep{mint},
			proposed: []seatStep{fileDocket},
			act:      seatStep{"motion", "petition", "rule", "--seat-id", "judge", "--id", "M1", "--as", "granted", "--relief", "Restate the section as the bench orders.", "--binds", "blue", "--reason", "the bench takes the chair's docket"},
			refused:  "was filed as a docket motion and you are ruling it as a petition",
			guards:   "The same hole from the other side — a docket motion answered with a petition's verdicts.",
		},
		{
			name:    "a motion is answered once",
			setup:   []seatStep{filePetition, {"motion", "petition", "rule", "--seat-id", "judge", "--id", "M1", "--as", "granted", "--relief", "Restate the section as the bench orders.", "--binds", "blue", "--reason", "granted on the papers"}},
			act:     seatStep{"motion", "petition", "rule", "--seat-id", "judge", "--id", "M1", "--as", "denied", "--reason", "reversing myself silently"},
			refused: "is already ruled",
			guards: "Two rulings replay as whichever the shard ordering favours and the report " +
				"shows ONE, with nothing saying the other exists. " +
				"The ruler's own answer to a wrong ruling is a new act, which keeps both.",
		},
		{
			name:    "a ruling is not appealed, and the refusal says what the subject holds",
			setup:   []seatStep{filePetition, {"motion", "petition", "rule", "--seat-id", "judge", "--id", "M1", "--as", "denied", "--reason", "the record shows the gate"}},
			act:     seatStep{"motion", "petition", "appeal", "--seat-id", "red-lens-evidence", "--id", "M1", "--reason", "pressing a ruling"},
			refused: "has no `appeal` verb — it has file",
			guards: "It answered `unknown flag: --id`, which sends a seat looking at its flags for " +
				"a problem that is not there. A motion is filed and ruled and the ruling stands — " +
				"and the refusal is where a seat meets that.",
		},
		{
			name:    "PASS is refused while a motion is unanswered",
			setup:   []seatStep{mint, filePetition, {"close", "--seat-id", "red-lens-evidence", "--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "closed on the merits"}},
			act:     seatStep{"verdict", "--seat-id", "red-chair"},
			refused: "filed and never ruled",
			guards: "A probe walked a run to `verdict PASS` AND `outcome VERIFIED` with a " +
				"motion never ruled. The gate counted open GAPS only. PASS claims nothing is left " +
				"open, and an unanswered ask is exactly that.",
		},
		{
			name: "PASS is allowed once every motion is answered",
			// The lens closes G1 at its next sitting, its full re-read, delivered before the close: a
			// PASS waits on both whole deliveries.
			setup: []seatStep{mint, filePetition, {"motion", "petition", "rule", "--seat-id", "judge", "--id", "M1", "--as", "denied", "--reason", "the record shows the gate"}, {"register", "--seat-id", "red-lens-evidence"}, {"close", "--seat-id", "red-lens-evidence", "--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "closed on the merits"}},
			act:   seatStep{"verdict", "--seat-id", "red-chair"},
			guards: "The gate above must not become unpassable. A check that no legitimate run can " +
				"satisfy is removed by the first person it blocks.",
		},
		{
			name:    "an inquiry check names an inquiry that exists",
			act:     seatStep{"inquiry", "check", "--seat-id", "red-lens-evidence", "--id", "Q9", "--reason", "checking an inquiry nobody proposed"},
			refused: "names no inquiry on the record",
			guards: "A check joins on the INQUIRY's own id, so the dangling-reference discipline has to " +
				"reach a different id space than a gap's or a motion's.",
		},
		{
			name:    "an inquiry check is the logic lens's, and another lens is told so",
			setup:   []seatStep{propose, {"inquiry", "record", "--seat-id", "blue-respond", "--id", "Q1", "--as", "confirmed", "--reason", "the approach ran and answered"}},
			act:     seatStep{"inquiry", "check", "--seat-id", "red-lens-evidence", "--id", "Q1", "--reason", "its fate stands"},
			refused: "the check of an inquiry is red-lens-logic's",
			guards: "Every lens carries `inquiry check` on its tree and one lens owns the judgement, so " +
				"two lenses dispatched together never race over one inquiry. Absent-because-not-yours " +
				"must not read as absent-by-design: the refusal names whose it is and what the other " +
				"lens does instead.",
		},
		{
			name:  "blue records a fate on a settled inquiry",
			setup: []seatStep{propose, {"inquiry", "record", "--seat-id", "blue-respond", "--id", "Q1", "--as", "confirmed", "--reason", "the approach ran and answered"}},
			act:   seatStep{"inquiry", "record", "--seat-id", "blue-respond", "--id", "Q1", "--as", "refuted", "--reason", "a later reading answered it the other way"},
			guards: "A SETTLED INQUIRY STAYS BLUE'S TO REVISIT: a new fate is how blue answers what it " +
				"learns, and it leaves any check of the old fate stale rather than being refused by it.",
		},
		{
			name: "a petition class outside the four is refused",
			act:  seatStep{"motion", "petition", "file", "--seat-id", "red-lens-evidence", "--class", "aesthetic", "--relief", "halt", "--reason", "petitioning on a standard nobody wrote"},
			// THE ORDER IS THE SCHEMA'S, and this expectation used to carry a different one — from
			// the hand-written table that listed four classes PetitionClass did not have. Two of
			// the four were refused at the write for words the help advertised. The set is derived
			// from the enum now, so its order follows declaration: integrity and safety kept their
			// numbers (they were the two that worked), ethical and constitutional took fresh ones
			// because the numbers they would have reused are reserved.
			refused: "--class takes one of integrity|safety|ethical|constitutional",
			guards: "The bench is convened PER CLASS; a fifth is a petition heard under whichever " +
				"standard the ruling seat happened to imagine. Refused at PARSE now, with the " +
				"four classes and what each is for — the seat learns the vocabulary from the refusal.",
		},
		{
			name:     "the gavel holds even with the right subject",
			setup:    []seatStep{mint},
			proposed: []seatStep{fileDocket},
			act:      seatStep{"motion", "docket", "rule", "--seat-id", "blue-respond", "--id", "M1", "--for-blue", "--reason", "ruling my own motion"},
			refused:  "is ruled by the bench",
			guards: "Filed by any seat, ruled by one. The asymmetry IS the mechanism, and blue " +
				"answering its own ask would make the exchange a monologue.",
		},
	}
}

// TestAHostileSeatIsRefused runs every scenario in its own run directory, in parallel.
func TestAHostileSeatIsRefused(t *testing.T) {
	// SET ONCE, IN THE PARENT. t.Setenv forbids t.Parallel on the test that calls it, and the
	// package's usual seatRun helper calls it — so the scenarios build their own run directory
	// and inherit this. The parent is not parallel, the subtests are, and the env is a
	// process-global that only needs setting once anyway.
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	for _, tc := range adversarialCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runDir := adversarialRun(t)

			for i, step := range tc.setup {
				args := byHandle(t, runDir, append(append([]string{}, step...), "--run", runDir))
				if _, err := run(t, args...); err != nil {
					t.Fatalf("setup step %d (%v) failed: %v\n\nA scenario whose setup fails asserts its outcome against a board that was never built, which passes for the wrong reason.", i, step, err)
				}
			}
			if len(tc.proposed) > 0 {
				exchangeOn(t, runDir, handleText(t, runDir, "G1"))
			}
			for i, step := range tc.proposed {
				args := byHandle(t, runDir, append(append([]string{}, step...), "--run", runDir))
				if _, err := run(t, args...); err != nil {
					t.Fatalf("proposal %d (%v) failed: %v\n\nA scenario whose setup fails asserts its outcome against a board that was never built, which passes for the wrong reason.", i, step, err)
				}
			}

			args := byHandle(t, runDir, append(append([]string{}, tc.act...), "--run", runDir))
			out, err := run(t, args...)

			if tc.refused == "" {
				if err != nil {
					t.Fatalf("the act was REFUSED and this scenario expects it to be allowed: %v\n\nguards: %s", err, tc.guards)
				}
				return
			}
			if err == nil {
				t.Fatalf("the act was ACCEPTED and must not be (stdout %q).\n\nguards: %s", out, tc.guards)
			}
			if !strings.Contains(err.Error(), tc.refused) {
				t.Errorf("refused, but not for the stated reason.\n got: %v\nwant it to contain: %q\n\nguards: %s\n\nA refusal with the wrong message is a seat sent to the wrong fix — under a scoped surface \"not yours\" reads as \"does not exist\", and a seat that believes a verb is missing works around it and loses the capability for the run.", err, tc.refused, tc.guards)
			}
		})
	}
}

// TestEveryAdversarialScenarioStatesWhatItGuards keeps the table honest.
//
// The args say WHAT a scenario does; only `guards` says what it would catch, and that is the half
// a later reader cannot reconstruct. A scenario added without it becomes, within one refactor, a
// line nobody dares delete and nobody can justify.
func TestEveryAdversarialScenarioStatesWhatItGuards(t *testing.T) {
	t.Parallel()
	cases := adversarialCases()
	if len(cases) < 10 {
		t.Fatalf("only %d scenarios — the table was gutted, and a table this small stops being a sweep", len(cases))
	}
	seen := map[string]bool{}
	for _, tc := range cases {
		if strings.TrimSpace(tc.guards) == "" {
			t.Errorf("scenario %q states no `guards` — say what it would catch", tc.name)
		}
		if len(tc.act) == 0 {
			t.Errorf("scenario %q has no act", tc.name)
		}
		if seen[tc.name] {
			t.Errorf("duplicate scenario name %q — the subtests would collide and one would be invisible", tc.name)
		}
		seen[tc.name] = true
	}
}

// adversarialRun is seatRun without the t.Setenv, so scenarios can be parallel. The four seats
// are registered because a motion names its filer from the seat id, and an unregistered seat is
// refused before any scenario's own check is reached.
func adversarialRun(t *testing.T) string {
	t.Helper()
	runDir := castRun(t)
	castSeats(t, runDir)
	return runDir
}
