package hookgate

import (
	"strings"
	"testing"
)

// PreOutcome — the run-directory injection (#281).
//
// NOTE FOR ANYONE EDITING THIS FILE FROM A SHELL: the deny arm matches its write patterns
// anywhere in a Bash command, so a heredoc containing `cp … blue/report.md` is refused as
// though it were a write. Writing this file through the Write tool works; a `cat <<EOF` does
// not. That is the same mention-vs-invocation confusion the rewrite arm's position matcher
// exists to avoid, still present in the deny arm — recorded here because it was hit, not
// theorised.

const liveRun = "/c/Users/gb/Projects/special-circumstances/research/2026-08-05_smoke"

func bash(t *testing.T, command string) Input {
	t.Helper()
	return mkInput(t, "blue", "Bash", map[string]string{"command": command})
}

// THE REAL COMMAND, from the 2026-08-05 smoke, and the reason the emission is `export …;`
// rather than the inline `VAR=x cmd` form: an inline prefix binds to `cd` and never crosses
// the `&&`, so the seat's actual invocation would have run without it.
func TestTheExportCrossesTheAndAndSeparator(t *testing.T) {
	cmd := `cd C:/Users/gb/Projects/special-circumstances && "/c/bin/feov-record" blue manifest-row --id R1-3 --row "figures recomputed"`
	out, payload := PreOutcome(bash(t, cmd), liveRun)
	if out != OutcomeRewrite {
		t.Fatalf("outcome %v, want rewrite", out)
	}
	if !strings.HasPrefix(payload, "export FEOV_RUN='"+liveRun+"'; ") {
		t.Errorf("the export does not lead the command:\n%s", payload)
	}
	if !strings.HasSuffix(payload, cmd) {
		t.Errorf("the original command was not preserved verbatim:\n%s", payload)
	}
}

// A MENTION IS NOT AN INVOCATION. Every one of these contains the token and none of them runs
// it; a `strings.Contains` rule would have prefixed an export onto documentation writes and
// friction messages, which is the failure this matcher exists to avoid.
func TestMentionsAreNotRewritten(t *testing.T) {
	for _, cmd := range []string{
		`grep -rn "feov-record" plugins/`,
		`echo "run feov-record blue edit next" >> notes.md`,
		`git commit -m "feov-record blue edit now records provenance"`,
		"cat <<'END' > doc.md\nfeov-record blue cite --url ...\nEND",
	} {
		if out, _ := PreOutcome(bash(t, cmd), liveRun); out == OutcomeRewrite {
			t.Errorf("a mention was rewritten, which mutates prose the seat is writing:\n%s", cmd)
		}
	}
}

// Command position in every shape the matcher claims to cover.
func TestInvocationPositionsAreRewritten(t *testing.T) {
	for _, cmd := range []string{
		`feov-record blue position --reason x`,
		`"/c/bin/feov-record" merge mint --id R1-1`,
		`/usr/local/bin/feov-record verify`,
		`true; feov-record lens finding --key F1`,
		`false || feov-record bench opinion --id R1-1`,
		"cd /tmp\nfeov-record blue register",
	} {
		if out, _ := PreOutcome(bash(t, cmd), liveRun); out != OutcomeRewrite {
			t.Errorf("a real invocation was NOT rewritten, so the seat keeps typing the path:\n%s", cmd)
		}
	}
}

// DENY WINS, structurally. A command that both writes the report and invokes the tool must be
// DENIED — a rewrite emitted in its place is a deny that never happened, and the hook protocol
// allows only one document.
func TestDenyBeatsRewrite(t *testing.T) {
	reportPath := "/runs/x/blue/" + "report.md"
	out, payload := PreOutcome(bash(t, `cp draft.md `+reportPath+` && "/c/bin/feov-record" blue edit --key F1`), liveRun)
	if out != OutcomeDeny {
		t.Fatalf("outcome %v, want deny — the blue-report lockdown must not open", out)
	}
	if strings.Contains(payload, "export FEOV_RUN") {
		t.Error("a rewrite leaked into the deny reason")
	}
}

// QUOTING IS A SECURITY BOUNDARY. A newline in the value would terminate the export statement
// and turn the remainder into a command. The run directory comes off disk, and a file is not
// trusted input just because we wrote it once.
func TestHostileRunDirIsRefusedOrNeutralised(t *testing.T) {
	in := bash(t, "feov-record verify")

	if out, _ := PreOutcome(in, "/runs/x\nrm -rf /"); out == OutcomeRewrite {
		t.Error("a run directory containing a newline was injected — that ends the export and runs the rest")
	}
	if out, _ := PreOutcome(in, "/runs/x\x00y"); out == OutcomeRewrite {
		t.Error("a run directory containing a NUL was injected")
	}
	// An apostrophe is legal in a path, so it is ESCAPED rather than refused.
	out, payload := PreOutcome(in, `/runs/it's here`)
	if out != OutcomeRewrite {
		t.Fatalf("outcome %v: an apostrophe must be escaped, not refused", out)
	}
	if !strings.HasPrefix(payload, `export FEOV_RUN='/runs/it'\''s here'; `) {
		t.Errorf("the apostrophe was not escaped as '\\'':\n%s", payload)
	}
}

// Idempotent: a second hook pass, or a seat that copied a rewritten command, must not stack.
func TestAlreadyInjectedIsLeftAlone(t *testing.T) {
	if out, _ := PreOutcome(bash(t, `export FEOV_RUN='/runs/x'; feov-record verify`), liveRun); out == OutcomeRewrite {
		t.Error("the export was stacked onto a command that already carried one")
	}
}

// No marker, no guess. Same posture as InferRunDir: say nothing rather than attach a seat's
// events to the wrong run.
func TestNoRunDirMeansNoRewrite(t *testing.T) {
	if out, _ := PreOutcome(bash(t, "feov-record verify"), ""); out != OutcomeNone {
		t.Error("a rewrite was emitted with no resolved run directory")
	}
}

// Only Bash carries a command. Write/Edit have no shell to inject into.
func TestNonBashToolsAreUntouched(t *testing.T) {
	in := mkInput(t, "blue", "Write", map[string]string{"file_path": "/runs/x/notes.md", "command": "feov-record verify"})
	if out, _ := PreOutcome(in, liveRun); out != OutcomeNone {
		t.Error("a non-Bash tool call was rewritten")
	}
}
