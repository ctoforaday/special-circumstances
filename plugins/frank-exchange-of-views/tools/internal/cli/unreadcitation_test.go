package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A VERIFICATION THAT READ NOTHING SAYS WHAT IT LEFT AND WHAT IT OWES. The citation stays unread,
// PASS is refused while the report cites it, and the lens that read nothing owes the gap about it,
// outside its budget; the gap minted, the verb names it instead.
func TestAnUnreadVerifySaysTheCitationStaysUnreadAndTheGapIsOwed(t *testing.T) {
	runDir, label := citedClause(t)
	out, err := verifyAs(t, runDir, label, "unreachable")
	if err != nil {
		t.Fatalf("verify --as unreachable: %v", err)
	}
	for _, want := range []string{"citation " + label + " stays unread", "PASS is refused while the report cites it unread",
		"you owe a gap about it", "which takes no mint of your budget"} {
		if !strings.Contains(out, want) {
			t.Errorf("the verify does not say %q:\n%s", want, out)
		}
	}
	mintOut, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--about-kind", "anchor", "--about", label,
		"--class", "spec_underspecification", "--problem", "the source cannot be read", "--fix", "retire the citation or cite a readable source",
		"--check-kind", "document", "--check", "a lens reads the cited source", "--severity", "medium",
		"--likelihood", "medium", "--impact", "medium", "--complexity", "low")
	if err != nil {
		t.Fatalf("mint the gap about the unread citation: %v", err)
	}
	gap := gapID(mintOut)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-logic"); err != nil {
		t.Fatal(err)
	}
	other, err := run(t, "verify", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", label,
		"--as", "unreachable", "--confidence", "high", "--reason", "it would not load")
	if err != nil {
		t.Fatalf("a second lens's unreachable verify: %v", err)
	}
	if want := "gap " + gap + " carries it to blue"; !strings.Contains(other, want) {
		t.Errorf("the second unread verify does not say %q:\n%s", want, other)
	}
}

// --ANCHOR ALONE RETIRES ONE CITATION STANDING ON A SENTENCE, AND NOTHING ELSE. Each refusal names
// what it was given and records nothing; the retire that lands shows in the report's retirements as
// the source a sentence that stays lost — in the changelog — never as a claim that left.
func TestACitationAloneRetiresOnlyACitationOnASentence(t *testing.T) {
	runDir := citedRun(t, "https://sky/1", "https://grass/1")
	sky := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	grass := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
	proof := computeSentence(t, runDir, "Fire is hot.")
	retire := func(anchors string) error {
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", anchors, "--reason", "unreadable")
		return err
	}
	refused := func(name string, err error, wants ...string) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: the retire was recorded", name)
		}
		for _, w := range wants {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%s: the refusal does not say %q:\n%v", name, w, err)
			}
		}
	}
	retires := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE)
	refused("two citations", retire(sky+","+grass), "names 2", sky, grass)
	refused("a proof", retire(proof), proof+" is a proof anchor", "--quote")
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != retires {
		t.Errorf("%d retire(s) recorded by refused retires", n-retires)
	}

	if err := retire(sky); err != nil {
		t.Fatalf("retire --anchor %s: %v", sky, err)
	}
	refused("a retired citation", retire(sky), "the report holds no", sky)
	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	changelog, err := os.ReadFile(filepath.Join(filepath.Dir(a.Report), report.FileChangelog))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(changelog), "**The sky is blue.** — stays in the report; a source it cited was retired from it") {
		t.Errorf("the changelog does not list the sentence that kept its place:\n%s", changelog)
	}

	// A bare citation — an archived edit left it so — leaves with its claim, not alone.
	tok := "<!--cite:" + grass + "-->"
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: blueSeat}, &recordpb.BlueEdit{
		Old: proto.String("The grass is green" + tok + "."), New: proto.String(tok), Text: proto.String("cut"),
	}); err != nil {
		t.Fatalf("the edit that left the citation bare: %v", err)
	}
	refused("a bare citation", retire(grass), grass+" stands bare", "retire that claim with --quote")
}
