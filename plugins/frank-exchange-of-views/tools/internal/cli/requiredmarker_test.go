package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// HELP SAYS REQUIRED EXACTLY WHERE OMITTING THE FLAG IS REFUSED — IN BOTH DIRECTIONS.
//
// helpcontract_test.go holds one direction: a flag the help calls REQUIRED must be refused when
// omitted. Nothing held the other, and the m13/m14 transcripts paid for it: `near-match` without
// --problem twice, `position` without --reason, `motion docket rule` without --reason — each a
// flag the tool refuses without and the help listed as if optional, one wasted call per lesson.
// Twenty-two flags rendered that way at 7b05e7cd, because required-ness is written by five
// mechanisms (the schema's `(sql).required`, seat.ProseRequired, cobra's MarkFlagRequired, a
// hand-typed marker, a check in the verb's own handler) and only the first two write the word a
// seat reads.
//
// This gate drives each verb with a COMPLETE argument set that runs clean, then omits each flag
// in turn against a fresh run and compares what happened with what the help says:
//
//	refused without it, help silent   → the m13 defect: a seat learns the requirement by losing a call
//	ran clean without it, help REQUIRED → a seat types what it does not have to (the halved defect)
//	a flag the clean set never passed  → optional by demonstration, so the help must not mark it
//
// The marker is read the way markRequired writes it: at the head of the usage. "REQUIRED to close
// a gap red minted with --check-kind computation" mid-sentence on `prove --answers` is a condition,
// not a marker, and `mint --fix`'s "the required fix" is a field name.

// requiredProbe is one verb with an argument set that RUNS CLEAN on the correction fixture.
type requiredProbe struct {
	seat  string
	setup func(t *testing.T, runDir string) corrVars
	args  func(v corrVars) []string
	// fullIsRefused says the clean set is expected to be REFUSED for a reason no flag can cure —
	// render-page needs a cached PDF the fixture does not build. Only the omissions are then
	// probed, and only flags the help marks are checked; an unmarked flag cannot be shown optional
	// against a verb that never runs.
	fullIsRefused bool
}

// oneOfAnother are flags the verb accepts in place of each other, with what stands in. Omitting
// one while the other is present runs clean, and omitting both is refused naming both — a
// requirement no single-flag marker states, so these are exempt from the marker check and the
// help carries the alternative in prose (helpcontract_test.go's satisfiedByAnother reads the
// same shape from the other side).
var oneOfAnother = map[string]string{
	"problem":    "reason",
	"reason":     "problem",
	"reopens-on": "final",
	"final":      "reopens-on",
	"ids":        "none",
	"none":       "ids",
}

// verbsWithAOneOf are the verbs on which oneOfAnother applies — --reason is ordinary everywhere
// but `mint`, where it stands in for --problem. A pair exempts a flag only where its partner is
// registered: `motion docket rule` has --final/--reopens-on and no --problem, so its --reason is
// held to the marker like any other.
var verbsWithAOneOf = map[string]bool{"mint": true, "motion docket rule": true, "spot-check": true}

