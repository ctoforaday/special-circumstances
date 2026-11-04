package migrate

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// migratedEvents is the named archive's migration: its manifest's refusals checked, its events.
func migratedEvents(t *testing.T, name string) (record.Merged, map[string]int, map[string]int, record.Run) {
	t.Helper()
	res, dir := MigratedArchive(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
	}
	run := runtest.Open(t, dir)
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	return m, res.In, res.Out, run
}

// THE ARCHIVED RAISES ARE WRITTEN AS NO EVENT (§III.9 items 1, 3). A raise named a reading by its
// event and answered nothing a gap does not; the migrated record holds the reading and the gap, and
// the gap's `found_by` keeps every finding but the raise.
func TestTheMigratedM30RaiseIsDropped(t *testing.T) {
	for _, c := range []struct {
		archive, gap string
		raises       []string
		foundBy      []string
	}{
		{"2026-10-19_is-91-prime-m30.tar.gz", "G-dc877265", []string{"F-8410bec7"}, []string{"F-9b6251e9"}},
		{"2026-10-24_is-91-prime-m32.tar.gz", "G-f417112d", []string{"F-d8329c9a", "F-1c990edb", "F-b59925d3", "F-5e7dfd3d"}, nil},
	} {
		t.Run(c.archive, func(t *testing.T) {
			m, in, out, _ := migratedEvents(t, c.archive)
			minted, closed := false, false
			for _, e := range m.Events {
				body, ok := recordpb.Body(e)
				if !ok {
					continue
				}
				switch b := body.(type) {
				case *recordpb.Finding:
					if slices.Contains(c.raises, b.GetId()) {
						t.Errorf("raise %s was written, as a finding about %s %q", b.GetId(), recordpb.Word(b.GetAboutKind()), b.GetAboutRef())
					}
				case *recordpb.Mint:
					if b.GetGapId() == c.gap {
						minted = true
						if !slices.Equal(b.GetFoundBy(), c.foundBy) {
							t.Errorf("%s found_by %q, want %q", c.gap, b.GetFoundBy(), c.foundBy)
						}
					}
				case *recordpb.Close:
					closed = closed || b.GetGapId() == c.gap
				}
			}
			// Both gaps were closed in their archives, and stay closed.
			if !minted || !closed {
				t.Errorf("%s minted %v, closed %v; the archive minted and closed it", c.gap, minted, closed)
			}
			if in["finding"] != out["finding"]+len(c.raises) {
				t.Errorf("findings in %d, out %d: each of the %d raise(s) is read and written as no event", in["finding"], out["finding"], len(c.raises))
			}
		})
	}
}

// A FINDING ABOUT AN ANCHOR IS NO RAISE (§III.9 item 1). A raise is recognised by a ref that names no
// anchor; a finding about a citation, and one about a proof (respelled to the computation's anchor
// before the test runs), come forward as findings about an anchor. Both are seeded into m30's stream,
// after the anchors they name, in the shape its epoch wrote.
func TestAFindingAboutAnAnchorIsNoRaise(t *testing.T) {
	const name = "2026-10-19_is-91-prime-m30.tar.gz"
	from := recordtest.ExtractArchive(t, name)
	evs, err := archivedSource(t, from).Events()
	if err != nil {
		t.Fatal(err)
	}
	finding := func(after OldEvent, id, kind, ref string) OldEvent {
		return OldEvent{ID: 1<<40 + after.ID, SeatID: "red-lens-evidence", Round: after.Round, TS: after.TS,
			Key: "red-lens-evidence:finding:" + id, Word: "finding", Fields: map[string]any{
				"id": id, "finding_key": "", "text": "the sentence overstates its source",
				"severity": "low", "likelihood": "low", "impact": "low", "location": "",
				"about_kind": kind, "about_ref": ref}}
	}
	var stream []OldEvent
	seeded := 0
	for _, old := range evs {
		stream = append(stream, old)
		if old.Word == "proof" && old.Fields["proof_id"] == "P-e8532ebe" {
			stream = append(stream,
				finding(old, "F-0000a1c0", "anchor", "C-73119c8d"),
				finding(old, "F-0000b2d0", "proof", "P-7de24bab"))
			seeded += 2
		}
	}
	if seeded != 2 {
		t.Fatalf("the archive holds no proof P-e8532ebe to seed after")
	}
	to := recordtest.TmpRun(t)
	if err := copyChannels(from, to); err != nil {
		t.Fatal(err)
	}
	dst := runtest.New(t, to)
	res, err := replayInOrder(sliceOf(stream), Entries(), dst, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"F-0000a1c0": "C-73119c8d", "F-0000b2d0": "K-7de24bab"}
	for _, e := range m.Events {
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if !ok {
			continue
		}
		ref, seededHere := want[f.GetId()]
		if !seededHere {
			continue
		}
		delete(want, f.GetId())
		if f.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_ANCHOR || f.GetAboutRef() != ref {
			t.Errorf("finding %s came forward about %s %q, want anchor %q", f.GetId(), recordpb.Word(f.GetAboutKind()), f.GetAboutRef(), ref)
		}
	}
	if len(want) != 0 {
		t.Errorf("findings %v were not written: a finding about an anchor is no raise", want)
	}
}

