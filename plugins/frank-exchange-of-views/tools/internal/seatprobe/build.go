package seatprobe

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// BUILDING A BOARD, THROUGH WHATEVER RUNS THE TOOL.
//
// The caller supplies Exec — the test drives the in-process cobra root, the harness shells out to
// the real binary — and everything else is shared. That split is the point: a board built one way
// for the test and another way for the probe is two fixtures with one name, and the probe's is the
// one that matters.
//
// EVERY STEP GOES THROUGH THE REAL VERBS. Writing events directly would let this record a state no
// seat could ever have reached — a gap with no acceptance check, a closure with no anchor, a
// ruling on a motion nobody filed — and a probe run against an impossible board teaches nothing
// about the possible one.
//
// EVERY FAILURE IS FATAL, and that is not fastidiousness. The first draft logged a failed `cite`
// and carried on, and two boards built with ZERO cited claims — which is precisely the state their
// expectations are about. `lens verify` had nothing to verify and `blue claim-index` had nothing
// to index, so both would have reported UNMET against a seat with no way to meet them, and the
// report would have read as a finding about the SEAT rather than about the fixture.

// Exec runs one tool invocation and returns its stdout.
type Exec func(args ...string) (string, error)

// ProbeAgentID is the harness handle the probe uses for a seat, at both ends of the binding: the
// register the build writes, and the environment the seat is dispatched with. Derived from the
// seat id so the two cannot drift, and prefixed so a probe record is never mistaken for a
// production agent handle.
func ProbeAgentID(seatID string) string { return "probe-" + seatID }

// Seats are registered on every board: a motion names its filer, a ruling names its ruler, and an
// unregistered seat is refused before any board state exists.
var Seats = []struct{ Role, ID string }{
	{"lens", "red-lens-r1-L1"},
	{"merge", "red-merge-r1"},
	{"blue", "blue-respond-r1"},
	{"bench", "judge-r2"},
}

