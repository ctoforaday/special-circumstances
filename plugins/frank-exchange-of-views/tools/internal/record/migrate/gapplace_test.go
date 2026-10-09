package migrate_test

import (
	"fmt"
	"path/filepath"
	"slices"
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
		Problem: proto.String("p"), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("c"),
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
	s.add(shapeLens, &recordpb.Finding{Id: proto.String(id),
		Location: proto.String(quote), Text: proto.String("t")})
	s.add(shapeLens, &recordpb.Anchor{Id: proto.String(id), Location: proto.String(quote)})
}

// cite, proof and corroboration are the acts as the archive holds them: recorded, and anchored by
// nothing else.
func (s *shapeSource) cite(id, quote string) *record.Event {
	s.t.Helper()
	ev, err := record.Append(record.Identity{Run: s.run, SeatID: shapeBlue}, &recordpb.Cite{Label: proto.String(id), Location: proto.String(quote),
		Url: proto.String("https://example.org/" + id), Title: proto.String("S"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(),
		WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum()})
	if err != nil {
		s.t.Fatalf("seeding a cite: %v", err)
	}
	return ev
}

func (s *shapeSource) proof(id, quote string) {
	s.t.Helper()
	s.add(shapeBlue, &recordpb.Proof{ProofId: proto.String(id), Location: proto.String(quote)})
}

func (s *shapeSource) corroboration(id, quote string) {
	s.t.Helper()
	s.add(shapeLens, &recordpb.Verify{Label: proto.String(id), Claim: proto.String(quote), Url: proto.String("https://example.org/" + id),
		Title: proto.String("R"), Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("read at the leaf")})
}

func (s *shapeSource) edit(old, new string, exact bool) {
	s.t.Helper()
	b := &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), Text: proto.String("r")}
	if exact {
		b.ExactSpan = proto.Bool(true)
	}
	s.add(shapeBlue, b)
}

