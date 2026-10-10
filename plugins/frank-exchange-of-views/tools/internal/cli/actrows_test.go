package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ONE WORKING ACT PER RECORDING COMMAND, driven through the command line: the invocation, the seat
// that makes it, and whatever the run must hold first.
//
// KEYED BY PATH, not by verb: `log` under four roles and `position`/`closing` under two are distinct
// commands with distinct handlers. The required-marker gate reads the rows to run each verb clean,
// and the answers gate reads them to file the act a refusal's row then answers.

type actVars map[string]string

type actRow struct {
	seat  string
	setup func(t *testing.T, runDir string) actVars
	// act is the invocation, less --run and --seat-id, with text as its free-text payload.
	act func(v actVars, text string) []string
}

func actFixture(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	for _, id := range []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"} {
		// The bench owes an OCCASION and no other seat may pass one — ask the same question the
		// write path asks rather than keeping a list of which ids are the bench.
		args := []string{"register", "--run", runDir, "--seat-id", id}
		if record.SeatOwesOccasion(id) {
			args = append(args, "--occasion", "docket")
		}
		if _, err := run(t, args...); err != nil {
			t.Fatalf("register %s: %v", id, err)
		}
	}
	seedBlueReport(t, runDir)
	if g := mintGap(t, runDir, "corr-seed", "self-attestation"); g != onlyID(t, runDir, "G") {
		t.Fatalf("the fixture mints %s and the run holds another gap — every row names the one gap as G1", g)
	}
	return runDir
}

func must(t *testing.T, runDir string, args ...string) string {
	t.Helper()
	out, err := runAt(t, append(args, "--run", runDir)...)
	if err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return out
}

