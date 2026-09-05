package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A REFUSAL COMES BACK AS AN ERROR. IT DOES NOT EXIT.
//
// `RunE` returns an error precisely so a handler need not decide the process's fate, and Execute
// already maps a returned error to exit 2 — but only AFTER EmitTopLevelError, which is what turns
// a refusal into an envelope for a --json caller. root.go states that contract at length: "A --json
// CALLER GETS JSON, INCLUDING WHEN THE FLAGS THEMSELVES ARE REFUSED… a bare sentence on a channel
// whose entire contract is that it is machine-readable."
//
// Three commands answered a bad argv by printing a usage line and calling os.Exit from inside
// their RunE, so they were the remaining holes in that fix (#716): a consumer parsing the channel
// got a usage sentence and could not see which argument was wrong. The exit also skipped Execute's
// deferred signal-guard release, and made those commands undrivable by any in-process test — which
// is how the class was found, when the fuzz's read-only sweep could convert every surface except
// these.

// TestAnOperatorRefusalIsAnEnvelopeAndNotAnExit drives each fixed command with the argv it used to
// exit on, and asserts the refusal arrives as a coded envelope.
//
// IT IS ALSO ITS OWN TRIPWIRE, and by an unpleasant mechanism worth naming: if os.Exit comes back
// to one of these handlers, this test does not fail — the test BINARY exits, mid-suite. That is
// loud (the package reports a failure) but it names the wrong thing, which is why the structural
// gate below exists alongside it rather than instead of it.
func TestAnOperatorRefusalIsAnEnvelopeAndNotAnExit(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		args []string
		code string
	}{
		{
			// A value present but ill-formed: the chair set is closed, and `banana` is not in it.
			name: "scorecard --chair outside the set",
			args: []string{"scorecard", "--chair", "banana"},
			code: "validation",
		},
		{
			// Both of these refuse on a MISSING positional, before anything opens a run.
			name: "dashboard with one positional",
			args: []string{"dashboard", "only-one"},
			code: "missing_field",
		},
		{
			name: "capture with one positional",
			args: []string{"capture", "only-one"},
			code: "missing_field",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			argv := append([]string{"--json"}, c.args...)
			argv = append(argv, "--run", recordtest.TmpRun(t), "--seat-id", "operator")

			out, err := run(t, argv...)
			if err == nil {
				t.Fatalf("%v was ACCEPTED; a bad argv must be refused: %s", c.args, out)
			}
			var env map[string]any
			if e := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); e != nil {
				t.Fatalf("refusal is not valid JSON (%v) — the usage line is back on the machine-readable channel: %s", e, out)
			}
			if env["ok"] != false {
				t.Errorf("refusal envelope = %v, want ok:false", env)
			}
			// THE CODE, NOT MERELY A CODE. A consumer branches on the KIND of failure, and a
			// refusal that arrives as the generic error is one it cannot tell from any other.
			if env["code"] != c.code {
				t.Errorf("refusal code = %v, want %q (env %v)", env["code"], c.code, env)
			}
			// The usage line is still THERE — this moved which channel carries it, not whether a
			// human is told what the command wanted.
			if msg, _ := env["error"].(string); !strings.Contains(msg, "usage:") {
				t.Errorf("refusal dropped the usage line a human reads: %v", env)
			}
		})
	}
}

// exemptExits are the os.Exit calls in this package that are NOT refusals, each with its reason.
//
// The distinction the gate is drawing: a REFUSAL is the command declining an argv, and it belongs
// in the error channel. An exit STATUS deliberately used as a result — a verdict, or a child
// process's own code passed through — is a different thing, and forcing it through the error
// channel would append a failure line to a command that did not fail.
var exemptExits = map[string]string{
	"capture.go": "the documented `exit 2 iff any audit FAILs` — a verdict on a run that COMPLETED, " +
		"reported after the report is printed, not a refusal of the argv",
	"setup.go": "propagates setup.Run's own exit code, which is the child's result rather than this " +
		"command's refusal",
	"root.go": "Execute itself — the one place that is SUPPOSED to end the process, and the reason " +
		"every handler can simply return",
}

// TestNoHandlerAnswersABadArgvByExiting walks this package's source for os.Exit outside the exempt
// set, so the class cannot come back one file at a time.
//
// WHY A SOURCE WALK, when this repository prefers a behavioural check: an os.Exit cannot be
// OBSERVED from inside the process it ends. The test above dies rather than fails when the defect
// returns, so it cannot be the gate. That is the stated reason a guard exists here rather than a
// generated fact.
//
// AND THE WALK GUARDS ITS OWN EMPTINESS. A pattern that stops matching returns no findings, and no
// findings reads exactly like a clean package — the failure mode this repository has already met
// once, when an Append-census regex went stale and reported full coverage of call sites it had
// stopped seeing. So the walk fails if it finds NO os.Exit anywhere: the exempt files above
// contain three between them, and a walk that cannot see those is not measuring this package.
func TestNoHandlerAnswersABadArgvByExiting(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no source to walk (%v); this gate is measuring nothing", err)
	}
	var offenders, seen []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			// The CALL, not the word: comments in this package discuss os.Exit at length, and a
			// gate that counted those would be unfixable.
			if !strings.Contains(line, "os.Exit(") || strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			seen = append(seen, f)
			if _, ok := exemptExits[filepath.Base(f)]; ok {
				continue
			}
			offenders = append(offenders, f+":"+itoa(i+1)+": "+strings.TrimSpace(line))
		}
	}
	if len(seen) == 0 {
		t.Fatal("found no os.Exit call anywhere in package cli — the exempt files hold three between " +
			"them, so this walk has gone stale and its silence is not a clean result")
	}
	if len(offenders) > 0 {
		t.Errorf("os.Exit inside a handler — a refusal must RETURN so Execute can render it as an "+
			"envelope for a --json caller (#716). Add a reason to exemptExits if the exit is a "+
			"result rather than a refusal:\n  %s", strings.Join(offenders, "\n  "))
	}
}