// archived spells every id of a migrated text as the source spelled it, so a fixture written in the
// archived spelling states what it expects in the spelling it was written in.
func archived(m *migrate.Manifest, s string) string {
	for old, id := range m.IDs {
		s = strings.ReplaceAll(s, id, old)
	}
	return s
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

// MIGRATION REWRITES NOTHING ELSE. An exact-span edit a gap anchor now stands inside, and a
// placement whose stored location does not place, refuse the run by event and say why.
func TestGapTranslationRefusesWhatItDoesNotRewrite(t *testing.T) {
	const base = "# H\n\nCosts rose sharply in Q1.\n\n(They stand on their own.).\n"
	for _, c := range []struct {
		name, word, says string
		build            func(s *shapeSource)
	}{
		{"an exact-span edit across a gap anchor", "blue_edit", "migration rewrites no exact span", func(s *shapeSource) {
			s.mint("G1", "They stand on their own")
			s.edit("on their own.). ", "on their own.)", true)
		}},
		{"a placement whose location names its section", "anchor", "migration rewrites no stored location", func(s *shapeSource) {
			s.finding("f-0000aaa1", `§ H: "Costs rose sharply in Q1."`)
		}},
		{"a citation whose location names its section", "cite", "migration rewrites no stored location", func(s *shapeSource) {
			s.cite("c-0000aaa1", `§ H: "Costs rose sharply in Q1."`)
		}},
		{"a second act placing one id", "verify", "UNIQUE constraint failed: anchor.id", func(s *shapeSource) {
			s.cite("c-0000aaa1", "Costs rose sharply in Q1.")
			s.corroboration("c-0000aaa1", "Costs rose sharply in Q1.")
		}},
		{"a correction placing again what a retire took out", "cite", `has already recorded a anchor on "C-0000aaa1"`, func(s *shapeSource) {
			cited := s.cite("c-0000aaa1", "Costs rose sharply in Q1.")
			s.add(shapeBlue, &recordpb.Retire{Claim: proto.String("Costs rose sharply in Q1."), Reason: proto.String("refuted"), Anchors: []string{"c-0000aaa1"}})
			fixed := proto.Clone(cited.GetCite()).(*recordpb.Cite)
			fixed.Title = proto.String("S, section 2")
			if _, err := record.Append(record.Identity{Run: s.run, SeatID: shapeBlue,
				Correct: &record.Correct{Type: recordpb.EventType_EVENT_TYPE_CITE, Key: cited.GetKey(), Why: "the title was wrong"}}, fixed); err != nil {
				s.t.Fatalf("seeding the correction: %v", err)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			src := newShapeSource(t, base)
			c.build(src)
			m, err := migrate.Migrate(src.dir, recordtest.TmpRun(t), migrate.Entries(), migrate.Options{})
			if m == nil || err == nil {
				t.Fatalf("migrate = %v, want a manifest and a refusal", err)
			}
			if len(m.Refusals) != 1 || m.Refusals[0].Word != c.word || !strings.Contains(m.Refusals[0].Err, c.says) {
				t.Errorf("refusals = %+v, want one %s saying %q", m.Refusals, c.word, c.says)
			}
		})
	}
}

// AN ARCHIVED CITATION, PROOF AND CORROBORATION PLACED THEIR MARKER BY BEING RECORDED, and each is
// brought forward with one Anchor right after it, from its seat. The place in the stream is what
// keeps the render: the edit that follows the citation carries its anchor, and replays only if the
// Anchor stands before it. A same-sitting correction re-carries its id and gains none; an act that
// named no quote placed nothing and gains none.
func TestPlacementTranslationAnchorsEveryPlacer(t *testing.T) {
	src := newShapeSource(t, "# H\n\nCosts rose sharply in Q1. Volume fell.\n\n- Water is wet.\n- Fire is hot.\n")
	cited := src.cite("c-0000aaa1", "Costs rose sharply in Q1.")
	src.edit("Costs rose sharply in Q1<!--cite:c-0000aaa1-->.", "Costs rose modestly in Q1<!--cite:c-0000aaa1-->.", false)
	src.proof("p-0000aaa1", "Water is wet.")
	src.proof("p-0000aaa2", "")
	fixed := proto.Clone(cited.GetCite()).(*recordpb.Cite)
	fixed.Title = proto.String("S, section 2")
	if _, err := record.Append(record.Identity{Run: src.run, SeatID: shapeBlue,
		Correct: &record.Correct{Type: recordpb.EventType_EVENT_TYPE_CITE, Key: cited.GetKey(), Why: "the title was wrong"}}, fixed); err != nil {
		t.Fatalf("seeding the correction: %v", err)
	}
	src.corroboration("c-0000aaa2", "Fire is hot.")

	m, md, _, dst := src.migrated()
	const want = "# H\n\nCosts rose modestly in Q1<!--cite:c-0000aaa1-->. Volume fell.\n\n- Water is wet<!--proof:p-0000aaa1-->.\n- Fire is hot<!--cite:c-0000aaa2-->.\n"
	if md = archived(m, md); md != want {
		t.Errorf("the migrated render:\n got  %q\n want %q", md, want)
	}
	if sh := m.GapAnchors.Shapes; m.Out["anchor"] != 3 || sh["cite"] != 1 || sh["proof"] != 1 || sh["verify"] != 1 {
		t.Errorf("anchors out = %d, added %+v; want one each for the citation, the proof and the corroboration", m.Out["anchor"], sh)
	}
	merged, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	var followed []string
	for i, e := range merged.Events[1:] {
		if a := e.GetAnchor(); a != nil {
			act := merged.Events[i]
			if e.GetSeatId() != act.GetSeatId() || a.GetId() != act.GetCite().GetLabel()+act.GetProof().GetProofId()+act.GetVerify().GetLabel() {
				t.Errorf("anchor %s (%s) does not follow the act it places: %v", a.GetId(), e.GetSeatId(), act)
			}
			followed = append(followed, a.GetId())
		}
	}
	if !slices.Equal(followed, []string{"C-0000aaa1", "P-0000aaa1", "C-0000aaa2"}) {
		t.Errorf("anchors in stream order = %v", followed)
	}
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
		{"(g) the fallback passes over a heading to the first prose sentence", func(s *shapeSource) {
			s.mint("G1", "Costs rose sharply in Q1.")
			s.edit("Costs rose sharply in Q1. Volume fell.", "## Costs\n\nCosts climbed in Q1. Volume fell.", false)
		}, "Costs climbed in Q1<!--gap:G1-->. Volume fell.", "sentence"},
		{"(h) a replacement whose only prose is a heading", func(s *shapeSource) {
			s.mint("G1", "Costs rose sharply in Q1.")
			s.edit("Costs rose sharply in Q1. Volume fell.", "## Costs", false)
		}, "## Costs.", "unplaced"},
	} {
		t.Run(c.name, func(t *testing.T) {
			src := newShapeSource(t, base)
			c.build(src)
			srcEdits := edits(t, src.run)
			m, md, got, _ := src.migrated()
			if !strings.Contains(archived(m, md), c.render+"\n") {
				t.Errorf("the render does not hold %q:\n%s", c.render, md)
			}
			if m.GapAnchors.Shapes[c.shape] == 0 {
				t.Errorf("the %s shape did not run: %+v", c.shape, m.GapAnchors)
			}
			for i, e := range got {
				if archived(m, stripGapTokens(e.GetOld())) != srcEdits[i].GetOld() || archived(m, stripGapTokens(e.GetNew())) != srcEdits[i].GetNew() {
					t.Errorf("a rewrite inserted more than gap anchors:\n old %q -> %q\n new %q -> %q", srcEdits[i].GetOld(), e.GetOld(), srcEdits[i].GetNew(), e.GetNew())
				}
			}
			if c.shape == "unplaced" && (strings.Contains(md, anchor.Token(m.IDs["G1"])) || !slices.Contains(m.GapAnchors.NeverPlaced, m.IDs["G1"])) {
				t.Errorf("a gap no prose sentence carries stands in the render or goes unnamed: %+v\n%s", m.GapAnchors, md)
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

	// (f) A RUN WRITTEN BY THIS BINARY carries its own anchors: the Anchor after each mint, citation,
	// proof and corroboration, and an edit that carried the gap's. It migrates with one Anchor per
	// id, nothing added, and the same render.
	//
	// Its ids are in the one shape and stand. A run that anchored its own acts under the spelling
	// before it migrates the same way: the source's Anchors name archived ids and the acts it is
	// asked about carry migrated ones, and a placement step that compared the two spellings would
	// add a second Anchor per id, which the record refuses.
	for name, ids := range map[string][4]string{
		"(f) a run written by this binary":               {"C-0000aaa1", "P-0000aaa1", "C-0000aaa2", "G-0000aaa1"},
		"(f) a run that anchored its acts before one id": {"c-0000aaa1", "p-0000aaa1", "c-0000aaa2", "G1"},
	} {
		t.Run(name, func(t *testing.T) {
			src := newShapeSource(t, base)
			gapToken := "<!--gap:" + ids[3] + "-->"
			src.cite(ids[0], "Volume fell.")
			src.add(shapeBlue, &recordpb.Anchor{Id: proto.String(ids[0]), Location: proto.String("Volume fell.")})
			src.proof(ids[1], "Water is wet.")
			src.add(shapeBlue, &recordpb.Anchor{Id: proto.String(ids[1]), Location: proto.String("Water is wet.")})
			src.corroboration(ids[2], "Fire is hot.")
			src.add(shapeLens, &recordpb.Anchor{Id: proto.String(ids[2]), Location: proto.String("Fire is hot.")})
			src.mint(ids[3], "Costs rose sharply in Q1.")
			src.add(shapeLens, &recordpb.Anchor{Id: proto.String(ids[3]), Location: proto.String("Costs rose sharply in Q1.")})
			src.add(shapeBlue, &recordpb.BlueEdit{Old: proto.String("Costs rose sharply in Q1" + gapToken + "."), New: proto.String("Costs rose only modestly in Q1" + gapToken + "."),
				Text: proto.String("accept"), Answers: proto.String(ids[3]), Accepted: proto.Bool(true), AppliedVerbatim: proto.Bool(true), Reopened: []string{ids[3]}})
			m, md, _, _ := src.migrated()
			const want = "# H\n\nCosts rose only modestly in Q1<!--gap:%s-->. Volume fell<!--cite:%s-->.\n\n- Water is wet<!--proof:%s-->.\n- Fire is hot<!--cite:%s-->.\n\n(They stand on their own.).\n"
			if got := archived(m, md); got != fmt.Sprintf(want, ids[3], ids[0], ids[1], ids[2]) {
				t.Errorf("the render moved:\n%q", got)
			}
			if n := m.Out["anchor"]; n != 4 || strings.Count(md, "<!--") != 4 {
				t.Errorf("the migrated run holds %d anchor events for its four placed ids, want 4:\n%s", n, md)
			}
			if len(m.GapAnchors.Shapes) != 0 {
				t.Errorf("a run carrying its own anchors was rewritten: %+v", m.GapAnchors)
			}
		})
	}

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
