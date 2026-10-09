package migrate_test

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// EVERY ARCHIVED RUN RENDERS, and renders the bytes it rendered before. Replay re-runs placement,
// the splice locators and the retire tidy on every read, so a change to any of them can move a
// historical report while every unit test stays green; this pins the render of every tarball that
// holds a base. No archived run replays a retire, so the retire tidy (reportproj.RemoveAnchorAt)
// is pinned by its own tests and the retire verb's, not by this one.
//
// One golden line per run: `<run> <rendered|no-base> <raw> <normalized> <skeleton> <markers per
// kind>`.
//   - raw is the sha256 of the render: any byte that moves, moves it.
//   - normalized is the sha256 with each anchor token replaced by ⟨kind#n⟩, n its first-appearance
//     ordinal within its kind, so a change that respells ids or tokens and nothing else leaves it.
//   - skeleton is the sha256 of the render with each gap token taken out as a retire takes it out
//     (reportproj.RemoveAnchorAt, husk included), then of the lines that hold prose
//     (claimcount.HasProse, read before the tokens are replaced), each normalized and its whitespace
//     runs folded to one space, so a change that only adds or takes out anchors and their husks
//     leaves it.
//   - the marker counts are the anchor tokens of each kind in the render, and the gap anchors that
//     stand bare of any prose.
//
// The token reader here is the test's own: the readers under test must not grade themselves.
func TestEveryArchivedRunRenders(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	var lines []string
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		run := strings.TrimSuffix(name, ".tar.gz")
		res, dst := migrateArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%s: %d event(s) refused migrating; the first: %+v", run, len(res.Refusals), res.Refusals[0])
		}
		_, haveBase, _, err := record.ReportProjection(dst)
		if err != nil {
			t.Fatalf("%s: reading the report projection: %v", run, err)
		}
		if !haveBase {
			lines = append(lines, run+" no-base - - - -")
			continue
		}
		md, err := reportproj.RenderFromRecord(dst)
		if err != nil {
			t.Fatalf("%s: the archived run does not render: %v", run, err)
		}
		lines = append(lines, fmt.Sprintf("%s rendered %s %s %s %s", run,
			digest(md), digest(normalizeAnchors(md)), digest(skeleton(md)), markerCounts(md)))
	}
	goldentest.Assert(t, "archived_renders", strings.Join(lines, "\n")+"\n")
}

// renderToken is an anchor token of any kind a render can hold.
var renderToken = regexp.MustCompile(`<!--(fx|cite|proof|gap):([FCPG]-[0-9a-f]{8})-->`)

var renderKinds = map[string]string{"fx": "finding", "cite": "citation", "proof": "proof", "gap": "gap"}

func digest(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }

// normalizeAnchors replaces each token with ⟨kind#n⟩, n the id's first-appearance ordinal within
// its kind.
func normalizeAnchors(md string) string {
	ordinal := map[string]int{}
	perKind := map[string]int{}
	return renderToken.ReplaceAllStringFunc(md, func(tok string) string {
		m := renderToken.FindStringSubmatch(tok)
		kind := renderKinds[m[1]]
		key := kind + " " + m[2]
		if _, ok := ordinal[key]; !ok {
			perKind[kind]++
			ordinal[key] = perKind[kind]
		}
		return fmt.Sprintf("⟨%s#%d⟩", kind, ordinal[key])
	})
}

// skeleton takes each gap token out as a retire would, then keeps the lines that hold prose,
// normalized, whitespace runs folded to one space.
func skeleton(md string) string {
	for {
		loc := gapToken.FindStringIndex(md)
		if loc == nil {
			break
		}
		md = reportproj.RemoveAnchorAt(md, loc[0], loc[1]-loc[0])
	}
	var kept []string
	for _, ln := range strings.Split(md, "\n") {
		if claimcount.HasProse(ln) {
			kept = append(kept, ln)
		}
	}
	return strings.Join(strings.Fields(normalizeAnchors(strings.Join(kept, "\n"))), " ")
}

var gapToken = regexp.MustCompile(`<!--gap:G-[0-9a-f]{8}-->`)

func markerCounts(md string) string {
	n := map[string]int{}
	for _, m := range renderToken.FindAllStringSubmatch(md, -1) {
		n[renderKinds[m[1]]]++
	}
	bare := 0
	for _, id := range claimcount.BareAnchorIDs(md) {
		if strings.Contains(md, "<!--gap:"+id+"-->") {
			bare++
		}
	}
	return fmt.Sprintf("finding=%d citation=%d proof=%d gap=%d bare-gap=%d", n["finding"], n["citation"], n["proof"], n["gap"], bare)
}
