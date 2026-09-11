package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY CORRECTABLE COMMAND ACCEPTS ONE CORRECTION, DRIVEN THROUGH THE COMMAND LINE
// (plans/same-sitting-correction.md III.C.4, layer 3).
//
// A guard that refuses because "this already happened" would refuse a correction of the act that
// happened — a closure's correction names a closed gap, a ruling's correction a ruled motion. The
// record layer threads the target into the guards it knows about; this is the proof that the census
// missed none, because it runs every correctable command's own handler end to end.
//
// KEYED BY PATH, not by verb: `log` under four roles and `position`/`closing` under two are distinct
// commands with distinct handlers. A correctable path with no row here FAILS, so a newly wired verb
// cannot ship without one.

type corrVars map[string]string

type corrRow struct {
	seat  string
	setup func(t *testing.T, runDir string) corrVars
	// act is the invocation, less --run and --seat-id, with text as its free-text payload. The
	// defective act and its correction differ ONLY in text.
	act func(v corrVars, text string) []string
}

func corrFixture(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	for _, id := range []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", id); err != nil {
			t.Fatalf("register %s: %v", id, err)
		}
	}
	seedBlueReport(t, runDir)
	if g := mintGap(t, runDir, "corr-seed", "self-attestation"); g != "G1" {
		t.Fatalf("the fixture's gap is %s, want G1", g)
	}
	return runDir
}

func must(t *testing.T, runDir string, args ...string) string {
	t.Helper()
	out, err := run(t, append(args, "--run", runDir)...)
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
	must(t, runDir, "line-of-inquiry", "propose", "--seat-id", "blue-respond", "--reason", "a seeded line")
}

