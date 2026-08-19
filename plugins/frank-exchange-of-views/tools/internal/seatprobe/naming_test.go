package seatprobe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// THE CONSTITUTIONS NAME NO VERB, AND THAT IS NOW THE INVARIANT.
//
// This asserted the opposite: that the redactor REMOVED names from files that carried them, back
// when a constitution's hand-kept list was the shipped configuration. The finding it was built to
// measure landed — the help page is the only page that instructs — so the shipped bytes ARE the
// `none` arm, and the property worth holding is that they stay that way.
//
// The redactor is kept and still asserted below: it is what makes the arm total against a file
// that acquires a name, and a treatment nobody exercises is one nobody would notice breaking.
func TestTheShippedConstitutionsNameNoVerb(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	for _, name := range []string{"red-auditor.md", "blue-researcher.md", "blue-synthesizer.md", "lead-judge.md"} {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("..", "..", "..", "agents", name))
			if err != nil {
				t.Fatal(err)
			}
			if named := NamesSurviving(string(b), sf); len(named) > 0 {
				var left []string
				for v, n := range named {
					left = append(left, v+"×"+itoa(n))
				}
				t.Errorf("%s names %d verb(s): %s\n\nA constitution that names a slice of the surface does not under-inform a seat, it SATISFIES it — 58%% surface exposure against 95%% with the names removed (2026-08-15), and re-measured 2026-08-19 over two models the arm that names the WHOLE surface is the worst of the four. Name the ACT; the verb is the help's to state.",
					name, len(named), strings.Join(left, ", "))
			}
		})
	}
}

// AND THE REDACTOR IS TOTAL ON TEXT THAT DOES NAME VERBS.
//
// Asserted against a constructed input rather than a shipped one, because the shipped ones no
// longer name anything — which is exactly the vacuous pass this file's own comment warns about.
// The `partial` arm is that constructed input, so the redactor is measured against the treatment
// it has to undo.
func TestRedactionIsTotalOnTheArmThatNames(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	for _, role := range Roles {
		t.Run(role, func(t *testing.T) {
			named := PartialSurfaceBlock(sf, role)
			before := NamesSurviving(named, sf)
			if len(before) == 0 {
				t.Fatalf("the partial arm for %s names no verb, so redacting it is a no-op and `none` would be byte-identical to `partial` — a null result manufactured by the instrument", role)
			}
			after := NamesSurviving(Redact(named, sf), sf)
			if len(after) > 0 {
				var left []string
				for v, n := range after {
					left = append(left, v+"×"+itoa(n))
				}
				t.Errorf("%s still names %d verb(s) after redaction: %s\n\nThe `none` arm would be dispatched with those names in front of the seat, so any difference it measures is against a treatment that did not happen.",
					role, len(after), strings.Join(left, ", "))
			}
		})
	}
}

// AND IT LEAVES THE SITUATION STANDING.
//
// The variable is the NAME, not the clause around it. A redaction that also removed the sentence
// explaining when the act is right would be testing whether constitutions teach at all, which is a
// different and much less interesting question.
func TestRedactionKeepsTheSituationItWithholdsTheVerbFor(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "agents", "red-auditor.md"))
	if err != nil {
		t.Fatal(err)
	}
	out := Redact(string(b), sf)
	// The clause this is drawn from: "A grade moves through `regrade`, and only through it".
	for _, keep := range []string{"A grade moves through", "and only through it"} {
		if !strings.Contains(out, keep) {
			t.Errorf("redaction removed the situation clause %q — it is meant to withhold the verb's name and leave the argument for when the verb is right", keep)
		}
	}
	if len(out) < len(string(b))/2 {
		t.Errorf("redaction cut the constitution from %d to %d bytes — that is a rewrite, not a withheld vocabulary", len(string(b)), len(out))
	}
}

