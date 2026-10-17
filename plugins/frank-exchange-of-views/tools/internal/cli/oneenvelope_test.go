package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/refusalcheck"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// EVERY WAY A CALL CAN BE REFUSED COMES OUT IN ONE SHAPE.
//
// A seat's `jq` reads one envelope whatever refused it: a handler, the record's own validation,
// cobra (an unknown flag, a required flag left out, a value that does not parse, an argument the
// verb does not take, a flag group), the identity checks, or a command that is not on the surface.
// There were two — a handler's `{verb, role, ok, code, error}` with the role typed again inside
// `error`, and `{ok, code, error}` for everything raised before a handler ran — so the same seat
// got a different object depending on WHERE its call failed, which it cannot know.
//
// THE REAL BINARY, because the shape is a property of what a seat receives: this drives the built
// feov-record once per row and per form, and reads its stdout, its stderr and its exit code.
//
// What each row holds the tool to:
//   - --json: exactly one object on stdout — the keys verb, role, ok, code, error in that order,
//     typed string, string, bool, string, string — nothing on stderr, exit 2;
//   - no --json: nothing on stdout, and on stderr `feov-record: <path>: <diagnosis>`, the path
//     printed once and left out only where no command resolved;
//   - both forms carry the SAME diagnosis, and neither opens it with the role or the path again;
//   - the envelope's `error` holds no help page: a refusal that teaches names the `--help` that
//     prints it;
//   - where the tool logs the refusal, the entry ends in the sentence the seat was shown.
func TestEveryRefusalKindHasOneEnvelope(t *testing.T) {
	bin := testbuild.Binary(t, "feov-record")
	runDir := seatRun(t)
	elsewhere := filepath.Join(t.TempDir(), "another-run")
	if err := os.MkdirAll(elsewhere, 0o755); err != nil {
		t.Fatal(err)
	}
	// AN AGENT THE RECORD BINDS TO THE CHAIR, for the rows where the seat id typed disagrees.
	const boundAgent = "agent_bound_as_the_chair"
	if out, err := execRecord(bin, []string{"FEOV_AGENT_ID=" + boundAgent}, "register", "--run", runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatalf("register the bound agent: %v\n%s", err, out)
	}

	const lens, blue, chair, bench = "red-lens-evidence", "blue-respond", "red-chair", "judge"
	type row struct {
		kind string
		seat string // "" for a call that names none
		env  []string
		args []string

		verb, path, role, code string
		// teaches: the refusal carries a help page in its human form and names it in the JSON one.
		teaches bool
		// short: the refusal answers a flag in a few lines — the same text in both forms, under
		// 1 KB, naming the page and printing none of it.
		short bool
		// logged: the tool writes this refusal to the log — of the seat named, or of logSeat where
		// the record binds the agent to another.
		logged  bool
		logSeat string
	}
	rows := []row{
		// A handler's own refusal, on a seat of each class.
		{kind: "handler", seat: lens, args: []string{"finding", "--quote", "a quoted sentence", "--severity", "high", "--likelihood", "high", "--impact", "high"},
			verb: "finding", path: "finding", role: "lens", code: "missing_field", logged: true},
		{kind: "handler", seat: blue, args: []string{"edit", "--quote", "a quoted sentence", "--new", "x", "--key", "E1"},
			verb: "edit", path: "edit", role: "blue", code: "missing_field", logged: true},
		{kind: "handler", seat: bench, args: []string{"declare", "--reason", ""},
			verb: "declare", path: "declare", role: "bench", code: "missing_field", logged: true},
		{kind: "handler on a nested verb", seat: lens, args: []string{"motion", "petition", "file", "--class", "safety", "--relief", "r"},
			verb: "file", path: "motion petition file", role: "lens", code: "validation", logged: true},
		{kind: "handler on a read", seat: lens, args: []string{"show", "report", "--window", "3"},
			verb: "report", path: "show report", role: "lens", code: "validation", logged: true},
		{kind: "handler on the operator's surface", seat: "operator", args: []string{"graph", "--format", "nonesuch"},
			verb: "graph", path: "graph", role: "", code: "validation"},

		// The record's validation, reached through a handler.
		{kind: "record validation", seat: lens, args: []string{"close", "--id", "G-0000beef", "--reason", "x", "--verified-by", "y", "--verified-with", "z", "--verified-against", "a"},
			verb: "close", path: "close", role: "lens", code: "error", logged: true},

		{kind: "record validation", seat: chair, args: []string{"spot-check", "--ids", "G-0000beef", "--areas", "a", "--reason", "r"},
			verb: "spot-check", path: "spot-check", role: "chair", code: "error", logged: true},
		{kind: "bad value, its message several lines long", seat: chair, args: []string{"verdict", "--as", "MAYBE"},
			verb: "verdict", path: "verdict", role: "chair", code: "validation", short: true, logged: true},

		// A failure nothing coded is not the caller's ill-formed input.
		{kind: "uncoded failure", seat: "operator", args: []string{"fetch", "--url", "http://127.0.0.1:1/x"},
			verb: "fetch", path: "fetch", role: "", code: "error"},

		// Cobra's refusals, each coded by its producer.
		{kind: "unknown flag", seat: lens, args: []string{"close", "--nonesuch", "x"},
			verb: "close", path: "close", role: "lens", code: "validation", short: true, logged: true},
		{kind: "unknown flag", seat: blue, args: []string{"retire", "--key", "k1"},
			verb: "retire", path: "retire", role: "blue", code: "validation", short: true, logged: true},
		{kind: "unknown flag", seat: chair, args: []string{"verdict", "--nonesuch", "x"},
			verb: "verdict", path: "verdict", role: "chair", code: "validation", short: true, logged: true},
		{kind: "unknown flag", seat: bench, args: []string{"declare", "--nonesuch", "x"},
			verb: "declare", path: "declare", role: "bench", code: "validation", short: true, logged: true},
		{kind: "unknown flag on the operator's surface", seat: "operator", args: []string{"graph", "--nonesuch", "x"},
			verb: "graph", path: "graph", role: "", code: "validation", short: true},
		{kind: "missing required flag", seat: lens, args: []string{"close", "--id", "G-0000beef"},
			verb: "close", path: "close", role: "lens", code: "missing_field", short: true, logged: true},
		{kind: "missing required flag", seat: bench, args: []string{"outcome"},
			verb: "outcome", path: "outcome", role: "bench", code: "missing_field", short: true, logged: true},
		{kind: "bad value", seat: lens, args: []string{"finding", "--quote", "q", "--reason", "r", "--severity", "banana"},
			verb: "finding", path: "finding", role: "lens", code: "validation", short: true, logged: true},
		{kind: "wrong arguments", seat: lens, args: []string{"near-match", "extra"},
			verb: "near-match", path: "near-match", role: "lens", code: "validation", short: true, logged: true},
		{kind: "wrong arguments, refused by the verb's own check", seat: chair, args: []string{"dispatch", "later"},
			verb: "dispatch", path: "dispatch", role: "chair", code: "validation", teaches: true, logged: true},
		// A FLAG GROUP has no row: no seat verb declares a group a call can violate (see
		// TestAFlagGroupViolatedIsRefusedWithItsFlagsLines). mint without --problem is now a missing
		// required flag.
		{kind: "missing required flag", seat: lens, args: []string{"mint", "--key", "M-probe", "--class", "scope-creep", "--check-kind", "document",
			"--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low", "--fix", "f", "--reason", "r"},
			verb: "mint", path: "mint", role: "lens", code: "missing_field", short: true, logged: true},

		// Who is asking.
		{kind: "seat id disagrees with the registration", seat: lens, env: []string{"FEOV_AGENT_ID=" + boundAgent},
			args: []string{"log", "--type", "defect", "--reason", "x"},
			verb: "log", path: "log", role: "lens", code: "conflict", logged: true, logSeat: chair},
		{kind: "seat id disagrees with the registration, on a read", seat: blue, env: []string{"FEOV_AGENT_ID=" + boundAgent},
			args: []string{"show", "board"},
			verb: "board", path: "show board", role: "blue", code: "conflict", logged: true, logSeat: chair},
		{kind: "no register", seat: bench, env: []string{"FEOV_AGENT_ID=agent_that_never_registered"},
			args: []string{"log", "--type", "defect", "--reason", "x"},
			verb: "log", path: "log", role: "bench", code: "validation", logged: true},
		{kind: "run disagrees with the dispatch", seat: lens, env: []string{"FEOV_RUN_FROM_WRAPPER=" + elsewhere},
			args: []string{"show", "board"},
			verb: "board", path: "show board", role: "lens", code: "conflict"},
		{kind: "no seat named", seat: "", args: []string{"finding", "--quote", "q"},
			verb: "", path: "", role: "", code: "missing_field", teaches: true},

		// What is on the surface.
		{kind: "unknown command", seat: lens, args: []string{"nonesuch"},
			verb: "", path: "", role: "lens", code: "validation", teaches: true, logged: true},
		{kind: "unknown command", seat: "operator", args: []string{"nonesuch"},
			verb: "", path: "", role: "", code: "validation", teaches: true},
		{kind: "no command", seat: chair, args: nil,
			verb: "", path: "", role: "chair", code: "missing_field", teaches: true, logged: true},
		{kind: "another seat's command", seat: lens, args: []string{"edit", "--quote", "q"},
			verb: "", path: "", role: "lens", code: "role_violation", teaches: true, logged: true},
		{kind: "another seat's act under a group", seat: blue, args: []string{"inquiry", "review", "--id", "A1"},
			verb: "inquiry", path: "inquiry", role: "blue", code: "role_violation", teaches: true, logged: true},
		{kind: "no such act under a group", seat: blue, args: []string{"inquiry", "nonesuch"},
			verb: "inquiry", path: "inquiry", role: "blue", code: "validation", teaches: true, logged: true},
		{kind: "no such view", seat: bench, args: []string{"show", "nonesuch"},
			verb: "show", path: "show", role: "bench", code: "validation", teaches: true, logged: true},
		{kind: "the ruler's verb on a filer's surface", seat: lens, args: []string{"motion", "petition", "rule", "--id", "M1"},
			verb: "petition", path: "motion petition", role: "lens", code: "role_violation", logged: true},
	}

	classes := map[string]bool{}
	for _, r := range rows {
		r := r
		classes[r.role] = true
		name := r.kind + "/" + r.seat
		if r.seat == "" {
			name = r.kind
		}
		t.Run(name, func(t *testing.T) {
			args := []string{"--run", runDir}
			if r.seat != "" {
				args = append(args, "--seat-id", r.seat)
			}
			args = append(args, r.args...)

			// --- the human form
			before := toolRefusals(t, runDir)
			stdout, stderr, code := execForms(t, bin, r.env, args...)
			if code != 2 {
				t.Fatalf("exit %d, want 2 — the call is not refused, so the row measures nothing.\nstdout: %s\nstderr: %s", code, clipTo(stdout, 400), clipTo(stderr, 400))
			}
			if stdout != "" {
				t.Errorf("a refusal without --json wrote to stdout:\n%s", clipTo(stdout, 400))
			}
			head := "feov-record: "
			if r.path != "" {
				head += r.path + ": "
			}
			if !strings.HasPrefix(stderr, head) {
				t.Fatalf("the human form opens %q, want %q", clipTo(stderr, len(head)+40), head)
			}
			human := strings.TrimSuffix(strings.TrimPrefix(stderr, head), "\n")
			// THE PATH IS RESOLVED FROM THE ARGV, the way every other harness resolves it, and has to be
			// the one this row states: a resolver that disagreed with the binary would count a head
			// the binary never printed.
			resolved, n, again := refusalcheck.Rendered(NewRootFor(r.seat), roleWords(), args, stderr)
			if resolved != r.path {
				t.Fatalf("the argv resolves to %q on the %s surface, and the row says %q", resolved, r.seat, r.path)
			}
			if n != 1 || strings.HasPrefix(human, "feov-record: ") {
				t.Errorf("the human form names its command %d times (%s): after %q the diagnosis opens %q", n, strings.Join(again, ", "), head, clipTo(human, 80))
			}
			logged := toolRefusals(t, runDir)[len(before):]

			// --- the JSON form
			jsonOut, jsonErr, code := execForms(t, bin, r.env, append(args, "--json")...)
			if code != 2 {
				t.Fatalf("--json: exit %d, want 2\nstdout: %s", code, clipTo(jsonOut, 400))
			}
			if jsonErr != "" {
				t.Errorf("--json wrote to stderr:\n%s", clipTo(jsonErr, 400))
			}
			env := decodeEnvelope(t, jsonOut)
			if env.Verb != r.verb || env.Role != r.role || env.Code != r.code {
				t.Errorf("envelope verb %q role %q code %q, want %q %q %q\n%s", env.Verb, env.Role, env.Code, r.verb, r.role, r.code, clipTo(jsonOut, 400))
			}
			if env.OK {
				t.Errorf("a refusal's envelope says ok: %s", clipTo(jsonOut, 200))
			}
			if again := refusalcheck.Again(r.path, roleWords(), env.Error); len(again) > 0 {
				t.Errorf("the envelope's error names the command its `verb` already carries (%s): %s", strings.Join(again, ", "), clipTo(env.Error, 200))
			}
			if strings.Contains(env.Error, "\nUsage:") || strings.Contains(env.Error, "\nFlags:") {
				t.Errorf("the envelope's error carries a help page (%d bytes): %s", len(env.Error), clipTo(env.Error, 300))
			}
			if strings.Contains(env.Error, "below") || strings.Contains(firstLine(human), "below") {
				t.Errorf("the diagnosis points at text \"below\", which the JSON form does not carry: %s", clipTo(env.Error, 300))
			}

			// --- one fact, two forms
			if firstLine(env.Error) != firstLine(human) {
				t.Errorf("the two forms give different diagnoses.\n human: %s\n  json: %s", firstLine(human), firstLine(env.Error))
			}
			pointer := "--help` for "
			switch {
			case r.short:
				if !strings.Contains(env.Error, pointer) || env.Error != human || strings.Contains(human, "\nUsage:") || strings.Contains(human, "\nFlags:") || len(stderr) >= 1024 {
					t.Errorf("a flag refusal is a few lines naming the page, the same in both forms (%d bytes).\n human: %s\n  json: %s", len(stderr), clipTo(human, 1200), clipTo(env.Error, 1200))
				}
			case r.teaches && !strings.Contains(env.Error, pointer):
				t.Errorf("the human form carries a help page and the envelope does not name the `--help` that prints it: %s", clipTo(env.Error, 300))
			case r.teaches && !strings.Contains(human, "\nUsage:") && !strings.Contains(human, "\nFlags:") && !strings.Contains(human, "--seat-id IS REQUIRED"):
				t.Errorf("a teaching refusal's human form carries no page:\n%s", clipTo(human, 300))
			case !r.teaches && (strings.Contains(env.Error, pointer) || strings.Contains(human, "\nUsage:")):
				t.Errorf("a refusal that already answers itself was given a help page: %s", clipTo(env.Error, 300))
			case !r.teaches && env.Error != human:
				t.Errorf("the two forms differ.\n human: %s\n  json: %s", clipTo(human, 300), clipTo(env.Error, 300))
			}

			// --- what the tool logged is what the seat was shown
			switch {
			case r.logged && len(logged) != 1:
				t.Errorf("the tool logged %d refusal(s) for this call, want 1: %q", len(logged), logged)
			case !r.logged && len(logged) != 0:
				t.Errorf("the tool logged a refusal this row does not expect — say so in the row: %q", logged)
			case r.logged:
				seatLogged := r.seat
				if r.logSeat != "" {
					seatLogged = r.logSeat
				}
				what := seatLogged + " | refused a verb this surface does not have"
				if r.path != "" {
					what = seatLogged + " | refused `" + r.path + "`"
				}
				// THE ENTRY IS THE SENTENCE AS QUOTED TEXT: the log holds no live anchor, and the one
				// reading that takes the annotation layer out also drops a closing full stop.
				shown := anchortext.Visible(firstLine(human))
				if rs := []rune(shown); len(rs) > 300 {
					shown = string(rs[:300]) + "…"
				}
				// THE WHOLE ENTRY: who, the command once, the flags given by name, and then exactly the
				// sentence shown — nothing in front of it, nothing after.
				flagsGiven, cut := strings.CutPrefix(strings.TrimSuffix(logged[0], ": "+shown), what)
				if !cut || !strings.HasSuffix(logged[0], ": "+shown) || !flagsNamed.MatchString(flagsGiven) {
					t.Errorf("the log entry is not what the seat was shown.\n logged: %s\n  shown: %s[ with --flags]: %s", logged[0], what, shown)
				}
			}
		})
	}
	for _, role := range []string{"lens", "blue", "chair", "bench", ""} {
		if !classes[role] {
			t.Errorf("no row refuses a seat of role %q — every class is held to the one shape", role)
		}
	}
}

