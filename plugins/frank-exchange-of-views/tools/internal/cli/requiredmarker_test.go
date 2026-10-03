package cli

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
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
// The marker is read through seat.IsMarked — the one reader of the shape every writer produces —
// so this gate and helpcontract_test.go cannot disagree about what counts. "REQUIRED to close a
// gap red minted with --check-kind computation" mid-sentence on `prove --answers` is a condition,
// not a marker, and `mint --fix`'s "the required fix" is a field name.
//
// EVERY LEAF A SEAT CAN RUN IS IN THE TABLE OR EXEMPTED WITH A REASON. The first table was
// corrRows plus a hand-typed dozen, and `verdict --as` and `fetch --url` — refused when omitted,
// unmarked — were neither probed nor exempt, so the gate was green over the defect it exists for.
// TestEveryRunnableLeafIsProbedForItsMarkers is the census.

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
	probes["fetch"] = requiredProbe{seat: lensSeat,
		setup: func(t *testing.T, runDir string) corrVars {
			withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/f": []byte("<html>the source</html>")}})
			return nil
		},
		args: fixed("fetch", "--url", "https://src/f")}
	probes["chair verdict"] = requiredProbe{seat: "red-chair", args: fixed("verdict", "--as", "FAIL")}
	// The fixture registers every seat; a second register is the bench's own again, under the
	// occasion the bench alone is asked for.
	probes["bench register"] = requiredProbe{seat: "judge", args: fixed("register", "--occasion", "docket")}
	probes["blue register"] = requiredProbe{seat: "blue-respond", args: fixed("register")}
	return probes
}

// unprobedLeaves are the runnable seat leaves with local flags that no probe drives, each with
// the reason. A leaf here is a leaf the gate does not measure, named rather than left looking
// covered.
var unprobedLeaves = map[string]string{
	"show":    "a read: its --id/--match/--quote narrow a projection and none is required",
	"inquest": "a read: --match/--quote narrow the record and neither is required",
}

// TestEveryRunnableLeafIsProbedForItsMarkers holds the probe table to the tree: every runnable
// verb on any seat's surface that takes a local flag is driven by requiredProbes or named in
// unprobedLeaves. Keyed on the verb's PATH, not the seat: a verb on several surfaces is built by
// one constructor, so one seat's probe measures its flags for all of them.
func TestEveryRunnableLeafIsProbedForItsMarkers(t *testing.T) {
	probed := map[string]bool{}
	for _, p := range requiredProbes() {
		probed[strings.Join(pathOf(p.args(nil)), " ")] = true
	}
	var seen int
	for role, r := range AllRoots() {
		if !isSeatRole(role) {
			continue
		}
		walk(r, func(c *cobra.Command, path []string) {
			if !c.Runnable() {
				return
			}
			var local []string
			c.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
				if f.Name != "help" {
					local = append(local, "--"+f.Name)
				}
			})
			if len(local) == 0 {
				return
			}
			seen++
			verb := strings.Join(path, " ")
			if probed[verb] || unprobedLeaves[path[0]] != "" {
				return
			}
			t.Errorf("%s: `%s` takes %s and no probe drives it — add it to requiredProbes, or to unprobedLeaves with the reason.\n\nA verb outside the table can refuse an unmarked flag and the gate stays green.", role, verb, strings.Join(local, " "))
		})
	}
	if seen < 40 {
		t.Fatalf("only %d leaves with local flags seen — the walk is not reaching the tree", seen)
	}
}

// TestNoUsageCarriesTwoMarkers holds the writers to one marker per flag. The schema's walk
// (markRequired) and a verb's own Require can both reach a flag; before they shared one
// idempotent writer the page read "REQUIRED — REQUIRED — …" wherever both did.
func TestNoUsageCarriesTwoMarkers(t *testing.T) {
	var checked int
	for role, r := range AllRoots() {
		walk(r, func(c *cobra.Command, path []string) {
			c.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
				checked++
				if p := markerProblem(f.Usage); p != "" {
					t.Errorf("%s: `%s --%s` %s:\n\n%s", role, strings.Join(path, " "), f.Name, p, f.Usage)
				}
			})
		})
	}
	if checked < 100 {
		t.Fatalf("only %d flags checked — the walk is not reaching the tree", checked)
	}
}