// M23'S READINGS COME FORWARD AS READINGS (§III.9). m23 raised nothing: its corroborations and its
// readings of citations are carried, none naming a gap, and its refuting corroboration (written as
// event 400) is a reading of the source its lens found.
func TestM23CorroborationReadingComesForward(t *testing.T) {
	m, in, out, _ := migratedEvents(t, "2026-10-13_is-91-prime-m23.tar.gz")
	if in["verify"] != out["verify"] || in["finding"] != out["finding"] {
		t.Errorf("verify in %d out %d, finding in %d out %d: m23 holds no raise, and every reading comes forward",
			in["verify"], out["verify"], in["finding"], out["finding"])
	}
	found, cited := 0, 0
	for _, e := range m.Events {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok {
			continue
		}
		if v.CarriedBy != nil || v.CarrierMinted != nil {
			t.Errorf("verify %d names a carrier (%q, %v); an archived reading names no gap", m.At.Of(e).ID, v.GetCarriedBy(), v.GetCarrierMinted())
		}
		if v.GetAnchor() == "" {
			found++
		} else {
			cited++
		}
		if m.At.Of(e).ID == 400 && (v.GetAnchor() != "" || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES || v.GetUrl() == "") {
			t.Errorf("event 400 is not the refuting reading of a source the lens found: %v", v)
		}
	}
	if found == 0 || cited == 0 {
		t.Errorf("m23 migrated %d reading(s) of found sources and %d of citations; it holds both", found, cited)
	}
}

// labelledArchive is one archive's lens-placed citations: each label a reading of a found source
// minted, with that reading's url and title.
type labelledArchive struct {
	name   string
	labels map[string]labelSource // archived spelling -> source
}

// lensPlacedLabels reads every archive's labelled readings off its source stream — the test's own
// read, not labelSources'.
func lensPlacedLabels(t *testing.T) []labelledArchive {
	t.Helper()
	var out []labelledArchive
	for _, name := range archivedRuns(t) {
		la := labelledArchive{name: name, labels: map[string]labelSource{}}
		for _, old := range sourceEvents(t, name) {
			if l, _ := old.Fields["label"].(string); l != "" && old.Word == "verify" {
				url, _ := old.Fields["url"].(string)
				title, _ := old.Fields["title"].(string)
				la.labels[l] = labelSource{url, title}
			}
		}
		if len(la.labels) > 0 {
			out = append(out, la)
		}
	}
	return out
}

// anyID is an anchor token, or an id in any spelling an archive or the migrated record holds.
var anyID = regexp.MustCompile(`<!--[a-z]+:[A-Za-z0-9-]+-->`)

// spellings is a label in both spellings, the archived lower-case one and the respelled one.
func spellings(label string) []string {
	hex := label[strings.IndexByte(label, '-')+1:]
	return []string{"c-" + hex, "C-" + hex}
}