// THE COMPLETE ARM IS GENERATED AND WHOLE.
func TestCompleteArmStatesEveryVerbTheRoleHas(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	for _, role := range Roles {
		block := CompleteSurfaceBlock(sf, role)
		verbs := sf.Verbs(role)
		if len(verbs) == 0 {
			t.Fatalf("%s has no verbs — the surface is empty and every arm would be identical", role)
		}
		// THE BARE FORM, because that is what a seat types: the tree is scoped to the dispatched
		// seat, so `merge mint` is not an invocation any more and rendering it would teach one
		// that cannot run. Asserted per LINE so a verb is not matched inside a longer one.
		for _, v := range verbs {
			if !strings.Contains("\n"+block, "\n  "+v+"\n") {
				t.Errorf("the complete arm for %s omits %q — a `complete` arm that is not complete measures a third, unnamed treatment", role, v)
			}
		}
	}
}

// AND THE ARMS ARE ACTUALLY DIFFERENT DOCUMENTS.
//
// The cheapest way for this experiment to produce a confident nothing is for two arms to be the
// same bytes. Asserted rather than assumed.
func TestTheThreeArmsDiffer(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "agents", "blue-researcher.md"))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]Naming{}
	for _, arm := range NamingArms {
		got := string(Constitution(b, sf, "blue", arm, false))
		if prev, dup := seen[got]; dup {
			t.Fatalf("arms %q and %q render identical bytes — the experiment would report no difference because there is none to report", prev, arm)
		}
		seen[got] = arm
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

// THE DUTY-DELIVERY READER IS ITSELF ASSERTED.
//
// It exists because an unmeasured channel co-varied with a treatment and a result was published on
// top of it. A reader built in that repair and never checked would be the same defect once more:
// its no-match and its honest zero are the same number, and "this seat opened no projection" is a
// real outcome the report prints.
func TestViewReadsCountsTheBareFormAsTheWorkList(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "traj.jsonl")
	// A bare `show` resolves to the role default, which is the work list for every role. Counting
	// it as an unnamed view would undercount the ONE carrier of the duty list — the number this
	// whole measurement exists to report.
	lines := []string{
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show --run /r"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show board"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/x/feov-record blue show work"}}]}}`,
		`{"message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"grep show board /etc/passwd"}}]}}`,
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := ReadViewReads(p, "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	if v.Work != 2 {
		t.Errorf("work list = %d, want 2 (the bare form plus the explicit one)", v.Work)
	}
	if v.Board != 1 {
		t.Errorf("board = %d, want 1", v.Board)
	}
	// The seat's own grep is not a read of this surface, exactly as Attempted scopes by binary.
	if v.Total != 3 {
		t.Errorf("total = %d, want 3 — a command that never invokes the tool is not a projection read", v.Total)
	}
}

// A REDACTION THAT REDACTS NOTHING MUST NOT PASS FOR THE TREATMENT.
//
// The directive arm inverted: it used to APPEND the surface-discovery duty, because that duty
// lived only in debate.js's dispatch prompt and no constitution carried it. It is constitutional
// now, so the arm strips instead. The failure mode of a stripper is silence — it finds no block,
// returns the shipped bytes, and the run reports itself as the treatment arm while measuring the
// control. That is the plausible zero, in the instrument.
func TestStrippingRefusesToBeANoOp(t *testing.T) {
	const body = "# seat\n\nidentity line\n\n" +
		"## Your surface comes from `--help`, and reading it is a first act\n\n" +
		"read it before you act.\n\n" +
		"## Something else\n\nkept.\n"
	if !HasHelpDirective(body) {
		t.Fatal("the fixture does not carry the duty, so this test proves nothing")
	}
	got := StripHelpDirective(body)
	if HasHelpDirective(got) {
		t.Error("the duty survived the strip — the arm would report a treatment it did not apply")
	}
	if !strings.Contains(got, "## Something else") || !strings.Contains(got, "kept.") {
		t.Errorf("the strip took more than the duty:\n%s", got)
	}
	if !strings.Contains(got, "identity line") {
		t.Errorf("the strip took the seat's identity with it:\n%s", got)
	}
}

// AND THE SHIPPED CONSTITUTIONS CARRY IT, so the strip has something to remove. If they ever stop,
// this arm silently becomes a second copy of the control.
func TestTheShippedConstitutionsAreWhatTheArmStrips(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "agents", "*.md"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no constitutions found (%v)", err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !HasHelpDirective(string(b)) {
			t.Errorf("%s carries no surface-discovery duty, so -strip-help-directive is a no-op against it and the arm measures nothing", filepath.Base(p))
		}
	}
}
