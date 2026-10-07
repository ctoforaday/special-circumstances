package migrate_test

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

const (
	shapeLens = "red-lens-evidence"
	shapeBlue = "blue-respond"
)

// shapeSource is a source record in the archived shape: a base, mints that anchor nothing, and
// edits written against a report that held no gap anchor.
type shapeSource struct {
	t   *testing.T
	dir string
	run record.Run
}

func newShapeSource(t *testing.T, base string) *shapeSource {
	t.Helper()
	dir := recordtest.TmpRun(t)
	s := &shapeSource{t: t, dir: dir, run: runtest.Open(t, dir)}
	s.add(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{shapeLens, shapeBlue}})
	if err := record.StageForRunWithDefaults(s.run, map[string]recordpb.ClassMaterial{"overclaim": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE}); err != nil {
		t.Fatal(err)
	}
	for _, seat := range []string{shapeLens, shapeBlue} {
		if _, _, err := record.RegisterSeat(record.Identity{Run: s.run, SeatID: seat}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	s.add(record.HarnessSeat, &recordpb.BaseIngest{Text: proto.String(base)})
	return s
}

func (s *shapeSource) add(seat string, body proto.Message) {
	s.t.Helper()
	if _, err := record.Append(record.Identity{Run: s.run, SeatID: seat}, body); err != nil {
		s.t.Fatalf("seeding %T: %v", body, err)
	}
}

func (s *shapeSource) mint(id, quote string) {
	s.t.Helper()
	s.add(shapeLens, &recordpb.Mint{GapId: proto.String(id), Class: proto.String("overclaim"), Location: proto.String(quote),
		Problem: proto.String("p " + id), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(), DistinctFrom: s.open()})
}

// open is every gap minted so far, which a fixture's gaps are distinct from by construction.
func (s *shapeSource) open() []string {
	b, err := record.BoardJSONOfRun(s.run)
	if err != nil {
		s.t.Fatal(err)
	}
	var out []string
	for _, g := range b.Open {
		out = append(out, g.ID)
	}
	return out
}

func (s *shapeSource) finding(id, quote string) {
	s.t.Helper()
	s.add(shapeLens, &recordpb.Finding{Label: proto.String("evidence-F" + id[len(id)-1:]), FindingId: proto.String(id),
		Location: proto.String(quote), Text: proto.String("t")})
	s.add(shapeLens, &recordpb.Anchor{Id: proto.String(id), Location: proto.String(quote)})
}

func (s *shapeSource) edit(old, new string, exact bool) {
	s.t.Helper()
	b := &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), Text: proto.String("r")}
	if exact {
		b.ExactSpan = proto.Bool(true)
	}
	s.add(shapeBlue, b)
}

// migrated is the source brought forward, its render, and the edits it carries.
func (s *shapeSource) migrated() (*migrate.Manifest, string, []*recordpb.BlueEdit, record.Run) {
	s.t.Helper()
	to := recordtest.TmpRun(s.t)
	m, err := migrate.Migrate(s.dir, to, migrate.Entries(), migrate.Options{})
	if err != nil || len(m.Refusals) != 0 {
		s.t.Fatalf("migrate: %v %+v", err, resRefusals(m))
	}
	dst := runtest.Open(s.t, to)
	md, err := reportproj.RenderFromRecord(dst)
	if err != nil {
		s.t.Fatalf("the migrated run does not render: %v", err)
	}
	return m, md, edits(s.t, dst), dst
}

func edits(t *testing.T, run record.Run) []*recordpb.BlueEdit {
	t.Helper()
	merged, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []*recordpb.BlueEdit
	for _, e := range merged.Events {
		if b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
			out = append(out, b)
		}
	}
	return out
}

func stripGapTokens(s string) string {
	return anchor.Replace(s, func(tok, id string) string {
		if anchor.Kind(id) == "gap" {
			return ""
		}
		return tok
	})
}