// THE RED-PLACED CITATIONS LEAVE THE MIGRATED RECORDS, AND THEIR READINGS STAY READINGS (§III.9 item
// 7). Six archives hold 15 labels a lens's reading of a found source placed; on each migrated record
// no anchor, cite or report text holds one, every reading that held or named one reads the label's
// source with no anchor, blue's cites are the only cited sources, and b9's edits that reopened one
// land reopening the rest of what they reopened.
func TestTheMigratedRedCitationsAreReadingsOfTheirSources(t *testing.T) {
	archives := lensPlacedLabels(t)
	var names []string
	total, onlyLabelURLs := 0, 0
	for _, la := range archives {
		names = append(names, strings.SplitN(la.name, "_", 2)[1])
		total += len(la.labels)
	}
	want := []string{"is-91-prime-b7.tar.gz", "is-91-prime-b9.tar.gz", "is-91-prime-m27.tar.gz", "is-91-prime-m29.tar.gz", "is-91-prime-m35.tar.gz", "quadratic-formula.tar.gz"}
	sort.Strings(names)
	if !slices.Equal(names, want) || total != 15 {
		t.Fatalf("labelled archives %v holding %d label(s); the plan's census is %v holding 15", names, total, want)
	}
	for _, la := range archives {
		t.Run(la.name, func(t *testing.T) {
			src := sourceEvents(t, la.name)
			m, in, out, run := migratedEvents(t, la.name)
			dropped := map[string]labelSource{}
			for l, s := range la.labels {
				for _, sp := range spellings(l) {
					dropped[sp] = s
				}
			}
			// A migrated reading is paired with its source by its claim (every token taken out) and
			// access date: the replay serialises sittings, so stream order is not the record's.
			var srcVerify []OldEvent
			citeURLs := map[string]bool{}
			for _, old := range src {
				switch old.Word {
				case "verify":
					srcVerify = append(srcVerify, old)
				case "cite":
					u, _ := old.Fields["url"].(string)
					citeURLs[u] = true
				}
			}
			if in["verify"] != out["verify"] || in["blue_edit"] != out["blue_edit"] {
				t.Fatalf("verify in %d out %d, blue_edit in %d out %d: each comes forward", in["verify"], out["verify"], in["blue_edit"], out["blue_edit"])
			}
			type readingKey struct{ claim, date string }
			verifies := map[readingKey][]*recordpb.Verify{}
			var edits []*recordpb.BlueEdit
			cites := map[string]bool{}
			for _, e := range m.Events {
				body, ok := recordpb.Body(e)
				if !ok {
					continue
				}
				switch b := body.(type) {
				case *recordpb.Verify:
					k := readingKey{strings.TrimSpace(anyID.ReplaceAllString(b.GetClaim(), "")), b.GetAccessDate()}
					verifies[k] = append(verifies[k], b)
				case *recordpb.BlueEdit:
					edits = append(edits, b)
				case *recordpb.Cite:
					cites[b.GetLabel()] = true
					if _, ok := dropped[b.GetLabel()]; ok {
						t.Errorf("cite %s holds a lens-placed label", b.GetLabel())
					}
				case *recordpb.Anchor:
					if _, ok := dropped[b.GetId()]; ok {
						t.Errorf("anchor %s is a lens-placed label", b.GetId())
					}
				}
			}
			reReadings := 0
			for _, old := range srcVerify {
				held, _ := old.Fields["label"].(string)
				named, _ := old.Fields["anchor"].(string)
				s, namedOne := dropped[named]
				if held == "" && !namedOne {
					continue
				}
				if !namedOne {
					s = dropped[held]
				} else {
					reReadings++
				}
				claim, _ := old.Fields["claim"].(string)
				date, _ := old.Fields["access_date"].(string)
				k := readingKey{strings.TrimSpace(anyID.ReplaceAllString(claim, "")), date}
				var match *recordpb.Verify
				for _, v := range verifies[k] {
					if v.GetAnchor() == "" && v.GetUrl() == s.url && v.GetTitle() == s.title {
						match = v
					}
				}
				if match == nil {
					t.Errorf("source verify %d (label %q, anchor %q) came forward as none of %d migrated reading(s) of its claim with no anchor, url %q and title %q",
						old.ID, held, named, len(verifies[k]), s.url, s.title)
				}
			}
			if wantRe := map[string]int{"2026-09-02_quadratic-formula.tar.gz": 3}[la.name]; reReadings != wantRe {
				t.Errorf("%d re-reading(s) of a label; the census holds %d", reReadings, wantRe)
			}
			sources, err := record.CitedSources(run)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range sources {
				if !cites[s.Label] {
					t.Errorf("cited source %s is no blue cite", s.Label)
				}
			}
			// b9's edits 164 and 167 reopened the evidence lens's finding and its label c-d49a32c4; they
			// land reopening exactly the finding and the gap they answer (which the migration of b9's
			// unlabelled edits 114 and 115 adds too) — the label removed, never respelled to "".
			if la.name == "2026-09-11_is-91-prime-b9.tar.gz" {
				landed := 0
				for _, e := range edits {
					if !slices.Contains(e.GetReopened(), "F-84381c8e") {
						continue
					}
					landed++
					if got := e.GetReopened(); !slices.Equal(got, []string{"F-84381c8e", "G-ddb5c38b"}) {
						t.Errorf("a b9 edit reopened %q; want exactly [F-84381c8e G-ddb5c38b], the list without its label", got)
					}
				}
				if landed != 2 {
					t.Errorf("%d b9 edit(s) reopen F-84381c8e; edits 164 and 167 both do", landed)
				}
			}
			// The render holds no label, and the Bibliography no url only a label carried.
			if _, err := report.Assemble(run); err != nil {
				t.Fatalf("assemble: %v", err)
			}
			md, err := os.ReadFile(filepath.Join(run.Dir(), report.FileReport))
			if err != nil {
				t.Fatal(err)
			}
			for sp := range dropped {
				if strings.Contains(string(md), sp) {
					t.Errorf("the assembled report holds %s", sp)
				}
			}
			_, bib, _ := strings.Cut(string(md), "## Bibliography")
			labelURLs := map[string]bool{}
			for _, s := range la.labels {
				labelURLs[s.url] = !citeURLs[s.url]
			}
			for u, only := range labelURLs {
				if !only {
					continue
				}
				onlyLabelURLs++
				if strings.Contains(bib, u) {
					t.Errorf("the Bibliography lists %s, which only a label carried", u)
				}
			}
			blockers, err := record.PassBlockers(run)
			if err != nil {
				t.Fatal(err)
			}
			for _, b := range blockers {
				t.Logf("end blocker: %s %s owner %s", b.Kind, b.Subject, b.Owner)
			}
		})
	}
	// Nine urls only a label carried: the eight the migrated Bibliographies lose (b7 4, m27 2, m29 1,
	// m35 1), and quadratic's web.archive copy of poshenloh.com/quadraticrelated, whose token no
	// archived render of quadratic held.
	if onlyLabelURLs != 9 {
		t.Errorf("%d url(s) only a label carried; the census holds 9", onlyLabelURLs)
	}
}