// flagsNamed is the part of a refusal's log entry between the command and the sentence: nothing, or
// the flags the call gave, by name.
var flagsNamed = regexp.MustCompile(`^( with( --[a-z-]+)+)?$`)

// refusalKeys is the envelope's key order, which a consumer that streams the object sees.
var refusalKeys = []string{"verb", "role", "ok", "code", "error"}

type envelopeSeen struct {
	Verb  string `json:"verb"`
	Role  string `json:"role"`
	OK    bool   `json:"ok"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

// decodeEnvelope reads stdout as exactly one refusal envelope: those keys, in that order, of
// those types, and nothing after it.
func decodeEnvelope(t *testing.T, stdout string) envelopeSeen {
	t.Helper()
	if strings.Count(strings.TrimSpace(stdout), "\n") != 0 {
		t.Fatalf("stdout is not one line holding one object:\n%s", clipTo(stdout, 600))
	}
	dec := json.NewDecoder(strings.NewReader(stdout))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		t.Fatalf("stdout does not open a JSON object (%v): %s", err, clipTo(stdout, 300))
	}
	var keys []string
	for dec.More() {
		k, err := dec.Token()
		if err != nil {
			t.Fatalf("reading a key: %v\n%s", err, clipTo(stdout, 300))
		}
		keys = append(keys, k.(string))
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("reading the value of %q: %v", k, err)
		}
	}
	if strings.Join(keys, ",") != strings.Join(refusalKeys, ",") {
		t.Fatalf("the envelope's keys are %v, want %v in that order\n%s", keys, refusalKeys, clipTo(stdout, 300))
	}
	var env envelopeSeen
	strict := json.NewDecoder(strings.NewReader(stdout))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&env); err != nil {
		t.Fatalf("the envelope's values are not string, string, bool, string, string: %v\n%s", err, clipTo(stdout, 300))
	}
	if env.Error == "" || env.Code == "" {
		t.Fatalf("the envelope has an empty code or error: %s", clipTo(stdout, 300))
	}
	return env
}

// execForms runs the binary and returns its two streams and its exit code.
func execForms(t *testing.T, bin string, env []string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Args[0] = "feov-record"
	cmd.Env = append(os.Environ(), env...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exit):
		code = exit.ExitCode()
	default:
		t.Fatalf("running %v: %v", args, err)
	}
	return out.String(), errOut.String(), code
}

func execRecord(bin string, env []string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func firstLine(s string) string {
	l, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return l
}

func clipTo(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