// Build materialises a board into runDir.
func Build(runDir string, b Board, exec Exec) error {
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(b.Report), 0o644); err != nil {
		return err
	}
	// THE BUILD REGISTERS THESE SEATS AS THE HARNESS, NOT AS THE AGENT THAT WILL HOLD THEM.
	//
	// Staging a board needs registered seats — a motion names its filer, a ruling names its ruler,
	// and an unregistered seat is refused before any board state exists — so these registers have
	// to happen. What must NOT happen is binding them to the handle the seat is later dispatched
	// with: that hands the seat a binding it did not create, and `register` stops being its first
	// act because the guard is already satisfied by work the harness did.
	//
	// MEASURED, WHICH IS WHY THIS CHANGED. In the 2026-08-20 run one seat in nine — judge-r1 on
	// the sitting board — never called `register` at all, made 22 tool calls, and recorded events
	// anyway. In production its first write would have been refused with "register is your first
	// act and it has not happened". The probe could not see that, because the probe had already
	// registered it. An instrument that satisfies the guard it is measuring reports the guard as
	// untested and the seat as compliant, and those look identical from the outside.
	//
	// No handle is set here, so these registers carry no agent_id — which is exactly right: the
	// harness staging a fixture is not a seat, and the record says so by the field's absence.
	for _, s := range Seats {
		if _, err := exec("register", "--run", runDir, "--seat-id", s.ID); err != nil {
			return fmt.Errorf("register %s: %w", s.ID, err)
		}
	}

	// THE BOARD DECLARES ITS CLASS VOCABULARY, because a real run does and this fixture stands in
	// for one. `run-setup` stages the registry on a production run; the probe does not go through
	// setup, so it stages its own.
	//
	// AFTER THE REGISTERS, AND THAT ORDER IS LOAD-BEARING. The probe separates a run's records to a
	// temp root so the seat cannot read the board off disk, and it does that by handing the record
	// ROOT to the tool subprocess in its environment — which this function does not carry. Staged
	// before the first register, `RecordsDir` resolves in THIS process, finds no separation marker,
	// and creates <runDir>/records; the first register subprocess then refuses to separate a run
	// that already keeps events locally. Staged here, the marker exists and the registry lands
	// where the tool will read it.
	//
	// MEASURED TWICE, both loudly. The advisory branch going strict stopped 8 of 9 boards on their
	// first mint; fixing that in the wrong place stopped 9 of 9 on their first register. The probe
	// refused to score either one ("this run is not a result"), which is the only reason both were
	// five-minute diagnoses rather than runs reported with a thinner board.
	if err := record.StageForRun(runDir, BoardClasses...); err != nil {
		return fmt.Errorf("stage the class registry: %w", err)
	}

	for i, g := range b.Gaps {
		if _, err := exec("mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--key", g.Key, "--class", g.Class,
			"--quote", g.Location, "--problem", g.Problem, "--fix", g.Fix,
			"--check", g.Check, "--check-kind", g.CheckKind,
			"--severity", g.Severity, "--likelihood", g.Likelihood,
			"--impact", g.Impact, "--complexity", g.Complexity,
			"--reason", g.Problem+" (baits "+g.Baits+": "+g.Why+")"); err != nil {
			return fmt.Errorf("mint %s: %w", g.Key, err)
		}
		// THE CLOSINGS, WHERE THE BOARD CARRIES THEM. The bench's prompt states its ruling basis
		// is "the two closings, the transcript, and the final state" and that both sides have
		// already filed — so a bench board without them hands the seat a docket it cannot rule on
		// the way it is told to. Measured 2026-08-20 by the seat: the boundary bench filed friction
		// naming the null closings, ruled on artifact state instead, and asked for a human check.
		gapID := fmt.Sprintf("R1-%d", i+1)
		for _, c := range []struct{ seat, text string }{
			{"red-merge-r1", g.RedClosing}, {"blue-respond-r1", g.BlueClosing},
		} {
			if c.text == "" {
				continue
			}
			if _, err := exec("closing", "--run", runDir, "--seat-id", c.seat,
				"--id", gapID, "--reason", c.text); err != nil {
				return fmt.Errorf("closing %s by %s: %w", gapID, c.seat, err)
			}
		}
		if !g.Closed {
			continue
		}
		// A CLOSED gap so the archive is not empty: `spot-check` against an empty one has nothing
		// to sample, so a board that wants the duty exercised has to give it something.
		if _, err := exec("close", "--run", runDir, "--seat-id", "red-merge-r1",
			"--id", gapID, "--as", "repaired", "--verified-by", "L1", "--verified-with", "git show",
			"--verified-against", "HEAD:config", "--reason", "verified at the leaf against the pinned config"); err != nil {
			return fmt.Errorf("close %s: %w", gapID, err)
		}
	}

	// RED'S ROUND-1 NARRATIVE, so the transcript blue is sent to read exists. Filed AFTER the gaps
	// it accounts for and BEFORE anything that answers it, which is the order a real round has.
	if b.RedNarrative != "" {
		if _, err := exec("position", "--run", runDir, "--seat-id", "red-merge-r1",
			"--reason", b.RedNarrative); err != nil {
			return fmt.Errorf("red narrative: %w", err)
		}
	}

	for i, a := range b.Inquiries {
		// THE ID COMES BACK FROM THE TOOL; IT IS NOT RECOMPOSED HERE.
		//
		// This used to propose the line, throw the output away, and rebuild the id as
		// `fmt.Sprintf("A%d", i+1)` — a guess that the minter counts the same things this loop
		// does. It held only while both agreed, and it stopped holding the moment the id moved
		// from A<n> to Q<n>: every probe board with a ruled line failed to build, and the probe
		// reports a failed board as a harness fault rather than as a fixture speaking a retired
		// model. That is `facts-are-fields` in a fixture — the record MINTS the id and says so on
		// stdout, and composing it from a loop index is a hope about someone else's counter.
		out, err := exec("line-of-inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--reason", a.Line, "--hypothesis", a.Hypothesis)
		if err != nil {
			return fmt.Errorf("line of inquiry %d: %w", i+1, err)
		}
		if a.Ruled == "" {
			continue
		}
		id := mintedInquiryID(out)
		if id == "" {
			return fmt.Errorf("line of inquiry %d: the tool did not report a minted id in %q — the ruling "+
				"below needs the id the RECORD assigned, and guessing one is how this broke before", i+1, out)
		}
		if _, err := exec("motion", "inquiry", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
			"--id", id, "--as", a.Ruled,
			"--reason", rulingReason(a.RuledWhy, a.Ruled)); err != nil {
			return fmt.Errorf("rule %s: %w", id, err)
		}
	}

	for i, m := range b.Motions {
		args := []string{"motion", m.Subject, "file", "--run", runDir, "--seat-id", m.Filer,
			"--reason", m.Basis}
		switch m.Subject {
		case "grade":
			args = append(args, "--id", m.GapID, "--dimension", m.Dimension, "--proposed", m.Proposed)
		case "petition":
			args = append(args, "--class", m.Class, "--relief", m.Relief)
		}
		if _, err := exec(args...); err != nil {
			return fmt.Errorf("motion %d (%s): %w", i+1, m.Subject, err)
		}
		if m.Ruled == "" {
			continue
		}
		ruler := map[string]string{"grade": "red-merge-r1", "petition": "judge-r2"}[m.Subject]
		if _, err := exec("motion", m.Subject, "rule", "--run", runDir, "--seat-id", ruler,
			"--id", fmt.Sprintf("M%d", i+1), "--as", m.Ruled,
			"--reason", rulingReason(m.RuledWhy, m.Ruled)); err != nil {
			return fmt.Errorf("rule motion %d: %w", i+1, err)
		}
	}

	for i, pr := range b.Proofs {
		script := filepath.Join(runDir, fmt.Sprintf("probe-proof-%d.py", i+1))
		if err := os.WriteFile(script, []byte(pr.Script+"\n"), 0o644); err != nil {
			return err
		}
		args := []string{"prove", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--quote", pr.Location, "--script", script,
			"--reason", "the computation behind this sentence"}
		if pr.Answers != "" {
			args = append(args, "--answers", pr.Answers)
		}
		if _, err := exec(args...); err != nil {
			return fmt.Errorf("prove %d: %w", i+1, err)
		}
	}

	for i, claim := range b.Claims {
		// A REACHABLE url, because `cite` FETCHES and caches. An unreachable one is refused and
		// logged as friction, which is correct behaviour and useless here.
		if _, err := exec("cite", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--key", fmt.Sprintf("C%d", i+1), "--quote", claim,
			"--title", "the pinned source", "--url", "https://example.com/",
			"--reason", "the source this claim rests on"); err != nil {
			return fmt.Errorf("cite %d (%q): %w — the board declares this claim and its expectations are about acting on it; building without it would produce a board whose demands cannot be met", i+1, claim, err)
		}
	}
	return nil
}

