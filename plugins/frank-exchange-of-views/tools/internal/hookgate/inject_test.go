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
	// INSIDE THE LIVE RUN, which is now what makes it the lockdown's business. The path used to
	// be an unrelated /runs/x/... and denied anyway, because the gate matched the path SHAPE and
	// never asked whether it was in a run.
	reportPath := liveRun + "/blue/" + "report.md"
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

// THE COMMAND THAT COST RUN 2 ITS BIBLIOGRAPHY.
//
// blue-respond-r1 ran this at tool call 53 of 61. The hook bailed on the heredoc, no
// FEOV_AGENT_ID reached the tool, and the tool answered "this agent has not registered" — which
// was false; it had registered fifty calls earlier. The seat believed the refusal, re-registered,
// rotated its shard nonce, and replay kept the newer shard: 26 events orphaned, including the ten
// citations whose anchors are still in the report and now render as
// "(unresolved citation c-… — no source on the record)".
//
// A heredoc is the form the tool's own help teaches for multi-line prose, so this was the
// recommended path.
func TestTheHeredocFormTheToolTeachesKeepsItsIdentity(t *testing.T) {
	for _, cmd := range []string{
		// the measured one, leading newline and all
		"\n/scratch/runbin/feov-record position \\\n  --reason-file - <<'EOF'\nRed raised five gaps.\nEOF",
		// the same shape a merge seat used, four times
		"cat > /tmp/p.txt << 'EOF'\nprose\nEOF\n/scratch/feov-record merge close --id R1-1 --reason-file /tmp/p.txt",
		// `<<-` and an unquoted delimiter
		"feov-record blue edit --reason-file - <<-END\n\ttext\n\tEND",
	} {
		out, rewritten := PreOutcome(bash(t, cmd), liveRun)
		if out != OutcomeRewrite {
			t.Errorf("a heredoc invocation was NOT rewritten, so the seat loses its identity and "+
				"will be told it never registered:\n%s", cmd)
			continue
		}
		// AND THE PROSE IS UNTOUCHED — the whole safety argument for stripping rather than bailing
		// is that injection only ever prepends.
		if !strings.HasSuffix(rewritten, cmd) {
			t.Errorf("the command body was modified, not merely prefixed:\ngot:  %q\nwant suffix: %q", rewritten, cmd)
		}
	}
}

// And a body that merely DOCUMENTS a verb still must not be treated as an invocation — the case
// the old bail was really protecting, now handled by blanking the body rather than by giving up.
func TestAHeredocBodyIsStillNotAnInvocation(t *testing.T) {
	cmd := "cat <<'END' > doc.md\nfeov-record blue cite --url ...\nEND"
	if out, _ := PreOutcome(bash(t, cmd), liveRun); out == OutcomeRewrite {
		t.Errorf("a heredoc BODY was read as command position:\n%s", cmd)
	}
}

// judge-r2's command, verbatim in shape: the invocation sits inside $( ), which the matcher did
// not count as command position. Same run, same false "you never registered", different opener.
func TestCommandSubstitutionIsCommandPosition(t *testing.T) {
	for _, cmd := range []string{
		"\n# comment\nevidence=$(\"/scratch/runbin/feov-record\" show evidence | grep x)\necho \"$evidence\"",
		"cites=`feov-record show evidence`",
		"( feov-record verify ) || true",
	} {
		if out, _ := PreOutcome(bash(t, cmd), liveRun); out != OutcomeRewrite {
			t.Errorf("an invocation in command-substitution position was not rewritten, so the seat "+
				"loses its identity:\n%s", cmd)
		}
	}
}
