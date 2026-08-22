package cli

import (
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
// awkward path: ruling under the wrong subgroup, ruling twice, appealing what nobody ruled,
// passing a verdict over an unanswered ask. Six of the ten scenarios below were ACCEPTED by the
// tool when first tried, and four of those six reached the rendered report.
//
// The sharpest was `motion grade rule --id M1` on a PETITION motion. It was accepted: the subject
// came from the seat's POSITION IN THE COMMAND TREE rather than from the motion, so the caller
// chose both the gavel and the verdict vocabulary by picking a subgroup. The report rendered
// "petition (safety) … ruled accepted" — a verdict that is not in the petition set at all. That is
// facts-are-fields one level below where the collapse applied it: a fact recovered from position
// instead of read from the record.
//
// # Massively parallel by construction
//
// Every scenario gets its own t.TempDir() and its own run directory, and `run` builds a fresh
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
	// act is the hostile move.
	act seatStep
	// refused is what the message must contain. Empty means the act must be ACCEPTED.
	refused string
	guards  string
}

func adversarialCases() []adversarialCase {
	// The board every motion scenario argues over: one gap, minted by the merge.
	mint := seatStep{"merge", "mint", "--seat-id", "red-merge-r1",
		"--key", "k1", "--class", "self-attestation", "--problem", "p",
		"--fix", "f", "--check", "c", "--check-kind", "document",
		"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "low",
		"--reason", "the board needs something to argue about"}
	fileGrade := seatStep{"motion", "grade", "file", "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation"}
	filePetition := seatStep{"motion", "petition", "file", "--seat-id", "red-lens-r1-L1",
		"--class", "safety", "--relief", "halt before the next round",
		"--reason", "continuing would require asserting a consent gate that does not exist"}
	propose := seatStep{"blue", "line-of-inquiry", "propose", "--seat-id", "blue-respond-r1",
		"--reason", "a line worth taking", "--hypothesis", "it would settle the open question"}

	return []adversarialCase{
		{
			name:    "a petition cannot be ruled through the grade subgroup",
			setup:   []seatStep{filePetition},
			act:     seatStep{"motion", "grade", "rule", "--seat-id", "red-merge-r1", "--id", "M1", "--as", "accepted", "--reason", "ruling a petition as if it were a grade"},
			refused: "was filed as a petition motion and you are ruling it as a grade",
			guards: "THE ONE THAT WAS ACCEPTED AND REACHED THE REPORT. The subject came from the " +
				"command tree, so the caller chose the gavel (the merge answering what only the " +
				"bench may answer) AND the verdict set (`accepted`, not a petition ruling at all).",
		},
		{
			name:    "a grade motion cannot be ruled through the petition subgroup",
			setup:   []seatStep{mint, fileGrade},
			act:     seatStep{"motion", "petition", "rule", "--seat-id", "judge-r1", "--id", "M1", "--as", "granted", "--reason", "the bench takes the merge's docket"},
			refused: "was filed as a grade motion and you are ruling it as a petition",
			guards:  "The same hole from the other side — the bench reaching a grade the merge owns.",
		},
		{
			name:    "a motion is answered once",
			setup:   []seatStep{filePetition, {"motion", "petition", "rule", "--seat-id", "judge-r1", "--id", "M1", "--as", "granted", "--reason", "granted on the papers"}},
			act:     seatStep{"motion", "petition", "rule", "--seat-id", "judge-r1", "--id", "M1", "--as", "denied", "--reason", "reversing myself silently"},
			refused: "is already ruled",
			guards: "Two rulings replay as whichever the shard ordering favours and the report " +
				"shows ONE, with nothing saying the other exists. Measured on both petition and " +
				"direction. The escalation path is an appeal, which keeps both positions.",
		},
		{
			name:  "a ruling can be appealed",
			setup: []seatStep{mint, fileGrade, {"motion", "grade", "rule", "--seat-id", "red-merge-r1", "--id", "M1", "--as", "rejected", "--reason", "the evidence does not reach it"}},
			act:   seatStep{"motion", "grade", "appeal", "--seat-id", "blue-respond-r1", "--id", "M1", "--reason", "pressing it to the bench on new grounds"},
			guards: "The positive case beside the refusal above: pressing a settled motion is the " +
				"SUPPORTED move, and a suite that only tested refusals would let it rot.",
		},
		{
			name:    "an appeal needs a ruling to appeal against",
			setup:   []seatStep{mint, fileGrade},
			act:     seatStep{"motion", "grade", "appeal", "--seat-id", "blue-respond-r1", "--id", "M1", "--reason", "pressing on before anyone answered"},
			refused: "no ruling to appeal",
			guards: "Otherwise the motion replays both unruled and appealed, a state the report " +
				"has no honest sentence for.",
		},
		{
			name:    "a petition has no appeal, and says so",
			setup:   []seatStep{filePetition, {"motion", "petition", "rule", "--seat-id", "judge-r1", "--id", "M1", "--as", "denied", "--reason", "denied on the papers"}},
			act:     seatStep{"motion", "petition", "appeal", "--seat-id", "blue-respond-r1", "--id", "M1", "--reason", "pressing a petition"},
			refused: "has no `appeal` verb",
			guards: "It answered `unknown flag: --id`, which sends a seat looking at its flags for " +
				"a problem that is not there. A petition is heard BEFORE the debate continues, so " +
				"there is nothing to escalate to — and the refusal is where a seat meets that.",
		},
		{
			name:    "PASS is refused while a motion is unanswered",
			setup:   []seatStep{mint, fileGrade, {"merge", "close", "--seat-id", "red-merge-r1", "--id", "R1-1", "--as", "closed", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "closed on the merits"}},
			act:     seatStep{"merge", "verdict", "--seat-id", "red-merge-r1", "--as", "PASS"},
			refused: "filed and never ruled",
			guards: "A probe walked a run to `verdict PASS` AND `outcome VERIFIED` with a grade " +
				"motion never ruled. The gate counted open GAPS only. PASS claims nothing is left " +
				"open, and an unanswered ask is exactly that.",
		},
		{
			name:  "PASS is allowed once every motion is answered",
			setup: []seatStep{mint, fileGrade, {"motion", "grade", "rule", "--seat-id", "red-merge-r1", "--id", "M1", "--as", "rejected", "--reason", "the grade stands"}, {"merge", "close", "--seat-id", "red-merge-r1", "--id", "R1-1", "--as", "closed", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "closed on the merits"}},
			act:   seatStep{"merge", "verdict", "--seat-id", "red-merge-r1", "--as", "PASS"},
			guards: "The gate above must not become unpassable. A check that no legitimate run can " +
				"satisfy is removed by the first person it blocks.",
		},
		{
			name:    "a direction ruling names a line of inquiry that exists",
			act:     seatStep{"motion", "inquiry", "rule", "--seat-id", "red-merge-r1", "--id", "Q9", "--as", "endorsed", "--reason", "ruling a line nobody proposed"},
			refused: "which no line of inquiry event proposed",
			guards: "A direction joins on the LINE's own id, so the dangling-reference discipline " +
				"has to reach a different id space than the other two subjects.",
		},
		{
			name:  "any seat may appeal, not only the filer",
			setup: []seatStep{propose, {"motion", "inquiry", "rule", "--seat-id", "red-merge-r1", "--id", "Q1", "--as", "out_of_scope", "--reason", "not this question"}},
			act:   seatStep{"motion", "inquiry", "appeal", "--seat-id", "red-lens-r1-L1", "--id", "Q1", "--reason", "a lens presses a direction blue proposed"},
			guards: "STANDING IS OPEN ON PURPOSE and this pins it. A motion belongs to the run, not " +
				"to its filer — a lens files a safety petition and it is BLUE the granted relief " +
				"binds. The code comment used to say `the filer` while the code checked nobody, so " +
				"the doc was the wrong half.",
		},
		{
			name:    "a grade motion names a gap that exists",
			act:     seatStep{"motion", "grade", "file", "--seat-id", "blue-respond-r1", "--id", "R9-9", "--dimension", "severity", "--proposed", "low", "--reason", "contesting a gap nobody minted"},
			refused: "no mint event created",
			guards: "Lost outright when the verb was replaced: `blue dispute` checked it and the " +
				"additive `motion grade file` did not, which nothing could see while both were live.",
		},
		{
			name:    "a grade motion is refused on a gap already disposed of",
			setup:   []seatStep{mint, {"merge", "close", "--seat-id", "red-merge-r1", "--id", "R1-1", "--as", "closed", "--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./x", "--reason", "closed on the merits"}},
			act:     seatStep{"motion", "grade", "file", "--seat-id", "blue-respond-r1", "--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "contesting a settled grade"},
			refused: "disposition has already been made",
			guards:  "The other check lost in the replacement.",
		},
		{
			name:    "a dimension outside the four is refused",
			setup:   []seatStep{mint},
			act:     seatStep{"motion", "grade", "file", "--seat-id", "blue-respond-r1", "--id", "R1-1", "--dimension", "vibes", "--proposed", "low", "--reason", "contesting an axis that does not exist"},
			refused: "--dimension must be one of severity|likelihood|impact|complexity",
			guards: "The ruling is matched to the filing on (gap, dimension), so an axis outside the " +
				"four is a motion filed against nothing. `blue dispute` enum-checked it; the " +
				"replacement took any string until the old verb was deleted and the two compared. " +
				"The refusal now happens at PARSE rather than at the write — strictly earlier, and " +
				"it arrives with the menu, so a seat is told what the four axes are FOR.",
		},
		{
			name:    "a non-grade --proposed is refused at parse",
			setup:   []seatStep{mint},
			act:     seatStep{"motion", "grade", "file", "--seat-id", "blue-respond-r1", "--id", "R1-1", "--dimension", "severity", "--proposed", "quite-bad", "--reason", "proposing a grade that is not one"},
			refused: "quite-bad",
			guards: "`--proposed` is a pflag.Value, so this fails BEFORE any RunE runs and the help " +
				"and the refusal come from one list. The replacement registered it as a bare string.",
		},
		{
			name:    "a petition class outside the four is refused",
			act:     seatStep{"motion", "petition", "file", "--seat-id", "red-lens-r1-L1", "--class", "aesthetic", "--relief", "halt", "--reason", "petitioning on a standard nobody wrote"},
			refused: "--class must be one of ethical|safety|integrity|constitutional",
			guards: "The bench is convened PER CLASS; a fifth is a petition heard under whichever " +
				"standard the ruling seat happened to imagine. Refused at PARSE now, with the " +
				"four classes and what each is for — the seat learns the vocabulary from the refusal.",
		},
		{
			name:    "the gavel holds even with the right subject",
			setup:   []seatStep{mint, fileGrade},
			act:     seatStep{"motion", "grade", "rule", "--seat-id", "blue-respond-r1", "--id", "M1", "--as", "accepted", "--reason", "ruling my own motion"},
			refused: "is ruled by the merge seat",
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
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, tc := range adversarialCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runDir := adversarialRun(t)

			for i, step := range tc.setup {
				args := append(append([]string{}, step...), "--run", runDir)
				if _, err := run(t, args...); err != nil {
					t.Fatalf("setup step %d (%v) failed: %v\n\nA scenario whose setup fails asserts its outcome against a board that was never built, which passes for the wrong reason.", i, step, err)
				}
			}

			args := append(append([]string{}, tc.act...), "--run", runDir)
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
	runDir := t.TempDir()
	for _, s := range []struct{ role, id string }{
		{"lens", "red-lens-r1-L1"},
		{"merge", "red-merge-r1"},
		{"blue", "blue-respond-r1"},
		{"bench", "judge-r1"},
	} {
		if _, err := run(t, s.role, "register", "--run", runDir, "--seat-id", s.id); err != nil {
			t.Fatalf("register %s: %v", s.id, err)
		}
	}
	seedBlueReport(t, runDir)
	return runDir
}