// motionID files a motion and reads its id off the envelope's field.
func motionID(t *testing.T, runDir string, args ...string) string {
	t.Helper()
	out := must(t, runDir, append(args, "--json")...)
	var env struct {
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.MotionID == "" {
		t.Fatalf("no motion id on the filing's envelope (%v): %s", err, out)
	}
	return env.Result.MotionID
}

func closeG1(t *testing.T, runDir, text string) {
	t.Helper()
	must(t, runDir, "close", "--seat-id", lensSeat, "--id", "G1", "--as", "repaired", "--verified-by", "L1",
		"--verified-with", "read", "--verified-against", "blue/report.md", "--reason", text)
}

func fileGradeMotion(t *testing.T, runDir string) string {
	return motionID(t, runDir, "motion", "grade", "file", "--seat-id", "blue-respond", "--id", "G1",
		"--dimension", "severity", "--proposed", "low", "--reason", "the consequence is bounded")
}

func proposeQ1(t *testing.T, runDir string) {
	t.Helper()
	must(t, runDir, "avenue", "propose", "--seat-id", "blue-respond", "--reason", "a seeded line")
}

func actRows() map[string]actRow {
	prose := func(verb ...string) func(actVars, string) []string {
		return func(_ actVars, text string) []string { return append(append([]string{}, verb...), "--reason", text) }
	}
	logRow := func(seatID string) actRow {
		return actRow{seat: seatID, act: func(_ actVars, text string) []string {
			return []string{"log", "--type", "defect", "--reason", text}
		}}
	}
	closingRow := func(seatID string) actRow {
		return actRow{seat: seatID, act: func(_ actVars, text string) []string {
			return []string{"closing", "--id", "G1", "--reason", text}
		}}
	}
	return map[string]actRow{
		"blue manifest-row": {seat: "blue-respond", act: func(_ actVars, text string) []string {
			return []string{"manifest-row", "--id", "G1", "--reason", text}
		}},
		"blue cite": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars {
				withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/corr": []byte("<html>a source on the finding</html>")}})
				return nil
			},
			act: func(_ actVars, text string) []string {
				return []string{"cite", "--quote", "§2 the finding prose lands in a quoted sentence.", "--url", "https://src/corr", "--title", text,
					"--reason", "it states the finding"}
			}},
		"blue prove": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars {
				return actVars{"script": script(t, runDir, "corr.js", "console.log(91 % 7)")}
			},
			act: func(v actVars, text string) []string {
				return []string{"prove", "--quote", "the parser accepts an empty body in this line.", "--script", v["script"], "--reason", text}
			}},
		"blue avenue propose": {seat: "blue-respond", act: func(_ actVars, text string) []string {
			return []string{"avenue", "propose", "--reason", text, "--hypothesis", "it would settle something", "--method", "a source class"}
		}},
		"blue avenue move": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars { proposeQ1(t, runDir); return nil },
			act: func(_ actVars, text string) []string {
				return []string{"avenue", "move", "--id", "Q1", "--as", "pursued", "--reason", text}
			}},
		"blue log":       logRow("blue-respond"),
		"lens log":       logRow(lensSeat),
		"chair log":      logRow("red-chair"),
		"bench log":      logRow("judge"),
		"blue position":  {seat: "blue-respond", act: prose("position")},
		"chair position": {seat: "red-chair", act: prose("position")},
		"blue closing":   closingRow("blue-respond"),
		"chair closing":  closingRow("red-chair"),
		"lens regrade": {seat: lensSeat, act: func(_ actVars, text string) []string {
			return []string{"regrade", "--id", "G1", "--severity", "high", "--reason", text}
		}},
		// A proof whose output differs on every run.
		"lens reproduce": {seat: lensSeat,
			setup: func(t *testing.T, runDir string) actVars {
				s := filepath.Join(runDir, "unstable.js")
				if err := os.WriteFile(s, []byte("console.log(Math.random());"), 0o644); err != nil {
					t.Fatal(err)
				}
				must(t, runDir, "prove", "--seat-id", "blue-respond", "--quote", "a quoted sentence",
					"--script", s, "--reason", "the computation a lens re-runs", "--expect-error")
				return actVars{"sha": lastBody(t, runDir, &recordpb.Proof{}).GetProofSha()}
			},
			act: func(v actVars, text string) []string {
				return []string{"reproduce", "--id", v["sha"], "--as", "sound", "--reason", text}
			}},
		"lens close": {seat: lensSeat, act: func(_ actVars, text string) []string {
			return []string{"close", "--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
				"--verified-against", "blue/report.md", "--reason", text}
		}},
		"chair carry": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) actVars {
				closeG1(t, runDir, "closed in the first sitting")
				must(t, runDir, "register", "--seat-id", "red-chair")
				return nil
			},
			act: func(_ actVars, text string) []string {
				return []string{"carry", "--id", "G1", "--carried-from", "1", "--as", "repaired", "--reason", text}
			}},
		"chair spot-check": {seat: "red-chair", act: func(_ actVars, text string) []string {
			return []string{"spot-check", "--none", "--reason", text}
		}},
		"chair avenue review": {seat: "red-chair", act: prose("avenue", "review")},
		"bench declare":       {seat: "judge", act: prose("declare")},
		"bench certify":       {seat: "judge", act: prose("certify")},
		"bench halt":          {seat: "judge", act: prose("halt")},
		"bench outcome": {seat: "judge", act: func(_ actVars, text string) []string {
			return []string{"outcome", "--as", "UNVERIFIED", "--reason", text}
		}},
		"motion grade rule": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) actVars { return actVars{"m": fileGradeMotion(t, runDir)} },
			act: func(v actVars, text string) []string {
				return []string{"motion", "grade", "rule", "--id", v["m"], "--as", "rejected", "--reason", text}
			}},
		"motion petition rule": {seat: "judge",
			setup: func(t *testing.T, runDir string) actVars {
				return actVars{"m": motionID(t, runDir, "motion", "petition", "file", "--seat-id", lensSeat,
					"--class", "safety", "--relief", "halt before the next round", "--reason", "a consent gate is missing")}
			},
			act: func(v actVars, text string) []string {
				return []string{"motion", "petition", "rule", "--id", v["m"], "--as", "denied", "--reason", text}
			}},
		"motion avenue rule": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) actVars { proposeQ1(t, runDir); return nil },
			act: func(_ actVars, text string) []string {
				return []string{"motion", "avenue", "rule", "--id", "Q1", "--as", "endorsed", "--reason", text}
			}},
		"motion docket rule": {seat: "judge",
			setup: func(t *testing.T, runDir string) actVars {
				return actVars{"m": docketFile(t, runDir, "red-chair", "G1", "red cannot settle G1")}
			},
			act: func(v actVars, text string) []string {
				return []string{"motion", "docket", "rule", "--id", v["m"], "--as", "remanded",
					"--principle", "thoroughness over speed", "--tension", "cost against certainty",
					"--review-flag", "none", "--settled", "nothing yet", "--reopens-on", "a reproduction", "--reason", text}
			}},
		"motion grade appeal": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars {
				m := fileGradeMotion(t, runDir)
				must(t, runDir, "motion", "grade", "rule", "--seat-id", "red-chair", "--id", m, "--as", "rejected", "--reason", "the evidence does not reach it")
				return actVars{"m": m}
			},
			act: func(v actVars, text string) []string {
				return []string{"motion", "grade", "appeal", "--id", v["m"], "--reason", text}
			}},
		"motion avenue appeal": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars {
				proposeQ1(t, runDir)
				must(t, runDir, "motion", "avenue", "rule", "--seat-id", "red-chair", "--id", "Q1", "--as", "out_of_scope", "--reason", "not this question")
				return nil
			},
			act: func(_ actVars, text string) []string {
				return []string{"motion", "avenue", "appeal", "--id", "Q1", "--reason", text}
			}},
	}
}