// THE TRANSLATION REWRITES EVERY SHAPE (F-c). One source stream per shape, each in the archived form:
// mints that anchored nothing and edits written against a report that held no gap anchor. Brought
// forward, each renders, each gap anchor lands where the rule puts it, and with gap anchors removed
// every old and new is the source's bytes. Over the archive, the golden counts the rewrites per run
// and shape, and the abutting shape — the commonest — must have run.
func TestGapTranslationRewritesEveryShape(t *testing.T) {
	const base = "# H\n\nCosts rose sharply in Q1. Volume fell.\n\n- Water is wet.\n- Fire is hot.\n\n(They stand on their own.).\n"
	for _, c := range []struct {
		name   string
		build  func(s *shapeSource)
		render string // the line of the render the shape is about
		shape  string
	}{
		{"(a) an edit quoting the gap's sentence without its anchor", func(s *shapeSource) {
			s.mint("G1", "Costs rose sharply in Q1.")
			s.edit("Costs rose sharply in Q1.", "Costs rose modestly in Q1.", false)
		}, "Costs rose modestly in Q1<!--gap:G1-->. Volume fell.", "abutting"},
		{"(a) with a finding anchor already in the run", func(s *shapeSource) {
			s.finding("f-0000aaa1", "Costs rose sharply in Q1.")
			s.mint("G1", "Costs rose sharply in Q1.")
			s.edit("Costs rose sharply in Q1<!--fx:f-0000aaa1-->.", "Costs rose modestly in Q1<!--fx:f-0000aaa1-->.", false)
		}, "Costs rose modestly in Q1<!--gap:G1--><!--fx:f-0000aaa1-->. Volume fell.", "abutting"},
		{"(b) an exact-span edit across the anchor", func(s *shapeSource) {
			s.mint("G1", "They stand on their own")
			s.edit("on their own.).", "on their own.)", true)
			// The edit's sentence is not the anchor's word for word, so the fallback takes its end.
		}, "(They stand on their own.)<!--gap:G1-->", "literal"},
		{"(c) a drop inside the span, its sentence kept", func(s *shapeSource) {
			s.mint("G1", "rose sharply")
			s.edit("Costs rose sharply in Q1.", "Costs rose sharply in Q1. Analysts disagree.", false)
		}, "Costs rose sharply<!--gap:G1--> in Q1. Analysts disagree. Volume fell.", "autoplace"},
		{"(c) a drop inside the span, its sentence rewritten", func(s *shapeSource) {
			s.mint("G1", "rose sharply")
			s.edit("Costs rose sharply in Q1", "Costs climbed in Q1. Then fell", false)
		}, "Costs climbed in Q1<!--gap:G1-->. Then fell. Volume fell.", "sentence"},
		{"(d) a cut to empty", func(s *shapeSource) {
			s.mint("G1", "Volume fell.")
			s.edit("Volume fell.", "", false)
		}, "Costs rose sharply in Q1. <!--gap:G1-->", "bare"},
		{"(e) the cut idiom, then its retire", func(s *shapeSource) {
			s.finding("f-0000aaa1", "Water is wet.")
			s.mint("G1", "Water is wet.")
			s.edit("Water is wet<!--fx:f-0000aaa1-->.", "<!--fx:f-0000aaa1-->", false)
			s.add(shapeBlue, &recordpb.Retire{Claim: proto.String("Water is wet."), Reason: proto.String("refuted"),
				Anchors: []string{"f-0000aaa1"}, RemovalBasis: proto.String(record.RemovalVerified)})
		}, "- <!--gap:G1-->", "bare"},
	} {
		t.Run(c.name, func(t *testing.T) {
			src := newShapeSource(t, base)
			c.build(src)
			srcEdits := edits(t, src.run)
			m, md, got, _ := src.migrated()
			if !strings.Contains(md, c.render+"\n") {
				t.Errorf("the render does not hold %q:\n%s", c.render, md)
			}
			if m.GapAnchors.Shapes[c.shape] == 0 {
				t.Errorf("the %s shape did not run: %+v", c.shape, m.GapAnchors)
			}
			for i, e := range got {
				if stripGapTokens(e.GetOld()) != srcEdits[i].GetOld() || stripGapTokens(e.GetNew()) != srcEdits[i].GetNew() {
					t.Errorf("a rewrite inserted more than gap anchors:\n old %q -> %q\n new %q -> %q", srcEdits[i].GetOld(), e.GetOld(), srcEdits[i].GetNew(), e.GetNew())
				}
			}
			if strings.HasPrefix(c.name, "(e)") {
				// The gap anchor stands bare beside the husk the retire kept, and taken out as a
				// retire would, it leaves the render of the same stream without the mint.
				without := newShapeSource(t, base)
				without.finding("f-0000aaa1", "Water is wet.")
				without.edit("Water is wet<!--fx:f-0000aaa1-->.", "<!--fx:f-0000aaa1-->", false)
				without.add(shapeBlue, &recordpb.Retire{Claim: proto.String("Water is wet."), Reason: proto.String("refuted"),
					Anchors: []string{"f-0000aaa1"}, RemovalBasis: proto.String(record.RemovalVerified)})
				_, plain, _, _ := without.migrated()
				if skeleton(md) != skeleton(plain) {
					t.Errorf("the skeleton moved:\n with the gap %q\n without      %q", md, plain)
				}
			}
		})
	}

	// (f) A RUN WRITTEN BY THIS BINARY carries its own anchors: the mint's Anchor and an edit that
	// carried it. It migrates with one Anchor per gap, nothing added, and the same render.
	t.Run("(f) a run written by this binary", func(t *testing.T) {
		src := newShapeSource(t, base)
		src.mint("G1", "Costs rose sharply in Q1.")
		src.add(shapeLens, &recordpb.Anchor{Id: proto.String("G1"), Location: proto.String("Costs rose sharply in Q1.")})
		src.add(shapeBlue, &recordpb.BlueEdit{Old: proto.String("Costs rose sharply in Q1<!--gap:G1-->."), New: proto.String("Costs rose only modestly in Q1<!--gap:G1-->."),
			Text: proto.String("accept"), Answers: proto.String("G1"), Accepted: proto.Bool(true), AppliedVerbatim: proto.Bool(true), Reopened: []string{"G1"}})
		before, err := reportproj.RenderFromRecord(src.run)
		if err != nil {
			t.Fatal(err)
		}
		m, md, _, dst := src.migrated()
		if md != before {
			t.Errorf("the render moved:\n source %q\n migrated %q", before, md)
		}
		if n := m.Out["anchor"]; n != 1 {
			t.Errorf("the migrated run holds %d anchor events for its one gap, want 1", n)
		}
		if len(m.GapAnchors.Shapes) != 0 {
			t.Errorf("a run carrying its own anchors was rewritten: %+v", m.GapAnchors)
		}
		_ = dst
	})

	t.Run("archive", func(t *testing.T) {
		var lines []string
		total := map[string]int{}
		for _, tb := range archiveTarballs(t) {
			m, _ := migrateArchive(t, filepath.Base(tb))
			var shapes []string
			for k, n := range m.GapAnchors.Shapes {
				shapes = append(shapes, fmt.Sprintf("%s=%d", k, n))
				total[k] += n
			}
			sort.Strings(shapes)
			line := strings.TrimSuffix(filepath.Base(tb), ".tar.gz") + " " + strings.Join(shapes, " ")
			if len(m.GapAnchors.NeverPlaced) > 0 {
				line += " never-placed=" + strings.Join(m.GapAnchors.NeverPlaced, ",")
			}
			if len(m.GapAnchors.Fallback) > 0 {
				line += " fallback=" + strings.Join(m.GapAnchors.Fallback, ",")
			}
			lines = append(lines, strings.TrimSpace(line))
		}
		if total["abutting"] == 0 {
			t.Errorf("the abutting shape counts zero across every archived run — it is the commonest, so the rule never ran")
		}
		goldentest.Assert(t, "archived_gap_rewrites", strings.Join(lines, "\n")+"\n")
	})
}