func corrRows() map[string]corrRow {
	prose := func(verb ...string) func(corrVars, string) []string {
		return func(_ corrVars, text string) []string { return append(append([]string{}, verb...), "--reason", text) }
	}
	logRow := func(seatID string) corrRow {
		return corrRow{seat: seatID, act: func(_ corrVars, text string) []string {
			return []string{"log", "--type", "defect", "--reason", text}
		}}
	}
	closingRow := func(seatID string) corrRow {
		return corrRow{seat: seatID, act: func(_ corrVars, text string) []string {
			return []string{"closing", "--id", "G1", "--reason", text}
		}}
	}
	return map[string]corrRow{
		"blue manifest-row": {seat: "blue-respond", act: func(_ corrVars, text string) []string {
			return []string{"manifest-row", "--id", "G1", "--reason", text}
		}},
		"blue revision": {seat: "blue-respond", act: prose("revision")},
		// F13: the correction keeps the line's id instead of minting a new one.
		"blue line-of-inquiry propose": {seat: "blue-respond", act: func(_ corrVars, text string) []string {
			return []string{"line-of-inquiry", "propose", "--reason", text, "--hypothesis", "it would settle something"}
		}},
		"blue line-of-inquiry move": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) corrVars { proposeQ1(t, runDir); return nil },
			act: func(_ corrVars, text string) []string {
				return []string{"line-of-inquiry", "move", "--id", "Q1", "--as", "pursued", "--reason", text}
			}},
		"blue log":       logRow("blue-respond"),
		"lens log":       logRow(lensSeat),
		"merge log":      logRow("red-chair"),
		"bench log":      logRow("judge"),
		"blue position":  {seat: "blue-respond", act: prose("position")},
		"merge position": {seat: "red-chair", act: prose("position")},
		"blue closing":   closingRow("blue-respond"),
		"merge closing":  closingRow("red-chair"),
		"lens regrade": {seat: lensSeat, act: func(_ corrVars, text string) []string {
			return []string{"regrade", "--id", "G1", "--severity", "high", "--reason", text}
		}},
		// F13: a proof whose output differs on every run. A correction that re-ran it would record
		// different outputs and be refused for changing a frozen field.
		"lens reproduce": {seat: lensSeat,
			setup: func(t *testing.T, runDir string) corrVars {
				s := filepath.Join(runDir, "unstable.js")
				if err := os.WriteFile(s, []byte("console.log(Math.random());"), 0o644); err != nil {
					t.Fatal(err)
				}
				must(t, runDir, "prove", "--seat-id", "blue-respond", "--quote", "a quoted sentence",
					"--script", s, "--reason", "the computation a lens re-runs", "--expect-error")
				return corrVars{"sha": lastBody(t, runDir, &recordpb.Proof{}).GetProofSha()}
			},
			act: func(v corrVars, text string) []string {
				return []string{"reproduce", "--id", v["sha"], "--as", "sound", "--reason", text}
			}},
		// Layer 1: the correction of a closure names a gap its own target closed.
		"lens close": {seat: lensSeat, act: func(_ corrVars, text string) []string {
			return []string{"close", "--id", "G1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
				"--verified-against", "blue/report.md", "--reason", text}
		}},
		"merge carry": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) corrVars {
				closeG1(t, runDir, "closed in the first sitting")
				must(t, runDir, "register", "--seat-id", "red-chair")
				return nil
			},
			act: func(_ corrVars, text string) []string {
				return []string{"carry", "--id", "G1", "--carried-from", "1", "--as", "repaired", "--reason", text}
			}},
		"merge spot-check": {seat: "red-chair", act: func(_ corrVars, text string) []string {
			return []string{"spot-check", "--none", "--reason", text}
		}},
		"merge inquiry-support": {seat: "red-chair", act: prose("inquiry-support")},
		"bench declare":         {seat: "judge", act: prose("declare")},
		"bench certify":         {seat: "judge", act: prose("certify")},
		"bench halt":            {seat: "judge", act: prose("halt")},
		// F13: the verdict's basis and reasoning are the corrected act's, not re-derived.
		"bench outcome": {seat: "judge", act: func(_ corrVars, text string) []string {
			return []string{"outcome", "--as", "UNVERIFIED", "--reason", text}
		}},
		"motion grade rule": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) corrVars { return corrVars{"m": fileGradeMotion(t, runDir)} },
			act: func(v corrVars, text string) []string {
				return []string{"motion", "grade", "rule", "--id", v["m"], "--as", "rejected", "--reason", text}
			}},
		// Layer 2: the first-wins guard passes a correction of the ruling it finds.
		"motion petition rule": {seat: "judge",
			setup: func(t *testing.T, runDir string) corrVars {
				return corrVars{"m": motionID(t, runDir, "motion", "petition", "file", "--seat-id", lensSeat,
					"--class", "safety", "--relief", "halt before the next round", "--reason", "a consent gate is missing")}
			},
			act: func(v corrVars, text string) []string {
				return []string{"motion", "petition", "rule", "--id", v["m"], "--as", "denied", "--reason", text}
			}},
		"motion inquiry rule": {seat: "red-chair",
			setup: func(t *testing.T, runDir string) corrVars { proposeQ1(t, runDir); return nil },
			act: func(_ corrVars, text string) []string {
				return []string{"motion", "inquiry", "rule", "--id", "Q1", "--as", "endorsed", "--reason", text}
			}},
		// OPEN-1: a docket ruling's settled/reopens-on/review-flag are prose, correctable until relied on.
		"motion docket rule": {seat: "judge",
			setup: func(t *testing.T, runDir string) corrVars {
				return corrVars{"m": docketFile(t, runDir, "red-chair", "G1", "red cannot settle G1")}
			},
			act: func(v corrVars, text string) []string {
				return []string{"motion", "docket", "rule", "--id", v["m"], "--as", "carried",
					"--principle", "thoroughness over speed", "--tension", "cost against certainty",
					"--review-flag", "none", "--settled", "nothing yet", "--reopens-on", "a reproduction", "--reason", text}
			}},
		// Layer 2: the appeal guard passes a correction of the appeal it finds.
		"motion grade appeal": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) corrVars {
				m := fileGradeMotion(t, runDir)
				must(t, runDir, "motion", "grade", "rule", "--seat-id", "red-chair", "--id", m, "--as", "rejected", "--reason", "the evidence does not reach it")
				return corrVars{"m": m}
			},
			act: func(v corrVars, text string) []string {
				return []string{"motion", "grade", "appeal", "--id", v["m"], "--reason", text}
			}},
		"motion inquiry appeal": {seat: "blue-respond",
			setup: func(t *testing.T, runDir string) corrVars {
				proposeQ1(t, runDir)
				must(t, runDir, "motion", "inquiry", "rule", "--seat-id", "red-chair", "--id", "Q1", "--as", "out_of_scope", "--reason", "not this question")
				return nil
			},
			act: func(_ corrVars, text string) []string {
				return []string{"motion", "inquiry", "appeal", "--id", "Q1", "--reason", text}
			}},
	}
}

