package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY ARCHIVED CLASS LANDS ON A VALUE. The slugs an archive holds are its staged registry's
// classes, every slug a run coined, and every class a mint carried; each resolves onto a GapClass
// value, by the fold or through gapclass.go. Read from the archives rather than from a list kept
// here, so a row of gapclass.go that no mint uses — a registry class nobody minted, a coined slug
// never minted — is still owned: deleting it leaves an archived slug that lands nowhere. Every row
// is used by some archive (a row no archive holds is dead), every key is NOT a current word under
// the fold (else the fold already resolves it and the row is dead), and every value IS one.
func TestEveryArchivedClassLandsOnAValue(t *testing.T) {
	ed := recordpb.GapClass(0).Descriptor()
	for from, to := range archivedGapClasses {
		if _, err := enumNumberOf(ed, from); err == nil {
			t.Errorf("gapclass.go maps %q, which the fold already reads as a GapClass word — the row is dead", from)
		}
		if n, err := enumNumberOf(ed, to); err != nil || n == 0 {
			t.Errorf("gapclass.go maps %q onto %q, which is no GapClass value: %v", from, to, err)
		}
	}
	held := map[string][]string{} // archived slug -> the archives holding it
	for _, name := range archivedRuns(t) {
		dir := recordtest.ExtractArchive(t, name)
		slugs := map[string]bool{}
		if b, err := os.ReadFile(filepath.Join(dir, "records", "class-registry.json")); err == nil {
			var reg struct {
				Classes []struct {
					Slug string `json:"slug"`
				} `json:"classes"`
			}
			if err := json.Unmarshal(b, &reg); err != nil {
				t.Fatalf("%s: the staged registry does not read: %v", name, err)
			}
			for _, c := range reg.Classes {
				slugs[c.Slug] = true
			}
		}
		evs, err := archivedSource(t, dir).Events()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, old := range evs {
			switch old.Word {
			case "mint":
				if s, _ := old.Fields["class"].(string); s != "" {
					slugs[s] = true
				}
			case "class_new":
				if s, _ := old.Fields["slug"].(string); s != "" {
					slugs[s] = true
				}
			}
		}
		for s := range slugs {
			held[s] = append(held[s], name)
		}
	}
	if len(held) == 0 {
		t.Fatal("no archive holds a gap class — nothing here measured the translation")
	}
	var slugs []string
	for s := range held {
		slugs = append(slugs, s)
	}
	sort.Strings(slugs)
	for _, s := range slugs {
		if n, err := enumValue(ed, s); err != nil || n == 0 {
			t.Errorf("the archived class %q (%v) lands on no GapClass value: %v", s, held[s], err)
		}
	}
	for from := range archivedGapClasses {
		if len(held[from]) == 0 {
			t.Errorf("gapclass.go maps %q, which no archive holds — the row is dead", from)
		}
	}
	t.Logf("%d archived class slug(s) across the archives; %d through an authored row", len(slugs), len(archivedGapClasses))
}
