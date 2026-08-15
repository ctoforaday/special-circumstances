package seatprobe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// THE REDACTION IS ASSERTED AGAINST THE REAL CONSTITUTIONS, NOT A FIXTURE.
//
// A fixture would prove the regex works on text written to make it work. The arms are dispatched
// with THESE files, so these are what the treatment has to reach — and the residue this reports is
// the honest scope of the `none` arm rather than a claim that it is total.
func TestRedactionRemovesTheNamesFromTheRealConstitutions(t *testing.T) {
	sf := NewSurface(cli.CommandPaths())
	for _, name := range []string{"red-auditor.md", "blue-researcher.md", "blue-synthesizer.md", "lead-judge.md"} {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("..", "..", "..", "agents", name))
			if err != nil {
				t.Fatal(err)
			}
			before := NamesSurviving(string(b), sf)
			// A VACUOUS PASS IS THE FAILURE MODE. If the input names nothing, redacting it
			// removes nothing, the `none` arm equals the `partial` arm, and the experiment
			// reports "naming does not matter" — a null result manufactured by the instrument.
			if len(before) == 0 {
				t.Fatalf("%s names no live verb at all, so redacting it is a no-op and the `none` arm would be byte-identical to `partial`. Either the constitution stopped naming verbs or NamesSurviving no longer matches how they are written — both make this experiment report nothing while looking like it ran", name)
			}
			after := NamesSurviving(Redact(string(b), sf), sf)
			if len(after) > 0 {
				var left []string
				for v, n := range after {
					left = append(left, v+"×"+itoa(n))
				}
				t.Errorf("%s still names %d verb(s) after redaction: %s\n\nThe `none` arm would be dispatched with those names in front of the seat, so any difference it measures is against a treatment that did not happen.",
					name, len(after), strings.Join(left, ", "))
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
		for _, v := range verbs {
			if !strings.Contains(block, role+" "+v) {
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