func requiredProbes() map[string]requiredProbe {
	probes := map[string]requiredProbe{}
	// The correctable verbs already carry a working act each; the text is the prose argument.
	for name, row := range corrRows() {
		row := row
		probes[name] = requiredProbe{seat: row.seat, setup: row.setup,
			args: func(v corrVars) []string { return row.act(v, "the argument for this act") }}
	}
	fixed := func(args ...string) func(corrVars) []string { return func(corrVars) []string { return args } }
	probes["lens finding"] = requiredProbe{seat: lensSeat, args: fixed("finding", "--key", "F1",
		"--quote", "§1 first — a finding sits in sec 1 here.",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason", "r")}
	probes["lens near-match"] = requiredProbe{seat: lensSeat, args: fixed("near-match", "--problem", "p",
		"--quote", "§1 first — a finding sits in sec 1 here.")}
	probes["lens mint"] = requiredProbe{seat: lensSeat, args: fixed("mint", "--key", "M-probe",
		"--class", "scope-creep", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--problem", "p", "--fix", "f", "--reason", "r")}
	probes["lens class new"] = requiredProbe{seat: lensSeat, args: fixed("class", "new", "--class", "a-new-class",
		"--definition", "d", "--neighbor", "scope-creep", "--distinguisher", "x", "--material-default", "by_grade")}
	probes["lens render-page"] = requiredProbe{seat: lensSeat, fullIsRefused: true,
		args: fixed("render-page", "--sha", strings.Repeat("a", 64), "--page", "1")}
	probes["lens verify"] = requiredProbe{seat: lensSeat,
		setup: func(t *testing.T, runDir string) corrVars {
			withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/v": []byte("<html>the source</html>")}})
			must(t, runDir, "cite", "--seat-id", "blue-respond", "--quote", "§2 the finding prose lands in a quoted sentence.",
				"--url", "https://src/v", "--title", "T")
			return corrVars{"anchor": firstCiteEvent(t, runDir).GetLabel()}
		},
		args: func(v corrVars) []string {
			return []string{"verify", "--anchor", v["anchor"], "--quote", "§2 the finding prose lands in a quoted sentence.",
				"--as", "supports", "--confidence", "high", "--reason", "r"}
		}}
	probes["lens corroborate"] = requiredProbe{seat: lensSeat,
		setup: func(t *testing.T, runDir string) corrVars {
			withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>the source</html>")}})
			return nil
		},
		args: fixed("corroborate", "--url", "https://src/c", "--title", "T",
			"--quote", "§2 the finding prose lands in a quoted sentence.", "--as", "supports", "--confidence", "high", "--reason", "r")}
	probes["blue edit"] = requiredProbe{seat: "blue-respond", args: fixed("edit",
		"--quote", "the parser accepts an empty body in this line.", "--new", "the parser accepts an empty body on this line.", "--reason", "r")}
	probes["blue retire"] = requiredProbe{seat: "blue-respond", args: fixed("retire",
		"--quote", "a claim that is no longer in the report.", "--reason", "r")}
	probes["motion grade file"] = requiredProbe{seat: "blue-respond", args: fixed("motion", "grade", "file",
		"--id", "G1", "--dimension", "severity", "--proposed", "low", "--reason", "r")}
	probes["motion petition file"] = requiredProbe{seat: lensSeat, args: fixed("motion", "petition", "file",
		"--class", "safety", "--relief", "stop", "--reason", "r")}
	probes["motion docket file"] = requiredProbe{seat: "red-chair", args: fixed("motion", "docket", "file",
		"--id", "G1", "--reason", "r")}
	return probes
}

func TestHelpSaysRequiredExactlyWhereOmissionIsRefused(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	probes := requiredProbes()
	names := make([]string, 0, len(probes))
	for n := range probes {
		names = append(names, n)
	}
	sort.Strings(names)

	var checked int
	for _, name := range names {
		p := probes[name]
		t.Run(name, func(t *testing.T) {
			// A FRESH RUN PER MUTATION. A refusal leaves the record as it was; a clean run does
			// not, and the next omission would then be refused for a reason that has nothing to
			// do with its flag (a duplicate key, a gap already closed).
			var runDir string
			var vars corrVars
			dirty := true
			fresh := func() {
				if !dirty {
					return
				}
				runDir = corrFixture(t)
				vars = nil
				if p.setup != nil {
					vars = p.setup(t, runDir)
				}
				dirty = false
			}
			fresh()
			full := p.args(vars)
			path := pathOf(full)
			c := cmdAt(NewRootFor(p.seat), path)
			if c == nil {
				t.Fatalf("%v is not on %s's surface", path, p.seat)
			}
			invoke := func(args []string) error {
				fresh()
				_, err := run(t, append(append([]string{}, args...), "--run", runDir, "--seat-id", p.seat)...)
				dirty = err == nil
				return err
			}
			fullErr := invoke(full)
			if (fullErr != nil) != p.fullIsRefused {
				if fullErr != nil {
					t.Fatalf("the complete set is refused, so this probe measures nothing: %v", fullErr)
				}
				t.Fatalf("the complete set ran clean on a fixture the probe says cannot satisfy it — drop fullIsRefused")
			}

			verb := strings.Join(path, " ")
			c.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "help" {
					return
				}
				marked := strings.HasPrefix(f.Usage, "REQUIRED")
				checked++
				omitted, present := without(full, f)
				if !present {
					if marked && !p.fullIsRefused {
						t.Errorf("`%s --%s` says REQUIRED and the verb ran clean without it.\n\nusage: %s\n\nA seat reading that supplies a value it did not have to choose.", verb, f.Name, f.Usage)
					}
					return
				}
				if other := oneOfAnother[f.Name]; verbsWithAOneOf[verb] && other != "" && c.Flags().Lookup(other) != nil {
					return // satisfied by --other, which the help states in prose
				}
				err := invoke(omitted)
				if p.fullIsRefused {
					// The clean set is refused too, so a refusal is THIS flag's only when it is a
					// different one that names the flag — cobra's "required flag(s) not set" against
					// the verb's own "no such document".
					isThisFlag := err != nil && (fullErr == nil || err.Error() != fullErr.Error()) && strings.Contains(err.Error(), f.Name)
					if isThisFlag && !marked {
						t.Errorf("`%s` is REFUSED without --%s and its help does not say so.\n\nusage: %s\nrefusal: %v", verb, f.Name, f.Usage, err)
					}
					return
				}
				switch {
				case err != nil && !marked:
					t.Errorf("`%s` is REFUSED without --%s and its help does not say so.\n\nusage: %s\nrefusal: %v\n\nA seat reads the flag as optional, omits it, and loses the call — measured twice on near-match --problem in one run.", verb, f.Name, f.Usage, err)
				case err == nil && marked:
					t.Errorf("`%s --%s` says REQUIRED and the verb ran clean without it.\n\nusage: %s", verb, f.Name, f.Usage)
				case err != nil && !strings.Contains(err.Error(), f.Name):
					t.Errorf("`%s` refused without --%s and the refusal does not NAME it:\n\n%v", verb, f.Name, err)
				}
			})
		})
	}
	if checked < 100 {
		t.Fatalf("only %d flags checked — the table is not reaching the tree, and a walk that finds nothing passes forever", checked)
	}
}

// pathOf is the command path at the head of an argument list: every token before the first flag.
func pathOf(args []string) []string {
	var path []string
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			break
		}
		path = append(path, a)
	}
	return path
}

// without drops one flag — and its value, when the flag takes one — from an argument list, and
// says whether the flag was there to drop.
func without(args []string, f *pflag.Flag) ([]string, bool) {
	var out []string
	present := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--"+f.Name {
			present = true
			if f.NoOptDefVal == "" && i+1 < len(args) {
				i++ // the value
			}
			continue
		}
		out = append(out, args[i])
	}
	return out, present
}