// keyOf runs the act with --json and reads the key off the envelope's field.
func correctionKeyOf(t *testing.T, runDir, seatID string, args []string) string {
	t.Helper()
	out, err := run(t, append(args, "--run", runDir, "--seat-id", seatID, "--json")...)
	if err != nil {
		t.Fatalf("the defective act was refused: %v", err)
	}
	var env struct {
		OK  bool   `json:"ok"`
		Key string `json:"key"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || !env.OK || env.Key == "" {
		t.Fatalf("the act's envelope carries no key (%v): %s", err, out)
	}
	return env.Key
}

func TestCorrectionAcceptedPerCorrectablePath(t *testing.T) {
	rows := corrRows()
	byPath := commandsByPath()
	var paths []string
	for p, c := range byPath {
		if seat.IsCorrectable(c) {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		if _, ok := rows[p]; !ok {
			t.Errorf("%s takes a correction and has no row here — drive one through the command line before it ships", p)
		}
	}
	for p := range rows {
		if c, ok := byPath[p]; !ok || !seat.IsCorrectable(c) {
			t.Errorf("the row %q names no correctable command", p)
		}
	}
	if len(paths) < 28 {
		t.Fatalf("only %d correctable paths — the walk is not seeing the surface", len(paths))
	}
	for _, p := range paths {
		row, ok := rows[p]
		if !ok {
			continue
		}
		t.Run(p, func(t *testing.T) {
			runDir := corrFixture(t)
			var v corrVars
			if row.setup != nil {
				v = row.setup(t, runDir)
			}
			k := correctionKeyOf(t, runDir, row.seat, row.act(v, "the text the shell cut short:  and nothing"))
			args := append(row.act(v, "the text as it was meant, whole"), "--run", runDir, "--seat-id", row.seat,
				"--corrects", k, "--correction-why", "the shell deleted a word")
			out, err := run(t, args...)
			if err != nil {
				t.Fatalf("the correction was refused: %v", err)
			}
			if want := "[key " + k + "~1]"; !strings.Contains(out, want) {
				t.Errorf("the correction's success line does not carry %s: %q", want, out)
			}
			c := lastBody(t, runDir, &recordpb.Correction{})
			if c.GetCorrects() != k || c.GetReplacement() != k+"~1" || c.GetWhy() != "the shell deleted a word" {
				t.Errorf("the correction on the record = %v, want %s struck by %s~1", c, k, k)
			}
		})
	}
}

// F12 AND F13 THROUGH THE COMMAND LINE: a proposal corrected after the line was moved keeps its id.
func TestCorrectingAProposalAfterItsMoveKeepsTheLine(t *testing.T) {
	runDir := corrFixture(t)
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"line-of-inquiry", "propose", "--reason", "try the  method"})
	must(t, runDir, "line-of-inquiry", "move", "--seat-id", "blue-respond", "--id", "Q1", "--as", "pursued", "--reason", "the method held")
	if _, err := run(t, "line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond",
		"--reason", "try the recorded method", "--corrects", k, "--correction-why", "the shell deleted a word"); err != nil {
		t.Fatalf("correcting a moved proposal was refused: %v", err)
	}
	a := lastBody(t, runDir, &recordpb.Avenue{})
	if a.GetAvenueId() != "Q1" || a.GetLine() != "try the recorded method" {
		t.Errorf("the replacement proposal is %s %q, want Q1 with the corrected line", a.GetAvenueId(), a.GetLine())
	}
}

// A CORRECTION FLAG WITHOUT ITS PAIR, and a correction that corrects nothing, are refused.
func TestCorrectionFlagsAreRefusedAlone(t *testing.T) {
	runDir := corrFixture(t)
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"revision", "--reason", "first"})
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"revision", "--reason", "second", "--correction-why", "w"}, "--corrects is not given"},
		{[]string{"revision", "--reason", "second", "--corrects", k}, "requires --correction-why"},
		{[]string{"log", "--type", "defect", "--reason", "x", "--corrects", k, "--correction-why", "w"}, "is a revision, and this command writes a log"},
	} {
		_, err := run(t, append(tc.args, "--run", runDir, "--seat-id", "blue-respond")...)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: got %v, want a refusal saying %q", tc.args, err, tc.want)
		}
	}
}
