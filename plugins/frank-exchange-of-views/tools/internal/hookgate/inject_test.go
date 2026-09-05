package hookgate

import (
	"encoding/json"
	"strings"
	"testing"
)

// PreOutcome — the run-directory and identity injection (#281, #510).

const liveRun = "/c/Users/gb/Projects/special-circumstances/research/2026-08-05_smoke"

func bash(t *testing.T, command string) Input {
	t.Helper()
	raw, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatal(err)
	}
	return Input{ToolName: "Bash", ToolInput: raw}
}

// seatBash is a Bash call from a DISPATCHED SEAT — one carrying both harness identity fields.
//
// bash() deliberately carries neither, which is a main-session call. Tests about the injection's
// per-variable behaviour need a payload with more than one variable on it, or they assert their
// property against a payload that had nothing to inject and pass without exercising anything.
func seatBash(t *testing.T, command string) Input {
	t.Helper()
	in := bash(t, command)
	in.AgentID = "agent_01"
	in.AgentType = "frank-exchange-of-views:red-auditor"
	return in
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

// A MENTION IS NOT AN INVOCATION — AND IT NO LONGER HAS TO BE TOLD APART.
//
// This test used to assert that none of these was rewritten, on the stated grounds that a
// `strings.Contains` rule "would have prefixed an export onto documentation writes and friction
// messages". That reasoning described a harm the mechanism cannot produce: injection PREPENDS.
// `echo "run feov-record next" >> notes.md` writes the same bytes to notes.md whether or not an
// export ran first. The matcher was guarding prose it could never have touched, and paid for the
// guard in seats whose identity it silently dropped (#510).
//
// So the same inputs stay, and the assertion becomes the property that actually matters: the
// command the seat wrote survives verbatim as the tail of whatever runs.
func TestAMentionIsPrefixedAndNeverModified(t *testing.T) {
	for _, cmd := range []string{
		`grep -rn "feov-record" plugins/`,
		`echo "run feov-record blue edit next" >> notes.md`,
		`git commit -m "feov-record blue edit now records provenance"`,
		"cat <<'END' > doc.md\nfeov-record blue cite --url ...\nEND",
		// and a command with no mention at all, which now also carries the identity
		`ls -la research/`,
	} {
		out, payload := PreOutcome(bash(t, cmd), liveRun)
		if out != OutcomeRewrite {
			t.Errorf("outcome %v, want rewrite — every Bash call in a live run carries the run and the identity:\n%s", out, cmd)
			continue
		}
		if !strings.HasSuffix(payload, cmd) {
			t.Errorf("the command body was modified, not merely prefixed:\ngot:  %q\nwant suffix: %q", payload, cmd)
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

// QUOTING IS A SECURITY BOUNDARY. A newline in the value would terminate the export statement
// and turn the remainder into a command. The run directory comes off disk, and a file is not
// trusted input just because we wrote it once.
func TestHostileRunDirIsRefusedOrNeutralised(t *testing.T) {
	in := bash(t, "feov-record verify")

	// THE PROPERTY IS THAT THE HOSTILE VALUE NEVER LANDS, not that nothing else does. This
	// asserted `out != OutcomeRewrite`, which held only while the fixture carried no other
	// variable to inject — the injection is PER-VARIABLE by design, so a refused FEOV_RUN
	// alongside an attested agent type is a rewrite that correctly omits the run. Asserting the
	// outcome tested the fixture; asserting the payload tests the boundary.
	for _, hostile := range []string{"/runs/x\nrm -rf /", "/runs/x\x00y"} {
		out, payload := PreOutcome(seatBash(t, "feov-record verify"), hostile)
		if out == OutcomeRewrite && strings.Contains(payload, "FEOV_RUN") {
			t.Errorf("a run directory containing a control character was injected — that ends the "+
				"export and runs the rest:\n%s", payload)
		}
		if strings.Contains(payload, "rm -rf") || strings.Contains(payload, "\x00") {
			t.Errorf("the hostile value reached the payload:\n%q", payload)
		}
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

// Idempotent PER VARIABLE: a second hook pass, or a seat that copied a rewritten command, must not
// stack the variable it already carries — and must still supply the ones it does not.
//
// This asserted that no rewrite happened at all, which was indistinguishable from the real
// property while the fixture had only one variable to inject. Making it all-or-nothing is the
// documented mistake: it would silently skip the identity of a seat that had copied a command
// carrying FEOV_RUN, and an absent identity reads exactly like a main-session call.
func TestAlreadyInjectedIsNotStacked(t *testing.T) {
	_, payload := PreOutcome(seatBash(t, `export FEOV_RUN='/runs/x'; feov-record verify`), liveRun)
	if n := strings.Count(payload, "export FEOV_RUN="); n > 1 {
		t.Errorf("FEOV_RUN was exported %d times — the prefix stacked:\n%s", n, payload)
	}
	if strings.Contains(payload, "export FEOV_RUN='"+liveRun+"'") {
		t.Errorf("the seat's own FEOV_RUN was overridden by a second export:\n%s", payload)
	}
}

// The variable a command does NOT already carry still arrives. This is the half the all-or-nothing
// shape lost, and the one whose absence is silent.
func TestAMissingVariableIsStillInjectedBesideOneAlreadyPresent(t *testing.T) {
	in := seatBash(t, `export FEOV_RUN='/runs/x'; feov-record verify`)
	out, payload := PreOutcome(in, liveRun)
	if out != OutcomeRewrite {
		t.Fatalf("outcome %v: the agent type was not supplied to a command already carrying the run", out)
	}
	if !strings.Contains(payload, "export FEOV_AGENT_TYPE='frank-exchange-of-views:red-auditor'; ") {
		t.Errorf("the attested agent type did not reach a command that already carried FEOV_RUN:\n%s", payload)
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
	in := Input{ToolName: "Write", ToolInput: []byte(`{"file_path":"/runs/x/notes.md","command":"feov-record verify"}`)}
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

// THE SEAT THAT ALIASES THE BINARY — the shape that cost 21 of 65 registers their agent_id
// across six runs (#510). No matcher that reads the command text can see `$RB` as an
// invocation, which is why none is consulted any more.
func TestAnAliasedBinaryStillCarriesTheIdentity(t *testing.T) {
	for _, cmd := range []string{
		"RB=\"/tmp/x/runbin/feov-record\"\n$RB --seat-id blue-lane-1 register",
		"RB=/tmp/x/runbin/feov-record\n$RB --seat-id blue-lane-1 line-of-inquiry propose --hypothesis h",
		"export RB=/tmp/x/feov-record\n\"$RB\" show board",
		"alias fr='/tmp/x/feov-record'\nfr verify",
	} {
		in := bash(t, cmd)
		in.AgentID = "a241ffe401fe7ebba"
		out, payload := PreOutcome(in, liveRun)
		if out != OutcomeRewrite {
			t.Errorf("an aliased invocation lost its identity, which is the whole of #510:\n%s", cmd)
			continue
		}
		if !strings.Contains(payload, "export FEOV_AGENT_ID='") {
			t.Errorf("no identity in the prefix:\n%s", payload)
		}
		if !strings.HasSuffix(payload, cmd) {
			t.Errorf("the command body was modified, not merely prefixed:\ngot: %q", payload)
		}
	}
}

// judge-r2's command, verbatim in shape: the invocation sits inside $( ), which the old matcher
// did not count as command position — one of the three shapes that cost a seat its identity
// before the matcher was removed outright. Kept as a regression fixture: these are real commands
// from real runs, and they must keep carrying the run and the identity.
func TestTheShapesThatOnceLostTheirIdentity(t *testing.T) {
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
