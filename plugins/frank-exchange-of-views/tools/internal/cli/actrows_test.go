package cli

import (
	"encoding/json"
	"testing"

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
	runDir := castRun(t)
	castSeats(t, runDir)
	if g := mintGap(t, runDir, "corr-seed", "spec_underspecification"); g != onlyID(t, runDir, "G") {
		t.Fatalf("the fixture mints %s and the run holds another gap — every row names the one gap as G1", g)
	}
	return runDir
}

func must(t *testing.T, runDir string, args ...string) string {
	t.Helper()
	out, err := runAt(t, ahead(args, "--run", runDir)...)
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
	must(t, runDir, "inquiry", "propose", "--seat-id", "blue-respond", "--question", "a seeded question")
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
	// A CLOSING ARGUES A DOCKETING: the write path refuses one on a gap nothing is before the bench on.
	closingRow := func(seatID string) actRow {
		return actRow{seat: seatID,
			setup: func(t *testing.T, runDir string) actVars {
				docketFile(t, runDir, "blue-respond", "G1", "the bench's to settle")
				return nil
			},
			act: func(_ actVars, text string) []string {
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
		"blue compute": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars {
				return actVars{"bundle": script(t, runDir, "corr.js", "console.log(91 % 7)")}
			},
			act: func(v actVars, text string) []string {
				return []string{"compute", "--quote", "the parser accepts an empty body in this line.", "--dir", v["bundle"], "--reason", text, "--", "node", "corr.js"}
			}},
		"blue inquiry propose": {seat: "blue-respond", act: func(_ actVars, text string) []string {
			return []string{"inquiry", "propose", "--question", text, "--hypothesis", "it would settle something"}
		}},
		"blue inquiry record": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) actVars { proposeQ1(t, runDir); return nil },
			act: func(_ actVars, text string) []string {
				return []string{"inquiry", "record", "--id", "Q1", "--as", "underway", "--reason", text, "--method", "a census of the sources"}
			}},
		"blue log":       logRow("blue-respond"),
		"lens log":       logRow(lensSeat),
		"chair log":      logRow("red-chair"),
		"bench log":      logRow("judge"),
		"chair position": {seat: "red-chair", act: prose("position")},
		"blue closing":   closingRow("blue-respond"),
		"lens closing":   closingRow(lensSeat),
		"lens regrade": {seat: lensSeat, act: func(_ actVars, text string) []string {
			return []string{"regrade", "--id", "G1", "--severity", "high", "--reason", text}
		}},
		// A computation whose output differs on every run.
		"lens reproduce": {seat: lensSeat,
			setup: func(t *testing.T, runDir string) actVars {
				s := script(t, runDir, "unstable.js", "console.log(Math.random());")
				must(t, runDir, "compute", "--seat-id", "blue-respond", "--quote", "a quoted sentence",
					"--dir", s, "--reason", "the computation a lens re-runs", "--variance", "a number in [0, 1)", "--", "node", "unstable.js")
				return actVars{"computation": lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()}
			},
			act: func(v actVars, text string) []string {
				return []string{"reproduce", "--anchor", v["computation"], "--as", "sound", "--reason", text}
			}},
		"lens close": {seat: lensSeat, act: func(_ actVars, text string) []string {
			return []string{"close", "--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
				"--verified-against", "blue/report.md", "--reason", text}
		}},
		"chair carry": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) actVars {
				closeG1(t, runDir, "closed in the chair's second sitting (epoch 2)")
				must(t, runDir, "register", "--seat-id", "red-chair")
				return nil
			},
			act: func(_ actVars, text string) []string {
				return []string{"carry", "--id", "G1", "--carried-from", "2", "--as", "repaired", "--reason", text}
			}},
		"chair spot-check": {seat: "red-chair", act: func(_ actVars, text string) []string {
			return []string{"spot-check", "--reason", text}
		}},
		"chair inquiry review": {seat: "red-chair", act: prose("inquiry", "review")},
		"bench declare":        {seat: "judge", act: prose("declare")},
		"bench certify":        {seat: "judge", act: prose("certify")},
		"bench halt":           {seat: "judge", act: prose("halt")},
		// The outcome is the assembly sitting's: the bench registers for it first.
		"bench outcome": {seat: "judge",
			setup: func(t *testing.T, runDir string) actVars {
				must(t, runDir, "register", "--seat-id", "judge", "--occasion", "assemble")
				return nil
			},
			act: func(_ actVars, text string) []string {
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
		"motion inquiry rule": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) actVars { proposeQ1(t, runDir); return nil },
			act: func(_ actVars, text string) []string {
				return []string{"motion", "inquiry", "rule", "--id", "Q1", "--as", "endorsed", "--reason", text}
			}},
		"motion docket rule": {seat: "judge",
			setup: func(t *testing.T, runDir string) actVars {
				return actVars{"m": docketFile(t, runDir, "red-chair", "G1", "red cannot settle G1")}
			},
			act: func(v actVars, text string) []string {
				return []string{"motion", "docket", "rule", "--id", v["m"], "--as", "remanded",
					"--principle", "thoroughness over speed", "--tension", "cost against certainty",
					"--review-flag", "none", "--settled", "nothing yet", "--direction", "a reproduction", "--reason", text}
			}},
	}
}
