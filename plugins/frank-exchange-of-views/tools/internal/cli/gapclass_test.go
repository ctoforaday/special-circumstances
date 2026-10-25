package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// mintClass mints one gap of class at severity through the lens verb, with reason when non-empty,
// and returns the verb's output and error.
func mintClass(t *testing.T, runDir, key, class, severity, reason string) (string, error) {
	t.Helper()
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	args := []string{"--run", runDir, "--seat-id", lensSeat, "--key", key, "--class", class,
		"--problem", "the defect " + key, "--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check",
		"--severity", severity, "--likelihood", severity, "--impact", severity}
	if reason != "" {
		args = append(args, "--reason", reason)
	}
	return runMint(t, args...)
}

// A GAP NO CLASS FITS SAYS WHY (D7). The catch-all without --reason is refused by the record's own
// check, naming the duty; with one it records, and like every placeholder it goes by grade.
func TestUnclassifiedRequiresAReason(t *testing.T) {
	runDir := newRun(t)
	if _, err := mintClass(t, runDir, "k1", "unclassified", "medium", ""); err == nil ||
		!strings.Contains(err.Error(), "a gap no class fits says why in --reason") {
		t.Fatalf("an unclassified mint without --reason must be refused naming the duty: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
		t.Fatalf("the refused mint recorded %d mint(s)", n)
	}
	if _, err := mintClass(t, runDir, "k1", "unclassified", "medium", "nearest is spec_underspecification, which covers a term left undefined; this is a term defined twice"); err != nil {
		t.Fatalf("an unclassified mint with a reason was refused: %v", err)
	}
	m := lastBody(t, runDir, &recordpb.Mint{})
	if m.GetClass() != recordpb.GapClass_GAP_CLASS_UNCLASSIFIED || m.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE {
		t.Errorf("recorded class %s stamped %s, want unclassified stamped by_grade", recordpb.Word(m.GetClass()), recordpb.Word(m.GetClassMaterial()))
	}
}

// EVERY MINT IS STAMPED WITH ITS CLASS'S MATERIAL DEFAULT, through the verb, one per materiality.
func TestAMintTakesItsClassMaterialDefault(t *testing.T) {
	runDir := newRun(t)
	for i, c := range []struct {
		class string
		want  recordpb.ClassMaterial
	}{
		{"false_universal", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS},
		{"structure_noncompliance", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER},
		{"spec_underspecification", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE},
	} {
		if _, err := mintClass(t, runDir, "k"+string(rune('1'+i)), c.class, "medium", ""); err != nil {
			t.Fatalf("mint %s: %v", c.class, err)
		}
		m := lastBody(t, runDir, &recordpb.Mint{})
		if recordpb.Word(m.GetClass()) != c.class || m.GetClassMaterial() != c.want {
			t.Errorf("mint %s: recorded class %s stamped %s, want %s", c.class, recordpb.Word(m.GetClass()),
				recordpb.Word(m.GetClassMaterial()), recordpb.Word(c.want))
		}
	}
}

// A LOW-GRADE PLACEHOLDER GAP DOES NOT HOLD THE GATE. Every fixture's placeholder class goes by
// grade, so an open low gap of it is not material and the chair's PASS lands over it; an `always`
// class at the same grade is material and the same PASS is refused. A stamp that ignored the class
// fails one arm or the other.
func TestALowGradePlaceholderMintDoesNotHoldTheGate(t *testing.T) {
	for _, c := range []struct {
		class  string
		passes bool
	}{{"spec_underspecification", true}, {"false_universal", false}} {
		t.Run(c.class, func(t *testing.T) {
			runDir := castRun(t)
			if _, err := mintClass(t, runDir, "k1", c.class, "low", ""); err != nil {
				t.Fatalf("mint: %v", err)
			}
			chairSitsAgain(t, runDir)
			_, err := run(t, "verdict", "--as", "PASS", "--run", runDir, "--seat-id", "red-chair")
			switch {
			case c.passes && err != nil:
				t.Fatalf("a PASS over one open low %s gap was refused: %v", c.class, err)
			case !c.passes && (err == nil || !strings.Contains(err.Error(), "material")):
				t.Fatalf("a PASS over one open low %s gap must be refused for its materiality: %v", c.class, err)
			}
		})
	}
}

// MINT --HELP LISTS EVERY GAP CLASS IN FULL (D9, F3): each value's word, materiality and definition
// on one line, and its nearest value with the distinguisher on the next. Rendered, not grepped.
// --reason's help on mint carries the catch-all's duty.
func TestMintHelpListsEveryGapClass(t *testing.T) {
	out, err := run(t, "mint", "--help", "--seat-id", lensSeat)
	if err != nil {
		t.Fatalf("mint --help: %v", err)
	}
	lines := strings.Split(out, "\n")
	ed := recordpb.GapClass(0).Descriptor()
	for i := 0; i < ed.Values().Len(); i++ {
		vd := ed.Values().Get(i)
		if vd.Number() == 0 {
			continue
		}
		c := recordpb.GapClass(vd.Number())
		means, err := recordpb.EnumValueDoc(vd)
		if err != nil {
			t.Fatal(err)
		}
		word, md := recordpb.Word(c), recordpb.Word(recordpb.MaterialDefault(c))
		at := -1
		for j, ln := range lines {
			f := strings.Fields(ln)
			if len(f) > 1 && f[0] == word && strings.Contains(ln, "["+md+"] "+means) {
				at = j
				break
			}
		}
		if at < 0 {
			t.Errorf("mint --help lacks %s's line `%s  [%s] %s`", word, word, md, means)
			continue
		}
		if n, ok := recordpb.Neighbor(c); ok {
			want := "nearest: " + recordpb.Word(n) + " — " + recordpb.Distinguisher(c)
			if at+1 >= len(lines) || strings.TrimSpace(lines[at+1]) != want {
				t.Errorf("mint --help: %s's next line must be %q", word, want)
			}
		}
	}
	if !strings.Contains(out, "Under unclassified it is REQUIRED: name the nearest class and what it does not cover") {
		t.Errorf("mint --help's --reason must carry the catch-all's duty:\n%s", out)
	}
}