// NO MIGRATED ARCHIVE HOLDS A LENS-PLACED TOKEN (§III.9 item 7). For each of the 15 labels, in both
// spellings, no string field of any migrated event's body outside argumentFields holds its token, no
// string value or list element is its bare id, and no migrated render holds either. A bare label in
// report text an archived edit wrote (quadratic's "c-95969800, placed by a red auditing seat") is
// prose naming a label, archived text kept as written. It is the gate the silent skip of an id the
// record does not hold in proseRefs lacks.
func TestNoMigratedArchiveHoldsALensPlacedToken(t *testing.T) {
	var all []string
	for _, la := range lensPlacedLabels(t) {
		for l := range la.labels {
			all = append(all, spellings(l)...)
		}
	}
	if len(all) != 30 {
		t.Fatalf("%d label spellings; the census holds 15 labels", len(all)/2)
	}
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			m, _, _, run := migratedEvents(t, name)
			for _, e := range m.Events {
				body, ok := recordpb.Body(e)
				if !ok {
					continue
				}
				eachString(body.ProtoReflect(), func(field, v string) {
					if argumentFields[field] {
						return
					}
					for _, l := range all {
						for _, tok := range []string{"<!--cite:" + l + "-->", "<!--" + l + "-->"} {
							if i := strings.Index(v, tok); i >= 0 {
								t.Errorf("event %d, %s holds %s: %q", m.At.Of(e).ID, field, tok, around(v, i, i+len(tok)))
							}
						}
						if v == l {
							t.Errorf("event %d, %s is %s", m.At.Of(e).ID, field, l)
						}
					}
				})
			}
			has, err := record.ReportHasParts(run)
			if err != nil || !has {
				return
			}
			md, err := record.RenderRun(run)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			for _, l := range all {
				if strings.Contains(md, l) {
					t.Errorf("the render holds %s", l)
				}
			}
		})
	}
}