// rulingReason is the ruler's argument, and it refuses to invent one.
//
// The boilerplate it replaces ("ruled <verdict> on the line as it was proposed") was measured, by
// asking a seat: it read the motions view twice and reported that it could not find red's
// reasoning, which was true. A board that scores whether a seat appeals a ruling must let the seat
// READ the ruling, or it is scoring a guess.
//
// A missing RuledWhy is loud rather than papered over. The alternative — falling back to the old
// boilerplate — would put the defect back exactly where it was, in a field that looks filled.
func rulingReason(why, verdict string) string {
	if strings.TrimSpace(why) == "" {
		return "NO ARGUMENT RECORDED — this board ruled " + verdict + " without saying why, which is a " +
			"defect in the fixture rather than a position the ruler took. Treat it as unreadable."
	}
	return why
}

// mintedInquiryID reads the id out of the propose verb's own confirmation
// ("line of inquiry Q1 recorded (proposed): …"). It returns "" rather than guessing, so a
// changed message surfaces as the explicit failure above instead of a wrong id reaching a ruling.
func mintedInquiryID(out string) string {
	m := mintedID.FindStringSubmatch(out)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

var mintedID = regexp.MustCompile(`\b(Q\d+)\b`)

// BoardClasses is the gap-class vocabulary the staged boards mint under.
//
// Every entry is a slug from the shipped registry (feov-memory/class-registry.json), not a
// placeholder: a board is a fixture for a REAL sitting, and a seat that opens `class` should find
// the taxonomy its gaps are actually filed under rather than a set invented for the harness.
var BoardClasses = []string{
	"citation-status-drift", "claim-contradicts-own-record", "cross-section-contradiction",
	"derivation-status-overclaim", "doctrine-vs-implementation", "false-universal",
	"figure-recount-fails", "incomplete-repair-propagation", "metric-conflation",
	"risk-coverage-omission", "self-attestation", "unverified-composition",
	"verification-scope-blindspot",
}
