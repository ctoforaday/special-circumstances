package migrate

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// A FINDING ID NEVER PRINTS WITHOUT ITS AREA (S9, F-f). The id says nothing of which lens found
// it, so every text view names a finding as `F-7cdcd115 (adversary)`. Over every text view and
// every assembled document of every migrated archived run, each finding id outside an anchor
// token is followed by its lens's area — except where the view is printing a seat's own wording,
// which prints as the seat wrote it (A-4): there the id and the words before it stand in a
// seat-argument field of the record.
func TestEveryFindingIdPrintsItsArea(t *testing.T) {
	finding := regexp.MustCompile(`\b` + anchor.IDPattern("finding") + `\b( \((?:` + strings.Join(record.LensAreas, "|") + `)\))?`)
	named, seatsWords := 0, 0
	for _, name := range archivedRuns(t) {
		toDir := recordtest.TmpRun(t)
		if _, err := Migrate(recordtest.ExtractArchive(t, name), toDir, Entries(), Options{}); err != nil {
			t.Fatalf("migrate %s: %v", name, err)
		}
		run, err := record.OpenRun(toDir)
		if err != nil {
			t.Fatal(err)
		}
		merged, err := record.MergedEvents(run)
		if err != nil || len(merged.Events) == 0 {
			t.Fatalf("%s: reading the migrated record: %d events, %v", name, len(merged.Events), err)
		}
		evs := merged.Events
		var arguments []string
		for _, e := range evs {
			if body, ok := recordpb.Body(e); ok {
				eachString(body.ProtoReflect(), func(field, v string) {
					if argumentFields[field] && finding.MatchString(v) {
						arguments = append(arguments, v)
					}
				})
			}
		}
		texts := map[string]string{}
		for _, v := range view.MarkdownViews() {
			md, err := view.Markdown(run, v, "")
			if err != nil {
				t.Fatalf("%s: view %s: %v", name, v, err)
			}
			texts["view "+v] = string(md)
		}
		docs, err := report.AssembleAll(run)
		if err != nil {
			t.Fatalf("%s: assemble: %v", name, err)
		}
		for _, d := range docs {
			texts[d.File] = d.Body
		}
		for where, text := range texts {
			text = anchor.Replace(text, func(_, _ string) string { return "" })
			for _, m := range finding.FindAllStringSubmatchIndex(text, -1) {
				if m[2] >= 0 {
					named++
					continue
				}
				// The id with the twelve bytes before it on its line, or with the twelve after: a view
				// sets its own words on one side of a seat's, never inside them.
				from := max(m[0]-12, strings.LastIndexByte(text[:m[0]], '\n')+1)
				to := min(m[1]+12, m[1]+len(strings.SplitN(text[m[1]:], "\n", 2)[0]))
				if !containsAny(arguments, text[from:m[1]]) && !containsAny(arguments, text[m[0]:to]) {
					t.Errorf("%s, %s: finding id printed without its area: %q", name, where, text[from:min(len(text), m[1]+40)])
				} else {
					seatsWords++
				}
			}
		}
	}
	if named == 0 {
		t.Error("no finding id printed with its area in any view of any archived run — the views were not read")
	}
	t.Logf("finding ids printed with their area: %d; printed as a seat wrote them: %d", named, seatsWords)
}

func containsAny(texts []string, s string) bool {
	for _, text := range texts {
		if strings.Contains(text, s) {
			return true
		}
	}
	return false
}