// anyMarker is the marker in either shape, ANYWHERE in a usage: bare ("REQUIRED — ") or conditional
// ("REQUIRED unless --accept — "). Counting only the bare spelling let "REQUIRED unless --accept —
// REQUIRED — …" read as one marker.
var anyMarker = regexp.MustCompile(`REQUIRED(?: unless [^—]+)? — `)

// markerProblem is what is wrong with a usage's marker, or "".
func markerProblem(usage string) string {
	if n := len(anyMarker.FindAllStringIndex(usage, -1)); n > 1 {
		return fmt.Sprintf("carries the marker %d times", n)
	}
	if strings.HasPrefix(usage, "REQUIRED") && !seat.IsMarked(usage) {
		return "begins with REQUIRED in a shape IsMarked does not read, so one gate counts it marked and another does not"
	}
	if seat.IsMarked(usage) && strings.TrimSpace(anyMarker.ReplaceAllString(usage, "")) == "" {
		return "says REQUIRED and nothing about what to supply"
	}
	return ""
}

// The marker gate's reader, on the shapes the tree does not hold today: a test of the walk above
// passes on a clean tree whether or not the reader would notice a doubled conditional marker.
func TestMarkerProblemSeesEveryShape(t *testing.T) {
	for usage, wantProblem := range map[string]bool{
		"REQUIRED — the gap id":                            false,
		"REQUIRED unless --accept — the span":              false,
		"the gap id — REQUIRED to close a computation gap": false,
		"REQUIRED — REQUIRED — the gap id":                 true,
		"REQUIRED unless --accept — REQUIRED — the span":   true,
		"REQUIRED — REQUIRED unless --accept — the span":   true,
		"REQUIRED: the gap id":                             true,
		"REQUIRED — ":                                      true,
		"REQUIRED unless --about names the subject — ":     true,
	} {
		if got := markerProblem(usage) != ""; got != wantProblem {
			t.Errorf("markerProblem(%q) found a problem: %v, want %v", usage, got, wantProblem)
		}
	}
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
				marked := seat.IsMarked(f.Usage)
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
					isThisFlag := err != nil && (fullErr == nil || err.Error() != fullErr.Error()) && refusalNames(err, f.Name)
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
				case err != nil && !refusalNames(err, f.Name):
					t.Errorf("`%s` refused without --%s and the refusal does not NAME it:\n\n%v", verb, f.Name, err)
				}
			})
		})
	}
	if checked < 100 {
		t.Fatalf("only %d flags checked — the table is not reaching the tree, and a walk that finds nothing passes forever", checked)
	}
}

// refusalNames says whether a refusal names the flag AS A FLAG — `--id`, not the letters "id"
// inside "valid". It reads the same `--flag` token helpcontract_test.go's assertNamedFlagsExist
// reads; cobra's own `required flag(s) "id" not set`, which quotes the bare name; and cobra's
// group refusals, which list bare names in brackets — `at least one of the flags in the group
// [quote anchor] is required`.
func refusalNames(err error, flag string) bool {
	for _, m := range flagToken.FindAllStringSubmatch(err.Error(), -1) {
		if m[1] == flag {
			return true
		}
	}
	for _, m := range flagGroup.FindAllStringSubmatch(err.Error(), -1) {
		if slices.Contains(strings.Fields(m[1]), flag) {
			return true
		}
	}
	return strings.Contains(err.Error(), `"`+flag+`"`)
}

// flagGroup is cobra's rendering of a flag group in its refusals: bare names, space-separated, in
// brackets.
var flagGroup = regexp.MustCompile(`\[([a-z][a-z-]*(?: [a-z][a-z-]*)*)\]`)

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
